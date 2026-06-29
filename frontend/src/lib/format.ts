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
