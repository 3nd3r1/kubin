// Package kube wraps the Kubernetes API calls the CLI makes.
package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	ClusterInfo(ctx context.Context) (ClusterInfo, error)
	GetNamespaces(ctx context.Context) ([]corev1.Namespace, error)
	// GetPods lists pods in namespace, or in all namespaces when it is empty.
	GetPods(ctx context.Context, namespace string) ([]corev1.Pod, error)
	GetPodLogs(ctx context.Context, namespace string, podName string) (string, error)
}

// ClusterInfo describes the cluster a snapshot was taken from.
type ClusterInfo struct {
	Context       string
	Name          string
	ServerVersion string
}

type KubeClient struct {
	clientset kubernetes.Interface
	context   string
	cluster   string
}

var _ Client = (*KubeClient)(nil)

// NewKubeClient connects with the current kubeconfig context, read the same
// way kubectl reads it (KUBECONFIG, then ~/.kube/config).
func NewKubeClient() (*KubeClient, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})

	raw, err := loader.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %w", err)
	}
	var cluster string
	if c, ok := raw.Contexts[raw.CurrentContext]; ok {
		cluster = c.Cluster
	}

	config, err := loader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig context %q: %w", raw.CurrentContext, err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewFromClientset(clientset, raw.CurrentContext, cluster), nil
}

// NewFromClientset wraps an existing clientset, like client-go's fake.
func NewFromClientset(clientset kubernetes.Interface, context, cluster string) *KubeClient {
	return &KubeClient{clientset: clientset, context: context, cluster: cluster}
}

func (k *KubeClient) ClusterInfo(ctx context.Context) (ClusterInfo, error) {
	v, err := k.clientset.Discovery().ServerVersion()
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("get server version: %w", err)
	}
	return ClusterInfo{Context: k.context, Name: k.cluster, ServerVersion: v.GitVersion}, nil
}

func (k *KubeClient) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	list, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (k *KubeClient) GetPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	list, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (k *KubeClient) GetPodLogs(ctx context.Context, namespace string, podName string) (string, error) {
	req := k.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", err
	}
	return string(logs), nil
}
