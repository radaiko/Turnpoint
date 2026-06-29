package backup

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/radaiko/turnpoint/internal/store"
)

func TestBackupRestore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "main.db")
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := db.Athletes().Create(ctx, store.Athlete{Name: "Backup Subject", Sex: "unspecified"}); err != nil {
		t.Fatal(err)
	}

	bak := filepath.Join(dir, "backup.db")
	if err := Backup(db.DB, bak); err != nil {
		t.Fatalf("backup: %v", err)
	}
	// backing up to an existing path is refused
	if err := Backup(db.DB, bak); err == nil {
		t.Error("expected error backing up over an existing file")
	}
	db.Close()

	reopened, err := Restore(dbPath, bak, func(p string) (*sql.DB, error) {
		return sql.Open("sqlite", "file:"+p+"?_pragma=foreign_keys(1)")
	})
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	reopened.Close()

	// verify via the repository layer
	verify, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer verify.Close()
	list, _ := verify.Athletes().List(ctx, "")
	if len(list) != 1 || list[0].Name != "Backup Subject" {
		t.Errorf("restored data wrong: %+v", list)
	}
}
