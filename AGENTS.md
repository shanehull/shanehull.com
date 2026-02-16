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
          <input
            type="number"
            id="input1"
            value="100"
          />
        </div>
        <div class="calculator-field">
          <label>Input Label (%)</label>
          <input
            type="number"
            id="input2"
            value="5"
          />
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
