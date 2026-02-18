/**
 * IRR Solver using Newton-Raphson method
 * Finds the discount rate where NPV = 0
 */
function solveIRR(cashflows) {
  const tolerance = 0.0001;
  const maxIterations = 100;
  let rate = 0.1; // Start guess at 10%

  for (let iter = 0; iter < maxIterations; iter++) {
    let npv = 0;
    let npvDerivative = 0;

    for (let i = 0; i < cashflows.length; i++) {
      const cf = cashflows[i];
      const discountFactor = Math.pow(1 + rate, -i);
      npv += cf * discountFactor;
      npvDerivative += (cf * -i * discountFactor) / (1 + rate);
    }

    if (Math.abs(npv) < tolerance) return rate;
    if (npvDerivative === 0) return rate; // Avoid division by zero

    rate = rate - npv / npvDerivative;
  }

  return rate; // Return best guess if doesn't converge
}

function calculateSMB() {
  const price = parseFloat(document.getElementById("price").value) || 0;
  const equityPct =
    parseFloat(document.getElementById("equityPct").value) / 100;
  const intRate = parseFloat(document.getElementById("intRate").value) / 100;
  const term = parseFloat(document.getElementById("term").value) || 0;
  const startSDE = parseFloat(document.getElementById("cashflow").value) || 0;
  const growth = parseFloat(document.getElementById("growth").value) / 100;
  const hold = parseFloat(document.getElementById("hold").value) || 5;
  const exitMult = parseFloat(document.getElementById("exitMult").value) || 0;
  const capex = parseFloat(document.getElementById("capex").value) || 0;
  const corpTax = parseFloat(document.getElementById("corpTax").value) / 100;
  const gainsTax = parseFloat(document.getElementById("gainsTax").value) / 100;

  const debt = price * (1 - equityPct);
  const equity = price * equityPct;

  // Debt service calculation
  let annualDebtService = 0;
  if (intRate > 0 && term > 0) {
    annualDebtService = (debt * intRate) / (1 - Math.pow(1 + intRate, -term));
  } else if (term > 0) {
    annualDebtService = debt / term;
  }

  // Calculate true ROIC with accumulated invested capital
  // Year 1 ROIC
  const year1OperatingProfit = startSDE - capex;
  const year1NOPAT = year1OperatingProfit * (1 - corpTax);
  const roic = price > 0 ? (year1NOPAT / price) * 100 : 0;

  // Year X ROIC = Year X NOPAT / Purchase Price
  // Denominator stays constant - shows operational return independent of growth
  // Growth flows through to ROIC proportionally (3% SDE growth = 3% ROIC growth)
  const exitYearSDE = startSDE * Math.pow(1 + growth, hold);
  const exitOperatingProfit = exitYearSDE - capex;
  const exitNOPAT = exitOperatingProfit * (1 - corpTax);
  const exitROIC = price > 0 ? (exitNOPAT / price) * 100 : 0;

  // Payback calculation
  let paybackYears = 0;
  let currentSDE = startSDE;
  let cumulativeEquityReturn = 0;

  for (let i = 1; i <= 30; i++) {
    let interestPaid = 0;
    let principalPaid = 0;

    // Calculate interest on remaining debt
    let remainingDebt = debt;
    for (let j = 1; j < i; j++) {
      principalPaid = annualDebtService - remainingDebt * intRate;
      remainingDebt = Math.max(0, remainingDebt - principalPaid);
    }
    interestPaid = i <= term ? remainingDebt * intRate : 0;

    let taxableIncome = currentSDE - interestPaid - capex;
    let taxPayment = Math.max(0, taxableIncome * corpTax);
    let flowToEquity =
      currentSDE - (i <= term ? annualDebtService : 0) - capex - taxPayment;

    cumulativeEquityReturn += flowToEquity;
    if (paybackYears === 0 && cumulativeEquityReturn >= equity) {
      const previousBalance = cumulativeEquityReturn - flowToEquity;
      const needed = equity - previousBalance;
      paybackYears = i - 1 + (flowToEquity > 0 ? needed / flowToEquity : 0);
    }
    currentSDE = currentSDE * (1 + growth);
  }

  // Exit year calculations
  const cashflows = [-equity]; // Initial investment (negative)
  currentSDE = startSDE;
  let totalPreTaxCF = 0;
  let remainingDebt = debt;
  let principalPaid = 0;

  for (let i = 1; i <= hold; i++) {
    // Calculate debt balance at end of year i
    let yearlyDebtService = i <= term ? annualDebtService : 0;
    let interestPaid = intRate > 0 ? remainingDebt * intRate : 0;
    principalPaid = yearlyDebtService - interestPaid;
    remainingDebt = Math.max(0, remainingDebt - principalPaid);

    let taxableIncome = currentSDE - interestPaid - capex;
    let taxPayment = Math.max(0, taxableIncome * corpTax);
    let afterTaxCF = currentSDE - yearlyDebtService - capex - taxPayment;

    if (i === hold) {
      // Exit year: add exit proceeds
      const exitValue = currentSDE * exitMult;
      const gainOnSale = Math.max(0, exitValue - remainingDebt - price);
      const gainsTaxPayment = gainOnSale * gainsTax;
      const exitProceeds = Math.max(
        0,
        exitValue - remainingDebt - gainsTaxPayment,
      );
      afterTaxCF += exitProceeds;
    }

    cashflows.push(afterTaxCF);
    totalPreTaxCF += afterTaxCF;
    currentSDE = currentSDE * (1 + growth);
  }

  // Calculate IRR and MoIC from cashflows
  const irr = solveIRR(cashflows) * 100;
  const moic =
    cashflows[cashflows.length - 1] > 0
      ? -cashflows[0] > 0
        ? (cashflows.reduce((a, b) => a + b) - cashflows[0]) / -cashflows[0] + 1
        : 0
      : 0;

  // Update UI
  document.getElementById("outROIC").innerText = roic.toFixed(1) + "%";
  document.getElementById("exitROICLabel").innerText =
    "Year " + Math.round(hold) + " ROIC";
  document.getElementById("outExitROIC").innerText = exitROIC.toFixed(1) + "%";
  document.getElementById("outPayback").innerText =
    paybackYears > 0 ? paybackYears.toFixed(1) + " Years" : ">30 Years";
  document.getElementById("outIRR").innerText = isFinite(irr)
    ? irr.toFixed(1) + "%"
    : "0%";
  document.getElementById("outMoIC").innerText = moic.toFixed(2) + "x MoIC";
}

function initSMB() {
  if (!document.getElementById("price")) return;

  // Attach event listeners (CSP-compliant, no inline handlers)
  document.getElementById("price").addEventListener("input", calculateSMB);
  document.getElementById("equityPct").addEventListener("input", calculateSMB);
  document.getElementById("intRate").addEventListener("input", calculateSMB);
  document.getElementById("term").addEventListener("input", calculateSMB);
  document.getElementById("cashflow").addEventListener("input", calculateSMB);
  document.getElementById("growth").addEventListener("input", calculateSMB);
  document.getElementById("hold").addEventListener("input", calculateSMB);
  document.getElementById("exitMult").addEventListener("input", calculateSMB);
  document.getElementById("capex").addEventListener("input", calculateSMB);
  document.getElementById("corpTax").addEventListener("input", calculateSMB);
  document.getElementById("gainsTax").addEventListener("input", calculateSMB);

  if (typeof toggleMode === "function") toggleMode();
  calculateSMB();

  if (typeof makeShareable === "function") {
    makeShareable(calculateSMB);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initSMB);
} else {
  initSMB();
}
