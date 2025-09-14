package sim

import (
	"math"
	"time"

	"prototype-game/backend/internal/spatial"
)

const (
	botSpeed = 1.5 // m/s

	retargetMin   = 3 // seconds
	retargetRange = 5 // seconds -> 3-7s
	retargetMax   = retargetMin + retargetRange - 1

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
			if distSq < sepDistSq {
				dist := math.Sqrt(distSq)
				repel.X += dx / dist
				repel.Z += dz / dist
			}
		}
		if repel.X != 0 || repel.Z != 0 {
			mag := math.Hypot(repel.X, repel.Z)
			// Blend repulsion with current direction
			repelDir := spatial.Vec2{X: repel.X / mag, Z: repel.Z / mag}
			blendWander := 0.7
			blendRepel := 0.3
			blended := spatial.Vec2{
				X: st.dir.X*blendWander + repelDir.X*blendRepel,
				Z: st.dir.Z*blendWander + repelDir.Z*blendRepel,
			}
			blendedMag := math.Hypot(blended.X, blended.Z)
			if blendedMag > 0 {
				st.dir = spatial.Vec2{X: blended.X / blendedMag, Z: blended.Z / blendedMag}
			} else {
				st.dir = repelDir
			}
			st.retargetAt = now.Add(time.Duration(retargetMin+e.rng.Intn(retargetRange)) * time.Second)
		}
	}
	if now.After(st.retargetAt) {
		angle := e.rng.Float64() * 2 * math.Pi
		st.dir = spatial.Vec2{X: math.Cos(angle), Z: math.Sin(angle)}
		st.retargetAt = now.Add(time.Duration(retargetMin+e.rng.Intn(retargetRange)) * time.Second)
	}
	b.Vel = spatial.Vec2{X: st.dir.X * botSpeed, Z: st.dir.Z * botSpeed}
}
