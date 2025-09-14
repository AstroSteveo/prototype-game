package sim

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

// TestBotSeparation ensures bots steer away when within 2 meters and move at clamped speed.
func TestBotSeparation(t *testing.T) {
	e := NewEngine(Config{CellSize: 10})
	e.rng = rand.New(rand.NewSource(1))

	k := spatial.CellKey{Cx: 0, Cz: 0}
	e.mu.Lock()
	c := e.getOrCreateCellLocked(k)
	b1 := &Entity{ID: "bot-1", Kind: KindBot, Pos: spatial.Vec2{X: 0, Z: 0}}
	b2 := &Entity{ID: "bot-2", Kind: KindBot, Pos: spatial.Vec2{X: 1, Z: 0}}
	c.Entities[b1.ID] = b1
	c.Entities[b2.ID] = b2
	e.bots[b1.ID] = &botState{OwnedCell: k}
	e.bots[b2.ID] = &botState{OwnedCell: k}
	e.mu.Unlock()

	e.Step(1 * time.Second)

	v1 := math.Hypot(b1.Vel.X, b1.Vel.Z)
	v2 := math.Hypot(b2.Vel.X, b2.Vel.Z)
	if math.Abs(v1-botSpeed) > 1e-9 || math.Abs(v2-botSpeed) > 1e-9 {
		t.Fatalf("bot speed not clamped: %v, %v", v1, v2)
	}
	dist := math.Hypot(b1.Pos.X-b2.Pos.X, b1.Pos.Z-b2.Pos.Z)
	if dist <= 2 {
		t.Fatalf("bots did not separate, dist=%.2f", dist)
	}
	if d := time.Until(e.bots[b1.ID].retargetAt); d < time.Duration(retargetMin)*time.Second || d > time.Duration(retargetMax)*time.Second {
		t.Fatalf("retarget window out of range: %v", d)
	}
}

// TestBotWanderRetargetTiming validates that bots retarget direction within 3-7 second windows.
func TestBotWanderRetargetTiming(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"seed-1", 1},
		{"seed-42", 42},
		{"seed-100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(Config{CellSize: 50})
			e.rng = rand.New(rand.NewSource(tt.seed))

			k := spatial.CellKey{Cx: 0, Cz: 0}
			e.mu.Lock()
			c := e.getOrCreateCellLocked(k)
			bot := &Entity{ID: "wander-bot", Kind: KindBot, Pos: spatial.Vec2{X: 25, Z: 25}}
			c.Entities[bot.ID] = bot
			st := &botState{OwnedCell: k}
			e.bots[bot.ID] = st
			e.mu.Unlock()

			// Force initial retarget
			e.Step(1 * time.Millisecond)
			initialRetarget := st.retargetAt

			// Verify initial retarget time is within 3-7s from now
			now := time.Now()
			retargetDelay := initialRetarget.Sub(now)
			// Allow small tolerance for timing precision
			const tolerance = 10 * time.Millisecond
			if retargetDelay < 3*time.Second-tolerance || retargetDelay > 7*time.Second+tolerance {
				t.Errorf("initial retarget delay out of range: %v (expected 3-7s)", retargetDelay)
			}

			// Simulate reaching retarget time
			e.mu.Lock()
			st.retargetAt = now.Add(-1 * time.Second) // Force retarget on next step
			e.mu.Unlock()

			e.Step(1 * time.Millisecond)
			newRetarget := st.retargetAt

			// Verify new retarget time is again within 3-7s
			newDelay := newRetarget.Sub(time.Now())
			if newDelay < 3*time.Second-tolerance || newDelay > 7*time.Second+tolerance {
				t.Errorf("retarget delay out of range: %v (expected 3-7s)", newDelay)
			}
		})
	}
}

