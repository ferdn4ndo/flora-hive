#!/bin/sh
set -e
cd /app

case "${SKIP_CONTAINER_PREPARE:-}" in
  1|true|TRUE|yes|YES) ;;
  *)
    node dist/containerPrepare.js
    ;;
esac

exec "$@"
