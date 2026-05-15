package handler

import (
	"practivo/internal/service"
)

type BackupHandler struct {
	backupService *service.BackupService
}

// NewBackupHandler creates a BackupHandler backed by the given service.
func NewBackupHandler(backupService *service.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// CreateBackup triggers a database backup to the specified directory
// (or the default backup directory if empty).
func (h *BackupHandler) CreateBackup(destinationDir string) (*service.BackupInfo, error) {
	result, err := h.backupService.CreateBackup(destinationDir)
	return result, safeError(err)
}

// RestoreFromBackup replaces the current database with a verified backup file.
func (h *BackupHandler) RestoreFromBackup(backupPath string) error {
	return safeError(h.backupService.RestoreFromBackup(backupPath))
}

// VerifyBackup checks a backup file's integrity via SQLite PRAGMA integrity_check.
func (h *BackupHandler) VerifyBackup(filePath string) (bool, error) {
	result, err := h.backupService.VerifyBackup(filePath)
	return result, safeError(err)
}

// ListBackups returns all available backup files sorted by date (newest first).
func (h *BackupHandler) ListBackups() ([]service.BackupInfo, error) {
	result, err := h.backupService.ListBackups()
	return result, safeError(err)
}

// GetAutoBackupPath returns the configured automatic backup directory path.
func (h *BackupHandler) GetAutoBackupPath() string {
	return h.backupService.GetAutoBackupPath()
}
