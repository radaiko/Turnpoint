# Turnpoint v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Full technical detail** for every task lives in `docs/DESIGN.md` (consolidated design + Appendices A–F with exact gonum signatures, DDL, algorithm code, design tokens). This plan is the task spine; `docs/DESIGN.md` is the reference. Requirements IDs (FR-/NFR-/OI-) refer to `docs/REQUIREMENTS.md`.

**Goal:** Build Turnpoint — a local-first Wails v2 desktop app that ingests blood-lactate step-test data, fits a lactate curve, computes ~16 literature threshold methods, derives a 5-zone training model, renders two interactive charts, and exports a PDF report — validated against the Appendix C reference report.

**Architecture:** Functional-core / imperative-shell. A pure gonum-only Go package `core/` does all numerics (data → fits → thresholds → LT1/LT2 → zones → derived metrics) with no UI/DB deps. An app layer `internal/` adds SQLite persistence (modernc, cgo-free), CSV/backup, and reporting. A Svelte+TypeScript frontend (custom SVG charts on LayerCake + d3-scale/shape, custom data grid) is bound to Go via Wails. One Go module, repo root `github.com/radaiko/turnpoint`.

**Tech Stack:** Go 1.26 · gonum v0.15.x · Wails v2.12 · Svelte + Vite + TS · modernc.org/sqlite · maroto/v2 + tdewolff/canvas (PDF/export fallback) · LayerCake + d3-scale + d3-shape · Geist Sans/Mono (self-hosted).

## Global Constraints

- **Module path** `github.com/radaiko/turnpoint`; `go 1.23` floor, toolchain 1.26. ONE module.
- **`core/` purity:** imports stdlib + `gonum.org/v1/gonum` ONLY. Enforced by a deps-allowlist test (`go list -deps ./core/...`). No `core/` file may import `internal/`, wails, or sqlite. (V6, SRS §10.)
- **cgo-free:** `CGO_ENABLED=0` must build the whole app. SQLite driver = `modernc.org/sqlite` (driver name `"sqlite"`).
- **Offline / no telemetry:** no network calls in any core feature; fonts self-hosted woff2, never a Google Fonts `<link>` (NFR-2/5, FR-M4).
- **Determinism:** identical `(Input, Config)` ⇒ byte-identical `Result` JSON. No `time.Now()`/`rand` in `core/` (NFR-6).
- **Warnings are data, never Go errors.** Only hard preconditions are sentinel errors (`errors.Is`-checkable). Everything the SRS calls "non-blocking" rides as `domain.Warning` (DESIGN §9 / Appendix B §9).
- **Intensity stored in native unit** (km/h or W) as REAL; pace is always derived; time stored as INTEGER seconds (OI-9, OI-21).
- **Parity ground truth = WinLactat (Appendix C)** now; `lactater` golden layer is regenerated out-of-band (R-gated) and frozen as JSON — CI never runs R (DESIGN §7 risk 1, Appendix F).
- **Tolerances (OI-1):** intensity(run) abs 0.1 **or** rel 1%; intensity(cycle) abs 2 or rel 1%; HR abs 1; lactate abs 0.05. Helper `core/internal/testutil/floatcmp.go`.
- **Commit after every green task.** TDD: failing test → minimal impl → green → commit.

---

## Phase P0 — `core/` (pure compute, zero UI/DB)

The credibility-critical core. Fully testable headless with `go test ./core/...`. Build bottom-up along the import DAG `unit ← domain ← {fit, zone, metrics} ← threshold ← analysis`.

### Task 0.0a: Repo scaffold + Go module + deps-purity guard

**Files:** Create `go.mod`, `core/doc.go`, `core/internal/depsguard/depsguard_test.go`.

**Interfaces:** Produces the module and the purity test all later tasks rely on.

