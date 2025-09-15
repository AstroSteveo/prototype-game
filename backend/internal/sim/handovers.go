package sim

import (
	"sync/atomic"
	"time"

	"prototype-game/backend/internal/metrics"
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

	// Anti-thrash logic: if returning to the previous cell, require 2x hysteresis
	hysteresis := e.cfg.HandoverHysteresisM
	if target == p.PrevCell {
		hysteresis *= 2.0 // Double hysteresis when returning to previous cell
	}

	// Verify past hysteresis threshold into target cell.
	if crossedBeyondHysteresis(p.Pos, p.OwnedCell, target, e.cfg.CellSize, hysteresis) {
		// Capture timestamp immediately when handover condition is detected
		// This ensures accurate latency measurement from detection to client notification
		p.HandoverAt = time.Now()
		old := p.OwnedCell

		// Check if this is a cross-node handover
		targetNodeID := e.nodeRegistry.GetCellOwner(target)
		if targetNodeID != e.nodeRegistry.GetLocalNodeID() {
			// This is a cross-node handover - initiate cross-node transfer
			e.initiateCrossNodeHandoverLocked(p, targetNodeID, old, target)
		} else {
			// Local handover - use existing logic
			e.moveEntityLocked(p, old, target)
			p.PrevCell = p.OwnedCell // Remember the cell we're leaving
			p.OwnedCell = target
			// metrics: record handover (logical ownership change)
			atomic.AddInt64(&e.met.handovers, 1)
			metrics.IncHandovers()
		}
	}
}

// initiateCrossNodeHandoverLocked initiates a cross-node player transfer
// e.mu must be held by caller
func (e *Engine) initiateCrossNodeHandoverLocked(p *Player, targetNodeID string, fromCell, toCell spatial.CellKey) {
	if e.crossNodeSvc == nil {
		// No cross-node service configured - fall back to local handover
		e.moveEntityLocked(p, fromCell, toCell)
		p.PrevCell = p.OwnedCell
		p.OwnedCell = toCell
		atomic.AddInt64(&e.met.handovers, 1)
		metrics.IncHandovers()
		return
	}

	// Mark player as pending cross-node handover
	p.CrossNodeHandover = &CrossNodeHandoverState{
		TargetNode: targetNodeID,
		FromCell:   fromCell,
		ToCell:     toCell,
		Status:     HandoverInProgress,
		InitiatedAt: time.Now(),
	}

	// For now, continue the local handover to avoid state inconsistency
	// The cross-node transfer will be handled asynchronously
	e.moveEntityLocked(p, fromCell, toCell)
	p.PrevCell = p.OwnedCell
	p.OwnedCell = toCell
	atomic.AddInt64(&e.met.handovers, 1)
	metrics.IncHandovers()
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
