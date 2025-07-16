package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type MockClient struct {
	GetNamespacesFunc func(ctx context.Context) ([]corev1.Namespace, error)
	GetPodsFunc       func(ctx context.Context, namespace string) ([]corev1.Pod, error)
	GetPodLogsFunc    func(ctx context.Context, namespace string, podName string) (string, error)
}

var _ Client = (*MockClient)(nil)

func (m *MockClient) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	return m.GetNamespacesFunc(ctx)
}

func (m *MockClient) GetPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	return m.GetPodsFunc(ctx, namespace)
}

func (m *MockClient) GetPodLogs(ctx context.Context, namespace string, podName string) (string, error) {
	return m.GetPodLogsFunc(ctx, namespace, podName)
}
