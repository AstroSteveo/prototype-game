# Repository Guidelines

## Instruction Scope
- This file defines default instructions for the repository.
- Subdirectories may provide their own `AGENTS.md` to override or extend these rules.
- **GitHub Copilot Instructions**: `.github/copilot-instructions.md` is the single source of truth for GitHub Copilot configuration. Do not create duplicate instruction files.
- Current subtree guides:
  - `backend/AGENTS.md` for Go services and libraries.
  - `docs/AGENTS.md` for reference documentation.

## Project Structure
- `backend/` – server code and libraries.
- `client/` – game client placeholder.
- `docs/` – design, developer, and process docs.

## Commit & Pull Request Guidelines
- Use concise, imperative commit messages.
- Run `make fmt vet test` and `make test-ws` before committing code changes.
- PRs include intent, summary, testing notes, and links to relevant issues.

## Security & Configuration
- Gateway listens on `:8080`; sim on `:8081`.
- Do not commit secrets, logs, or build artifacts.

## Agent Rituals
- Daily async standup: Use `docs/process/sessions/STANDUP.md` as the template. Each role (PO, Architect, Gameplay, Networking, SRE/QA) posts a 3-bullet update: yesterday/today/blockers. Prefer a single GitHub issue created via `.github/ISSUE_TEMPLATE/standup.yml`.
- Weekly planning: Use `docs/process/sessions/PLANNING.md` to confirm scope, break down stories, and seed issues. Output is a short milestone plan and linked story/tasks.
- Roadmap planning: Use `docs/process/sessions/ROADMAP.md` for quarterly or pre-release roadmap sessions. Define release goals, prioritize features, and update the project roadmap. Create using `.github/ISSUE_TEMPLATE/roadmap.yml`.
- Decision panels: For high-impact changes, run the 60–90m decision panel using `docs/process/sessions/DECISION_PANEL.md`. Capture the outcome as an ADR in `docs/process/adr/` using the ADR template.
- Review + Retro: Close each sprint with `docs/process/sessions/REVIEW_RETRO.md`. Record wins, misses, metrics, and improvements; open follow-up issues.

Utilities
- Context prep: `scripts/agents/prepare-context.sh` emits a current repo snapshot for session pre-reads. Redirect output to a file (e.g., `docs/process/sessions/_latest_context.md`).
- Agent validation: `scripts/agents/validate-onboarding.sh` verifies that an agent can access all necessary repository resources and that the onboarding framework is working correctly.
