# Persistence Verification Guide

This guide covers how to verify that inventory and equipment persistence is working correctly in the prototype game, including both happy path scenarios and failure injection testing.

## Prerequisites

- Go 1.23+ installed
- Prototype game built (`make build`)
- Optional: PostgreSQL database for production-like testing

## Happy Path Verification

### 1. File Store Testing (Recommended for Development)

Start the sim service with file-based persistence:

```bash
cd backend
./bin/sim -store-file=/tmp/test_persistence.json
```

In another terminal, start the gateway:

```bash
cd backend  
./bin/gateway
```

### 2. Test Basic Persistence Flow

1. **Connect and verify initial state:**
   ```bash
   TOKEN=$(make login)
   make wsprobe TOKEN="$TOKEN"
   ```
   
   Verify the join_ack contains:
   - `inventory` with empty items array and default compartment caps
   - `equipment` with empty slots object
   - `skills` with empty object
   - `encumbrance` with default weight/bulk limits

2. **Check persistence file:**
   ```bash
   cat /tmp/test_persistence.json
   ```
   
   Should show player state with all required fields:
   - `pos`: Player position
   - `inventory_data`: Serialized inventory state
   - `equipment_data`: Serialized equipment state
   - `skills_data`: Serialized skills
   - `cooldown_timers`: Equipment cooldown data
   - `encumbrance_config`: Weight/bulk limits

3. **Test reconnection (simulated disconnect/reconnect):**
   ```bash
   # Connect again with new token (different player ID for now)
   TOKEN2=$(make login)
   make wsprobe TOKEN="$TOKEN2"
   ```

### 3. Verify Persistence Metrics

Check persistence metrics during operation:

```bash
curl http://localhost:8081/persistence/metrics.json
```

Should return metrics including:
- `persist_attempts`: Number of persistence operations attempted
- `persist_successes`: Number of successful saves
- `persist_failures`: Number of failed saves
- `avg_persist_duration`: Average persistence latency
- Queue lengths for checkpoints and disconnects

### 4. Check Combined Metrics

View simulation and persistence metrics together:

```bash
curl http://localhost:8081/metrics.json
```

Should return nested object with both `simulation` and `persistence` sections.

## Advanced Testing

### 1. PostgreSQL Store Testing

For production-like testing with optimistic locking:

1. **Set up test database:**
   ```sql
   CREATE DATABASE prototype_game_test;
   CREATE USER test_user WITH PASSWORD 'test_pass';
   GRANT ALL PRIVILEGES ON DATABASE prototype_game_test TO test_user;
   ```

2. **Start sim with PostgreSQL:**
   ```bash
   ./bin/sim -store-postgres="postgres://test_user:test_pass@localhost/prototype_game_test?sslmode=disable"
   ```

3. **Verify schema creation:**
   ```sql
   \dt  -- Should show player_state table
   SELECT * FROM player_state;  -- Should be empty initially
   ```

### 2. Load Testing Persistence

Test persistence under load using multiple concurrent connections:

```bash
# Terminal 1: Monitor persistence metrics
watch -n 1 'curl -s http://localhost:8081/persistence/metrics.json | jq'

# Terminal 2-5: Create multiple concurrent connections
for i in {1..10}; do
  TOKEN=$(make login)
  timeout 10 make wsprobe TOKEN="$TOKEN" &
done
wait
```

Monitor persistence queue lengths and success rates during load.

## Failure Injection Testing

### 1. Database Connection Failures

Test persistence resilience when database is unavailable:

1. **Start with working database, then disconnect:**
   ```bash
   # Start normally
   ./bin/sim -store-postgres="postgres://user:pass@localhost/test?sslmode=disable"
   
   # Stop PostgreSQL service
   sudo systemctl stop postgresql
   # or
   docker stop postgres-container
   ```

2. **Monitor error handling:**
   - Check sim logs for persistence failures
   - Verify application continues running
   - Check metrics show increased failure count

3. **Restart database and verify recovery:**
   ```bash
   sudo systemctl start postgresql
   # Check if persistence resumes working
   ```

### 2. Disk Space Exhaustion (File Store)

Test file store behavior when disk is full:

1. **Fill up disk space in /tmp:**
   ```bash
   # Fill up space (be careful with this!)
   dd if=/dev/zero of=/tmp/fill_disk bs=1M count=1000
   ```

2. **Attempt persistence operations:**
   ```bash
   TOKEN=$(make login)
   make wsprobe TOKEN="$TOKEN"
   ```

3. **Check error handling:**
   - Verify application doesn't crash
   - Check logs for disk space errors
   - Verify graceful degradation

4. **Clean up:**
   ```bash
   rm /tmp/fill_disk
   ```

### 3. Optimistic Locking Conflicts

Test optimistic locking with concurrent updates:

```bash
# This requires custom test script or manual database manipulation
# Example SQL to simulate conflict:
UPDATE player_state SET version = version + 1 WHERE player_id = 'test-player';
```

### 4. Serialization Failures

Test handling of corrupted persistence data:

1. **Corrupt persistence file:**
   ```bash
   echo "invalid json" > /tmp/test_persistence.json
   ```

2. **Restart sim and verify error handling:**
   ```bash
   ./bin/sim -store-file=/tmp/test_persistence.json
   # Should log error but continue with empty state
   ```

## Performance Verification

### 1. Persistence Latency

Monitor persistence performance under normal load:

```bash
# Expected latency targets (from technical design):
# - Disconnect persistence: < 500ms
# - Checkpoint persistence: < 1s
# - Reconnect hydration: < 2s

curl http://localhost:8081/persistence/metrics.json | jq '.avg_persist_duration'
```

### 2. Memory Usage

Monitor memory growth during extended operation:

```bash
# Monitor sim process memory
top -p $(pgrep sim)

# Check for persistence queue buildup
watch -n 5 'curl -s http://localhost:8081/persistence/metrics.json | jq "{checkpoint_queue: .checkpoint_queue_len, disconnect_queue: .disconnect_queue_len}"'
```

## Expected Behavior

### Normal Operation

- **Persistence latency:** < 100ms for in-memory store, < 500ms for database
- **No data loss:** Player state always recoverable after disconnect
- **Queue management:** Queues should remain small (< 10 items) under normal load
- **Error resilience:** Application continues running even if persistence fails

### Failure Scenarios

- **Database unavailable:** Application logs errors but continues, queues may grow
- **Disk full:** File store operations fail gracefully, application continues
- **Optimistic lock conflicts:** Conflicting saves are rejected with appropriate error
- **Serialization errors:** Corrupted data is detected and logged, defaults used

## Troubleshooting

### Common Issues

1. **High persistence failure rate:**
   - Check database connectivity
   - Verify disk space for file store
   - Monitor database performance

2. **Growing persistence queues:**
   - Check persistence worker health
   - Monitor database/file system performance
   - Verify no deadlocks in persistence operations

3. **Data not persisting:**
   - Verify store configuration (-store-file or database URL)
   - Check file/database permissions
   - Monitor persistence manager startup logs

### Debug Commands

```bash
# Check sim service status
curl http://localhost:8081/healthz

# View detailed persistence metrics
curl http://localhost:8081/persistence/metrics.json | jq

# Check persistence file contents (file store)
cat /tmp/test_persistence.json | jq

# Monitor real-time persistence activity
tail -f backend/logs/sim.log | grep -i persist
```

This verification approach ensures the persistence system meets the requirements outlined in US-006 for reliable inventory and equipment preservation through reconnects.