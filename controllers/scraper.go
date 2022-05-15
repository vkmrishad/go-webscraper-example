package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/vkmrishad/go-webscraper-example/models"
	"github.com/vkmrishad/go-webscraper-example/utility"
)

func ScraperHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	requestBody := &models.ScraperRequestBody{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(requestBody); err != nil {
		msg := models.ErrorMessage{
			Error: models.Error{Code: http.StatusBadRequest, Message: "Request body is empty"},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	if errors := utility.Validate(requestBody); len(errors) > 0 {
		err := map[string]interface{}{"validationError": errors}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	} else {
		_url := requestBody.Url

		doc, err := goquery.NewDocument(_url)

		if err != nil {
			msg := models.ErrorMessage{
				Error: models.Error{Code: http.StatusBadRequest, Message: "URL is not reachable"},
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(msg)
			return
		}

		// Return webscraper response
		webScraperResponse := utility.WebScraper(_url, *doc)

		jsonResponse, jsonError := json.Marshal(webScraperResponse)

		if jsonError != nil {
			fmt.Println("Unable to encode JSON")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
