package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Weather Forecast</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.16.22/css/uikit.min.css" />
    <script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.16.22/js/uikit.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/uikit/3.16.22/js/uikit-icons.min.js"></script>
    <style>
        .forecast-header {
            text-align: center;
        }
        .forecast-details {
            text-align: center;
            margin-top: 20px;
            font-size: 1.2rem;
        }
		.forecast-table-container {
    		overflow-x: auto;
    		white-space: nowrap;
			text-align: center;
		}
		.forecast-legend-container {
    		overflow-x: auto;
    		white-space: nowrap;
			text-align: left;
		}
        .forecast-legend {
			display: inline-block;
            border: none;
			text-align: left;
            box-shadow: none;
        }
        .forecast-about-container {
            text-align: left;
            word-wrap: break-word;
        }  
        .forecast-about {
            border: none;
            text-align: left;
            box-shadow: none;
            white-space: normal;
        }
        .forecast-table {
			display: inline-block;
            border: none;
			text-align: center;
            box-shadow: none;
        }
        .footer {
            margin-top: 50px;
            text-align: center;
            font-size: 0.9rem;
            color: #555;
        }
    </style>
</head>
<body>
    <div class="uk-container uk-margin-top">
        <h1 class="uk-heading-divider forecast-header">Weather for Astrophotographers</h1>
        
        <div class="uk-grid-small uk-flex-middle" uk-grid>
            <div class="uk-width-expand">
                <input class="uk-input" id="city" type="text" placeholder="e.g., London" value="%s" oninput="fetchSuggestions(this.value)">
                <ul id="suggestions" class="uk-list uk-list-divider" style="position: absolute; z-index: 1000; background: white; display: none;"></ul>
            </div>
            <div>
                <button class="uk-button uk-button-primary" onclick="fetchWeather()">Get Forecast</button>
            </div>
        </div>

        <!-- Hidden fields to store latitude and longitude -->
        <input type="hidden" id="latitude" value="%s">
        <input type="hidden" id="longitude" value="%s">

        <div id="forecastDetails" class="forecast-details" style="display: none;"></div>

		<div class="forecast-table-container">
	        <pre id="weatherResult" class="uk-margin-top forecast-table"></pre>
		</div>

        <div class="forecast-legend-container">
            <pre class="forecast-legend">
                <h4>Available Infomation</h3>
Hour           -> time of the day
Status         -> weather status. "Good" means cloud cover is below 25% and wind speed is below 15 km/h
Moon           -> Moon illumination percentage
Low, Mid, High -> cloud cover percentage at different altitudes
Wind           -> wind speed in km/h
Gusts          -> wind gusts in km/h
Seeing         -> seeing conditions in arcseconds
            </pre>
        </div>

        <div class="forecast-about-container">
            <pre class="forecast-about">
                <h4>About</h3>
This page is the result of my long time struggle to find clean weather forecast for astrophotography.
The one that gives simple answer to the simple question: "Is weather good for astrophotography tonight/next 3 hours/tomorrow/etc.?"
So here it is. Ultra-minimalistic weather forecast for astrophotographers. No humidity, no precipitation, no pressure or dew point. 
Just cloud cover and wind speed. And a clear status Good/Bad. From my experience this is all that matters.
            </pre>
        </div>

        <div class=forecast-legend-container>
            <pre class=forecast-legend>
                <h4>Contact & Links</h3>
