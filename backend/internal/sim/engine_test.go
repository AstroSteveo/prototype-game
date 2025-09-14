package sim

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func newTestEngine() *Engine {
	return NewEngine(Config{
		CellSize:            10,
		AOIRadius:           5,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
}

func TestIntegratesVelocityStep(t *testing.T) {
	e := newTestEngine()
	p := e.DevSpawn("p3", "Carol", spatial.Vec2{X: 0, Z: 0})
	if !e.DevSetVelocity(p.ID, spatial.Vec2{X: 1, Z: -0.5}) {
		t.Fatalf("failed to set velocity")
	}
	// Advance 2 seconds
	e.Step(2 * time.Second)
	snap, ok := e.GetPlayer(p.ID)
	if !ok {
		t.Fatalf("player not found")
	}
	if abs(snap.Pos.X-2.0) > 1e-6 || abs(snap.Pos.Z+1.0) > 1e-6 {
		t.Fatalf("unexpected position after integrate: %#v", snap.Pos)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestAddOrUpdatePlacesPlayerInCorrectCell(t *testing.T) {
	e := newTestEngine()
	p := e.AddOrUpdatePlayer("p1", "Alice", spatial.Vec2{X: 9.9, Z: -0.1}, spatial.Vec2{})
	if p.OwnedCell.Cx != 0 || p.OwnedCell.Cz != -1 {
		t.Fatalf("expected OwnedCell (0,-1), got (%d,%d)", p.OwnedCell.Cx, p.OwnedCell.Cz)
	}
}

func TestHandoverAfterHysteresis(t *testing.T) {
	e := newTestEngine()
	// Start near east border of cell (0,0)
	p := e.AddOrUpdatePlayer("p2", "Bob", spatial.Vec2{X: 9.9, Z: 0}, spatial.Vec2{})
	if p.OwnedCell.Cx != 0 || p.OwnedCell.Cz != 0 {
		t.Fatalf("expected OwnedCell (0,0), got (%d,%d)", p.OwnedCell.Cx, p.OwnedCell.Cz)
	}
	// Move east at 1 m/s. Hysteresis = 2m, border = x=10; need x >= 12 to transfer.
	e.DevSetVelocity("p2", spatial.Vec2{X: 1, Z: 0})

	// Step 1.0s: x = 10.9; should still be in cell (0,0)
	e.Step(1 * time.Second)
	pSnap, _ := e.GetPlayer("p2")
	if pSnap.OwnedCell.Cx != 0 || pSnap.OwnedCell.Cz != 0 {
		t.Fatalf("handover occurred too early at x=%.2f; owned=(%d,%d)", pSnap.Pos.X, pSnap.OwnedCell.Cx, pSnap.OwnedCell.Cz)
	}

	// Step another 0.2s: x = 11.1; still below 12, should still not handover
	e.Step(200 * time.Millisecond)
	pSnap, _ = e.GetPlayer("p2")
	if pSnap.OwnedCell.Cx != 0 {
		t.Fatalf("handover occurred too early at x=%.2f; owned=(%d,%d)", pSnap.Pos.X, pSnap.OwnedCell.Cx, pSnap.OwnedCell.Cz)
	}

	// Step 1.0s more: x = 12.1; should handover to (1,0)
	e.Step(1 * time.Second)
	pSnap, _ = e.GetPlayer("p2")
	if pSnap.OwnedCell.Cx != 1 || pSnap.OwnedCell.Cz != 0 {
		t.Fatalf("expected handover to (1,0), got (%d,%d) at x=%.2f", pSnap.OwnedCell.Cx, pSnap.OwnedCell.Cz, pSnap.Pos.X)
	}
}

func TestEngine_StopTwiceIsIdempotent(t *testing.T) {
	e := newTestEngine()
	e.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// First stop should cleanly terminate the loop
	e.Stop(ctx)
	// Second stop should return immediately without panic
	done := make(chan struct{})
	go func() { e.Stop(ctx); close(done) }()
	select {
	case <-done:
		// ok
	case <-ctx.Done():
		t.Fatalf("second Stop did not return before context deadline")
	}
}

func TestEngine_StartTwiceIsIdempotent(t *testing.T) {
	e := newTestEngine()
	// Calling Start multiple times must not panic or create issues
	e.Start()
	e.Start()
	// Give it a moment to spin
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e.Stop(ctx)
}

func TestEngine_StopWithoutStartReturns(t *testing.T) {
	e := newTestEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	// Should be a no-op and return promptly
	start := time.Now()
	e.Stop(ctx)
	if time.Since(start) > 300*time.Millisecond {
		t.Fatalf("Stop without Start took too long")
	}
}
