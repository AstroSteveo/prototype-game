# Repository Guidelines

## Instruction Scope
- This file defines default instructions for the repository.
- Subdirectories may provide their own `AGENTS.md` to override or extend these rules.
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
