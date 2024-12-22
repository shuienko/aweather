package main

import (
	"fmt"
	"math"
	"time"
)

type DataPoint struct {
	Time              time.Time
	Temperature2M     float64
	Temperature500hPa float64
	LowClouds         int64
	MidClouds         int64
	HighClouds        int64
	MoonIllum         int64
	WindSpeed         float64
	WindGusts         float64
	Seeing            float64
	WindSpeed200hPa   float64
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
		point.MoonIllum = int64(moonIllumination(point.Time))
		updatedPoints = append(updatedPoints, point)
	}

	return updatedPoints
}

// setSeeing() sets Seeing value for point in DataPoints
func (dp DataPoints) setSeeing() DataPoints {
	updatedPoints := DataPoints{}

	for _, point := range dp {
		// Calculate temperature gradient (°C/km)
		// 5 km difference between surface and 500 hPa
		tempGradient := (point.Temperature2M - point.Temperature500hPa) / 5.0

		// Approximate seeing using empirical formula
		// ε ∝ V^0.6 * T_grad^0.4
		point.Seeing = 0.1 * math.Pow(point.WindSpeed200hPa/3.6, 0.6) * math.Pow(math.Abs(tempGradient), 0.4)

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
			out += fmt.Sprintf("%s - %s\n", date, dayOfWeek)
			out += " Hour | Status | Temp  | Moon  | Low | Mid  | High | Wind  | Gusts | Seeing \n"
			out += "-----|--------|-------|-------|-----|------|------|-------|-------|-------\n"
			currentDate = date
		}

		status := "Bad"
		if point.isGood(MaxCloudCover, MaxWindSpeed) {
			status = "Good"
		}

		out += fmt.Sprintf("%02d | %6s | %5.1f | %3d%%  | %3d | %3d  | %3d  | %5.1f | %5.1f | %3.1f \n",
			point.Time.Hour(), status, point.Temperature2M, point.MoonIllum, point.LowClouds, point.MidClouds, point.HighClouds, point.WindSpeed, point.WindGusts, point.Seeing)
	}

	return out
}
