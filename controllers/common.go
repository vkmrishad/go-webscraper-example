package controllers

import (
	"fmt"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Health check endpoint
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "API is up and running")
}
