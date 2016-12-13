package conductor

import "net/http"

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

// ServeHTTP dispatches the request to the app mux
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.mux.ServeHTTP(w, r)
}
