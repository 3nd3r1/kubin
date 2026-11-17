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
)

type Manager struct {
	client     kube.Client
	collectors []collector.Collector
	cliVersion string
	now        func() time.Time
}

func NewManager(client kube.Client, cliVersion string) *Manager {
	return &Manager{
		client: client,
		collectors: []collector.Collector{
			collector.NewCoreCollector(client),
			collector.NewEventsCollector(client),
		},
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

	for _, c := range m.collectors {
		resources, err := c.Collect(ctx)
		if err != nil {
			return Manifest{}, fmt.Errorf("collector %s: %w", c.Name(), err)
		}
		for _, r := range resources {
			if err := w.AddResource(r); err != nil {
				return Manifest{}, err
			}
		}
	}

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
	}
	if err := w.Finish(manifest, out); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}
