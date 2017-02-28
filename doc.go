/*
Package conductor provides a library of web application utilities.

Create a new Router instance and define middleware components:

	mux := conductor.NewRouter()
	mux.Use(conductor.RequestIDComponent)
	mux.Use(conductor.DefaultRequestLogComponent)
	mux.Use(conductor.ComponentFunc(myMiddlwareHandler))

A Router instance is ServeMux compatible Handler. Register routes to Router:

	// Static resources
	files := http.FileServer(http.Dir("./assets"))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	// Handler func
	mux.HandleFunc("/hello", helloHandler)


Run server via ListenAndServe:

	log.Fatal(":8080", mux)

*/
package conductor
