package sim

import (
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
