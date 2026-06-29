import { writable } from "svelte/store";
import { App, type AnalysisDTO } from "$lib/api";

export const analysis = writable<AnalysisDTO | null>(null);
export const analyzing = writable(false);

let seq = 0;

// runAnalysis is latest-wins: a newer call supersedes an in-flight older one.
export async function runAnalysis(testId: number) {
  const mine = ++seq;
  analyzing.set(true);
  try {
    const res = await App.Analyze(testId);
    if (mine === seq) analysis.set(res);
  } finally {
    if (mine === seq) analyzing.set(false);
  }
}

// recomputeZones is the drag fast path (FR-C2); also latest-wins.
export async function recomputeZones(testId: number, lt1: number, lt2: number) {
  const mine = ++seq;
  const res = await App.RecomputeZones(testId, lt1, lt2);
  if (mine === seq) analysis.set(res);
}
