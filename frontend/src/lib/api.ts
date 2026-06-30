// Thin typed facade over the generated Wails bindings. Views import from here so
// the binding paths live in one place (the one mock seam). When running outside
// the Wails runtime (plain browser `npm run dev`), fall back to mock data.
import * as realApp from "$wails/go/main/App";
import { service, store, csvio, main } from "$wails/go/models";
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
// Config-related DTOs are declared as plain interfaces (not the generated Wails
// classes) so they can be spread/cloned in stores without losing class methods.
export interface MethodParamDTO {
  oblaConc: number;
  baselineDelta: number;
}
export interface AnalysisConfigDTO {
  displayFit: string;
  includeBaselineInFit: boolean;
  enabledMarkers: string[];
  methodParams: Record<string, MethodParamDTO>;
  lt1Anchor: string;
  lt2Anchor: string;
  lt1Override?: number;
  lt2Override?: number;
  profileName: string;
}
export interface MarkerOption {
  name: string;
  fitType: string;
}
export interface ProfileOption {
  name: string;
  sport: string;
  calibrated: boolean;
}

export type Athlete = store.Athlete;
export type AthleteSummary = store.AthleteSummary;
export type Test = store.Test;
export type Step = store.Step;
export type Template = store.Template;
export type TrainingProfile = store.TrainingProfile;
export type ImportReport = csvio.ImportReport;
export type UpdateInfo = main.UpdateInfo;

export { store, service };
