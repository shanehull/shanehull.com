---
title: Let's Pretend Valuation Matters
date: 2026-11-12
description: An exercise in pricing an Australian housing reversion to a realistic valuation.
tags:
  - housing
  - valuation
  - australian-property
  - capital-cycle
  - mean-reversion
---

The exercise prices a reversion. It does not forecast one. It assumes valuation matters, even though the last three decades say otherwise. The question is how far Australian housing prices would have to fall to return to a realistic valuation.

The market values a house against the one next door, the rent it earns, the cost of building it, and the credit available to buy it. The measures here use earnings and replacement cost, anchored to their own history rather than today's relative prices. That is the contrarian part: asking what it would cost for prices to return to those historical anchors.

Whether they fall is a separate question.

Two measures do the work. Price-to-income compares the house price index with average weekly earnings. The Q ratio compares the ABS balance sheet value of residential land and dwellings, less housing credit, with the replacement cost of the dwellings.

The Q series starts in 1990 and price-to-income in 1995. The 1995-96 trough is the reference point. It is the last trough in the data, but the cycle is longer than the sample, so the previous trough may sit outside it.

## The measures

Price-to-income divides the Organisation for Economic Co-operation and Development (OECD) nominal house price index by Australian Bureau of Statistics (ABS) average weekly earnings. Earnings per employee is the denominator. Household income per household drifts with composition, from second earners to household size, and that drift would sit inside a valuation signal.

The numerator is an index, so the level is index points per dollar of weekly earnings. Read it against its own history. There is no fixed threshold. Each reading is ranked against the full history. The rank is the share of observations at or below the value.

The Q ratio divides the ABS balance sheet value of residential land and dwellings, less the housing credit outstanding against it, by the current replacement cost of the dwellings. The balance sheet values land at market and dwellings at replacement cost, so the numerator is a mixed value. Two choices matter. The numerator is equity, since debt is fixed in nominal terms and the owner's exposure is the residual. The denominator is replacement cost, the Tobin's q question: is it cheaper to buy the existing stock or to build new?

A raw Q of 1 is the theoretical mark where that balance sheet equity equals the replacement cost of the dwellings. It is not an empirical fair value. Because land is in the numerator and not the denominator, the ratio has sat above 1 for the whole sample. The 1995 trough was 1.710. Above 1, the equity carries a premium against dwelling replacement cost. Below 1, a discount.

The ratio is then divided by its running geometric mean, which removes the multi-decade drift and leaves a cycle. The mean is built from the history to each date. The series starts in 1990, short of a full housing cycle, so the earliest readings are the least settled.

<img src="/blog/images/housing-valuation-2026-q2.svg" alt="Three panels: price to weekly income, the raw Q ratio and the detrended Q ratio, with 25th and 75th percentile lines, data to 2026-Q2" width="900" />

<p><em>Data to 2026-Q2. Dashed lines mark the 25th and 75th percentiles of the full history. Sources: OECD, ABS and the Reserve Bank of Australia (RBA). Live charts: <a href="/tools/housing-q-ratio/">Q ratio</a> and <a href="/tools/house-prices-to-weekly-income/">price to weekly income</a>.</em></p>

## Where we are

| Measure           | Latest | Date    | Percentile |
| ----------------- | ------ | ------- | ---------- |
| Price-to-income   | 0.1120 | 2026-Q1 | 100th      |
| Raw Q ratio       | 3.006  | 2026-Q2 | 96th       |
| Detrended Q ratio | 1.375  | 2026-Q2 | 95th       |

The raw Q of 3.006 means the balance sheet value of land and dwellings, less housing credit, is about three times the replacement cost of the dwellings. Land is about 74% of that combined value, so the Q is mostly a land and leverage measure.

The raw Q trends, so a percentile against its own history is mechanically high. The detrended series strips that out and is the cleaner cyclical read.

Price-to-income is at its all-time high. The Q ratio peaked at 3.163 in 2021. It is now 3.006 and has turned lower. Replacement costs have risen 34% since then against a 27% rise in housing equity, and the construction boom of 2021-22 has caught up with the denominator.

## The destinations

Nominal prices ground higher by about 15% from 1989 to 1996, roughly 2% a year, while earnings kept climbing. The Q ratio fell through that stretch and bottomed in 1995-Q4. Price-to-income bottomed three quarters later, in 1996-Q3. Its earnings series starts in 1994, so it cannot see earlier troughs.

