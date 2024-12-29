package main

import (
	"testing"
	"time"
)

// TestMoonIllumination verifies the correctness of moonIllumination function.
func TestMoonIllumination(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected float64
	}{
		{
			name:     "New Moon",
			input:    time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
			expected: 0.0,
		},
		{
			name:     "Full Moon",
			input:    time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			expected: 100.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := moonIllumination(tc.input)
			if result < tc.expected-1.0 || result > tc.expected+1.0 {
				t.Errorf("expected %.1f, got %.1f", tc.expected, result)
			}
		})
	}
}

// TestCalculateMoonRiseSet verifies the correctness of calculateMoonRiseSet function.
func TestCalculateMoonRiseSet(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		lat, lon float64
	}{
		{
			name:  "Moonrise and Moonset",
			input: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
			lat:   37.7749,
			lon:   -122.4194,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			moonrise, moonset := calculateMoonRiseSet(tc.input, tc.lat, tc.lon)

			if moonrise.IsZero() || moonset.IsZero() {
				t.Errorf("expected valid moonrise and moonset, got moonrise: %v, moonset: %v", moonrise, moonset)
			}
		})
	}
}

// TestCalculateSunRiseSet verifies the correctness of calculateSunRiseSet function.
func TestCalculateSunRiseSet(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		lat, lon float64
	}{
		{
			name:  "Sunrise and Sunset",
			input: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
			lat:   37.7749,
			lon:   -122.4194,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sunrise, sunset := calculateSunRiseSet(tc.input, tc.lat, tc.lon)

			if sunrise.IsZero() || sunset.IsZero() {
				t.Errorf("expected valid sunrise and sunset, got sunrise: %v, sunset: %v", sunrise, sunset)
			}
		})
	}
}
