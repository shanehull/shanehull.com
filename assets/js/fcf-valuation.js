let assetIdCounter = 0;

function debounce(func, wait) {
  let timeout;
  return function (...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
}

function updateShareableURL(input) {
  const url = new URL(window.location);
  if (input.id) {
    if (input.value) {
      url.searchParams.set(input.id, input.value);
    } else {
      url.searchParams.delete(input.id);
    }
    window.history.replaceState({}, "", url);
  }
}

function addAsset() {
  assetIdCounter++;
  const list = document.getElementById("assetList");
  const card = document.createElement("div");

  card.className = "calculator-dynamic-card";
  card.id = `asset-${assetIdCounter}`;

  card.innerHTML = `
        ${assetIdCounter > 1 ? `<span class="calculator-remove-btn" data-id="${assetIdCounter}">✕</span>` : ""}
        
        <div class="calculator-field" style="margin-bottom: 1rem;">
            <label>Asset Description</label>
            <input type="text" placeholder="e.g., Gold Production" id="asset-${assetIdCounter}-desc" class="asset-input">
        </div>
        
        <div class="calculator-row">
            <div class="calculator-field">
                <label>Production</label>
                <input type="number" id="asset-${assetIdCounter}-prod" class="p-prod" value="100000">
            </div>
            <div class="calculator-field">
                <label>Price ($)</label>
                <input type="number" id="asset-${assetIdCounter}-price" class="p-price" value="2500">
            </div>
        </div>
        
        <div class="calculator-row">
            <div class="calculator-field">
                <label>Unit Cost ($)</label>
                <input type="number" id="asset-${assetIdCounter}-cost" class="p-cost" value="1200">
            </div>
        </div>
    `;
  list.appendChild(card);

  // Attach event listeners to new inputs (CSP-compliant)
  const newInputs = card.querySelectorAll(
    ".asset-input, .p-prod, .p-price, .p-cost",
  );
  newInputs.forEach((input) => {
    input.addEventListener("input", calculate);
    // Also attach URL update listeners if makeShareable is available
    if (typeof makeShareable !== "undefined") {
      input.addEventListener(
        "input",
        debounce(() => updateShareableURL(input), 300),
      );
      input.addEventListener("change", updateShareableURL.bind(null, input));
    }
  });

  const removeBtn = card.querySelector(".calculator-remove-btn");
  if (removeBtn) {
    removeBtn.addEventListener("click", () => {
      const id = removeBtn.dataset.id;
      removeAsset(id);
    });
  }

  calculate();
}

function removeAsset(id) {
  document.getElementById(`asset-${id}`).remove();
  calculate();
}

function calculate() {
  const shares =
    (parseFloat(document.getElementById("shares").value) || 0) * 1000000;
  const currPrice = parseFloat(document.getElementById("currPrice").value) || 0;
  const multiple = parseFloat(document.getElementById("multiple").value) || 0;
  const tax = (parseFloat(document.getElementById("tax").value) || 0) / 100;

  let totalGrossFCF = 0;

  const cards = document.querySelectorAll(".calculator-dynamic-card");

  cards.forEach((card) => {
    const prod = parseFloat(card.querySelector(".p-prod").value) || 0;
    const price = parseFloat(card.querySelector(".p-price").value) || 0;
    const cost = parseFloat(card.querySelector(".p-cost").value) || 0;
    totalGrossFCF += (price - cost) * prod;
  });

  const netFCF = totalGrossFCF * (1 - tax);
  const mCap = netFCF * multiple;
  const targetPrice = shares > 0 ? mCap / shares : 0;
  const upsidePct =
    currPrice > 0 ? ((targetPrice - currPrice) / currPrice) * 100 : 0;
  const upsideX = currPrice > 0 ? targetPrice / currPrice : 0;

  document.getElementById("outFCF").innerText = formatCompact(netFCF);
  document.getElementById("outMCap").innerText = formatCompact(mCap);
  document.getElementById("outPrice").innerText = "$" + targetPrice.toFixed(2);
  document.getElementById("outUpside").innerText =
    `${upsidePct.toLocaleString(undefined, { maximumFractionDigits: 0 })}% (${upsideX.toFixed(1)}x)`;
}

function formatCompact(num) {
  if (Math.abs(num) >= 1e9) return "$" + (num / 1e9).toFixed(2) + "B";
  if (Math.abs(num) >= 1e6) return "$" + (num / 1e6).toFixed(2) + "M";
  return "$" + Math.round(num).toLocaleString();
}

function initCashflow() {
  if (!document.getElementById("shares")) return;

  // Attach event listeners to main inputs (CSP-compliant, no inline handlers)
  document.getElementById("shares").addEventListener("input", calculate);
  document.getElementById("currPrice").addEventListener("input", calculate);
  document.getElementById("multiple").addEventListener("input", calculate);
  document.getElementById("tax").addEventListener("input", calculate);

  const addAssetBtn = document.getElementById("addAssetBtn");
  if (addAssetBtn) {
    addAssetBtn.addEventListener("click", addAsset);
  }

  const list = document.getElementById("assetList");
  if (list && list.children.length === 0) {
    // Check URL for asset count before creating
    const params = new URLSearchParams(window.location.search);
    let maxAssetId = 0;
    for (const key of params.keys()) {
      if (key.startsWith("asset-") && key.includes("-prod")) {
        const id = parseInt(key.split("-")[1]);
        maxAssetId = Math.max(maxAssetId, id);
      }
    }

    // Create all assets that exist in URL (at least one)
    const assetsToCreate = Math.max(1, maxAssetId);
    for (let i = 0; i < assetsToCreate; i++) {
      addAsset();
    }
  } else {
    calculate();
  }

  if (typeof makeShareable === "function") {
    makeShareable(calculate);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initCashflow);
} else {
  initCashflow();
}
