// Package persister manages the persistation of snapshots
package persister

import (
	"github.com/3nd3r1/kubin/cli/pkg/collector"
)

type Persister interface {
	Persist(resource collector.ClusterResource) error
	Finalize() error
}
