# Standard Operating Procedure for Calculators

This document outlines the strict architecture for creating new interactive calculators in this Hugo project. All new tools **must** follow this pattern to ensure consistency, correct styling, and URL shareability.

## 1. File Structure

Each calculator requires exactly four files in these specific locations:

1.  **Content:** `content/tools/[tool-name].md` (Defines metadata & URL)
2.  **Layout:** `layouts/tools/[tool-name].html` (The HTML structure)
3.  **Logic:** `assets/js/[tool-name].js` (The math, interaction, and shareable logic)
4.  **Style:** `assets/scss/tools/_calculator.scss` (Shared styles - rarely needs editing)

---

## 2. Step-by-Step Implementation Guide

### Step 1: Create the Content File

**File:** `content/tools/[tool-name].md`
This defines the page title, description, and layout link.

```markdown
---
title: "Tool Name"
description: "A short, punchy description of what this tool calculates."
type: "tools/[tool-name]"
layout: "tools/[tool-name]"
tool_type: "calculator"
---

### Instructions

(Optional) Markdown content here appears below the calculator as a user guide.
```

### Step 2: Create the Layout (HTML)

**File:** `layouts/tools/[tool-name].html`
**Critical:** Use the standard CSS classes (`calculator-wrapper`, `calculator-row`, `calculator-field`) to inherit the site's theme automatically.

```html
{{ define "main" }}
<main class="container">
  <h1>{{ .Title }}</h1>
  <p class="description"><i>{{ .Description }}</i></p>
  <hr />
  <br />

  <div class="calculator-wrapper">
    <div class="calculator-sticky-header">
      <div class="calculator-main-display">
        <span class="label">Primary Result</span>
        <h1 id="outResult">$0.00</h1>
      </div>
    </div>

    <div class="calculator-body">
      <h3 class="calculator-section-head">1. Input Section Name</h3>

      <div class="calculator-row">
        <div class="calculator-field">
          <label>Input Label ($)</label>
          <input type="number" id="input1" value="100" />
        </div>
        <div class="calculator-field">
          <label>Input Label (%)</label>
          <input type="number" id="input2" value="5" />
        </div>
      </div>
    </div>
  </div>

  <br />
  <div class="tool-instructions">{{ .Content }}</div>
  <br />
  <hr />
  <br />
  <div class="center-items">
    <a href="/tools/" class="unchanging-link back-link"><- back to tools</a>
  </div>
</main>

{{ $shareable := resources.Get "js/shareable.js" | minify | fingerprint }}
<script src="{{ $shareable.RelPermalink }}"></script>

{{ $script := resources.Get "js/[tool-name].js" | minify | fingerprint }}
<script src="{{ $script.RelPermalink }}"></script>
{{ end }}
```

### Step 3: Create the Logic (JavaScript)

**File:** `assets/js/[tool-name].js`
**Requirement:** Must include the Universal Loader pattern to handle URL parameters and navigation states.

```javascript
/**
 * Main Calculation Function
 * Triggers on every input change
 */
function calculateMyTool() {
  // 1. Get Inputs (Use parseFloat || 0 for safety)
  const val1 = parseFloat(document.getElementById("input1").value) || 0;
  const val2 = parseFloat(document.getElementById("input2").value) || 0;

  // 2. Perform Math
  const result = val1 * (val2 / 100);

  // 3. Update UI
  const outResult = document.getElementById("outResult");
  outResult.innerText = formatCurrency(result);

  // Optional: Conditional Formatting (Red for negative, standard for positive)
  if (result < 0) {
    outResult.style.color = "#ef4444";
  } else {
    outResult.style.color = ""; // Resets to theme default (White/Grey)
  }
}

// Helper: Currency Formatting
function formatCurrency(num) {
  return (
    "$" +
    num.toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  );
}

/**
 * Universal Initialization Pattern
 * Handles:
 * 1. Hard Refresh (DOMContentLoaded)
 * 2. Soft Navigation (Turbo/HTMX)
 * 3. URL Parameter Sharing (shareable.js)
 */
function initMyTool() {
  // Safety check: specific ID for this tool to prevent errors on other pages
  if (!document.getElementById("input1")) return;

  // 1. Attach event listeners (CSP-compliant, no inline handlers)
  document.getElementById("input1").addEventListener("input", calculateMyTool);
  document.getElementById("input2").addEventListener("input", calculateMyTool);

  // 2. Run Defaults
  calculateMyTool();

  // 3. Enable URL Sharing (if helper exists)
  if (typeof makeShareable === "function") {
    makeShareable(calculateMyTool);
  }
}

// Loader Logic (Handles both refresh and navigation)
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initMyTool);
} else {
  initMyTool();
}
```

