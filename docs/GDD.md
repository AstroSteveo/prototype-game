# Game Design Document (GDD)

## Vision
- Build a “micro‑MMO”: a single global world that feels alive even for solo players.
- Signature feature: seamless local instancing (server meshing) so nearby players share the same experience automatically.
- Low‑commitment sessions: jump in, see activity (players or bots), make progress, log off without FOMO.

## Design Pillars
- Seamless world: no manual server selection; proximity‑based phasing.
- Always a crowd: server‑side bots fill gaps where population is low.
- Respect time: short, meaningful sessions; quick rejoin to last location.
- Fair play: server‑authoritative simulation; client prediction for feel, but server wins conflicts.

## Target Audience & Platform
- Audience: MMO fans who want social presence without the time sink.
- Platform: PC first. Networked client with server authoritative backend.

## Glossary (Player‑Facing Terms)
- Cell: invisible square region of the world that helps the server group nearby players.
- Instance: the server’s internal owner of a cell; players in adjacent cells still see each other if within radius.
- AOI (Area of Interest): how far you can “sense” other entities (players/bots) around you.
- Handover: transparent server step when you cross from one cell into a neighboring one.

## Core Player Loop (MVP)
1. Authenticate and spawn at last location (or a default spawn).
2. Move around freely in a shared world (WASD + mouse or gamepad).
3. See other entities within a radius (players and bots) with smooth updates.
4. Simple interaction: emote/ping and nameplate hover; optional proximity chat later.
5. Earn small, persistent progress (e.g., a “renown” or “steps walked” counter) to validate persistence.

## Content Scope (MVP)
- World: a simple test zone (flat plane or blocky test map) with coordinates in meters.
- Entities: players, ambient bots (wanderers), and static POIs (spawn stones, markers).
- Interactions: emotes (wave), proximity ping.
- Progression: lightweight stat/counter that persists; no inventories yet.

## AI “Population Filler” (Bots)
- Behavior: wander within a cell, occasionally change direction, avoid clustering too tightly.
- Density targets: configurable desired entities per cell; spawn/despawn to maintain range.
- Visibility: indistinguishable network replication from players (same interest management).

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
- Minimal HUD with nameplates, entity count in vicinity for debugging, and a net status indicator (latency, tick rate).
- On reconnect, restore position and instance gracefully.

### Accessibility & Comfort
- Play sessions target 10–20 minutes with immediate re‑entry to last location.
- Clear, readable nameplates; color‑blind‑safe debug overlays.
- Frame‑independent input, server reconciles to avoid sluggishness.

## Telemetry (Player‑Facing Success)
- The world feels populated (≥ N visible entities when moving through cells).
- Movement and entity updates feel smooth (rare rubber‑banding within MVP targets).
- Crossing “invisible” borders does not break immersion (handover < 250ms target).

## MVP Definition (Player Experience)
- Move, see others/bots within a defined radius.
- Automatic co‑presence with nearby players without choosing servers.
- Basic persistence (position + a simple stat) across sessions.

## User Stories (MVP)
- As a player, I log in and spawn where I last logged out within 1 meter.
- As a player, I move toward a crowd and immediately see their nameplates appear smoothly.
- As a player, I cross an invisible border without losing control or desync.
- As a solo player at off‑hours, I still see a few bots wandering nearby.
- As a returning player, my progress counter (e.g., renown/steps) increased since last session.

## Acceptance Criteria (Experience)
- Handover occurs within 250ms and does not eject me from AOI streaming.
- Entity popping is minimized; add/remove events are coherent while moving.
- Minimum population target is met: if < target in my vicinity, bots appear within 10s.
- Reconnect restores my state in < 2s with same appearance and position.

## Post‑MVP Ideas
- Proximity chat/VOIP, simple cooperative events, inventory/crafting, party sync, exploration goals, cosmetics.

## Debugging & Playtesting Aids
- Toggle overlay: current cell `(cx,cz)`, AOI count, tick rate, RTT.
- Visual borders (optional): faint debug lines for cells in dev builds.
- Command palette: teleport to cell, spawn N bots, freeze AI.

## Content Roadmap (Short‑Term)
- Biome props: sparse markers/landmarks per cell to improve orientation.
- Emote wheel with 2–3 expressive actions.
- Ambient audio zones tied to cells for subtle variation.

