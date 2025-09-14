package join

import (
	"context"

	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/spatial"
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
	ResumeToken string `json:"resume,omitempty"`
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
	// TODO (M5): load last known position; for now default to origin.
	pos := spatial.Vec2{}
	eng.AddOrUpdatePlayer(pid, name, pos, spatial.Vec2{})
	// Read player fields via snapshot accessor to avoid data races with the tick loop.
	snap, _ := eng.GetPlayer(pid)

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
	return ack, nil
}
