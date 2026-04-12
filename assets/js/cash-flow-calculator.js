let assetIdCounter = 0;

function debounce(func, wait) {
  let timeout;
  return function (...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
}

let initComplete = false;

function addAsset() {
  assetIdCounter++;
  const list = document.getElementById("assetList");
  const card = document.createElement("div");

  card.className = "calculator-dynamic-card";
  card.id = `asset-${assetIdCounter}`;

  card.innerHTML = `
        ${assetIdCounter > 1 ? `<span class="calculator-remove-btn" data-id="${assetIdCounter}">✕</span>` : ""}
        
        <div class="calculator-field" style="margin-bottom: 1rem;">
            <label for="asset-${assetIdCounter}-desc">Asset Description</label>
            <input type="text" placeholder="e.g., Gold Production" id="asset-${assetIdCounter}-desc" class="asset-input">
        </div>
        
        <div class="calculator-row">
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-prod">Production</label>
                <input type="number" id="asset-${assetIdCounter}-prod" class="p-prod" value="100000">
            </div>
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-price">Price ($)</label>
                <input type="number" id="asset-${assetIdCounter}-price" class="p-price" value="3500">
            </div>
        </div>
        
        <div class="calculator-row">
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-cost">Unit Cost (AISC $)</label>
                <input type="number" id="asset-${assetIdCounter}-cost" class="p-cost" value="1200">
            </div>
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-royalty">Royalty (%)</label>
                <input type="number" id="asset-${assetIdCounter}-royalty" class="p-royalty" value="0">
            </div>
        </div>

        <div class="calculator-row">
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-stream-pct">Stream (%)</label>
                <input type="number" id="asset-${assetIdCounter}-stream-pct" class="p-stream-pct" value="0">
            </div>
            <div class="calculator-field">
                <label for="asset-${assetIdCounter}-stream-price">Stream Price ($)</label>
                <input type="number" id="asset-${assetIdCounter}-stream-price" class="p-stream-price" value="0">
            </div>
        </div>
    `;
  list.appendChild(card);

  // Attach event listeners to new inputs (CSP-compliant)
  const newInputs = card.querySelectorAll(
    ".asset-input, .p-prod, .p-price, .p-cost, .p-royalty, .p-stream-pct, .p-stream-price",
  );
  newInputs.forEach((input) => {
    input.addEventListener("input", calculate);
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
  const dilution =
    (parseFloat(document.getElementById("dilution").value) || 0) / 100;
  const ga = (parseFloat(document.getElementById("ga").value) || 0) * 1000000;

  let totalGrossFCF = 0;

  const cards = document.querySelectorAll(".calculator-dynamic-card");

  cards.forEach((card) => {
    const prod = parseFloat(card.querySelector(".p-prod").value) || 0;
    const price = parseFloat(card.querySelector(".p-price").value) || 0;
    const cost = parseFloat(card.querySelector(".p-cost").value) || 0;
    const royalty =
      (parseFloat(card.querySelector(".p-royalty").value) || 0) / 100;
    const streamPct =
      (parseFloat(card.querySelector(".p-stream-pct").value) || 0) / 100;
    const streamPrice =
      parseFloat(card.querySelector(".p-stream-price").value) || 0;

    // Realized Price Before Royalty = (unstreamed portion * spot) + (streamed portion * stream price)
    const realizedBeforeRoyalty =
      price * (1 - streamPct) + streamPrice * streamPct;

    // Final Realized Price = (Total Realized Revenue) * (1 - Royalty)
    const finalRealizedPrice = realizedBeforeRoyalty * (1 - royalty);

    totalGrossFCF += (finalRealizedPrice - cost) * prod;
  });

  const netFCF = (totalGrossFCF - ga) * (1 - tax);
  const mCap = netFCF * multiple;
  const dilutedShares = shares * (1 + dilution);
  const targetPrice = dilutedShares > 0 ? mCap / dilutedShares : 0;
  const upsidePct =
    currPrice > 0 ? ((targetPrice - currPrice) / currPrice) * 100 : 0;
  const upsideX = currPrice > 0 ? (targetPrice - currPrice) / currPrice : 0;

  document.getElementById("outFCF").innerText = formatCompact(netFCF);
  document.getElementById("outMCap").innerText = formatCompact(mCap);
  document.getElementById("outPrice").innerText = "$" + targetPrice.toFixed(2);
  document.getElementById("outUpside").innerText =
    `${upsidePct.toLocaleString(undefined, { maximumFractionDigits: 0 })}% (+${upsideX.toFixed(2)}x)`;
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
  document.getElementById("dilution").addEventListener("input", calculate);
  document.getElementById("ga").addEventListener("input", calculate);

  const addAssetBtn = document.getElementById("addAssetBtn");
  if (addAssetBtn) {
    addAssetBtn.addEventListener("click", addAsset);
  }

  // Restore assets from URL if they exist
  const params = new URLSearchParams(window.location.search);
  let maxAssetId = 0;
  for (const key of params.keys()) {
    if (key.startsWith("asset-") && key.includes("-prod")) {
      const id = parseInt(key.split("-")[1]);
      maxAssetId = Math.max(maxAssetId, id);
    }
  }

  // Create assets: either restore from URL or create 1 default
  const assetsToCreate = Math.max(1, maxAssetId);
  for (let i = 0; i < assetsToCreate; i++) {
    addAsset();
  }

  // Run initial calculation
  calculate();

  // Mark initialization complete - from now on, URL updates on user input
  initComplete = true;

  // Use universal makeShareable for URL management
  if (typeof makeShareable === "function") {
    makeShareable(calculate);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initCashflow);
} else {
  initCashflow();
}
