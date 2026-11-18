// Package snapshot builds snapshots in the format in docs/snapshot-format.md.
package snapshot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Manager struct {
	client     kube.Client
	collectors []collector.Collector
	logs       *collector.LogsCollector
	cliVersion string
	now        func() time.Time
}

func NewManager(client kube.Client, cliVersion string, logs collector.LogOptions) *Manager {
	return &Manager{
		client: client,
		collectors: []collector.Collector{
			collector.NewCoreCollector(client),
			collector.NewEventsCollector(client),
		},
		logs:       collector.NewLogsCollector(client, logs),
		cliVersion: cliVersion,
		now:        time.Now,
	}
}

// Create collects the cluster and writes the snapshot to out.
func (m *Manager) Create(ctx context.Context, out io.Writer) (_ Manifest, err error) {
	createdAt := m.now().UTC().Truncate(time.Second)

	info, err := m.client.ClusterInfo(ctx)
	if err != nil {
		return Manifest{}, err
	}

	w, err := NewWriter()
	if err != nil {
		return Manifest{}, err
	}
	defer func() { err = errors.Join(err, w.Cleanup()) }()

	var pods []corev1.Pod
	for _, c := range m.collectors {
		resources, err := c.Collect(ctx)
		if err != nil {
			return Manifest{}, fmt.Errorf("collector %s: %w", c.Name(), err)
		}
		for _, r := range resources {
			if r.Group == "" && r.Kind == "Pod" {
				var p corev1.Pod
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.Object, &p); err != nil {
					return Manifest{}, fmt.Errorf("pod %s/%s: %w", r.Namespace, r.Name, err)
				}
				pods = append(pods, p)
			}
			if err := w.AddResource(r); err != nil {
				return Manifest{}, err
			}
		}
	}

	// Logs come from the pods in the snapshot, so every log has its pod.
	if err := m.logs.Collect(ctx, pods, w.AddLog); err != nil {
		return Manifest{}, fmt.Errorf("collect logs: %w", err)
	}
	logOpts := m.logs.Options()

	manifest := Manifest{
		FormatVersion: FormatVersion,
		CreatedAt:     createdAt,
		CLIVersion:    m.cliVersion,
		Cluster: Cluster{
			Context:       info.Context,
			Name:          info.Name,
			ServerVersion: info.ServerVersion,
		},
		Resources: w.Counts(),
		Logs: Logs{
			TailLines: logOpts.TailLines,
			MaxBytes:  logOpts.MaxBytes,
			Files:     w.Logs(),
		},
	}
	if err := w.Finish(manifest, out); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}
