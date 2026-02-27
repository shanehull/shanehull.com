/**
 * SMB Acquisition Financial Model
 *
 * Assumptions & Limitations:
 * - Linear SDE growth (compounded annually)
 * - Constant annual CapEx (not tax-deductible)
 * - No depreciation/amortization tax shields
 * - No working capital adjustments
 * - No transaction fees or earnouts
 * - Full debt payoff at exit (no refinance)
 * - Debt service is standard amortization (not interest-only)
 * - Exit gains taxed at capital gains rate
 */

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

function updateFinancingEquation() {
  const equity = parseFloat(document.getElementById("equityPct").value) || 0;
  const bank = parseFloat(document.getElementById("bankDebtPct").value) || 0;
  const seller =
    parseFloat(document.getElementById("sellerDebtPct").value) || 0;
  const total = equity + bank + seller;

  const equationEl = document.getElementById("financingEquation");
  const equation = `Equity ${equity}% + Bank ${bank}% + Seller ${seller}% = ${total}%`;
  equationEl.innerText = equation;

  // Green if equals 100%, red if less than 100%
  if (total === 100) {
    equationEl.style.color = "#16a34a";
  } else if (total < 100) {
    equationEl.style.color = "#ef4444";
  } else {
    // Exceeds 100% - also red
    equationEl.style.color = "#ef4444";
  }
}

