package main

import (
	"fmt"
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
	router := conductor.NewRouter()

	// Define middleware in order
	router.Use(conductor.RequestIDComponent)
	router.Use(conductor.DefaultRequestLogComponent)
	router.Use(conductor.ComponentFunc(mw))

	// Add pattern routes
	rmux := conductor.NewRegexpMux()
	rmux.HandleFunc(`/list/([0-9]+)$`, listHandler)

	// Route everything to rmux
	router.Handle("/", rmux)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
