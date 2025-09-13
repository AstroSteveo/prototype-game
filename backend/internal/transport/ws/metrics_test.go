//go:build ws

package ws

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/metrics"
	"prototype-game/backend/internal/sim"
)

type fakeAuthM struct{}

func (fakeAuthM) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

// Ensures Prometheus metrics endpoint exposes ws_connected and reflects an open connection.
func TestMetrics_WsConnectedGauge(t *testing.T) {
	metrics.Init() // fresh registry for test

	eng := sim.NewEngine(sim.Config{CellSize: 4, AOIRadius: 2, TickHz: 60, SnapshotHz: 20, HandoverHysteresisM: 0.25})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthM{}, eng)
	mux.Handle("/metrics", metrics.Handler())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Connect a WS client and complete join
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("hello: %v", err)
	}
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c, &raw); err != nil {
		t.Fatalf("join_ack: %v", err)
	}

	// Scrape metrics
	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("metrics status: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	if !strings.Contains(s, "ws_connected") {
		t.Fatalf("expected ws_connected metric to be present")
	}
	// Expect a positive gauge value while the connection is open
	re := regexp.MustCompile(`(?m)^ws_connected\s+([0-9.]+)$`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		t.Fatalf("ws_connected sample not found in metrics output")
	}
	// Basic check > 0
	if m[1] == "0" || m[1] == "0.0" {
		t.Fatalf("expected ws_connected > 0, got %s", m[1])
	}
}
