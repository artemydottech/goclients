package repository

import (
	"context"
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
