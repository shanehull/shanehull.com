let API_HOST = "yahoo-finance15.p.rapidapi.com";
let currentChain = null;
let plChart = null;
let searchTimeout = null;
let currentHeatmapMode = "dollar";
let latestStrategyCalc = null;

const N = (x) => {
  let a1 = 0.31938153,
    a2 = -0.356563782,
    a3 = 1.781477937,
    a4 = -1.821255978,
    a5 = 1.330274429;
  let L = Math.abs(x),
    k = 1 / (1 + 0.2316419 * L);
  let w =
    1 -
    (1 / Math.sqrt(2 * Math.PI)) *
      Math.exp((-L * L) / 2) *
      (a1 * k +
        a2 * k * k +
        a3 * Math.pow(k, 3) +
        a4 * Math.pow(k, 4) +
        a5 * Math.pow(k, 5));
  return x < 0 ? 1 - w : w;
};

const BlackScholes = (type, S, K, T, r, sigma) => {
  if (T <= 0) return Math.max(0, type === "call" ? S - K : K - S);
  let d1 =
    (Math.log(S / K) + (r + (sigma * sigma) / 2) * T) / (sigma * Math.sqrt(T));
  let d2 = d1 - sigma * Math.sqrt(T);
  if (type === "call") return S * N(d1) - K * Math.exp(-r * T) * N(d2);
  return K * Math.exp(-r * T) * N(-d2) - S * N(-d1);
};

function calculateOptions() {
  const mode = document.getElementById("inputMode").value;
  const side = document.getElementById("sideSelect").value;
  const type = document.getElementById("typeSelect").value;
  const S = parseFloat(document.getElementById("currentPrice").value) || 0;
  const contracts = parseFloat(document.getElementById("contracts").value) || 1;
  const rangePct =
    parseFloat(document.getElementById("chartRange").value) / 100 || 0.2;
  const iv = parseFloat(document.getElementById("ivInput").value) / 100 || 0.3;
  const r = 0.05;
  const lotSize = parseFloat(document.getElementById("lotSize").value) || 100;

  let K = 0,
    T = 0.1;

  if (mode === "manual") {
    K = parseFloat(document.getElementById("strikeInputManual").value) || S;
    const dateStr = document.getElementById("expiryInputManual").value;
    if (dateStr) {
      const expiryDate = new Date(dateStr);
      const now = new Date();
      T = (expiryDate - now) / (1000 * 60 * 60 * 24 * 365);
    }
  } else {
    K = parseFloat(document.getElementById("strikeSelect").value) || S;
    const expTimestamp = parseInt(
      document.getElementById("expirySelect").value,
    );
    if (expTimestamp) {
      const now = Math.floor(Date.now() / 1000);
      T = (expTimestamp - now) / (365 * 24 * 60 * 60);
    }
  }
  if (T < 0.001) T = 0.001;

  const fairValue = BlackScholes(type, S, K, T, r, iv);

  const userPrice = parseFloat(document.getElementById("contractPrice").value);
  let contractPrice = fairValue;

  if (!isNaN(userPrice) && userPrice > 0) {
    contractPrice = userPrice;
  } else {
    document.getElementById("contractPrice").value = fairValue.toFixed(2);
  }

  const totalEntry = contractPrice * lotSize * contracts;

  document.getElementById("outFairValue").innerText =
    `$${fairValue.toFixed(2)}`;

  const breakeven = type === "call" ? K + contractPrice : K - contractPrice;
  document.getElementById("outBreakeven").innerText =
    `$${breakeven.toFixed(2)}`;

  let maxProfit = 0;
  let maxLoss = 0;

  if (side === "long") {
    maxLoss = totalEntry;
    if (type === "call") {
      maxProfit = Infinity;
    } else {
      maxProfit = Math.max(0, K - contractPrice) * lotSize * contracts;
    }
  } else {
    maxProfit = totalEntry;
    if (type === "call") {
      maxLoss = Infinity;
    } else {
      maxLoss = Math.max(0, contractPrice - K) * lotSize * contracts;
    }
  }

  document.getElementById("outMaxProfit").innerText =
    maxProfit === Infinity ? "Unlimited" : `$${maxProfit.toFixed(0)}`;
  document.getElementById("outMaxLoss").innerText =
    maxLoss === Infinity ? "Unlimited" : `$${maxLoss.toFixed(0)}`;

  const outEntry = document.getElementById("outEntryCost");
  if (side === "long") {
    outEntry.innerText = `Debit $${totalEntry.toFixed(0)}`;
    outEntry.style.color = "";
    document.getElementById("outProfitBadge").innerText =
      "Max Risk: $" + totalEntry.toFixed(0);
    document.getElementById("outProfitBadge").style.color = "#ef4444";
  } else {
    outEntry.innerText = `Credit $${totalEntry.toFixed(0)}`;
    outEntry.style.color = "#16a34a";
    document.getElementById("outProfitBadge").innerText =
      "Max Profit: $" + totalEntry.toFixed(0);
    document.getElementById("outProfitBadge").style.color = "#16a34a";
  }

  latestStrategyCalc = {
    S,
    K,
    T,
    iv,
    type,
    side,
    contracts,
    lotSize,
    cost: contractPrice,
    rangePct,
    entry: totalEntry,
  };

  renderChart(
    S,
    K,
    T,
    iv,
    type,
    side,
    contracts,
    lotSize,
    contractPrice,
    rangePct,
    totalEntry,
  );
  renderHeatmap(
    S,
    K,
    T,
    iv,
    type,
    side,
    contracts,
    lotSize,
    contractPrice,
    rangePct,
    totalEntry,
    currentHeatmapMode,
  );
}

