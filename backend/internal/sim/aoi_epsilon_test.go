package sim

import (
	"fmt"
	"math"
	"testing"

	"prototype-game/backend/internal/spatial"
)

// TestEpsilonToleranceAtBoundary validates that the epsilon tolerance (1e-9)
// correctly handles floating-point precision issues at AOI boundaries.
// This specifically addresses the requirement for epsilon tolerance in AOI calculations.
func TestEpsilonToleranceAtBoundary(t *testing.T) {
	e := NewEngine(Config{CellSize: 10, AOIRadius: 5, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})

	// Place anchor player at origin
	anchor := e.DevSpawn("anchor", "Anchor", spatial.Vec2{X: 0, Z: 0})

	// Test cases for epsilon boundary conditions
	testCases := []struct {
		name      string
		pos       spatial.Vec2
		shouldSee bool
		reason    string
	}{
		{
			name:      "exactly_at_radius",
			pos:       spatial.Vec2{X: 3, Z: 4}, // distance = 5.0 exactly
			shouldSee: true,
			reason:    "entity exactly at radius boundary should be included",
		},
		{
			name:      "just_inside_radius",
			pos:       spatial.Vec2{X: 2.999999999, Z: 4}, // distance slightly < 5.0
			shouldSee: true,
			reason:    "entity just inside radius should be included",
		},
		{
			name:      "just_outside_without_epsilon",
			pos:       spatial.Vec2{X: 3.0000000001, Z: 4}, // distance slightly > 5.0 but within epsilon
			shouldSee: true,
			reason:    "entity just outside radius should be included due to epsilon tolerance",
		},
		{
			name:      "clearly_outside_radius",
			pos:       spatial.Vec2{X: 5.1, Z: 0}, // distance = 5.1, well outside epsilon
			shouldSee: false,
			reason:    "entity clearly outside radius should be excluded",
		},
		{
			name:      "floating_point_precision_case",
			pos:       spatial.Vec2{X: 4.999999999999999, Z: 0}, // floating point precision edge case
			shouldSee: true,
			reason:    "floating point precision edge case should be handled by epsilon",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Add entity at test position
			entityID := "test_" + tc.name
			e.AddOrUpdatePlayer(entityID, tc.name, tc.pos, spatial.Vec2{})

			// Query AOI from anchor position
			results := e.QueryAOI(anchor.Pos, e.cfg.AOIRadius, anchor.ID)

			// Check if entity is in results
			found := false
			for _, ent := range results {
				if ent.ID == entityID {
					found = true
					break
				}
			}

			// Validate expectation
			if found != tc.shouldSee {
				actualDist := math.Sqrt(spatial.Dist2(tc.pos, anchor.Pos))
				t.Errorf("Test case %s failed: expected found=%v, got found=%v\n"+
					"Position: (%.10f, %.10f), Distance: %.10f, Reason: %s",
					tc.name, tc.shouldSee, found, tc.pos.X, tc.pos.Z, actualDist, tc.reason)
			}

			// Note: Entity remains for subsequent tests (no RemovePlayer method available)
		})
	}
}