---

## 3. Style Guidelines (SCSS)

**File:** `assets/scss/tools/_calculator.scss`

Do not write new CSS unless absolutely necessary. Use these existing classes:

- **Wrapper:** `.calculator-wrapper` (Handles dark mode, shadows, and borders)
- **Header:** `.calculator-sticky-header` (Keeps results visible on mobile)
- **Result Text:** `.calculator-main-display h1` (Big text, auto-adapts color)
- **Section Headers:** `.calculator-section-head` (Uppercase, dividers)
- **Input Groups:** `.calculator-field` (Handles labels, inputs, focus states)
- **Badges:** `.upside-badge` (Green/Red text for secondary stats)

### Color Variables (REQUIRED - No Hardcoded Colors)

Always use the site's color variables defined in `assets/scss/style.scss`:

- **Light Mode:** `$light-background`, `$light-text`, `$light-icon`
- **Dark Mode:** `$dark-background`, `$dark-text`, `$dark-icon`
- **Buttons:** `$button-color-dark`, `$button-color-light`

**Never use hardcoded colors like `#ef4444`, `#ffffff`, `#000000`, etc.** Always include `@media (prefers-color-scheme: dark)` media queries to support both light and dark themes.

Example:

```scss
.my-component {
  background-color: $light-background;
  color: $light-text;
  border-color: $light-icon;

  @media (prefers-color-scheme: dark) {
    background-color: $dark-background;
    color: $dark-text;
    border-color: $dark-icon;
  }
}
```

## 4. Common Pitfalls to Avoid

1.  **Duplicate IDs:** Ensure all input IDs are unique within the page.
2.  **Hardcoded Colors:** Never use `color: black` or `white` in inline styles. Use the standard classes so Dark Mode works automatically.
3.  **Missing `shareable.js`:** If you forget to include the shareable script in the HTML, URL params won't work.
4.  **Inline Event Handlers:** Avoid `oninput="..."` attributes in HTML. Use `addEventListener()` in your JavaScript instead (required for CSP compliance).
5.  **Inline JS:** Avoid `<script>` tags inside the HTML. Keep logic in `assets/js/`.
6.  **Inline Styles:** Avoid inline styles in the HTML. Use the standard classes for styling or add classes via `assets/scss/tools/_calculator.scss`.

---

# Standard Operating Procedure for Chart Tools

Chart tools are reusable, interactive line charts with optional quartile bands. These tools use HTMX for interactivity and Chart.js for rendering. The `LineChart` templ component outputs data attributes that are consumed by an external JavaScript file—this approach ensures CSP compliance without inline scripts.

## CSP Compliance

✅ **No inline event handlers** – Use `hx-trigger` attributes instead of `onclick`  
✅ **No inline styles** – Use CSS classes from `_chart-tools.scss`  
✅ **No inline scripts** – Chart initialization via external `assets/js/chart-init.js`  
✅ **Data-driven rendering** – Chart.js setup passed via `data-chart` attributes

## 1. File Structure

Each chart tool requires four files:

1.  **Content:** `content/tools/[tool-name].md` (Defines metadata & URL)
2.  **Layout:** `layouts/tools/[tool-name].html` (HTMX-powered HTML structure—copy from msindex)
3.  **Handler:** `internal/handlers/[tool-name].go` (Data fetching and calculation logic)
4.  **Templ:** `internal/templates/linechart.templ` (Generic `LineChart` component—shared across all chart tools)

The generic `LineChart` templ component (in `internal/templates/linechart.templ`) outputs chart config as `data-chart` attributes. The external `assets/js/chart-init.js` file handles all initialization and reinitializes on HTMX swaps.

---

## 2. Step-by-Step Implementation Guide

### Step 1: Create the Content File

**File:** `content/tools/[tool-name].md`

```markdown
---
title: "Tool Name"
description: "A short description of what this metric represents."
layout: "[tool-name]"
tool_type: "chart"
---

### About [Tool Name]

Detailed explanation of the metric, calculation method, and data sources.

**Data Source:** U.S. Federal Reserve Economic Data (FRED)

- Series 1: SERIES_ID_1
- Series 2: SERIES_ID_2
- Frequency: Quarterly
```

### Step 2: Create the Handler

**File:** `internal/handlers/[tool-name].go`

The handler must:

1. Parse `range` query param (default: "max", options: "1y", "5y", "10y")
2. Parse `quartiles` query param (present = true, absent = false)
3. Fetch data from FRED API or other source
4. Convert to `LineChartData` format
5. Call `templates.LineChart()`