- [ ] Init module `github.com/radaiko/turnpoint` (`go mod init`), `go get gonum.org/v1/gonum@latest`.
- [ ] Write `core/internal/depsguard/depsguard_test.go`: shell out to `go list -deps ./...` for each `core/...` package, assert every dep is stdlib or under `gonum.org/v1/gonum`. (V6.)
- [ ] Run it — fails (no core pkgs yet) or passes trivially; commit.

### Task 0.0b: `core/unit` — Sport, Unit, Intensity, Pace, clock (OI-9, FR-T4/T5)

**Files:** Create `core/unit/{unit.go,sport.go,pace.go}`, `core/unit/unit_test.go`.

**Interfaces (Produces):** `Sport{Running,Cycling}` + `.Unit() .HasPace() .String()`; `Unit{UnitKmh,UnitWatt}` + `.Symbol()`; `PaceFromKmh(float64) Pace`; `Pace.MMSS() string`; `ParseClock(string)(time.Duration,error)`; `FormatClock(time.Duration) string`. (Signatures: DESIGN Appendix B §1.)

- [ ] Test: `PaceFromKmh(16.1).MMSS()=="03:43"`; `PaceFromKmh(20)=="03:00"`; `kmh<=0 ⇒ "—"/zero`; `Running.HasPace()` true, `Cycling` false; `UnitWatt.Symbol()=="W"`; `ParseClock("22:10")==22m10s`; `FormatClock` round-trip.
- [ ] Implement; green; commit.

### Task 0.0c: `core/domain` — records + Warning (FR-T1/T2/T3, OI-13, DAG anchor)

**Files:** Create `core/domain/{protocol.go,step.go,test.go,warning.go}`, `core/domain/domain_test.go`.

**Interfaces (Produces):** `Protocol`, `Step` (incl. `HasLactate,Aborted,Excluded,RPE *int`), `Test{Protocol,Steps,BodyMassKg}`; `Test.FitPoints(includeBaseline bool) []fit.Point`?? — to avoid `domain→fit` cycle, `FitPoints` returns `[]struct{X,Y float64}` or lives where `fit.Point` is; **decision:** `domain.FitPoint{X,Y float64}` defined in domain, `fit.Point` aliases it. `Test.Baseline()(float64,bool)`, `Test.MaxIntensity() float64`; `Warning{Code,Severity,Subject,Message}` + `WarnCode` consts (DESIGN Appendix B §2–3).

- [ ] Test: `FitPoints` sorts by X, dedups, drops `!HasLactate`/`Excluded`, includes intensity-0 only when `includeBaseline`; `Baseline()` returns the intensity-0 row; `MaxIntensity()==20` for Appendix A.
- [ ] Implement; green; commit.

### Task 0.0d: Foundation testdata — Appendix A dataset + Appendix C goldens

**Files:** Create `core/testdata/datasets/appendix_a.json`, `core/testdata/golden/winlactat/{appendix_c_markers.json,appendix_c_zones.json}`, `core/internal/testutil/{floatcmp.go,golden.go}`.

**Interfaces (Produces):** loader `testutil.LoadAppendixA() domain.Test`; `testutil.EqualWithinTol(...)` wrappers over `gonum/floats/scalar.EqualWithinAbsOrRel`; golden loader with `-update` flag. Exact numbers: DESIGN §6 Fixtures.

- [ ] Encode Appendix A (9 rows incl. baseline; 8 km/h dip; 20 km/h aborted at 22:10) and the Appendix C marker/zone tables as JSON.
- [ ] Implement loaders + tolerance helpers; a smoke test loads them; commit.

### Task 0.1: `core/numeric` — roots + segmented regression (NFR-6)

**Files:** Create `core/numeric/{roots.go,segreg.go}`, `core/numeric/numeric_test.go`.

