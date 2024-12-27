package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/allegro/bigcache/v3"
)

const (
	OpenMeteoAPIEndpoint    = "https://api.open-meteo.com/v1/forecast?"
	OpenMeteoGeoAPIEndpoint = "https://geocoding-api.open-meteo.com/v1/search"
	OpenMeteoAPIParams      = "temperature_2m,cloud_cover_low,cloud_cover_mid,cloud_cover_high,wind_speed_10m,wind_gusts_10m,wind_speed_200hPa,temperature_500hPa,temperature_850hPa,wind_speed_850hPa,geopotential_height_850hPa,geopotential_height_500hPa"
	MaxCloudCover           = 25               // percentage
	MaxWindSpeed            = 15               // km/h
	CacheTTL                = 10 * time.Minute // cache TTL
)

var cache *bigcache.BigCache

func main() {
	cache, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(CacheTTL))

	mux := http.NewServeMux()

	// Handle static files (favicon and icons)
	mux.Handle("/static/", http.FileServer(http.FS(StaticFiles)))

	// Define all routes
	mux.HandleFunc("/weather", handleWeather)
	mux.HandleFunc("/suggestions", handleSuggestions)
	mux.HandleFunc("/robots.txt", handleRobots)

	// Use a custom NotFound handler for unknown routes and handleIndex
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/weather" && r.URL.Path != "/suggestions" && r.URL.Path != "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		handleIndex(w, r)
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
