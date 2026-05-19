#!/usr/bin/env bash
set -euo pipefail

APP_NAME="meal_back"
DOCKER_ENV_FILE=".env.docker"
DOCKER_ENV_EXAMPLE=".env.docker.example"
DEFAULT_JWT_PLACEHOLDER="change-this-to-a-long-random-string"
COMPOSE_CMD=()

info() {
  printf '[setup] %s\n' "$1"
}

warn() {
  printf '[setup] WARNING: %s\n' "$1" >&2
}

fail() {
  printf '[setup] ERROR: %s\n' "$1" >&2
  exit 1
}

need_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "$1 is not installed or not in PATH"
  fi
}

detect_compose() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD=(docker compose)
    return
  fi

  if command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD=(docker-compose)
    return
  fi

  fail "neither 'docker compose' nor 'docker-compose' is available"
}

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
    return
  fi

  LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64
  printf '\n'
}

set_env_var() {
  local file="$1"
  local key="$2"
  local value="$3"
  local tmp

  tmp="$(mktemp)"
  awk -F= -v k="$key" -v v="$value" '
    BEGIN { found = 0 }
    $1 == k { print k "=" v; found = 1; next }
    { print }
    END { if (!found) print k "=" v }
  ' "$file" >"$tmp"
  mv "$tmp" "$file"
}

prepare_env_file() {
  if [[ ! -f "$DOCKER_ENV_EXAMPLE" ]]; then
    fail "missing ${DOCKER_ENV_EXAMPLE}"
  fi

  if [[ ! -f "$DOCKER_ENV_FILE" ]]; then
    cp "$DOCKER_ENV_EXAMPLE" "$DOCKER_ENV_FILE"
    info "created ${DOCKER_ENV_FILE} from ${DOCKER_ENV_EXAMPLE}"
  else
    info "${DOCKER_ENV_FILE} already exists"
  fi

  local current_secret
  current_secret="$(grep -E '^JWT_SECRET=' "$DOCKER_ENV_FILE" | sed 's/^JWT_SECRET=//' || true)"
  if [[ -z "$current_secret" || "$current_secret" == "$DEFAULT_JWT_PLACEHOLDER" ]]; then
    set_env_var "$DOCKER_ENV_FILE" "JWT_SECRET" "$(generate_secret)"
    info "generated JWT_SECRET in ${DOCKER_ENV_FILE}"
  fi
}

check_docker() {
  detect_compose

  if command -v docker >/dev/null 2>&1 && ! docker info >/dev/null 2>&1; then
    fail "docker daemon is not running"
  fi
}

main() {
  cd "$(dirname "$0")/.."

  info "setting up ${APP_NAME} with Docker"
  check_docker
  prepare_env_file

  info "validating compose config"
  "${COMPOSE_CMD[@]}" --env-file "$DOCKER_ENV_FILE" config >/dev/null

  info "setup complete"
  info "run the backend with: make run-env"
}

main "$@"