**Interfaces (Produces):** `LevelSetRoot(f func(float64)float64, lo,hi,target float64)(root float64, ok bool)` (grid sign-change + Brent/bisection); `PolyRoots(coef []float64) []complex128` (companion-matrix `mat.Eigen`); `SegmentedFit(x,y []float64, k int)(knots []float64, rss float64)` (continuous piecewise-linear basis, deterministic grid search). Detail: DESIGN Appendix A §3.

- [ ] Test: root of `x²−2` on `[0,2]`→√2; segmented k=1 recovers a synthetic breakpoint; determinism (two runs identical); commit per primitive.

### Task 0.2: `core/fit` — fits + quality guard (FR-F1/F2/F3, OI-14)

**Files:** Create `core/fit/{fit.go,poly.go,exp.go,spline.go,loglog.go,quality.go,eval.go}`, `core/fit/fit_test.go`.

**Interfaces (Produces):** `Kind{Poly3,Exp,Spline,LogLog,Segmented,None}`; `Point{X,Y}`; `Fit` interface (`Kind/Predict/Derivative/SecondDerivative/Domain/Quality`); `Quality{R2,Monotonic,Conditioned,LocalExtremum,Warnings}`; factories `New/Poly/Exponential/Spline/LogLogSeg`; sentinels `ErrTooFewPoints,ErrSingular,ErrNonPositive`. Full code: DESIGN Appendix A §2 + Appendix B §4.

- [ ] **Poly3:** centred/scaled Vandermonde QR (`mat.QR.SolveVecTo`); analytic `Predict/Derivative/SecondDerivative`; `mat.Cond` for ill-conditioning. Test: recover known cubic coeffs; `n<4 ⇒ ErrTooFewPoints`; `n==4 ⇒ R²≈1 + WarnFewSteps-class flag`.
- [ ] **quality.go:** sample 200 pts; non-monotone sign change → `WarnNonMonotonicFit`+`LocalExtremum`; `R²<0.95 → WarnLowR2`; cond>1e8 → `WarnIllConditioned`. Test each flag fires/doesn't on crafted data.
- [ ] **Exp:** `a+b·e^{cx}` NLS via `optimize.Minimize`+`NelderMead`, log-linear warm start; `c≤0 ⇒` flag inappropriate; non-convergence ⇒ `Computable:false`. Test on synthetic exp data.
- [ ] **Spline:** `interp.FritschButland` wrapper (monotone ⇒ passes FR-F3); 2nd deriv via `fd.Central2nd`. (Eilers–Marx P-spline deferred — DESIGN risk 12.)
- [ ] **LogLogSeg:** segmented k=1 on `(ln x, ln y)` raw points; breakpoint `exp(ψ)`; `y<=0 ⇒ ErrNonPositive`. Commit per fit.

### Task 0.3a: `core/metrics` — derived metrics + HR curve (FR-D3, OI-15)

**Files:** Create `core/metrics/{derived.go,hrcurve.go}`, `core/metrics/metrics_test.go`.

**Interfaces (Produces):** `DerivedMetrics{Intensity,PctMax,Pace,HasPace,HeartRate,KcalPerHour,HasKcal,Lactate}`; `PctMax(i,max)`; `KcalPerHourRunning(m,v)=m*v*1.036`, `KcalPerHourCycling(W)=W*3.6`; `HRCurve` (wraps `interp.PiecewiseLinear` over HasLactate steps, baseline excluded) + `NewHRCurve/At`; `Derive(...)`. Detail: DESIGN Appendix B §7.

- [ ] Test: `PctMax(16.1,20)==80.5`; `HRCurve.At(16.1)==167` (interp 16→166,18→180); `BodyMassKg==0 ⇒ HasKcal=false`; pace zero for cycling. Commit.

### Task 0.3b: `core/zone` — 5-zone %-of-IANS bands + profiles (FR-Z1/Z4/Z5, OI-17)

**Files:** Create `core/zone/{zone.go,profile.go}`, `core/zone/zone_test.go`.

