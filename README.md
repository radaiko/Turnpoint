# Turnpoint

A local-first desktop app for analysing blood-lactate step tests — fits a lactate
curve, computes training thresholds with a suite of literature methods, derives a
5-zone training model, and produces a printable report. All data stays on your
machine. A hobbyist-focused alternative to clinical lactate software (WinLactat).

See [`docs/REQUIREMENTS.md`](docs/REQUIREMENTS.md) for the full specification and
[`docs/DESIGN.md`](docs/DESIGN.md) for the technical design.

## Architecture

Functional-core / imperative-shell, one Go module, single binary (Wails v2):

```
core/        pure compute (gonum only) — fits, 16 threshold methods, zones, pipeline
internal/    app layer — SQLite store, CSV I/O, backup, service DTOs
frontend/    Svelte 4 + TypeScript UI (custom SVG charts on d3-scale/shape)
app.go       Wails-bound facade (the API the frontend calls)
```

- **Core** depends only on the Go stdlib + `gonum`; a deps-purity test enforces it.
- **Storage** is a single SQLite file (`modernc.org/sqlite`, cgo-free).
- **Validation:** the core reproduces the WinLactat reference report (SRS Appendix
  A/C) within a defined tolerance — e.g. OBLA 4.0 → 16.1 km/h / 167 bpm.

## Prerequisites

- Go 1.23+
- Node 18+ / npm
- [Wails v2 CLI](https://wails.io): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Develop

```sh
wails dev        # hot-reloading desktop app
```

## Build

```sh
wails build      # → build/bin/Turnpoint.app (or .exe / binary)
```

## Test

```sh
go test ./...                        # Go: core (parity), persistence, app facade
cd frontend && npm run check         # frontend type-check (svelte-check)
cd frontend && npm run build         # frontend production bundle
```

The Go suite includes the binding-level happy-path (`app_test.go`) and the core
parity suite asserting the Appendix C reference values.

## License

Source-available, all rights reserved. See [`LICENSE`](LICENSE).
