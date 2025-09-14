package state

import (
	"context"
	"sync"
	"time"

	"prototype-game/backend/internal/spatial"
)

// PlayerState captures minimal persistent state for a player.
type PlayerState struct {
	Pos     spatial.Vec2 `json:"pos"`
	Logins  int          `json:"logins"`
	Updated time.Time    `json:"updated"`
}

// Store is a minimal interface for persisting player state.
type Store interface {
	Load(ctx context.Context, playerID string) (PlayerState, bool, error)
	Save(ctx context.Context, playerID string, st PlayerState) error
}

// MemStore is a simple in-memory store for development/testing.
type MemStore struct {
	mu   sync.RWMutex
	data map[string]PlayerState
}

func NewMemStore() *MemStore { return &MemStore{data: make(map[string]PlayerState)} }

func (m *MemStore) Load(_ context.Context, playerID string) (PlayerState, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.data[playerID]
	return st, ok, nil
}

func (m *MemStore) Save(_ context.Context, playerID string, st PlayerState) error {
	m.mu.Lock()
	m.data[playerID] = st
	m.mu.Unlock()
	return nil
}
