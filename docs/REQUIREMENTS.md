# Turnpoint — Software Requirements Specification

| | |
|---|---|
| **Project** | Turnpoint |
| **Type** | Lactate threshold analysis (desktop application) |
| **Status** | Draft |
| **Version** | 0.7 |
| **License** | Source-available / proprietary (all rights reserved) |
| **Author** | Aiko |

---

## How to read this document

This is a Software Requirements Specification (SRS). It states **what** the
system must do, not how to build it (architecture lives in §10 as supporting
context). Each requirement is written to be atomic, uniquely identified, and
testable.

**Conventions**

- **"shall"** marks a normative, mandatory requirement. **"should"** marks a
  recommendation. **"may"** marks an optional capability.
- Requirement IDs are stable: `FR-<area><n>` for functional, `NFR-<n>` for
  non-functional. Once assigned, an ID is not reused for a different
  requirement.
- Each requirement is followed by **Acceptance criteria** — the observable
  conditions a tester or automated test verifies. A requirement is "done" only
  when every acceptance criterion passes.
- **⚠ OI-<n>** marks a point that had an open decision. Each now carries a
  **proposed** resolution in §15 (Proposed Resolutions), pending author
  confirmation. Inline acceptance criteria state the proposed value; §15 is the
  authoritative list with rationale and a confidence rating. Until confirmed,
  treat each proposal as a working default, not a final decision.
- See §16 (Glossary) for domain terms (IAS, IANS, OBLA, LT1/LT2, etc.).

---

## 1. Overview

Turnpoint is a desktop application for analysing blood-lactate data from
incremental step tests and deriving training thresholds and zones. It is a
hobbyist-focused, local-first alternative to clinical lactate software (e.g.
WinLactat), without the clinical/lab integrations or the price tag.

The user enters lactate, intensity, and heart-rate values from a step test;
Turnpoint fits a lactate curve, computes thresholds using a suite of established
scientific methods, derives training zones, and produces a clean report. All
data stays on the user's machine.

The name refers to the **lactate turn point** (LTP), one of the core threshold
concepts the application computes.

---

## 2. Goals & Non-Goals

**Goals**

- Accurate, literature-validated threshold detection (parity with the
  `lactater` R package, within a defined tolerance — see ⚠ OI-1).
- A modern, responsive desktop UI that makes a normally lab-bound workflow
  approachable.
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

- **The self-coached endurance athlete** — runs their own lactate tests
  (finger-prick analyser at the track or on the trainer) and wants objective
  thresholds and zones instead of guesswork.
- **The hobbyist/grassroots coach** — tests a handful of athletes and wants
  repeatable, presentable results without buying clinical software.
- **The quantified-self enthusiast** — tracks their own physiology over a
  season and wants longitudinal insight.

Assumed context: technically comfortable, desktop-first, values data ownership.
Not a clinician.

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

#### FR-A1 — Create athlete
The system **shall** allow the user to create a new athlete record.

*Acceptance criteria*
- A created athlete persists to the database and appears in the athlete list
  (FR-A4) without an application restart.
- Creation is rejected with a clear message if a required field is missing.
  Required field: **name** only; all other fields optional (proposed — ⚠ OI-2).

#### FR-A2 — Edit athlete
The system **shall** allow the user to edit an existing athlete record.

*Acceptance criteria*
- Edits persist and are reflected in dependent views (test history, reports) on
  next render.
- Editing an athlete does **not** retroactively alter stored `ThresholdResult`
  or `Zone` rows of existing tests (those are snapshots). ⚠ OI-3 — confirm
  whether body-mass-dependent derived metrics (kcal/h) recompute on athlete edit
  or remain snapshot.

#### FR-A3 — Delete athlete
The system **shall** allow the user to delete an athlete record.

*Acceptance criteria*
- Deletion requires explicit user confirmation.
- Deleting an athlete cascades to (or blocks on) that athlete's tests, steps,
  threshold results, and zones. ⚠ OI-4 — choose cascade-delete vs.
  block-if-tests-exist vs. soft-delete/archive.

#### FR-A4 — List and search athletes
The system **shall** display a list of athletes and allow the user to search/
filter it.

*Acceptance criteria*
- The list shows at minimum each athlete's name and (⚠ OI-5) other summary
  columns TBD.
- Search matches on name (case-insensitive, substring). ⚠ OI-5 — confirm
  additional searchable fields.

#### FR-A5 — Athlete fields
An athlete record **shall** store: name, date of birth **or** age, sex, body
mass, primary sport, and free-text notes.

*Acceptance criteria*
- Date of birth and age are mutually derivable; storing one is sufficient. ⚠
  OI-6 — store DOB (preferred, age auto-derives) or age-at-test? Decide which is
  canonical.
- Sex accepts a defined enumeration. ⚠ OI-7 — define allowed sex values and
  whether the field is used in any calculation (e.g. metabolic model).
- Body mass is stored in kilograms, 1 decimal place, range 20.0–250.0
  (proposed — ⚠ OI-8); an empty body mass is permitted and disables kcal/h
  (FR-Z6, FR-D5).

### 5.2 Test / Session Entry

