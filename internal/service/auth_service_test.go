package service

import (
	"errors"
	"testing"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/utils"
)

// --- Mock User Repository for Auth Tests ---

type mockUserRepoForAuth struct {
	users     map[string]*models.User
	byName    map[string]*models.User
	lastLogin map[string]bool
}

func newMockUserRepoForAuth() *mockUserRepoForAuth {
	return &mockUserRepoForAuth{
		users:     make(map[string]*models.User),
		byName:    make(map[string]*models.User),
		lastLogin: make(map[string]bool),
	}
}

func (m *mockUserRepoForAuth) Create(user *models.User) error {
	if _, exists := m.byName[user.Username]; exists {
		return errors.New("duplicate username")
	}
	m.users[user.ID] = user
	m.byName[user.Username] = user
	return nil
}

func (m *mockUserRepoForAuth) FindByID(id string) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepoForAuth) FindByUsername(username string) (*models.User, error) {
	u, ok := m.byName[username]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepoForAuth) Update(user *models.User) error {
	m.users[user.ID] = user
	m.byName[user.Username] = user
	return nil
}

func (m *mockUserRepoForAuth) UpdateLastLogin(id string) error {
	m.lastLogin[id] = true
	return nil
}

func (m *mockUserRepoForAuth) Count() (int64, error) {
	return int64(len(m.users)), nil
}

// --- Helper to create AuthService for tests ---

func newTestAuthServiceWithDir(dir string) (*AuthService, *mockUserRepoForAuth) {
	cfg := &config.Config{
		AppName:          "TestApp",
		Version:          "1.0.0",
		DataDir:          dir,
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       4, // Fast for tests
	}
	sessionManager := auth.NewSessionManager(cfg.SessionHours)
	loginLimiter := auth.NewLoginLimiter(cfg.MaxLoginAttempts, cfg.LockoutMinutes)
	sessionStore := auth.NewSessionStore(dir)
	auditService := newTestAuditService()
	userRepo := newMockUserRepoForAuth()

	svc := &AuthService{
		userRepo:       userRepo,
		sessionManager: sessionManager,
		loginLimiter:   loginLimiter,
		sessionStore:   sessionStore,
		auditService:   auditService,
		cfg:            cfg,
	}
	return svc, userRepo
}

func createTestUser(repo *mockUserRepoForAuth, username, password string, role models.UserRole, active bool) *models.User {
	hash, _ := auth.HashPassword(password, 4)
	user := &models.User{
		BaseModel:    models.BaseModel{ID: "user-" + username},
		Username:     username,
		PasswordHash: hash,
		FullName:     "Test " + username,
		Role:         role,
		IsActive:     active,
	}
	repo.users[user.ID] = user
	repo.byName[user.Username] = user
	return user
}

// --- Login Tests ---

func TestAuthService_Login_Success(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "Password1", models.RoleAdmin, true)

	resp, err := svc.Login("admin", "Password1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.LoggedIn {
		t.Fatal("expected LoggedIn=true")
	}
	if resp.User.Username != "admin" {
		t.Errorf("expected username 'admin', got: %s", resp.User.Username)
	}
	if resp.User.Role != models.RoleAdmin {
		t.Errorf("expected role admin, got: %s", resp.User.Role)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "Password1", models.RoleAdmin, true)

	_, err := svc.Login("admin", "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got: %s", appErr.Code)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	_, err := svc.Login("nonexistent", "password")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	// Should return generic error, not leak that username doesn't exist
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got: %s", appErr.Code)
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "inactive", "Password1", models.RoleAdmin, false)

	_, err := svc.Login("inactive", "Password1")
	if err == nil {
		t.Fatal("expected error for inactive user")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got: %s", appErr.Code)
	}
}

