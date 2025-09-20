package sim

import (
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

// TestDeserializePlayerData_T051 validates T-051 "Restore from persisted state"
// This test ensures the DeserializePlayerData function correctly restores player
// state from persistence using item templates
func TestDeserializePlayerData_T051(t *testing.T) {
	// Create a player manager with test item templates
	playerMgr := NewPlayerManager()
	playerMgr.CreateTestItemTemplates()
	templates := playerMgr.GetAllItemTemplates()

	// Create a test player with initial state
	player := &Player{
		Entity: Entity{
			ID:   "test-player-t051",
			Name: "TestPlayer",
			Pos:  spatial.Vec2{X: 100, Z: 200},
		},
	}

	// Initialize player with empty state
	playerMgr.InitializePlayer(player)

	// Add some items to inventory
	ironSword := ItemInstance{
		InstanceID: "sword-instance-001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}

	err := playerMgr.AddItemToInventory(player, ironSword, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add item to inventory: %v", err)
	}

	// Equip an item
	err = playerMgr.EquipItem(player, ironSword.InstanceID, SlotMainHand, time.Now())
	if err != nil {
		t.Fatalf("Failed to equip item: %v", err)
	}

	// Add some skills
	player.Skills["melee"] = 15
	player.Skills["defense"] = 8

	// Serialize the player state
	serializedState, err := SerializePlayerData(player)
	if err != nil {
		t.Fatalf("Failed to serialize player data: %v", err)
	}

	// Create a new player to restore to
	newPlayer := &Player{
		Entity: Entity{
			ID:   "test-player-t051",
			Name: "TestPlayer",
			Pos:  spatial.Vec2{X: 0, Z: 0}, // Different initial position
		},
	}

	// This is the core functionality being tested for T-051:
	// Deserialize player state using templates
	err = DeserializePlayerData(serializedState, newPlayer, templates)
	if err != nil {
		t.Fatalf("T-051 FAILED: DeserializePlayerData failed: %v", err)
	}

	// Verify position restoration
	if newPlayer.Pos.X != 100 || newPlayer.Pos.Z != 200 {
		t.Errorf("T-051 FAILED: Position not restored correctly: got %v, want (100, 200)", newPlayer.Pos)
	}

	// Verify inventory restoration
	if newPlayer.Inventory == nil {
		t.Fatal("T-051 FAILED: Inventory should be restored")
	}

	if len(newPlayer.Inventory.Items) != 1 {
		t.Errorf("T-051 FAILED: Expected 1 item in inventory, got %d", len(newPlayer.Inventory.Items))
	}

	if len(newPlayer.Inventory.Items) > 0 {
		restoredItem := newPlayer.Inventory.Items[0]
		if restoredItem.Instance.InstanceID != ironSword.InstanceID {
			t.Errorf("T-051 FAILED: Item instance ID not preserved: got %s, want %s",
				restoredItem.Instance.InstanceID, ironSword.InstanceID)
		}
		if restoredItem.Instance.TemplateID != ironSword.TemplateID {
			t.Errorf("T-051 FAILED: Item template ID not preserved: got %s, want %s",
				restoredItem.Instance.TemplateID, ironSword.TemplateID)
		}
	}

	// Verify equipment restoration
	if newPlayer.Equipment == nil {
		t.Fatal("T-051 FAILED: Equipment should be restored")
	}

	equippedItem := newPlayer.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil {
		t.Fatal("T-051 FAILED: Main hand item should be restored")
	}

	if equippedItem.Instance.InstanceID != ironSword.InstanceID {
		t.Errorf("T-051 FAILED: Equipped item not restored correctly: got %s, want %s",
			equippedItem.Instance.InstanceID, ironSword.InstanceID)
	}

	// Verify skills restoration
	if newPlayer.Skills == nil {
		t.Fatal("T-051 FAILED: Skills should be restored")
	}

	if newPlayer.Skills["melee"] != 15 {
		t.Errorf("T-051 FAILED: Melee skill not restored correctly: got %d, want 15",
			newPlayer.Skills["melee"])
	}

	if newPlayer.Skills["defense"] != 8 {
		t.Errorf("T-051 FAILED: Defense skill not restored correctly: got %d, want 8",
			newPlayer.Skills["defense"])
	}

	t.Logf("T-051 SUCCESS: DeserializePlayerData correctly restored all player state")
}

// TestEngineRestorePlayerState_T051 validates the engine-level restore functionality
func TestEngineRestorePlayerState_T051(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	playerMgr := eng.GetPlayerManager()
	playerMgr.CreateTestItemTemplates()
	templates := playerMgr.GetAllItemTemplates()

	playerID := "test-engine-restore-t051"
	playerName := "EngineTestPlayer"
	initialPos := spatial.Vec2{X: 150, Z: 250}

	// Add player to engine
	player := eng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
	playerMgr.InitializePlayer(player)

	// Set up some state to restore
	player.Skills = map[string]int{
		"magic":   20,
		"archery": 12,
	}

	// Create persisted state
	persistedState := state.PlayerState{
		Pos:     spatial.Vec2{X: 300, Z: 400},
		Logins:  5,
		Updated: time.Now(),
		Version: 1,
	}

	// Add some test data to the persisted state
	skillsData := `{"combat": 25, "crafting": 10}`
	persistedState.SkillsData = []byte(skillsData)

	// Test the engine-level restore function (core of T-051)
	err := eng.RestorePlayerState(playerID, persistedState, templates)
	if err != nil {
		t.Fatalf("T-051 FAILED: Engine RestorePlayerState failed: %v", err)
	}

	// Verify the player state was restored
	restoredPlayer, exists := eng.GetPlayer(playerID)
	if !exists {
		t.Fatal("T-051 FAILED: Player should exist after restore")
	}

	// Check position restoration
	if restoredPlayer.Pos.X != 300 || restoredPlayer.Pos.Z != 400 {
		t.Errorf("T-051 FAILED: Position not restored by engine: got %v, want (300, 400)", restoredPlayer.Pos)
	}

	// Check skills restoration
	if restoredPlayer.Skills["combat"] != 25 {
		t.Errorf("T-051 FAILED: Combat skill not restored: got %d, want 25", restoredPlayer.Skills["combat"])
	}

	if restoredPlayer.Skills["crafting"] != 10 {
		t.Errorf("T-051 FAILED: Crafting skill not restored: got %d, want 10", restoredPlayer.Skills["crafting"])
	}

	t.Logf("T-051 SUCCESS: Engine RestorePlayerState correctly restored player state")
}
