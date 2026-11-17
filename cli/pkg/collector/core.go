package collector

import (
	"context"
	"fmt"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreCollector collects namespaces and pods.
type CoreCollector struct {
	client kube.Client
}

func NewCoreCollector(client kube.Client) *CoreCollector {
	return &CoreCollector{client: client}
}

func (c *CoreCollector) Name() string {
	return "core"
}

func (c *CoreCollector) Collect(ctx context.Context) ([]Resource, error) {
	namespaces, err := c.client.GetNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	pods, err := c.client.GetPods(ctx, metav1.NamespaceAll)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	resources := make([]Resource, 0, len(namespaces)+len(pods))
	nsKind := corev1.SchemeGroupVersion.WithKind("Namespace")
	for i := range namespaces {
		r, err := fromTyped(nsKind, &namespaces[i])
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	podKind := corev1.SchemeGroupVersion.WithKind("Pod")
	for i := range pods {
		r, err := fromTyped(podKind, &pods[i])
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}
