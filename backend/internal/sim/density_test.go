package sim

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

// testDensityEngine creates an engine with density control settings for testing.
func testDensityEngine(targetDensity, maxBots int) *Engine {
	return NewEngine(Config{
		CellSize:             10,
		AOIRadius:            5,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: targetDensity,
		MaxBots:              maxBots,
	})
}

// TestDensityControllerBasicSpawn verifies bots spawn when density is below target.
func TestDensityControllerBasicSpawn(t *testing.T) {
	e := testDensityEngine(5, 50) // target 5 per cell, max 50 bots
	e.rng = rand.New(rand.NewSource(1))

	// Create a player in cell (0,0)
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})

	// Run density maintenance once - should spawn 1 bot per cycle (ramp = ceil(5/10) = 1)
	e.mu.Lock()
	e.maintainBotDensityLocked()
	cell := e.cells[spatial.CellKey{Cx: 0, Cz: 0}]
	e.mu.Unlock()

	players, bots := countEntitiesInCell(cell)
	if players != 1 {
		t.Fatalf("expected 1 player, got %d", players)
	}

	// Should spawn bots gradually. Low bound is 80% of 5 = 4, but ramping is 1 per cycle
	// So first cycle should spawn min(3 needed, 1 ramp) = 1 bot
	if bots != 1 {
		t.Fatalf("expected 1 bot after first cycle, got %d", bots)
	}

	// Run several more cycles to reach target
	for i := 0; i < 5; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	players, bots = countEntitiesInCell(e.cells[spatial.CellKey{Cx: 0, Cz: 0}])
	e.mu.RUnlock()

	total := players + bots
	if total < 4 {
		t.Fatalf("expected at least 4 total entities (low bound) after multiple cycles, got %d", total)
	}
}

// TestDensityControllerBasicDespawn verifies bots despawn when density is above target.
func TestDensityControllerBasicDespawn(t *testing.T) {
	e := testDensityEngine(3, 50) // target 3 per cell, max 50 bots
	e.rng = rand.New(rand.NewSource(1))

	// Manually spawn many bots in cell (0,0)
	key := spatial.CellKey{Cx: 0, Cz: 0}
	e.mu.Lock()
	for i := 0; i < 8; i++ {
		e.spawnBotInCellLocked(key)
	}
	e.mu.Unlock()

	// Count before maintenance
	e.mu.RLock()
	cell := e.cells[key]
	_, botsBefore := countEntitiesInCell(cell)
	e.mu.RUnlock()

	if botsBefore != 8 {
		t.Fatalf("expected 8 bots before maintenance, got %d", botsBefore)
	}

	// Run density maintenance multiple times (ramp = ceil(3/10) = 1 per cycle)
	// High bound is 120% of 3 = 3.6, so max 4. Need to remove 4 bots (8-4).
	for i := 0; i < 5; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	_, botsAfter := countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	// Should have removed excess bots to reach high bound
	if botsAfter > 4 {
		t.Fatalf("expected at most 4 bots after maintenance (high bound), got %d", botsAfter)
	}
	if botsAfter >= botsBefore {
		t.Fatalf("expected bots to be removed, had %d before, %d after", botsBefore, botsAfter)
	}
}

// TestDensityControllerGlobalBotCap verifies that global bot cap is respected.
func TestDensityControllerGlobalBotCap(t *testing.T) {
	e := testDensityEngine(10, 5) // target 10 per cell, but only 5 bots max globally
	e.rng = rand.New(rand.NewSource(1))

	// Add players to multiple cells to trigger bot spawning
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})  // cell (0,0)
	e.AddOrUpdatePlayer("p2", "Player2", spatial.Vec2{X: 11, Z: 1}, spatial.Vec2{}) // cell (1,0)

	// Run density maintenance multiple times
	for i := 0; i < 5; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	// Count total bots globally
	e.mu.RLock()
	totalBots := len(e.bots)
	e.mu.RUnlock()

	if totalBots > 5 {
		t.Fatalf("expected at most 5 bots globally (cap), got %d", totalBots)
	}
}

