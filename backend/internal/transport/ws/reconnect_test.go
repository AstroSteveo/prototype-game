//go:build ws

package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/sim"
)

// fakeAuth implements the join.AuthService interface without importing join in tests.
type fakeAuthR struct{}

func (fakeAuthR) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

func TestWS_ReconnectAndResume(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 60, SnapshotHz: 30, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthR{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First connection
	c1, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial1: %v", err)
	}
	// hello
	if err := wsjson.Write(ctx, c1, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("hello1: %v", err)
	}
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c1, &raw); err != nil {
		t.Fatalf("join_ack1: %v", err)
	}
	var env struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal env1: %v", err)
	}
	if env.Type != "join_ack" {
		t.Fatalf("expected join_ack, got %q", env.Type)
	}
	resume, _ := env.Data["resume"].(string)
	if resume == "" {
		t.Fatalf("expected resume token in join_ack data")
	}

	// Send one input to advance seq/pos
	_ = wsjson.Write(ctx, c1, map[string]any{"type": "input", "seq": 1, "dt": 0.05, "intent": map[string]float64{"x": 1, "z": 0}})

	// Close first connection
	_ = c1.Close(nws.StatusNormalClosure, "bye")

	// Reconnect with resume and last_seq
	c2, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial2: %v", err)
	}
	defer c2.Close(nws.StatusNormalClosure, "bye")
	if err := wsjson.Write(ctx, c2, map[string]any{"token": "tok", "resume": resume, "last_seq": 1}); err != nil {
		t.Fatalf("hello2: %v", err)
	}
	var raw2 json.RawMessage
	if err := wsjson.Read(ctx, c2, &raw2); err != nil {
		t.Fatalf("join_ack2: %v", err)
	}

	// Expect next state with ack >= 1 soon
	deadline := time.Now().Add(2 * time.Second)
	sawAck := false
	for time.Now().Before(deadline) {
		var msg json.RawMessage
		rctx, cancelR := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := wsjson.Read(rctx, c2, &msg)
		cancelR()
		if err != nil {
			continue
		}
		var e struct {
			Type string         `json:"type"`
			Data map[string]any `json:"data"`
		}
		if json.Unmarshal(msg, &e) != nil {
			continue
		}
		if e.Type != "state" {
			continue
		}
		if v, ok := e.Data["ack"].(float64); ok && int(v) >= 1 {
			sawAck = true
			break
		}
	}
	if !sawAck {
		t.Fatalf("did not observe ack >= 1 after resume")
	}
}

func TestResumeManager_Validate(t *testing.T) {
	rm := NewResumeManager(time.Second)

	// Test validation with empty inputs
	if rm.Validate("", "") {
		t.Error("expected false for empty token and playerID")
	}
	if rm.Validate("token", "") {
		t.Error("expected false for empty playerID")
	}
	if rm.Validate("", "player1") {
		t.Error("expected false for empty token")
	}

	// Issue a token for player1
	token := rm.Issue("player1")
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Test successful validation
	if !rm.Validate(token, "player1") {
		t.Error("expected true for valid token and matching playerID")
	}

	// Test validation with wrong player ID
	if rm.Validate(token, "player2") {
		t.Error("expected false for valid token but wrong playerID")
	}

	// Test validation with invalid token
	if rm.Validate("invalid", "player1") {
		t.Error("expected false for invalid token")
	}

	// Test validation after expiration
	rmShort := NewResumeManager(1 * time.Millisecond)
	tokenShort := rmShort.Issue("player1")
	time.Sleep(2 * time.Millisecond)
	if rmShort.Validate(tokenShort, "player1") {
		t.Error("expected false for expired token")
	}
}
