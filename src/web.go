package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//go:embed templates/index.html
var indexHTML string

//go:embed static/*
var StaticFiles embed.FS

var indexTmpl = template.Must(template.New("index").Parse(indexHTML))

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Only serve the root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Only allow GET
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Log HTTP request
	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)

	// Retrieve cookies
	cityNameCookie, _ := r.Cookie("cityName")
	latCookie, _ := r.Cookie("latitude")
	lonCookie, _ := r.Cookie("longitude")

	cityName := ""
	latitude := ""
	longitude := ""

	// Set defaults if cookies exist
	if cityNameCookie != nil {
		cityName = cityNameCookie.Value
	}
	if latCookie != nil {
		latitude = latCookie.Value
	}
	if lonCookie != nil {
		longitude = lonCookie.Value
	}

	// Render template with automatic HTML escaping
	w.Header().Set("Content-Type", "text/html")
	data := struct {
		CityName  string
		Latitude  string
		Longitude string
	}{cityName, latitude, longitude}
	if err := indexTmpl.Execute(w, data); err != nil {
		log.Printf("ERROR: rendering index: %v", err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		return
	}
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	unitTemp := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("unit_temp")))
	unitWind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("unit_wind")))
	time12h := strings.TrimSpace(r.URL.Query().Get("time_12h")) == "1"

	if lat == "" || lon == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	latitude, err1 := strconv.ParseFloat(lat, 64)
	longitude, err2 := strconv.ParseFloat(lon, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid latitude or longitude", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Requested weather data for lat: %s, lon: %s", lat, lon)

	data := OpenMeteoAPIResponse{}
	data.FetchData(OpenMeteoAPIEndpoint, OpenMeteoAPIParams, float64ToString(latitude), float64ToString(longitude))
	opts := PrintOptions{TemperatureUnit: unitTemp, WindSpeedUnit: unitWind, Use12Hour: time12h}
	weatherTable := data.Points().setMoonIllumination().setSeeing().PrintWithOptions(opts)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, weatherTable)
}

func handleSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	query = strings.TrimSpace(query)

	if query == "" {
		http.Error(w, "Query parameter is missing", http.StatusBadRequest)
		return
	}

	suggestions, err := fetchSuggestions(query)
	if err != nil {
		http.Error(w, "Unable to fetch suggestions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		http.Error(w, "Unable to encode suggestions", http.StatusInternalServerError)
		return
	}
}

func handleReverseGeocoding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	latStr := strings.TrimSpace(r.URL.Query().Get("lat"))
	lonStr := strings.TrimSpace(r.URL.Query().Get("lon"))

	if latStr == "" || lonStr == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	// Validate numeric input
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid latitude or longitude", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Reverse geocoding for lat: %s, lon: %s", latStr, lonStr)
	suggestion, _ := fetchReverseGeocoding(float64ToString(lat), float64ToString(lon))

	w.Header().Set("Content-Type", "application/json")
	// Always return 200 with either a suggestion or an empty object to keep UX smooth
	if suggestion == nil {
		if _, err := w.Write([]byte(`{}`)); err != nil {
			log.Printf("ERROR: reverse-geocoding write: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(suggestion); err != nil {
		http.Error(w, "Unable to encode result", http.StatusInternalServerError)
		return
	}
}

func handleRobots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)
	serveEmbeddedFile(w, "static/robots.txt", "text/plain")
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)
	serveEmbeddedFile(w, "static/favicon.ico", "image/x-icon")
}

func handleSitemap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		// Prefer https unless explicitly forwarded otherwise
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "https"
		}
	}
	host := r.Host
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>%s/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>`, baseURL)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(sitemap)); err != nil {
		log.Printf("ERROR: Could not write sitemap.xml: %v", err)
	}
}

func float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

// serveEmbeddedFile writes an embedded static file with the given content type
func serveEmbeddedFile(w http.ResponseWriter, path, contentType string) {
	data, err := StaticFiles.ReadFile(path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		log.Printf("ERROR: Could not read %s: %v", path, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("ERROR: Could not write %s: %v", path, err)
	}
}
