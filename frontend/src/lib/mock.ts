// Browser-dev mock: when the app runs outside the Wails runtime (plain `npm run
// dev` in a browser), api.ts falls back to these stubs so the UI renders with
// realistic sample data. No-op in the packaged desktop app.
import analysisDTO from "./mockAnalysis.json";

const athletes = [
  { id: 1, name: "Bogner Markus", primarySport: "running", lastTestDate: "2025-02-01", testCount: 1 },
  { id: 2, name: "Sara Klein", primarySport: "cycling", lastTestDate: "2025-03-12", testCount: 3 },
  { id: 3, name: "Tomáš Novák", primarySport: "running", lastTestDate: "2024-11-20", testCount: 2 },
];

const athleteFull: Record<number, any> = {
  1: { id: 1, name: "Bogner Markus", dob: "1991-06-04", sex: "male", bodyMassKg: 72, primarySport: "running", notes: "" },
  2: { id: 2, name: "Sara Klein", dob: "1996-02-18", sex: "female", bodyMassKg: 61, primarySport: "cycling", notes: "" },
  3: { id: 3, name: "Tomáš Novák", dob: "1988-09-30", sex: "male", bodyMassKg: 78, primarySport: "running", notes: "" },
};

const tests = [
  { id: 10, athleteId: 1, testDate: "2025-02-01", sport: "running", stepDurationS: 180, increment: 2, startIntensity: 6, mode: "continuous", pretestNote: "", remarks: "Nicht ausbelastet · EndLac 8.72 mmol" },
];

const steps = [
  [0, 0, 0, 0], [6, 180, 98, 1.24], [8, 360, 111, 1.19], [10, 540, 120, 1.32], [12, 720, 131, 1.66],
  [14, 900, 149, 2.38], [16, 1080, 166, 3.89], [18, 1260, 180, 6.66], [20, 1330, 185, 7.74],
].map(([i, ts, hr, lac], idx) => ({
  id: idx + 1, testId: 10, stepOrder: idx, intensity: i, timePointS: ts, heartRate: hr, lactate: lac,
  isBaseline: i === 0, excluded: false, aborted: i === 20,
}));

const templates = [
  { id: 1, name: "Running (Lauf)", sport: "running", stepDurationS: 180, increment: 2, startIntensity: 6, endIntensity: 22, mode: "continuous", visibleColumns: "", isPredefined: true },
  { id: 2, name: "Cycling (Rad)", sport: "cycling", stepDurationS: 240, increment: 40, startIntensity: 80, endIntensity: 440, mode: "continuous", visibleColumns: "", isPredefined: true },
];

const defaultConfig = {
  displayFit: "poly3",
  includeBaselineInFit: false,
  enabledMarkers: ["OBLA 2.0", "OBLA 4.0", "OBLA 6.0", "Log-log", "Dmax", "ModDmax", "LTP1", "LTP2", "IAT", "MAX"],
  methodParams: {},
  lt1Anchor: "Log-log",
  lt2Anchor: "OBLA 4.0",
  lt1Override: null,
  lt2Override: null,
  profileName: "Laufen Leistungssportler (6×/Woche)",
};
let cfgState: any = { ...defaultConfig };

const markerOptions = [
  { name: "OBLA 2.0", fitType: "spline" }, { name: "OBLA 3.0", fitType: "spline" },
  { name: "OBLA 4.0", fitType: "spline" }, { name: "OBLA 6.0", fitType: "spline" },
  { name: "Bsln+0.5", fitType: "spline" }, { name: "Bsln+1.0", fitType: "spline" },
  { name: "Bsln+1.5", fitType: "spline" }, { name: "Log-log", fitType: "loglog" },
  { name: "Dmax", fitType: "poly3" }, { name: "ModDmax", fitType: "poly3" },
  { name: "Exp-Dmax", fitType: "exp" }, { name: "LTP1", fitType: "segmented" },
  { name: "LTP2", fitType: "segmented" }, { name: "IAT", fitType: "poly3" },
  { name: "LTratio", fitType: "poly3" }, { name: "D2Lmax", fitType: "poly4" },
  { name: "MAX", fitType: "none" },
];
const profileOptions = [
  { name: "Laufen Leistungssportler (6×/Woche)", sport: "Running", calibrated: true },
  { name: "Laufen Ambitioniert (4–5×/Woche)", sport: "Running", calibrated: false },
  { name: "Laufen Freizeit (3×/Woche)", sport: "Running", calibrated: false },
];

const wait = <T>(v: T): Promise<T> => new Promise((r) => setTimeout(() => r(v), 60));

export const mockApp = {
  ListAthletes: (search: string) =>
    wait(athletes.filter((a) => a.name.toLowerCase().includes(search.toLowerCase()))),
  GetAthlete: (id: number) => wait(athleteFull[id] ?? athleteFull[1]),
  SaveAthlete: (_a: any) => wait(1),
  DeleteAthlete: (_id: number) => wait(undefined),
  ListTests: (athleteId: number) => wait(tests.filter((t) => t.athleteId === athleteId)),
  GetTest: (_id: number) => wait(tests[0]),
  SaveTest: (_t: any) => wait(10),
  DeleteTest: (_id: number) => wait(undefined),
  GetSteps: (_id: number) => wait(steps),
  SaveSteps: (_id: number, _s: any) => wait(undefined),
  ListTemplates: () => wait(templates),
  SaveTemplate: (_t: any) => wait(99),
  DeleteTemplate: (_id: number) => wait(undefined),
  ListProfiles: (_sport: string) => wait([]),
  Analyze: (_id: number) => wait(analysisDTO),
  AnalyzeWith: (_id: number, cfg: any) => {
    cfgState = cfg;
    return wait(analysisDTO);
  },
  GetAnalysisConfig: (_id: number) => wait({ ...cfgState }),
  ResetAnalysisConfig: (_id: number) => {
    cfgState = { ...defaultConfig };
    return wait(analysisDTO);
  },
  GetMarkerOptions: () => wait(markerOptions),
  GetProfileOptions: (_sport: string) => wait(profileOptions),
  RecomputeZones: (_id: number, _lt1: number, lt2: number) =>
    wait({ ...analysisDTO, lt2: { ...(analysisDTO as any).lt2, intensity: lt2, manual: true } }),
  ImportCSV: (_id: number, _t: string) => wait({ imported: 9, skipped: 0, errors: [] }),
  ParsePaste: (_t: string) => wait(steps),
  ExportCSV: (_id: number) => wait("intensity,time,hr,lactate,rpe\n6,03:00,98,1.24,6\n"),
  BackupDatabase: () => wait("/Users/demo/turnpoint-backup.db"),
  RestoreDatabase: () => wait(""),
};
