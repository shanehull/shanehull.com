function calculateAuto() {
  const freqPerDay = parseFloat(document.getElementById("frequency").value);
  const timeSaved = parseFloat(document.getElementById("timeSaved").value);
  const unitMultiplier = parseFloat(document.getElementById("unit").value);
  const years = parseFloat(document.getElementById("duration").value);

  // Calculate total seconds saved over the period
  const days = years * 365;
  const totalRuns = freqPerDay * days;
  const totalSecondsSaved = totalRuns * (timeSaved * unitMultiplier);

  // Convert to readable format
  let result = "";

  if (totalSecondsSaved < 60) {
    result = Math.round(totalSecondsSaved) + " Seconds";
  } else if (totalSecondsSaved < 3600) {
    result = (totalSecondsSaved / 60).toFixed(1) + " Minutes";
  } else if (totalSecondsSaved < 28800) {
    // Less than 8 hours
    result = (totalSecondsSaved / 3600).toFixed(1) + " Hours";
  } else if (totalSecondsSaved < 172800) {
    // Less than 2 days (48h)
    // Convert to Work Days (8 hour days)
    result = (totalSecondsSaved / 28800).toFixed(1) + " Work Days";
  } else {
    result = (totalSecondsSaved / 604800).toFixed(1) + " Weeks";
  }

  document.getElementById("outTime").innerText = result;
}

function initAuto() {
  if (!document.getElementById("frequency")) return;

  // Attach event listeners (CSP-compliant, no inline handlers)
  document
    .getElementById("frequency")
    .addEventListener("change", calculateAuto);
  document.getElementById("timeSaved").addEventListener("input", calculateAuto);
  document.getElementById("unit").addEventListener("change", calculateAuto);
  document.getElementById("duration").addEventListener("input", calculateAuto);

  calculateAuto();

  if (typeof makeShareable === "function") {
    makeShareable(calculateAuto);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initAuto);
} else {
  initAuto();
}
