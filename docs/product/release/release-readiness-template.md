# Release Readiness Analysis Template

Use this template to evaluate whether a release candidate is ready for launch. Update the callouts, metrics, and risks with current data before sharing with stakeholders.

## Release Snapshot
- **Analysis Date**: _YYYY-MM-DD_
- **Analyzed By**: _Name / Role_
- **Target Release**: _Release theme or codename_
- **Overall Status**: 🟢 READY / 🟡 READY WITH CONDITIONS / 🔴 NOT READY (choose one and justify below)

## Executive Summary
Provide a concise summary of the project’s readiness. Highlight blockers, notable achievements, and the recommendation (proceed, proceed with conditions, or delay).

### Decision Matrix
| Criteria | Status | Notes |
|----------|--------|-------|
| Foundation Complete | ⬜️ | Core milestones implemented and tested |
| Technical Health | ⬜️ | Build/test results, performance budgets |
| Team Readiness | ⬜️ | Operational coverage, on-call rotations |
| Timeline Realism | ⬜️ | Buffers for late-breaking work |
| Risk Management | ⬜️ | Mitigations in place with owners |

_Swap the placeholders with ✅/🟡/❌ (or similar) and expand in the Notes column._

## Milestone Status
Document current completion state with evidence.

| Milestone | Status | Evidence | Performance Highlights |
|-----------|--------|----------|------------------------|
| Example: Inventory & Equipment | ✅ Complete | Demo video link, PR references | Encumbrance enforcement, equip cooldown metrics |
| Example: Targeting & Skills | 🟡 In Progress | Integration tests pending | AOI targeting verified, XP telemetry captured |
| Example: Persistence | 🔴 Not Started | Schema review scheduled | Pending reconnect benchmark |

Add or remove rows to mirror the active roadmap.

## Technical Health
### Build & Test Results
Summarize the latest CI runs, including command outputs (`make fmt vet test test-ws`) and any race or integration suites.

### Performance Validation
Record the most recent measurements (tick rate, reconnect latency, snapshot size, etc.) and compare them to targets defined in the [Technical Design Document](../../architecture/technical-design-document.md).

### Operational Readiness
Confirm service start/stop reliability, deployment automation, logging, metrics, and alerting coverage. Link to runbooks (e.g., [`operations/branch-maintenance-runbook.md`](../../operations/branch-maintenance-runbook.md)) as proof of preparedness.

## Documentation & Process
- **Design Docs**: [Game Design Document](../../product/vision/game-design-document.md) / [Technical Design Document](../../architecture/technical-design-document.md) updated?
- **Developer Workflows**: [Developer Guide](../../development/developer-guide.md) reflects new commands?
- **Roadmap**: [Roadmap handbook](../roadmap/roadmap.md) refreshed within the last two business days?
- **Release Notes**: Drafted and approved by stakeholders?

## Risks & Mitigations
List active release risks with owners and mitigation plans.

| Risk | Probability | Impact | Mitigation | Owner | Status |
|------|-------------|--------|------------|-------|--------|
| Example: Reconnect latency exceeds 2s | Medium | High | Run soak test with persistence flag, profile hotspots | Backend | In progress |

## Recommendation
State the recommended decision (GO, GO WITH CONDITIONS, or NO-GO) and enumerate the conditions if applicable. Include a brief action plan with deadlines for each outstanding task.

## Sign-off
Capture approvals from required stakeholders (Product, Engineering, QA/SRE, Leadership). When conditions are present, document follow-up verification steps.

Maintain this document as a rolling template—duplicate it for each release assessment so historical reports remain accessible.
