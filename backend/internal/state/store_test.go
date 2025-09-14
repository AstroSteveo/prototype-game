package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func TestFileStore_BasicOperations(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "players.json")
	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Test loading non-existent player
	_, found, err := store.Load(ctx, "player1")
	if err != nil {
		t.Errorf("Load should not error for non-existent player: %v", err)
	}
	if found {
		t.Error("Expected not found for non-existent player")
	}

	// Test saving player state
	state1 := PlayerState{
		Pos:     spatial.Vec2{X: 10.5, Z: 20.5},
		Logins:  3,
		Updated: time.Now().Truncate(time.Second), // Truncate for JSON comparison
	}

	err = store.Save(ctx, "player1", state1)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test loading saved player
	loaded, found, err := store.Load(ctx, "player1")
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}
	if !found {
		t.Error("Expected to find saved player")
	}
	if loaded.Pos.X != state1.Pos.X || loaded.Pos.Z != state1.Pos.Z {
		t.Errorf("Position mismatch: expected %v, got %v", state1.Pos, loaded.Pos)
	}
	if loaded.Logins != state1.Logins {
		t.Errorf("Logins mismatch: expected %d, got %d", state1.Logins, loaded.Logins)
	}
}

func TestFileStore_Persistence(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "players.json")

	// Create first store instance and save data
	{
		store, err := NewFileStore(storePath)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		state1 := PlayerState{
			Pos:     spatial.Vec2{X: 15.5, Z: 25.5},
			Logins:  5,
			Updated: time.Now().Truncate(time.Second),
		}
		state2 := PlayerState{
			Pos:     spatial.Vec2{X: 35.5, Z: 45.5},
			Logins:  2,
			Updated: time.Now().Truncate(time.Second),
		}

		err = store.Save(ctx, "player1", state1)
		if err != nil {
			t.Fatal(err)
		}
		err = store.Save(ctx, "player2", state2)
		if err != nil {
			t.Fatal(err)
		}

		// Force flush to disk
		err = store.Flush()
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create second store instance and verify data persisted
	{
		store, err := NewFileStore(storePath)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()

		// Check player1
		loaded1, found1, err := store.Load(ctx, "player1")
		if err != nil {
			t.Errorf("Load player1 failed: %v", err)
		}
		if !found1 {
			t.Error("Expected to find player1 after reload")
		}
		if loaded1.Pos.X != 15.5 || loaded1.Pos.Z != 25.5 {
			t.Errorf("Player1 position mismatch: expected (15.5, 25.5), got (%v, %v)", loaded1.Pos.X, loaded1.Pos.Z)
		}
		if loaded1.Logins != 5 {
			t.Errorf("Player1 logins mismatch: expected 5, got %d", loaded1.Logins)
		}

		// Check player2
		loaded2, found2, err := store.Load(ctx, "player2")
		if err != nil {
			t.Errorf("Load player2 failed: %v", err)
		}
		if !found2 {
			t.Error("Expected to find player2 after reload")
		}
		if loaded2.Pos.X != 35.5 || loaded2.Pos.Z != 45.5 {
			t.Errorf("Player2 position mismatch: expected (35.5, 45.5), got (%v, %v)", loaded2.Pos.X, loaded2.Pos.Z)
		}
		if loaded2.Logins != 2 {
			t.Errorf("Player2 logins mismatch: expected 2, got %d", loaded2.Logins)
		}
	}
}

func TestFileStore_FlushBehavior(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "players.json")
	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Save without flush - should not write to disk immediately
	state1 := PlayerState{
		Pos:     spatial.Vec2{X: 1.0, Z: 2.0},
		Logins:  1,
		Updated: time.Now().Truncate(time.Second),
	}
	err = store.Save(ctx, "player1", state1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file doesn't exist yet (no flush)
	if _, err := os.Stat(storePath); err == nil {
		t.Error("File should not exist before first flush")
	}

	// Flush and verify file exists
	err = store.Flush()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(storePath); err != nil {
		t.Errorf("File should exist after flush: %v", err)
	}

	// Additional flush should be no-op (no dirty data)
	err = store.Flush()
	if err != nil {
		t.Errorf("Second flush should not error: %v", err)
	}
}

func TestFileStore_GracefulShutdown(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "players.json")
	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Save some data
	state1 := PlayerState{
		Pos:     spatial.Vec2{X: 100.0, Z: 200.0},
		Logins:  10,
		Updated: time.Now().Truncate(time.Second),
	}
	err = store.Save(ctx, "shutdown_test", state1)
	if err != nil {
		t.Fatal(err)
	}

	// Test graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = store.GracefulShutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Verify data was saved
	newStore, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	loaded, found, err := newStore.Load(context.Background(), "shutdown_test")
	if err != nil {
		t.Errorf("Load after shutdown failed: %v", err)
	}
	if !found {
		t.Error("Expected to find data after graceful shutdown")
	}
	if loaded.Pos.X != 100.0 || loaded.Pos.Z != 200.0 {
		t.Errorf("Data mismatch after shutdown: expected (100.0, 200.0), got (%v, %v)", loaded.Pos.X, loaded.Pos.Z)
	}
}

func TestFileStore_DirectoryCreation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test creating store in non-existent subdirectory
	storePath := filepath.Join(tmpDir, "subdir", "nested", "players.json")
	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	// Verify directory was created
	dir := filepath.Dir(storePath)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("Directory should have been created: %v", err)
	}

	// Test that store works
	ctx := context.Background()
	state1 := PlayerState{
		Pos:     spatial.Vec2{X: 5.0, Z: 10.0},
		Logins:  1,
		Updated: time.Now().Truncate(time.Second),
	}
	err = store.Save(ctx, "test_player", state1)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Flush()
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was created in the nested directory
	if _, err := os.Stat(storePath); err != nil {
		t.Errorf("Store file should exist: %v", err)
	}
}

func TestFileStore_PeriodicFlush(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storePath := filepath.Join(tmpDir, "players.json")
	store, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	// Start periodic flush with very short interval for testing
	stopCh := store.StartPeriodicFlush(50 * time.Millisecond)

	ctx := context.Background()

	// Save some data
	state1 := PlayerState{
		Pos:     spatial.Vec2{X: 777.0, Z: 888.0},
		Logins:  7,
		Updated: time.Now().Truncate(time.Second),
	}
	err = store.Save(ctx, "periodic_test", state1)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for periodic flush to happen
	time.Sleep(100 * time.Millisecond)

	// Stop periodic flush
	close(stopCh)

	// Create new store to verify data was flushed
	newStore, err := NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	loaded, found, err := newStore.Load(context.Background(), "periodic_test")
	if err != nil {
		t.Errorf("Load after periodic flush failed: %v", err)
	}
	if !found {
		t.Error("Expected to find data after periodic flush")
	}
	if loaded.Pos.X != 777.0 || loaded.Pos.Z != 888.0 {
		t.Errorf("Data mismatch after periodic flush: expected (777.0, 888.0), got (%v, %v)", loaded.Pos.X, loaded.Pos.Z)
	}
}
