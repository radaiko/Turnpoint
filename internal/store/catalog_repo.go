package store

import (
	"context"
	"errors"
)

// TemplateRepo, ProfileRepo and ReportRepo cover templates, training profiles and
// report settings (FR-T7/T8, FR-Z5, FR-R3/R4).
type TemplateRepo struct{ db *DB }
type ProfileRepo struct{ db *DB }
type ReportRepo struct{ db *DB }

func (db *DB) Templates() *TemplateRepo { return &TemplateRepo{db} }
func (db *DB) Profiles() *ProfileRepo   { return &ProfileRepo{db} }
func (db *DB) Reports() *ReportRepo     { return &ReportRepo{db} }

// ErrPredefinedReadOnly is returned when editing/deleting a shipped predefined row.
var ErrPredefinedReadOnly = errors.New("store: predefined entries are read-only")

func (r *TemplateRepo) List(ctx context.Context) ([]Template, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,sport,step_duration_s,increment,start_intensity,end_intensity,mode,
		        rest_duration_s,visible_columns,is_predefined
		 FROM template ORDER BY is_predefined DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Template
	for rows.Next() {
		var t Template
		var pre int
		if err := rows.Scan(&t.ID, &t.Name, &t.Sport, &t.StepDurationS, &t.Increment, &t.StartIntensity,
			&t.EndIntensity, &t.Mode, &t.RestDurationS, &t.VisibleColumns, &pre); err != nil {
			return nil, err
		}
		t.IsPredefined = pre == 1
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TemplateRepo) Create(ctx context.Context, t Template) (int64, error) {
	vc := t.VisibleColumns
	if vc == "" {
		vc = `["intensity","time","hr","lactate","rpe"]`
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO template (name,sport,step_duration_s,increment,start_intensity,end_intensity,mode,
		   rest_duration_s,visible_columns,is_predefined) VALUES (?,?,?,?,?,?,?,?,?,0)`,
		t.Name, t.Sport, t.StepDurationS, t.Increment, t.StartIntensity, t.EndIntensity, defMode(t.Mode),
		t.RestDurationS, vc)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TemplateRepo) Delete(ctx context.Context, id int64) error {
	var pre int
	if err := r.db.QueryRowContext(ctx, `SELECT is_predefined FROM template WHERE id=?`, id).Scan(&pre); err != nil {
		return err
	}
	if pre == 1 {
		return ErrPredefinedReadOnly
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM template WHERE id=?`, id)
	return err
}

func (r *ProfileRepo) List(ctx context.Context, sport string) ([]TrainingProfile, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,sport,level,weekly_frequency,spread_json,is_predefined
		 FROM training_profile WHERE (?='' OR sport=?) ORDER BY is_predefined DESC, name`, sport, sport)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TrainingProfile
	for rows.Next() {
		var p TrainingProfile
		var pre int
		if err := rows.Scan(&p.ID, &p.Name, &p.Sport, &p.Level, &p.WeeklyFrequency, &p.SpreadJSON, &pre); err != nil {
			return nil, err
		}
		p.IsPredefined = pre == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProfileRepo) Create(ctx context.Context, p TrainingProfile) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO training_profile (name,sport,level,weekly_frequency,spread_json,is_predefined)
		 VALUES (?,?,?,?,?,0)`, p.Name, p.Sport, p.Level, p.WeeklyFrequency, p.SpreadJSON)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetForTest returns the per-test report settings, falling back to the global
// default when none exist for the test.
func (r *ReportRepo) GetForTest(ctx context.Context, testID int64) (ReportSettings, error) {
	rs, err := r.scan(ctx, `SELECT id,test_id,header_logo,header_text,footer_text,page_size,orientation,
		block_config_json,commentary,show_page_numbers FROM report_settings WHERE test_id=?`, testID)
	if err == nil {
		return rs, nil
	}
	return r.GetGlobal(ctx)
}

// GetGlobal returns the global default report settings (test_id IS NULL).
func (r *ReportRepo) GetGlobal(ctx context.Context) (ReportSettings, error) {
	return r.scan(ctx, `SELECT id,test_id,header_logo,header_text,footer_text,page_size,orientation,
		block_config_json,commentary,show_page_numbers FROM report_settings WHERE test_id IS NULL`)
}

func (r *ReportRepo) scan(ctx context.Context, q string, args ...any) (ReportSettings, error) {
	var rs ReportSettings
	var spn int
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&rs.ID, &rs.TestID, &rs.HeaderLogo, &rs.HeaderText,
		&rs.FooterText, &rs.PageSize, &rs.Orientation, &rs.BlockConfigJSON, &rs.Commentary, &spn)
	rs.ShowPageNumbers = spn == 1
	return rs, err
}

// UpsertForTest creates or updates the per-test report settings.
func (r *ReportRepo) UpsertForTest(ctx context.Context, rs ReportSettings) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO report_settings (test_id,header_logo,header_text,footer_text,page_size,orientation,
		   block_config_json,commentary,show_page_numbers)
		 VALUES (?,?,?,?,?,?,?,?,?)
		 ON CONFLICT(test_id) WHERE test_id IS NOT NULL DO UPDATE SET header_logo=excluded.header_logo, header_text=excluded.header_text,
		   footer_text=excluded.footer_text, page_size=excluded.page_size, orientation=excluded.orientation,
		   block_config_json=excluded.block_config_json, commentary=excluded.commentary,
		   show_page_numbers=excluded.show_page_numbers,
		   updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')`,
		rs.TestID, rs.HeaderLogo, rs.HeaderText, rs.FooterText, defStr(rs.PageSize, "A4"),
		defStr(rs.Orientation, "portrait"), rs.BlockConfigJSON, rs.Commentary, b2i(rs.ShowPageNumbers))
	return err
}

func defStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
