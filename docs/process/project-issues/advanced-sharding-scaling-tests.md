# Advanced Sharding and Scaling Test Suite

## Issue Summary

Develop a comprehensive test suite that focuses on advanced sharding and scaling capabilities of the server, including cross-node sharding (Phase B), load testing, and distributed system resilience scenarios.

## Context

The current system implements **Phase A (local sharding)** within a single process with cells as local instances. The existing test suite covers:
- Basic handovers between local cells
- Bot density control within single nodes
- AOI streaming and continuity 
- Single-node performance under moderate load

**Missing capabilities that need test coverage:**
- **Phase B (Cross-Node Sharding)**: Multi-node cell distribution and cross-node handovers
- **High-Load Stress Testing**: Performance under hundreds of concurrent players
- **Distributed System Resilience**: Network partitions, node failures, split-brain scenarios
- **Load Balancing**: Dynamic cell reassignment and load distribution
- **Scaling Bottlenecks**: Database, session state, and memory scaling limits

## Proposed Test Suite Categories

### 1. Cross-Node Sharding Tests (Phase B)

**Missing Logic**: The current system has no cross-node handover implementation. Need to design and test:

- **Cell Assignment Algorithm**: How cells are distributed across multiple sim nodes
- **Cross-Node Handover Protocol**: Player state serialization and node-to-node transfer
- **Gateway Routing**: How gateway directs clients to appropriate sim nodes
- **State Synchronization**: Ensuring consistency during cross-node transfers

**Test Scenarios:**
```go
// Test cross-node handover latency and success rate
func TestCrossNodeHandoverLatency(t *testing.T)
func TestCrossNodeHandoverStateConsistency(t *testing.T) 
func TestCrossNodeHandoverWithHighLoad(t *testing.T)

// Test cell ownership and reassignment
func TestCellReassignmentDuringNodeFailure(t *testing.T)
func TestLoadBasedCellMigration(t *testing.T)
```

### 2. Stress and Performance Tests

**Test high entity counts and concurrent players:**
```go
// Benchmark tests for scaling limits
func BenchmarkSimulationWith1000Players(b *testing.B)
func BenchmarkAOIQueriesUnderLoad(b *testing.B) 
func BenchmarkHandoverThroughput(b *testing.B)

// Stress tests for resource limits  
func TestMemoryUsageUnder10000Entities(t *testing.T)
func TestTickLatencyUnderMaxLoad(t *testing.T)
func TestSnapshotPayloadSizeRegression(t *testing.T)
```

### 3. Distributed System Resilience Tests

**Test network partitions and node failures:**
```go
// Network partition scenarios
func TestNetworkPartitionBetweenSimNodes(t *testing.T)
func TestGatewayIsolationRecovery(t *testing.T)
func TestSplitBrainPrevention(t *testing.T)

// Node failure scenarios
func TestSimNodeFailureDuringHandover(t *testing.T)
func TestGracefulNodeShutdownAndDrain(t *testing.T) 
func TestRapidNodeRecoveryAfterCrash(t *testing.T)
```

### 4. Load Balancing and Auto-Scaling Tests

**Test dynamic load distribution:**
```go
// Load balancing algorithms
func TestLoadBalancerCellDistribution(t *testing.T)
func TestHotspotDetectionAndMitigation(t *testing.T)
func TestElasticScalingUnderVariableLoad(t *testing.T)

// Resource utilization optimization
func TestCellConsolidationDuringLowLoad(t *testing.T)
func TestResourceThrottlingUnderStress(t *testing.T)
```

## Architecture Dependencies

### Missing Components Requiring ADR + Implementation

1. **Cross-Node Communication Protocol** (ADR needed)
   - Message format for node-to-node handovers
   - Authentication and authorization between sim nodes
   - Failure detection and retry mechanisms

2. **Distributed Cell Assignment System** (ADR needed)
   - Consistent hashing vs. centralized assignment
   - Load balancing algorithm (round-robin, least-loaded, geographic)
   - Cell migration triggers and procedures

3. **Gateway Enhancement for Multi-Node** (Implementation needed)
   - Node discovery and health checking  
   - Client routing to appropriate sim nodes
   - Connection tunneling vs. client reconnection for handovers

4. **Distributed Session State Management** (ADR needed)
   - Redis/shared state vs. node-to-node replication
   - Session failover and recovery mechanisms
   - Consistency guarantees and conflict resolution

## Success Criteria

### Performance Targets (Phase B)
- Cross-node handover latency: < 500ms (per TDD)
- Support 1000+ concurrent players across multiple nodes
- Graceful degradation under 150% target load
- Zero data loss during planned node shutdowns
- < 1% failure rate for cross-node handovers

### Test Coverage Targets
- 90%+ code coverage for cross-node logic
- Automated stress tests that can run in CI/CD
- Performance regression detection (alert on 20%+ degradation)
- Chaos engineering integration (random failure injection)

## Implementation Plan

### Phase 1: Foundation (ADRs + Basic Implementation)
1. Create ADR for cross-node handover protocol design
2. Create ADR for distributed cell assignment strategy  
3. Implement basic cross-node communication framework
4. Create test infrastructure for multi-node scenarios

### Phase 2: Core Cross-Node Logic
1. Implement cross-node handover state machine
2. Add gateway multi-node routing logic
3. Create distributed session state management
4. Build comprehensive unit tests for new components

### Phase 3: Integration and Stress Testing
1. Create end-to-end cross-node handover tests
2. Build load testing framework with synthetic players
3. Implement chaos testing for node failures
4. Add performance monitoring and alerting

### Phase 4: Advanced Scenarios
1. Create partition tolerance tests
2. Implement load balancing optimization tests
3. Add auto-scaling and elastic capacity tests
4. Build long-running stability and soak tests

## Acceptance Criteria

- [ ] Cross-node handover tests demonstrate < 500ms latency
- [ ] Load tests validate 1000+ concurrent player capacity  
- [ ] Chaos tests prove resilience to single node failures
- [ ] Performance tests detect regressions automatically
- [ ] All tests integrate into existing CI/CD pipeline (`make test-scaling`)
- [ ] Documentation covers scaling architecture and deployment

## Risk Mitigation

- Start with simplest cross-node protocol (HTTP-based)
- Use feature flags to enable/disable cross-node features
- Implement comprehensive monitoring before load testing
- Create rollback procedures for distributed deployments
- Design tests to be deterministic and reproducible

## Related Issues/ADRs

- **TDD Section**: "Handover (Phase B: Cross‑Node, Post‑MVP)" (lines 156-159)
- **Performance Budgets**: "< 500ms cross‑node" target (line 362)
- **Existing ADR**: `docs/process/adr/0001-local-sharding.md`
- **Relevant User Stories**: Need to identify multi-node user scenarios

---

**Priority**: High (foundational for post-MVP scaling)
**Effort**: 3-4 sprints (requires significant architecture work)
**Dependencies**: Cross-node protocol design, distributed state management