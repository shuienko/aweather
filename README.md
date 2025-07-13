# 🌟 [aweather.info](https://aweather.info)

<p align="center">
  <img src="https://github.com/shuienko/aweather.info/blob/main/src/static/favicon-192x192.png?raw=true" alt="Logo" width="192"/>
</p>

<p align="center">
    <b>🔭 Weather forecast for astrophotographers 🌙</b>
</p>

## 🌌 Why This Exists

Ever spent hours planning the perfect astrophotography session, only to discover the "clear skies" forecast didn't account for high-altitude clouds ruining your shots? Yeah, me too. 😤

Most weather apps are designed for people who want to know if they need an umbrella. We need to know if the atmosphere will cooperate with our telescopes and cameras.

**The Mission**: Answer one simple question with scientific precision:
> *"Will tonight be good for astrophotography?"*

No more scrolling through humidity percentages, dew points, or precipitation chances that don't matter when you're pointing a camera at the stars. Just the essentials: **cloud cover**, **wind speed**, **atmospheric seeing**, and a satisfying **"OK"** when conditions align. ✨

Built with real meteorological data from [Open-Meteo.com](https://open-meteo.com/) and actual atmospheric physics!

## ✨ What Makes It Special

- **🌫️ Smart Cloud Tracking**: Monitors low, mid, and high-altitude clouds separately (because that wispy cirrus at 30,000ft matters!)
- **🌬️ Atmospheric Seeing Calculation**: Advanced turbulence estimation using temperature gradients and wind shear across multiple atmospheric levels
- **🌙 Lunar Intelligence**: Moon phase, illumination percentage, and rise/set times
- **☀️ Solar Timing**: Precise sunrise/sunset calculations for golden hour planning
- **⚡ Lightning Fast**: 10-minute API caching keeps responses snappy
- **🎯 Location Smart**: Type any city name and get instant geocoding suggestions
- **📱 No Bloat**: Clean, responsive interface that loads in milliseconds

## 🛠️ Built With
- **Backend**: Go 1.23.4 (because speed matters when you're chasing clear skies)
- **Frontend**: Vanilla HTML/JS + UIkit (keeping it simple and fast)
- **Caching**: bigcache for blazing-fast in-memory storage
- **APIs**: Open-Meteo for weather data and geocoding
- **Science**: Meeus astronomical algorithms and atmospheric physics

## 🚀 Quick Start

### Try It Live
Just visit **[aweather.info](https://aweather.info)** - no setup required!

### Run Your Own Instance
```bash
# Clone and run locally
git clone https://github.com/shuienko/aweather.info.git
cd aweather.info/src
go run .
# Visit http://localhost:8080

# Or use Docker
docker build -t aweather .
docker run -d -p 8080:8080 aweather
```

> 💡 **Pro tip**: No API key required! Open-Meteo is generous enough to let us query their free tier.

## 🔬 The Science Behind the Magic

### Atmospheric Seeing Calculation
One of the coolest features is our **atmospheric seeing estimation** - we don't just tell you if it's cloudy, we tell you how **steady** the atmosphere is.

Think of it like this: even on a perfectly clear night, atmospheric turbulence can turn stars into dancing blobs in your telescope. Our algorithm crunches real meteorological data to predict seeing conditions:

**What We Calculate:**
- 🌡️ **Temperature gradients** across different atmospheric layers (surface → 850hPa → 500hPa)
- 💨 **Wind shear** between ground level and jet stream altitudes  
- ⚡ **Richardson Number** for atmospheric stability
- 🌪️ **Jet stream influence** (those 200hPa winds matter!)

**The Formula:**
> seeing ∝ V^0.6 × T_grad^0.4

Where V is jet stream wind speed and T_grad is temperature gradient.

**Real-World Impact:**
- 🟢 **Good seeing** (< 2"): Pin-sharp planetary details, tight star images
- 🟡 **Moderate seeing** (2-4"): Still great for deep sky, softer planetary views  
- 🔴 **Poor seeing** (> 4"): Stars look like donuts, planets are fuzzy blobs

This isn't just weather data - it's **atmospheric physics** tailored for astrophotographers! 🤓

### Astronomical Calculations
We use the industry-standard **Meeus algorithms** for precise:
- 🌅 **Sunrise/sunset** times (down to the second)
- 🌙 **Moon phases** and illumination percentages
- 🌔 **Moonrise/moonset** times (including next/previous day calculations)

Perfect for planning those new moon sessions or golden hour timelapses!

## 📊 Detailed Seeing Evaluation

### Formula for Seeing in Arcseconds

The seeing ( $\epsilon$ ) in arcseconds can be approximated using the following formula, which is based on the Kolmogorov turbulence theory:

$\epsilon \approx 0.98 \cdot \lambda^{-1/5} \cdot r_0^{-6/5}$

Where:
* $\lambda$ : Observing wavelength in meters (e.g., 500 nm =  $5 \times 10^{-7}$  m for visible light).
* $r_0$ : Fried's parameter (coherence length) in meters, representing the largest aperture over which turbulence is coherent.

Estimating Fried's Parameter ( $r_0$ ):

Fried's parameter depends on atmospheric conditions and can be estimated as:

$r_0 = \left( \frac{0.423 \cdot (2\pi)^2}{k^2 \cdot \sec(\theta)} \int_0^\infty C_n^2(h) \cdot dh \right)^{-3/5}$

Where:
* $k = 2\pi / \lambda$ : Wavenumber.
* $C_n^2(h)$ : Refractive index structure constant at height  $h$ , describing turbulence strength.

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
