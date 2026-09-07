package repository

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT
);

CREATE TABLE IF NOT EXISTS companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    address TEXT,
    geolocation TEXT,
    schedule TEXT,
    site TEXT,
    socials TEXT,
    logo TEXT
);

CREATE TABLE IF NOT EXISTS employees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    position TEXT,
    avatar TEXT
);
`

// Open connects to the database and checks that it answers — sql.Open alone
// only parses the path, so a broken one surfaces at the first request instead.
func Open(path string) (*sql.DB, error) {
	// SQLite ignores foreign keys unless the pragma is set, and it has to ride
	// on the DSN: database/sql pools connections, so a one-off PRAGMA would
	// only cover whichever connection happened to run it.
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", path))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping %s: %w", path, err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}

	return nil
}
