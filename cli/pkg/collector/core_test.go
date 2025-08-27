package collector

import (
	"context"
	"errors"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCoreCollector_Name(t *testing.T) {
	mockClient := &kube.MockClient{}
	collector := NewCoreCollector(mockClient)

	assert.Equal(t, "core", collector.Name())
}

func TestCoreCollector_Collect_Success(t *testing.T) {
	// Setup test data
	testNamespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		},
	}

	testPods := map[string][]corev1.Pod{
		"default": {
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app-123",
					Namespace: "default",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-server-456",
					Namespace: "default",
				},
			},
		},
		"kube-system": {
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-dns-789",
					Namespace: "kube-system",
				},
			},
		},
	}

	// Create mock client
	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return testNamespaces, nil
		},
		GetPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
			if pods, exists := testPods[namespace]; exists {
				return pods, nil
			}
			return []corev1.Pod{}, nil
		},
	}

	collector := NewCoreCollector(mockClient)
	resources, err := collector.Collect(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, resources, 5) // 3 namespaces, 2 pods

	// Check namespaces
	namespaceCount := 0
	podCount := 0
	for _, resource := range resources {
		switch resource.Kind {
		case "namespace":
			namespaceCount++
		case "pod":
			podCount++
		default:
			t.Errorf("Unexpected resource kind: %s", resource.Kind)
		}
	}
	assert.Equal(t, 2, namespaceCount)
	assert.Equal(t, 3, podCount)

	// Check pod naming format and metadata
	var expectedPodNames []string
	for _, pods := range testPods {
		for _, pod := range pods {
			expectedPodNames = append(expectedPodNames, pod.Name)
		}
	}

	for _, resource := range resources {
		if resource.Kind == "pod" {
			assert.Equal(t, "pod", resource.Kind)
			assert.Contains(t, expectedPodNames, resource.Name)
			assert.NotNil(t, resource.Metadata)
			assert.Contains(t, resource.Metadata, "namespace")
		}
	}
}

func TestCoreCollector_Collect_NamespaceError(t *testing.T) {
	// Create mock client that returns error for namespaces
	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return nil, errors.New("failed to connect to cluster")
		},
	}

	collector := NewCoreCollector(mockClient)
	resources, err := collector.Collect(context.Background())

	// Should fail early on namespace collection
	assert.Error(t, err)
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), "failed to connect to cluster")
}

func TestCoreCollector_Collect_PodError(t *testing.T) {
	// Setup test data
	testNamespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
	}

	// Create mock client that returns error for pods
	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return testNamespaces, nil
		},
		GetPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
			return nil, errors.New("failed to get pods")
		},
	}

	collector := NewCoreCollector(mockClient)
	resources, err := collector.Collect(context.Background())

	// Should fail on pod collection
	assert.Error(t, err)
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), "failed to get pods")
}

func TestCoreCollector_CollectPods_Success(t *testing.T) {
	testNamespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "production",
			},
		},
	}

	testPods := map[string][]corev1.Pod{
		"default": {
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-pod",
					Namespace: "default",
				},
			},
		},
		"production": {
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-pod",
					Namespace: "production",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db-pod",
					Namespace: "production",
				},
			},
		},
	}

	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return testNamespaces, nil
		},
		GetPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
			if pods, exists := testPods[namespace]; exists {
				return pods, nil
			}
			return []corev1.Pod{}, nil
		},
	}

	collector := NewCoreCollector(mockClient)
	podResources, err := collector.collectPods(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, podResources, 3)

	// Check pod naming format and metadata
	expectedPodNamespaces := make(map[string]string)
	for _, pods := range testPods {
		for _, pod := range pods {
			expectedPodNamespaces[pod.Name] = pod.Namespace
		}
	}

	for _, pod := range podResources {
		assert.Equal(t, pod.Kind, "pod")
		assert.Contains(t, expectedPodNamespaces, pod.Name)
		assert.Equal(t, pod.Metadata["namespace"], expectedPodNamespaces[pod.Name])
	}
}

func TestCoreCollector_CollectPods_EmptyNamespace(t *testing.T) {
	testNamespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "empty-namespace",
			},
		},
	}

	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return testNamespaces, nil
		},
		GetPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
			// Return empty pod list for the namespace
			return []corev1.Pod{}, nil
		},
	}

	collector := NewCoreCollector(mockClient)
	podResources, err := collector.collectPods(context.Background())

	// Should succeed with no pods
	assert.NoError(t, err)
	assert.Len(t, podResources, 0)
}
