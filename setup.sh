#!/bin/sh
# Local / CI bootstrap: optional Postgres role+DB, migrations, then uServer-Auth HTTP bootstrap.
# Mirrors userver-filemgr setup.sh (MIGRATE_BIN, migrate:up, bootstrap:auth; auth failure is non-fatal).
set -eu
cd "$(dirname "$0")"

MIGRATE_BIN="${MIGRATE_BIN:-./bin/flora-hive}"

if [ -f .env ]; then
	set -a
	# shellcheck disable=SC1091
	. ./.env
	set +a
fi

if [ ! -f "$MIGRATE_BIN" ]; then
	echo "setup.sh: binary not found at $MIGRATE_BIN — run: make build" >&2
	exit 1
fi

scripts/docker-bootstrap-postgres.sh
"$MIGRATE_BIN" migrate:up

if [ "${SKIP_AUTH_BOOTSTRAP:-}" = "1" ] || [ "${SKIP_USERVER_AUTH_SETUP:-}" = "1" ]; then
	echo "setup.sh: skipping bootstrap:auth (SKIP_AUTH_BOOTSTRAP / SKIP_USERVER_AUTH_SETUP)"
else
	if ! "$MIGRATE_BIN" bootstrap:auth; then
		echo "setup.sh: bootstrap:auth failed or incomplete — set USERVER_AUTH_* / SYSTEM_CREATION_TOKEN or SKIP_USERVER_AUTH_SETUP=1" >&2
	fi
fi

echo "setup.sh: done"
