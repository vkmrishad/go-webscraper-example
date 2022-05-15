package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vkmrishad/go-webscraper-example/controllers"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health-check", controllers.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/scraper/", controllers.ScraperHandler).Methods("POST")
	http.Handle("/", router)

	//start and listen to requests
	fmt.Println("Starting API server")
	log.Fatal(http.ListenAndServe(":8080", router))
}
