package conductor

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

type routeKey string

const routeParamsKey routeKey = "github.com/ascarter/conductor/RouteParamsKey"

// newContextWithRegexpMatch creates context with regular expression matches
func newContextWithRouteParams(ctx context.Context, params RouteParams) context.Context {
	return context.WithValue(ctx, routeParamsKey, params)
}

// RouteParamsFromContext returns a map of params and values extracted from the route match.
func RouteParamsFromContext(ctx context.Context) (RouteParams, bool) {
	params, ok := ctx.Value(routeParamsKey).(RouteParams)
	return params, ok
}

type RouteParams map[string]string

// A route is a RegExp pattern and the handler to call.
type route struct {
	re *regexp.Regexp
	h  http.Handler
}

func newRoute(p string, h http.Handler) *route {
	// TODO: Fix up pattern...
	return &route{h: h, re: regexp.MustCompile(p)}
}

func (r *route) GetParams(path string) RouteParams {
	keys := r.re.SubexpNames()
	values := r.re.FindStringSubmatch(path)

	params := RouteParams{}
	for i, k := range keys {
		if k == "" {
			k = strconv.Itoa(i)
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
// Patterns may optionally begin with a host name, restricting matches to URLs on that
// host only. Host specific patterns take precedence over general patterns.
//
// Patterns can take the following forms:
//	`/posts`
//	`/posts/`
//	`/posts/:id`
//	`/posts/(\d+)`
//	`/posts/(?<id>\d+)`
//	`host/posts`
//
// RouterMux follows the general approach used by http.ServeMux.
type RouterMux struct {
	mu     sync.RWMutex
	routes map[string]*route
	hosts  bool
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

// Handler returns the handler to use for the given request, consulting r.Host, r.Method,
// and r.URL.Path. It always returns a non-nil handler.
//
// Handler also returns the registered pattern that matches the request.
//
// If there is no registered handler that applies to the request, Handler returns
// a ``page not found'' handler and an empty pattern.
func (mux *RouterMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	if mux.hosts {
		h, pattern = mux.match(r.Method, r.Host+r.URL.Path)
	}

	// If no host match, match generic patterns
	if h == nil {
		h, pattern = mux.match(r.Method, r.URL.Path)
	}

	// No handler matches
	if h == nil {
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

// Handle registers the handler for a give pattern.
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

	if pattern[0] != '/' {
		mux.hosts = true
	}
}

// HandleFunc registers the handler function for the given pattern.
func (mux *RouterMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	mux.Handle(pattern, http.HandlerFunc(handler))
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

// Handle registers the handler for the given pattern.
func (router *Router) Handle(pattern string, handler http.Handler) {
	for i := range router.components {
		handler = router.components[len(router.components)-1-i].Next(handler)
	}

	router.rmux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
func (router *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	router.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP dispatches the request to the internal ServeMux for URL normalization then
// is passed to the internal RouterMux for dispatch.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
