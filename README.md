# [aweather](https://aweather.shnk.net/)

<p align="center">
  <img src="https://github.com/shuienko/aweather/blob/main/src/static/favicon-192x192.png?raw=true" alt="Logo" width="192"/>
</p>

<p align="center">
    <b>Weather forecast for astrophotographers</b>
</p>


## Overview

This project comes from my long-time struggle to find a clean, no-nonsense weather forecast for astrophotography.

I wanted something that answers one simple question: "Is the weather good for astrophotography tonight, in the next 3 hours, or tomorrow?"

So, here it is – an ultra-minimalist weather forecast built specifically for astrophotographers. 

No humidity, no precipitation, no pressure, or dew point to fuss over.

Just cloud cover, wind speed, and a simple "Ok" when conditions are good.

All information comes from [Open-Meteo.com](https://open-meteo.com/) API.

## Features
- **Detailed Weather Forecast**: Provides temperature, cloud cover (low/mid/high), wind speed and gusts, moon illumination, and seeing.
- **Sun & Moon Calculations**: Calculates sunrise, sunset, moonrise, and moonset times.
- **Minimalist Interface**: Focuses on relevant metrics for astrophotography, avoiding unnecessary weather details.
- **Caching**: API responses are cached to improve performance and reduce API calls.
- **Location Suggestions**: Offers geolocation suggestions for easier city selection.

## Tech Stack
- **Frontend**: HTML, JavaScript, Tailwind CSS (via CDN)
- **Backend**: Golang (net/http for web server, Open‑Meteo API integration)
- **Caching**: bigcache for in‑memory caching
- **Geolocation**: Open‑Meteo Geocoding API

## Local development

### Prerequisites
- Go 1.23+

### Run locally
```bash
cd src
go run .
```

Open `http://localhost:8080`.

### Run tests
```bash
cd src
go test ./...
```

## HTTP endpoints
- `GET /` – HTML UI (served with embedded templates and static assets)
- `GET /weather?lat=<lat>&lon=<lon>` – returns a plain‑text table forecast
- `GET /suggestions?q=<query>` – JSON location suggestions (Open‑Meteo Geocoding)
- `GET /robots.txt`, `GET /favicon.ico`, `GET /static/*`

## Deployment
This website is deployed to [aweather.shnk.net/](https://aweather.shnk.net/)

### Build image
Build a Docker image from the repo root (the Dockerfile expects sources under `src/`).
```bash
docker build -t aweather:latest .
```
* The `-t` flag allows you to tag the image with a name and version.

### Run image
Run the container and map port 8080.
```bash
docker run -d -p 8080:8080 --name aweather aweather:latest
```
* The `-d` flag runs the container in detached mode (in the background).
* The `-p 8080:8080` maps the container's `8080` port to your host's `8080` port.
* The `--name aweather` assigns a name to the container for easier management.

#### Open `http://localhost:8080` in your browser to access the website

#### No API key required (Open‑Meteo does not require authentication).

## Configuration
- **Thresholds**: `ok` status means cloud cover ≤ 25% at all levels and wind speed/gusts < 15 km/h (see `MaxCloudCover`, `MaxWindSpeed`).
- **Cache**: in‑memory cache TTL is 10 minutes.
- **Port**: the server listens on port `8080`.

## Seeing evaluation

### Formula for Seeing in Arcseconds

The seeing ( $\epsilon$ ) in arcseconds can be approximated using the following formula, which is based on the Kolmogorov turbulence theory:


$\epsilon \approx 0.98 \cdot \lambda^{-1/5} \cdot r_0^{-6/5}$


Where:
* $\lambda$ : Observing wavelength in meters (e.g., 500 nm =  $5 \times 10^{-7}$  m for visible light).
* $r_0$ : Fried’s parameter (coherence length) in meters, representing the largest aperture over which turbulence is coherent.

Estimating Fried’s Parameter ( $r_0$ ):

Fried’s parameter depends on atmospheric conditions and can be estimated as:

$r_0 = \left( \frac{0.423 \cdot (2\pi)^2}{k^2 \cdot \sec(\theta)} \int_0^\infty C_n^2(h) \cdot dh \right)^{-3/5}$


Where:
	•	 $k = 2\pi / \lambda$ : Wavenumber.
	•	 $C_n^2(h)$ : Refractive index structure constant at height  $h$ , describing turbulence strength.

If $C_n^2(h)$  is not directly available, meteorological proxies can help approximate seeing conditions.

### Practical Approach Using Meteorological Data

#### Overview
The `setSeeing` function, located in `datapoints.go`, enhances the weather forecasting capability by estimating atmospheric seeing conditions. This is crucial for astrophotographers, as good seeing conditions directly affect the clarity of celestial observations.

#### What the Function Does
The `setSeeing` function computes a value representing atmospheric turbulence, which can distort the quality of astronomical images. It uses meteorological data to approximate the level of turbulence at different altitudes by evaluating:

- **Temperature Gradients** – Differences in temperature between various atmospheric layers (surface to $850 hPa$ and $850 hPa$ to $500 hPa$).
- **Wind Shear** – Variations in wind speed between the ground level, $850 hPa$, and $200 hPa$.
- **Jet Stream Influence** – Penalizes seeing conditions when wind speeds at $200 hPa$ exceed a threshold (15 m/s by default).
- **Richardson Number ($Ri$)** – A measure of atmospheric stability, further modifying the seeing value when turbulence increases.

#### How it Works
1. **Temperature Gradients**  
   The function calculates the temperature difference between:
   - Surface and $850 hPa$ (low altitude)
   - $850 hPa$ and $500 hPa$ (mid-altitude)

   These differences are used to derive a temperature gradient across approximately 5 km of the atmosphere.

2. **Wind Shear Calculation**  
   Wind shear is computed by determining the absolute difference between:
   - Wind speeds at $200 hPa$ and surface wind speed
   - Surface wind speed and $850 hPa$ wind speed

3. **Seeing Formula**  
   The formula to estimate seeing is:   $\epsilon \propto V^{0.6} \cdot T_{\text{grad}}^{0.4}$
    
    Where:
    * $V$ : Wind speed at $200–300 hPa$ (jet stream).
    * $T_{\text{grad}}$ : Temperature gradient between ground and upper atmosphere.
4. **Jet Stream Adjustment**  
   If wind speeds at $200 hPa$ surpass 15 m/s, the seeing value is penalized proportionally, representing increased turbulence due to jet streams.

5. **Richardson Number Adjustment**  
   The Richardson Number ($Ri$) is calculated by dividing the temperature gradient by the square of the wind speed.  
   If $Ri$ is:
   - Less than 0.25, seeing is increased by 50%.
   - Between 0.25 and 0.5, seeing is increased by 20%.

#### Why This Matters
 Seeing gives the astrophotographers an easy-to-understand measure of atmospheric stability. This allows for better planning of observation sessions by identifying times with favorable conditions for clear imaging.
