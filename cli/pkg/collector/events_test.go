package collector

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func event(namespace, name, reason string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: namespace, Name: name},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Namespace: namespace, Name: "api"},
		Reason:         reason,
		Type:           corev1.EventTypeWarning,
	}
}

func TestEventsCollectorCollectsAllNamespaces(t *testing.T) {
	cs := fake.NewClientset(
		event("shop", "api.1", "BackOff"),
		event("kube-system", "dns.1", "Pulled"),
	)
	got, err := NewEventsCollector(kube.NewFromClientset(cs, "", "")).Collect(context.Background())
	require.NoError(t, err)

	var keys []string
	for _, r := range got {
		keys = append(keys, r.Namespace+"/"+r.Name)
		assert.Equal(t, "Event", r.Kind)
		assert.Equal(t, "", r.Group)
		assert.Equal(t, "v1", r.Object["apiVersion"])
		assert.Equal(t, "Event", r.Object["kind"])
	}
	assert.ElementsMatch(t, []string{"shop/api.1", "kube-system/dns.1"}, keys)

	var backOff map[string]any
	for _, r := range got {
		if r.Name == "api.1" {
			backOff = r.Object
		}
	}
	assert.Equal(t, "BackOff", backOff["reason"])
	assert.Equal(t, "api", backOff["involvedObject"].(map[string]any)["name"])
}

func TestEventsCollectorSkipsForbidden(t *testing.T) {
	cs := fake.NewClientset(event("shop", "api.1", "BackOff"))
	cs.PrependReactor("list", "events", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "events"}, "", errors.New("RBAC"))
	})
	c := NewEventsCollector(kube.NewFromClientset(cs, "", ""))
	var warnings []string
	c.warn = func(msg string, fields ...any) { warnings = append(warnings, fmt.Sprint(msg, fields)) }

	got, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "not allowed")
}

func TestEventsCollectorReturnsOtherErrors(t *testing.T) {
	cs := fake.NewClientset()
	cs.PrependReactor("list", "events", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection reset")
	})
	_, err := NewEventsCollector(kube.NewFromClientset(cs, "", "")).Collect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list events: connection reset")
}
