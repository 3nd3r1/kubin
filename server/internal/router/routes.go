package router

import (
	"net/http"
	"strings"

	"github.com/3nd3r1/kubin/server/internal/api"
)

// Methods restricts the handler to specific HTTP methods
func (r *Router) Methods(methods ...string) *methodRouter {
	return &methodRouter{
		router:  r,
		methods: methods,
	}
}

type methodRouter struct {
	router  *Router
	methods []string
}

func (mr *methodRouter) Handle(pattern string, handler http.Handler) {
	mr.router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		for _, method := range mr.methods {
			if r.Method == method {
				handler.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Allow", strings.Join(mr.methods, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func (mr *methodRouter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mr.Handle(pattern, http.HandlerFunc(handler))
}

func RegisterRoutes(r *Router, handler *api.SnapshotHandler) {
	// API v1 routes
	apiV1 := r.Subrouter("/api/v1")
	apiV1.Use(jsonMiddleware)

	// Snapshots collection
	apiV1.Methods("GET").HandleFunc("/snapshots", handler.ListSnapshots)
	apiV1.Methods("POST").HandleFunc("/snapshots", handler.UploadSnapshot)

	// Health check (no middleware)
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
