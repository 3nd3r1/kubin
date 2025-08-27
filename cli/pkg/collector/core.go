// Package collector contains all the logic for collecting data from the cluster
package collector

import (
	"context"
	"fmt"

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

	pods, err := c.collectPods(ctx)
	if err != nil {
		return nil, err
	}
	resources = append(resources, pods...)

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

func (c *CoreCollector) collectPods(ctx context.Context) ([]ClusterResource, error) {
	var resources []ClusterResource

	namespaces, err := c.client.GetNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces for pod collection: %w", err)
	}

	for _, namespace := range namespaces {
		pods, err := c.client.GetPods(ctx, namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods from namespace %s: %w", namespace.Name, err)
		}

		for _, pod := range pods {
			metadata := map[string]string{
				"namespace": pod.Namespace,
			}

			resources = append(resources, ClusterResource{
				Kind:     "pod",
				Name:     pod.Name,
				Data:     pod,
				Metadata: metadata,
			})
		}
	}

	return resources, nil
}
