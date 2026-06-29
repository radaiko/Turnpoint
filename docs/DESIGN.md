# Turnpoint — Technical Design

> Source of truth for implementation. Derived from `docs/REQUIREMENTS.md` via a six-expert design pass, reconciled into the consolidated design below. Appendices A–F preserve each expert+s full detail.

---


---

# Turnpoint — Consolidated Technical Design (v1 / P0–P3)

## 1. Module / repo layout

**Resolved:** ONE Go module rooted at the repo (`github.com/radaiko/turnpoint`, `go 1.23`, toolchain 1.26). `turnpoint-core` lives as the isolated `core/` package tree — gonum-only, no wails/sqlite import. See §3 for why single-module beats a separate module.

```
turnpoint/                              # repo root · module github.com/radaiko/turnpoint
├── go.mod  go.sum                      # gonum, modernc.org/sqlite, wails/v2, maroto/v2, tdewolff/canvas
├── wails.json                          # "wailsjsdir":"./frontend/wailsjs", frontend cmds
├── main.go                             # wails.Run(&options.App{...})            (Doc 4)
├── app.go                              # App struct = Wails-bound facade (orchestrates core+store+report)
├── LICENSE
├── build/                              # appicon + darwin/ windows/ generated platform assets
│
├── core/                               # ═══ turnpoint-core: PURE compute · gonum-only · NO wails/sqlite ═══
│   ├── unit/        unit.go sport.go pace.go              # Sport/Unit/Intensity/Pace (OI-9)   [leaf]
│   ├── domain/      test.go step.go protocol.go warning.go# data records + Warning (data, never error)
│   ├── numeric/     roots.go segreg.go                    # bracketed Brent + mat.Eigen polyroot; segmented LS
│   ├── fit/         fit.go poly.go exp.go spline.go loglog.go quality.go eval.go
│   ├── threshold/   method.go obla.go baselineplus.go loglog.go dmax.go moddmax.go
│   │                expdmax.go ltp.go iat.go ltratio.go d2lmax.go max.go
│   ├── metrics/     derived.go hrcurve.go                 # %max, pace, kcal/h, HR interp (FR-D3/D5)
│   ├── zone/        zone.go profile.go                    # 5-zone %-of-IANS bands (FR-Z*)
│   ├── analysis/    analysis.go pipeline.go recompute.go  # ◄ PUBLIC API the Wails layer calls
│   ├── internal/testutil/  floatcmp.go golden.go          # OI-1 tolerances, golden loader, -update flag
│   ├── testdata/
│   │   ├── datasets/   appendix_a.json  lactater_demo.json
│   │   └── golden/
│   │       ├── winlactat/  appendix_c_markers.json  appendix_c_zones.json   # FROZEN ground truth
│   │       ├── lactater/   appendix_a.json  demo_data.json                  # out-of-band regen (R-gated)
│   │       └── turnpoint/  appendix_a_methods.json                          # -update drift snapshot
│   └── tools/regen-lactater/  main.go  lactater_export.R                    # NOT in CI
│
├── internal/                           # ═══ app-only Go (imports core; may import sqlite/wails) ═══
│   ├── store/
│   │   ├── db.go                        # Open() + DSN pragmas (driver "sqlite")
│   │   ├── migrate.go                   # PRAGMA user_version migrator (embed.FS)
│   │   ├── migrations/  0001_init.sql  0002_seed_predefined.sql
│   │   ├── models.go                    # DTO structs (nullable col = pointer)
│   │   └── athlete_repo.go test_repo.go step_repo.go threshold_repo.go
│   │       zone_repo.go template_repo.go profile_repo.go report_repo.go
│   ├── csvio/       csv.go time.go      # OI-21 import/export + paste parse (OI-10)
│   ├── backup/      backup.go           # VACUUM INTO / restore
│   ├── report/      blocks.go html.go maroto.go chartexport.go
│   └── service/     dto.go              # AnalysisConfig/Result DTOs marshalled to JS
│
├── frontend/
│   ├── index.html package.json tsconfig.json vite.config.ts svelte.config.js
│   ├── wailsjs/                          # GENERATED — App.{js,d.ts}, go/core/models.ts, runtime/
│   └── src/
│       ├── main.ts  App.svelte  app.css
│       ├── lib/
│       │   ├── api/        # typed wrappers over $wails/go/main/App (one mock seam)
│       │   ├── stores/     athletes.ts session.ts analysis.ts ui.ts config.ts
│       │   ├── components/ Button IconButton Field NumberField Select Tabs Table Modal Toast Toggle Tag EmptyState DataGrid.svelte
│       │   ├── charts/     FitChart.svelte TemporalChart.svelte  Axis ZoneBands DraggableMarker StepBars
│       │   ├── design/     tokens.css  fonts/{GeistSans,GeistMono}.woff2
│       │   └── format/     mmss.ts pacePer1000.ts tabularNum.ts parseTabular.ts
│       └── views/  Athletes TestEntry Analysis Zones Comparison Report  (.svelte)
│
└── docs/  REQUIREMENTS.md
```

Reconciliations: Doc 3's `app/store|csvio|backup|services` and Doc 4's `db/` → unified under `internal/` (app-only, un-importable externally; bound `App` stays at root per Wails `Bind`). Doc 1's `numeric/` + `derive/hr.go` → kept as `core/numeric` and folded into `core/metrics/hrcurve.go`. Doc 2's package set is the spine.

---

## 2. Build order (phase sequence, each independently testable)

| Phase | Scope | Independently testable at exit |
|---|---|---|
| **P0.0 Foundation** | `wails init -t svelte-ts`; set module path; vite aliases (`$wails`,`$lib`); embed assets; `tokens.css` + Geist woff2; `core/unit`, `core/domain` (+`Warning`). | Blank frameless shell runs on win/mac/linux (NFR-1 smoke). `go test ./core/unit/... ./core/domain/...` (pace mm:ss, FitPoints dedup/sort, unit symbol switch). |
| **P0.1 Numeric primitives** | `core/numeric` roots + segreg. | Brent/companion-matrix roots on known polynomials; segmented regression recovers synthetic breakpoints; determinism. |
| **P0.2 Fits** | `core/fit`: poly3 (QR, centred/scaled), analytic derivatives, exp NLS, FritschButland spline, log-log; `quality.go` FR-F3 guard. | Reproduce known cubic coeffs; R²/condition/monotonicity flags fire correctly; pinned fits ignore display selection (FR-D4). |
| **P0.3 Methods + metrics + zones** | `core/threshold` (one method/file), `core/metrics`, `core/zone` (Laufen Leistung 6×). | Each marker vs WinLactat golden within OI-1; HR interp = 167@16.1; zone bands reproduce Appendix C. |
| **P0.4 Pipeline** | `core/analysis` `Analyze`/`RecomputeZones`/`Validate`. | Full Appendix A → Appendix C parity (V2 anchors A+B); NFR-6 byte-identical reruns; V4 edge cases. **Entire core ships with zero UI/DB.** |
| **P1.0 Persistence** | `internal/store` DDL + user_version migrator + repos + DSN; `internal/csvio`; `internal/backup`. | Migrate-from-empty; CRUD round-trip; cascade delete + FK enforced; CSV Appendix A round-trip; VACUUM INTO backup + integrity-checked restore — all on a temp-file DB. |
| **P1.1 App shell** | `app.go` App + `OnStartup` ctx + `Bind`; `wails generate module`; frontend stores (session debounced autosave, analysis latest-wins), nav rail, Athletes CRUD, TestEntry + custom `DataGrid`. | Enter Appendix A → SaveSteps → Analyze returns Result; happy-path walkthrough (V5 partial); paste auto-detect (OI-10). |
| **P2 Charts & interaction** | `FitChart` (draggable markers → `RecomputeZones`), Analysis table + method toggles, `TemporalChart`, Zones view + anchor override controls, Comparison view. | Drag p95 < 100 ms on reference HW (NFR-3); temporal chart reproduces Appendix C layout; manual-vs-algorithmic styling (FR-C3). |
| **P3 Reporting** | `internal/report`: block model, HTML `@page` print route, `window.print()`, live-SVG reuse; maroto fallback + tdewolff/canvas export; CSV results/zones; `SaveFileDialog`. | Default report = pages 2&3; PNG/SVG/CSV valid; multi-page pagination without clipping. |
| **P4 History + remainder** | Longitudinal/cross-sectional; remaining methods (IAT, D2Lmax, LTratio polish); lactater golden layer once R run on a dev box. | Curve-only methods vs frozen lactater JSON within OI-1. |

---

## 3. Locked technical decisions

| Concern | Decision | Rationale / disagreement resolution |
|---|---|---|
| **Module structure** | **Single** Go module; core isolated under `core/`; firewall enforced by a `go list -deps ./core/...` allowlist test (fails on any import outside stdlib+gonum). | **Resolves Doc 4 (single) vs Doc 1/2/6 (separate module).** Single wins: no `go.work`/`replace` friction with `wails build` + cross-compile; the deps-guard test gives the compile-firewall benefit without multi-module overhead. Trivial to promote to its own module later. |
| **Module path** | `github.com/radaiko/turnpoint`. | git user `radaiko`. Resolves Doc 2 `radaiko/turnpoint-core` vs Doc 4 `aiko/turnpoint`. |
| **Numerical lib** | gonum v0.15.x — sole external dep of `core`. | SRS §10. |
| **Poly3 least squares** | `mat.QR` (`Factorize` + `SolveVecTo(dst,false,y)`) on **centred/scaled** Vandermonde `t=(v−mean)/sd`; `mat.Cond(X,2)` for the FR-F3 ill-conditioning flag. | Doc 1+2 agree on QR; adopt Doc 1's centring (raw Vandermonde is ill-conditioned). Read residual + condition number cheaply. |
| **Poly/exp derivatives** | **Analytic** (closed-form, exact). | No `fd` error; chain-rule the `1/sd` scale. |
| **Exp fit** | `optimize.Minimize` + `NelderMead` (optional `BFGS` polish), `fd.Gradient` for grad, **log-linear warm start** (`a₀=min(l)−ε`, regress `ln(lᵢ−a₀)`). | gonum is minimisation-only; NLS for `a+b·e^{cx}` diverges without the warm start. |
| **Smoothing/penalised spline (display fit)** | gonum **`interp.FritschButland`** (monotone cubic, gonum-native, has `PredictDerivative`); 2nd deriv via `diff/fd` `Central2nd`. **Defer** custom Eilers–Marx P-spline. | **Resolves Doc 1 (FritschButland) vs Doc 2 (AkimaSpline).** FritschButland passes FR-F3 by construction (monotone). No threshold method pins the display spline → deferring the P-spline carries **zero parity risk**. |
| **HR interpolation** | `interp.PiecewiseLinear` over (intensity, HR), baseline (0,0) excluded; shared by prod + tests. | FR-D3; reproduces 167 bpm @16.1. |
| **Level-set root** `L(v)=c` | hand-rolled bracketed **Brent/bisection** in `numeric/roots.go`; companion-matrix `mat.Eigen` polyroot fallback (matches R `polyroot`). Smallest in-range root + flag if multiple. | gonum `optimize` has no 1-D root finder. |
| **Segmented regression** | `numeric/segreg.go` continuous piecewise-linear basis + deterministic **grid search** over ordered knots minimising RSS. k=1=log-log, k=2=LTP. | NFR-6 reproducibility; optional Muggeo polish later. |
| **D2Lmax fit** | **Pinned 4th-order polynomial** (or P-spline), NOT the default cubic. FR-D4 flag in the row. | On a cubic `L''=2a₂+6a₃v` is linear → max always at an endpoint (degenerate). Highest-impact numeric call (Doc 1). |
| **LTratio fit** | Argmin of `L(v)/v` on the **default poly3 fine grid** (defer custom B-spline). | Avoids custom B-spline code; flagged as a lactater parity caveat. |
| **Float comparison** | `gonum/floats/scalar.{EqualWithinAbs, EqualWithinAbsOrRel, EqualWithinRel, Round}`. | OI-1 "abs OR rel, whichever larger" maps exactly to `EqualWithinAbsOrRel`. |
| **SQLite driver** | `modernc.org/sqlite`, driver name **`"sqlite"`**, `CGO_ENABLED=0`. | Pure-Go, clean cross-compile (SRS §10). |
| **DSN pragmas** | `foreign_keys(1)` + `busy_timeout(5000)` + `journal_mode(WAL)` + `synchronous(NORMAL)`; `db.SetMaxOpenConns(1)`. | FKs are per-connection + OFF by default → must be in DSN; single writer cures SQLITE_BUSY. |
| **Migrator** | Hand-rolled `PRAGMA user_version` + `embed.FS`, one tx per file. Fallback: golang-migrate `database/sqlite` (cgo-free) if complexity grows. | Zero-dep, ~60 LOC; golang-migrate default `sqlite3` driver is cgo (disqualified). |
| **FK strategy** | `ON DELETE CASCADE` athlete→test→{step, threshold_result, zone, report_settings}; template/profile → test/zone use `ON DELETE SET NULL`. | OI-4 single-statement subtree delete; provenance nulled, snapshots preserved (FR-T8). |
| **Backup / restore** | Backup = `VACUUM INTO`; restore = close → `PRAGMA integrity_check` → `os.Rename` → drop `-wal`/`-shm` → reopen + Migrate. | Online-consistent under WAL; raw cp of open WAL DB is unsafe. |
| **Storage canon** | time = INTEGER seconds (display mm:ss); intensity = REAL native unit, pace derived (OI-9); logo = BLOB. | No parse ambiguity; portable self-contained single file (NFR-4). |
| **report_settings DDL** | **Merged:** nullable `test_id` (global default + per-test, Doc 3) + `blocks_json` ordered `[]ReportBlock` (Doc 5) + `header_logo BLOB` + `show_page_numbers` + `page_size`/`orientation`. | **Resolves Doc 3 vs Doc 5 conflict** (logo BLOB vs path; column names). BLOB keeps backup self-contained. |
| **PDF approach** | **Primary:** HTML/CSS Paged Media (`@page`) + `window.print()` (OS "Save as PDF"), reusing the live Svelte chart `<svg>` as **true vector**. **Fallback:** maroto/v2 for headless/batch, chart embedded via tdewolff/canvas. | Zero-conversion vector chart + the editor IS the report. Wails exposes no headless print-to-PDF → fallback covers batch (see risks). |
| **Chart SVG → PDF/PNG/SVG** | `tdewolff/canvas` `ParseSVG` + `renderers.{PDF,PNG(DPI 300),SVG}`. | One engine for FR-R5 export + fallback report chart page. |
| **Charts** | Custom Svelte SVG on **LayerCake + d3-scale + d3-shape** (modular, not full d3). | Draggable markers + dual-axis + zone-bands + step-bars are a fight against any charting lib; few KB (NFR-8). |
| **Entry grid** | **Custom `DataGrid.svelte`** (~300 LOC). | Bespoke 5-col schema, unit-switch header, mm:ss mask, paste auto-detect; ag-Grid/Handsontable bust NFR-8 + look templated. |
| **Routing / state** | No router (ui-store selection state); functional-core/imperative-shell; debounced **latest-wins** async `analysis` store; drag → `RecomputeZones`. | Desktop app needs no deep-linking; referential transparency gives FR-F4 + NFR-6 for free. |
| **Design system** | Geist Sans + Geist Mono self-hosted woff2; cool-neutral + single deep-teal accent; hairline borders over shadow; tabular numerics everywhere. | §5. Self-host for offline (NFR-2/M4); never Google Fonts `<link>`. |
| **Parity ground truth** | **WinLactat (Appendix C) primary now**; lactater golden layer regenerated **once** out-of-band (R-gated), frozen JSON in `core/testdata/golden/lactater/`. CI never touches R. | **Resolves Doc 6 framing:** WinLactat ≠ lactater; CI stays R-free/cgo-light. |
| **fit.Kind enum** | `{Poly3, Exp, Spline, LogLog, Segmented, None}`; only Poly3/Exp/Spline are constructible `Fit` objects; LogLog/Segmented are breakpoint pseudo-fits (label only); None = MAX. DB `fit_type CHECK ∈ ('poly3','exp','spline','loglog','segmented','none')`. | Reconciles Doc 2 (4 kinds) + Doc 3 DDL (drop `'linear'`, add `'none'`) + Doc 1's loglog-vs-segmented distinction. |

---

## 4. Public Go API surface of `core/analysis` (what `app.go` calls)

```go
package analysis

type Input  struct{ Test domain.Test }
type Config struct {
    DisplayFit           fit.Kind                                  // default KindPoly3 (FR-F2)
    IncludeBaselineInFit bool                                      // parity knob, default false
    EnabledMarkers       []threshold.Marker                        // FR-D2
    MethodParams         map[threshold.Marker]threshold.Params
    LT1Anchor, LT2Anchor threshold.Marker                          // OI-16: default LogLog / OBLA4
    LT1Override, LT2Override *Override                             // FR-Z3/C3 manual
    Profile              zone.TrainingProfile                      // FR-Z5
}
type Override struct{ Intensity float64 }

type Anchor struct {
    Marker    threshold.Marker
    Intensity float64
    Manual    bool                                                 // FR-C3
    Metrics   metrics.DerivedMetrics
}
type Result struct {
    Fits         map[fit.Kind]fit.Fit
    DisplayFit   fit.Fit
    Thresholds   []threshold.Result                                // FR-D5 rows
    Markers      map[threshold.Marker]metrics.DerivedMetrics
    LT1, LT2     Anchor                                            // IAS / IANS
    Zones        []zone.Zone                                       // five (FR-Z1)
    MaxIntensity float64
    Warnings     []domain.Warning
}

// Full pipeline: data → fits → thresholds → LT1/LT2 → zones → metrics. Pure & total (NFR-6).
func Analyze(in Input, cfg Config) (Result, error)

// Marker-drag fast path (FR-C2/F4, NFR-3 <100ms): reuse prev.Fits, recompute ONLY anchors+zones+metrics.
func RecomputeZones(prev Result, in Input, cfg Config, lt1, lt2 float64) (Result, error)

// FR-T6 pre-analysis checks (step count OI-11, ranges OI-12) without computing.
func Validate(in Input) []domain.Warning

var (
    ErrInsufficientSteps = errors.New("analysis: need ≥4 fit steps (FR-T6)")
    ErrNoBaseline        = errors.New("analysis: method requires a baseline row")
)
```

