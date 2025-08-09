package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
	Time                  []string  `json:"time"`
	Temperature2M         []float64 `json:"temperature_2m"`
	Temperature500hPa     []float64 `json:"temperature_500hPa"`
	Temperature850hPa     []float64 `json:"temperature_850hPa"`
	CloudCoverLow         []int64   `json:"cloud_cover_low"`
	CloudCoverMid         []int64   `json:"cloud_cover_mid"`
	CloudCoverHigh        []int64   `json:"cloud_cover_high"`
	WindSpeed10M          []float64 `json:"wind_speed_10m"`
	WindGusts10M          []float64 `json:"wind_gusts_10m"`
	WindSpeed200hPa       []float64 `json:"wind_speed_200hPa"`
	WindSpeed850hPa       []float64 `json:"wind_speed_850hPa"`
	GeopotentialHeight850 []float64 `json:"geopotential_height_850hPa"`
	GeopotentialHeight500 []float64 `json:"geopotential_height_500hPa"`
}

type HourlyUnits struct {
	Time                  string `json:"time"`
	Temperature2M         string `json:"temperature_2m"`
	Temperature500hPa     string `json:"temperature_500hPa"`
	Temperature850hPa     string `json:"temperature_850hPa"`
	CloudCoverLow         string `json:"cloud_cover_low"`
	CloudCoverMid         string `json:"cloud_cover_mid"`
	CloudCoverHigh        string `json:"cloud_cover_high"`
	WindSpeed10M          string `json:"wind_speed_10m"`
	WindGusts10M          string `json:"wind_gusts_10m"`
	WindSpeed200hPa       string `json:"wind_speed_200hPa"`
	WindSpeed850hPa       string `json:"wind_speed_850hPa"`
	GeopotentialHeight850 string `json:"geopotential_height_850hPa"`
	GeopotentialHeight500 string `json:"geopotential_height_500hPa"`
}

type Suggestion struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Admin1      string  `json:"admin1"`
	Admin2      string  `json:"admin2"`
	Admin3      string  `json:"admin3"`
	Admin4      string  `json:"admin4"`
	Lat         float64 `json:"latitude"`
	Lon         float64 `json:"longitude"`
}

// shared HTTP client with reasonable timeout
var httpClient = &http.Client{Timeout: 12 * time.Second}

// FetchData goes to OpenMeteoEndpoint, makes HTTPS request and stores result as OpenMeteoAPIResponse object
// Returns error when upstream is unavailable or response cannot be parsed.
func (response *OpenMeteoAPIResponse) FetchData(apiEndpoint, parameters, lat, lon string) error {
	cacheKey := fmt.Sprintf("weather:%s,%s:%s", lat, lon, parameters)
	weatherData, err := cache.Get(cacheKey)

	if err != nil {
		log.Println("INFO: Making request to Open-Meteo API and parsing response")

		// Set parameters
		params := url.Values{}
		params.Add("latitude", lat)
		params.Add("longitude", lon)
		params.Add("hourly", parameters)
		params.Add("timezone", "auto")

		// Make request to Open-Meteo API
		req, err := http.NewRequest("GET", apiEndpoint+params.Encode(), nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("do request: %w", err)
		}
		defer resp.Body.Close()

		// Read Response Body
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("upstream status: %s", resp.Status)
		}

		log.Println("INFO: Got API response", resp.Status)
		weatherData, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}

		// Save response to cache
		if err := cache.Set(cacheKey, weatherData); err != nil {
			log.Printf("WARN: cache set weather failed for %s: %v", cacheKey, err)
		}
	} else {
		log.Println("INFO: Using cached data for", cacheKey)
	}

	// Save response as OpenMeteoAPIResponse object
	err = json.Unmarshal(weatherData, response)
	if err != nil {
		return fmt.Errorf("unmarshal weather json: %w", err)
	}
	return nil
}

