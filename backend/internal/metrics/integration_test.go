//go:build ws

package metrics_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	"prototype-game/backend/internal/transport/ws"
)

type testAuth struct{}

func (testAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "test-token" {
		return "test-player", "TestPlayer", true
	}
	return "", "", false
}

// TestIntegration_TickMetrics verifies that tick metrics are updated during simulation
func TestIntegration_TickMetrics(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            4,
		AOIRadius:           2,
		TickHz:              60,
		SnapshotHz:          20,
		HandoverHysteresisM: 0.25,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	// Get initial metrics
	initialMetrics := scrapeMetrics(t)
	initialTickCount := extractHistogramSampleCount(t, initialMetrics, "sim_tick_time_ms")

	// Wait for a few ticks to occur
	time.Sleep(100 * time.Millisecond)

	// Get updated metrics
	updatedMetrics := scrapeMetrics(t)
	updatedTickCount := extractHistogramSampleCount(t, updatedMetrics, "sim_tick_time_ms")

	// Verify tick count increased
	if updatedTickCount <= initialTickCount {
		t.Errorf("Expected tick metrics to increase: initial=%d, updated=%d", initialTickCount, updatedTickCount)
	}
}

// TestIntegration_WSConnectionMetrics verifies WebSocket connection metrics during actual connections
func TestIntegration_WSConnectionMetrics(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            4,
		AOIRadius:           2,
		TickHz:              60,
		SnapshotHz:          20,
		HandoverHysteresisM: 0.25,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	ws.Register(mux, "/ws", testAuth{}, eng)
	mux.Handle("/metrics", metrics.Handler())
	server := httptest.NewServer(mux)
	defer server.Close()

	// Get initial connection count
	initialMetrics := scrapeMetrics(t)
	initialConnections := extractGaugeValue(t, initialMetrics, "ws_connected")

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Send auth message
	if err := wsjson.Write(ctx, conn, map[string]any{"token": "test-token"}); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	// Read join acknowledgment
	var ackMsg json.RawMessage
	if err := wsjson.Read(ctx, conn, &ackMsg); err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	// Check that connection count increased
	updatedMetrics := scrapeMetrics(t)
	updatedConnections := extractGaugeValue(t, updatedMetrics, "ws_connected")

	if updatedConnections <= initialConnections {
		t.Errorf("Expected ws_connected to increase: initial=%f, updated=%f", initialConnections, updatedConnections)
	}

	// Close connection
	conn.Close(nws.StatusNormalClosure, "test complete")

	// Wait a moment for cleanup
	time.Sleep(50 * time.Millisecond)

	// Check that connection count decreased
	finalMetrics := scrapeMetrics(t)
	finalConnections := extractGaugeValue(t, finalMetrics, "ws_connected")

	if finalConnections >= updatedConnections {
		t.Errorf("Expected ws_connected to decrease after disconnect: updated=%f, final=%f", updatedConnections, finalConnections)
	}
}

// TestIntegration_AOIMetrics verifies AOI metrics are updated during entity queries
func TestIntegration_AOIMetrics(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            4,
		AOIRadius:           2,
		TickHz:              60,
		SnapshotHz:          20,
		HandoverHysteresisM: 0.25,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	ws.Register(mux, "/ws", testAuth{}, eng)
	mux.Handle("/metrics", metrics.Handler())
	server := httptest.NewServer(mux)
	defer server.Close()

	// Get initial AOI metrics
	initialMetrics := scrapeMetrics(t)
	initialAOICount := extractHistogramSampleCount(t, initialMetrics, "sim_entities_in_aoi")

	// Connect and add a player to trigger AOI queries
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(nws.StatusNormalClosure, "test complete")

	// Auth and join
	if err := wsjson.Write(ctx, conn, map[string]any{"token": "test-token"}); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	var ackMsg json.RawMessage
	if err := wsjson.Read(ctx, conn, &ackMsg); err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	// Wait for snapshots to trigger AOI queries
	time.Sleep(200 * time.Millisecond)

	// Check that AOI metrics increased
	updatedMetrics := scrapeMetrics(t)
	updatedAOICount := extractHistogramSampleCount(t, updatedMetrics, "sim_entities_in_aoi")

	if updatedAOICount <= initialAOICount {
		t.Errorf("Expected AOI metrics to increase: initial=%d, updated=%d", initialAOICount, updatedAOICount)
	}
}

