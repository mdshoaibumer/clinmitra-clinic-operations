package auth

import (
	"sync"
	"time"
)

type LoginAttempt struct {
	FailedCount int
	LockedUntil time.Time
}

type LoginLimiter struct {
	attempts    map[string]*LoginAttempt
	mu          sync.RWMutex
	maxAttempts int
	lockMinutes int
	maxEntries  int
}

// NewLoginLimiter creates a rate limiter that locks accounts after
// maxAttempts failed logins for lockMinutes duration. Caps the internal
// map at 1000 entries to prevent memory exhaustion from brute-force.
func NewLoginLimiter(maxAttempts, lockMinutes int) *LoginLimiter {
	return &LoginLimiter{
		attempts:    make(map[string]*LoginAttempt),
		maxAttempts: maxAttempts,
		lockMinutes: lockMinutes,
		maxEntries:  1000, // Prevent unbounded memory growth from brute-force attempts
	}
}

// IsLocked returns true if the given username is currently locked out
// due to exceeding the maximum number of failed login attempts.
func (ll *LoginLimiter) IsLocked(username string) bool {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	attempt, exists := ll.attempts[username]
	if !exists {
		return false
	}

	if time.Now().After(attempt.LockedUntil) {
		return false
	}

	return attempt.FailedCount >= ll.maxAttempts
}

// RecordFailure increments the failure count for a username. If the
// count reaches maxAttempts, the account is locked for lockMinutes.
// Evicts expired entries when the map exceeds maxEntries.
func (ll *LoginLimiter) RecordFailure(username string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	// Evict expired entries if map is too large to prevent unbounded growth
	if len(ll.attempts) >= ll.maxEntries {
		now := time.Now()
		for k, a := range ll.attempts {
			// Remove entries that are either:
			// - locked but lock has expired, or
			// - not locked (stale partial failure records)
			if a.FailedCount >= ll.maxAttempts && now.After(a.LockedUntil) {
				delete(ll.attempts, k)
			} else if a.FailedCount < ll.maxAttempts && now.After(a.LockedUntil.Add(time.Duration(ll.lockMinutes)*time.Minute)) {
				// Stale partial failures older than lockout window
				delete(ll.attempts, k)
			}
		}
	}

	attempt, exists := ll.attempts[username]
	if !exists {
		attempt = &LoginAttempt{}
		ll.attempts[username] = attempt
	}

	// Reset if lock has expired
	if time.Now().After(attempt.LockedUntil) && attempt.FailedCount >= ll.maxAttempts {
		attempt.FailedCount = 0
	}

	attempt.FailedCount++
	if attempt.FailedCount >= ll.maxAttempts {
		attempt.LockedUntil = time.Now().Add(time.Duration(ll.lockMinutes) * time.Minute)
	}
}

// ResetAttempts clears the failure record for a username (called on
// successful login).
func (ll *LoginLimiter) ResetAttempts(username string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	delete(ll.attempts, username)
}

// RemainingAttempts returns the number of login attempts remaining
// before the account is locked.
func (ll *LoginLimiter) RemainingAttempts(username string) int {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	attempt, exists := ll.attempts[username]
	if !exists {
		return ll.maxAttempts
	}

	remaining := ll.maxAttempts - attempt.FailedCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
