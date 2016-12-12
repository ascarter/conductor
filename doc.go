/*
Package conductor provides a library of web application utilities.

Create a new App instance and define middleware components:

	app := conductor.NewApp()
	app.Use(requestid.RequestIDComponent)
	app.Use(logging.DefaultLoggingComponent)
	app.Use(conductor.ComponentFunc(myMiddlwareHandler))

An App instance is ServeMux compatible Handler. Register routes to App:

	// Static resources
	files := http.FileServer(http.Dir("./assets"))
	app.Handle("/static/", http.StripPrefix("/static/", files))

	// Handler func
	app.HandleFunc("/hello", helloHandler)


Run server via ListenAndServe:

	log.Fatal(":8080", app)

*/
package conductor
