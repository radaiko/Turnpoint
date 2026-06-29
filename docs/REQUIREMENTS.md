# Turnpoint — Software Requirements

|||
|-|-|
|**Project**|Turnpoint|
|**Type**|Lactate threshold analysis (desktop application)|
|**Status**|Draft|
|**Version**|0.1|
|**License**|Source-available / proprietary (all rights reserved)|
|**Author**|Aiko|

\---

## 1\. Overview

Turnpoint is a desktop application for analysing blood-lactate data from incremental step tests and deriving training thresholds and zones. It is a hobbyist-focused, local-first alternative to clinical lactate software (e.g. WinLactat), without the clinical/lab integrations or the price tag.

The user enters lactate, intensity, and heart-rate values from a step test; Turnpoint fits a lactate curve, computes thresholds using a suite of established scientific methods, derives training zones, and produces a clean report. All data stays on the user's machine.

The name refers to the **lactate turn point** (LTP), one of the core threshold concepts the application computes.

\---

## 2\. Goals \& Non-Goals

**Goals**

* Accurate, literature-validated threshold detection (parity with the `lactater` R package).
* A modern, responsive desktop UI that makes a normally lab-bound workflow approachable.
* Fully offline and local: no account, no cloud, no telemetry.
* Portable, exportable data the user fully owns.
* Professional-looking PDF reports.

**Non-Goals**

* Not a clinical/medical device. No diagnostic claims.
* No patient-management or lab-system integration (HL7, GDT).
* No automatic device/hardware import in the initial scope.
* No mobile app in the initial scope.

\---

## 3\. Target Users

* **The self-coached endurance athlete** — runs their own lactate tests (finger-prick analyser at the track or on the trainer) and wants objective thresholds and zones instead of guesswork.
* **The hobbyist/grassroots coach** — tests a handful of athletes and wants repeatable, presentable results without buying clinical software.
* **The quantified-self enthusiast** — tracks their own physiology over a season and wants longitudinal insight.

Assumed context: technically comfortable, desktop-first, values data ownership. Not a clinician.

\---

## 4\. Scope

### 4.1 In scope (v1)

* Athlete records and test/session management.
* Manual step-test data entry.
* Lactate curve fitting (multiple strategies).
* Threshold detection across the methods in §6.
* Training-zone derivation from thresholds.
* Interactive chart with manual threshold correction.
* Longitudinal and cross-sectional comparison.
* PDF report generation.
* Local SQLite storage; CSV import/export.

### 4.2 Out of scope (v1) — candidate future work

|Deferred|Notes|
|-|-|
|Automatic device import (spiroergometry, breath-by-breath, analyser exports)|Long tail of per-device formats; needs sample files. Manual entry + CSV covers v1.|
|Mobile companion app|Workflow is desktop-shaped. A read-only/companion app may follow; keep the data layer portable to allow it.|
|Clinical interfaces (HL7, GDT)|Only relevant to medical practices, not the target user.|
|Multi-user / cloud sync / accounts|Conflicts with local-first goal.|

\---

## 5\. Functional Requirements

### 5.1 Athlete Management

* **FR-A1** — Create, edit, and delete athlete records.
* **FR-A2** — Athlete fields: name, date of birth (or age), sex, body mass, primary sport, free-text notes.
* **FR-A3** — List and search athletes; open an athlete to see their test history.

### 5.2 Test / Session Entry

* **FR-T1** — Create a test bound to an athlete, with a protocol definition: sport, step duration, intensity increment, starting intensity, and continuous vs. intermittent mode (rest duration if intermittent).
* **FR-T2** — Record per-step values: intensity, end-of-step heart rate, blood lactate (mmol/L), and optional RPE.
* **FR-T3** — Support a resting/baseline lactate value and an optional pre-test note.
* **FR-T4** — Intensity unit follows the sport: running = speed or pace, cycling = power (W, optionally W/kg), rowing/ski-erg = power or pace, swimming = pace. Pace handling must respect the inverted relationship (lower pace = faster).
* **FR-T5** — Spreadsheet-style entry grid: keyboard navigation, add/remove/reorder rows, paste from clipboard.
* **FR-T6** — Validate input: minimum step count for analysis, flag implausible values, handle a lactate dip at low intensity and an aborted final step gracefully.

### 5.3 Curve Fitting

* **FR-F1** — Fit a lactate-vs-intensity curve from the step data.
* **FR-F2** — Offer multiple fit strategies, user-selectable: 3rd-order polynomial, exponential, and a smoothing/penalised spline.
* **FR-F3** — Guard against non-physiological fits (e.g. non-monotonic wiggle producing spurious local minima); surface a warning when the fit is poorly conditioned.
* **FR-F4** — Recompute the fit and all dependent thresholds reactively when data changes.

