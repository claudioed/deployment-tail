#!/bin/bash

# Smoke Test Script for Post-Deployment Validation
#
# This script validates that authentication and core functionality
# are working correctly after deployment.
#
# Usage:
#   export API_URL=http://localhost:8080
#   export ADMIN_TOKEN=your-jwt-token
#   ./scripts/smoke_test.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
VERBOSE="${VERBOSE:-false}"

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Helper functions
print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_test() {
    echo -e "${YELLOW}▶ Test $((TOTAL_TESTS + 1)): $1${NC}"
}

pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${GREEN}  ✓ PASS${NC}"
}

fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${RED}  ✗ FAIL: $1${NC}"
}

make_request() {
    local method=$1
    local endpoint=$2
    local auth_header=$3
    local data=$4

    local curl_opts="-s -w \n%{http_code}"

    if [ "$VERBOSE" = "true" ]; then
        curl_opts="-v $curl_opts"
    fi

    if [ -n "$auth_header" ]; then
        curl_opts="$curl_opts -H 'Authorization: Bearer $auth_header'"
    fi

    if [ -n "$data" ]; then
        curl_opts="$curl_opts -H 'Content-Type: application/json' -d '$data'"
    fi

    eval curl $curl_opts -X "$method" "$API_URL$endpoint"
}

# Start tests
print_header "Deployment Tail - Smoke Test Suite"

echo "Configuration:"
echo "  API URL: $API_URL"
echo "  Admin Token: $([ -n "$ADMIN_TOKEN" ] && echo 'Provided' || echo 'Not provided')"
echo ""

# Test 1: Health Check
print_header "Basic Health Checks"

print_test "Health endpoint is accessible"
response=$(make_request GET /health "")
status_code=$(echo "$response" | tail -n1)
if [ "$status_code" = "200" ]; then
    pass
else
    fail "Expected 200, got $status_code"
fi

# Test 2: Authentication Required
print_test "Unauthenticated requests are rejected"
response=$(make_request GET /api/v1/schedules "")
status_code=$(echo "$response" | tail -n1)
if [ "$status_code" = "401" ]; then
    pass
else
    fail "Expected 401 Unauthorized, got $status_code"
fi

# Test 3: OAuth Login Endpoint
print_test "Google OAuth login endpoint exists"
response=$(make_request GET /auth/google/login "")
status_code=$(echo "$response" | tail -n1)
if [ "$status_code" = "302" ] || [ "$status_code" = "307" ] || [ "$status_code" = "200" ]; then
    pass
else
    fail "Expected redirect (302/307) or 200, got $status_code"
fi

