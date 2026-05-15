package db

import (
	"fmt"
	"log/slog"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"practivo/internal/config"
)

// NewDatabase opens a SQLite database with WAL mode, foreign keys, and
// performance pragmas. Runs an integrity check before returning.
// The database is configured for single-writer access (MaxOpenConns=1).
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000", cfg.DBPath)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// SQLite supports single writer
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Enable WAL mode and other pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-20000", // 20MB cache
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return nil, fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	// Run integrity check on startup to detect corruption early
	// Time complexity: O(n) where n is database page count
	// Space complexity: O(1) — single row result
	if err := CheckIntegrity(db); err != nil {
		slog.Error("database integrity check failed", "error", err)
		return nil, fmt.Errorf("database integrity check failed: %w", err)
	}

	return db, nil
}

// CheckIntegrity runs SQLite PRAGMA integrity_check to detect corruption.
func CheckIntegrity(db *gorm.DB) error {
	var result string
	if err := db.Raw("PRAGMA integrity_check").Scan(&result).Error; err != nil {
		return fmt.Errorf("integrity check query failed: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("database corruption detected: %s", result)
	}
	slog.Info("database integrity check passed")
	return nil
}

// CheckpointWAL forces a WAL checkpoint to flush pending writes to the main database file.
// Should be called before backup and on graceful shutdown.
// Time complexity: O(n) where n is number of WAL frames
// Space complexity: O(1)
func CheckpointWAL(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// CloseDatabase gracefully shuts down the database by flushing WAL
// (via TRUNCATE checkpoint) and closing the underlying sql.DB connection.
func CloseDatabase(db *gorm.DB) error {
	// Checkpoint WAL before closing to ensure all writes are flushed
	if err := CheckpointWAL(db); err != nil {
		slog.Error("WAL checkpoint failed during shutdown", "error", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
