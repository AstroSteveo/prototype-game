# Cross-Node Handover Implementation

This document describes the cross-node handover functionality implemented for US-601.

## Overview

Cross-node handover enables players to seamlessly move between cells owned by different simulation nodes. The implementation supports two modes:

1. **Reconnect Mode** (default): Client disconnects and reconnects to the target node
2. **Tunnel Mode**: Client maintains connection to original node, which proxies data from target node

## Architecture

### Components

- **NodeRegistry**: Tracks which nodes own which cells
- **CrossNodeHandoverService**: Manages handover coordination between nodes
- **HTTPCrossNodeService**: HTTP-based implementation for inter-node communication
- **WebSocket Transport**: Enhanced to handle cross-node events

### Data Flow

1. Player movement triggers handover detection in `checkAndHandoverLocked()`
2. If target cell belongs to remote node, cross-node handover is initiated
3. Handover token is generated for secure transfer
4. Client receives handover event with reconnection/tunneling instructions
5. Target node accepts player using handover token

## Configuration

### Command Line Options

```bash
# Default reconnect mode
./sim -port 8081 -node-id node1

# Enable tunnel mode  
./sim -port 8081 -node-id node1 -handover-mode tunnel
```

### Node Registration

```bash
# Register remote node
curl "http://localhost:8081/dev/node/register?id=node2&address=localhost&port=8082"

# Assign cell to node
curl "http://localhost:8081/dev/node/assign-cell?node_id=node2&cx=1&cz=0"

# View node info
curl "http://localhost:8081/dev/node/info"
```

## Protocol Messages

### Reconnect Mode

```json
{
  "type": "handover_start",
  "data": {
    "from": {"Cx": 0, "Cz": 0},
    "to": {"Cx": 1, "Cz": 0}, 
    "target_node": "node2",
    "reason": "cross_node_transfer"
  }
}
```

### Tunnel Mode

```json
{
  "type": "handover_tunnel",
  "data": {
    "from": {"Cx": 0, "Cz": 0},
    "to": {"Cx": 1, "Cz": 0},
    "target_node": "node2", 
    "reason": "cross_node_transfer",
    "tunnel_active": true
  }
}
```

## Testing

### Manual Testing

```bash
# Setup cross-node scenario
make run
curl "http://localhost:8081/dev/node/register?id=node2&address=localhost&port=8082"
curl "http://localhost:8081/dev/node/assign-cell?node_id=node2&cx=1&cz=0"

# Spawn and move player
curl "http://localhost:8081/dev/spawn?id=test&name=Test&x=250&z=100"
curl "http://localhost:8081/dev/vel?id=test&vx=10&vz=0"

# Verify handover
sleep 1
curl "http://localhost:8081/dev/players"
```

### Automated Tests

```bash
# Run cross-node integration tests
go test ./internal/sim -run TestEngineWithCrossNodeHandover -v

# Run all tests
make test test-ws
```

## Implementation Details

### Handover Detection

The `checkAndHandoverLocked()` function in `handovers.go` detects when a player crosses into a cell owned by a different node and initiates the cross-node transfer process.

### Token Management

Handover tokens are generated with 30-second expiration and cleaned up automatically. They ensure secure player transfers between nodes.

### Backward Compatibility

Local handovers continue to work unchanged. Cross-node functionality is only activated when:
1. A CrossNodeHandoverService is configured
2. Target cell is owned by a different node

## Future Enhancements

- Load balancing for automatic cell assignment
- Encrypted inter-node communication
- State synchronization for tunnel mode
- Metrics and monitoring for cross-node transfers