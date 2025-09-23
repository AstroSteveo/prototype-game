package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestInit verifies that metrics initialization is idempotent and safe
func TestInit(t *testing.T) {
	// Multiple calls to Init() should not panic
	Init()
	Init()
	Init()
	
	// Handler should be available after init
	handler := Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil after Init()")
	}
}

// TestObserveTickDuration verifies tick duration histogram is updated
func TestObserveTickDuration(t *testing.T) {
	// Record a tick duration
	duration := 5 * time.Millisecond
	ObserveTickDuration(duration)
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_tick_time_ms histogram exists
	if !strings.Contains(metrics, "sim_tick_time_ms") {
		t.Fatal("Expected sim_tick_time_ms metric to be present")
	}
	
	// Check bucket count increased (at least one sample in a bucket)
	bucketPattern := regexp.MustCompile(`sim_tick_time_ms_bucket{[^}]*le="[^"]*"}\s+([1-9]\d*)`)
	if !bucketPattern.MatchString(metrics) {
		t.Fatal("Expected at least one sample in sim_tick_time_ms histogram buckets")
	}
}

// TestObserveSnapshotBytes verifies snapshot bytes histogram is updated
func TestObserveSnapshotBytes(t *testing.T) {
	// Record snapshot size
	snapshotSize := 1024
	ObserveSnapshotBytes(snapshotSize)
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that ws_snapshot_bytes histogram exists
	if !strings.Contains(metrics, "ws_snapshot_bytes") {
		t.Fatal("Expected ws_snapshot_bytes metric to be present")
	}
	
	// Check bucket count increased
	bucketPattern := regexp.MustCompile(`ws_snapshot_bytes_bucket{[^}]*le="[^"]*"}\s+([1-9]\d*)`)
	if !bucketPattern.MatchString(metrics) {
		t.Fatal("Expected at least one sample in ws_snapshot_bytes histogram buckets")
	}
}

// TestObserveEntitiesInAOI verifies AOI entities histogram is updated
func TestObserveEntitiesInAOI(t *testing.T) {
	// Record AOI query result
	entityCount := 8
	ObserveEntitiesInAOI(entityCount)
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_entities_in_aoi histogram exists
	if !strings.Contains(metrics, "sim_entities_in_aoi") {
		t.Fatal("Expected sim_entities_in_aoi metric to be present")
	}
	
	// Check bucket count increased
	bucketPattern := regexp.MustCompile(`sim_entities_in_aoi_bucket{[^}]*le="[^"]*"}\s+([1-9]\d*)`)
	if !bucketPattern.MatchString(metrics) {
		t.Fatal("Expected at least one sample in sim_entities_in_aoi histogram buckets")
	}
}

// TestObserveHandoverLatency verifies handover latency histogram is updated
func TestObserveHandoverLatency(t *testing.T) {
	// Record handover latency
	latency := 10 * time.Millisecond
	ObserveHandoverLatency(latency)
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_handover_latency_ms histogram exists
	if !strings.Contains(metrics, "sim_handover_latency_ms") {
		t.Fatal("Expected sim_handover_latency_ms metric to be present")
	}
	
	// Check bucket count increased
	bucketPattern := regexp.MustCompile(`sim_handover_latency_ms_bucket{[^}]*le="[^"]*"}\s+([1-9]\d*)`)
	if !bucketPattern.MatchString(metrics) {
		t.Fatal("Expected at least one sample in sim_handover_latency_ms histogram buckets")
	}
}

// TestWSConnectedGauge verifies WS connection gauge operations
func TestWSConnectedGauge(t *testing.T) {
	// Test increment
	IncWSConnected()
	IncWSConnected()
	
	// Scrape metrics after increments
	metrics := scrapeMetrics(t)
	
	// Check that ws_connected gauge exists and has positive value
	if !strings.Contains(metrics, "ws_connected") {
		t.Fatal("Expected ws_connected metric to be present")
	}
	
	re := regexp.MustCompile(`(?m)^ws_connected\s+([0-9.]+)$`)
	matches := re.FindStringSubmatch(metrics)
	if len(matches) < 2 {
		t.Fatal("ws_connected value not found in metrics output")
	}
	
	// Should be at least 2 (we incremented twice)
	if matches[1] == "0" || matches[1] == "0.0" {
		t.Fatal("Expected ws_connected > 0 after increments")
	}
	
	// Test decrement
	DecWSConnected()
	
	// Gauge should still be positive (we had 2, decremented 1)
	metrics2 := scrapeMetrics(t)
	matches2 := re.FindStringSubmatch(metrics2)
	if len(matches2) < 2 {
		t.Fatal("ws_connected value not found after decrement")
	}
	if matches2[1] == "0" || matches2[1] == "0.0" {
		t.Fatal("Expected ws_connected > 0 after single decrement")
	}
}

