# Contributing

This repository uses a lightweight, test-first workflow. Even for solo work with an AI agent, these guardrails keep quality high and context tight.

Key docs
- Workflow and conventions: see [AGENTS.md](../AGENTS.md)
- Daily dev commands: see [DEV guide](../docs/dev/DEV.md)
- Backlog and status: see [BACKLOG](../docs/process/BACKLOG.md) and [PROGRESS](../docs/process/PROGRESS.md)
- Makefile helpers: see [Makefile](../Makefile) or run `make help`

Branches
- Create one feature branch per story/task. Suggested names: `feat/us-201-aoi-visibility`, `fix/handover-hysteresis`, `docs/progress-snapshot`.
- Do not push directly to `main`. Open a Draft PR early to run CI and collect feedback.

Commits
- Prefer small, focused commits with green tests.
- Use imperative subjects and an area prefix: `sim: tighten AOI boundary`.
- Reference story IDs when applicable (e.g., `US-201`).

Tests and CI gate
- Run locally before pushing: `make fmt vet test` and `make test-ws` (includes WebSocket-tagged tests).
- Guard WS-only code/tests with `//go:build ws`.

Docs and traceability
- When behavior or status changes, update the backlog and progress docs:
  - [docs/BACKLOG.md](../docs/BACKLOG.md) and [docs/PROGRESS.md](../docs/PROGRESS.md)
- Keep PRs concise: describe intent, summarize changes, note tests added/updated, and link story IDs.

House rules
- Prefer rebase/fast-forward to keep a linear history on `main`.
- Do not commit secrets, logs, or build artifacts. Use `.gitignore` appropriately.
