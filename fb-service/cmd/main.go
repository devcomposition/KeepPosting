package main

import (
	p "fb_service/pkg"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// test db connectivity
	p.ConnectDatabase()

	// add routes
	route := mux.NewRouter()
	p.AddRoutes(route)

	log.Println("Starting the server at 8080...")
	log.Fatal(http.ListenAndServe(":8080", route))
}