function renderChart(
  S,
  K,
  T,
  iv,
  type,
  side,
  contracts,
  lotSize,
  cost,
  range,
  entryTotal,
) {
  const ctx = document.getElementById("plChart");
  if (!ctx) return;

  const steps = 20;
  const labels = [],
    dataExp = [],
    dataNow = [];

  for (let i = -steps; i <= steps; i++) {
    let p = S * (1 + (i / steps) * range);
    labels.push(p.toFixed(2));

    let valE = BlackScholes(type, p, K, 0.0001, 0.05, iv);
    let plE =
      (side === "long" ? valE - cost : cost - valE) * lotSize * contracts;
    dataExp.push(plE);

    let valN = BlackScholes(type, p, K, T, 0.05, iv);
    let plN =
      (side === "long" ? valN - cost : cost - valN) * lotSize * contracts;
    dataNow.push(plN);
  }

  if (plChart) plChart.destroy();

  plChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Expiration",
          data: dataExp,
          borderColor: "#58a6ff",
          borderWidth: 2,
          pointRadius: 0,
        },
        {
          label: "Today",
          data: dataNow,
          borderColor: "#8b949e",
          borderDash: [5, 5],
          borderWidth: 2,
          pointRadius: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { display: true } },
      scales: {
        x: {
          ticks: { maxTicksLimit: 8, color: "#888" },
          grid: { color: "#333" },
        },
        y: { ticks: { color: "#888" }, grid: { color: "#333" } },
      },
    },
  });
}

function renderHeatmap(
  S,
  K,
  T,
  iv,
  type,
  side,
  contracts,
  lotSize,
  cost,
  range,
  entryTotal,
  mode = "dollar",
) {
  const container = document.getElementById("heatmapContainer");

  let html =
    '<table class="heatmap" style="width:100%; border-collapse:collapse; font-size:0.8rem; text-align:center;">';

  html +=
    '<tr><th style="padding:5px; background:rgba(255,255,255,0.1);">Price</th>';
  const dateSteps = 5;
  const now = new Date();
  const dates = [];

  for (let i = 0; i <= dateSteps; i++) {
    let t_rem = T * (1 - i / dateSteps);
    if (t_rem < 0.001) t_rem = 0.0001;
    let d = new Date(now.getTime() + (T - t_rem) * 365 * 24 * 60 * 60 * 1000);
    dates.push({ t: t_rem });
    html += `<th style="padding:5px; background:rgba(255,255,255,0.1);">${d.toLocaleDateString(undefined, { month: "short", day: "numeric" })}</th>`;
  }
  html += "</tr>";

  const priceSteps = 8;
  for (let i = priceSteps; i >= -priceSteps; i--) {
    let p = S * (1 + (i / priceSteps) * range);
    html += `<tr><td style="padding:5px; font-weight:bold; border:1px solid rgba(255,255,255,0.1);">${p.toFixed(2)}</td>`;

    dates.forEach((d) => {
      let theo = BlackScholes(type, p, K, d.t, 0.05, iv);
      let pl =
        (side === "long" ? theo - cost : cost - theo) * lotSize * contracts;
      let val = theo * lotSize * contracts;

      let displayTxt = Math.round(pl);
      if (mode === "percent") {
        let roi = entryTotal > 0 ? (pl / entryTotal) * 100 : 0;
        displayTxt = roi.toFixed(0) + "%";
      } else if (mode === "value") {
        displayTxt = Math.round(val);
      }

      let opacity = Math.min(0.8, Math.abs(pl) / (entryTotal || 500));
      let color =
        pl >= 0
          ? `rgba(35, 134, 54, ${Math.max(0.1, opacity)})`
          : `rgba(218, 54, 51, ${Math.max(0.1, opacity)})`;

      html += `<td style="background:${color}; padding:5px; border:1px solid rgba(255,255,255,0.1); color:white;">${displayTxt}</td>`;
    });
    html += "</tr>";
  }
  html += "</table>";
  container.innerHTML = html;
}

