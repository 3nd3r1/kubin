package collector

import (
	"context"
	"errors"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func ns(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func pod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
}

func TestCoreCollectorCollectsNamespacesAndPods(t *testing.T) {
	cs := fake.NewClientset(ns("default"), ns("prod"), pod("default", "web"), pod("prod", "web"), pod("prod", "db"))
	got, err := NewCoreCollector(kube.NewFromClientset(cs, "", "")).Collect(context.Background())
	require.NoError(t, err)

	type key struct{ kind, ns, name string }
	var keys []key
	for _, r := range got {
		keys = append(keys, key{r.Kind, r.Namespace, r.Name})
		assert.Equal(t, "", r.Group)
		assert.Equal(t, "v1", r.Version)
		assert.Equal(t, "v1", r.Object["apiVersion"], "apiVersion set on %s", r.Name)
		assert.Equal(t, r.Kind, r.Object["kind"])
	}
	assert.ElementsMatch(t, []key{
		{"Namespace", "", "default"},
		{"Namespace", "", "prod"},
		{"Pod", "default", "web"},
		{"Pod", "prod", "web"},
		{"Pod", "prod", "db"},
	}, keys)
}

func TestCoreCollectorReturnsListErrors(t *testing.T) {
	for _, resource := range []string{"namespaces", "pods"} {
		t.Run(resource, func(t *testing.T) {
			cs := fake.NewClientset(ns("default"))
			cs.PrependReactor("list", resource, func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New("forbidden")
			})
			_, err := NewCoreCollector(kube.NewFromClientset(cs, "", "")).Collect(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), "forbidden")
		})
	}
}
