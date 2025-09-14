## Commit & Pull Request Guidelines
- **Commits:** use concise, imperative subjects (e.g., "sim: fix handover hysteresis"). Group related changes; keep diffs focused.
- **PRs:** include intent, summary of changes, and testing notes. Link relevant GitHub issues or project items. Add screenshots or CLI transcripts for behavior changes (e.g., `/dev/players` output).
- **Checks:** ensure `go fmt`, `go vet`, and tests pass; update docs (GDD/TDD) when behavior or APIs change.
- **Frequency:** make small, focused commits with green tests. Reference GitHub issue IDs in messages where relevant (e.g., `#123`).
- **Branching:** one feature branch per story/task. Use names like `feat/us-201-aoi-visibility`, `fix/handover-hysteresis`, or `docs/progress-snapshot`. Don’t push to `main` directly; open a Draft PR early to get CI.
- **Scope:** keep PRs narrowly scoped (aim <300 lines changed when practical). Include tests for new logic and any doc updates.
- **CI gate:** run `make fmt vet test` and `make test-ws` locally before pushing; all checks must pass in CI.
- **Docs updates:** when story status or behavior changes, update relevant design docs and ensure the corresponding GitHub issue or project item reflects the latest status; reference these updates in the PR description.
- **History:** prefer rebase/fast-forward onto `main` for a linear history.
- **Build tags:** guard WS-only code/tests with `//go:build ws`; include them with `make test-ws`.

## Security & Configuration Tips
