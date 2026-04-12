#!/bin/sh
# Optional bootstrap: create POSTGRES_DB + POSTGRES_USER using a superuser (same idea as userver-filemgr setup.sh).
# Requires: POSTGRES_ROOT_USER, POSTGRES_ROOT_PASS, POSTGRES_HOST, POSTGRES_DB, POSTGRES_USER,
#           POSTGRES_PASS or POSTGRES_PASSWORD.
# Uses POSTGRES_ADMIN_DATABASE (default: postgres) and POSTGRES_SSLMODE (default: disable).

set -eu

case "${SKIP_CONTAINER_PREPARE:-}" in
1 | true | TRUE | yes | YES) exit 0 ;;
esac

if [ -z "${POSTGRES_ROOT_USER:-}" ] || [ -z "${POSTGRES_ROOT_PASS:-}" ]; then
	exit 0
fi

if [ -z "${POSTGRES_HOST:-}" ] || [ -z "${POSTGRES_DB:-}" ] || [ -z "${POSTGRES_USER:-}" ]; then
	echo "docker-bootstrap-postgres: POSTGRES_HOST, POSTGRES_DB, POSTGRES_USER are required when POSTGRES_ROOT_* is set" >&2
	exit 1
fi

app_pass="${POSTGRES_PASS:-}"
if [ -z "$app_pass" ] && [ -n "${POSTGRES_PASSWORD:-}" ]; then
	app_pass="$POSTGRES_PASSWORD"
fi
if [ -z "$app_pass" ]; then
	echo "docker-bootstrap-postgres: POSTGRES_PASS or POSTGRES_PASSWORD is required when POSTGRES_ROOT_* is set" >&2
	exit 1
fi

admin_db="${POSTGRES_ADMIN_DATABASE:-postgres}"
port="${POSTGRES_PORT:-5432}"
sslmode="${POSTGRES_SSLMODE:-disable}"

export PGPASSWORD="$POSTGRES_ROOT_PASS"
export PGSSLMODE="$sslmode"

wait_secs=90
i=0
while ! pg_isready -h "$POSTGRES_HOST" -p "$port" -U "$POSTGRES_ROOT_USER" -d "$admin_db" -q 2>/dev/null; do
	i=$((i + 1))
	if [ "$i" -ge "$wait_secs" ]; then
		echo "docker-bootstrap-postgres: timed out waiting for PostgreSQL at ${POSTGRES_HOST}:${port}" >&2
		exit 1
	fi
	sleep 1
done

# Create role and database if missing. Use psql :'var' (string literal) for format(%I/%L) args —
# :"var" is a quoted identifier and is parsed as a column name here, which breaks.
psql -v ON_ERROR_STOP=1 \
	-h "$POSTGRES_HOST" \
	-p "$port" \
	-U "$POSTGRES_ROOT_USER" \
	-d "$admin_db" \
	-v "app_db=$POSTGRES_DB" \
	-v "app_user=$POSTGRES_USER" \
	-v "app_password=$app_pass" <<'EOSQL'
SELECT format(
	'CREATE ROLE %I WITH LOGIN PASSWORD %L',
	:'app_user',
	:'app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user')\gexec

SELECT format(
	'CREATE DATABASE %I OWNER %I ENCODING ''UTF8''',
	:'app_db',
	:'app_user'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'app_db')\gexec
EOSQL

# PostgreSQL 15+: ensure app role can use public schema in the app database.
psql -v ON_ERROR_STOP=1 \
	-h "$POSTGRES_HOST" \
	-p "$port" \
	-U "$POSTGRES_ROOT_USER" \
	-d "$POSTGRES_DB" \
	-v "app_user=$POSTGRES_USER" <<'EOSQL'
SELECT format('GRANT ALL ON SCHEMA public TO %I', :'app_user')
WHERE EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'public')\gexec
EOSQL

unset PGPASSWORD
