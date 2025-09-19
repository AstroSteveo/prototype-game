# Copilot & AI Assistant Playbook

Large language models and GitHub Copilot can accelerate editing, but automation must apply them responsibly.

## When to Use Copilot
- Drafting boilerplate code or repetitive Markdown sections.
- Generating test scaffolding that mirrors existing patterns.
- Summarizing logs or code snippets for investigation notes.

## Guardrails
- Review every suggestion for accuracy and repository compliance before accepting.
- Reject suggestions that fabricate APIs, commands, or references.
- Never accept placeholder text (e.g., `TODO`, `lorem ipsum`)—replace with concrete guidance.
- Keep context windows focused by selecting only relevant files; avoid leaking unrelated secrets or data.

## Collaboration Workflow
1. Craft a clear prompt summarizing the intent and referencing required docs (TDD, roadmap, etc.).
2. Evaluate Copilot output against the acceptance criteria and update as needed.
3. Run tests or lint commands even if Copilot indicates success.
4. Capture useful prompts or pitfalls discovered during the session in the PR description or follow-up issues.

## Troubleshooting Suggestions
- If Copilot repeatedly suggests incorrect patterns, narrow the context (select specific files) or temporarily disable it.
- Prefer manual editing when updating governance docs or policies to ensure deliberate wording.
- Validate generated code for concurrency, error handling, and performance characteristics; Copilot may miss edge cases.

Automation should treat Copilot as an assistant, not an authority. Human review remains the final safety net.
