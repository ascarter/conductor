package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/ascarter/conductor"
	"github.com/ascarter/conductor/components/logging"
	"github.com/ascarter/conductor/components/requestid"
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

	// Define middleware
	app.Use(requestid.RequestIDComponent)
	app.Use(logging.DefaultLoggingComponent)
	app.Use(conductor.ComponentFunc(mw))

	// Add handlers
	app.HandleFunc("/hello", helloHandler)
	app.HandleFunc("/goodbye", goodbyeHandler)

	// Start server

	//	log.Fatal(app.ListenAndServe())

	log.Fatal(http.ListenAndServe(":8080", app))
}