function showApiKeyModal(errorMsg = null) {
  const modal = document.getElementById("apiKeyModal");
  const errorEl = document.getElementById("modalErrorMsg");

  if (errorMsg) {
    errorEl.innerText = errorMsg;
    errorEl.classList.add("modal-error-visible");
  } else {
    errorEl.classList.remove("modal-error-visible");
    errorEl.innerText = "";
  }

  modal.classList.add("modal-visible");
}

function closeApiKeyModal() {
  const modal = document.getElementById("apiKeyModal");
  modal.classList.remove("modal-visible");
}

async function fetchTickerSuggestions() {
  const fullValue = document
    .getElementById("ticker")
    .value.trim()
    .toUpperCase();
  // Extract just the ticker symbol (before any parentheses)
  const query = fullValue.split("(")[0].trim();
  const apiKey = document.getElementById("apiKey").value;
  const suggestionsEl = document.getElementById("suggestions");

  // Don't search if query is empty
  if (!query) {
    suggestionsEl.innerHTML = "";
    return;
  }

  if (!apiKey) {
    suggestionsEl.innerHTML = "";
    return;
  }

  // Show loading state
  suggestionsEl.innerHTML = '<div style="padding: 0.5rem; text-align: center; color: inherit;">Searching...</div>';

  try {
    const res = await fetch(
      `https://${API_HOST}/api/v1/markets/search?search=${query}`,
      {
        headers: { "X-RapidAPI-Key": apiKey, "X-RapidAPI-Host": API_HOST },
      },
    );
    const data = await res.json();

    if (!res.ok) {
      if (res.status === 429) {
        suggestionsEl.innerHTML = '<div style="padding: 0.75rem; color: #ef4444; text-align: center; font-size: 0.9rem;">API quota exceeded. Upgrade at RapidAPI.</div>';
      } else {
        suggestionsEl.innerHTML = "";
      }
      return;
    }

    if (data.body && Array.isArray(data.body) && data.body.length > 0) {
      suggestionsEl.innerHTML = "";
      data.body.slice(0, 8).forEach((result) => {
        const div = document.createElement("div");
        const name = result.name || result.shortname || result.longname || "";
        div.innerHTML = `<strong>${result.symbol}</strong><br><span style="font-size: 0.75rem; opacity: 0.7;">${name}</span>`;
        div.className = "options-pnl-suggestion-item";

        div.addEventListener("click", () => {
          document.getElementById("ticker").value = result.symbol;
          suggestionsEl.innerHTML = "";
          fetchOptionsForTicker(result.symbol);
        });

        suggestionsEl.appendChild(div);
      });
    } else {
      suggestionsEl.innerHTML = "";
    }
  } catch (e) {
    suggestionsEl.innerHTML = "";
  }
}

