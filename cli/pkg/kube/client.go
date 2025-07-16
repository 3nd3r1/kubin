package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubeClient struct {
	clientset *kubernetes.Clientset
}

func NewKubeClient() (*KubeClient, error) {
    kubeconfigPath := os.Getenv("KUBECONFIG")
    if kubeconfigPath == "" {
        if home := homedir.HomeDir(); home != "" {
            kubeconfigPath = filepath.Join(home, ".kube", "config")
        } else {
            return nil, fmt.Errorf("failed to find kubeconfig path")
        }
    }

    config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
    if err != nil {
        return nil, err
    }

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubeClient{clientset: clientset}, nil
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
