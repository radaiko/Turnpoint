export namespace csvio {
	
	export class RowError {
	    line: number;
	    msg: string;
	
	    static createFrom(source: any = {}) {
	        return new RowError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.line = source["line"];
	        this.msg = source["msg"];
	    }
	}
	export class ImportReport {
	    imported: number;
	    skipped: number;
	    errors: RowError[];
	
	    static createFrom(source: any = {}) {
	        return new ImportReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.skipped = source["skipped"];
	        this.errors = this.convertValues(source["errors"], RowError);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace service {
	
	export class WarningDTO {
	    severity: string;
	    subject: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new WarningDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.subject = source["subject"];
	        this.message = source["message"];
	    }
	}
	export class ZoneDTO {
	    index: number;
	    label: string;
	    intensityLow: number;
	    intensityHigh: number;
	    hrLow: number;
	    hrHigh: number;
	    lactateLow: number;
	    lactateHigh: number;
	    paceLow: string;
	    paceHigh: string;
	
	    static createFrom(source: any = {}) {
	        return new ZoneDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.label = source["label"];
	        this.intensityLow = source["intensityLow"];
	        this.intensityHigh = source["intensityHigh"];
	        this.hrLow = source["hrLow"];
	        this.hrHigh = source["hrHigh"];
	        this.lactateLow = source["lactateLow"];
	        this.lactateHigh = source["lactateHigh"];
	        this.paceLow = source["paceLow"];
	        this.paceHigh = source["paceHigh"];
	    }
	}
	export class AnchorDTO {
	    marker: string;
	    intensity: number;
	    lactate: number;
	    heartRate: number;
	    pace: string;
	    pctMax: number;
	    manual: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AnchorDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marker = source["marker"];
	        this.intensity = source["intensity"];
	        this.lactate = source["lactate"];
	        this.heartRate = source["heartRate"];
	        this.pace = source["pace"];
	        this.pctMax = source["pctMax"];
	        this.manual = source["manual"];
	    }
	}
	export class MarkerRow {
	    marker: string;
	    intensity: number;
	    lactate: number;
	    heartRate: number;
	    pctMax: number;
	    pace: string;
	    kcalPerHr: number;
	    hasKcal: boolean;
	    fitType: string;
	    computable: boolean;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new MarkerRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marker = source["marker"];
	        this.intensity = source["intensity"];
	        this.lactate = source["lactate"];
	        this.heartRate = source["heartRate"];
	        this.pctMax = source["pctMax"];
	        this.pace = source["pace"];
	        this.kcalPerHr = source["kcalPerHr"];
	        this.hasKcal = source["hasKcal"];
	        this.fitType = source["fitType"];
	        this.computable = source["computable"];
	        this.reason = source["reason"];
	    }
	}
	export class StepBar {
	    startS: number;
	    endS: number;
	    intensity: number;
	
	    static createFrom(source: any = {}) {
	        return new StepBar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startS = source["startS"];
	        this.endS = source["endS"];
	        this.intensity = source["intensity"];
	    }
	}
	export class XY {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new XY(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class AnalysisDTO {
	    sport: string;
	    unit: string;
	    hasPace: boolean;
	    rawPoints: XY[];
	    curve: XY[];
	    hrPoints: XY[];
	    timeHR: XY[];
	    timeLactate: XY[];
	    stepBars: StepBar[];
	    markers: MarkerRow[];
	    lt1: AnchorDTO;
	    lt2: AnchorDTO;
	    zones: ZoneDTO[];
	    maxIntensity: number;
	    domainLow: number;
	    domainHigh: number;
	    warnings: WarningDTO[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sport = source["sport"];
	        this.unit = source["unit"];
	        this.hasPace = source["hasPace"];
	        this.rawPoints = this.convertValues(source["rawPoints"], XY);
	        this.curve = this.convertValues(source["curve"], XY);
	        this.hrPoints = this.convertValues(source["hrPoints"], XY);
	        this.timeHR = this.convertValues(source["timeHR"], XY);
	        this.timeLactate = this.convertValues(source["timeLactate"], XY);
	        this.stepBars = this.convertValues(source["stepBars"], StepBar);
	        this.markers = this.convertValues(source["markers"], MarkerRow);
	        this.lt1 = this.convertValues(source["lt1"], AnchorDTO);
	        this.lt2 = this.convertValues(source["lt2"], AnchorDTO);
	        this.zones = this.convertValues(source["zones"], ZoneDTO);
	        this.maxIntensity = source["maxIntensity"];
	        this.domainLow = source["domainLow"];
	        this.domainHigh = source["domainHigh"];
	        this.warnings = this.convertValues(source["warnings"], WarningDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	

}

export namespace store {
	
	export class Athlete {
	    id: number;
	    name: string;
	    dob?: string;
	    sex: string;
	    bodyMassKg?: number;
	    primarySport?: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Athlete(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.dob = source["dob"];
	        this.sex = source["sex"];
	        this.bodyMassKg = source["bodyMassKg"];
	        this.primarySport = source["primarySport"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class AthleteSummary {
	    id: number;
	    name: string;
	    primarySport?: string;
	    lastTestDate?: string;
	    testCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AthleteSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.primarySport = source["primarySport"];
	        this.lastTestDate = source["lastTestDate"];
	        this.testCount = source["testCount"];
	    }
	}
	export class Step {
	    id: number;
	    testId: number;
	    stepOrder: number;
	    intensity: number;
	    timePointS?: number;
	    heartRate?: number;
	    lactate?: number;
	    rpe?: number;
	    isBaseline: boolean;
	    excluded: boolean;
	    aborted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Step(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.testId = source["testId"];
	        this.stepOrder = source["stepOrder"];
	        this.intensity = source["intensity"];
	        this.timePointS = source["timePointS"];
	        this.heartRate = source["heartRate"];
	        this.lactate = source["lactate"];
	        this.rpe = source["rpe"];
	        this.isBaseline = source["isBaseline"];
	        this.excluded = source["excluded"];
	        this.aborted = source["aborted"];
	    }
	}
	export class Template {
	    id: number;
	    name: string;
	    sport: string;
	    stepDurationS: number;
	    increment: number;
	    startIntensity: number;
	    endIntensity?: number;
	    mode: string;
	    restDurationS?: number;
	    visibleColumns: string;
	    isPredefined: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sport = source["sport"];
	        this.stepDurationS = source["stepDurationS"];
	        this.increment = source["increment"];
	        this.startIntensity = source["startIntensity"];
	        this.endIntensity = source["endIntensity"];
	        this.mode = source["mode"];
	        this.restDurationS = source["restDurationS"];
	        this.visibleColumns = source["visibleColumns"];
	        this.isPredefined = source["isPredefined"];
	    }
	}
	export class Test {
	    id: number;
	    athleteId: number;
	    testDate: string;
	    sport: string;
	    stepDurationS: number;
	    increment: number;
	    startIntensity: number;
	    mode: string;
	    restDurationS?: number;
	    baselineLactate?: number;
	    bodyMassSnapshot?: number;
	    pretestNote: string;
	    remarks: string;
	    templateId?: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Test(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.athleteId = source["athleteId"];
	        this.testDate = source["testDate"];
	        this.sport = source["sport"];
	        this.stepDurationS = source["stepDurationS"];
	        this.increment = source["increment"];
	        this.startIntensity = source["startIntensity"];
	        this.mode = source["mode"];
	        this.restDurationS = source["restDurationS"];
	        this.baselineLactate = source["baselineLactate"];
	        this.bodyMassSnapshot = source["bodyMassSnapshot"];
	        this.pretestNote = source["pretestNote"];
	        this.remarks = source["remarks"];
	        this.templateId = source["templateId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TrainingProfile {
	    id: number;
	    name: string;
	    sport: string;
	    level: string;
	    weeklyFrequency?: number;
	    spreadJson: string;
	    isPredefined: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TrainingProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sport = source["sport"];
	        this.level = source["level"];
	        this.weeklyFrequency = source["weeklyFrequency"];
	        this.spreadJson = source["spreadJson"];
	        this.isPredefined = source["isPredefined"];
	    }
	}

}

