package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

type session struct {
	PlayerID string    `json:"player_id"`
	Name     string    `json:"name"`
	Token    string    `json:"token"`
	LastSeen time.Time `json:"last_seen"`
}

type gateway struct {
	mu         sync.Mutex
	sessions   map[string]session // token -> session
	simAddress string
}

func newGateway(simAddr string) *gateway {
	return &gateway{sessions: make(map[string]session), simAddress: simAddr}
}

func (g *gateway) handleLogin(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Player"
	}
	tok := randomToken()
	s := session{
		PlayerID: randomToken()[:8],
		Name:     name,
		Token:    tok,
		LastSeen: time.Now(),
	}
	g.mu.Lock()
	g.sessions[tok] = s
	g.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":     tok,
		"player_id": s.PlayerID,
		"sim": map[string]any{
			"address":  g.simAddress,
			"protocol": "http-json-dev", // placeholder until WS is added
		},
	})
}

func (g *gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// handleValidate returns player identity for a valid token.
// GET /validate?token=...
func (g *gateway) handleValidate(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	g.mu.Lock()
	s, ok := g.sessions[tok]
	g.mu.Unlock()
	if !ok {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"player_id": s.PlayerID,
		"name":      s.Name,
	})
}

func randomToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// fallback time-based
		now := time.Now().UnixNano()
		return hex.EncodeToString([]byte{byte(now), byte(now >> 8), byte(now >> 16), byte(now >> 24)})
	}
	return hex.EncodeToString(b[:])
}

func main() {
	var port = flag.String("port", "8080", "gateway port")
	var simAddr = flag.String("sim", "localhost:8081", "sim service address")
	flag.Parse()

	g := newGateway(*simAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", g.handleHealth)
	mux.HandleFunc("/login", g.handleLogin)
	mux.HandleFunc("/validate", g.handleValidate)

	log.Printf("gateway: listening on :%s (sim=%s)", *port, *simAddr)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}
