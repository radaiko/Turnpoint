// Display formatting helpers (tabular, locale-stable).

export function mmss(totalSeconds: number): string {
  if (!isFinite(totalSeconds) || totalSeconds <= 0) return "—";
  const s = Math.floor(totalSeconds);
  return `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
}

// parse "mm:ss" → seconds (null on blank/invalid)
export function parseMMSS(s: string): number | null {
  const t = s.trim();
  if (!t) return null;
  const parts = t.split(":");
  if (parts.length < 2 || parts.length > 3) return null;
  let total = 0;
  for (const p of parts) {
    const n = parseInt(p, 10);
    if (isNaN(n) || n < 0) return null;
    total = total * 60 + n;
  }
  return total;
}

export function num(v: number | null | undefined, places = 1): string {
  if (v === null || v === undefined || !isFinite(v)) return "—";
  return v.toFixed(places);
}

export function intStr(v: number | null | undefined): string {
  if (v === null || v === undefined || !isFinite(v) || v === 0) return "—";
  return String(Math.round(v));
}

// Region preference drives date (and number) formatting. "system" follows the OS.
export type Region = "system" | "eu" | "us";

export function getRegion(): Region {
  const r = typeof localStorage !== "undefined" ? localStorage.getItem("tp-region") : null;
  return r === "eu" || r === "us" ? r : "system";
}

function dateLocale(region = getRegion()): string | undefined {
  if (region === "eu") return "de-DE"; // 31.12.2025
  if (region === "us") return "en-US"; // 12/31/2025
  return undefined; // system locale
}

// formatDate renders an ISO yyyy-mm-dd date per the chosen region (FR — region
// dependent display). The `dep` arg lets a Svelte view pass $ui.region so the
// markup re-renders when the region changes.
export function formatDate(iso: string | null | undefined, _dep?: unknown): string {
  if (!iso) return "—";
  const d = new Date(iso.length <= 10 ? iso + "T00:00:00" : iso);
  if (isNaN(d.getTime())) return String(iso);
  return new Intl.DateTimeFormat(dateLocale(), { year: "numeric", month: "2-digit", day: "2-digit" }).format(d);
}

// formatDecimal renders a number with the region's decimal separator.
export function formatDecimal(v: number | null | undefined, places = 1, _dep?: unknown): string {
  if (v === null || v === undefined || !isFinite(v)) return "—";
  const region = getRegion();
  const loc = region === "eu" ? "de-DE" : region === "us" ? "en-US" : undefined;
  return new Intl.NumberFormat(loc, { minimumFractionDigits: places, maximumFractionDigits: places }).format(v);
}

// age from an ISO yyyy-mm-dd date of birth
export function ageFromDOB(dob: string | null | undefined): string {
  if (!dob) return "—";
  const d = new Date(dob);
  if (isNaN(d.getTime())) return "—";
  const now = new Date();
  let age = now.getFullYear() - d.getFullYear();
  const m = now.getMonth() - d.getMonth();
  if (m < 0 || (m === 0 && now.getDate() < d.getDate())) age--;
  return String(age);
}

export const ZONE_COLORS: Record<string, string> = {
  REKOM: "var(--zone-rekom)",
  GA1: "var(--zone-ga1)",
  GA2: "var(--zone-ga2)",
  EB: "var(--zone-eb)",
  SB: "var(--zone-sb)",
};
