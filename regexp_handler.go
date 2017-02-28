package conductor

import (
	"context"
	"net/http"
	"regexp"
	"sync"
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

// A RegexpMux is an HTTP request multiplexer for regular expression patterns.
// It matches URL of each incoming request against a list of registered regular
// expressions and calls the handler for best pattern match.
//
// Patterns are regular expressions. Longer patterns take precedence over shorter
// ones. If there are patterns that match both "/images\/.*" and "/images/thumbnails\/.*"
// the path "/images/thumbnails" would use the later handler.
//
// Patterns may optionally begin with a host name, restricting matches to URLs on that
// host only. Host specific patterns take precedence over general patterns.
//
// RegexpMux follows the general approach used by http.ServeMux.
type RegexpMux struct {
	mu    sync.RWMutex
	m     map[string]regexpMuxEntry
	hosts bool
}

type regexpMuxEntry struct {
	h       http.Handler
	pattern *regexp.Regexp
}

// NewRegexpMux allocates and returns a new RegexpMux.
func NewRegexpMux() *RegexpMux {
	return new(RegexpMux)
}

// match finds the best regular expression match for the method and path.
func (mux *RegexpMux) match(path string) (h http.Handler, pattern string) {
	var n = 0
	for k, v := range mux.m {
		if !v.pattern.MatchString(path) {
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

// Handler returns the handler to use for the given request, consulting r.Host
// and r.URL.Path. It always returns a non-nil handler.
//
// Handler also returns the registered regular expression pattern that matches the request.
//
// If there is no registered handler that applies to the request, Handler returns
// a ``page not found'' handler and an empty pattern.
func (mux *RegexpMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	if mux.hosts {
		h, pattern = mux.match(r.Host + r.URL.Path)
	}

	// If no host match, match generic patterns
	if h == nil {
		h, pattern = mux.match(r.URL.Path)
	}

	// No handler matches
	if h == nil {
		h, pattern = http.NotFoundHandler(), ""
	}

	return
}

// ServeHTTP dispatches request to the handler whose regular expression pattern most
// closely matches the request URL.
func (mux *RegexpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, pattern := mux.Handler(r)
	entry, ok := mux.m[pattern]
	if ok {
		matches := entry.pattern.FindStringSubmatch(r.URL.Path)
		ctx := newContextWithRegexpMatch(r.Context(), matches)
		r = r.WithContext(ctx)
	}
	h.ServeHTTP(w, r)
}

// Handle registers the handler for a give regular expression pattern.
//
// If the handler already exists for pattern or the regular expression does not compile,
// Handle panics.
func (mux *RegexpMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Verify parameters
	if pattern == "" {
		panic("regexp mux: invalid pattern " + pattern)
	}
	re := regexp.MustCompile(pattern)

	if handler == nil {
		panic("regexp mux: nil handler")
	}

	if _, ok := mux.m[pattern]; ok {
		panic("regexp mux: multiple registrations for " + pattern)
	}

	if mux.m == nil {
		mux.m = make(map[string]regexpMuxEntry)
	}

	mux.m[pattern] = regexpMuxEntry{h: handler, pattern: re}

	if pattern[0] != '/' {
		mux.hosts = true
	}
}

// HandleFunc registers the handler function for the given regular expression pattern.
func (mux *RegexpMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}
