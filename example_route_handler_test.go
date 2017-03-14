package conductor_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ascarter/conductor"
)

func postsHandler(w http.ResponseWriter, r *http.Request) {
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

func Example_RouteHandler() {
	// Define routes
	h := conductor.NewRouteHandler()
	h.HandleRouteFunc(http.MethodGet, `/posts/:id`, postsHandler)
	h.HandleRouteFunc(http.MethodGet, `/posts`, postsHandler)
	h.HandleRouteFunc(http.MethodPost, `/posts`, postsHandler)

	// Setup router
	mux := conductor.NewRouter()
	mux.Handle("/posts/", h)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", mux))
}
