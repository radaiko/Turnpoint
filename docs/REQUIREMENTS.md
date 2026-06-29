# Turnpoint — Software Requirements

| | |
|---|---|
| **Project** | Turnpoint |
| **Type** | Lactate threshold analysis (desktop application) |
| **Status** | Draft |
| **Version** | 0.3 |
| **License** | Source-available / proprietary (all rights reserved) |
| **Author** | Aiko |

---

## 1. Overview

Turnpoint is a desktop application for analysing blood-lactate data from incremental step tests and deriving training thresholds and zones. It is a hobbyist-focused, local-first alternative to clinical lactate software (e.g. WinLactat), without the clinical/lab integrations or the price tag.

The user enters lactate, intensity, and heart-rate values from a step test; Turnpoint fits a lactate curve, computes thresholds using a suite of established scientific methods, derives training zones, and produces a clean report. All data stays on the user's machine.

The name refers to the **lactate turn point** (LTP), one of the core threshold concepts the application computes.

---

## 2. Goals & Non-Goals

**Goals**

- Accurate, literature-validated threshold detection (parity with the `lactater` R package).
- A modern, responsive desktop UI that makes a normally lab-bound workflow approachable.
- Fully offline and local: no account, no cloud, no telemetry.
- Portable, exportable data the user fully owns.
- Professional-looking PDF reports.

**Non-Goals**

- Not a clinical/medical device. No diagnostic claims.
- No patient-management or lab-system integration (HL7, GDT).
- No automatic device/hardware import in the initial scope.
- No mobile app in the initial scope.

---

## 3. Target Users

- **The self-coached endurance athlete** — runs their own lactate tests (finger-prick analyser at the track or on the trainer) and wants objective thresholds and zones instead of guesswork.
- **The hobbyist/grassroots coach** — tests a handful of athletes and wants repeatable, presentable results without buying clinical software.
- **The quantified-self enthusiast** — tracks their own physiology over a season and wants longitudinal insight.

Assumed context: technically comfortable, desktop-first, values data ownership. Not a clinician.

---

## 4. Scope

### 4.1 In scope (v1)

- Athlete records and test/session management.
- Manual step-test data entry.
- Lactate curve fitting (multiple strategies).
- Threshold detection across the methods in §6.
- Training-zone derivation from thresholds.
- Interactive chart with manual threshold correction.
- Longitudinal and cross-sectional comparison.
- PDF report generation.
- Local SQLite storage; CSV import/export.

### 4.2 Out of scope (v1) — candidate future work

| Deferred | Notes |
|---|---|
| Automatic device import (spiroergometry, breath-by-breath, analyser exports) | Long tail of per-device formats; needs sample files. Manual entry + CSV covers v1. |
| Mobile companion app | Workflow is desktop-shaped. A read-only/companion app may follow; keep the data layer portable to allow it. |
| Clinical interfaces (HL7, GDT) | Only relevant to medical practices, not the target user. |
| Multi-user / cloud sync / accounts | Conflicts with local-first goal. |

---

## 5. Functional Requirements

### 5.1 Athlete Management
- **FR-A1** — Create, edit, and delete athlete records.
- **FR-A2** — Athlete fields: name, date of birth (or age), sex, body mass, primary sport, free-text notes.
- **FR-A3** — List and search athletes; open an athlete to see their test history.

### 5.2 Test / Session Entry
- **FR-T1** — Create a test bound to an athlete, with a protocol definition: sport, step duration, intensity increment, starting intensity, and continuous vs. intermittent mode (rest duration if intermittent).
- **FR-T2** — Record per-step values: intensity, time point, end-of-step heart rate, blood lactate (mmol/L), and optional Borg/RPE.
- **FR-T3** — Support a resting/baseline row (intensity 0) and an optional pre-test note.
- **FR-T4** — Intensity unit follows the sport: running = speed (km/h) or pace, cycling = power (W, optionally W/kg), rowing/ski-erg = power or pace, swimming = pace. Pace handling must respect the inverted relationship (lower pace = faster).
- **FR-T5** — Raw values are entered in a spreadsheet-style grid with the canonical column layout (see Appendix A): **intensity**, **time point** (mm:ss), **HR (bpm)**, **lactate (mmol/L)**, **Borg/RPE**. The intensity column header/unit switches with the active sport (e.g. `[km/h]` for running, `[W]` for cycling). The grid supports keyboard navigation, add/remove/reorder rows, and paste from clipboard.
- **FR-T6** — Validate input: minimum step count for analysis, flag implausible values, and handle a lactate dip at low intensity and an aborted final step gracefully (e.g. a final step shorter than the protocol step duration).
- **FR-T7 (Predefined templates)** — Ship predefined entry templates for **Running (Lauf)** and **Cycling (Rad)** that preset the sport, intensity unit, default step duration, default increment, starting intensity, and visible columns (see Appendix B).
- **FR-T8 (User templates)** — The user can create, save, edit, and delete their own templates (a named protocol + column configuration) and select one when starting a new test.

