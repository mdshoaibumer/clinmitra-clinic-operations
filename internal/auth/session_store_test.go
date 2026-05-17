package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clinmitra/internal/models"
)

func TestSessionStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewSessionStore(tmpDir)

	session := &Session{
		Token:     "test-token-abc123",
		UserID:    "user-1",
		Username:  "admin",
		FullName:  "Test Admin",
		Role:      models.RoleAdmin,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}

	// Save session
	if err := store.Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists with restrictive permissions
	filePath := filepath.Join(tmpDir, "session.json")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Session file not created: %v", err)
	}
	if info.Mode().Perm()&0077 != 0 {
		// On Windows permissions work differently, so only check on Unix
		if os.Getenv("OS") != "Windows_NT" {
			t.Errorf("Session file has insecure permissions: %v", info.Mode())
		}
	}

	// Load session
	loaded := store.Load()
	if loaded == nil {
		t.Fatal("Load returned nil")
	}
	if loaded.Token != session.Token {
		t.Errorf("Token mismatch: got %s, want %s", loaded.Token, session.Token)
	}
	if loaded.UserID != session.UserID {
		t.Errorf("UserID mismatch: got %s, want %s", loaded.UserID, session.UserID)
	}
	if loaded.Username != session.Username {
		t.Errorf("Username mismatch: got %s, want %s", loaded.Username, session.Username)
	}
	if loaded.Role != session.Role {
		t.Errorf("Role mismatch: got %s, want %s", loaded.Role, session.Role)
	}
}

func TestSessionStore_LoadExpired(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewSessionStore(tmpDir)

	// Save an already-expired session
	session := &Session{
		Token:     "expired-token",
		UserID:    "user-1",
		Username:  "admin",
		FullName:  "Admin",
		Role:      models.RoleAdmin,
		CreatedAt: time.Now().Add(-10 * time.Hour),
		ExpiresAt: time.Now().Add(-2 * time.Hour),
	}

	if err := store.Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load should return nil for expired session
	loaded := store.Load()
	if loaded != nil {
		t.Error("Expected nil for expired session, got non-nil")
	}

	// File should be cleaned up
	filePath := filepath.Join(tmpDir, "session.json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Expected expired session file to be removed")
	}
}

func TestSessionStore_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewSessionStore(tmpDir)

	session := &Session{
		Token:     "test-token",
		UserID:    "user-1",
		Username:  "admin",
		FullName:  "Admin",
		Role:      models.RoleAdmin,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}

	_ = store.Save(session)

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	loaded := store.Load()
	if loaded != nil {
		t.Error("Expected nil after Clear, got non-nil")
	}
}

func TestSessionStore_LoadNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewSessionStore(tmpDir)

	// Load with no file should return nil without error
	loaded := store.Load()
	if loaded != nil {
		t.Error("Expected nil when no session file exists")
	}
}

func TestSessionStore_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewSessionStore(tmpDir)

	// Write garbage data
	filePath := filepath.Join(tmpDir, "session.json")
	if err := os.WriteFile(filePath, []byte("not-json{{{"), 0600); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Load should return nil and clean up the file
	loaded := store.Load()
	if loaded != nil {
		t.Error("Expected nil for corrupt file")
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Expected corrupt session file to be removed")
	}
}

func TestSessionManager_RestoreSession(t *testing.T) {
	sm := NewSessionManager(8)

	session := &Session{
		Token:     "restored-token",
		UserID:    "user-1",
		Username:  "admin",
		FullName:  "Admin",
		Role:      models.RoleAdmin,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(4 * time.Hour),
	}

	sm.RestoreSession(session)

	// Should be findable via ValidateSession
	validated := sm.ValidateSession(session.Token)
	if validated == nil {
		t.Fatal("Restored session not found via ValidateSession")
	}
	if validated.UserID != session.UserID {
		t.Errorf("UserID mismatch: got %s, want %s", validated.UserID, session.UserID)
	}
}

func TestSessionManager_RestoreExpiredSession(t *testing.T) {
	sm := NewSessionManager(8)

	expired := &Session{
		Token:     "expired-token",
		UserID:    "user-1",
		Username:  "admin",
		FullName:  "Admin",
		Role:      models.RoleAdmin,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	sm.RestoreSession(expired)

	// Should not be in the map
	validated := sm.ValidateSession(expired.Token)
	if validated != nil {
		t.Error("Expired session should not be restored")
	}
}
