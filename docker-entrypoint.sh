#!/bin/sh
set -e
cd /app

case "${SKIP_CONTAINER_PREPARE:-}" in
  1|true|TRUE|yes|YES) ;;
  *)
    /usr/local/bin/docker-bootstrap-postgres.sh
    ./flora-hive migrate:up
    # Same pattern as userver-filemgr setup.sh (optional; non-fatal).
    if ! ./flora-hive bootstrap:auth; then
      echo "flora-hive: bootstrap:auth failed or incomplete — set USERVER_AUTH_* / SYSTEM_CREATION_TOKEN or SKIP_USERVER_AUTH_SETUP=1" >&2
    fi
    ;;
esac

exec "$@"