### 5.3 Curve Fitting
- **FR-F1** — Fit a lactate-vs-intensity curve from the step data.
- **FR-F2** — Offer multiple fit strategies, user-selectable: 3rd-order polynomial, exponential, and a smoothing/penalised spline.
- **FR-F3** — Guard against non-physiological fits (e.g. non-monotonic wiggle producing spurious local minima); surface a warning when the fit is poorly conditioned.
- **FR-F4** — Recompute the fit and all dependent thresholds reactively when data changes.

### 5.4 Threshold Detection
- **FR-D1** — Compute the threshold methods listed in §6.
- **FR-D2** — Let the user enable/disable individual methods and configure parameters (e.g. fixed OBLA values, baseline delta).
- **FR-D3** — For each computed threshold, report the corresponding intensity and the interpolated heart rate.
- **FR-D4** — Clearly label which fit each method depends on, and never silently mix incompatible fit/method pairs.
- **FR-D5 (Results table)** — Produce a threshold results table à la the reference report (Appendix C): for each marker — fixed thresholds (2 / 4 / 6 mmol/L), IAS, IANS, and MAX — list intensity, lactate, heart rate, % of max performance, pace (per 1000 m), and (optional, requires body mass) kcal/h.

### 5.5 Training Zones
- **FR-Z1** — Derive a **5-zone training model** (§7) anchored on two thresholds: the aerobic threshold **LT1 (IAS — individuelle aerobe Schwelle)** and the anaerobic threshold **LT2 (IANS — individuelle anaerobe Schwelle)**.
- **FR-Z2 (Auto LT1/LT2)** — LT1 and LT2 are computed automatically from the threshold methods (§6) via configurable default mappings (e.g. LT1 ← Log-log / LTP1 / Bsln+; LT2 ← ModDmax / LTP2 / OBLA 4.0). *(In the reference report, IANS was configured as OBLA 4.0 — see Appendix C.)*
- **FR-Z3 (Manual override)** — The user can override LT1 and LT2: by selecting a different computed method as the anchor, by editing the intensity directly, or by dragging the marker on the chart (§5.6). An overridden anchor is persisted and flagged as manual.
- **FR-Z4** — Zones are expressed as ranges for intensity (km/h / pace / W), heart rate, lactate, and pace (per 1000 m), and recompute live whenever LT1 or LT2 changes.
- **FR-Z5 (Training profiles)** — Zone boundaries derive from a selectable **training profile** (sport + athlete level + weekly frequency, e.g. *"Laufen Leistungssportler (6×/Woche)"*). Ship a small set of predefined profiles and allow user-defined ones; the profile determines how the zones spread around IAS/IANS (§7).
- **FR-Z6 (Derived metrics)** — For each threshold and zone bound, derive pace (min/km), % of max intensity, and interpolated HR. kcal/h is optional and requires body mass + a metabolic formula.

### 5.6 Visualisation & Interactive Correction
- **FR-C1 (Fitting view)** — Lactate-vs-intensity chart: fitted curve, raw points, heart-rate overlay, threshold markers, and zone bands. Used for curve fitting and threshold placement.
- **FR-C2** — Threshold markers are draggable; dragging recomputes the corresponding intensity, heart rate, pace, and downstream zones live.
- **FR-C3** — A manually corrected threshold is visually distinguished from an algorithmic one and is persisted as an override.
- **FR-C4 (Temporal view)** — Primary results chart, "Zeitliche Darstellung" (cf. Appendix C image): X-axis = time; dual Y-axes = lactate (left) and heart rate (right); intensity (km/h **or W**) shown as step bars; the five zone bands overlaid on the curves; **IAS** and **IANS** drawn as vertical lines; configurable reference lines (e.g. 4 mmol/L, threshold HR, threshold speed).

### 5.7 Comparison & History
- **FR-H1** — Longitudinal view: plot a single athlete's threshold/zone progression across tests over time.
- **FR-H2** — Cross-sectional view: compare multiple tests/athletes side by side.
- **FR-H3** — Overlay multiple lactate curves on one chart for visual comparison.

