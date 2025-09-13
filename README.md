# prototype-game

Design docs:
- `docs/design/GDD.md` — Game Design Document (vision, player experience, scope)
- `docs/design/TDD.md` — Technical Design Document (architecture, sharding plan, milestones)
- `docs/process/PROGRESS.md` — Running progress log with milestone status and next steps
- `docs/process/BACKLOG.md` — Prioritized, testable user stories for MVP → stretch
- `docs/dev/DEV.md` — Developer Guide (build, run, tests, Makefile)
- `.github/copilot-instructions.md` — GitHub Copilot/AI agent instructions

Quick start (Go backend, local dev):
- Makefile (recommended): `make run` then `make login`
- Manual:
  - `cd backend`
  - Run sim (WS enabled): `go run -tags ws ./cmd/sim --port 8081`
  - Run gateway: `go run ./cmd/gateway --port 8080 --sim localhost:8081`
  - Health checks: `curl localhost:8081/healthz` and `curl localhost:8080/healthz`
  - Login (dev): `curl 'http://localhost:8080/login?name=Test'`
  - Validate token (dev): `curl 'http://localhost:8080/validate?token=<token>'`

WebSocket (US-101)
- Sim registers `/ws` endpoint. By default, it is a stub returning `501` until built with the `ws` build tag.
- Enable WS: `go run -tags ws ./cmd/sim --gateway http://localhost:8080`
- Login response includes WebSocket URL: `{ "sim": { "address": "ws://host:port/ws", "protocol": "ws-json", "version": "1" } }`
- First message from client: `{"token":"..."}`. Server replies with `{"type":"join_ack","data":{...}}` or `{"type":"error",...}`.

Notes:
- M0 focuses on the simulation loop, cell math, and handover logic in-process (local sharding). Networking to clients is stubbed until WebSocket transport is added.
