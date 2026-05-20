package auth

import (
	"testing"
	"time"

	"clinmitra/internal/models"
)

func TestSessionManager_CreateSession(t *testing.T) {
	sm := NewSessionManager(8)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "user-1"},
		Username:  "admin",
		FullName:  "Admin User",
		Role:      models.RoleAdmin,
	}

	session, err := sm.CreateSession(user)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if session.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", session.UserID)
	}
	if session.Username != "admin" {
		t.Fatalf("expected admin, got %s", session.Username)
	}
	if session.Role != models.RoleAdmin {
		t.Fatalf("expected admin role, got %s", session.Role)
	}

	// Validate the session
	validated := sm.ValidateSession(session.Token)
	if validated == nil {
		t.Fatal("expected valid session")
	}
}

func TestSessionManager_CreateSession_ReplacesExisting(t *testing.T) {
	sm := NewSessionManager(8)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "user-1"},
		Username:  "admin",
		FullName:  "Admin User",
		Role:      models.RoleAdmin,
	}

	session1, _ := sm.CreateSession(user)
	session2, _ := sm.CreateSession(user)

	// First session should be invalidated
	if sm.ValidateSession(session1.Token) != nil {
		t.Fatal("old session should be invalidated")
	}
	if sm.ValidateSession(session2.Token) == nil {
		t.Fatal("new session should be valid")
	}
}

func TestSessionManager_DestroySession(t *testing.T) {
	sm := NewSessionManager(8)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "user-1"},
		Username:  "admin",
		FullName:  "Admin User",
		Role:      models.RoleAdmin,
	}

	session, _ := sm.CreateSession(user)

	sm.DestroySession(session.Token)

	if sm.ValidateSession(session.Token) != nil {
		t.Fatal("destroyed session should not be valid")
	}
}

func TestSessionManager_GetActiveSession(t *testing.T) {
	sm := NewSessionManager(8)

	user := &models.User{
		BaseModel: models.BaseModel{ID: "user-1"},
		Username:  "admin",
		FullName:  "Admin User",
		Role:      models.RoleAdmin,
	}

	// No active session initially
	active := sm.GetActiveSession("user-1")
	if active != nil {
		t.Fatal("expected no active session initially")
	}

	// Create session
	sm.CreateSession(user)

	// Now should find active session
	active = sm.GetActiveSession("user-1")
	if active == nil {
		t.Fatal("expected active session")
	}
	if active.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", active.UserID)
	}
}

func TestSessionManager_RestoreSession_ValidAndExpired(t *testing.T) {
	sm := NewSessionManager(8)

	// Restore a valid session
	session := &Session{
		Token:     "restored-token",
		UserID:    "user-1",
		Username:  "admin",
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	sm.RestoreSession(session)

	validated := sm.ValidateSession("restored-token")
	if validated == nil {
		t.Fatal("restored session should be valid")
	}

	// Restore nil - should not panic
	sm.RestoreSession(nil)

	// Restore expired session - should not be added
	expired := &Session{
		Token:     "expired-token",
		UserID:    "user-2",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	sm.RestoreSession(expired)

	if sm.ValidateSession("expired-token") != nil {
		t.Fatal("expired restored session should not be valid")
	}
}

func TestSessionManager_EvictsExpiredOnOverflow(t *testing.T) {
	sm := NewSessionManager(8)
	sm.maxSessions = 2

	// Fill up with expired sessions
	sm.mu.Lock()
	sm.sessions["old1"] = &Session{Token: "old1", UserID: "u1", ExpiresAt: time.Now().Add(-1 * time.Hour)}
	sm.sessions["old2"] = &Session{Token: "old2", UserID: "u2", ExpiresAt: time.Now().Add(-1 * time.Hour)}
	sm.mu.Unlock()

	// New session should trigger eviction
	user := &models.User{
		BaseModel: models.BaseModel{ID: "user-3"},
		Username:  "new",
		FullName:  "New User",
		Role:      models.RoleAdmin,
	}
	session, err := sm.CreateSession(user)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	// New session should work
	if sm.ValidateSession(session.Token) == nil {
		t.Fatal("new session should be valid")
	}
}

func TestLoginLimiter_RecordFailure_Lockout(t *testing.T) {
	limiter := NewLoginLimiter(3, 1) // 3 attempts, 1 minute lockout

	limiter.RecordFailure("user1")
	limiter.RecordFailure("user1")
	limiter.RecordFailure("user1")

	// Should now be locked
	if !limiter.IsLocked("user1") {
		t.Fatal("expected user to be locked after 3 failures")
	}

	// Remaining attempts should be 0
	remaining := limiter.RemainingAttempts("user1")
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}