Supporting public surface the shell also touches: `threshold.Default() []ThresholdMethod`, `threshold.For(...Marker) []ThresholdMethod`, `threshold.Marker.String()`; `zone.Predefined() []TrainingProfile`, `zone.LaufLeistung6() TrainingProfile`; `fit.New(k, pts)`, `fit.Poly/Exponential/Spline/LogLogSeg`; `unit.PaceFromKmh`, `unit.ParseClock/FormatClock`. **Error style:** sentinel `errors.Is`-checkable values for the few hard preconditions; everything the SRS calls "non-blocking" rides as `domain.Warning` **data** (never a Go error) inside `Result`/`fit.Quality`/`threshold.Result`. The Wails-bound `App.Analyze(testID int64, cfg AnalysisConfig)(AnalysisResult,error)` wraps these and persists the snapshot rows.

`threshold.Method.RequiredFit()` drives FR-D4 no-silent-mixing: OBLA·/Bsln+/Dmax/ModDmax/IAT/LTratio→`Poly3`; LogLog/LTP1/LTP2→`Segmented`; ExpDmax→`Exp`; D2Lmax→**`Poly4`** (pinned). The pipeline builds each required `Kind` once and dispatches, propagating each fit's `Quality().Warnings` into every matching `threshold.Result` (FR-F3).

---

## 5. Minimal-modern design system — "Instrument minimal"

**Type — Geist Sans (UI) + Geist Mono (all numerics), self-hosted woff2.** 13px base, weights 400/500/600 only. `font-variant-numeric: tabular-nums slashed-zero` on every grid cell, results/zone table, axis label.

```
--fs-eyebrow 11/16 600 +0.06em upper   --fs-caption 12/16 400   --fs-body 13/20 400
--fs-label   13/20 500                 --fs-h3 15/22 600        --fs-h2 18/26 600 -0.005em
--fs-h1 22/30 600 -0.01em              --fs-display 28/36 600 -0.015em
--font-mono "Geist Mono", ui-monospace, monospace
fallback: -apple-system, "Segoe UI", Roboto, sans-serif
```

**Color (`lib/design/tokens.css`, `[data-theme="dark"]` on `<html>`):**

```
LIGHT  --bg #F5F6F8  --surface #FFFFFF  --surface-2 #EEF0F3  --inset #E9ECF0
       --border #DEE1E6  --border-strong #C7CCD3
       --text #16181D  --text-muted #5A616B  --text-faint #8A919C
       --accent #15616D  --accent-hover #114E58  --accent-contrast #FFFFFF
       --focus rgba(21,97,109,.40)  --danger #B42318  --warn #B54708  --ok #207A4C
DARK   --bg #0E1014  --surface #16191F  --surface-2 #1E222A  --inset #11141A
       --border #272C35  --border-strong #353B45
       --text #E6E8EC  --text-muted #9AA2AD  --text-faint #6B727C
       --accent #4FB3C0  --accent-hover #6AC5D1  --accent-contrast #08191C
       --focus rgba(79,179,192,.45)  --danger #F0795C  --warn #E5A23B  --ok #5DBE8A
```

**Data-viz** (fills ~14% alpha, edges full): zone bands REKOM `#5B8DB8` · GA1 `#3E9C8F` · GA2 `#A6B24D` · EB `#D99A2B` · SB `#C2543F`; lactate `#15616D`; HR `#C2693A`; step-bars `--border-strong` @55%. Threshold lines: **algorithmic = `--text` solid 1px; manual override = `--accent` dashed** (FR-C3).

**Spacing/radius/elevation:** `--space` 4·8·12·16·24·32·48·64; `--radius-sm 4 / -md 6 / -lg 10 / -pill 9999`. Separation = **1px hairline borders, not shadows**; only `--shadow-pop` (popovers) + `--shadow-modal`. Focus ring `box-shadow: 0 0 0 3px var(--focus)`. 150ms ease, honour `prefers-reduced-motion`.

**Layout shell:** frameless window (`Frameless:true`, `CSSDragProperty:"--wails-draggable"`, mac `TitlebarAppearsTransparent`), custom 44px titlebar (app mark + theme toggle; OS draws controls). 240px **nav rail** (Athletes · Comparison · Settings-pinned-bottom) + a **segmented stage tab bar** in the test workspace (Entry → Analysis → Zones → Report) — stages are sequential, not peer destinations. No router; `ui` store holds `{activeSection, activeAthleteId, activeTestId, activeStage, theme}`.

---

## 6. Validation anchors (exact golden numbers + tolerances)

**Tolerances (OI-1), `core/internal/testutil/floatcmp.go`:** intensity(run) `EqualWithinAbsOrRel(got,want,0.1,0.01)`; intensity(cycle) `(…,2,0.01)`; HR `EqualWithinAbs(…,1)`; lactate `EqualWithinAbs(…,0.05)`.

**Fixture 1 — Appendix C markers** (poly3 on Appendix A, `winlactat/appendix_c_markers.json`):

| Marker | km/h (±0.1/1%) | Lactate (±0.05) | HR (±1) | %max | Pace/km |
|---|---|---|---|---|---|
| OBLA 2.0 | 13.1 | 2.0 | 140 | 65.5 | 04:34 |
| OBLA 4.0 | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| OBLA 6.0 | 17.6 | 6.0 | 177 | 87.8 | 03:25 |
| IAS | 10.5 | 1.4 | 122 | 52.7 | 05:41 |
| IANS | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| MAX | 20.0 | 7.7 | 185 | 100.0 | 03:00 |

**Derived-metric rounding caveat:** assert **intensity** vs displayed value (±0.1 km/h), but compute `%max = intensity/maxIntensity·100` and `pace = 60/kmh` from the engine's **unrounded** intensity and assert `%max` ±0.3, `pace` ±1 s — absorbs WinLactat's inconsistent display rounding (e.g. 17.6 display ⇒ internal ≈17.56).

**Anchor A — OBLA 4.0 ⇒ IANS (V2, binding P0):** LT2←OBLA 4.0, poly3. Curve crosses 4.0 between 16 (3.89) and 18 (6.66) ⇒ **16.1 km/h ±0.1**; HR piecewise-linear `16→166,18→180` at 16.1 ⇒ `166+0.05·14=166.7→167 ±1`; lactate 4.0 by definition. The IANS row must equal the OBLA 4.0 row field-for-field.

**Anchor B — zones (V2, "Laufen Leistung 6×/Wo", `Rule=SpreadPctIANS`, IANS=16.1):** boundary km/h = `pct·16.1`.

| Zone | %IANS | km/h (±0.1) | Lactate (±0.05) | HR (±1) | Pace |
|---|---|---|---|---|---|
| REKOM | 0–46 | <7.4 | — | — | — |
| GA1 | 46–70 | 7.4–11.3 | 1.2–1.5 | 107–126 | 08:06–05:19 |
| GA2 | 70–88 | 11.3–14.2 | 1.5–2.5 | 126–150 | 05:19–04:14 |
| EB | 88–102 | 14.2–16.4 | 2.5–4.4 | 150–169 | 04:14–03:39 |
| SB | 102–125 | 16.4–20.1 | 4.4–7.7 | 169–185 | 03:39–02:58 |

Check: `0.46·16.1=7.41→7.4`, `0.70→11.27→11.3`, `0.88→14.17→14.2`, `1.02→16.42→16.4`, `1.25→20.13→20.1` ✓. **Keep marker-anchor HRs (167@16.1, 177@17.6) as hard asserts**; treat **zone-edge HRs as within-tolerance-with-a-note** — they sit on the ±1 bpm boundary (interp at 11.3→127.15 vs report 126; at 14.2→150.7 vs report 150). A 1-bpm edge failure is the signal to revisit floor-vs-nearest rounding (the "tighten after inspecting output" OI-1 anticipates).

Plus: NFR-6 determinism (analysis run twice → byte-identical JSON); V4 edges (3 steps→`ErrInsufficientSteps`, 4→runs+`WarnFewSteps`; raw dip accepted, synthetic dip trips FR-F3; aborted step included→MAX 20.0/185/7.7, exclude variant shifts; pinned-fit methods identical across display fits); V6 core deps-allowlist guard.

---

## 7. Top risks & open caveats

1. **R/`lactater` absence (load-bearing).** WinLactat (Appendix C) ≠ `lactater`; they diverge for curve-shape methods (Dmax variants, LTP, log-log). Ship V1/V2 against frozen **WinLactat** goldens now; generate the **`lactater`** golden layer once on a dev box (`remotes::install_github("fmmattioni/lactater")`), commit with pinned commit SHA, CI consumes frozen JSON only. Curve-only methods have no parity oracle until then.
2. **D2Lmax is undefined on the default cubic** (highest numeric-ambiguity call). `L''` of a cubic is linear → no interior max. **Pin a 4th-order polynomial** (RequiredFit=`Poly4`) and flag the row (FR-D4). No `lactater` target and not in Appendix C → validate against a hand-computed fixture.
3. **IAS default mapping (OI-16) unconfirmed.** LT2←OBLA 4.0 is locked (reproduces 16.1/167). LT1←Log-log is **provisional** — must empirically confirm it reproduces 10.5 km/h / 1.4 mmol; if not, switch IAS default to LTP1 or Bsln+. Decide after running against `lactater`.
4. **LTP/log-log pre-interpolation parity.** `lactater` segments *interpolated* data (LTP) vs *raw* log-log points; the LTP grid step materially shifts breakpoints. Replicate by augmenting the canonical fit to a fixed grid (e.g. 0.1 km/h) and treat the step as a parity-tuned constant.
5. **Bsln+ baseline source.** Appendix A resting lactate `0.00` is a placeholder → Bsln+0.5 below min-measured = not computable. Rule: `L_base = resting if present and >0, else min measured lactate`. No Appendix-C parity row → calibrate against `lactater` directly.
6. **IAT (Dickhuth) & D2Lmax not in `lactater`.** No V1 oracle; need literature/hand-computed fixtures distinct from the `lactater` suite.
7. **kcal/h formula unvalidated (OI-15, Low).** Body mass = 0 in the reference report → cannot validate. Isolated in `metrics` (`KcalPerHourRunning = m·v·1.036`, `…Cycling = W·3.6`) for easy later swap; `HasKcal=false` + `WarnNoBodyMass` when mass absent.
8. **Wails v2 has no programmatic/headless print-to-PDF.** FR-R1's "export to a PDF file" rides on the OS dialog's Save-as-PDF in the primary HTML path. If silent/batch PDF is later mandated, promote the maroto fallback or wire native webview PDF APIs per platform.
9. **report_settings DDL reconciliation** (Doc 3 vs Doc 5) is resolved to nullable-`test_id` + `blocks_json` + `header_logo BLOB` + `show_page_numbers`; the plan must use the merged shape, not either source verbatim.
10. **LTratio B-spline parity.** Shipping `argmin L(v)/v` on the poly3 grid instead of `lactater`'s B-spline augmentation may shift the result; flagged for verification against `lactater`.
11. **Exponential NLS divergence** without the log-linear warm start; `c≤0` (no upward curvature) ⇒ flag model inappropriate, `Computable:false`.
12. **Smoothing P-spline deferred.** v1 ships FritschButland as the display spline; the full Eilers–Marx P-spline (custom de-Boor basis + GCV) is out of P0 scope — acceptable because no threshold method pins it.

Source SRS: `/Users/radaiko/dev/private/Turnpoint/docs/REQUIREMENTS.md`. Core root: `/Users/radaiko/dev/private/Turnpoint/core/`. App root: `/Users/radaiko/dev/private/Turnpoint/internal/` + `/Users/radaiko/dev/private/Turnpoint/frontend/`.

---

# Expert Design Appendices



---

## Appendix A — Numerical Algorithm Design

I have everything I need (gonum APIs confirmed via pkg.go.dev; `lactater` method internals confirmed from its source on rdrr.io). Here is the numerical design for `turnpoint-core`.

---

# turnpoint-core — Numerical algorithm design

## 0. gonum sub-packages → role

| Package | Use |
|---|---|
| `gonum.org/v1/gonum/mat` | Vandermonde least squares (`QR`), penalised-spline normal equations (`Cholesky`), conditioning (`Cond`), companion-matrix roots (`Eigen`) |
| `gonum.org/v1/gonum/optimize` | Exponential NLS (`Minimize` + `NelderMead`, refine `BFGS`) |
| `gonum.org/v1/gonum/diff/fd` | Numerical gradient for the NLS objective (`Gradient`); 2nd derivative of non-polynomial fits (`Derivative` + `Central2nd`) |
| `gonum.org/v1/gonum/interp` | HR-vs-intensity interpolation (`PiecewiseLinear`); spline display fit (`FritschButland`) |
| `gonum.org/v1/gonum/stat` | `MeanStdDev` (centre/scale), `LinearRegression`, `RSquared` for fit quality + segment fits |

Confirmed signatures (current gonum): `func (qr *mat.QR) SolveVecTo(dst *VecDense, trans bool, b Vector) error`; `func optimize.Minimize(p Problem, initX []float64, s *Settings, m Method) (*Result, error)` with `Problem{Func func([]float64) float64; Grad func(grad,x []float64)}`; `func fd.Derivative(f func(float64) float64, x float64, s *Settings) float64` with `fd.Central2nd` (2nd-deriv stencil, step 1e-4); `func fd.Gradient(dst []float64, f func([]float64) float64, x []float64, s *Settings) []float64`; `interp.PiecewiseLinear`/`FritschButland` implement `Fit(xs,ys []float64) error` + `Predict(x float64) float64` (latter also `PredictDerivative`); `func stat.LinearRegression(x,y,weights []float64, origin bool) (alpha,beta float64)`; `func stat.RSquared(...) float64`; `func stat.MeanStdDev(x,weights []float64) (mean,std float64)`.

Note on `(*mat.Dense).Solve`: gonum's `Solve` *does* return the least-squares solution for full-rank `m≥n` (a small-model doc fetch claimed otherwise — it is wrong). I still recommend explicit `mat.QR` so we can read back the residual and call `mat.Cond` for the FR-F3 ill-conditioning warning.

## 1. File layout

```
turnpoint-core/
  fit/        fit.go (Fit interface)  poly.go  exp.go  pspline.go  bspline.go  eval.go
  threshold/  method.go (Method iface, Result)  obla.go  bslnplus.go  loglog.go
              dmax.go  ltp.go  iat.go  ltratio.go  d2lmax.go
  numeric/    roots.go (bracketed Brent + companion-matrix polyroot)
              segreg.go (continuous piecewise-linear LS, k knots, grid search)
  derive/     hr.go (HR interp, %max, pace)
```

Core interface:

```go
type Fit interface {
    Eval(v float64) float64     // fitted lactate L(v)
    Deriv(v float64) float64    // L'(v)
    Deriv2(v float64) float64   // L''(v)
    Domain() (lo, hi float64)   // [min,max] tested intensity
    Kind() FitKind
}
type Result struct {
    Method   string
    Intensity, Lactate, HR float64
    PctMax, PaceSecPerKm   float64
    FitKind  FitKind
    Computable bool
    Note     string          // reason when !Computable, or flag (FR-F3/D4)
}
```

---

## 2. Curve fits

### (1) 3rd-order polynomial LS — **default**
Model `L(v)=Σ aⱼ tʲ, j=0..3`, fit on **centred/scaled** `t=(v−mean)/sd` (Vandermonde on raw `v` is ill-conditioned).

```go
func FitPoly(v, l []float64, deg int) (*Poly, error) {
    n, p := len(v), deg+1
    mean, sd := stat.MeanStdDev(v, nil)
    X := mat.NewDense(n, p, nil)
    for i := range v { t, pw := (v[i]-mean)/sd, 1.0
        for j := 0; j < p; j++ { X.Set(i,j,pw); pw *= t } }
    y := mat.NewVecDense(n, append([]float64(nil), l...))
    var qr mat.QR; qr.Factorize(X)
    var c mat.VecDense
    if err := qr.SolveVecTo(&c, false, y); err != nil { return nil, err }   // least squares
    // R² via residuals; cond := mat.Cond(X, 2)
}
```
`Eval/Deriv/Deriv2` are closed-form on the cubic (chain-rule the `1/sd` scale). FR-F3 warning when `cond > 1e8` **or** R² < 0.95 **or** an interior `L'(v)=0` sign change exists (non-monotone wiggle). Needs ≥4 points (FR-T6); at n=4 it interpolates (R²=1) — warn (OI-11).

