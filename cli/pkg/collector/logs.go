package collector

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"sync"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	DefaultLogTailLines = 1000
	DefaultLogMaxBytes  = 1 << 20
	DefaultLogWorkers   = 8
)

// LogOptions bound how much log a snapshot carries.
type LogOptions struct {
	// TailLines is the most lines kept per log.
	TailLines int
	// MaxBytes is the most bytes kept per log. It guards against huge
	// lines, which TailLines alone doesn't.
	MaxBytes int
	// Workers is how many logs are fetched at once.
	Workers int
}

// DefaultLogOptions are the limits kubin create uses without flags.
func DefaultLogOptions() LogOptions {
	return LogOptions{TailLines: DefaultLogTailLines, MaxBytes: DefaultLogMaxBytes, Workers: DefaultLogWorkers}
}

// Log is the end of one container's log. Previous is the log of the
// container's last run before a restart. Truncated means older lines were
// dropped to stay within the limits.
type Log struct {
	Namespace string
	Pod       string
	Container string
	Previous  bool
	Data      []byte
	Truncated bool
}

// LogsCollector fetches the logs of containers that have run, and the
// previous log of containers that restarted.
//
// Credentials: logs are kept as the containers wrote them. Kubin can't tell
// a token in a log line from any other text, so nothing is redacted. The
// tail limit keeps what leaks small, and TASK-13 lets users leave logs out.
type LogsCollector struct {
	client kube.Client
	opts   LogOptions
	warn   func(msg string, fields ...any)
}

func NewLogsCollector(client kube.Client, opts LogOptions) *LogsCollector {
	d := DefaultLogOptions()
	if opts.TailLines <= 0 {
		opts.TailLines = d.TailLines
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = d.MaxBytes
	}
	if opts.Workers <= 0 {
		opts.Workers = d.Workers
	}
	return &LogsCollector{client: client, opts: opts, warn: log.Warn}
}

func (c *LogsCollector) Options() LogOptions {
	return c.opts
}

type logJob struct {
	namespace, pod, container string
	previous                  bool
}

// Collect fetches the logs of pods and hands each to emit, one at a time
// and from the calling goroutine. A log that can't be fetched is skipped
// with a warning: a snapshot without one log is still worth sharing. It
// stops at the first error from emit or when ctx is done.
func (c *LogsCollector) Collect(ctx context.Context, pods []corev1.Pod, emit func(Log) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan logJob)
	go func() {
		defer close(jobs)
		for _, j := range logJobs(pods) {
			select {
			case jobs <- j:
			case <-ctx.Done():
				return
			}
		}
	}()

	logs := make(chan Log)
	var wg sync.WaitGroup
	var forbidden sync.Once
	for range c.opts.Workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				l, ok := c.fetch(ctx, j, &forbidden)
				if !ok {
					continue
				}
				select {
				case logs <- l:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(logs)
	}()

	var err error
	for l := range logs {
		if err != nil {
			continue
		}
		if err = emit(l); err != nil {
			cancel()
		}
	}
	if err != nil {
		return err
	}
	return ctx.Err()
}

func (c *LogsCollector) fetch(ctx context.Context, j logJob, forbidden *sync.Once) (Log, bool) {
	// One line more than the limit tells a log that fits from one that
	// doesn't.
	tail := int64(c.opts.TailLines) + 1
	rc, err := c.client.GetPodLogs(ctx, j.namespace, j.pod, corev1.PodLogOptions{
		Container: j.container,
		Previous:  j.previous,
		TailLines: &tail,
	})
	var data []byte
	var truncated bool
	if err == nil {
		data, truncated, err = tailLog(rc, c.opts.TailLines, c.opts.MaxBytes)
		err = errors.Join(err, rc.Close())
	}
	if err == nil && logGone(data) {
		// The kubelet already removed this container. Nothing to warn about:
		// it keeps one dead container per name, so this is routine for the
		// previous log of a container that crashed again.
		return Log{}, false
	}
	if err != nil {
		switch {
		case ctx.Err() != nil:
		case apierrors.IsForbidden(err):
			forbidden.Do(func() { c.warn("Skipping logs: not allowed to read them", "error", err.Error()) })
		default:
			c.warn("Skipping a log", "namespace", j.namespace, "pod", j.pod,
				"container", j.container, "previous", j.previous, "error", err.Error())
		}
		return Log{}, false
	}
	return Log{
		Namespace: j.namespace, Pod: j.pod, Container: j.container, Previous: j.previous,
		Data: data, Truncated: truncated,
	}, true
}

// logGone reports the body the kubelet sends, with status 200, in place of
// a log whose container was garbage collected.
func logGone(data []byte) bool {
	const prefix = "unable to retrieve container logs for "
	return bytes.HasPrefix(data, []byte(prefix)) && len(data) < 256 && bytes.Count(data, []byte{'\n'}) <= 1
}

// logJobs lists the logs to fetch, sorted. A container that never started
// has no log, so it is left out instead of costing a failed request.
func logJobs(pods []corev1.Pod) []logJob {
	var jobs []logJob
	for _, p := range pods {
		statuses := map[string]corev1.ContainerStatus{}
		for _, s := range p.Status.InitContainerStatuses {
			statuses[s.Name] = s
		}
		for _, s := range p.Status.ContainerStatuses {
			statuses[s.Name] = s
		}
		containers := append(append([]corev1.Container{}, p.Spec.InitContainers...), p.Spec.Containers...)
		for _, ctr := range containers {
			s, ok := statuses[ctr.Name]
			if !ok {
				continue
			}
			ran := s.State.Running != nil || s.State.Terminated != nil || s.LastTerminationState.Terminated != nil
			if ran {
				jobs = append(jobs, logJob{namespace: p.Namespace, pod: p.Name, container: ctr.Name})
			}
			if s.RestartCount > 0 {
				jobs = append(jobs, logJob{namespace: p.Namespace, pod: p.Name, container: ctr.Name, previous: true})
			}
		}
	}
	sort.Slice(jobs, func(i, k int) bool {
		a, b := jobs[i], jobs[k]
		if a.namespace != b.namespace {
			return a.namespace < b.namespace
		}
		if a.pod != b.pod {
			return a.pod < b.pod
		}
		if a.container != b.container {
			return a.container < b.container
		}
		return !a.previous && b.previous
	})
	return jobs
}

// tailLog reads r and keeps its last maxLines lines, within maxBytes. When
// it drops anything it reports truncated, and the kept part starts at a
// line boundary unless a single line is longer than maxBytes.
func tailLog(r io.Reader, maxLines, maxBytes int) (data []byte, truncated bool, err error) {
	var buf []byte
	chunk := make([]byte, 32*1024)
	for {
		n, rerr := r.Read(chunk)
		buf = append(buf, chunk[:n]...)
		// Keep memory near maxBytes however long the stream is.
		if len(buf) > 2*maxBytes {
			buf = append(buf[:0], buf[len(buf)-maxBytes:]...)
			truncated = true
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return nil, false, rerr
		}
	}
	if len(buf) > maxBytes {
		buf = buf[len(buf)-maxBytes:]
		truncated = true
	}
	if truncated {
		if i := bytes.IndexByte(buf, '\n'); i >= 0 && i < len(buf)-1 {
			buf = buf[i+1:]
		}
	}

	lines := bytes.Count(buf, []byte{'\n'})
	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		lines++
	}
	for ; lines > maxLines; lines-- {
		i := bytes.IndexByte(buf, '\n')
		buf = buf[i+1:]
		truncated = true
	}
	return buf, truncated, nil
}
