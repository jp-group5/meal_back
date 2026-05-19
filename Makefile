SHELL := /bin/sh
DOCKER_ENV_FILE ?= .env.docker
COMPOSE ?= $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; elif command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo "docker compose"; fi)

.PHONY: help setup teardown run run-env up down logs ps test fmt vet tidy

help:
	@echo "Targets:"
	@echo "  make setup     - prepare Docker env file and validate compose config"
	@echo "  make run-env   - start api + postgres in Docker (foreground)"
	@echo "  make up        - start api + postgres in Docker (detached)"
	@echo "  make down      - stop services (keep db volume)"
	@echo "  make teardown  - stop services and remove db volume"
	@echo "                  - optional: TEARDOWN_ARGS='--yes --keep-db --remove-env' make teardown"
	@echo "  make logs      - stream compose logs"
	@echo "  make ps        - show compose service status"
	@echo "  make test      - run go test ./... inside Docker"
	@echo "  make fmt       - run gofmt on all go files inside Docker"
	@echo "  make vet       - run go vet ./... inside Docker"
	@echo "  make tidy      - run go mod tidy inside Docker"

setup:
	./scripts/setup_env.sh

teardown:
	./scripts/teardown_env.sh $(TEARDOWN_ARGS)

run: run-env

run-env: setup
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) up --build

up: setup
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) up --build -d

down:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) down --remove-orphans

logs:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) logs -f

ps:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) ps

test:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) run --rm api go test ./...

fmt:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) run --rm api sh -lc "gofmt -w \$$(find . -name '*.go' -not -path './.git/*')"

vet:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) run --rm api go vet ./...

tidy:
	$(COMPOSE) --env-file $(DOCKER_ENV_FILE) run --rm api go mod tidy
