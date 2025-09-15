package sim

import (
	"context"
	"time"

	"prototype-game/backend/internal/spatial"
)

// HandoverStatus represents the status of a cross-node handover
type HandoverStatus int

const (
	HandoverInProgress HandoverStatus = iota
	HandoverCompleted
	HandoverFailed
)

// CrossNodeHandoverState tracks the state of an ongoing cross-node handover
type CrossNodeHandoverState struct {
	TargetNode  string           `json:"target_node"`
	FromCell    spatial.CellKey  `json:"from_cell"`
	ToCell      spatial.CellKey  `json:"to_cell"`
	Status      HandoverStatus   `json:"status"`
	InitiatedAt time.Time        `json:"initiated_at"`
	Token       string           `json:"token,omitempty"`
	Error       string           `json:"error,omitempty"`
}

// HandoverToken represents a token for cross-node player handover
type HandoverToken struct {
	PlayerID   string            `json:"player_id"`
	FromNode   string            `json:"from_node"`
	ToNode     string            `json:"to_node"`
	FromCell   spatial.CellKey   `json:"from_cell"`
	ToCell     spatial.CellKey   `json:"to_cell"`
	PlayerData *PlayerData       `json:"player_data"`
	IssuedAt   time.Time         `json:"issued_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
}

// PlayerData represents serialized player state for handover
type PlayerData struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Pos        spatial.Vec2    `json:"pos"`
	Vel        spatial.Vec2    `json:"vel"`
	Yaw        float64         `json:"yaw"`
	OwnedCell  spatial.CellKey `json:"owned_cell"`
	PrevCell   spatial.CellKey `json:"prev_cell"`
	LastSeq    int             `json:"last_seq"`
}

// HandoverRequest represents a request to transfer a player to another node
type HandoverRequest struct {
	Token      string      `json:"token"`
	PlayerData *PlayerData `json:"player_data"`
}

// HandoverResponse represents the response to a handover request
type HandoverResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ResumeToken  string `json:"resume_token,omitempty"`
	TargetWSURL  string `json:"target_ws_url,omitempty"`
}

// CrossNodeHandoverService manages cross-node player transfers
type CrossNodeHandoverService interface {
	// InitiateHandover starts a cross-node handover process
	InitiateHandover(ctx context.Context, playerID string, targetNode string, targetCell spatial.CellKey) (*HandoverToken, error)
	
	// AcceptHandover accepts an incoming player from another node
	AcceptHandover(ctx context.Context, req *HandoverRequest) (*HandoverResponse, error)
	
	// ValidateHandoverToken validates a handover token
	ValidateHandoverToken(ctx context.Context, tokenStr string) (*HandoverToken, error)
}

// HandoverResult indicates the outcome of a handover attempt
type HandoverResult int

const (
	HandoverResultSuccess HandoverResult = iota
	HandoverResultFailed
	HandoverResultRetry
)

// CrossNodeHandoverEvent represents a cross-node handover event
type CrossNodeHandoverEvent struct {
	Type         string          `json:"type"` // "handover_start" or "handover_complete"
	PlayerID     string          `json:"player_id"`
	FromNode     string          `json:"from_node"`
	ToNode       string          `json:"to_node"`
	FromCell     spatial.CellKey `json:"from_cell"`
	ToCell       spatial.CellKey `json:"to_cell"`
	ResumeToken  string          `json:"resume_token,omitempty"`
	TargetWSURL  string          `json:"target_ws_url,omitempty"`
	Result       HandoverResult  `json:"result,omitempty"`
	Error        string          `json:"error,omitempty"`
}