// TestBotSpeedClamping validates that bot speed is always clamped to botSpeed regardless of direction changes.
func TestBotSpeedClamping(t *testing.T) {
	e := NewEngine(Config{CellSize: 10})
	e.rng = rand.New(rand.NewSource(1))

	k := spatial.CellKey{Cx: 0, Cz: 0}
	e.mu.Lock()
	c := e.getOrCreateCellLocked(k)
	bot := &Entity{ID: "speed-bot", Kind: KindBot, Pos: spatial.Vec2{X: 5, Z: 5}}
	c.Entities[bot.ID] = bot
	st := &botState{OwnedCell: k}
	e.bots[bot.ID] = st
	e.mu.Unlock()

	// Test multiple simulation steps to ensure speed remains clamped
	for i := 0; i < 10; i++ {
		e.Step(100 * time.Millisecond)
		speed := math.Hypot(bot.Vel.X, bot.Vel.Z)
		if math.Abs(speed-botSpeed) > 1e-9 {
			t.Fatalf("step %d: bot speed not clamped: got %.6f, want %.6f", i, speed, botSpeed)
		}
	}
}

// TestBotClusteringPrevention ensures multiple bots spread out and don't cluster tightly.
func TestBotClusteringPrevention(t *testing.T) {
	e := NewEngine(Config{CellSize: 20})
	e.rng = rand.New(rand.NewSource(42))

	k := spatial.CellKey{Cx: 0, Cz: 0}
	e.mu.Lock()
	c := e.getOrCreateCellLocked(k)

	// Create 4 bots clustered in the center
	bots := make([]*Entity, 4)
	for i := 0; i < 4; i++ {
		bot := &Entity{
			ID:   fmt.Sprintf("cluster-bot-%d", i),
			Kind: KindBot,
			Pos:  spatial.Vec2{X: 10 + float64(i)*0.5, Z: 10 + float64(i)*0.3}, // Start clustered
		}
		bots[i] = bot
		c.Entities[bot.ID] = bot
		e.bots[bot.ID] = &botState{OwnedCell: k}
	}
	e.mu.Unlock()

	// Run simulation for several seconds to allow separation
	for step := 0; step < 50; step++ {
		e.Step(100 * time.Millisecond)
	}

	// Verify bots have spread out - check all pairwise distances
	for i := 0; i < len(bots); i++ {
		for j := i + 1; j < len(bots); j++ {
			dist := math.Hypot(bots[i].Pos.X-bots[j].Pos.X, bots[i].Pos.Z-bots[j].Pos.Z)
			if dist < sepDist {
				t.Errorf("bots %d and %d too close: %.2fm (expected >= %.2fm)", i, j, dist, sepDist)
			}
		}
	}
}

// TestBotSeparationDeterministic validates separation behavior with deterministic RNG.
func TestBotSeparationDeterministic(t *testing.T) {
	// Test the same scenario with the same seed should produce identical results
	runTest := func(seed int64) (spatial.Vec2, spatial.Vec2) {
		e := NewEngine(Config{CellSize: 10})
		e.rng = rand.New(rand.NewSource(seed))

		k := spatial.CellKey{Cx: 0, Cz: 0}
		e.mu.Lock()
		c := e.getOrCreateCellLocked(k)
		b1 := &Entity{ID: "det-bot-1", Kind: KindBot, Pos: spatial.Vec2{X: 0, Z: 0}}
		b2 := &Entity{ID: "det-bot-2", Kind: KindBot, Pos: spatial.Vec2{X: 1.5, Z: 0}}
		c.Entities[b1.ID] = b1
		c.Entities[b2.ID] = b2
		e.bots[b1.ID] = &botState{OwnedCell: k}
		e.bots[b2.ID] = &botState{OwnedCell: k}
		e.mu.Unlock()

		e.Step(500 * time.Millisecond)
		return b1.Pos, b2.Pos
	}

	// Run the same test twice with the same seed
	pos1a, pos2a := runTest(12345)
	pos1b, pos2b := runTest(12345)

	// Results should be identical with deterministic RNG
	const eps = 1e-9
	if math.Abs(pos1a.X-pos1b.X) > eps || math.Abs(pos1a.Z-pos1b.Z) > eps {
		t.Errorf("bot1 position not deterministic: %v vs %v", pos1a, pos1b)
	}
	if math.Abs(pos2a.X-pos2b.X) > eps || math.Abs(pos2a.Z-pos2b.Z) > eps {
		t.Errorf("bot2 position not deterministic: %v vs %v", pos2a, pos2b)
	}
}
