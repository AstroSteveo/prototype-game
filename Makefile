SHELL := /bin/bash

# Paths
BACKEND := backend
BIN := $(BACKEND)/bin
PIDDIR := $(BACKEND)/.pids
LOGDIR := $(BACKEND)/logs

# Ports
GATEWAY_PORT ?= 8080
SIM_PORT ?= 8081
GATEWAY_URL := http://localhost:$(GATEWAY_PORT)
SIM_URL := http://localhost:$(SIM_PORT)

.PHONY: help fmt fmt-check vet test test-ws test-race test-ws-race build run run-gateway run-sim wait-up stop login wsprobe e2e-join e2e-move pr clean

help:
	@echo "Common targets:"
	@echo "  make run         # start gateway and sim (WS) in background"
	@echo "  make stop        # stop both services"
	@echo "  make login       # prints a dev token from gateway"
	@echo "  make wsprobe TOKEN=... [MOVE_X=1 MOVE_Z=0]  # probe WS"
	@echo "  make test        # run unit tests"
	@echo "  make test-ws     # run tests with -tags ws (includes WS integration)"
	@echo "  make test-race   # run unit tests with race detection"
	@echo "  make test-ws-race # run WS tests with race detection"
	@echo "  make build       # build gateway, sim (WS), and wsprobe binaries"
	@echo "  make fmt         # format Go code with gofmt"
	@echo "  make fmt-check   # check if Go code is formatted (non-mutating)"
	@echo "  make pr TITLE=... [BODY=... BASE=main HEAD=current DRAFT=1]  # open a PR (uses gh or curl)"

fmt:
	cd $(BACKEND) && go fmt ./...

