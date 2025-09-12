package sim

import (
	"prototype-game/backend/internal/spatial"
)

// checkAndHandoverLocked decides whether to move the player to a new cell based on hysteresis.
// e.mu must be held by caller.
func (e *Engine) checkAndHandoverLocked(p *Player) {
	// If player is already inside its owned cell (with hysteresis) do nothing.
	// We require the player to be at least H meters past the border into the new cell.
	cx, cz := spatial.WorldToCell(p.Pos.X, p.Pos.Z, e.cfg.CellSize)
	target := spatial.CellKey{Cx: cx, Cz: cz}
	if target == p.OwnedCell {
		return
	}
	// Verify past hysteresis threshold into target cell.
	if crossedBeyondHysteresis(p.Pos, p.OwnedCell, target, e.cfg.CellSize, e.cfg.HandoverHysteresisM) {
		old := p.OwnedCell
		e.moveEntityLocked(p, old, target)
		p.OwnedCell = target
	}
}

// crossedBeyondHysteresis returns true if pos is sufficiently inside the target cell relative to the origin cell.
func crossedBeyondHysteresis(pos spatial.Vec2, from spatial.CellKey, to spatial.CellKey, cellSize, H float64) bool {
	// Determine which border was crossed and check that the position is at least H beyond that border inside 'to'.
	// Horizontal move (east/west)
	if to.Cx > from.Cx {
		// crossed east border at x = (from.Cx+1)*cellSize
		border := float64(from.Cx+1) * cellSize
		return pos.X >= border+H
	}
	if to.Cx < from.Cx {
		// crossed west border at x = from.Cx*cellSize
		border := float64(from.Cx) * cellSize
		return pos.X <= border-H
	}
	// Vertical move (north/south) using Z axis
	if to.Cz > from.Cz {
		// crossed north border at z = (from.Cz+1)*cellSize
		border := float64(from.Cz+1) * cellSize
		return pos.Z >= border+H
	}
	if to.Cz < from.Cz {
		// crossed south border at z = from.Cz*cellSize
		border := float64(from.Cz) * cellSize
		return pos.Z <= border-H
	}
	return false
}
