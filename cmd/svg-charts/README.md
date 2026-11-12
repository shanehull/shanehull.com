# svg-charts

Renders the static SVG figures used in blog posts from a JSON spec. It handles two chart types:

- `series`: one or more time-series panels, one line per panel, optional quartile lines.
- `lollipop`: one row per destination, with a stem from zero and a labelled dot.

It is not a general chart engine. A new chart layout needs a new type and renderer.

## Usage

```bash
go run ./cmd/svg-charts content/blog/data/<post>/figures.json
```

Data paths in the spec are relative to the spec file. Output paths are relative to the repository root. The command writes each figure and prints its path.

## Series data

A CSV with a `date` column in `YYYY-MM-DD` and one column for each panel `field`. Blank values are skipped, so series with different start dates can share a file.

## Lollipop data

A CSV with `measure`, `destination`, `value` and `move_required_pct`. Rows are sorted by `move_required_pct`, descending. A move can be a fall or a rise; the label says which.

## Spec

| Field       | Applies to | Notes                                                              |
| ----------- | ---------- | ------------------------------------------------------------------ |
| `type`      | all        | `series` or `lollipop`.                                            |
| `output`    | all        | Output SVG path, relative to the repository root.                  |
| `title`     | all        | Figure title.                                                      |
| `subtitle`  | all        | One line under the title.                                          |
| `source`    | all        | Footer text.                                                       |
| `data`      | all        | CSV path, relative to the spec file.                               |
| `start`     | `series`   | Optional ISO date. Rows before it are not plotted; the data file can keep the full history. |
| `quartiles` | `series`   | Draw 25th and 75th percentile lines when `true`.                   |
| `panels`    | `series`   | Array of `{title, field, format}`. `format` is a Go format string. |
| `labels`    | `lollipop` | Map keyed by `measure\|destination`, or `measure`.                 |

Each label takes `{label, detail, accent}`. `detail` is a Go format string applied to the value. `accent` colours that dot and stem with the site accent.

## Example

See `content/blog/data/housing-valuation-2026-q2/figures.json` for a three-panel series figure and a lollipop figure.

## Layout

The width, margins, font sizes and colours are fixed to keep figures consistent across posts. The height grows with the number of panels or rows. The SVG uses `prefers-color-scheme` for light and dark.
