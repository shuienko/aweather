package main

import (
	"strings"
	"testing"
	"time"
)

func TestIsGood(t *testing.T) {
	pointGood := DataPoint{
		HighClouds: 10,
		MidClouds:  15,
		LowClouds:  20,
		WindSpeed:  10,
		WindGusts:  12,
	}

	pointBadClouds := DataPoint{
		HighClouds: 10,
		MidClouds:  35,
		LowClouds:  20,
		WindSpeed:  4,
		WindGusts:  12,
	}

	pointBadWind := DataPoint{
		HighClouds: 10,
		MidClouds:  15,
		LowClouds:  20,
		WindSpeed:  18,
		WindGusts:  24,
	}

	pointWeird := DataPoint{
		HighClouds: -2,
		MidClouds:  0,
		LowClouds:  200,
		WindSpeed:  10,
		WindGusts:  3,
	}

	if !pointGood.isGood(MaxCloudCover, MaxWindSpeed) {
		t.Errorf("Expected point to be 'good', but got 'bad'")
	}

	if pointBadClouds.isGood(MaxCloudCover, MaxWindSpeed) {
		t.Errorf("Expected point to be 'bad', but got 'good'")
	}

	if pointBadWind.isGood(MaxCloudCover, MaxWindSpeed) {
		t.Errorf("Expected point to be 'bad', but got 'good'")
	}

	if pointWeird.isGood(MaxCloudCover, MaxWindSpeed) {
		t.Errorf("Expected point to be 'bad', but got 'good'")
	}
}

func TestSetMoonIllumination(t *testing.T) {
	points := DataPoints{
		{
			Time: time.Now(),
		},
	}

	updatedPoints := points.setMoonIllumination()
	if updatedPoints[0].MoonIllum == 0 {
		t.Errorf("Expected MoonIllum to be set, but got 0")
	}
}

func TestSetSeeing(t *testing.T) {
	points := DataPoints{
		{
			Temperature2M:         15,
			Temperature500hPa:     -10,
			Temperature850hPa:     5,
			WindSpeed200hPa:       50,
			WindSpeed850hPa:       20,
			WindSpeed:             15,
			GeopotentialHeight850: 1500,
			GeopotentialHeight500: 5000,
		},
	}

	updatedPoints := points.setSeeing()
	if updatedPoints[0].Seeing == 0 {
		t.Errorf("Expected Seeing to be set, but got 0")
	}
}

func TestPrint(t *testing.T) {
	points := DataPoints{
		{
			Time:          time.Now(),
			Temperature2M: 10,
			MoonIllum:     50,
			LowClouds:     10,
			MidClouds:     15,
			HighClouds:    20,
			WindSpeed:     10,
			WindGusts:     12,
			Seeing:        1.5,
			Lat:           51.5,
			Lon:           -0.12,
		},
	}

	output := points.Print()
	if output == "" {
		t.Errorf("Expected output to contain forecast data, but got empty string")
	}
}

func TestPrintWithOptions_UnitsAnd12Hour(t *testing.T) {
	points := DataPoints{
		{
			Time:                  time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			Temperature2M:         10,
			Temperature500hPa:     -20,
			WindSpeed:             18, // km/h
			WindGusts:             36, // km/h
			LowClouds:             5,
			MidClouds:             5,
			HighClouds:            5,
			GeopotentialHeight500: 5500,
			Elevation:             100,
			Lat:                   40,
			Lon:                   -120,
		},
	}

	out := points.setMoonIllumination().setSeeing().PrintWithOptions(PrintOptions{
		TemperatureUnit: "f",
		WindSpeedUnit:   "mph",
		Use12Hour:       true,
	})
	if !strings.Contains(out, "1pm") {
		t.Fatalf("Expected 12h time formatting with '1pm', got: %s", out)
	}
	if !strings.Contains(out, "50.0") {
		t.Fatalf("Expected temperature 50.0F for 10C, got: %s", out)
	}
	if strings.Contains(out, "46.0") {
		t.Fatalf("Did not expect 46.0F in output, got: %s", out)
	}
}

func TestPoints_TruncatesMismatchedArrays(t *testing.T) {
	data := OpenMeteoAPIResponse{
		Latitude:  1,
		Longitude: 2,
		Hourly: Hourly{
			Time:                  []string{"2024-01-01T00:00", "2024-01-01T01:00"},
			Temperature2M:         []float64{1}, // shorter
			Temperature500hPa:     []float64{0},
			Temperature850hPa:     []float64{0},
			CloudCoverLow:         []int64{0},
			CloudCoverMid:         []int64{0},
			CloudCoverHigh:        []int64{0},
			WindSpeed10M:          []float64{0},
			WindGusts10M:          []float64{0},
			WindSpeed200hPa:       []float64{0},
			WindSpeed850hPa:       []float64{0},
			GeopotentialHeight850: []float64{0},
			GeopotentialHeight500: []float64{0},
		},
		Timezone: "UTC",
	}
	pts := data.Points()
	if len(pts) != 1 {
		t.Fatalf("Expected 1 point after truncation, got %d", len(pts))
	}
}

func TestPrintWithOptions_DefaultsNormalization(t *testing.T) {
	points := DataPoints{{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}}
	out := points.PrintWithOptions(PrintOptions{TemperatureUnit: "", WindSpeedUnit: "", Use12Hour: false})
	if !strings.Contains(out, "hour") || !strings.Contains(out, "wind") {
		t.Fatalf("expected default headers present, got: %s", out)
	}
}