#### FR-T1 — Create test with protocol
The system **shall** allow the user to create a test bound to an athlete, with a
protocol definition: sport, step duration, intensity increment, starting
intensity, and continuous vs. intermittent mode (including rest duration when
intermittent).

*Acceptance criteria*
- A test cannot be created without an associated athlete.
- All protocol fields persist with the test and are shown in the report
  (FR-R2).
- When mode = intermittent, rest duration is required; when continuous, it is
  absent/ignored.

#### FR-T2 — Record per-step values
The system **shall** record, for each test step: intensity, time point,
end-of-step heart rate, blood lactate (mmol/L), and optional Borg/RPE.

*Acceptance criteria*
- Intensity, time point, heart rate, and lactate are accepted per the ranges in
  FR-T6.
- Borg/RPE is optional and may be left empty on any step.
- A step with an empty **lactate** cell is excluded from curve fitting (FR-F1)
  and flagged in the UI.

#### FR-T3 — Baseline row and test remarks
The system **shall** support a resting/baseline step (intensity 0) and free-text
test remarks (pre-test note and post-test remarks).

*Acceptance criteria*
- Exactly one baseline row (intensity 0) is supported per test; it supplies the
  baseline lactate used by Bsln+ methods (§6) and IAT.
- Free-text remarks persist and are available to the report's remarks block
  (FR-R2), e.g. "Nicht ausbelastet", "EndLac 8.72 mmol".

#### FR-T4 — Supported sports and units (v1)
The system **shall** support **Running** (intensity in km/h) and **Cycling**
(intensity in W) as active sports in v1.

*Acceptance criteria*
- Pace (min/km) is a **derived** display/output metric (FR-Z6, Appendix C), not
  a primary input unit.
- Rowing, ski-erg, and swimming are **not** selectable in v1.
- The data model and unit handling do not assume a monotone-increasing
  pace↔speed relationship in a way that would block future inverted-unit sports
  (lower pace = faster). ⚠ OI-9 — confirm this is a design constraint to test,
  or defer entirely.

#### FR-T5 — Spreadsheet entry grid
The system **shall** present raw values in a spreadsheet-style grid with the
canonical column layout (Appendix A): intensity, time point (mm:ss), HR (bpm),
lactate (mmol/L), Borg/RPE.

*Acceptance criteria*
- The intensity column header/unit switches with the active sport (`[km/h]`
  running, `[W]` cycling).
- The grid supports keyboard navigation between cells, add/remove/reorder rows,
  and paste from the clipboard (⚠ OI-10 — define accepted paste formats:
  tab-separated, CSV, both).
- Time point is entered and displayed as `mm:ss`.

#### FR-T6 — Input validation
The system **shall** validate entered values and surface implausible or
incomplete data.

*Acceptance criteria*
- A test must have at least **4** valid steps (intensity + lactate present)
  before analysis runs — the minimum a 3rd-order polynomial requires; below 5,
  analysis runs but warns (proposed — ⚠ OI-11). (Reference dataset has 8 loaded
  steps + baseline.)
- Each numeric field is range-checked (proposed bounds — ⚠ OI-12): lactate
  0.00–30.00 mmol/L (2 dp); HR 0–250 bpm (integer); running intensity 0.0–45.0
  km/h (1 dp); cycling intensity 0–2000 W (integer); Borg/RPE 6–20 (integer).
- A **low-intensity lactate dip** (a lactate value lower than the previous step
  at low intensity, e.g. 8 km/h 1.19 < 6 km/h 1.24) is accepted without error
  and handled by fitting (FR-F3), not rejected.
- An **aborted final step** shorter than the protocol step duration (e.g. a 20
  km/h step ending at 22:10, ~1:10 in) is accepted and marked; it is included
  in analysis. ⚠ OI-13 — confirm whether an aborted step is included or excluded
  from the fit by default.
- Implausible values are flagged non-blockingly (the user may keep them).

#### FR-T7 — Predefined templates
The system **shall** ship predefined entry templates for **Running (Lauf)** and
**Cycling (Rad)** (Appendix B).

*Acceptance criteria*
- Selecting a template presets sport, intensity unit, default step duration,
  default increment, starting intensity, and visible columns to the Appendix B
  values.
- All preset values remain editable per test after the template is applied.

#### FR-T8 — User templates
The system **shall** allow the user to create, save, edit, and delete their own
templates (a named protocol + column configuration), including by cloning a
predefined template.

*Acceptance criteria*
- A saved user template is selectable when starting a new test.
- Editing or deleting a user template does not alter tests already created from
  it.

### 5.3 Curve Fitting

#### FR-F1 — Fit lactate–intensity curve
The system **shall** fit a lactate-vs-intensity curve from the step data.

*Acceptance criteria*
- Only steps with both intensity and lactate present participate in the fit.
- The fitted curve is available to all curve-dependent methods (§6) and to the
  fitting view (FR-C1).

#### FR-F2 — Fit strategies
The system **shall** default to a **3rd-order polynomial** fit and **shall**
offer user-selectable alternative fits (exponential, smoothing/penalised
spline). Methods that require a specific fit (log-log, Exp-Dmax) **shall** use
their own canonical fit per §6, independent of the displayed-curve default.

