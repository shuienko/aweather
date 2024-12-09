package main

import (
	"fmt"
	"net/http"
)

const (
	OpenMeteoAPIEndpoint = "https://api.open-meteo.com/v1/forecast?"
	OpenMeteoAPIParams   = "temperature_2m,cloud_cover_low,cloud_cover_mid,cloud_cover_high,wind_speed_10m,wind_gusts_10m,wind_speed_200hPa,temperature_500hPa"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := OpenMeteoAPIResponse{}
		data.FetchData(OpenMeteoAPIEndpoint, OpenMeteoAPIParams, "36.0", "5.0")
		table := data.Points().setMoonIllumination().setSeeing().Print()
		fmt.Fprintf(w, "<html><body><pre>%s</pre></body></html>", table)
	})

	http.ListenAndServe(":8080", nil)
}
