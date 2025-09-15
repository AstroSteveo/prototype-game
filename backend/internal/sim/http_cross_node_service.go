package sim

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"prototype-game/backend/internal/spatial"
)

// HTTPCrossNodeService implements cross-node handover over HTTP
type HTTPCrossNodeService struct {
	mu          sync.RWMutex
	nodeID      string
	tokens      map[string]*HandoverToken // token -> HandoverToken
	httpClient  *http.Client
	localPort   int
}

// NewHTTPCrossNodeService creates a new HTTP-based cross-node handover service
func NewHTTPCrossNodeService(nodeID string, localPort int) *HTTPCrossNodeService {
	return &HTTPCrossNodeService{
		nodeID:     nodeID,
		tokens:     make(map[string]*HandoverToken),
		httpClient: &http.Client{Timeout: 5 * time.Second},
		localPort:  localPort,
	}
}

// InitiateHandover starts a cross-node handover process
func (s *HTTPCrossNodeService) InitiateHandover(ctx context.Context, playerID string, targetNode string, targetCell spatial.CellKey) (*HandoverToken, error) {
	// Generate handover token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate handover token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)
	
	token := &HandoverToken{
		PlayerID:  playerID,
		FromNode:  s.nodeID,
		ToNode:    targetNode,
		ToCell:    targetCell,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Second), // 30 second expiry
	}
	
	// Store token locally
	s.mu.Lock()
	s.tokens[tokenStr] = token
	s.mu.Unlock()
	
	return token, nil
}

// AcceptHandover accepts an incoming player from another node
func (s *HTTPCrossNodeService) AcceptHandover(ctx context.Context, req *HandoverRequest) (*HandoverResponse, error) {
	// Validate the handover token
	token, err := s.ValidateHandoverToken(ctx, req.Token)
	if err != nil {
		return &HandoverResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid handover token: %v", err),
		}, nil
	}
	
	// Verify token is for this node
	if token.ToNode != s.nodeID {
		return &HandoverResponse{
			Success: false,
			Error:   "handover token not intended for this node",
		}, nil
	}
	
	// For MVP: just accept the handover and generate a resume token
	// In a full implementation, this would coordinate with the Engine
	resumeToken := s.generateResumeToken()
	targetWSURL := fmt.Sprintf("ws://localhost:%d/ws", s.localPort)
	
	return &HandoverResponse{
		Success:     true,
		ResumeToken: resumeToken,
		TargetWSURL: targetWSURL,
	}, nil
}

// ValidateHandoverToken validates a handover token
func (s *HTTPCrossNodeService) ValidateHandoverToken(ctx context.Context, tokenStr string) (*HandoverToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	token, exists := s.tokens[tokenStr]
	if !exists {
		return nil, fmt.Errorf("handover token not found")
	}
	
	if time.Now().After(token.ExpiresAt) {
		delete(s.tokens, tokenStr)
		return nil, fmt.Errorf("handover token expired")
	}
	
	return token, nil
}

// RequestHandover sends a handover request to a target node
func (s *HTTPCrossNodeService) RequestHandover(ctx context.Context, targetNodeInfo *NodeInfo, playerData *PlayerData, token string) (*HandoverResponse, error) {
	req := &HandoverRequest{
		Token:      token,
		PlayerData: playerData,
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handover request: %w", err)
	}
	
	url := fmt.Sprintf("http://%s:%d/handover", targetNodeInfo.Address, targetNodeInfo.Port)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("handover request failed with status %d", resp.StatusCode)
	}
	
	var handoverResp HandoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&handoverResp); err != nil {
		return nil, fmt.Errorf("failed to decode handover response: %w", err)
	}
	
	return &handoverResp, nil
}

// generateResumeToken generates a resume token for reconnection
func (s *HTTPCrossNodeService) generateResumeToken() string {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "" // Return empty token on error
	}
	return hex.EncodeToString(tokenBytes)
}

// CleanupExpiredTokens removes expired handover tokens
func (s *HTTPCrossNodeService) CleanupExpiredTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for tokenStr, token := range s.tokens {
		if now.After(token.ExpiresAt) {
			delete(s.tokens, tokenStr)
		}
	}
}

// RegisterHandoverEndpoint registers the HTTP endpoint for accepting handovers
func (s *HTTPCrossNodeService) RegisterHandoverEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/handover", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		var req HandoverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		resp, err := s.AcceptHandover(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}