package conductor

import "net/http"

// A Component is a middleware element that can chain http.Handler requests
type Component interface {
	Next(http.Handler) http.Handler
}

// A ComponentFunc is an adapter that allows the use of ordinary function as
// a App component
type ComponentFunc func(http.Handler) http.Handler

// Next calls f(h)
func (f ComponentFunc) Next(h http.Handler) http.Handler {
	return f(h)
}
