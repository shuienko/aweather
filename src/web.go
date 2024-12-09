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
            font-weight: bold;
        }
		.forecast-table-container {
    		overflow-x: auto;
    		white-space: nowrap;
			text-align: center;
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
        <h1 class="uk-heading-divider forecast-header">Weather Forecast for Astrophotographers</h1>
        
        <div class="uk-grid-small uk-flex-middle" uk-grid>
            <div class="uk-width-expand">
                <input class="uk-input" id="city" type="text" placeholder="e.g., London" oninput="fetchSuggestions(this.value)">
                <ul id="suggestions" class="uk-list uk-list-divider" style="position: absolute; z-index: 1000; background: white; display: none;"></ul>
            </div>
            <div>
                <button class="uk-button uk-button-primary" onclick="fetchWeather()">Get Forecast</button>
            </div>
        </div>

        <!-- Hidden fields to store latitude and longitude -->
        <input type="hidden" id="latitude" value="">
        <input type="hidden" id="longitude" value="">

        <div id="forecastDetails" class="forecast-details" style="display: none;"></div>

		<div class="forecast-table-container">
	        <pre id="weatherResult" class="uk-margin-top forecast-table"></pre>
		</div>

        <div class="footer">
            © aweather.info
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
            const latitude = document.getElementById('latitude').value;
            const longitude = document.getElementById('longitude').value;

            if (!latitude || !longitude || isNaN(latitude) || isNaN(longitude)) {
                return alert('Please select a valid suggestion from the list');
            }

            const cityName = document.getElementById('city').value;
			const shortName = cityName.split(',')[0].trim();
			const country = cityName.split(',')[1].trim();

            // Display forecast details
            const forecastDetails = document.getElementById('forecastDetails');
            forecastDetails.textContent = shortName + ", " + country + "  |  lat: " + latitude + ", lon: " + longitude;
            forecastDetails.style.display = 'block';

            // Fetch weather data
            fetch('/weather?lat=' + encodeURIComponent(latitude) + '&lon=' + encodeURIComponent(longitude))
                .then(response => response.text())
                .then(data => {
                    document.getElementById('weatherResult').textContent = data;
                })
                .catch(console.error);
        }
    </script>
</body>
</html>
`

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexHTML)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func float64ToSting(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
