# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

aweather.info is a minimalist weather forecast application built specifically for astrophotographers. It provides clean, focused weather data answering the question: "Is the weather good for astrophotography tonight?"

**Tech Stack:**
- **Backend**: Go 1.23.4 with net/http standard library
- **Frontend**: HTML/JavaScript with UIkit framework  
- **Caching**: bigcache for in-memory API response caching
- **APIs**: Open-Meteo.com for weather data and geocoding
- **Deployment**: Docker containerized application

## Commands

### Development
```bash
# Navigate to source directory
cd src/

# Run the application locally
go run .

# Build the application
go build -o app .

# Run tests
go test ./...

# Run specific test
go test -run TestFunctionName

# Check dependencies
go mod tidy
go mod verify
```

### Docker
```bash
# Build Docker image
docker build -t aweather:latest .

# Run Docker container
docker run -d -p 8080:8080 --name aweather aweather:latest
```

## Architecture

### Core Components

**main.go**: Application entry point and HTTP server setup
- Defines global constants (MaxCloudCover=25%, MaxWindSpeed=15km/h, CacheTTL=10min)
- Initializes bigcache and HTTP routes
- Serves on port 8080

**web.go**: HTTP handlers and routing
- `handleIndex()`: Serves main page with embedded HTML template
- `handleWeather()`: API endpoint for weather data
- `handleSuggestions()`: Geocoding suggestions endpoint
- Uses embedded static files and templates

**open-meteo.go**: External API integration
- Fetches weather data from Open-Meteo API
- Handles geocoding requests for location suggestions
- Implements caching layer for API responses
- Parses complex meteorological data (temperature, clouds, wind at multiple pressure levels)

**datapoints.go**: Core weather data processing
- `DataPoint` struct: Represents weather conditions at specific time
- `isGood()` method: Determines if conditions are suitable for astrophotography
- `setMoonIllumination()`: Calculates moon illumination percentage
- `setSeeing()`: **Complex atmospheric seeing calculation** using meteorological proxies

**sun-and-moon.go**: Astronomical calculations
- Sunrise/sunset times using Meeus astronomical algorithms  
- Moonrise/moonset calculations
- Moon phase and illumination calculations

### Key Dependencies
- `github.com/allegro/bigcache/v3`: High-performance in-memory caching
- `github.com/soniakeys/meeus/v3`: Astronomical calculations library
- `github.com/hablullah/go-sampa`: Solar position algorithms

### Data Flow
1. User requests weather forecast for location
2. Geocoding API suggests locations if needed  
3. Weather data fetched from Open-Meteo API (cached for 10 minutes)
4. Raw meteorological data processed into DataPoints
5. Atmospheric seeing calculated using temperature gradients and wind shear
6. Moon illumination and sun/moon times calculated
7. Each datapoint evaluated for astrophotography suitability
8. Frontend displays simplified "good/bad" conditions

### Seeing Calculation
The `setSeeing()` function in datapoints.go implements a sophisticated atmospheric turbulence estimation using:
- Temperature gradients between surface, 850hPa, and 500hPa levels
- Wind shear between surface, 850hPa, and 200hPa
- Richardson Number for atmospheric stability
- Jet stream influence (>15 m/s at 200hPa penalized)
- Formula: seeing ∝ V^0.6 × T_grad^0.4

This provides astrophotographers with practical seeing estimates for planning observation sessions.