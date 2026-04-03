package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrCaptchaRequired = errors.New("captcha verification required")
	ErrCaptchaInvalid  = errors.New("invalid captcha response")
	ErrCaptchaExpired  = errors.New("captcha challenge expired")
)

// CaptchaChallenge represents a slider captcha challenge
type CaptchaChallenge struct {
	Token      string    `json:"token"`
	ImageIndex int       `json:"image_index"` // Index for selecting background image (0-9)
	TargetX    int       `json:"target_x"`    // Target X position (percentage 20-80)
	CreatedAt  time.Time `json:"-"`
	ExpiresAt  time.Time `json:"-"`
}

// CaptchaVerifyRequest represents the verification request from frontend
type CaptchaVerifyRequest struct {
	Token    string `json:"token"`
	SliderX  int    `json:"slider_x"` // User's slider position (percentage)
}

// CaptchaManager manages slider captcha challenges
type CaptchaManager struct {
	mu         sync.RWMutex
	challenges map[string]*CaptchaChallenge
	tolerance  int           // Allowed deviation in percentage points
	lifetime   time.Duration // Challenge lifetime
}

// NewCaptchaManager creates a new captcha manager
func NewCaptchaManager() *CaptchaManager {
	cm := &CaptchaManager{
		challenges: make(map[string]*CaptchaChallenge),
		tolerance:  5,               // 5% tolerance
		lifetime:   2 * time.Minute, // 2 minute lifetime
	}
	
	// Start cleanup goroutine
	go cm.cleanupLoop()
	
	return cm
}

// GenerateChallenge creates a new captcha challenge
func (cm *CaptchaManager) GenerateChallenge() (*CaptchaChallenge, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Generate random image index (0-9)
	imageIndexBytes := make([]byte, 1)
	rand.Read(imageIndexBytes)
	imageIndex := int(imageIndexBytes[0]) % 10

	// Generate random target X position (20-80%)
	targetXBytes := make([]byte, 1)
	rand.Read(targetXBytes)
	targetX := 20 + int(targetXBytes[0])%61 // 20 to 80

	now := time.Now()
	challenge := &CaptchaChallenge{
		Token:      token,
		ImageIndex: imageIndex,
		TargetX:    targetX,
		CreatedAt:  now,
		ExpiresAt:  now.Add(cm.lifetime),
	}

	cm.challenges[token] = challenge
	return challenge, nil
}

// VerifyChallenge verifies a captcha response
func (cm *CaptchaManager) VerifyChallenge(token string, sliderX int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	challenge, exists := cm.challenges[token]
	if !exists {
		return ErrCaptchaInvalid
	}

	// Remove challenge after verification attempt (one-time use)
	delete(cm.challenges, token)

	// Check expiration
	if time.Now().After(challenge.ExpiresAt) {
		return ErrCaptchaExpired
	}

	// Check if slider position is within tolerance
	diff := sliderX - challenge.TargetX
	if diff < 0 {
		diff = -diff
	}
	if diff > cm.tolerance {
		return ErrCaptchaInvalid
	}

	return nil
}

// cleanupLoop periodically removes expired challenges
func (cm *CaptchaManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.cleanup()
	}
}

func (cm *CaptchaManager) cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for token, challenge := range cm.challenges {
		if now.After(challenge.ExpiresAt) {
			delete(cm.challenges, token)
		}
	}
}