*Acceptance criteria*
- Changing the displayed fit does not change a method that pins its own fit.
- The fixed-threshold values for the Appendix A data, computed on the default
  3rd-order polynomial, reproduce the Appendix C parity targets within the OI-1
  tolerance.

#### FR-F3 — Guard against non-physiological fits
The system **shall** detect and warn when a fit is poorly conditioned or
non-physiological (e.g. a non-monotonic wiggle producing spurious local minima).

*Acceptance criteria*
- A warning is surfaced (non-blocking) when the fit produces ⚠ OI-14
  (define detection rule, e.g. local minimum above baseline, or R²/condition
  threshold).
- Threshold detection still runs but flags affected methods.

#### FR-F4 — Reactive recompute
The system **shall** recompute the fit and all dependent thresholds and zones
when the underlying data changes.

*Acceptance criteria*
- Editing a step value updates the fitted curve, thresholds, and zones without a
  manual "recompute" action.
- Recompute latency meets NFR-3.

### 5.4 Threshold Detection

#### FR-D1 — Implement threshold methods
The system **shall** compute the threshold methods listed in §6.

*Acceptance criteria*
- Each method in §6 is implemented and produces an intensity for the Appendix A
  data (or an explicit "not computable" result with a reason).
- Each method's output matches `lactater` within the OI-1 tolerance (see V1).

#### FR-D2 — Enable/configure methods
The system **shall** let the user enable/disable individual methods and
configure their parameters (e.g. fixed OBLA concentrations, baseline delta).

*Acceptance criteria*
- Disabled methods are excluded from the results table and zone anchoring.
- Configured parameters persist with the test and are reflected in recompute.

#### FR-D3 — Report intensity and interpolated HR
For each computed threshold, the system **shall** report the corresponding
intensity and the interpolated heart rate.

*Acceptance criteria*
- HR is interpolated on the HR-vs-intensity relationship (Appendix C: ≈167 bpm
  at 16.1 km/h).

#### FR-D4 — Label fit dependency; no silent mixing
The system **shall** clearly label which fit each method depends on and **shall
not** silently combine an incompatible fit/method pair.

*Acceptance criteria*
- Each results-table row indicates its underlying fit.
- Selecting an incompatible fit for a method either auto-pins the method's
  canonical fit or warns; it never produces a silently wrong value.

#### FR-D5 — Threshold results table
The system **shall** produce a threshold results table (Appendix C) listing, for
each marker: intensity, lactate, heart rate, % of max performance, pace (per
1000 m), and kcal/h (optional, requires body mass).

*Acceptance criteria*
- Markers include fixed thresholds (2 / 4 / 6 mmol/L), IAS, IANS, and MAX.
- Derived values use the confirmed formulas in Appendix C (% of max, pace/1000m,
  interpolated HR).
- kcal/h shows 0 / blank when body mass is absent. ⚠ OI-15 — kcal/h requires a
  named metabolic formula; the model is unspecified.
- For the Appendix A data with IANS ← OBLA 4.0, the table reproduces **16.1
  km/h / 167 bpm** for IANS (parity check, V2).

### 5.5 Training Zones

#### FR-Z1 — 5-zone model anchored on LT1/LT2
The system **shall** derive a **5-zone training model** (§7) anchored on the
aerobic threshold **LT1 (IAS)** and anaerobic threshold **LT2 (IANS)**.

*Acceptance criteria*
- Exactly five zones are produced (REKOM, GA1, GA2, EB, SB).

#### FR-Z2 — Automatic LT1/LT2 selection
The system **shall** compute LT1 and LT2 automatically from the §6 methods via
configurable default mappings.

*Acceptance criteria*
- Default mappings are configurable (e.g. LT1 ← Log-log / LTP1 / Bsln+; LT2 ←
  ModDmax / LTP2 / OBLA 4.0). ⚠ OI-16 — fix the **shipped default** mapping for
  each sport (the reference report used IANS ← OBLA 4.0).
- Changing the mapping recomputes LT1/LT2 and all zones.

#### FR-Z3 — Manual LT1/LT2 override
The system **shall** allow the user to override LT1 and LT2 by (a) selecting a
different computed method as anchor, (b) editing the intensity directly, or (c)
dragging the marker on the chart (FR-C2).

*Acceptance criteria*
- An overridden anchor persists and is flagged as manual (distinct from
  algorithmic).
- Removing the override restores the automatic value.

#### FR-Z4 — Live zone recompute
The system **shall** express each zone as ranges for intensity (km/h / pace /
W), heart rate, lactate, and pace (per 1000 m), recomputed live whenever LT1 or
LT2 changes.

*Acceptance criteria*
- A change to LT1 or LT2 updates all five zones' ranges within NFR-3 latency.

#### FR-Z5 — Training profiles
The system **shall** derive zone boundaries from a selectable **training
profile** (sport + athlete level + weekly frequency) and **shall** ship a small
set of predefined profiles plus allow user-defined ones.

*Acceptance criteria*
- At least the profile "Laufen Leistungssportler (6×/Woche)" is shipped and
  reproduces the Appendix C reference zones within the OI-1 tolerance.
