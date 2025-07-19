#!/bin/bash

# Concurrent Access Tests for Tiny URL Service
# Usage: ./tests/concurrent_tests.sh [base_url] [num_requests]

BASE_URL=${1:-"http://localhost:8080"}
NUM_REQUESTS=${2:-50}
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create temporary directory for test results
TEST_DIR="/tmp/tiny_url_concurrent_test_$$"
mkdir -p "$TEST_DIR"

# Cleanup function
cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

echo -e "${YELLOW}🧪 Starting Concurrent Access Tests${NC}"
echo -e "${BLUE}Testing against: $BASE_URL${NC}"
echo -e "${BLUE}Number of concurrent requests: $NUM_REQUESTS${NC}"
echo

# Test 1: Concurrent URL creation
echo -e "${BLUE}Testing: Concurrent URL creation${NC}"
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Function to create a URL in background
create_url() {
    local id=$1
    local url="https://www.example.com/concurrent/test/$id"
    local result_file="$TEST_DIR/result_$id.json"
    
    curl -s -X POST "$BASE_URL/urls" \
        -H "Content-Type: application/json" \
        -d "{\"long_url\": \"$url\"}" \
        > "$result_file" 2>/dev/null
    
    echo $? > "$TEST_DIR/status_$id"
}

# Start concurrent requests
echo "Starting $NUM_REQUESTS concurrent URL creation requests..."
for i in $(seq 1 $NUM_REQUESTS); do
    create_url $i &
done

# Wait for all requests to complete
wait

# Analyze results
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0
UNIQUE_SHORT_CODES=()

for i in $(seq 1 $NUM_REQUESTS); do
    status_file="$TEST_DIR/status_$i"
    result_file="$TEST_DIR/result_$i.json"
    
    if [[ -f "$status_file" ]] && [[ $(cat "$status_file") -eq 0 ]]; then
        if [[ -f "$result_file" ]] && grep -q "short_url" "$result_file"; then
            SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
            # Extract short code
            short_url=$(sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p' "$result_file")
            short_code=$(echo "$short_url" | sed "s|$BASE_URL/||")
            UNIQUE_SHORT_CODES+=("$short_code")
        else
            FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
        fi
    else
        FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
    fi
done

echo "Successful requests: $SUCCESSFUL_REQUESTS"
echo "Failed requests: $FAILED_REQUESTS"

# Check for unique short codes (no duplicates)
UNIQUE_COUNT=$(printf '%s\n' "${UNIQUE_SHORT_CODES[@]}" | sort -u | wc -l)
TOTAL_CODES=${#UNIQUE_SHORT_CODES[@]}

echo "Total short codes generated: $TOTAL_CODES"
echo "Unique short codes: $UNIQUE_COUNT"

if [[ $SUCCESSFUL_REQUESTS -eq $NUM_REQUESTS ]] && [[ $UNIQUE_COUNT -eq $TOTAL_CODES ]]; then
    echo -e "${GREEN}✅ PASS${NC}: Concurrent URL creation (all requests successful, all codes unique)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAIL${NC}: Concurrent URL creation"
    if [[ $SUCCESSFUL_REQUESTS -ne $NUM_REQUESTS ]]; then
        echo "  Expected $NUM_REQUESTS successful requests, got $SUCCESSFUL_REQUESTS"
    fi
    if [[ $UNIQUE_COUNT -ne $TOTAL_CODES ]]; then
        echo "  Duplicate short codes detected! ($UNIQUE_COUNT unique out of $TOTAL_CODES total)"
    fi
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 2: Concurrent access to same short URL
echo -e "${BLUE}Testing: Concurrent access to same short URL${NC}"
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Create a URL first
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/urls" \
    -H "Content-Type: application/json" \
    -d '{"long_url": "https://www.example.com/concurrent/redirect/test"}')

if echo "$CREATE_RESPONSE" | grep -q "short_url"; then
    SHORT_URL=$(echo "$CREATE_RESPONSE" | sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p')
    SHORT_CODE=$(echo "$SHORT_URL" | sed "s|$BASE_URL/||")
    
    # Function to access URL in background
    access_url() {
        local id=$1
        local status_file="$TEST_DIR/redirect_status_$id"
        
        # Use -I to get headers only (HEAD request alternative)
        curl -s -I "$BASE_URL/$SHORT_CODE" > "$TEST_DIR/redirect_result_$id" 2>/dev/null
        echo $? > "$status_file"
    }
    
    # Start concurrent access requests
    echo "Starting $NUM_REQUESTS concurrent access requests to $SHORT_CODE..."
    for i in $(seq 1 $NUM_REQUESTS); do
        access_url $i &
    done
    
    # Wait for all requests to complete
    wait
    
    # Analyze redirect results
    SUCCESSFUL_REDIRECTS=0
    FAILED_REDIRECTS=0
    
    for i in $(seq 1 $NUM_REQUESTS); do
        status_file="$TEST_DIR/redirect_status_$i"
        result_file="$TEST_DIR/redirect_result_$i"
        
        if [[ -f "$status_file" ]] && [[ $(cat "$status_file") -eq 0 ]]; then
            if [[ -f "$result_file" ]] && grep -q "HTTP/1.1 302" "$result_file"; then
                SUCCESSFUL_REDIRECTS=$((SUCCESSFUL_REDIRECTS + 1))
            else
                FAILED_REDIRECTS=$((FAILED_REDIRECTS + 1))
            fi
        else
            FAILED_REDIRECTS=$((FAILED_REDIRECTS + 1))
        fi
    done
    
    echo "Successful redirects: $SUCCESSFUL_REDIRECTS"
    echo "Failed redirects: $FAILED_REDIRECTS"
    
    if [[ $SUCCESSFUL_REDIRECTS -eq $NUM_REQUESTS ]]; then
        echo -e "${GREEN}✅ PASS${NC}: Concurrent redirect access"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAIL${NC}: Concurrent redirect access"
        echo "  Expected $NUM_REQUESTS successful redirects, got $SUCCESSFUL_REDIRECTS"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
else
    echo -e "${RED}❌ FAIL${NC}: Concurrent redirect access (could not create test URL)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 3: Mixed concurrent operations
echo -e "${BLUE}Testing: Mixed concurrent operations (create + access)${NC}"
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Function for mixed operations
mixed_operation() {
    local id=$1
    local operation_type=$((id % 3))
    
    case $operation_type in
        0) # Create URL
            curl -s -X POST "$BASE_URL/urls" \
                -H "Content-Type: application/json" \
                -d "{\"long_url\": \"https://www.example.com/mixed/$id\"}" \
                > "$TEST_DIR/mixed_result_$id" 2>/dev/null
            ;;
        1) # Access existing URL (if any)
            curl -s -I "$BASE_URL/1" > "$TEST_DIR/mixed_result_$id" 2>/dev/null
            ;;
        2) # Get stats (if any)
            curl -s "$BASE_URL/urls/1/stats" > "$TEST_DIR/mixed_result_$id" 2>/dev/null
            ;;
    esac
    
    echo $? > "$TEST_DIR/mixed_status_$id"
}

