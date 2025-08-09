package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	handleFavicon(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestHandleSuggestions_NoQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/suggestions", nil)
	rec := httptest.NewRecorder()

	handleSuggestions(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestHandleSuggestions_Success(t *testing.T) {
	setupCache()

	// Mock upstream geo API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[{"name":"Berlin","country":"Germany","latitude":52.52,"longitude":13.405}]}`))
	}))
	defer ts.Close()

	original := OpenMeteoGeoAPIEndpoint
	OpenMeteoGeoAPIEndpoint = ts.URL
	defer func() { OpenMeteoGeoAPIEndpoint = original }()

	req := httptest.NewRequest(http.MethodGet, "/suggestions?q=Berlin", nil)
	rec := httptest.NewRecorder()
	handleSuggestions(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Expected application/json, got %s", ct)
	}
	body, _ := io.ReadAll(res.Body)
	var got []Suggestion
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Berlin" {
		t.Fatalf("Unexpected suggestions: %+v", got)
	}
}

func TestHandleReverseGeocoding_InvalidParams(t *testing.T) {
	// Missing both
	req := httptest.NewRequest(http.MethodGet, "/reverse-geocoding", nil)
	rec := httptest.NewRecorder()
	handleReverseGeocoding(rec, req)
	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d for missing params, got %d", http.StatusBadRequest, rec.Result().StatusCode)
	}

	// Invalid lat
	req = httptest.NewRequest(http.MethodGet, "/reverse-geocoding?lat=bad&lon=10", nil)
	rec = httptest.NewRecorder()
	handleReverseGeocoding(rec, req)
	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d for invalid lat, got %d", http.StatusBadRequest, rec.Result().StatusCode)
	}
}

func TestHandleReverseGeocoding_Success(t *testing.T) {
	setupCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[{"name":"Berlin","country":"Germany","latitude":52.52,"longitude":13.405}]}`))
	}))
	defer ts.Close()

	original := OpenMeteoGeoReverseAPIEndpoint
	OpenMeteoGeoReverseAPIEndpoint = ts.URL
	defer func() { OpenMeteoGeoReverseAPIEndpoint = original }()

	req := httptest.NewRequest(http.MethodGet, "/reverse-geocoding?lat=52.5200&lon=13.4050", nil)
	rec := httptest.NewRecorder()
	handleReverseGeocoding(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Expected application/json, got %s", ct)
	}
	var got Suggestion
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if got.Name != "Berlin" {
		t.Fatalf("Unexpected suggestion: %+v", got)
	}
}

func TestHandleReverseGeocoding_UpstreamNon200(t *testing.T) {
	setupCache()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	original := OpenMeteoGeoReverseAPIEndpoint
	OpenMeteoGeoReverseAPIEndpoint = ts.URL
	defer func() { OpenMeteoGeoReverseAPIEndpoint = original }()

	req := httptest.NewRequest(http.MethodGet, "/reverse-geocoding?lat=52.52&lon=13.405", nil)
	rec := httptest.NewRecorder()
	handleReverseGeocoding(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if strings.TrimSpace(string(body)) != "{}" {
		t.Fatalf("Expected empty object, got: %s", string(body))
	}
}

func TestHandleSitemap(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	req.Host = "example.com"
	req.Header.Set("X-Forwarded-Proto", "http")
	rec := httptest.NewRecorder()

	handleSitemap(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/xml") {
		t.Fatalf("Expected application/xml, got %s", ct)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "<loc>http://example.com/</loc>") {
		t.Fatalf("Sitemap body unexpected: %s", string(body))
	}
}

func TestMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		path    string
	}{
		{"index", handleIndex, "/"},
		{"weather", handleWeather, "/weather"},
		{"suggestions", handleSuggestions, "/suggestions"},
		{"reverse", handleReverseGeocoding, "/reverse-geocoding"},
		{"robots", handleRobots, "/robots.txt"},
		{"favicon", handleFavicon, "/favicon.ico"},
		{"sitemap", handleSitemap, "/sitemap.xml"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			res := rec.Result()
			if res.StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("Expected 405, got %d", res.StatusCode)
			}
			if allow := res.Header.Get("Allow"); allow != http.MethodGet {
				t.Fatalf("Expected Allow header %q, got %q", http.MethodGet, allow)
			}
		})
	}
}

func TestFloat64ToString(t *testing.T) {
	if got := float64ToString(1.2); got != "1.200000" {
		t.Fatalf("Expected 1.200000, got %s", got)
	}
}
