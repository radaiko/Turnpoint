import { writable } from "svelte/store";

import type { Region } from "$lib/format";

export type Section = "athletes" | "comparison" | "settings";
export type Stage = "entry" | "analysis" | "zones" | "report";
export type Theme = "light" | "dark";

export interface UIState {
  section: Section;
  activeAthleteId: number | null;
  activeTestId: number | null;
  stage: Stage;
  theme: Theme;
  region: Region;
}

function initialTheme(): Theme {
  const saved = localStorage.getItem("tp-theme");
  if (saved === "light" || saved === "dark") return saved;
  return "dark";
}

function initialRegion(): Region {
  const r = localStorage.getItem("tp-region");
  return r === "eu" || r === "us" ? r : "system";
}

export const ui = writable<UIState>({
  section: "athletes",
  activeAthleteId: null,
  activeTestId: null,
  stage: "entry",
  theme: initialTheme(),
  region: initialRegion(),
});

ui.subscribe((s) => {
  document.documentElement.setAttribute("data-theme", s.theme);
  localStorage.setItem("tp-theme", s.theme);
  localStorage.setItem("tp-region", s.region);
});

export function setRegion(region: Region) {
  ui.update((s) => ({ ...s, region }));
}

export function toggleTheme() {
  ui.update((s) => ({ ...s, theme: s.theme === "dark" ? "light" : "dark" }));
}

export function openAthlete(id: number) {
  ui.update((s) => ({ ...s, activeAthleteId: id, activeTestId: null, section: "athletes" }));
}

export function openTest(id: number) {
  ui.update((s) => ({ ...s, activeTestId: id, stage: "entry" }));
}

export function setStage(stage: Stage) {
  ui.update((s) => ({ ...s, stage }));
}

export function setSection(section: Section) {
  ui.update((s) => ({ ...s, section }));
}
