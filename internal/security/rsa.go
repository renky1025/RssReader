package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrKeyNotFound    = errors.New("RSA key not found")
	ErrKeyExpired     = errors.New("RSA key expired")
	ErrDecryptFailed  = errors.New("failed to decrypt password")
	ErrInvalidKeyData = errors.New("invalid key data")
)

// RSAKeyPair holds a generated RSA key pair
type RSAKeyPair struct {
	KeyID      string
	PublicKey  string // PEM encoded
	PrivateKey string // PEM encoded
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// RSAManager manages RSA key pairs for password encryption
type RSAManager struct {
	mu          sync.RWMutex
	currentKey  *RSAKeyPair
	privateKeys map[string]*rsa.PrivateKey // keyID -> private key
	keyLifetime time.Duration
}

// NewRSAManager creates a new RSA manager
func NewRSAManager(keyLifetime time.Duration) *RSAManager {
	if keyLifetime == 0 {
		keyLifetime = 24 * time.Hour // default 24 hours
	}
	return &RSAManager{
		privateKeys: make(map[string]*rsa.PrivateKey),
		keyLifetime: keyLifetime,
	}
}

// GenerateKeyPair generates a new RSA key pair
func (m *RSAManager) GenerateKeyPair() (*RSAKeyPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate key ID based on timestamp
	keyID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	now := time.Now()
	keyPair := &RSAKeyPair{
		KeyID:      keyID,
		PublicKey:  string(publicKeyPEM),
		PrivateKey: string(privateKeyPEM),
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.keyLifetime),
	}

	// Store private key for decryption
	m.privateKeys[keyID] = privateKey
	m.currentKey = keyPair

	return keyPair, nil
}

// GetCurrentPublicKey returns the current public key for encryption
func (m *RSAManager) GetCurrentPublicKey() (*RSAKeyPair, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKey == nil || time.Now().After(m.currentKey.ExpiresAt) {
		return nil, ErrKeyNotFound
	}

	return m.currentKey, nil
}

// LoadPrivateKey loads a private key from PEM string
func (m *RSAManager) LoadPrivateKey(keyID, privateKeyPEM string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return ErrInvalidKeyData
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	m.privateKeys[keyID] = privateKey
	return nil
}

// DecryptPassword decrypts a base64-encoded RSA-OAEP encrypted password
func (m *RSAManager) DecryptPassword(keyID, encryptedPassword string) (string, error) {
	m.mu.RLock()
	privateKey, exists := m.privateKeys[keyID]
	m.mu.RUnlock()

	if !exists {
		return "", ErrKeyNotFound
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decrypt using RSA-OAEP with SHA-256 (matching frontend Web Crypto API)
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptFailed
	}

	return string(plaintext), nil
}

// CleanupExpiredKeys removes expired keys from memory
func (m *RSAManager) CleanupExpiredKeys() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Keep keys for a grace period after expiration (for in-flight requests)
	// In production, you'd track expiration per key and remove old ones
	// For now, we keep all keys in memory as they're lightweight
}
