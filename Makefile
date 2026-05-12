SHELL := /bin/sh

.PHONY: help run run-env test fmt vet tidy

help:
	@echo "Targets:"
	@echo "  make run       - run backend (requires DB_DSN and JWT_SECRET in env)"
	@echo "  make run-env   - run backend by loading variables from .env first"
	@echo "  make test      - run go test ./..."
	@echo "  make fmt       - run gofmt on all go files"
	@echo "  make vet       - run go vet ./..."
	@echo "  make tidy      - run go mod tidy"

check-env:
	@if [ -z "$$DB_DSN" ]; then echo "ERROR: DB_DSN is not set"; exit 1; fi
	@if [ -z "$$JWT_SECRET" ]; then echo "ERROR: JWT_SECRET is not set"; exit 1; fi

run: check-env
	go run .

run-env:
	@if [ ! -f .env ]; then echo "ERROR: .env not found. Copy .env.example to .env first."; exit 1; fi
	@set -a; . ./.env; set +a; $(MAKE) run

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

vet:
	go vet ./...

tidy:
	go mod tidy
