---
title: Let's Pretend Valuation Matters
date: 2026-11-12
description: An exercise in pricing an Australian housing reversion against household income and rent.
tags:
  - housing
  - valuation
  - australian-property
  - capital-cycle
  - mean-reversion
---

The exercise prices a reversion. It does not forecast one. It assumes valuation matters, even though the last three decades say otherwise. The question is how far Australian housing prices would have to fall to return to a realistic valuation.

Two measures do the work. Price-to-income compares the house price index with household disposable income per capita. Price-to-rent compares the same price index with rent, the investment earnings of the dwelling. Both are at the top of their history, but the rent measure is more stretched.

## The measures

Price-to-income divides the Organisation for Economic Co-operation and Development (OECD) nominal house price index by Australian Bureau of Statistics (ABS) household gross disposable income per capita. The income is seasonally adjusted and divided by estimated resident population.

Disposable income is after tax and transfers, before housing costs. It is the standard affordability denominator. It includes rent received by household landlords and excludes corporate, institutional and foreign investors' income. The series begins in 1981.

Price-to-rent divides the same house price index by the OECD rent price index. The ratio is the inverse of the gross rental yield. It is the housing equivalent of a price-to-earnings multiple: how many years of gross rent buy the dwelling. The series begins in 1972.

The gross measure is the standard long-run comparison. A net cap rate after vacancy, operating costs and depreciation would be lower, but the expense ratio is stable over long periods, so the ranking is similar.

<img src="/blog/images/housing-valuation-2026-q2.svg?v=2026-11-12" alt="Two panels: price to household income and price to rent, with 25th and 75th percentile lines, data to 2026-Q1" width="900" />

<p><em>Data to 2026-Q1. Dashed lines mark the 25th and 75th percentiles of the full history. Price-to-income begins in 1981 and price-to-rent in 1972. The figure starts in 1981 to align the two; the earlier rent observations are in the CSV and JSON and do not change the percentiles or destinations. Sources: OECD and ABS. Live charts: <a href="/tools/house-prices-to-income/">price to household income</a> and <a href="/tools/house-prices-to-rent/">price to rent</a>.</em></p>

## Where we are

| Measure              | Latest | Date    | Percentile |
| -------------------- | ------ | ------- | ---------- |
| Price-to-income      | 0.0107 | 2026-Q1 | 100th      |
| Price-to-rent        | 133.9  | 2025-Q3 | 98th       |

Price-to-income is at its all-time high. Price-to-rent is just below its 2022 peak. Both earnings measures are at the top of their history.

## The destinations

| Measure         | Destination | Value  | Fall required |
| --------------- | ----------- | ------ | ------------- |
| Price-to-income | 1996 trough | 0.005345 | 50%         |
| Price-to-income | 1983 low    | 0.004616 | 57%         |
| Price-to-rent   | 1996 trough | 50.0   | 63%           |

The 1983 low is the all-time low in the income series. The 1996 trough is the last trough. The rent measure has one trough, in 1996.

Price-to-income is about 2.0 times its 1996 low and 2.3 times its 1983 low. Price-to-rent is about 2.7 times its 1996 low. The rent measure is more stretched than the income measure.

The lollipop chart shows the same distances:

<img src="/blog/images/housing-valuation-2026-q2-destinations.svg?v=2026-11-12" alt="Lollipop chart of the price fall required to reach each valuation destination: 50% for price-to-income at its 1996 trough, 57% at its 1983 low, and 63% for price-to-rent at its 1996 trough." width="900" />

<p><em>Price fall required to reach each destination.</em></p>

## The ladder

Every row holds disposable income and rents at their current levels. Only the price of the dwelling stock moves. The 0% row uses 2026-Q1 for price-to-income and 2025-Q3 for price-to-rent.

| Price fall  | Price-to-income | Price-to-income percentile | Price-to-rent | Price-to-rent percentile |
| ----------- | --------------- | -------------------------- | ------------- | ------------------------ |
| 0% (latest) | 0.0107          | 100th                      | 133.9         | 98th                     |
| 10%         | 0.0097          | 87th                       | 120.5         | 91st                     |
| 20%         | 0.0086          | 77th                       | 107.1         | 85th                     |
| 30%         | 0.0075          | 52nd                       | 93.7          | 76th                     |
| 40%         | 0.0064          | 46th                       | 80.3          | 58th                     |
| 50%         | 0.0054          | 18th                       | 67.0          | 55th                     |
| 60%         | 0.0043          | 0th                        | 53.6          | 50th                     |

The income measure's 50% row is 0.0054, at the 1996 trough. Its 60% row is 0.0043, below the 1983 low. The rent measure's 60% row is 53.6, near its 1996 trough. The rent measure needs more of a fall to reach its last trough.

## What the gap means

Price-to-income needs 50% to its 1996 trough, or 57% to its 1983 low. Price-to-rent needs 63% to its 1996 trough.

The two measures no longer agree at 60%. The rent measure is more stretched. That is a meaningful divergence: household income is high relative to house prices by the standards of the 1980s and 1990s, but rent is even lower relative to prices.

The land and debt effects sit inside the market price. They are why the price is high relative to both income and rent. The two measures capture that. A separate leverage or replacement-cost ratio does not add a valuation signal.

## Time does part of the work

A fall is not the only path. With nominal prices flat:

| Measure         | Assumption                          | Years to the trough |
| --------------- | ----------------------------------- | ------------------- |
| Price-to-income | disposable income +3.5% a year      | 20                  |
| Price-to-rent   | rents +4% a year                    | 25                  |
| Price-to-income | 30% fall, then flat, income +3.5%   | 7                   |

The income measure needs about 20 years of flat prices, or a 30% fall and 7 years of income growth, to return to its 1996 trough. The rent measure needs about 25 years of flat prices, because rents grow more slowly against the price.

## Caveats

Nominal throughout. Income and rents are held at their current levels unless a row states otherwise. It measures distance in today's prices.

The exercise assumes valuation matters. Whether it does is up to you.

Price-to-income uses household disposable income per capita. It is after tax and transfers, before housing costs. It includes rent received by household landlords and excludes corporate, institutional and foreign investors' income. Price-to-rent is gross. A net cap rate after vacancy, operating costs and depreciation would be lower, but the expense ratio is stable over long periods, so the ranking is similar.

The data is quarterly and published with a lag. Price-to-income ends 2026-Q1 and price-to-rent ends 2025-Q3.

This follows [Dousing the Pyre](/blog/dousing-the-pyre/), which argued that policy is delaying the downturn by pushing the cost onto younger buyers. Their mortgages at today's prices let existing owners sell and lenders keep the loans performing. This post prices a destination that policy resists. If that resistance becomes monetisation, the debt behind the bubble is inflated away. The correction arrives through inflation, and nominal house prices do less of the work. The ratios above would still normalise. Just not the way this exercise assumes.

Download the data: <a href="/blog/data/housing-valuation-2026-q2/data.json" download hx-boost="false">JSON</a>, <a href="/blog/data/housing-valuation-2026-q2/series.csv" download hx-boost="false">CSV</a>.
