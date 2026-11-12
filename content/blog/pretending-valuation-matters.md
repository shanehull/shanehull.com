---
title: Let's Pretend Valuation Matters
date: 2026-11-12
description: An exercise in pricing an Australian housing reversion against earnings and rent.
tags:
  - housing
  - valuation
  - australian-property
  - capital-cycle
  - mean-reversion
---

The exercise prices a reversion. It does not forecast one. It assumes valuation matters, even though the last three decades say otherwise. The question is how far Australian housing prices would have to fall to return to a realistic valuation.

Two measures do the work. Price-to-income compares the house price index with average weekly earnings. Price-to-rent compares the same price index with rent, the investment earnings of the dwelling. Both are at the top of their history.

## The measures

Price-to-income divides the Organisation for Economic Co-operation and Development (OECD) nominal house price index by Australian Bureau of Statistics (ABS) average weekly earnings. Earnings per employee is the denominator. Household income per household drifts with composition, from second earners to household size, and that drift would sit inside a valuation signal.

The numerator is an index, so the level is index points per dollar of weekly earnings. Read it against its own history. There is no fixed threshold. Each reading is ranked against the full history. The rank is the share of observations at or below the value.

Price-to-rent divides the same house price index by the OECD rent price index. The ratio is the inverse of the gross rental yield. It is the housing equivalent of a price-to-earnings multiple: how many years of gross rent buy the dwelling.

The gross measure is the standard long-run comparison. A net cap rate after vacancy, operating costs and depreciation would be lower, but the expense ratio is stable over long periods, so the ranking is similar. The gross ratio is the simpler and more consistent series across five decades. The series begins in 1972. The chart starts in 1990 to align with price-to-income; the earlier observations remain in the CSV and JSON and do not change the percentile or the destinations.

<img src="/blog/images/housing-valuation-2026-q2.svg?v=2026-11-12" alt="Two panels: price to weekly income and price to rent, with 25th and 75th percentile lines, data to 2026-Q2" width="900" />

<p><em>Data to 2026-Q2. Dashed lines mark the 25th and 75th percentiles of the full history. The price-to-rent series begins in 1972; the chart starts in 1990 to align with price-to-income. The earlier observations are in the CSV and JSON and do not change the percentiles or destinations. Sources: OECD and ABS. Live charts: <a href="/tools/house-prices-to-weekly-income/">price to weekly income</a> and <a href="/tools/house-prices-to-rent/">price to rent</a>.</em></p>

## Where we are

| Measure         | Latest | Date    | Percentile |
| --------------- | ------ | ------- | ---------- |
| Price-to-income | 0.1120 | 2026-Q1 | 100th      |
| Price-to-rent   | 133.9  | 2025-Q3 | 98th       |

Price-to-income is at its all-time high. Price-to-rent is just below its 2022 peak. Both earnings measures are at the top of their history.

## The destinations

| Measure         | Destination | Value  | Fall required |
| --------------- | ----------- | ------ | ------------- |
| Price-to-income | 1996 trough | 0.0460 | 59%           |
| Price-to-rent   | 1996 trough | 50.0   | 63%           |

Each trough is the lowest reading in its history. Price-to-income is about 2.4 times its 1996 low. Price-to-rent is about 2.7 times its 1996 low. The two measures agree at about 60%.

The lollipop chart shows the same distances:

<img src="/blog/images/housing-valuation-2026-q2-destinations.svg?v=2026-11-12" alt="Lollipop chart of the price fall required to reach each valuation destination: 59% for price-to-income and 63% for price-to-rent." width="900" />

<p><em>Price fall required to reach each destination.</em></p>

## The ladder

Every row holds earnings and rents at their current levels. Only the price of the dwelling stock moves. The 0% row uses the latest reading for each measure: 2026-Q1 for price-to-income and 2025-Q3 for price-to-rent.

| Price fall  | Price-to-income | Price-to-income percentile | Price-to-rent | Price-to-rent percentile |
| ----------- | --------------- | -------------------------- | ------------- | ------------------------ |
| 0% (latest) | 0.1120          | 100th                      | 133.9         | 98th                     |
| 10%         | 0.1008          | 84th                       | 120.5         | 91st                     |
| 20%         | 0.0896          | 72nd                       | 107.1         | 85th                     |
| 30%         | 0.0784          | 50th                       | 93.7          | 76th                     |
| 40%         | 0.0672          | 25th                       | 80.3          | 58th                     |
| 50%         | 0.0560          | 20th                       | 67.0          | 55th                     |
| 60%         | 0.0448          | 0th                        | 53.6          | 50th                     |

## What the gap means

Price-to-income needs 59%. Price-to-rent needs 63%. The two earnings measures agree. Housing is about 2.4 to 2.7 times its last affordable level.

The land and debt effects sit inside the market price. They are why the price is high relative to both earnings and rent. The two measures capture that. A separate leverage or replacement-cost ratio does not add a valuation signal.

## Time does part of the work

A fall is not the only path. With nominal prices flat, there are several ways to get there:

| Measure         | Assumption                          | Years to the trough |
| --------------- | ----------------------------------- | ------------------- |
| Price-to-income | earnings +3.5% a year               | 26                  |
| Price-to-rent   | rents +4% a year                    | 25                  |
| Price-to-income | 30% fall, then flat, earnings +3.5% | 16                  |

The two earnings measures need about 25 years of flat prices, or a 30% fall and 16 years of earnings growth, to return to their 1996 troughs.

## Caveats

Nominal throughout. Earnings and rents are held at their current levels unless a row states otherwise. It measures distance in today's prices.

The exercise assumes valuation matters. Whether it does is up to you.

Price-to-income uses average weekly earnings, not household disposable income. Price-to-rent is gross. A net cap rate after vacancy, operating costs and depreciation would be lower, but the expense ratio is stable over long periods, so the ranking is similar.

The data is quarterly and published with a lag. The OECD price-to-rent ends 2025-Q3 and the OECD house price index and average weekly earnings end 2026-Q1, so the current quarter is not in both numbers.

This follows [Dousing the Pyre](/blog/dousing-the-pyre/), which argued that policy is delaying the downturn by pushing the cost onto younger buyers. Their mortgages at today's prices let existing owners sell and lenders keep the loans performing. This post prices a destination that policy resists. If that resistance becomes monetisation, the debt behind the bubble is inflated away. The correction arrives through inflation, and nominal house prices do less of the work. The ratios above would still normalise. Just not the way this exercise assumes.

Download the data: <a href="/blog/data/housing-valuation-2026-q2/data.json" download hx-boost="false">JSON</a>, <a href="/blog/data/housing-valuation-2026-q2/series.csv" download hx-boost="false">CSV</a>.
