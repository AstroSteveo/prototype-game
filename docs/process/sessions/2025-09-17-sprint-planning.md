# Sprint Planning — 2025-09-17

Game Roadmap Project #2 will run this sprint to land the persistence + inventory foundation alongside the combat targeting vertical slice.

## Sprint Backlog Overview
- **Iteration window**: 2025-09-17 → 2025-09-30 (two-week cadence).
- **Committed stories**: 10
- **Total estimate**: 52 pts
- **Themes**: Persistence reliability, inventory/equipment UX, combat targeting & skills telemetry.

## Story Bundle

### Inventory & Persistence (23 pts)
- #117 `US-003 — Establish PostgreSQL persistence foundation` (5 pts)
- #118 `US-004 — Implement inventory compartments and encumbrance in the sim` (8 pts)
- #119 `US-005 — Implement equip flow with cooldown and skill gating` (5 pts)
- #120 `US-006 — Persist inventory and equipment through reconnect` (5 pts)

### Targeting & Skills Vertical Slice (24 pts)
- #121 `US-007 — Implement server-side target acquisition & validation` (8 pts)
- #122 `US-008 — Build client feedback loop for targeting & skill use` (5 pts)
- #123 `US-009 — Deliver skill XP/level-up progression pipeline` (8 pts)
- #124 `US-010 — Telemetry & observability for combat interactions` (3 pts)

### Supporting & Hardening (5 pts)
- #125 `CH-011 — Harden auth/session edges for combat reconnect` (3 pts)
- #126 `CH-012 — Sandbox data seeding for combat QA scenarios` (2 pts)

## Tracking Notes
- Issues #117–#120 already sit in Project #2 (column: Todo) with the `story` label and milestone M5 applied.
- Queue #121–#126 into Project #2 with the same labels/milestone so the sprint backlog is fully represented.
- Set iteration assignments and owners once the team confirms availability; automation depends on the `Sprint` field being populated.
- Audit dependencies: confirm Postgres credentials rotation, telemetry quotas, and QA sandbox access before kickoff.

## Follow-Ups
- Review automation signals during stand-up to make sure new issues sync correctly.
- Flag outstanding blockers (e.g., DB provisioning, auth token updates) in daily stand-up threads until resolved.
- Schedule a combat telemetry dashboard review mid-sprint to validate metric coverage.
