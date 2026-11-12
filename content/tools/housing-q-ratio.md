---
title: "Housing Q Ratio"
description: "Australian residential land at market plus dwellings at replacement cost, net of housing debt, over dwelling replacement cost, detrended by default."
layout: "housing-q-ratio"
tool_type: "chart"
math: true
---

The Q ratio compares the ABS balance sheet value of residential land and dwellings, less housing credit, with the replacement cost of the dwellings. The balance sheet values land at market and dwellings at replacement cost, so the ratio is a hybrid: land is in the numerator, not the denominator. A Q of 1 is the theoretical mark where the balance sheet equity equals dwelling replacement cost. It is not an empirical fair value.

$$
Q_t = \frac{\text{residential land and dwellings}_t - \text{housing credit}_t}{\text{dwellings at replacement cost}_t}
$$

The numerator is the ABS balance sheet value of residential land and dwellings, less the housing credit outstanding against it. Land is at market; dwellings are at replacement cost. The denominator is the replacement cost of the dwellings. Both come from the Australian Bureau of Statistics household balance sheet, with housing credit from the Reserve Bank of Australia.

Housing carries a long upward trend as land becomes a larger share of value. The default view divides each reading by the running geometric mean of every reading to that date:

$$
\tilde{Q}_t = \frac{Q_t}{\left(\prod_{i=1}^{t} Q_i\right)^{1/t}}
$$

That leaves an oscillator centred near one. Untick "Detrended" to plot the raw ratio.

The detrending mean is built from the history available to that date. The series begins in 1990, well short of a full housing cycle, so the mean reflects the sample so far rather than a complete cycle and the earliest readings are the least settled.

The quartile lines mark the 25th and 75th percentiles of the selected range, so the current reading can be read against its own recent distribution.

**Data Sources**

- Australian Bureau of Statistics, Australian National Accounts: Finance and Wealth (5232.0), household balance sheet, Table 35, columns "Residential land and dwellings" and "Non-financial - Produced - Fixed assets - Dwellings"
- Reserve Bank of Australia, D2 Lending and Credit Aggregates, owner-occupier and investor housing credit
- Frequency: quarterly, from 1990
