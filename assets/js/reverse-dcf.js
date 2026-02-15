function solveReverseDCF() {
  const priceInput = document.getElementById("sharePrice");
  const epsInput = document.getElementById("eps");
  const discountInput = document.getElementById("discount");
  const terminalInput = document.getElementById("terminal");
  const yearsInput = document.getElementById("years");

  if (!priceInput || !epsInput) return;

  const price = parseFloat(priceInput.value);
  const eps = parseFloat(epsInput.value);
  const discount = (parseFloat(discountInput.value) || 0) / 100;
  const terminalMult = parseFloat(terminalInput.value);
  const years = parseFloat(yearsInput.value);

  let low = -0.5;
  let high = 1.0;
  let impliedGrowth = 0;

  for (let i = 0; i < 25; i++) {
    impliedGrowth = (low + high) / 2;
    const estimatedValue = calculateDCF(
      eps,
      impliedGrowth,
      discount,
      terminalMult,
      years,
    );

    if (estimatedValue < price) {
      low = impliedGrowth;
    } else {
      high = impliedGrowth;
    }
  }

  // Result
  const growthPct = (impliedGrowth * 100).toFixed(2);
  document.getElementById("outImpliedGrowth").innerText = growthPct + "%";

  // Verdict Logic
  const badge = document.getElementById("outVerdict");
  if (impliedGrowth > 0.15) {
    badge.innerText = "Priced for Perfection";
    badge.style.color = "#ef4444"; // Red
  } else if (impliedGrowth < 0.05) {
    badge.innerText = "Margin of Safety";
    badge.style.color = "#16a34a"; // Green
  } else {
    badge.innerText = "Fairly Valued";
    badge.style.color = "#f59e0b"; // Orange
  }
}

function calculateDCF(startEPS, growth, discount, terminalMult, years) {
  let sumPV = 0;
  let currentEPS = startEPS;

  for (let t = 1; t <= years; t++) {
    currentEPS = currentEPS * (1 + growth);
    sumPV += currentEPS / Math.pow(1 + discount, t);
  }

  const terminalValue = currentEPS * terminalMult;
  const pvTerminal = terminalValue / Math.pow(1 + discount, years);

  return sumPV + pvTerminal;
}

function initRevDCF() {
  if (!document.getElementById("sharePrice")) return;

  // Attach event listeners (CSP-compliant, no inline handlers)
  document.getElementById("sharePrice").addEventListener("input", solveReverseDCF);
  document.getElementById("eps").addEventListener("input", solveReverseDCF);
  document.getElementById("discount").addEventListener("input", solveReverseDCF);
  document.getElementById("terminal").addEventListener("input", solveReverseDCF);
  document.getElementById("years").addEventListener("input", solveReverseDCF);

  solveReverseDCF();

  if (typeof makeShareable === "function") {
    makeShareable(solveReverseDCF);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initRevDCF);
} else {
  initRevDCF();
}
