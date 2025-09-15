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
type fakeAuth struct{}

func (fakeAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

func TestWS_InputState_AckAndMotion(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuth{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Send hello
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c, &raw); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}
	var env struct {
		Type  string          `json:"type"`
		Data  json.RawMessage `json:"data"`
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Type != "join_ack" {
		t.Fatalf("expected join_ack, got %q", env.Type)
	}

	// Send one input and expect a state with ack >= 1 and pos.X > 0
	if err := wsjson.Write(ctx, c, map[string]any{
		"type":   "input",
		"seq":    1,
		"dt":     0.05,
		"intent": map[string]float64{"x": 1, "z": 0},
	}); err != nil {
		t.Fatalf("write input: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	gotAck, gotMotion := false, false
	for time.Now().Before(deadline) {
		var msg json.RawMessage
		rctx, cancelR := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := wsjson.Read(rctx, c, &msg)
		cancelR()
		if err != nil {
			continue
		}
		var e struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(msg, &e) != nil {
			continue
		}
		if e.Type != "state" {
			continue
		}
		// State payload is {"ack":N, "player": {"id":..., "pos": {"X":...}}}
		// Use a relaxed decoder into a map first to extract fields.
		var m map[string]any
		if err := json.Unmarshal(e.Data, &m); err != nil {
			continue
		}
		if v, ok := m["ack"].(float64); ok && int(v) >= 1 {
			gotAck = true
		}
		if pl, ok := m["player"].(map[string]any); ok {
			if pos, ok := pl["pos"].(map[string]any); ok {
				if x, ok := pos["X"].(float64); ok && x > 0 {
					gotMotion = true
				}
			}
		}
		if gotAck && gotMotion {
			break
		}
	}
	if !gotAck {
		t.Fatalf("did not observe ack >= 1 in time")
	}
	if !gotMotion {
		t.Fatalf("did not observe positive X motion in time")
	}
}

// slowAuth simulates a slow auth service to test timeout behavior
type slowAuth struct {
	delay time.Duration
}

func (s slowAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "slow" {
		select {
		case <-time.After(s.delay):
			return "p1", "Alice", true
		case <-ctx.Done():
			// Context was cancelled/timed out
			return "", "", false
		}
	}
	return "", "", false
}

func TestWS_JoinTimeout_HandlesAuthTimeout(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	// Create an auth service that takes 10 seconds to respond
	auth := slowAuth{delay: 10 * time.Second}

	mux := http.NewServeMux()
	Register(mux, "/ws", auth, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Send hello with slow token
	start := time.Now()
	if err := wsjson.Write(ctx, c, map[string]any{"token": "slow"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	// Try to read response, expecting either an error message or connection close
	var raw json.RawMessage
	readCtx, readCancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer readCancel()

	err = wsjson.Read(readCtx, c, &raw)
	elapsed := time.Since(start)

	// Should complete within 6 seconds (the 5s timeout + buffer), not the full 10s delay
	if elapsed > 6*time.Second {
		t.Fatalf("WebSocket join took too long (%v), timeout not working properly", elapsed)
	}

	if err != nil {
		// Connection closed, which is acceptable for timeout scenarios
		t.Logf("✓ Auth timeout handled correctly - connection closed in %v (< 6s)", elapsed)
		return
	}

	// If we got a response, it should be an auth error
	var env struct {
		Type  string `json:"type"`
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	// Should get an error response (auth timeout)
	if env.Type != "error" || env.Error == nil || env.Error.Code != "auth" {
		t.Fatalf("expected auth error, got: %s", string(raw))
	}

	t.Logf("✓ Auth timeout handled correctly with error message in %v (< 6s)", elapsed)
}
