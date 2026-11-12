---
title: Let's Pretend Reality Matters
date: 2026-11-12
description: An exercise in pricing an Australian housing reversion against household income and rent.
tags:
  - housing
  - housing-affordability
  - australian-property
  - capital-cycle
  - mean-reversion
---

The exercise prices a reversion to reality. It does not forecast one. It assumes the historical relationships between house prices, income and rent still matter, even though the last three decades say otherwise. The question is how far prices would have to fall to reach the levels of the 1996 trough and the 1983 low.

Two measures do the work. Price-to-income compares the house price index with household disposable income per capita. Price-to-rent compares the same price index with rent, the investment earnings of the dwelling. They sit near the top of their history, but the rent measure needs a larger fall to reach its 1996 trough.

## The measures

Price-to-income divides the Organisation for Economic Co-operation and Development (OECD) nominal house price index by Australian Bureau of Statistics (ABS) household gross disposable income per capita. The income is seasonally adjusted and divided by estimated resident population.

Disposable income is after tax and transfers, before housing costs. It is the standard affordability denominator used by the OECD and the Reserve Bank of Australia. It includes rent received by household landlords and excludes corporate, institutional and foreign investors' income. The series begins in 1981.

Price-to-rent divides the same house price index by the OECD rent price index. The ratio is the inverse of the gross rental yield. A higher ratio means more years of gross rent to buy the dwelling. The series begins in 1972.

The gross measure is the usual long-run comparison. A net cap rate after vacancy, operating costs and depreciation would be lower. The ranking is similar if the expense ratio is broadly stable over long periods, which this exercise assumes.

<img src="/blog/images/housing-valuation-2026-q2.svg?v=2026-11-12" alt="Two panels: price to household income and price to rent, with 25th and 75th percentile lines, income to 2026-Q1 and rent to 2025-Q3" width="900" />

<p><em>Data to 2026-Q1 for price-to-income and 2025-Q3 for price-to-rent. Dashed lines mark the 25th and 75th percentiles of the period shown. Price-to-income begins in 1981 and price-to-rent in 1972; the figure starts in 1981 to align the two. The earlier rent observations are in the CSV and JSON. Sources: OECD and ABS. Live charts: <a href="/tools/house-prices-to-income/">price to household income</a> and <a href="/tools/house-prices-to-rent/">price to rent</a>.</em></p>

## Where we are

| Measure         | Latest | Date    | Percentile |
| --------------- | ------ | ------- | ---------- |
| Price-to-income | 0.0107 | 2026-Q1 | 100th      |
| Price-to-rent   | 133.9  | 2025-Q3 | 98th       |

Price-to-income is at its series high. Price-to-rent is about 6% below its 2022 peak. Both sit near the top of their history.

## Why the old relationship

The pre-2000 multiples are the reference because the rules changed after that period. Financial deregulation and the credit boom let household debt grow much faster than income from the 1990s. Negative gearing and the 50% capital gains tax discount (from 1999) made investment property more attractive. Lower nominal interest rates raised the price a given income could service. These changes stretched the relationship between prices, income and rent. The exercise uses the pre-2000 relationship because it is the last period before the stretch. That is the sense in which it is reality. It is the old regime. It is not a natural constant.

## The destinations

Price-to-income needs a 50% fall to its 1996 trough and a 57% fall to its 1983 low. Price-to-rent needs a 63% fall to its 1996 trough.

<img src="/blog/images/housing-valuation-2026-q2-destinations.svg?v=2026-11-12" alt="Lollipop chart of the price fall required to reach each destination: 50% for price-to-income at its 1996 trough, 57% at its 1983 low, and 63% for price-to-rent at its 1996 trough." width="900" />

<p><em>Price fall required to reach each destination.</em></p>

The 1983 low is the all-time low in the income series. The 1996 trough is the last major trough before the current high, and the common destination for both measures. The rent series' all-time low is 1972, at the start of its history.

Price-to-income is about 2.0 times its 1996 low and 2.3 times its 1983 low. Price-to-rent is about 2.7 times its 1996 low. The rent measure is more stretched.

## The ladder

Every row holds disposable income and rents at their current levels. Only the price of the dwelling stock moves. The 0% row uses 2026-Q1 for price-to-income and 2025-Q3 for price-to-rent. Percentiles use the full history of each series: income from 1981, rent from 1972.

| Price fall  | Price-to-income | Price-to-income percentile | Price-to-rent | Price-to-rent percentile |
| ----------- | --------------- | -------------------------- | ------------- | ------------------------ |
| 0% (latest) | 0.0107          | 100th                      | 133.9         | 98th                     |
| 10%         | 0.0097          | 87th                       | 120.5         | 91st                     |
| 20%         | 0.0086          | 77th                       | 107.1         | 85th                     |
| 30%         | 0.0075          | 52nd                       | 93.7          | 76th                     |
| 40%         | 0.0064          | 46th                       | 80.3          | 58th                     |
| 50%         | 0.0054          | 18th                       | 67.0          | 55th                     |
| 60%         | 0.0043          | 0th                        | 53.6          | 50th                     |

The income measure's 50% row is 0.0054, just above the 1996 trough. Its 60% row is 0.0043, below the 1983 low. The rent measure's 60% row is 53.6, near its 1996 trough. The rent measure needs more of a fall to reach its last major trough.

## Time does part of the work

A fall is not the only path. With nominal prices flat:

| Measure         | Assumption                        | Years to the trough |
| --------------- | --------------------------------- | ------------------- |
| Price-to-income | disposable income +3.5% a year    | 20                  |
| Price-to-rent   | rents +4% a year                  | 25                  |
| Price-to-income | 30% fall, then flat, income +3.5% | 10                  |

The income measure needs about 20 years of flat prices, or a 30% fall and 10 years of income growth, to return to its 1996 trough. The rent measure needs about 25 years of flat prices at 4% rent growth, because the gap to its 1996 trough is larger.

## Caveats

Nominal throughout. In the ladder, income and rents are held at their current levels; the time estimates state their growth assumptions. The exercise measures distance in today's prices.

The exercise treats the historical relationships as a useful guide. It does not require prices to return there. The trough is a reference point. The likely mechanism is debt: the price falls relative to income and rent when credit stops expanding or becomes more expensive.

The data is quarterly and published with a lag. Price-to-income ends 2026-Q1 and price-to-rent ends 2025-Q3.

This follows [Dousing the Pyre](/blog/dousing-the-pyre/), which argued that policy is delaying the downturn by pushing the cost onto younger buyers. Their mortgages at today's prices let existing owners sell and lenders keep the loans performing. This post prices a destination that policy resists. If that resistance becomes monetisation, the debt behind the bubble is inflated away. A correction through inflation would do less of the work through nominal house prices. The numbers above are an exercise. They do not forecast. They would look different in that scenario, if at all.

Download the data: <a href="/blog/data/housing-valuation-2026-q2/data.json" download hx-boost="false">JSON</a>, <a href="/blog/data/housing-valuation-2026-q2/data.csv" download hx-boost="false">CSV</a>.
