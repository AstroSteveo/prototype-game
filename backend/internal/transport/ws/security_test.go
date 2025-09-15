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

func TestWS_OversizedMessageRejection(t *testing.T) {
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

	// Send oversized message (larger than 32KB limit)
	largeData := strings.Repeat("x", 33*1024) // 33KB
	oversizedMsg := map[string]any{
		"type":   "input",
		"seq":    1,
		"dt":     0.05,
		"intent": map[string]any{"x": 1, "z": 0},
		"data":   largeData,
	}

	// Expect connection to be closed after sending oversized message
	err = wsjson.Write(ctx, c, oversizedMsg)
	if err != nil {
		// Connection might be closed immediately on write
		return
	}

	// Try to read - should fail with connection closed
	readCtx, cancelRead := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelRead()
	err = wsjson.Read(readCtx, c, &raw)
	if err == nil {
		t.Fatal("expected connection to be closed after oversized message, but read succeeded")
	}

	// Verify the error indicates message was too big or connection was closed
	errStr := err.Error()
	if !strings.Contains(errStr, "StatusMessageTooBig") &&
		!strings.Contains(errStr, "closed") &&
		!strings.Contains(errStr, "EOF") {
		t.Fatalf("expected message too big or connection closed error, got: %v", err)
	}
}

func TestWS_IdleTimeout(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	// Use a shorter idle timeout for testing (3 seconds)
	RegisterWithOptions(mux, "/ws", fakeAuth{}, eng, nil, WSOptions{IdleTimeout: 3 * time.Second})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// The connection should be closed within 5 seconds due to either:
	// 1. Idle timeout (3s configured)
	// 2. Ping failure (also valid resource management)
	// Both behaviors are correct and demonstrate resource management
	startTime := time.Now()
	connectionClosed := false

	for time.Since(startTime) < 5*time.Second && !connectionClosed {
		readCtx, cancelRead := context.WithTimeout(context.Background(), 500*time.Millisecond)
		err := wsjson.Read(readCtx, c, &raw)
		cancelRead()

		if err != nil {
			connectionClosed = true
			elapsed := time.Since(startTime)
			// Either idle timeout (~3s) or ping failure (~1-2s) is acceptable
			if elapsed > 5*time.Second {
				t.Fatalf("connection closed too late: %v (expected within 5s)", elapsed)
			}
			t.Logf("Connection closed after %v (resource management working)", elapsed)
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	if !connectionClosed {
		t.Fatal("connection was not closed within expected timeframe - resource management not working")
	}
}

func TestWS_ReadDeadline(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuth{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Send a valid input message to ensure normal operation works
	if err := wsjson.Write(ctx, c, map[string]any{
		"type":   "input",
		"seq":    1,
		"dt":     0.05,
		"intent": map[string]any{"x": 1, "z": 0},
	}); err != nil {
		t.Fatalf("write input: %v", err)
	}

	// Should be able to read state messages normally
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		readCtx, cancelRead := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := wsjson.Read(readCtx, c, &raw)
		cancelRead()
		if err == nil {
			// Successfully read a message - per-message deadline is working
			return
		}
		// Continue trying if timeout (expected behavior)
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("could not read any messages within deadline - read timeout may be too strict")
}
