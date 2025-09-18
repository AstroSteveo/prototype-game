# Vision–Requirements Traceability Matrix (VRTM)

This document maps the project vision, design pillars, and user stories from the Game Design Document (GDD) to the technical systems, milestones, and implementation components defined in the Technical Design Document (TDD).

**Last Updated**: 2025-01-15  
**Next Review**: During roadmap alignment sessions (see [ROADMAP_ALIGNMENT_SESSION.md](../sessions/ROADMAP_ALIGNMENT_SESSION.md))

## Purpose

The VRTM ensures that:
- Every design pillar has corresponding technical systems
- User stories are mapped to specific milestones and components
- Implementation gaps are identified and tracked
- Technical decisions support the product vision

## Vision Mapping

### Core Vision
**"Build a 'micro-MMO': a single global world that feels alive even for solo players."**

| Vision Element | Technical Implementation | Status |
|----------------|-------------------------|---------|
| Single global world | Cell-based spatial partitioning with seamless handovers | ✅ Implemented (M3) |
| Feels alive | Bot density management and AOI streaming | ✅ Implemented (M4) |
| Solo player support | Always-populated cells with AI entities | ✅ Implemented (M4) |
| Micro-MMO scale | Local sharding architecture (Phase A) | ✅ Implemented (M3) |

## Design Pillars Traceability

### Pillar 1: Seamless World
**"No manual server selection; proximity-based phasing"**

| Component | System/Module | Milestone | Implementation Status | Risk Level |
|-----------|---------------|-----------|----------------------|------------|
| Spatial partitioning | `internal/spatial` | M1-M3 | ✅ Complete | Low |
| Cell ownership | `sim/cells` in TDD | M3 | ✅ Complete | Low |
| Handover system | `sim/engine`, `HandoverManager` | M3 | ✅ Complete | Medium |
| Cross-cell AOI | `AOIIndex` interface | M2-M3 | ✅ Complete | Low |

**Acceptance Criteria Met:**
- ✅ Handover occurs within 250ms target
- ✅ No AOI streaming interruption during handover
- ✅ Position continuity across cell boundaries

### Pillar 2: Always a Crowd
**"Server-side bots fill gaps where population is low"**

| Component | System/Module | Milestone | Implementation Status | Risk Level |
|-----------|---------------|-----------|----------------------|------------|
| Bot spawning | `sim/bots` in TDD | M4 | ✅ Complete | Low |
| Density targets | `BotSpawner` interface | M4 | ✅ Complete | Low |
| Wandering behavior | Bot AI system | M4 | ✅ Complete | Low |
| Population management | PID-lite density control | M4 | ✅ Complete | Medium |

**Acceptance Criteria Met:**
- ✅ Maintains configured min entities ±20%
- ✅ Bots indistinguishable from players in networking
- ✅ Simple avoidance and wandering behaviors

### Pillar 3: Respect Time
**"Short, meaningful sessions; quick rejoin to last location"**

| Component | System/Module | Milestone | Implementation Status | Risk Level |
|-----------|---------------|-----------|----------------------|------------|
| Position persistence | Player data model | M5 | ✅ Complete | Low |
| Fast reconnect | Session management | M5 | ✅ Complete | Medium |
| Resume tokens | Gateway service | M5 | ✅ Complete | Low |
| State restoration | Full player state | M5 | ✅ Complete | Medium |

**Acceptance Criteria Met:**
- ✅ Reconnect in < 2s
- ✅ Spawn within 1m of saved position
- ✅ Inventory and equipment parity
- ✅ No XP loss on reconnect

### Pillar 4: Fair Play
**"Server-authoritative simulation; client prediction for feel"**

| Component | System/Module | Milestone | Implementation Status | Risk Level |
|-----------|---------------|-----------|----------------------|------------|
| Authoritative physics | `sim/engine` | M1 | ✅ Complete | Low |
| Input validation | Server tick loop | M1 | ✅ Complete | Low |
| Prediction reconciliation | Client-server protocol | M1 | ✅ Complete | Medium |
| Anti-cheat validation | Velocity/position checks | M1 | ✅ Complete | Medium |

**Acceptance Criteria Met:**
- ✅ 20 Hz server tick
- ✅ Server wins all conflicts
- ✅ Impossible velocities rejected

### Pillar 5: Meaningful Loadouts
**"Constrained inventory, explicit equipment slots, clear target intel"**

| Component | System/Module | Milestone | Implementation Status | Risk Level |
|-----------|---------------|-----------|----------------------|------------|
| Inventory system | Data model + UI | Post-MVP | 🟡 In Progress | High |
| Equipment slots | `equipment_slots` table | Post-MVP | 🟡 In Progress | High |
| Weight/bulk limits | Encumbrance system | Post-MVP | 🟡 In Progress | Medium |
| Skill requirements | `skill_req` validation | Post-MVP | 🟡 In Progress | High |
| Target difficulty | Level band color coding | Post-MVP | 🟡 In Progress | Medium |

