# Agents Guidelines for This Repository

## Engineering Principles: A Philosophy of Software Design

To maintain high velocity and low technical debt, all agents (human or AI) must adhere to the principles outlined in _A Philosophy of Software Design_ by John Ousterhout. Our primary goal is **Complexity Management**.

---

### 1. Strategic vs. Tactical Programming

- **Strategic Mindset:** Working code is not enough. The primary goal is a great design. High-quality design is a proactive investment that prevents future slowdowns.
- **Avoid Tactical Tornadoes:** Do not take shortcuts to finish a feature faster if it introduces complexity. A "quick fix" is often a permanent debt.

### 2. Modules Should Be Deep

- **Deep Modules:** The best modules provide powerful functionality through a simple, narrow interface. They hide significant implementation complexity.
- **Shallow Modules:** Avoid modules where the interface is complex relative to the small amount of work it performs.
- **Goal:** Maximize the _benefit_ of the module while minimizing its _cost_ (interface complexity).

### 3. Information Hiding

- **Encapsulation:** Each module should encapsulate specific design decisions.
- **Information Leakage:** This occurs when a design decision is reflected in multiple modules. If changing one class requires changing another, the design is leaking.
- **Pull Complexity Downwards:** It is better for a module's implementation to be complex than for its interface to be complex. The **developer of the module** should suffer complexity so the **users of the module** don't have to.

### 4. General-Purpose vs. Special-Purpose

- **"Somewhat General-Purpose":** Design modules to be flexible enough to handle current needs and potential near-future needs without being overly tied to a single specific use case.
- **Clean API:** A general-purpose interface is usually simpler and deeper than a specialized one.

### 5. Define Errors Out of Existence

- **Minimize Exceptions:** Exceptions contribute significantly to complexity. Design APIs so that "edge cases" are handled naturally by the normal flow (e.g., returning an empty list instead of throwing `null` or an error).

### 6. Comments and Documentation

- **Describe what is NOT obvious:** Comments should capture information that was in the mind of the designer but isn't clear from the code itself.
- **Different Levels:**
  - **Interface:** Describe the _intent_ and _results_, not the implementation.
  - **Implementation:** Explain _what_ the code is doing and _why_, specifically for complex logic.

---

### Design Red Flags

Check for these during code reviews and generation:

| Red Flag                 | Description                                                                   |
| :----------------------- | :---------------------------------------------------------------------------- |
| **Information Leakage**  | A change in one place requires coordinated changes elsewhere.                 |
| **Temporal Coupling**    | Methods must be called in a specific order that isn't enforced by the API.    |
| **Pass-Through Methods** | A method does nothing but call another method with a similar signature.       |
| **Cognitive Load**       | The amount of "mental state" a dev must hold to understand a single function. |
| **Shallow Module**       | The interface is nearly as complex as the logic it hides.                     |

### 🧠 Core Mental Models for Engineering

#### 1. The Inversion Principle

- **Definition:** Instead of asking "How do I make this feature work?", ask "What would make this feature fail, be unmaintainable, or slow down the system?"
- **Application:** Identify the "anti-goals" first. By avoiding the things that cause failure (spaghetti code, tight coupling, global state), you arrive at a better design by default.

#### 2. First Principles Thinking

- **Definition:** Deconstruct a problem into its fundamental truths rather than reasoning by analogy ("we've always done it this way").
- **Application:** Before using a heavy framework or library because it's "industry standard," evaluate if the core problem can be solved with a simpler, more direct implementation.

#### 3. Second-Order Thinking

- **Definition:** Ask "And then what?" for every design decision.
- **Application:** A "quick fix" might solve a bug today (first-order), but it might create a rigid dependency that prevents a major refactor six months from now (second-order).

#### 4. Occam’s Razor for APIs

- **Definition:** The simplest explanation or solution is usually the right one.
- **Application:** If you are debating between two designs, choose the one with the fewest assumptions and moving parts.

## Stylesheet Architecture

### Overview

- **`assets/scss/style.scss`** - All site styles (loaded on every page) ~25 KiB
  - Contains color and theme variables at the top
  - Imports tool-specific styles and fonts

### File Structure

