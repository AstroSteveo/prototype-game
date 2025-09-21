#!/bin/bash
# Performance testing script for Prototype Game Backend

set -e

# Configuration
GATEWAY_URL=${GATEWAY_URL:-"http://localhost:8080"}
SIM_URL=${SIM_URL:-"http://localhost:8081"}
NUM_CLIENTS=${NUM_CLIENTS:-100}
TEST_DURATION=${TEST_DURATION:-60}
RAMP_UP_TIME=${RAMP_UP_TIME:-10}

echo "🚀 Starting performance tests for Prototype Game Backend"
echo "Configuration:"
echo "  Gateway URL: $GATEWAY_URL"
echo "  Simulation URL: $SIM_URL"
echo "  Number of clients: $NUM_CLIENTS"
echo "  Test duration: ${TEST_DURATION}s"
echo "  Ramp-up time: ${RAMP_UP_TIME}s"
echo ""

# Check if services are running
echo "📋 Checking service health..."
if ! curl -f "$GATEWAY_URL/healthz" > /dev/null 2>&1; then
    echo "❌ Gateway is not healthy at $GATEWAY_URL"
    exit 1
fi

if ! curl -f "$SIM_URL/healthz" > /dev/null 2>&1; then
    echo "❌ Simulation is not healthy at $SIM_URL"
    exit 1
fi
echo "✅ All services are healthy"

# Create test directory
TEST_DIR="performance_tests_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$TEST_DIR"

echo "📊 Running performance tests..."
echo "Results will be saved to: $TEST_DIR"

# Function to run a single client test
run_client_test() {
    local client_id=$1
    local output_file="$TEST_DIR/client_${client_id}.log"
    
    {
        echo "Client $client_id starting at $(date)"
        
        # Get auth token
        TOKEN=$(curl -s "$GATEWAY_URL/login?name=PerfTest$client_id" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])' 2>/dev/null || echo "failed")
        
        if [ "$TOKEN" = "failed" ]; then
            echo "Client $client_id: Failed to get auth token"
            return 1
        fi
        
        # Run WebSocket probe
        cd backend
        timeout $TEST_DURATION go run ./cmd/wsprobe -url "ws://localhost:8081/ws" -token "$TOKEN" -move_x 1 -move_z 1 2>&1 | while IFS= read -r line; do
            echo "$(date '+%Y-%m-%d %H:%M:%S') Client $client_id: $line"
        done
        cd ..
        
        echo "Client $client_id finished at $(date)"
    } > "$output_file" 2>&1 &
    
    echo $!
}

# Function to monitor system resources
monitor_resources() {
    local output_file="$TEST_DIR/resources.log"
    
    {
        echo "Resource monitoring started at $(date)"
        echo "Time,CPU%,Memory%,Load1,Load5,Load15,NetRX,NetTX"
        
        while true; do
            # Get CPU and memory usage
            CPU=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
            MEM=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
            
            # Get load averages
            LOAD=$(uptime | awk -F'load average:' '{print $2}' | sed 's/,//g')
            
            # Get network stats (simplified)
            NET_RX=$(cat /proc/net/dev | grep eth0 | awk '{print $2}' 2>/dev/null || echo "0")
            NET_TX=$(cat /proc/net/dev | grep eth0 | awk '{print $10}' 2>/dev/null || echo "0")
            
            echo "$(date '+%Y-%m-%d %H:%M:%S'),$CPU,$MEM,$LOAD,$NET_RX,$NET_TX"
            
            sleep 1
        done
    } > "$output_file" &
    
    echo $!
}

# Function to collect application metrics
collect_metrics() {
    local output_file="$TEST_DIR/metrics.log"
    
    {
        echo "Metrics collection started at $(date)"
        
        while true; do
            echo "=== $(date) ==="
            echo "Gateway metrics:"
            curl -s "$GATEWAY_URL/metrics" 2>/dev/null | grep -E "(http_requests_total|websocket_connections|response_time)" | head -10
            echo ""
            echo "Simulation metrics:"
            curl -s "$SIM_URL/metrics" 2>/dev/null | grep -E "(players_count|entities_count|tick_duration)" | head -10
            echo ""
            sleep 5
        done
    } > "$output_file" &
    
    echo $!
}

# Start resource monitoring
MONITOR_PID=$(monitor_resources)
echo "📈 Started resource monitoring (PID: $MONITOR_PID)"

# Start metrics collection
METRICS_PID=$(collect_metrics)
echo "📊 Started metrics collection (PID: $METRICS_PID)"

# Array to store client PIDs
CLIENT_PIDS=()

# Ramp up clients gradually
echo "🚀 Starting client ramp-up..."
RAMP_INTERVAL=$(echo "scale=2; $RAMP_UP_TIME / $NUM_CLIENTS" | bc)

for ((i=1; i<=NUM_CLIENTS; i++)); do
    CLIENT_PID=$(run_client_test $i)
    CLIENT_PIDS+=($CLIENT_PID)
    echo "Started client $i (PID: $CLIENT_PID)"
    
    # Wait for ramp-up interval
    if [ $i -lt $NUM_CLIENTS ]; then
        sleep $RAMP_INTERVAL
    fi
