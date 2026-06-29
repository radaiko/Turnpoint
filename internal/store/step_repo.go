package store

import "context"

// StepRepo persists test steps (FR-T2).
type StepRepo struct{ db *DB }

func (db *DB) Steps() *StepRepo { return &StepRepo{db} }

// ReplaceAll replaces a test's entire step grid in one transaction (the canonical
// full-grid save).
func (r *StepRepo) ReplaceAll(ctx context.Context, testID int64, steps []Step) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM step WHERE test_id=?`, testID); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO step (test_id, step_order, intensity, time_point_s, heart_rate, lactate, rpe,
		   is_baseline, excluded, aborted) VALUES (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i, s := range steps {
		order := s.StepOrder
		if order == 0 && i != 0 {
			order = i
		}
		if _, err := stmt.ExecContext(ctx, testID, order, s.Intensity, s.TimePointS, s.HeartRate,
			s.Lactate, s.RPE, b2i(s.IsBaseline), b2i(s.Excluded), b2i(s.Aborted)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListByTest returns a test's steps in order.
func (r *StepRepo) ListByTest(ctx context.Context, testID int64) ([]Step, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, test_id, step_order, intensity, time_point_s, heart_rate, lactate, rpe,
		        is_baseline, excluded, aborted
		 FROM step WHERE test_id=? ORDER BY step_order`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Step{}
	for rows.Next() {
		var s Step
		var base, excl, abort int
		if err := rows.Scan(&s.ID, &s.TestID, &s.StepOrder, &s.Intensity, &s.TimePointS, &s.HeartRate,
			&s.Lactate, &s.RPE, &base, &excl, &abort); err != nil {
			return nil, err
		}
		s.IsBaseline, s.Excluded, s.Aborted = base == 1, excl == 1, abort == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
