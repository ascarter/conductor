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

// An App is a web application object that has a stack of components for the requests
type App struct {
	components []Component
	mux        *http.ServeMux
}

// NewApp returns a new App instance
func NewApp() *App {
	return &App{
		mux: http.NewServeMux(),
	}
}

// Use adds a Component to the middleware stack for an App
func (app *App) Use(component ...Component) {
	app.components = append(app.components, component...)
}

// Handle registers the handler for the given pattern.
// Handler is run using the middleware stack.
func (app *App) Handle(pattern string, handler http.Handler) {
	for i := range app.components {
		handler = app.components[len(app.components)-1-i].Next(handler)
	}

	app.mux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
// Handler is run using the middleware stack
func (app *App) HandleFunc(pattern string, handler http.HandlerFunc) {
	app.Handle(pattern, http.HandlerFunc(handler))
}

// HandleView registers the a handler for view anchored at pattern.
// Handler is run using the middleware stack
//func (app *App) HandleView(pattern, View) {
//	// TODO: Wrap view handler...
//}

// ServeHTTP dispatches the request to the app mux
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.mux.ServeHTTP(w, r)
}