// TestDensityControllerChurnScenario tests density maintenance under player churn.
func TestDensityControllerChurnScenario(t *testing.T) {
	e := testDensityEngine(4, 50) // target 4 per cell, max 50 bots
	e.rng = rand.New(rand.NewSource(1))

	key := spatial.CellKey{Cx: 0, Cz: 0}

	// Initial state: 2 players
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})
	e.AddOrUpdatePlayer("p2", "Player2", spatial.Vec2{X: 2, Z: 2}, spatial.Vec2{})

	// Run maintenance several times to reach target (ramp = ceil(4/10) = 1 per cycle)
	// Low bound is floor(80% of 4) = floor(3.2) = 3. Already have 2 players, need 1 bot.
	for i := 0; i < 3; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	players, bots := countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	if players != 2 {
		t.Fatalf("expected 2 players, got %d", players)
	}
	if players+bots < 3 {
		t.Fatalf("expected total >= 3 after spawning, got %d", players+bots)
	}

	// Simulate player leaving (remove from players map and cell)
	e.mu.Lock()
	delete(e.players, "p2")
	delete(e.cells[key].Entities, "p2")
	e.mu.Unlock()

	// Run maintenance again - should spawn more bots (need 2 more now: 3-1=2)
	botsBefore := bots
	for i := 0; i < 3; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	_, botsAfter := countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	if botsAfter <= botsBefore {
		t.Fatalf("expected more bots after player left, had %d, now %d", botsBefore, botsAfter)
	}
}

// TestDensityControllerSpawnDespawnBounds verifies spawn/despawn boundaries.
func TestDensityControllerSpawnDespawnBounds(t *testing.T) {
	e := testDensityEngine(10, 100) // target 10 per cell
	e.rng = rand.New(rand.NewSource(1))

	key := spatial.CellKey{Cx: 0, Cz: 0}

	// Test low bound: 80% of 10 = 8
	// Add 7 players (below low bound)
	for i := 0; i < 7; i++ {
		playerID := fmt.Sprintf("p%d", i+1)
		pos := spatial.Vec2{X: float64(i), Z: 1}
		e.AddOrUpdatePlayer(playerID, playerID, pos, spatial.Vec2{})
	}

	// Run maintenance multiple times (ramp = ceil(10/10) = 1 per cycle)
	// Need to spawn 1 bot to reach low bound (8-7=1)
	for i := 0; i < 2; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	players, bots := countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	if players != 7 {
		t.Fatalf("expected 7 players, got %d", players)
	}
	if players+bots < 8 {
		t.Fatalf("expected total >= 8 (low bound), got %d", players+bots)
	}

	// Test high bound: 120% of 10 = 12
	// Manually add many bots to exceed high bound
	e.mu.Lock()
	for i := 0; i < 10; i++ {
		e.spawnBotInCellLocked(key)
	}
	e.mu.Unlock()

	e.mu.RLock()
	_, botsBeforeCleanup := countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	// Run maintenance multiple times to clean up excess
	for i := 0; i < 10; i++ {
		e.mu.Lock()
		e.maintainBotDensityLocked()
		e.mu.Unlock()
	}

	e.mu.RLock()
	players, bots = countEntitiesInCell(e.cells[key])
	e.mu.RUnlock()

	if players+bots > 12 {
		t.Fatalf("expected total <= 12 (high bound), got %d", players+bots)
	}
	if bots >= botsBeforeCleanup {
		t.Fatalf("expected bot reduction, had %d, now %d", botsBeforeCleanup, bots)
	}
}

// TestDensityControllerTimingConvergence tests that density reaches target within reasonable time.
func TestDensityControllerTimingConvergence(t *testing.T) {
	e := testDensityEngine(6, 100) // target 6 per cell
	e.rng = rand.New(rand.NewSource(1))
	e.Start()
	defer func() {
		e.Stop(testContext(t))
	}()

	// Add one player to trigger density maintenance
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})

	// Wait for convergence (density maintenance runs at 1Hz)
	// Should reach low bound (floor(80% of 6) = floor(4.8) = 4) within reasonable time
	// With ramp = ceil(6/10) = 1 per second, need 3 seconds to spawn 3 bots (1 player + 3 bots = 4 total)
	converged := false
	for i := 0; i < 8; i++ { // wait up to 8 seconds with some margin
		time.Sleep(1 * time.Second)
		e.mu.RLock()
		if cell, exists := e.cells[spatial.CellKey{Cx: 0, Cz: 0}]; exists {
			players, bots := countEntitiesInCell(cell)
			total := players + bots
			if total >= 4 { // reached low bound (floor(4.8) = 4)
				converged = true
				e.mu.RUnlock()
				break
			}
		}
		e.mu.RUnlock()
	}

	if !converged {
		t.Fatalf("density did not converge to target within 8 seconds")
	}
}

