package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type DataPoint struct {
	Time                  time.Time
	Temperature2M         float64
	Temperature500hPa     float64
	Temperature850hPa     float64
	LowClouds             int64
	MidClouds             int64
	HighClouds            int64
	MoonIllum             int64
	WindSpeed             float64
	WindGusts             float64
	Seeing                float64
	WindSpeed200hPa       float64
	WindSpeed850hPa       float64
	GeopotentialHeight850 float64
	GeopotentialHeight500 float64
	Elevation             float64
	Lat                   float64
	Lon                   float64
}

type DataPoints []DataPoint

// PrintOptions controls display units and time formatting
type PrintOptions struct {
	TemperatureUnit string // "c" or "f"
	WindSpeedUnit   string // "kmh" or "mph"
	Use12Hour       bool
}

// Shared column widths for printing header and rows
const (
	colWidthHour   = 4
	colWidthOK     = 3
	colWidthTemp   = 5
	colWidthMoon   = 4
	colWidthLow    = 3
	colWidthMid    = 3
	colWidthHigh   = 4
	colWidthWind   = 5
	colWidthGusts  = 5
	colWidthSeeing = 6
)

// isGood() returns true if Low, Mid and High clouds percentage is less than maxCloudCover and wind is less than maxWind
func (d DataPoint) isGood(maxCloudCover int64, maxWind float64) bool {
	return d.HighClouds <= maxCloudCover &&
		d.MidClouds <= maxCloudCover &&
		d.LowClouds <= maxCloudCover &&
		d.WindSpeed <= maxWind &&
		d.WindGusts <= maxWind
}

// setMoonIllumination() sets MoonIllum value for point in DataPoints
func (dp DataPoints) setMoonIllumination() DataPoints {
	updatedPoints := make(DataPoints, 0, len(dp))

	for _, point := range dp {
		point.MoonIllum = int64(math.Round(moonIllumination(point.Time)))
		updatedPoints = append(updatedPoints, point)
	}

	return updatedPoints
}

// setSeeing() sets Seeing value for point in DataPoints
func (dp DataPoints) setSeeing() DataPoints {
	updatedPoints := make(DataPoints, 0, len(dp))

	for _, point := range dp {
		// Jet stream penalty configuration (using 200 hPa winds)
		jetStreamThreshold := 22.0 // m/s
		jetStreamFactor := 0.02    // per m/s above threshold
		maxJetMultiplier := 1.5    // cap the penalty

		// Elevation-aware single-layer temperature lapse (°C/km)
		// Use total depth from site elevation to 500 hPa geopotential height
		depthMeters := point.GeopotentialHeight500 - point.Elevation
		if depthMeters < 100.0 {
			depthMeters = 100.0 // avoid divide-by-zero and negative depths
		}
		depthKm := depthMeters / 1000.0
		tempLapse := (point.Temperature2M - point.Temperature500hPa) / depthKm

		// Wind shear proxy (m/s): combine vertical shear across 10m–850hPa and 850hPa–200hPa
		v10 := point.WindSpeed / 3.6
		v850 := point.WindSpeed850hPa / 3.6
		v200 := point.WindSpeed200hPa / 3.6
		windShear := math.Abs(v200-v850) + math.Abs(v850-v10)

		// Heuristic seeing index (dimensionless): lower is better
		// Keep scaling similar to previous behavior without arcsec claim
		base := 0.12 * math.Pow(windShear, 0.6) * math.Pow(math.Abs(tempLapse), 0.4)

		// Jet stream penalty above threshold, capped
		if v200 > jetStreamThreshold {
			penalty := 1.0 + jetStreamFactor*(v200-jetStreamThreshold)
			if penalty > maxJetMultiplier {
				penalty = maxJetMultiplier
			}
			base *= penalty
		}

		// Clamp to a reasonable range
		if base < 0.5 {
			base = 0.5
		}
		if base > 5.0 {
			base = 5.0
		}

		point.Seeing = base

		updatedPoints = append(updatedPoints, point)
	}

	return updatedPoints
}

