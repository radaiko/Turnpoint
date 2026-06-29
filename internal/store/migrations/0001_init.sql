-- 0001_init.sql — Turnpoint schema (SRS §9, DESIGN Appendix C).
PRAGMA foreign_keys = ON;

-- ---- Athlete (FR-A5, OI-2/6/7/8) ----
CREATE TABLE athlete (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,
    dob           TEXT,
    sex           TEXT    NOT NULL DEFAULT 'unspecified'
                          CHECK (sex IN ('male','female','unspecified')),
    body_mass_kg  REAL    CHECK (body_mass_kg IS NULL OR body_mass_kg BETWEEN 20.0 AND 250.0),
    primary_sport TEXT    CHECK (primary_sport IS NULL OR primary_sport IN ('running','cycling')),
    notes         TEXT    NOT NULL DEFAULT '',
    created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_athlete_name ON athlete(name COLLATE NOCASE);

-- ---- Template (FR-T7/T8) ----
CREATE TABLE template (
    id              INTEGER PRIMARY KEY,
    name            TEXT    NOT NULL,
    sport           TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    step_duration_s INTEGER NOT NULL CHECK (step_duration_s > 0),
    increment       REAL    NOT NULL CHECK (increment > 0),
    start_intensity REAL    NOT NULL CHECK (start_intensity >= 0),
    end_intensity   REAL    CHECK (end_intensity IS NULL OR end_intensity >= start_intensity),
    mode            TEXT    NOT NULL DEFAULT 'continuous' CHECK (mode IN ('continuous','intermittent')),
    rest_duration_s INTEGER CHECK (rest_duration_s IS NULL OR rest_duration_s > 0),
    visible_columns TEXT    NOT NULL DEFAULT '["intensity","time","hr","lactate","rpe"]',
    is_predefined   INTEGER NOT NULL DEFAULT 0 CHECK (is_predefined IN (0,1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_template_name ON template(name);

-- ---- TrainingProfile (FR-Z5, OI-17) ----
CREATE TABLE training_profile (
    id               INTEGER PRIMARY KEY,
    name             TEXT    NOT NULL,
    sport            TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    level            TEXT    NOT NULL CHECK (level IN ('freizeit','ambitioniert','leistung')),
    weekly_frequency INTEGER CHECK (weekly_frequency IS NULL OR weekly_frequency > 0),
    spread_json      TEXT    NOT NULL,
    is_predefined    INTEGER NOT NULL DEFAULT 0 CHECK (is_predefined IN (0,1)),
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_profile_name ON training_profile(name);

-- ---- Test (FR-T1/T3, OI-3) ----
CREATE TABLE test (
    id                 INTEGER PRIMARY KEY,
    athlete_id         INTEGER NOT NULL
                       REFERENCES athlete(id) ON DELETE CASCADE ON UPDATE CASCADE,
    test_date          TEXT    NOT NULL,
    sport              TEXT    NOT NULL CHECK (sport IN ('running','cycling')),
    step_duration_s    INTEGER NOT NULL CHECK (step_duration_s > 0),
    increment          REAL    NOT NULL CHECK (increment > 0),
    start_intensity    REAL    NOT NULL CHECK (start_intensity >= 0),
    mode               TEXT    NOT NULL DEFAULT 'continuous' CHECK (mode IN ('continuous','intermittent')),
    rest_duration_s    INTEGER CHECK (
                           (mode='intermittent' AND rest_duration_s IS NOT NULL AND rest_duration_s > 0)
                        OR (mode='continuous'   AND rest_duration_s IS NULL)),
    baseline_lactate   REAL    CHECK (baseline_lactate IS NULL OR baseline_lactate BETWEEN 0 AND 30),
    body_mass_snapshot REAL    CHECK (body_mass_snapshot IS NULL OR body_mass_snapshot BETWEEN 20.0 AND 250.0),
    pretest_note       TEXT    NOT NULL DEFAULT '',
    remarks            TEXT    NOT NULL DEFAULT '',
    template_id        INTEGER REFERENCES template(id) ON DELETE SET NULL,
    created_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_test_athlete ON test(athlete_id, test_date);

-- ---- Step (FR-T2/T3/T6, OI-12/13) ----
CREATE TABLE step (
    id           INTEGER PRIMARY KEY,
    test_id      INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    step_order   INTEGER NOT NULL,
    intensity    REAL    NOT NULL CHECK (intensity >= 0),
    time_point_s INTEGER CHECK (time_point_s IS NULL OR time_point_s >= 0),
    heart_rate   INTEGER CHECK (heart_rate IS NULL OR heart_rate BETWEEN 0 AND 250),
    lactate      REAL    CHECK (lactate IS NULL OR lactate BETWEEN 0 AND 30),
    rpe          INTEGER CHECK (rpe IS NULL OR rpe BETWEEN 6 AND 20),
    is_baseline  INTEGER NOT NULL DEFAULT 0 CHECK (is_baseline IN (0,1)),
    excluded     INTEGER NOT NULL DEFAULT 0 CHECK (excluded IN (0,1)),
    aborted      INTEGER NOT NULL DEFAULT 0 CHECK (aborted IN (0,1)),
    UNIQUE (test_id, step_order)
);
CREATE INDEX idx_step_test ON step(test_id, step_order);
CREATE UNIQUE INDEX uq_step_one_baseline ON step(test_id) WHERE is_baseline = 1;

-- ---- ThresholdResult (FR-D1..D5) snapshot rows ----
CREATE TABLE threshold_result (
    id                    INTEGER PRIMARY KEY,
    test_id               INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    method                TEXT    NOT NULL,
    intensity             REAL,
    lactate               REAL,
    heart_rate            REAL,
    pct_max               REAL,
    pace_s_per_km         REAL,
    kcal_h                REAL,
    is_override           INTEGER NOT NULL DEFAULT 0 CHECK (is_override IN (0,1)),
    fit_type              TEXT    NOT NULL CHECK (fit_type IN
                              ('poly3','poly4','exp','spline','loglog','segmented','none')),
    not_computable_reason TEXT,
    params_json           TEXT,
    UNIQUE (test_id, method)
);
CREATE INDEX idx_tr_test ON threshold_result(test_id);

-- ---- Zone (FR-Z1/Z4, §7) snapshot rows ----
CREATE TABLE zone (
    id                INTEGER PRIMARY KEY,
    test_id           INTEGER NOT NULL REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    model             TEXT    NOT NULL DEFAULT '5zone',
    zone_index        INTEGER NOT NULL CHECK (zone_index BETWEEN 1 AND 5),
    zone_name         TEXT    NOT NULL CHECK (zone_name IN ('REKOM','GA1','GA2','EB','SB')),
    profile_id        INTEGER REFERENCES training_profile(id) ON DELETE SET NULL,
    intensity_low     REAL, intensity_high     REAL,
    hr_low            REAL, hr_high            REAL,
    lactate_low       REAL, lactate_high       REAL,
    pace_low_s_per_km REAL, pace_high_s_per_km REAL,
    UNIQUE (test_id, model, zone_index)
);
CREATE INDEX idx_zone_test ON zone(test_id);

-- ---- ReportSettings (FR-R3/R4, OI-20) global default (test_id NULL) + per-test ----
CREATE TABLE report_settings (
    id                INTEGER PRIMARY KEY,
    test_id           INTEGER REFERENCES test(id) ON DELETE CASCADE,
    header_logo       BLOB,
    header_text       TEXT    NOT NULL DEFAULT '',
    footer_text       TEXT    NOT NULL DEFAULT '',
    page_size         TEXT    NOT NULL DEFAULT 'A4' CHECK (page_size IN ('A4','Letter')),
    orientation       TEXT    NOT NULL DEFAULT 'portrait' CHECK (orientation IN ('portrait','landscape')),
    block_config_json TEXT    NOT NULL,
    commentary        TEXT    NOT NULL DEFAULT '',
    show_page_numbers INTEGER NOT NULL DEFAULT 1 CHECK (show_page_numbers IN (0,1)),
    updated_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX uq_report_global ON report_settings(test_id) WHERE test_id IS NULL;
CREATE UNIQUE INDEX uq_report_test   ON report_settings(test_id) WHERE test_id IS NOT NULL;
