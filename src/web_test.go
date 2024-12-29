package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handleIndex(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestHandleWeather(t *testing.T) {
	// Valid request
	req := httptest.NewRequest(http.MethodGet, "/weather?lat=51.509865&lon=-0.118092", nil)
	rec := httptest.NewRecorder()

	handleWeather(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Missing parameters
	req = httptest.NewRequest(http.MethodGet, "/weather", nil)
	rec = httptest.NewRecorder()

	handleWeather(rec, req)

	res = rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestHandleSuggestions(t *testing.T) {
	// Mock API response
	mockResponse := GeocodingResponse{
		Results: []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Name      string  `json:"name"`
			Country   string  `json:"country"`
			Admin1    string  `json:"admin1"`
			Admin2    string  `json:"admin2"`
			Admin3    string  `json:"admin3"`
			Admin4    string  `json:"admin4"`
		}{
			{
				Name:      "Mock City",
				Country:   "Mockland",
				Latitude:  51.509865,
				Longitude: -0.118092,
			},
		},
	}

	// Start mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Override the API endpoint with the mock server URL
	OpenMeteoGeoAPIEndpoint = mockServer.URL

	// Encode the query to handle spaces properly
	encodedQuery := url.QueryEscape("Mock City")

	// Test the handler with encoded URL
	req := httptest.NewRequest(http.MethodGet, "/suggestions?q="+encodedQuery, nil)
	rec := httptest.NewRecorder()

	handleSuggestions(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var suggestions []Suggestion
	err := json.NewDecoder(res.Body).Decode(&suggestions)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(suggestions) == 0 || suggestions[0].Name != "Mock City" {
		t.Errorf("Expected Mock City in response, got %+v", suggestions)
	}
}

func TestHandleRobots(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()

	handleRobots(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", contentType)
	}
}

func TestHandleFavicon(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/static/favicon.ico", nil)
	rec := httptest.NewRecorder()

	handleFavicon(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}
