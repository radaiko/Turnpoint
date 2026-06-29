// Thin typed facade over the generated Wails bindings. Views import from here so
// the binding paths live in one place (the one mock seam). When running outside
// the Wails runtime (plain browser `npm run dev`), fall back to mock data.
import * as realApp from "$wails/go/main/App";
import { service, store, csvio } from "$wails/go/models";
import { mockApp } from "./mock";

const wailsReady =
  typeof window !== "undefined" && !!(window as any).go?.main?.App;

export const App = (wailsReady ? realApp : mockApp) as typeof realApp;

export type AnalysisDTO = service.AnalysisDTO;
export type MarkerRow = service.MarkerRow;
export type AnchorDTO = service.AnchorDTO;
export type ZoneDTO = service.ZoneDTO;
export type XY = service.XY;
export type StepBar = service.StepBar;
export type WarningDTO = service.WarningDTO;

export type Athlete = store.Athlete;
export type AthleteSummary = store.AthleteSummary;
export type Test = store.Test;
export type Step = store.Step;
export type Template = store.Template;
export type TrainingProfile = store.TrainingProfile;
export type ImportReport = csvio.ImportReport;

export { store, service };
