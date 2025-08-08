package main

import (
	"time"

	"github.com/hablullah/go-sampa"
	"github.com/soniakeys/meeus/v3/base"
	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/moonillum"
)

// moonIllumination calculates the Moon's illumination percentage for a given time
func moonIllumination(t time.Time) float64 {

	// Convert the date to Julian Day
	julianDate := julian.TimeToJD(t.UTC())

	// Calculate illumination
	angle := moonillum.PhaseAngle3(julianDate)
	illumination := base.Illuminated(angle) * 100

	return illumination
}

// calculateRiseSet returns rise and set local time for Sun or Moon
// body must be "sun" or "moon"
func calculateRiseSet(t time.Time, lat, lon float64, body string) (time.Time, time.Time) {
	city := makeLocation(lat, lon)
	if body == "moon" {
		moonEvents, _ := sampa.GetMoonEvents(t, city, nil)
		return moonEvents.Moonrise.DateTime, moonEvents.Moonset.DateTime
	}
	sunEvents, _ := sampa.GetSunEvents(t, city, nil)
	return sunEvents.Sunrise.DateTime, sunEvents.Sunset.DateTime
}

// makeLocation builds a sampa.Location from coordinates
func makeLocation(lat, lon float64) sampa.Location {
	return sampa.Location{Latitude: lat, Longitude: lon}
}

// Backwards-compatible wrappers used by tests and callers
func calculateMoonRiseSet(t time.Time, lat, lon float64) (time.Time, time.Time) {
	return calculateRiseSet(t, lat, lon, "moon")
}

func calculateSunRiseSet(t time.Time, lat, lon float64) (time.Time, time.Time) {
	return calculateRiseSet(t, lat, lon, "sun")
}
