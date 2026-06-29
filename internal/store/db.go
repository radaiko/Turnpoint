// Package store is the local SQLite persistence layer (FR-M1). It uses the pure-Go
// modernc.org/sqlite driver (cgo-free) under driver name "sqlite". This package is
// app-only and must never be imported by core/ (enforced by the deps-purity guard).
package store

import (
	"database/sql"

	_ "modernc.org/sqlite" // registers driver "sqlite"
)

// DB wraps the configured connection pool.
type DB struct {
	*sql.DB
	path string
}

// Open opens (creating if needed) the database at path, applies the DSN pragmas
// to every pooled connection, caps the pool to a single writer, and migrates to
// the latest schema.
func Open(path string) (*DB, error) {
	dsn := "file:" + path +
		"?_pragma=foreign_keys(1)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)"
	sdb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sdb.SetMaxOpenConns(1) // SQLite has a single writer; avoid SQLITE_BUSY storms
	if err := sdb.Ping(); err != nil {
		sdb.Close()
		return nil, err
	}
	db := &DB{DB: sdb, path: path}
	if err := Migrate(sdb); err != nil {
		sdb.Close()
		return nil, err
	}
	return db, nil
}

// Path returns the database file path.
func (db *DB) Path() string { return db.path }
