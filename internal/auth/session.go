package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"clinmitra/internal/models"
)

type Session struct {
	Token     string
	UserID    string
	Username  string
	FullName  string
	Role      models.UserRole
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions     map[string]*Session
	mu           sync.RWMutex
	sessionHours int
	maxSessions  int
}

// NewSessionManager creates a SessionManager that issues sessions lasting
// the given number of hours. Sessions are stored in-memory with a cap to
// prevent unbounded memory growth.
func NewSessionManager(sessionHours int) *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*Session),
		sessionHours: sessionHours,
		maxSessions:  100,
	}
}

// CreateSession generates a new cryptographic token, invalidates any
// existing session for the same user, and stores a new session.
// Expired sessions are cleaned up if the map exceeds maxSessions.
func (sm *SessionManager) CreateSession(user *models.User) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Invalidate any existing session for this user
	for k, s := range sm.sessions {
		if s.UserID == user.ID {
			delete(sm.sessions, k)
		}
	}

	// Evict expired sessions if map exceeds capacity
	if len(sm.sessions) >= sm.maxSessions {
		now := time.Now()
		for k, s := range sm.sessions {
			if now.After(s.ExpiresAt) {
				delete(sm.sessions, k)
			}
		}
	}

	session := &Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(sm.sessionHours) * time.Hour),
	}

	sm.sessions[token] = session
	return session, nil
}

// ValidateSession checks if a token maps to a non-expired session.
// Returns the Session if valid, or nil if expired/not found.
// Expired sessions are automatically cleaned up.
func (sm *SessionManager) ValidateSession(token string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[token]
	if !exists {
		return nil
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, token)
		return nil
	}

	return session
}

// DestroySession removes a session by its token (used on logout).
func (sm *SessionManager) DestroySession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, token)
}

// GetActiveSession finds an active (non-expired) session for a user by
// their ID. Returns nil if no active session exists.
func (sm *SessionManager) GetActiveSession(userID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, s := range sm.sessions {
		if s.UserID == userID && time.Now().Before(s.ExpiresAt) {
			return s
		}
	}
	return nil
}

// RestoreSession re-registers a persisted session into the in-memory map.
// Used on application startup to restore sessions that survive restart.
func (sm *SessionManager) RestoreSession(session *Session) {
	if session == nil || time.Now().After(session.ExpiresAt) {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.Token] = session
}

// generateToken creates a 32-byte cryptographically random token encoded
// as a 64-character hex string for use as a session identifier.
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