-> All information comes from <a href="https://open-meteo.com/">Open-Meteo.com</a> API.
-> Source code is available on <a href="https://github.com/shuienko/aweather.info">GitHub</a>. Pull requests are welcome.
-> For any questions, suggestions or feedback please contact me via email: <b>contact@aweather.info</b>
-> My <a href="https://www.astrobin.com/users/maffei/">Astrobin</a> profile. Don't judge! I have feelings too :)
-> And if you like this page, please consider supporting me using the button below. Thank you!
            </pre>
        </div>
        
        <div class="footer">
                <form action="https://www.paypal.com/donate" method="post" target="_top">
                <input type="hidden" name="hosted_button_id" value="R3TWPEY9X4YKA" />
                <input type="image" src="https://www.paypalobjects.com/en_US/i/btn/btn_donate_SM.gif" border="0" name="submit" title="PayPal - The safer, easier way to pay online!" alt="Donate with PayPal button" />
                <img alt="" border="0" src="https://www.paypal.com/en_ES/i/scr/pixel.gif" width="1" height="1" />
            </form>
            © aweather.info | <a href="https://open-meteo.com/">Weather data by Open-Meteo.com</a>
        </div>
    </div>

    <script>
        function fetchSuggestions(query) {
            if (query.length < 2) {
                document.getElementById('suggestions').style.display = 'none';
                return;
            }
            fetch('/suggestions?q=' + encodeURIComponent(query))
                .then(response => response.json())
                .then(data => {
                    const suggestions = document.getElementById('suggestions');
                    suggestions.innerHTML = '';
                    if (data.length > 0) {
                        suggestions.style.display = 'block';
                        data.forEach(function(item) {
                            const regions = [item.admin1, item.admin2, item.admin3, item.admin4]
                                .filter(function(region) {
                                    return region && region.trim().length > 0;
                                })
                                .join(', ');
                            const fullName = item.name + (regions ? ', ' + regions : '') + ', ' + item.country;

                            const li = document.createElement('li');
                            li.textContent = fullName;
                            li.onclick = function() {
                                document.getElementById('city').value = fullName;

                                // Ensure the latitude and longitude inputs exist and set their values
                                const latitudeInput = document.getElementById('latitude');
                                const longitudeInput = document.getElementById('longitude');
                                if (latitudeInput && longitudeInput) {
                                    latitudeInput.value = item.latitude;
                                    longitudeInput.value = item.longitude;
                                } else {
                                    console.error("Latitude or longitude input not found in the DOM.");
                                }

                                suggestions.style.display = 'none';
                            };
                            suggestions.appendChild(li);
                        });
                    } else {
                        suggestions.style.display = 'none';
                    }
                })
                .catch(console.error);
        }

        function fetchWeather() {
            const cityNameInput = document.getElementById("city");
            const latitudeInput = document.getElementById("latitude");
            const longitudeInput = document.getElementById("longitude");

            const cityName = cityNameInput.value;
            const latitude = latitudeInput.value;
            const longitude = longitudeInput.value;

            if (!latitude || !longitude || isNaN(latitude) || isNaN(longitude)) {
                return alert("Please select a valid suggestion from the list");
            }

            // Save city name, latitude, and longitude in cookies (encoded)
            document.cookie = "cityName=" + encodeURIComponent(cityName) + "; path=/";
            document.cookie = "latitude=" + encodeURIComponent(latitude) + "; path=/";
            document.cookie = "longitude=" + encodeURIComponent(longitude) + "; path=/";

            const shortName = cityName.split(",")[0].trim();
            const country = cityName.split(",").slice(-1)[0].trim();

            // Display forecast details
            const forecastDetails = document.getElementById("forecastDetails");
            forecastDetails.textContent = shortName + ", " + country + "  |  lat: " + latitude + ", lon: " + longitude;
            forecastDetails.style.display = "block";

            // Fetch weather data from the backend
            fetch("/weather?lat=" + encodeURIComponent(latitude) + "&lon=" + encodeURIComponent(longitude))
                .then(function (response) {
                    if (!response.ok) {
                        throw new Error("Error fetching weather data: " + response.statusText);
                    }
                    return response.text();
                })
                .then(function (data) {
                    document.getElementById("weatherResult").textContent = data;
                })
                .catch(function (error) {
                    console.error(error);
                    alert("Failed to fetch weather data. Please try again later.");
                });
        }
        // Decode cookie values and populate the input fields
        function loadCookies() {
            const cookies = document.cookie.split("; ");
            const cookieMap = {};
            cookies.forEach(cookie => {
                const [key, value] = cookie.split("=");
                cookieMap[key] = decodeURIComponent(value);
            });

            // Populate fields if cookies exist
            const cityNameInput = document.getElementById("city");
            const latitudeInput = document.getElementById("latitude");
            const longitudeInput = document.getElementById("longitude");

            if (cookieMap.cityName) {
                cityNameInput.value = cookieMap.cityName;
            }
            if (cookieMap.latitude) {
                latitudeInput.value = cookieMap.latitude;
            }
            if (cookieMap.longitude) {
                longitudeInput.value = cookieMap.longitude;
            }
        }
    document.addEventListener("DOMContentLoaded", loadCookies);
    </script>
</body>
</html>
`

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Retrieve cookies
	cityNameCookie, _ := r.Cookie("cityName")
	latCookie, _ := r.Cookie("latitude")
	lonCookie, _ := r.Cookie("longitude")

	cityName := ""
	latitude := ""
	longitude := ""

	// Set defaults if cookies exist
	if cityNameCookie != nil {
		cityName = cityNameCookie.Value
	}
	if latCookie != nil {
		latitude = latCookie.Value
	}
	if lonCookie != nil {
		longitude = lonCookie.Value
	}

	// Insert cookies into JavaScript for the frontend
	pageWithCookies := fmt.Sprintf(indexHTML, cityName, latitude, longitude)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageWithCookies)
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	if lat == "" || lon == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	latitude, err1 := strconv.ParseFloat(lat, 64)
	longitude, err2 := strconv.ParseFloat(lon, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid latitude or longitude", http.StatusBadRequest)
		return
	}

	data := OpenMeteoAPIResponse{}
	data.FetchData(OpenMeteoAPIEndpoint, OpenMeteoAPIParams, float64ToSting(latitude), float64ToSting(longitude))
	markdownTable := data.Points().setMoonIllumination().setSeeing().Print()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, markdownTable)
}

func handleSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter is missing", http.StatusBadRequest)
		return
	}

	suggestions, err := fetchSuggestions(query)
	if err != nil {
		http.Error(w, "Unable to fetch suggestions", http.StatusInternalServerError)
		return
	}

	// If a suggestion is selected, save it in cookies
	if len(suggestions) > 0 {
		selected := suggestions[0] // Assuming the first result is selected by default

		// Save cookies for city name, latitude, and longitude
		http.SetCookie(w, &http.Cookie{
			Name:  "cityName",
			Value: selected.Name,
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "latitude",
			Value: fmt.Sprintf("%f", selected.Lat),
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "longitude",
			Value: fmt.Sprintf("%f", selected.Lon),
			Path:  "/",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func float64ToSting(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
