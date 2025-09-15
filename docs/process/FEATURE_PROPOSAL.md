# Feature Proposal Workflow

When a new feature or scope adjustment is discovered, follow this lightweight process to bring it into the roadmap and keep documentation aligned.

## 1. Discovery
- Capture the idea as a GitHub issue using the appropriate template (`Story` or `Task`).
- Provide a short problem statement and acceptance criteria.
- Add the issue to the [Game Roadmap project](https://github.com/users/AstroSteveo/projects/2).

## 2. Design
- For minor tweaks, reference existing design docs:
  - [Game Design Document](../design/GDD.md)
  - [Technical Design Document](../design/TDD.md)
- For **major architectural changes**, create an Architecture Decision Record (ADR) using [`docs/process/adr/TEMPLATE.md`](adr/TEMPLATE.md).
- Update relevant design docs with new details or rationale.

## 3. Task Breakdown
- Decompose the issue into sub-issues or add a checklist covering implementation, tests, docs, and deployment updates.
- Link related issues to track progress.

## 4. Implementation
- Follow the steps in [`docs/dev/DEV.md`](../dev/DEV.md) (fmt, vet, tests).
- Reference the issue ID in commit messages and PR descriptions.

## 5. Release & Follow-up
- After merge, update the project board and any release notes.
- Close the issue or move it to the appropriate column.
- Revisit the ADR if the decision evolves.

## Related Resources
- Issue templates: [`.github/ISSUE_TEMPLATE`](../../.github/ISSUE_TEMPLATE)
- Project board: [Game Roadmap](https://github.com/users/AstroSteveo/projects/2)
- Developer guide: [`docs/dev/DEV.md`](../dev/DEV.md)
- ADRs: [`docs/process/adr/`](adr/)

