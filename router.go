package conductor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type routeKey string

const routeParamsKey routeKey = "github.com/ascarter/conductor/RouteParamsKey"

// newContextWithRegexpMatch creates context with regular expression matches
func newContextWithRouteParams(ctx context.Context, params RouteParams) context.Context {
	return context.WithValue(ctx, routeParamsKey, params)
}

// RouteParamsFromContext returns a map of params and values extracted from the route match.
// Unnamed parameters are returned using position (for example `$1`).
func RouteParamsFromContext(ctx context.Context) (RouteParams, bool) {
	params, ok := ctx.Value(routeParamsKey).(RouteParams)
	return params, ok
}

// RouteParams is a map of param names to values that are matched with a pattern to a path.
// Param ID's are expected to be unique.
type RouteParams map[string]string

// A route is a RegExp pattern and the handler to call.
type route struct {
	re *regexp.Regexp
	h  http.Handler
}

// expandParams replaces parameterized placeholders like `:id` with named
// regular expression match group like `(?P<id>\w+)`
func expandParams(s string) string {
	// If pattern does not start with `/`, treat as pure regular expression
	if s[0] != '/' {
		return s
	}

	var n int
	parts := strings.Split(s, "/")
	for i, v := range parts {
		if strings.HasPrefix(v, ":") {
			// Replace it with named match group
			parts[i] = fmt.Sprintf(`(?P<%s>\w+)`, v[1:])
			n++
		}
	}

	// Any parameters?
	if n > 0 {
		return strings.Join(parts, "/")
	}

	return s
}

func newRoute(p string, h http.Handler) *route {
	n := len(p)
	pattern := expandParams(p)

	if p[0] == '/' && p[n-1] != '/' {
		if p[n-1] != '$' {
			// Terminate for exact match
			pattern += "$"
		}
	}

	return &route{h: h, re: regexp.MustCompile(pattern)}
}

func (r *route) GetParams(path string) RouteParams {
	keys := r.re.SubexpNames()
	values := r.re.FindStringSubmatch(path)

	params := RouteParams{}
	for i, k := range keys {
		// Skip entire match string
		if i == 0 {
			continue
		}
		if k == "" {
			k = "$" + strconv.Itoa(i)
		}
		params[k] = values[i]
	}

	return params
}

// Match checks if p matches route
func (r *route) Match(p string) bool {
	return r.re.MatchString(p)
}

// A RouterMux is an HTTP request multiplexer for URL patterns.
// It matches URL of each incoming request against a list of registered patterns
// and calls the handler for best pattern match.
//
// Patterns are static matches, parameterized patterns, or regular expressions.
// Longer patterns take precedence over shorter ones. If there are patterns that
// match both "/images\/.*" and "/images/thumbnails\/.*", the path "/images/thumbnails"
// would use the later handler.
//
// Patterns can take the following forms:
//
//	Static
//	------
//	`/posts`
//	`/posts/`
//
//	Parameterized
//	-------------
//	`/posts/:id`
//	`/posts/:id/comments/:author`
//
//	Regular Expression
//	------------------
//	`/posts/(\d+)`
//	`/posts/(?<id>\d+)`
//	`.*/author`
//
// RouterMux follows the general approach used by http.ServeMux. If host patterns are
// desired, a RouterMux can be assigned to routes on http.ServeMux using `host/`
type RouterMux struct {
	mu     sync.RWMutex
	routes map[string]*route
}

// NewRouterMux allocates and returns a new RouterMux.
func NewRouterMux() *RouterMux {
	return &RouterMux{}
}

// match finds the best pattern for the method and path.
func (mux *RouterMux) match(method, path string) (h http.Handler, pattern string) {
	var n = 0
	for k, v := range mux.routes {
		if !v.Match(path) {
			continue
		}

		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
			pattern = k
		}
	}

	return
}

// Handler returns the handler to use for the given request.
// It always returns a non-nil handler.
//
// Handler also returns the registered pattern that matches the request.
//
// If there is no registered handler that applies to the request, Handler returns
// a ``page not found'' handler and an empty pattern.
func (mux *RouterMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	h, pattern = mux.match(r.Method, r.URL.Path)

	if h == nil {
		// No handler matched
		h, pattern = http.NotFoundHandler(), ""
	}

	return
}

// ServeHTTP dispatches request to the handler whose pattern most closely matches the request URL.
func (mux *RouterMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, pattern := mux.Handler(r)
	route, ok := mux.routes[pattern]
	if ok {
		ctx := newContextWithRouteParams(r.Context(), route.GetParams(r.URL.Path))
		r = r.WithContext(ctx)
	}
	h.ServeHTTP(w, r)
}

// Handle registers a handler for a pattern.
//
// If the handler already exists for pattern or the expression does not compile,
// Handle panics.
func (mux *RouterMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Verify parameters
	if pattern == "" {
		panic("mux: invalid pattern " + pattern)
	}

	if handler == nil {
		panic("mux: nil handler")
	}

	if _, ok := mux.routes[pattern]; ok {
		panic("mux: multiple registrations for " + pattern)
	}

	if mux.routes == nil {
		mux.routes = make(map[string]*route)
	}

	mux.routes[pattern] = newRoute(pattern, handler)
}

// HandleFunc registers a handler function for a pattern.
func (mux *RouterMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// A RouteHandler dispatch requests based on method and path.
// It is useful for mapping a heterogenous mix of methods and path patterns.
type RouteHandler struct {
	routes map[string]*RouterMux
}

// NewRouteHandler returns a new RouteHandler instance
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{}
}

// HandleRoute defines route for HTTP method and pattern to http.Handler.
func (h *RouteHandler) HandleRoute(method, pattern string, handler http.Handler) error {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		if h.routes == nil {
			h.routes = make(map[string]*RouterMux)
		}

		mux, ok := h.routes[method]
		if !ok {
			mux = NewRouterMux()
			h.routes[method] = mux
		}

		mux.Handle(pattern, handler)
	default:
		return errors.New("invalid HTTP method")
	}

	return nil
}

// HandleRouteFunc defines route for HTTP method and pattern to http.HandlerFunc.
func (h *RouteHandler) HandleRouteFunc(method, pattern string, fn http.HandlerFunc) error {
	return h.HandleRoute(method, pattern, http.HandlerFunc(fn))
}

// ServeHTTP dispatches request to the RouteMux handler that matches the request method.
func (h *RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux, ok := h.routes[r.Method]
	if !ok {
		http.NotFound(w, r)
		return
	}

	mux.ServeHTTP(w, r)
}

// A Router is an http.Handler that can route requests with a stack of middleware components.
type Router struct {
	mux        *http.ServeMux
	rmux       *RouterMux
	components []Component
}

// NewRouter returns a new Router instance
func NewRouter() *Router {
	router := Router{mux: http.NewServeMux(), rmux: NewRouterMux()}

	// Take advantage of ServeMux URL normalization and then send
	// to RouterMux for dispatch
	router.mux.Handle("/", router.rmux)

	return &router
}

// Use adds a Component to the middleware stack for an Router
func (router *Router) Use(component ...Component) {
	router.components = append(router.components, component...)
}

// Handle registers a handler for a pattern.
func (router *Router) Handle(pattern string, handler http.Handler) {
	for i := range router.components {
		handler = router.components[len(router.components)-1-i].Next(handler)
	}

	router.rmux.Handle(pattern, handler)
}

// HandleFunc registers a handler function for a pattern.
func (router *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	router.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP dispatches the request to the internal http.ServeMux for URL normalization then
// is passed to the internal RouterMux for dispatch.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
