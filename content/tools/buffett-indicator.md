---
title: "Buffett Indicator"
description: "The Buffett Indicator, adjusted for federal debt by default, on the view that the Fed will eventually monetize it."
layout: "buffett-indicator"
tool_type: "chart"
math: true
---

The Buffett Indicator measures whether the US stock market is expensive or cheap relative to the size of the economy:

$$
\text{Buffett Indicator}=\frac{\text{total market capitalization}}{\text{GDP}}\times 100
$$

Warren Buffett has called it "probably the best single measure of where valuations stand at any given moment."

This chart adjusts that ratio by adding federal debt to the numerator, and it does so by default:

$$
\text{Adjusted Buffett Indicator}=\frac{\text{total market capitalization}+\text{federal debt}}{\text{GDP}}\times 100
$$

The adjustment follows the framing popularised by Luke Gromen. Federal debt is added because the Fed will eventually have to monetize it. Monetization creates money that has to go somewhere, and equities are a primary destination. Leaving debt out of the numerator understates how much future money creation is already priced into stocks. On that view, every time the adjusted ratio has reached 100%, stocks have subsequently underperformed gold.

Both versions are available. The debt switch is on by default. Turn it off to chart the classic market-cap-only ratio.

When the ratio is high, it suggests stocks are overpriced. When it's low, the market may be undervalued. The average line shows the historical mean, use it as a reference point to see if current valuations are above or below the norm.

**Data Source:** U.S. Federal Reserve Economic Data (FRED)

- **Numerator:** Nonfinancial Corporate Business; Corporate Equities (NCBEILQ027S)
- **Debt adjustment:** Federal Debt: Total Public Debt (GFDEBTN), added to the numerator
- **Denominator:** Gross Domestic Product (GDP)
- **Frequency:** Quarterly
- **Coverage:** GFDEBTN begins in 1966, so both modes start there. The window stays fixed when the debt switch is toggled.
