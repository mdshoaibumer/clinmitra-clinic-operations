package auth

import (
	"clinmitra/internal/models"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// persistedSession is the on-disk representation of a session token.
// Only the token is stored; the full session is re-validated from memory.
type persistedSession struct {
	Token     string `json:"token"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expiresAt"`
}

// SessionStore handles persisting the active session to disk so it
// survives application restarts within the session expiry window.
type SessionStore struct {
	filePath string
}

// NewSessionStore creates a store backed by a file in the given data directory.
func NewSessionStore(dataDir string) *SessionStore {
	return &SessionStore{
		filePath: filepath.Join(dataDir, "session.json"),
	}
}

// Save persists the current session to disk with restrictive permissions.
func (ss *SessionStore) Save(session *Session) error {
	if session == nil {
		return ss.Clear()
	}

	data := persistedSession{
		Token:     session.Token,
		UserID:    session.UserID,
		Username:  session.Username,
		FullName:  session.FullName,
		Role:      string(session.Role),
		ExpiresAt: session.ExpiresAt.Unix(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(ss.filePath, bytes, 0600)
}

// Load reads a persisted session from disk. Returns nil if no session
// file exists or if the session has expired.
func (ss *SessionStore) Load() *Session {
	bytes, err := os.ReadFile(ss.filePath)
	if err != nil {
		return nil
	}

	var data persistedSession
	if err := json.Unmarshal(bytes, &data); err != nil {
		// Corrupt file — remove it
		_ = os.Remove(ss.filePath)
		return nil
	}

	expiresAt := time.Unix(data.ExpiresAt, 0)
	if time.Now().After(expiresAt) {
		// Session expired — clean up
		_ = os.Remove(ss.filePath)
		return nil
	}

	return &Session{
		Token:     data.Token,
		UserID:    data.UserID,
		Username:  data.Username,
		FullName:  data.FullName,
		Role:      models.UserRole(data.Role),
		ExpiresAt: expiresAt,
	}
}

// Clear removes the persisted session file.
func (ss *SessionStore) Clear() error {
	if err := os.Remove(ss.filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