```
assets/scss/
  _fonts.scss              ← Font declarations
  style.scss               ← Main stylesheet with variables at top
  tools/
    _calculator.scss       ← Imported by style.scss
    _chart-tools.scss      ← Imported by style.scss
    _options-pnl.scss      ← Imported by style.scss
    _list.scss             ← Imported by style.scss
```

### Color Variables

Defined at the top of `style.scss`:

- Light mode: `$light-background`, `$light-text`, `$light-icon`
- Dark mode: `$dark-background`, `$dark-text`, `$dark-icon`
- Buttons: `$button-color-dark`, `$button-color-light`

### When Creating New Tool Styles

1. Create `assets/scss/tools/_newtool.scss`
2. Import it in `assets/scss/style.scss`
3. Use color variables (defined in style.scss)
4. Follow dark mode pattern: `@media (prefers-color-scheme: dark) { ... }`

---

## Code Standards

### Go

All Go code must pass `golangci-lint run ./...` without warnings. Key requirements:

- **Error handling**: Check all error return values. Use blank identifiers `_` to explicitly ignore errors when appropriate:
  ```go
  _, _ = w.Write(buf.Bytes())  // Ignore write errors in error handlers
  defer func() {
      _ = resp.Body.Close()
  }()
  ```
- **Blank imports**: Remove unused imports
- **Unused variables**: All assigned variables must be used or discarded with `_`
- **Formatting**: Run `gofmt` on all edited files before committing
- Run linting before committing: `golangci-lint run ./...`

### JavaScript

All JavaScript files are formatted with Prettier. Key rules:

- **Single quotes**: Use single quotes for strings
- **Semicolons**: Include semicolons at end of statements
- **Indentation**: 2 spaces
- **Format before committing**: Run Prettier on all `assets/js/*.js` files
- **Configuration**: Prettier via none-ls with `--single-quote` flag

---

## Standard Operating Procedure for Calculators

This document outlines the strict architecture for creating new interactive calculators in this Hugo project. All new tools **must** follow this pattern to ensure consistency, correct styling, and URL shareability.

### 1. File Structure

Each calculator requires exactly four files in these specific locations:

1.  **Content:** `content/tools/[tool-name].md` (Defines metadata & URL)
2.  **Layout:** `layouts/tools/[tool-name].html` (The HTML structure)
3.  **Logic:** `assets/js/[tool-name].js` (The math, interaction, and shareable logic)
4.  **Style:** `assets/scss/tools/_calculator.scss` (Shared styles - rarely needs editing)

---

### 2. Step-by-Step Implementation Guide

#### Step 1: Create the Content File

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

#### Instructions

(Optional) Markdown content here appears below the calculator as a user guide.
```

#### Step 2: Create the Layout (HTML)

**File:** `layouts/tools/[tool-name].html`
**Critical:** Use the standard CSS classes (`calculator-wrapper`, `calculator-row`, `calculator-field`) to inherit the site's theme automatically.

**Layout structure:**

1. `<h1>{{ .Title }}</h1>` – Page title only
2. `<div class="tool-instructions">{{ .Content }}</div>` – Instructions from markdown
3. `<div class="calculator-wrapper">` – Calculator container
4. Sticky header with results
5. Calculator body with inputs

**Sticky header structure:**

- Use `calculator-stats-grid` with `calculator-stat-item` for secondary metrics
- Use `calculator-main-display` for primary result with optional `upside-badge` for status

```html
{{ define "main" }}
<main class="container">
  <h1>{{ .Title }}</h1>
  <div class="tool-instructions">{{ .Content }}</div>

  <div class="calculator-wrapper">
    <div class="calculator-sticky-header">
      <div class="calculator-stats-grid">
        <div class="calculator-stat-item">
          <span class="label">Secondary Metric</span>
          <span id="outSecondary" class="val">$0.00</span>
        </div>
      </div>
      <div class="calculator-main-display">
        <span class="label">Primary Result</span>
        <h1 id="outResult">$0.00</h1>
        <div id="outStatus" class="upside-badge">0%</div>
      </div>
    </div>

    <div class="calculator-body">
      <h3 class="calculator-section-head">Input Section Name</h3>

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

