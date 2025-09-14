## Quick Start (Makefile)
The repo includes a Makefile with common workflows.

- Start services (gateway + sim with WS):
  - `make run`
- Stop services:
  - `make stop`
- Get a dev token from the gateway:
  - `make login`
- Probe WebSocket (join only):
  - `make wsprobe TOKEN=<value>`

## Reconnect / Resume (WS)