- ⚠ OI-17 — enumerate the predefined profiles shipped in v1 and define the
  spread rule (how each profile positions zone bounds around IAS/IANS).

#### FR-Z6 — Derived metrics per threshold/zone
The system **shall** derive, for each threshold and zone bound: pace (min/km), %
of max intensity, and interpolated HR; kcal/h is optional and requires body mass
plus a metabolic formula.

*Acceptance criteria*
- Derived values match Appendix C formulas.
- kcal/h is omitted/zero without body mass (see OI-15 for the formula gap).

### 5.6 Visualisation & Interactive Correction

#### FR-C1 — Fitting view
The system **shall** provide a lactate-vs-intensity chart (fitting view) showing
the fitted curve, raw points, heart-rate overlay, threshold markers, and zone
bands.

*Acceptance criteria*
- All listed layers can be displayed together; ⚠ OI-18 — confirm which layers
  are individually toggleable.

#### FR-C2 — Draggable threshold markers
The system **shall** make threshold markers draggable; dragging **shall**
recompute the corresponding intensity, heart rate, pace, and downstream zones
live.

*Acceptance criteria*
- Dragging a marker updates intensity/HR/pace and all dependent zones within
  NFR-3 latency.
- The dragged threshold becomes a persisted manual override (FR-Z3 / FR-C3).

#### FR-C3 — Distinguish manual corrections
The system **shall** visually distinguish a manually corrected threshold from an
algorithmic one and persist it as an override.

*Acceptance criteria*
- Manual vs. algorithmic state is visible in both the chart and the results
  table.

#### FR-C4 — Temporal view (primary results chart)
The system **shall** provide a temporal results chart ("Zeitliche Darstellung",
Appendix C) with: X-axis = time; dual Y-axes = lactate (left) and heart rate
(right); intensity (km/h **or** W) as step bars; the five zone bands overlaid;
IAS and IANS as vertical lines; and configurable reference lines.

*Acceptance criteria*
- The chart reproduces the Appendix C layout for the Appendix A data.
- Reference lines (e.g. 4 mmol/L, threshold HR, threshold speed) are
  user-configurable.
- The chart exports at print/vector quality (FR-R2, FR-R5).

### 5.7 Comparison & History

#### FR-H1 — Longitudinal view
The system **shall** plot a single athlete's threshold/zone progression across
their tests over time.

*Acceptance criteria*
- Selecting an athlete with ≥2 tests renders a time series of at least IAS and
  IANS. ⚠ OI-19 — confirm which metrics are plotted longitudinally.

#### FR-H2 — Cross-sectional view
The system **shall** compare multiple tests/athletes side by side.

*Acceptance criteria*
- The user can select ≥2 tests (possibly across athletes) and view their
  thresholds/zones in a comparison layout.

#### FR-H3 — Overlay lactate curves
The system **shall** overlay multiple lactate curves on one chart for visual
comparison.

*Acceptance criteria*
- ≥2 selected tests render their fitted curves on shared axes, each visually
  distinguishable.

### 5.8 Reporting

#### FR-R1 — Printable PDF report
The system **shall** generate a printable PDF report for a test (system print
dialog and export to PDF).

*Acceptance criteria*
- The report exports to a PDF file and is printable via the OS print dialog.
- ⚠ OI-20 — define default page size (A4 vs. Letter) and orientation.

#### FR-R2 — Report content blocks
The report **shall** be composed of these content blocks: cover +
athlete/anamnesis; test remarks; the temporal chart (vector, print quality); the
raw step-data table; the threshold results table (FR-D5); the training zones
(§7); and an evaluation (Bewertung) section.

*Acceptance criteria*
- Each listed block can appear in the report and renders the corresponding data.

#### FR-R3 — User-editable template blocks
The report **shall** be composed of user-editable template blocks the user can
include, omit, reorder, and edit (free-text commentary/evaluation).

*Acceptance criteria*
- The **default** report shows reference-report **page 2** (raw step-data table
  + temporal chart) and **page 3** (threshold results table + training zones).
- The cover/anamnesis (page 1) and evaluation (page 4) blocks are optional
  add-ons, off by default.
- Reordering and omitting blocks is reflected in the generated PDF.

#### FR-R4 — Custom header/logo and pagination
The report **shall** support an optional custom header/logo and footer and
**shall** produce a multi-page, paginated layout.

*Acceptance criteria*
- A user-supplied logo and footer text appear on the rendered report.
- Content spanning multiple pages paginates without clipping.

#### FR-R5 — Export chart and data
The system **shall** export the chart as an image (PNG/SVG) and results/zones as
CSV.

*Acceptance criteria*
- Chart export produces a valid PNG and a valid SVG.
- Results/zones export produces a CSV matching the on-screen tables.

### 5.9 Data Management

#### FR-M1 — Local SQLite persistence
The system **shall** persist all data in a single local SQLite database.

*Acceptance criteria*
- All athletes, tests, steps, results, zones, templates, and report settings
  persist across restarts in one database file.

#### FR-M2 — CSV import/export
The system **shall** import step data from CSV and export athletes/tests to CSV.

*Acceptance criteria*
- Importing a CSV in the canonical column layout (Appendix A) populates a test's
  steps.
