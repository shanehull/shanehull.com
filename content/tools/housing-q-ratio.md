---
title: "Housing Q Ratio"
description: "The owner's equity in the Australian dwelling stock against the replacement cost of the dwellings."
layout: "housing-q-ratio"
tool_type: "chart"
math: true
---

The Q ratio compares the owner's equity in the housing stock with the cost of replacing the dwelling structures. Land is in the numerator. The denominator is the replacement cost of the buildings.

$$
Q_t = \frac{\text{residential land and dwellings}_t - \text{housing credit}_t}{\text{dwellings at replacement cost}_t}
$$

The numerator is the market value of residential land and dwellings less the housing credit outstanding against it: the owner's equity. The denominator is the replacement cost of the dwellings themselves. Both come from the Australian Bureau of Statistics household balance sheet, with housing credit from the Reserve Bank of Australia.

A Q of 1 is the mark where equity equals the replacement cost of the structures. It is a theoretical reference, not an empirical fair value. Because land is in the numerator and not the denominator, the ratio has sat above 1 for the whole sample. The current raw reading is 2.90. The 1995 trough was 1.71. The 2021 peak was 3.37.

The old Q divided equity by the market value of the dwellings. The missing piece is the replacement cost. The Q now measures equity against the cost of rebuilding the structures. It rises when land values rise faster than construction costs, and falls when credit grows or replacement costs rise faster than prices.

The ratio is quarterly, from 1990. The chart is detrended by default. Untick "Detrended" to plot the raw ratio. The quartile lines mark the 25th and 75th percentiles of the selected range.

**Data Sources**

- Australian Bureau of Statistics, Australian National Accounts: Finance and Wealth (5232.0), household balance sheet, "Residential land and dwellings" (market value)
- Australian Bureau of Statistics, Australian System of National Accounts (5204.0), household balance sheet, "Dwellings" (replacement cost)
- Reserve Bank of Australia, D2 Lending and Credit Aggregates, owner-occupier and investor housing credit
- Frequency: quarterly, from 1990