### 5.8 Reporting
- **FR-R1** — Generate a **printable** PDF report for a test (system print dialog + export to PDF).
- **FR-R2** — Report content blocks: cover + athlete/anamnesis, test remarks, the temporal chart (vector, print quality), the raw step-data table, the threshold results table (FR-D5), the training zones (§7), and an evaluation (Bewertung) section.
- **FR-R3 (Free configuration)** — The report is **freely configurable**: the user can choose which blocks appear, reorder them, and edit free-text sections (commentary, evaluation). The default layout mirrors the reference report's four-part structure (Appendix C).
- **FR-R4** — Optional custom header/logo and footer. Multi-page, paginated layout.
- **FR-R5** — Export the chart as an image (PNG/SVG) and results/zones as CSV.

### 5.9 Data Management
- **FR-M1** — Persist all data in a single local SQLite database.
- **FR-M2** — Import step data from CSV; export athletes/tests to CSV.
- **FR-M3** — Back up / restore the database as a single portable file.
- **FR-M4** — No data leaves the device; no network calls required for any core feature.

---

## 6. Threshold Methods

Methods to implement, validated against `lactater`. References are the canonical originators; confirm exact DOIs against `lactater`'s documented sources during implementation.

| ID | Method | Definition (summary) | Requires | Reference |
|---|---|---|---|---|
| OBLA 2.0 / 3.0 / 4.0 | Onset of Blood Lactate Accumulation (fixed) | Intensity at a fixed lactate concentration on the fitted curve | Curve fit | Sjödin & Jacobs (1981); Heck et al. (1985) |
| Bsln+0.5 / 1.0 / 1.5 | Baseline plus | Intensity at baseline lactate + a fixed delta | Baseline + fit | Berg et al. (1990); Zoladz et al. (1995) |
| Log-log | Log-log (Beaver) | Breakpoint between two linear segments in log(lactate) vs log(intensity) | Log-log regression | Beaver et al. (1985) |
| Dmax | Dmax | Point of maximum perpendicular distance from the chord joining first and last fitted points | Polynomial fit | Cheng et al. (1992) |
| ModDmax | Modified Dmax | As Dmax, but chord starts at the first point rising > 0.4 mmol/L above the previous | Polynomial fit | Bishop et al. (1998) |
| Exp-Dmax | Exponential Dmax | Dmax computed on an exponential fit | Exponential fit | Newell et al. (2007) |
| LTP1 / LTP2 | Lactate Turn Points | First and second turn points via three-phase / segmented regression | Segmented regression | Hofmann & Tschakert; Pokan et al. |
| IAT (Dickhuth) | Individual Anaerobic Threshold | Lactate minimum equivalent + 1.5 mmol/L | Curve fit | Dickhuth et al. (1991); Stegmann |
| LTratio | Minimum Lactate-Intensity Ratio | Intensity at the minimum lactate/intensity ratio | Curve fit | Jamnick et al. (2018) |
| D2Lmax | Maximum acceleration | Intensity at the maximum of the second derivative of the fitted curve | Curve fit | Newell et al. (2006/2007) |

> **Lactate-minimum test (LTmin):** noted but lower priority — it requires a specific (non-standard) protocol with a post-loading ramp, so it is a distinct workflow rather than a method over a standard step test. Tegtbur et al. (1993).

---

## 7. Training Zone Model (5-Zone)

The v1 zone model is the classic German/Austrian **5-zone endurance model**, anchored on two automatically-computed, manually-overridable thresholds — **IAS** (individuelle aerobe Schwelle ≈ LT1) and **IANS** (individuelle anaerobe Schwelle ≈ LT2). Zone boundaries are produced by the selected **training profile** (FR-Z5), which defines how the zones spread around IAS/IANS; all boundaries remain configurable.

| Zone | German | English | Position |
|---|---|---|---|
| Z1 | REKOM | Recovery / compensation | below GA1 (regeneration) |
| Z2 | GA1 | Basic endurance 1 | around / below IAS |
| Z3 | GA2 | Basic endurance 2 | between IAS and IANS (lower) |
| Z4 | EB | Development | around IANS |
| Z5 | SB | Peak | above IANS |

Each zone is output as ranges for intensity (km/h / pace / W), heart rate, lactate, and pace (per 1000 m), interpolated on the test data. Zones recompute live whenever IAS or IANS is adjusted.

**Reference zones (profile "Laufen Leistungssportler, 6×/Woche"), from Appendix C:**

