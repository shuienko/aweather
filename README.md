# aweather

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
- **Detailed Weather Forecast**: Provides temperature, cloud cover (low/mid/high), wind speed and gusts, moon illumination, and seeing index.
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

## Seeing index

The application derives a heuristic “seeing index” from available meteorological fields to help rank time slots for astrophotography. It is not a physically calibrated arcsecond value and should be interpreted as: lower is better.

### Inputs used
- Temperature: `temperature_2m`, `temperature_500hPa`
- Wind: `wind_speed_10m`, `wind_speed_850hPa`, `wind_speed_200hPa`
- Heights: `geopotential_height_500hPa` (and site elevation)

### Method (summary)
- Compute an elevation‑aware temperature lapse across the total depth from site elevation to 500 hPa: `(T2m − T500) / depth_km`.
- Form a vertical wind shear proxy: `|V200 − V850| + |V850 − V10|` (all in m/s).
- Seeing index ∝ `shear^0.6 * |lapse|^0.4`. A small jet‑stream penalty applies above ~22 m/s at 200 hPa, capped to avoid runaway values.
- The result is clamped to a reasonable range for readability (≈0.5–5.0).

This index is intended for relative comparison between hours/nights rather than absolute image resolution.