**Interfaces (Produces):** `Index{REKOM,GA1,GA2,EB,SB}`+German/English; `SpreadRule{SpreadPctIANS}`; `Band{Zone,LowPct,HighPct}`; `TrainingProfile{Name,Sport,Level,WeeklyFreq,Rule,Bands,GermanLabels}`; `Zone{Index,Label,Intensity/HR/Lactate/Pace Low/High}`; `Derive(p,ias,ians,curve,hr,sport)[]Zone`; `Predefined()`; `LaufLeistung6()`. Calibrated bands (REKOM 0–.46, GA1 .46–.70, GA2 .70–.88, EB .88–1.02, SB 1.02–1.25): DESIGN Appendix B §6.

- [ ] Test: `LaufLeistung6` at IANS=16.1 reproduces Appendix C zone km/h (7.4/11.3/14.2/16.4/20.1 within ±0.1); lactate/HR/pace read off curve+HR at bounds. Other 5 profiles ship `TODO`-flagged (OI-17). Commit.

### Task 0.4a: `core/threshold` — method interface + all 16 markers (FR-D1/D2/D4)

**Files:** Create `core/threshold/{method.go,obla.go,baselineplus.go,loglog.go,dmax.go,moddmax.go,expdmax.go,ltp.go,iat.go,ltratio.go,d2lmax.go,max.go}`, `core/threshold/threshold_test.go`.

**Interfaces (Produces):** `Marker` enum (OBLA2/3/4/6, Bsln05/10/15, LogLog, Dmax, ModDmax, ExpDmax, LTP1, LTP2, IAT, LTratio, D2Lmax, MAX) + `.String()`; `Params{OBLAConc,BaselineDelta}`; `Context{Points,Steps,BaselineLactate,HasBaseline,Params}`; `Result{Marker,Intensity,Lactate,FitKind,Computable,Reason,Warnings}`; `ThresholdMethod{Marker,RequiredFit,Compute}`; `Default()`, `For(...)`. Per-method algorithm + `RequiredFit` table: DESIGN Appendix A §4 + Appendix B §5. **D2Lmax pins Poly4** (DESIGN risk 2); **Bsln+ baseline = resting if>0 else min(l)** (risk 5); **ModDmax start = first raw Δ>0.4** (=14 km/h on Appendix A).

- [ ] Implement one method per file, each with a table-driven test. **Binding anchor (V2):** OBLA4 → 16.1 km/h / 167 bpm / 4.0 (hard assert, ±tol). Dmax/ModDmax via quadratic root; ExpDmax closed form in `c`; LTP via segmented k=2 on interpolated grid; IAT = curve-min +1.5; LTratio = argmin L/x; MAX = peak step. Methods not in lactater (IAT, D2Lmax) use hand-computed fixtures.
- [ ] `RequiredFit()` test: each marker maps to its canonical `Kind`. Commit per method.

### Task 0.4b: `core/analysis` — pipeline + public API (FR-F4, NFR-3/6)

**Files:** Create `core/analysis/{analysis.go,pipeline.go,recompute.go}`, `core/analysis/analysis_test.go`.

**Interfaces (Produces — the surface `app.go` binds):** `Input{Test}`; `Config{DisplayFit,IncludeBaselineInFit,EnabledMarkers,MethodParams,LT1Anchor,LT2Anchor,LT1Override,LT2Override,Profile}`; `Override{Intensity}`; `Anchor{Marker,Intensity,Manual,Metrics}`; `Result{Fits,DisplayFit,Thresholds,Markers,LT1,LT2,Zones,MaxIntensity,Warnings}`; `Analyze(in,cfg)(Result,error)`; `RecomputeZones(prev,in,cfg,lt1,lt2)(Result,error)`; `Validate(in)[]Warning`; sentinels `ErrInsufficientSteps,ErrNoBaseline`. Pipeline stages 1–7 + defaults (LT1←LogLog, LT2←OBLA4): DESIGN §4 + Appendix B §8.