| Zone | km/h | Lactate | HR | Pace /km |
|---|---|---|---|---|
| GA1 | 7.4–11.3 | 1.2–1.5 | 107–126 | 08:06–05:19 |
| GA2 | 11.3–14.2 | 1.5–2.5 | 126–150 | 05:19–04:14 |
| EB | 14.2–16.4 | 2.5–4.4 | 150–169 | 04:14–03:39 |
| SB | 16.4–20.1 | 4.4–7.7 | 169–185 | 03:39–02:58 |

> The reference report displays the upper four zones (GA1–SB); **REKOM** is the recovery zone below GA1 and completes the 5-zone model. Zone naming is configurable (German labels by default, English optional).

---

## 8. Non-Functional Requirements

- **NFR-1 (Platform)** — Cross-platform desktop: Windows, macOS, Linux.
- **NFR-2 (Offline-first)** — All core features function with no network connection. No telemetry.
- **NFR-3 (Performance)** — Live threshold/zone recompute during marker dragging feels instant (target < 100 ms).
- **NFR-4 (Data ownership)** — Single-file, portable database; open, documented schema; CSV import/export.
- **NFR-5 (Privacy)** — No data transmitted off-device by default. Personal athlete data treated as sensitive.
- **NFR-6 (Correctness)** — Threshold outputs reproducible and validated against reference values (§11).
- **NFR-7 (Usability)** — Approachable to a non-clinician; sensible defaults; the happy path (enter → fit → thresholds → zones → report) requires minimal configuration.
- **NFR-8 (Footprint)** — Small install and low resource use befitting a desktop utility.

---

## 9. Data Model (indicative)

- **Athlete** — id, name, dob/age, sex, body_mass, sport, notes.
- **Test** — id, athlete_id, date, sport, protocol (step_duration, increment, start_intensity, mode, rest), baseline_lactate, notes.
- **Step** — id, test_id, order, intensity, heart_rate, lactate, rpe (optional).
- **ThresholdResult** — id, test_id, method, intensity, heart_rate, is_override, fit_type.
- **Zone** — id, test_id, model, zone_index, intensity_low, intensity_high, hr_low, hr_high.
- **ReportSettings / Template** — header/logo, commentary text.

---

## 10. Technical Architecture

**Stack**

- **Core logic:** Go — threshold models, curve fitting, zone derivation. Fast, testable, packaged as a standalone library/binary.
- **Frontend:** Svelte — reactive UI, entry grid, interactive chart.
- **Desktop shell:** Tauri v2 — native window, packaging, cross-platform builds.
- **Storage:** SQLite — local embedded database.

**Build order (bottom-up, risk-first)**

1. **Go threshold-model package** with a test suite validated against `lactater`. Independently useful; the credibility-critical core.
2. **Svelte + Tauri shell** — data entry, persistence, and the interactive chart on top of the core.
3. **Report export** — templating and PDF generation last.

**Effort (recalled estimate):** usable MVP (entry → fit → ~8 methods → chart → simple PDF) ≈ 3–4 weeks part-time; polished v1 ≈ 5–7 weeks. Weight concentrates in two places: the **curve-fitting + validation core** and the **report generator**. Everything else is routine.

---

## 11. Validation & Testing

