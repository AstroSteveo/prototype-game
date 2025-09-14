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