- [ ] **Pipeline test (V2, binding):** full Appendix A `Analyze` reproduces the Appendix C marker table + zone table within tol; IANS row == OBLA4 row field-for-field.
- [ ] **Determinism (NFR-6):** `Analyze` twice → byte-identical JSON.
- [ ] **Edge (V4):** 3 steps → `ErrInsufficientSteps`; 4 → runs + `WarnFewSteps`; raw 8 km/h dip accepted; synthetic dip trips FR-F3; aborted step included → MAX 20.0/185/7.7, exclude variant shifts; pinned-fit methods identical across `DisplayFit` choices.
- [ ] `RecomputeZones` reuses `prev.Fits`, only re-derives anchors+zones+metrics; result equals a full `Analyze` for the same LT1/LT2. Commit. **← core complete; tag `core-v1`.**

---

## Phase P1 — Persistence + App shell

### Task 1.0a: `internal/store` — schema, migrator, DSN (FR-M1, OI-4)

**Files:** Create `internal/store/{db.go,migrate.go,models.go}`, `internal/store/migrations/{0001_init.sql,0002_seed_predefined.sql}`, `internal/store/store_test.go`.

**Interfaces (Produces):** `Open(path string)(*DB,error)` (DSN: `_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)`, `SetMaxOpenConns(1)`); `Migrate(*DB) error` (`PRAGMA user_version` + `embed.FS`, one tx/file). Full DDL (all §9 tables, `report_settings` merged shape, FK cascade athlete→test→{step,threshold_result,zone,report_settings}, `fit_type CHECK`): DESIGN Appendix C.

- [ ] Test on a temp-file DB: migrate-from-empty sets user_version; re-migrate is a no-op; `PRAGMA foreign_keys` is ON; cascade delete removes child rows; seed inserts predefined templates+profiles. Commit.

### Task 1.0b: repositories (FR-A1–A5, FR-T1–T8, FR-D5, FR-Z*)

**Files:** Create `internal/store/{athlete_repo.go,test_repo.go,step_repo.go,threshold_repo.go,zone_repo.go,template_repo.go,profile_repo.go,report_repo.go}` + tests.

**Interfaces (Produces):** CRUD per entity, e.g. `AthleteRepo.{Create,Get,List,Search,Update,Delete}`; `TestRepo.*` (copies `body_mass_snapshot` from athlete at create, OI-3); `StepRepo.ReplaceForTest`; snapshot writers for thresholds/zones.

- [ ] Round-trip test per repo on temp DB; athlete delete cascades; `body_mass_snapshot` frozen on athlete edit (OI-3). Commit per repo.

### Task 1.0c: `internal/csvio` + `internal/backup` (FR-M2/M3, OI-10/21)

**Files:** Create `internal/csvio/{csv.go,time.go}`, `internal/backup/backup.go` + tests.

**Interfaces (Produces):** `ImportSteps(r,opts)([]domain.Step,[]Warning,error)`, `ExportTests(w,...)`, `ParsePasted(text)([]domain.Step,error)` (auto-detect TSV/CSV, dot/comma decimal); `Backup(db,dst)` (`VACUUM INTO`), `Restore(src,dst)` (integrity_check → rename → drop wal/shm → reopen+migrate).

- [ ] Test: Appendix A CSV round-trips; paste auto-detect (tab + comma + semicolon, comma-decimal); backup file restores to identical dataset. Commit.

### Task 1.1a: Wails shell — `wails init`, `app.go`, bindings (NFR-1)

**Files:** `wails init -t svelte-ts`; create `main.go`, `app.go`; configure `wails.json`, frameless window options (DESIGN §5 shell).

**Interfaces (Produces):** Wails-bound `App` methods wrapping `analysis` + repos: `App.{ListAthletes,SaveAthlete,DeleteAthlete,GetTest,SaveSteps,Analyze(testID,cfg),...}` returning JSON-marshalled DTOs (`internal/service/dto.go`). `OnStartup` opens+migrates the DB under the OS app-data dir.

