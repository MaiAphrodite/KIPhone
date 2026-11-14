#!/bin/sh
# Start the whole KIPhone demo locally: Postgres, Go API (:8080), SPA (:8088).
# Usage: sh run-local.sh          then open http://127.0.0.1:8088
set -eu
ROOT="$(cd "$(dirname "$0")" && pwd)"
PORT_API=8080
PORT_WEB=8088

say() { printf '\033[2m%s\033[0m\n' "$*"; }

# 1) database
if ! docker ps --format '{{.Names}}' | grep -q '^hackstack-db-1$'; then
  say "starting postgres…"
  (cd "$ROOT" && docker compose up -d >/dev/null)
  sleep 4
fi

# 2) api
if ! curl -sf -m 2 "http://127.0.0.1:$PORT_API/health" >/dev/null 2>&1; then
  say "starting go api on :$PORT_API…"
  (cd "$ROOT/backend" && DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5432/hackstack" \
    PORT=$PORT_API setsid ./backend >/tmp/kiphone-api.log 2>&1 < /dev/null &)
  sleep 3
fi

# 3) seed + detect if the batch is empty
RECORDS=$(curl -s -m 5 "http://127.0.0.1:$PORT_API/api/stats" -H 'X-Dev-User: admin' | sed -n 's/.*"records":\([0-9]*\).*/\1/p')
if [ "${RECORDS:-0}" = "0" ]; then
  say "seeding 215 synthetic records…"
  curl -s -X POST "http://127.0.0.1:$PORT_API/api/seed/load" -H 'X-Dev-User: admin' >/dev/null
  curl -s -X POST "http://127.0.0.1:$PORT_API/api/detect" -H 'X-Dev-User: admin' >/dev/null
fi

# 3b) seed the dummy partner-bank feed if it is empty
TXNS=$(curl -s -m 5 "http://127.0.0.1:$PORT_API/api/tx/stats" -H 'X-Dev-User: admin' | sed -n 's/.*"transactions":\([0-9]*\).*/\1/p')
if [ "${TXNS:-0}" = "0" ]; then
  say "seeding dummy bank transactions…"
  curl -s -X POST "http://127.0.0.1:$PORT_API/api/tx/seed" -H 'X-Dev-User: admin' >/dev/null
  curl -s -X POST "http://127.0.0.1:$PORT_API/api/tx/detect" -H 'X-Dev-User: admin' >/dev/null
fi

# 4) web (production build + preview server that proxies /api)
say "building spa…"
(cd "$ROOT/frontend" && npm run build >/dev/null 2>&1)

if ! curl -sf -m 2 "http://127.0.0.1:$PORT_WEB/" >/dev/null 2>&1; then
  say "starting preview server on :$PORT_WEB…"
  (cd "$ROOT/frontend" && setsid npx vite preview >/tmp/kiphone-web.log 2>&1 < /dev/null &)
  sleep 4
fi

printf '\n  KIPhone  →  http://127.0.0.1:%s\n  API      →  http://127.0.0.1:%s/api/stats\n\n' "$PORT_WEB" "$PORT_API"
