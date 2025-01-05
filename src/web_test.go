package main

import (
	"net/http"
	"net/http/httptest"
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
	req := httptest.NewRequest(http.MethodGet, "/static/favicon.ico", nil)
	rec := httptest.NewRecorder()

	handleFavicon(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}
