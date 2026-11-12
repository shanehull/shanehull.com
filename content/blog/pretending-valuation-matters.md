---
title: Let's Pretend Valuation Matters
date: 2026-11-12
description: An exercise in pricing an Australian housing reversion against earnings, rent and replacement cost.
tags:
  - housing
  - valuation
  - australian-property
  - capital-cycle
  - mean-reversion
---

The exercise prices a reversion. It does not forecast one. It assumes valuation matters, even though the last three decades say otherwise. The question is how far Australian housing prices would have to fall to return to a realistic valuation.

Three measures do the work. Price-to-income compares the house price index with average weekly earnings. Price-to-rent compares the same price index with rent, the investment earnings of the dwelling. The Q ratio compares the owner's equity with the replacement cost of the dwellings.

The price-to-income and price-to-rent series are at the top of their history. The Q is high but has turned lower from its 2021 peak. That divergence is the point: the earnings measures say housing is expensive, while the leverage measure says the cycle has already begun to turn.

## The measures

Price-to-income divides the Organisation for Economic Co-operation and Development (OECD) nominal house price index by Australian Bureau of Statistics (ABS) average weekly earnings. Earnings per employee is the denominator. Household income per household drifts with composition, from second earners to household size, and that drift would sit inside a valuation signal.

The numerator is an index, so the level is index points per dollar of weekly earnings. Read it against its own history. There is no fixed threshold. Each reading is ranked against the full history. The rank is the share of observations at or below the value.

Price-to-rent divides the same house price index by the OECD rent price index. The ratio is the inverse of the gross rental yield. It is the housing equivalent of a price-to-earnings multiple: how many years of gross rent buy the dwelling.

The gross measure is the standard long-run comparison. A net cap rate after vacancy, operating costs and depreciation would be lower, but the expense ratio is stable over long periods, so the ranking is similar. The gross price-to-rent ratio is the simpler and more consistent series across five decades. The series begins in 1972. The chart starts in 1990 to align with the Q and price-to-income; the earlier observations remain in the CSV and JSON and do not change the percentile or the destinations.

The Q ratio divides owner equity by the replacement cost of the dwelling structures. Equity is the market value of residential land and dwellings less housing credit. Land is in the numerator. The denominator is the replacement cost of the buildings. A Q of 1 is the mark where equity equals the replacement cost of the structures. It is a theoretical reference, not an empirical fair value.

Because land is in the numerator and not the denominator, the Q has sat above 1 for the whole sample. The raw Q is 2.90. The 1995 trough was 1.71. The 2021 peak was 3.37. The old Q divided equity by the market value of the dwellings. The missing piece is the replacement cost. The Q now measures equity against the cost of rebuilding the structures.

The raw Q trends, so a percentile against its own history is mechanically high. The detrended series divides each reading by the running geometric mean of every reading to that date. That leaves an oscillator centred near one and is the cleaner cyclical read.

<img src="/blog/images/housing-valuation-2026-q2.svg?v=2026-11-12" alt="Four panels: price to weekly income, price to rent, the raw housing Q ratio and the detrended housing Q ratio, with 25th and 75th percentile lines, data to 2026-Q2" width="900" />

<p><em>Data to 2026-Q2. Dashed lines mark the 25th and 75th percentiles of the full history. Sources: OECD, ABS and the Reserve Bank of Australia (RBA). The price-to-rent series begins in 1972; the chart starts in 1990 to align with the other measures. The earlier observations are in the CSV and JSON and do not change the percentiles or destinations. Live charts: <a href="/tools/house-prices-to-weekly-income/">price to weekly income</a>, <a href="/tools/house-prices-to-rent/">price to rent</a> and <a href="/tools/housing-q-ratio/">Q ratio</a>.</em></p>

## Where we are

| Measure         | Latest | Date    | Percentile |
| --------------- | ------ | ------- | ---------- |
| Price-to-income | 0.1120 | 2026-Q1 | 100th      |
| Price-to-rent   | 133.9  | 2025-Q3 | 98th       |
| Raw Q           | 2.902  | 2026-Q2 | 96th       |
| Detrended Q     | 1.319  | 2026-Q2 | 93rd       |

The two earnings measures are at their records. Price-to-income is at its all-time high. Price-to-rent is just below its 2022 peak. The Q is at the 96th percentile, but it is 14% below its 2021 peak. Replacement costs have risen 48% since 2021 against a 27% rise in equity, so the Q has turned lower while the earnings measures have kept rising.

The raw Q of 2.902 means owner equity is about 2.9 times the replacement cost of the dwelling structures. Land is about 74% of the combined value, so the Q is mostly a land and leverage measure. The detrended Q is 1.319, above its running mean.

## The destinations

| Measure         | Destination | Value  | Fall required |
| --------------- | ----------- | ------ | ------------- |
| Price-to-income | 1996 trough | 0.0460 | 59%           |
| Price-to-rent   | 1996 trough | 50.0   | 63%           |
| Raw Q           | 1995 trough | 1.707  | 32%           |
| Raw Q           | Q=1 reference | 1.000  | 52%           |
| Detrended Q     | 1995 trough | 0.935  | 23%           |

Each trough is the lowest reading in its history. Price-to-income is about 2.4 times its 1996 low. Price-to-rent is about 2.7 times its 1996 low. The two earnings measures agree at about 60%.