// Print() returns Markdown string which represents DataPoints
func (dp DataPoints) Print() string {
	out := ""
	currentDate := ""

	for _, point := range dp {
		date := point.Time.Format("January 2")
		dayOfWeek := point.Time.Format("Monday")

		if date != currentDate {
			if currentDate != "" {
				out += "\n"
			}
			// Get Moon and Sun rise and set time
			moonRise, moonSet := calculateRiseSet(point.Time, point.Lat, point.Lon, "moon")
			sunRise, sunSet := calculateRiseSet(point.Time, point.Lat, point.Lon, "sun")

			// Format for Moon
			moonRiseString := moonRise.Format("15:04")
			moonSetString := moonSet.Format("15:04")

			// Handle special cases when Moon is not rising or setting on that day
			if moonRise.Day() != point.Time.Day() {
				moonRiseString = moonRiseString + "*"
			}

			if moonSet.Day() != point.Time.Day() {
				moonSetString = moonSetString + "*"
			}

			// Header
			header := fmt.Sprintf("%*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s\n",
				colWidthHour, "hour",
				colWidthOK, "ok?",
				colWidthTemp, "temp",
				colWidthMoon, "moon",
				colWidthLow, "low",
				colWidthMid, "mid",
				colWidthHigh, "high",
				colWidthWind, "wind",
				colWidthGusts, "gusts",
				colWidthSeeing, "seeing")

			// Separator matching column widths
			sep := strings.Join([]string{
				strings.Repeat("-", colWidthHour),
				strings.Repeat("-", colWidthOK),
				strings.Repeat("-", colWidthTemp),
				strings.Repeat("-", colWidthMoon),
				strings.Repeat("-", colWidthLow),
				strings.Repeat("-", colWidthMid),
				strings.Repeat("-", colWidthHigh),
				strings.Repeat("-", colWidthWind),
				strings.Repeat("-", colWidthGusts),
				strings.Repeat("-", colWidthSeeing),
			}, "-|-") + "\n"

			// Print out results
			out += fmt.Sprintf("%s - %s\n", date, dayOfWeek)
			out += fmt.Sprintf("moon: %s - %s | sun: %s - %s\n", moonRiseString, moonSetString, sunRise.Format("15:04"), sunSet.Format("15:04"))
			out += strings.Repeat("-", len(strings.TrimRight(header, "\n"))) + "\n"
			out += header
			out += sep
			currentDate = date
		}

		status := "-"
		if point.isGood(MaxCloudCover, MaxWindSpeed) {
			status = "ok"
		}

		hourStr := fmt.Sprintf("%02d", point.Time.Hour())
		okStr := status
		tempStr := fmt.Sprintf("%.1f", point.Temperature2M)
		moonStr := fmt.Sprintf("%d%%", point.MoonIllum)
		lowStr := fmt.Sprintf("%d", point.LowClouds)
		midStr := fmt.Sprintf("%d", point.MidClouds)
		highStr := fmt.Sprintf("%d", point.HighClouds)
		windStr := fmt.Sprintf("%.1f", point.WindSpeed)
		gustsStr := fmt.Sprintf("%.1f", point.WindGusts)
		seeingStr := fmt.Sprintf("%.1f", point.Seeing)

		out += fmt.Sprintf("%*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s\n",
			colWidthHour, hourStr,
			colWidthOK, okStr,
			colWidthTemp, tempStr,
			colWidthMoon, moonStr,
			colWidthLow, lowStr,
			colWidthMid, midStr,
			colWidthHigh, highStr,
			colWidthWind, windStr,
			colWidthGusts, gustsStr,
			colWidthSeeing, seeingStr)
	}

	return out
}

