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
	// Initialize cache with bounded size
	cacheConfig := bigcache.DefaultConfig(CacheTTL)
	cacheConfig.MaxEntrySize = 4096   // bytes, avoid oversized entries
	cacheConfig.HardMaxCacheSize = 32 // MB, keeps memory bounded on Cloud Run
	c, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		log.Fatalf("failed to init cache: %v", err)
	}
	cache = c

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
	mux.HandleFunc("/sitemap.xml", handleSitemap)
	mux.HandleFunc("/favicon.ico", handleFavicon)

	// Root index
	mux.HandleFunc("/", handleIndex)

	// Wrap with canonical host/scheme redirector (skip in local dev)
	canonicalHost := "aweather.shnk.net"
	canonicalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow plain HTTP and any host in local dev
		if r.Host == "localhost:8080" || r.Host == "127.0.0.1:8080" {
			mux.ServeHTTP(w, r)
			return
		}

		scheme := "http"
		if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
			scheme = "https"
		}
		if r.Host != canonicalHost || scheme != "https" {
			target := "https://" + canonicalHost + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// Harden server with reasonable timeouts
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           canonicalHandler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}
