package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vkmrishad/go-webscraper-example/controllers"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health-check", controllers.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/scraper", controllers.ScraperHandler).Methods("POST")
	http.Handle("/", router)

	//start and listen to requests
	fmt.Println("Starting API server")
	http.ListenAndServe(":8080", router)
}
