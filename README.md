# [aweather.info](https://aweather.info)
Weather for Astrophotographers

## About

This project is the result of my long time struggle to find clean weather forecast for astrophotography.
The one that gives simple answer to the simple question: "Is weather good for astrophotography tonight/next 3 hours/tomorrow/etc.?"

So here it is. Ultra-minimalistic weather forecast for astrophotographers. 

No humidity, no precipitation, no pressure or dew point. 
Just cloud cover and wind speed. And a clear status Good/Bad. 
From my experience this is all that matters.

All information comes from [Open-Meteo.com](https://open-meteo.com/) API.

## Deployment
This website is deployed to [aweather.info](https://aweather.info)

Feel free to build your own docker container and deploy using any convinient method.
```
docker build
```
Please keep in mind no Open-Meteo API key set by default.

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

If detailed turbulence data isn’t available, you can estimate seeing based on wind speed, temperature gradients, and altitude.
1.	Wind Speed:
    * Strong winds at high altitudes (jet stream) contribute to turbulence.
	* Seeing is worse when high-altitude wind speeds exceed $~15 m/s$.
2.	Temperature Gradients:
	* Sudden temperature changes between ground and upper atmosphere (inversions) increase turbulence.
3. Empirical Approximation:

    $\epsilon \propto V^{0.6} \cdot T_{\text{grad}}^{0.4}$

    Where:
    * $V$ : Wind speed at $200–300 hPa$ (jet stream).
    * $T_{\text{grad}}$ : Temperature gradient between ground and upper atmosphere.
4.	Example Seeing Estimation:
If wind speed at 10 km altitude ( $V$ ) is $20 m/s$ and the temperature gradient is $5°C/km$: 
    * $\epsilon \approx 1.5$  arcseconds.