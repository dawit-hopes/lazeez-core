#!/bin/sh
# Apply SQL migrations in order. Safe to re-run (uses IF NOT EXISTS / IF EXISTS).
set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/scripts/migrations}"

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "No migrations directory at $MIGRATIONS_DIR, skipping."
  exit 0
fi

echo "Running database migrations from $MIGRATIONS_DIR ..."

for f in "$MIGRATIONS_DIR"/*.sql; do
  [ -f "$f" ] || continue
  echo "  -> $(basename "$f")"
  psql -v ON_ERROR_STOP=1 \
    -h "${DB_HOST:-localhost}" \
    -p "${DB_PORT:-5432}" \
    -U "${DB_USER:-lazeez}" \
    -d "${DB_NAME:-lazeez}" \
    -f "$f"
done

echo "Migrations complete."