{{ $shareable := resources.Get "js/shareable.js" | minify | fingerprint "sha384"
}}
<script
  src="{{ $shareable.RelPermalink }}"
  integrity="{{ $shareable.Data.Integrity }}"
  defer
></script>

{{ $script := resources.Get "js/[tool-name].js" | minify | fingerprint "sha384"
}}
<script
  src="{{ $script.RelPermalink }}"
  integrity="{{ $script.Data.Integrity }}"
  defer
></script>
{{ end }}
```

#### Step 3: Create the Logic (JavaScript)

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

### 3. Style Guidelines (SCSS)

**File:** `assets/scss/tools/_calculator.scss`

Do not write new CSS unless absolutely necessary. Use these existing classes:

- **Wrapper:** `.calculator-wrapper` (Handles dark mode, shadows, and borders)
- **Header:** `.calculator-sticky-header` (Keeps results visible on mobile)
- **Result Text:** `.calculator-main-display h1` (Big text, auto-adapts color)
- **Section Headers:** `.calculator-section-head` (Plain text, no numbering. E.g., "Assumptions" not "1. Assumptions")
- **Input Groups:** `.calculator-field` (Handles labels, inputs, focus states). **Every `<label>` must have a `for` attribute matching the input's `id`**
- **Badges:** `.upside-badge` (Green/Red text for secondary stats)

#### Color Variables (REQUIRED - No Hardcoded Colors)

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

### 4. Common Pitfalls to Avoid

1.  **Duplicate IDs:** Ensure all input IDs are unique within the page.
2.  **Hardcoded Colors:** Never use `color: black` or `white` in inline styles. Use the standard classes so Dark Mode works automatically.
3.  **Missing `shareable.js`:** If you forget to include the shareable script in the HTML, URL params won't work.
4.  **Inline Event Handlers:** Avoid `oninput="..."` attributes in HTML. Use `addEventListener()` in your JavaScript instead (required for CSP compliance).
5.  **Inline JS:** Avoid `<script>` tags inside the HTML. Keep logic in `assets/js/`.
6.  **Inline Styles:** Avoid inline styles in the HTML. Use the standard classes for styling or add classes via `assets/scss/tools/_calculator.scss`.

---

## Standard Operating Procedure for Chart Tools

Chart tools are reusable, interactive line charts with optional quartile bands. These tools use HTMX for interactivity and Chart.js for rendering. The `LineChart` templ component outputs data attributes that are consumed by an external JavaScript file—this approach ensures CSP compliance without inline scripts.

### CSP Compliance

✅ **No inline event handlers** – Use `hx-trigger` attributes instead of `onclick`  
✅ **No inline styles** – Use CSS classes from `_chart-tools.scss`  
✅ **No inline scripts** – Chart initialization via external `assets/js/chart-init.js`  
✅ **Data-driven rendering** – Chart.js setup passed via `data-chart` attributes

### 1. File Structure

Each chart tool requires four files:

1.  **Content:** `content/tools/[tool-name].md` (Defines metadata & URL)
2.  **Layout:** `layouts/tools/[tool-name].html` (HTMX-powered HTML structure—copy from msindex)
3.  **Handler:** `internal/handlers/[tool-name].go` (Data fetching and calculation logic)
4.  **Templ:** `internal/templates/linechart.templ` (Generic `LineChart` component—shared across all chart tools)

#### Shared Packages

For FRED-based chart tools, two reusable internal packages are available:

- **`internal/fred`** – FRED API client for fetching economic data series
  - `FetchSeries(seriesID, opts)` – Fetches observations with configurable options
  - `FetchOptions` struct – Configure observation_start, observation_end, frequency, units, etc.
- **`internal/charts`** – Shared chart utilities
  - `CalculateRangeStart(rangeParam)` – Converts UI range params ("1y", "5y", "max", etc.) to date filters

The generic `LineChart` templ component (in `internal/templates/linechart.templ`) outputs chart config as `data-chart` attributes. The external `assets/js/chart-init.js` file handles all initialization and reinitializes on HTMX swaps.

---

### 2. Step-by-Step Implementation Guide

#### Step 1: Create the Content File

**File:** `content/tools/[tool-name].md`

**Important:** Setting `tool_type: "chart"` automatically loads Chart.js and the zoom plugin in baseof.html. Do NOT include these scripts in your layout.

```markdown
---
title: "Tool Name"
description: "A short description of what this metric represents."
layout: "[tool-name]"
tool_type: "chart"
---

