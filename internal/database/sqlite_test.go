package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSQLiteCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dessego.db")
	db, err := NewSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("database directory was not created: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("database is not usable: %v", err)
	}
}
