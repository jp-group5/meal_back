#!/usr/bin/env bash
set -euo pipefail

USAGE="Usage: DB_DSN='<dsn>' ./scripts/cleanup_all_users.sh --yes\n\nThis will permanently remove all rows from users, user_profiles, and user_sessions."

if [[ "${1:-}" != "--yes" ]]; then
  echo "$USAGE"
  exit 1
fi

if [[ -z "${DB_DSN:-}" ]]; then
  echo "ERROR: DB_DSN is not set"
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "ERROR: psql is not installed"
  exit 1
fi

echo "[cleanup-users] counting users before cleanup..."
BEFORE_COUNT=$(psql "$DB_DSN" -Atqc "SELECT COUNT(*) FROM users;")
echo "[cleanup-users] users before: ${BEFORE_COUNT}"

echo "[cleanup-users] truncating users/user_profiles/user_sessions..."
psql "$DB_DSN" -v ON_ERROR_STOP=1 -qc "TRUNCATE TABLE user_sessions, user_profiles, users RESTART IDENTITY CASCADE;"

echo "[cleanup-users] counting users after cleanup..."
AFTER_COUNT=$(psql "$DB_DSN" -Atqc "SELECT COUNT(*) FROM users;")
echo "[cleanup-users] users after: ${AFTER_COUNT}"

echo "[cleanup-users] done"
