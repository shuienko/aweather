package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// Embed the index.html file
//
//go:embed templates/index.html
var indexHTML string

// Embed the favicon.ico file
//
//go:embed static/favicon.ico
var Favicon []byte

func handleIndex(w http.ResponseWriter, r *http.Request) {
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

	// Insert cookies into JavaScript for the frontend
	pageWithCookies := fmt.Sprintf(indexHTML, cityName, latitude, longitude)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageWithCookies)
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

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
	data.FetchData(OpenMeteoAPIEndpoint, OpenMeteoAPIParams, float64ToSting(latitude), float64ToSting(longitude))
	markdownTable := data.Points().setMoonIllumination().setSeeing().Print()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, markdownTable)
}

func handleSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter is missing", http.StatusBadRequest)
		return
	}

	suggestions, err := fetchSuggestions(query)
	if err != nil {
		http.Error(w, "Unable to fetch suggestions", http.StatusInternalServerError)
		return
	}

	// If a suggestion is selected, save it in cookies
	if len(suggestions) > 0 {
		selected := suggestions[0] // Assuming the first result is selected by default

		// Save cookies for city name, latitude, and longitude
		http.SetCookie(w, &http.Cookie{
			Name:  "cityName",
			Value: url.QueryEscape(selected.Name),
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "latitude",
			Value: fmt.Sprintf("%f", selected.Lat),
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "longitude",
			Value: fmt.Sprintf("%f", selected.Lon),
			Path:  "/",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(Favicon)
	if err != nil {
		log.Printf("ERROR: Could not write favicon: %v", err)
	}
}

func float64ToSting(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
