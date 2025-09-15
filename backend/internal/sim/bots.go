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
	var appliedSeparation bool
	if cell, ok := e.cells[st.OwnedCell]; ok {
		var repel spatial.Vec2
		minDistSq := sepDistSq
		for id, other := range cell.Entities {
			if id == b.ID || other.Kind != KindBot {
				continue
			}
			dx := b.Pos.X - other.Pos.X
			dz := b.Pos.Z - other.Pos.Z
			distSq := dx*dx + dz*dz
			if distSq < sepDistSq {
				dist := math.Sqrt(distSq)
				if distSq < minDistSq {
					minDistSq = distSq
				}
				// Stronger repulsion for closer bots
				strength := sepDist / dist // 1/d scaling
				repel.X += dx * strength / dist
				repel.Z += dz * strength / dist
			}
		}
		if repel.X != 0 || repel.Z != 0 {
			mag := math.Hypot(repel.X, repel.Z)
			repelDir := spatial.Vec2{X: repel.X / mag, Z: repel.Z / mag}

			// For very close bots (< 1.2m), use pure separation
			minDist := math.Sqrt(minDistSq)
			if minDist < 1.2 {
				st.dir = repelDir
			} else {
				// Handle uninitialized direction (zero vector)
				if st.dir.X == 0 && st.dir.Z == 0 {
					// Pure separation when no existing direction
					st.dir = repelDir
				} else {
					// Adaptive blending: stronger separation for closer bots
					separationStrength := math.Min(1.0, sepDist/minDist) // 0.0 at sepDist, 1.0 at 0 distance
					blendRepel := 0.8 + 0.15*separationStrength          // 0.8 to 0.95 based on proximity
					blendWander := 1.0 - blendRepel

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
				}
			}
			// Delay retargeting while actively separating
			st.retargetAt = now.Add(time.Duration(retargetMin+e.rng.Intn(retargetRange)) * time.Second)
			appliedSeparation = true
		}
	}

	// Random retargeting when not actively separating
	// Only retarget if we have a valid retarget time set and it has passed
	if !appliedSeparation && !st.retargetAt.IsZero() && now.After(st.retargetAt) {
		angle := e.rng.Float64() * 2 * math.Pi
		st.dir = spatial.Vec2{X: math.Cos(angle), Z: math.Sin(angle)}
		st.retargetAt = now.Add(time.Duration(retargetMin+e.rng.Intn(retargetRange)) * time.Second)
	}

	// Ensure bot has direction - for initial updates in tests
	if st.dir.X == 0 && st.dir.Z == 0 && st.retargetAt.IsZero() {
		angle := e.rng.Float64() * 2 * math.Pi
		st.dir = spatial.Vec2{X: math.Cos(angle), Z: math.Sin(angle)}
		st.retargetAt = now.Add(time.Duration(retargetMin+e.rng.Intn(retargetRange)) * time.Second)
	}

	b.Vel = spatial.Vec2{X: st.dir.X * botSpeed, Z: st.dir.Z * botSpeed}
}
