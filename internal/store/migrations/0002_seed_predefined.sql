-- 0002_seed_predefined.sql — Appendix B templates, OI-17 profiles, global report defaults.

-- Predefined entry templates (Appendix B; all values editable per test).
INSERT INTO template (name, sport, step_duration_s, increment, start_intensity, end_intensity, mode, is_predefined)
VALUES
    ('Running (Lauf)',  'running', 180, 2,  6,  22,  'continuous', 1),
    ('Cycling (Rad)',   'cycling', 240, 40, 80, 440, 'continuous', 1);

-- Predefined training profiles (OI-17). Only the Leistung running profile is calibrated.
INSERT INTO training_profile (name, sport, level, weekly_frequency, spread_json, is_predefined)
VALUES
    ('Laufen Leistungssportler (6×/Woche)', 'running', 'leistung', 6,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1),
    ('Laufen Ambitioniert (4–5×/Woche)', 'running', 'ambitioniert', 5,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1),
    ('Laufen Freizeit (3×/Woche)', 'running', 'freizeit', 3,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1),
    ('Rad Leistungssportler (6×/Woche)', 'cycling', 'leistung', 6,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1),
    ('Rad Ambitioniert (4–5×/Woche)', 'cycling', 'ambitioniert', 5,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1),
    ('Rad Freizeit (3×/Woche)', 'cycling', 'freizeit', 3,
     '{"REKOM":[0,0.46],"GA1":[0.46,0.70],"GA2":[0.70,0.88],"EB":[0.88,1.02],"SB":[1.02,1.25]}', 1);

-- Global default report settings (test_id NULL): reference report pages 2 & 3 on by default.
INSERT INTO report_settings (test_id, block_config_json)
VALUES (NULL,
    '[{"block":"cover","visible":false},{"block":"remarks","visible":false},{"block":"raw_table","visible":true},{"block":"temporal_chart","visible":true},{"block":"threshold_table","visible":true},{"block":"zones","visible":true},{"block":"evaluation","visible":false}]');
