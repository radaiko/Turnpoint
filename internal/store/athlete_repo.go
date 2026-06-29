package store

import "context"

// AthleteRepo is the athlete CRUD repository (FR-A1..A5).
type AthleteRepo struct{ db *DB }

func (db *DB) Athletes() *AthleteRepo { return &AthleteRepo{db} }

// Create inserts an athlete and returns its id.
func (r *AthleteRepo) Create(ctx context.Context, a Athlete) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO athlete (name, dob, sex, body_mass_kg, primary_sport, notes)
		 VALUES (?,?,?,?,?,?)`,
		a.Name, a.DOB, defSex(a.Sex), a.BodyMassKg, a.PrimarySport, a.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update modifies an athlete (does not touch existing tests' snapshots, OI-3).
func (r *AthleteRepo) Update(ctx context.Context, a Athlete) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE athlete SET name=?, dob=?, sex=?, body_mass_kg=?, primary_sport=?, notes=?,
		   updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
		 WHERE id=?`,
		a.Name, a.DOB, defSex(a.Sex), a.BodyMassKg, a.PrimarySport, a.Notes, a.ID)
	return err
}

// Delete removes an athlete; FK cascade removes their tests/steps/results/zones (OI-4).
func (r *AthleteRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM athlete WHERE id=?`, id)
	return err
}

// Get returns one athlete.
func (r *AthleteRepo) Get(ctx context.Context, id int64) (Athlete, error) {
	var a Athlete
	err := r.db.QueryRowContext(ctx,
		`SELECT id,name,dob,sex,body_mass_kg,primary_sport,notes,created_at,updated_at
		 FROM athlete WHERE id=?`, id).
		Scan(&a.ID, &a.Name, &a.DOB, &a.Sex, &a.BodyMassKg, &a.PrimarySport, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

// List returns athlete summaries, optionally filtered by a case-insensitive name
// or notes substring (FR-A4, OI-5).
func (r *AthleteRepo) List(ctx context.Context, search string) ([]AthleteSummary, error) {
	like := "%" + search + "%"
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.name, a.primary_sport,
		        (SELECT MAX(t.test_date) FROM test t WHERE t.athlete_id=a.id) AS last_test,
		        (SELECT COUNT(*) FROM test t WHERE t.athlete_id=a.id) AS test_count
		 FROM athlete a
		 WHERE (?='' OR a.name LIKE ? COLLATE NOCASE OR a.notes LIKE ? COLLATE NOCASE)
		 ORDER BY a.name COLLATE NOCASE`,
		search, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AthleteSummary
	for rows.Next() {
		var s AthleteSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.PrimarySport, &s.LastTestDate, &s.TestCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func defSex(s string) string {
	if s == "" {
		return "unspecified"
	}
	return s
}
