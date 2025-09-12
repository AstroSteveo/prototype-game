package sim

import (
	"prototype-game/backend/internal/spatial"
	"testing"
)

func TestCrossedBeyondHysteresis(t *testing.T) {
	cell := 10.0
	H := 2.0
	from := spatial.CellKey{Cx: 0, Cz: 0}
	toE := spatial.CellKey{Cx: 1, Cz: 0}
	// crossing east: border at x=10; require x>=12
	if crossedBeyondHysteresis(spatial.Vec2{X: 11.9, Z: 0}, from, toE, cell, H) {
		t.Fatalf("should not handover before hysteresis")
	}
	if !crossedBeyondHysteresis(spatial.Vec2{X: 12.0, Z: 0}, from, toE, cell, H) {
		t.Fatalf("should handover after hysteresis")
	}

	toW := spatial.CellKey{Cx: -1, Cz: 0}
	// crossing west: border at x=0; require x<=-2
	if crossedBeyondHysteresis(spatial.Vec2{X: -1.9, Z: 0}, from, toW, cell, H) {
		t.Fatalf("should not handover before hysteresis (west)")
	}
	if !crossedBeyondHysteresis(spatial.Vec2{X: -2.0, Z: 0}, from, toW, cell, H) {
		t.Fatalf("should handover after hysteresis (west)")
	}

	toN := spatial.CellKey{Cx: 0, Cz: 1}
	// crossing north: border at z=10; require z>=12
	if crossedBeyondHysteresis(spatial.Vec2{X: 0, Z: 11.9}, from, toN, cell, H) {
		t.Fatalf("should not handover before hysteresis (north)")
	}
	if !crossedBeyondHysteresis(spatial.Vec2{X: 0, Z: 12.0}, from, toN, cell, H) {
		t.Fatalf("should handover after hysteresis (north)")
	}

	toS := spatial.CellKey{Cx: 0, Cz: -1}
	// crossing south: border at z=0; require z<=-2
	if crossedBeyondHysteresis(spatial.Vec2{X: 0, Z: -1.9}, from, toS, cell, H) {
		t.Fatalf("should not handover before hysteresis (south)")
	}
	if !crossedBeyondHysteresis(spatial.Vec2{X: 0, Z: -2.0}, from, toS, cell, H) {
		t.Fatalf("should handover after hysteresis (south)")
	}
}
