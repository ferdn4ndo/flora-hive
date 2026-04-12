#!/bin/sh
set -e
cd /app

case "${SKIP_CONTAINER_PREPARE:-}" in
  1|true|TRUE|yes|YES) ;;
  *)
    ./flora-hive migrate:up
    ;;
esac

exec "$@"
