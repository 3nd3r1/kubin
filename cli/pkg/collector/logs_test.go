package collector

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestTailLog(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		lines     int
		bytes     int
		want      string
		truncated bool
	}{
		{"empty", "", 3, 100, "", false},
		{"fits", "a\nb\nc\n", 3, 100, "a\nb\nc\n", false},
		{"fits without final newline", "a\nb\nc", 3, 100, "a\nb\nc", false},
		{"one line too many", "a\nb\nc\nd\n", 3, 100, "b\nc\nd\n", true},
		{"too many without final newline", "a\nb\nc\nd", 3, 100, "b\nc\nd", true},
		{"byte cap cuts at a line start", "aaaa\nbbbb\ncc\n", 10, 8, "cc\n", true},
		{"one line over the byte cap", strings.Repeat("x", 20) + "\n", 10, 8, strings.Repeat("x", 7) + "\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, truncated, err := tailLog(strings.NewReader(tt.in), tt.lines, tt.bytes)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}

func TestTailLogKeepsTheEndOfALongStream(t *testing.T) {
	var b strings.Builder
	for i := range 100_000 {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	got, truncated, err := tailLog(strings.NewReader(b.String()), 1_000_000, 64)
	require.NoError(t, err)
	assert.True(t, truncated)
	assert.LessOrEqual(t, len(got), 64)
	assert.True(t, strings.HasSuffix(string(got), "line 99999\n"))
	assert.True(t, strings.HasPrefix(string(got), "line "), "starts at a line: %q", got)
}

// logClient serves logs from a function, counting how many run at once.
type logClient struct {
	kube.Client
	logs func(ns, pod string, opts corev1.PodLogOptions) (string, error)

	mu      sync.Mutex
	opts    []corev1.PodLogOptions
	running atomic.Int32
	peak    atomic.Int32
}

func (c *logClient) GetPodLogs(_ context.Context, ns, pod string, opts corev1.PodLogOptions) (io.ReadCloser, error) {
	n := c.running.Add(1)
	defer c.running.Add(-1)
	for {
		p := c.peak.Load()
		if n <= p || c.peak.CompareAndSwap(p, n) {
			break
		}
	}
	time.Sleep(time.Millisecond)

	c.mu.Lock()
	c.opts = append(c.opts, opts)
	c.mu.Unlock()
	s, err := c.logs(ns, pod, opts)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(s)), nil
}

type warnings struct {
	mu   sync.Mutex
	msgs []string
}

func (w *warnings) warn(msg string, _ ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.msgs = append(w.msgs, msg)
}

func runningPod(ns, name string, containers ...string) corev1.Pod {
	p := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}
	for _, c := range containers {
		p.Spec.Containers = append(p.Spec.Containers, corev1.Container{Name: c})
		p.Status.ContainerStatuses = append(p.Status.ContainerStatuses, corev1.ContainerStatus{
			Name: c, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
		})
	}
	return p
}

func collectLogs(t *testing.T, c *LogsCollector, pods ...corev1.Pod) []Log {
	t.Helper()
	var got []Log
	require.NoError(t, c.Collect(context.Background(), pods, func(l Log) error {
		got = append(got, l)
		return nil
	}))
	return got
}

func TestLogJobs(t *testing.T) {
	crashed := corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}
	waiting := corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "web"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "init"}},
			Containers:     []corev1.Container{{Name: "app"}, {Name: "pending"}, {Name: "nostatus"}},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{{Name: "init", State: crashed}},
			ContainerStatuses: []corev1.ContainerStatus{
				// Crash looping: waiting now, ran before.
				{Name: "app", State: waiting, LastTerminationState: crashed, RestartCount: 4},
				{Name: "pending", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}}},
			},
		},
	}
	assert.Equal(t, []logJob{
		{namespace: "prod", pod: "web", container: "app"},
		{namespace: "prod", pod: "web", container: "app", previous: true},
		{namespace: "prod", pod: "web", container: "init"},
	}, logJobs([]corev1.Pod{p}))
}

