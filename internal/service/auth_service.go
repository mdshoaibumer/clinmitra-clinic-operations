package service

import (
	"fmt"
	"sync"
	"time"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/google/uuid"
)

type AuthResponse struct {
	User     UserInfo `json:"user"`
	LoggedIn bool     `json:"loggedIn"`
}

type UserInfo struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	FullName string          `json:"fullName"`
	Role     models.UserRole `json:"role"`
}

type AuthService struct {
	userRepo       repository.UserRepository
	sessionManager *auth.SessionManager
	loginLimiter   *auth.LoginLimiter
	auditService   *AuditService
	cfg            *config.Config
	mu             sync.RWMutex
	currentSession *auth.Session
}

// NewAuthService creates an AuthService with the required dependencies for
// user authentication, session management, and brute-force protection.
func NewAuthService(
	userRepo repository.UserRepository,
	sessionManager *auth.SessionManager,
	loginLimiter *auth.LoginLimiter,
	auditService *AuditService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionManager: sessionManager,
		loginLimiter:   loginLimiter,
		auditService:   auditService,
		cfg:            cfg,
	}
}

// Login authenticates a user by username/password. It checks the brute-force
// limiter, verifies credentials, creates a session, and records an audit log.
// Returns an AuthResponse with user info on success.
func (s *AuthService) Login(username, password string) (*AuthResponse, error) {
	// Guard against excessively long inputs
	if len(username) > 50 || len(password) > 128 {
		return nil, utils.ErrInvalidCredentials
	}

	if s.loginLimiter.IsLocked(username) {
		s.auditService.LogAction("", models.AuditLogin, "user", "", nil, map[string]string{
			"username": username,
			"status":   "locked",
		})
		return nil, utils.ErrAccountLocked
	}

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// User not found — record failure but return generic error to avoid
		// leaking whether the username exists.
		s.loginLimiter.RecordFailure(username)
		s.auditService.LogAction("", models.AuditLogin, "user", "", nil, map[string]string{
			"username": username,
			"status":   "failed",
		})
		return nil, utils.ErrInvalidCredentials
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		s.loginLimiter.RecordFailure(username)
		remaining := s.loginLimiter.RemainingAttempts(username)
		s.auditService.LogAction("", models.AuditLogin, "user", "", nil, map[string]string{
			"username":  username,
			"status":    "failed",
			"remaining": fmt.Sprintf("%d", remaining),
		})
		return nil, utils.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, utils.ErrForbidden
	}

	session, err := s.sessionManager.CreateSession(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.mu.Lock()
	s.currentSession = session
	s.mu.Unlock()
	s.loginLimiter.ResetAttempts(username)
	_ = s.userRepo.UpdateLastLogin(user.ID)

	s.auditService.LogAction(user.ID, models.AuditLogin, "user", user.ID, nil, map[string]string{
		"status": "success",
	})

	return &AuthResponse{
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Role:     user.Role,
		},
		LoggedIn: true,
	}, nil
}

// Logout destroys the current session and logs the action to the audit trail.
func (s *AuthService) Logout() error {
	s.mu.Lock()
	session := s.currentSession
	s.currentSession = nil
	s.mu.Unlock()

	if session != nil {
		s.auditService.LogAction(session.UserID, models.AuditLogout, "user", session.UserID, nil, nil)
		s.sessionManager.DestroySession(session.Token)
	}
	return nil
}

// GetCurrentUser returns the currently authenticated user's info.
// If no valid session exists, returns LoggedIn=false.
func (s *AuthService) GetCurrentUser() (*AuthResponse, error) {
	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return &AuthResponse{LoggedIn: false}, nil
	}

	validated := s.sessionManager.ValidateSession(session.Token)
	if validated == nil {
		s.mu.Lock()
		s.currentSession = nil
		s.mu.Unlock()
		return &AuthResponse{LoggedIn: false}, nil
	}

	return &AuthResponse{
		User: UserInfo{
			ID:       validated.UserID,
			Username: validated.Username,
			FullName: validated.FullName,
			Role:     validated.Role,
		},
		LoggedIn: true,
	}, nil
}

// ChangePassword verifies the old password and updates to the new one.
// Requires an active session. Minimum password length: 6 characters.
func (s *AuthService) ChangePassword(oldPassword, newPassword string) error {
	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return utils.ErrUnauthorized
	}

	if len(newPassword) < 6 {
		return utils.ValidationError("Password must be at least 6 characters")
	}
	if len(newPassword) > 128 {
		return utils.ValidationError("Password must not exceed 128 characters")
	}

	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		return utils.ErrNotFound
	}

	if !auth.VerifyPassword(user.PasswordHash, oldPassword) {
		return utils.ValidationError("Current password is incorrect")
	}

	hash, err := auth.HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hash
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.auditService.LogAction(user.ID, models.AuditUpdate, "user", user.ID, nil, map[string]string{
		"field": "password",
	})

	return nil
}

// CreateInitialAdmin creates the first admin user during setup.
// Fails if any user already exists (setup already completed).
func (s *AuthService) CreateInitialAdmin(username, password, fullName string) error {
	count, err := s.userRepo.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return utils.ErrSetupAlreadyDone
	}

	if err := utils.ValidateMinLength("Username", username, 3); err != nil {
		return err
	}
	if err := utils.ValidateMinLength("Password", password, 6); err != nil {
		return err
	}
	if err := utils.ValidateRequired("Full name", fullName); err != nil {
		return err
	}

	hash, err := auth.HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &models.User{
		BaseModel: models.BaseModel{
			ID: uuid.New().String(),
		},
		Username:     username,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         models.RoleAdmin,
		IsActive:     true,
		LastLoginAt:  &now,
	}

	return s.userRepo.Create(user)
}

// GetCurrentUserID returns the user ID from the active session,
// or an empty string if no session is active.
func (s *AuthService) GetCurrentUserID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentSession != nil {
		return s.currentSession.UserID
	}
	return ""
}

// IsAuthenticated returns true if there is an active user session.
func (s *AuthService) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSession != nil
}
