# Go backend common tasks
# 使用方式：make <target>

SHELL := /bin/zsh

APP_NAME := meal_back
APP_ENTRY := main.go
API_BASE ?= http://localhost:8080/api/v1
TEST_EMAIL ?= test@example.com
TEST_USERNAME ?= testuser01
TEST_PASSWORD ?= 12345678

.PHONY: help run test tidy fmt lint register login me refresh logout smoke userctl cleanup-users

help:
	@echo "Available targets:"
	@echo "  make run       - 启动后端服务"
	@echo "  make test      - 运行 go test ./..."
	@echo "  make tidy      - 整理依赖 go mod tidy"
	@echo "  make fmt       - 格式化代码（gofmt）"
	@echo "  make lint      - 基础静态检查（go vet）"
	@echo "  make smoke     - 编译检查（go test ./...）"
	@echo "  make register  - 调用注册接口"
	@echo "  make login     - 调用登录接口"
	@echo "  make me        - 调用 /private/me（需 ACCESS_TOKEN）"
	@echo "  make refresh   - 调用 /refresh（需 REFRESH_TOKEN）"
	@echo "  make logout    - 调用 /private/logout（需 ACCESS_TOKEN）"
	@echo "  make userctl   - 运行数据库用户账号管理工具（通过 ARGS 传子命令）"
	@echo "  make cleanup-users - 清理所有已注册用户（危险，需 CONFIRM=YES）"
	@echo ""
	@echo "Required env vars for run: DB_DSN, JWT_SECRET"
	@echo "Optional vars: API_BASE, TEST_USERNAME, TEST_EMAIL, TEST_PASSWORD, ACCESS_TOKEN, REFRESH_TOKEN"
	@echo "userctl vars: ARGS (示例: make userctl ARGS='list --limit 50')"
	@echo "cleanup vars: CONFIRM=YES"

run:
	@if [ -z "$$DB_DSN" ]; then echo "ERROR: DB_DSN 未设置"; exit 1; fi
	@if [ -z "$$JWT_SECRET" ]; then echo "ERROR: JWT_SECRET 未设置"; exit 1; fi
	@DB_HOST=$$(echo "$$DB_DSN" | tr ' ' '\n' | awk -F= '$$1=="host"{print $$2}'); \
	DB_PORT=$$(echo "$$DB_DSN" | tr ' ' '\n' | awk -F= '$$1=="port"{print $$2}'); \
	DB_NAME=$$(echo "$$DB_DSN" | tr ' ' '\n' | awk -F= '$$1=="dbname"{print $$2}'); \
	DB_USER=$$(echo "$$DB_DSN" | tr ' ' '\n' | awk -F= '$$1=="user"{print $$2}'); \
	if [ -z "$$DB_HOST" ]; then DB_HOST=127.0.0.1; fi; \
	if [ -z "$$DB_PORT" ]; then DB_PORT=5432; fi; \
	if [ -z "$$DB_NAME" ]; then DB_NAME="(未设置)"; fi; \
	if [ -z "$$DB_USER" ]; then DB_USER="(未设置)"; fi; \
	echo "DB check => host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME user=$$DB_USER"; \
	if command -v pg_isready >/dev/null 2>&1; then \
		if ! pg_isready -h "$$DB_HOST" -p "$$DB_PORT" -U "$$DB_USER" >/dev/null 2>&1; then \
			echo "ERROR: PostgreSQL 未就绪或不可达。请先启动数据库，并确认 DB_DSN 配置正确。"; \
			exit 1; \
		fi; \
	else \
		echo "WARN: 未找到 pg_isready，跳过数据库健康检查。"; \
	fi
	go run $(APP_ENTRY)

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...

smoke:
	go test ./...

register:
	@echo "POST $(API_BASE)/register"
	@curl -sS -X POST "$(API_BASE)/register" \
		-H "Content-Type: application/json" \
		-d '{"username":"$(TEST_USERNAME)","email":"$(TEST_EMAIL)","password":"$(TEST_PASSWORD)"}' | jq .

login:
	@echo "POST $(API_BASE)/login"
	@curl -sS -X POST "$(API_BASE)/login" \
		-H "Content-Type: application/json" \
		-d '{"username":"$(TEST_USERNAME)","password":"$(TEST_PASSWORD)"}' | jq .

me:
	@if [ -z "$$ACCESS_TOKEN" ]; then echo "ERROR: ACCESS_TOKEN 未设置"; exit 1; fi
	@echo "GET $(API_BASE)/private/me"
	@curl -sS -X GET "$(API_BASE)/private/me" \
		-H "Authorization: Bearer $$ACCESS_TOKEN" | jq .

refresh:
	@if [ -z "$$REFRESH_TOKEN" ]; then echo "ERROR: REFRESH_TOKEN 未设置"; exit 1; fi
	@echo "POST $(API_BASE)/refresh"
	@curl -sS -X POST "$(API_BASE)/refresh" \
		-H "Content-Type: application/json" \
		-d '{"refresh_token":"'"$$REFRESH_TOKEN"'"}' | jq .

logout:
	@if [ -z "$$ACCESS_TOKEN" ]; then echo "ERROR: ACCESS_TOKEN 未设置"; exit 1; fi
	@echo "POST $(API_BASE)/private/logout"
	@curl -sS -X POST "$(API_BASE)/private/logout" \
		-H "Authorization: Bearer $$ACCESS_TOKEN" | jq .

userctl:
	@if [ -z "$$DB_DSN" ]; then echo "ERROR: DB_DSN 未设置"; exit 1; fi
	go run ./cmd/userctl $(ARGS)

cleanup-users:
	@if [ -z "$$DB_DSN" ]; then echo "ERROR: DB_DSN 未设置"; exit 1; fi
	@if [ "$$CONFIRM" != "YES" ]; then echo "ERROR: 该操作会清空所有用户。请显式确认：CONFIRM=YES make cleanup-users"; exit 1; fi
	./scripts/cleanup_all_users.sh --yes