### 5.4 Threshold Detection

* **FR-D1** — Compute the threshold methods listed in §6.
* **FR-D2** — Let the user enable/disable individual methods and configure parameters (e.g. fixed OBLA values, baseline delta).
* **FR-D3** — For each computed threshold, report the corresponding intensity and the interpolated heart rate.
* **FR-D4** — Clearly label which fit each method depends on, and never silently mix incompatible fit/method pairs.

### 5.5 Training Zones

* **FR-Z1** — Derive training zones anchored on the aerobic threshold (LT1) and anaerobic threshold (LT2) the user selects.
* **FR-Z2** — Provide selectable zone models (§7), expressed in both intensity (pace/power) and heart-rate ranges.
* **FR-Z3** — Allow the anchor thresholds to be chosen from any computed method or set manually.

### 5.6 Interactive Correction

* **FR-C1** — Interactive lactate chart: curve, raw points, heart-rate overlay, threshold markers, and zone bands.
* **FR-C2** — Threshold markers are draggable; dragging recomputes the corresponding intensity, heart rate, and downstream zones live.
* **FR-C3** — A manually corrected threshold is visually distinguished from an algorithmic one and is persisted as an override.

### 5.7 Comparison \& History

* **FR-H1** — Longitudinal view: plot a single athlete's threshold/zone progression across tests over time.
* **FR-H2** — Cross-sectional view: compare multiple tests/athletes side by side.
* **FR-H3** — Overlay multiple lactate curves on one chart for visual comparison.

### 5.8 Reporting

* **FR-R1** — Generate a PDF report for a test.
* **FR-R2** — Report includes: athlete and test metadata, the lactate chart (as vector graphics, print quality), a results table of thresholds per method, and the derived zones.
* **FR-R3** — Template with merge fields plus an editable free-text commentary section.
* **FR-R4** — Optional custom header/logo. Multi-page, paginated layout.
* **FR-R5** — Export the chart as an image (PNG/SVG) and results as CSV.

### 5.9 Data Management

* **FR-M1** — Persist all data in a single local SQLite database.
* **FR-M2** — Import step data from CSV; export athletes/tests to CSV.
* **FR-M3** — Back up / restore the database as a single portable file.
* **FR-M4** — No data leaves the device; no network calls required for any core feature.

\---

## 6\. Threshold Methods

Methods to implement, validated against `lactater`. References are the canonical originators; confirm exact DOIs against `lactater`'s documented sources during implementation.

|ID|Method|Definition (summary)|Requires|Reference|
|-|-|-|-|-|
|OBLA 2.0 / 3.0 / 4.0|Onset of Blood Lactate Accumulation (fixed)|Intensity at a fixed lactate concentration on the fitted curve|Curve fit|Sjödin \& Jacobs (1981); Heck et al. (1985)|
|Bsln+0.5 / 1.0 / 1.5|Baseline plus|Intensity at baseline lactate + a fixed delta|Baseline + fit|Berg et al. (1990); Zoladz et al. (1995)|
|Log-log|Log-log (Beaver)|Breakpoint between two linear segments in log(lactate) vs log(intensity)|Log-log regression|Beaver et al. (1985)|
|Dmax|Dmax|Point of maximum perpendicular distance from the chord joining first and last fitted points|Polynomial fit|Cheng et al. (1992)|
|ModDmax|Modified Dmax|As Dmax, but chord starts at the first point rising > 0.4 mmol/L above the previous|Polynomial fit|Bishop et al. (1998)|
|Exp-Dmax|Exponential Dmax|Dmax computed on an exponential fit|Exponential fit|Newell et al. (2007)|
|LTP1 / LTP2|Lactate Turn Points|First and second turn points via three-phase / segmented regression|Segmented regression|Hofmann \& Tschakert; Pokan et al.|
|IAT (Dickhuth)|Individual Anaerobic Threshold|Lactate minimum equivalent + 1.5 mmol/L|Curve fit|Dickhuth et al. (1991); Stegmann|
|LTratio|Minimum Lactate-Intensity Ratio|Intensity at the minimum lactate/intensity ratio|Curve fit|Jamnick et al. (2018)|
|D2Lmax|Maximum acceleration|Intensity at the maximum of the second derivative of the fitted curve|Curve fit|Newell et al. (2006/2007)|

> \*\*Lactate-minimum test (LTmin):\*\* noted but lower priority — it requires a specific (non-standard) protocol with a post-loading ramp, so it is a distinct workflow rather than a method over a standard step test. Tegtbur et al. (1993).

\---

## 7\. Training Zone Models

Zones derive from two anchors: **LT1** (aerobic threshold) and **LT2** (anaerobic threshold).

