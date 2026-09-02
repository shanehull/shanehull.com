---
title: "Misesian Stationarity Index"
description: "Visualize the Misesian Stationarity Index based on stock market equity and net worth data from the Federal Reserve."
layout: "msindex"
tool_type: "chart"
math: true
---

The Misesian Stationarity Index is a financial metric that measures the deviation of the equity-to-net-worth ratio from its geometric mean over time. This indicator can help identify periods of economic anomaly or instability.

$$
\text{MSI}_t=\frac{E_t\,/\,NW_t}{\left(\prod_{i=1}^{t}E_i/NW_i\right)^{1/t}}
$$

The calculation divides each period's equity-to-net-worth ratio by the geometric mean of all previous ratios, providing a stationary measure that adjusts for long-term trends.

**Data Source:** U.S. Federal Reserve Economic Data (FRED)

- Corporate Equity: NCBCEL
- Net Worth: TNWMVBSNNCB
- Frequency: Quarterly
