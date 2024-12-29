package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock handler to simulate the Open-Meteo Geocoding API response
func mockGeocodingHandler(w http.ResponseWriter, r *http.Request) {
	response := GeocodingResponse{
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
				Latitude:  51.5074,
				Longitude: -0.1278,
				Name:      "London",
				Country:   "United Kingdom",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func TestGetSuggestions(t *testing.T) {
	// Start a test server with the mock handler
	server := httptest.NewServer(http.HandlerFunc(mockGeocodingHandler))
	defer server.Close()

	// Override the API endpoint with the test server URL
	OpenMeteoGeoAPIEndpoint = server.URL

	suggestions, err := getSuggestions("London")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("Expected 1 suggestion, got %d", len(suggestions))
	}

	expected := "London, United Kingdom"
	if suggestions[0] != expected {
		t.Errorf("Expected %s, got %s", expected, suggestions[0])
	}
}

func TestGetCoordinates(t *testing.T) {
	// Start a test server with the mock handler
	server := httptest.NewServer(http.HandlerFunc(mockGeocodingHandler))
	defer server.Close()

	// Override the API endpoint with the test server URL
	OpenMeteoGeoAPIEndpoint = server.URL

	lat, lon, err := getCoordinates("London")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lat != 51.5074 || lon != -0.1278 {
		t.Errorf("Expected coordinates (51.5074, -0.1278), got (%f, %f)", lat, lon)
	}
}

func TestGetCoordinatesNoResults(t *testing.T) {
	// Mock handler with no results
	noResultHandler := func(w http.ResponseWriter, r *http.Request) {
		response := GeocodingResponse{}
		json.NewEncoder(w).Encode(response)
	}

	server := httptest.NewServer(http.HandlerFunc(noResultHandler))
	defer server.Close()

	OpenMeteoGeoAPIEndpoint = server.URL

	_, _, err := getCoordinates("UnknownCity")
	if err == nil {
		t.Fatal("Expected an error for no results, got nil")
	}

	expectedError := "no results found for location"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}
