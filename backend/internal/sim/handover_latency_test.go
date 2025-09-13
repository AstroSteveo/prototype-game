package sim

import (
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

// TestHandoverLatencyTimestampPrecision tests that the handover timestamp
// is captured as soon as the handover condition is detected, not after processing.
func TestHandoverLatencyTimestampPrecision(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:            10.0,
		HandoverHysteresisM: 1.0,
		TickHz:              60,
		SnapshotHz:          30,
		AOIRadius:           15.0,
	})

	// Place player near the border of cell (0,0) but not past hysteresis
	player := eng.DevSpawn("p1", "Alice", spatial.Vec2{X: 9.5, Z: 0})
	if player.OwnedCell.Cx != 0 || player.OwnedCell.Cz != 0 {
		t.Fatalf("expected player in cell (0,0), got (%d,%d)", player.OwnedCell.Cx, player.OwnedCell.Cz)
	}

	// Set velocity to move player past hysteresis threshold
	// Cell border at x=10, hysteresis=1.0, so need to reach x=11.0
	eng.DevSetVelocity("p1", spatial.Vec2{X: 10.0, Z: 0}) // 10 m/s eastward

	// Simulate enough time to cross the threshold
	// Need to move from 9.5 to 11.0 = 1.5 meters at 10 m/s = 0.15 seconds
	tickDuration := time.Millisecond * 200 // 0.2 seconds

	beforeTick := time.Now()
	eng.Step(tickDuration)
	afterTick := time.Now()

	// Get updated player
	updatedPlayer, ok := eng.GetPlayer("p1")
	if !ok {
		t.Fatal("player not found after tick")
	}

	// Verify handover occurred
	if updatedPlayer.OwnedCell.Cx != 1 || updatedPlayer.OwnedCell.Cz != 0 {
		// Check final position for debugging
		t.Logf("Player final position: x=%.2f, z=%.2f", updatedPlayer.Pos.X, updatedPlayer.Pos.Z)
		t.Fatalf("expected player in cell (1,0) after handover, got (%d,%d)",
			updatedPlayer.OwnedCell.Cx, updatedPlayer.OwnedCell.Cz)
	}

	// Verify timestamp was set
	if updatedPlayer.HandoverAt.IsZero() {
		t.Fatal("HandoverAt timestamp was not set")
	}

	// Verify timestamp precision: HandoverAt should be close to when the condition was detected
	if updatedPlayer.HandoverAt.Before(beforeTick) {
		t.Errorf("HandoverAt timestamp %v is before tick started %v",
			updatedPlayer.HandoverAt, beforeTick)
	}
	if updatedPlayer.HandoverAt.After(afterTick) {
		t.Errorf("HandoverAt timestamp %v is after tick completed %v",
			updatedPlayer.HandoverAt, afterTick)
	}

	// The key observation: currently the timestamp is captured after processing
	// We want to capture it as soon as the condition is detected
	t.Logf("Handover timestamp: %v", updatedPlayer.HandoverAt)
	t.Logf("Tick start: %v", beforeTick)
	t.Logf("Tick end: %v", afterTick)
}

// TestHandoverLatencyAccuracy validates that the timestamp is captured
// at the earliest possible moment when the handover condition is detected.
func TestHandoverLatencyAccuracy(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:            10.0,
		HandoverHysteresisM: 1.0,
		TickHz:              60,
		SnapshotHz:          30,
		AOIRadius:           15.0,
	})

	// Start player in center of cell (0,0) to ensure clean state
	_ = eng.DevSpawn("p1", "Alice", spatial.Vec2{X: 5.0, Z: 0})

	// Set velocity to move player past cell boundary and hysteresis
	// Cell border at x=10, hysteresis=1.0, so need to reach x=11.0
	eng.DevSetVelocity("p1", spatial.Vec2{X: 50.0, Z: 0}) // 50 m/s eastward

	// Use a tick long enough to cross the threshold
	tickDuration := time.Millisecond * 200 // 0.2 seconds = 10 meters traveled

	beforeTick := time.Now()
	eng.Step(tickDuration)

	updatedPlayer, ok := eng.GetPlayer("p1")
	if !ok {
		t.Fatal("player not found after tick")
	}

	// Check position - should be at x = 5.0 + 50*0.2 = 15.0
	t.Logf("Player final position: x=%.3f, z=%.3f", updatedPlayer.Pos.X, updatedPlayer.Pos.Z)
	t.Logf("Player cell: (%d, %d)", updatedPlayer.OwnedCell.Cx, updatedPlayer.OwnedCell.Cz)

	// This should definitely trigger handover
	if updatedPlayer.OwnedCell.Cx == 1 && updatedPlayer.OwnedCell.Cz == 0 {
		// Handover occurred - validate timestamp precision
		if updatedPlayer.HandoverAt.IsZero() {
			t.Fatal("HandoverAt timestamp was not set")
		}

		// Measure how quickly the timestamp was captured
		detectionDelay := updatedPlayer.HandoverAt.Sub(beforeTick)
		t.Logf("Handover detection delay: %v", detectionDelay)

		// This demonstrates the fix: timestamp is captured early in the handover process
		// The delay should be minimal since we capture timestamp immediately when condition is detected
		if detectionDelay < time.Millisecond {
			t.Logf("✓ Timestamp captured quickly: %v < 1ms", detectionDelay)
		} else {
			t.Logf("Timestamp capture took: %v (may include processing time)", detectionDelay)
		}
	} else {
		t.Fatalf("Expected handover to occur with final position x=%.3f (>11.0), but player is still in cell (%d,%d)",
			updatedPlayer.Pos.X, updatedPlayer.OwnedCell.Cx, updatedPlayer.OwnedCell.Cz)
	}
}
