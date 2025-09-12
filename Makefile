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

.PHONY: help fmt vet test test-ws build run run-gateway run-sim wait-up stop login wsprobe e2e-join e2e-move clean

help:
	@echo "Common targets:"
	@echo "  make run         # start gateway and sim (WS) in background"
	@echo "  make stop        # stop both services"
	@echo "  make login       # prints a dev token from gateway"
	@echo "  make wsprobe TOKEN=... [MOVE_X=1 MOVE_Z=0]  # probe WS"
	@echo "  make test        # run unit tests"
	@echo "  make test-ws     # run tests with -tags ws (includes WS integration)"
	@echo "  make build       # build gateway, sim (WS), and wsprobe binaries"

fmt:
	cd $(BACKEND) && go fmt ./...

vet:
	cd $(BACKEND) && go vet ./...

test:
	cd $(BACKEND) && go test ./...

test-ws:
	cd $(BACKEND) && go test -tags ws ./...

build:
	mkdir -p $(BIN)
	cd $(BACKEND) && go build -o bin/gateway ./cmd/gateway
	cd $(BACKEND) && go build -tags ws -o bin/sim ./cmd/sim
	cd $(BACKEND) && go build -o bin/wsprobe ./cmd/wsprobe

run-gateway:
	mkdir -p $(PIDDIR) $(LOGDIR)
	cd $(BACKEND) && nohup go run ./cmd/gateway -port $(GATEWAY_PORT) -sim localhost:$(SIM_PORT) > $(LOGDIR)/gateway.log 2>&1 & echo $$! > $(PIDDIR)/gateway.pid ; \
	echo "gateway running on :$(GATEWAY_PORT) (pid $$(cat $(PIDDIR)/gateway.pid))"

run-sim:
	mkdir -p $(PIDDIR) $(LOGDIR)
	cd $(BACKEND) && nohup go run -tags ws ./cmd/sim -port $(SIM_PORT) > $(LOGDIR)/sim.log 2>&1 & echo $$! > $(PIDDIR)/sim.pid ; \
	echo "sim running on :$(SIM_PORT) (pid $$(cat $(PIDDIR)/sim.pid))"

wait-up:
	@echo "waiting for services..."
	@for i in {1..50}; do curl -sf $(GATEWAY_URL)/healthz >/dev/null 2>&1 && break || sleep 0.1; done
	@for i in {1..50}; do curl -sf $(SIM_URL)/healthz >/dev/null 2>&1 && break || sleep 0.1; done
	@echo "services healthy"

run: run-gateway run-sim wait-up
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

