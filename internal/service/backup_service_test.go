package service

import (
	"os"
	"path/filepath"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// --- Helpers for Backup Tests ---

func setupBackupTestDB(t *testing.T) (*gorm.DB, *config.Config) {
	t.Helper()
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Create a simple table so integrity check has something to verify
	sqlDB, _ := db.DB()
	_, err = sqlDB.Exec("CREATE TABLE test_data (id TEXT PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	_, err = sqlDB.Exec("INSERT INTO test_data VALUES ('1', 'hello')")
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Ensure DB is closed on test cleanup
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	cfg := &config.Config{
		AppName:          "TestApp",
		Version:          "1.0.0",
		DataDir:          dir,
		DBPath:           dbPath,
		BackupDir:        filepath.Join(dir, "backups"),
		LogDir:           filepath.Join(dir, "logs"),
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       4,
	}

	os.MkdirAll(cfg.BackupDir, 0700)

	return db, cfg
}

func setupBackupService(t *testing.T, role models.UserRole) *BackupService {
	t.Helper()
	db, cfg := setupBackupTestDB(t)

	userRepo := newMockUserRepoForAuth()
	sessionManager := auth.NewSessionManager(cfg.SessionHours)
	loginLimiter := auth.NewLoginLimiter(cfg.MaxLoginAttempts, cfg.LockoutMinutes)
	sessionStore := auth.NewSessionStore(cfg.DataDir)
	auditService := newTestAuditService()

	authService := &AuthService{
		userRepo:       userRepo,
		sessionManager: sessionManager,
		loginLimiter:   loginLimiter,
		sessionStore:   sessionStore,
		auditService:   auditService,
		cfg:            cfg,
	}

	// Create and login a user with the specified role
	createTestUser(userRepo, "testuser", "password1", role, true)
	authService.Login("testuser", "password1")

	clinicRepo := &mockClinicRepoForBackup{}

	return &BackupService{
		db:           db,
		cfg:          cfg,
		authService:  authService,
		auditService: auditService,
		clinicRepo:   clinicRepo,
	}
}

type mockClinicRepoForBackup struct{}

func (m *mockClinicRepoForBackup) Get() (*models.ClinicSettings, error) {
	return nil, nil
}
func (m *mockClinicRepoForBackup) Upsert(settings *models.ClinicSettings) error {
	return nil
}
func (m *mockClinicRepoForBackup) IsSetupComplete() (bool, error) {
	return true, nil
}

// --- CreateBackup Tests ---

func TestBackupService_CreateBackup_Success(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	info, err := svc.CreateBackup("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info == nil {
		t.Fatal("expected backup info, got nil")
	}
	if info.FileName == "" {
		t.Error("expected non-empty filename")
	}
	if info.Size <= 0 {
		t.Error("expected positive file size")
	}

	// Verify file exists
	if _, err := os.Stat(info.FilePath); os.IsNotExist(err) {
		t.Errorf("backup file does not exist: %s", info.FilePath)
	}
}

func TestBackupService_CreateBackup_CustomDirectory(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	customDir := filepath.Join(t.TempDir(), "custom-backups")
	info, err := svc.CreateBackup(customDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify it's in the custom directory
	dir := filepath.Dir(info.FilePath)
	if dir != customDir {
		t.Errorf("expected backup in %s, got: %s", customDir, dir)
	}
}

func TestBackupService_CreateBackup_RequiresAdmin(t *testing.T) {
	svc := setupBackupService(t, models.RoleDoctor)

	_, err := svc.CreateBackup("")
	if err == nil {
		t.Fatal("expected error for non-admin user")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got: %s", appErr.Code)
	}
}

func TestBackupService_CreateBackup_ReceptionistDenied(t *testing.T) {
	svc := setupBackupService(t, models.RoleReceptionist)

	_, err := svc.CreateBackup("")
	if err == nil {
		t.Fatal("expected error for receptionist")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got: %s", appErr.Code)
	}
}

// --- VerifyBackup Tests ---

func TestBackupService_VerifyBackup_ValidDB(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	// Create a backup first
	info, err := svc.CreateBackup("")
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	valid, err := svc.VerifyBackup(info.FilePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !valid {
		t.Error("expected backup to be valid")
	}
}

func TestBackupService_VerifyBackup_CorruptFile(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	// Create a corrupt file
	corruptPath := filepath.Join(t.TempDir(), "corrupt.db")
	os.WriteFile(corruptPath, []byte("not a database"), 0600)

	valid, err := svc.VerifyBackup(corruptPath)
	// Either returns error or valid=false
	if valid && err == nil {
		t.Error("expected corrupt file to fail verification")
	}
}

func TestBackupService_VerifyBackup_NonexistentFile(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	_, err := svc.VerifyBackup("/nonexistent/path/backup.db")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// --- ListBackups Tests ---

func TestBackupService_ListBackups_Empty(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got: %d", len(backups))
	}
}

func TestBackupService_ListBackups_WithBackups(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	// Create two backups with different names (timestamps may collide in fast tests)
	info1, err := svc.CreateBackup("")
	if err != nil {
		t.Fatalf("first backup failed: %v", err)
	}

	// Rename first backup to ensure unique filename
	newPath := filepath.Join(svc.cfg.BackupDir, "clinmitra_backup_20250101_100000.db")
	os.Rename(info1.FilePath, newPath)

	_, err = svc.CreateBackup("")
	if err != nil {
		t.Fatalf("second backup failed: %v", err)
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("expected 2 backups, got: %d", len(backups))
	}
}

func TestBackupService_ListBackups_IgnoresNonBackupFiles(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	// Create a non-backup file in the backup directory
	nonBackup := filepath.Join(svc.cfg.BackupDir, "random_file.db")
	os.WriteFile(nonBackup, []byte("data"), 0600)

	// Create a file with wrong extension
	wrongExt := filepath.Join(svc.cfg.BackupDir, "clinmitra_backup_20240101.txt")
	os.WriteFile(wrongExt, []byte("data"), 0600)

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 valid backups, got: %d", len(backups))
	}
}

// --- RestoreFromBackup Tests ---

func TestBackupService_RestoreFromBackup_EmptyPath(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	err := svc.RestoreFromBackup("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestBackupService_RestoreFromBackup_NonDBExtension(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	err := svc.RestoreFromBackup("/some/file.txt")
	if err == nil {
		t.Fatal("expected error for non-.db extension")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestBackupService_RestoreFromBackup_FileNotFound(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	err := svc.RestoreFromBackup("/nonexistent/backup.db")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestBackupService_RestoreFromBackup_RequiresAdmin(t *testing.T) {
	svc := setupBackupService(t, models.RoleDoctor)

	err := svc.RestoreFromBackup("/some/backup.db")
	if err == nil {
		t.Fatal("expected error for non-admin")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got: %s", appErr.Code)
	}
}

func TestBackupService_RestoreFromBackup_CorruptBackup(t *testing.T) {
	svc := setupBackupService(t, models.RoleAdmin)

	// Create a corrupt .db file
	corruptPath := filepath.Join(t.TempDir(), "corrupt.db")
	os.WriteFile(corruptPath, []byte("not a database at all"), 0600)

	err := svc.RestoreFromBackup(corruptPath)
	if err == nil {
		t.Fatal("expected error for corrupt backup")
	}
}