- ⚠ OI-21 — define the exact CSV schema(s) for import and export (headers,
  delimiter, decimal separator, time format).

#### FR-M3 — Backup / restore
The system **shall** back up and restore the database as a single portable file.

*Acceptance criteria*
- Backup produces one file that, when restored on a clean install, reproduces
  the full dataset.

#### FR-M4 — No off-device data flow
The system **shall not** transmit data off-device, and no core feature **shall**
require a network connection.

*Acceptance criteria*
- With networking disabled, every §5 feature still functions.
- No outbound network calls are made by core features (verifiable by network
  monitoring during an end-to-end run).

---

## 6. Threshold Methods

Methods to implement, validated against `lactater`. References are the canonical
originators; confirm exact DOIs against `lactater`'s documented sources during
implementation.

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

> **Lactate-minimum test (LTmin):** noted but lower priority — it requires a
> specific (non-standard) protocol with a post-loading ramp, so it is a distinct
> workflow rather than a method over a standard step test. Tegtbur et al.
> (1993).

---

## 7. Training Zone Model (5-Zone)

The v1 zone model is the classic German/Austrian **5-zone endurance model**,
anchored on two automatically-computed, manually-overridable thresholds — **IAS**
(individuelle aerobe Schwelle ≈ LT1) and **IANS** (individuelle anaerobe
Schwelle ≈ LT2). Zone boundaries are produced by the selected **training
profile** (FR-Z5), which defines how the zones spread around IAS/IANS; all
boundaries remain configurable.

| Zone | German | English | Position |
|---|---|---|---|
| Z1 | REKOM | Recovery / compensation | below GA1 (regeneration) |
| Z2 | GA1 | Basic endurance 1 | around / below IAS |
| Z3 | GA2 | Basic endurance 2 | between IAS and IANS (lower) |
| Z4 | EB | Development | around IANS |
| Z5 | SB | Peak | above IANS |

Each zone is output as ranges for intensity (km/h / pace / W), heart rate,
lactate, and pace (per 1000 m), interpolated on the test data. Zones recompute
live whenever IAS or IANS is adjusted.

**Reference zones (profile "Laufen Leistungssportler, 6×/Woche"), from Appendix C:**

| Zone | km/h | Lactate | HR | Pace /km |
|---|---|---|---|---|
| GA1 | 7.4–11.3 | 1.2–1.5 | 107–126 | 08:06–05:19 |
| GA2 | 11.3–14.2 | 1.5–2.5 | 126–150 | 05:19–04:14 |
| EB | 14.2–16.4 | 2.5–4.4 | 150–169 | 04:14–03:39 |
| SB | 16.4–20.1 | 4.4–7.7 | 169–185 | 03:39–02:58 |

> The reference report displays the upper four zones (GA1–SB); **REKOM** is the
> recovery zone below GA1 and completes the 5-zone model. Zone naming is
> configurable (German labels by default, English optional).

---

## 8. Non-Functional Requirements

#### NFR-1 — Platform
The system **shall** run as a cross-platform desktop application on Windows,
macOS, and Linux.

*Acceptance criteria*
- A build runs and passes a smoke test on each of the three platforms. ⚠ OI-22 —
  define minimum OS versions per platform.

#### NFR-2 — Offline-first; no telemetry
The system **shall** function fully offline and **shall not** emit telemetry.

*Acceptance criteria*
- See FR-M4 verification. No telemetry endpoints exist in the build.

#### NFR-3 — Interactive recompute performance
Live threshold/zone recompute during marker dragging **shall** feel instant
(target **< 100 ms** per update).

*Acceptance criteria*
- For the Appendix A dataset on reference hardware (a mid-range 2020 laptop:
  4-core x86-64 ~2.5 GHz, 8 GB RAM, no discrete GPU — proposed, ⚠ OI-23), the
  p95 recompute latency on a drag step is < 100 ms.

#### NFR-4 — Data ownership
The system **shall** store data in a single-file, portable database with an
open, documented schema and CSV import/export.

*Acceptance criteria*
- The schema is documented in-repo.
- The database file is portable between installs (FR-M3).

#### NFR-5 — Privacy
The system **shall** treat personal athlete data as sensitive and **shall not**
transmit it off-device by default.

*Acceptance criteria*
- No personal data leaves the device under any default code path (FR-M4).

#### NFR-6 — Correctness / reproducibility
Threshold outputs **shall** be reproducible and validated against reference
values (§11).

*Acceptance criteria*
- Re-running analysis on identical input yields identical outputs.
- Outputs match `lactater` within the OI-1 tolerance.

#### NFR-7 — Usability
The system **should** be approachable to a non-clinician, with sensible defaults
such that the happy path (enter → fit → thresholds → zones → report) requires
minimal configuration.

*Acceptance criteria*
- A new user can complete the happy path for the Appendix A data without editing
  any default setting. ⚠ OI-24 — define how this is measured (task walkthrough
  vs. usability test).

#### NFR-8 — Footprint
The system **should** have a small install size and low resource use befitting a
desktop utility.

*Acceptance criteria*
- ⚠ OI-25 — set concrete targets (install size, idle RAM) or mark as
  best-effort/non-testable.

