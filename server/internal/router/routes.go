package router

import (
	"net/http"

	"github.com/3nd3r1/kubin/server/internal/api"
)

func RegisterRoutes(r *Router, handler *api.SnapshotHandler) {
	// Health check endpoints (no middleware)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add readiness checks (database, storage, etc.)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1 routes
	apiV1 := r.Subrouter("/api/v1")
	apiV1.Use(jsonMiddleware)

	// Snapshots collection
	apiV1.Methods("GET").HandleFunc("/snapshots", handler.ListSnapshots)
	apiV1.Methods("POST").HandleFunc("/snapshots", handler.UploadSnapshot)

	// Internal API for UI
	internalAPI := r.Subrouter("/internal/api/v1")
	internalAPI.Use(jsonMiddleware)

	// Internal snapshot endpoints
	internalAPI.Methods("GET").HandleFunc("/snapshots/{id}/resources", handler.GetSnapshotResources)
	internalAPI.Methods("GET").HandleFunc("/snapshots/{id}/pods", handler.GetSnapshotPods)
	internalAPI.Methods("GET").HandleFunc("/snapshots/{id}/logs", handler.GetPodLogs)
	internalAPI.Methods("GET").HandleFunc("/snapshots/{id}/namespaces", handler.GetSnapshotNamespaces)
}
