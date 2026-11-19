package collector

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// Kinds is every kind a snapshot holds. The version of each comes from
// discovery, as the version the server prefers.
//
// Credentials: every object goes through redact.Redactor in
// snapshot.Manager. Workload templates and pods lose env var values there.
// ConfigMaps are kept as they are: they are meant for configuration that is
// not secret, and docs/snapshot-format.md says they are not redacted.
// Secrets are not collected.
var Kinds = []schema.GroupKind{
	{Group: "", Kind: "Namespace"},
	{Group: "", Kind: "Node"},
	{Group: "", Kind: "Pod"},
	{Group: "", Kind: "Event"},
	{Group: "", Kind: "Service"},
	{Group: "", Kind: "ConfigMap"},
	{Group: "", Kind: "PersistentVolumeClaim"},
	{Group: "", Kind: "PersistentVolume"},
	{Group: "apps", Kind: "Deployment"},
	{Group: "apps", Kind: "ReplicaSet"},
	{Group: "apps", Kind: "StatefulSet"},
	{Group: "apps", Kind: "DaemonSet"},
	{Group: "batch", Kind: "Job"},
	{Group: "batch", Kind: "CronJob"},
	{Group: "networking.k8s.io", Kind: "Ingress"},
	// Not core Endpoints: they are deprecated since v1.33, and listing
	// them prints a warning.
	{Group: "discovery.k8s.io", Kind: "EndpointSlice"},
}

// Scope limits collection to some namespaces. The zero Scope is the whole
// cluster.
type Scope struct {
	// Namespaces, when set, are the only namespaces collected. Namespaced
	// kinds are listed in each of them, so a user allowed only those
	// namespaces can still take a snapshot. Cluster-scoped kinds other
	// than Namespace, like Nodes, are still collected.
	Namespaces []string
	// Exclude are namespaces left out, applied after Namespaces.
	Exclude []string
}

// keep reports whether an object in namespace ns is in scope. Cluster-scoped
// objects pass ns "".
func (s Scope) keep(ns string) bool {
	if ns == "" {
		return true
	}
	if len(s.Namespaces) > 0 && !slices.Contains(s.Namespaces, ns) {
		return false
	}
	return !slices.Contains(s.Exclude, ns)
}

// ResourcesCollector lists every kind in Kinds that the cluster serves.
type ResourcesCollector struct {
	client kube.Client
	kinds  []schema.GroupKind
	scope  Scope
	warn   func(msg string, fields ...any)
}

func NewResourcesCollector(client kube.Client, scope Scope) *ResourcesCollector {
	return &ResourcesCollector{client: client, kinds: Kinds, scope: scope, warn: log.Warn}
}

func (c *ResourcesCollector) Name() string {
	return "resources"
}

// Collect skips, with a warning, kinds the user isn't allowed to list and
// API groups discovery could not read: a snapshot with some kinds missing
// is still worth sharing. Kinds the cluster doesn't serve are left out.
func (c *ResourcesCollector) Collect(ctx context.Context) ([]Resource, error) {
	gvrs, err := c.resolve()
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, gvr := range gvrs {
		items, err := c.list(ctx, gvr)
		if err != nil {
			return nil, err
		}
		for _, u := range items {
			ns := u.GetNamespace()
			if gvr.Group == "" && gvr.kind == "Namespace" {
				ns = u.GetName()
			}
			if !c.scope.keep(ns) {
				continue
			}
			// Lists leave apiVersion and kind empty on items.
			u.SetAPIVersion(gvr.GroupVersion().String())
			u.SetKind(gvr.kind)
			resources = append(resources, Resource{
				Group:     gvr.Group,
				Version:   gvr.Version,
				Kind:      gvr.kind,
				Namespace: u.GetNamespace(),
				Name:      u.GetName(),
				Object:    u.Object,
			})
		}
	}
	return resources, nil
}

// list lists one kind in scope. A list the user isn't allowed to make is
// skipped with a warning.
func (c *ResourcesCollector) list(ctx context.Context, r kindResource) ([]unstructured.Unstructured, error) {
	namespaces := []string{""}
	if r.namespaced && len(c.scope.Namespaces) > 0 {
		namespaces = c.scope.Namespaces
	}
	var items []unstructured.Unstructured
	for _, ns := range namespaces {
		got, err := c.client.List(ctx, r.GroupVersionResource, ns)
		where := ""
		if ns != "" {
			where = " in " + ns
		}
		if apierrors.IsForbidden(err) {
			c.warn(fmt.Sprintf("Skipping %s%s: not allowed to list them", r.resource(), where), "error", err.Error())
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("list %s%s: %w", r.resource(), where, err)
		}
		items = append(items, got...)
	}
	return items, nil
}

type kindResource struct {
	schema.GroupVersionResource
	kind       string
	namespaced bool
}

// resource names the resource for messages, like "deployments.apps".
func (r kindResource) resource() string {
	return r.GroupResource().String()
}

// resolve finds the resource and preferred version of each allowed kind,
// in the order of c.kinds.
func (c *ResourcesCollector) resolve() ([]kindResource, error) {
	lists, err := c.client.PreferredResources()
	var failed *discovery.ErrGroupDiscoveryFailed
	if errors.As(err, &failed) {
		groups := make([]string, 0, len(failed.Groups))
		for gv := range failed.Groups {
			groups = append(groups, gv.String())
		}
		sort.Strings(groups)
		c.warn("Skipping API groups discovery could not read", "groups", strings.Join(groups, ","))
	} else if err != nil {
		return nil, fmt.Errorf("discover resources: %w", err)
	}

	found := map[schema.GroupKind]kindResource{}
	for _, list := range lists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("discover resources: %w", err)
		}
		for _, r := range list.APIResources {
			if strings.Contains(r.Name, "/") || !slices.Contains(r.Verbs, "list") {
				continue // a subresource, like pods/log
			}
			gk := schema.GroupKind{Group: gv.Group, Kind: r.Kind}
			if _, ok := found[gk]; !ok {
				found[gk] = kindResource{gv.WithResource(r.Name), r.Kind, r.Namespaced}
			}
		}
	}

	var out []kindResource
	for _, gk := range c.kinds {
		if r, ok := found[gk]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}
