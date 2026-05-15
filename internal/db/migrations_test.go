package db

import (
	"testing"
)

func TestMigrations_OrderedAndUnique(t *testing.T) {
	allMigrations := migrations()

	if len(allMigrations) == 0 {
		t.Fatal("expected at least one migration")
	}

	seen := make(map[int]bool)
	prevVersion := 0

	for _, m := range allMigrations {
		if m.Version <= 0 {
			t.Errorf("migration version must be positive, got %d", m.Version)
		}
		if seen[m.Version] {
			t.Errorf("duplicate migration version: %d", m.Version)
		}
		if m.Version <= prevVersion {
			t.Errorf("migrations must be in ascending order: version %d after %d", m.Version, prevVersion)
		}
		if m.Description == "" {
			t.Errorf("migration %d has no description", m.Version)
		}
		if m.Up == nil {
			t.Errorf("migration %d has nil Up function", m.Version)
		}

		seen[m.Version] = true
		prevVersion = m.Version
	}
}

func TestMigrationVersion_FirstVersion(t *testing.T) {
	allMigrations := migrations()
	if allMigrations[0].Version != 1 {
		t.Errorf("first migration should be version 1, got %d", allMigrations[0].Version)
	}
}
