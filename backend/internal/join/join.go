package join

import (
	"context"
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

	// Read player fields via snapshot accessor to avoid data races with the tick loop.
	snap, _ := eng.GetPlayer(pid)

	// Initialize or restore player data
	if persistedState != nil {
		// Restore from persistence including inventory, equipment, and skills
		if err := sim.DeserializePlayerData(*persistedState, &snap, templates); err != nil {
			// If deserialization fails, log warning and fall back to defaults
			// In production, we might want more sophisticated error handling
			playerMgr.InitializePlayer(&snap)
		}
	} else {
		// New player - initialize with defaults
		playerMgr.InitializePlayer(&snap)
	}

	// Ensure all player components are properly initialized
	if snap.Inventory == nil || snap.Equipment == nil || snap.Skills == nil {
		playerMgr.InitializePlayer(&snap)
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
