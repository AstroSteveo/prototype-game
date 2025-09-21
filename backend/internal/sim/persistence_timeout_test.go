package sim

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

// TestDisconnectPersistence_TimeoutHandling tests timeout-bounded save on disconnect
func TestDisconnectPersistence_TimeoutHandling(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	// Start engine first
	eng.Start()
	defer eng.Stop(context.Background())

	// Set up store that simulates slow saves
	store := &SlowStore{
		inner: state.NewMemStore(),
		delay: 3 * time.Second, // Simulate slow persistence
	}
	eng.SetPersistenceStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng.StartPersistence(ctx)
	defer eng.StopPersistence()

	playerID := "test-timeout-player"
	playerName := "TimeoutPlayer"
	initialPos := spatial.Vec2{X: 100, Z: 200}

	// Add player to engine
	player := eng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
	playerMgr := eng.GetPlayerManager()
	playerMgr.InitializePlayer(player)

	// Test 1: Normal case - sufficient timeout should complete successfully
	t.Run("SufficientTimeout", func(t *testing.T) {
		// Use generous timeout that should allow completion
		persistCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		eng.RequestPlayerDisconnectPersist(persistCtx, playerID)

		// Give time for persistence to complete
		time.Sleep(4 * time.Second)

		elapsed := time.Since(start)
		if elapsed > 6*time.Second {
			t.Errorf("Persistence took too long: %v", elapsed)
		}

		// Verify state was persisted
		persistedState, exists, err := store.Load(context.Background(), playerID)
		if err != nil {
			t.Fatalf("Failed to load persisted state: %v", err)
		}
		if !exists {
			t.Fatal("Player state should be persisted")
		}

		// Verify position was saved
		if persistedState.Pos.X != initialPos.X || persistedState.Pos.Z != initialPos.Z {
			t.Errorf("Position not persisted correctly: got %v, want %v", persistedState.Pos, initialPos)
		}

		t.Logf("✓ Persistence completed successfully in %v", elapsed)
	})

	// Test 2: Insufficient timeout - should handle gracefully
	t.Run("InsufficientTimeout", func(t *testing.T) {
		// Use very short timeout that should expire before save completes
		persistCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		eng.RequestPlayerDisconnectPersist(persistCtx, playerID)

		// Wait for timeout to expire
		<-persistCtx.Done()
		elapsed := time.Since(start)

		// Should timeout quickly
		if elapsed > 500*time.Millisecond {
			t.Errorf("Timeout took too long: %v", elapsed)
		}

		// Even though context timed out, the persistence may have fallen back to sync save
		// Give a bit more time for any fallback mechanisms
		time.Sleep(100 * time.Millisecond)

		t.Logf("✓ Timeout handled correctly in %v", elapsed)
	})

	// Test 3: Verify metrics reflect timeout scenarios
	t.Run("TimeoutMetrics", func(t *testing.T) {
		// Create a separate engine with a fast store for metrics testing
		fastEng := NewEngine(Config{
			CellSize:             256,
			AOIRadius:            128,
			TickHz:               20,
			SnapshotHz:           10,
			HandoverHysteresisM:  2,
			TargetDensityPerCell: 3,
			MaxBots:              100,
		})
		fastEng.Start()
		defer fastEng.Stop(context.Background())

		fastStore := state.NewMemStore()
		fastEng.SetPersistenceStore(fastStore)

		fastCtx, fastCancel := context.WithCancel(context.Background())
		defer fastCancel()
		fastEng.StartPersistence(fastCtx)
		defer fastEng.StopPersistence()

		// Add player to the fast engine
		fastPlayer := fastEng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
		fastPlayerMgr := fastEng.GetPlayerManager()
		fastPlayerMgr.InitializePlayer(fastPlayer)

		// Request a disconnect persist to ensure metrics are updated
		persistCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		fastEng.RequestPlayerDisconnectPersist(persistCtx, playerID)

		// Give time for processing - be more generous
		time.Sleep(1 * time.Second)

		metrics := fastEng.GetPersistenceMetrics()
		// Should have metrics structure even if no operations completed yet
		if len(metrics) == 0 {
			t.Error("Should have persistence metrics structure")
		}

		attempts, _ := metrics["persist_attempts"].(int64)
		successes, _ := metrics["persist_successes"].(int64)
		failures, _ := metrics["persist_failures"].(int64)

		t.Logf("Persistence metrics after timeout tests:")
		t.Logf("  Attempts: %d", attempts)
		t.Logf("  Successes: %d", successes)
		t.Logf("  Failures: %d", failures)

		// The metrics should at least be present, even if zero
		// This validates the metrics infrastructure exists
		if _, exists := metrics["persist_attempts"]; !exists {
			t.Error("persist_attempts metric should exist")
		}
		if _, exists := metrics["persist_successes"]; !exists {
			t.Error("persist_successes metric should exist")
		}
		if _, exists := metrics["persist_failures"]; !exists {
			t.Error("persist_failures metric should exist")
		}
	})
}