- [ ] `wails build` (or `wails dev`) produces a running frameless shell on this platform (NFR-1 smoke). `wails generate module` emits `frontend/wailsjs`. Commit.

### Task 1.1b: Frontend foundation — tokens, fonts, shell, stores

**Files:** Create `frontend/src/lib/design/tokens.css` + Geist woff2; `app.css`; `App.svelte` (44px titlebar + 240px nav rail + stage tab bar); `lib/stores/{ui,athletes,session,analysis,config}.ts`; `lib/api/*` typed wrappers; base components (`Button,Field,NumberField,Select,Tabs,Table,Modal,Toast,Toggle,Tag,EmptyState`). Tokens/shell: DESIGN §5.

- [ ] App renders the shell with light/dark toggle; `analysis` store is debounced latest-wins. Manual smoke via `wails dev`. Commit.

### Task 1.1c: Athletes CRUD view + Test Entry grid (FR-A*, FR-T5/T6/T7/T8, OI-10)

**Files:** Create `frontend/src/views/{Athletes,TestEntry}.svelte`, `lib/components/DataGrid.svelte`, `lib/format/{mmss,pacePer1000,parseTabular}.ts`.

**Interfaces (Consumes):** `App.ListAthletes/SaveAthlete/...`, `App.SaveSteps`. Custom `DataGrid`: 5-col canonical layout, unit-switch header, keyboard nav, add/remove/reorder rows, paste auto-detect.

- [ ] Walkthrough (V5 partial): create athlete → new test from Running template → paste Appendix A → rows validated (min-steps, ranges) → save. Commit.

---

## Phase P2 — Charts & interaction

### Task 2.0: `FitChart.svelte` — fitting view + draggable markers (FR-C1/C2/C3, NFR-3)

**Files:** Create `frontend/src/lib/charts/{FitChart,Axis,ZoneBands,DraggableMarker,StepBars}.svelte`, `views/Analysis.svelte`. Custom SVG on LayerCake + d3-scale/shape.

- [ ] Fitted curve + raw points + HR overlay + threshold markers + zone bands; layer toggles (OI-18). Dragging a marker calls `App.RecomputeZones`, updates intensity/HR/pace/zones live; manual override styled distinctly (dashed accent, FR-C3) and persisted. Verify drag latency feels instant (NFR-3). Commit.

### Task 2.1: Analysis results table + method toggles (FR-D5/D2)

**Files:** Extend `Analysis.svelte`; results table component.

- [ ] FR-D5 table: per marker intensity/lactate/HR/%max/pace/kcal-h, fit label, computable/warn state; enable/disable methods + configure OBLA/baseline params; recompute. Commit.

### Task 2.2: `TemporalChart.svelte` — primary results chart (FR-C4)

**Files:** Create `frontend/src/lib/charts/TemporalChart.svelte`.

- [ ] X=time; dual Y (lactate left / HR right); intensity step bars; five zone bands; IAS/IANS vertical lines; configurable reference lines. Reproduces Appendix C layout for Appendix A. Commit.

### Task 2.3: Zones view + Comparison view (FR-Z3/Z5, FR-H1/H2/H3)

**Files:** Create `frontend/src/views/{Zones,Comparison}.svelte`.

- [ ] Zones: ranges table + profile selector + LT1/LT2 anchor override (method select / direct edit). Comparison: longitudinal progression (single athlete IAS/IANS/MAX over time), cross-sectional side-by-side, overlaid lactate curves. Commit.

---

## Phase P3 — Reporting

### Task 3.0: `internal/report` block model + HTML print route (FR-R1/R2/R3/R4, OI-20)

**Files:** Create `internal/report/{blocks.go,html.go}`, `frontend/src/views/Report.svelte` (print `@page` CSS, A4 portrait).