done

echo "✅ All $NUM_CLIENTS clients started"
echo "⏱️ Running test for ${TEST_DURATION} seconds..."

# Wait for test duration
sleep $TEST_DURATION

echo "⏹️ Stopping all clients..."

# Stop all clients
for pid in "${CLIENT_PIDS[@]}"; do
    kill $pid 2>/dev/null || true
done

# Stop monitoring
kill $MONITOR_PID 2>/dev/null || true
kill $METRICS_PID 2>/dev/null || true

echo "📋 Collecting final results..."

# Generate test summary
{
    echo "Performance Test Summary"
    echo "======================="
    echo "Test completed at: $(date)"
    echo "Configuration:"
    echo "  - Clients: $NUM_CLIENTS"
    echo "  - Duration: ${TEST_DURATION}s"
    echo "  - Ramp-up: ${RAMP_UP_TIME}s"
    echo ""
    
    echo "Client Results:"
    echo "---------------"
    
    SUCCESSFUL_CLIENTS=0
    FAILED_CLIENTS=0
    
    for ((i=1; i<=NUM_CLIENTS; i++)); do
        if [ -f "$TEST_DIR/client_$i.log" ]; then
            if grep -q "finished at" "$TEST_DIR/client_$i.log"; then
                SUCCESSFUL_CLIENTS=$((SUCCESSFUL_CLIENTS + 1))
            else
                FAILED_CLIENTS=$((FAILED_CLIENTS + 1))
            fi
        else
            FAILED_CLIENTS=$((FAILED_CLIENTS + 1))
        fi
    done
    
    echo "  - Successful clients: $SUCCESSFUL_CLIENTS"
    echo "  - Failed clients: $FAILED_CLIENTS"
    echo "  - Success rate: $(echo "scale=2; $SUCCESSFUL_CLIENTS * 100 / $NUM_CLIENTS" | bc)%"
    echo ""
    
    echo "Error Summary:"
    echo "--------------"
    grep -h "error\|Error\|ERROR\|failed\|Failed\|FAILED" "$TEST_DIR"/client_*.log 2>/dev/null | sort | uniq -c | head -10
    echo ""
    
    echo "Resource Usage Peak:"
    echo "-------------------"
    if [ -f "$TEST_DIR/resources.log" ]; then
        echo "Peak CPU: $(tail -n +2 "$TEST_DIR/resources.log" | cut -d, -f2 | sort -n | tail -1)%"
        echo "Peak Memory: $(tail -n +2 "$TEST_DIR/resources.log" | cut -d, -f3 | sort -n | tail -1)%"
        echo "Peak Load: $(tail -n +2 "$TEST_DIR/resources.log" | cut -d, -f4 | sort -n | tail -1)"
    fi
    
} > "$TEST_DIR/summary.txt"

echo "📊 Test Results:"
cat "$TEST_DIR/summary.txt"

echo ""
echo "📁 Full results available in: $TEST_DIR"
echo "   - summary.txt: Test summary"
echo "   - client_*.log: Individual client logs"
echo "   - resources.log: System resource usage"
echo "   - metrics.log: Application metrics"

# Generate simple HTML report
cat > "$TEST_DIR/report.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Performance Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: green; }
        .error { color: red; }
        pre { background: #f8f8f8; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Prototype Game Backend Performance Test</h1>
        <p><strong>Test Date:</strong> $(date)</p>
        <p><strong>Clients:</strong> $NUM_CLIENTS | <strong>Duration:</strong> ${TEST_DURATION}s | <strong>Success Rate:</strong> $(echo "scale=2; $SUCCESSFUL_CLIENTS * 100 / $NUM_CLIENTS" | bc)%</p>
    </div>
    
    <div class="section">
        <h2>📊 Summary</h2>
        <pre>$(cat "$TEST_DIR/summary.txt")</pre>
    </div>
    
    <div class="section">
        <h2>📈 Quick Analysis</h2>
        <ul>
            <li class="$([ $SUCCESSFUL_CLIENTS -gt $((NUM_CLIENTS * 80 / 100)) ] && echo 'success' || echo 'error')">
                Client Success Rate: $(echo "scale=1; $SUCCESSFUL_CLIENTS * 100 / $NUM_CLIENTS" | bc)%
            </li>
            <li>Total Test Duration: ${TEST_DURATION} seconds</li>
            <li>Concurrent Users: $NUM_CLIENTS</li>
        </ul>
    </div>
</body>
</html>
EOF

echo "📄 HTML report generated: $TEST_DIR/report.html"
echo ""
echo "🎉 Performance testing completed!"

# Open report if running on desktop
if command -v xdg-open > /dev/null 2>&1; then
    xdg-open "$TEST_DIR/report.html" 2>/dev/null &
elif command -v open > /dev/null 2>&1; then
    open "$TEST_DIR/report.html" 2>/dev/null &
fi