#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api}"
EMAIL="${E2E_EMAIL:?set E2E_EMAIL}"
PASSWORD="${E2E_PASSWORD:?set E2E_PASSWORD}"
COOKIE_JAR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR"' EXIT

login_payload="$(printf '{"email":"%s","password":"%s"}' "$EMAIL" "$PASSWORD")"
curl -fsS -c "$COOKIE_JAR" \
  -H "Content-Type: application/json" \
  -d "$login_payload" \
  "$API_BASE_URL/auth/login" >/dev/null

submission_payload='{"problem_slug":"echo-input","language":"python","source_code":"print(input())\n","kind":"submit"}'
submission_id="$(
  curl -fsS -b "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    -d "$submission_payload" \
    "$API_BASE_URL/submissions" |
    python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["submission_id"])'
)"

for _ in $(seq 1 20); do
  response="$(curl -fsS -b "$COOKIE_JAR" "$API_BASE_URL/submissions/$submission_id")"
  status="$(printf '%s' "$response" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["status"])')"
  verdict="$(printf '%s' "$response" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["verdict"])')"

  if [ "$status" = "done" ]; then
    if [ "$verdict" = "AC" ]; then
      echo "submission $submission_id accepted"
      exit 0
    fi
    echo "submission $submission_id finished with verdict $verdict" >&2
    exit 1
  fi

  sleep 1
done

echo "submission $submission_id did not finish in time" >&2
exit 1
