---
title: "task: Refine GDD/TDD for core MMO systems"
labels:
  - task
project: https://github.com/users/AstroSteveo/projects/2
---

## Summary
The design docs need to capture the MMO gameplay pillars discussed by the team: inventories with equipment slots, a server-driven targeting panel, per-skill experience gains, and modular skills with varied damage types. The revised documentation should cite reliable references to justify each mechanic and keep options open while the exact leveling cadence is still under evaluation.

## Scope
- Update the Game Design Document (GDD) with inventory management, equipment slot flow, targeting UX, damage taxonomy, and skill progression beats supported by external references.
- Update the Technical Design Document (TDD) with data models, network events, and server logic for inventory/equipment, targeting validation, skill stanzas, and XP pipelines.
- Highlight that players earn experience through in-world actions even though the long-term leveling curve is undecided.
- Capture telemetry, testing, and security implications introduced by the new systems.

## Definition of Done
- GDD and TDD sections describe inventory, equipment, targeting, skills, and XP with traceable citations.
- Revised docs note the tentative nature of overall leveling pacing while affirming XP gains per action.
- Data models and message flows cover inventory diffs, equip cooldowns, target updates, and skill progress events.
- Issue is associated with Game Roadmap Project #2 for tracking.

## Notes
- Leveling structure remains flexible; document the assumptions that can shift without rework.
- Use the existing task issue template language (Scope / Definition of Done) when creating the live GitHub issue.
