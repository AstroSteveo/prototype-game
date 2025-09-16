# ADR 0002: Cross-Node Handover Protocol

- **Status**: Proposed
- **Context**: 

The current system implements Phase A (local sharding) where all cells exist within a single simulation process. Players can handover between cells through in-process memory transfers with ~1ms latency. To scale beyond a single node's capacity, we need Phase B (cross-node sharding) where cells are distributed across multiple simulation nodes, requiring network-based handovers between processes.

**Key Challenges:**
1. **State Transfer**: Player state must be serialized, transmitted, and deserialized between nodes
2. **Connection Management**: Client WebSocket connections must either be tunneled or require reconnection
3. **Consistency**: Player state must not be lost or duplicated during node-to-node transfers
4. **Performance**: Target < 500ms handover latency while maintaining game feel
5. **Failure Handling**: Network failures, node crashes, or timeouts during handover must be recoverable

**Current Architecture Limitations:**
- `Engine.checkAndHandoverLocked()` only handles local cell transfers
- No node-to-node communication protocol exists
- Gateway has no concept of multiple simulation nodes
- Session state is stored locally in each simulation process

**Design Constraints:**
- Must be compatible with existing Phase A local handovers
- Should minimize client-visible interruptions (target: transparent handovers)
- Must handle node failures gracefully without player state loss
- Should support gradual rollout (feature flag controlled)

- **Decision**: 

Implement a **two-phase commit protocol** for cross-node handovers with the following components:

## 1. Handover Protocol State Machine

```
[Source Node]    [Target Node]    [Gateway]        [Client]
     |                |              |               |
     |-- PREPARE ----->|              |               |
     |<-- RESERVED ----|              |               |
     |-- COMMIT ------>|              |               |
     |<-- CONFIRMED ---|              |               |
     |                 |-- ROUTE ---->|               |
     |                 |<-- ACK ------|               |
     |                 |              |-- HANDOVER -->|
     |                 |              |<-- ACK -------|
     |-- CLEANUP ----->|              |               |
```

## 2. Network Protocol

**HTTP-based inter-node communication** (simple, reliable, debuggable):

```http
POST /handover/prepare
{
  "player_id": "p123",
  "from_cell": {"cx": 0, "cz": 0},
  "to_cell": {"cx": 1, "cz": 0},
  "player_state": {
    "pos": {"x": 10.1, "z": 5.0},
    "vel": {"x": 2.0, "z": 0.0},
    "yaw": 1.57,
    "sequence": 1234,
    "session_data": "...",
    "checksum": "abc123"
  },
  "handover_token": "uuid-v4",
  "expires_at": "2024-01-15T10:30:00Z"
}
```

## 3. Client Connection Handling

**Option A: Connection Tunneling** (Phase 1)
- Gateway proxies WebSocket messages between client and target node
- Client unaware of node change; connection remains stable
- Higher gateway load but simpler client logic

**Option B: Client Reconnection** (Phase 2 optimization)
- Gateway sends `handover_start` with new endpoint
- Client reconnects to target node with handover token
- Lower gateway overhead but requires client reconnection logic

## 4. Failure Recovery Mechanisms

**Timeout Handling:**
- PREPARE timeout (5s): Abort handover, player stays on source node
- COMMIT timeout (3s): Target node assumes success, source cleans up
- Gateway routing timeout (2s): Retry routing, fallback to old node

**Node Failure Scenarios:**
- Source node crash during PREPARE: Player reconnects to original node
- Target node crash during COMMIT: Source node retains player, retry later
- Gateway crash: Nodes use cached routing until gateway recovers

## 5. State Consistency Guarantees

**Exactly-once semantics** through:
- Unique handover tokens (prevent duplicates)
- State checksums (detect corruption)  
- Sequence number validation (prevent replay)
- Expiration timestamps (cleanup stale reservations)

**Rollback procedures:**
- PREPARE rejection: Source node continues normally
- COMMIT failure: Target releases reservation, source retains player
- Network partition: Players stay on last known good node

- **Consequences**: 

**Positive:**
- **Scalability**: Enables horizontal scaling beyond single-node limits
- **Load Distribution**: Cells can be assigned to least-loaded nodes
- **Fault Isolation**: Node failures only affect subset of players
- **Operational Flexibility**: Nodes can be drained for maintenance
- **Performance Visibility**: Clear metrics for cross-node handover latency

**Negative:**
- **Complexity**: Significantly more complex than local handovers
- **Network Dependency**: Handovers now subject to network failures
- **Latency**: Cross-node handovers will be 10-100x slower than local
- **Resource Overhead**: Additional HTTP connections and serialization
- **Debugging Difficulty**: Distributed system issues harder to diagnose

**Mitigation Strategies:**
- Implement comprehensive logging and tracing for handover flows
- Use feature flags to enable cross-node handovers gradually  
- Build extensive integration tests for failure scenarios
- Create monitoring dashboards for handover success rates and latency
- Design graceful degradation (fallback to local handovers if needed)

**Architecture Impact:**
- Gateway becomes stateful (node routing table, handover tokens)
- Simulation nodes need HTTP client/server for inter-node communication
- Player state serialization must be more robust (checksums, versioning)
- Session management becomes distributed (Redis or node replication)
- Metrics and monitoring must track cross-node operations

**Implementation Priority:**
1. **Phase 1**: Basic PREPARE/COMMIT protocol with connection tunneling
2. **Phase 2**: Failure recovery and timeout handling
3. **Phase 3**: Performance optimizations (client reconnection, compression)
4. **Phase 4**: Advanced load balancing and cell migration

**Testing Strategy:**
- Unit tests for handover state machine and serialization  
- Integration tests with real node-to-node transfers
- Chaos testing (network partitions, node failures)
- Load testing with concurrent cross-node handovers
- End-to-end validation of handover latency targets

**Open Questions for Future ADRs:**
- Cell assignment strategy (consistent hashing vs. centralized)
- Load balancing algorithm and trigger conditions
- Session state storage (Redis, node replication, or hybrid)
- Monitoring and alerting requirements for distributed operations