async function fetchOptionsForTicker(symbol) {
  const apiKey = document.getElementById("apiKey").value;

  if (!apiKey) {
    showApiKeyModal();
    return;
  }

  try {
    const res = await fetch(
      `https://${API_HOST}/api/v1/markets/options?ticker=${symbol}&display=list`,
      {
        headers: { "X-RapidAPI-Key": apiKey, "X-RapidAPI-Host": API_HOST },
      },
    );
    const data = await res.json();

    if (!res.ok) {
      const errorMsg = data.message || `Error: ${res.status} ${res.statusText}`;
      showApiKeyModal("Couldn't fetch data for " + symbol + ": " + errorMsg);
      return;
    }

    if (data.body && data.body[0]) {
      const result = data.body[0];
      const price = result.quote.regularMarketPrice;
      document.getElementById("currentPrice").value = price;

      const dates = result.expirationDates || [];
      const sel = document.getElementById("expirySelect");
      sel.innerHTML = "";
      dates.forEach((d) => {
        let opt = document.createElement("option");
        opt.value = d;
        opt.innerText = new Date(d * 1000).toISOString().split("T")[0];
        sel.appendChild(opt);
      });

      if (result.options && result.options[0]) {
        processOptionsChain(result.options[0]);
      }
    } else {
      showApiKeyModal(
        "No options data found for " + symbol + ". Try a different ticker.",
      );
    }
  } catch (e) {
    console.error("Fetch failed", e);
    showApiKeyModal("Network error: " + e.message);
  }
}

async function handleSearch() {
  const fullValue = document.getElementById("ticker").value.toUpperCase();
  // Extract just the ticker symbol (before any parentheses)
  const symbol = fullValue.split("(")[0].trim();
  const apiKey = document.getElementById("apiKey").value;

  // Check if user is trying to use demo mode (has parentheses)
  if (fullValue.includes("(")) {
    simulateData();
    return;
  }

  if (!apiKey) {
    showApiKeyModal();
    return;
  }

  await fetchOptionsForTicker(symbol);
}

function processOptionsChain(options) {
  const strikes = new Set();
  (options.calls || []).forEach((o) => strikes.add(o.strike));
  (options.puts || []).forEach((o) => strikes.add(o.strike));

  const sorted = Array.from(strikes).sort((a, b) => a - b);
  const sel = document.getElementById("strikeSelect");
  sel.innerHTML = "";

  let closest = sorted[0],
    minDiff = Infinity;
  const current = parseFloat(document.getElementById("currentPrice").value);

  sorted.forEach((s) => {
    let opt = document.createElement("option");
    opt.value = s;
    opt.innerText = s;
    sel.appendChild(opt);
    if (Math.abs(s - current) < minDiff) {
      minDiff = Math.abs(s - current);
      closest = s;
    }
  });
  sel.value = closest;
  calculateOptions();
}

function simulateData() {
  document.getElementById("currentPrice").value = 150;
  const sel = document.getElementById("expirySelect");
  sel.innerHTML = "";
  const now = Math.floor(Date.now() / 1000);
  [30, 60, 90].forEach((days) => {
    let d = now + days * 86400;
    let opt = document.createElement("option");
    opt.value = d;
    opt.innerText = new Date(d * 1000).toISOString().split("T")[0];
    sel.appendChild(opt);
  });

  const strSel = document.getElementById("strikeSelect");
  strSel.innerHTML = "";
  for (let i = 100; i <= 200; i += 5) {
    let opt = document.createElement("option");
    opt.value = i;
    opt.innerText = i;
    strSel.appendChild(opt);
    if (i === 150) opt.selected = true;
  }
  calculateOptions();
}

function updateUiMode() {
  const mode = document.getElementById("inputMode").value;
  const isManual = mode === "manual";

  const toggleVisibility = (id, show) => {
    const el = document.getElementById(id);
    if (show) {
      el.classList.add("options-pnl-visible");
      el.classList.remove("options-pnl-hidden");
    } else {
      el.classList.add("options-pnl-hidden");
      el.classList.remove("options-pnl-visible");
    }
  };

  toggleVisibility("apiKeyField", !isManual);
  toggleVisibility("tickerField", !isManual);
  toggleVisibility("expirySelectWrapper", !isManual);
  toggleVisibility("expiryManualWrapper", isManual);
  toggleVisibility("strikeSelectWrapper", !isManual);
  toggleVisibility("strikeManualWrapper", isManual);
}

function setHeatmapMode(mode) {
  currentHeatmapMode = mode;

  if (latestStrategyCalc) {
    renderHeatmap(
      latestStrategyCalc.S,
      latestStrategyCalc.K,
      latestStrategyCalc.T,
      latestStrategyCalc.iv,
      latestStrategyCalc.type,
      latestStrategyCalc.side,
      latestStrategyCalc.contracts,
      latestStrategyCalc.lotSize,
      latestStrategyCalc.cost,
      latestStrategyCalc.rangePct,
      latestStrategyCalc.entry,
      mode,
    );
  }
}

