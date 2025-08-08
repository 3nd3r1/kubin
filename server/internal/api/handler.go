package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SnapshotHandler struct {
}

func NewSnapshotHandler() *SnapshotHandler {
	return &SnapshotHandler{}
}

func (h *SnapshotHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	respondWithNotImplemented(w)
}

func (h *SnapshotHandler) UploadSnapshot(w http.ResponseWriter, r *http.Request) {
	respondWithNotImplemented(w)
}

func (h *SnapshotHandler) GetSnapshotResources(w http.ResponseWriter, r *http.Request) {
	snapshotID := chi.URLParam(r, "id")
	if snapshotID == "" {
		http.Error(w, "Snapshot ID is required", http.StatusBadRequest)
		return
	}
	respondWithNotImplemented(w)
}

func (h *SnapshotHandler) GetSnapshotPods(w http.ResponseWriter, r *http.Request) {
	snapshotID := chi.URLParam(r, "id")
	if snapshotID == "" {
		http.Error(w, "Snapshot ID is required", http.StatusBadRequest)
		return
	}
	respondWithNotImplemented(w)
}

func (h *SnapshotHandler) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	snapshotID := chi.URLParam(r, "id")
	if snapshotID == "" {
		http.Error(w, "Snapshot ID is required", http.StatusBadRequest)
		return
	}
	respondWithNotImplemented(w)
}

func (h *SnapshotHandler) GetSnapshotNamespaces(w http.ResponseWriter, r *http.Request) {
	snapshotID := chi.URLParam(r, "id")
	if snapshotID == "" {
		http.Error(w, "Snapshot ID is required", http.StatusBadRequest)
		return
	}
	respondWithNotImplemented(w)
}

func respondWithNotImplemented(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "Not implemented yet"}`))
}
