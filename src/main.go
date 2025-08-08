package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/allegro/bigcache/v3"
)

const (
	MaxCloudCover = 25               // percentage
	MaxWindSpeed  = 15               // km/h
	CacheTTL      = 10 * time.Minute // cache TTL
)

var (
	OpenMeteoAPIEndpoint           = "https://api.open-meteo.com/v1/forecast?"
	OpenMeteoGeoAPIEndpoint        = "https://geocoding-api.open-meteo.com/v1/search"
	OpenMeteoGeoReverseAPIEndpoint = "https://geocoding-api.open-meteo.com/v1/reverse"
	OpenMeteoAPIParams             = "temperature_2m,cloud_cover_low,cloud_cover_mid,cloud_cover_high,wind_speed_10m,wind_gusts_10m,wind_speed_200hPa,temperature_500hPa,temperature_850hPa,wind_speed_850hPa,geopotential_height_850hPa,geopotential_height_500hPa"
)

var cache *bigcache.BigCache

func main() {
	cache, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(CacheTTL))

	mux := http.NewServeMux()

	// Handle static files (favicon, icons, JS)
	staticRoot, err := fs.Sub(StaticFiles, "static")
	if err != nil {
		log.Fatalf("failed to set static sub FS: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticRoot))))

	// Define all routes
	mux.HandleFunc("/weather", handleWeather)
	mux.HandleFunc("/suggestions", handleSuggestions)
	mux.HandleFunc("/reverse-geocoding", handleReverseGeocoding)
	mux.HandleFunc("/robots.txt", handleRobots)
	mux.HandleFunc("/favicon.ico", handleFavicon)

	// Root index
	mux.HandleFunc("/", handleIndex)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