**Gap Analysis:**
- ❌ Equipment cooldown system not implemented
- ❌ Skill gating for equipment not implemented
- ❌ Target difficulty color system not implemented

## User Stories Traceability

### Core Player Loop Stories

| User Story | System Component | Milestone | Status | Acceptance Criteria |
|------------|------------------|-----------|--------|-------------------|
| "I log in and spawn where I last logged out within 1 meter" | Position persistence | M5 | ✅ Complete | Reconnect < 2s, position ±1m |
| "I loot a weapon upgrade, equip it, and feel its stats reflected" | Inventory + Equipment | Post-MVP | 🔴 Not Started | Equipment affects combat stats |
| "I move toward a crowd and see their nameplates appear smoothly" | AOI streaming | M2 | ✅ Complete | Smooth add/remove, no flapping |
| "I tab-target an enemy and read their relative difficulty" | Targeting system | Post-MVP | 🔴 Not Started | Difficulty colors, target metadata |
| "I cross an invisible border without losing control or desync" | Handover system | M3 | ✅ Complete | < 250ms handover, no desync |
| "I still see a few bots wandering nearby during off-hours" | Bot density | M4 | ✅ Complete | Maintains target ±20% |
| "My skill line shows increased XP and new stanza options" | Skill progression | Post-MVP | 🔴 Not Started | XP persistence, unlock notifications |

### Authentication & Session Stories

| User Story | System Component | Milestone | Status | Notes |
|------------|------------------|-----------|--------|----- |
| Authenticate and spawn at last location | Gateway + Join logic | M0-M5 | ✅ Complete | Token-based auth working |
| Resume session with minimal hitch | Session management | M5 | ✅ Complete | Resume tokens implemented |
| Maintain connection across handovers | Handover transparency | M3 | ✅ Complete | Same WS connection maintained |

### Movement & Presence Stories

| User Story | System Component | Milestone | Status | Notes |
|------------|------------------|-----------|--------|----- |
| Move around freely (WASD + mouse) | Input handling | M1 | ✅ Complete | 20 Hz input processing |
| See other entities within radius | AOI system | M2 | ✅ Complete | 128m radius, 3x3 cell query |
| Smooth updates for visible entities | Replication system | M2 | ✅ Complete | 10 Hz snapshots |

## Technical Milestones Mapping

### M0: Project Skeleton ✅ Complete
- **Vision Support**: Infrastructure foundation
- **Deliverables**: Gateway and sim service scaffolding
- **User Value**: Enables all subsequent development

### M1: Presence & Movement ✅ Complete
- **Vision Support**: Fair play pillar (server authoritative)
- **Deliverables**: Connect, spawn, move with prediction
- **User Value**: Basic interaction with game world

### M2: Interest Management ✅ Complete
- **Vision Support**: Always a crowd, seamless world
- **Deliverables**: AOI streaming, entity visibility
- **User Value**: See other players and activity

### M3: Local Sharding ✅ Complete
- **Vision Support**: Seamless world pillar
- **Deliverables**: Multi-cell handover system
- **User Value**: Uninterrupted movement across world

### M4: Bots & Density ✅ Complete
- **Vision Support**: Always a crowd pillar
- **Deliverables**: AI entities maintaining population
- **User Value**: World feels alive even when alone

### M5: Persistence ✅ Complete
- **Vision Support**: Respect time pillar
- **Deliverables**: Position, state persistence
- **User Value**: Quick resume of previous session

### Post-MVP: Advanced Systems 🔴 Not Started
- **Vision Support**: Meaningful loadouts pillar
- **Deliverables**: Inventory, equipment, skills, targeting
- **User Value**: Character progression and combat

## Gap Analysis & Roadmap Issues

**Updated**: 2025-01-15 (Post US-002 Roadmap Alignment Workshop)

### Roadmap Alignment Workshop Outcomes
Following the comprehensive roadmap alignment session for US-002, the following post-MVP milestone structure has been established to address the "Meaningful Loadouts" design pillar:

**M6-M9 Post-MVP Timeline**: 13-17 weeks total
- **M6**: Equipment Foundation (4-5 weeks)
- **M7**: Skill Progression System (3-4 weeks) 
- **M8**: Targeting & Combat Resolution (4-5 weeks)
- **M9**: Integration & Polish (2-3 weeks)

### High Priority Gaps (M6-M8 Scope)
1. **Equipment System** - **M6 Milestone** - Critical for "meaningful loadouts" pillar
   - Missing: Equipment slots, cooldowns, stat effects, persistence
   - Impact: Core user story "equip weapon and feel stats" blocked
   - Implementation: Database schema, server validation, basic equip/unequip
   - Timeline: 4-5 weeks
   - Dependencies: None (foundation system)
   - **Status**: Ready for M6 implementation planning

