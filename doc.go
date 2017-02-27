/*
Package conductor provides a library of web application utilities.

Create a new Router instance and define middleware components:

	router := conductor.NewRouter()
	router.Use(conductor.RequestIDComponent)
	router.Use(conductor.DefaultRequestLogComponent)
	router.Use(conductor.ComponentFunc(myMiddlwareHandler))

A Router instance is ServeMux compatible Handler. Register routes to Router:

	// Static resources
	files := http.FileServer(http.Dir("./assets"))
	router.Handle("/static/", http.StripPrefix("/static/", files))

	// Handler func
	router.HandleFunc("/hello", helloHandler)


Run server via ListenAndServe:

	log.Fatal(":8080", app)

*/
package conductor