// PrintWithOptions returns Markdown-like string using provided formatting options
func (dp DataPoints) PrintWithOptions(opts PrintOptions) string {
	// normalize options
	tempUnit := strings.ToLower(strings.TrimSpace(opts.TemperatureUnit))
	if tempUnit != "f" {
		tempUnit = "c"
	}
	windUnit := strings.ToLower(strings.TrimSpace(opts.WindSpeedUnit))
	if windUnit != "mph" {
		windUnit = "kmh"
	}

	out := ""
	currentDate := ""

	for _, point := range dp {
		date := point.Time.Format("January 2")
		dayOfWeek := point.Time.Format("Monday")

		if date != currentDate {
			if currentDate != "" {
				out += "\n"
			}
			// Get Moon and Sun rise and set time
			moonRise, moonSet := calculateRiseSet(point.Time, point.Lat, point.Lon, "moon")
			sunRise, sunSet := calculateRiseSet(point.Time, point.Lat, point.Lon, "sun")

			// Format time strings depending on 12/24h preference
			timeFmt := "15:04"
			if opts.Use12Hour {
				timeFmt = "3:04pm"
			}

			// Format for Moon
			moonRiseString := moonRise.Format(timeFmt)
			moonSetString := moonSet.Format(timeFmt)

			// Handle special cases when Moon is not rising or setting on that day
			if moonRise.Day() != point.Time.Day() {
				moonRiseString = moonRiseString + "*"
			}

			if moonSet.Day() != point.Time.Day() {
				moonSetString = moonSetString + "*"
			}

			// Header
			header := fmt.Sprintf("%*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s\n",
				colWidthHour, "hour",
				colWidthOK, "ok?",
				colWidthTemp, "temp",
				colWidthMoon, "moon",
				colWidthLow, "low",
				colWidthMid, "mid",
				colWidthHigh, "high",
				colWidthWind, "wind",
				colWidthGusts, "gusts",
				colWidthSeeing, "seeing")

			// Separator matching column widths
			sep := strings.Join([]string{
				strings.Repeat("-", colWidthHour),
				strings.Repeat("-", colWidthOK),
				strings.Repeat("-", colWidthTemp),
				strings.Repeat("-", colWidthMoon),
				strings.Repeat("-", colWidthLow),
				strings.Repeat("-", colWidthMid),
				strings.Repeat("-", colWidthHigh),
				strings.Repeat("-", colWidthWind),
				strings.Repeat("-", colWidthGusts),
				strings.Repeat("-", colWidthSeeing),
			}, "-|-") + "\n"

			// Print out results
			out += fmt.Sprintf("%s - %s\n", date, dayOfWeek)
			out += fmt.Sprintf("moon: %s - %s | sun: %s - %s\n", moonRiseString, moonSetString, sunRise.Format(timeFmt), sunSet.Format(timeFmt))
			out += strings.Repeat("-", len(strings.TrimRight(header, "\n"))) + "\n"
			out += header
			out += sep
			currentDate = date
		}

		status := "-"
		if point.isGood(MaxCloudCover, MaxWindSpeed) {
			status = "ok"
		}

		// hour string
		hourStr := ""
		if opts.Use12Hour {
			hour := point.Time.Hour()
			ampm := "am"
			if hour >= 12 {
				ampm = "pm"
			}
			h12 := hour % 12
			if h12 == 0 {
				h12 = 12
			}
			hourStr = fmt.Sprintf("%d%s", h12, ampm)
		} else {
			hourStr = fmt.Sprintf("%02d", point.Time.Hour())
		}

		// temperature conversion (display only)
		temp := point.Temperature2M
		if tempUnit == "f" {
			temp = temp*9.0/5.0 + 32.0
		}

		// wind conversion (display only)
		wind := point.WindSpeed
		gusts := point.WindGusts
		if windUnit == "mph" {
			wind = wind / 1.609344
			gusts = gusts / 1.609344
		}

		okStr := status
		tempStr := fmt.Sprintf("%.1f", temp)
		moonStr := fmt.Sprintf("%d%%", point.MoonIllum)
		lowStr := fmt.Sprintf("%d", point.LowClouds)
		midStr := fmt.Sprintf("%d", point.MidClouds)
		highStr := fmt.Sprintf("%d", point.HighClouds)
		windStr := fmt.Sprintf("%.1f", wind)
		gustsStr := fmt.Sprintf("%.1f", gusts)
		seeingStr := fmt.Sprintf("%.1f", point.Seeing)

		out += fmt.Sprintf("%*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s | %*s\n",
			colWidthHour, hourStr,
			colWidthOK, okStr,
			colWidthTemp, tempStr,
			colWidthMoon, moonStr,
			colWidthLow, lowStr,
			colWidthMid, midStr,
			colWidthHigh, highStr,
			colWidthWind, windStr,
			colWidthGusts, gustsStr,
			colWidthSeeing, seeingStr)
	}

	return out
}
