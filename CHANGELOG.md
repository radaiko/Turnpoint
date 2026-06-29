# Changelog

All notable changes to Turnpoint are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com); the project aims to follow
semantic versioning. The in-app **What's New** dialog renders this file.

## [0.0.1] — 2026-06-29

First internal test build — the initial implementation, bundled for early
testers. Everything below ships in this build.

### Added
- **Athletes** — create, edit, delete and search athletes; per-athlete test
  history.
- **Step-test entry** — spreadsheet-style grid with clipboard paste and
  predefined Running and Cycling templates.
- **Analysis** — lactate curve fitting and 16 threshold methods (OBLA 2/4/6,
  Bsln+, Log-log, Dmax, ModDmax, Exp-Dmax, LTP1/2, IAT, LTratio, D2Lmax, MAX),
  validated against the WinLactat reference report (e.g. OBLA 4.0 → 16.1 km/h /
  167 bpm).
- **Training zones** — the 5-zone model (REKOM–SB) anchored on the aerobic
  (IAS) and anaerobic (IANS) thresholds.
- **Charts** — an interactive fitting chart with draggable IAS/IANS markers and
  a temporal results chart (lactate + heart rate over time with intensity step
  bars).
- **Reporting** — a printable report (system print / save-as-PDF) and CSV
  export of the step data.
- **Data** — local-first single-file SQLite storage with database backup; all
  data stays on your machine, no network required.
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
