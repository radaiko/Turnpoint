// Package backup provides single-file backup and restore of the SQLite database
// (FR-M3). Backup is online-consistent via VACUUM INTO; restore is a controlled
// file swap after an integrity check.
package backup

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Backup writes a consistent, defragmented copy of the live database to destPath
// (which must not already exist). Safe while the app holds the DB open (WAL).
func Backup(db *sql.DB, destPath string) error {
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("backup: destination %q already exists", destPath)
	}
	_, err := db.Exec(`VACUUM INTO ?`, destPath)
	return err
}

// Restore replaces the database at dbPath with a copy of the file at srcPath after
// an integrity check, and drops stale WAL/SHM sidecars. The caller must have
// closed the live *sql.DB first and must reopen dbPath afterwards. srcPath is
// copied (not moved) so the user's backup file is preserved.
func Restore(dbPath, srcPath string) error {
	if err := verifyIntegrity(srcPath); err != nil {
		return err
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dbPath, data, 0o644); err != nil {
		return err
	}
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")
	return nil
}

func verifyIntegrity(path string) error {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	var result string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&result); err != nil {
		return fmt.Errorf("backup: integrity check failed: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("backup: integrity check returned %q", result)
	}
	return nil
}
