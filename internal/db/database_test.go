package db

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return db
}

func TestCheckIntegrity_Passes(t *testing.T) {
	db := openTestDB(t)
	err := CheckIntegrity(db)
	if err != nil {
		t.Errorf("CheckIntegrity should pass on fresh DB: %v", err)
	}
}

func TestCheckpointWAL_InMemory(t *testing.T) {
	db := openTestDB(t)
	// WAL checkpoint on in-memory DB should not error
	err := CheckpointWAL(db)
	if err != nil {
		t.Errorf("CheckpointWAL should not error on in-memory DB: %v", err)
	}
}

func TestCloseDatabase(t *testing.T) {
	db := openTestDB(t)
	err := CloseDatabase(db)
	if err != nil {
		t.Errorf("CloseDatabase error: %v", err)
	}

	// After close, operations should fail
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err == nil {
		t.Error("expected ping to fail after CloseDatabase")
	}
}

func TestSeedTreatments(t *testing.T) {
	db := openTestDB(t)

	// Create treatments table
	db.Exec(`CREATE TABLE IF NOT EXISTS treatments (
		id TEXT PRIMARY KEY,
		name TEXT,
		code TEXT,
		default_price INTEGER,
		category TEXT,
		description TEXT,
		is_active INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	err := SeedTreatments(db)
	if err != nil {
		t.Fatalf("SeedTreatments error: %v", err)
	}

	// Verify treatments were seeded
	var count int64
	db.Raw("SELECT COUNT(*) FROM treatments").Scan(&count)
	if count == 0 {
		t.Error("expected treatments to be seeded, got 0")
	}

	// Should be idempotent - calling again should not error or duplicate
	err = SeedTreatments(db)
	if err != nil {
		t.Fatalf("SeedTreatments (second call) error: %v", err)
	}

	var count2 int64
	db.Raw("SELECT COUNT(*) FROM treatments").Scan(&count2)
	if count2 != count {
		t.Errorf("SeedTreatments should be idempotent: first=%d, second=%d", count, count2)
	}
}
