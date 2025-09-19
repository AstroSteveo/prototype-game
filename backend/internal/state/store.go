package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"prototype-game/backend/internal/spatial"
)

// PlayerState captures persistent state for a player including inventory and equipment.
type PlayerState struct {
	Pos               spatial.Vec2    `json:"pos"`
	Logins            int             `json:"logins"`
	Updated           time.Time       `json:"updated"`
	Version           int64           `json:"version"`            // For optimistic locking
	InventoryData     json.RawMessage `json:"inventory_data"`     // Serialized inventory state
	EquipmentData     json.RawMessage `json:"equipment_data"`     // Serialized equipment state
	SkillsData        json.RawMessage `json:"skills_data"`        // Serialized skills state
	CooldownTimers    json.RawMessage `json:"cooldown_timers"`    // Serialized equipment cooldowns
	EncumbranceConfig json.RawMessage `json:"encumbrance_config"` // Weight/bulk limits
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

// FileStore persists player state to a JSON file on disk.
type FileStore struct {
	mu       sync.RWMutex
	data     map[string]PlayerState
	filePath string
	dirty    bool
	lastSync time.Time
}

// NewFileStore creates a new file-backed store.
// The file will be created if it doesn't exist.
func NewFileStore(filePath string) (*FileStore, error) {
	fs := &FileStore{
		data:     make(map[string]PlayerState),
		filePath: filePath,
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Load existing data if file exists
	if err := fs.load(); err != nil {
		return nil, fmt.Errorf("failed to load from %s: %w", filePath, err)
	}

	fs.lastSync = time.Now()
	return fs, nil
}

func (fs *FileStore) Load(_ context.Context, playerID string) (PlayerState, bool, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	st, ok := fs.data[playerID]
	return st, ok, nil
}

func (fs *FileStore) Save(_ context.Context, playerID string, st PlayerState) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.data[playerID] = st
	fs.dirty = true
	return nil
}

// Flush writes the current state to disk if there are changes.
func (fs *FileStore) Flush() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.dirty {
		return nil
	}

	return fs.flushLocked()
}

// flushLocked writes to disk without acquiring the lock (caller must hold lock).
func (fs *FileStore) flushLocked() error {
	// Write to temporary file first for atomicity
	tempPath := fs.filePath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(fs.data); err != nil {
		file.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempPath, fs.filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to move temp file: %w", err)
	}

	fs.dirty = false
	fs.lastSync = time.Now()
	return nil
}

// load reads the state from disk into memory.
func (fs *FileStore) load() error {
	file, err := os.Open(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start with empty state
			return nil
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&fs.data); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}

// StartPeriodicFlush starts a goroutine that periodically flushes dirty data to disk.
// Returns a channel that should be closed to stop the periodic flushing.
func (fs *FileStore) StartPeriodicFlush(interval time.Duration) chan<- struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := fs.Flush(); err != nil {
					// Log error but continue - this is best-effort persistence
					// In a real system, we'd use a proper logger here
					fmt.Printf("FileStore: periodic flush failed: %v\n", err)
				}
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// GracefulShutdown performs a final flush of all data to disk.
func (fs *FileStore) GracefulShutdown(ctx context.Context) error {
	// Wait for context or perform flush
	done := make(chan error, 1)
	go func() {
		done <- fs.Flush()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
