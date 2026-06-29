// Thin typed facade over the generated Wails bindings. Views import from here so
// the binding paths live in one place (the one mock seam).
import * as App from "$wails/go/main/App";
import { service, store, csvio } from "$wails/go/models";

export { App };

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
