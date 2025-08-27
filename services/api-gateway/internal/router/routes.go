package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/3nd3r1/kubin/api-gateway/internal/api"
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

	// Snapshots collection using chi's method routing
	apiV1.mux.Route("/snapshots", func(r chi.Router) {
		r.Get("/", handler.ListSnapshots)
		r.Post("/", handler.UploadSnapshot)
	})

	// Internal API for UI
	internalAPI := r.Subrouter("/internal/api/v1")
	internalAPI.Use(jsonMiddleware)

	// Internal snapshot endpoints
	internalAPI.mux.Route("/snapshots/{id}", func(r chi.Router) {
		r.Get("/resources", handler.GetSnapshotResources)
		r.Get("/pods", handler.GetSnapshotPods)
		r.Get("/logs", handler.GetPodLogs)
		r.Get("/namespaces", handler.GetSnapshotNamespaces)
	})
}
