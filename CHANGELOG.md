# Changelog

All notable changes to Turnpoint are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com); the project aims to follow
semantic versioning. The in-app **What's New** dialog renders this file.

## [0.0.2] — Unreleased

Feedback from the first test build. _Work in progress._

### Added
- **Automatic updates** — on launch Turnpoint checks GitHub for a newer release
  and shows an update banner with a one-click **Update now** (downloads and runs
  the installer on Windows). This is the only time the app goes online, and it
  can be turned off in Settings.
- **About & updates in Settings** — the version, release notes (**What's New**)
  and a manual **Check for updates** now sit at the top of Settings; the version
  in the bottom-left corner also opens the changelog.
- **Lactate at the thresholds** — the IAS/IANS anchor cards now show the lactate
  value (mmol/L), and it updates live while you drag a marker (e.g. LTP1/LTP2).
- **Training zones in the report** — the 5-zone table is now part of the report.
- Reports save with a filename built from the **athlete name and test date**
  (e.g. `Bogner-Markus_2026-06-30`).

### Changed
- **Editable templates** — every template, including the predefined Running and
  Cycling protocols, can now be edited — not only cloned.
- **Bolder zone colours** — the training-zone colours are stronger and easier to
  read at a glance.
- **Report in colour** — the report's charts and tables now use the same colours
  as the on-screen analysis instead of greyscale.
- **One sheet for raw data + curve** — the raw step-data table and the lactate
  performance curve are shown together on a single page (reference-report page 2).

## [0.0.1] — 2026-06-29

First internal test build — the initial implementation, bundled for early
testers. Everything below ships in this build.

### Added
- **Athletes** — create, edit, delete and search athletes; per-athlete test
  history.
- **Step-test entry** — spreadsheet-style grid with clipboard paste; predefined
  Running and Cycling templates, plus custom templates you can create, clone,
  edit and delete.
- **Analysis** — lactate curve fitting and 16 threshold methods (OBLA 2/4/6,
  Bsln+, Log-log, Dmax, ModDmax, Exp-Dmax, LTP1/2, IAT, LTratio, D2Lmax, MAX),
  validated against the WinLactat reference report (e.g. OBLA 4.0 → 16.1 km/h /
  167 bpm).
- **Configurable analysis** — enable/disable individual methods, configure their
  parameters (OBLA concentrations, baseline deltas), and choose the displayed
  curve fit (polynomial / exponential / spline).
- **Training zones** — the 5-zone model (REKOM–SB) anchored on the aerobic
  (IAS) and anaerobic (IANS) thresholds, with a selectable training profile and
  per-anchor controls (pick the source method, type an exact intensity, or drag
  the marker on the chart; reset to automatic anytime).
- **Charts** — an interactive fitting chart with draggable IAS/IANS markers and
  toggleable layers (curve, heart rate, points, zones), plus a temporal results
  chart (lactate + heart rate over time with intensity step bars).
- **Reporting** — a configurable report: include/omit/reorder content blocks,
  add a header, footer, logo and free-text evaluation; print or save as PDF, and
  export the chart as PNG/SVG and the data as CSV.
- **Data** — local-first single-file SQLite storage with one-click database
  backup and restore; all data stays on your machine, no network required.
- **Look & feel** — light and dark themes; a frameless window with
  platform-appropriate controls (macOS traffic lights, Windows/Linux buttons).

### Known limitations
- Threshold values are validated against the WinLactat reference report;
  parity with the `lactater` R package is still pending.
- Only the "Laufen Leistungssportler (6×/Woche)" zone profile is calibrated;
  the other profiles are provisional.
- Energy expenditure (kcal/h) uses a provisional formula and requires a body
  mass to be entered.
- PDF export goes through the system print dialog; there is no headless/batch
  PDF yet.