---

## 9. Data Model (indicative)

- **Athlete** — id, name, dob/age, sex, body_mass, sport, notes.
- **Test** — id, athlete_id, date, sport, protocol (step_duration, increment,
  start_intensity, mode, rest), baseline_lactate, body_mass_snapshot, notes.
  (`body_mass_snapshot` is copied from the athlete at test creation and editable
  per test, so editing the athlete never alters a past test's kcal/h — proposed,
  ⚠ OI-3.)
- **Step** — id, test_id, order, intensity, time_point, heart_rate, lactate, rpe
  (optional).
- **ThresholdResult** — id, test_id, method, intensity, heart_rate, is_override,
  fit_type.
- **Zone** — id, test_id, model, zone_index, intensity_low, intensity_high,
  hr_low, hr_high, lactate_low, lactate_high.
- **Template** — id, name, sport, protocol, visible_columns, is_predefined.
- **TrainingProfile** — id, name, sport, level, weekly_frequency, spread rule.
- **ReportSettings** — id, header/logo, footer, block order/visibility,
  commentary text.

> The model above is indicative, not normative. The authoritative schema is the
> documented SQLite schema (NFR-4).

---

## 10. Technical Architecture

*(Supporting context — informs but does not constrain the requirements above.)*

**Stack**

- **Application:** **Wails v2** — a single Go binary that hosts a native OS
  webview. The Go core and the desktop shell are the *same* binary; there is no
  separate shell process and no IPC/cgo bridge. Frontend calls Go directly via
  Wails bindings.
- **Core logic:** Go — threshold models, curve fitting, zone derivation. A
  standalone, separately-testable package (`turnpoint-core`) imported by the
  Wails app.
- **Frontend:** Svelte — reactive UI, entry grid, both chart views; assets
  embedded into the Go binary at build time.
- **Storage:** SQLite via a pure-Go driver (`modernc.org/sqlite`) to keep
  cross-compilation clean and the build cgo-light.

**Why Wails:** the core is Go and Wails *is* Go, so the Go↔shell integration
question disappears — no sidecar, no cgo FFI, one language, one `wails build`.
Trade-off accepted: Wails is desktop-first; a future mobile companion (if
pursued) will be a separate app rather than this same codebase.

**Build order (bottom-up, risk-first)**

1. **`turnpoint-core` Go package** with a test suite validated against
   `lactater`. Independently useful; the credibility-critical core.
2. **Wails + Svelte shell** — data entry, SQLite persistence, and both chart
   views on top of the core.
3. **Report export** — templating and PDF generation last.

**Effort (recalled estimate):** usable MVP (entry → fit → ~8 methods → chart →
simple PDF) ≈ 3–4 weeks part-time; polished v1 ≈ 5–7 weeks. Weight concentrates
in two places: the **curve-fitting + validation core** and the **report
generator**. Everything else is routine.

---

## 11. Validation & Testing

- **V1** — Each threshold method has unit tests asserting parity with `lactater`
  outputs within the defined tolerance (⚠ OI-1).
- **V2** — Maintain a set of reference datasets (`lactater`'s demo data, the
  Appendix A dataset, and published examples) with expected threshold values.
  The Appendix C parity targets are a required fixture.
- **V3** — Each method documents its source paper/DOI in code and in this spec.
- **V4** — Fit-strategy edge cases covered by tests: minimal step count (FR-T6),
  low-intensity lactate dip, aborted final step.
- **V5** — End-to-end test of the happy path (enter → fit → thresholds → zones →
  report) on the Appendix A data.
- **V6** — FR-M4/NFR-2 verified by confirming no outbound network activity
  during an end-to-end run.

---

## 12. Licensing

Source-available, **all rights reserved**. The repository may be published for
review/reference only; no license is granted to use, run, copy, modify, or
distribute. License text lives in `LICENSE` at the repo root.

---

## 13. Roadmap

| Phase | Deliverable |
|---|---|
| P0 | Go core: curve fitting + ~8 threshold methods, validated against `lactater` |
| P1 | Wails/Svelte shell: athlete + test CRUD, entry grid, SQLite persistence |
| P2 | Interactive chart with draggable threshold correction; zone derivation |
| P3 | PDF report generation |
| P4 | Longitudinal / cross-sectional comparison; remaining methods |
| Later | Mobile companion; device import (re-evaluate) |

---

## 14. Resolved Decisions

These were prior open questions, now settled (retained for traceability).

1. **Go ↔ shell integration** — Resolved: Go + **Wails v2** (single Go binary,
   native webview, no bridge). See §10. Mobile companion, if built, will be a
   separate app.
2. **Default fit strategy** — Resolved: default 3rd-order polynomial; per-method
   canonical fits otherwise (FR-F2). Standard markers fixed at 2 & 4 mmol/L;
   LT1/LT2 freely selectable.
3. **Pace representation** — Resolved: primary units km/h (running) and W
   (cycling); pace is a derived display metric. Row/swim deferred (FR-T4).
4. **Zone model defaults** — Resolved: v1 ships the 5-zone model (§7); 3-zone
   derivable later.
5. **Report templating** — Resolved: user-editable template blocks; default =
   reference report pages 2 & 3 (FR-R3).