```go
package handlers

import (
	"bytes"
	"net/http"
	"github.com/shanehull/shanehull.com/internal/templates"
)

func MyChartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse query parameters
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	showQuartiles := r.URL.Query().Has("quartiles")

	// 2. Fetch data (implement your data fetching logic)
	data := fetchMyData(rangeParam, showQuartiles)

	// 3. Convert to LineChartData
	chartData := make([]templates.LineChartData, len(data))
	for i, d := range data {
		chartData[i] = templates.LineChartData{
			Date:      d.DateString,
			Value:     d.MetricValue,
			Quartile1: d.Q1,
			Quartile3: d.Q3,
		}
	}

	// 4. Render using generic LineChart template
	options := map[string]string{
		"mainLabel":  "Tool Display Name",
		"yAxisLabel": "Metric Unit",
	}
	component := templates.LineChart("chart-canvas", chartData, showQuartiles, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// Implement your data fetching and calculation logic here
func fetchMyData(rangeParam string, showQuartiles bool) []MyData {
	// TODO: Fetch from FRED API or database
	// TODO: If showQuartiles is true, calculate Q1 and Q3
	return []MyData{}
}
```

### Step 3: Create the Layout

**File:** `layouts/tools/[tool-name].html`

Copy this template exactly from `layouts/tools/msindex.html`. The layout is entirely generic and requires zero modifications.

**⚠️ Important:** Do NOT include a separate htmx script tag—htmx is already loaded globally on every page via the site's base template. Including it twice causes conflicts.

```html
{{ define "main" }}
<main class="container">
  <h1>{{ .Title }}</h1>
  <p class="description"><i>{{ .Description }}</i></p>
  <hr />
  <br />

  <div class="tool-instructions">{{ .Content }}</div>

  <div class="chart-tool-wrapper">
    <div class="chart-controls">
      <div class="control-group">
        <label>Time Range:</label>
        <div class="button-group">
          <label class="radio-label">
            <input
              type="radio"
              name="range"
              value="max"
              checked
              hx-get="/[tool-name]/chart"
              hx-target="#chart-inner"
              hx-include="[name=quartiles]"
              hx-swap="innerHTML"
              hx-trigger="change"
            />
            Max
          </label>
          <label class="radio-label">
            <input
              type="radio"
              name="range"
              value="10y"
              hx-get="/[tool-name]/chart"
              hx-target="#chart-inner"
              hx-include="[name=quartiles]"
              hx-swap="innerHTML"
              hx-trigger="change"
            />
            10 Year
          </label>
          <label class="radio-label">
            <input
              type="radio"
              name="range"
              value="5y"
              hx-get="/[tool-name]/chart"
              hx-target="#chart-inner"
              hx-include="[name=quartiles]"
              hx-swap="innerHTML"
              hx-trigger="change"
            />
            5 Year
          </label>
          <label class="radio-label">
            <input
              type="radio"
              name="range"
              value="1y"
              hx-get="/[tool-name]/chart"
              hx-target="#chart-inner"
              hx-include="[name=quartiles]"
              hx-swap="innerHTML"
              hx-trigger="change"
            />
            1 Year
          </label>
        </div>
      </div>

      <div class="control-group">
        <label class="checkbox-label">
          <input
            type="checkbox"
            name="quartiles"
            checked
            hx-get="/[tool-name]/chart"
            hx-target="#chart-inner"
            hx-include="[name=range]"
            hx-swap="innerHTML"
            hx-trigger="change"
          />
          Show Quartiles
        </label>
      </div>
    </div>

    <div class="chart-container">
      <canvas id="chart-canvas"></canvas>
      <div
        id="chart-inner"
        hx-get="/[tool-name]/chart?range=max&quartiles=on"
        hx-trigger="load"
        hx-swap="innerHTML"
      ></div>
    </div>
  </div>

  <br />
  <hr />
  <br />
  <div class="center-items">
    <a href="/tools/" class="unchanging-link back-link"><- back to tools</a>
  </div>
</main>

<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-zoom@2.1.0/dist/chartjs-plugin-zoom.min.js"></script>

{{ $script := resources.Get "js/chart-init.js" | minify | fingerprint }}
<script src="{{ $script.RelPermalink }}"></script>

{{ end }}
```

### Step 4: Create Download Handlers

You need three handlers: one for the chart rendering, one for generating download links, and two for returning data (JSON and CSV).

**Downloads Handler** (generates HTMX-swapped links):

```go
func MyChartDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	showQuartiles := r.URL.Query().Has("quartiles")

	// Use the reusable templ component - just pass your tool name
	component := templates.ChartDownloads("[tool-name]", rangeParam, showQuartiles)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}
```

