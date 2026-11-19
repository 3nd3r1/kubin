package collector

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube/kubetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	eventsv1 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	k8stesting "k8s.io/client-go/testing"
)

func meta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Namespace: namespace, Name: name}
}

// one object of every kind in Kinds, plus kinds kubin leaves out.
func cluster() []runtime.Object {
	return []runtime.Object{
		&corev1.Namespace{ObjectMeta: meta("", "shop")},
		&corev1.Node{ObjectMeta: meta("", "node-1")},
		&corev1.Pod{ObjectMeta: meta("shop", "api-1")},
		&corev1.Event{ObjectMeta: meta("shop", "api-1.1"), Reason: "BackOff"},
		&corev1.Service{ObjectMeta: meta("shop", "api")},
		&corev1.ConfigMap{ObjectMeta: meta("shop", "api")},
		&corev1.PersistentVolumeClaim{ObjectMeta: meta("shop", "data")},
		&corev1.PersistentVolume{ObjectMeta: meta("", "pv-1")},
		&appsv1.Deployment{ObjectMeta: meta("shop", "api")},
		&appsv1.ReplicaSet{ObjectMeta: meta("shop", "api-7d9f")},
		&appsv1.StatefulSet{ObjectMeta: meta("shop", "db")},
		&appsv1.DaemonSet{ObjectMeta: meta("kube-system", "proxy")},
		&batchv1.Job{ObjectMeta: meta("shop", "migrate")},
		&batchv1.CronJob{ObjectMeta: meta("shop", "backup")},
		&networkingv1.Ingress{ObjectMeta: meta("shop", "api")},
		&discoveryv1.EndpointSlice{ObjectMeta: meta("shop", "api-x7k2p")},

		&corev1.Secret{ObjectMeta: meta("shop", "db"), Data: map[string][]byte{"password": []byte("x")}},
		&corev1.ServiceAccount{ObjectMeta: meta("shop", "api")},
		&networkingv1.NetworkPolicy{ObjectMeta: meta("shop", "deny")},
		&eventsv1.Event{ObjectMeta: meta("shop", "api-1.2")},
	}
}

type key struct{ group, version, kind, ns, name string }

func keys(rs []Resource) []key {
	var out []key
	for _, r := range rs {
		out = append(out, key{r.Group, r.Version, r.Kind, r.Namespace, r.Name})
	}
	return out
}

type warnLog []string

func (w *warnLog) warn(msg string, fields ...any) { *w = append(*w, fmt.Sprint(msg, fields)) }

func TestResourcesCollectorCollectsAllowedKinds(t *testing.T) {
	c := NewResourcesCollector(kubetest.New(cluster()...), Scope{})
	var warnings warnLog
	c.warn = warnings.warn

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.ElementsMatch(t, []key{
		{"", "v1", "Namespace", "", "shop"},
		{"", "v1", "Node", "", "node-1"},
		{"", "v1", "Pod", "shop", "api-1"},
		{"", "v1", "Event", "shop", "api-1.1"},
		{"", "v1", "Service", "shop", "api"},
		{"", "v1", "ConfigMap", "shop", "api"},
		{"", "v1", "PersistentVolumeClaim", "shop", "data"},
		{"", "v1", "PersistentVolume", "", "pv-1"},
		{"apps", "v1", "Deployment", "shop", "api"},
		{"apps", "v1", "ReplicaSet", "shop", "api-7d9f"},
		{"apps", "v1", "StatefulSet", "shop", "db"},
		{"apps", "v1", "DaemonSet", "kube-system", "proxy"},
		{"batch", "v1", "Job", "shop", "migrate"},
		{"batch", "v1", "CronJob", "shop", "backup"},
		{"networking.k8s.io", "v1", "Ingress", "shop", "api"},
		{"discovery.k8s.io", "v1", "EndpointSlice", "shop", "api-x7k2p"},
	}, keys(got))

	for _, r := range got {
		want := r.Version
		if r.Group != "" {
			want = r.Group + "/" + r.Version
		}
		assert.Equal(t, want, r.Object["apiVersion"], "%s apiVersion", r.Kind)
		assert.Equal(t, r.Kind, r.Object["kind"])
	}
}

func TestResourcesCollectorSkipsForbiddenKinds(t *testing.T) {
	fake := kubetest.New(cluster()...)
	for _, resource := range []string{"pods", "deployments"} {
		fake.Dynamic.PrependReactor("list", resource, func(a k8stesting.Action) (bool, runtime.Object, error) {
			gr := a.GetResource().GroupResource()
			return true, nil, apierrors.NewForbidden(gr, "", errors.New("RBAC"))
		})
	}
	c := NewResourcesCollector(fake, Scope{})
	var warnings warnLog
	c.warn = warnings.warn

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	kinds := map[string]bool{}
	for _, r := range got {
		kinds[r.Kind] = true
	}
	assert.False(t, kinds["Pod"])
	assert.False(t, kinds["Deployment"])
	assert.True(t, kinds["ReplicaSet"], "other kinds are still collected")
	require.Len(t, warnings, 2)
	assert.Contains(t, warnings[0], "Skipping pods: not allowed")
	assert.Contains(t, warnings[1], "Skipping deployments.apps: not allowed")
}

