package collector

import (
	"context"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
)

type CoreCollector struct {
	client kube.Client
}

func NewCoreCollector(client kube.Client) *CoreCollector {
	return &CoreCollector{client: client}
}

func (c *CoreCollector) Name() string {
	return "core"
}

func (c *CoreCollector) Collect(ctx context.Context) ([]ClusterResource, error) {
	var resources []ClusterResource

	namespaces, err := c.collectNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	resources = append(resources, namespaces...)

	return resources, nil
}

func (c *CoreCollector) collectNamespaces(ctx context.Context) ([]ClusterResource, error) {
	var resources []ClusterResource

	namespaces, err := c.client.GetNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces {
		resources = append(resources, ClusterResource{
			Kind:     "namespace",
			Name:     namespace.Name,
			Data:     namespace,
			Metadata: nil,
		})
	}

	return resources, nil
}
