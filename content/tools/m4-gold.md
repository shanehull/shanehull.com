---
title: "Divisia M4 to Gold"
description: "The broadest honest money aggregate, Divisia M4, measured per ounce of gold."
layout: "m4-gold"
tool_type: "chart"
math: true
---

The Divisia M4 to gold ratio measures the broadest honest money aggregate against the hardest money there is. M4 is the tinder gauge. It captures the institutional money that M2 misses, the large time deposits, wholesale money funds and eurodollars, and it is the correct continuation of M3, which the Federal Reserve stopped publishing in 2006.

The Center for Financial Stability publishes Divisia M4 as a quantity index at 1967 = 100, not a dollar level. Dividing it by the gold price as an index cancels the unknown base level, so both legs sit on the same 1967 scale:

$$
\text{M4 per ounce}(t)=\frac{\text{Divisia M4 index}(t)\times \text{gold price (1967)}}{\text{gold price}(t)}
$$

The line starts at 100 in 1967. It rises when money creation outruns the metal. It falls when gold revalues faster than money is created, as in 1971, the 1980 mania, 2011, and the 2024 to 2026 rally.

**Data**

- Numerator: Divisia M4 index, Center for Financial Stability, monthly, 1967 to present. Fetched live from the CFS workbook and cached for 24 hours.
- Denominator: gold price in USD per ounce, World Bank Pink Sheet 1960 to 2024, extended live by the COMEX settlement price (GC=F).
- Ratio: monthly, from 1967.