---

## 15. Proposed Resolutions (pending confirmation)

Each item below had an open decision (flagged ⚠ above). A **proposed** default is
now stated so the spec reads as complete and testable. These are working
defaults, **not final** — the author should confirm or override each.
**Confidence** indicates how firmly the proposal is grounded (High = obvious /
reference-backed; Medium = reasonable convention; Low = needs real validation).

| ID | Area | Proposed resolution | Confidence |
|---|---|---|---|
| OI-1 | Validation | Parity tolerance vs. `lactater`, per metric: **intensity** \|Δ\| ≤ 0.1 km/h (running) / ≤ 2 W (cycling), **or** ≤ 1 % relative, whichever is larger; **HR** ≤ 1 bpm; **lactate** ≤ 0.05 mmol/L. Tighten after inspecting `lactater`'s actual output precision. | Medium |
| OI-2 | Athlete | Only **name** is required at creation; sex, sport, DOB, body mass, notes all optional. | High |
| OI-3 | Athlete | Store **`body_mass_snapshot` on the Test** (copied from the athlete at creation, editable per test). Editing the athlete never alters a past test's kcal/h. Reflected in §9. | High |
| OI-4 | Athlete | **Cascade-delete** the athlete's tests/steps/results/zones, gated behind a typed confirmation (re-type the athlete name) and an "export first" prompt. (Soft-delete/archive is the fallback if data loss is a concern.) | Medium |
| OI-5 | Athlete | List columns: **name, primary sport, last-test date, test count**. Search matches name + notes (case-insensitive substring). | Medium |
| OI-6 | Athlete | Store **date of birth** as canonical; age (age-at-test) is derived from the test date. | High |
| OI-7 | Athlete | Sex enumeration = **{male, female, unspecified}**; **not used** in any v1 calculation. Revisit if the kcal/h model (OI-15) becomes sex-dependent. | Medium |
| OI-8 | Athlete | Body mass in **kg, 1 decimal place, range 20.0–250.0**. | High |
| OI-9 | Sports | **Deferred** for v1, but a design constraint to honour: persist intensity as a canonical numeric in the sport's native unit (km/h or W); never store pace as the primary value. Verified by code review of the storage layer, not a runtime test. | Medium |
| OI-10 | Entry | Accept paste as **TSV and CSV**, auto-detecting delimiter (tab / comma / semicolon) and decimal separator (dot / comma). | High |
| OI-11 | Entry | **Minimum 4** steps with intensity + lactate (the minimum a 3rd-order polynomial needs); below **5**, analysis runs but warns. | High |
| OI-12 | Entry | Ranges: lactate 0.00–30.00 mmol/L (2 dp); HR 0–250 bpm (int); running 0.0–45.0 km/h (1 dp); cycling 0–2000 W (int); Borg/RPE 6–20 (int). Sanity bounds, not strict physiology. | Medium |
| OI-13 | Entry | **Include** the aborted final step in the fit by default (it is the MAX point in the reference report); mark it; allow per-step exclusion. | High |
| OI-14 | Fitting | Warn when **(a)** the fitted curve is non-monotonic over the tested intensity range (an interior local extremum), **or (b)** fit R² < 0.95. | Medium |
| OI-15 | Results | Proposed formulas (validate later): **Running** kcal/h = body_mass_kg × speed_kmh × 1.036 (net running cost ≈ 1 kcal·kg⁻¹·km⁻¹); **Cycling** kcal/h = power_W × 3.6 (mechanical 0.86 kcal/h per W ÷ ~24 % gross efficiency). Cannot be validated against the reference report (body mass = 0 there). | Low |
| OI-16 | Zones | Default anchors: **LT1 (IAS) ← Log-log**, **LT2 (IANS) ← OBLA 4.0** (reproduces the reference IANS 16.1 km/h / 167 bpm). **P0 caveat:** confirm which method reproduces the reference **IAS** (10.5 km/h / 1.4 mmol/L); if Log-log does not, switch the IAS default to the method that does (likely LTP1 or Bsln+). | Medium |
| OI-17 | Zones | Zones = **bands expressed as % of IANS intensity**. Ship 3 running + 3 cycling profiles: **Freizeit (3×/Wo), Ambitioniert (4–5×/Wo), Leistung (6×/Wo)**. The "Laufen Leistung 6×/Wo" bands, back-derived from Appendix C: **REKOM < 46 %, GA1 46–70 %, GA2 70–88 %, EB 88–102 %, SB 102–125 %** of IANS. Other profiles' percentages need reference tables (set in P0). | Medium (calibrated profile) / Low (others) |
| OI-18 | Chart | Independently toggleable layers: **fitted curve, HR overlay, threshold markers, zone bands**. Raw points always shown. | High |
| OI-19 | History | Longitudinal plot of **IAS and IANS** (intensity + interpolated HR) and **MAX intensity** over test date. | Medium |
| OI-20 | Report | Default **A4 portrait**; Letter selectable. | High |
| OI-21 | Data | CSV: **UTF-8**, header row of canonical column names, delimiter **comma** (semicolon selectable), decimal **dot** (comma selectable), time **mm:ss**. | High |
| OI-22 | Platform | Minimum OS: **Windows 10 (1903+), macOS 12 (Monterey), Linux with WebKit2GTK (glibc ≥ 2.31, e.g. Ubuntu 20.04+)** — aligned with Wails v2 support. | Medium |
| OI-23 | Perf | Reference hardware: **mid-range 2020 laptop — 4-core x86-64 ~2.5 GHz, 8 GB RAM, no discrete GPU**. | Medium |
| OI-24 | Usability | Measure NFR-7 by a **documented happy-path walkthrough** on the Appendix A data with no settings changed; optionally an informal usability test with 1–2 target users. | Medium |
| OI-25 | Footprint | Best-effort (non-blocking) targets: **installed size < 50 MB, idle RAM < 250 MB**. | Medium |

