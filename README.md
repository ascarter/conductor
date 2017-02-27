# conductor

Conductor is an HTTP routing and handling library for Go. It provides structure for
stacking middleware components and handling routes based on named paramters and regular
expressions in addition to default standard library handlers.

It draws some inspiration from how [Ruby on Rails](http://rubyonrails.org) applications are structured.

## Usage:

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
	router.Use(conductor.RequestIDComponent)
	router.Use(conductor.DefaultRequestLogComponent)
	router.Use(conductor.ComponentFunc(mw))

	// Add handlers
	router.HandleFunc("/hello", helloHandler)
	router.HandleFunc("/goodbye", goodbyeHandler)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

```

# References

**conductor** was influenced by the following:

* [Joe Shaw - Revisiting context and http.Handler for Go 1.7](https://joeshaw.org/revisiting-context-and-http-handler-for-go-17/)
* [alice by justinas](https://github.com/justinas/alice)
* [Ruby on Rails](http://rubyonrails.org)

[![Go Report Card](https://goreportcard.com/badge/github.com/ascarter/conductor)](https://goreportcard.com/report/github.com/ascarter/conductor)
