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
