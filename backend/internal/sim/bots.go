package sim

import (
	"math"
	"time"

	"prototype-game/backend/internal/spatial"
)

const (
	botSpeed  = 1.5 // m/s
	sepDist   = 2.0 // meters
	sepDistSq = sepDist * sepDist
)

type botState struct {
	dir        spatial.Vec2
	retargetAt time.Time
	OwnedCell  spatial.CellKey
}

// maintainBotDensity no-op for now.
func (e *Engine) maintainBotDensity() {}

// updateBot applies wander behavior with simple separation to avoid clustering.
func (e *Engine) updateBot(b *Entity, dt time.Duration, st *botState) {
	now := time.Now()
	// Separation: steer away from nearby bots (<2m).
	if cell, ok := e.cells[st.OwnedCell]; ok {
		var repel spatial.Vec2
		for id, other := range cell.Entities {
			if id == b.ID || other.Kind != KindBot {
				continue
			}
			dx := b.Pos.X - other.Pos.X
			dz := b.Pos.Z - other.Pos.Z
			distSq := dx*dx + dz*dz
			if distSq < sepDistSq && distSq > 0 {
				dist := math.Sqrt(distSq)
				repel.X += dx / dist
				repel.Z += dz / dist
			}
		}
		if repel.X != 0 || repel.Z != 0 {
			mag := math.Hypot(repel.X, repel.Z)
			st.dir = spatial.Vec2{X: repel.X / mag, Z: repel.Z / mag}
			st.retargetAt = now.Add(time.Duration(3+e.rng.Intn(5)) * time.Second)
		}
	}
	if now.After(st.retargetAt) {
		angle := e.rng.Float64() * 2 * math.Pi
		st.dir = spatial.Vec2{X: math.Cos(angle), Z: math.Sin(angle)}
		st.retargetAt = now.Add(time.Duration(3+e.rng.Intn(5)) * time.Second)
	}
	b.Vel = spatial.Vec2{X: st.dir.X * botSpeed, Z: st.dir.Z * botSpeed}
}
