package store

import (
	"context"
	"database/sql"
)

// TestRepo is the test CRUD repository (FR-T1).
type TestRepo struct{ db *DB }

func (db *DB) Tests() *TestRepo { return &TestRepo{db} }

// Create inserts a test. If BodyMassSnapshot is nil it is copied from the athlete
// at creation time (OI-3), so later athlete edits don't change this test.
func (r *TestRepo) Create(ctx context.Context, t Test) (int64, error) {
	if t.BodyMassSnapshot == nil {
		var bm sql.NullFloat64
		if err := r.db.QueryRowContext(ctx, `SELECT body_mass_kg FROM athlete WHERE id=?`, t.AthleteID).Scan(&bm); err == nil && bm.Valid {
			v := bm.Float64
			t.BodyMassSnapshot = &v
		}
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO test (athlete_id, test_date, sport, step_duration_s, increment, start_intensity,
		   mode, rest_duration_s, baseline_lactate, body_mass_snapshot, pretest_note, remarks, template_id)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.AthleteID, t.TestDate, t.Sport, t.StepDurationS, t.Increment, t.StartIntensity,
		defMode(t.Mode), t.RestDurationS, t.BaselineLactate, t.BodyMassSnapshot, t.PretestNote, t.Remarks, t.TemplateID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TestRepo) Update(ctx context.Context, t Test) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE test SET test_date=?, sport=?, step_duration_s=?, increment=?, start_intensity=?,
		   mode=?, rest_duration_s=?, baseline_lactate=?, body_mass_snapshot=?, pretest_note=?, remarks=?,
		   updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
		 WHERE id=?`,
		t.TestDate, t.Sport, t.StepDurationS, t.Increment, t.StartIntensity,
		defMode(t.Mode), t.RestDurationS, t.BaselineLactate, t.BodyMassSnapshot, t.PretestNote, t.Remarks, t.ID)
	return err
}

func (r *TestRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM test WHERE id=?`, id)
	return err
}

func (r *TestRepo) Get(ctx context.Context, id int64) (Test, error) {
	var t Test
	err := r.db.QueryRowContext(ctx,
		`SELECT id, athlete_id, test_date, sport, step_duration_s, increment, start_intensity,
		        mode, rest_duration_s, baseline_lactate, body_mass_snapshot, pretest_note, remarks,
		        template_id, created_at, updated_at
		 FROM test WHERE id=?`, id).
		Scan(&t.ID, &t.AthleteID, &t.TestDate, &t.Sport, &t.StepDurationS, &t.Increment, &t.StartIntensity,
			&t.Mode, &t.RestDurationS, &t.BaselineLactate, &t.BodyMassSnapshot, &t.PretestNote, &t.Remarks,
			&t.TemplateID, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *TestRepo) ListByAthlete(ctx context.Context, athleteID int64) ([]Test, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, athlete_id, test_date, sport, step_duration_s, increment, start_intensity,
		        mode, rest_duration_s, baseline_lactate, body_mass_snapshot, pretest_note, remarks,
		        template_id, created_at, updated_at
		 FROM test WHERE athlete_id=? ORDER BY test_date DESC, id DESC`, athleteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Test{}
	for rows.Next() {
		var t Test
		if err := rows.Scan(&t.ID, &t.AthleteID, &t.TestDate, &t.Sport, &t.StepDurationS, &t.Increment, &t.StartIntensity,
			&t.Mode, &t.RestDurationS, &t.BaselineLactate, &t.BodyMassSnapshot, &t.PretestNote, &t.Remarks,
			&t.TemplateID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func defMode(m string) string {
	if m == "" {
		return "continuous"
	}
	return m
}
