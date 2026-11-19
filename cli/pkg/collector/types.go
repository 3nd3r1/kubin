// Package collector reads objects from the cluster for a snapshot.
package collector

import (
	"context"
)

type Collector interface {
	Name() string
	Collect(ctx context.Context) ([]Resource, error)
}

// Resource is one object from the cluster. Group is "" for the core group.
// Namespace is "" for cluster-scoped objects. Object is the object as JSON
// would decode it, with apiVersion and kind set.
type Resource struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	Name      string
	Object    map[string]any
}
