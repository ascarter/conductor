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
	params, ok := conductor.RouteParamsFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	log.Printf("route params: %+v", params)
	count, err := strconv.Atoi(params["1"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "%d\n", i)
	}
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Running posts handler")
	params, ok := conductor.RouteParamsFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	log.Printf("route params: %+v", params)

	id, ok := params["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "post %s", id)
}

func main() {
	mux := conductor.NewRouter()

	// Define middleware in order
	mux.Use(conductor.RequestIDComponent)
	mux.Use(conductor.DefaultRequestLogComponent)
	mux.Use(conductor.ComponentFunc(mw))

	// Add simple routes
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/goodbye", goodbyeHandler)

	// Add regexp routes
	mux.HandleFunc(`/list/([0-9]+)$`, listHandler)

	// Add parameterized routes
	mux.HandleFunc(`/posts/:id`, postsHandler)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
