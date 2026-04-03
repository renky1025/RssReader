package security

import (
	"sync"
	"time"
)

// LoginAttempt records a login attempt
type LoginAttempt struct {
	Username    string
	IPAddress   string
	Success     bool
	AttemptedAt time.Time
	Nonce       string
}

// LoginLimiter tracks login attempts and determines when captcha is required
type LoginLimiter struct {
	mu              sync.RWMutex
	attempts        map[string][]time.Time // username -> attempt times
	maxAttempts     int                    // Max failed attempts before captcha
	windowDuration  time.Duration          // Time window for counting attempts
	usedNonces      map[string]time.Time   // nonce -> used time (for replay protection)
	nonceExpiration time.Duration          // How long to keep nonces
}

// NewLoginLimiter creates a new login limiter
func NewLoginLimiter(maxAttempts int, windowDuration time.Duration) *LoginLimiter {
	ll := &LoginLimiter{
		attempts:        make(map[string][]time.Time),
		maxAttempts:     maxAttempts,
		windowDuration:  windowDuration,
		usedNonces:      make(map[string]time.Time),
		nonceExpiration: 5 * time.Minute,
	}
	
	// Start cleanup goroutine
	go ll.cleanupLoop()
	
	return ll
}

// RecordAttempt records a login attempt
func (ll *LoginLimiter) RecordAttempt(username string, success bool) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if success {
		// Clear attempts on successful login
		delete(ll.attempts, username)
		return
	}

	// Record failed attempt
	now := time.Now()
	ll.attempts[username] = append(ll.attempts[username], now)
}

// RequiresCaptcha checks if captcha is required for a username
func (ll *LoginLimiter) RequiresCaptcha(username string) bool {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	attempts, exists := ll.attempts[username]
	if !exists {
		return false
	}

	// Count recent attempts within window
	cutoff := time.Now().Add(-ll.windowDuration)
	recentCount := 0
	for _, t := range attempts {
		if t.After(cutoff) {
			recentCount++
		}
	}

	return recentCount >= ll.maxAttempts
}

// GetFailedAttempts returns the number of recent failed attempts
func (ll *LoginLimiter) GetFailedAttempts(username string) int {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	attempts, exists := ll.attempts[username]
	if !exists {
		return 0
	}

	cutoff := time.Now().Add(-ll.windowDuration)
	count := 0
	for _, t := range attempts {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// ValidateNonce checks if a nonce is valid (not used before)
func (ll *LoginLimiter) ValidateNonce(nonce string) bool {
	if nonce == "" {
		return false
	}

	ll.mu.Lock()
	defer ll.mu.Unlock()

	if _, used := ll.usedNonces[nonce]; used {
		return false // Nonce already used (replay attack)
	}

	// Mark nonce as used
	ll.usedNonces[nonce] = time.Now()
	return true
}

// ValidateTimestamp checks if timestamp is within acceptable range
func (ll *LoginLimiter) ValidateTimestamp(timestamp int64, maxAge time.Duration) bool {
	requestTime := time.Unix(timestamp, 0)
	now := time.Now()
	
	// Check if timestamp is not too old
	if now.Sub(requestTime) > maxAge {
		return false
	}
	
	// Check if timestamp is not in the future (with small tolerance)
	if requestTime.Sub(now) > 30*time.Second {
		return false
	}
	
	return true
}

// cleanupLoop periodically removes old data
func (ll *LoginLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ll.cleanup()
	}
}

func (ll *LoginLimiter) cleanup() {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-ll.windowDuration)
	nonceCutoff := now.Add(-ll.nonceExpiration)

	// Cleanup old attempts
	for username, attempts := range ll.attempts {
		var recent []time.Time
		for _, t := range attempts {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(ll.attempts, username)
		} else {
			ll.attempts[username] = recent
		}
	}

	// Cleanup old nonces
	for nonce, usedAt := range ll.usedNonces {
		if usedAt.Before(nonceCutoff) {
			delete(ll.usedNonces, nonce)
		}
	}
}

// ClearAttempts clears all attempts for a username (e.g., after successful captcha)
func (ll *LoginLimiter) ClearAttempts(username string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	delete(ll.attempts, username)
}
