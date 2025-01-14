package main

import (
	"fmt"
	"math"
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
	Lat                   float64
	Lon                   float64
}

type DataPoints []DataPoint

// isGood() returns true if Low, Mid and High clouds percentage is less than maxCloudCover and wind is less than maxWind
func (d DataPoint) isGood(maxCloudCover int64, maxWind float64) bool {
	if d.HighClouds <= maxCloudCover && d.MidClouds <= maxCloudCover && d.LowClouds <= maxCloudCover && d.WindSpeed <= maxWind && d.WindGusts <= maxWind {
		return true
	} else {
		return false
	}
}

// setMoonIllumination() sets MoonIllum value for point in DataPoints
func (dp DataPoints) setMoonIllumination() DataPoints {
	updatedPoints := DataPoints{}

	for _, point := range dp {
		point.MoonIllum = int64(math.Round(moonIllumination(point.Time)))
		updatedPoints = append(updatedPoints, point)
	}

	return updatedPoints
}

// setSeeing() sets Seeing value for point in DataPoints
func (dp DataPoints) setSeeing() DataPoints {
	updatedPoints := DataPoints{}

	for _, point := range dp {
		// Threshold and Sensitivity.
		// Jet stream penalties are applied only if wind speeds exceed <jetStreamThreshold> m/s at 200 hPa.
		jetStreamThreshold := 15.0 // Jet stream speed threshold in m/s
		jetStreamFactor := 0.03    // Sensitivity factor

		// Calculate temperature gradient (°C/km)
		// ~3.5 km difference between 850 hPa (1.5 km) and 500 hPa (5 km)
		// ~1.5 km difference between surface and 850 hPa (1.5 km)
		// ~5 km gradient total

		// Calculate height difference dynamically
		heightDiff := point.GeopotentialHeight500 - point.GeopotentialHeight850

		// Gradient between surface and 850 hPa (1.5 km approx)
		lowGradient := (point.Temperature2M - point.Temperature850hPa) / (point.GeopotentialHeight850 / 1000.0)

		// Gradient between 850 hPa and 500 hPa (3.5 km approx)
		midGradient := (point.Temperature850hPa - point.Temperature500hPa) / (heightDiff / 1000.0)

		// Combine gradients
		tempGradient := lowGradient + midGradient

		// High wind shear (difference in wind speeds between altitudes) increases turbulence
		windShear := math.Abs(point.WindSpeed200hPa/3.6-point.WindSpeed/3.6) + math.Abs(point.WindSpeed/3.6-point.WindSpeed850hPa/3.6)

		// Approximate seeing using empirical formula: ε ∝ V^0.6 * T_grad^0.4
		// 0.12 coefficient set to be less optimistic
		point.Seeing = 0.12 * math.Pow(windShear, 0.6) * math.Pow(math.Abs(tempGradient), 0.4)

		// Jet stream impact (penalize if above threshold)
		if point.WindSpeed200hPa/3.6 > jetStreamThreshold {
			point.Seeing *= 1 + jetStreamFactor*(point.WindSpeed200hPa/3.6-jetStreamThreshold)
		}

		// Richardson Number (Ri) to estimate turbulence
		Ri := tempGradient / math.Pow(point.WindSpeed/3.6, 2)
		if Ri < 0.25 {
			point.Seeing *= 1.5
		} else if Ri < 0.5 {
			point.Seeing *= 1.2
		}

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
			moonRise, moonSet := calculateMoonRiseSet(point.Time, point.Lat, point.Lon)
			sunRise, sunSet := calculateSunRiseSet(point.Time, point.Lat, point.Lon)

			// Format for Moon
			moonRiseString := moonRise.Format("15:04")
			moonSetString := moonSet.Format("15:04")

			// Handle spacial cases when Moon is not rising or setting on that day
			if moonRise.Day() != point.Time.Day() {
				moonRiseString = "no"
			}

			if moonSet.Day() != point.Time.Day() {
				moonSetString = "no"
			}

			// Print out results
			out += fmt.Sprintf("%s - %s\n", date, dayOfWeek)
			out += fmt.Sprintf("Moon: %s - %s | Sun: %s - %s\n", moonRiseString, moonSetString, sunRise.Format("15:04"), sunSet.Format("15:04"))
			out += "-----------------------------------------------------------------------\n"
			out += " Hour | Ok? | Temp  | Moon  | Low | Mid  | High | Wind  | Gusts | Seeing \n"
			out += "-----|-----|-------|-------|-----|------|------|-------|-------|-------\n"
			currentDate = date
		}

		status := "-"
		if point.isGood(MaxCloudCover, MaxWindSpeed) {
			status = "ok"
		}

		out += fmt.Sprintf("%02d | %3s | %5.1f | %3d%%  | %3d | %3d  | %3d  | %5.1f | %5.1f | %4.1f\n",
			point.Time.Hour(), status, point.Temperature2M, point.MoonIllum, point.LowClouds, point.MidClouds, point.HighClouds, point.WindSpeed, point.WindGusts, point.Seeing)
	}

	return out
}
