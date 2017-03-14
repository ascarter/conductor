package conductor_test

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
	params, ok := conductor.FromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get count from first regular expression match
	count, err := strconv.Atoi(params["$1"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "%d\n", i)
	}
}

func paramHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Running posts handler")
	params, ok := conductor.FromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get post id
	id, ok := params["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "post %s", id)
}

func Example() {
	mux := conductor.NewRouter()

	// Define middleware
	mux.Use(conductor.ComponentFunc(mw))

	// Add simple routes
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/goodbye", goodbyeHandler)

	// Add regexp routes
	mux.HandleFunc(`/list/([0-9]+)$`, listHandler)

	// Add parameterized routes
	mux.HandleFunc(`/posts/:id`, paramHandler)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", mux))
}
