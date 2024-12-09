package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	openMeteoGeoAPIEndpoint = "https://geocoding-api.open-meteo.com/v1/search"
)

type GeocodingResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
		Admin2    string  `json:"admin2"`
		Admin3    string  `json:"admin3"`
		Admin4    string  `json:"admin4"`
	} `json:"results"`
}

func getSuggestions(query string) ([]string, error) {
	apiURL := fmt.Sprintf("%s?name=%s", openMeteoGeoAPIEndpoint, url.QueryEscape(query))
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geoResponse GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
		return nil, err
	}

	var suggestions []string
	for _, result := range geoResponse.Results {
		suggestions = append(suggestions, fmt.Sprintf("%s, %s", result.Name, result.Country))
	}
	return suggestions, nil
}

func getCoordinates(location string) (float64, float64, error) {
	query := url.QueryEscape(location)
	apiURL := fmt.Sprintf("%s?name=%s", openMeteoGeoAPIEndpoint, query)

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling geocoding API: %w", err)
	}
	defer resp.Body.Close()

	var geoResponse GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
		return 0, 0, fmt.Errorf("error decoding geocoding API response: %w", err)
	}

	if len(geoResponse.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for location")
	}

	return geoResponse.Results[0].Latitude, geoResponse.Results[0].Longitude, nil
}
