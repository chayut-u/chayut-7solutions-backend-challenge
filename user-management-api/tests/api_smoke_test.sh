#!/usr/bin/env bash
# Deep QA regression suite for the User Management API.
# Usage: bash tests/api_smoke_test.sh
# Base URL can be overridden: BASE_URL=http://localhost:8080 bash tests/api_smoke_test.sh
#
# Requires: curl, and (optionally) jq for prettier diagnostics (not required).

set -u
BASE_URL="${BASE_URL:-http://localhost:8080}"

PASS=0
FAIL=0
FAILED_TESTS=()

# run <name> <expected_status> <curl args...>
# The LAST arg group after expected_status is passed straight to curl.
run() {
  local name="$1" expected="$2"; shift 2
  local body status
  body=$(curl -s -o - -w "\n%{http_code}" "$@")
  status=$(echo "$body" | tail -n1)
  body=$(echo "$body" | sed '$d')

  if [ "$status" = "$expected" ]; then
    echo "[PASS] $name -> $status"
    PASS=$((PASS+1))
  else
    echo "[FAIL] $name -> expected $expected got $status"
    echo "       body: $body"
    FAIL=$((FAIL+1))
    FAILED_TESTS+=("$name (expected $expected got $status): $body")
  fi
  LAST_BODY="$body"
  LAST_STATUS="$status"
}

echo "=== Base URL: $BASE_URL ==="
TS=$(date +%s)

# ---------------------------------------------------------------------------
# 1. Registration validation edge cases
# ---------------------------------------------------------------------------
echo
echo "--- Registration validation ---"

run "register invalid email format" 422 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Bad Email\",\"email\":\"not-an-email\",\"password\":\"secret123\"}"

run "register password 5 chars (too short)" 422 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Short Pw\",\"email\":\"shortpw_$TS@example.com\",\"password\":\"abcde\"}"

run "register password exactly 6 chars (valid)" 201 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Six Char Pw\",\"email\":\"sixcharpw_$TS@example.com\",\"password\":\"abcdef\"}"

run "register empty name" 400 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"\",\"email\":\"emptyname_$TS@example.com\",\"password\":\"secret123\"}"

run "register missing all fields" 400 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{}"

