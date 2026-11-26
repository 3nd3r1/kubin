// Package kube wraps the Kubernetes API calls the CLI makes.
package kube

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// listPageSize is how many objects one list call asks for, so a big
// cluster is read in pages instead of one huge response.
const listPageSize = 500

type Client interface {
	ClusterInfo(ctx context.Context) (ClusterInfo, error)
	// PreferredResources lists the resources the server serves, each at
	// the version the server prefers. If some API groups could not be
	// read it returns what it got together with a
	// *discovery.ErrGroupDiscoveryFailed.
	PreferredResources() ([]*metav1.APIResourceList, error)
	// List returns every object of a resource in namespace, or in all
	// namespaces when it is empty.
	List(ctx context.Context, gvr schema.GroupVersionResource, namespace string) ([]unstructured.Unstructured, error)
	// GetPodLogs streams one container's log. The caller closes it.
	GetPodLogs(ctx context.Context, namespace, pod string, opts corev1.PodLogOptions) (io.ReadCloser, error)
}

// ClusterInfo describes the cluster a snapshot was taken from.
type ClusterInfo struct {
	Context       string
	Name          string
	Server        string
	ServerVersion string
}

type KubeClient struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
	context   string
	cluster   string
	server    string
}

var _ Client = (*KubeClient)(nil)

// NewKubeClient connects with a kubeconfig context, read the same way
// kubectl reads it (KUBECONFIG, then ~/.kube/config). An empty context
// means the current one.
func NewKubeClient(context string) (*KubeClient, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{CurrentContext: context})

	raw, err := loader.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %w", err)
	}
	if context == "" {
		context = raw.CurrentContext
	}
	var cluster string
	if c, ok := raw.Contexts[context]; ok {
		cluster = c.Cluster
	}

	config, err := loader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig context %q: %w", context, err)
	}
	// client-go defaults to 5 requests per second, which serialises the
	// parallel log fetches. The API server's own fairness still applies.
	config.QPS = 50
	config.Burst = 100
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewFromClients(clientset, dyn, context, cluster).WithServer(config.Host), nil
}

// NewFromClients wraps existing clients, like client-go's fakes.
func NewFromClients(clientset kubernetes.Interface, dyn dynamic.Interface, context, cluster string) *KubeClient {
	return &KubeClient{clientset: clientset, dynamic: dyn, context: context, cluster: cluster}
}

// WithServer records the Kubernetes API server address reported by ClusterInfo.
func (k *KubeClient) WithServer(server string) *KubeClient {
	k.server = server
	return k
}

func (k *KubeClient) ClusterInfo(ctx context.Context) (ClusterInfo, error) {
	v, err := k.clientset.Discovery().ServerVersion()
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("get server version: %w", err)
	}
	return ClusterInfo{Context: k.context, Name: k.cluster, Server: k.server, ServerVersion: v.GitVersion}, nil
}

func (k *KubeClient) PreferredResources() ([]*metav1.APIResourceList, error) {
	return discovery.ServerPreferredResources(k.clientset.Discovery())
}

func (k *KubeClient) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string) ([]unstructured.Unstructured, error) {
	var items []unstructured.Unstructured
	opts := metav1.ListOptions{Limit: listPageSize}
	for {
		list, err := k.dynamic.Resource(gvr).Namespace(namespace).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		items = append(items, list.Items...)
		opts.Continue = list.GetContinue()
		if opts.Continue == "" {
			return items, nil
		}
	}
}

func (k *KubeClient) GetPodLogs(ctx context.Context, namespace, pod string, opts corev1.PodLogOptions) (io.ReadCloser, error) {
	return k.clientset.CoreV1().Pods(namespace).GetLogs(pod, &opts).Stream(ctx)
}
