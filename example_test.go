package conductor_test

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

func Example() {
	c := conductor.New()

	// Define middleware
	c.Use(logRequest)

	// Add routes
	http.Handle("/hello", c.HandlerFunc(helloHandler))
	http.Handle("/goodbye", c.HandlerFunc(goodbyeHandler))

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