func TestAuthService_Login_AccountLocked(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "Password1", models.RoleAdmin, true)

	// Exhaust login attempts
	for i := 0; i < 5; i++ {
		svc.Login("admin", "wrongpass")
	}

	// Next attempt should be locked
	_, err := svc.Login("admin", "Password1")
	if err == nil {
		t.Fatal("expected error for locked account")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "ACCOUNT_LOCKED" {
		t.Errorf("expected ACCOUNT_LOCKED, got: %s", appErr.Code)
	}
}

func TestAuthService_Login_OversizedInput(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	longUsername := make([]byte, 51)
	for i := range longUsername {
		longUsername[i] = 'a'
	}

	_, err := svc.Login(string(longUsername), "password")
	if err == nil {
		t.Fatal("expected error for oversized username")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got: %s", appErr.Code)
	}
}

func TestAuthService_Login_ResetsAttemptsOnSuccess(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "Password1", models.RoleAdmin, true)

	// Fail a few times
	for i := 0; i < 3; i++ {
		svc.Login("admin", "wrongpass")
	}

	// Succeed
	resp, err := svc.Login("admin", "Password1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !resp.LoggedIn {
		t.Fatal("expected LoggedIn=true")
	}

	// Fail more — should not lock because attempts were reset
	for i := 0; i < 4; i++ {
		svc.Login("admin", "wrongpass")
	}
	// 4th attempt should still work (not locked, only 4 failures)
	resp, err = svc.Login("admin", "Password1")
	if err != nil {
		t.Fatalf("expected success after reset, got: %v", err)
	}
	if !resp.LoggedIn {
		t.Fatal("expected LoggedIn=true after reset")
	}
}

// --- Logout Tests ---

func TestAuthService_Logout_WithSession(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "Password1", models.RoleAdmin, true)

	// Login first
	svc.Login("admin", "Password1")
	if !svc.IsAuthenticated() {
		t.Fatal("expected authenticated after login")
	}

	// Logout
	err := svc.Logout()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if svc.IsAuthenticated() {
		t.Fatal("expected not authenticated after logout")
	}
}

func TestAuthService_Logout_WithoutSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	// Logout without login — should not panic
	err := svc.Logout()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// --- GetCurrentUser Tests ---

func TestAuthService_GetCurrentUser_NoSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	resp, err := svc.GetCurrentUser()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.LoggedIn {
		t.Fatal("expected LoggedIn=false with no session")
	}
}

func TestAuthService_GetCurrentUser_ValidSession(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "doctor", "pass1234", models.RoleDoctor, true)

	svc.Login("doctor", "pass1234")

	resp, err := svc.GetCurrentUser()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.LoggedIn {
		t.Fatal("expected LoggedIn=true")
	}
	if resp.User.Role != models.RoleDoctor {
		t.Errorf("expected role doctor, got: %s", resp.User.Role)
	}
}

// --- ChangePassword Tests ---

func TestAuthService_ChangePassword_Success(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "OldPass12", models.RoleAdmin, true)

	svc.Login("admin", "OldPass12")

	err := svc.ChangePassword("OldPass12", "NewPass12")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Logout and login with new password
	svc.Logout()
	resp, err := svc.Login("admin", "NewPass12")
	if err != nil {
		t.Fatalf("expected login with new password, got: %v", err)
	}
	if !resp.LoggedIn {
		t.Fatal("expected LoggedIn=true with new password")
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "correct1", models.RoleAdmin, true)

	svc.Login("admin", "correct1")

	err := svc.ChangePassword("WrongOld1", "NewPass12")
	if err == nil {
		t.Fatal("expected error for wrong old password")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAuthService_ChangePassword_TooShort(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "correct1", models.RoleAdmin, true)

	svc.Login("admin", "correct1")

	err := svc.ChangePassword("correct1", "12345")
	if err == nil {
		t.Fatal("expected error for short password")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAuthService_ChangePassword_TooLong(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "correct1", models.RoleAdmin, true)

	svc.Login("admin", "correct1")

	longPass := make([]byte, 129)
	for i := range longPass {
		longPass[i] = 'x'
	}
	err := svc.ChangePassword("correct1", string(longPass))
	if err == nil {
		t.Fatal("expected error for long password")
	}
}

func TestAuthService_ChangePassword_NoSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	err := svc.ChangePassword("old", "NewPass12")
	if err == nil {
		t.Fatal("expected error with no session")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got: %s", appErr.Code)
	}
}

