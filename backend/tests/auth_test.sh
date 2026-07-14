#!/bin/bash
# ============================================
# TaskFlow Auth API - Manual Test Script (curl)
# ============================================
# Usage:
#   1. Start the backend server: cd backend && go run cmd/server/main.go
#   2. Run this script: bash tests/auth_test.sh
#
# Requirements: curl, jq (optional, for pretty JSON)
# ============================================

BASE_URL="http://localhost:3000/api/v1"
PASS=0
FAIL=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Generate unique email
EMAIL="curl_test_$(date +%s)@test.com"
PASSWORD="TestPass123"
NAME="Curl Test User"

echo -e "${YELLOW}============================================${NC}"
echo -e "${YELLOW}  TaskFlow Auth API - Curl Test Suite${NC}"
echo -e "${YELLOW}============================================${NC}"
echo ""

# Helper function
check_status() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"

    if [ "$actual" -eq "$expected" ]; then
        echo -e "  ${GREEN}[PASS]${NC} $test_name (HTTP $actual)"
        PASS=$((PASS + 1))
    else
        echo -e "  ${RED}[FAIL]${NC} $test_name (Expected $expected, Got $actual)"
        FAIL=$((FAIL + 1))
    fi
}

# ============================================
# TEST 1: Health Check
# ============================================
echo -e "${YELLOW}--- Health Check ---${NC}"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check_status "GET /health" 200 "$STATUS"

# ============================================
# TEST 2: Register - Success
# ============================================
echo -e "\n${YELLOW}--- Register ---${NC}"
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -c /tmp/taskflow_cookies.txt \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"$NAME\"}")

REGISTER_STATUS=$(echo "$REGISTER_RESPONSE" | tail -1)
REGISTER_BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')
check_status "POST /auth/register (valid)" 201 "$REGISTER_STATUS"
echo "  Response: $REGISTER_BODY"

# Extract access token
ACCESS_TOKEN=$(echo "$REGISTER_BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
echo "  Token: ${ACCESS_TOKEN:0:30}..."

# TEST 3: Register - Duplicate Email
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"$NAME\"}")
check_status "POST /auth/register (duplicate)" 409 "$STATUS"

# TEST 4: Register - Invalid Email
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"not-an-email","password":"TestPass123","name":"Test"}')
check_status "POST /auth/register (invalid email)" 422 "$STATUS"

# TEST 5: Register - Short Password
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"short@test.com","password":"short","name":"Test"}')
check_status "POST /auth/register (short password)" 422 "$STATUS"

# TEST 6: Register - Missing Fields
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{}')
check_status "POST /auth/register (empty body)" 422 "$STATUS"

# ============================================
# TEST 7: Login - Success
# ============================================
echo -e "\n${YELLOW}--- Login ---${NC}"
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -c /tmp/taskflow_cookies.txt \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

LOGIN_STATUS=$(echo "$LOGIN_RESPONSE" | tail -1)
LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')
check_status "POST /auth/login (valid)" 200 "$LOGIN_STATUS"

# Extract new access token
ACCESS_TOKEN=$(echo "$LOGIN_BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

# TEST 8: Login - Wrong Password
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"WrongPassword\"}")
check_status "POST /auth/login (wrong password)" 401 "$STATUS"

# TEST 9: Login - Non-existent Email
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"nonexistent@test.com","password":"TestPass123"}')
check_status "POST /auth/login (non-existent)" 401 "$STATUS"

# ============================================
# TEST 10: Me (Protected) - Success
# ============================================
echo -e "\n${YELLOW}--- Me (Protected) ---${NC}"
ME_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X GET "$BASE_URL/auth/me" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

ME_STATUS=$(echo "$ME_RESPONSE" | tail -1)
ME_BODY=$(echo "$ME_RESPONSE" | sed '$d')
check_status "GET /auth/me (valid token)" 200 "$ME_STATUS"
echo "  Response: $ME_BODY"

# TEST 11: Me - No Token
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "$BASE_URL/auth/me")
check_status "GET /auth/me (no token)" 401 "$STATUS"

# TEST 12: Me - Invalid Token
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "$BASE_URL/auth/me" \
    -H "Authorization: Bearer invalid-token")
check_status "GET /auth/me (invalid token)" 401 "$STATUS"

# ============================================
# TEST 13: Refresh Token
# ============================================
echo -e "\n${YELLOW}--- Refresh Token ---${NC}"
REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/auth/refresh" \
    -H "Content-Type: application/json" \
    -b /tmp/taskflow_cookies.txt \
    -c /tmp/taskflow_cookies.txt)

REFRESH_STATUS=$(echo "$REFRESH_RESPONSE" | tail -1)
REFRESH_BODY=$(echo "$REFRESH_RESPONSE" | sed '$d')
check_status "POST /auth/refresh (valid cookie)" 200 "$REFRESH_STATUS"
echo "  Response: $REFRESH_BODY"

# TEST 14: Refresh - No Cookie
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/refresh" \
    -H "Content-Type: application/json")
check_status "POST /auth/refresh (no cookie)" 401 "$STATUS"

# ============================================
# TEST 15: Logout
# ============================================
echo -e "\n${YELLOW}--- Logout ---${NC}"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/logout" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
check_status "POST /auth/logout (valid token)" 200 "$STATUS"

# TEST 16: Logout - No Token
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/auth/logout")
check_status "POST /auth/logout (no token)" 401 "$STATUS"

# ============================================
# TEST 17: 404 - Endpoint Not Found
# ============================================
echo -e "\n${YELLOW}--- 404 ---${NC}"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "$BASE_URL/nonexistent")
check_status "GET /nonexistent (404)" 404 "$STATUS"

# ============================================
# RESULTS
# ============================================
echo ""
echo -e "${YELLOW}============================================${NC}"
TOTAL=$((PASS + FAIL))
echo -e "  Results: ${GREEN}$PASS passed${NC} / ${RED}$FAIL failed${NC} / $TOTAL total"
echo -e "${YELLOW}============================================${NC}"

# Cleanup
rm -f /tmp/taskflow_cookies.txt

if [ "$FAIL" -eq 0 ]; then
    echo -e "  ${GREEN}ALL TESTS PASSED!${NC}"
    exit 0
else
    echo -e "  ${RED}SOME TESTS FAILED${NC}"
    exit 1
fi
