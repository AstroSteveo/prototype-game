# Roadmap Meeting Guide

Use this guide to facilitate roadmap planning sessions and keep the project aligned on upcoming releases. The intent is to focus on reusable cadence rather than one-off calendar dates.

## When to Hold Roadmap Meetings
Schedule a roadmap session when:
- A new quarter or major release window is starting.
- Stakeholders need to re-evaluate priorities after major discoveries or delivery risks.
- The team completes a milestone and must confirm the next scope.
- External signals (player feedback, market changes) require course correction.

## Scheduling Checklist
1. **Create an issue** using the "Roadmap Planning Meeting" template under `.github/ISSUE_TEMPLATE/roadmap.yml`. Title format: `roadmap: <theme> - YYYY-MM-DD`.
2. **Invite participants** with calendar holds and attach the issue link for asynchronous context gathering.
3. **Pre-read packet**: compile metric snapshots, roadmap deltas, and open risks so attendees arrive informed.

## Preparation Materials
- Review the facilitation prompts in [`process/sessions/ROADMAP.md`](../../process/sessions/ROADMAP.md).
- Bring the latest [Project Roadmap Handbook](roadmap.md) and mark sections needing updates.
- Capture metrics from the [implementation playbook](implementation-playbook.md) or active issues that prove progress.
- Collect user research, playtest feedback, and technical health signals to inform prioritization.

## Meeting Flow (60–75 minutes)
1. **Current State (15 min)** — Highlight progress, blockers, and relevant metrics.
2. **Vision Refresh (10 min)** — Reconfirm the release theme and desired player outcomes.
3. **Options & Trade-offs (20 min)** — Review candidate initiatives, capacity, and dependencies.
4. **Commitments (15 min)** — Lock the milestone list, timelines, and success metrics.
5. **Actions & Owners (5–10 min)** — Assign follow-ups, document risks, and schedule checkpoints.

## Post-Meeting Actions
- Update the roadmap issue with decisions, metrics, and risk adjustments.
- Refresh the [roadmap handbook](roadmap.md) and [update log template](update-template.md) with the new plan.
- Ensure downstream artifacts (stories, tasks, ADRs) reflect the chosen scope.
- Communicate the outcomes asynchronously to stakeholders and automation agents.

## Participants
- **Product Owner** — Drives scope and success metrics, records decisions.
- **Technical Architect** — Evaluates feasibility, architecture impact, and integration risks.
- **Engineering Lead(s)** — Provide delivery estimates, sequencing, and capacity signals.
- **SRE/QA** — Represent quality gates, reliability budgets, and release readiness.
- **Optional** — Domain experts, analytics partners, or player-research representatives as needed.

## Success Signals
- Clear release narrative that ties to player or stakeholder value.
- Prioritized milestone list with acceptance criteria and owners.
- Updated risk register with mitigation owners and review dates.
- Communication plan for sharing decisions with the broader team.

## Tips
1. Time-box each agenda item; capture parking-lot topics in the issue for offline follow-up.
2. Require data to back status shifts—attach logs, dashboards, or test results where possible.
3. Encourage automation participation by flagging tasks in [../../.github/instructions/](../../.github/instructions/) that need agent support.
4. Close with explicit next steps so no decision remains ambiguous.

Revisit this guide whenever facilitation duties rotate or process adjustments are needed. Keeping the ritual consistent ensures the roadmap stays truthful and actionable.
