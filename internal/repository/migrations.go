package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL
)`

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, migrationsTable); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	return nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("reading schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("reading schema_migrations: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func markApplied(ctx context.Context, db *sql.DB, version int, name string) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		version, name, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("recording migration %d: %w", version, err)
	}

	return nil
}

type migration struct {
	version int
	name    string
	apply   func(ctx context.Context, db *sql.DB) error
}

var migrations = []migration{
	{1, "initial schema", func(ctx context.Context, db *sql.DB) error {
		if _, err := db.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("applying schema: %w", err)
		}

		return nil
	}},
	{2, "add user profile columns", func(ctx context.Context, db *sql.DB) error {
		for _, column := range []string{"surname", "username", "avatar"} {
			if err := addColumn(ctx, db, "users", column, "TEXT NOT NULL DEFAULT ''"); err != nil {
				return err
			}
		}

		return nil
	}},
	{3, "add company timezone", func(ctx context.Context, db *sql.DB) error {
		return addColumn(ctx, db, "companies", "timezone", "TEXT NOT NULL DEFAULT 'UTC'")
	}},
	{4, "add schedule breaks", func(ctx context.Context, db *sql.DB) error {
		for _, column := range []string{"break_starts_at", "break_ends_at"} {
			if err := addColumn(ctx, db, "employee_schedules", column, "TEXT NOT NULL DEFAULT ''"); err != nil {
				return err
			}
		}

		return nil
	}},
	{5, "add appointment price", func(ctx context.Context, db *sql.DB) error {
		return addColumn(ctx, db, "appointments", "price", "REAL NOT NULL DEFAULT 0")
	}},
	{6, "add appointment reminded_at", func(ctx context.Context, db *sql.DB) error {
		return addColumn(ctx, db, "appointments", "reminded_at", "TEXT NOT NULL DEFAULT ''")
	}},
	{7, "index employees by company", func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_employees_company ON employees(company_id, id)")
		return err
	}},
	{8, "index company appointment searches", func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_appointments_company_time ON appointments(company_id, starts_at, id);
 CREATE INDEX IF NOT EXISTS idx_appointments_company_status ON appointments(company_id, status, starts_at);`)
		return err
	}},
	{9, "persist idempotent booking requests", func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS booking_requests (
 client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
 request_key TEXT NOT NULL,fingerprint TEXT NOT NULL,appointment_id INTEGER NOT NULL DEFAULT 0,
 PRIMARY KEY(client_id,request_key));`)
		return err
	}},
	{10, "store client tags", func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS client_tags (
 client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,tag TEXT NOT NULL,
 PRIMARY KEY(client_id,tag));CREATE INDEX IF NOT EXISTS idx_client_tags_tag ON client_tags(tag,client_id);`)
		return err
	}},
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		if err := m.apply(ctx, db); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.version, m.name, err)
		}

		if err := markApplied(ctx, db, m.version, m.name); err != nil {
			return err
		}
	}

	return nil
}

func SchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}

	return int(version.Int64), nil
}