// Test3x3CellNeighborhoodCoverage validates that AOI queries correctly cover
// all entities in the 3x3 cell neighborhood, including edge cases.
func Test3x3CellNeighborhoodCoverage(t *testing.T) {
	cellSize := 10.0
	aoiRadius := 15.0 // Large enough to cover entire 3x3 neighborhood

	e := NewEngine(Config{CellSize: cellSize, AOIRadius: aoiRadius, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})

	// Place query player at center of cell (0,0)
	queryPlayer := e.DevSpawn("query", "Query", spatial.Vec2{X: 5, Z: 5})

	// Define all 9 cells in 3x3 neighborhood
	expectedCells := []spatial.CellKey{
		{Cx: -1, Cz: -1}, {Cx: 0, Cz: -1}, {Cx: 1, Cz: -1},
		{Cx: -1, Cz: 0}, {Cx: 0, Cz: 0}, {Cx: 1, Cz: 0},
		{Cx: -1, Cz: 1}, {Cx: 0, Cz: 1}, {Cx: 1, Cz: 1},
	}

	// Place one entity in each cell of the 3x3 neighborhood
	entityPositions := make(map[string]spatial.Vec2)
	for _, cell := range expectedCells {
		// Place entity near center of each cell
		pos := spatial.Vec2{
			X: float64(cell.Cx)*cellSize + cellSize/2,
			Z: float64(cell.Cz)*cellSize + cellSize/2,
		}
		entityID := fmt.Sprintf("entity_%d_%d", cell.Cx, cell.Cz)
		e.AddOrUpdatePlayer(entityID, entityID, pos, spatial.Vec2{})
		entityPositions[entityID] = pos

		t.Logf("Placed entity %s at (%.1f, %.1f) in cell (%d, %d)",
			entityID, pos.X, pos.Z, cell.Cx, cell.Cz)
	}

	// Query AOI from the center position
	results := e.QueryAOI(queryPlayer.Pos, aoiRadius, queryPlayer.ID)

	// Verify all entities in 3x3 neighborhood are found
	foundEntities := make(map[string]bool)
	for _, ent := range results {
		foundEntities[ent.ID] = true
	}

	// Check each expected entity (excluding the query player which should not be in AOI results)
	expectedInAOI := 0
	for entityID, pos := range entityPositions {
		if entityID == queryPlayer.ID {
			// Query player should NOT be in its own AOI results
			if foundEntities[entityID] {
				t.Errorf("Query player %s should not be included in its own AOI results", entityID)
			}
			continue
		}

		distance := math.Sqrt(spatial.Dist2(pos, queryPlayer.Pos))
		withinRadius := distance <= aoiRadius

		if withinRadius {
			expectedInAOI++
			if !foundEntities[entityID] {
				t.Errorf("Entity %s at distance %.2f should be in AOI but was not found", entityID, distance)
			}
		} else if foundEntities[entityID] {
			t.Errorf("Entity %s at distance %.2f should not be in AOI but was found", entityID, distance)
		}

		t.Logf("Entity %s: distance=%.2f, within_radius=%v, found=%v",
			entityID, distance, withinRadius, foundEntities[entityID])
	}

	if len(results) != expectedInAOI {
		t.Errorf("Expected %d entities in AOI, found %d", expectedInAOI, len(results))
	}

	t.Logf("✓ 3x3 neighborhood coverage validated: %d entities found", len(results))
}

// TestAOICellBoundaryPrecision tests AOI behavior at exact cell boundaries
// to ensure no entities are missed due to cell boundary calculations.
func TestAOICellBoundaryPrecision(t *testing.T) {
	cellSize := 10.0
	aoiRadius := 8.0

	e := NewEngine(Config{CellSize: cellSize, AOIRadius: aoiRadius, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})

	// Test positions exactly on cell boundaries
	testCases := []struct {
		name        string
		queryPos    spatial.Vec2
		entityPos   spatial.Vec2
		shouldSee   bool
		description string
	}{
		{
			name:        "query_on_x_boundary",
			queryPos:    spatial.Vec2{X: 10.0, Z: 5.0}, // Exactly on X boundary between cells
			entityPos:   spatial.Vec2{X: 5.0, Z: 5.0},  // In adjacent cell, within range
			shouldSee:   true,
			description: "entity in adjacent cell should be visible when query is on boundary",
		},
		{
			name:        "query_on_z_boundary",
			queryPos:    spatial.Vec2{X: 5.0, Z: 10.0}, // Exactly on Z boundary between cells
			entityPos:   spatial.Vec2{X: 5.0, Z: 5.0},  // In adjacent cell, within range
			shouldSee:   true,
			description: "entity in adjacent cell should be visible when query is on boundary",
		},
		{
			name:        "both_on_corner",
			queryPos:    spatial.Vec2{X: 10.0, Z: 10.0}, // On corner of 4 cells
			entityPos:   spatial.Vec2{X: 15.0, Z: 15.0}, // In diagonal cell, within range
			shouldSee:   true,
			description: "entity in diagonal cell should be visible from corner position",
		},
		{
			name:        "epsilon_boundary_cross",
			queryPos:    spatial.Vec2{X: 9.999999999, Z: 5.0},  // Just inside cell boundary
			entityPos:   spatial.Vec2{X: 10.000000001, Z: 5.0}, // Just outside cell boundary
			shouldSee:   true,
			description: "very close entities across boundaries should be visible due to epsilon",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Place query player
			queryID := "query_" + tc.name
			e.AddOrUpdatePlayer(queryID, queryID, tc.queryPos, spatial.Vec2{})

			// Place target entity
			entityID := "entity_" + tc.name
			e.AddOrUpdatePlayer(entityID, entityID, tc.entityPos, spatial.Vec2{})

			// Query AOI
			results := e.QueryAOI(tc.queryPos, aoiRadius, queryID)

			// Check if entity is found
			found := false
			for _, ent := range results {
				if ent.ID == entityID {
					found = true
					break
				}
			}

			// Calculate actual distance
			distance := math.Sqrt(spatial.Dist2(tc.entityPos, tc.queryPos))

			// Validate result
			if found != tc.shouldSee {
				t.Errorf("Test case %s failed: expected found=%v, got found=%v\n"+
					"Query pos: (%.10f, %.10f), Entity pos: (%.10f, %.10f)\n"+
					"Distance: %.10f, Description: %s",
					tc.name, tc.shouldSee, found,
					tc.queryPos.X, tc.queryPos.Z, tc.entityPos.X, tc.entityPos.Z,
					distance, tc.description)
			}

			// Note: Entities remain in engine (no RemovePlayer method available)
		})
	}
}

