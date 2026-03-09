/**
 * Main DCF Calculation Function
 * Handles exit multiple, 2-stage growth, and dividend discount models
 */
function calculateDCF() {
  const modelType = document.querySelector('input[name="dcfModel"]:checked').value;

  if (modelType === "exit") {
    calculateDCFExit();
  } else if (modelType === "perpetual") {
    calculateDCFPerpetual();
  } else if (modelType === "ddm") {
    calculateDDM();
  }
}

/**
 * DCF Exit Multiple Model
 */
function calculateDCFExit() {
  const initialFCF =
    parseFloat(document.getElementById("exitInitialFCF").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("exitGrowthRate").value) / 100 || 0;
  const discountRate =
    parseFloat(document.getElementById("discountRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("exitYearsToExit").value) || 1;
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

  // Expected return: IRR of cash flows
  let expectedReturn = calculateIRR(currentPrice, initialFCF, growthRate, yearsToExit, terminalValue);

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

  // Clear perpetual-specific styling
  document.getElementById("outIntrinsicValue").style.color = "";
}

/**
 * DCF Perpetual Growth Model
 * Projects explicit growth period, then assumes perpetual growth thereafter using Gordon Growth Model
 */
function calculateDCFPerpetual() {
  const initialFCF =
    parseFloat(document.getElementById("perpetualInitialFCF").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("perpetualGrowthRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("perpetualYearsToExit").value) || 1;
  const perpetualGrowthRate =
    parseFloat(document.getElementById("perpetualTerminalGrowthRate").value) / 100 || 0;
  const discountRate =
    parseFloat(document.getElementById("discountRate").value) / 100 || 0;
  const currentPrice =
    parseFloat(document.getElementById("currentPrice").value) || 0;

  let pvCashFlows = 0;

  // Calculate PV of cash flows during explicit period
  for (let i = 0; i < yearsToExit; i++) {
    const fcf = initialFCF * Math.pow(1 + growthRate, i);
    const discountFactor = Math.pow(1 + discountRate, i + 1);
    pvCashFlows += fcf / discountFactor;
  }

  // Calculate terminal value at end of explicit period using Gordon Growth Model
  let pvTerminalValue = 0;
  if (discountRate <= perpetualGrowthRate) {
    pvTerminalValue = 0; // Invalid: growth must be less than discount rate
  } else {
    const fcfAtTerminal =
      initialFCF * Math.pow(1 + growthRate, yearsToExit);
    const terminalValue =
      (fcfAtTerminal * (1 + perpetualGrowthRate)) /
      (discountRate - perpetualGrowthRate);
    const discountFactorTerminal = Math.pow(1 + discountRate, yearsToExit);
    pvTerminalValue = terminalValue / discountFactorTerminal;
  }

  // Intrinsic value
  const intrinsicValue = pvCashFlows + pvTerminalValue;

  // Terminal value at end of explicit period (before discounting)
  const terminalValueAtEnd =
    discountRate > perpetualGrowthRate
      ? (initialFCF * Math.pow(1 + growthRate, yearsToExit) *
          (1 + perpetualGrowthRate)) /
        (discountRate - perpetualGrowthRate)
      : 0;

  // Expected return: IRR of cash flows
  let expectedReturn = calculateIRR(currentPrice, initialFCF, growthRate, yearsToExit, terminalValueAtEnd);

  // Upside/Downside
  const upsideDownside = (intrinsicValue - currentPrice) / currentPrice;

  // Update UI
  document.getElementById("outIntrinsicValue").innerText =
    formatCurrency(intrinsicValue);
  document.getElementById("outPVCashFlows").innerText =
    formatCurrency(pvCashFlows);
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

  // Show warning if perpetual growth >= discount rate
  if (perpetualGrowthRate >= discountRate) {
    document.getElementById("outIntrinsicValue").style.color = "#ef4444";
  } else {
    document.getElementById("outIntrinsicValue").style.color = "";
  }
}

/**
 * Dividend Discount Model (DDM)
 * Projects dividends for explicit period, then assumes perpetual growth thereafter
 */
function calculateDDM() {
  const initialDividend =
    parseFloat(document.getElementById("ddmInitialDividend").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("ddmGrowthRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("ddmYearsToExit").value) || 1;
  const terminalGrowthRate =
    parseFloat(document.getElementById("ddmTerminalGrowthRate").value) / 100 || 0;
  const discountRate =
    parseFloat(document.getElementById("discountRate").value) / 100 || 0;
  const currentPrice =
    parseFloat(document.getElementById("currentPrice").value) || 0;

  let pvDividends = 0;

  // Calculate PV of dividends during explicit period
  for (let i = 0; i < yearsToExit; i++) {
    const dividend = initialDividend * Math.pow(1 + growthRate, i);
    const discountFactor = Math.pow(1 + discountRate, i + 1);
    pvDividends += dividend / discountFactor;
  }

  // Calculate terminal value at end of explicit period using Gordon Growth Model
  let pvTerminalValue = 0;
  if (discountRate <= terminalGrowthRate) {
    pvTerminalValue = 0; // Invalid: growth must be less than discount rate
  } else {
    const dividendAtTerminal =
      initialDividend * Math.pow(1 + growthRate, yearsToExit);
    const terminalValue =
      (dividendAtTerminal * (1 + terminalGrowthRate)) /
      (discountRate - terminalGrowthRate);
    const discountFactorTerminal = Math.pow(1 + discountRate, yearsToExit);
    pvTerminalValue = terminalValue / discountFactorTerminal;
  }

  // Intrinsic value
  const intrinsicValue = pvDividends + pvTerminalValue;

  // Terminal value at end of explicit period (before discounting)
  const terminalValueAtEnd =
    discountRate > terminalGrowthRate
      ? (initialDividend * Math.pow(1 + growthRate, yearsToExit) *
          (1 + terminalGrowthRate)) /
        (discountRate - terminalGrowthRate)
      : 0;

  // Expected return: IRR of dividends
  let expectedReturn = calculateIRR(currentPrice, initialDividend, growthRate, yearsToExit, terminalValueAtEnd);

  // Upside/Downside
  const upsideDownside = (intrinsicValue - currentPrice) / currentPrice;

  // Update UI
  document.getElementById("outIntrinsicValue").innerText =
    formatCurrency(intrinsicValue);
  document.getElementById("outPVCashFlows").innerText =
    formatCurrency(pvDividends);
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

  // Show warning if terminal growth >= discount rate
  if (terminalGrowthRate >= discountRate) {
    document.getElementById("outIntrinsicValue").style.color = "#ef4444";
  } else {
    document.getElementById("outIntrinsicValue").style.color = "";
  }
}

/**
 * Calculate IRR using Newton-Raphson method
 * Solves: -initialInvestment + sum(cf_t / (1+r)^t) = 0
 */
function calculateIRR(initialInvestment, initialCashFlow, growthRate, yearsToExit, terminalValue) {
  if (initialInvestment <= 0) return 0;

  // Build cash flow array
  const cashFlows = [-initialInvestment];
  for (let i = 1; i <= yearsToExit; i++) {
    const cf = initialCashFlow * Math.pow(1 + growthRate, i - 1);
    if (i === yearsToExit) {
      cashFlows.push(cf + terminalValue);
    } else {
      cashFlows.push(cf);
    }
  }

  // Newton-Raphson method to find IRR
  let rate = growthRate; // Initial guess
  for (let iteration = 0; iteration < 100; iteration++) {
    let npv = 0;
    let npvDerivative = 0;

    for (let t = 0; t < cashFlows.length; t++) {
      const discountFactor = Math.pow(1 + rate, t);
      npv += cashFlows[t] / discountFactor;
      if (t > 0) {
        npvDerivative -= (t * cashFlows[t]) / Math.pow(1 + rate, t + 1);
      }
    }

    if (Math.abs(npv) < 0.01) break; // Converged
    if (Math.abs(npvDerivative) < 1e-10) break; // Avoid division by zero

    rate = rate - npv / npvDerivative;
    if (rate < -0.99) rate = -0.99; // Prevent negative infinity
  }

  return rate;
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
 * Toggle visibility between model sections
 */
function toggleDCFModel() {
  const modelType = document.querySelector('input[name="dcfModel"]:checked').value;
  
  document.getElementById("exitSection").style.display = 
    modelType === "exit" ? "block" : "none";
  document.getElementById("perpetualSection").style.display = 
    modelType === "perpetual" ? "block" : "none";
  document.getElementById("ddmSection").style.display = 
    modelType === "ddm" ? "block" : "none";
  document.getElementById("exitValueItem").style.display = 
    modelType === "exit" ? "block" : "none";
  
  // Update label based on model
  const pvLabel = document.getElementById("pvFlowsLabel");
  if (modelType === "ddm") {
    pvLabel.innerText = "PV of Dividends";
  } else {
    pvLabel.innerText = "PV of Cash Flows";
  }
  
  calculateDCF();
}

/**
 * Universal Initialization Pattern
 * Handles:
 * 1. Hard Refresh (DOMContentLoaded)
 * 2. Soft Navigation (Turbo/HTMX)
 * 3. URL Parameter Sharing (shareable.js)
 */
function initDCF() {
  if (!document.getElementById("exitInitialFCF")) return;

  // Restore model selection from URL params (before shareable processes other inputs)
  const params = new URLSearchParams(window.location.search);
  if (params.has("dcfModel")) {
    const modelValue = params.get("dcfModel");
    if (modelValue === "perpetual") {
      document.getElementById("modelPerpetual").checked = true;
    } else if (modelValue === "ddm") {
      document.getElementById("modelDDM").checked = true;
    } else {
      document.getElementById("modelExit").checked = true;
    }
  }

  // Model selection listeners
  document
    .getElementById("modelExit")
    .addEventListener("change", toggleDCFModel);
  document
    .getElementById("modelPerpetual")
    .addEventListener("change", toggleDCFModel);
  document
    .getElementById("modelDDM")
    .addEventListener("change", toggleDCFModel);

  // Exit model listeners
  document.getElementById("exitInitialFCF").addEventListener("input", calculateDCF);
  document.getElementById("exitGrowthRate").addEventListener("input", calculateDCF);
  document
    .getElementById("discountRate")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("exitYearsToExit")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("exitMultiple")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("currentPrice")
    .addEventListener("input", calculateDCF);

  // Perpetual model listeners
  document
    .getElementById("perpetualInitialFCF")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("perpetualGrowthRate")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("perpetualYearsToExit")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("perpetualTerminalGrowthRate")
    .addEventListener("input", calculateDCF);

  // DDM model listeners
  document
    .getElementById("ddmInitialDividend")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("ddmGrowthRate")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("ddmYearsToExit")
    .addEventListener("input", calculateDCF);
  document
    .getElementById("ddmTerminalGrowthRate")
    .addEventListener("input", calculateDCF);

  // Load from URL params, then ensure model sections match selection
  toggleDCFModel();
  if (typeof makeShareable === "function") {
    makeShareable(calculateDCF);
  } else {
    calculateDCF();
  }
}

// Loader Logic (Handles both refresh and navigation)
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initDCF);
} else {
  initDCF();
}
