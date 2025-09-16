# Agent Roles & Responsibilities

This repository coordinates specialized agents to simulate a small cross‑functional team. Each role contributes predictable outputs to keep the system aligned and auditable.

- Product Owner (PO): Scope, acceptance criteria, user value, risks, and KPI impact. Owns roadmap and milestone clarity.
- Lead Architect: System boundaries, trade‑off analysis, and ADR authorship. Guards consistency and evolution path.
- Gameplay/Sim Engineer: Tick loop, AOI/handovers, progression logic, data shapes. Provides test hooks and deterministic sims.
- Networking/Gateway Engineer: Protocol contracts, reliability/ordering/backpressure, upgrade/versioning plan. Ensures compat and perf budgets.
- SRE/QA: SLOs/SLIs, test strategy (unit/integration/load/soak), deploy/rollback, observability and incident hygiene.

Outputs per Session Type
- Standup: 3 bullets per role — Yesterday, Today, Blockers.
- Planning: Epic/Story list with DoR/DoD, dependencies, risk notes, estimates.
- Decision Panel: ADR draft or update; follow‑up issue checklist.
- Review + Retro: Metrics snapshot, demo links, “start/stop/continue” and owner for improvements.

Reference
- Session templates live under `docs/process/sessions/`.
- ADRs live under `docs/process/adr/` using `TEMPLATE.md`.
