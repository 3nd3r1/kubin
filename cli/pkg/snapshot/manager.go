package snapshot

import (
	"context"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/persister"
)

type Manager struct {
	collectors []collector.Collector
	persister  persister.Persister
}

func NewManager() (*Manager, error) {
	mgr := &Manager{}

	kubeClient, err := kube.NewKubeClient()
	if err != nil {
		return nil, err
	}

	mgr.collectors = []collector.Collector{
		collector.NewCoreCollector(kubeClient),
	}

	mgr.persister, err = persister.NewTarGzPersister()
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (mgr *Manager) CreateSnapshot(ctx context.Context) error {
	for _, c := range mgr.collectors {
		resources, err := c.Collect(ctx)
		if err != nil {
			return err
		}

		for _, resource := range resources {
			mgr.persister.Persist(resource)
		}
	}

    if err := mgr.persister.Finalize("kubin-snapshot.tar.gz"); err != nil {
        return err
    }

    return nil
}
