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

  // Clear perpetual-specific styling
  document.getElementById("outIntrinsicValue").style.color = "";
}

/**
 * DCF Perpetual Growth Model
 * Projects explicit growth period, then assumes perpetual growth thereafter using Gordon Growth Model
 */
function calculateDCFPerpetual() {
  const initialFCF =
    parseFloat(document.getElementById("initialFCF").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("growthRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("yearsToExit").value) || 1;
  const perpetualGrowthRate =
    parseFloat(document.getElementById("perpetualGrowthRate").value) / 100 || 0;
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

  // Expected return: growth rate + FCF yield (simplified)
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
    parseFloat(document.getElementById("initialDividend").value) || 0;
  const growthRate =
    parseFloat(document.getElementById("growthRate").value) / 100 || 0;
  const yearsToExit =
    parseInt(document.getElementById("yearsToExit").value) || 1;
  const terminalGrowthRate =
    parseFloat(document.getElementById("perpetualGrowthRate").value) / 100 || 0;
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

  // Expected return: growth rate + dividend yield
  let expectedReturn = 0;
  if (currentPrice > 0) {
    const dividendYield = initialDividend / currentPrice;
    expectedReturn = growthRate + dividendYield;
  }

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
  if (!document.getElementById("initialFCF")) return;

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

  // Perpetual model listeners
  document
    .getElementById("perpetualGrowthRate")
    .addEventListener("input", calculateDCF);

  // DDM model listeners
  document
    .getElementById("initialDividend")
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
