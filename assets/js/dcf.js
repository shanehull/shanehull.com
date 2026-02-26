/**
 * Main DCF Calculation Function
 * Triggers on every input change
 */
function calculateDCF() {
  const initialFCF =
    parseFloat(document.getElementById("initialFCF").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("growthRate").value) / 100 || 0;
  const discountRate =
    parseFloat(document.getElementById("discountRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("yearsToExit").value) || 1;
  const exitMultiple =
    parseFloat(document.getElementById("exitMultiple").value) || 0;
  const currentPrice =
    parseFloat(document.getElementById("currentPrice").value) || 0;

  let pvCashFlows = 0;

  // Calculate PV of cash flows for each year in the holding period
  for (let i = 0; i < yearsToExit; i++) {
    const fcf = initialFCF * Math.pow(1 + growthRate, i);
    const discountFactor = Math.pow(1 + discountRate, i + 1);
    pvCashFlows += fcf / discountFactor;
  }

  // Calculate exit FCF (at year numYears) and terminal value
  const exitFCF = initialFCF * Math.pow(1 + growthRate, yearsToExit);
  const terminalValue = exitFCF * exitMultiple;
  const discountFactorExit = Math.pow(1 + discountRate, yearsToExit);
  const pvExit = terminalValue / discountFactorExit;

  // Intrinsic value
  const intrinsicValue = pvCashFlows + pvExit;

  // Expected return: growth rate + FCF yield
  let expectedReturn = 0;
  if (currentPrice > 0) {
    const fcfYield = initialFCF / currentPrice;
    expectedReturn = growthRate + fcfYield;
  }

  // Upside/Downside
  const upsideDownside = (intrinsicValue - currentPrice) / currentPrice;

  // Update UI
  document.getElementById("outIntrinsicValue").innerText =
    formatCurrency(intrinsicValue);
  document.getElementById("outPVCashFlows").innerText =
    formatCurrency(pvCashFlows);
  document.getElementById("outPVExit").innerText = formatCurrency(pvExit);
  document.getElementById("outExpectedReturn").innerText =
    formatPercent(expectedReturn) + "% /y";

  // Update upside/downside badge
  const upsideElement = document.getElementById("outUpsideDownside");
  upsideElement.innerText = formatPercent(upsideDownside) + "%";
  if (upsideDownside > 0) {
    upsideElement.style.color = "#22c55e";
  } else if (upsideDownside < 0) {
    upsideElement.style.color = "#ef4444";
  } else {
    upsideElement.style.color = "";
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

// Helper: Percent Formatting
function formatPercent(num) {
  return (num * 100).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

/**
 * Universal Initialization Pattern
 * Handles:
 * 1. Hard Refresh (DOMContentLoaded)
 * 2. Soft Navigation (Turbo/HTMX)
 * 3. URL Parameter Sharing (shareable.js)
 */
function initDCF() {
  if (!document.getElementById("initialFCF")) return;

  document.getElementById("initialFCF").addEventListener("input", calculateDCF);
  document.getElementById("growthRate").addEventListener("input", calculateDCF);
  document
    .getElementById("discountRate")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("yearsToExit")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("exitMultiple")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("currentPrice")
    .addEventListener("input", calculateDCF);

  calculateDCF();

  if (typeof makeShareable === "function") {
    makeShareable(calculateDCF);
  }
}

// Loader Logic (Handles both refresh and navigation)
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initDCF);
} else {
  initDCF();
}
