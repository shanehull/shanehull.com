/**
 * Initialize a Chart.js line chart with zoom and pan support
 */
function initChartFromData() {
  const chartElement = document.querySelector("[data-chart]");
  if (!chartElement) return;

  try {
    const config = JSON.parse(chartElement.dataset.chart);
    const canvasId = config.canvasId;

    // Destroy existing chart
    if (window.lineChartInstance) {
      window.lineChartInstance.destroy();
    }

    const ctx = document.getElementById(canvasId).getContext("2d");
    const isMobile = window.innerWidth < 768;
    const legendFontSize = isMobile ? 8 : 12;
    const tickFontSize = isMobile ? 8 : 11;
    const titleFontSize = isMobile ? 9 : 12;

    window.lineChartInstance = new Chart(ctx, {
      type: "line",
      data: {
        labels: config.labels,
        datasets: config.datasets,
      },
      options: {
        responsive: true,
        maintainAspectRatio: true,
        interaction: {
          mode: "index",
          intersect: false,
        },
        plugins: {
          legend: {
            display: true,
            position: "top",
            labels: {
              font: {
                size: legendFontSize,
              },
              padding: isMobile ? 10 : 15,
            },
          },
          zoom: {
            zoom: {
              wheel: {
                enabled: true,
                speed: 0.05,
              },
              pinch: {
                enabled: true,
              },
              mode: "x",
              drag: {
                enabled: true,
                backgroundColor: "rgba(225,225,225,0.3)",
              },
            },
            pan: {
              enabled: false,
            },
          },
        },
        scales: {
          x: {
            ticks: {
              font: {
                size: tickFontSize,
              },
            },
          },
          y: {
            ticks: {
              font: {
                size: tickFontSize,
              },
            },
            title: {
              display: true,
              text: config.yAxisLabel,
              font: {
                size: titleFontSize,
                weight: "bold",
              },
            },
          },
        },
      },
    });
  } catch (e) {
    console.error("Failed to initialize chart:", e);
  }
}

// Initialize on page load
document.addEventListener("DOMContentLoaded", initChartFromData);

// Reinitialize chart when HTMX swaps content
document.body.addEventListener("htmx:afterSwap", function (event) {
  if (event.detail.target.id === "chart-inner") {
    initChartFromData();
  }
});