---

## 16. Glossary

| Term | Meaning |
|---|---|
| **LT1** | First (aerobic) lactate threshold. Mapped to IAS in v1. |
| **LT2** | Second (anaerobic) lactate threshold. Mapped to IANS in v1. |
| **IAS** | *Individuelle aerobe Schwelle* — individual aerobic threshold (≈ LT1). |
| **IANS** | *Individuelle anaerobe Schwelle* — individual anaerobic threshold (≈ LT2). |
| **LTP / LTP1 / LTP2** | Lactate turn point(s); first/second turn points via segmented regression. |
| **OBLA** | Onset of Blood Lactate Accumulation — intensity at a fixed lactate concentration (2/3/4/6 mmol/L). |
| **Bsln+** | Baseline-plus — intensity at baseline lactate + a fixed delta. |
| **Dmax / ModDmax / Exp-Dmax** | Maximum-perpendicular-distance threshold methods (standard, modified, exponential-fit variants). |
| **IAT** | Individual Anaerobic Threshold (Dickhuth): lactate minimum equivalent + 1.5 mmol/L. |
| **MAX** | Maximal step / peak performance reached in the test. |
| **REKOM / GA1 / GA2 / EB / SB** | The five training zones (recovery, basic endurance 1, basic endurance 2, development, peak). |
| **RPE / Borg** | Rating of Perceived Exertion (Borg 6–20 scale). |
| **`lactater`** | Reference R package; the parity benchmark for threshold outputs. |

---

## Appendix A — Reference Dataset (Running)

A real running step test, used as the canonical input-format example and as a
**P0 validation fixture**. It exercises two edge cases: a low-intensity lactate
dip (8 km/h, 1.19 < 1.24) and an aborted final step (the 20 km/h step ends at
22:10, ~1:10 in).

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

Protocol: 3:00 steps, +2 km/h increment, continuous. The 0 km/h row is the
resting/baseline row; the Borg/RPE column is constant here and may be sparse in
real use.

## Appendix B — Predefined Templates

Default protocols (all values editable per test):

| Template | Sport | Intensity unit | Step duration | Increment | Start | End | Steps |
|---|---|---|---|---|---|---|---|
| Running (Lauf) | Running | km/h | 3:00 | +2 km/h | 6 km/h | 22 km/h | 9 |
| Cycling (Rad) | Cycling | W | 4:00 | +40 W | 80 W | 440 W | 10 |

- **Running default:** 6, 8, 10, 12, 14, 16, 18, 20, 22 km/h — each step 3:00.
- **Cycling default:** 80, 120, 160, 200, 240, 280, 320, 360, 400, 440 W — each
  step 4:00.

Both templates use the canonical column set (intensity, time, HR, lactate,
Borg/RPE). Templates preset the protocol and visible columns; users can clone a
predefined template and save their own variants (FR-T8), and any field can be
changed per test.

## Appendix C — Reference Report (mesics "Lactate EXPRESS")

A full WinLactat / Lactate EXPRESS report for the Appendix A dataset (athlete
"Bogner Markus", 01.02.2025). Used as the **output reference** (chart style,
tables, report structure) and as **parity targets** for P0.

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
- **Pace per 1000 m** = 60 ÷ (km/h), formatted mm:ss (e.g. 60 ÷ 16.1 = 3.727 →
  03:43).
- **HR at threshold** = interpolated on HR-vs-intensity (≈ 167 at 16.1 km/h).
- **kcal/h** = 0 in this report because body mass was 0; requires body mass + a
  metabolic model (⚠ OI-15).

**Notes for parity:**

- IANS here equals the 4 mmol/L fixed threshold exactly (same km/h, lactate, HR)
  — i.e. IANS was configured as OBLA 4.0. P0 check: with LT2 ← OBLA 4.0,
  Turnpoint must reproduce **16.1 km/h / 167 BPM**.
- The fixed thresholds are reproducible from the Appendix A raw data by curve
  fit + interpolation (e.g. 4 mmol/L falls between the 16 km/h and 18 km/h steps
  → 16.1 km/h).
- Default report structure: p1 cover + anamnesis + test remarks; p2 raw-data
  table + temporal chart; p3 threshold table + training zones; p4 evaluation
  (Bewertung).
- "Nicht ausbelastet" (not maximally exhausted) and "EndLac 8.72 mmol" are
  free-text test remarks — supports FR-T3 / FR-R3.
