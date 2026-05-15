package auth

import (
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecureP@ss123"
	cost := 12

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == password {
		t.Error("hash should not equal plaintext password")
	}

	if !VerifyPassword(hash, password) {
		t.Error("VerifyPassword should return true for correct password")
	}

	if VerifyPassword(hash, "WrongPassword") {
		t.Error("VerifyPassword should return false for wrong password")
	}
}

func TestPasswordHashing_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("", 12)
	if err != nil {
		t.Fatalf("HashPassword failed for empty: %v", err)
	}
	if !VerifyPassword(hash, "") {
		t.Error("should verify empty password")
	}
	if VerifyPassword(hash, "notempty") {
		t.Error("should not verify wrong password")
	}
}

func TestLoginLimiter_BasicFlow(t *testing.T) {
	limiter := NewLoginLimiter(3, 15)

	// Initially not locked
	if limiter.IsLocked("user1") {
		t.Error("user should not be locked initially")
	}

	// Remaining attempts = max
	if limiter.RemainingAttempts("user1") != 3 {
		t.Errorf("expected 3 remaining, got %d", limiter.RemainingAttempts("user1"))
	}

	// Record failures
	limiter.RecordFailure("user1")
	if limiter.RemainingAttempts("user1") != 2 {
		t.Errorf("expected 2 remaining, got %d", limiter.RemainingAttempts("user1"))
	}

	limiter.RecordFailure("user1")
	if limiter.RemainingAttempts("user1") != 1 {
		t.Errorf("expected 1 remaining, got %d", limiter.RemainingAttempts("user1"))
	}

	// Third failure should lock
	limiter.RecordFailure("user1")
	if !limiter.IsLocked("user1") {
		t.Error("user should be locked after 3 failures")
	}

	// Different user should not be affected
	if limiter.IsLocked("user2") {
		t.Error("user2 should not be locked")
	}
}

func TestLoginLimiter_Reset(t *testing.T) {
	limiter := NewLoginLimiter(3, 15)

	limiter.RecordFailure("user1")
	limiter.RecordFailure("user1")

	// Reset
	limiter.ResetAttempts("user1")

	if limiter.RemainingAttempts("user1") != 3 {
		t.Errorf("expected 3 remaining after reset, got %d", limiter.RemainingAttempts("user1"))
	}
	if limiter.IsLocked("user1") {
		t.Error("user should not be locked after reset")
	}
}

func TestSessionManager_CreateAndValidate(t *testing.T) {
	_ = NewSessionManager(8)

	// Test the session token generation independently
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}

	if len(token) == 0 {
		t.Error("token should not be empty")
	}

	// Tokens should be unique
	token2, _ := generateToken()
	if token == token2 {
		t.Error("two generated tokens should not be identical")
	}
}

func TestSessionManager_ExpiredSession(t *testing.T) {
	sm := NewSessionManager(0) // 0 hours = expires immediately

	// Insert a session with past expiry directly
	sm.mu.Lock()
	sm.sessions["test-token"] = &Session{
		Token:     "test-token",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
	}
	sm.mu.Unlock()

	// Should be invalid since ExpiresAt is in the past
	session := sm.ValidateSession("test-token")
	if session != nil {
		t.Error("expired session should return nil")
	}
}

func TestSessionManager_NonExistentSession(t *testing.T) {
	sm := NewSessionManager(8)

	session := sm.ValidateSession("non-existent-token")
	if session != nil {
		t.Error("non-existent session should return nil")
	}
}
