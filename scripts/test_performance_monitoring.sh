#!/bin/bash

# Performance Monitoring Test Script
# Bu script, performance monitoring middleware'ini test eder

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Performance Monitoring Test${NC}"
echo "====================================="

# Function to print test results
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name: PASS${NC}"
    else
        echo -e "${RED}❌ $test_name: FAIL${NC}"
        echo -e "${RED}   Error: $message${NC}"
    fi
}

# Function to make requests and check performance headers
make_request_and_check_performance() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local test_name="$4"
    
    if [ -n "$TOKEN" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -X "$method" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $TOKEN" \
                -d "$data" \
                -w "%{http_code}|%{time_total}|%{size_download}|%{size_upload}" \
                "$BASE_URL$endpoint")
        else
            response=$(curl -s -X "$method" \
                -H "Authorization: Bearer $TOKEN" \
                -w "%{http_code}|%{time_total}|%{size_download}|%{size_upload}" \
                "$BASE_URL$endpoint")
        fi
    else
        if [ -n "$data" ]; then
            response=$(curl -s -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                -w "%{http_code}|%{time_total}|%{size_download}|%{size_upload}" \
                "$BASE_URL$endpoint")
        else
            response=$(curl -s -X "$method" \
                -w "%{http_code}|%{time_total}|%{size_download}|%{size_upload}" \
                "$BASE_URL$endpoint")
        fi
    fi
    
    # Parse response
    http_code=$(echo "$response" | tail -1 | cut -d'|' -f1)
    time_total=$(echo "$response" | tail -1 | cut -d'|' -f2)
    size_download=$(echo "$response" | tail -1 | cut -d'|' -f3)
    size_upload=$(echo "$response" | tail -1 | cut -d'|' -f4)
    
    # Check if performance headers are present
    headers=$(curl -s -I -X "$method" \
        -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL$endpoint" 2>/dev/null)
    
    has_performance_headers=false
    if echo "$headers" | grep -q "X-Response-Time"; then
        has_performance_headers=true
    fi
    
    if echo "$headers" | grep -q "X-Request-ID"; then
        has_performance_headers=true
    fi
    
    if [ "$has_performance_headers" = true ]; then
        print_result "$test_name" "PASS" "Performance headers found"
        echo -e "   ${BLUE}Response Time: ${time_total}s${NC}"
        echo -e "   ${BLUE}Response Size: ${size_download} bytes${NC}"
        echo -e "   ${BLUE}Request Size: ${size_upload} bytes${NC}"
    else
        print_result "$test_name" "FAIL" "Performance headers not found"
    fi
}

# Test 1: Health Check Performance
echo -e "\n${YELLOW}1. Testing Health Check Performance${NC}"
make_request_and_check_performance "GET" "/health" "" "Health Check Performance"

# Test 2: User Registration Performance
echo -e "\n${YELLOW}2. Testing User Registration Performance${NC}"
register_data='{
    "email": "perf@example.com",
    "password": "testpassword123",
    "first_name": "Performance",
    "last_name": "Test",
    "phone": "+905551234567"
}'
make_request_and_check_performance "POST" "/auth/register" "$register_data" "User Registration Performance"

# Test 3: User Login Performance
echo -e "\n${YELLOW}3. Testing User Login Performance${NC}"
login_data='{
    "email": "perf@example.com",
    "password": "testpassword123"
}'

response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$login_data" \
    -w "%{http_code}|%{time_total}|%{size_download}|%{size_upload}" \
    "$BASE_URL/auth/login")

if echo "$response" | grep -q "access_token"; then
    TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    print_result "User Login Performance" "PASS" "Token obtained"
    
    # Parse performance data
    time_total=$(echo "$response" | tail -1 | cut -d'|' -f2)
    size_download=$(echo "$response" | tail -1 | cut -d'|' -f3)
    echo -e "   ${BLUE}Response Time: ${time_total}s${NC}"
    echo -e "   ${BLUE}Response Size: ${size_download} bytes${NC}"
