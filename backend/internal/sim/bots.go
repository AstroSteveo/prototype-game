package sim

import (
	"math"
	"math/rand"
	"time"

	"prototype-game/backend/internal/spatial"
)

// Placeholder bot logic for MVP skeleton. Does nothing yet but holds structure.

const botSpeed = 1.5 // m/s

type botState struct {
	dir        spatial.Vec2
	retargetAt time.Time
}

// maintainBotDensity no-op for now.
func (e *Engine) maintainBotDensity() {}

// updateBot no-op for now.
func (e *Engine) updateBot(b *Entity, dt time.Duration, st *botState) {
	if time.Now().After(st.retargetAt) {
		// random unit vector
		angle := rand.Float64() * 2 * math.Pi
		st.dir = spatial.Vec2{X: math.Cos(angle), Z: math.Sin(angle)}
		st.retargetAt = time.Now().Add(time.Duration(3+rand.Intn(5)) * time.Second)
	}
	b.Vel = spatial.Vec2{X: st.dir.X * botSpeed, Z: st.dir.Z * botSpeed}
}
