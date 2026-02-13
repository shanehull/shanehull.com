function calculateSMB() {
  const price = parseFloat(document.getElementById("price").value) || 0;
  const equityPct =
    parseFloat(document.getElementById("equityPct").value) / 100;
  const intRate = parseFloat(document.getElementById("intRate").value) / 100;
  const term = parseFloat(document.getElementById("term").value) || 0;
  const startEBITDA =
    parseFloat(document.getElementById("cashflow").value) || 0;
  const growth = parseFloat(document.getElementById("growth").value) / 100;
  const hold = parseFloat(document.getElementById("hold").value) || 5;
  const exitMult = parseFloat(document.getElementById("exitMult").value) || 0;

  const debt = price * (1 - equityPct);
  const equity = price * equityPct;

  let annualDebtService = 0;
  if (intRate > 0) {
    annualDebtService = (debt * intRate) / (1 - Math.pow(1 + intRate, -term));
  } else {
    annualDebtService = debt / term;
  }

  const roic = price > 0 ? (startEBITDA / price) * 100 : 0;

  let cumulativeEquityReturn = 0;
  let paybackYears = 0;
  let currentEBITDA = startEBITDA;

  for (let i = 1; i <= 30; i++) {
    let flowToEquity = currentEBITDA - (i <= term ? annualDebtService : 0);

    cumulativeEquityReturn += flowToEquity;
    if (paybackYears === 0 && cumulativeEquityReturn >= equity) {
      const previousBalance = cumulativeEquityReturn - flowToEquity;
      const needed = equity - previousBalance;
      paybackYears = i - 1 + needed / flowToEquity;
    }
    currentEBITDA = currentEBITDA * (1 + growth);
  }

  let totalCashReturned = 0;
  currentEBITDA = startEBITDA;
  let finalYearEBITDA = 0;

  for (let i = 1; i <= hold; i++) {
    let netCF = currentEBITDA - (i <= term ? annualDebtService : 0);
    totalCashReturned += netCF;
    if (i === hold) finalYearEBITDA = currentEBITDA;
    currentEBITDA = currentEBITDA * (1 + growth);
  }

  const exitVal = finalYearEBITDA * exitMult;

  let remainingDebt = 0;
  if (hold < term) {
    remainingDebt =
      (debt * (Math.pow(1 + intRate, term) - Math.pow(1 + intRate, hold))) /
      (Math.pow(1 + intRate, term) - 1);
  }

  const exitProceeds = Math.max(0, exitVal - remainingDebt);
  const totalReturn = totalCashReturned + exitProceeds;

  const moic = equity > 0 ? totalReturn / equity : 0;
  const irr = (Math.pow(moic, 1 / hold) - 1) * 100;

  document.getElementById("outROIC").innerText = roic.toFixed(1) + "%";

  document.getElementById("outPayback").innerText =
    paybackYears > 0 ? paybackYears.toFixed(1) + " Years" : ">30 Years";

  document.getElementById("outIRR").innerText = isFinite(irr)
    ? irr.toFixed(1) + "%"
    : "0%";
  document.getElementById("outMoIC").innerText = moic.toFixed(2) + "x MoIC";
}

function initSMB() {
  if (!document.getElementById("price")) return;

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
