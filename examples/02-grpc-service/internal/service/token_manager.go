package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// TokenManager handles token generation and validation
type TokenManager struct {
	tokens map[string]*TokenInfo
	mu     sync.RWMutex
}

// TokenInfo stores token metadata
type TokenInfo struct {
	UserID    string
	ExpiresAt time.Time
	IsRefresh bool
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]*TokenInfo),
	}
}

// GenerateToken creates a new access token
func (tm *TokenManager) GenerateToken(userID string) (string, time.Time) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	tm.tokens[token] = &TokenInfo{
		UserID:    userID,
		ExpiresAt: expiresAt,
		IsRefresh: false,
	}

	return token, expiresAt
}

// GenerateRefreshToken creates a new refresh token
func (tm *TokenManager) GenerateRefreshToken(userID string) (string, time.Time) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	tm.tokens[token] = &TokenInfo{
		UserID:    userID,
		ExpiresAt: expiresAt,
		IsRefresh: true,
	}

	return token, expiresAt
}

// ValidateToken checks if a token is valid
func (tm *TokenManager) ValidateToken(token string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tokens[token]
	if !exists {
		return "", false
	}

	if time.Now().After(info.ExpiresAt) {
		return "", false
	}

	return info.UserID, true
}

// RevokeToken removes a token
func (tm *TokenManager) RevokeToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tokens, token)
}

// IsRefreshToken checks if token is a refresh token
func (tm *TokenManager) IsRefreshToken(token string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tokens[token]
	if !exists {
		return false
	}

	return info.IsRefresh
}