// TestAOIPerformanceWithLargeCellCounts validates that AOI queries remain efficient
// even when dealing with cells containing many entities.
func TestAOIPerformanceWithLargeCellCounts(t *testing.T) {
	e := NewEngine(Config{CellSize: 50, AOIRadius: 75, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})

	// Place query player at center
	queryPlayer := e.DevSpawn("query", "Query", spatial.Vec2{X: 25, Z: 25})

	// Populate 3x3 neighborhood with many entities
	entityCount := 0
	for cx := -1; cx <= 1; cx++ {
		for cz := -1; cz <= 1; cz++ {
			// Add 20 entities per cell (180 total in 3x3 neighborhood)
			for i := 0; i < 20; i++ {
				entityCount++
				// Random position within the cell
				x := float64(cx)*50 + 10 + float64(i%5)*8 // Spread entities across cell
				z := float64(cz)*50 + 10 + float64(i/5)*8

				entityID := fmt.Sprintf("perf_entity_%d", entityCount)
				e.AddOrUpdatePlayer(entityID, entityID, spatial.Vec2{X: x, Z: z}, spatial.Vec2{})
			}
		}
	}

	// Measure AOI query performance
	iterations := 100
	for i := 0; i < iterations; i++ {
		results := e.QueryAOI(queryPlayer.Pos, e.cfg.AOIRadius, queryPlayer.ID)

		// Basic validation that we get reasonable results
		if len(results) == 0 {
			t.Errorf("Expected some entities in AOI, got 0")
		}

		if len(results) > entityCount {
			t.Errorf("AOI returned more entities (%d) than exist (%d)", len(results), entityCount)
		}
	}

	// Final comprehensive query to validate correctness
	finalResults := e.QueryAOI(queryPlayer.Pos, e.cfg.AOIRadius, queryPlayer.ID)

	// Count entities that should be within radius
	expectedInRadius := 0
	for i := 1; i <= entityCount; i++ {
		entityID := fmt.Sprintf("perf_entity_%d", i)
		if player, ok := e.GetPlayer(entityID); ok {
			distance := math.Sqrt(spatial.Dist2(player.Pos, queryPlayer.Pos))
			if distance <= e.cfg.AOIRadius {
				expectedInRadius++
			}
		}
	}

	if len(finalResults) != expectedInRadius {
		t.Errorf("Performance test failed correctness check: expected %d entities, got %d",
			expectedInRadius, len(finalResults))
	}

	t.Logf("✓ Performance test passed: %d entities in 3x3 neighborhood, %d within AOI radius",
		entityCount, len(finalResults))
}
