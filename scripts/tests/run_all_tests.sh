#!/bin/bash

# Master Test Runner for Tiny URL Service
# Usage: ./tests/run_all_tests.sh [base_url]

BASE_URL=${1:-"http://localhost:8080"}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test results
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

echo -e "${CYAN}🧪 Tiny URL Service - Complete Test Suite${NC}"
echo -e "${BLUE}Testing against: $BASE_URL${NC}"
echo -e "${BLUE}$(date)${NC}"
echo

# Function to run a test suite
run_test_suite() {
    local suite_name="$1"
    local script_path="$2"
    local description="$3"
    
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}🔍 $suite_name${NC}"
    echo -e "${BLUE}$description${NC}"
    echo

    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    
    if [[ -f "$script_path" ]]; then
        if bash "$script_path" "$BASE_URL"; then
            echo -e "${GREEN}✅ $suite_name PASSED${NC}"
            PASSED_SUITES=$((PASSED_SUITES + 1))
        else
            echo -e "${RED}❌ $suite_name FAILED${NC}"
            FAILED_SUITES=$((FAILED_SUITES + 1))
        fi
    else
        echo -e "${RED}❌ Test script not found: $script_path${NC}"
        FAILED_SUITES=$((FAILED_SUITES + 1))
    fi
    
    echo
}

# Check if server is running
echo -e "${BLUE}Checking if server is running...${NC}"
if ! curl -s -f "$BASE_URL/health" > /dev/null; then
    echo -e "${RED}❌ Server is not running at $BASE_URL${NC}"
    echo -e "${YELLOW}Please start the server first:${NC}"
    echo -e "${BLUE}  ./tiny-url-service${NC}"
    echo -e "${BLUE}  # or with custom port:${NC}"
    echo -e "${BLUE}  PORT=8080 ./tiny-url-service${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo

# Run all test suites
run_test_suite "Basic Functionality Tests" \
    "$SCRIPT_DIR/basic_tests.sh" \
    "Testing core URL shortening functionality, redirects, and API endpoints"

run_test_suite "Error Case Tests" \
    "$SCRIPT_DIR/error_tests.sh" \
    "Testing error handling, validation, and edge cases"

run_test_suite "Concurrent Access Tests" \
    "$SCRIPT_DIR/concurrent_tests.sh" \
    "Testing thread safety, race conditions, and concurrent access"

# Summary
echo -e "${YELLOW}========================================${NC}"
echo -e "${CYAN}📊 FINAL TEST RESULTS${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "Total test suites: $TOTAL_SUITES"
echo -e "${GREEN}Passed suites: $PASSED_SUITES${NC}"
echo -e "${RED}Failed suites: $FAILED_SUITES${NC}"
echo

if [[ $FAILED_SUITES -eq 0 ]]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
    echo -e "${GREEN}✅ Your Tiny URL service is working correctly${NC}"
    echo -e "${GREEN}✅ All error cases are handled properly${NC}"
    echo -e "${GREEN}✅ Thread safety and concurrency verified${NC}"
    echo
    echo -e "${CYAN}Your service is ready for production! 🚀${NC}"
    exit 0
else
    echo -e "${RED}💥 SOME TESTS FAILED! 💥${NC}"
    echo -e "${RED}Please review the failed tests above and fix any issues.${NC}"
    exit 1
fi 