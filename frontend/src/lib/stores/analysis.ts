import { writable } from "svelte/store";
import { App, type AnalysisDTO } from "$lib/api";

export const analysis = writable<AnalysisDTO | null>(null);
export const analyzing = writable(false);

let seq = 0;

// runAnalysis is latest-wins: a newer call supersedes an in-flight older one.
// Errors (e.g. too few steps to analyse) clear the result rather than throwing.
export async function runAnalysis(testId: number) {
  const mine = ++seq;
  analyzing.set(true);
  try {
    const res = await App.Analyze(testId);
    if (mine === seq) analysis.set(res);
  } catch {
    if (mine === seq) analysis.set(null);
  } finally {
    if (mine === seq) analyzing.set(false);
  }
}

// ensureAnalysis runs the analysis only if no result is loaded yet.
export async function ensureAnalysis(testId: number) {
  let current: AnalysisDTO | null = null;
  const unsub = analysis.subscribe((v) => (current = v));
  unsub();
  if (!current) await runAnalysis(testId);
}

// recomputeZones is the drag fast path (FR-C2); also latest-wins.
export async function recomputeZones(testId: number, lt1: number, lt2: number) {
  const mine = ++seq;
  const res = await App.RecomputeZones(testId, lt1, lt2);
  if (mine === seq) analysis.set(res);
}
