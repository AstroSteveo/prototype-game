# Documentation Overview

This directory captures the enduring knowledge for the prototype game project. The files are grouped by purpose so contributors and automation can find living guidance quickly.

**Last Updated**: 2025-09-19  
**Total Documentation Files**: 50+

## Quick Navigation

| Section | Purpose | Key Files |
|---------|---------|-----------|
| [Product Direction](#product-direction) | Vision, roadmap, releases | roadmap.md, game-design-document.md |
| [Architecture](#architecture--quality) | System design, technical decisions | technical-design-document.md, ADRs |
| [Development](#development-workflow) | Developer guides, testing | developer-guide.md |
| [Analysis](#project-analysis) | Comprehensive project analysis | Discovery of advanced capabilities |
| [Operations](#operations--governance) | Project management, automation | Board automation, branch maintenance |
| [Process](#process--rituals) | Templates, roles, sessions | Feature proposals, meeting guides |

## Product Direction
- `product/vision/` — player and experience goals, including the Game Design Document and design refresh template.
- `product/roadmap/` — roadmap source of truth, meeting guide, update template, and implementation playbook for upcoming releases.
- `product/release/` — release readiness analysis template used to judge launches.

## Architecture & Quality
- `architecture/technical-design-document.md` — system-level plan for the simulation stack.
- `architecture/design.md` — architectural design documentation.
- `architecture/requirements.md` — system requirements and specifications.
- `architecture/testing/` — deep dives for advanced load and sharding validation.
- `architecture/health/` — reusable templates for technical health reviews.
- `process/adr/` — architectural decision records with the canonical template, peer review guidelines, and recorded ADRs.

## Development Workflow
- `development/developer-guide.md` — comprehensive developer setup and workflow guide.
- `development/server-feature-test-plan.md` — testing strategy and coverage mapping.
- `development/persistence-verification-guide.md` — database and persistence testing guide.
- `development/tasks.md` — development task tracking.

## Project Analysis
**🚨 CRITICAL DISCOVERY**: Comprehensive analysis revealing sophisticated backend capabilities already implemented.
- `analysis/README.md` — analysis overview and key discoveries.
- `analysis/VALIDATION_RESULTS.md` — detailed findings on implemented systems.
- `analysis/REFLECTION_AND_RECOMMENDATIONS.md` — strategic recommendations based on discoveries.
- `analysis/HANDOFF_PACKAGE.md` — complete handoff documentation for stakeholders.

## Operations & Governance
- `operations/project-board-automation.md` and `operations/project-sync.md` — how issues and pull requests flow into GitHub Projects.
- `operations/branch-maintenance-runbook.md` — guidelines for pruning stale branches and keeping `main` authoritative.
- `governance/llm-adoption-whitepaper.md` — background research on large-language-model usage for the project.

## Process & Rituals
- `process/FEATURE_PROPOSAL.md` — template and guardrails for feature pitches.
- `process/roles.md` — responsibilities for each simulated team role.
- `process/sessions/` — facilitation guides for standups, planning, roadmap, decision panels, and retrospectives.
- `process/traceability/` — verification and requirements traceability matrix.

## AI/LLM Agent Guidance
- **Primary Instructions**: [.github/instructions/](.github/instructions/) — specification-driven workflow and task implementation guides.
- **Scripts**: `scripts/agents/` — automation validation and onboarding scripts.
- **Governance**: `governance/llm-adoption-whitepaper.md` — strategic approach to AI adoption.

## Documentation Maintenance

**When updating documentation:**
1. Update the date stamp at the top of this file
2. Keep section descriptions current with actual file contents
3. Add new major documents to the appropriate section
4. Validate internal links after structural changes
5. Ensure consistency between README.md, docs/README.md, and roadmap documentation

**Documentation Standards:**
- Use relative links for internal references
- Include brief descriptions for all major documents
- Maintain the table of contents for easy navigation
- Cross-reference related documents where appropriate

Use this file as the entry point before diving into specific guides. Update it whenever new long-lived documents are added.