function calculateSMB() {
  // Check if financing structure equals 100%
  const equityInput =
    parseFloat(document.getElementById("equityPct").value) || 0;
  const bankInput =
    parseFloat(document.getElementById("bankDebtPct").value) || 0;
  const sellerInput =
    parseFloat(document.getElementById("sellerDebtPct").value) || 0;
  const total = equityInput + bankInput + sellerInput;

  // If structure doesn't equal 100%, show dashes
  if (total !== 100) {
    document.getElementById("outROIC").innerText = "--";
    document.getElementById("outPayback").innerText = "--";
    document.getElementById("outIRR").innerText = "--";
    document.getElementById("outMoIC").innerText = "--";
    return;
  }

  const price = parseFloat(document.getElementById("price").value) || 0;
  const equityPct = equityInput / 100;
  const bankDebtPct = bankInput / 100;
  const sellerDebtPct = sellerInput / 100;
  const bankRate = parseFloat(document.getElementById("bankRate").value) / 100;
  const bankTerm = parseFloat(document.getElementById("bankTerm").value) || 0;
  const sellerRate =
    parseFloat(document.getElementById("sellerRate").value) / 100;
  const sellerTerm =
    parseFloat(document.getElementById("sellerTerm").value) || 0;
  const startSDE = parseFloat(document.getElementById("cashflow").value) || 0;
  const growth = parseFloat(document.getElementById("growth").value) / 100;
  const hold = parseFloat(document.getElementById("hold").value) || 5;
  const exitMult = parseFloat(document.getElementById("exitMult").value) || 0;
  const capex = parseFloat(document.getElementById("capex").value) || 0;
  const corpTax = parseFloat(document.getElementById("corpTax").value) / 100;
  const gainsTax = parseFloat(document.getElementById("gainsTax").value) / 100;

  const bankDebt = price * bankDebtPct;
  const sellerDebt = price * sellerDebtPct;
  const equity = price * equityPct;

  // Bank debt service calculation
  let annualBankDebtService = 0;
  if (bankRate > 0 && bankTerm > 0) {
    annualBankDebtService =
      (bankDebt * bankRate) / (1 - Math.pow(1 + bankRate, -bankTerm));
  } else if (bankTerm > 0) {
    annualBankDebtService = bankDebt / bankTerm;
  }

  // Seller debt service calculation
  let annualSellerDebtService = 0;
  if (sellerRate > 0 && sellerTerm > 0) {
    annualSellerDebtService =
      (sellerDebt * sellerRate) / (1 - Math.pow(1 + sellerRate, -sellerTerm));
  } else if (sellerTerm > 0) {
    annualSellerDebtService = sellerDebt / sellerTerm;
  }

  const totalAnnualDebtService =
    annualBankDebtService + annualSellerDebtService;

  // Calculate true ROIC with accumulated invested capital
  // Year 1 ROIC
  const year1OperatingProfit = startSDE - capex;
  const year1NOPAT = year1OperatingProfit * (1 - corpTax);
  const roic = price > 0 ? (year1NOPAT / price) * 100 : 0;

  // Payback calculation
  let paybackYears = 0;
  let currentSDE = startSDE;
  let cumulativeEquityReturn = 0;
  let trackingBankDebt = bankDebt;
  let trackingSellerDebt = sellerDebt;

  for (let i = 1; i <= 30; i++) {
    // Bank debt interest and principal for this year
    const bankInterestPaid =
      i <= bankTerm && trackingBankDebt > 0 ? trackingBankDebt * bankRate : 0;
    const bankPrincipalPaid =
      i <= bankTerm ? annualBankDebtService - bankInterestPaid : 0;
    trackingBankDebt = Math.max(0, trackingBankDebt - bankPrincipalPaid);

    // Seller debt interest and principal for this year
    const sellerInterestPaid =
      i <= sellerTerm && trackingSellerDebt > 0
        ? trackingSellerDebt * sellerRate
        : 0;
    const sellerPrincipalPaid =
      i <= sellerTerm ? annualSellerDebtService - sellerInterestPaid : 0;
    trackingSellerDebt = Math.max(0, trackingSellerDebt - sellerPrincipalPaid);

    const totalInterestPaid = bankInterestPaid + sellerInterestPaid;
    let debtServiceThisYear = 0;
    if (i <= bankTerm) debtServiceThisYear += annualBankDebtService;
    if (i <= sellerTerm) debtServiceThisYear += annualSellerDebtService;

    // Taxable income: SDE - interest (capex is not deductible)
    let taxableIncome = currentSDE - totalInterestPaid;
    let taxPayment = Math.max(0, taxableIncome * corpTax);
    let flowToEquity = currentSDE - debtServiceThisYear - capex - taxPayment;

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
  let remainingBankDebt = bankDebt;
  let remainingSellerDebt = sellerDebt;
  let bankPrincipalPaid = 0;
  let sellerPrincipalPaid = 0;

  for (let i = 1; i <= hold; i++) {
    // Calculate debt balances at end of year i
    let yearlyBankDebtService = i <= bankTerm ? annualBankDebtService : 0;
    let yearlySellerDebtService = i <= sellerTerm ? annualSellerDebtService : 0;

    let bankInterestPaid = bankRate > 0 ? remainingBankDebt * bankRate : 0;
    bankPrincipalPaid = yearlyBankDebtService - bankInterestPaid;
    remainingBankDebt = Math.max(0, remainingBankDebt - bankPrincipalPaid);

    let sellerInterestPaid =
      sellerRate > 0 ? remainingSellerDebt * sellerRate : 0;
    sellerPrincipalPaid = yearlySellerDebtService - sellerInterestPaid;
    remainingSellerDebt = Math.max(
      0,
      remainingSellerDebt - sellerPrincipalPaid,
    );

    const totalInterestPaid = bankInterestPaid + sellerInterestPaid;
    const totalDebtService = yearlyBankDebtService + yearlySellerDebtService;

    // Taxable income: EBITDA - interest (capex is not deductible, only interest)
    let taxableIncome = currentSDE - totalInterestPaid;
    let taxPayment = Math.max(0, taxableIncome * corpTax);
    // After-tax cashflow: pay debt service and capex after taxes
    let afterTaxCF = currentSDE - totalDebtService - capex - taxPayment;

    if (i === hold) {
      // Exit year: add exit proceeds
      const exitValue = currentSDE * exitMult;
      const totalRemainingDebt = remainingBankDebt + remainingSellerDebt;
      const gainOnSale = Math.max(0, exitValue - totalRemainingDebt - price);
      const gainsTaxPayment = gainOnSale * gainsTax;
      const exitProceeds = Math.max(
        0,
        exitValue - totalRemainingDebt - gainsTaxPayment,
      );
      afterTaxCF += exitProceeds;
    }

    cashflows.push(afterTaxCF);
    totalPreTaxCF += afterTaxCF;
    currentSDE = currentSDE * (1 + growth);
  }

  // Calculate IRR and MoIC from cashflows
  const irr = solveIRR(cashflows) * 100;
  // MoIC = (total distributions) / (initial equity investment)
  // cashflows[0] is negative (equity invested), remaining are returns
  const totalDistributions =
    cashflows.reduce((a, b) => a + b, 0) - cashflows[0];
  const moic = -cashflows[0] > 0 ? totalDistributions / -cashflows[0] + 1 : 0;

  // Calculate leverage ratios
  const entryDebt = bankDebt + sellerDebt;
  const entryLeverage = startSDE > 0 ? entryDebt / startSDE : 0;

  const exitSDE = startSDE * Math.pow(1 + growth, hold);
  const exitLeverage =
    exitSDE > 0 ? (remainingBankDebt + remainingSellerDebt) / exitSDE : 0;

  // Update UI
  document.getElementById("outROIC").innerText = roic.toFixed(1) + "%";
  document.getElementById("outPayback").innerText =
    paybackYears > 0 ? paybackYears.toFixed(1) + " Years" : ">30 Years";
  document.getElementById("outIRR").innerText = isFinite(irr)
    ? irr.toFixed(1) + "%"
    : "0%";
  document.getElementById("outMoIC").innerText = moic.toFixed(2) + "x MoIC";
  document.getElementById("outEntryLeverage").innerText =
    entryLeverage.toFixed(2) + "x";
  document.getElementById("outExitLeverage").innerText =
    exitLeverage.toFixed(2) + "x";
}

function initSMB() {
  if (!document.getElementById("price")) return;

  // Attach event listeners (CSP-compliant, no inline handlers)
  document.getElementById("price").addEventListener("input", calculateSMB);

  // Financing structure inputs update equation and trigger calculation
  document
    .getElementById("equityPct")
    .addEventListener("input", updateFinancingEquation);
  document
    .getElementById("bankDebtPct")
    .addEventListener("input", updateFinancingEquation);
  document
    .getElementById("sellerDebtPct")
    .addEventListener("input", updateFinancingEquation);

  document.getElementById("equityPct").addEventListener("input", calculateSMB);
  document
    .getElementById("bankDebtPct")
    .addEventListener("input", calculateSMB);
  document
    .getElementById("sellerDebtPct")
    .addEventListener("input", calculateSMB);

  document.getElementById("bankRate").addEventListener("input", calculateSMB);
  document.getElementById("bankTerm").addEventListener("input", calculateSMB);
  document.getElementById("sellerRate").addEventListener("input", calculateSMB);
  document.getElementById("sellerTerm").addEventListener("input", calculateSMB);
  document.getElementById("cashflow").addEventListener("input", calculateSMB);
  document.getElementById("growth").addEventListener("input", calculateSMB);
  document.getElementById("hold").addEventListener("input", calculateSMB);
  document.getElementById("exitMult").addEventListener("input", calculateSMB);
  document.getElementById("capex").addEventListener("input", calculateSMB);
  document.getElementById("corpTax").addEventListener("input", calculateSMB);
  document.getElementById("gainsTax").addEventListener("input", calculateSMB);

  if (typeof toggleMode === "function") toggleMode();
  updateFinancingEquation();
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
