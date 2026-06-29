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

// Restore replaces the database at dbPath with the file at srcPath after an
// integrity check, drops stale WAL/SHM sidecars, and reopens via reopen. The
// caller must have closed the live *sql.DB first.
func Restore(dbPath, srcPath string, reopen func(path string) (*sql.DB, error)) (*sql.DB, error) {
	if err := verifyIntegrity(srcPath); err != nil {
		return nil, err
	}
	if err := os.Rename(srcPath, dbPath); err != nil {
		return nil, err
	}
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")
	return reopen(dbPath)
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
