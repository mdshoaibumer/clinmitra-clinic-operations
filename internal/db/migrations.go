package db

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"clinmitra/internal/models"

	"gorm.io/gorm"
)

// Migration represents a single versioned schema migration.
type Migration struct {
	Version     int
	Description string
	Up          func(tx *gorm.DB) error
}

// MigrationVersion tracks applied migrations in the database.
type MigrationVersion struct {
	Version   int    `gorm:"primaryKey"`
	AppliedAt int64  `gorm:"autoCreateTime"`
	Name      string `gorm:"type:text"`
}

// migrations returns all registered migrations in order.
func migrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "initial_schema",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&models.ClinicSettings{},
					&models.User{},
					&models.Patient{},
					&models.Treatment{},
					&models.Appointment{},
					&models.Invoice{},
					&models.InvoiceItem{},
					&models.Payment{},
					&models.PatientTreatment{},
					&models.AuditLog{},
				)
			},
		},
		{
			Version:     2,
			Description: "add_composite_indexes_for_performance",
			Up: func(tx *gorm.DB) error {
				// Composite index for appointment conflict detection:
				// covers queries on (appointment_date, status) and (start_time, end_time)
				indexes := []string{
					"CREATE INDEX IF NOT EXISTS idx_appt_date_status ON appointments(appointment_date, status)",
					"CREATE INDEX IF NOT EXISTS idx_appt_time_range ON appointments(start_time, end_time)",
					// Index for audit log queries by entity
					"CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at)",
				}
				for _, idx := range indexes {
					if err := tx.Exec(idx).Error; err != nil {
						return fmt.Errorf("failed to create index: %s: %w", idx, err)
					}
				}
				return nil
			},
		},
		{
			Version:     3,
			Description: "add_doctor_qualification_to_settings",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.ClinicSettings{})
			},
		},
		{
			Version:     4,
			Description: "add_cloud_backup_fields_to_settings",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.ClinicSettings{})
			},
		},
		{
			Version:     5,
			Description: "add_whatsapp_template_fields_to_settings",
			Up: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.ClinicSettings{})
			},
		},
	}
}

// RunMigrations applies all pending versioned migrations.
// Creates an automatic backup before applying any new migrations.
// Time complexity: O(m) where m is number of pending migrations
// Space complexity: O(n) where n is total migration count (for applied map)
func RunMigrations(db *gorm.DB) error {
	// Ensure migration tracking table exists
	if err := db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create migration_versions table: %w", err)
	}

	allMigrations := migrations()
	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	// Get already applied versions
	var applied []MigrationVersion
	if err := db.Find(&applied).Error; err != nil {
		return fmt.Errorf("failed to read migration history: %w", err)
	}
	appliedMap := make(map[int]bool, len(applied))
	for _, m := range applied {
		appliedMap[m.Version] = true
	}

	// Check if there are pending migrations
	var hasPending bool
	for _, m := range allMigrations {
		if !appliedMap[m.Version] {
			hasPending = true
			break
		}
	}

	// Create pre-migration backup if there are pending migrations
	if hasPending {
		if err := createPreMigrationBackup(db); err != nil {
			slog.Warn("pre-migration backup failed, proceeding with migration", "error", err)
		}
	}

	// Apply pending migrations in order
	for _, m := range allMigrations {
		if appliedMap[m.Version] {
			continue
		}

		slog.Info("applying migration", "version", m.Version, "description", m.Description)

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Description, err)
			}

			return tx.Create(&MigrationVersion{
				Version: m.Version,
				Name:    m.Description,
			}).Error
		})
		if err != nil {
			return err
		}

		slog.Info("migration applied", "version", m.Version)
	}

	return nil
}

// createPreMigrationBackup copies the database file before applying migrations.
// Time complexity: O(n) where n is database file size
// Space complexity: O(1) — streaming copy with fixed buffer
func createPreMigrationBackup(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Get the database file path from the connection
	var dbPath string
	row := sqlDB.QueryRow("PRAGMA database_list")
	var seq int
	var name string
	if err := row.Scan(&seq, &name, &dbPath); err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	if dbPath == "" || dbPath == ":memory:" {
		return nil // Skip backup for in-memory databases
	}

	// Checkpoint WAL before backup
	if err := CheckpointWAL(db); err != nil {
		return fmt.Errorf("WAL checkpoint failed: %w", err)
	}

	// Create backup directory
	backupDir := filepath.Join(filepath.Dir(dbPath), "pre_migration_backups")
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("pre_migration_%s.db", timestamp))

	src, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		return err
	}

	slog.Info("pre-migration backup created", "path", backupPath)
	return nil
}
