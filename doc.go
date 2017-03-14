/*
Package conductor is an HTTP middleware library.

Conductor provides an easy way to define a sequence of middleware and apply
to handlers when executed.

Create a new Conductor instance and add components:

	c := conductor.New()
	c.Use(myMiddlwareHandler)

Components are handlers that wrap another handler. Any func with the following
signature is a Component:

	func(http.Handler) http.Handler

A Conductor can return a handler for any http.Handler that applies the
entire component sequence calling the input http.Handler last. There is also
a variation that accepts a http.HandlerFunc and wraps it.

	mux := http.NewServeMux()
	mux.Handle("/posts", c.Handler(postsHandler))
	mux.Handle("/comments", c.HandlerFunc(commentsFn))

*/
package conductor
