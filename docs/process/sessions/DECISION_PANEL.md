# Architecture Decision Panel (AI Team)

Purpose: Drive a clear choice on a high-impact change, capture it as an ADR, and seed follow-ups.

Duration: 60-90 minutes (time-boxed roles keep it brief).

Pre-read: `scripts/agents/prepare-context.sh > docs/process/sessions/_latest_context.md` and relevant design docs/benchmarks.

## Agenda
- 0-10m: PO frames goals, constraints, and success criteria.
- 10-30m: Architect outlines options and trade-offs; engineers give brief positions.
- 30-55m: Debate trade-offs; focus on consequences and reversibility.
- 55-70m: Decision and rationale; assign ADR author.
- 70-90m: Create follow-up issue checklist and owners.

## Outputs
- ADR in `docs/process/adr/` using `TEMPLATE.md`.
- Issue checklist linking to code, tests, and docs updates.
- Rollback/mitigation note if the change is risky.
