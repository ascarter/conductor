package conductor

import (
	"net/http"
)

// A Component is a handler that wraps another handler
type Component func(http.Handler) http.Handler

// A Conductor is a sequence of components
type Conductor struct {
	components []Component
}

// New returns a new Conductor instance
func New() *Conductor { return &Conductor{} }

// Use adds a Component to the sequence of components
func (c *Conductor) Use(component ...Component) {
	c.components = append(c.components, component...)
}

// Handler returns a handler that wraps h with the component sequence
func (c *Conductor) Handler(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	for i := range c.components {
		h = c.components[len(c.components)-1-i](h)
	}

	return h
}

// HandlerFunc returns a handler that wraps fn with the component sequence
func (c *Conductor) HandlerFunc(fn http.HandlerFunc) http.Handler {
	return c.Handler(fn)
}