run "register malformed JSON" 400 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{name: broken"

run "register no body at all" 400 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json"

run "register whitespace-only name" 201 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"   \",\"email\":\"wsname_$TS@example.com\",\"password\":\"secret123\"}"
WHITESPACE_NAME_BODY="$LAST_BODY"
WHITESPACE_NAME_STATUS="$LAST_STATUS"

# ---------------------------------------------------------------------------
# 2. Login validation edge cases
# ---------------------------------------------------------------------------
echo
echo "--- Login validation ---"

run "login malformed JSON" 400 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{email: broken"

run "login missing fields" 400 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{}"

run "login unknown email" 401 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"doesnotexist_$TS@example.com\",\"password\":\"whatever1\"}"

# ---------------------------------------------------------------------------
# Set up two real users for the rest of the suite
# ---------------------------------------------------------------------------
echo
echo "--- Setup: register userA / userB ---"

EMAIL_A="usera_$TS@example.com"
EMAIL_B="userb_$TS@example.com"
PASSWORD="password123"

run "register userA" 201 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"User A\",\"email\":\"$EMAIL_A\",\"password\":\"$PASSWORD\"}"
USER_A_ID=$(echo "$LAST_BODY" | grep -o '"id":"[a-f0-9]*"' | head -1 | cut -d'"' -f4)

run "register userB" 201 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"User B\",\"email\":\"$EMAIL_B\",\"password\":\"$PASSWORD\"}"
USER_B_ID=$(echo "$LAST_BODY" | grep -o '"id":"[a-f0-9]*"' | head -1 | cut -d'"' -f4)

echo "userA id=$USER_A_ID userB id=$USER_B_ID"

run "register duplicate email (userA again)" 409 -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"User A Dup\",\"email\":\"$EMAIL_A\",\"password\":\"$PASSWORD\"}"

# Informational only (spec is silent on case-folding) - not scored pass/fail.
CASE_FOLD_RESP=$(curl -s -o - -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Case Test\",\"email\":\"$(echo "$EMAIL_A" | sed 's/.*/\U&/')\",\"password\":\"$PASSWORD\"}")
CASE_FOLD_STATUS=$(echo "$CASE_FOLD_RESP" | tail -n1)
CASE_FOLD_BODY=$(echo "$CASE_FOLD_RESP" | sed '$d')
echo "[INFO] register different case of existing email -> $CASE_FOLD_STATUS"

run "login userA correct password" 200 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL_A\",\"password\":\"$PASSWORD\"}"
TOKEN_A=$(echo "$LAST_BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

run "login userA wrong password" 401 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL_A\",\"password\":\"wrongpassword\"}"
WRONG_PW_MSG="$LAST_BODY"

run "login unknown email (compare msg)" 401 -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"nobody_$TS@example.com\",\"password\":\"whatever1\"}"
UNKNOWN_EMAIL_MSG="$LAST_BODY"

if [ "$WRONG_PW_MSG" = "$UNKNOWN_EMAIL_MSG" ]; then
  echo "[PASS] anti-enumeration: wrong-password and unknown-email messages are identical"
  PASS=$((PASS+1))
else
  echo "[FAIL] anti-enumeration: messages differ!"
  echo "       wrong-password: $WRONG_PW_MSG"
  echo "       unknown-email : $UNKNOWN_EMAIL_MSG"
  FAIL=$((FAIL+1))
  FAILED_TESTS+=("anti-enumeration message mismatch")
fi

AUTH_A=(-H "Authorization: Bearer $TOKEN_A")

# ---------------------------------------------------------------------------
# 3. Auth / JWT edge cases
# ---------------------------------------------------------------------------
echo
echo "--- Auth / JWT edge cases ---"

run "protected endpoint no token" 401 "$BASE_URL/api/users"

run "protected endpoint garbage bearer token" 401 "$BASE_URL/api/users" \
  -H "Authorization: Bearer not-a-real-jwt-at-all"

run "protected endpoint missing Bearer prefix" 401 "$BASE_URL/api/users" \
  -H "Authorization: $TOKEN_A"

run "protected endpoint empty bearer" 401 "$BASE_URL/api/users" \
  -H "Authorization: Bearer "

# Mutate the signature of a real token
MUTATED_TOKEN="${TOKEN_A%??}XX"
run "protected endpoint tampered signature" 401 "$BASE_URL/api/users" \
  -H "Authorization: Bearer $MUTATED_TOKEN"

# alg=none attack: header {"alg":"none","typ":"JWT"} + same payload + empty sig
HEADER_NONE=$(echo -n '{"alg":"none","typ":"JWT"}' | base64 | tr -d '=' | tr '/+' '_-')
PAYLOAD=$(echo "$TOKEN_A" | cut -d. -f2)
ALG_NONE_TOKEN="${HEADER_NONE}.${PAYLOAD}."
run "protected endpoint alg=none token" 401 "$BASE_URL/api/users" \
  -H "Authorization: Bearer $ALG_NONE_TOKEN"

# ---------------------------------------------------------------------------
# 4. ObjectID edge cases
# ---------------------------------------------------------------------------
echo
echo "--- ObjectID edge cases ---"

run "GET user with invalid objectid" 400 "$BASE_URL/api/users/not-a-valid-objectid" "${AUTH_A[@]}"

run "GET user very short id" 400 "$BASE_URL/api/users/123" "${AUTH_A[@]}"

run "GET user very long garbage id" 400 "$BASE_URL/api/users/$(printf 'a%.0s' {1..50})" "${AUTH_A[@]}"

run "GET users trailing slash no id" 404 "$BASE_URL/api/users/" "${AUTH_A[@]}"

run "GET user well-formed but nonexistent objectid" 404 "$BASE_URL/api/users/507f1f77bcf86cd799439011" "${AUTH_A[@]}"

# ---------------------------------------------------------------------------
# 5. List / Get happy path shape checks
# ---------------------------------------------------------------------------
echo
echo "--- Response shape checks ---"

run "GET users list" 200 "$BASE_URL/api/users" "${AUTH_A[@]}"
if echo "$LAST_BODY" | grep -q '"password"'; then
  echo "[FAIL] password field leaked in list response!"
  FAIL=$((FAIL+1))
  FAILED_TESTS+=("password leaked in GET /api/users")
else
  echo "[PASS] no password field in list response"
  PASS=$((PASS+1))
fi

run "GET userA by id" 200 "$BASE_URL/api/users/$USER_A_ID" "${AUTH_A[@]}"
if echo "$LAST_BODY" | grep -q '"password"'; then
  echo "[FAIL] password field leaked in get-by-id response!"
  FAIL=$((FAIL+1))
  FAILED_TESTS+=("password leaked in GET /api/users/:id")
else
  echo "[PASS] no password field in get-by-id response"
  PASS=$((PASS+1))
fi

# ---------------------------------------------------------------------------
# 6. Update business logic
# ---------------------------------------------------------------------------
echo
echo "--- Update business logic ---"

run "PUT no fields at all" 400 -X PUT "$BASE_URL/api/users/$USER_A_ID" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{}"

run "PUT userA email to userB's email (conflict)" 409 -X PUT "$BASE_URL/api/users/$USER_A_ID" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"email\":\"$EMAIL_B\"}"

run "PUT userA email to itself (should be allowed)" 200 -X PUT "$BASE_URL/api/users/$USER_A_ID" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"email\":\"$EMAIL_A\"}"

run "PUT userA name to empty string" 400 -X PUT "$BASE_URL/api/users/$USER_A_ID" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"name\":\"\"}"

run "PUT nonexistent user" 404 -X PUT "$BASE_URL/api/users/507f1f77bcf86cd799439011" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"name\":\"Ghost\"}"

run "PUT invalid objectid" 400 -X PUT "$BASE_URL/api/users/garbage-id" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"name\":\"Ghost\"}"

# ---------------------------------------------------------------------------
# 7. Delete business logic
# ---------------------------------------------------------------------------
echo
echo "--- Delete business logic ---"

run "DELETE userB" 200 -X DELETE "$BASE_URL/api/users/$USER_B_ID" "${AUTH_A[@]}"

run "DELETE userB again (already deleted)" 404 -X DELETE "$BASE_URL/api/users/$USER_B_ID" "${AUTH_A[@]}"

run "PUT deleted userB" 404 -X PUT "$BASE_URL/api/users/$USER_B_ID" "${AUTH_A[@]}" \
  -H "Content-Type: application/json" -d "{\"name\":\"Ghost\"}"

run "GET deleted userB" 404 "$BASE_URL/api/users/$USER_B_ID" "${AUTH_A[@]}"

run "DELETE invalid objectid" 400 -X DELETE "$BASE_URL/api/users/garbage-id" "${AUTH_A[@]}"

# ---------------------------------------------------------------------------
# 8. Concurrent duplicate registration
# ---------------------------------------------------------------------------
echo
echo "--- Concurrency: duplicate registration race ---"

RACE_EMAIL="race_$TS@example.com"
tmp1=$(mktemp)
tmp2=$(mktemp)

curl -s -o "$tmp1" -w "%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Racer 1\",\"email\":\"$RACE_EMAIL\",\"password\":\"password123\"}" > "$tmp1.status" &
PID1=$!
curl -s -o "$tmp2" -w "%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Racer 2\",\"email\":\"$RACE_EMAIL\",\"password\":\"password123\"}" > "$tmp2.status" &
PID2=$!
wait $PID1 $PID2

STATUS1=$(cat "$tmp1.status")
STATUS2=$(cat "$tmp2.status")
BODY1=$(cat "$tmp1")
BODY2=$(cat "$tmp2")

echo "Racer1: $STATUS1 / $BODY1"
echo "Racer2: $STATUS2 / $BODY2"

SUCCESS_COUNT=0
[ "$STATUS1" = "201" ] && SUCCESS_COUNT=$((SUCCESS_COUNT+1))
[ "$STATUS2" = "201" ] && SUCCESS_COUNT=$((SUCCESS_COUNT+1))

OTHER_STATUS=""
[ "$STATUS1" != "201" ] && OTHER_STATUS="$STATUS1"
[ "$STATUS2" != "201" ] && OTHER_STATUS="$STATUS2"

if [ "$SUCCESS_COUNT" = "1" ] && [ "$OTHER_STATUS" = "409" ]; then
  echo "[PASS] concurrent duplicate registration: exactly one 201, other 409"
  PASS=$((PASS+1))
elif [ "$SUCCESS_COUNT" = "1" ]; then
  echo "[FAIL] concurrent duplicate registration: one succeeded (201) but loser returned $OTHER_STATUS instead of 409"
  FAIL=$((FAIL+1))
  FAILED_TESTS+=("race condition: loser got $OTHER_STATUS not 409 - body: $([ "$STATUS1" != "201" ] && echo "$BODY1" || echo "$BODY2")")
else
  echo "[FAIL] concurrent duplicate registration: expected exactly one 201, got statuses $STATUS1 / $STATUS2"
  FAIL=$((FAIL+1))
  FAILED_TESTS+=("race condition: unexpected statuses $STATUS1 / $STATUS2")
fi
rm -f "$tmp1" "$tmp2" "$tmp1.status" "$tmp2.status"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo
echo "=== Deferred findings (not pass/fail, informational) ==="
echo "Whitespace-only name register: status=$WHITESPACE_NAME_STATUS body=$WHITESPACE_NAME_BODY"
echo "Case-different duplicate email register: status=$CASE_FOLD_STATUS body=$CASE_FOLD_BODY"

echo
echo "=== SUMMARY: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  echo "Failing tests:"
  for t in "${FAILED_TESTS[@]}"; do
    echo "  - $t"
  done
fi

exit 0
