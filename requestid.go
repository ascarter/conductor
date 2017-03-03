package conductor

import (
	"context"
	"net/http"
)

type ridKey string

const requestIDKey ridKey = "github.com/ascarter/conductor/RequestID"

// newContextWithRequestID creates a context with request id set to
// X-Request-ID header or generates a new unique ID
func newContextWithRequestID(ctx context.Context, r *http.Request) context.Context {
	rid := r.Header.Get("X-Request-ID")
	if rid == "" {
		rid = NewUUID().String()
		r.Header.Set("X-Request-ID", rid)
	}
	return context.WithValue(ctx, requestIDKey, rid)
}

// RequestIDFromContext returns the request id from request context
func RequestIDFromContext(ctx context.Context) (string, bool) {
	rid, ok := ctx.Value(requestIDKey).(string)
	return rid, ok
}

// RequestIDHandler sets unique request id.
// If header `X-Request-ID` is already present in the request, that is considered the
// request id.
func RequestIDHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithRequestID(r.Context(), r)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDComponent returns a RequestIDHandler as a component
func RequestIDComponent() Component { return ComponentFunc(RequestIDHandler) }
