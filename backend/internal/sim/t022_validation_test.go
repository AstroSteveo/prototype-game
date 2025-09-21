package sim

import (
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

// TestT022_HandoverHysteresisAndAntiThrash validates the specific acceptance criteria for T-022:
// - Handover latency measured; event emitted once per change.
// - Validate hysteresis H and doubled hysteresis when returning.
func TestT022_HandoverHysteresisAndAntiThrash(t *testing.T) {
	t.Run("ValidateHysteresisH", func(t *testing.T) {
		eng := NewEngine(Config{
			CellSize:            10.0,
			HandoverHysteresisM: 2.0, // H = 2 meters
			TickHz:              60,
			SnapshotHz:          30,
			AOIRadius:           15.0,
		})

		// Start player in center of cell (0,0)
		_ = eng.DevSpawn("p1", "Alice", spatial.Vec2{X: 5.0, Z: 5.0})
		initialHandovers := eng.MetricsSnapshot().Handovers

		// Move player close to border but not past hysteresis
		// Border at x=10, hysteresis=2.0, so position at x=11.5 should NOT trigger handover
		eng.DevSetVelocity("p1", spatial.Vec2{X: 50.0, Z: 0}) // fast eastward
		eng.Step(time.Millisecond * 130)                      // Move to ~x=11.5

		player1, _ := eng.GetPlayer("p1")
		handovers1 := eng.MetricsSnapshot().Handovers

		t.Logf("Position x=%.2f (border at x=10, hysteresis=2.0)", player1.Pos.X)

		// Should NOT have handover yet since x=11.5 < (border + hysteresis) = 12.0
		if player1.Pos.X < 12.0 && handovers1 > initialHandovers {
			t.Errorf("Handover occurred too early at x=%.2f (should wait until x≥12.0)", player1.Pos.X)
		}

		// Continue moving to definitely cross hysteresis threshold
		eng.Step(time.Millisecond * 20) // Move more to x ≈ 12.5

		player2, _ := eng.GetPlayer("p1")
		handovers2 := eng.MetricsSnapshot().Handovers

		t.Logf("Final position x=%.2f, handovers: %d → %d", player2.Pos.X, initialHandovers, handovers2)

		// Should have handover now since we're past x=12.0
		if player2.Pos.X >= 12.0 && handovers2 == initialHandovers {
			t.Errorf("Expected handover at x=%.2f (past threshold x≥12.0)", player2.Pos.X)
		}

		// Verify handover occurred to correct cell
		if player2.OwnedCell.Cx != 1 || player2.OwnedCell.Cz != 0 {
			t.Errorf("Expected player in cell (1,0), got (%d,%d)", player2.OwnedCell.Cx, player2.OwnedCell.Cz)
		}

		t.Logf("✓ Hysteresis H=2.0 validated: handover at x=%.2f", player2.Pos.X)
	})

	t.Run("ValidateDoubleHysteresisWhenReturning", func(t *testing.T) {
		eng := NewEngine(Config{
			CellSize:            10.0,
			HandoverHysteresisM: 1.0, // H = 1 meter, so double = 2 meters
			TickHz:              60,
			SnapshotHz:          30,
			AOIRadius:           15.0,
		})

		// Start player and move to cell (1,0)
		_ = eng.DevSpawn("p1", "Alice", spatial.Vec2{X: 5.0, Z: 5.0})
		eng.DevSetVelocity("p1", spatial.Vec2{X: 50.0, Z: 0})
		eng.Step(time.Millisecond * 140) // Move to x ≈ 12.0, triggering handover to (1,0)

		player1, _ := eng.GetPlayer("p1")
		if player1.OwnedCell.Cx != 1 {
			t.Fatalf("Expected initial handover to cell (1,0), got (%d,%d)", player1.OwnedCell.Cx, player1.OwnedCell.Cz)
		}

		// Now move back west past normal hysteresis (1.0) but not past double hysteresis (2.0)
		// Border from (1,0) to (0,0) is at x=10
		// Normal hysteresis: need x ≤ 9.0
		// Double hysteresis: need x ≤ 8.0
		eng.DevSetVelocity("p1", spatial.Vec2{X: -50.0, Z: 0})
		eng.Step(time.Millisecond * 60) // Move to x ≈ 9.0

		player2, _ := eng.GetPlayer("p1")
		handovers2 := eng.MetricsSnapshot().Handovers

		t.Logf("After return move: pos=(%.2f,%.2f) cell=(%d,%d) handovers=%d",
			player2.Pos.X, player2.Pos.Z, player2.OwnedCell.Cx, player2.OwnedCell.Cz, handovers2)

		// Should still be in cell (1,0) due to double hysteresis preventing thrash
		if player2.Pos.X > 8.0 && player2.OwnedCell.Cx != 1 {
			t.Errorf("Double hysteresis failed: player at x=%.2f should still be in cell (1,0), got (%d,%d)",
				player2.Pos.X, player2.OwnedCell.Cx, player2.OwnedCell.Cz)
		}

		// Now move further west to trigger double hysteresis
		eng.Step(time.Millisecond * 40) // Move to x ≈ 7.0, past double hysteresis

		player3, _ := eng.GetPlayer("p1")
		handovers3 := eng.MetricsSnapshot().Handovers

		t.Logf("After far west move: pos=(%.2f,%.2f) cell=(%d,%d) handovers=%d",
			player3.Pos.X, player3.Pos.Z, player3.OwnedCell.Cx, player3.OwnedCell.Cz, handovers3)

		// Should now be in cell (0,0) after crossing double hysteresis
		if player3.Pos.X <= 8.0 && handovers3 <= handovers2 {
			t.Errorf("Expected handover when crossing double hysteresis at x=%.2f", player3.Pos.X)
		}

		t.Logf("✓ Double hysteresis validated: prevented thrash, then allowed handover at x=%.2f", player3.Pos.X)
	})

	t.Run("ValidateHandoverLatencyMeasuredAndEventEmittedOnce", func(t *testing.T) {
		eng := NewEngine(Config{
			CellSize:            10.0,
			HandoverHysteresisM: 1.0,
			TickHz:              60,
			SnapshotHz:          30,
			AOIRadius:           15.0,
		})

		// Start player and measure handover latency
		_ = eng.DevSpawn("p1", "Alice", spatial.Vec2{X: 5.0, Z: 5.0})
		initialHandovers := eng.MetricsSnapshot().Handovers

		beforeMove := time.Now()
		eng.DevSetVelocity("p1", spatial.Vec2{X: 50.0, Z: 0})
		eng.Step(time.Millisecond * 140) // Trigger handover

		player1, _ := eng.GetPlayer("p1")
		finalHandovers := eng.MetricsSnapshot().Handovers

		// Validate handover occurred
		if finalHandovers <= initialHandovers {
			t.Fatalf("Expected handover to occur, but handover count did not increase")
		}

		// Validate handover latency timestamp was set
		if player1.HandoverAt.IsZero() {
			t.Fatalf("HandoverAt timestamp was not set - latency not measured")
		}

		// Validate latency measurement timing
		handoverLatency := player1.HandoverAt.Sub(beforeMove)
		if handoverLatency <= 0 || handoverLatency > 500*time.Millisecond {
			t.Errorf("Handover latency %v seems unreasonable (should be >0 and <500ms)", handoverLatency)
		}

		// Validate exactly one handover event per change
		expectedHandovers := initialHandovers + 1
		if finalHandovers != expectedHandovers {
			t.Errorf("Expected exactly one handover event, got %d (from %d to %d)",
				finalHandovers-initialHandovers, initialHandovers, finalHandovers)
		}

		t.Logf("✓ Handover latency measured: %v", handoverLatency)
		t.Logf("✓ Event emitted once per change: %d → %d", initialHandovers, finalHandovers)
	})
}

// TestT022_AcceptanceCriteria validates the exact acceptance criteria:
// "Handover latency measured; event emitted once per change."
func TestT022_AcceptanceCriteria(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:            10.0,
		HandoverHysteresisM: 1.0,
		TickHz:              60,
		SnapshotHz:          30,
		AOIRadius:           15.0,
	})

	// Start player in cell (0,0)
	_ = eng.DevSpawn("p1", "TestPlayer", spatial.Vec2{X: 5.0, Z: 5.0})

	// Record initial state
	initialHandovers := eng.MetricsSnapshot().Handovers

	// Move player to trigger exactly one handover
	beforeHandover := time.Now()
	eng.DevSetVelocity("p1", spatial.Vec2{X: 50.0, Z: 0}) // fast eastward
	eng.Step(time.Millisecond * 140)                      // cross border + hysteresis

	// Check final state
	finalPlayer, ok := eng.GetPlayer("p1")
	if !ok {
		t.Fatal("Player not found after movement")
	}
	finalHandovers := eng.MetricsSnapshot().Handovers

	// ACCEPTANCE CRITERIA VALIDATION:

	// 1. "Handover latency measured"
	if finalPlayer.HandoverAt.IsZero() {
		t.Error("❌ ACCEPTANCE CRITERIA FAILED: Handover latency not measured (HandoverAt timestamp not set)")
	} else {
		latency := finalPlayer.HandoverAt.Sub(beforeHandover)
		if latency <= 0 {
			t.Error("❌ ACCEPTANCE CRITERIA FAILED: Invalid handover latency measurement")
		} else {
			t.Logf("✅ ACCEPTANCE CRITERIA PASSED: Handover latency measured (%v)", latency)
		}
	}

	// 2. "event emitted once per change"
	handoverCount := finalHandovers - initialHandovers
	if handoverCount != 1 {
		t.Errorf("❌ ACCEPTANCE CRITERIA FAILED: Expected exactly 1 handover event, got %d", handoverCount)
	} else {
		t.Logf("✅ ACCEPTANCE CRITERIA PASSED: Event emitted once per change (%d handover)", handoverCount)
	}

	// Verify the handover actually occurred
	if finalPlayer.OwnedCell.Cx != 1 || finalPlayer.OwnedCell.Cz != 0 {
		t.Errorf("Expected handover to cell (1,0), got (%d,%d)", finalPlayer.OwnedCell.Cx, finalPlayer.OwnedCell.Cz)
	}

	t.Log("🎯 T-022 ACCEPTANCE CRITERIA VALIDATION COMPLETE")
}
