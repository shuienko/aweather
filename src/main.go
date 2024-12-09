package main

import (
	"log"
	"net/http"
)

const (
	OpenMeteoAPIEndpoint    = "https://api.open-meteo.com/v1/forecast?"
	OpenMeteoGeoAPIEndpoint = "https://geocoding-api.open-meteo.com/v1/search"
	OpenMeteoAPIParams      = "temperature_2m,cloud_cover_low,cloud_cover_mid,cloud_cover_high,wind_speed_10m,wind_gusts_10m,wind_speed_200hPa,temperature_500hPa"
	MaxCloudCover           = 25 // percentage
	MaxWindSpeed            = 15 // km/h
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/weather", handleWeather)
	http.HandleFunc("/suggestions", handleSuggestions)

	log.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