- **V1** — Each threshold method has unit tests asserting parity with `lactater` outputs within a defined tolerance.
- **V2** — Maintain a set of reference datasets (e.g. `lactater`'s demo data and published examples) with expected threshold values.
- **V3** — Each method documents its source paper/DOI in code and in this spec.
- **V4** — Fit-strategy edge cases covered: minimal step count, low-intensity lactate dip, aborted final step.

---

## 12. Licensing

Source-available, **all rights reserved**. The repository may be published for review/reference only; no license is granted to use, run, copy, modify, or distribute. License text lives in `LICENSE` at the repo root.

---

## 13. Roadmap

| Phase | Deliverable |
|---|---|
| P0 | Go core: curve fitting + ~8 threshold methods, validated against `lactater` |
| P1 | Tauri/Svelte shell: athlete + test CRUD, entry grid, SQLite persistence |
| P2 | Interactive chart with draggable threshold correction; zone derivation |
| P3 | PDF report generation |
| P4 | Longitudinal / cross-sectional comparison; remaining methods |
| Later | Mobile companion; device import (re-evaluate) |

---

## 14. Open Questions

1. **Go ↔ Tauri integration** — Go core as a sidecar binary the shell spawns (local IPC/HTTP), via cgo/FFI, or a pure-Go binary with an embedded webview instead of Tauri? Decide before P1; it shapes packaging and the dev loop.
2. **Default fit strategy** — which fit is the sensible default, and should it vary by method?
3. **Pace representation** — internal unit (m/s) with pace as a display layer, confirmed across run/row/swim?
4. ~~Zone model defaults~~ — **Resolved:** v1 ships the 5-zone model (§7); 3-zone derivable later.
5. **Report templating** — fixed layout with merge fields, or user-editable template blocks?

---

## Appendix A — Reference Dataset (Running)

A real running step test, used as the canonical input-format example and as a **P0 validation fixture**. It exercises two edge cases: a low-intensity lactate dip (8 km/h, 1.19 < 1.24) and an aborted final step (the 20 km/h step ends at 22:10, ~1:10 in).

| [km/h] | Time | HR | Lactate [mmol/L] | Borg/RPE |
|---|---|---|---|---|
| 0.0 | 00:00 | 0 | 0.00 | 6 |
| 6.0 | 03:00 | 98 | 1.24 | 6 |
| 8.0 | 06:00 | 111 | 1.19 | 6 |
| 10.0 | 09:00 | 120 | 1.32 | 6 |
| 12.0 | 12:00 | 131 | 1.66 | 6 |
| 14.0 | 15:00 | 149 | 2.38 | 6 |
| 16.0 | 18:00 | 166 | 3.89 | 6 |
| 18.0 | 21:00 | 180 | 6.66 | 6 |
| 20.0 | 22:10 | 185 | 7.74 | 6 |

Protocol: 3:00 steps, +2 km/h increment, continuous. The 0 km/h row is the resting/baseline row; the Borg/RPE column is constant here and may be sparse in real use.

## Appendix B — Predefined Templates

| Template | Sport | Intensity unit | Default step | Default increment | Start | Columns |
|---|---|---|---|---|---|---|
| Running (Lauf) | Running | km/h (pace optional) | 3:00 | +2 km/h | 6 km/h | intensity, time, HR, lactate, Borg/RPE |
| Cycling (Rad) | Cycling | W (W/kg optional) | 3:00 | +20–40 W (configurable) | configurable | intensity, time, HR, lactate, Borg/RPE |

Templates preset the protocol and the visible grid columns. Users can clone a predefined template and save their own variants (FR-T8).

## Appendix C — Reference Report (mesics "Lactate EXPRESS")

A full WinLactat / Lactate EXPRESS report for the Appendix A dataset (athlete "Bogner Markus", 01.02.2025). Used as the **output reference** (chart style, tables, report structure) and as **parity targets** for P0.

**Threshold markers (Schwellen):**

| Marker | km/h | Lactate | HR | % max | Pace /km |
|---|---|---|---|---|---|
| 2 mmol/L | 13.1 | 2.0 | 140 | 65.5 | 04:34 |
| 4 mmol/L | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| 6 mmol/L | 17.6 | 6.0 | 177 | 87.8 | 03:25 |
| IAS | 10.5 | 1.4 | 122 | 52.7 | 05:41 |
| IANS | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| MAX | 20.0 | 7.7 | 185 | 100.0 | 03:00 |

**Derived-metric formulas (confirmed against the report):**

- **% of max** = intensity ÷ max intensity × 100 (e.g. 16.1 ÷ 20.0 = 80.5 %).
- **Pace per 1000 m** = 60 ÷ (km/h), formatted mm:ss (e.g. 60 ÷ 16.1 = 3.727 → 03:43).
- **HR at threshold** = interpolated on HR-vs-intensity (≈ 167 at 16.1 km/h).
- **kcal/h** = 0 in this report because body mass was 0; requires body mass + a metabolic model.

**Notes for parity:**

- IANS here equals the 4 mmol/L fixed threshold exactly (same km/h, lactate, HR) — i.e. IANS was configured as OBLA 4.0. P0 check: with LT2 ← OBLA 4.0, Turnpoint must reproduce **16.1 km/h / 167 BPM**.
- The fixed thresholds are reproducible from the Appendix A raw data by curve fit + interpolation (e.g. 4 mmol/L falls between the 16 km/h and 18 km/h steps → 16.1 km/h).
- Default report structure: p1 cover + anamnesis + test remarks; p2 raw-data table + temporal chart; p3 threshold table + training zones; p4 evaluation (Bewertung).
- "Nicht ausbelastet" (not maximally exhausted) and "EndLac 8.72 mmol" are free-text test remarks — supports FR-T3 / FR-R3.
