// Package kubetest builds a kube.Client on client-go's fakes, with the
// discovery of a plain cluster, for tests.
package kubetest

import (
	"github.com/3nd3r1/kubin/cli/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

// Context and Cluster name the fake kubeconfig context; ServerVersion is
// what the fake API server reports.
const (
	Context       = "kind-kubin"
	Cluster       = "kind-kubin"
	ServerVersion = "v1.33.1"
)

type resource struct {
	name, kind string
	namespaced bool
}

// builtin is part of what a plain cluster serves: more kinds than kubin
// collects, subresources, and events in two groups.
var builtin = map[string][]resource{
	"v1": {
		{"namespaces", "Namespace", false},
		{"nodes", "Node", false},
		{"persistentvolumes", "PersistentVolume", false},
		{"pods", "Pod", true},
		{"pods/log", "Pod", true},
		{"pods/status", "Pod", true},
		{"events", "Event", true},
		{"services", "Service", true},
		{"endpoints", "Endpoints", true},
		{"configmaps", "ConfigMap", true},
		{"secrets", "Secret", true},
		{"serviceaccounts", "ServiceAccount", true},
		{"persistentvolumeclaims", "PersistentVolumeClaim", true},
	},
	"apps/v1": {
		{"deployments", "Deployment", true},
		{"deployments/scale", "Scale", true},
		{"replicasets", "ReplicaSet", true},
		{"statefulsets", "StatefulSet", true},
		{"daemonsets", "DaemonSet", true},
	},
	"batch/v1": {
		{"jobs", "Job", true},
		{"cronjobs", "CronJob", true},
	},
	"networking.k8s.io/v1": {
		{"ingresses", "Ingress", true},
		{"networkpolicies", "NetworkPolicy", true},
	},
	"discovery.k8s.io/v1": {
		{"endpointslices", "EndpointSlice", true},
	},
	"events.k8s.io/v1": {
		{"events", "Event", true},
	},
}

// Fake is a kube.Client on client-go's fakes. Clientset serves logs and the
// server version, Dynamic serves lists; add reactors to either.
type Fake struct {
	*kube.KubeClient
	Clientset *fake.Clientset
	Dynamic   *dynamicfake.FakeDynamicClient
}

// New returns a fake cluster holding objs, which are typed objects from
// k8s.io/api.
func New(objs ...runtime.Object) *Fake {
	cs := fake.NewClientset(objs...)
	disco := cs.Discovery().(*fakediscovery.FakeDiscovery)
	disco.FakedServerVersion = &version.Info{GitVersion: ServerVersion}

	listKinds := map[schema.GroupVersionResource]string{}
	for gv, resources := range builtin {
		list := &metav1.APIResourceList{GroupVersion: gv}
		parsed, err := schema.ParseGroupVersion(gv)
		if err != nil {
			panic(err)
		}
		for _, r := range resources {
			list.APIResources = append(list.APIResources, metav1.APIResource{
				Name:       r.name,
				Kind:       r.kind,
				Namespaced: r.namespaced,
				Verbs:      metav1.Verbs{"get", "list", "watch"},
			})
			listKinds[parsed.WithResource(r.name)] = r.kind + "List"
		}
		disco.Resources = append(disco.Resources, list)
	}

	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme.Scheme, listKinds, objs...)
	return &Fake{
		KubeClient: kube.NewFromClients(cs, dyn, Context, Cluster),
		Clientset:  cs,
		Dynamic:    dyn,
	}
}
