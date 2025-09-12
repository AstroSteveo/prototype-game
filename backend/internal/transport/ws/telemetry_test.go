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

type fakeAuthT struct{}

func (fakeAuthT) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

func TestTelemetry_TickAndRTT(t *testing.T) {
	// Use higher tick for snappier loop, but telemetry is at 1Hz
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthT{}, eng)
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

	// hello
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("hello: %v", err)
	}
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c, &raw); err != nil {
		t.Fatalf("join_ack: %v", err)
	}

	// Wait up to 2s for a telemetry message
	deadline := time.Now().Add(2 * time.Second)
	got := false
	for time.Now().Before(deadline) {
		rctx, cancelR := context.WithTimeout(context.Background(), 250*time.Millisecond)
		var msg json.RawMessage
		err := wsjson.Read(rctx, c, &msg)
		cancelR()
		if err != nil {
			continue
		}
		var env struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(msg, &env) != nil {
			continue
		}
		if env.Type != "telemetry" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(env.Data, &m); err != nil {
			continue
		}
		tr, ok1 := m["tick_rate"].(float64)
		rtt, ok2 := m["rtt_ms"].(float64)
		if !ok1 || !ok2 {
			continue
		}
		if int(tr) != 50 {
			t.Fatalf("tick_rate mismatch: got %v want 50", tr)
		}
		if rtt <= 0 || rtt > 500 {
			t.Fatalf("rtt_ms unreasonable: %vms", rtt)
		}
		got = true
		break
	}
	if !got {
		t.Fatalf("did not receive telemetry in time")
	}
}
