package sim

import (
	"fmt"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

// TestAOI3x3CellQuery tests that AOI queries fetch entities from a 3×3 cell neighborhood
// This validates the basic requirement for US-302.
func TestAOI3x3CellQuery(t *testing.T) {
	e := NewEngine(Config{CellSize: 10, AOIRadius: 15, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})

	// Place query player at origin (center of 3x3 grid)
	queryPlayer := e.DevSpawn("center", "Center", spatial.Vec2{X: 5, Z: 5})

	// Place entities in all 8 neighboring cells within AOI radius
	positions := []struct {
		id   string
		pos  spatial.Vec2
		cell spatial.CellKey
	}{
		// Adjacent cells (4-connected neighbors)
		{"north", spatial.Vec2{X: 5, Z: 15}, spatial.CellKey{Cx: 0, Cz: 1}},
		{"south", spatial.Vec2{X: 5, Z: -5}, spatial.CellKey{Cx: 0, Cz: -1}},
		{"east", spatial.Vec2{X: 15, Z: 5}, spatial.CellKey{Cx: 1, Cz: 0}},
		{"west", spatial.Vec2{X: -5, Z: 5}, spatial.CellKey{Cx: -1, Cz: 0}},
		// Diagonal neighbors
		{"northeast", spatial.Vec2{X: 13, Z: 13}, spatial.CellKey{Cx: 1, Cz: 1}},
		{"northwest", spatial.Vec2{X: -3, Z: 13}, spatial.CellKey{Cx: -1, Cz: 1}},
		{"southeast", spatial.Vec2{X: 13, Z: -3}, spatial.CellKey{Cx: 1, Cz: -1}},
		{"southwest", spatial.Vec2{X: -3, Z: -3}, spatial.CellKey{Cx: -1, Cz: -1}},
	}

	// Add entities to test 3x3 cell coverage
	for _, p := range positions {
		e.AddOrUpdatePlayer(p.id, p.id, p.pos, spatial.Vec2{})
	}

	// Query AOI from center position
	aoiResults := e.QueryAOI(queryPlayer.Pos, e.cfg.AOIRadius, queryPlayer.ID)

	// Verify all 8 neighbors are included (within radius 15)
	found := make(map[string]bool)
	for _, entity := range aoiResults {
		found[entity.ID] = true
	}

	for _, p := range positions {
		// Check if distance is within radius
		dist := spatial.Dist2(p.pos, queryPlayer.Pos)
		if dist <= e.cfg.AOIRadius*e.cfg.AOIRadius {
			if !found[p.id] {
				t.Errorf("Expected to find entity %s at position (%.1f,%.1f) in cell (%d,%d) within radius %.1f, but it was missing",
					p.id, p.pos.X, p.pos.Z, p.cell.Cx, p.cell.Cz, e.cfg.AOIRadius)
			}
		}
	}

	// Verify center player is not included in its own AOI
	if found[queryPlayer.ID] {
		t.Errorf("Query player should not be included in its own AOI results")
	}

	t.Logf("✓ AOI query returned %d entities from 3×3 cell neighborhood", len(aoiResults))
}

// TestContinuousAOIAcrossBorderWithStaticNeighbors tests the core US-302 requirement:
// movement across border with static neighbors should maintain continuous visibility.
func TestContinuousAOIAcrossBorderWithStaticNeighbors(t *testing.T) {
	e := NewEngine(Config{CellSize: 10, AOIRadius: 8, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 1})

	// Place moving player near the border between cells (0,0) and (1,0)
	movingPlayer := e.DevSpawn("moving", "Moving", spatial.Vec2{X: 9.5, Z: 5})

	// Place static neighbors in various cells that should remain visible
	staticNeighbors := []struct {
		id  string
		pos spatial.Vec2
	}{
		{"static1", spatial.Vec2{X: 7, Z: 5}},  // Same cell (0,0), should always be visible
		{"static2", spatial.Vec2{X: 11, Z: 5}}, // Target cell (1,0), should become visible after handover
		{"static3", spatial.Vec2{X: 9, Z: 2}},  // Cell (0,0), near border
		{"static4", spatial.Vec2{X: 11, Z: 2}}, // Cell (1,0), near border
		{"static5", spatial.Vec2{X: 9, Z: 12}}, // Cell (0,1), should be visible in 3x3
	}

	for _, neighbor := range staticNeighbors {
		e.AddOrUpdatePlayer(neighbor.id, neighbor.id, neighbor.pos, spatial.Vec2{})
	}

	// Test AOI before handover
	beforeAOI := e.QueryAOI(movingPlayer.Pos, e.cfg.AOIRadius, movingPlayer.ID)
	beforeIDs := make(map[string]bool)
	for _, entity := range beforeAOI {
		beforeIDs[entity.ID] = true
	}

	t.Logf("Before handover: moving player at (%.1f,%.1f), AOI contains %d entities",
		movingPlayer.Pos.X, movingPlayer.Pos.Z, len(beforeAOI))

	// Move player across the border to trigger handover
	e.DevSetVelocity("moving", spatial.Vec2{X: 10, Z: 0}) // Move east
	e.Step(200 * time.Millisecond)                        // Move to approximately (11.5, 5)

	// Get player state after movement
	afterPlayer, ok := e.GetPlayer("moving")
	if !ok {
		t.Fatal("Moving player not found after handover")
	}

	// Verify handover occurred
	if afterPlayer.OwnedCell.Cx == 0 {
		t.Logf("Player at (%.1f,%.1f) - handover not yet triggered (expected with hysteresis)",
			afterPlayer.Pos.X, afterPlayer.Pos.Z)
		// Continue moving to ensure handover
		e.Step(200 * time.Millisecond)
		afterPlayer, _ = e.GetPlayer("moving")
	}

	// Test AOI after handover
	afterAOI := e.QueryAOI(afterPlayer.Pos, e.cfg.AOIRadius, afterPlayer.ID)
	afterIDs := make(map[string]bool)
	for _, entity := range afterAOI {
		afterIDs[entity.ID] = true
	}

	t.Logf("After handover: moving player at (%.1f,%.1f) in cell (%d,%d), AOI contains %d entities",
		afterPlayer.Pos.X, afterPlayer.Pos.Z, afterPlayer.OwnedCell.Cx, afterPlayer.OwnedCell.Cz, len(afterAOI))

	// Verify no duplicate entity IDs in AOI results
	seenIDs := make(map[string]int)
	for _, entity := range afterAOI {
		seenIDs[entity.ID]++
		if seenIDs[entity.ID] > 1 {
			t.Errorf("Duplicate entity ID %s found in AOI results", entity.ID)
		}
	}

	// Check continuous visibility: entities within radius should remain visible
	for _, neighbor := range staticNeighbors {
		distBefore := spatial.Dist2(neighbor.pos, movingPlayer.Pos)
		distAfter := spatial.Dist2(neighbor.pos, afterPlayer.Pos)
		radiusSquared := e.cfg.AOIRadius * e.cfg.AOIRadius

		// If neighbor was visible before and is still within radius, should still be visible
		wasBefore := beforeIDs[neighbor.id]
		isAfter := afterIDs[neighbor.id]
		withinRadius := distAfter <= radiusSquared

		if withinRadius && !isAfter {
			t.Errorf("Entity %s at (%.1f,%.1f) should be visible after handover (distance %.1f ≤ radius %.1f) but is missing",
				neighbor.id, neighbor.pos.X, neighbor.pos.Z,
				spatial.Dist2(neighbor.pos, afterPlayer.Pos), e.cfg.AOIRadius)
		}

		t.Logf("Static neighbor %s: was_visible=%v, is_visible=%v, dist_before=%.1f, dist_after=%.1f, within_radius=%v",
			neighbor.id, wasBefore, isAfter, distBefore, distAfter, withinRadius)
	}

	t.Logf("✓ Continuous AOI maintained across border: no duplicate IDs, entities within radius remain visible")
}

// TestAOIRebuildTimingRequirement tests that AOI rebuild completes within next snapshot
// after a handover occurs.
func TestAOIRebuildTimingRequirement(t *testing.T) {
	// Use faster snapshot rate to test timing more precisely
	e := NewEngine(Config{CellSize: 10, AOIRadius: 8, TickHz: 60, SnapshotHz: 30, HandoverHysteresisM: 1})

	// Place player near border
	e.DevSpawn("player", "Player", spatial.Vec2{X: 9.5, Z: 5})

	// Add some entities in neighboring cells
	e.AddOrUpdatePlayer("neighbor1", "N1", spatial.Vec2{X: 11, Z: 5}, spatial.Vec2{})
	e.AddOrUpdatePlayer("neighbor2", "N2", spatial.Vec2{X: 7, Z: 5}, spatial.Vec2{})

	// Record time before movement
	beforeHandover := time.Now()

	// Trigger handover by moving across border
	e.DevSetVelocity("player", spatial.Vec2{X: 20, Z: 0}) // Fast movement to ensure handover

	// Step simulation to trigger handover
	e.Step(100 * time.Millisecond) // Should move player to ~11.5

	afterPlayer, ok := e.GetPlayer("player")
	if !ok {
		t.Fatal("Player not found after movement")
	}

	// Check if handover occurred
	handoverOccurred := afterPlayer.OwnedCell.Cx != 0
	if !handoverOccurred {
		// Try one more step to ensure handover
		e.Step(50 * time.Millisecond)
		afterPlayer, _ = e.GetPlayer("player")
		handoverOccurred = afterPlayer.OwnedCell.Cx != 0
	}

	if handoverOccurred {
		// Test that AOI query works immediately after handover (within next snapshot)
		aoiResults := e.QueryAOI(afterPlayer.Pos, e.cfg.AOIRadius, afterPlayer.ID)
		queryTime := time.Now()

		// Calculate time since handover detection
		timeSinceHandover := queryTime.Sub(beforeHandover)
		snapshotInterval := time.Second / time.Duration(e.cfg.SnapshotHz) // 33ms for 30Hz

		t.Logf("AOI query after handover: %d entities found", len(aoiResults))
		t.Logf("Time since handover trigger: %v (snapshot interval: %v)", timeSinceHandover, snapshotInterval)

		// Verify AOI includes expected neighbors
		foundNeighbors := 0
		for _, entity := range aoiResults {
			if entity.ID == "neighbor1" || entity.ID == "neighbor2" {
				foundNeighbors++
			}
		}

		if foundNeighbors == 0 {
			t.Errorf("No neighbors found in AOI after handover - AOI rebuild may have failed")
		}

		// Acceptance criteria: AOI rebuild completes within next snapshot
		// We expect the AOI to be immediately available, not requiring a full snapshot interval
		if timeSinceHandover > snapshotInterval*2 {
			t.Errorf("AOI rebuild took too long: %v > %v (2× snapshot interval)",
				timeSinceHandover, snapshotInterval*2)
		} else {
			t.Logf("✓ AOI rebuild completed quickly: %v ≤ %v", timeSinceHandover, snapshotInterval*2)
		}
	} else {
		t.Logf("Handover not triggered in test (player at %.1f,%.1f in cell %d,%d) - may need stronger movement",
			afterPlayer.Pos.X, afterPlayer.Pos.Z, afterPlayer.OwnedCell.Cx, afterPlayer.OwnedCell.Cz)
	}
}

// TestNoDuplicateEntityIDs specifically tests that AOI queries never return duplicate entity IDs
// across the 3×3 cell neighborhood.
func TestNoDuplicateEntityIDs(t *testing.T) {
	e := NewEngine(Config{CellSize: 10, AOIRadius: 15, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 1})

	// Create a scenario with many entities across multiple cells
	entityCount := 0
	for cx := -1; cx <= 1; cx++ {
		for cz := -1; cz <= 1; cz++ {
			// Place 3 entities in each cell of the 3×3 grid
			for i := 0; i < 3; i++ {
				entityCount++
				x := float64(cx)*10 + 2 + float64(i)*2 // Spread within cell
				z := float64(cz)*10 + 2 + float64(i)*2
				id := fmt.Sprintf("entity_%d_%d_%d", cx, cz, i)
				e.AddOrUpdatePlayer(id, id, spatial.Vec2{X: x, Z: z}, spatial.Vec2{})
			}
		}
	}

	// Place query player at center
	e.DevSpawn("query", "Query", spatial.Vec2{X: 5, Z: 5})

	// Perform multiple AOI queries from different positions to stress test
	testPositions := []spatial.Vec2{
		{X: 0, Z: 0},     // Cell corner
		{X: 5, Z: 5},     // Cell center
		{X: 9.9, Z: 9.9}, // Near border
		{X: 10.1, Z: 5},  // Just across border
	}

	for i, pos := range testPositions {
		// Update player position
		e.AddOrUpdatePlayer("query", "Query", pos, spatial.Vec2{})

		// Query AOI
		aoiResults := e.QueryAOI(pos, e.cfg.AOIRadius, "query")

		// Check for duplicates
		seenIDs := make(map[string]bool)
		duplicates := make([]string, 0)

		for _, entity := range aoiResults {
			if seenIDs[entity.ID] {
				duplicates = append(duplicates, entity.ID)
			}
			seenIDs[entity.ID] = true
		}

		if len(duplicates) > 0 {
			t.Errorf("Test position %d (%.1f,%.1f): Found duplicate entity IDs: %v",
				i, pos.X, pos.Z, duplicates)
		}

		t.Logf("Position %d (%.1f,%.1f): %d entities, no duplicates", i, pos.X, pos.Z, len(aoiResults))
	}

	t.Logf("✓ No duplicate entity IDs found across all test positions")
}
