## UX suggestions for aweather

A prioritized list of UX improvements and feature ideas tailored for astrophotographers.

### High‑impact quick wins
- **Use my location**: Geolocation API button to auto‑fill `lat/lon` and city, then fetch. Fallback with a friendly error if denied.
- **Shareable links**: Support `?lat=..&lon=..&city=..` so users can bookmark/share a specific spot. Update page title with city after selection.
- **Favorites**: Allow saving multiple locations (localStorage). Quick dropdown under the search bar.
- **Night/Red mode**: Add dark and red‑light themes to preserve night vision. Persist preference.
- **Next good window summary**: Above the cards, show contiguous “ok” windows: e.g., “Good 22:00–01:00; tomorrow 21:30–00:30.”

### Personalization controls
- **Threshold presets**: Small settings popover with Strict/Normal/Lenient presets for cloud cover and wind caps. Persist choice.
- **Unit toggles**: °C/°F and km/h/mph toggles. Persist choice.
- **12h time option**: Toggle between 24h and 12h display.

### Visual expressiveness (keep minimalism)
- **Inline day heatmap**: Compact 24‑segment bar per day colored by ok/neutral/bad; keep the text table.
- **Highlight “ok” rows**: Subtle background or badge to speed scanning.
- **Seeing labels**: Add Good (<1”), Fair (1–2”), Poor (>2”) labels next to arcsec values with a tiny legend.

### Astro‑specific context
- **Astronomical darkness window**: Show astronomical twilight end/start alongside sunrise/sunset.
- **Moon altitude cue**: Dim hours where the Moon is above the horizon or add a small ↑/↓ next to moon % using existing rise/set.
- **Light pollution hint**: Static Bortle estimate per location as a small tag near the location line (approximate dataset).

### Polish and accessibility
- **Better empty/error states**: When no city is selected, show a hint; keep validation near the input.
- **Keyboard accessibility**: Arrow‑key navigation for suggestions, `aria-*` roles, and `aria-expanded` on the dropdown.
- **Mobile UX**: Sticky “Get forecast” on mobile while loading; ensure cards scroll horizontally well.
- **Freshness hint**: Show “Updated just now” with API time; if cached, “From cache (<n> min old)”.

### Performance and SEO
- **PWA**: Manifest + Service Worker for offline caching of the last viewed location; installable on mobile.
- **Meta/OG tags**: Add meta description; when a location is active, update `<title>` and meta for better share cards.
- **Font optimization**: Preload required Inter weights; keep `display=swap`. Defer non‑critical scripts.

### Integrations and retention (optional)
- **Alerts**: In‑tab notifications: “Notify me when tonight becomes ‘ok’” using periodic fetch + Notification API (while tab is open).
- **Calendar export**: “Add good window to Calendar” by downloading an `.ics` for the day’s windows.
- **External links**: Tiny links to compare with regional astro forecast sites for the selected coordinates.

### Copy/structure
- **Condense About/Table**: Collapse into accordions to keep results above the fold.
- **Feedback link**: Add mailto or GitHub Issues near Donate; optionally add Ko‑fi.


