package sim

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func TestHTTPCrossNodeService_InitiateHandover(t *testing.T) {
	service := NewHTTPCrossNodeService("node1", 8081)
	
	ctx := context.Background()
	playerID := "player123"
	targetNode := "node2"
	targetCell := spatial.CellKey{Cx: 1, Cz: 1}
	
	token, err := service.InitiateHandover(ctx, playerID, targetNode, targetCell)
	if err != nil {
		t.Fatalf("InitiateHandover failed: %v", err)
	}
	
	if token == nil {
		t.Fatal("Expected token, got nil")
	}
	
	if token.PlayerID != playerID {
		t.Errorf("Expected player ID %s, got %s", playerID, token.PlayerID)
	}
	
	if token.FromNode != "node1" {
		t.Errorf("Expected from node 'node1', got %s", token.FromNode)
	}
	
	if token.ToNode != targetNode {
		t.Errorf("Expected to node %s, got %s", targetNode, token.ToNode)
	}
	
	if token.ToCell != targetCell {
		t.Errorf("Expected to cell %v, got %v", targetCell, token.ToCell)
	}
	
	// Verify token is stored
	if len(service.tokens) != 1 {
		t.Errorf("Expected 1 token stored, got %d", len(service.tokens))
	}
}

func TestHTTPCrossNodeService_ValidateHandoverToken(t *testing.T) {
	service := NewHTTPCrossNodeService("node1", 8081)
	ctx := context.Background()
	
	// Create a valid token
	token, err := service.InitiateHandover(ctx, "player123", "node2", spatial.CellKey{Cx: 1, Cz: 1})
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	
	// Find the token string
	var tokenStr string
	for k, v := range service.tokens {
		if v == token {
			tokenStr = k
			break
		}
	}
	
	if tokenStr == "" {
		t.Fatal("Token string not found")
	}
	
	// Validate the token
	validatedToken, err := service.ValidateHandoverToken(ctx, tokenStr)
	if err != nil {
		t.Errorf("ValidateHandoverToken failed: %v", err)
	}
	
	if validatedToken != token {
		t.Error("Validated token doesn't match original")
	}
	
	// Test with invalid token
	_, err = service.ValidateHandoverToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestHTTPCrossNodeService_AcceptHandover(t *testing.T) {
	service := NewHTTPCrossNodeService("node2", 8082)
	ctx := context.Background()
	
	// Create a handover token (simulate receiving from node1)
	token, err := service.InitiateHandover(ctx, "player123", "node2", spatial.CellKey{Cx: 1, Cz: 1})
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	
	// Find the token string
	var tokenStr string
	for k, v := range service.tokens {
		if v == token {
			tokenStr = k
			break
		}
	}
	
	playerData := &PlayerData{
		ID:   "player123",
		Name: "TestPlayer",
		Pos:  spatial.Vec2{X: 100, Z: 100},
		Vel:  spatial.Vec2{X: 0, Z: 0},
		Yaw:  0,
	}
	
	req := &HandoverRequest{
		Token:      tokenStr,
		PlayerData: playerData,
	}
	
	resp, err := service.AcceptHandover(ctx, req)
	if err != nil {
		t.Fatalf("AcceptHandover failed: %v", err)
	}
	
	if !resp.Success {
		t.Errorf("Expected success, got failure: %s", resp.Error)
	}
	
	if resp.ResumeToken == "" {
		t.Error("Expected resume token")
	}
	
	if resp.TargetWSURL == "" {
		t.Error("Expected target WebSocket URL")
	}
}

func TestHTTPCrossNodeService_CleanupExpiredTokens(t *testing.T) {
	service := NewHTTPCrossNodeService("node1", 8081)
	ctx := context.Background()
	
	// Create a token
	_, err := service.InitiateHandover(ctx, "player123", "node2", spatial.CellKey{Cx: 1, Cz: 1})
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	
	// Verify token exists
	if len(service.tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(service.tokens))
	}
	
	// Manually expire the token
	for _, token := range service.tokens {
		token.ExpiresAt = time.Now().Add(-1 * time.Hour)
	}
	
	// Cleanup expired tokens
	service.CleanupExpiredTokens()
	
	// Verify token is removed
	if len(service.tokens) != 0 {
		t.Errorf("Expected 0 tokens after cleanup, got %d", len(service.tokens))
	}
}