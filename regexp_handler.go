package conductor

import (
	"context"
	"errors"
	"net/http"
	"regexp"
)

type reKey string

const regexpMatchKey reKey = "github.com/ascarter/conductor/RegexpRouteMatchKey"

// newContextWithRegexpMatch creates context with regular expression matches
func newContextWithRegexpMatch(ctx context.Context, matches []string) context.Context {
	return context.WithValue(ctx, regexpMatchKey, matches)
}

// RegexpMatchesFromContext returns slice of regular expression matches from context if any
func RegexpMatchesFromContext(ctx context.Context) ([]string, bool) {
	matches, ok := ctx.Value(regexpMatchKey).([]string)
	return matches, ok
}

// A RegexpRoute defines the Handler for a regular expression Pattern
type RegexpRoute struct {
	Pattern *regexp.Regexp
	Handler http.Handler
}

// NewRegexpRoute returns a RegexpRoute for a pattern to a handler
func NewRegexpRoute(pattern string, handler http.Handler) (*RegexpRoute, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexpRoute{re, handler}, nil
}

// A RegexpRouteMap associates HTTP methods to slices of RegexpRoutes
//
// Example:
//	routes := RegexpRouteMap{}
//	routes.AddRouteFunc(http.MethodGet, `/posts[/]?$`, handleGetAllPosts)
//	routes.AddRoute(http.MethodGet, `/posts/[0-9]+$`, postHandler)
type RegexpRouteMap map[string][]*RegexpRoute

// AddRoute defines route for HTTP method and pattern to handler
func (m RegexpRouteMap) AddRoute(method, pattern string, handler http.Handler) error {
	route, err := NewRegexpRoute(pattern, handler)
	if err != nil {
		return err
	}

	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		m[method] = append(m[method], route)
	default:
		return errors.New("invalid HTTP method")
	}

	return nil
}

// AddRouteFunc defines route for HTTP method and pattern to handler func
func (m RegexpRouteMap) AddRouteFunc(method, pattern string, fn http.HandlerFunc) error {
	return m.AddRoute(method, pattern, http.HandlerFunc(fn))
}

type regexpHandler struct {
	routes RegexpRouteMap
}

func (rh *regexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find routes for HTTP method
	routes, ok := rh.routes[r.Method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	for _, route := range routes {
		if route.Pattern.MatchString(r.URL.Path) {
			matches := route.Pattern.FindStringSubmatch(r.URL.Path)
			ctx := newContextWithRegexpMatch(r.Context(), matches)
			route.Handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}

	// No matching route found
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

// RegexpHandler returns a request handler that
func RegexpHandler(routes RegexpRouteMap) http.Handler {
	return &regexpHandler{routes}
}
