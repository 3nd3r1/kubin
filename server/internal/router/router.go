package router

import (
	"net/http"
)

type Router struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use adds middleware to the router that will wrap subsequent handlers
func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middleware)
}

// Handle registers a route with the router after applying all middleware
func (r *Router) Handle(pattern string, handler http.Handler) {
	wrapped := r.wrapMiddleware(handler)
	r.mux.Handle(pattern, wrapped)
}

// HandleFunc registers a route with the router after applying all middleware
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	wrapped := r.wrapMiddleware(http.HandlerFunc(handler))
	r.mux.Handle(pattern, wrapped)
}

// ServeHTTP makes the router implement http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// wrapMiddleware applies all middleware in reverse order (first added is outermost)
func (r *Router) wrapMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order so first added runs first
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	return handler
}

// Subrouter creates a new router with a path prefix
func (r *Router) Subrouter(prefix string) *Router {
	sub := New()
	r.mux.Handle(prefix+"/", http.StripPrefix(prefix, sub.mux))
	return sub
}
