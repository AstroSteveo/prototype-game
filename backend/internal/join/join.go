package join

import (
	"context"
	"log"
	"time"

	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

// AuthService validates a client token and returns player identity.
type AuthService interface {
	Validate(ctx context.Context, token string) (playerID, name string, ok bool)
}

// Hello represents the minimal client hello payload.
type Hello struct {
	Token   string `json:"token"`
	Resume  string `json:"resume,omitempty"`
	LastSeq int    `json:"last_seq,omitempty"`
}

// JoinAck is sent on successful join.
type JoinAck struct {
	PlayerID string          `json:"player_id"`
	Pos      spatial.Vec2    `json:"pos"`
	Cell     spatial.CellKey `json:"cell"`
	Config   struct {
		TickHz              int     `json:"tick_hz"`
		SnapshotHz          int     `json:"snapshot_hz"`
		AOIRadius           float64 `json:"aoi_radius"`
		CellSize            float64 `json:"cell_size"`
		HandoverHysteresisM float64 `json:"handover_hysteresis"`
	} `json:"config"`
	Inventory   *sim.Inventory       `json:"inventory"`
	Equipment   *sim.Equipment       `json:"equipment"`
	Skills      map[string]int       `json:"skills"`
	Encumbrance sim.EncumbranceState `json:"encumbrance"`
	ResumeToken string               `json:"resume,omitempty"`
}

// ErrorMsg is a structured error for transport.
type ErrorMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HandleJoin performs auth, spawns/attaches the player, and builds a JoinAck.
// It is transport-agnostic so we can test without websockets.
func HandleJoin(ctx context.Context, auth AuthService, eng *sim.Engine, hello Hello) (JoinAck, *ErrorMsg) {
	if hello.Token == "" {
		return JoinAck{}, &ErrorMsg{Code: "bad_request", Message: "missing token"}
	}
	pid, name, ok := auth.Validate(ctx, hello.Token)
	if !ok || pid == "" {
		return JoinAck{}, &ErrorMsg{Code: "auth", Message: "invalid token"}
	}

	playerMgr := eng.GetPlayerManager()
	templates := playerMgr.GetAllItemTemplates()

	// Load last known state if available with full persistence data (US-006)
	pos := spatial.Vec2{}
	var persistedState *state.PlayerState

	if playerStore != nil {
		if st, ok, err := playerStore.Load(ctx, pid); ok && err == nil {
			pos = st.Pos
			persistedState = &st
		}
	}

	eng.AddOrUpdatePlayer(pid, name, pos, spatial.Vec2{})

	// Initialize or restore player data directly on the engine's authoritative record
	if persistedState != nil {
		// Restore from persistence including inventory, equipment, and skills
		if err := eng.RestorePlayerState(pid, *persistedState, templates); err != nil {
			// If deserialization fails, log warning and fall back to defaults
			log.Printf("join: failed to deserialize player state for %s: %v", pid, err)
			log.Printf("join: corrupted persisted state for %s: %+v", pid, *persistedState)
			// Get the player reference and initialize with defaults
			if snap, ok := eng.GetPlayer(pid); ok {
				playerMgr.InitializePlayer(&snap)
				// Attempt to preserve non-corrupted fields from persistedState
				restored := state.PlayerState{
					Pos:       snap.Pos,
					Inventory: persistedState.Inventory,
					Equipment: persistedState.Equipment,
					Skills:    persistedState.Skills,
				}
				eng.RestorePlayerState(pid, restored, templates)
			}
		}
	}

	// Read player fields via snapshot accessor after restoration for the response
	snap, _ := eng.GetPlayer(pid)

	// Ensure all player components are properly initialized if this is a new player
	if persistedState == nil && (snap.Inventory == nil || snap.Equipment == nil || snap.Skills == nil) {
		playerMgr.InitializePlayer(&snap)
		// Note: For new players, AddOrUpdatePlayer creates the record, but full initialization of components is performed by InitializePlayer
	}

	cfg := eng.GetConfig()
	ack := JoinAck{
		PlayerID: snap.ID,
		Pos:      snap.Pos,
		Cell:     snap.OwnedCell,
	}
	ack.Config.TickHz = cfg.TickHz
	ack.Config.SnapshotHz = cfg.SnapshotHz
	ack.Config.AOIRadius = cfg.AOIRadius
	ack.Config.CellSize = cfg.CellSize
	ack.Config.HandoverHysteresisM = cfg.HandoverHysteresisM

	// Include inventory and equipment data in join response
	ack.Inventory = snap.Inventory
	ack.Equipment = snap.Equipment
	ack.Skills = snap.Skills

	// Calculate current encumbrance
	ack.Encumbrance = playerMgr.GetPlayerEncumbrance(&snap)

	// Persist updated state immediately (best-effort) for login tracking
	if playerStore != nil {
		if persistedState != nil {
			// Update existing player record
			persistedState.Logins++
			persistedState.Pos = snap.Pos
			persistedState.Updated = time.Now()
			_ = playerStore.Save(ctx, pid, *persistedState)
		} else {
			// Create new player record with default state
			newState := sim.CreateDefaultPlayerState(pid, snap.Pos)
			_ = playerStore.Save(ctx, pid, newState)
		}
	}

	return ack, nil
}

// Pluggable store for player persistence; set by the service (e.g., sim main).
var playerStore state.Store

// SetStore configures the package-level persistence store.
func SetStore(s state.Store) { playerStore = s }

// now is an indirection for tests.
// no-op