// TestIncHandovers verifies handover counter is updated
func TestIncHandovers(t *testing.T) {
	// Increment handover counter
	IncHandovers()
	IncHandovers()
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_handovers_total counter exists
	if !strings.Contains(metrics, "sim_handovers_total") {
		t.Fatal("Expected sim_handovers_total metric to be present")
	}
	
	// Check counter value
	re := regexp.MustCompile(`(?m)^sim_handovers_total\s+([0-9.]+)$`)
	matches := re.FindStringSubmatch(metrics)
	if len(matches) < 2 {
		t.Fatal("sim_handovers_total value not found in metrics output")
	}
	
	// Should be at least 2 (we incremented twice)
	if matches[1] == "0" || matches[1] == "0.0" {
		t.Fatal("Expected sim_handovers_total > 0 after increments")
	}
}

// TestObserveEquipOperation verifies equipment operation counters are updated
func TestObserveEquipOperation(t *testing.T) {
	// Record successful equip operation
	ObserveEquipOperation("equip", true)
	
	// Record failed unequip operation
	ObserveEquipOperation("unequip", false)
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_equip_operations_total counter exists
	if !strings.Contains(metrics, "sim_equip_operations_total") {
		t.Fatal("Expected sim_equip_operations_total metric to be present")
	}
	
	// Check for specific label combinations
	successPattern := regexp.MustCompile(`sim_equip_operations_total{[^}]*operation="equip"[^}]*result="success"[^}]*}\s+([1-9]\d*|1)`)
	if !successPattern.MatchString(metrics) {
		t.Fatal("Expected sim_equip_operations_total with operation=equip,result=success")
	}
	
	failPattern := regexp.MustCompile(`sim_equip_operations_total{[^}]*operation="unequip"[^}]*result="failed"[^}]*}\s+([1-9]\d*|1)`)
	if !failPattern.MatchString(metrics) {
		t.Fatal("Expected sim_equip_operations_total with operation=unequip,result=failed")
	}
}

// TestIncEquipCooldownBlocks verifies equip cooldown counter is updated
func TestIncEquipCooldownBlocks(t *testing.T) {
	// Increment cooldown blocks counter
	IncEquipCooldownBlocks()
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check that sim_equip_cooldown_blocks_total counter exists
	if !strings.Contains(metrics, "sim_equip_cooldown_blocks_total") {
		t.Fatal("Expected sim_equip_cooldown_blocks_total metric to be present")
	}
	
	// Check counter value
	re := regexp.MustCompile(`(?m)^sim_equip_cooldown_blocks_total\s+([1-9]\d*|1)`)
	matches := re.FindStringSubmatch(metrics)
	if len(matches) < 2 {
		t.Fatal("Expected sim_equip_cooldown_blocks_total > 0")
	}
}

// TestMetricsEndpointFormat verifies the metrics endpoint returns valid Prometheus format
func TestMetricsEndpointFormat(t *testing.T) {
	// Generate some sample data first
	ObserveTickDuration(time.Millisecond)
	ObserveSnapshotBytes(512)
	ObserveEntitiesInAOI(4)
	ObserveHandoverLatency(5 * time.Millisecond)
	IncWSConnected()
	IncHandovers()
	ObserveEquipOperation("equip", true)
	IncEquipCooldownBlocks()
	
	// Scrape metrics
	metrics := scrapeMetrics(t)
	
	// Check for required metric families
	requiredMetrics := []string{
		"sim_tick_time_ms",
		"ws_snapshot_bytes", 
		"sim_entities_in_aoi",
		"sim_handover_latency_ms",
		"ws_connected",
		"sim_handovers_total",
		"sim_equip_operations_total",
		"sim_equip_cooldown_blocks_total",
	}
	
	for _, metric := range requiredMetrics {
		if !strings.Contains(metrics, metric) {
			t.Errorf("Missing required metric: %s", metric)
		}
	}
	
	// Basic format validation - should contain HELP and TYPE comments
	if !strings.Contains(metrics, "# HELP") {
		t.Error("Expected HELP comments in metrics output")
	}
	if !strings.Contains(metrics, "# TYPE") {
		t.Error("Expected TYPE comments in metrics output")
	}
}

// scrapeMetrics creates a test server and scrapes the metrics endpoint
func scrapeMetrics(t *testing.T) string {
	handler := Handler()
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