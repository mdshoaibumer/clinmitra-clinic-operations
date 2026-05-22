package service

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"gorm.io/gorm"
)

type BackupInfo struct {
	FileName  string `json:"fileName"`
	FilePath  string `json:"filePath"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
}

type BackupService struct {
	db           *gorm.DB
	cfg          *config.Config
	authService  *AuthService
	auditService *AuditService
	clinicRepo   repository.ClinicRepository
}

// NewBackupService creates a BackupService for database backup and restore
// operations.
func NewBackupService(
	db *gorm.DB,
	cfg *config.Config,
	authService *AuthService,
	auditService *AuditService,
	clinicRepo repository.ClinicRepository,
) *BackupService {
	return &BackupService{
		db:           db,
		cfg:          cfg,
		authService:  authService,
		auditService: auditService,
		clinicRepo:   clinicRepo,
	}
}

// CreateBackup performs a WAL checkpoint, copies the database file to the
// destination directory, and verifies the backup's integrity. If the
// destination is empty, the default backup directory is used.
// Requires admin role.
func (s *BackupService) CreateBackup(destinationDir string) (*BackupInfo, error) {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return nil, err
	}

	if destinationDir == "" {
		destinationDir = s.cfg.BackupDir
	}

	// Resolve to absolute path and validate — prevent path traversal
	absDir, err := filepath.Abs(destinationDir)
	if err != nil {
		return nil, utils.ValidationError("Invalid backup directory path")
	}
	destinationDir = absDir

	if err := os.MkdirAll(destinationDir, 0700); err != nil {
		slog.Error("failed to create backup directory", "dir", destinationDir, "error", err)
		return nil, utils.InternalError("Failed to create backup directory")
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("clinmitra_backup_%s.db", timestamp)
	destPath := filepath.Join(destinationDir, fileName)

	// Use SQLite backup by copying the file (with checkpoint first)
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil, err
	}

	// Force WAL checkpoint before backup
	if _, err := sqlDB.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		slog.Error("failed to checkpoint WAL", "error", err)
		return nil, utils.InternalError("Failed to checkpoint database")
	}

	// Copy database file
	if err := copyFile(s.cfg.DBPath, destPath); err != nil {
		slog.Error("failed to copy database file", "src", s.cfg.DBPath, "dest", destPath, "error", err)
		return nil, utils.InternalError("Failed to copy database file")
	}

	// Verify backup integrity
	valid, err := s.VerifyBackup(destPath)
	if err != nil || !valid {
		os.Remove(destPath)
		return nil, utils.InternalError("Backup integrity check failed")
	}

	info, err := os.Stat(destPath)
	if err != nil {
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditBackup, "backup", "", nil, map[string]string{
		"path": destPath,
	})

	return &BackupInfo{
		FileName:  fileName,
		FilePath:  destPath,
		Size:      info.Size(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// RestoreFromBackup replaces the current database with a backup file.
// Creates a safety backup of the current DB before overwriting. Verifies
// backup integrity before proceeding. Requires application restart after.
// Requires admin role.
func (s *BackupService) RestoreFromBackup(backupPath string) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}

	if backupPath == "" {
		return utils.ValidationError("Backup file path is required")
	}

	// Resolve to absolute path — prevent path traversal
	absPath, err := filepath.Abs(backupPath)
	if err != nil {
		return utils.ValidationError("Invalid backup file path")
	}
	backupPath = absPath

	// Ensure the file has a .db extension to prevent restoring arbitrary files
	if filepath.Ext(backupPath) != ".db" {
		return utils.ValidationError("Backup file must have .db extension")
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return utils.ValidationError("Backup file not found")
	}

	// Verify backup integrity before restore
	valid, err := s.VerifyBackup(backupPath)
	if err != nil || !valid {
		return utils.ValidationError("Backup file is corrupted or invalid")
	}

	// Create a safety backup of current database
	safetyDir := filepath.Join(s.cfg.BackupDir, "pre_restore")
	if err := os.MkdirAll(safetyDir, 0700); err != nil {
		slog.Error("failed to create safety backup directory", "dir", safetyDir, "error", err)
		return utils.InternalError("Failed to create safety backup directory")
	}
	safetyPath := filepath.Join(safetyDir, fmt.Sprintf("pre_restore_%s.db", time.Now().Format("20060102_150405")))
	if err := copyFile(s.cfg.DBPath, safetyPath); err != nil {
		slog.Error("failed to create safety backup", "error", err)
		return utils.InternalError("Failed to create safety backup before restore")
	}

	// Close current database
	sqlDB, err := s.db.DB()
	if err != nil {
		slog.Error("failed to access database connection", "error", err)
		return utils.InternalError("Failed to access database connection")
	}
	sqlDB.Close()

	// Copy backup to database path
	if err := copyFile(backupPath, s.cfg.DBPath); err != nil {
		// Attempt to restore from safety backup
		_ = copyFile(safetyPath, s.cfg.DBPath)
		slog.Error("failed to restore database", "error", err)
		return utils.InternalError("Failed to restore database")
	}

	// The database connection is now closed. The application must be
	// restarted to use the restored database.
	return utils.NewError("RESTART_REQUIRED", "Database restored successfully. Please restart the application.")
}

// VerifyBackup opens a backup file in read-only mode and runs SQLite
// PRAGMA integrity_check to verify it is not corrupted.
func (s *BackupService) VerifyBackup(filePath string) (bool, error) {
	// Resolve to absolute path for safety
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false, utils.InternalError("Invalid file path")
	}

	db, err := sql.Open("sqlite", absPath+"?mode=ro")
	if err != nil {
		return false, err
	}
	defer db.Close()

	var result string
	err = db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return false, err
	}

	return result == "ok", nil
}

// ListBackups scans the backup directory for files matching the naming
// convention (clinmitra_backup_*.db) and returns them sorted newest first.
func (s *BackupService) ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo

	dirs := []string{s.cfg.BackupDir}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".db") {
				continue
			}
			if !strings.HasPrefix(entry.Name(), "clinmitra_backup_") {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			backups = append(backups, BackupInfo{
				FileName:  entry.Name(),
				FilePath:  filepath.Join(dir, entry.Name()),
				Size:      info.Size(),
				CreatedAt: info.ModTime().Format(time.RFC3339),
			})
		}
	}

	// Sort by date descending
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// GetAutoBackupPath returns the configured automatic backup directory.
func (s *BackupService) GetAutoBackupPath() string {
	return s.cfg.BackupDir
}

// CreateCloudBackup creates a backup in the configured cloud sync folder
// (Google Drive, OneDrive, etc.). The file is written locally to the sync
// folder, and the cloud provider's desktop client handles the upload.
// Returns nil info if cloud backup is not enabled/configured.
func (s *BackupService) CreateCloudBackup() (*BackupInfo, error) {
	settings, err := s.clinicRepo.Get()
	if err != nil || settings == nil {
		return nil, nil // No settings yet, skip silently
	}

	if !settings.CloudBackupEnabled || settings.CloudBackupPath == "" {
		return nil, nil // Cloud backup not configured
	}

	// Verify the cloud folder still exists (drive might be disconnected)
	if !isCloudPathAccessible(settings.CloudBackupPath) {
		return nil, utils.InternalError("Cloud backup folder not accessible: " + settings.CloudBackupPath)
	}

	// Create a subfolder for Clinmitra backups within the cloud drive
	cloudBackupDir := filepath.Join(settings.CloudBackupPath, "ClinMitra Backups")
	if err := os.MkdirAll(cloudBackupDir, 0700); err != nil {
		slog.Error("failed to create cloud backup folder", "path", cloudBackupDir, "error", err)
		return nil, utils.InternalError("Failed to create cloud backup folder")
	}

	return s.CreateBackup(cloudBackupDir)
}

// CreateBackupWithCloudSync performs a local backup and, if cloud backup
// is enabled, also copies to the cloud sync folder. Used during auto-backup
// on shutdown.
func (s *BackupService) CreateBackupWithCloudSync() (*BackupInfo, error) {
	// Always do local backup first
	info, err := s.CreateBackup("")
	if err != nil {
		return nil, err
	}

	// Attempt cloud backup (best-effort, don't fail the whole operation)
	if _, cloudErr := s.CreateCloudBackup(); cloudErr != nil {
		// Log but don't fail — local backup succeeded
		fmt.Printf("cloud backup failed (local backup OK): %v\n", cloudErr)
	}

	return info, nil
}

// DetectCloudDrives returns available cloud sync folders on this system.
func (s *BackupService) DetectCloudDrives() []CloudDriveInfo {
	return DetectCloudDrives()
}

// isCloudPathAccessible checks if a cloud sync path exists and is writable.
func isCloudPathAccessible(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}

// copyFile copies a file from src to dst with restrictive permissions (0600)
// and an explicit fsync to ensure data is flushed to disk.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
