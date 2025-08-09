package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
)

func setupCache() {
	var err error
	cache, err = bigcache.New(context.Background(), bigcache.DefaultConfig(5*time.Minute))
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
}

func TestFetchData_Success(t *testing.T) {
	setupCache() // Initialize cache

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"latitude": 52.52,
			"longitude": 13.405,
			"hourly": {
				"time": ["2024-01-01T00:00"],
				"temperature_2m": [5.0]
			}
		}`))
	}))
	defer server.Close()

	response := OpenMeteoAPIResponse{}
	response.FetchData(server.URL+"?", "temperature_2m", "52.52", "13.405")

	if response.Latitude != 52.52 {
		t.Errorf("Expected latitude 52.52, got %f", response.Latitude)
	}

	if len(response.Hourly.Time) == 0 || response.Hourly.Temperature2M[0] != 5.0 {
		t.Error("Hourly temperature data not fetched correctly")
	}
}

func TestFetchData_Error(t *testing.T) {
	setupCache() // Initialize cache

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	response := OpenMeteoAPIResponse{}
	response.FetchData(server.URL+"?", "temperature_2m", "52.52", "13.405")

	if response.Latitude != 0 {
		t.Error("Expected latitude 0 on error response")
	}
}

func TestFetchSuggestions_Success(t *testing.T) {
	setupCache() // Initialize cache

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"results": [{
				"name": "Berlin",
				"country": "Germany",
				"latitude": 52.52,
				"longitude": 13.405
			}]
		}`))
	}))
	defer server.Close()

	OpenMeteoGeoAPIEndpoint = server.URL
	suggestions, err := fetchSuggestions("Berlin")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(suggestions) != 1 || suggestions[0].Name != "Berlin" {
		t.Error("Failed to fetch or parse suggestions correctly")
	}
}

func TestFetchReverseGeocoding_CacheAndNormalize(t *testing.T) {
	setupCache()
	// Pre-cache a value under normalized key reverse:52.520,13.405
	key := "reverse:52.520,13.405"
	payload, _ := json.Marshal(Suggestion{Name: "Cached Berlin", Country: "DE"})
	if err := cache.Set(key, payload); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	// Values with more precision should normalize to the same key
	got, err := fetchReverseGeocoding("52.52000", "13.40500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Name != "Cached Berlin" {
		t.Fatalf("expected cached suggestion, got: %+v", got)
	}
}

func TestFetchSuggestions_Cache(t *testing.T) {
	setupCache() // Initialize cache

	// Store a suggestion in cache for testing
	suggestion := []Suggestion{
		{Name: "CachedCity", Country: "Testland"},
	}
	jsonData, _ := json.Marshal(suggestion)
	cache.Set("CachedCity", jsonData)

	suggestions, err := fetchSuggestions("CachedCity")

	if err != nil {
		t.Fatalf("Unexpected error fetching cached data: %v", err)
	}

	if len(suggestions) == 0 || suggestions[0].Name != "CachedCity" {
		t.Error("Expected CachedCity from cache, but did not find it")
	}
}

func TestPoints(t *testing.T) {
	response := OpenMeteoAPIResponse{
		Latitude:  52.52,
		Longitude: 13.405,
		Hourly: Hourly{
			Time:                  []string{"2024-01-01T00:00"},
			Temperature2M:         []float64{5.0},
			Temperature500hPa:     []float64{0.0}, // Matching length
			Temperature850hPa:     []float64{0.0}, // Matching length
			WindSpeed200hPa:       []float64{0.0}, // Matching length
			WindSpeed850hPa:       []float64{0.0}, // Matching length
			CloudCoverLow:         []int64{0},     // Matching length
			CloudCoverMid:         []int64{0},     // Matching length
			CloudCoverHigh:        []int64{0},     // Matching length
			WindSpeed10M:          []float64{0.0}, // Matching length
			WindGusts10M:          []float64{0.0}, // Matching length
			GeopotentialHeight850: []float64{0.0},
			GeopotentialHeight500: []float64{0.0},
		},
		Timezone: "Europe/Berlin",
	}

	points := response.Points()
	if len(points) != 1 {
		t.Fatalf("Expected 1 data point, got %d", len(points))
	}

	if points[0].Temperature2M != 5.0 {
		t.Errorf("Expected temperature 5.0, got %f", points[0].Temperature2M)
	}
}
