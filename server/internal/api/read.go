package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/3nd3r1/kubin/server/internal/store"
)

// clusterScoped stands in for the namespace of cluster-scoped resources in
// URLs and filters, as in the tarball layout.
const clusterScoped = "_cluster"

// Reader reads stored snapshots. store.Store implements it.
type Reader interface {
	GetSnapshot(ctx context.Context, id string) (store.Snapshot, error)
	CountResources(ctx context.Context, id string) ([]store.KindCount, error)
	ListResources(ctx context.Context, id string, f store.Filter) ([]store.Resource, error)
	GetResource(ctx context.Context, id string, k store.Key) (store.Resource, error)
}

type kindCount struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Count   int    `json:"count"`
}

type snapshotResponse struct {
	ID         string          `json:"id"`
	UploadedAt time.Time       `json:"uploaded_at"`
	Manifest   json.RawMessage `json:"manifest"`
	// Counts is what was indexed. The manifest's own counts are what the
	// CLI reported and may differ.
	Counts []kindCount `json:"counts"`
}

type resourceItem struct {
	Group     string          `json:"group"`
	Version   string          `json:"version"`
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Name      string          `json:"name"`
	Object    json.RawMessage `json:"object"`
}

type resourceList struct {
	Items []resourceItem `json:"items"`
}

// snapshot loads the snapshot named in the URL, writing a 404 or 500 and
// returning false if it can't.
func (h *handlers) snapshot(w http.ResponseWriter, r *http.Request) (store.Snapshot, bool) {
	snap, err := h.opts.Store.GetSnapshot(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "snapshot not found")
		return store.Snapshot{}, false
	}
	if err != nil {
		h.serverError(w, "get snapshot", err)
		return store.Snapshot{}, false
	}
	return snap, true
}

func (h *handlers) getSnapshot(w http.ResponseWriter, r *http.Request) {
	snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	counts, err := h.opts.Store.CountResources(r.Context(), snap.ID)
	if err != nil {
		h.serverError(w, "count resources", err)
		return
	}
	resp := snapshotResponse{
		ID:         snap.ID,
		UploadedAt: snap.UploadedAt.UTC(),
		Manifest:   snap.Manifest,
		Counts:     make([]kindCount, len(counts)),
	}
	for i, c := range counts {
		resp.Counts[i] = kindCount(c)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *handlers) listResources(w http.ResponseWriter, r *http.Request) {
	snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	f := store.Filter{Group: q.Get("group"), Kind: q.Get("kind"), Namespace: q.Get("namespace")}
	if f.Namespace == clusterScoped {
		f.Namespace, f.ClusterScoped = "", true
	}
	rs, err := h.opts.Store.ListResources(r.Context(), snap.ID, f)
	if err != nil {
		h.serverError(w, "list resources", err)
		return
	}
	resp := resourceList{Items: make([]resourceItem, len(rs))}
	for i, res := range rs {
		resp.Items[i] = resourceItem{
			Group: res.Group, Version: res.Version, Kind: res.Kind,
			Namespace: res.Namespace, Name: res.Name, Object: res.Object,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// getResource returns the object itself, as kubectl get -o json would.
func (h *handlers) getResource(w http.ResponseWriter, r *http.Request) {
	snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	k := store.Key{
		Group:     r.URL.Query().Get("group"),
		Kind:      chi.URLParam(r, "kind"),
		Namespace: chi.URLParam(r, "ns"),
		Name:      chi.URLParam(r, "name"),
	}
	if k.Namespace == clusterScoped {
		k.Namespace = ""
	}
	res, err := h.opts.Store.GetResource(r.Context(), snap.ID, k)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	if err != nil {
		h.serverError(w, "get resource", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(res.Object)
}

func (h *handlers) serverError(w http.ResponseWriter, what string, err error) {
	h.opts.Log.Error(what, "err", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