| Measure           | Destination | Value  | Fall required |
| ----------------- | ----------- | ------ | ------------- |
| Price-to-income   | 1996 trough | 0.0460 | 59%           |
| Raw Q ratio       | Q=1 reference | 1.000  | 52%           |
| Raw Q ratio       | 1995 trough | 1.710  | 34%           |
| Detrended Q ratio | 1995 trough | 0.939  | 25%           |

Each trough is the lowest reading in its history. Price-to-income is about 2.4 times its 1996 low.

The detrended trough was measured against a mean built on about six years of data; today's mean is built on 36. So 0.939 carries more sample uncertainty than today's 1.375.

The lollipop chart shows the same distances from zero:

<img src="/blog/images/housing-valuation-2026-q2-destinations.svg" alt="Lollipop chart of the price fall required to reach each valuation destination: 59% for price-to-income, 52% for the raw Q at Q=1, 34% for the raw Q at its 1995 trough, and 25% for the detrended Q at its 1995 trough." width="900" />

<p><em>Price fall required to reach each destination. The raw Q has two: its 1995 trough and Q=1.</em></p>

## The ladder

Every row holds earnings, housing credit and replacement costs at their current levels. Only the price of the dwelling stock moves. The 0% row uses the latest reading for each measure: 2026-Q1 for price-to-income and 2026-Q2 for Q. In the raw Q column, 1 is the Q=1 reference, where the balance sheet equity equals dwelling replacement cost. Above 1 is a premium and below 1 is a discount.

| Price fall  | Price-to-income | Price-to-income Percentile | Raw Q | Detrended Q |
| ----------- | --------------- | -------------------------- | ----- | ----------- |
| 0% (latest) | 0.1120          | 100th                      | 3.006 | 1.375       |
| 10%         | 0.1008          | 84th                       | 2.624 | 1.201       |
| 20%         | 0.0896          | 71st                       | 2.241 | 1.027       |
| 30%         | 0.0784          | 49th                       | 1.859 | 0.853       |
| 40%         | 0.0672          | 24th                       | 1.477 | 0.679       |
| 50%         | 0.0560          | 20th                       | 1.094 | 0.504       |
| 60%         | 0.0448          | 0th                        | 0.712 | 0.329       |

## What the gap means

Price-to-income needs 59%. The raw Q needs 52% to reach Q=1, or 34% to reach its 1995 trough. The detrended Q needs 25%. Debt explains why both raw Q destinations need a smaller price fall than price-to-income. The running mean explains why the detrended Q needs less than the raw Q. Housing credit is 21.4% of the dwelling stock, so equity is 78.6% of value and a price fall is amplified 1.27 times. A 34% fall cuts the Q numerator by 43%; the 52% fall to Q=1 cuts it by 66%.

The detrended Q needs less again because its running mean falls with the numerator. The mean absorbs part of the move.

## Time does part of the work

A fall is not the only path. With nominal prices flat, there are several ways to get there:

| Measure         | Assumption                                  | Years to the trough |
| --------------- | ------------------------------------------- | ------------------- |
| Price-to-income | earnings +3.5% a year                       | 26                  |
| Raw Q ratio     | credit +7.4%, replacement cost +3.5% a year | 8.7                 |
| Price-to-income | 30% fall, then flat, earnings +3.5%         | 16                  |

The raw Q gets there sooner because two forces work against the owner without any price fall. Housing credit keeps rising, which shrinks equity, and replacement costs keep rising, which grows the denominator. Price-to-income has only earnings on its side.

## Caveats

Nominal throughout. Earnings, credit and replacement costs are held at their current levels unless a row states otherwise. It measures distance in today's prices.

The exercise assumes valuation matters. Whether it does is up to you.

The data is quarterly and published with a lag. The OECD index ends 2026-Q1 and the balance sheet ends 2026-Q2, so the current quarter is not in these numbers.

This follows [Dousing the Pyre](/blog/dousing-the-pyre/), which argued that policy is delaying the downturn by pushing the cost onto younger buyers. Their mortgages at today's prices let existing owners sell and lenders keep the loans performing. This post prices a destination that policy resists. If that resistance becomes monetisation, the debt behind the bubble is inflated away. The correction arrives through inflation, and nominal house prices do less of the work. The ratios above would still normalise. Just not the way this exercise assumes.

Download the data: <a href="/blog/data/housing-valuation-2026-q2/data.json" download hx-boost="false">JSON</a>, <a href="/blog/data/housing-valuation-2026-q2/series.csv" download hx-boost="false">CSV</a>.
