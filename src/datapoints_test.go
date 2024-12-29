package main

import (
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
