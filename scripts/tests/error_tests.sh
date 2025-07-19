#!/bin/bash

# Error Case Tests for Tiny URL Service
# Usage: ./tests/error_tests.sh [base_url]

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

# Helper function to test expected failures
test_error_case() {
    local test_name="$1"
    local curl_command="$2"
    local expected_status="$3"
    local expected_error_text="$4"
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Run the command and capture output and HTTP status
    local response
    local http_status
    response=$(eval "$curl_command" 2>/dev/null)
    http_status=$(eval "$curl_command -w '%{http_code}'" 2>/dev/null | tail -c 3)
    
    local passed=true
    local error_msg=""
    
    # Check HTTP status
    if [[ "$http_status" != "$expected_status" ]]; then
        passed=false
        error_msg="Expected HTTP $expected_status, got $http_status"
    fi
    
    # Check error message content if specified
    if [[ -n "$expected_error_text" ]] && [[ "$passed" == true ]]; then
        if ! echo "$response" | grep -q "$expected_error_text"; then
            passed=false
            error_msg="Expected error text '$expected_error_text' not found in response"
        fi
    fi
    
    if [[ "$passed" == true ]]; then
        echo -e "${GREEN}✅ PASS${NC}: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAIL${NC}: $test_name ($error_msg)"
        echo "Response: $response"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo
}

echo -e "${YELLOW}🧪 Starting Error Case Tests${NC}"
echo -e "${BLUE}Testing against: $BASE_URL${NC}"
echo

# Test 1: Missing Content-Type header
test_error_case "Missing Content-Type header" \
    "curl -s -X POST $BASE_URL/urls -d '{\"long_url\": \"https://www.example.com\"}'" \
    "400" \
    "Content-Type must be application/json"

# Test 2: Wrong Content-Type header
test_error_case "Wrong Content-Type header" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: text/plain' -d '{\"long_url\": \"https://www.example.com\"}'" \
    "400" \
    "Content-Type must be application/json"

# Test 3: Invalid JSON format
test_error_case "Invalid JSON format" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"invalid\": json}'" \
    "400" \
    "Invalid JSON format"

# Test 4: Missing required field (long_url)
test_error_case "Missing long_url field" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"other_field\": \"value\"}'" \
    "400" \
    "Invalid JSON format"

# Test 5: Empty long_url
test_error_case "Empty long_url" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"\"}'" \
    "400" \
    "Invalid URL format"

# Test 6: Invalid URL format - no protocol
test_error_case "Invalid URL - no protocol" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"www.example.com\"}'" \
    "400" \
    "Invalid URL format"

# Test 7: Invalid URL format - wrong protocol
test_error_case "Invalid URL - wrong protocol" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"ftp://example.com\"}'" \
    "400" \
    "Invalid URL format"

# Test 8: Invalid URL format - malformed
test_error_case "Invalid URL - malformed" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"not-a-url\"}'" \
    "400" \
    "Invalid URL format"

# Test 9: Non-existent short code
test_error_case "Non-existent short code redirect" \
    "curl -s $BASE_URL/nonexistent123" \
    "404" \
    "Short URL not found"

# Test 10: Non-existent short code stats
test_error_case "Non-existent short code stats" \
    "curl -s $BASE_URL/urls/nonexistent123/stats" \
    "404" \
    "Short URL not found"

# Test 11: Invalid HTTP method on redirect endpoint
test_error_case "Invalid HTTP method on redirect" \
    "curl -s -X POST $BASE_URL/1" \
    "404" \
    ""

# Test 12: Invalid expiration date format
test_error_case "Invalid expiration date format" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d '{\"long_url\": \"https://www.example.com\", \"expiration_date\": \"invalid-date\"}'" \
    "400" \
    ""

# Test 13: Empty request body
test_error_case "Empty request body" \
    "curl -s -X POST $BASE_URL/urls -H 'Content-Type: application/json' -d ''" \
    "400" \
    ""

# Test 14: Request to non-existent endpoint
test_error_case "Non-existent endpoint" \
    "curl -s $BASE_URL/nonexistent/endpoint" \
    "404" \
    ""

# Summary
echo -e "${YELLOW}📊 Error Test Summary${NC}"
echo -e "Total tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "${GREEN}🎉 All error tests passed!${NC}"
    exit 0
else
    echo -e "${RED}💥 Some error tests failed!${NC}"
    exit 1
fi 