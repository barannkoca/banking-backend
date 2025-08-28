#!/bin/bash

# Redis Cache Test Script
# Bu script Redis cache'inin düzgün çalışıp çalışmadığını test eder

set -e

echo "🧪 Redis Cache Test Script"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test functions
test_redis_connection() {
    echo -e "${BLUE}🔍 Testing Redis connection...${NC}"
    
    if redis-cli ping > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Redis connection successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Redis connection failed${NC}"
        return 1
    fi
}

test_cache_health_endpoint() {
    echo -e "${BLUE}🔍 Testing cache health endpoint...${NC}"
    
    response=$(curl -s http://localhost:8080/health/cache)
    
    if echo "$response" | grep -q "cache_healthy"; then
        echo -e "${GREEN}✅ Cache health endpoint working${NC}"
        echo -e "${YELLOW}Response: $response${NC}"
        return 0
    else
        echo -e "${RED}❌ Cache health endpoint failed${NC}"
        echo -e "${YELLOW}Response: $response${NC}"
        return 1
    fi
}

test_basic_cache_operations() {
    echo -e "${BLUE}🔍 Testing basic cache operations...${NC}"
    
    # Test SET
    if redis-cli set "test:key" "test:value" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ SET operation successful${NC}"
    else
        echo -e "${RED}❌ SET operation failed${NC}"
        return 1
    fi
    
    # Test GET
    value=$(redis-cli get "test:key" 2>/dev/null)
    if [ "$value" = "test:value" ]; then
        echo -e "${GREEN}✅ GET operation successful${NC}"
    else
        echo -e "${RED}❌ GET operation failed${NC}"
        return 1
    fi
    
    # Test DELETE
    if redis-cli del "test:key" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ DELETE operation successful${NC}"
    else
        echo -e "${RED}❌ DELETE operation failed${NC}"
        return 1
    fi
    
    # Test GET after DELETE
    value=$(redis-cli get "test:key" 2>/dev/null)
    if [ -z "$value" ]; then
        echo -e "${GREEN}✅ GET after DELETE successful (key not found)${NC}"
    else
        echo -e "${RED}❌ GET after DELETE failed (key still exists)${NC}"
        return 1
    fi
}

test_banking_cache_operations() {
    echo -e "${BLUE}🔍 Testing banking-specific cache operations...${NC}"
    
    # Test balance cache
    user_id="test-user-123"
    balance="1000.50"
    
    # Set balance
    if redis-cli set "balance:$user_id" "$balance" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Balance SET operation successful${NC}"
    else
        echo -e "${RED}❌ Balance SET operation failed${NC}"
        return 1
    fi
    
    # Get balance
    cached_balance=$(redis-cli get "balance:$user_id" 2>/dev/null)
    if [ "$cached_balance" = "$balance" ]; then
        echo -e "${GREEN}✅ Balance GET operation successful${NC}"
    else
        echo -e "${RED}❌ Balance GET operation failed${NC}"
        return 1
    fi
    
    # Test transaction cache
    transaction_id="test-tx-456"
    transaction_data='{"id":"test-tx-456","amount":500.00,"type":"credit"}'
    
    # Set transaction
    if redis-cli set "transaction:$transaction_id" "$transaction_data" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Transaction SET operation successful${NC}"
    else
        echo -e "${RED}❌ Transaction SET operation failed${NC}"
        return 1
    fi
    
    # Get transaction
    cached_transaction=$(redis-cli get "transaction:$transaction_id" 2>/dev/null)
    if [ "$cached_transaction" = "$transaction_data" ]; then
        echo -e "${GREEN}✅ Transaction GET operation successful${NC}"
    else
        echo -e "${RED}❌ Transaction GET operation failed${NC}"
        return 1
    fi
    
    # Cleanup
    redis-cli del "balance:$user_id" > /dev/null 2>&1
    redis-cli del "transaction:$transaction_id" > /dev/null 2>&1
    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

test_cache_statistics() {
    echo -e "${BLUE}🔍 Testing cache statistics...${NC}"
    
    # Get Redis info
    info=$(redis-cli info memory 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Redis INFO command successful${NC}"
        echo -e "${YELLOW}Memory usage: $(echo "$info" | grep "used_memory_human" | cut -d: -f2)${NC}"
    else
        echo -e "${RED}❌ Redis INFO command failed${NC}"
        return 1
    fi
    
    # Get Redis keys count
    keys_count=$(redis-cli dbsize 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Redis DBSIZE command successful${NC}"
        echo -e "${YELLOW}Total keys: $keys_count${NC}"
    else
        echo -e "${RED}❌ Redis DBSIZE command failed${NC}"
        return 1
    fi
}

test_cache_performance() {
    echo -e "${BLUE}🔍 Testing cache performance...${NC}"
    
    # Test bulk operations
    start_time=$(date +%s.%N)
    
    for i in {1..100}; do
        redis-cli set "perf:key:$i" "value:$i" > /dev/null 2>&1
    done
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc -l)
    
    echo -e "${GREEN}✅ Bulk SET operations completed${NC}"
    echo -e "${YELLOW}Duration: ${duration}s for 100 operations${NC}"
    
    # Cleanup
    for i in {1..100}; do
        redis-cli del "perf:key:$i" > /dev/null 2>&1
    done
    
    echo -e "${GREEN}✅ Performance test cleanup completed${NC}"
}

# Main test execution
main() {
    echo -e "${YELLOW}Starting Redis cache tests...${NC}"
    echo ""
    
    # Check if Redis is running
    if ! test_redis_connection; then
        echo -e "${RED}❌ Redis is not running. Please start Redis first.${NC}"
        exit 1
    fi
    
    # Check if backend is running
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${RED}❌ Backend is not running. Please start the backend first.${NC}"
        exit 1
    fi
    
    echo ""
    
    # Run tests
    test_cache_health_endpoint
    echo ""
    
    test_basic_cache_operations
    echo ""
    
    test_banking_cache_operations
    echo ""
    
    test_cache_statistics
    echo ""
    
    test_cache_performance
    echo ""
    
    echo -e "${GREEN}🎉 All Redis cache tests completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}📊 Cache Statistics:${NC}"
    echo -e "${YELLOW}Total keys in Redis: $(redis-cli dbsize)${NC}"
    echo -e "${YELLOW}Memory usage: $(redis-cli info memory | grep "used_memory_human" | cut -d: -f2)${NC}"
}

# Run main function
main "$@"
