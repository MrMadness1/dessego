package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// SQLite driver.
	_ "github.com/mattn/go-sqlite3"
)

// NewSQLite returns a new SQLite database.
func NewSQLite(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("create DB directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open DB: %w", err)
	}

	return db, nil
}