fmt-check:
	@UNFORMATTED=$$(cd $(BACKEND) && gofmt -s -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Unformatted files:" && echo "$$UNFORMATTED"; \
		exit 1; \
	fi

vet:
	cd $(BACKEND) && go vet ./...

test:
	cd $(BACKEND) && go test ./...

test-ws:
	cd $(BACKEND) && go test -tags ws ./...

test-race:
	cd $(BACKEND) && go test -race ./...

test-ws-race:
	cd $(BACKEND) && go test -race -tags ws ./...

build:
	mkdir -p $(BIN)
	cd $(BACKEND) && go build -o bin/gateway ./cmd/gateway
	cd $(BACKEND) && go build -tags ws -o bin/sim ./cmd/sim
	cd $(BACKEND) && go build -o bin/wsprobe ./cmd/wsprobe

run-gateway:
	mkdir -p $(PIDDIR) $(LOGDIR)
	@( cd $(BACKEND) ; mkdir -p .pids logs ; nohup ./bin/gateway -port $(GATEWAY_PORT) -sim localhost:$(SIM_PORT) > logs/gateway.log 2>&1 & echo $$! > .pids/gateway.pid )
	@echo "gateway running on :$(GATEWAY_PORT) (pid $$(cat $(PIDDIR)/gateway.pid))"

run-sim:
	mkdir -p $(PIDDIR) $(LOGDIR)
	@( cd $(BACKEND) ; mkdir -p .pids logs ; nohup ./bin/sim -port $(SIM_PORT) > logs/sim.log 2>&1 & echo $$! > .pids/sim.pid )
	@echo "sim running on :$(SIM_PORT) (pid $$(cat $(PIDDIR)/sim.pid))"

wait-up:
	@echo "waiting for services..."
	@ok=0; for i in {1..100}; do curl -sf $(GATEWAY_URL)/healthz >/dev/null 2>&1 && { ok=1; break; } || sleep 0.1; done; \
	if [ $$ok -ne 1 ]; then \
		echo "gateway failed to start at $(GATEWAY_URL)"; \
		echo "-- gateway log (last 200 lines) --"; \
		tail -n 200 $(LOGDIR)/gateway.log 2>/dev/null || true; \
		exit 1; \
	fi
	@ok=0; for i in {1..100}; do curl -sf $(SIM_URL)/healthz >/dev/null 2>&1 && { ok=1; break; } || sleep 0.1; done; \
	if [ $$ok -ne 1 ]; then \
		echo "sim failed to start at $(SIM_URL)"; \
		echo "-- sim log (last 200 lines) --"; \
		tail -n 200 $(LOGDIR)/sim.log 2>/dev/null || true; \
		exit 1; \
	fi
	@echo "services healthy"

run: build run-gateway run-sim wait-up
	@echo "gateway: $(GATEWAY_URL)  sim: $(SIM_URL)"

stop:
	-@[ -f $(PIDDIR)/gateway.pid ] && kill $$(cat $(PIDDIR)/gateway.pid) >/dev/null 2>&1 || true
	-@[ -f $(PIDDIR)/sim.pid ] && kill $$(cat $(PIDDIR)/sim.pid) >/dev/null 2>&1 || true
	@rm -f $(PIDDIR)/*.pid 2>/dev/null || true
	@echo "stopped"

login:
	@curl -s "$(GATEWAY_URL)/login?name=Dev" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])'

MOVE_X ?= 0
MOVE_Z ?= 0
TOKEN ?=

wsprobe:
	@[ -n "$(TOKEN)" ] || (echo "Set TOKEN=<value> (hint: make login)"; exit 1)
	cd $(BACKEND) && go run ./cmd/wsprobe -url ws://localhost:$(SIM_PORT)/ws -token "$(TOKEN)" -move_x $(MOVE_X) -move_z $(MOVE_Z)

pr:
	@HEADBR=$(HEAD); \
	[ -n "$$HEADBR" ] || HEADBR=$$(git rev-parse --abbrev-ref HEAD); \
	BASEBR=$(BASE); \
	[ -n "$$BASEBR" ] || BASEBR=main; \
	REPO=$$(git config --get remote.origin.url | sed -E 's#(git@github.com:|https://github.com/)##; s#\.git$$##'); \
	if command -v gh >/dev/null 2>&1; then \
		CMD="gh pr create --base $$BASEBR --head $$HEADBR"; \
		[ -n "$(DRAFT)" ] && CMD="$$CMD --draft"; \
		if [ -n "$(TITLE)" ]; then CMD="$$CMD --title \"$(TITLE)\""; else CMD="$$CMD --fill"; fi; \
		[ -n "$(BODY)" ] && CMD="$$CMD --body \"$(BODY)\""; \
		echo "Running: $$CMD"; eval $$CMD; \
	else \
		[ -n "$$GITHUB_TOKEN" ] || { echo "GITHUB_TOKEN not set. Set it and re-run, or install gh CLI."; exit 1; }; \
		JSON=$$(TITLE="$(TITLE)" BODY="$(BODY)" BASE="$$BASEBR" HEAD="$$HEADBR" DRAFT="$(DRAFT)" python3 -c 'import json,os; draft=os.environ.get("DRAFT","").lower() in ("1","true","yes","y"); d={"head":os.environ.get("HEAD"),"base":(os.environ.get("BASE") or "main"),"draft":draft}; t=os.environ.get("TITLE"); b=os.environ.get("BODY"); d.update({k:v for k,v in (("title",t),("body",b)) if v}); print(json.dumps(d))'); \
		URL="https://api.github.com/repos/$$REPO/pulls"; \
		echo "POST $$URL"; \
		RESP=$$(curl -sfSL -H "Authorization: token $$GITHUB_TOKEN" -H "Accept: application/vnd.github+json" -X POST "$$URL" -d "$$JSON" || true); \
		if echo "$$RESP" | python3 -c "import sys, json; d=json.load(sys.stdin); import sys; sys.exit(0 if 'html_url' in d else 1)"; then \
			URL=$$(echo "$$RESP" | python3 -c "import sys, json; print(json.load(sys.stdin)['html_url'])"); \
			echo "PR created: $$URL"; \
		else \
			echo "Create PR via browser: https://github.com/$$REPO/pull/new/$$HEADBR"; \
			echo "API response:"; echo "$$RESP"; exit 1; \
		fi; \
	fi

e2e-join: run wait-up
	@TOKEN=$$(curl -s "$(GATEWAY_URL)/login?name=E2E" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])'); \
	cd $(BACKEND) && go run ./cmd/wsprobe -url ws://localhost:$(SIM_PORT)/ws -token "$$TOKEN"
	@$(MAKE) stop

e2e-move: run wait-up
	@TOKEN=$$(curl -s "$(GATEWAY_URL)/login?name=E2E" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])'); \
	cd $(BACKEND) && go run ./cmd/wsprobe -url ws://localhost:$(SIM_PORT)/ws -token "$$TOKEN" -move_x 1
	@$(MAKE) stop

clean:
	rm -rf $(BIN) $(LOGDIR) $(PIDDIR)