func TestResourcesCollectorReturnsOtherListErrors(t *testing.T) {
	fake := kubetest.New(cluster()...)
	fake.Dynamic.PrependReactor("list", "services", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection reset")
	})
	_, err := NewResourcesCollector(fake, Scope{}).Collect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list services: connection reset")
}

func TestResourcesCollectorLeavesOutKindsTheClusterDoesNotServe(t *testing.T) {
	c := NewResourcesCollector(kubetest.New(cluster()...), Scope{})
	c.kinds = []schema.GroupKind{{Group: "", Kind: "Pod"}, {Group: "example.com", Kind: "Widget"}}

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []key{{"", "v1", "Pod", "shop", "api-1"}}, keys(got))
}

// partialDiscovery fails discovery for one group and serves the rest.
type partialDiscovery struct {
	*kubetest.Fake
	failed schema.GroupVersion
}

func (d partialDiscovery) PreferredResources() ([]*metav1.APIResourceList, error) {
	lists, err := d.Fake.PreferredResources()
	if err != nil {
		return nil, err
	}
	var kept []*metav1.APIResourceList
	for _, l := range lists {
		if l.GroupVersion != d.failed.String() {
			kept = append(kept, l)
		}
	}
	return kept, &discovery.ErrGroupDiscoveryFailed{
		Groups: map[schema.GroupVersion]error{d.failed: errors.New("service unavailable")},
	}
}

func TestResourcesCollectorWarnsAboutFailedDiscovery(t *testing.T) {
	c := NewResourcesCollector(partialDiscovery{kubetest.New(cluster()...), schema.GroupVersion{Group: "batch", Version: "v1"}}, Scope{})
	var warnings warnLog
	c.warn = warnings.warn

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	for _, r := range got {
		assert.NotEqual(t, "batch", r.Group)
	}
	assert.Len(t, got, len(Kinds)-2)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "batch/v1")
}

func TestResourcesCollectorFailsWhenDiscoveryFails(t *testing.T) {
	fake := kubetest.New()
	fake.Clientset.PrependReactor("get", "group", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection refused")
	})
	_, err := NewResourcesCollector(fake, Scope{}).Collect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "discover resources")
}

func TestResourcesCollectorListsOnlyChosenNamespaces(t *testing.T) {
	fake := kubetest.New(
		&corev1.Namespace{ObjectMeta: meta("", "shop")},
		&corev1.Namespace{ObjectMeta: meta("", "pay")},
		&corev1.Namespace{ObjectMeta: meta("", "kube-system")},
		&corev1.Node{ObjectMeta: meta("", "node-1")},
		&corev1.Pod{ObjectMeta: meta("shop", "api-1")},
		&corev1.Pod{ObjectMeta: meta("pay", "card-1")},
		&corev1.Pod{ObjectMeta: meta("kube-system", "dns-1")},
	)
	c := NewResourcesCollector(fake, Scope{Namespaces: []string{"shop", "pay"}})
	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []key{
		{"", "v1", "Namespace", "", "shop"},
		{"", "v1", "Namespace", "", "pay"},
		{"", "v1", "Node", "", "node-1"},
		{"", "v1", "Pod", "shop", "api-1"},
		{"", "v1", "Pod", "pay", "card-1"},
	}, keys(got))

	// Namespaced kinds are listed per namespace, never cluster-wide.
	for _, a := range fake.Dynamic.Actions() {
		if a.GetResource().Resource == "pods" {
			assert.Contains(t, []string{"shop", "pay"}, a.GetNamespace())
		}
	}
}

func TestResourcesCollectorExcludesNamespaces(t *testing.T) {
	fake := kubetest.New(
		&corev1.Namespace{ObjectMeta: meta("", "shop")},
		&corev1.Namespace{ObjectMeta: meta("", "kube-system")},
		&corev1.Pod{ObjectMeta: meta("shop", "api-1")},
		&corev1.Pod{ObjectMeta: meta("kube-system", "dns-1")},
	)
	got, err := NewResourcesCollector(fake, Scope{Exclude: []string{"kube-system"}}).Collect(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []key{
		{"", "v1", "Namespace", "", "shop"},
		{"", "v1", "Pod", "shop", "api-1"},
	}, keys(got))
}

func TestResourcesCollectorSkipsForbiddenNamespace(t *testing.T) {
	fake := kubetest.New(
		&corev1.Pod{ObjectMeta: meta("shop", "api-1")},
		&corev1.Pod{ObjectMeta: meta("pay", "card-1")},
	)
	fake.Dynamic.PrependReactor("list", "*", func(a k8stesting.Action) (bool, runtime.Object, error) {
		if a.GetNamespace() != "pay" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewForbidden(a.GetResource().GroupResource(), "", errors.New("RBAC"))
	})
	c := NewResourcesCollector(fake, Scope{Namespaces: []string{"shop", "pay"}})
	var warnings warnLog
	c.warn = warnings.warn

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	var pods []key
	for _, k := range keys(got) {
		if k.kind == "Pod" {
			pods = append(pods, k)
		}
	}
	assert.Equal(t, []key{{"", "v1", "Pod", "shop", "api-1"}}, pods)
	assert.Contains(t, fmt.Sprint(warnings), "Skipping pods in pay: not allowed")
}