#### About [Tool Name]

Detailed explanation of the metric, calculation method, and data sources.

**Data Source:** U.S. Federal Reserve Economic Data (FRED)

- Series 1: SERIES_ID_1
- Series 2: SERIES_ID_2
- Frequency: Quarterly
```

#### Step 2: Create the Handler

**File:** `internal/handlers/[tool-name].go`

The handler must:

1. Parse `range` query param (default: "max", options: "1y", "5y", "10y", "20y", "50y")
2. Parse `quartiles` query param (present = true, absent = false)
3. Use `internal/charts.CalculateRangeStart(rangeParam)` to get start date
4. Use `internal/fred.FetchSeries(seriesID, opts)` to fetch FRED data
5. Convert to `LineChartData` format
6. Cache results using `internal/cache` for 24 hours
7. Call `templates.LineChart()`

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
		"mainLabel":    "Tool Display Name",
		"yAxisLabel":   "Metric Unit",
		"showQuartiles": "false",
		"showAverage":  "false",
	}
	if showQuartiles {
		options["showQuartiles"] = "true"
	}
	// Set showAverage to "true" if your tool wants to display the average line
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

#### Caching Chart Data

For better performance, cache chart calculations using the `internal/cache` package. Example pattern for your handler:

```go
package handlers

import (
	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/templates"
	"time"
)

const cacheTTL = 24 * time.Hour

var chartCache = cache.New()

func MyChartHandler(w http.ResponseWriter, r *http.Request) {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	showQuartiles := r.URL.Query().Has("quartiles")

	// Generate cache key
	cacheKey := fmt.Sprintf("mychart:%s:%v", rangeParam, showQuartiles)

	// Check cache first
	if cached, found := chartCache.Get(cacheKey); found {
		chartData := cached.([]templates.LineChartData)
		// Render and return cached data
		renderChart(w, r, chartData, showQuartiles)
		return
	}

	// Fetch and calculate data
	data := fetchMyData(rangeParam)
	chartData := convertToLineChartData(data, showQuartiles)

	// Cache the result for 24 hours
	chartCache.Set(cacheKey, chartData, cacheTTL)

	// Render and return
	renderChart(w, r, chartData, showQuartiles)
}
```

**Cache key format:** `toolname:rangeParam:showQuartiles`
**TTL:** 24 hours (configurable via constant)
**Note:** Each unique combination of range and quartiles settings gets its own cache entry.

#### Step 3: Create the Layout

**File:** `layouts/tools/[tool-name].html`

Copy this template exactly from `layouts/tools/msindex.html`. The layout is entirely generic and requires zero modifications.

**⚠️ Important:** Do NOT include a separate htmx script tag—htmx is already loaded globally on every page via the site's base template. Including it twice causes conflicts.

**⚠️ Important:** Chart.js and the internal init script are automatically loaded in `baseof.html` for any page with `tool_type: "chart"`. Do NOT manually load these scripts in individual tool layouts.

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
              hx-trigger="load, change"
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
      <div id="chart-inner"></div>
    </div>

    <div
      id="chart-downloads"
      hx-get="/[tool-name]/downloads"
      hx-include="[name=range],[name=quartiles]"
      hx-trigger="load, change from:[name=range], change from:[name=quartiles]"
      hx-swap="innerHTML"
    ></div>
  </div>

  <br />
  <hr />
  <br />
  <div class="center-items">
    <a href="/tools/" class="unchanging-link back-link"><- back to tools</a>
  </div>
</main>

{{ end }}
```

#### Step 4: Create Download Handlers

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

#### Step 5: Register the Handlers

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

### 3. Chart Overlay Options

The `LineChart` template supports optional overlays via the `options` map:

- **`showQuartiles`**: Set to `"true"` to display Q1 (25th percentile) and Q3 (75th percentile) lines. Defaults to `"false"`. Only shows if quartile data is provided.
- **`showAverage`**: Set to `"true"` to display a horizontal average line. Defaults to `"false"`. Only shows if average data is provided.

