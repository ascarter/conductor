package conductor

import (
	"errors"
	"net/http"
	"regexp"
)

// A RegexpRoute defines the Handler for a regular expression Pattern
type RegexpRoute struct {
	Pattern *regexp.Regexp
	Handler http.HandlerFunc
}

// NewRegexpRoute returns a RegexpRoute for a pattern to a handler
func NewRegexpRoute(pattern string, handler http.HandlerFunc) (*RegexpRoute, error) {
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
//	routes.AddRoute(http.MethodGet, RegexpRoute{`/posts[/]?$`, handleGetAllPosts})
//	routes.AddRoute(http.MethodGet, RegexpRoute{`/posts/[0-9]+$`, handleGetPost})
type RegexpRouteMap map[string][]*RegexpRoute

// AddRoute attaches route to map for http method
func (m RegexpRouteMap) AddRoute(method, pattern string, handler http.HandlerFunc) error {
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

type regexpHandler struct {
	routes RegexpRouteMap
}

func (rh *regexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	// Find routes for HTTP method
	routes, ok := rh.routes[r.Method]

	if ok {
		for _, route := range routes {
			if route.Pattern.MatchString(r.URL.Path) {
				route.Handler(w, r)
				return
			}
		}
	}

	// No matching route found
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// RegexpHandler returns a request handler that
func RegexpHandler(routes RegexpRouteMap) http.Handler {
	return &regexpHandler{routes}
}
