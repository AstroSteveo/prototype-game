package join

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

func TestHandleJoin_UsesSavedPosition(t *testing.T) {
	eng := newTestEngine()
	st := state.NewMemStore()
	SetStore(st)
	defer SetStore(nil)

	// Seed saved state for player p2
	_ = st.Save(context.Background(), "p2", state.PlayerState{Pos: spatial.Vec2{X: 7.5, Z: -1.25}, Logins: 3, Updated: time.Now()})

	auth := fakeAuth{"tok": {"p2", "Eve"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ack, err := HandleJoin(ctx, auth, eng, Hello{Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if diff := abs(ack.Pos.X-7.5) + abs(ack.Pos.Z-(-1.25)); diff > 1e-9 {
		t.Fatalf("expected spawn at saved pos (7.5,-1.25), got %#v", ack.Pos)
	}
	// Verify login count incremented
	saved, ok, _ := st.Load(context.Background(), "p2")
	if !ok || saved.Logins != 4 {
		t.Fatalf("expected logins incremented to 4, got %+v", saved)
	}
}

func TestHandleJoin_WorksWithFileStore(t *testing.T) {
	// Create temporary directory for file store test
	tmpDir, err := os.MkdirTemp("", "join_filestore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	eng := newTestEngine()
	storePath := filepath.Join(tmpDir, "players.json")
	fileStore, err := state.NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	SetStore(fileStore)
	defer SetStore(nil)

	// Save some initial state for player p3
	ctx := context.Background()
	initialState := state.PlayerState{
		Pos:     spatial.Vec2{X: 100.5, Z: 200.5},
		Logins:  5,
		Updated: time.Now(),
	}
	err = fileStore.Save(ctx, "p3", initialState)
	if err != nil {
		t.Fatal(err)
	}
	err = fileStore.Flush()
	if err != nil {
		t.Fatal(err)
	}

	auth := fakeAuth{"token123": {"p3", "FilePlayer"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Test join with existing file store data
	ack, errMsg := HandleJoin(ctx, auth, eng, Hello{Token: "token123"})
	if errMsg != nil {
		t.Fatalf("unexpected error: %+v", errMsg)
	}
	if diff := abs(ack.Pos.X-100.5) + abs(ack.Pos.Z-200.5); diff > 1e-9 {
		t.Fatalf("expected spawn at saved pos (100.5,200.5), got %#v", ack.Pos)
	}

	// Verify login count incremented and persisted
	saved, ok, _ := fileStore.Load(context.Background(), "p3")
	if !ok || saved.Logins != 6 {
		t.Fatalf("expected logins incremented to 6, got %+v", saved)
	}

	// Verify data persists by flushing and reloading
	err = fileStore.Flush()
	if err != nil {
		t.Fatal(err)
	}

	// Create new file store instance to verify persistence
	newFileStore, err := state.NewFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	reloaded, found, err := newFileStore.Load(context.Background(), "p3")
	if err != nil {
		t.Errorf("Failed to reload from new store: %v", err)
	}
	if !found {
		t.Error("Expected to find player after persistence")
	}
	if reloaded.Logins != 6 {
		t.Errorf("Expected 6 logins after persistence, got %d", reloaded.Logins)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