The raw Q has two destinations. Q=1 is the theoretical mark where equity equals the replacement cost of the structures. The 1995 trough is the last leverage trough. The raw Q needs 52% to reach Q=1, or 32% to reach its 1995 trough. The detrended Q needs 23%.

The Q's lower falls are not a contradiction. The numerator is equity, and land is in the numerator. A price fall hits equity harder than it hits the asset, so the Q falls faster than the earnings measures. The Q is a leverage measure, not a valuation measure.

The detrended trough was measured against a mean built on about six years of data; today's mean is built on 36. So 0.935 carries more sample uncertainty than today's 1.319.

The lollipop chart shows the same distances:

<img src="/blog/images/housing-valuation-2026-q2-destinations.svg?v=2026-11-12" alt="Lollipop chart of the price fall required to reach each valuation destination: 59% for price-to-income, 63% for price-to-rent, 52% for the raw Q at Q=1, 32% for the raw Q at its 1995 trough, and 23% for the detrended Q at its 1995 trough." width="900" />

<p><em>Price fall required to reach each destination. The raw Q has two: its 1995 trough and Q=1.</em></p>

## The ladder

Every row holds earnings, rents, housing credit and replacement costs at their current levels. Only the price of the dwelling stock moves. The 0% row uses the latest reading for each measure: 2026-Q1 for price-to-income, 2025-Q3 for price-to-rent and 2026-Q2 for Q.

| Price fall  | Price-to-income | Price-to-rent | Raw Q | Detrended Q |
| ----------- | --------------- | ------------- | ----- | ----------- |
| 0% (latest) | 0.1120          | 133.9         | 2.902 | 1.319       |
| 10%         | 0.1008          | 120.5         | 2.533 | 1.151       |
| 20%         | 0.0896          | 107.1         | 2.164 | 0.984       |
| 30%         | 0.0784          | 93.7          | 1.795 | 0.816       |
| 40%         | 0.0672          | 80.3          | 1.426 | 0.648       |
| 50%         | 0.0560          | 67.0          | 1.057 | 0.480       |
| 60%         | 0.0448          | 53.6          | 0.687 | 0.312       |

The raw Q's 50% row is 1.057, close to Q=1. The detrended Q's 20% row is 0.984, close to its 1995 trough. The earnings measures are still far from their 1996 lows at a 60% price fall.

## What the gap means

Price-to-income needs 59%. Price-to-rent needs 63%. The raw Q needs 52% to reach Q=1, or 32% to reach its 1995 trough. The detrended Q needs 23%.

The Q's Q=1 is close to the earnings measures. The Q's 1995 trough and the detrended Q are the nearer leverage anchors. The difference is the numerator. Equity is the residual after debt, so a price fall is amplified. Housing credit is 21.4% of the dwelling stock, so equity is 78.6% of value and a price fall is amplified 1.27 times. A 32% fall cuts the Q numerator by 41%; the 52% fall to Q=1 cuts it by 66%.

The detrended Q needs less again because its running mean falls with the numerator. The mean absorbs part of the move.

The Q is high but no longer at its peak. The earnings measures are at their records. The valuation story is in the earnings measures. The Q is the transmission channel: credit growth and replacement costs have already begun to compress the equity premium.

## Time does part of the work

A fall is not the only path. With nominal prices flat, there are several ways to get there:

| Measure         | Assumption                                  | Years to the trough |
| --------------- | ------------------------------------------- | ------------------- |
| Price-to-income | earnings +3.5% a year                       | 26                  |
| Price-to-rent   | rents +4% a year                            | 25                  |
| Raw Q           | credit +7.4%, replacement cost +3.5% a year | 8.7                 |
| Price-to-income | 30% fall, then flat, earnings +3.5%         | 16                  |

The raw Q gets there sooner because two forces work against the owner without any price fall. Housing credit keeps rising, which shrinks equity, and replacement costs keep rising, which grows the denominator. Price-to-income and price-to-rent have only earnings and rents on their side.

## Caveats

Nominal throughout. Earnings, rents, credit and replacement costs are held at their current levels unless a row states otherwise. It measures distance in today's prices.

The exercise assumes valuation matters. Whether it does is up to you.

The data is quarterly and published with a lag. The OECD price-to-rent ends 2025-Q3, the OECD house price index and average weekly earnings end 2026-Q1, and the balance sheet ends 2026-Q2, so the current quarter is not in all three numbers.

The Q denominator is the replacement cost of the dwelling structures, not the full property. Land is in the numerator. A full property Q with land in the denominator is an identity and tells us nothing. The Q is the equity premium over the cost of the structure.

This follows [Dousing the Pyre](/blog/dousing-the-pyre/), which argued that policy is delaying the downturn by pushing the cost onto younger buyers. Their mortgages at today's prices let existing owners sell and lenders keep the loans performing. This post prices a destination that policy resists. If that resistance becomes monetisation, the debt behind the bubble is inflated away. The correction arrives through inflation, and nominal house prices do less of the work. The ratios above would still normalise. Just not the way this exercise assumes.

Download the data: <a href="/blog/data/housing-valuation-2026-q2/data.json" download hx-boost="false">JSON</a>, <a href="/blog/data/housing-valuation-2026-q2/series.csv" download hx-boost="false">CSV</a>.
