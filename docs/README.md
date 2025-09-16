# Documentation Overview

This directory captures the enduring knowledge for the prototype game project. The files are grouped by purpose so contributors and automation can find living guidance quickly.

## Product Direction
- `product/vision/` — player and experience goals, including the Game Design Document and design refresh template.
- `product/roadmap/` — roadmap source of truth, meeting guide, update template, and implementation playbook for upcoming releases.
- `product/release/` — release readiness analysis template used to judge launches.

## Architecture & Quality
- `architecture/technical-design-document.md` — system-level plan for the simulation stack.
- `architecture/testing/` — deep dives for advanced load and sharding validation.
- `architecture/health/` — reusable templates for technical health reviews.
- `process/adr/` — architectural decision records with the canonical template alongside recorded ADRs.

## Development Workflow
- `development/` — developer guide, server feature coverage map, and other day-to-day engineering references.
- `governance/llm-adoption-whitepaper.md` — background research on large-language-model usage for the project.

## Operations & Governance
- `operations/project-board-automation.md` and `operations/project-sync.md` — how issues and pull requests flow into GitHub Projects.
- `operations/branch-maintenance-runbook.md` — guidelines for pruning stale branches and keeping `main` authoritative.

## Process & Rituals
- `process/FEATURE_PROPOSAL.md` — template and guardrails for feature pitches.
- `process/roles.md` — responsibilities for each simulated team role.
- `process/sessions/` — facilitation guides for standups, planning, roadmap, decision panels, and retrospectives.

## LLM Agent Guidance
- `.llm/AGENTS.md` — onboarding, governance, branching standards, and authoring templates tailored for automation.

Use this file as the entry point before diving into specific guides. Update it whenever new long-lived documents are added.
