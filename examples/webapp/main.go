package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"

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

func listHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Running list handler")
	matches, ok := conductor.RegexpMatchesFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	log.Printf("matches: %+v", matches)
	count, err := strconv.Atoi(matches[1])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "%d\n", i)
	}
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

	// Add pattern route
	listRoutes := conductor.RegexpRouteMap{}
	listRoutes.AddRouteFunc(http.MethodGet, `/list/([0-9]+)$`, listHandler)
	app.Handle("/list/", conductor.RegexpHandler(listRoutes))

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, app))
}
