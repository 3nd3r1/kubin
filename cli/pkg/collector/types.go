package collector

import "context"

type Collector interface {
	Name() string
	Collect(ctx context.Context) ([]ClusterResource, error)
}

type ClusterResource struct {
	Kind     string
	Name     string
	Data     interface{}
	Metadata map[string]string
}
