package sim

import (
	"encoding/json"
	"fmt"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

// PlayerPersistenceData represents the complete serializable player state
type PlayerPersistenceData struct {
	Inventory         *Inventory           `json:"inventory"`
	Equipment         *Equipment           `json:"equipment"`
	Skills            map[string]int       `json:"skills"`
	CooldownTimers    map[SlotID]time.Time `json:"cooldown_timers"`
	EncumbranceConfig struct {
		WeightLimit     float64                 `json:"weight_limit"`
		CompartmentCaps map[CompartmentType]int `json:"compartment_caps"`
	} `json:"encumbrance_config"`
}

// SerializePlayerData converts player game state to persistent state
func SerializePlayerData(player *Player) (state.PlayerState, error) {
	persistData := PlayerPersistenceData{
		Inventory:      player.Inventory,
		Equipment:      player.Equipment,
		Skills:         player.Skills,
		CooldownTimers: make(map[SlotID]time.Time),
	}

	// Extract cooldown timers from equipment
	if player.Equipment != nil {
		for slotID, equippedItem := range player.Equipment.Slots {
			if equippedItem != nil {
				persistData.CooldownTimers[slotID] = equippedItem.CooldownUntil
			}
		}
	}

	// Capture encumbrance configuration
	if player.Inventory != nil {
		persistData.EncumbranceConfig.WeightLimit = player.Inventory.WeightLimit
		persistData.EncumbranceConfig.CompartmentCaps = player.Inventory.CompartmentCaps
	}

	// Serialize individual components
	inventoryData, err := json.Marshal(persistData.Inventory)
	if err != nil {
		return state.PlayerState{}, fmt.Errorf("failed to serialize inventory: %w", err)
	}

	equipmentData, err := json.Marshal(persistData.Equipment)
	if err != nil {
		return state.PlayerState{}, fmt.Errorf("failed to serialize equipment: %w", err)
	}

	skillsData, err := json.Marshal(persistData.Skills)
	if err != nil {
		return state.PlayerState{}, fmt.Errorf("failed to serialize skills: %w", err)
	}

	cooldownData, err := json.Marshal(persistData.CooldownTimers)
	if err != nil {
		return state.PlayerState{}, fmt.Errorf("failed to serialize cooldowns: %w", err)
	}

	encumbranceData, err := json.Marshal(persistData.EncumbranceConfig)
	if err != nil {
		return state.PlayerState{}, fmt.Errorf("failed to serialize encumbrance config: %w", err)
	}

	return state.PlayerState{
		Pos:               player.Pos,
		Logins:            0, // Will be set by caller
		Updated:           time.Now(),
		Version:           0, // Will be managed by database
		InventoryData:     inventoryData,
		EquipmentData:     equipmentData,
		SkillsData:        skillsData,
		CooldownTimers:    cooldownData,
		EncumbranceConfig: encumbranceData,
	}, nil
}

// DeserializePlayerData converts persistent state back to player game state
func DeserializePlayerData(state state.PlayerState, player *Player, templates map[ItemTemplateID]*ItemTemplate) error {
	// Restore position
	player.Pos = state.Pos

	// Deserialize inventory
	if len(state.InventoryData) > 0 {
		var inventory Inventory
		if err := json.Unmarshal(state.InventoryData, &inventory); err != nil {
			return fmt.Errorf("failed to deserialize inventory: %w", err)
		}

		// Restore template catalog and rebuild internal index map
		inventory.SetTemplateCatalog(templates)
		inventory.rebuildIndex()
		player.Inventory = &inventory
	}

	// Deserialize equipment
	if len(state.EquipmentData) > 0 {
		var equipment Equipment
		if err := json.Unmarshal(state.EquipmentData, &equipment); err != nil {
			return fmt.Errorf("failed to deserialize equipment: %w", err)
		}
		player.Equipment = &equipment
	}

	// Deserialize skills
	if len(state.SkillsData) > 0 {
		var skills map[string]int
		if err := json.Unmarshal(state.SkillsData, &skills); err != nil {
			return fmt.Errorf("failed to deserialize skills: %w", err)
		}
		player.Skills = skills
	}

	// Restore cooldown timers to equipment
	if len(state.CooldownTimers) > 0 && player.Equipment != nil {
		var cooldowns map[SlotID]time.Time
		if err := json.Unmarshal(state.CooldownTimers, &cooldowns); err != nil {
			return fmt.Errorf("failed to deserialize cooldowns: %w", err)
		}

		// Apply cooldowns to equipped items
		for slotID, cooldownTime := range cooldowns {
			if equippedItem := player.Equipment.GetSlot(slotID); equippedItem != nil {
				equippedItem.CooldownUntil = cooldownTime
			}
		}
	}

	// Restore encumbrance configuration
	if len(state.EncumbranceConfig) > 0 && player.Inventory != nil {
		var encumbranceConfig struct {
			WeightLimit     float64                 `json:"weight_limit"`
			CompartmentCaps map[CompartmentType]int `json:"compartment_caps"`
		}
		if err := json.Unmarshal(state.EncumbranceConfig, &encumbranceConfig); err != nil {
			return fmt.Errorf("failed to deserialize encumbrance config: %w", err)
		}

		player.Inventory.WeightLimit = encumbranceConfig.WeightLimit
		if encumbranceConfig.CompartmentCaps != nil {
			player.Inventory.CompartmentCaps = encumbranceConfig.CompartmentCaps
		}
	}

	return nil
}

// CreateDefaultPlayerState creates a default player state for new players
func CreateDefaultPlayerState(playerID string, pos spatial.Vec2) state.PlayerState {
	// Create default inventory and equipment
	inventory := NewInventory()
	equipment := NewEquipment()
	skills := make(map[string]int)

	// Serialize defaults - these should never fail, but handle errors for robustness
	inventoryData, err := json.Marshal(inventory)
	if err != nil {
		// This should never happen with default inventory, but panic if it does
		panic(fmt.Sprintf("failed to marshal default inventory during player state creation: %v", err))
	}

	equipmentData, err := json.Marshal(equipment)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default equipment: %v", err))
	}

	skillsData, err := json.Marshal(skills)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default skills: %v", err))
	}

	cooldownData, err := json.Marshal(make(map[SlotID]time.Time))
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default cooldowns: %v", err))
	}

	encumbranceConfig := struct {
		WeightLimit     float64                 `json:"weight_limit"`
		CompartmentCaps map[CompartmentType]int `json:"compartment_caps"`
	}{
		WeightLimit:     inventory.WeightLimit,
		CompartmentCaps: inventory.CompartmentCaps,
	}
	encumbranceData, err := json.Marshal(encumbranceConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal default encumbrance config: %v", err))
	}

	return state.PlayerState{
		Pos:               pos,
		Logins:            1,
		Updated:           time.Now(),
		Version:           1,
		InventoryData:     inventoryData,
		EquipmentData:     equipmentData,
		SkillsData:        skillsData,
		CooldownTimers:    cooldownData,
		EncumbranceConfig: encumbranceData,
	}
}
