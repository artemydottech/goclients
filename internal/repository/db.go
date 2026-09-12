package repository

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    surname TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT ''
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

CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL,
    price REAL NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_services_company ON services(company_id);

CREATE TABLE IF NOT EXISTS employee_services (
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    PRIMARY KEY (employee_id, service_id)
);

CREATE INDEX IF NOT EXISTS idx_employee_services_service ON employee_services(service_id);

CREATE TABLE IF NOT EXISTS clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    UNIQUE (company_id, phone)
);

CREATE TABLE IF NOT EXISTS appointments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    starts_at TEXT NOT NULL,
    ends_at TEXT NOT NULL,
    status TEXT NOT NULL,
    comment TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_appointments_employee ON appointments(employee_id, starts_at);
CREATE INDEX IF NOT EXISTS idx_appointments_client ON appointments(client_id);

CREATE TABLE IF NOT EXISTS employee_schedules (
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    weekday INTEGER NOT NULL,
    starts_at TEXT NOT NULL,
    ends_at TEXT NOT NULL,
    PRIMARY KEY (employee_id, weekday)
);
`

// Open connects to the database and checks that it answers — sql.Open alone
// only parses the path, so a broken one surfaces at the first request instead.
func Open(path string) (*sql.DB, error) {
	// SQLite ignores foreign keys unless the pragma is set, and it has to ride
	// on the DSN: database/sql pools connections, so a one-off PRAGMA would
	// only cover whichever connection happened to run it.
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", path))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	// SQLite takes one writer at a time; without this the second concurrent
	// request fails outright with "database is locked".
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping %s: %w", path, err)
	}

	return db, nil
}

// Columns added to users after the first release. SQLite has no
// ADD COLUMN IF NOT EXISTS, so each one is checked before it is added.
var addedUserColumns = []string{"surname", "username", "avatar"}

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}

	for _, column := range addedUserColumns {
		exists, err := columnExists(db, "users", column)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		statement := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s TEXT NOT NULL DEFAULT ''", column)
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("adding users.%s: %w", column, err)
		}
	}

	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("reading %s columns: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return false, fmt.Errorf("reading %s columns: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}

	return false, rows.Err()
}
