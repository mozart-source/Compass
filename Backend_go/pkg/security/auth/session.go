package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Token        string    `json:"token"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsValid      bool      `json:"is_valid"`
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session // token -> session
	mu       sync.RWMutex
}

var (
	sessionStore     *SessionStore
	sessionStoreOnce sync.Once
)

// GetSessionStore returns the singleton instance of SessionStore
func GetSessionStore() *SessionStore {
	sessionStoreOnce.Do(func() {
		sessionStore = &SessionStore{
			sessions: make(map[string]*Session),
		}
	})
	return sessionStore
}

// CreateSession creates a new session
func (ss *SessionStore) CreateSession(userID uuid.UUID, deviceInfo, ipAddress string, token string, expiryDuration time.Duration) *Session {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		Token:        token,
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(expiryDuration),
		IsValid:      true,
	}

	ss.sessions[token] = session
	return session
}

// GetSession retrieves a session by token
func (ss *SessionStore) GetSession(token string) (*Session, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	session, exists := ss.sessions[token]
	if !exists || !session.IsValid || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

// InvalidateSession marks a session as invalid
func (ss *SessionStore) InvalidateSession(token string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if session, exists := ss.sessions[token]; exists {
		session.IsValid = false
	}
	delete(ss.sessions, token)
}

// GetUserSessions returns all active sessions for a user
func (ss *SessionStore) GetUserSessions(userID uuid.UUID) []*Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var userSessions []*Session
	now := time.Now()

	for _, session := range ss.sessions {
		if session.UserID == userID && session.IsValid && now.Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions
}

// CleanupExpiredSessions removes expired sessions
func (ss *SessionStore) CleanupExpiredSessions() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	now := time.Now()
	for token, session := range ss.sessions {
		if !session.IsValid || now.After(session.ExpiresAt) {
			delete(ss.sessions, token)
		}
	}
}

// UpdateSessionActivity updates the last activity time of a session
func (ss *SessionStore) UpdateSessionActivity(token string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if session, exists := ss.sessions[token]; exists {
		session.LastActivity = time.Now()
	}
}