|Model|Zones|Basis|
|-|-|-|
|3-zone|Below LT1 / LT1–LT2 / above LT2|Polarised model|
|5-zone|Recovery / endurance / tempo / threshold / VO₂|Anchored on LT1, LT2, and interpolated bands|

Each zone is expressed as an intensity range (pace or power) **and** a heart-rate range. Zone boundaries and the active model are configurable.

\---

## 8\. Non-Functional Requirements

* **NFR-1 (Platform)** — Cross-platform desktop: Windows, macOS, Linux.
* **NFR-2 (Offline-first)** — All core features function with no network connection. No telemetry.
* **NFR-3 (Performance)** — Live threshold/zone recompute during marker dragging feels instant (target < 100 ms).
* **NFR-4 (Data ownership)** — Single-file, portable database; open, documented schema; CSV import/export.
* **NFR-5 (Privacy)** — No data transmitted off-device by default. Personal athlete data treated as sensitive.
* **NFR-6 (Correctness)** — Threshold outputs reproducible and validated against reference values (§11).
* **NFR-7 (Usability)** — Approachable to a non-clinician; sensible defaults; the happy path (enter → fit → thresholds → zones → report) requires minimal configuration.
* **NFR-8 (Footprint)** — Small install and low resource use befitting a desktop utility.

\---

## 9\. Data Model (indicative)

* **Athlete** — id, name, dob/age, sex, body\_mass, sport, notes.
* **Test** — id, athlete\_id, date, sport, protocol (step\_duration, increment, start\_intensity, mode, rest), baseline\_lactate, notes.
* **Step** — id, test\_id, order, intensity, heart\_rate, lactate, rpe (optional).
* **ThresholdResult** — id, test\_id, method, intensity, heart\_rate, is\_override, fit\_type.
* **Zone** — id, test\_id, model, zone\_index, intensity\_low, intensity\_high, hr\_low, hr\_high.
* **ReportSettings / Template** — header/logo, commentary text.

\---

## 10\. Technical Architecture

**Stack**

* **Core logic:** Go — threshold models, curve fitting, zone derivation. Fast, testable, packaged as a standalone library/binary.
* **Frontend:** Svelte — reactive UI, entry grid, interactive chart.
* **Desktop shell:** Tauri v2 — native window, packaging, cross-platform builds.
* **Storage:** SQLite — local embedded database.

**Build order (bottom-up, risk-first)**

1. **Go threshold-model package** with a test suite validated against `lactater`. Independently useful; the credibility-critical core.
2. **Svelte + Tauri shell** — data entry, persistence, and the interactive chart on top of the core.
3. **Report export** — templating and PDF generation last.

**Effort (recalled estimate):** usable MVP (entry → fit → \~8 methods → chart → simple PDF) ≈ 3–4 weeks part-time; polished v1 ≈ 5–7 weeks. Weight concentrates in two places: the **curve-fitting + validation core** and the **report generator**. Everything else is routine.

\---

## 11\. Validation \& Testing

* **V1** — Each threshold method has unit tests asserting parity with `lactater` outputs within a defined tolerance.
* **V2** — Maintain a set of reference datasets (e.g. `lactater`'s demo data and published examples) with expected threshold values.
* **V3** — Each method documents its source paper/DOI in code and in this spec.
* **V4** — Fit-strategy edge cases covered: minimal step count, low-intensity lactate dip, aborted final step.

\---

## 12\. Licensing

Source-available, **all rights reserved**. The repository may be published for review/reference only; no license is granted to use, run, copy, modify, or distribute. License text lives in `LICENSE` at the repo root.

\---

## 13\. Roadmap

|Phase|Deliverable|
|-|-|
|P0|Go core: curve fitting + \~8 threshold methods, validated against `lactater`|
|P1|Tauri/Svelte shell: athlete + test CRUD, entry grid, SQLite persistence|
|P2|Interactive chart with draggable threshold correction; zone derivation|
|P3|PDF report generation|
|P4|Longitudinal / cross-sectional comparison; remaining methods|
|Later|Mobile companion; device import (re-evaluate)|

\---

## 14\. Open Questions

1. **Go ↔ Tauri integration** — Go core as a sidecar binary the shell spawns (local IPC/HTTP), via cgo/FFI, or a pure-Go binary with an embedded webview instead of Tauri? Decide before P1; it shapes packaging and the dev loop.
2. **Default fit strategy** — which fit is the sensible default, and should it vary by method?
3. **Pace representation** — internal unit (m/s) with pace as a display layer, confirmed across run/row/swim?
4. **Zone model defaults** — ship 3-zone, 5-zone, or both enabled by default?
5. **Report templating** — fixed layout with merge fields, or user-editable template blocks?

