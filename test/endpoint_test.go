package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vkmrishad/go-webscraper-example/controllers"
)

func TestHealthCheckHandler(t *testing.T) {
	// Test HealthCheck endpoint
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.HealthCheckHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `API is up and running`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestScraperHandlerWithoutRequetBody(t *testing.T) {
	// Test ScraperHandler endpoint without request body
	var jsonStr = []byte(``)

	req, err := http.NewRequest("POST", "/scraper/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.ScraperHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `{"errors":{"code":400,"message":"Request body is empty"}}`
	if strings.Trim(rr.Body.String(), " \r\n") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.Trim(rr.Body.String(), " \r\n"), expected)
	}
}

func TestScraperHandlerUrlRequired(t *testing.T) {
	// Test ScraperHandler endpoint URL required
	var jsonStr = []byte(`{}`)

	req, err := http.NewRequest("POST", "/scraper/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.ScraperHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `{"validation_errors":{"Url":["This field is required"]}}`
	if strings.Trim(rr.Body.String(), " \r\n") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.Trim(rr.Body.String(), " \r\n"), expected)
	}
}

func TestScraperHandlerUrlFormatCheck(t *testing.T) {
	// Test ScraperHandler endpoint URL format check
	var jsonStr = []byte(`{"url":"example"}`)

	req, err := http.NewRequest("POST", "/scraper/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.ScraperHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `{"validation_errors":{"Url":["Not a valid URL"]}}`
	if strings.Trim(rr.Body.String(), " \r\n") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.Trim(rr.Body.String(), " \r\n"), expected)
	}
}

func TestScraperHandlerWithValidAndUnReachableUrl(t *testing.T) {
	// Test ScraperHandler endpoint with valid and unreachable URL
	var jsonStr = []byte(`{"url":"https://mohammedrishad.co.in"}`)

	req, err := http.NewRequest("POST", "/scraper/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.ScraperHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `{"errors":{"code":400,"message":"URL is not reachable"}}`
	if strings.Trim(rr.Body.String(), " \r\n") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.Trim(rr.Body.String(), " \r\n"), expected)
	}
}

func TestScraperHandlerValidUrl(t *testing.T) {
	// Test ScraperHandler endpoint with valid URL
	var jsonStr = []byte(`{"url":"https://mohammedrishad.com"}`)

	req, err := http.NewRequest("POST", "/scraper/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.ScraperHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
