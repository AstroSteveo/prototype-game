//go:build ws

package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/sim"
)

func TestOriginValidation_DevMode(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	// Register with dev mode enabled
	RegisterWithStoreAndDevMode(mux, "/ws", fakeAuth{}, eng, nil, true)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with cross-origin request - should succeed in dev mode
	headers := http.Header{}
	headers.Set("Origin", "https://evil.example.com")

	c, _, err := nws.Dial(ctx, wsURL, &nws.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("dial with cross-origin should succeed in dev mode: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Verify we can actually use the connection
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
	// No need to read response, just verify connection works
}

func TestOriginValidation_ProductionMode(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	// Register with dev mode disabled (production mode)
	RegisterWithStoreAndDevMode(mux, "/ws", fakeAuth{}, eng, nil, false)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with cross-origin request - should fail in production mode
	headers := http.Header{}
	headers.Set("Origin", "https://evil.example.com")

	_, _, err := nws.Dial(ctx, wsURL, &nws.DialOptions{HTTPHeader: headers})
	if err == nil {
		t.Fatalf("dial with cross-origin should fail in production mode")
	}
	// Expect an error containing "origin" or "cross-origin"
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "origin") && !strings.Contains(errStr, "cross-origin") {
		t.Logf("got error (expected origin-related): %v", err)
	}
}

func TestOriginValidation_ProductionMode_LocalhostAllowed(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	// Register with dev mode disabled (production mode)
	RegisterWithStoreAndDevMode(mux, "/ws", fakeAuth{}, eng, nil, false)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with localhost origin - should succeed in production mode
	headers := http.Header{}
	headers.Set("Origin", "http://localhost:3000")

	c, _, err := nws.Dial(ctx, wsURL, &nws.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("dial with localhost origin should succeed in production mode: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Verify we can actually use the connection
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
}

func TestOriginValidation_ProductionMode_SameOriginAllowed(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 20, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	// Register with dev mode disabled (production mode)
	RegisterWithStoreAndDevMode(mux, "/ws", fakeAuth{}, eng, nil, false)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with same origin as the server - should succeed in production mode
	serverOrigin := srv.URL
	headers := http.Header{}
	headers.Set("Origin", serverOrigin)

	c, _, err := nws.Dial(ctx, wsURL, &nws.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("dial with same origin should succeed in production mode: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Verify we can actually use the connection
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
}