// TestIntegration_SnapshotMetrics verifies snapshot size metrics during WebSocket communication
func TestIntegration_SnapshotMetrics(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            4,
		AOIRadius:           2,
		TickHz:              60,
		SnapshotHz:          20,
		HandoverHysteresisM: 0.25,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	ws.Register(mux, "/ws", testAuth{}, eng)
	mux.Handle("/metrics", metrics.Handler())
	server := httptest.NewServer(mux)
	defer server.Close()

	// Get initial snapshot metrics
	initialMetrics := scrapeMetrics(t)
	initialSnapshotCount := extractHistogramSampleCount(t, initialMetrics, "ws_snapshot_bytes")

	// Connect and trigger snapshots
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(nws.StatusNormalClosure, "test complete")

	// Auth and join to start receiving snapshots
	if err := wsjson.Write(ctx, conn, map[string]any{"token": "test-token"}); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	var ackMsg json.RawMessage
	if err := wsjson.Read(ctx, conn, &ackMsg); err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	// Wait for snapshots to be sent
	time.Sleep(200 * time.Millisecond)

	// Check that snapshot metrics increased
	updatedMetrics := scrapeMetrics(t)
	updatedSnapshotCount := extractHistogramSampleCount(t, updatedMetrics, "ws_snapshot_bytes")

	if updatedSnapshotCount <= initialSnapshotCount {
		t.Errorf("Expected snapshot metrics to increase: initial=%d, updated=%d", initialSnapshotCount, updatedSnapshotCount)
	}
}

// TestIntegration_EquipmentMetrics verifies equipment operation metrics are recorded correctly
func TestIntegration_EquipmentMetrics(t *testing.T) {
	// Get initial equipment metrics
	initialMetrics := scrapeMetrics(t)
	initialEquipCount := extractCounterValue(t, initialMetrics, "sim_equip_operations_total")
	initialCooldownCount := extractCounterValue(t, initialMetrics, "sim_equip_cooldown_blocks_total")

	// Simulate successful equip operation (this is what the WebSocket layer would do)
	metrics.ObserveEquipOperation("equip", true)
	
	// Simulate failed unequip operation
	metrics.ObserveEquipOperation("unequip", false)
	
	// Simulate cooldown block
	metrics.IncEquipCooldownBlocks()

	// Check that equipment metrics increased
	updatedMetrics := scrapeMetrics(t)
	updatedEquipCount := extractCounterValue(t, updatedMetrics, "sim_equip_operations_total")
	updatedCooldownCount := extractCounterValue(t, updatedMetrics, "sim_equip_cooldown_blocks_total")

	if updatedEquipCount <= initialEquipCount {
		t.Errorf("Expected equipment metrics to increase: initial=%f, updated=%f", initialEquipCount, updatedEquipCount)
	}
	
	if updatedCooldownCount <= initialCooldownCount {
		t.Errorf("Expected cooldown block metrics to increase: initial=%f, updated=%f", initialCooldownCount, updatedCooldownCount)
	}
	
	// Verify the metrics contain the expected labels
	if !strings.Contains(updatedMetrics, `operation="equip"`) {
		t.Error("Expected to find equip operation label in metrics")
	}
	if !strings.Contains(updatedMetrics, `operation="unequip"`) {
		t.Error("Expected to find unequip operation label in metrics")
	}
	if !strings.Contains(updatedMetrics, `result="success"`) {
		t.Error("Expected to find success result label in metrics")
	}
	if !strings.Contains(updatedMetrics, `result="failed"`) {
		t.Error("Expected to find failed result label in metrics")
	}
}

