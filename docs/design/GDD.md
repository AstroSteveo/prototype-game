# Game Design Document (GDD)

## Vision
- Build a “micro‑MMO”: a single global world that feels alive even for solo players.
- Signature feature: seamless local instancing (server meshing) so nearby players share the same experience automatically.
- Low‑commitment sessions: jump in, see activity (players or bots), make progress, log off without FOMO.
- Layer lightweight RPG systems—inventory, equipment, targeting, and skills—so every short session still produces tangible growth.

## Design Pillars
- Seamless world: no manual server selection; proximity‑based phasing.
- Always a crowd: server‑side bots fill gaps where population is low.
- Respect time: short, meaningful sessions; quick rejoin to last location.
- Fair play: server‑authoritative simulation; client prediction for feel, but server wins conflicts.
- Meaningful loadouts: constrained inventory, explicit equipment slots, and clear target intel push interesting choices.[^1][^2][^3][^4]

## Target Audience & Platform
- Audience: MMO fans who want social presence without the time sink.
- Platform: PC first. Networked client with server authoritative backend.

## Glossary (Player‑Facing Terms)
- Cell: invisible square region of the world that helps the server group nearby players.
- Instance: the server’s internal owner of a cell; players in adjacent cells still see each other if within radius.
- AOI (Area of Interest): how far you can “sense” other entities (players/bots) around you.
- Handover: transparent server step when you cross from one cell into a neighboring one.
- Loadout: the combination of equipped gear, slotted abilities, and consumables for a given situation.

## Core Player Loop (MVP)
1. Authenticate and spawn at last location (or a default spawn).
2. Move around freely in a shared world (WASD + mouse or gamepad).
3. See other entities within a radius (players and bots) with smooth updates.
4. Loot or craft items and manage a limited inventory with weight/bulk trade‑offs.[^2]
5. Equip items into defined slots (hands, armor, tools) to update your loadout.[^1][^3]
6. Acquire a target via soft‑lock/tab targeting; read difficulty and status cues from the HUD.[^4]
7. Trigger skills that combine stanzas (ability components) to apply damage types and utility effects.[^5][^6]
8. Earn skill‑specific experience from resolved actions; bank spendable skill points.[^7]
9. Log out or swap roles; persistence saves position, inventory, equipment, and skill progress.

## Content Scope (MVP)
- World: a simple test zone (flat plane or blocky test map) with coordinates in meters.
- Entities: players, ambient bots (wanderers), static POIs (spawn stones, markers), and lootable nodes/crates (interactable objects that can be looted by players; once looted, they become inactive and respawn after a fixed interval or when the cell is empty; interaction is proximity-based with a short pickup animation and clear feedback).
- Interactions: emotes (wave), proximity ping, item pickup, targeted ability use with clear cast feedback.
- Progression: per‑skill experience, gated abilities, and spendable skill points to slot new stanzas.[^6][^7]
- Inventory & Equipment: backpack with capacity (weight + bulk), quick access belt, hand slots, and contextual equip cooldown.[^1][^2]
- Combat/Utility: baseline damage types (slash, pierce, blunt) and simple status riders (slow, resist buff) surfaced through stanzas.[^5][^6]

## Systems Overview (MVP)
### Inventory & Equipment
- Dedicated UI for hands and gear slots; equipping from inventory triggers a short lockout before use.[^1]
- Items track weight and bulk; exceeding limits slows movement and blocks new pickups until space frees up.[^2]
- Equipment requirements reference skill level and role mastery to reinforce build identity.[^3]

### Targeting & Combat Awareness
- HUD panel mirrors target name, title, vital gauges, and aggression/difficulty cues with region color coding.[^4]
- Target selection defaults to soft lock with optional manual cycling; server validates target visibility within AOI.
- Difficulty colors plus star ranks set expectations for solo vs. group viability, guiding social play.

### Skills, Damage Types, and Status Effects
- Actions assemble modular “stanzas” to mix delivery (melee/ranged), damage type, and secondary effects.[^6]
- Core damage taxonomy includes slash, pierce, and blunt mitigation along with magical resistances.[^5]
- Ability unlocks cost skill points and can be remixed to build bespoke rotations or crafting macros.

### Experience & Skill Progression
- Each resolved action (combat hit, craft completion, gather, support cast) awards experience to the relevant skill line.[^7]
- Skill rank thresholds unlock additional stanzas and item proficiencies; surplus XP feeds a shared pool for hybrid builds.
- Progress persists per character, enabling experimentation without rerolls.

## AI “Population Filler” (Bots)
- Behavior: wander within a cell, occasionally change direction, avoid clustering too tightly.
- Density targets: configurable desired entities per cell; spawn/despawn to maintain range.
- Visibility: indistinguishable network replication from players (same interest management). Bots drop basic loot tables for inventory validation.

## Sharding/Phasing Experience
- Players are automatically co‑located with nearby players into a local instance for shared visibility.
- Transfers are seamless when crossing cell boundaries (no loading screens; minimal hitch acceptable for MVP).

