//go:build ws

package ws

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type resumeEntry struct {
	playerID string
	exp      time.Time
}

type ResumeManager struct {
	mu   sync.Mutex
	ttl  time.Duration
	data map[string]resumeEntry
}

func NewResumeManager(ttl time.Duration) *ResumeManager {
	return &ResumeManager{ttl: ttl, data: make(map[string]resumeEntry)}
}

func (m *ResumeManager) Issue(playerID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// If random number generation fails, do not issue a token.
		return ""
	}
	tok := hex.EncodeToString(b[:])
	m.data[tok] = resumeEntry{playerID: playerID, exp: time.Now().Add(m.ttl)}
	return tok
}

func (m *ResumeManager) Lookup(token string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.data[token]
	if !ok {
		return "", false
	}
	if time.Now().After(e.exp) {
		delete(m.data, token)
		return "", false
	}
	return e.playerID, true
}

// Validate checks if a resume token is valid for the specified player ID.
// This method provides a cleaner interface for resume token validation.
func (m *ResumeManager) Validate(token, playerID string) bool {
	if token == "" || playerID == "" {
		return false
	}
	resumePlayerID, ok := m.Lookup(token)
	return ok && resumePlayerID == playerID
}

var defaultResume = NewResumeManager(60 * time.Second)