Example:

```go
options := map[string]string{
	"mainLabel":    "My Chart",
	"yAxisLabel":   "Values",
	"showQuartiles": "true",  // Show quartile bands
	"showAverage":  "true",   // Show average line
}
```

Both overlays are optional and can be independently controlled per chart.

---

### 4. LineChartData Format

All chart tools use this generic data structure:

```go
type LineChartData struct {
	Date      string  `json:"date"`           // ISO date: "2024-01-01"
	Value     float64 `json:"value"`          // Main metric value
	Quartile1 float64 `json:"quartile1,omitempty"` // Q1 (25th percentile)
	Quartile3 float64 `json:"quartile3,omitempty"` // Q3 (75th percentile)
	Average   float64 `json:"average,omitempty"`   // Average line
}
```

Use `omitempty` to exclude fields from JSON when they're zero-valued. Only set fields that your tool needs.

---

### 5. How the LineChart Templ Works

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

### 6. Key Features (Automatic)

✅ **Dark mode support** – Uses site color variables (`$light-icon`, `$dark-icon`)  
✅ **HTMX interactivity** – Controls trigger API requests, swapped content reinitializes chart  
✅ **Time range selection** – Max, 10Y, 5Y, 1Y buttons  
✅ **Optional quartile bands** – Toggle with checkbox  
✅ **Chart zoom/pan** – Drag to select time range, scroll wheel to zoom, pinch on mobile (via chartjs-plugin-zoom)  
✅ **Chart destruction/recreation** – Handles HTMX swaps correctly  
✅ **Responsive design** – Works on mobile and desktop  
✅ **CSP compliant** – No inline scripts, all JS external

---

### 7. Download UI in Layout

Add this section to your template to display download links that update when controls change:

```html
<div
  id="chart-downloads"
  hx-get="/[tool-name]/downloads"
  hx-include="[name=range],[name=overlayParam]"
  hx-trigger="load, change from:.chart-controls input"
  hx-swap="innerHTML"
></div>
```

The `/[tool-name]/downloads` handler uses the reusable `ChartDownloads` templ component:

```go
component := templates.ChartDownloads("[tool-name]", rangeParam, "overlayParamName", showOverlay)
```

Pass your specific overlay parameter name (e.g., "quartiles", "average") and the template handles building the correct download links.

### 8. Common Pitfalls to Avoid

1. **Missing FRED API key:** Store in `FRED_API_KEY` env var
2. **Wrong query param names:** Use `range` exactly (lowercase) and your custom overlay param name
3. **Forgetting `templates.LineChart` import:** Add to handler file
4. **Not destroying old chart:** Chart.js destruction handled in `assets/js/chart-init.js` via `window.lineChartInstance.destroy()`
5. **Download links intercepted by HTMX:** Add `hx-boost="false"` to download links
6. **Downloads not reflecting UI state:** Use `hx-include="[name=range],[name=overlayParam]"` to send form values
7. **Trigger syntax:** Use `change from:.chart-controls input` to listen to control changes on a parent
8. **Hardcoded endpoint URLs:** Always use relative paths like `/[tool-name]/chart`
9. **Forgetting to include chart-init.js:** The external script is required for CSP compliance and chart initialization. It listens to `htmx:afterSwap` events automatically.
10. **HTMX includes for missing controls:** If you remove a control (e.g., quartiles checkbox), remove it from `hx-include` attributes too. Orphaned form fields in HTMX will cause requests to fail silently.
11. **Overlay parameter names:** Use `ChartDownloads(toolName, rangeParam, overlayParamName, showOverlay)` and pass your specific overlay parameter name (e.g., "quartiles" for msindex, "average" for buffett-indicator). The template handles building the correct download links dynamically.
12. **hx-boost navigation:** Add `hx-trigger="load delay:200ms, change"` to the default "max" radio button for the chart, and `hx-trigger="load, change from:..."` on the downloads div. These elements are always in the DOM, so `load` fires on both hard refreshes and hx-boost navigation. The 200ms delay ensures `chart-init.js` has loaded and registered the `htmx:afterSwap` listener before the trigger fires, preventing race conditions on slow mobile connections.
