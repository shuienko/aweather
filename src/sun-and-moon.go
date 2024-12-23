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

	// Claculate illumination
	angle := moonillum.PhaseAngle3(julianDate)
	illumination := base.Illuminated(angle) * 100

	return illumination
}

// calculateMoonRiseSet returns Moon rise and set local time
func calculateMoonRiseSet(t time.Time, lat, lon float64) (time.Time, time.Time) {
	city := sampa.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	moonEvents, _ := sampa.GetMoonEvents(t, city, nil)
	return moonEvents.Moonrise.DateTime, moonEvents.Moonset.DateTime
}

// calculateSunRiseSet returns Sun rise and set local time
func calculateSunRiseSet(t time.Time, lat, lon float64) (time.Time, time.Time) {
	city := sampa.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	moonEvents, _ := sampa.GetSunEvents(t, city, nil)
	return moonEvents.Sunrise.DateTime, moonEvents.Sunset.DateTime
}
