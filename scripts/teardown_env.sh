#!/usr/bin/env bash
set -euo pipefail

APP_NAME="meal_back"
DOCKER_ENV_FILE=".env.docker"
COMPOSE_CMD=()

REMOVE_ENV=0
REMOVE_VOLUMES=1
ASSUME_YES=0

usage() {
  cat <<'EOF'
Usage: ./scripts/teardown_env.sh [options]

Tear down meal_back Docker runtime.

Options:
  --yes         Skip interactive confirmation
  --keep-db     Keep Docker volume data (do not pass --volumes to compose down)
  --remove-env  Remove .env.docker after teardown
  -h, --help    Show this help
EOF
}

info() {
  printf '[teardown] %s\n' "$1"
}

fail() {
  printf '[teardown] ERROR: %s\n' "$1" >&2
  exit 1
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

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --yes)
        ASSUME_YES=1
        ;;
      --keep-db)
        REMOVE_VOLUMES=0
        ;;
      --remove-env)
        REMOVE_ENV=1
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "unknown option: $1"
        ;;
    esac
    shift
  done
}

confirm() {
  if [[ "$ASSUME_YES" -eq 1 ]]; then
    return
  fi

  if [[ ! -t 0 ]]; then
    fail "non-interactive shell detected; use --yes to continue"
  fi

  printf '[teardown] This will stop %s containers. Continue? [y/N]: ' "$APP_NAME"
  read -r ans
  case "$ans" in
    y|Y|yes|YES)
      ;;
    *)
      info "aborted"
      exit 0
      ;;
  esac
}

run_compose_down() {
  detect_compose

  local cmd
  cmd=("${COMPOSE_CMD[@]}" down --remove-orphans)
  if [[ "$REMOVE_VOLUMES" -eq 1 ]]; then
    cmd+=(--volumes)
  fi

  if [[ -f "$DOCKER_ENV_FILE" ]]; then
    cmd=("${COMPOSE_CMD[@]}" --env-file "$DOCKER_ENV_FILE" down --remove-orphans)
    if [[ "$REMOVE_VOLUMES" -eq 1 ]]; then
      cmd+=(--volumes)
    fi
  fi

  "${cmd[@]}"
}

remove_env_file() {
  if [[ "$REMOVE_ENV" -eq 1 && -f "$DOCKER_ENV_FILE" ]]; then
    rm -f "$DOCKER_ENV_FILE"
    info "removed ${DOCKER_ENV_FILE}"
  fi
}

main() {
  cd "$(dirname "$0")/.."
  parse_args "$@"
  confirm
  run_compose_down
  remove_env_file
  info "teardown complete"
}

main "$@"
