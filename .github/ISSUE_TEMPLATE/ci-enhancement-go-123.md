---
name: CI: Enhance Go 1.23 pipeline (race, cache, matrix)
about: Improve CI to better validate Go 1.23 builds and tests
labels: ci, enhancement
---

## Summary
Enhance the backend CI workflow for Go 1.23 to increase coverage and reliability.

## Scope
- Add race detector step: `go test -race ./...`
- Add OS matrix: `ubuntu-latest`, `macos-latest`
- Ensure module caching and tidy check (fail if `go.mod`/`go.sum` change)
- Optional: add `staticcheck`, upload test artifacts (if applicable)

## Acceptance Criteria
- Workflow runs on PRs and pushes to `main`
- All steps pass on both OSes
- No regressions in ws-tag tests (`go test -tags ws ./...`)

## References
- Current workflow: `.github/workflows/ci.yml`
- Backend directory: `backend/`

