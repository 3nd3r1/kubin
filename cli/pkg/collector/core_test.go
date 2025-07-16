package collector

import (
	"context"
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

	// Create mock client
	mockClient := &kube.MockClient{
		GetNamespacesFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
			return testNamespaces, nil
		},
	}

	collector := NewCoreCollector(mockClient)
	resources, err := collector.Collect(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "namespace", resources[0].Kind)
	assert.Equal(t, "default", resources[0].Name)
	assert.Equal(t, "namespace", resources[1].Kind)
	assert.Equal(t, "kube-system", resources[1].Name)
}