# Tests requiring admin token
if [ -n "$ADMIN_TOKEN" ]; then
    print_header "Authenticated Endpoint Tests"

    # Test 4: Authenticated request works
    print_test "Authenticated requests are accepted"
    response=$(make_request GET /api/v1/schedules "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    if [ "$status_code" = "200" ]; then
        pass
    else
        fail "Expected 200, got $status_code"
    fi

    # Test 5: User profile endpoint
    print_test "Get current user profile"
    response=$(make_request GET /users/me "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    if [ "$status_code" = "200" ]; then
        # Check if response contains expected fields
        if echo "$body" | grep -q '"email"' && echo "$body" | grep -q '"role"'; then
            pass
        else
            fail "Response missing required fields"
        fi
    else
        fail "Expected 200, got $status_code"
    fi

    # Test 6: List users (admin only)
    print_test "List users endpoint (admin only)"
    response=$(make_request GET /users "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    if [ "$status_code" = "200" ] || [ "$status_code" = "403" ]; then
        pass
    else
        fail "Expected 200 or 403, got $status_code"
    fi

    # Test 7: Token refresh
    print_test "Token refresh endpoint"
    response=$(make_request POST /auth/refresh "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    if [ "$status_code" = "200" ]; then
        if echo "$body" | grep -q '"token"'; then
            pass
        else
            fail "Response missing token field"
        fi
    else
        fail "Expected 200, got $status_code"
    fi

    # Test 8: Create schedule
    print_test "Create schedule with authentication"
    schedule_data='{
        "scheduledAt": "'$(date -u -d '+1 day' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+1d +%Y-%m-%dT%H:%M:%SZ)'",
        "serviceName": "smoke-test-service",
        "environments": ["development"],
        "owners": ["smoke-test"]
    }'
    response=$(make_request POST /api/v1/schedules "$ADMIN_TOKEN" "$schedule_data")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    if [ "$status_code" = "201" ] || [ "$status_code" = "200" ]; then
        # Extract schedule ID for cleanup
        SCHEDULE_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        if [ -n "$SCHEDULE_ID" ]; then
            pass
        else
            fail "Schedule created but no ID returned"
        fi
    else
        fail "Expected 201 or 200, got $status_code"
    fi

    # Test 9: Get schedule (verify audit trail)
    if [ -n "$SCHEDULE_ID" ]; then
        print_test "Get schedule and verify audit trail"
        response=$(make_request GET "/api/v1/schedules/$SCHEDULE_ID" "$ADMIN_TOKEN")
        status_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | head -n -1)
        if [ "$status_code" = "200" ]; then
            # Check for createdBy field
            if echo "$body" | grep -q '"createdBy"'; then
                pass
            else
                fail "Schedule missing createdBy audit field"
            fi
        else
            fail "Expected 200, got $status_code"
        fi

        # Test 10: Delete schedule (cleanup)
        print_test "Delete schedule (cleanup)"
        response=$(make_request DELETE "/api/v1/schedules/$SCHEDULE_ID" "$ADMIN_TOKEN")
        status_code=$(echo "$response" | tail -n1)
        if [ "$status_code" = "204" ] || [ "$status_code" = "200" ]; then
            pass
        else
            fail "Expected 204 or 200, got $status_code"
        fi
    fi

    # Test 11: Logout (token revocation)
    print_test "Logout (token revocation)"
    response=$(make_request POST /auth/logout "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    if [ "$status_code" = "200" ] || [ "$status_code" = "204" ]; then
        pass
    else
        fail "Expected 200 or 204, got $status_code"
    fi

    # Test 12: Revoked token is rejected
    print_test "Revoked token is rejected"
    sleep 2  # Wait for revocation to propagate
    response=$(make_request GET /api/v1/schedules "$ADMIN_TOKEN")
    status_code=$(echo "$response" | tail -n1)
    if [ "$status_code" = "401" ]; then
        pass
    else
        fail "Expected 401 for revoked token, got $status_code"
    fi

else
    echo ""
    echo -e "${YELLOW}ℹ  Admin token not provided. Skipping authenticated tests.${NC}"
    echo -e "${YELLOW}   To run full test suite:${NC}"
    echo -e "${YELLOW}     1. Login: deployment-tail auth login${NC}"
    echo -e "${YELLOW}     2. Get token from ~/.deployment-tail/auth.json${NC}"
    echo -e "${YELLOW}     3. Export: export ADMIN_TOKEN=your-token${NC}"
    echo -e "${YELLOW}     4. Re-run: ./scripts/smoke_test.sh${NC}"
    echo ""
fi

# Test 13: Database connectivity
print_header "Infrastructure Checks"

print_test "Database is accessible (via health check)"
# We already tested this, but let's verify again
response=$(make_request GET /health "")
status_code=$(echo "$response" | tail -n1)
if [ "$status_code" = "200" ]; then
    pass
else
    fail "Health check failed - database may be down"
fi

# Summary
print_header "Test Summary"

echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
fi

SUCCESS_RATE=$((TESTS_PASSED * 100 / TOTAL_TESTS))
echo "Success Rate: ${SUCCESS_RATE}%"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   ✓ All smoke tests passed!          ║${NC}"
    echo -e "${GREEN}║   Deployment appears healthy.         ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════╗${NC}"
    echo -e "${RED}║   ✗ Some tests failed!                ║${NC}"
    echo -e "${RED}║   Check logs for details.             ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════╝${NC}"
    exit 1
fi
