---
title: "Housing Q Ratio"
description: "The Australian dwelling stock at market value, net of housing debt, against the cost of replacing it, detrended by default."
layout: "housing-q-ratio"
tool_type: "chart"
math: true
---

The Q ratio compares what the housing stock is worth with what it would cost to rebuild. A reading above one means the market values the stock above its replacement cost.

$$
Q_t = \frac{\text{residential land and dwellings}_t - \text{housing credit}_t}{\text{dwellings at replacement cost}_t}
$$

The numerator is the market value of residential land and dwellings, less the housing credit outstanding against it: the equity in the stock. The denominator is the replacement cost of the dwellings themselves. Both come from the Australian Bureau of Statistics household balance sheet, with housing credit from the Reserve Bank of Australia.

Housing carries a long upward trend as land becomes a larger share of value. The default view divides each reading by the running geometric mean of every reading to that date:

$$
\tilde{Q}_t = \frac{Q_t}{\left(\prod_{i=1}^{t} Q_i\right)^{1/t}}
$$

That leaves an oscillator centred near one. Untick "Detrended" to plot the raw ratio.

The detrending mean is built from the history available to that date. The series begins in 1990, well short of a full housing cycle, so the mean reflects the sample so far rather than a complete cycle and the earliest readings are the least settled.

**Data Sources**

- Australian Bureau of Statistics, Australian National Accounts: Finance and Wealth (5232.0), household balance sheet, Table 35, columns "Residential land and dwellings" and "Non-financial - Produced - Fixed assets - Dwellings"
- Reserve Bank of Australia, D2 Lending and Credit Aggregates, owner-occupier and investor housing credit
- Frequency: quarterly, from 1990
