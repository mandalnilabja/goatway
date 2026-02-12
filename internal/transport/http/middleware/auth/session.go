package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// Session represents an authenticated web session.
type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore manages web UI sessions (in-memory).
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewSessionStore creates a new session store with the given TTL.
func NewSessionStore(ttl time.Duration) *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	go store.cleanup() // Background cleanup of expired sessions
	return store
}

// Create creates a new session and returns it.
func (s *SessionStore) Create() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateSessionID()
	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.ttl),
	}
	s.sessions[id] = session
	return session
}

// Get retrieves a session by ID.
func (s *SessionStore) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[id]
	if session == nil {
		return nil
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil
	}

	return session
}

// Delete removes a session.
func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// cleanup removes expired sessions every minute.
func (s *SessionStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

// generateSessionID creates a cryptographically secure session ID.
func generateSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SessionAuth middleware protects web UI routes with session authentication.
// Redirects to login page if no valid session exists.
func SessionAuth(sessions *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("goatway_session")
			if err != nil || cookie.Value == "" {
				http.Redirect(w, r, "/web/login", http.StatusFound)
				return
			}

			session := sessions.Get(cookie.Value)
			if session == nil {
				http.Redirect(w, r, "/web/login", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetSessionCookie creates and sets a session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, r *http.Request, session *Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "goatway_session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt,
	})
}

// ClearSessionCookie clears the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "goatway_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
