-- 0003_analysis_config.sql — persisted per-test analysis configuration (FR-D2,
-- FR-F2, FR-Z2/Z3/Z5): enabled methods, parameters, display fit, anchors, profile.
CREATE TABLE analysis_config (
    test_id     INTEGER PRIMARY KEY REFERENCES test(id) ON DELETE CASCADE,
    config_json TEXT    NOT NULL,
    updated_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
