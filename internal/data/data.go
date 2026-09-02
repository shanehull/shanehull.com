// Package data embeds source-controlled data assets used by chart tools.
package data

import _ "embed"

// GoldPriceCSV is the monthly gold price in USD per troy ounce, one row per
// month as `month,price`. Source: World Bank Pink Sheet (London fix), 1960
// through 2024. Months after the last embedded row are appended at runtime by
// callers from a live source scaled to this series.
//
//go:embed gold_price.csv
var GoldPriceCSV []byte