### What “Local Sharding” Means (MVP)
- The world is divided into a grid of cells, each owned by a local “instance” that runs inside the same server process.
- When you cross a cell edge, the server hands your character to the neighboring instance instantly; your network connection stays the same.
- You still see entities in adjacent cells if they are within your AOI.

```
   cx-1,cz+1   cx,cz+1     cx+1,cz+1
   ┌────────┬────────┬────────┐
   │        │        │        │
   │  NW    │   N    │   NE   │
   ├────────┼────────┼────────┤
   │        │  You→  │        │
   │   W    │  Cx,Cz │    E   │  ← handover triggers when crossing lines
   ├────────┼────────┼────────┤
   │        │        │        │
   │  SW    │   S    │   SE   │
   └────────┴────────┴────────┘

AOI = circle around you; the server streams entities from your cell and neighbors.
```

## UX & Presentation
- Minimal HUD with nameplates, inventory quick slots, target panel, and a net status indicator (latency, tick rate).
- Equip cooldown overlay and encumbrance warnings keep loadout choices legible.[^1][^2]
- On reconnect, restore position, equipment, and target states gracefully.

### Accessibility & Comfort
- Play sessions target 10–20 minutes with immediate re‑entry to last location.
- Clear, readable nameplates; color‑blind‑safe debug overlays and damage icons.
- Frame‑independent input, server reconciles to avoid sluggishness.

## Telemetry (Player‑Facing Success)
- The world feels populated (≥ N visible entities when moving through cells).
- Loadout friction is meaningful but not punitive (average encumbrance warnings < 10% of playtime).
- Combat clarity: 90% of ability uses show target feedback within 300ms; targeting errors < 2%.
- Crossing “invisible” borders does not break immersion (handover < 250ms target).

## MVP Definition (Player Experience)
- Move, see others/bots within a defined radius.
- Manage a constrained inventory, equip contextually appropriate gear, and feel the impact immediately.
- Acquire targets and fire abilities that respect damage types and telegraph outcomes.
- Earn per‑skill experience and unlock at least one new stanza or item proficiency per short session.
- Automatic co‑presence with nearby players without choosing servers.
- Basic persistence (position + loadout + skill progress) across sessions.

## User Stories (MVP)
- As a player, I log in and spawn where I last logged out within 1 meter.
- As a player, I loot a weapon upgrade, equip it, and feel its stats reflected on my next attack.
- As a player, I move toward a crowd and immediately see their nameplates appear smoothly.
- As a player, I tab‑target an enemy and read their relative difficulty before engaging.[^4]
- As a player, I cross an invisible border without losing control or desync.
- As a solo player at off‑hours, I still see a few bots wandering nearby.
- As a returning player, my skill line shows increased experience and new stanza options since last session.[^6][^7]

## Acceptance Criteria (Experience)
- Handover occurs within 250ms and does not eject me from AOI streaming.
- Inventory UI prevents equipping items I do not meet requirements for.[^3]
- Encumbrance feedback fires before movement penalties apply, and weight/bulk caps are enforced server side.[^2]
- Target panel always reflects current difficulty color within 200ms of target change.[^4]
- Ability execution resolves damage type correctly and applies promised status effects.[^5][^6]
- Per‑skill XP increments only once per validated action and persists through reconnects.[^7]
- Reconnect restores my state in < 2s with same appearance, position, and equipped gear.

## Post‑MVP Ideas
- Proximity chat/VOIP, simple cooperative events, inventory crafting tiers, party sync, exploration goals, cosmetics.
- Trading posts and shared stash tabs once inventory UX hardens.
- Advanced targeting modes (cone/ground reticle) layered atop tab targeting.

## Debugging & Playtesting Aids
- Toggle overlay: current cell `(cx,cz)`, AOI count, tick rate, RTT, encumbrance %, and currently selected target metadata.
- Visual borders (optional): faint debug lines for cells in dev builds.
- Command palette: teleport to cell, spawn N bots, freeze AI, grant temporary gear, and reset XP for a skill line.

## Content Roadmap (Short‑Term)
- Biome props: sparse markers/landmarks per cell to improve orientation.
- Emote wheel with 2–3 expressive actions.
- Prototype three starter loadouts (melee, caster, crafter) with curated inventories and ability kits.[^6]
- Ambient audio zones tied to cells for subtle variation.

## References
[^1]: Ryzom Core – “How to change the items in your hands” help page. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/how_to_changeitemsinhand_en.html
[^2]: Ryzom Core – Craft tool item info highlighting weight/bulk limits. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_item_craft_tool_en.html
[^3]: Ryzom Core – Abilities and items guide noting skill requirements for equipment. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/abilities_item_step4_en.html
[^4]: Ryzom Core – Target interface documentation with level color coding. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_target_en.html
[^5]: Ryzom Core – Shield item info describing damage type mitigation. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_item_shield_en.html
[^6]: Ryzom Core – Stanza system overview for modular ability construction. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_sbrick_en.html
[^7]: Ryzom Core – Server command definitions showing per-skill XP and skill points. https://github.com/ryzom/ryzomcore/blob/master/ryzom/server/data_shard/egs/client_commands_privileges_open.txt