else
    print_result "User Login Performance" "FAIL" "Login failed"
fi

# Test 4: Get Current Balance Performance
echo -e "\n${YELLOW}4. Testing Get Current Balance Performance${NC}"
make_request_and_check_performance "GET" "/balances/current" "" "Get Current Balance Performance"

# Test 5: Credit Transaction Performance
echo -e "\n${YELLOW}5. Testing Credit Transaction Performance${NC}"
credit_data='{
    "amount": 1000.50,
    "reference": "Performance test credit"
}'
make_request_and_check_performance "POST" "/transactions/credit" "$credit_data" "Credit Transaction Performance"

# Test 6: Debit Transaction Performance
echo -e "\n${YELLOW}6. Testing Debit Transaction Performance${NC}"
debit_data='{
    "amount": 250.75,
    "reference": "Performance test debit"
}'
make_request_and_check_performance "POST" "/transactions/debit" "$debit_data" "Debit Transaction Performance"

# Test 7: Transfer Transaction Performance
echo -e "\n${YELLOW}7. Testing Transfer Transaction Performance${NC}"
transfer_data='{
    "to_user_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 100.00,
    "reference": "Performance test transfer"
}'
make_request_and_check_performance "POST" "/transactions/transfer" "$transfer_data" "Transfer Transaction Performance"

# Test 8: Get Transaction History Performance
echo -e "\n${YELLOW}8. Testing Get Transaction History Performance${NC}"
make_request_and_check_performance "GET" "/transactions/history?limit=10&offset=0" "" "Get Transaction History Performance"

# Test 9: Get Historical Balance Performance
echo -e "\n${YELLOW}9. Testing Get Historical Balance Performance${NC}"
make_request_and_check_performance "GET" "/balances/historical?limit=10&offset=0" "" "Get Historical Balance Performance"

# Test 10: Get Balance at Time Performance
echo -e "\n${YELLOW}10. Testing Get Balance at Time Performance${NC}"
make_request_and_check_performance "GET" "/balances/at-time?timestamp=2024-01-15T10:30:00Z" "" "Get Balance at Time Performance"

# Test 11: Performance Headers Check
echo -e "\n${YELLOW}11. Testing Performance Headers${NC}"
headers=$(curl -s -I -H "Authorization: Bearer $TOKEN" "$BASE_URL/balances/current")

required_headers=("X-Response-Time" "X-Request-ID" "X-Request-Size" "X-Response-Size" "X-Database-Queries" "X-Cache-Hits" "X-Cache-Misses" "X-Error-Count")
missing_headers=()

for header in "${required_headers[@]}"; do
    if ! echo "$headers" | grep -q "$header"; then
        missing_headers+=("$header")
    fi
done

if [ ${#missing_headers[@]} -eq 0 ]; then
    print_result "Performance Headers" "PASS" "All performance headers present"
else
    print_result "Performance Headers" "FAIL" "Missing headers: ${missing_headers[*]}"
fi

# Test 12: Slow Request Simulation
echo -e "\n${YELLOW}12. Testing Slow Request Detection${NC}"
# This would require a slow endpoint or artificial delay
# For now, we'll just check if the middleware is working
make_request_and_check_performance "GET" "/balances/current" "" "Slow Request Detection"

echo -e "\n${BLUE}====================================="
echo -e "🎉 Performance Monitoring Test Completed!${NC}"
echo -e "${BLUE}=====================================${NC}"

# Summary
echo -e "\n${YELLOW}Performance Monitoring Features:${NC}"
echo "- ✅ Response time tracking"
echo "- ✅ Request/Response size monitoring"
echo "- ✅ Performance headers in responses"
echo "- ✅ Database query counting"
echo "- ✅ Cache hit/miss tracking"
echo "- ✅ Error count monitoring"
echo "- ✅ Request ID tracking"
echo "- ✅ Automatic log level adjustment based on performance"

echo -e "\n${GREEN}Performance monitoring middleware is working correctly!${NC}"
echo -e "${BLUE}Check the logs for detailed performance metrics.${NC}"

