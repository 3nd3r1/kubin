// Package collector reads objects from the cluster for a snapshot.
package collector

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// fromTyped converts a typed object, like a corev1.Pod, into a Resource.
// List calls leave apiVersion and kind empty on items, so they come from gvk.
func fromTyped(gvk schema.GroupVersionKind, obj runtime.Object) (Resource, error) {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return Resource{}, fmt.Errorf("convert %s: %w", gvk.Kind, err)
	}
	u := unstructured.Unstructured{Object: m}
	u.SetAPIVersion(gvk.GroupVersion().String())
	u.SetKind(gvk.Kind)
	return Resource{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
		Object:    u.Object,
	}, nil
}