2. **Skill Progression** - **M7 Milestone** - Required for character advancement
   - Missing: XP pipeline, skill unlocks, stanza system, skill gating
   - Impact: Progression user stories blocked, equipment requirements
   - Implementation: XP events, skill validation, unlock notifications
   - Timeline: 3-4 weeks
   - Dependencies: Equipment system (M6) for skill requirements
   - **Status**: Blocked until M6 completion

3. **Targeting System** - **M8 Milestone** - Essential for combat readiness
   - Missing: Target selection, difficulty display, metadata, combat resolution
   - Impact: Combat preparation user stories blocked
   - Implementation: Target validation, difficulty colors, damage types
   - Timeline: 4-5 weeks
   - Dependencies: Equipment (M6) and Skills (M7) for stat calculations
   - **Status**: Blocked until M6-M7 completion

### Medium Priority Gaps (M8-M9 Scope)
4. **Inventory UI** - **M9 Milestone** - Needed for equipment management
   - Missing: Weight/bulk visualization, equip interactions, drag-drop
   - Impact: User experience for loadout management
   - Implementation: Frontend integration with backend inventory system
   - Timeline: Included in M9 polish phase
   - Dependencies: Equipment system (M6) completion
   - **Status**: Deferred to M9 integration phase

5. **Combat Resolution** - **M8 Milestone** - Required for ability execution
   - Missing: Damage types, status effects, ability system integration
   - Impact: Combat user stories blocked
   - Implementation: Damage calculation, mitigation tables, status effects
   - Timeline: Included in M8 with targeting system
   - Dependencies: Equipment (M6) and Skills (M7) for stat effects
   - **Status**: Blocked until M6-M7 completion

### Low Priority Gaps (Post M9)
6. **Advanced Networking** - Cross-node handover (Phase B)
   - Missing: Multi-node architecture
   - Impact: Scalability beyond single process
   - Recommendation: Defer to Phase B milestone (post-loadouts)
   - **Status**: Unchanged - deferred to future phase

### Risk Assessment Updates (Post-Workshop)

#### Newly Identified Risks
- **Equipment-Skill System Complexity**: High interdependency between M6/M7 systems
  - Mitigation: Phased delivery with equipment foundation first
  - Owner: Technical Lead
  - Review: Every 2 weeks during M6-M7 implementation

- **Performance Under Load**: Encumbrance calculations and stat effects may impact tick budget
  - Mitigation: Early performance testing with synthetic equipment data
  - Owner: Backend Team
  - Review: During M6 implementation, before M7 start

- **UI/UX Complexity**: Inventory interface design complexity underestimated
  - Mitigation: Early prototype and user testing during M6
  - Owner: Frontend Team (when available)
  - Review: M8 planning phase

## Risk Assessment

### High Risk Items
- **Equipment System Complexity**: Interdependent with inventory, skills, and combat
- **Skill Progression Balance**: Complex game design decisions required
- **Performance Under Load**: Need validation with realistic player counts

### Medium Risk Items
- **Database Schema Evolution**: Inventory/equipment schema changes
- **Client-Server Protocol**: Addition of equipment/skill messages
- **Testing Coverage**: Need comprehensive integration tests for new systems

### Low Risk Items
- **Documentation Updates**: Well-established process
- **Monitoring Integration**: Existing metrics infrastructure

## Review Process

### Quarterly Reviews
- Validate all mappings remain current
- Update milestone status and acceptance criteria
- Reassess risk levels based on implementation progress
- Identify new gaps from evolving requirements

### Release Planning Integration
- **Before each release planning session**: Review VRTM for current status
- Use VRTM to guide feature prioritization based on vision alignment
- Ensure milestone dependencies support design pillars
- Validate technical decisions against user story requirements
- **After each release planning session**: Update VRTM with new roadmap decisions
- Update milestone status and timelines based on planning outcomes
- Reassess risk levels for upcoming milestones
- Document any new gaps or dependencies identified

### Post-US-002 Workshop Integration
- **January 15, 2025**: Comprehensive roadmap alignment workshop completed
- **M6-M9 Milestone Structure**: Established for "Meaningful Loadouts" implementation
- **GitHub Issues**: [Defined for M6-M9 milestones](../sessions/github-issues-m6-m9.md) (pending creation)
- **Stakeholder Communication**: [Workshop outcomes documented](../sessions/stakeholder-communication-us002.md)
- **Next Workshop**: Scheduled after M6 completion to assess progress and refine M7-M9 scope

### Continuous Maintenance
- Update status after milestone completion
- Add new user stories as they emerge
- Track implementation decisions and their vision alignment
- Maintain links to supporting technical documentation

---

## References
- [Game Design Document](../../product/vision/game-design-document.md) - Source of design pillars and user stories
- [Technical Design Document](../../architecture/technical-design-document.md) - Implementation details and milestones
- [Developer Guide](../../development/developer-guide.md) - Build and test procedures
- [Roadmap Planning](../sessions/ROADMAP.md) - Planning session template