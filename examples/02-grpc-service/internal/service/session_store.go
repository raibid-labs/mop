package service

import (
	"sync"
)

// Session represents a user session
type Session struct {
	UserID   string
	Username string
	Email    string
	Roles    []string
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (ss *SessionStore) CreateSession(userID, username, email string, roles []string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.sessions[userID] = &Session{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
	}
}

// GetSession retrieves a session by user ID
func (ss *SessionStore) GetSession(userID string) (*Session, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	session, exists := ss.sessions[userID]
	return session, exists
}

// DeleteSession removes a session
func (ss *SessionStore) DeleteSession(userID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.sessions, userID)
}