// fetchSuggestions() makes request to OpenMeteoGeoAPI and returns Suggestion object
func fetchSuggestions(query string) ([]Suggestion, error) {
	// Encode query
	encodedQuery := url.QueryEscape(query)

	// Check if query is in cache
	cacheKey := "geo:" + encodedQuery
	resultByte, err := cache.Get(cacheKey)
	if err != nil {
		// Fallback to legacy key used previously
		if legacy, legacyErr := cache.Get(encodedQuery); legacyErr == nil {
			resultByte = legacy
			err = nil
		}
	}

	// If not in cache, make request to OpenMeteoGeoAPI
	if err != nil {
		log.Println("INFO: Making request to Open-Meteo Geo API and parsing response for query: ", encodedQuery)
		requestURL := fmt.Sprintf("%s?name=%s", OpenMeteoGeoAPIEndpoint, encodedQuery)
		resp, err := httpClient.Get(requestURL)
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
		cache.Set(cacheKey, jsonData)
		// Also store under legacy key for backward compatibility (tests rely on it)
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

// fetchReverseGeocoding queries Open‑Meteo Reverse Geocoding API for a single best match
// It caches the first result under key "reverse:lat,lon" and returns it.
func fetchReverseGeocoding(lat string, lon string) (*Suggestion, error) {
	// Normalize to 3 decimal places (~111m) to avoid cache misses due to small GPS jitter
	// This significantly increases cache hit rate and reduces upstream calls.
	normLat, normLon := lat, lon
	if lf, err1 := strconv.ParseFloat(lat, 64); err1 == nil {
		normLat = strconv.FormatFloat(lf, 'f', 3, 64)
	}
	if lf, err2 := strconv.ParseFloat(lon, 64); err2 == nil {
		normLon = strconv.FormatFloat(lf, 'f', 3, 64)
	}

	cacheKey := fmt.Sprintf("reverse:%s,%s", normLat, normLon)

	if cached, err := cache.Get(cacheKey); err == nil {
		var suggestion Suggestion
		if err := json.Unmarshal(cached, &suggestion); err == nil {
			log.Println("INFO: Using cached reverse geocoding for", cacheKey)
			return &suggestion, nil
		} else {
			log.Println("WARN: Cached reverse geocoding unmarshal failed, refetching:", cacheKey, err)
		}
		// fallthrough to refetch on unmarshal error
	} else {
		log.Println("INFO: Reverse geocoding cache miss for", cacheKey, "reason:", err)
	}

	log.Println("INFO: Making request to Open-Meteo Reverse Geo API for:", normLat, normLon)
	requestURL := fmt.Sprintf("%s?latitude=%s&longitude=%s", OpenMeteoGeoReverseAPIEndpoint, url.QueryEscape(normLat), url.QueryEscape(normLon))
	resp, err := httpClient.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Upstream returned non-200; treat as no result to avoid surfacing errors to clients
		return nil, nil
	}

	var result struct {
		Results []Suggestion `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, nil
	}

	top := result.Results[0]
	if data, err := json.Marshal(top); err == nil {
		if err := cache.Set(cacheKey, data); err != nil {
			log.Println("WARN: Failed to cache reverse geocoding result for", cacheKey, "error:", err)
		} else {
			log.Println("INFO: Cached reverse geocoding for", cacheKey)
		}
	} else {
		log.Println("WARN: Failed to marshal reverse geocoding result for", cacheKey, "error:", err)
	}
	return &top, nil
}

// Points() return DataPoints object based on OpenMeteoAPIResponse fields
func (data OpenMeteoAPIResponse) Points() DataPoints {
	points := DataPoints{}

	h := data.Hourly
	// Determine safe length across all hourly arrays
	minLen := len(h.Time)
	candidates := []int{
		len(h.Temperature2M),
		len(h.Temperature500hPa),
		len(h.Temperature850hPa),
		len(h.CloudCoverLow),
		len(h.CloudCoverMid),
		len(h.CloudCoverHigh),
		len(h.WindSpeed10M),
		len(h.WindGusts10M),
		len(h.WindSpeed200hPa),
		len(h.WindSpeed850hPa),
		len(h.GeopotentialHeight850),
		len(h.GeopotentialHeight500),
	}
	for _, n := range candidates {
		if n < minLen {
			minLen = n
		}
	}

	if minLen == 0 {
		log.Println("WARN: Open-Meteo: no hourly data to build points")
		return points
	}

	// Log if arrays are mismatched; we will safely truncate to minLen
	if len(h.Time) != minLen ||
		len(h.Temperature2M) != minLen ||
		len(h.Temperature500hPa) != minLen ||
		len(h.Temperature850hPa) != minLen ||
		len(h.CloudCoverLow) != minLen ||
		len(h.CloudCoverMid) != minLen ||
		len(h.CloudCoverHigh) != minLen ||
		len(h.WindSpeed10M) != minLen ||
		len(h.WindGusts10M) != minLen ||
		len(h.WindSpeed200hPa) != minLen ||
		len(h.WindSpeed850hPa) != minLen ||
		len(h.GeopotentialHeight850) != minLen ||
		len(h.GeopotentialHeight500) != minLen {
		log.Printf("WARN: Open-Meteo hourly array length mismatch; truncating to %d (time=%d t2m=%d t500=%d t850=%d cl=%d cm=%d ch=%d w10=%d gust=%d w200=%d w850=%d gph850=%d gph500=%d)",
			minLen,
			len(h.Time), len(h.Temperature2M), len(h.Temperature500hPa), len(h.Temperature850hPa),
			len(h.CloudCoverLow), len(h.CloudCoverMid), len(h.CloudCoverHigh),
			len(h.WindSpeed10M), len(h.WindGusts10M), len(h.WindSpeed200hPa), len(h.WindSpeed850hPa),
			len(h.GeopotentialHeight850), len(h.GeopotentialHeight500))
	}

	// Resolve location once; fall back to UTC if unknown
	location, locErr := time.LoadLocation(data.Timezone)
	if locErr != nil || location == nil {
		location = time.UTC
	}

	for i := 0; i < minLen; i++ {
		parsedTime, err := time.ParseInLocation("2006-01-02T15:04", h.Time[i], location)
		if err != nil {
			// Fallback to UTC parsing to avoid zero time
			if t, err2 := time.Parse("2006-01-02T15:04", h.Time[i]); err2 == nil {
				parsedTime = t
			}
		}

		point := DataPoint{
			Time:                  parsedTime,
			Temperature2M:         h.Temperature2M[i],
			Temperature500hPa:     h.Temperature500hPa[i],
			Temperature850hPa:     h.Temperature850hPa[i],
			WindSpeed200hPa:       h.WindSpeed200hPa[i],
			WindSpeed850hPa:       h.WindSpeed850hPa[i],
			LowClouds:             h.CloudCoverLow[i],
			MidClouds:             h.CloudCoverMid[i],
			HighClouds:            h.CloudCoverHigh[i],
			WindSpeed:             h.WindSpeed10M[i],
			WindGusts:             h.WindGusts10M[i],
			GeopotentialHeight850: h.GeopotentialHeight850[i],
			GeopotentialHeight500: h.GeopotentialHeight500[i],
			Elevation:             data.Elevation,
			Lat:                   data.Latitude,
			Lon:                   data.Longitude,
		}

		points = append(points, point)
	}

	return points
}
