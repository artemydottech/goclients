package repository

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationsTableTracksVersions(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := ensureMigrationsTable(ctx, db); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := ensureMigrationsTable(ctx, db); err != nil {
		t.Fatalf("ensure must be idempotent: %v", err)
	}

	if err := markApplied(ctx, db, 900, "test migration"); err != nil {
		t.Fatalf("markApplied: %v", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		t.Fatalf("appliedVersions: %v", err)
	}
	if !applied[900] {
		t.Errorf("version 900 missing from %v", applied)
	}
}

func TestMarkAppliedRejectsTheSameVersionTwice(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if err := ensureMigrationsTable(ctx, db); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	if err := markApplied(ctx, db, 901, "first"); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := markApplied(ctx, db, 901, "second"); err == nil {
		t.Fatal("expected the primary key to reject a duplicate version")
	}
}

func TestMigrateRecordsEveryVersion(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		t.Fatalf("appliedVersions: %v", err)
	}

	for _, m := range migrations {
		if !applied[m.version] {
			t.Errorf("migration %d (%s) was not recorded", m.version, m.name)
		}
	}

	version, err := SchemaVersion(ctx, db)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if want := migrations[len(migrations)-1].version; version != want {
		t.Errorf("version = %d, want %d", version, want)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("recorded %d migrations, want %d", count, len(migrations))
	}
}

func TestSchemaVersionOfAnEmptyDatabase(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := ensureMigrationsTable(context.Background(), db); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	version, err := SchemaVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if version != 0 {
		t.Errorf("version = %d, want 0 before any migration", version)
	}
}