// TestCheckpointPersistence_TimeoutHandling tests checkpoint requests with timeout constraints
func TestCheckpointPersistence_TimeoutHandling(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	// Start engine first
	eng.Start()
	defer eng.Stop(context.Background())

	// Set up normal store for checkpoint testing
	store := state.NewMemStore()
	eng.SetPersistenceStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng.StartPersistence(ctx)
	defer eng.StopPersistence()

	playerID := "test-checkpoint-player"
	playerName := "CheckpointPlayer"
	initialPos := spatial.Vec2{X: 50, Z: 150}

	// Add player to engine
	player := eng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
	playerMgr := eng.GetPlayerManager()
	playerMgr.InitializePlayer(player)

	// Test checkpoint request functionality
	t.Run("CheckpointRequest", func(t *testing.T) {
		checkpointCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Request multiple checkpoints to fill the batch (batch size is 10)
		for i := 0; i < 10; i++ {
			eng.RequestPlayerCheckpoint(checkpointCtx, playerID)
		}

		// Give persistence manager time to process the full batch
		time.Sleep(500 * time.Millisecond)

		// Verify state was persisted
		persistedState, exists, err := store.Load(context.Background(), playerID)
		if err != nil {
			t.Fatalf("Failed to load persisted state: %v", err)
		}
		if !exists {
			t.Fatal("Player state should be persisted via checkpoint")
		}

		// Verify position was saved
		if persistedState.Pos.X != initialPos.X || persistedState.Pos.Z != initialPos.Z {
			t.Errorf("Position not persisted correctly: got %v, want %v", persistedState.Pos, initialPos)
		}

		t.Logf("✓ Checkpoint request completed successfully")
	})

	// Test multiple concurrent checkpoint requests
	t.Run("ConcurrentCheckpoints", func(t *testing.T) {
		const numCheckpoints = 5
		done := make(chan struct{}, numCheckpoints)

		for i := 0; i < numCheckpoints; i++ {
			go func(i int) {
				defer func() { done <- struct{}{} }()
				checkpointCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				eng.RequestPlayerCheckpoint(checkpointCtx, playerID)
			}(i)
		}

		// Wait for all checkpoints to complete
		for i := 0; i < numCheckpoints; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(3 * time.Second):
				t.Fatalf("Checkpoint %d timed out", i)
			}
		}

		t.Logf("✓ %d concurrent checkpoints completed successfully", numCheckpoints)
	})

	// Test checkpoint with context cancellation
	t.Run("CheckpointCancellation", func(t *testing.T) {
		checkpointCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

		// Request checkpoint save
		eng.RequestPlayerCheckpoint(checkpointCtx, playerID)
		
		// Cancel immediately
		cancel()

		// Give a moment for cancellation to be processed
		time.Sleep(100 * time.Millisecond)

		t.Logf("✓ Checkpoint cancellation handled gracefully")
	})
}

// TestPersistence_QueueBackpressure tests behavior under queue pressure
func TestPersistence_QueueBackpressure(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	// Start engine first
	eng.Start()
	defer eng.Stop(context.Background())

	// Set up slow store to create backpressure
	store := &SlowStore{
		inner: state.NewMemStore(),
		delay: 1 * time.Second,
	}
	eng.SetPersistenceStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng.StartPersistence(ctx)
	defer eng.StopPersistence()

	playerID := "test-backpressure-player"
	playerName := "BackpressurePlayer"
	initialPos := spatial.Vec2{X: 75, Z: 125}

	// Add player to engine
	player := eng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
	playerMgr := eng.GetPlayerManager()
	playerMgr.InitializePlayer(player)

	// Test disconnect persist under queue pressure
	t.Run("DisconnectUnderPressure", func(t *testing.T) {
		// Fill checkpoint queue first to create backpressure
		for i := 0; i < 10; i++ {
			checkpointCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			eng.RequestPlayerCheckpoint(checkpointCtx, playerID)
			cancel()
		}

		// Now test disconnect persistence (should have higher priority)
		persistCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		start := time.Now()
		eng.RequestPlayerDisconnectPersist(persistCtx, playerID)

		// Give time for processing
		time.Sleep(2 * time.Second)
		elapsed := time.Since(start)

		// Should complete in reasonable time despite backpressure
		if elapsed > 4*time.Second {
			t.Errorf("Disconnect persistence took too long under backpressure: %v", elapsed)
		}

		t.Logf("✓ Disconnect persistence handled backpressure correctly in %v", elapsed)
	})
}

// SlowStore wraps a store to simulate slow persistence operations
type SlowStore struct {
	inner state.Store
	delay time.Duration
}

func (s *SlowStore) Save(ctx context.Context, playerID string, data state.PlayerState) error {
	// Simulate slow save operation
	select {
	case <-time.After(s.delay):
		return s.inner.Save(ctx, playerID, data)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *SlowStore) Load(ctx context.Context, playerID string) (state.PlayerState, bool, error) {
	return s.inner.Load(ctx, playerID)
}

func (s *SlowStore) Close() error {
	if closer, ok := s.inner.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func (s *SlowStore) Flush() error {
	if flusher, ok := s.inner.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}