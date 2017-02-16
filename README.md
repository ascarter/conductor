# conductor

Conductor is a web application library for Go. It provides some helpers for wiring middleware. It draws some inspiration from how [Ruby on Rails](http://rubyonrails.org) applications are structured.

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
	app := conductor.NewApp()

	// Define middleware in order
	app.Use(conductor.RequestIDComponent)
	app.Use(conductor.DefaultRequestLogComponent)
	app.Use(conductor.ComponentFunc(mw))

	// Add handlers
	app.HandleFunc("/hello", helloHandler)
	app.HandleFunc("/goodbye", goodbyeHandler)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, app))
}

```

# References

**conductor** was influenced by the following:

* [Joe Shaw - Revisiting context and http.Handler for Go 1.7](https://joeshaw.org/revisiting-context-and-http-handler-for-go-17/)
* [alice by justinas](https://github.com/justinas/alice)
* [Ruby on Rails](http://rubyonrails.org)

[![Go Report Card](https://goreportcard.com/badge/github.com/ascarter/conductor)](https://goreportcard.com/report/github.com/ascarter/conductor)
