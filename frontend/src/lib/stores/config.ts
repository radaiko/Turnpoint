import { writable, get } from "svelte/store";
import { App, type AnalysisConfigDTO, type MarkerOption, type ProfileOption } from "$lib/api";
import { analysis } from "./analysis";

// The active test's analysis configuration (FR-D2/F2/Z2/Z3/Z5).
export const config = writable<AnalysisConfigDTO | null>(null);
export const markerOptions = writable<MarkerOption[]>([]);
export const profileOptions = writable<ProfileOption[]>([]);

export async function loadConfig(testId: number, sport: string) {
  config.set(await App.GetAnalysisConfig(testId));
  if (get(markerOptions).length === 0) markerOptions.set(await App.GetMarkerOptions());
  profileOptions.set(await App.GetProfileOptions(sport));
}

// applyConfig persists the edited config and re-runs the analysis (FR-D2/F2/Z5).
export async function applyConfig(testId: number) {
  const c = get(config);
  if (!c) return;
  // cast at the binding boundary: the runtime just JSON-serialises the object
  analysis.set(await App.AnalyzeWith(testId, c as any));
}

// resetConfig reverts to the sport default.
export async function resetConfig(testId: number) {
  analysis.set(await App.ResetAnalysisConfig(testId));
  config.set(await App.GetAnalysisConfig(testId));
}

// setAnchor overrides an LT1/LT2 anchor by method or by direct intensity (FR-Z3).
export async function setAnchorMethod(testId: number, which: "lt1" | "lt2", marker: string) {
  config.update((c) => {
    if (!c) return c;
    if (which === "lt1") return { ...c, lt1Anchor: marker, lt1Override: undefined } as AnalysisConfigDTO;
    return { ...c, lt2Anchor: marker, lt2Override: undefined } as AnalysisConfigDTO;
  });
  await applyConfig(testId);
}

export async function setAnchorIntensity(testId: number, which: "lt1" | "lt2", intensity: number) {
  config.update((c) => {
    if (!c) return c;
    if (which === "lt1") return { ...c, lt1Override: intensity } as AnalysisConfigDTO;
    return { ...c, lt2Override: intensity } as AnalysisConfigDTO;
  });
  await applyConfig(testId);
}
