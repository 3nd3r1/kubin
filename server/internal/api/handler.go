package api

import (
	"net/http"
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

func respondWithNotImplemented(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotImplemented)
}
