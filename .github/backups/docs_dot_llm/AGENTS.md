# LLM Agent Operating Manual

## Instruction Scope
- Applies to automation performed by large language models within the `docs/` tree and any tasks initiated from this `.llm/` directory.
- Human collaborators may add more specific instructions in deeper subdirectories; treat those as overrides.

## Onboarding Sequence
1. Read `docs/README.md` to understand documentation categories.
2. Follow the quick-start steps in `onboarding/quick-start.md` before making changes.
3. Review `onboarding/file-organization-guide.md` to understand file management best practices.
4. Complete the validation checklist in `onboarding/agent-validation-checklist.md` to ensure full repository access.
5. Use `onboarding/contribution-checklist.md` as the gating checklist prior to submitting work or opening PRs.
6. Reference `onboarding/story-template.md` when drafting roadmap stories, planning issues, or doc change logs.

## Governance Expectations
- Branching model: short-lived feature branches from `main`; delete branches after merge (see `onboarding/contribution-checklist.md`).
- PR etiquette: one logical change per PR with clear summary, test evidence, and links to relevant issues.
- Documentation hierarchy: update `docs/README.md` whenever new long-lived docs are created or reorganized.
- Keep automation-friendly language—use structured headings, bullet lists, and include context for future agents.

## Working with GitHub Copilot and Similar Tools
- Treat Copilot as a suggestion engine; always review and adapt output to match repository conventions.
- Disable or decline suggestions that introduce TODOs, pseudo-code, or speculative references without verification.
- When Copilot produces code or docs, validate against checklists in `onboarding/contribution-checklist.md` and add citations in PR summaries.
- Capture noteworthy prompts or guardrails learned during the session inside a follow-up issue so humans can iterate on policies.

## Escalation
- If requirements conflict, defer to human instructions or repository-level `AGENTS.md` files.
- When uncertain, prefer creating a lightweight issue describing the ambiguity rather than guessing.

Maintain this manual as the authoritative source for LLM operations. Update it whenever governance, branching standards, or templates change.