function initOptions() {
  if (!document.getElementById("currentPrice")) return;

  // Restore API key from localStorage
  const savedApiKey = localStorage.getItem("options-pnl-api-key");
  if (savedApiKey) {
    document.getElementById("apiKey").value = savedApiKey;
  }

  // Save API key to localStorage when changed
  document.getElementById("apiKey").addEventListener("input", (e) => {
    localStorage.setItem("options-pnl-api-key", e.target.value);
  });

  const inputs = [
    "inputMode",
    "ticker",
    "currentPrice",
    "sideSelect",
    "typeSelect",
    "expirySelect",
    "expiryInputManual",
    "strikeSelect",
    "strikeInputManual",
    "contractPrice",
    "ivInput",
    "contracts",
    "lotSize",
    "chartRange",
    "heatmapMode",
  ];

  inputs.forEach((id) => {
    const el = document.getElementById(id);
    if (el) {
      el.addEventListener(el.tagName === "SELECT" ? "change" : "input", () => {
        if (id === "inputMode") {
          updateUiMode();
          calculateOptions();
        } else if (
          id === "ticker" &&
          document.getElementById("inputMode").value === "search"
        ) {
          clearTimeout(searchTimeout);
          searchTimeout = setTimeout(() => {
            const tickerVal = document.getElementById("ticker").value.trim();
            if (tickerVal.length > 0) fetchTickerSuggestions();
          }, 500);
        } else if (id === "heatmapMode") {
          setHeatmapMode(el.value);
        } else {
          calculateOptions();
        }
      });
    }
  });

  document.getElementById("ticker").addEventListener("keydown", (e) => {
    if (e.key === "Enter") handleSearch();
  });

  // Modal event listeners
  const modalClose = document.getElementById("modalClose");
  const modalContinue = document.getElementById("modalContinue");
  const modalOverlay = document.getElementById("apiKeyModal");

  if (modalClose) {
    modalClose.addEventListener("click", closeApiKeyModal);
  }

  if (modalContinue) {
    modalContinue.addEventListener("click", closeApiKeyModal);
  }

  if (modalOverlay) {
    modalOverlay.addEventListener("click", (e) => {
      if (e.target === modalOverlay) {
        closeApiKeyModal();
      }
    });
  }

  const d = new Date();
  d.setMonth(d.getMonth() + 1);
  document.getElementById("expiryInputManual").value = d
    .toISOString()
    .split("T")[0];

  updateUiMode();

  // Check if ticker contains "(Demo)" to trigger demo mode on load
  const tickerValue = document.getElementById("ticker").value.toUpperCase();
  if (tickerValue.includes("(DEMO)")) {
    simulateData();
  } else {
    calculateOptions();
  }

  // Restore heatmapMode from URL if present
  const params = new URLSearchParams(window.location.search);
  const urlHeatmapMode = params.get("heatmapMode");
  if (
    urlHeatmapMode &&
    ["dollar", "percent", "value"].includes(urlHeatmapMode)
  ) {
    document.getElementById("heatmapMode").value = urlHeatmapMode;
    currentHeatmapMode = urlHeatmapMode;
  }

  if (typeof makeShareable === "function") {
    makeShareable(calculateOptions);
    updateUiMode();

    const expirySelect = document.getElementById("expirySelect");
    if (expirySelect && expirySelect.children.length > 0) {
      const selectedValue = expirySelect.value;
      const optionExists = Array.from(expirySelect.options).some(
        (opt) => opt.value === selectedValue,
      );
      if (!optionExists) {
        expirySelect.value = expirySelect.options[0].value;
      }
    }

    const strikeSelect = document.getElementById("strikeSelect");
    if (strikeSelect && strikeSelect.children.length > 0) {
      const selectedValue = strikeSelect.value;
      const optionExists = Array.from(strikeSelect.options).some(
        (opt) => opt.value === selectedValue,
      );
      if (!optionExists) {
        strikeSelect.value = strikeSelect.options[0].value;
      }
    }

    calculateOptions();
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initOptions);
} else {
  initOptions();
}

document.addEventListener("htmx:afterSwap", initOptions);
