package conductor

import "net/http"

// A Router is an http.Handler that can route requests with a stack of middleware components.
type Router struct {
	components []Component
	mux        *http.ServeMux
}

// NewRouter returns a new Router instance
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use adds a Component to the middleware stack for an Router
func (router *Router) Use(component ...Component) {
	router.components = append(router.components, component...)
}

// Handle registers the handler for the given pattern.
// Handler is run using the middleware stack.
func (router *Router) Handle(pattern string, handler http.Handler) {
	for i := range router.components {
		handler = router.components[len(router.components)-1-i].Next(handler)
	}

	router.mux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
// Handler is run using the middleware stack
func (router *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	router.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP dispatches the request to the Router mux
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