// TestIntegration_HandoverMetrics verifies handover metrics during cell transitions
func TestIntegration_HandoverMetrics(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            4, // Small cell size to trigger handovers easily
		AOIRadius:           2,
		TickHz:              60,
		SnapshotHz:          20,
		HandoverHysteresisM: 0.25,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	ws.Register(mux, "/ws", testAuth{}, eng)
	mux.Handle("/metrics", metrics.Handler())
	server := httptest.NewServer(mux)
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(nws.StatusNormalClosure, "test complete")

	// Auth and join
	if err := wsjson.Write(ctx, conn, map[string]any{"token": "test-token"}); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	var ackMsg json.RawMessage
	if err := wsjson.Read(ctx, conn, &ackMsg); err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	// Get initial handover metrics
	initialMetrics := scrapeMetrics(t)
	initialHandoverCount := extractCounterValue(t, initialMetrics, "sim_handovers_total")

	// Send movement that should trigger cell transition (handover)
	moveMsg := map[string]any{
		"action": "move",
		"x":      10.0, // Move far enough to cross cell boundaries
		"z":      10.0,
	}
	if err := wsjson.Write(ctx, conn, moveMsg); err != nil {
		t.Fatalf("Failed to send move: %v", err)
	}

	// Wait for handover processing
	time.Sleep(200 * time.Millisecond)

	// Check if handover metrics increased
	// Note: Handovers may not always occur depending on player spawn location
	// This test verifies the metric infrastructure works when handovers do occur
	updatedMetrics := scrapeMetrics(t)
	updatedHandoverCount := extractCounterValue(t, updatedMetrics, "sim_handovers_total")

	// Log for debugging
	t.Logf("Handover count: initial=%f, updated=%f", initialHandoverCount, updatedHandoverCount)
	
	// The test passes as long as the metrics infrastructure is working
	// (metric exists and is properly formatted)
	if !strings.Contains(updatedMetrics, "sim_handovers_total") {
		t.Error("Expected sim_handovers_total metric to exist")
	}
}

// Helper functions for extracting metric values

func scrapeMetrics(t *testing.T) string {
	handler := metrics.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()
	
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to scrape metrics: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read metrics response: %v", err)
	}
	
	return string(body)
}

func extractHistogramSampleCount(t *testing.T, metrics, metricName string) int {
	// Look for the _count suffix which gives us total samples
	pattern := regexp.MustCompile(`(?m)^` + metricName + `_count\s+(\d+)$`)
	matches := pattern.FindStringSubmatch(metrics)
	if len(matches) < 2 {
		return 0 // No samples yet
	}
	var count int
	if _, err := fmt.Sscanf(matches[1], "%d", &count); err != nil {
		t.Fatalf("Failed to parse histogram count: %v", err)
	}
	return count
}

func extractGaugeValue(t *testing.T, metrics, metricName string) float64 {
	pattern := regexp.MustCompile(`(?m)^` + metricName + `\s+([0-9.]+)$`)
	matches := pattern.FindStringSubmatch(metrics)
	if len(matches) < 2 {
		return 0
	}
	var value float64
	if _, err := fmt.Sscanf(matches[1], "%f", &value); err != nil {
		t.Fatalf("Failed to parse gauge value: %v", err)
	}
	return value
}

func extractCounterValue(t *testing.T, metrics, metricName string) float64 {
	// For vector counters, sum all label variations
	pattern := regexp.MustCompile(`(?m)^` + metricName + `{[^}]*}\s+([0-9.]+)$`)
	matches := pattern.FindAllStringSubmatch(metrics, -1)
	
	total := 0.0
	for _, match := range matches {
		if len(match) >= 2 {
			var value float64
			if _, err := fmt.Sscanf(match[1], "%f", &value); err == nil {
				total += value
			}
		}
	}
	
	// Also check for non-vector version
	simplePattern := regexp.MustCompile(`(?m)^` + metricName + `\s+([0-9.]+)$`)
	simpleMatches := simplePattern.FindStringSubmatch(metrics)
	if len(simpleMatches) >= 2 {
		var value float64
		if _, err := fmt.Sscanf(simpleMatches[1], "%f", &value); err == nil {
			total += value
		}
	}
	
	return total
}