#!/bin/bash

# Basic Functionality Tests for Tiny URL Service
# Usage: ./tests/basic_tests.sh [base_url]

BASE_URL=${1:-"http://localhost:8080"}
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_status="$3"
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Run the command and capture output and status
    local output
    local status
    output=$(eval "$test_command" 2>&1)
    status=$?
    
    # Check HTTP status if specified
    if [[ -n "$expected_status" ]]; then
        local http_status="$output"
        if [[ "$http_status" == "$expected_status" ]]; then
            echo -e "${GREEN}✅ PASS${NC}: $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAIL${NC}: $test_name (Expected HTTP $expected_status, got $http_status)"
            echo "Output: $output"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        if [[ $status -eq 0 ]]; then
            echo -e "${GREEN}✅ PASS${NC}: $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAIL${NC}: $test_name"
            echo "Output: $output"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
    echo
}

echo -e "${YELLOW}🧪 Starting Basic Functionality Tests${NC}"
echo -e "${BLUE}Testing against: $BASE_URL${NC}"
echo

# Test 1: Health Check
run_test "Health check endpoint" \
    "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/health" \
    "200"

# Test 2: Create short URL with valid long URL
echo -e "${BLUE}Testing: Create short URL${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/urls \
    -H "Content-Type: application/json" \
    -d '{"long_url": "https://www.example.com/test"}')

if echo "$RESPONSE" | grep -q "short_url"; then
    SHORT_URL=$(echo "$RESPONSE" | sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p')
    SHORT_CODE=$(echo "$SHORT_URL" | sed "s|$BASE_URL/||")
    echo -e "${GREEN}✅ PASS${NC}: Create short URL"
    echo "  Short URL: $SHORT_URL"
    echo "  Short Code: $SHORT_CODE"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAIL${NC}: Create short URL"
    echo "Response: $RESPONSE"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))
echo

# Test 3: Redirect functionality (only if we got a short code)
if [[ -n "$SHORT_CODE" ]]; then
    run_test "Redirect functionality" \
        "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/$SHORT_CODE" \
        "302"
        
    # Test 4: Stats endpoint
    run_test "Stats endpoint" \
        "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/urls/$SHORT_CODE/stats" \
        "200"
else
    echo -e "${YELLOW}⚠️  SKIP${NC}: Redirect and stats tests (no short code available)"
    TOTAL_TESTS=$((TOTAL_TESTS + 2))
fi

# Test 5: Multiple URL formats
TEST_URLS=(
    "https://www.google.com"
    "http://example.com/path/to/page"
    "https://subdomain.example.com:8080/path?query=value&other=param"
    "http://192.168.1.1:3000/api/endpoint"
)

for url in "${TEST_URLS[@]}"; do
    run_test "Create short URL for: $url" \
        "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"$url\"}' | grep -q short_url"
done

# Test 6: URL with expiration (macOS compatible date)
EXPIRATION=$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-12-31T23:59:59Z")
run_test "Create short URL with expiration" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"https://www.example.com/expiring\", \"expiration_date\": \"$EXPIRATION\"}' | grep -q short_url"

# Summary
echo -e "${YELLOW}📊 Test Summary${NC}"
echo -e "Total tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "${GREEN}🎉 All basic tests passed!${NC}"
    exit 0
else
    echo -e "${RED}💥 Some tests failed!${NC}"
    exit 1
fi 