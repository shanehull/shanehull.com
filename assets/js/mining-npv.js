function calculateMineNPV() {
  const reserves = parseFloat(document.getElementById("reserves").value) || 0;
  const production =
    parseFloat(document.getElementById("production").value) || 0;
  const price = parseFloat(document.getElementById("price").value) || 0;
  const statedCost = parseFloat(document.getElementById("cost").value) || 0;
  const statedCapex = parseFloat(document.getElementById("capex").value) || 0;
  const discount =
    (parseFloat(document.getElementById("discount").value) || 0) / 100;
  const tax = (parseFloat(document.getElementById("tax").value) || 0) / 100;

  const capexOverrunPct =
    parseFloat(document.getElementById("capexOverrun").value) || 0;
  const costInflationPct =
    parseFloat(document.getElementById("costInflation").value) || 0;

  const realCapex = statedCapex * (1 + capexOverrunPct / 100);
  const realCost = statedCost * (1 + costInflationPct / 100);

  let lifeOfMine = 0;
  if (production > 0) {
    lifeOfMine = reserves / production;
  }

  const margin = price - realCost;
  const grossIncome = margin * production;
  const netFCF = grossIncome * (1 - tax);

  let totalNPV = -realCapex;

  const fullYears = Math.floor(lifeOfMine);
  const partialYear = lifeOfMine - fullYears;

  for (let t = 1; t <= fullYears; t++) {
    totalNPV += netFCF / Math.pow(1 + discount, t);
  }

  if (partialYear > 0) {
    const partialFCF = netFCF * partialYear;
    totalNPV += partialFCF / Math.pow(1 + discount, fullYears + 1);
  }

  document.getElementById("outLife").innerText =
    lifeOfMine.toFixed(1) + " Years";
  document.getElementById("outAnnualFCF").innerText = formatCurrency(netFCF);

  const outNPV = document.getElementById("outNPV");
  outNPV.innerText = formatCurrency(totalNPV);
}

function formatCurrency(num) {
  if (Math.abs(num) >= 1e9) return "$" + (num / 1e9).toFixed(2) + "B";
  if (Math.abs(num) >= 1e6) return "$" + (num / 1e6).toFixed(2) + "M";
  return "$" + num.toLocaleString(undefined, { maximumFractionDigits: 0 });
}

function initNPV() {
  if (!document.getElementById("reserves")) return;

  // Attach event listeners (CSP-compliant, no inline handlers)
  document.getElementById("reserves").addEventListener("input", calculateMineNPV);
  document.getElementById("production").addEventListener("input", calculateMineNPV);
  document.getElementById("price").addEventListener("input", calculateMineNPV);
  document.getElementById("cost").addEventListener("input", calculateMineNPV);
  document.getElementById("capex").addEventListener("input", calculateMineNPV);
  document.getElementById("discount").addEventListener("input", calculateMineNPV);
  document.getElementById("capexOverrun").addEventListener("input", calculateMineNPV);
  document.getElementById("costInflation").addEventListener("input", calculateMineNPV);
  document.getElementById("tax").addEventListener("input", calculateMineNPV);

  calculateMineNPV();

  if (typeof makeShareable === "function") {
    makeShareable(calculateMineNPV);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initNPV);
} else {
  initNPV();
}
