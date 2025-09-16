# Documentation Guidelines

## Instruction Scope
- Applies to the entire `docs/` tree unless a subdirectory defines stricter rules.

## Structure Expectations
- Keep the category layout described in `docs/README.md` up to date when adding or moving documents.
- `product/` holds roadmap, vision, and release guidance that should age gracefully.
- `architecture/` contains the technical design, ADRs, health reviews, and specialized test plans.
- `development/` captures day-to-day engineering workflows and coverage references.
- `operations/` documents project automation, maintenance runbooks, and governance mechanics.
- `.llm/` houses automation onboarding instructions—treat it as the single source of truth for agent behavior.

## Authoring Guidelines
- Use Markdown with clear headings and scannable lists.
- Cross-link related documents when it improves navigation.
- Feature proposals must follow `process/FEATURE_PROPOSAL.md` and ADRs belong in `process/adr/` using the template provided there.
- Prefer evergreen language over time-bound milestones; if specific dates are required, clearly mark them for future review.
- Doc-only changes do not require running Go builds or tests.