func TestLogsCollectorTailsAndMarksTruncation(t *testing.T) {
	client := &logClient{logs: func(_, pod string, opts corev1.PodLogOptions) (string, error) {
		// The API server honours TailLines.
		lines := []string{"1", "2", "3", "4", "5"}
		if n := int(*opts.TailLines); n < len(lines) {
			lines = lines[len(lines)-n:]
		}
		if pod == "short" {
			lines = lines[len(lines)-2:]
		}
		return strings.Join(lines, "\n") + "\n", nil
	}}
	c := NewLogsCollector(client, LogOptions{TailLines: 3})
	got := collectLogs(t, c, runningPod("prod", "long", "app"), runningPod("prod", "short", "app"))

	byPod := map[string]Log{}
	for _, l := range got {
		byPod[l.Pod] = l
	}
	assert.Equal(t, "3\n4\n5\n", string(byPod["long"].Data))
	assert.True(t, byPod["long"].Truncated)
	assert.Equal(t, "4\n5\n", string(byPod["short"].Data))
	assert.False(t, byPod["short"].Truncated)
	for _, o := range client.opts {
		assert.Equal(t, int64(4), *o.TailLines, "asks for one line more than it keeps")
		assert.Equal(t, "app", o.Container)
	}
}

func TestLogsCollectorBoundsWorkers(t *testing.T) {
	var pods []corev1.Pod
	for i := range 40 {
		pods = append(pods, runningPod("prod", fmt.Sprintf("web-%d", i), "app", "sidecar"))
	}
	client := &logClient{logs: func(_, pod string, _ corev1.PodLogOptions) (string, error) { return pod + "\n", nil }}
	got := collectLogs(t, NewLogsCollector(client, LogOptions{Workers: 3}), pods...)
	assert.Len(t, got, 80)
	assert.LessOrEqual(t, client.peak.Load(), int32(3))
	assert.Greater(t, client.peak.Load(), int32(1), "fetches run in parallel")
}

func TestLogsCollectorSkipsFailedLogs(t *testing.T) {
	forbidden := apierrors.NewForbidden(schema.GroupResource{Resource: "pods/log"}, "x", errors.New("no"))
	client := &logClient{logs: func(_, pod string, _ corev1.PodLogOptions) (string, error) {
		switch pod {
		case "denied-1", "denied-2":
			return "", forbidden
		case "gone":
			return "", errors.New("container not found")
		case "collected":
			return "unable to retrieve container logs for containerd://2e97b0ce", nil
		}
		return "ok\n", nil
	}}
	c := NewLogsCollector(client, LogOptions{})
	var w warnings
	c.warn = w.warn

	got := collectLogs(t, c, runningPod("a", "denied-1", "app"), runningPod("a", "denied-2", "app"),
		runningPod("a", "gone", "app"), runningPod("a", "collected", "app"), runningPod("a", "fine", "app"))
	require.Len(t, got, 1)
	assert.Equal(t, "fine", got[0].Pod)
	assert.ElementsMatch(t, []string{"Skipping logs: not allowed to read them", "Skipping a log"}, w.msgs,
		"forbidden warns once, a garbage-collected container not at all")
}

func TestLogsCollectorStopsOnEmitError(t *testing.T) {
	var pods []corev1.Pod
	for i := range 50 {
		pods = append(pods, runningPod("prod", fmt.Sprintf("web-%d", i), "app"))
	}
	client := &logClient{logs: func(string, string, corev1.PodLogOptions) (string, error) { return "x\n", nil }}
	boom := errors.New("disk full")
	calls := 0
	err := NewLogsCollector(client, LogOptions{Workers: 2}).Collect(context.Background(), pods, func(Log) error {
		calls++
		return boom
	})
	require.ErrorIs(t, err, boom)
	assert.Equal(t, 1, calls)
	assert.Less(t, len(client.opts), 50, "stops fetching")
}
