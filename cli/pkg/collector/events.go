package collector

import (
	"context"
	"fmt"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventsCollector collects core/v1 events from every namespace. They say
// what Kubernetes saw happen to a pod, like back-offs and failed pulls.
//
// Credentials: events carry messages written by controllers and kubelets,
// not object data. The redact rules still run on them, as on every object.
type EventsCollector struct {
	client kube.Client
	warn   func(msg string, fields ...any)
}

func NewEventsCollector(client kube.Client) *EventsCollector {
	return &EventsCollector{client: client, warn: log.Warn}
}

func (c *EventsCollector) Name() string {
	return "events"
}

// Collect returns no events, and no error, when the user isn't allowed to
// list them: a snapshot without events is still worth sharing.
func (c *EventsCollector) Collect(ctx context.Context) ([]Resource, error) {
	events, err := c.client.GetEvents(ctx, metav1.NamespaceAll)
	if apierrors.IsForbidden(err) {
		c.warn("Skipping events: not allowed to list them", "error", err.Error())
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	kind := corev1.SchemeGroupVersion.WithKind("Event")
	resources := make([]Resource, 0, len(events))
	for i := range events {
		r, err := fromTyped(kind, &events[i])
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}
