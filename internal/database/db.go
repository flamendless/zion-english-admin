package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS processing_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		google_drive_url TEXT NOT NULL,
		name TEXT NOT NULL,
		template TEXT,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		excluded_rows TEXT,
		useragent TEXT,
		output_path TEXT,
		errors TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		return err
	}

	DB = db
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
