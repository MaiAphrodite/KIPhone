#!/usr/bin/env bash
set -u

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
DEV_USER="${DEV_USER:-qa}"
PASS=0
FAIL=0
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 2
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi

req() {
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -sS -o "$TMP/body" -w '%{http_code}' -X "$method" "$BASE_URL$path" \
      -H "X-Dev-User: $DEV_USER" -H 'Content-Type: application/json' --data "$body"
  else
    curl -sS -o "$TMP/body" -w '%{http_code}' -X "$method" "$BASE_URL$path" \
      -H "X-Dev-User: $DEV_USER"
  fi
}

jget() {
  jq -r "$1" "$TMP/body" 2>/dev/null
}

check_eq() {
  local name="$1" want="$2" got="$3"
  if [ "$got" = "$want" ]; then
    printf 'PASS  %s (%s)\n' "$name" "$got"
    PASS=$((PASS + 1))
  else
    printf 'FAIL  %s (want %s, got %s)\n' "$name" "$want" "$got"
    FAIL=$((FAIL + 1))
  fi
}

check_gt() {
  local name="$1" min="$2" got="$3"
  if [ -n "$got" ] && [ "$got" -gt "$min" ] 2>/dev/null; then
    printf 'PASS  %s (%s > %s)\n' "$name" "$got" "$min"
    PASS=$((PASS + 1))
  else
    printf 'FAIL  %s (want > %s, got %s)\n' "$name" "$min" "$got"
    FAIL=$((FAIL + 1))
  fi
}

code="$(req GET /health)"
check_eq "GET /health" "200" "$code"
check_eq "health.ok" "true" "$(jget '.ok')"

code="$(req POST /api/seed/load)"
check_eq "POST /api/seed/load" "200" "$code"
check_eq "seed.students" "215" "$(jget '.students')"
check_eq "seed.accounts" "215" "$(jget '.accounts')"
check_eq "seed.disbursements" "215" "$(jget '.disbursements')"
check_eq "seed.planted" "15" "$(jget '.planted')"

code="$(req GET /api/stats)"
check_eq "GET /api/stats" "200" "$code"
check_gt "stats.records after seed" "0" "$(jget '.records')"

code="$(req POST /api/detect)"
check_eq "POST /api/detect" "200" "$code"
check_gt "detect.findings_created" "0" "$(jget '.findings_created')"

code="$(req GET /api/findings)"
check_eq "GET /api/findings" "200" "$code"
check_eq "findings rows" "15" "$(jget '.findings | length')"

FINDING_ID="$(jget '.findings[0].id')"
check_eq "first finding status" "baru" "$(jget '.findings[0].status')"

code="$(req POST "/api/findings/$FINDING_ID/status" '{"to_status":"salah_deteksi","note":"qa contract"}')"
check_eq "transition baru->salah_deteksi" "200" "$code"
check_eq "transition records from_status" "baru" "$(jget '.event.from_status')"
check_eq "transition response status" "salah_deteksi" "$(jget '.status')"

code="$(req GET "/api/findings/$FINDING_ID")"
check_eq "GET /api/findings/{id} after transition" "200" "$code"
check_eq "persisted status" "salah_deteksi" "$(jget '.finding.status')"

code="$(req POST "/api/findings/$FINDING_ID/status" '{"to_status":"baru"}')"
check_eq "invalid transition salah_deteksi->baru" "409" "$code"

echo
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED: $FAIL check(s), $PASS passed"
  exit 1
fi
echo "PASSED: $PASS check(s)"