**Interfaces (Produces):** `ReportBlock` ordered model (include/omit/reorder/edit); default = reference pages 2&3 (raw table + temporal chart, threshold table + zones); cover/anamnesis + evaluation optional. Reuses the live Svelte chart `<svg>` as true vector. `window.print()` → OS Save-as-PDF.

- [ ] Default report renders pages 2&3; reorder/omit reflected; multi-page paginates without clipping; custom header/logo + footer. Commit.

### Task 3.1: Export — PNG/SVG chart + CSV results/zones + maroto fallback (FR-R5)

**Files:** Create `internal/report/{chartexport.go,maroto.go}`.

**Interfaces (Produces):** `tdewolff/canvas` `ParseSVG` → `renderers.{PDF,PNG(300dpi),SVG}`; CSV writers for results/zones; maroto/v2 headless report fallback (DESIGN §7 risk 8). `App.SaveFileDialog` wiring.

- [ ] PNG/SVG/CSV exports valid; maroto fallback produces a multi-page PDF headless. Commit.

---

## Phase P4 — History tail + lactater goldens

### Task 4.0: lactater golden layer (V1, R-gated, out-of-band)

**Files:** Create `core/tools/regen-lactater/{main.go,lactater_export.R}`, `core/testdata/golden/lactater/*.json`.

- [ ] On a dev box with R: install `fmmattioni/lactater` (pin SHA), export method outputs for Appendix A + demo data to frozen JSON. Add curve-method parity tests consuming the frozen JSON within OI-1. (Not in CI.) Confirm/661 revise IAS default mapping (OI-16) and LTP/log-log/LTratio parity caveats (DESIGN risks 3/4/10). Commit.

---

## Self-Review — spec coverage

- **Athlete (FR-A1–5):** 1.0b repos + 1.1c view. ✓ Delete cascade OI-4 in 1.0a DDL.
- **Test entry (FR-T1–8):** domain 0.0c, repos 1.0b, grid 1.1c, templates seed 1.0a + 1.1c. ✓
- **Curve fitting (FR-F1–4):** 0.2 fits, 0.4b reactive pipeline. ✓ FR-F3/OI-14 in quality.go.
- **Thresholds (FR-D1–5):** 0.4a methods + 0.4b table; FR-D4 RequiredFit; FR-D5 metrics. ✓
- **Zones (FR-Z1–6):** 0.3b zone math + 0.4b anchors + 2.3 override UI + profiles. ✓
- **Visualisation (FR-C1–4):** 2.0 FitChart, 2.2 TemporalChart. ✓
- **Comparison (FR-H1–3):** 2.3. ✓
- **Reporting (FR-R1–5):** 3.0 + 3.1. ✓
- **Data mgmt (FR-M1–4):** 1.0a/b/c; offline guard global constraint. ✓
- **Threshold methods §6:** all 16 markers in 0.4a; D2Lmax/Poly4 + IAT fixtures flagged. ✓
- **Zone model §7:** 0.3b calibrated + others TODO (OI-17). ✓
- **NFR-1–8:** 1.1a platform smoke, global offline/determinism, 2.0 perf, schema doc, deps-guard. ✓
- **Validation V1–6:** 0.0d fixtures, 0.4b parity+determinism+edges, depsguard V6, 4.0 lactater V1. ✓
- **Known gaps (carried, not silent):** R/lactater absence (P4), D2Lmax/IAT no-oracle, IAS default OI-16, LTratio/LTP parity, kcal/h OI-15, Wails headless-PDF — all in DESIGN §7.

---

## Execution

Build order = phase order. Each task ends green + committed. `core/` (P0) is the independently shippable, fully-verifiable milestone; tag `core-v1` at end of 0.4b. P1–P3 depend on Wails build working on this machine; the frontend interaction tasks (P2) are verified via `wails dev` smoke + the headless-testable Go underneath.