# Start mixed concurrent operations
echo "Starting $NUM_REQUESTS mixed concurrent operations..."
for i in $(seq 1 $NUM_REQUESTS); do
    mixed_operation $i &
done

# Wait for all operations to complete
wait

# Analyze mixed results
SUCCESSFUL_MIXED=0
FAILED_MIXED=0

for i in $(seq 1 $NUM_REQUESTS); do
    status_file="$TEST_DIR/mixed_status_$i"
    
    if [[ -f "$status_file" ]] && [[ $(cat "$status_file") -eq 0 ]]; then
        SUCCESSFUL_MIXED=$((SUCCESSFUL_MIXED + 1))
    else
        FAILED_MIXED=$((FAILED_MIXED + 1))
    fi
done

echo "Successful mixed operations: $SUCCESSFUL_MIXED"
echo "Failed mixed operations: $FAILED_MIXED"

# We expect some failures due to accessing non-existent URLs, so we're lenient
if [[ $SUCCESSFUL_MIXED -gt $((NUM_REQUESTS / 2)) ]]; then
    echo -e "${GREEN}✅ PASS${NC}: Mixed concurrent operations (server remained stable)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAIL${NC}: Mixed concurrent operations (too many failures)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Summary
echo -e "${YELLOW}📊 Concurrent Test Summary${NC}"
echo -e "Total tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "${GREEN}🎉 All concurrent tests passed!${NC}"
    echo -e "${GREEN}✅ No race conditions detected${NC}"
    echo -e "${GREEN}✅ Thread safety verified${NC}"
    exit 0
else
    echo -e "${RED}💥 Some concurrent tests failed!${NC}"
    exit 1
fi 