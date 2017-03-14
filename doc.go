/*
Package conductor is an HTTP handling library.
It provides `net/http` compatible extensions for routing and middleware.

Create a new Router instance and define middleware components:

	mux := conductor.NewRouter()
	mux.Use(conductor.ComponentFunc(myMiddlwareHandler))

A Router instance is an http.ServeMux compatible http.Handler.

Register routes to Router:

	// Static resources
	files := http.FileServer(http.Dir("./assets"))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	// Handler func
	mux.HandleFunc("/hello", helloHandler)


Run server via ListenAndServe:

	log.Fatal(":8080", mux)

*/
package conductor
