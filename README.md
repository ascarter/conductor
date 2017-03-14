# conductor [![GoDoc](https://godoc.org/github.com/ascarter/conductor?status.svg)](http://godoc.org/github.com/ascarter/conductor)[![Go Report Card](https://goreportcard.com/badge/github.com/ascarter/conductor)](https://goreportcard.com/report/github.com/ascarter/conductor)

Conductor is an HTTP component library.

Conductor provides an easy way to define a sequence of middleware components that is applied to handlers when executed.

Create a new Conductor instance and add components:

```go
c := conductor.New()
c.Use(myMiddlwareHandler)
```

Components are handlers that wrap another handler. Any func with the following
signature is a Component:

```go
func(http.Handler) http.Handler
```

A Conductor can return a handler for any http.Handler that applies the
entire component sequence calling the input http.Handler last. There is also
a variation that accepts a http.HandlerFunc and wraps it.

```go
mux := http.NewServeMux()
mux.Handle("/posts", c.Handler(postsHandler))
mux.Handle("/comments", c.HandlerFunc(commentsFn))
```

# Usage

This is an example of a simple web app using conductor. For more detailed example,
see `example_test.go` file.

```go

package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/ascarter/conductor"
)

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("starting request %s", r.URL.Path)
		h.ServeHTTP(w, r)
		log.Println("completed request %s", r.URL.Path)
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Running hello handler")
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func goodbyeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Running goodbye handler")
	fmt.Fprintf(w, "Goodbye, %q", html.EscapeString(r.URL.Path))
}

func main() {
	// Define middleware
	c := conductor.New()
	c.Use(mw)

	// Add routes
	http.Handle("/hello", c.HandlerFunc(helloHandler))
	http.Handle("/goodbye", c.HandlerFunc(goodbyeHandler))

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

# References

**conductor** was influenced by the following:

* [Joe Shaw - Revisiting context and http.Handler for Go 1.7](https://joeshaw.org/revisiting-context-and-http-handler-for-go-17/)
* [alice by justinas](https://github.com/justinas/alice)
* [Mat Ryer - The http.Handler wrapper technique in golang](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.x92zrij2u)