// --- CreateInitialAdmin Tests ---

func TestAuthService_CreateInitialAdmin_Success(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())

	err := svc.CreateInitialAdmin("admin", "Password1", "Dr. Admin")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify user was created
	user, exists := repo.byName["admin"]
	if !exists {
		t.Fatal("expected user to be created")
	}
	if user.Role != models.RoleAdmin {
		t.Errorf("expected admin role, got: %s", user.Role)
	}
	if user.FullName != "Dr. Admin" {
		t.Errorf("expected 'Dr. Admin', got: %s", user.FullName)
	}
	if !user.IsActive {
		t.Error("expected user to be active")
	}
}

func TestAuthService_CreateInitialAdmin_AlreadySetup(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "existing", "pass1234", models.RoleAdmin, true)

	err := svc.CreateInitialAdmin("admin2", "Password1", "Dr. Admin")
	if err == nil {
		t.Fatal("expected error when users already exist")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "SETUP_COMPLETE" {
		t.Errorf("expected SETUP_COMPLETE, got: %s", appErr.Code)
	}
}

func TestAuthService_CreateInitialAdmin_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		fullName string
	}{
		{"short username", "ab", "Password1", "Admin"},
		{"short password", "admin", "12345", "Admin"},
		{"empty full name", "admin", "Password1", ""},
		{"whitespace full name", "admin", "Password1", "   "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newTestAuthServiceWithDir(t.TempDir())

			err := svc.CreateInitialAdmin(tc.username, tc.password, tc.fullName)
			if err == nil {
				t.Fatalf("expected validation error for: %s", tc.name)
			}
		})
	}
}

// --- RequireRole Tests ---

func TestAuthService_RequireRole_Authorized(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "admin", "password1", models.RoleAdmin, true)

	svc.Login("admin", "password1")

	err := svc.RequireRole(models.RoleAdmin)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestAuthService_RequireRole_MultipleRolesAllowed(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "doctor", "password1", models.RoleDoctor, true)

	svc.Login("doctor", "password1")

	err := svc.RequireRole(models.RoleAdmin, models.RoleDoctor)
	if err != nil {
		t.Fatalf("expected no error for doctor in [admin, doctor], got: %v", err)
	}
}

func TestAuthService_RequireRole_Forbidden(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	createTestUser(repo, "receptionist", "password1", models.RoleReceptionist, true)

	svc.Login("receptionist", "password1")

	err := svc.RequireRole(models.RoleAdmin)
	if err == nil {
		t.Fatal("expected error for receptionist requiring admin")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got: %s", appErr.Code)
	}
}

func TestAuthService_RequireRole_NoSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	err := svc.RequireRole(models.RoleAdmin)
	if err == nil {
		t.Fatal("expected error with no session")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got: %s", appErr.Code)
	}
}

// --- GetCurrentUserID / GetCurrentUserRole ---

func TestAuthService_GetCurrentUserID_NoSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	id := svc.GetCurrentUserID()
	if id != "" {
		t.Errorf("expected empty string, got: %s", id)
	}
}

func TestAuthService_GetCurrentUserID_WithSession(t *testing.T) {
	svc, repo := newTestAuthServiceWithDir(t.TempDir())
	user := createTestUser(repo, "admin", "password1", models.RoleAdmin, true)

	svc.Login("admin", "password1")

	id := svc.GetCurrentUserID()
	if id != user.ID {
		t.Errorf("expected %s, got: %s", user.ID, id)
	}
}

func TestAuthService_GetCurrentUserRole_NoSession(t *testing.T) {
	svc, _ := newTestAuthServiceWithDir(t.TempDir())

	role := svc.GetCurrentUserRole()
	if role != "" {
		t.Errorf("expected empty role, got: %s", role)
	}
}