### (2) Exponential fit — own canonical fit for Exp-Dmax
Model **`L(v)=a+b·exp(c·v)`** (this is `lactater`'s form; its Exp-Dmax helper depends only on `c`). Nonlinear LS via `optimize` minimising `S(θ)=Σ(Lᵢ−a−b·e^{c·vᵢ})²`:

```go
prob := optimize.Problem{
    Func: sse,                                   // closure over v,l
    Grad: func(g, x []float64){ fd.Gradient(g, sse, x, nil) },
}
res, _ := optimize.Minimize(prob, init, &optimize.Settings{GradientThreshold:1e-8},
                            &optimize.NelderMead{})   // robust, derivative-free
// optional polish: re-Minimize from res.X with &optimize.BFGS{}
```
Init guess (critical for NLS convergence): `a₀=min(l)−ε`; regress `ln(lᵢ−a₀)` on `v` with `stat.LinearRegression` → slope=`c₀`, `b₀=exp(intercept)`. Edge cases: non-convergence → return `Computable:false`; if `c≤0` (no upward curvature) flag — exponential model inappropriate.

### (3) Penalised / smoothing spline
**Decision:** P-spline (Eilers–Marx): cubic B-spline basis `B` (≈10–15 equally spaced interior knots) with a 2nd-order difference penalty `D`; solve the penalised normal equations with Cholesky:

`(BᵀB + λ DᵀD) ĉ = Bᵀy` →
```go
var chol mat.Cholesky          // BtB+λDtD is SymDense
chol.Factorize(sym); chol.SolveVecTo(&coef, Bty)
```
λ chosen by minimising GCV over a log-spaced grid. **Flag:** gonum has no B-spline basis or smoothing-spline routine, so `bspline.go` (de Boor recursion) and the penalty matrix are custom; gonum supplies only the linear algebra. Pragmatic no-tuning alternative shipped alongside: `interp.FritschButland` (gonum-native, monotone cubic → passes FR-F3 by construction). Use the P-spline only as the user-selected "smoothing" display fit; **no threshold method pins it** (avoids parity drift).

---

## 3. Shared numerical primitives

- **Level-set root** `L(v)=c` on `[lo,hi]`: scan the fitted curve on a fine grid for a sign change of `L(v)−c`, then bracketed **Brent/bisection** (hand-rolled in `numeric/roots.go`; gonum `optimize` is minimisation-only, no 1-D root finder). General polyroot fallback = eigenvalues of the companion matrix via `mat.Eigen`, matching R's `polyroot()`. Pick the root in `[lo,hi]`; multiple roots ⇒ smallest in range + flag.
- **Quadratic root** (Dmax tangent): closed-form.
- **Segmented regression** `numeric/segreg.go`: continuous piecewise-linear basis `{1, v, (v−ψ₁)₊, …, (v−ψ_k)₊}`, LS per candidate via `mat.QR`; **grid-search** ordered knot tuples minimising RSS (deterministic ⇒ reproducible per NFR-6; optional Muggeo polish). k=1 → log-log, k=2 → LTP.
- **HR / derived** `derive/hr.go`: `interp.PiecewiseLinear` on (loaded-step intensity, HR) → `Predict(v*)`; `%max=v*/vmax·100`; `pace=60/v*` km/h → mm:ss. Baseline (0,0) row excluded from HR interp.

---

## 4. Per-method specification

Legend — **fit**: which canonical fit; **lt** = in `lactater` (has a hard parity target); ◆ = ambiguity flag.

### OBLA 2.0 / 3.0 / 4.0 — fit: cubic (default) · lt
- **Def:** intensity where `L(v)=c`, c∈{2,3,4,6}.
- **Algo:** `FitPoly` → level-set root of `L(v)−c` on `[lo,hi]` (Brent / `mat.Eigen` polyroot). HR via `PiecewiseLinear`.
- **Inputs:** loaded steps (v,l), HR series, c.
- **Edge:** c below min or above max fitted ⇒ not computable; non-monotone ⇒ smallest in-range root + flag.
- **Appendix A:** 4.0 → **16.1 km/h / 167 bpm** ✓ (linear bracket gives 16.08; cubic → 16.1). 2.0→13.1/140, 6.0→17.6/177. **This is the binding P0 parity (V2): IANS←OBLA 4.0 = 16.1/167.**

### Bsln+ 0.5 / 1.0 / 1.5 — fit: cubic · lt
- **Def:** intensity where `L(v)=L_base+Δ`.
- **Algo:** identical level-set root with `c=L_base+Δ`.
- **◆ Disambiguation:** Appendix A's resting row lactate = **0.00** (placeholder, not a real baseline) ⇒ Bsln+0.5 target 0.5 < min measured 1.19 = not computable. **Choice:** `L_base` = resting (intensity-0) lactate **if present and >0**, else **min measured lactate** (`lactater` uses the first/lowest stage value). No Appendix-C parity row exists for Bsln+, so calibrate against `lactater` directly.

### Log-log (Beaver) — fit: **own** (segmented on log-log points) · lt
- **Def:** `lactater` log-transforms **both** axes, fits `lm(ln L ~ ln v)`, then `segmented(npsi=1)`; breakpoint back-transformed `v*=exp(ψ)`.
- **Algo:** `segreg` k=1 over `(ln v, ln L)` **raw** points (not interpolated for log-log), grid-search ψ minimising RSS; `v*=exp(ψ)`. HR/lactate interpolated at `v*`.
- **Edge:** needs ≥4 points for 2 segments; any `L≤0` invalid for log.
- **Appendix A:** plausibly **≈10.5 km/h** (lactate ≈1.4) ⇒ candidate for **IAS** default (OI-16). ◆ Must verify the exact ψ against `lactater`; if log-log ≠ 10.5, switch IAS default to LTP1/Bsln+ (OI-16 caveat).

### Dmax — fit: cubic · lt
- **Def:** point of max perpendicular distance from the chord joining first & last **fitted** points. Equivalently (and as `lactater` computes it): the v where `L'(v)=` chord slope `m=(L(v_f)−L(v_0))/(v_f−v_0)`.
- **Algo:** solve quadratic `3a₃v²+2a₂v+(a₁−m)=0` (closed-form, on rescaled coeffs); pick root in `(v_0,v_f)`. (`lactater` uses `polyroot` on `P'−m`.)
- **Edge:** no in-range root ⇒ not computable.

### ModDmax — fit: cubic · lt
- **Def:** as Dmax but chord starts at `v_0`* = intensity of the **first step whose lactate rises >0.4 mmol/L above the previous step**; chord end = max intensity.
- **Algo:** find `v_0`* from raw steps, chord slope from **fitted** `L(v_0*),L(v_f)`, then same quadratic.
- **◆ Disambiguation:** "first rise >0.4" tested on **raw consecutive steps**; chord endpoints evaluated on the **fitted** curve (matches SRS "fitted points"). Appendix A: first Δ>0.4 is 14→16 (2.38→3.89) ⇒ `v_0*`=14.

### Exp-Dmax — fit: **exponential** (own) · lt
- **Def:** Dmax on `L=a+b·e^{cv}`. Tangent-parallel-to-chord reduces (as in `lactater`'s `exponential_dmax`) to a closed form in `c` only:
  `v* = ln( (e^{c·s_f} − e^{c·s_i}) / (c·(s_f − s_i)) ) / c`, with `s_i,s_f` = first/last intensity.
- **Algo:** `FitExp` → evaluate the closed form; lactate=`Eval(v*)`, HR interp.
- **Edge:** NLS failure ⇒ not computable; pinned to exponential fit regardless of displayed curve (FR-D4).

### LTP1 / LTP2 — fit: **own** (segmented k=2 on interpolated data) · lt
- **Def:** `lm(L~v)` on **interpolated** data → `segmented(npsi=2)`; LTP1=ψ₁, LTP2=ψ₂ (3 segments).
- **Algo:** interpolate fitted curve to a fine grid, `segreg` k=2 grid-search over ordered `(ψ₁<ψ₂)` minimising RSS.
- **◆ Disambiguation:** `lactater` segments **interpolated** (pre-smoothed) data, not raw — the interpolation grid step & interpolant materially shift the breakpoints. **Choice:** replicate by augmenting the canonical fit to a fixed grid (e.g. 0.1 km/h) and segmenting that; treat grid step as a parity-tuned constant. LTP1 is a strong **IAS** alternative; LTP2 an **IANS** alternative (OI-16).

### IAT (Dickhuth) — fit: cubic · **not in lactater**
- **Def:** lactate-minimum equivalent + **1.5 mmol/L**: `L_min=min_v L(v)` at the curve's lactate minimum; target `c=L_min+1.5`; IAT = intensity where `L(v)=c`, `v>v_min`.
- **Algo:** `L'(v)=0` interior root that is a minimum (the 8 km/h dip gives one) → `L_min`; level-set root for `L_min+1.5`.
- **Edge:** monotone curve (no interior min) ⇒ `v_min=lo`, `L_min=L(lo)`.
- **◆ Parity:** **no `lactater` target and not in Appendix C** — validate against literature / a hand-computed fixture. Appendix A trace: `L_min≈1.19`, target 2.69 ⇒ **≈14.4 km/h**.

### LTratio — fit: **B-spline** (own) · lt
- **Def:** intensity at the **minimum of `L(v)/v`**. `lactater` computes the ratio on its **B-spline-augmented** fine grid and takes `which.min`.
- **Algo:** B-spline (P-spline above with λ→small / interpolating) augment to fine grid, evaluate `L(v)/v`, `argmin` (or `optimize.Minimize` 1-D + `NelderMead`). Exclude v=0.
- **◆ Disambiguation:** ratio on the **fitted** (B-spline) curve, not raw points (`lactater` behaviour); minimum typically at low intensity.

### D2Lmax — fit: **higher-order poly or spline (pinned)** · not in lactater
- **Def:** intensity at the **maximum of `L''(v)`** (max acceleration).
- **◆ Critical ambiguity:** on a pure **cubic**, `L''(v)=2a₂+6a₃v` is linear ⇒ its max is always at a domain endpoint (degenerate); the exponential's `L''=bc²e^{cv}` is likewise monotone. An interior max of `L''` requires `L'''` to change sign. **Choice:** pin D2Lmax to a **≥4th-order polynomial** (or the P-spline), per Newell's higher-order fits, so `L''` has an interior maximum. **Algo:** poly → solve `L'''(v)=0` analytically and take the max; spline → `fd.Derivative(L, v, &fd.Settings{Formula:fd.Central2nd})` on a fine grid, then `argmax`. Flag clearly (FR-D4) that this method does **not** use the default cubic.

---

## 5. Ambiguity flags (for the plan author)

1. **D2Lmax is undefined on the default cubic** — must pin a ≥4th-order poly/spline (§4 D2Lmax). Highest-impact flag.
2. **Bsln+ baseline source** — resting 0.00 is a placeholder; define `L_base = resting if >0 else min(l)`.
3. **LTP/log-log pre-interpolation** — `lactater` segments *interpolated* data (LTP) vs *raw* log points (log-log); the LTP grid step is parity-sensitive. Replicate `lactater`'s augmentation, treat grid as a fixed constant.
4. **IAS default mapping (OI-16)** — must empirically confirm which method reproduces 10.5 km/h / 1.4 mmol (likely log-log ≈10.5 or LTP1); fix after running against `lactater`.
5. **IAT (Dickhuth) & D2Lmax are not in `lactater`** — no V1 parity oracle; need literature/hand-computed fixtures, distinct from the V1 `lactater` suite.
6. **ModDmax start rule** — "first rise >0.4" on raw consecutive steps; chord endpoints on the fitted curve.
7. **Exponential model form** — `a+b·e^{cv}` (matches `lactater`'s Exp-Dmax helper which is `c`-only); NLS needs the log-linear warm start or it diverges.

## 6. Appendix A → Appendix C parity trace (default cubic + linear HR)
Loaded steps {6..20 km/h}; `v_max=20`. OBLA 2/4/6 → 13.1/16.1/17.6 km/h with HR 140/167/177 (interp on bracketing steps: HR(16.1)=166+0.05·(180−166)=166.7→167). IANS←OBLA 4.0 reproduces **16.1 km/h / 167 bpm** (V2). IAS←log-log expected ≈10.5/1.4 (confirm). MAX=20.0/185 (last step). Derived: %max=v/20·100, pace=60/v.

Relevant files to create are under `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/` per the layout in §1 (`fit/`, `threshold/`, `numeric/`, `derive/`).

Sources: [lactater method-dmax](https://rdrr.io/cran/lactater/src/R/method-dmax.R), [method-ltp](https://rdrr.io/cran/lactater/src/R/method-ltp.R), [method-loglog](https://rdrr.io/cran/lactater/src/R/method-loglog.R), [method-ltratio](https://rdrr.io/cran/lactater/man/method_ltratio.html), [gonum optimize](https://pkg.go.dev/gonum.org/v1/gonum/optimize), [gonum interp](https://pkg.go.dev/gonum.org/v1/gonum/interp), [gonum diff/fd](https://pkg.go.dev/gonum.org/v1/gonum/diff/fd), [gonum stat](https://pkg.go.dev/gonum.org/v1/gonum/stat).

---

## Appendix B — Core Package Architecture

Read the SRS and verified the load-bearing gonum signatures (interp, diff/fd, mat QR, stat). Here is the architecture.

---

# turnpoint-core — Go package architecture

## 0. Module, layout, dependency DAG

Standalone module (no Wails, no `modernc.org/sqlite`, no DB — pure compute, SRS §10). Depends only on stdlib + `gonum.org/v1/gonum`.

```
turnpoint-core/                 # module github.com/radaiko/turnpoint-core, go 1.23
  go.mod                        # require gonum.org/v1/gonum v0.15.x  (ONLY external dep)
  unit/        unit.go sport.go pace.go      # Sport, Unit, Intensity, Pace — leaf
  domain/      test.go step.go protocol.go warning.go   # data records + Warning
  fit/         fit.go poly.go exp.go spline.go loglog.go quality.go
  threshold/   method.go obla.go baselineplus.go dmax.go ltp.go iat.go ltratio.go d2lmax.go loglog.go
  zone/        zone.go profile.go
  metrics/     derived.go hrcurve.go
  analysis/    analysis.go recompute.go pipeline.go   # PUBLIC API the Wails layer calls
  testdata/    appendix_a.json appendix_c.json        # P0 parity fixtures (V2)
```

Import DAG (acyclic): `unit ← domain ← {fit, zone, metrics} ← threshold ← analysis`. `fit`/`threshold` import gonum; everything else is gonum-free and trivially unit-testable. **Constraint that fixes the DAG:** `domain.Warning` must not reference `threshold.Marker` (would cycle) — Warning carries a `Subject string` (set to `marker.String()`), see §3/§9.

---

## 1. `unit` — Sport enum & unit handling (OI-9)

```go
package unit

type Sport uint8
const ( SportUnknown Sport = iota; Running; Cycling )   // Running→km/h, Cycling→W
func (s Sport) Unit() Unit      // UnitKmh | UnitWatt
func (s Sport) HasPace() bool   // true only for distance sports (Running)
func (s Sport) String() string

type Unit uint8
const ( UnitKmh Unit = iota; UnitWatt )
func (u Unit) Symbol() string   // "km/h" | "W"  (FR-T5 header switch)

// Intensity is ALWAYS the sport's native numeric (km/h or W). Pace is never
// stored as a primary value (OI-9 design constraint).
type Intensity = float64

// Pace = time per 1000 m. Derived display metric only (FR-T4).
type Pace time.Duration
func PaceFromKmh(kmh float64) Pace  // 3600/kmh seconds; Pace(0) if kmh<=0
func (p Pace) MMSS() string          // "mm:ss", e.g. 60/16.1 → "03:43"

// mm:ss clock helper for Step.TimePoint entry/display (FR-T5).
func ParseClock(s string) (time.Duration, error)
func FormatClock(d time.Duration) string
```

---

## 2. `domain` — pure data records (maps to §9 schema; DB layer owns DDL, out of my scope)

```go
package domain

type Mode uint8
const ( Continuous Mode = iota; Intermittent )

type Protocol struct {
    Sport          unit.Sport
    StepDuration   time.Duration
    Increment      float64        // native unit (+2 km/h | +40 W)
    StartIntensity float64
    Mode           Mode
    RestDuration   time.Duration  // Intermittent only (FR-T1)
}

type Step struct {
    Order      int
    Intensity  float64        // native unit; 0 == baseline row (FR-T3)
    TimePoint  time.Duration  // mm:ss from start
    HeartRate  int            // bpm
    Lactate    float64        // mmol/L
    HasLactate bool           // false ⇒ excluded from fit (empty cell, FR-T2)
    RPE        *int           // optional Borg 6..20
    Aborted    bool           // aborted final step, still in fit (OI-13)
    Excluded   bool           // user per-step fit exclusion (OI-13)
}

type Test struct {
    Protocol   Protocol
    Steps      []Step
    BodyMassKg float64        // body_mass_snapshot (OI-3); 0 ⇒ kcal/h disabled
}

// Helpers (pure): FitPoints returns sorted, dedup'd (intensity,lactate) for steps
// with HasLactate && !Excluded; baseline (intensity==0) inclusion governed by
// analysis.Config.IncludeBaselineInFit (parity knob, default false).
func (t Test) Baseline() (lactate float64, ok bool)  // the intensity==0 step
func (t Test) MaxIntensity() float64                 // peak loaded step ⇒ MAX / %max denom
```

---

## 3. `domain.Warning` — non-blocking diagnostics (FR-F3, FR-T6)

Warnings are **data, never Go errors** (everything in the SRS is "non-blocking"). They ride along in every result struct.

```go
type Severity uint8
const ( Info Severity = iota; Warn )

type WarnCode uint8
const (
    WarnFewSteps WarnCode = iota // <5 fit steps (OI-11)
    WarnNonMonotonicFit          // interior local extremum (OI-14a)
    WarnLowR2                    // R²<0.95 (OI-14b)
    WarnIllConditioned           // QR/Vandermonde near-singular
    WarnImplausibleValue         // out of OI-12 range
    WarnMethodNotComputable      // FR-D1 "not computable"
    WarnAbortedStep              // OI-13
    WarnNoBodyMass               // kcal/h disabled (FR-D5)
    WarnExtrapolated             // marker outside fitted domain
)
type Warning struct {
    Code     WarnCode
    Severity Severity
    Subject  string  // free key, e.g. "OBLA 4.0" or "fit:poly3" (avoids dep cycle)
    Message  string
}
```

---

## 4. `fit` — the `Fit` interface (FR-F1/F2/F3)

A fit is an immutable continuous model `lactate(intensity)` with analytic-or-numeric derivatives. **Polynomials differentiate analytically** (exact, fast); splines use `interp.AkimaSpline.PredictDerivative` for 1st and `diff/fd` `Central2nd` for 2nd.

```go
package fit

type Kind uint8
const ( KindPoly3 Kind = iota; KindExp; KindSpline; KindLogLog )
func (k Kind) String() string

type Point struct{ X, Y float64 }  // X intensity (native), Y lactate

type Fit interface {
    Kind() Kind
    Predict(x float64) float64          // fitted lactate at x
    Derivative(x float64) float64       // dL/dx  (Dmax/ModDmax families)
    SecondDerivative(x float64) float64 // d²L/dx² (D2Lmax)
    Domain() (xmin, xmax float64)       // fitted input range
    Quality() Quality                   // FR-F3/OI-14 diagnostics
}

type Quality struct {
    R2            float64
    Monotonic     bool      // strictly non-decreasing over Domain (OI-14a)
    Conditioned   bool      // R²≥0.95 AND well-conditioned solve (OI-14b)
    LocalExtremum *float64  // x of interior extremum, nil if none
    Warnings      []domain.Warning
}

// Factory + pinned builders (methods that pin a fit call these directly, FR-D4):
func New(k Kind, pts []Point) (Fit, error)
func Poly(pts []Point, order int) (*PolyFit, error)  // order=3 default
func Exponential(pts []Point) (*ExpFit, error)        // L = a + b·e^{c·x} via optimize
func Spline(pts []Point, lambda float64) (*SplineFit, error)
func LogLogSeg(pts []Point) (*LogLogFit, error)       // breakpoint in log–log

var (
    ErrTooFewPoints = errors.New("fit: need ≥ order+1 distinct points")
    ErrSingular     = errors.New("fit: design matrix ill-conditioned")
    ErrNonPositive  = errors.New("fit: log-log needs positive x and y")
)
```

**Poly core** (credibility-critical; uses the confirmed gonum `mat` QR API — Vandermonde least squares):

```go
func Poly(pts []Point, order int) (*PolyFit, error) {
    n := len(pts)
    if n < order+1 { return nil, ErrTooFewPoints }
    v := mat.NewDense(n, order+1, nil)
    y := mat.NewVecDense(n, nil)
    for i, p := range pts {
        xp := 1.0
        for j := 0; j <= order; j++ { v.Set(i, j, xp); xp *= p.X }
        y.SetVec(i, p.Y)
    }
    var qr mat.QR
    qr.Factorize(v)                 // *mat.QR.Factorize(Matrix)
    c := mat.NewVecDense(order+1, nil)
    if err := qr.SolveVecTo(c, false, y); err != nil {   // SolveVecTo(dst,trans,b)
        return nil, fmt.Errorf("%w: %v", ErrSingular, err)
    }
    coef := make([]float64, order+1)
    for j := range coef { coef[j] = c.AtVec(j) }
    return &PolyFit{coef: coef, xmin: pts[0].X, xmax: pts[n-1].X,
        q: assess(pts, coef)}, nil   // Predict=Horner, derivatives analytic
}
```

**FR-F3 / OI-14 mechanism** (`quality.go`): after building, sample the curve on ~200 points across `Domain`; (a) if `Derivative` changes sign at an interior point → set `LocalExtremum`, append `WarnNonMonotonicFit`; (b) `R² = stat.RSquaredFrom(estimates, values, nil)`; if `<0.95` append `WarnLowR2`; (c) condition check via `mat.Cond(v, 2)` over a threshold → `WarnIllConditioned`. These are non-blocking: detection still runs; the pipeline **propagates each fit's warnings into every threshold `Result` whose `FitKind` matches** (FR-F3 "flags affected methods").

---

## 5. `threshold` — `ThresholdMethod` interface & fit declaration (FR-D1/D2/D4)

```go
package threshold

type Marker uint8
const (
    OBLA2 Marker = iota; OBLA3; OBLA4; OBLA6
    Bsln05; Bsln10; Bsln15
    LogLog; Dmax; ModDmax; ExpDmax
    LTP1; LTP2; IAT; LTratio; D2Lmax
    MAX  // peak step, not curve-derived
)
func (m Marker) String() string  // "OBLA 4.0","Bsln+1.0","Log-log","ModDmax",...

type Params struct {            // FR-D2 configurable, persisted per test
    OBLAConc      float64       // 2.0/3.0/4.0/6.0
    BaselineDelta float64       // 0.5/1.0/1.5
}

type Context struct {
    Points          []fit.Point   // the fit input (for segmented methods, ModDmax chord)
    Steps           []domain.Step  // for MAX
    BaselineLactate float64
    HasBaseline     bool
    Params          Params
}

type Result struct {            // FR-D5 row; snapshot-friendly (§9 ThresholdResult)
    Marker     Marker
    Intensity  float64           // native unit; NaN if not computable
    Lactate    float64           // fit.Predict(Intensity)
    FitKind    fit.Kind          // FR-D4: row shows its underlying fit
    Computable bool
    Reason     string
    Warnings   []domain.Warning
}

// A method computes ONE marker. RequiredFit() is how it DECLARES its canonical
// fit (FR-D4) — the pipeline guarantees it receives a Fit of exactly that Kind,
// never the user's displayed-curve choice. No silent mixing.
type ThresholdMethod interface {
    Marker() Marker
    RequiredFit() fit.Kind
    Compute(f fit.Fit, ctx Context) Result
}

func Default() []ThresholdMethod          // all §6 methods, shipped params
func For(ms ...Marker) []ThresholdMethod  // subset for enabled markers (FR-D2)
```

**`RequiredFit()` per method** (drives no-silent-mixing, FR-D4): OBLA·/Bsln+/Dmax/ModDmax/IAT/LTratio/D2Lmax → `KindPoly3`; LogLog & LTP1/LTP2 → `KindLogLog`; ExpDmax → `KindExp`. The pipeline groups enabled methods by `RequiredFit()`, builds each `Kind` **once**, and dispatches. Compute strategies: OBLA = monotone root-find on curve to target conc; Bsln+ = root-find to `BaselineLactate+delta`; Dmax = `x` where `Derivative(x)==chordSlope(first,last)`; ModDmax = same with chord anchored at first raw step rising `>0.4` mmol/L; D2Lmax = argmax `SecondDerivative`; IAT = curve-min lactate `+1.5`, inverted; LTratio = argmin `Predict(x)/x`; LTP/LogLog = segmented-regression breakpoints; MAX = peak step (no fit, `FitKind` unset).

---

## 6. `zone` — 5-zone model, %-of-IANS band math (FR-Z1/Z5, OI-17)

The shipped band rule is **% of IANS intensity**, not interpolation between IAS↔IANS. The profile owns the percentage cut points; `IANS` is the single scaling anchor; `IAS` is passed through for future spread rules and as a displayed marker.

```go
package zone

type Index uint8
const ( REKOM Index = iota; GA1; GA2; EB; SB )
func (i Index) German() string   // "REKOM","GA1","GA2","EB","SB"
func (i Index) English() string  // "Recovery",...

type SpreadRule uint8
const ( SpreadPctIANS SpreadRule = iota /* future: SpreadIAStoIANS, SpreadPctHRmax */ )

type Band struct { Zone Index; LowPct, HighPct float64 } // fraction of IANS (1.0==IANS)

type TrainingProfile struct {     // §9 TrainingProfile; FR-Z5
    Name         string
    Sport        unit.Sport
    Level        string           // "Freizeit"|"Ambitioniert"|"Leistung"
    WeeklyFreq   int              // 3 | 4 | 6
    Rule         SpreadRule
    Bands        []Band           // 5, ascending, contiguous
    GermanLabels bool
}

type Zone struct {                // FR-Z4 output ranges
    Index                   Index
    Label                   string
    IntensityLow, IntensityHigh float64
    HRLow, HRHigh           int
    LactateLow, LactateHigh float64
    PaceLow, PaceHigh       unit.Pace  // distance sports only
}

// Derive: for each band, intensityBound = pct × ians; lactate = curve.Predict;
// HR = hr.At; pace = PaceFromKmh (Running). ias accepted for non-PctIANS rules.
func Derive(p TrainingProfile, ias, ians float64, curve fit.Fit, hr metrics.HRCurve, s unit.Sport) []Zone

func Predefined() []TrainingProfile   // 3 running + 3 cycling (OI-17)
func LaufLeistung6() TrainingProfile  // CALIBRATED to Appendix C
```

**Calibrated "Laufen Leistung 6×/Wo" bands** (back-derived from Appendix C, OI-17), `Rule=SpreadPctIANS`:

| Zone | LowPct | HighPct | @ IANS=16.1 km/h | Appendix C |
|---|---|---|---|---|
| REKOM | 0.00 | 0.46 | <7.4 | (below GA1) |
| GA1 | 0.46 | 0.70 | 7.4–11.3 | 7.4–11.3 ✓ |
| GA2 | 0.70 | 0.88 | 11.3–14.2 | 11.3–14.2 ✓ |
| EB | 0.88 | 1.02 | 14.2–16.4 | 14.2–16.4 ✓ |
| SB | 1.02 | 1.25 | 16.4–20.1 | 16.4–20.1 ✓ |

Other 5 profiles (Freizeit/Ambitioniert running + 3 cycling) ship as `TODO`-flagged tables to be filled from reference data in P0 (OI-17 marks them Low-confidence). **Math is `intensity = pct × IANS`**; lactate/HR/pace ranges are then read off the fit and HR curve at those intensities — so changing IANS (LT2) is the only thing that moves bands under this rule (drives the FR-C2 drag fast-path).

---

## 7. `metrics` — DerivedMetrics, exact Appendix C / OI-15 formulas

```go
package metrics

type DerivedMetrics struct {
    Intensity   float64
    PctMax      float64    // intensity / maxIntensity × 100
    Pace        unit.Pace  // 60/kmh → mm:ss; zero for cycling
    HasPace     bool
    HeartRate   int        // interpolated on HR-vs-intensity
    KcalPerHour float64
    HasKcal     bool       // false when BodyMassKg==0 (FR-D5)
    Lactate     float64
}

func PctMax(intensity, maxIntensity float64) float64 // intensity/max*100  (16.1/20=80.5)
// OI-15 proposed formulas (Low confidence — validate later):
func KcalPerHourRunning(bodyMassKg, speedKmh float64) float64 // m·v·1.036
func KcalPerHourCycling(powerW float64) float64               // W·3.6

// HRCurve: interpolated HR-vs-intensity (FR-D3; Appendix C ⇒ 167 bpm @16.1 km/h).
type HRCurve struct{ /* wraps interp.PiecewiseLinear over (intensity,HR) of HasLactate steps */ }
func NewHRCurve(steps []domain.Step) (HRCurve, error)  // PiecewiseLinear.Fit
func (c HRCurve) At(intensity float64) int             // PiecewiseLinear.Predict, rounded

func Derive(intensity, maxIntensity float64, s unit.Sport, hr HRCurve, bodyMassKg, lactate float64) DerivedMetrics
```

Pace uses `unit.PaceFromKmh` = `60/kmh` minutes formatted `mm:ss`. `PctMax` denom = `Test.MaxIntensity()`. kcal/h returns `HasKcal=false`, `0` when `bodyMassKg==0` and the pipeline emits `WarnNoBodyMass`.

---

## 8. `analysis` — the pipeline & PUBLIC API (what Wails binds)

```go
package analysis

type Input struct { Test domain.Test }

type Config struct {                            // user selections; persisted
    DisplayFit           fit.Kind               // default KindPoly3 (FR-F2)
    IncludeBaselineInFit bool                   // parity knob (default false)
    EnabledMarkers       []threshold.Marker     // FR-D2
    MethodParams         map[threshold.Marker]threshold.Params
    LT1Anchor            threshold.Marker        // OI-16 default LogLog (verify vs IAS 10.5)
    LT2Anchor            threshold.Marker        // OI-16 default OBLA4
    LT1Override          *Override               // FR-Z3/C3 manual
    LT2Override          *Override
    Profile              zone.TrainingProfile    // FR-Z5
}
type Override struct { Intensity float64 }       // presence ⇒ manual

type Anchor struct {
    Marker    threshold.Marker
    Intensity float64
    Manual    bool                  // FR-C3 distinguish manual vs algorithmic
    Metrics   metrics.DerivedMetrics
}

type Result struct {
    Fits         map[fit.Kind]fit.Fit
    DisplayFit   fit.Fit
    Thresholds   []threshold.Result                            // FR-D5 table
    Markers      map[threshold.Marker]metrics.DerivedMetrics   // %max/pace/HR/kcal per marker
    LT1, LT2     Anchor                                        // IAS / IANS
    Zones        []zone.Zone                                   // five (FR-Z1)
    MaxIntensity float64
    Warnings     []domain.Warning
}

// ── PUBLIC API SURFACE the Wails layer calls ──────────────────────────────
// Full pipeline: data → fits → thresholds → LT1/LT2 select → zones → metrics.
// Pure & total: identical (Input,Config) ⇒ identical Result (NFR-6).
func Analyze(in Input, cfg Config) (Result, error)

// Marker-drag fast path (FR-C2/F4, NFR-3 <100ms): reuse prev.Fits, recompute
// ONLY anchors+zones+their metrics for new LT1/LT2 intensities. No refit.
func RecomputeZones(prev Result, in Input, cfg Config, lt1, lt2 float64) (Result, error)

// FR-T6 pre-analysis checks (step count OI-11, ranges OI-12) without computing.
func Validate(in Input) []domain.Warning

var (
    ErrInsufficientSteps = errors.New("analysis: need ≥4 fit steps (FR-T6)")
    ErrNoBaseline        = errors.New("analysis: method requires a baseline row")
)
```

**Pipeline stages** (`pipeline.go`, all pure private funcs, composed by `Analyze`):

1. `pts := test.FitPoints(cfg.IncludeBaselineInFit)` → guard `len<4` ⇒ `ErrInsufficientSteps`; `len<5` ⇒ append `WarnFewSteps`.
2. `buildFits` — set of `Kind`s = `{m.RequiredFit() for enabled m} ∪ {cfg.DisplayFit}`; build each once → `map[Kind]Fit`. Fit warnings collected here.
3. `computeThresholds` — for each enabled method `m`: `r := m.Compute(fits[m.RequiredFit()], ctx)`; set `r.FitKind`; **merge that fit's `Quality().Warnings`** into `r.Warnings` (FR-F3 propagation).
4. `selectAnchors` — LT1 = `LT1Override` if set (Manual=true) else result of `LT1Anchor`; same for LT2 (FR-Z2/Z3).
5. `zone.Derive(cfg.Profile, LT1.Intensity, LT2.Intensity, DisplayFit, hrCurve, sport)` (FR-Z1/Z4).
6. `metrics.Derive` for every marker row + each zone bound + the two anchors (Appendix C).
7. aggregate warnings (fit + method + `WarnNoBodyMass` when `BodyMassKg==0`).

**Reactive recompute (FR-F4) modelled as pure functions — functional core / imperative shell:** the core holds **no state and no recompute trigger**. The Wails shell owns `(Input, Config)`; any edit yields a *new* `Input`/`Config` and re-invokes `Analyze` (cheap enough for NFR-3) — referential transparency gives FR-F4 "no manual recompute action" for free and NFR-6 reproducibility by construction. Marker dragging specifically routes through `RecomputeZones`, which skips stages 1–3 (the expensive curve fits are unchanged, only LT1/LT2 moved) and reruns stages 4–7, comfortably under the <100ms p95 budget.

---

## 9. Error-handling style — sentinel for control flow, typed value for non-blocking

- **Sentinel `errors.New` + `%w` wrapping** for the few hard preconditions a caller branches on with `errors.Is`: `fit.ErrTooFewPoints`, `fit.ErrSingular`, `fit.ErrNonPositive`, `analysis.ErrInsufficientSteps`, `analysis.ErrNoBaseline`. These abort the affected operation.
- **Typed value `domain.Warning` (NOT an error)** for everything the SRS calls "non-blocking": ill-conditioned/non-monotonic fit (FR-F3/OI-14), `<5` steps (OI-11), implausible inputs (OI-12), a method that is "not computable" (FR-D1 → `Result.Computable=false`, `Reason` set, `WarnMethodNotComputable`), aborted step (OI-13), missing body mass (FR-D5), extrapolation. Warnings never abort; they are carried as data in `Result`/`fit.Quality`/`threshold.Result` and surfaced by the UI. This matches the SRS's pervasive "flagged, the user may keep them" semantics and keeps the pipeline total.

No typed error structs — markers/subjects are conveyed via the `Warning.Subject string` field to keep `domain` dependency-free.

---

## 10. gonum API used (verified against current pkg.go.dev, not memory)

| Need | gonum symbol (confirmed signature) |
|---|---|
| Poly least squares | `mat.NewDense`, `mat.NewVecDense`, `(*mat.QR).Factorize(Matrix)`, `(*mat.QR).SolveVecTo(dst *VecDense, trans bool, b Vector) error`, `(*VecDense).AtVec(int)`, `mat.Cond` |
| R² / log-log segments | `stat.RSquaredFrom(estimates, values, weights []float64) float64`, `stat.LinearRegression(x, y, weights []float64, origin bool) (alpha, beta float64)` |
| HR + spline interpolation | `interp.FittablePredictor`; `(*interp.PiecewiseLinear).Fit(xs, ys []float64) error` / `.Predict(x float64) float64`; `(*interp.AkimaSpline).PredictDerivative(x float64) float64` |
| Spline 2nd derivative | `fd.Derivative(f func(float64) float64, x float64, s *fd.Settings) float64` with `fd.Central` / `fd.Central2nd` formulas |
| Exponential fit | `optimize.Minimize` (nonlinear least squares for `a+b·e^{cx}`) |

Polynomial 1st/2nd derivatives are computed **analytically** from coefficients (exact, no `fd` error) — `fd`/`AkimaSpline.PredictDerivative` are used only for the spline fit.

Key cross-references for the plan author: parity fixtures live in `testdata/` (V2); the `Config` defaults encode OI-16 (anchors) and OI-17 (Leistung profile); kcal/h formulas (OI-15, Low confidence) are isolated in `metrics` for easy later swap; `IncludeBaselineInFit` is the single parity knob if Appendix C targets miss.

---

## Appendix C — Persistence & App Backend

I have everything I need. Confirmed two decision-critical facts from current docs: (1) `modernc.org/sqlite` registers under driver name `"sqlite"` and is pure-Go/cgo-free; (2) golang-migrate's default `sqlite3` driver needs cgo, but its separate `sqlite` driver uses `modernc.org/sqlite` and is cgo-free. Here is the spec.

---

# Turnpoint — Persistence & Backend Services Design

## 0. Decisions at a glance

| Topic | Decision | One-line justification |
|---|---|---|
| Driver | `modernc.org/sqlite`, driver name `"sqlite"` | Pure Go, no cgo, clean cross-compile (SRS §10). |
| FK cascade | DB-level `ON DELETE CASCADE` from athlete down (OI-4) | Single `DELETE athlete` row removes tests/steps/results/zones atomically; UI adds typed-confirm + export-first prompt. |
| Migrations | **Hand-rolled versioned migrator keyed on `PRAGMA user_version`**, embedded `.sql` via `embed.FS` | Zero extra deps, cgo-free, ~60 LOC, right-sized for a single-file desktop DB. golang-migrate `database/sqlite` is the drop-in fallback if migration complexity grows. |
| Time storage | `time_point_s` as INTEGER seconds; display/CSV as `mm:ss` | Canonical numeric, no parse ambiguity, matches data-ownership (NFR-4). |
| Intensity storage | REAL in native unit (km/h or W); pace derived | OI-9: never store pace as primary. |
| Backup | `VACUUM INTO` (online, consistent) | Safe with open handles + WAL; file copy is not. |

---

## 1. Full SQLite Schema (DDL)

`PRAGMA foreign_keys` is **off by default** in SQLite — it must be enabled per connection (see §6 DSN). All booleans are `INTEGER CHECK (x IN (0,1))`; all timestamps are ISO-8601 TEXT; dates are `YYYY-MM-DD` TEXT.

```sql
-- ============ 0001_init.sql ============
PRAGMA foreign_keys = ON;

-- ---- Athlete (FR-A5, OI-2/6/7/8) ----
CREATE TABLE athlete (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,                          -- only required field (OI-2)
    dob           TEXT,                                       -- canonical; age derived (OI-6)
    sex           TEXT    NOT NULL DEFAULT 'unspecified'
                          CHECK (sex IN ('male','female','unspecified')),   -- OI-7
    body_mass_kg  REAL    CHECK (body_mass_kg IS NULL OR body_mass_kg BETWEEN 20.0 AND 250.0), -- OI-8
    primary_sport TEXT    CHECK (primary_sport IS NULL OR primary_sport IN ('running','cycling')),
    notes         TEXT    NOT NULL DEFAULT '',
    created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
-- name search is case-insensitive substring (OI-5): COLLATE NOCASE index
CREATE INDEX idx_athlete_name ON athlete(name COLLATE NOCASE);

-- ---- Template (FR-T7/T8) ---- (defined before test for the FK)
CREATE TABLE template (
    id              INTEGER PRIMARY KEY,
    name            TEXT    NOT NULL,
    sport           TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    step_duration_s INTEGER NOT NULL CHECK (step_duration_s > 0),
    increment       REAL    NOT NULL CHECK (increment > 0),
    start_intensity REAL    NOT NULL CHECK (start_intensity >= 0),
    end_intensity   REAL    CHECK (end_intensity IS NULL OR end_intensity >= start_intensity),
    mode            TEXT    NOT NULL DEFAULT 'continuous' CHECK (mode IN ('continuous','intermittent')),
    rest_duration_s INTEGER CHECK (rest_duration_s IS NULL OR rest_duration_s > 0),
    visible_columns TEXT    NOT NULL DEFAULT '["intensity","time","hr","lactate","rpe"]', -- JSON
    is_predefined   INTEGER NOT NULL DEFAULT 0 CHECK (is_predefined IN (0,1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_template_name ON template(name);

-- ---- Test (FR-T1/T3, OI-3) ----
CREATE TABLE test (
    id                 INTEGER PRIMARY KEY,
    athlete_id         INTEGER NOT NULL
                       REFERENCES athlete(id) ON DELETE CASCADE ON UPDATE CASCADE,  -- OI-4
    test_date          TEXT    NOT NULL,
    sport              TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    -- protocol snapshot (independent of template, FR-T8)
    step_duration_s    INTEGER NOT NULL CHECK (step_duration_s > 0),
    increment          REAL    NOT NULL CHECK (increment > 0),
    start_intensity    REAL    NOT NULL CHECK (start_intensity >= 0),
    mode               TEXT    NOT NULL DEFAULT 'continuous' CHECK (mode IN ('continuous','intermittent')),
    rest_duration_s    INTEGER CHECK (                                          -- FR-T1
                           (mode='intermittent' AND rest_duration_s IS NOT NULL AND rest_duration_s > 0)
                        OR (mode='continuous'   AND rest_duration_s IS NULL)),
    baseline_lactate   REAL    CHECK (baseline_lactate IS NULL OR baseline_lactate BETWEEN 0 AND 30),
    body_mass_snapshot REAL    CHECK (body_mass_snapshot IS NULL OR body_mass_snapshot BETWEEN 20.0 AND 250.0), -- OI-3
    pretest_note       TEXT    NOT NULL DEFAULT '',          -- FR-T3
    remarks            TEXT    NOT NULL DEFAULT '',          -- FR-T3 post-test
    template_id        INTEGER REFERENCES template(id) ON DELETE SET NULL,      -- provenance only; FR-T8
    created_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_test_athlete ON test(athlete_id, test_date);

-- ---- Step (FR-T2/T3/T6, OI-12/13) ----
CREATE TABLE step (
    id           INTEGER PRIMARY KEY,
    test_id      INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    step_order   INTEGER NOT NULL,                            -- 0..n; baseline conventionally 0
    intensity    REAL    NOT NULL CHECK (intensity >= 0),
    time_point_s INTEGER CHECK (time_point_s IS NULL OR time_point_s >= 0),  -- seconds, shown mm:ss
    heart_rate   INTEGER CHECK (heart_rate IS NULL OR heart_rate BETWEEN 0 AND 250),   -- OI-12
    lactate      REAL    CHECK (lactate IS NULL OR lactate BETWEEN 0 AND 30),          -- NULL ⇒ excluded from fit (FR-T2)
    rpe          INTEGER CHECK (rpe IS NULL OR rpe BETWEEN 6 AND 20),                   -- optional (FR-T2)
    is_baseline  INTEGER NOT NULL DEFAULT 0 CHECK (is_baseline IN (0,1)),               -- FR-T3
    excluded     INTEGER NOT NULL DEFAULT 0 CHECK (excluded IN (0,1)),                  -- per-step exclude (OI-13)
    aborted      INTEGER NOT NULL DEFAULT 0 CHECK (aborted IN (0,1)),                   -- aborted final step (OI-13)
    UNIQUE (test_id, step_order)
);
CREATE INDEX idx_step_test ON step(test_id, step_order);
CREATE UNIQUE INDEX uq_step_one_baseline ON step(test_id) WHERE is_baseline = 1;  -- exactly one baseline (FR-T3)

-- ---- ThresholdResult (FR-D1..D5) ----  snapshot rows (FR-A2: not retroactively edited)
CREATE TABLE threshold_result (
    id                    INTEGER PRIMARY KEY,
    test_id               INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    method                TEXT    NOT NULL,   -- 'OBLA_2.0','OBLA_4.0','BSLN_1.0','LOGLOG','DMAX','MODDMAX',
                                              -- 'EXPDMAX','LTP1','LTP2','IAT','LTRATIO','D2LMAX','IAS','IANS','MAX'
    intensity             REAL,               -- NULL ⇒ not computable (reason below)
    lactate               REAL,
    heart_rate            REAL,               -- interpolated (FR-D3) → REAL not INTEGER
    pct_max               REAL,               -- FR-D5
    pace_s_per_km         REAL,               -- FR-D5
    kcal_h                REAL,               -- NULL/0 when no body mass (FR-D5, OI-15)
    is_override           INTEGER NOT NULL DEFAULT 0 CHECK (is_override IN (0,1)),  -- FR-Z3/FR-C3
    fit_type              TEXT    NOT NULL CHECK (fit_type IN
                              ('poly3','exp','spline','loglog','segmented','linear')),  -- FR-D4
    not_computable_reason TEXT,               -- set iff intensity IS NULL
    params_json           TEXT,               -- configured params, e.g. {"obla":4.0} / {"delta":1.0} (FR-D2)
    UNIQUE (test_id, method)
);
CREATE INDEX idx_tr_test ON threshold_result(test_id);

-- ---- Zone (FR-Z1/Z4, §7) ----  snapshot rows
CREATE TABLE zone (
    id                INTEGER PRIMARY KEY,
    test_id           INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    model             TEXT    NOT NULL DEFAULT '5zone',
    zone_index        INTEGER NOT NULL CHECK (zone_index BETWEEN 1 AND 5),
    zone_name         TEXT    NOT NULL CHECK (zone_name IN ('REKOM','GA1','GA2','EB','SB')),
    profile_id        INTEGER REFERENCES training_profile(id) ON DELETE SET NULL,  -- which profile produced it
    intensity_low     REAL, intensity_high     REAL,
    hr_low            REAL, hr_high            REAL,
    lactate_low       REAL, lactate_high       REAL,
    pace_low_s_per_km REAL, pace_high_s_per_km REAL,
    UNIQUE (test_id, model, zone_index)
);
CREATE INDEX idx_zone_test ON zone(test_id);

-- ---- TrainingProfile (FR-Z5, OI-17) ----
CREATE TABLE training_profile (
    id               INTEGER PRIMARY KEY,
    name             TEXT    NOT NULL,
    sport            TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    level            TEXT    NOT NULL CHECK (level IN ('freizeit','ambitioniert','leistung')),
    weekly_frequency INTEGER CHECK (weekly_frequency IS NULL OR weekly_frequency > 0),
    -- spread rule: zone bands as fractions of IANS intensity (OI-17), e.g.
    -- {"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}
    spread_json      TEXT    NOT NULL,
    is_predefined    INTEGER NOT NULL DEFAULT 0 CHECK (is_predefined IN (0,1)),
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_profile_name ON training_profile(name);

-- ---- ReportSettings (FR-R3/R4, OI-20) ----  global default (test_id NULL) + optional per-test
CREATE TABLE report_settings (
    id                INTEGER PRIMARY KEY,
    test_id           INTEGER REFERENCES test(id) ON DELETE CASCADE,   -- NULL ⇒ global default
    header_logo       BLOB,                                            -- FR-R4 (image bytes)
    header_text       TEXT    NOT NULL DEFAULT '',
    footer_text       TEXT    NOT NULL DEFAULT '',
    page_size         TEXT    NOT NULL DEFAULT 'A4' CHECK (page_size IN ('A4','Letter')),       -- OI-20
    orientation       TEXT    NOT NULL DEFAULT 'portrait' CHECK (orientation IN ('portrait','landscape')),
    -- ordered blocks + visibility (FR-R3), e.g.
    -- [{"block":"cover","visible":false},{"block":"raw_table","visible":true},...]
    block_config_json TEXT    NOT NULL,
    commentary        TEXT    NOT NULL DEFAULT '',                      -- FR-R3 free-text evaluation
    updated_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_report_global ON report_settings(test_id) WHERE test_id IS NULL;     -- one global
CREATE UNIQUE INDEX uq_report_test   ON report_settings(test_id) WHERE test_id IS NOT NULL; -- one per test
```

**Cascade behaviour (OI-4), verified by FK graph:** `athlete → test → {step, threshold_result, zone, report_settings}` all carry `ON DELETE CASCADE`. A single `DELETE FROM athlete WHERE id=?` (with `foreign_keys=ON`) removes the entire subtree atomically. `template`/`training_profile` use `ON DELETE SET NULL` from `test`/`zone` so deleting a template or profile never deletes tests (FR-T8) — provenance is nulled, the snapshotted protocol/zone values stay intact.

---

## 2. Migration strategy — `PRAGMA user_version` migrator (cgo-free, zero-dep)

Embedded numbered SQL files; the migrator reads `PRAGMA user_version`, applies every file with index > current inside one transaction each, and bumps the version. This keeps the binary single-file, adds no dependency, and is fully cgo-free.

**Why not golang-migrate:** its default `database/sqlite3` driver pulls in `mattn/go-sqlite3` (**cgo — disqualified**). The pure-Go `database/sqlite` driver (built on `modernc.org/sqlite`) *is* cgo-free and is the sanctioned fallback, but it adds a dependency + a `schema_migrations` table for a problem that `user_version` solves in ~60 LOC. Pick the in-house migrator for v1; the `.sql` file convention stays compatible if we ever switch.

```
app/store/
  migrations/
    0001_init.sql
    0002_seed_predefined.sql   -- Appendix B templates + OI-17 profiles + global report_settings
  migrate.go
```

```go
package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies all embedded migrations with index > PRAGMA user_version.
func Migrate(db *sql.DB) error {
	var cur int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&cur); err != nil {
		return err
	}
	entries, _ := migrationsFS.ReadDir("migrations")
	type mig struct{ v int; name string }
	var migs []mig
	for _, e := range entries {
		v, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0]) // "0001_init.sql" -> 1
		if err != nil {
			return fmt.Errorf("bad migration name %q: %w", e.Name(), err)
		}
		migs = append(migs, mig{v, e.Name()})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].v < migs[j].v })

	for _, m := range migs {
		if m.v <= cur {
			continue
		}
		body, _ := migrationsFS.ReadFile("migrations/" + m.name)
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		// user_version can't be parameterized; m.v is our own int → safe to format.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.v)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
```

---

## 3. Repository + service (Wails-bound) layer

Two layers: **repositories** (narrow CRUD over `*sql.DB`/`*sql.Tx`, ctx-aware, no business logic) and a **service facade** (`App`) whose exported methods are the Wails bindings — they take/return JSON-friendly DTOs and `error` (Wails marshals the second return to a JS rejection). Analysis itself lives in `turnpoint-core` (out of scope here); the service calls it and persists the snapshot rows.

```go
// app/store/models.go (pointers = nullable/optional columns)
type Athlete struct {
	ID int64; Name string; DOB *string; Sex string
	BodyMassKg *float64; PrimarySport *string; Notes string
	CreatedAt, UpdatedAt string
}
type AthleteSummary struct { // FR-A4 / OI-5 list columns
	ID int64; Name string; PrimarySport *string; LastTestDate *string; TestCount int
}
type AthleteQuery struct { Search string; Limit, Offset int }

type Test struct {
	ID, AthleteID int64; TestDate, Sport string
	StepDurationS int; Increment, StartIntensity float64; Mode string; RestDurationS *int
	BaselineLactate, BodyMassSnapshot *float64; PretestNote, Remarks string; TemplateID *int64
}
type Step struct {
	ID, TestID int64; StepOrder int; Intensity float64; TimePointS *int
	HeartRate *int; Lactate *float64; RPE *int; IsBaseline, Excluded, Aborted bool
}
type ThresholdResult struct {
	ID, TestID int64; Method string; Intensity, Lactate, HeartRate, PctMax, PaceSPerKm, KcalH *float64
	IsOverride bool; FitType string; NotComputableReason *string; ParamsJSON *string
}
type Zone struct {
	ID, TestID int64; Model string; ZoneIndex int; ZoneName string; ProfileID *int64
	IntensityLow, IntensityHigh, HRLow, HRHigh, LactateLow, LactateHigh,
	PaceLowSPerKm, PaceHighSPerKm *float64
}
```

```go
// app/store/repos.go
type AthleteRepository interface {
	Create(ctx context.Context, a Athlete) (int64, error)
	Update(ctx context.Context, a Athlete) error
	Delete(ctx context.Context, id int64) error            // cascades (OI-4); FK=ON required
	Get(ctx context.Context, id int64) (Athlete, error)
	List(ctx context.Context, q AthleteQuery) ([]AthleteSummary, error)  // FR-A4
}
type TestRepository interface {
	Create(ctx context.Context, t Test) (int64, error)
	Update(ctx context.Context, t Test) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (Test, error)
	ListByAthlete(ctx context.Context, athleteID int64) ([]Test, error)
}
type StepRepository interface {
	ReplaceAll(ctx context.Context, testID int64, steps []Step) error // full-grid save in one tx
	ListByTest(ctx context.Context, testID int64) ([]Step, error)
	Upsert(ctx context.Context, s Step) (int64, error)
	Delete(ctx context.Context, id int64) error
}
type ThresholdResultRepository interface {
	ReplaceAll(ctx context.Context, testID int64, rows []ThresholdResult) error // recompute snapshot
	ListByTest(ctx context.Context, testID int64) ([]ThresholdResult, error)
	SetOverride(ctx context.Context, testID int64, method string, intensity float64) error // FR-Z3/C2
	ClearOverride(ctx context.Context, testID int64, method string) error
}
type ZoneRepository interface {
	ReplaceAll(ctx context.Context, testID int64, zones []Zone) error
	ListByTest(ctx context.Context, testID int64) ([]Zone, error)
}
type TemplateRepository interface {
	Create(ctx context.Context, t Template) (int64, error)
	Update(ctx context.Context, t Template) error // predefined are read-only → reject if is_predefined
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]Template, error)
}
type TrainingProfileRepository interface {
	Create(ctx context.Context, p TrainingProfile) (int64, error)
	Update(ctx context.Context, p TrainingProfile) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, sport string) ([]TrainingProfile, error)
}
type ReportSettingsRepository interface {
	GetForTest(ctx context.Context, testID int64) (ReportSettings, error) // falls back to global default
	UpsertForTest(ctx context.Context, rs ReportSettings) error
	GetGlobal(ctx context.Context) (ReportSettings, error)
	UpsertGlobal(ctx context.Context, rs ReportSettings) error
}
```

```go
// app/services/app.go — exported methods are the Wails bindings (JS-callable)
type App struct { db *sql.DB; ath store.AthleteRepository; /* ...other repos... */ }

func (a *App) CreateAthlete(in Athlete) (int64, error)
func (a *App) UpdateAthlete(in Athlete) error
func (a *App) DeleteAthlete(id int64) error                // UI gates with typed-confirm + export prompt
func (a *App) ListAthletes(search string) ([]AthleteSummary, error)

func (a *App) CreateTest(in Test) (int64, error)           // snapshots body_mass from athlete (OI-3)
func (a *App) SaveSteps(testID int64, steps []Step) error  // → StepRepository.ReplaceAll
func (a *App) Analyze(testID int64, cfg AnalysisConfig) (AnalysisResult, error) // calls core, persists TR+Zone
func (a *App) OverrideThreshold(testID int64, method string, intensity float64) (AnalysisResult, error)

func (a *App) ImportStepsCSV(testID int64, csv string, opt csvio.Options) (ImportReport, error) // FR-M2
func (a *App) ExportTestCSV(testID int64, opt csvio.Options) (string, error)                    // FR-R5
func (a *App) BackupDatabase(destPath string) error        // FR-M3
func (a *App) RestoreDatabase(srcPath string) error
```

**Concurrency:** `ReplaceAll`/`Analyze` run inside a single `BEGIN…COMMIT` so the grid + snapshot rows stay consistent. One process, one `*sql.DB`; WAL + `busy_timeout` (see §6) handle the rare reader/writer overlap.

---

## 4. CSV import/export (OI-21, Appendix A) + paste (OI-10)

**Canonical column order (Appendix A):** `intensity, time, hr, lactate, rpe`.
**Format defaults (OI-21):** UTF-8 (no BOM on write; strip BOM on read), header row, delimiter `,` (`;` selectable), decimal `.` (`,` selectable), time `mm:ss`. Hard rule: **if decimal = `,` then delimiter must be `;`** (otherwise commas are ambiguous) — the writer enforces this, the detector respects it.

Header is unit-implicit by the test's sport (running ⇒ km/h, cycling ⇒ W). Export writes lowercase canonical tokens; import is tolerant: case-insensitive, trims bracketed units (`Lactate [mmol/L]` → `lactate`), and accepts aliases (`speed|kmh|watt|power`→intensity, `bpm|hr|heartrate`→hr, `lac|lactate`→lactate, `borg|rpe`→rpe, `time|t`→time).

```go
// app/csvio/csv.go
package csvio

type Options struct {
	Delimiter rune // ',' or ';'   (0 ⇒ auto-detect on read)
	Decimal   rune // '.' or ','   (0 ⇒ auto-detect on read)
	Sport     string
}
type RowError struct { Line int; Col string; Value string; Msg string }
type ImportReport struct { Imported int; Skipped int; Errors []RowError }

// DetectDialect sniffs delimiter + decimal from a paste/file sample (OI-10).
func DetectDialect(sample []byte) Options

// ParseSteps reads canonical CSV/TSV → steps. Tolerant header mapping; per-row
// validation against OI-12 ranges; bad rows reported, not fatal (FR-T6).
func ParseSteps(r io.Reader, opt Options) (steps []store.Step, rep ImportReport, err error)

// WriteSteps emits canonical-order CSV for a test's grid (FR-M2/R5).
func WriteSteps(w io.Writer, steps []store.Step, opt Options) error

// --- app/csvio/time.go ---
func ParseMMSS(s string) (sec int, err error)  // "03:00"→180, "22:10"→1330; accepts "h:mm:ss" too
func FormatMMSS(sec int) string                // 1330→"22:10"

// --- numeric helpers ---
func parseDecimal(s string, dec rune) (float64, error) // dec==',' ⇒ swap comma→dot before ParseFloat
func formatDecimal(f float64, dec rune, places int) string
```

`ImportStepsCSV` (service) = `DetectDialect` (if opts zero) → `ParseSteps` → `StepRepository.ReplaceAll` inside a tx, returning the `ImportReport` so the UI can flag skipped/invalid rows (FR-T6). The same parse path backs clipboard paste into the grid (FR-T5/OI-10) — paste is just an in-memory `ParseSteps` with auto-detected dialect.

---

## 5. Backup / restore (FR-M3) — open-handle safe

**Backup = `VACUUM INTO`.** It is an online statement on the live connection that writes a fully consistent, defragmented single-file copy — safe while the app holds the DB open and correct under WAL (a raw `cp` of an open WAL DB can miss un-checkpointed `-wal` pages and is unsafe). No need to close the app.

```go
// app/backup/backup.go
func Backup(db *sql.DB, destPath string) error {
	// destPath must not exist; VACUUM INTO refuses to overwrite.
	_, err := db.Exec(`VACUUM INTO ?`, destPath) // single quoted literal also works
	return err
}
```

**Restore = controlled file replacement, not an online op.** Steps: (1) close the live `*sql.DB`; (2) integrity-check the incoming file by opening it read-only and running `PRAGMA integrity_check`; (3) move it over the primary DB path and delete any stale `-wal`/`-shm` sidecars; (4) reopen and run `Migrate` (forward-migrate an older backup). Because restore swaps the file, it must run when no other connection is open — the service quiesces all repo activity first.

```go
func Restore(dbPath, srcPath string, reopen func(path string) (*sql.DB, error)) (*sql.DB, error) {
	// caller has already Close()d the live *sql.DB
	if err := verifyIntegrity(srcPath); err != nil { return nil, err }
	if err := os.Rename(srcPath, dbPath); err != nil { return nil, err }      // same-volume atomic swap
	_ = os.Remove(dbPath + "-wal"); _ = os.Remove(dbPath + "-shm")            // drop stale sidecars
	db, err := reopen(dbPath); if err != nil { return nil, err }
	return db, Migrate(db)
}
```

---

## 6. `modernc.org/sqlite` — cgo-free, import path, minimal example

**Confirmed cgo-free.** `modernc.org/sqlite` is a pure-Go transpilation of SQLite; no cgo, no system SQLite, cross-compiles cleanly (the constraint is GOOS/GOARCH coverage, which includes win/amd64, darwin/amd64+arm64, linux/amd64+arm64 — all three SRS targets). Import path `modernc.org/sqlite`; it registers under driver name **`"sqlite"`** (note: *not* `"sqlite3"`).

```go
package store

import (
	"database/sql"
	_ "modernc.org/sqlite" // registers driver "sqlite"
)

// Open returns a configured pool. DSN pragmas apply to *every* pooled
// connection — critical for foreign_keys, which is per-connection and OFF by default.
func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path +
		"?_pragma=foreign_keys(1)" +   // enforce OI-4 cascades on all conns
		"&_pragma=busy_timeout(5000)" + // wait out transient locks
		"&_pragma=journal_mode(WAL)" +  // concurrent reads during writes
		"&_pragma=synchronous(NORMAL)"  // safe + fast under WAL for a desktop app
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite has a single writer; cap the pool to avoid SQLITE_BUSY storms.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil { // §2
		return nil, err
	}
	return db, nil
}
```

Minimal exec/query sanity (mirrors the upstream example, current API):

```go
db, _ := sql.Open("sqlite", "file:turnpoint.db?_pragma=foreign_keys(1)")
_, err := db.Exec(`INSERT INTO athlete(name, sex) VALUES (?, ?)`, "Bogner Markus", "male")
row := db.QueryRow(`SELECT id, name FROM athlete WHERE name = ?`, "Bogner Markus")
var id int64; var name string
err = row.Scan(&id, &name)
```

Notes that bite: (1) `_pragma=foreign_keys(1)` in the DSN is the reliable way to keep FKs on across the pool — setting it via a one-off `Exec` only affects one pooled connection. (2) Driver name is `"sqlite"`. (3) `SetMaxOpenConns(1)` (or a dedicated single-conn writer) plus `busy_timeout` is the standard desktop-app cure for write contention; reads stay concurrent under WAL if you later raise the cap with a separate read pool.

---

## 7. File layout (this scope)

```
app/
  store/
    db.go            # Open() + pragmas (§6)
    migrate.go       # user_version migrator (§2)
    migrations/0001_init.sql, 0002_seed_predefined.sql
    models.go        # DTO structs (§3)
    *_repo.go        # one file per repository
  csvio/
    csv.go, time.go  # §4
  backup/
    backup.go        # §5
  services/
    app.go           # Wails-bound facade (§3)
```

Relevant absolute paths for the plan author: `/Users/radaiko/dev/private/Turnpoint/docs/REQUIREMENTS.md` (source SRS); target implementation root `/Users/radaiko/dev/private/Turnpoint/app/` (store, csvio, backup, services packages as above). `turnpoint-core` (analysis) is a sibling package and out of this scope.

---

## Appendix D — Wails Shell, Frontend & Design System

Read the SRS and pulled current Wails v2 docs. Here is the design.

---

# Turnpoint — App Shell, Frontend Structure & Design System

## 1. Wails v2 project layout

**Single Go module, two packages.** Keep `core` as a normal exported package inside the app module (not `internal/`, so it stays independently `go test`-able and publishable later). This avoids multi-module/`go.work` friction that complicates `wails build` and cross-compilation — split it into its own module only if/when it's reused elsewhere.

```
turnpoint/                     # repo root = Wails app, module github.com/aiko/turnpoint
├── go.mod  go.sum
├── wails.json                 # build config (wailsjsdir, frontend cmds)
├── main.go                    # wails.Run(&options.App{...})
├── app.go                     # App struct: bound methods (thin: orchestrates core + db)
├── build/                     # appicon, darwin/ windows/ platform assets (generated)
├── core/                      # package core — pure domain, NO wails import
│   ├── fit/                   # gonum curve fitting (poly3, exp, spline)
│   ├── threshold/             # §6 methods
│   ├── zones/                 # §7 5-zone derivation
│   ├── dto.go                 # AnalysisInput / AnalysisResult structs (surface to TS)
│   └── *_test.go              # lactater parity fixtures
├── db/                        # modernc.org/sqlite: migrations, repositories
└── frontend/
    ├── index.html  package.json  tsconfig.json
    ├── vite.config.ts  svelte.config.js
    ├── wailsjs/               # GENERATED — do not edit, git-ignored or committed
    │   ├── go/main/App.{js,d.ts}     # one async fn per bound App method
    │   ├── go/core/models.ts         # TS mirror of core.* structs (namespace `core`)
    │   └── runtime/                  # EventsOn/Emit, window, dialog
    └── src/
        ├── main.ts  App.svelte  app.css
        ├── lib/{api,stores,components,charts,design,format}/
        └── views/{Athletes,TestEntry,Analysis,Zones,Comparison,Report}.svelte
```

**Scaffold:** `wails init -n turnpoint -t svelte-ts` (plain Svelte + Vite + TS SPA — **not** SvelteKit; SvelteKit's SSR/adapter layer is dead weight in a webview and fights the embedded-asset model). Then move generated `app.go`/`main.go` logic into the layout above and add `core/`, `db/`.

**Bindings:** generated into `frontend/wailsjs/` (set `"wailsjsdir": "./frontend/wailsjs"` in `wails.json`). Wails emits one **Promise-returning** JS function per exported `App` method, plus `models.ts` TS types for every Go struct used as a param/return, **namespaced by the Go package that defines it** (so return `core.AnalysisResult` → `core` namespace in TS). Regenerate without launching via `wails generate module`; `wails dev` regenerates on save.

**main.go / App binding (confirmed API):**
```go
//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()                       // holds *sql.DB, ctx
    wails.Run(&options.App{
        Title: "Turnpoint", Width: 1280, Height: 840, MinWidth: 1024, MinHeight: 720,
        AssetServer: &assetserver.Options{Assets: assets},
        Frameless:        true,            // custom titlebar (see shell)
        CSSDragProperty:  "--wails-draggable", CSSDragValue: "drag",
        BackgroundColour: &options.RGBA{R: 14, G: 16, B: 20, A: 255},
        Mac: &mac.Options{TitleBar: &mac.TitleBar{TitlebarAppearsTransparent: true,
            HideTitle: true, FullSizeContent: true}},
        OnStartup: app.startup,            // func(ctx context.Context) — stash ctx for runtime.*
        Bind:      []interface{}{app},
    })
}
```
Bound method signature pattern (request/response, not events): `func (a *App) Analyze(in core.AnalysisInput) (core.AnalysisResult, error)`. Use `runtime.EventsEmit(ctx, ...)` only for Go→JS push (e.g. autosave-done toast); all compute is await-a-Promise.

**Build commands:**
- Dev: `wails dev` (Vite HMR for Svelte + Go rebuild + live bindings + devtools).
- Release: `wails build -clean` → single binary in `build/bin/`. modernc.org/sqlite is pure-Go, so `CGO_ENABLED=0` and `-platform windows/amd64|darwin/universal|linux/amd64` cross-compile cleanly. Add `-upx` if footprint (NFR-8 <50 MB) is tight.
- Frontend cmds in `wails.json`: `"frontend:install": "npm install"`, `"frontend:build": "npm run build"`.

**Vite aliases** (set in `vite.config.ts` + `tsconfig.json` paths): `$wails → frontend/wailsjs`, `$lib → src/lib`.

---

## 2. Svelte app structure

**Routing:** no router library. Views are master-detail stages tied to *selection state*, not URLs (a desktop app needs no deep-linking). A `ui` store holds `{ activeSection, activeAthleteId, activeTestId, activeStage }`; `App.svelte` switches on it. Saves ~5–15 KB and keeps nav state in one place.

**Shell (sidebar, two-tier master-detail):**
```
┌──────────────────────────────────────────────────────────┐
│  ⠿ Turnpoint            [titlebar — --wails-draggable]  ◑ │  44px
├───────────┬──────────────────────────────────────────────┤
│ NAV RAIL  │  WORKSPACE                                    │
│ 240px     │  ┌ Entry · Analysis · Zones · Report ┐ (tabs) │
│ Athletes  │  │  active stage view                  │      │
│ Compare   │  │                                     │      │
│ ───       │  └─────────────────────────────────────┘      │
│ Settings  │                                               │
└───────────┴──────────────────────────────────────────────┘
```
Left **nav rail** = primary destinations (Athletes, Comparison, Settings pinned bottom). Selecting an athlete → its tests; opening a test enters the **workspace**, whose stages (Test Entry → Analysis/Charts → Zones → Report) are a horizontal **segmented tab bar**, because they're sequential stages of one test, not peer destinations. Frameless window + custom titlebar (theme toggle, app mark) for the native-minimal look; OS draws window controls.

**Views (`src/views/`):**
| View | Drives | Key FRs |
|---|---|---|
| `Athletes.svelte` | list/search + CRUD form (master-detail split) | FR-A1–A5 |
| `TestEntry.svelte` | protocol header + `DataGrid` | FR-T1–T8 |
| `Analysis.svelte` | `FitChart` + threshold results table + method toggles | FR-C1–C3, FR-D* |
| `Zones.svelte` | `TemporalChart` + zone table + LT1/LT2 anchor controls | FR-C4, FR-Z* |
| `Comparison.svelte` | longitudinal time-series + overlaid curves | FR-H1–H3 |
| `Report.svelte` | block list (include/omit/reorder) + print-styled preview | FR-R* |

**Shared components (`lib/components/`):** `Button`, `IconButton`, `Field`/`NumberField`/`Select`, `Tabs`, `Table`, `Modal`, `Toast`, `Toggle`, `Tag` (algorithmic/manual badge), `EmptyState`. **`lib/charts/`:** `FitChart`, `TemporalChart`, plus primitives (`Axis`, `ZoneBands`, `DraggableMarker`, `StepBars`). **`lib/format/`:** `mmss`, `pacePer1000`, `tabularNum`.

**Stores & data flow (`lib/stores/`):** Go is the single source of truth for *computed* results; Svelte stores cache them and re-request on input change.

- `athletes.ts` — list + CRUD, calls `App.*`.
- `session.ts` — **writable**: current test + editable steps + dirty flag; debounced autosave → `App.SaveTest`.
- `analysis.ts` — **async-derived**: fitted curve + thresholds + zones. Can't be a synchronous `derived` (compute lives in Go), so use a debounced, latest-wins recompute:

```ts
// lib/stores/analysis.ts
import { writable } from 'svelte/store';
import { Analyze } from '$wails/go/main/App';   // generated Promise binding
import type { core } from '$wails/go/core/models';
import { session } from './session';

function createAnalysis() {
  const { subscribe, set } = writable<core.AnalysisResult | null>(null);
  let seq = 0, timer: ReturnType<typeof setTimeout>;
  function recompute(input: core.AnalysisInput, debounceMs = 40) {
    clearTimeout(timer);
    timer = setTimeout(async () => {
      const id = ++seq;
      const res = await Analyze(input);           // in-process call, sub-ms overhead
      if (id === seq) set(res);                    // latest-wins → drop stale drags
    }, debounceMs);
  }
  return { subscribe, recompute };
}
export const analysis = createAnalysis();
session.subscribe(s => analysis.recompute(s.toAnalysisInput()));
```
- Drag (FR-C2, NFR-3 <100 ms): call `analysis.recompute(input, 0)` on each `pointermove` (latest-wins already guards stale responses); Wails binding overhead is negligible, the cost is the Go fit, which is small. No optimistic JS math needed.
- `ui.ts` (nav/stage/theme), `config.ts` (enabled methods, LT1/LT2 anchor mapping, training profile — persisted via `App.SaveConfig`).

`lib/api/` = thin typed wrappers over `$wails/go/main/App` so views never import generated paths directly (one seam to mock in tests).

---

## 3. Design system — "Instrument minimal"

Direction: a **measured, instrument-panel** aesthetic — flat surfaces, hairline separation over shadows, generous whitespace, monospaced numerics everywhere data appears, one restrained teal accent. Reads as a precision tool, not a Bootstrap dashboard.

### Typography
**Pairing: Geist Sans (UI) + Geist Mono (all numerics).** Both SIL OFL — **self-host as woff2** bundled via Vite `@font-face`; **never** a Google Fonts `<link>` (NFR-2/M4 offline). Geist is a contemporary neutral grotesque that avoids the Inter/system-default templated feel; its mono gives lab-readout character to tables, the grid, and chart axes. Fallback stack: `-apple-system, "Segoe UI", Roboto, sans-serif`.

Numerics are load-bearing: apply `font-variant-numeric: tabular-nums slashed-zero;` to every grid cell, results table, zone table, and axis label so digits align in columns.

Scale (desktop-dense 13px base), weights 400/500/600 only (no 700+ — keeps it quiet):
```
--fs-eyebrow: 11px / 16  600  +0.06em uppercase   (section micro-labels)
--fs-caption: 12px / 16  400
--fs-body:    13px / 20  400
--fs-label:   13px / 20  500
--fs-h3:      15px / 22  600
--fs-h2:      18px / 26  600  -0.005em
--fs-h1:      22px / 30  600  -0.01em
--fs-display: 28px / 36  600  -0.015em          (report headers)
--font-mono: "Geist Mono", ui-monospace, monospace;
```

### Color tokens (`lib/design/tokens.css`)
Cool near-neutrals (slight blue), single deep-teal accent. Theme via `[data-theme="dark"]` on `<html>`.

**Light**
```
--bg:#F5F6F8  --surface:#FFFFFF  --surface-2:#EEF0F3  --inset:#E9ECF0
--border:#DEE1E6  --border-strong:#C7CCD3
--text:#16181D  --text-muted:#5A616B  --text-faint:#8A919C
--accent:#15616D  --accent-hover:#114E58  --accent-contrast:#FFFFFF
--focus:rgba(21,97,109,.40)
--danger:#B42318 --warn:#B54708 --ok:#207A4C
```
**Dark**
```
--bg:#0E1014  --surface:#16191F  --surface-2:#1E222A  --inset:#11141A
--border:#272C35  --border-strong:#353B45
--text:#E6E8EC  --text-muted:#9AA2AD  --text-faint:#6B727C
--accent:#4FB3C0  --accent-hover:#6AC5D1  --accent-contrast:#08191C
--focus:rgba(79,179,192,.45)
--danger:#F0795C --warn:#E5A23B --ok:#5DBE8A
```

**Data-viz palette** (shared, desaturated so it never looks garish — fills at ~14% alpha, edges/series at full):
```
Zone bands (REKOM→SB, recovery→peak intensity gradient):
  REKOM #5B8DB8 · GA1 #3E9C8F · GA2 #A6B24D · EB #D99A2B · SB #C2543F
Series:  lactate #15616D (accent)  ·  heart-rate #C2693A (burnt orange)
         intensity step-bars: --border-strong fill @ 55% (neutral, recedes)
Threshold lines: algorithmic = --text solid 1px; manual override = --accent dashed (FR-C3)
```

### Spacing / radius / elevation
```
--space: 4 8 12 16 24 32 48 64               (4px base; --space-1..8)
--radius-sm:4  --radius-md:6  --radius-lg:10  --radius-pill:9999
```
Small radii (6px default) read intentional, not bubbly. **Separation = 1px hairline borders, not shadows.** Only two shadows, reserved for floating layers:
```
--shadow-pop:   0 1px 2px rgba(16,24,40,.04), 0 4px 12px rgba(16,24,40,.10)   (menus/popovers)
--shadow-modal: 0 8px 32px rgba(16,24,40,.18)
```
Dark theme leans on `--surface-2` elevation + borders instead of shadow. Focus = 2px `--accent` ring via `box-shadow: 0 0 0 3px var(--focus)` (keyboard-visible, NFR-7). Default 150ms ease transitions; honor `prefers-reduced-motion`.

---

## 4. Data-entry grid (FR-T5) — **build custom**

Build a ~300-line `DataGrid.svelte`, do **not** pull a library. The schema is tiny and bespoke (5 fixed columns, sport-dependent unit header `[km/h]`/`[W]`, `mm:ss` mask, one baseline row, per-cell validation flags from OI-12, exclude-from-fit toggle), and the heavyweight grids are wrong on every axis: ag-Grid / Handsontable are 200 KB–1 MB+ (busts NFR-8), look templated, and carry license terms; canvas grids (glide) and RevoGrid are React/web-component-first and fight Svelte reactivity. A hand-rolled grid binds straight to the `session` store and styles to the design system.

Contract to implement:
- **Keyboard nav:** Arrow/Tab/Shift-Tab move cells; Enter commits + moves down; typing replaces; Esc reverts cell. Roving `tabindex`, single focused `{row,col}` in component state.
- **Rows:** add (Enter on last row / "+ row"), delete (per-row control + `Cmd/Ctrl+Backspace`), reorder (drag handle + `Alt+↑/↓`). Baseline row pinned at top, visually distinct.
- **Paste (OI-10):** intercept `paste`, read `clipboardData`, **auto-detect delimiter** (tab/comma/semicolon) and **decimal separator** (dot/comma), map onto columns from the focused cell, grow rows as needed. Single `parseTabular(text): Cell[][]` in `lib/format/`.
- **Validation:** range-check per column (OI-12) → non-blocking inline flag (`--warn` underline + tooltip); empty-lactate row tagged "excluded from fit" (FR-T2). Edits flow to `session` → `analysis.recompute` fires automatically (FR-F4).

---

## 5. Charts (FR-C1, FR-C4) — **custom SVG on LayerCake + d3-scale/d3-shape**

Recommendation: render both charts as **declarative Svelte SVG**, using **LayerCake** (Svelte-native chart scaffold: handles dimensions, scales, layered SVG/HTML) plus only the modular **`d3-scale`** (`scaleLinear`) and **`d3-shape`** (`line`, `area`, `curveNatural`) — a few KB total, not full `d3`.

Why not the alternatives: a charting lib (Chart.js/uPlot/ECharts) makes **draggable threshold markers** (FR-C2) and the bespoke **dual-axis + zone-bands + step-bars + vertical IAS/IANS lines** (FR-C4) a fight against the lib's model, and pulls 50–300 KB (NFR-8). Full `d3` is imperative and duplicates Svelte's DOM/reactivity. Custom SVG is the opposite: every layer is a Svelte component reading the `analysis` store, a marker drag is just `pointerdown/move` writing back to the store, and print/SVG export (FR-R5) is the DOM you already rendered. This is also what makes the charts look distinctive rather than templated.

**`FitChart.svelte`** (FR-C1/C2/C3): layers = `ZoneBands` (`<rect>`) → fitted curve (`<path>` from `d3-shape.line().curve(curveNatural)`) → raw points (`<circle>`) → HR overlay → `DraggableMarker × n`. `DraggableMarker` emits `on:drag={x → analysis.recompute(withAnchor(x), 0)}`; algorithmic vs manual styling per the threshold-line tokens.

**`TemporalChart.svelte`** (FR-C4): X = time; **two y-scales** (`yLactate` left, `yHR` right) — LayerCake's single `yScale` plus a second locally-derived `scaleLinear` for HR; `StepBars` (`<rect>` per step, intensity) behind `ZoneBands`; lactate + HR `<path>` series; IAS/IANS vertical `<line>`s; user-configurable reference lines (4 mmol/L, threshold HR/speed). Print-quality vector export = serialize the SVG node (SVG) or rasterize via canvas (PNG) for FR-R5; the same SVG reused in `Report.svelte` under an `@page A4` / `@media print` stylesheet.

---

### Decisions summary
- **Scaffold:** `wails init -t svelte-ts`, plain Svelte+Vite+TS SPA, single Go module with exported `core/` package.
- **Bindings:** `frontend/wailsjs/` Promise functions + per-Go-package `models.ts`; aliased `$wails`.
- **Shell:** frameless window, 240px nav rail (Athletes/Compare/Settings) + segmented stage tabs in the test workspace.
- **State:** Go = compute source of truth; debounced latest-wins async store re-requests on every input/drag.
- **Type:** Geist Sans + Geist Mono (self-hosted woff2), tabular numerics everywhere.
- **Color:** cool neutrals + single deep-teal accent; desaturated 5-zone + 2-series data palette; hairline borders over shadow.
- **Grid:** build custom (footprint, bespoke schema, paste auto-detect).
- **Charts:** custom SVG via LayerCake + d3-scale/d3-shape (draggable markers, dual-axis, small footprint).

Relevant file: `/Users/radaiko/dev/private/Turnpoint/docs/REQUIREMENTS.md`.

---

## Appendix E — Reporting / PDF

# Reporting / PDF Generation + Chart & Data Export — Recommendation

## 1. Decision: PDF report engine

**Primary — HTML report route in the Wails webview + browser print-to-PDF (CSS Paged Media + `window.print()`).**
**Fallback — server-side deterministic Go PDF via `maroto/v2` (gofpdf), used for headless/batch export.**

Justification (one line each):
- **Vector chart:** the temporal chart is already a Svelte **SVG**; the print path embeds it verbatim as true vector with **zero conversion and guaranteed screen parity** — the single strongest reason. A Go engine cannot embed that SVG as vector without a re-render or PDF-import detour (see §3).
- **Editable/reorderable blocks (FR-R3):** the editor *is* the live Svelte UI; reorder/omit/edit are plain DOM/state changes — no second layout engine to keep in sync.
- **Pagination + header/footer (FR-R4):** handled by CSS `@page` + running fixed header/footer + `break-*` rules; the webview's print engine paginates without clipping.
- **Default = reference pages 2 & 3:** trivially expressed as the default block list (§4), rendered as two A4 sections.
- **FR-R1 is satisfied directly:** `window.print()` opens the OS print dialog, which on all three targets offers "Save as PDF" (Windows WebView2 → *Microsoft Print to PDF*; macOS WKWebView → *PDF ▸ Save as PDF*; Linux WebKitGTK → *Print to File (PDF)*). No PDF library needed for the happy path.

Why a fallback exists: the print path has **no Wails-exposed silent/headless export** (no `runtime.PrintToPDF`). Programmatic file output needs native webview calls — WebView2 `ICoreWebView2.PrintToPdf`, WKWebView `createPDF`, WebKitGTK `webkit_print_operation_*` — none surfaced by the Wails v2 runtime. When deterministic, dialog-free, cross-platform-identical, or batch export is required, generate server-side with maroto instead. (Don't go maroto-primary: its image component is **raster-only**, so the chart would lose vector quality — exactly the requirement we must protect.)

## 2. How the temporal chart reaches the report at vector quality

**Reuse the Svelte SVG — do not re-draw it in Go.** Author the chart once (the FR-C4 component: dual Y-axes, zone bands, step bars, IAS/IANS lines). Two consumers of that one SVG:

1. **Primary (HTML print):** the same `<svg>` is in the report-route DOM → printed as vector. Nothing to convert.
2. **Fallback PDF + image exports:** serialize the live SVG (`svgEl.outerHTML`), hand the string to Go, and parse it with **`github.com/tdewolff/canvas`** (a vector-graphics lib that *parses SVG* and renders to vector PDF / SVG / raster). This is the shared chart-export engine:

```go
import (
    "github.com/tdewolff/canvas"
    "github.com/tdewolff/canvas/renderers"
)
c, err := canvas.ParseSVG(strings.NewReader(svgMarkup)) // *canvas.Canvas
// vector PDF (for fallback report / standalone chart)
err = c.WriteFile("chart.pdf", renderers.PDF())
// raster PNG at chosen DPI (FR-R5)
err = c.WriteFile("chart.png", renderers.PNG(canvas.DPI(300)))
// normalized SVG (FR-R5) — or just re-emit the source markup
err = c.WriteFile("chart.svg", renderers.SVG())
```

For the **maroto** fallback specifically, embed the chart as a **300-dpi PNG** (maroto images are raster). If the fallback must also be vector, bypass maroto for the chart page only and import the `chart.pdf` produced above with **`github.com/go-pdf/fpdf`** + **`github.com/go-pdf/fpdf/contrib/gofpdi`** (`ImportPage`/`UseImportedTemplate`). (`gonum.org/v1/plot` + `gonum.org/v1/plot/vg/vgpdf` is a third option — vector PDF straight from Go — but only worth it if you ever re-implement the chart in Go; it risks divergence from the on-screen chart, so it is *not* recommended here.)

## 3. Report block model (include / omit / reorder / edit)

Persist as an ordered list (array index = order). Stored as JSON on `ReportSettings` (matches §9: "block order/visibility, commentary text").

```go
type BlockType string
const (
    BlockCover         BlockType = "cover_anamnesis"   // page 1, off by default
    BlockRemarks       BlockType = "test_remarks"
    BlockRawSteps      BlockType = "raw_step_table"     // page 2
    BlockTemporalChart BlockType = "temporal_chart"     // page 2, vector
    BlockThresholds    BlockType = "threshold_results"  // page 3 (FR-D5)
    BlockZones         BlockType = "training_zones"     // page 3 (§7)
    BlockEvaluation    BlockType = "evaluation"         // page 4, off by default
)

type ReportBlock struct {
    ID              string         `json:"id"`
    Type            BlockType      `json:"type"`
    Included        bool           `json:"included"`        // omit = false
    PageBreakBefore bool           `json:"pageBreakBefore"` // start a new page
    Title           string         `json:"title,omitempty"` // editable heading
    Body            string         `json:"body,omitempty"`  // editable free-text (remarks/evaluation/commentary)
    Options         map[string]any `json:"options,omitempty"` // chart layers, ref lines, visible columns
}
```

- **Reorder** = reorder the slice (drag in UI). **Omit** = `Included=false`. **Edit** = `Title`/`Body` (free-text per FR-R3) and `Options` (e.g. which chart layers/ref lines, which step columns).
- **Default block set** (FR-R3: reference pages 2 & 3 only):

```go
Default = []ReportBlock{
  {Type: BlockCover,         Included: false},
  {Type: BlockRemarks,       Included: false},
  {Type: BlockRawSteps,      Included: true,  PageBreakBefore: true}, // page 2
  {Type: BlockTemporalChart, Included: true},                        // page 2
  {Type: BlockThresholds,    Included: true,  PageBreakBefore: true}, // page 3
  {Type: BlockZones,         Included: true},                        // page 3
  {Type: BlockEvaluation,    Included: false},                       // page 4
}
```

## 4. Persistence (DDL)

```sql
CREATE TABLE report_settings (
  id                INTEGER PRIMARY KEY,
  test_id           INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE,
  page_size         TEXT    NOT NULL DEFAULT 'A4',        -- 'A4' | 'Letter'  (OI-20)
  orientation       TEXT    NOT NULL DEFAULT 'portrait',  -- A4 portrait default (OI-20)
  header_logo_path  TEXT,                                  -- user logo (PNG/SVG)
  header_text       TEXT,
  footer_text       TEXT,
  show_page_numbers INTEGER NOT NULL DEFAULT 1,
  blocks_json       TEXT    NOT NULL,                      -- ordered []ReportBlock
  updated_at        TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX ux_report_settings_test ON report_settings(test_id);
```

(A normalized `report_block` table is an alternative, but JSON matches the §9 model and keeps reorder/edit atomic.)

## 5. Page setup, header/logo, footer, pagination (FR-R4, OI-20)

**Default A4 portrait; Letter selectable.**

- **HTML/print (primary):**
  ```css
  @page { size: A4 portrait; margin: 18mm 16mm 16mm 16mm; }      /* Letter swaps size */
  .report-page  { break-before: page; }
  table         { break-inside: auto; }  tr { break-inside: avoid; }
  .no-split     { break-inside: avoid; } /* chart + each table block */
  header.running{ position: fixed; top: 0; }   /* logo <img src=…> + header_text */
  footer.running{ position: fixed; bottom: 0; content: footer_text " " counter(page); }
  ```
  Logo + footer repeat on every printed page; page numbers via the CSS `page` counter. Trigger from the toolbar with `window.print()`; Go can request it via `runtime.EventsEmit(ctx,"report:print")`.
- **maroto (fallback):** `cfg := config.NewBuilder().WithPageNumber().Build()` (A4 portrait is maroto's default; `WithPageSize(pagesize.Letter)` to switch). `m.RegisterHeader(row.New(20).Add(image.NewFromFileCol(3, logo, props.Rect{...}), text.NewCol(9, headerText, ...)))` and `m.RegisterFooter(row.New(12).Add(text.NewCol(12, footerText, ...)))`. Body rows auto-paginate; `m.Generate()` → `core.Document` → `.Save(path)` or `.GetBytes()`.

## 6. Chart export PNG/SVG + CSV export (FR-R5)

- **SVG export:** serialize the live chart (`svgEl.outerHTML`) → write bytes. (Optionally normalize through `renderers.SVG()`.) No raster step → stays vector.
- **PNG export:** `canvas.ParseSVG(svg)` → `c.WriteFile(path, renderers.PNG(canvas.DPI(300)))` for DPI control; or a pure-frontend `<canvas>.toBlob()` if DPI control isn't needed. Prefer the Go path to reuse one pipeline and get deterministic DPI.
- **Results & zones CSV:** pure Go `encoding/csv`. Two writers — **results** (markers 2/4/6 mmol·L⁻¹, IAS, IANS, MAX × {km/h|W, lactate, HR, %max, pace/km, kcal/h}) and **zones** (zone × {intensity range, lactate range, HR range, pace range}) — matching the on-screen tables. Honor **OI-21**: UTF-8, header row, comma delimiter (semicolon selectable), dot decimal (comma selectable), `mm:ss` time.
- **Saving any export:** `path, _ := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{DefaultFilename:"report.pdf", Filters:[]runtime.FileFilter{{DisplayName:"PDF",Pattern:"*.pdf"}}}); os.WriteFile(path, data, 0644)` (import `github.com/wailsapp/wails/v2/pkg/runtime`).

## 7. Exact libraries / APIs to pin

| Purpose | Import path | Key calls |
|---|---|---|
| Print path trigger / save dialog | `github.com/wailsapp/wails/v2/pkg/runtime` | `SaveFileDialog`, `EventsEmit` |
| Save-as-PDF (primary) | webview print engine via JS | `window.print()` (OS dialog → Save as PDF) |
| SVG → vector PDF / PNG / SVG (shared chart engine) | `github.com/tdewolff/canvas`, `…/canvas/renderers` | `canvas.ParseSVG`, `renderers.PDF/PNG/SVG`, `c.WriteFile` |
| Server-side PDF (fallback) | `github.com/johnfercher/maroto/v2` (+ `/pkg/config`,`/pkg/components/{row,col,text,image,line}`,`/pkg/consts`,`/pkg/props`,`/pkg/core`) | `maroto.New(cfg)`, `RegisterHeader/Footer`, `AddRow/AddRows/AddAutoRow`, `Generate().Save/GetBytes` |
| Vector chart inside fallback PDF (optional) | `github.com/go-pdf/fpdf` + `…/contrib/gofpdi` | `ImportPage`, `UseImportedTemplate` |
| Go-rendered vector chart (only if re-implemented) | `gonum.org/v1/plot` + `gonum.org/v1/plot/vg/vgpdf` | `plot.New`, `p.Save("…​.pdf")` |
| CSV | stdlib `encoding/csv`, `os` | `csv.NewWriter`, `os.WriteFile` |

**Risk to flag for the plan author:** Wails v2 exposes no programmatic print-to-PDF; FR-R1's "export to a PDF file" rides on the OS dialog's Save-as-PDF in the primary path. If silent/batch PDF is later mandated, either wire the native webview PDF API per platform or promote the maroto fallback — both already specified above.

Sources:
- [Wails Dialog runtime (SaveFileDialog)](https://wails.io/docs/reference/runtime/dialog/)
- [WebView2 PrintToPdf spec](https://github.com/MicrosoftEdge/WebView2Feedback/blob/main/specs/PrintToPdf.md) · [Printing from WebView2](https://learn.microsoft.com/en-us/microsoft-edge/webview2/how-to/print)
- [tdewolff/canvas (SVG in → PDF/PNG/SVG out)](https://github.com/tdewolff/canvas) · [renderers godoc](https://pkg.go.dev/github.com/tdewolff/canvas/renderers)
- [maroto v2 docs (header/footer, images, page number)](https://github.com/johnfercher/maroto)
- [gonum vgpdf (vector PDF backend)](https://pkg.go.dev/gonum.org/v1/plot/vg/vgpdf)

---

## Appendix F — Validation & Test Strategy

R is absent (premise confirmed), Go 1.26.3 present, repo is bare scaffolding. All library APIs verified against current docs. Here is the validation & test strategy.

---

# Turnpoint P0 — Validation & Test Strategy (no local `lactater`)

## 0. Framing decision (load-bearing)
**WinLactat (Appendix C) ≠ `lactater`.** The SRS names `lactater` as the parity benchmark (V1, NFR-6, OI-1) but every concrete number we actually possess (Appendix C) is a **WinLactat / Lactate EXPRESS** output, computed on its own curve fit. They will not be bit-identical for curve-shape-dependent methods (Dmax variants, LTP, log-log). Therefore:
- **Use WinLactat (Appendix C) as the primary frozen ground truth now** for the methods it exposes (fixed OBLA, IAS/IANS, MAX, zones). These are well-defined and fit-reproducible.
- **Treat `lactater` parity as a separate, data-driven golden layer** generated out-of-band when R exists, covering the curve-only methods WinLactat's report doesn't enumerate. CI never needs R.

This split is reflected in the directory layout (§2) and the install recommendation (§6).

---

## 1. Primary golden fixtures: exact assertions + OI-1 tolerances

**Tolerances (OI-1), encoded once** in `turnpoint-core/internal/testutil/floatcmp.go`:

| Metric | Rule | gonum call |
|---|---|---|
| intensity (run) | ≤ 0.1 km/h **or** ≤ 1 % rel | `scalar.EqualWithinAbsOrRel(got,want,0.1,0.01)` |
| intensity (cycle) | ≤ 2 W **or** ≤ 1 % rel | `scalar.EqualWithinAbsOrRel(got,want,2,0.01)` |
| heart rate | ≤ 1 bpm | `scalar.EqualWithinAbs(got,want,1)` |
| lactate | ≤ 0.05 mmol/L | `scalar.EqualWithinAbs(got,want,0.05)` |

`EqualWithinAbsOrRel(a,b,absTol,relTol)` returns true if **either** envelope holds → exactly OI-1's "whichever is larger". Package path confirmed current: `gonum.org/v1/gonum/floats/scalar`.

**Fixture 1 — threshold markers** (`testdata/golden/winlactat/appendix_c_markers.json`), 3rd-order polynomial on Appendix A. Assert per row:

| Marker | km/h (±0.1/1%) | Lactate (±0.05) | HR (±1) | %max | Pace/km |
|---|---|---|---|---|---|
| OBLA 2.0 | 13.1 | 2.0 | 140 | 65.5 | 04:34 |
| OBLA 4.0 | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| OBLA 6.0 | 17.6 | 6.0 | 177 | 87.8 | 03:25 |
| IAS | 10.5 | 1.4 | 122 | 52.7 | 05:41 |
| IANS | 16.1 | 4.0 | 167 | 80.5 | 03:43 |
| MAX | 20.0 | 7.7 | 185 | 100.0 | 03:00 |

**Derived-metric rounding caveat (important):** Appendix C's displayed km/h is rounded to 1 dp, but `%max` and `pace` are computed from the **full-precision** intensity (e.g. OBLA 6.0 displays 17.6 but %max 87.8 ⇒ internal ≈ 17.56). So:
- Assert **intensity** against the displayed value within ±0.1 km/h.
- Compute **%max = intensity/maxIntensity·100** and **pace = 60/kmh** from the engine's *unrounded* intensity, and assert `%max` within **±0.3** and `pace` within **±1 s** (parse `mm:ss`→seconds) — this absorbs WinLactat's display-rounding ambiguity rather than demanding exact string equality (its sec rounding is inconsistent: 13.1→34.8 s floors to 04:34, 17.6→24.5 s rounds to 03:25).

---

## 2. `lactater` golden encoding + `testdata/` layout

Three golden layers, namespaced by **source** so provenance and update policy are explicit:

```
turnpoint-core/                       # go.mod: module turnpoint-core
  fit/  threshold/  zone/
  internal/testutil/
    floatcmp.go        # OI-1 tolerance consts + Equal* wrappers (gonum scalar)
    golden.go          # load JSON, -update flag, assertMarkers/assertZones
  testdata/
    datasets/
      appendix_a.json              # canonical input (8 steps + baseline)
      lactater_demo.json           # snapshot of lactater::demo_data() (added at regen)
    golden/
      winlactat/                   # GROUND TRUTH, hand-encoded, NEVER auto-updated
        appendix_c_markers.json
        appendix_c_zones.json
      lactater/                    # regenerated out-of-band by tools/regen-lactater
        appendix_a.json
        demo_data.json
      turnpoint/                   # OUR engine snapshots, refreshed by `go test -update`
        appendix_a_methods.json    # regression guard (catches accidental drift)
  tools/regen-lactater/
    main.go            # shells Rscript, writes golden/lactater/*.json
    lactater_export.R  # calls lactate_threshold(); records pkg commit SHA
```

**Golden file format** — stable, diffable JSON keyed by `method`+`fitting` so a regenerated `lactater` file maps 1:1 onto the tibble it returns (`method_category, method, fitting, intensity, lactate, heart_rate`):

```json
{
  "source": "lactater",
  "dataset": "appendix_a",
  "fit": "3rd degree polynomial",
  "tool_version": "fmmattioni/lactater@<commit-sha>",
  "results": [
    {"method":"Log-log","fitting":"3rd degree polynomial","intensity":10.5,"lactate":1.4,"heart_rate":122},
    {"method":"Dmax","fitting":"3rd degree polynomial","intensity":16.0,"lactate":3.8,"heart_rate":166},
    {"method":"ModDmax","fitting":"3rd degree polynomial","intensity":15.8,"lactate":3.6,"heart_rate":164}
  ]
}
```

The `regen-lactater` R step (run only where R exists), confirmed against current docs:
```r
lactate_threshold(.data, intensity_column, lactate_column, heart_rate_column,
  method = c("Log-log","OBLA","Bsln+","Dmax","LTP","LTratio"),
  fit = "3rd degree polynomial", include_baseline = TRUE, sport = "running")
# returns tibble: method_category, method, fitting, intensity, lactate, heart_rate
```
The Go `main.go` writes the tibble (minus the `plot` list-column) as the JSON above and stamps the package commit SHA into `tool_version` for reproducibility. **`go test -update` must not touch `winlactat/` or `lactater/`** — only `turnpoint/` snapshots are auto-refreshable.

---

## 3. Go testing approach

- **Layout:** test code lives beside the package (`threshold/parity_test.go`, `fit/edge_test.go`, `zone/zone_test.go`). Input/golden loaded from `testdata/` (Go's `go test` sets CWD to the package dir, so relative `testdata/...` paths are stable).
- **Table-driven** for the parity matrix and edge cases (one row per marker/method/fit/dataset).
- **testify: yes, but thin.** Use `require` for fatal setup (file load, fit error) and `assert` for non-fatal field checks — de-facto standard, low cost. **Do not** use `assert.InDelta` for numbers: it can't express abs-OR-rel and gives weak messages. Route every numeric compare through a custom helper that names the field:
  ```go
  func assertKmh(t *testing.T, field string, got, want float64) {
      t.Helper()
      if !testutil.EqualIntensityKmh(got, want) {
          t.Errorf("%s: got %.4f want %.4f (tol 0.1 km/h or 1%%)", field, got, want)
      }
  }
  ```
- **Float comparison:** wrappers over `gonum.org/v1/gonum/floats/scalar` (signatures verified): `EqualWithinAbs`, `EqualWithinAbsOrRel`, `EqualWithinRel`, plus `scalar.Round(x,prec)` for display rounding. HR interpolation uses `gonum.org/v1/gonum/interp.PiecewiseLinear` (`Fit(xs,ys) / Predict(x)`), which is the same engine the production code should use for FR-D3 — test and prod share it.
- **Golden pattern:** `-update` flag (`flag.Bool("update",false,...)`) gates **only** the `turnpoint/` regression snapshots (marshal engine output, write file). Reference goldens are read-only inputs, asserted within tolerance — never overwritten.
- **Determinism (NFR-6):** a test runs analysis twice on identical input and asserts byte-identical JSON output.

---

## 4. V4 edge-case tests (`fit/edge_test.go`, table-driven)

1. **Minimum step count (FR-T6 / OI-11):** rows `{steps:3 → AnalysisRefused error; 4 → runs + LowStepWarning; 5 → runs, no warning; 8 → runs}`. Degree-3 polynomial needs ≥4 points — assert the 3-step case errors *before* fitting, not a panic.
2. **Low-intensity lactate dip (built into Appendix A: 8 km/h 1.19 < 6 km/h 1.24):**
   - Assert the dip is accepted (no rejection), fit succeeds, and OBLA/IAS/IANS still reproduce Appendix C — i.e. the raw dip does **not** trip the FR-F3/OI-14 guard because the *fitted* curve stays monotonic over the fitted range.
   - Plus a **synthetic** dataset whose dip is large enough to make the fitted curve non-monotonic ⇒ assert the FR-F3 warning fires (interior local extremum) and affected methods are flagged, analysis still runs.
3. **Aborted final step (Appendix A: 20 km/h ends 22:10, ~1:10):**
   - Default (OI-13 include): step marked `aborted`, included in fit, MAX reproduces 20.0 / 185 / 7.7.
   - Variant with per-step exclude → MAX recomputes to the 18 km/h point and downstream thresholds shift; assert they move (guards the exclude path).
4. **Fit-strategy independence (FR-F2/FR-D4):** run Appendix A across `{3rd-poly default, exponential, spline}`; assert each yields finite thresholds, and that **pinned-fit methods (log-log, Exp-Dmax) ignore the displayed-fit selection** — their output is identical regardless of the default fit chosen (no silent fit/method mixing).

---

## 5. The two parity anchors — exact assertions

**Anchor A — OBLA 4.0 ⇒ IANS (V2):** configure LT2 ← OBLA 4.0, 3rd-poly on Appendix A.
- Intensity: fitted curve crosses 4.0 mmol/L between 16 km/h (3.89) and 18 km/h (6.66) ⇒ **16.1 km/h** (assert ±0.1 km/h / 1 %).
- HR: piecewise-linear on raw HR-vs-kmh, `16→166, 18→180`, at 16.1 ⇒ `166 + 0.05·14 = 166.7 → round 167` (assert ±1 bpm).
- Lactate: 4.0 by definition (±0.05).
- FR-D5 cross-check: the IANS results row must equal the OBLA 4.0 row field-for-field.

**Anchor B — GA1–SB zone table (V2, profile "Laufen Leistung 6×/Wo"):** zones = % of IANS (OI-17), IANS = 16.1. Each km/h boundary = `pct·16.1` rounded 1 dp; HR via the same `interp.PiecewiseLinear`; lactate via fitted curve; pace = `60/kmh`.

| Zone | %IANS | km/h (±0.1) | Lactate (±0.05) | HR (±1) | Pace |
|---|---|---|---|---|---|
| GA1 | 46–70 | 7.4–11.3 | 1.2–1.5 | 107–126 | 08:06–05:19 |
| GA2 | 70–88 | 11.3–14.2 | 1.5–2.5 | 126–150 | 05:19–04:14 |
| EB | 88–102 | 14.2–16.4 | 2.5–4.4 | 150–169 | 04:14–03:39 |
| SB | 102–125 | 16.4–20.1 | 4.4–7.7 | 169–185 | 03:39–02:58 |

Check: `0.46·16.1=7.41→7.4`, `0.70·16.1=11.27→11.3`, `0.88·16.1=14.17→14.2`, `1.02·16.1=16.42→16.4`, `1.25·16.1=20.13→20.1` — all reproduce within ±0.1.

**Known sensitivity to surface in the plan:** zone-**edge** HRs sit exactly at the ±1 bpm boundary (e.g. interp at 11.3 → 127.15 vs report 126; at 14.2 → 150.7 vs report 150). The marker anchors (167 @ 16.1, 177 @ 17.6) interpolate cleanly, but if zone-edge HRs fail by 1 bpm that's the signal to revisit the rounding rule (floor vs nearest) — precisely the "tighten after inspecting actual output" OI-1 anticipates. Keep the marker assertions hard and the zone-edge HRs within-tolerance-with-a-note.

---

## 6. Recommendation: R+`lactater` install — cost/benefit

**Do NOT install R+`lactater` in the build/CI environment. Proceed with WinLactat (Appendix C) as ground truth.**

| | Install R+lactater in CI | WinLactat-as-ground-truth + one-time lactater regen |
|---|---|---|
| Cost | R runtime (~150 MB) + tidyverse deps + GitHub-only pkg (`fmmattioni/lactater`, not on CRAN); flaky cross-platform CI; ongoing maintenance | Hand-encode Appendix C once; run R once on a dev box |
| Covers now | All methods | Fixed OBLA, IAS, IANS, MAX, zones (the credibility-critical, fit-reproducible ones) |
| Gap | — | Curve-only methods (log-log, Dmax/ModDmax/Exp-Dmax, LTP1/2, LTratio, D2Lmax, IAT) until regen runs |

**Plan:** ship V1/V2 against WinLactat goldens immediately. Before P0 sign-off, run `tools/regen-lactater` **once** on a developer machine (R + `remotes::install_github("fmmattioni/lactater")`), commit `golden/lactater/*.json` (frozen, with pinned commit SHA), and assert curve-only methods against it within OI-1. CI consumes the frozen JSON and never touches R. This keeps the build cgo-light/R-free per §10 while still satisfying V1 for every method.

---

## 7. CI-style test command list

```bash
# Static / hygiene
go vet ./...
go build ./...
# Core suite (race + no test cache)
go test ./... -race -count=1
# V1 parity (WinLactat now; lactater once frozen JSON exists)
go test ./threshold/... -run 'TestParity_WinLactat|TestParity_Lactater' -v
# V2 anchors + Appendix C fixtures
go test ./... -run 'TestAppendixC_Markers|TestAnchorA_OBLA4|TestAnchorB_Zones' -v
# V4 edge cases
go test ./fit/... ./zone/... -run 'TestEdge_(MinSteps|LactateDip|AbortedStep|FitIndependence)' -v
# Layer-3 regression snapshots (exact-ish, drift guard)
go test ./... -run TestGoldenRegression
# Coverage gate
go test ./... -coverprofile=cover.out && go tool cover -func=cover.out

# --- Manual / out-of-band, NOT in CI ---
go test ./... -run TestGoldenRegression -update            # refresh OUR snapshots only
go run ./tools/regen-lactater -out ./testdata/golden/lactater  # needs R; refresh lactater goldens
```

**V6 (no off-device flow) in the core:** add a guard test that fails if `turnpoint-core` transitively imports a network package — parse `go list -deps -json turnpoint-core/...` and assert no `net`, `net/http`, etc. (cheap static proof for the core; the full end-to-end network-monitoring check belongs to the Wails/e2e layer, out of this scope).

---

## Relevant files (to be created)
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/internal/testutil/floatcmp.go` — OI-1 tolerance consts + gonum `scalar` wrappers
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/internal/testutil/golden.go` — golden loader, `-update` flag, marker/zone asserters
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/testdata/datasets/appendix_a.json` — canonical input fixture
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/testdata/golden/winlactat/{appendix_c_markers,appendix_c_zones}.json` — frozen ground truth
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/testdata/golden/lactater/appendix_a.json` — regenerated out-of-band
- `/Users/radaiko/dev/private/Turnpoint/turnpoint-core/tools/regen-lactater/{main.go,lactater_export.R}` — one-time golden generator (R-gated)

**Verified APIs:** `gonum.org/v1/gonum/floats/scalar.{EqualWithinAbs(a,b,tol), EqualWithinAbsOrRel(a,b,absTol,relTol), EqualWithinRel(a,b,tol), Round(x,prec)}`; `gonum.org/v1/gonum/interp.PiecewiseLinear{Fit(xs,ys) error, Predict(x) float64}`; `lactater::lactate_threshold(.data, intensity_column, lactate_column, heart_rate_column, method=, fit=, include_baseline=, sport=)` → tibble `{method_category, method, fitting, intensity, lactate, heart_rate}`.