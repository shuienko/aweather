package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type OpenMeteoAPIResponse struct {
	Latitude             float64     `json:"latitude"`
	Longitude            float64     `json:"longitude"`
	GenerationtimeMS     float64     `json:"generationtime_ms"`
	UTCOffsetSeconds     int64       `json:"utc_offset_seconds"`
	Timezone             string      `json:"timezone"`
	TimezoneAbbreviation string      `json:"timezone_abbreviation"`
	Elevation            float64     `json:"elevation"`
	HourlyUnits          HourlyUnits `json:"hourly_units"`
	Hourly               Hourly      `json:"hourly"`
}

type Hourly struct {
	Time              []string  `json:"time"`
	Temperature2M     []float64 `json:"temperature_2m"`
	Temperature500hPa []float64 `json:"temperature_500hPa"`
	CloudCoverLow     []int64   `json:"cloud_cover_low"`
	CloudCoverMid     []int64   `json:"cloud_cover_mid"`
	CloudCoverHigh    []int64   `json:"cloud_cover_high"`
	WindSpeed10M      []float64 `json:"wind_speed_10m"`
	WindGusts10M      []float64 `json:"wind_gusts_10m"`
	WindSpeed200hPa   []float64 `json:"wind_speed_200hPa"`
}

type HourlyUnits struct {
	Time              string `json:"time"`
	Temperature2M     string `json:"temperature_2m"`
	Temperature500hPa string `json:"temperature_500hPa"`
	CloudCoverLow     string `json:"cloud_cover_low"`
	CloudCoverMid     string `json:"cloud_cover_mid"`
	CloudCoverHigh    string `json:"cloud_cover_high"`
	WindSpeed10M      string `json:"wind_speed_10m"`
	WindGusts10M      string `json:"wind_gusts_10m"`
	WindSpeed200hPa   string `json:"wind_speed_200hPa"`
}

type Suggestion struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Admin1  string  `json:"admin1"`
	Admin2  string  `json:"admin2"`
	Admin3  string  `json:"admin3"`
	Admin4  string  `json:"admin4"`
	Lat     float64 `json:"latitude"`
	Lon     float64 `json:"longitude"`
}

// FetchData() goes to OpenMeteoEndpoint makes HTTPS request and stores result as OpenMeteoAPIResponse object
func (response *OpenMeteoAPIResponse) FetchData(apiEndpoint, parameters, lat, lon string) {
	latlon := lat + lon
	weatherData, err := cache.Get(latlon)

	if err != nil {
		log.Println("INFO: Making request to Open-Meteo API and parsing response")
		client := &http.Client{}

		// Set paramenters
		params := url.Values{}
		params.Add("latitude", lat)
		params.Add("longitude", lon)
		params.Add("hourly", parameters)
		params.Add("timezone", "auto")

		// Make request to Open-Meteo API
		req, err := http.NewRequest("GET", apiEndpoint+params.Encode(), nil)
		if err != nil {
			log.Println("ERROR: Couldn't create new Open-Meteo API request", err)
			return
		}

		parseFormErr := req.ParseForm()
		if parseFormErr != nil {
			log.Println(parseFormErr)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println("ERROR:", err)
			return
		}

		// Read Response Body
		if resp.StatusCode != 200 {
			fmt.Println("ERROR: Open-Meteo API response code:", resp.Status)
			return
		}

		log.Println("INFO: Got API response", resp.Status)
		weatherData, _ = io.ReadAll(resp.Body)

		// Save response to cache
		cache.Set(latlon, weatherData)
	} else {
		log.Println("INFO: Using cached data for latlon", latlon)
	}

	// Save response as OpenMeteoAPIResponse object
	err = json.Unmarshal(weatherData, response)
	if err != nil {
		log.Println("ERROR: cannot Unmarshal JSON", err)
		return
	}
}

// fetchSuggestions() makes request to OpenMeteoGeoAPI and returns Suggestion object
func fetchSuggestions(query string) ([]Suggestion, error) {
	// Encode query
	encodedQuery := url.QueryEscape(query)

	// Check if query is in cache
	resultByte, err := cache.Get(encodedQuery)

	// If not in cache, make request to OpenMeteoGeoAPI
	if err != nil {
		log.Println("INFO: Making request to Open-Meteo Geo API and parsing response for query: ", encodedQuery)
		url := fmt.Sprintf("%s?name=%s", OpenMeteoGeoAPIEndpoint, encodedQuery)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Results []Suggestion `json:"results"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		// Save response to cache
		jsonData, _ := json.Marshal(result.Results)
		cache.Set(encodedQuery, jsonData)

		// Return results
		return result.Results, nil
	} else {
		log.Println("INFO: Using cached data for query", encodedQuery)

		// Unmarshal cached data
		result := []Suggestion{}
		json.Unmarshal(resultByte, &result)

		// Return results
		return result, nil
	}
}

// Points() return DataPoints object based on OpenMeteoAPIResponse fields
func (data OpenMeteoAPIResponse) Points() DataPoints {
	points := DataPoints{}

	for i := 0; i < len(data.Hourly.Time); i++ {
		location, _ := time.LoadLocation(data.Timezone)
		time, _ := time.ParseInLocation("2006-01-02T15:04", data.Hourly.Time[i], location)

		if i == 0 {
			log.Printf("DEBUG: %v | %v | %v\n", data.Hourly.Time[i], time, location)
		}

		point := DataPoint{
			Time:              time,
			Temperature2M:     data.Hourly.Temperature2M[i],
			Temperature500hPa: data.Hourly.Temperature500hPa[i],
			WindSpeed200hPa:   data.Hourly.WindSpeed200hPa[i],
			LowClouds:         data.Hourly.CloudCoverLow[i],
			MidClouds:         data.Hourly.CloudCoverMid[i],
			HighClouds:        data.Hourly.CloudCoverHigh[i],
			WindSpeed:         data.Hourly.WindSpeed10M[i],
			WindGusts:         data.Hourly.WindGusts10M[i],
		}

		points = append(points, point)
	}

	return points
}
