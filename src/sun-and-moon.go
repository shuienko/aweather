package main

import (
	"time"

	"github.com/soniakeys/meeus/v3/base"
	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/moonillum"
)

// MoonIllumination calculates the Moon's illumination percentage for a given time
func moonIllumination(date time.Time) float64 {

	// Convert the date to Julian Day
	julianDate := julian.TimeToJD(date.UTC())

	// Claculate illumination
	angle := moonillum.PhaseAngle3(julianDate)
	illumination := base.Illuminated(angle) * 100

	return illumination
}
