# Roadmap Planning Update

_Use this template to log the outcome of a roadmap planning meeting. Duplicate it for each session so history remains auditable._

- **Meeting Date**: _YYYY-MM-DD_
- **Reference Issue/Notes**: _Link to meeting issue or agenda_
- **Release Theme**: _Codename or value statement_

## Summary
Provide a succinct summary of decisions, shifts in scope, and headline risks.

## Documents Updated
List every artifact touched during the session and what changed.

1. **`product/roadmap/roadmap.md`** — _e.g., updated timeline and risk register_
2. **`product/roadmap/implementation-playbook.md`** — _e.g., refreshed workstreams, instrumentation tasks_
3. **`process/sessions/PLANNING.md`** — _e.g., added facilitation notes_

Adjust or extend the list to match reality.

## Key Decisions
- **Milestones**: _Summarize chosen milestones or releases and their acceptance criteria._
- **Technical Priorities**: _Highlight architecture investments or debt paydown decisions._
- **Process Adjustments**: _Note any changes to rituals, tooling, or automation expectations._

## Risks & Mitigations
Capture new or evolving risks, along with mitigation owners and timelines.

| Risk | Mitigation | Owner | Due |
|------|------------|-------|-----|
| Example: Persistence reconnect budget | Profile reconnect flow, add soak test checkpoint | Backend | YYYY-MM-DD |

## Validation & Next Steps
- ✅ Tests required to confirm current state (e.g., `make fmt vet test test-ws`).
- ✅ Follow-up actions assigned (link to GitHub issues or project items).
- ✅ Communication plan (who needs to hear about the update and when).

## Change Log Classification
- **Change Type**: _Documentation / Process / Code_
- **Impact**: _Planning, Execution, Release, etc._
- **Testing Performed**: _List commands or N/A_

Maintain these updates in chronological order to build a reliable history of roadmap decisions.