**Data Handlers** (JSON and CSV) follow the same pattern as the chart handler—fetch data, accept `range` and `quartiles` params, and return files with proper headers.

### Step 5: Register the Handlers

**File:** `cmd/server/main.go`

Add all handlers to the mux:

```go
mux.HandleFunc(
	"/[tool-name]/chart",
	middleware.CORS(
		http.HandlerFunc(
			handlers.[ToolName]Handler,
		),
		allowedOrigin,
	),
)

mux.HandleFunc(
	"/[tool-name]/downloads",
	middleware.CORS(
		http.HandlerFunc(
			handlers.[ToolName]DownloadsHandler,
		),
		allowedOrigin,
	),
)

mux.HandleFunc(
	"/[tool-name]/data",
	middleware.CORS(
		http.HandlerFunc(
			handlers.[ToolName]DataHandler,
		),
		allowedOrigin,
	),
)

mux.HandleFunc(
	"/[tool-name]/data.csv",
	middleware.CORS(
		http.HandlerFunc(
			handlers.[ToolName]CSVHandler,
		),
		allowedOrigin,
	),
)
```

---

## 3. LineChartData Format

All chart tools use this generic data structure:

```go
type LineChartData struct {
	Date      string  `json:"date"`      // ISO date: "2024-01-01"
	Value     float64 `json:"value"`     // Main metric value
	Quartile1 float64 `json:"quartile1"` // Q1 (25th percentile)
	Quartile3 float64 `json:"quartile3"` // Q3 (75th percentile)
}
```

If quartiles are not needed, set `Q1` and `Q3` to `0`.

---

## 4. How the LineChart Templ Works

The `LineChart` templ function outputs a `<div>` element with a `data-chart` attribute containing JSON-serialized chart configuration. This is consumed by `assets/js/chart-init.js`.

**Key Points:**

- The templ does NOT output inline `<script>` tags
- Chart config is passed as JSON via the `data-chart` attribute
- `initChartFromData()` is called on page load and after each HTMX swap via the event listener in `chart-init.js`
- The function destroys any existing chart instance before creating a new one

Example output:

```html
<div
  data-chart='{"labels":["2024-01","2024-02"],"datasets":[...],"yAxisLabel":"Value","canvasId":"chart-canvas"}'
></div>
```

---

## 5. Key Features (Automatic)

✅ **Dark mode support** – Uses site color variables (`$light-icon`, `$dark-icon`)  
✅ **HTMX interactivity** – Controls trigger API requests, swapped content reinitializes chart  
✅ **Time range selection** – Max, 10Y, 5Y, 1Y buttons  
✅ **Optional quartile bands** – Toggle with checkbox  
✅ **Chart zoom/pan** – Drag to select time range, scroll wheel to zoom, pinch on mobile (via chartjs-plugin-zoom)  
✅ **Chart destruction/recreation** – Handles HTMX swaps correctly  
✅ **Responsive design** – Works on mobile and desktop  
✅ **CSP compliant** – No inline scripts, all JS external

---

## 6. Download UI in Layout

Add this section to your template to display download links that update when controls change:

```html
<div
  id="chart-downloads"
  hx-get="/[tool-name]/downloads"
  hx-include="[name=range],[name=quartiles]"
  hx-trigger="load, change from:.chart-controls input"
  hx-swap="innerHTML"
></div>
```

The `/[tool-name]/downloads` handler (using the reusable `ChartDownloads` templ component) returns HTML links with proper filenames and `hx-boost="false"` to bypass HTMX. No custom code needed—just the pattern from Step 4.

## 7. Common Pitfalls to Avoid

1. **Missing FRED API key:** Store in `FRED_API_KEY` env var
2. **Wrong query param names:** Use `range` and `quartiles` exactly (lowercase)
3. **Forgetting `templates.LineChart` import:** Add to handler file
4. **Not destroying old chart:** Chart.js destruction handled in `assets/js/chart-init.js` via `window.lineChartInstance.destroy()`
5. **Download links intercepted by HTMX:** Add `hx-boost="false"` to download links
6. **Downloads not reflecting UI state:** Use `hx-include="[name=range],[name=quartiles]"` to send form values
7. **Trigger syntax:** Use `change from:.chart-controls input` to listen to control changes on a parent
8. **Hardcoded endpoint URLs:** Always use relative paths like `/[tool-name]/chart`
9. **Forgetting to include chart-init.js:** The external script is required for CSP compliance and chart initialization. It listens to `htmx:afterSwap` events automatically.
