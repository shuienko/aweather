// aweather frontend script
// Organised: event wiring, helpers, UI actions

// ---- Event wiring ----
document.addEventListener("DOMContentLoaded", () => {
  const cityInput = document.getElementById("city");
  const latitudeInput = document.getElementById("latitude");
  const longitudeInput = document.getElementById("longitude");
  const suggestionsEl = document.getElementById("suggestions");
  const unitTempC = document.getElementById("unitTempC");
  const unitTempF = document.getElementById("unitTempF");
  const unitWindKmh = document.getElementById("unitWindKmh");
  const unitWindMph = document.getElementById("unitWindMph");
  const time24h = document.getElementById("time24h");
  const time12h = document.getElementById("time12h");

  // Clear coordinates if user backspaces the query
  cityInput.addEventListener("input", (event) => {
    if (event.inputType === "deleteContentBackward") {
      cityInput.value = "";
      latitudeInput.value = "";
      longitudeInput.value = "";
      hideSuggestions();
      clearDebounce();
    }
  });

  // Hide suggestions when clicking outside
  document.addEventListener("click", (e) => {
    if (e.target !== cityInput && !suggestionsEl.contains(e.target)) {
      hideSuggestions();
    }
  });

  // Pressing Enter triggers fetch
  cityInput.addEventListener("keypress", (event) => {
    if (event.key === "Enter") {
      event.preventDefault();
      fetchWeather();
    }
  });

  loadCookies();

  // Wire unit/time toggle listeners (if elements exist)
  const maybeRefetch = () => {
    const lat = latitudeInput.value;
    const lon = longitudeInput.value;
    if (lat && lon && !isNaN(lat) && !isNaN(lon)) {
      fetchWeather();
    }
  };

  function setCookie(name, value) {
    const maxAge = "max-age=" + 365 * 24 * 60 * 60;
    document.cookie = `${name}=${encodeURIComponent(value)}; path=/; ${maxAge}`;
  }

  if (unitTempC && unitTempF) {
    unitTempC.addEventListener("click", () => {
      setCookie("unitTemp", "c");
      unitTempC.classList.add("bg-blue-600", "text-white");
      unitTempC.classList.remove("bg-white", "text-blue-600");
      unitTempF.classList.remove("bg-blue-600", "text-white");
      unitTempF.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
    unitTempF.addEventListener("click", () => {
      setCookie("unitTemp", "f");
      unitTempF.classList.add("bg-blue-600", "text-white");
      unitTempF.classList.remove("bg-white", "text-blue-600");
      unitTempC.classList.remove("bg-blue-600", "text-white");
      unitTempC.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
  }
  if (unitWindKmh && unitWindMph) {
    unitWindKmh.addEventListener("click", () => {
      setCookie("unitWind", "kmh");
      unitWindKmh.classList.add("bg-blue-600", "text-white");
      unitWindKmh.classList.remove("bg-white", "text-blue-600");
      unitWindMph.classList.remove("bg-blue-600", "text-white");
      unitWindMph.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
    unitWindMph.addEventListener("click", () => {
      setCookie("unitWind", "mph");
      unitWindMph.classList.add("bg-blue-600", "text-white");
      unitWindMph.classList.remove("bg-white", "text-blue-600");
      unitWindKmh.classList.remove("bg-blue-600", "text-white");
      unitWindKmh.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
  }
  if (time24h && time12h) {
    time24h.addEventListener("click", () => {
      setCookie("time12h", "0");
      time24h.classList.add("bg-blue-600", "text-white");
      time24h.classList.remove("bg-white", "text-blue-600");
      time12h.classList.remove("bg-blue-600", "text-white");
      time12h.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
    time12h.addEventListener("click", () => {
      setCookie("time12h", "1");
      time12h.classList.add("bg-blue-600", "text-white");
      time12h.classList.remove("bg-white", "text-blue-600");
      time24h.classList.remove("bg-blue-600", "text-white");
      time24h.classList.add("bg-white", "text-blue-600");
      maybeRefetch();
    });
  }
});

// ---- Helpers ----
let debounceTimer;
function clearDebounce() {
  if (debounceTimer) {
    clearTimeout(debounceTimer);
  }
}

function formatPlaceName(item) {
  const regions = [item.admin1, item.admin2, item.admin3, item.admin4]
    .filter((region) => region && region.trim().length > 0)
    .join(", ");
  const country = item.country && item.country.trim().length > 0 ? item.country : item.country_code;
  return item.name + (regions ? ", " + regions : "") + ", " + country;
}

function parseCookies() {
  return document.cookie.split("; ").reduce((acc, cookie) => {
    const [k, v] = cookie.split("=");
    if (k) acc[k] = decodeURIComponent(v || "");
    return acc;
  }, {});
}

function setError(msg) {
  const el = document.getElementById("error");
  el.textContent = msg;
  el.style.display = "block";
}

function clearError() {
  const el = document.getElementById("error");
  el.textContent = "";
  el.style.display = "none";
}

// Safely escape HTML for rendering inside elements where we want to inject markup selectively
function escapeHtml(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

// Detect if a given table line has "ok" in the ok? column
function isOkRow(line) {
  if (!line || line.indexOf("|") === -1) return false;
  const cols = line.split("|").map((s) => s.trim());
  // The ok? column is the second column in our table
  // Header is "ok?"; separator rows are dashes; data rows contain "ok" or "-"
  return cols.length >= 2 && cols[1] === "ok";
}

// Convert plain-text table into HTML, bolding rows that are "ok"
function tableTextToHtml(tableText) {
  const lines = tableText.split("\n");
  const htmlLines = lines.map((line) => {
    const escaped = escapeHtml(line);
    return isOkRow(line) ? `<span class="ok-row">${escaped}</span>` : escaped;
  });
  return htmlLines.join("\n");
}

function showSuggestions() {
  document.getElementById("suggestions").style.display = "block";
}

function hideSuggestions() {
  document.getElementById("suggestions").style.display = "none";
}

// ---- Actions exposed to HTML inline handlers ----
async function fetchSuggestions(query) {
  clearDebounce();
  const suggestionsEl = document.getElementById("suggestions");

  debounceTimer = setTimeout(async () => {
    if (!query || query.length < 2) {
      hideSuggestions();
      suggestionsEl.innerHTML = "";
      return;
    }

    try {
      const resp = await fetch("/suggestions?q=" + encodeURIComponent(query));
      const data = await resp.json();

      suggestionsEl.innerHTML = "";
      if (!Array.isArray(data) || data.length === 0) {
        hideSuggestions();
        return;
      }

      showSuggestions();
      data.forEach((item) => {
        const li = document.createElement("li");
        const fullName = formatPlaceName(item);
        li.textContent = fullName;
        li.onclick = (e) => {
          e.stopPropagation();
          const cityInput = document.getElementById("city");
          const latitudeInput = document.getElementById("latitude");
          const longitudeInput = document.getElementById("longitude");
          cityInput.value = fullName;
          latitudeInput.value = item.latitude;
          longitudeInput.value = item.longitude;
          hideSuggestions();
          clearDebounce();
          fetchWeather();
        };
        suggestionsEl.appendChild(li);
      });
    } catch (err) {
      console.error("Error fetching suggestions:", err);
      hideSuggestions();
    }
  }, 300);
}

async function fetchWeather() {
  const cityNameInput = document.getElementById("city");
  const latitudeInput = document.getElementById("latitude");
  const longitudeInput = document.getElementById("longitude");
  const forecastDetails = document.getElementById("forecastDetails");
  const weatherResult = document.getElementById("weatherResult");
  const loaderEl = document.getElementById("loader");
  const fetchBtn = document.getElementById("fetchBtn");

  hideSuggestions();
  clearError();

  const cityName = cityNameInput.value;
  const latitude = latitudeInput.value;
  const longitude = longitudeInput.value;

  if (!latitude || !longitude || isNaN(latitude) || isNaN(longitude)) {
    setError("Please select a valid suggestion from the list or use the location button.");
    return;
  }
  if (cityName && cityName.length < 2) {
    setError("Please enter a city name with at least 2 characters.");
    return;
  }

  const shortName = cityName ? cityName.split(",")[0].trim() : "My location";
  const country = cityName && cityName.includes(",") ? cityName.split(",").slice(-1)[0].trim() : "";
  forecastDetails.textContent = `${shortName}${country ? ", " + country : ""}  |  ${latitude},  ${longitude}`;
  forecastDetails.style.display = "block";

  weatherResult.innerHTML = "";
  loaderEl.style.display = "block";
  fetchBtn.disabled = true;
  cityNameInput.disabled = true;

  try {
    const cookies = parseCookies();
    const unitTemp = (cookies.unitTemp || "c").toLowerCase();
    const unitWind = (cookies.unitWind || "kmh").toLowerCase();
    const time12h = cookies.time12h === "1" ? "1" : "0";
    const resp = await fetch(`/weather?lat=${encodeURIComponent(latitude)}&lon=${encodeURIComponent(longitude)}&unit_temp=${encodeURIComponent(unitTemp)}&unit_wind=${encodeURIComponent(unitWind)}&time_12h=${encodeURIComponent(time12h)}`);
    if (!resp.ok) throw new Error("Error fetching weather data: " + resp.statusText);
    const text = await resp.text();
    renderWeather(text);
  } catch (err) {
    console.error(err);
    setError("Failed to fetch weather data. Please try again later.");
  } finally {
    loaderEl.style.display = "none";
    fetchBtn.disabled = false;
    cityNameInput.disabled = false;
  }

  const maxAge = "max-age=" + 365 * 24 * 60 * 60;
  const cityCookie = country ? `${shortName}, ${country}` : shortName;
  document.cookie = `cityName=${encodeURIComponent(cityCookie)}; path=/; ${maxAge}`;
  document.cookie = `latitude=${encodeURIComponent(latitude)}; path=/; ${maxAge}`;
  document.cookie = `longitude=${encodeURIComponent(longitude)}; path=/; ${maxAge}`;
}

function renderWeather(text) {
  const container = document.getElementById("weatherResult");
  container.innerHTML = "";
  if (!text || text.trim().length === 0) return;

  const dayBlocks = text
    .split(/\n{2,}/)
    .map((s) => s.trim())
    .filter(Boolean);

  dayBlocks.forEach((block) => {
    const lines = block.split("\n");
    if (lines.length === 0) return;

    const dateLine = lines[0] || "";
    const sunMoonLine = lines[1] || "";
    const tableText = lines.slice(2).join("\n");
    const tableTextWithBlankLine = tableText.replace(/\n*$/, "\n\n");

    const card = document.createElement("div");
    card.className = "rounded-xl border border-slate-200 bg-white px-4 py-3 text-[13.5px] shadow-sm";

    const header = document.createElement("div");
    header.className = "text-center";

    const dateEl = document.createElement("div");
    dateEl.className = "font-medium text-slate-700 font-mono";
    dateEl.textContent = dateLine;

    const sunMoonEl = document.createElement("div");
    sunMoonEl.className = "text-slate-500 font-mono";
    sunMoonEl.textContent = sunMoonLine;

    header.appendChild(dateEl);
    header.appendChild(sunMoonEl);

    const preWrap = document.createElement("div");
    preWrap.className = "mt-2 text-center overflow-x-auto";

    const pre = document.createElement("pre");
    pre.className = "inline-block text-left font-mono whitespace-pre leading-relaxed max-w-full";
    pre.innerHTML = tableTextToHtml(tableTextWithBlankLine);

    preWrap.appendChild(pre);
    card.appendChild(header);
    card.appendChild(preWrap);
    container.appendChild(card);
  });
}

function loadCookies() {
  const cookies = parseCookies();
  const cityNameInput = document.getElementById("city");
  const latitudeInput = document.getElementById("latitude");
  const longitudeInput = document.getElementById("longitude");
  if (cookies.cityName) cityNameInput.value = cookies.cityName;
  if (cookies.latitude) latitudeInput.value = cookies.latitude;
  if (cookies.longitude) longitudeInput.value = cookies.longitude;

  // Apply unit/time preferences to toggles
  const unitTemp = (cookies.unitTemp || "c").toLowerCase();
  const unitWind = (cookies.unitWind || "kmh").toLowerCase();
  const time12h = cookies.time12h === "1";
  const unitTempC = document.getElementById("unitTempC");
  const unitTempF = document.getElementById("unitTempF");
  const unitWindKmh = document.getElementById("unitWindKmh");
  const unitWindMph = document.getElementById("unitWindMph");
  const time24h = document.getElementById("time24h");
  const time12hBtn = document.getElementById("time12h");

  if (unitTempC && unitTempF) {
    if (unitTemp === "f") {
      unitTempF.classList.add("bg-blue-600", "text-white");
      unitTempF.classList.remove("bg-white", "text-blue-600");
      unitTempC.classList.remove("bg-blue-600", "text-white");
      unitTempC.classList.add("bg-white", "text-blue-600");
    } else {
      unitTempC.classList.add("bg-blue-600", "text-white");
      unitTempC.classList.remove("bg-white", "text-blue-600");
      unitTempF.classList.remove("bg-blue-600", "text-white");
      unitTempF.classList.add("bg-white", "text-blue-600");
    }
  }
  if (unitWindKmh && unitWindMph) {
    if (unitWind === "mph") {
      unitWindMph.classList.add("bg-blue-600", "text-white");
      unitWindMph.classList.remove("bg-white", "text-blue-600");
      unitWindKmh.classList.remove("bg-blue-600", "text-white");
      unitWindKmh.classList.add("bg-white", "text-blue-600");
    } else {
      unitWindKmh.classList.add("bg-blue-600", "text-white");
      unitWindKmh.classList.remove("bg-white", "text-blue-600");
      unitWindMph.classList.remove("bg-blue-600", "text-white");
      unitWindMph.classList.add("bg-white", "text-blue-600");
    }
  }
  if (time24h && time12hBtn) {
    if (time12h) {
      time12hBtn.classList.add("bg-blue-600", "text-white");
      time12hBtn.classList.remove("bg-white", "text-blue-600");
      time24h.classList.remove("bg-blue-600", "text-white");
      time24h.classList.add("bg-white", "text-blue-600");
    } else {
      time24h.classList.add("bg-blue-600", "text-white");
      time24h.classList.remove("bg-white", "text-blue-600");
      time12hBtn.classList.remove("bg-blue-600", "text-white");
      time12hBtn.classList.add("bg-white", "text-blue-600");
    }
  }
}

async function useMyLocation() {
  const errorEl = document.getElementById("error");
  const geoBtn = document.getElementById("geoBtn");
  const cityInput = document.getElementById("city");
  const latitudeInput = document.getElementById("latitude");
  const longitudeInput = document.getElementById("longitude");

  hideSuggestions();
  clearError();

  if (location.protocol !== "https:" && location.hostname !== "localhost") {
    setError("Using your location requires HTTPS. Please use the search box instead.");
    return;
  }
  if (!("geolocation" in navigator)) {
    setError("Geolocation is not supported by your browser. Please use the search box instead.");
    return;
  }

  const geoIcon = geoBtn.querySelector("svg");
  const originalAriaLabel = geoBtn.getAttribute("aria-label") || "";
  geoBtn.disabled = true;
  geoBtn.setAttribute("aria-label", "Locating…");
  if (geoIcon) geoIcon.classList.add("animate-spin");

  const getPosition = () =>
    new Promise((resolve, reject) =>
      navigator.geolocation.getCurrentPosition(resolve, reject, {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0,
      })
    );

  try {
    const pos = await getPosition();
    const lat = pos.coords.latitude.toFixed(6);
    const lon = pos.coords.longitude.toFixed(6);
    latitudeInput.value = lat;
    longitudeInput.value = lon;

    try {
      const resp = await fetch(
        `/reverse-geocoding?lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`
      );
      if (resp.ok) {
        const data = await resp.json();
        if (data && data.name) {
          cityInput.value = formatPlaceName(data);
        } else {
          cityInput.value = "My location";
        }
      } else {
        cityInput.value = "My location";
      }
    } catch (_) {
      cityInput.value = "My location";
    } finally {
      geoBtn.disabled = false;
      geoBtn.setAttribute("aria-label", originalAriaLabel || "Use my location");
      if (geoIcon) geoIcon.classList.remove("animate-spin");
      fetchWeather();
    }
  } catch (err) {
    geoBtn.disabled = false;
    geoBtn.setAttribute("aria-label", originalAriaLabel || "Use my location");
    if (geoIcon) geoIcon.classList.remove("animate-spin");

    let msg = "";
    switch (err.code) {
      case err.PERMISSION_DENIED:
        msg = "Location permission denied. You can use the search box instead.";
        break;
      case err.POSITION_UNAVAILABLE:
        msg = "Unable to determine your location. Please try the search box.";
        break;
      case err.TIMEOUT:
        msg = "Timed out while trying to get your location. Please try again.";
        break;
      default:
        msg = "Couldn't get your location. Please use the search box.";
    }
    setError(msg);
  }
}

// Expose functions for inline handlers
window.fetchSuggestions = fetchSuggestions;
window.fetchWeather = fetchWeather;
window.useMyLocation = useMyLocation;


