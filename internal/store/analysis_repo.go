package store

import "context"

// ThresholdRepo and ZoneRepo persist the analysis snapshot rows (FR-D5, FR-Z4).
type ThresholdRepo struct{ db *DB }
type ZoneRepo struct{ db *DB }

func (db *DB) Thresholds() *ThresholdRepo { return &ThresholdRepo{db} }
func (db *DB) ZonesRepo() *ZoneRepo       { return &ZoneRepo{db} }

// ReplaceAll replaces a test's threshold result snapshot in one transaction.
func (r *ThresholdRepo) ReplaceAll(ctx context.Context, testID int64, rows []ThresholdResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM threshold_result WHERE test_id=?`, testID); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO threshold_result (test_id, method, intensity, lactate, heart_rate, pct_max,
		   pace_s_per_km, kcal_h, is_override, fit_type, not_computable_reason, params_json)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, t := range rows {
		if _, err := stmt.ExecContext(ctx, testID, t.Method, t.Intensity, t.Lactate, t.HeartRate, t.PctMax,
			t.PaceSPerKm, t.KcalH, b2i(t.IsOverride), t.FitType, t.NotComputableReason, t.ParamsJSON); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ThresholdRepo) ListByTest(ctx context.Context, testID int64) ([]ThresholdResult, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, test_id, method, intensity, lactate, heart_rate, pct_max, pace_s_per_km, kcal_h,
		        is_override, fit_type, not_computable_reason, params_json
		 FROM threshold_result WHERE test_id=? ORDER BY id`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ThresholdResult
	for rows.Next() {
		var t ThresholdResult
		var ovr int
		if err := rows.Scan(&t.ID, &t.TestID, &t.Method, &t.Intensity, &t.Lactate, &t.HeartRate, &t.PctMax,
			&t.PaceSPerKm, &t.KcalH, &ovr, &t.FitType, &t.NotComputableReason, &t.ParamsJSON); err != nil {
			return nil, err
		}
		t.IsOverride = ovr == 1
		out = append(out, t)
	}
	return out, rows.Err()
}

// ReplaceAll replaces a test's zone snapshot in one transaction.
func (r *ZoneRepo) ReplaceAll(ctx context.Context, testID int64, zones []Zone) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM zone WHERE test_id=?`, testID); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO zone (test_id, model, zone_index, zone_name, profile_id, intensity_low, intensity_high,
		   hr_low, hr_high, lactate_low, lactate_high, pace_low_s_per_km, pace_high_s_per_km)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, z := range zones {
		model := z.Model
		if model == "" {
			model = "5zone"
		}
		if _, err := stmt.ExecContext(ctx, testID, model, z.ZoneIndex, z.ZoneName, z.ProfileID,
			z.IntensityLow, z.IntensityHigh, z.HRLow, z.HRHigh, z.LactateLow, z.LactateHigh,
			z.PaceLowSPerKm, z.PaceHighSPerKm); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ZoneRepo) ListByTest(ctx context.Context, testID int64) ([]Zone, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, test_id, model, zone_index, zone_name, profile_id, intensity_low, intensity_high,
		        hr_low, hr_high, lactate_low, lactate_high, pace_low_s_per_km, pace_high_s_per_km
		 FROM zone WHERE test_id=? ORDER BY zone_index`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Zone
	for rows.Next() {
		var z Zone
		if err := rows.Scan(&z.ID, &z.TestID, &z.Model, &z.ZoneIndex, &z.ZoneName, &z.ProfileID,
			&z.IntensityLow, &z.IntensityHigh, &z.HRLow, &z.HRHigh, &z.LactateLow, &z.LactateHigh,
			&z.PaceLowSPerKm, &z.PaceHighSPerKm); err != nil {
			return nil, err
		}
		out = append(out, z)
	}
	return out, rows.Err()
}
