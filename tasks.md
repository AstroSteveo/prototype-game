# tasks.md

Purpose: Track implementation tasks, owners, estimates, dependencies, and acceptance criteria.

## Overview
- Generated: 2025-09-19 (DESIGN pass)
- Linked requirements: See `requirements.md` (R1–R23)

## Task Template
- ID: T-###
- Title: Short title
- Description: Full description
- Related Requirements: R#,...
- Dependencies: Task IDs or external deps
- Estimate: 4h / 1d / 3d
- Priority: High/Medium/Low
- Assignee: Unassigned
- Acceptance Criteria: Clear, testable criteria
- Status: TODO / In Progress / Done

## Phase 1 — Documentation & Baseline
1. T-001 — Finalize requirements and design
   - Description: Validate `requirements.md` and complete `design.md` with interfaces and error matrix.
   - Related Requirements: all
   - Estimate: 1d
   - Priority: High
   - Acceptance Criteria: `requirements.md` and `design.md` reviewed; no TBD sections.

2. T-002 — Tasks plan completion
   - Description: Populate this `tasks.md` with all required tasks mapped to R1–R23.
   - Related Requirements: all
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: Each requirement mapped to at least one task.

## Phase 2 — WS Session, Security & Telemetry
3. T-010 — WS origin policy & dev mode
   - Description: Verify `WSOptions` origin patterns enforcement in prod; relaxed in dev.
   - Related Requirements: R6
   - Dependencies: T-001
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: Integration tests pass for dev/prod modes.

4. T-011 — Resume token validation
   - Description: Validate resume token path; ignore invalid tokens.
   - Related Requirements: R4
   - Dependencies: T-001
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: Reconnect scenario passes and resumes ack.

5. T-012 — Idle timeout and ping behavior
   - Description: Ensure idle disconnect and ping/pong telemetry.
   - Related Requirements: R3, R22
   - Estimate: 0.5d
   - Priority: Medium
   - Acceptance Criteria: WS tests verify timeout and RTT reporting.

## Phase 3 — Movement, AOI, Handovers
6. T-020 — Input clamping and velocity
   - Description: Validate clamp [-1,1], speed scaling, dt handling.
   - Related Requirements: R7
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: Movement tests pass; no overspeed.

7. T-021 — AOI neighborhood and epsilon
   - Description: Verify 3x3 cell AOI and epsilon tolerance.
   - Related Requirements: R9–R10
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: AOI tests pass including boundary cases.

8. T-022 — Handover hysteresis and anti-thrash
   - Description: Validate hysteresis H and doubled hysteresis when returning.
   - Related Requirements: R11–R12
   - Estimate: 0.5d
   - Priority: High
   - Acceptance Criteria: Handover latency measured; event emitted once per change.

## Phase 4 — Inventory, Equipment, Skills
9. T-030 — Equip validation matrix
   - Description: Tests for slot compatibility, skill gate, cooldown; error codes.
   - Related Requirements: R14–R15
   - Estimate: 1d
   - Priority: High
   - Acceptance Criteria: Integration tests assert error codes and success.

10. T-031 — Inventory/equipment deltas & versions
    - Description: Ensure versioned deltas included in state messages.
    - Related Requirements: R16
    - Estimate: 0.5d
    - Priority: Medium
    - Acceptance Criteria: State contains deltas when versions change.

11. T-032 — Encumbrance including equipped items
    - Description: Validate encumbrance computation and movement penalty.
    - Related Requirements: R17
    - Estimate: 0.5d
    - Priority: Medium
    - Acceptance Criteria: Encumbrance math matches expected scenarios.

## Phase 5 — Bots & Density
12. T-040 — Bot density maintenance bounds
    - Description: Keep actors per cell within ±20% target; respect MaxBots; ramp behavior.
    - Related Requirements: R13
    - Estimate: 1d
    - Priority: Medium
    - Acceptance Criteria: Density tests meet thresholds and caps.

## Phase 6 — Persistence
13. T-050 — Disconnect save and checkpoint
    - Description: Timeout-bounded save on disconnect; checkpoint requests.
    - Related Requirements: R18–R19
    - Estimate: 1d
    - Priority: Medium
    - Acceptance Criteria: Integration tests verify writes; timeouts handled.

14. T-051 — Restore from persisted state
    - Description: Deserialize to player from store using templates.
    - Related Requirements: R20
    - Estimate: 0.5d
    - Priority: Medium
    - Acceptance Criteria: Round-trip save/restore preserves inventory/equipment/skills.

## Phase 7 — Metrics
15. T-060 — Metrics coverage
    - Description: Ensure counters/histograms updated for ticks, AOI, equip, handovers, snapshots, WS connections.
    - Related Requirements: R21
    - Estimate: 0.5d
    - Priority: Low
    - Acceptance Criteria: Metrics observed in unit/integration tests.

## Phase 8 — Hardening & Docs
16. T-070 — Error matrix validation
    - Description: Verify error handling paths match matrix; add docs.
    - Related Requirements: multiple
    - Estimate: 0.5d
    - Priority: Low
    - Acceptance Criteria: Design error matrix aligns with behavior.

17. T-071 — Update PR template + workflows (optional)
    - Description: Ensure PRs link to requirements/design/tasks and validation artifacts.
    - Estimate: 0.5d
    - Priority: Low
    - Acceptance Criteria: Template present and used in PRs.

<!-- sync: trigger Phase 2 issues to Project 2 on 2025-09-19 -->