// TestDensityControllerRampingRate tests that spawning/despawning is gradual.
func TestDensityControllerRampingRate(t *testing.T) {
	e := testDensityEngine(20, 200) // target 20 per cell (ramp should be 2 per cycle)
	e.rng = rand.New(rand.NewSource(1))

	key := spatial.CellKey{Cx: 0, Cz: 0}

	// Add 1 player (need to reach 16 total for low bound of 80%)
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})

	// Run maintenance once
	e.mu.Lock()
	e.maintainBotDensityLocked()
	_, botsAfter1 := countEntitiesInCell(e.cells[key])
	e.mu.Unlock()

	// Run maintenance again
	e.mu.Lock()
	e.maintainBotDensityLocked()
	_, botsAfter2 := countEntitiesInCell(e.cells[key])
	e.mu.Unlock()

	// Should have ramped up gradually (ramp = ceil(20/10) = 2 per cycle)
	rampRate := int(math.Ceil(float64(20) / 10.0))
	maxSpawnPerCycle := rampRate

	botsSpawned1 := botsAfter1
	botsSpawned2 := botsAfter2 - botsAfter1

	if botsSpawned1 > maxSpawnPerCycle {
		t.Fatalf("first cycle spawned too many bots: expected <= %d, got %d", maxSpawnPerCycle, botsSpawned1)
	}
	if botsSpawned2 > maxSpawnPerCycle {
		t.Fatalf("second cycle spawned too many bots: expected <= %d, got %d", maxSpawnPerCycle, botsSpawned2)
	}
}

// TestDensityControllerZeroTarget tests behavior when target is 0.
func TestDensityControllerZeroTarget(t *testing.T) {
	e := testDensityEngine(0, 50) // target 0 per cell (no density control)
	e.rng = rand.New(rand.NewSource(1))

	// Add player
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})

	// Run maintenance
	e.mu.Lock()
	e.maintainBotDensityLocked()
	cell := e.cells[spatial.CellKey{Cx: 0, Cz: 0}]
	e.mu.Unlock()

	_, bots := countEntitiesInCell(cell)
	if bots != 0 {
		t.Fatalf("expected no bots with target 0, got %d", bots)
	}
}

// TestDensityControllerNegativeMaxBots tests behavior when MaxBots is negative (disabled).
func TestDensityControllerNegativeMaxBots(t *testing.T) {
	e := testDensityEngine(5, -1) // target 5 per cell, no global limit (disabled)
	e.rng = rand.New(rand.NewSource(1))

	// Add player
	e.AddOrUpdatePlayer("p1", "Player1", spatial.Vec2{X: 1, Z: 1}, spatial.Vec2{})

	// Run maintenance - should NOT spawn bots when MaxBots < 0
	e.mu.Lock()
	e.maintainBotDensityLocked()
	e.mu.Unlock()

	e.mu.RLock()
	totalBots := len(e.bots)
	e.mu.RUnlock()

	// With MaxBots < 0, density control is disabled, so no bots should spawn
	if totalBots != 0 {
		t.Fatalf("expected no bots with negative MaxBots (disabled), got %d", totalBots)
	}
}

// countEntitiesInCell returns (players, bots) counts for a cell.
func countEntitiesInCell(cell *CellInstance) (int, int) {
	if cell == nil {
		return 0, 0
	}
	players, bots := 0, 0
	for _, ent := range cell.Entities {
		switch ent.Kind {
		case KindPlayer:
			players++
		case KindBot:
			bots++
		}
	}
	return players, bots
}

// testContext creates a test context with timeout.
func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
