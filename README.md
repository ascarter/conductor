# conductor [![GoDoc](https://godoc.org/github.com/ascarter/conductor?status.svg)](http://godoc.org/github.com/ascarter/conductor)

Conductor is an HTTP handling library for Go. It provides `net/http` compatible extensions for routing and middleware.

Conductor draws some inspiration from how [Ruby on Rails](http://rubyonrails.org) applications are structured.


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

func mw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("middleware before")
		h.ServeHTTP(w, r)
		log.Println("middleware after")
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
	router := conductor.NewRouter()

	// Define middleware in order
	router.Use(conductor.RequestIDComponent())
	router.Use(conductor.DefaultRequestLogComponent)
	router.Use(conductor.ComponentFunc(mw))

	// Add handlers
	router.HandleFunc("/hello", helloHandler)
	router.HandleFunc("/goodbye", goodbyeHandler)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

# Routing

Conductor provides a routing component `Router` which can handle several flavors of route patterns. The `Router` utilizes `http.ServeMux` and `RouterMux` to dispatch requests. The `http.ServeMux` provides URL canonicalization and maps all routes to `RouterMux`.

`RouterMux` provides a map of patterns to routes. The patterns can take the following forms:

Pattern | Description
------- | -----------
`/posts` | Exact match for `/posts`
`/posts/` | Match anything that begins with `/posts/`
`/posts/:id` | Match `/posts/123` with params `{ "id": "23" }` 
`/posts/:id/comments/:author` | Match `/posts/1/comments/obama` with params `{ "id": "1", "author": "obama" }`
`/posts/(\d+)` | Match using regular expression with params `{ "$1": "<match1>" }`
`/posts/(?<id>\d+)` | Match using regular expression with params `{ "id": "<match_id>" }`
`.*/author` | Match using regular expression

# Middleware

Conductor defines an interface for building middleware components. A `Component` is anything that implements `Next(http.Handler) http.Handler`. This provides a pattern for wrapping handlers.

An example middleware:

```go
	func logMiddleware(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("middleware before")
			h.ServeHTTP(w, r)
			log.Println("middleware after")
		})
	}
```

`logMiddleware` can be treated as a component because it implements a `ComponentFunc` that takes an `http.Handler` and returns a new `http.Handler`. In the `logMiddleware` case, this is an adapted func using `http.HandleFunc` which prints logging before and after.

Middleware is added in order to a `Router` by the `Use` func:

```go
	// Define middleware in order
	router := conductor.NewRouter()
	router.Use(conductor.RequestIDComponent())
	router.Use(conductor.DefaultRequestLogComponent)
	router.Use(conductor.ComponentFunc(mw))
```

The middleware stack will be executed in order before calling the route handler.


# References

**conductor** was influenced by the following:

* [Joe Shaw - Revisiting context and http.Handler for Go 1.7](https://joeshaw.org/revisiting-context-and-http-handler-for-go-17/)
* [alice by justinas](https://github.com/justinas/alice)
* [Ruby on Rails](http://rubyonrails.org)

[![Go Report Card](https://goreportcard.com/badge/github.com/ascarter/conductor)](https://goreportcard.com/report/github.com/ascarter/conductor)
