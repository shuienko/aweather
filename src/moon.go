package main

import (
	"math"
	"time"
)

// MoonIllumination calculates the Moon's illumination percentage for a given time
func moonIllumination(date time.Time) float64 {
	const synodicMonth float64 = 29.53059      // Average length of a synodic month in days
	const newMoonReference float64 = 2451549.5 // Julian date for a known new moon (Jan 6, 2000 18:14 UTC)

	// Convert the date to Julian Day
	julianDate := dateToJulianDate(date)

	// Calculate days since known new moon
	daysSinceNewMoon := julianDate - newMoonReference

	// Normalize to the Moon phase cycle (0 to 1)
	moonPhase := math.Mod(daysSinceNewMoon/synodicMonth, 1.0)
	if moonPhase < 0 {
		moonPhase += 1.0
	}

	// Calculate illumination percentage
	illumination := (1.0 - math.Cos(2.0*math.Pi*moonPhase)) / 2.0 * 100.0

	return illumination
}

// dateToJulianDate converts a time.Time to Julian Date
func dateToJulianDate(date time.Time) float64 {
	year, month, day := date.UTC().Date()
	hour, min, sec := date.UTC().Clock()

	// If the month is January or February, adjust the year and month
	if month <= 2 {
		year--
		month += 12
	}

	// Calculate Julian Day Number (JDN)
	a := int(float64(year) / 100.0)
	b := 2 - a + int(float64(a)/4.0)
	jdn := int(365.25*float64(year)) + int(30.6001*float64(month+1)) + day + 1720994 + b

	// Add fractional day for the time of day
	fracDay := (float64(hour) + float64(min)/60.0 + float64(sec)/3600.0) / 24.0

	return float64(jdn) + fracDay
}
