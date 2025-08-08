package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	mux *chi.Mux
}

func New() *Router {
	return &Router{
		mux: chi.NewRouter(),
	}
}

// Use adds middleware to the router
func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.mux.Use(middleware...)
}

// Handle registers a route with the router
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// HandleFunc registers a route with the router
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, handler)
}

// ServeHTTP makes the router implement http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Subrouter creates a new router with a path prefix
func (r *Router) Subrouter(prefix string) *Router {
	sub := New()
	r.mux.Mount(prefix, sub.mux)
	return sub
}

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
	mr.router.mux.With(middleware.Method(mr.methods...)).Handle(pattern, handler)
}

func (mr *methodRouter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mr.Handle(pattern, http.HandlerFunc(handler))
}
