---
goal: Fix overlapping collections toggle via effective-state precedence and performant apply
version: 1.0
date_created: 2025-09-21
last_updated: 2025-09-21
owner: awesome-copilot maintainers
status: 'Planned'
tags: ['feature','architecture','bug','cli','performance']
---

# Introduction

![Status: Planned](https://img.shields.io/badge/status-Planned-blue)

This plan implements correct, non-destructive collection toggling using an "effective-state" computation that respects explicit per-item overrides and overlapping collections, improves apply-time performance with Set-based lookups, and preserves idempotent behavior and UX.

## 1. Requirements & Constraints

- **REQ-001**: Shared items remain enabled if any enabled collection includes them (no accidental disable when another collection is still on).
- **REQ-002**: Explicit per-item boolean overrides take precedence over collections (true => enabled, false => disabled). Undefined => inherit from collections.
- **REQ-003**: Toggling a collection must not mass-edit per-item flags; only the collection's `enabled` field changes.
- **REQ-004**: The CLI must surface “effective state” and the reason (explicit vs via collections) in `list` output.
- **REQ-005**: `apply` operates on effective states, not raw booleans only, and remains idempotent.
- **REQ-006**: Improve performance by replacing repeated linear scans with precomputed Sets/Maps for O(1) lookups.
- **REQ-007**: Handle `undefined` item settings correctly (never treat as explicitly disabled).
- **SEC-001**: No secrets added; do not introduce insecure hashing or file writes outside the project.
- **CON-001**: Maintain backward-compatible CLI options and config file structure.
- **CON-002**: Avoid adding dependencies unless necessary; if a stable config hash is required, use sorted key stringify or Node `crypto` with deterministic input.
- **GUD-001**: Keep code changes minimal and localized; avoid broad refactors not tied to this feature.
- **PAT-001**: Single source of truth for effective-state via a dedicated helper in `config-manager.js`.

## 2. Implementation Steps

### Implementation Phase 1

- GOAL-001: Add effective-state computation and stop clobbering per-item flags during collection toggles

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-001 | In `instructions/README.md` (or project README if applicable), document precedence rules: explicit override > collections; undefined inherits. |  |  |
| TASK-002 | In `config-manager.js`, add `computeEffectiveItemStates(config)` to build membership maps per section and return, for each section, a set of effectively enabled items and a reasons map (source is explicit or collections, plus an optional list of collections via which it is enabled). |  |  |
| TASK-003 | In `config-manager.js`, update `toggleCollection(name, enabled)` to only flip the collection's `enabled` flag and return a summary; remove any unconditional per-item writes (e.g., `configCopy.prompts[prompt] = enabled`). |  |  |
| TASK-004 | Ensure `config-manager.js` collection helpers treat `undefined` item flags as "no explicit override". |  |  |

### Implementation Phase 2

- GOAL-002: Apply uses effective-state with O(1) lookups; fix undefined checks and improve hashing stability

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-005 | In `apply-config.js`, precompute Sets of effectively enabled basenames/IDs per section from `computeEffectiveItemStates(config)`; replace linear helper (e.g., `isItemInEnabledCollection`) with Set lookups. |  |  |
| TASK-006 | Update any "explicitly disabled" checks to use strict comparison `=== false` and not treat `undefined` as disabled. |  |  |
| TASK-007 | Stabilize `configHash` by using a sorted-key JSON stringify utility (local function) before hashing or base64 encoding; keep the rest of the state file structure unchanged. |  |  |

### Implementation Phase 3

- GOAL-003: CLI surfaces effective state and concise deltas after toggles; keep auto-apply behavior consistent

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-008 | In `awesome-copilot.js` (or the CLI entry that handles `list`), show for each item: effective state and reason (`explicit:true`, `explicit:false`, or `via:[collectionA,...]`). |  |  |
| TASK-009 | In `toggle collection` command, after flipping the `enabled` flag, recompute effective states and print a delta summary: counts of newly-enabled, newly-disabled, and items blocked by explicit overrides. |  |  |
| TASK-010 | Ensure any auto-apply step uses effective states (no functional change besides the new source of truth). |  |  |

### Implementation Phase 4

- GOAL-004: Tests for correctness, idempotency, and performance behavior

| Task | Description | Completed | Date |
|------|-------------|-----------|------|
| TASK-011 | Unit tests in `scripts/` or test folder for `computeEffectiveItemStates`: overlapping collections, explicit:true/false, undefined cases. |  |  |
| TASK-012 | Integration tests to run `toggle` + `apply` on a sample config with overlapping collections; verify shared items are not removed when still required or explicitly enabled; rerun `apply` for idempotency. |  |  |
| TASK-013 | CLI tests for `list` and `toggle` summaries reflecting reasons and counts. |  |  |

## 3. Alternatives

- **ALT-001**: Continue mass-writing per-item flags when a collection toggles. Rejected: destructive, causes conflicts for shared items, violates REQ-001/003.
- **ALT-002**: Remove explicit per-item overrides entirely. Rejected: breaks backward compatibility and removes user control; violates CON-001.
- **ALT-003**: Introduce a heavy dependency for stable JSON hashing. Rejected: local sorted-stringify is sufficient and avoids new deps; aligns with CON-002.

## 4. Dependencies

- **DEP-001**: Node.js runtime used by this repo.
- **DEP-002**: None new required; implement local stable stringify. If later needed, consider `json-stable-stringify` as an optional enhancement.

## 5. Files

- **FILE-001**: `config-manager.js` — add `computeEffectiveItemStates`, update `toggleCollection`, remove per-item clobbering.
- **FILE-002**: `apply-config.js` — consume effective states, replace linear scans with Sets, fix undefined checks, stable `configHash` input.
- **FILE-003**: `awesome-copilot.js` (CLI) — update `list`/`toggle` output to show effective reasons and deltas.
- **FILE-004**: `README.md` and/or `README.instructions.md` — document precedence rules and behavior.
- **FILE-005**: Tests under existing test harness (e.g., `scripts/test-functionality.js` or similar) for unit/integration coverage.

## 6. Testing

- **TEST-001**: Unit — Item in two collections: disable one, item remains enabled via the other (no explicit flags).
- **TEST-002**: Unit — Explicit true keeps item enabled when all collections off; explicit false keeps item disabled when collections on.
- **TEST-003**: Unit — Undefined + enabled collection => enabled; undefined + all collections off => disabled.
- **TEST-004**: Integration — `toggle` then `apply` on sample config; verify copied/skipped/removed counts align with effective Sets; second `apply` is a no-op.
- **TEST-005**: CLI — `list` shows effective state and reasons; `toggle` prints concise deltas and notes explicit override conflicts.
- **TEST-006**: Performance — assert no linear per-file scans in hot path; confirm Set membership checks are used.

## 7. Risks & Assumptions

- **RISK-001**: Edge cases where items are referenced by collections but do not exist. Mitigation: warn and continue.
- **RISK-002**: Misalignment between basenames and physical paths. Mitigation: standardize on consistent keying (prefer canonical item IDs or basenames consistently across config and FS).
- **RISK-003**: CLI output changes could confuse users. Mitigation: keep messages concise and add a short README note.
- **ASSUMPTION-001**: Current CLI commands (`list`, `toggle`, `apply`) are the integration points and can be updated without breaking consumers.
- **ASSUMPTION-002**: Tests can run in the existing Node environment without new dependencies.

## 8. Related Specifications / Further Reading

- PR Review Context: Overlapping collections and explicit override handling (internal reference to PR #8).
- Internal docs: `README.instructions.md`, `README.chatmodes.md`.
