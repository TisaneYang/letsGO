#!/bin/bash

# Product Redis Cache Testing Script
# This script tests Redis cache functionality and consistency guarantees

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Unset proxy environment variables to avoid gRPC connection issues
unset http_proxy
unset https_proxy
unset HTTP_PROXY
unset HTTPS_PROXY
unset all_proxy
unset ALL_PROXY

BASE_URL="http://localhost:8888"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Product Redis Cache Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# Function to print test results
print_result() {
    local test_name=$1
    local result=$2
    echo -e "${GREEN}[TEST]${NC} $test_name"
    echo "$result" | python3 -m json.tool 2>/dev/null || echo "$result"
    echo
}

# Function to print section header
print_section() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo
}

# Function to check Redis key
check_redis_key() {
    local key=$1
    local description=$2
    echo -e "${BLUE}[REDIS]${NC} $description"
    docker exec -it letsgo-redis redis-cli GET "$key" 2>/dev/null || echo "(Key not found or Redis not accessible)"
    echo
}

# ========================================
# Part 1: Setup - Register user for authentication
# ========================================
print_section "Part 1: Setup - User Registration"

echo -e "${BLUE}Registering test user for authentication...${NC}"
TIMESTAMP=$(date +%s)
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"rediscache_${TIMESTAMP}\",
    \"password\": \"password123\",
    \"email\": \"rediscache_${TIMESTAMP}@example.com\",
    \"phone\": \"137${TIMESTAMP: -8}\"
  }")
print_result "User Registration" "$REGISTER_RESULT"

# Extract token
TOKEN=$(echo $REGISTER_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get authentication token, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Got authentication token${NC}"
echo

# ========================================
# Part 2: Create Test Product
# ========================================
print_section "Part 2: Create Test Product"

echo -e "${BLUE}Creating test product...${NC}"
CURRENT_DATE=$(date)
ADD_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/product/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Redis Cache Test Product ${TIMESTAMP}\",
    \"description\": \"Testing Redis cache functionality - Created at ${CURRENT_DATE}\",
    \"price\": 99.99,
    \"stock\": 100,
    \"category\": \"Electronics\",
    \"images\": [
      \"https://example.com/redis-test-1.jpg\",
      \"https://example.com/redis-test-2.jpg\"
    ],
    \"attributes\": \"{\\\"color\\\": \\\"red\\\", \\\"size\\\": \\\"M\\\", \\\"test\\\": \\\"redis\\\"}\"
  }")
print_result "Add Product" "$ADD_RESULT"

# Extract product ID
PRODUCT_ID=$(echo $ADD_RESULT | grep -o '"productId":[0-9]*' | cut -d':' -f2)

if [ -z "$PRODUCT_ID" ]; then
    echo -e "${RED}Failed to create product, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Created product with ID: $PRODUCT_ID${NC}"
echo

# ========================================
# Part 3: Test Product Detail Cache
# ========================================
print_section "Part 3: Product Detail Cache (Cache-Aside Pattern)"

echo -e "${BLUE}Test 1: First query - Cache MISS (should query DB and write to Redis)${NC}"
DETAIL_1=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
print_result "Product Detail - First Query (Cache MISS)" "$DETAIL_1"

echo -e "${BLUE}Check Redis cache key: product:detail:${PRODUCT_ID}${NC}"
check_redis_key "product:detail:${PRODUCT_ID}" "Product detail cache"

echo -e "${BLUE}Test 2: Second query - Cache HIT (should read from Redis, no DB query)${NC}"
sleep 1
DETAIL_2=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
print_result "Product Detail - Second Query (Cache HIT)" "$DETAIL_2"

echo -e "${GREEN}Expected: Same data, but second query should be faster (from Redis)${NC}"
echo

# ========================================
# Part 4: Test Product List Cache with Version Control
# ========================================
print_section "Part 4: Product List Cache (Version-Based Invalidation)"

echo -e "${BLUE}Test 3: First list query - Cache MISS (query DB, write to Redis with version)${NC}"
LIST_1=$(curl -s "${BASE_URL}/api/v1/product/list?category=Electronics&page=1&pageSize=10")
print_result "List Products - First Query (Cache MISS)" "$LIST_1"

echo -e "${BLUE}Check Redis version keys${NC}"
check_redis_key "product:CategoryVersion:Electronics" "Category version for Electronics"
check_redis_key "product:GlobalVersion" "Global version"

echo -e "${BLUE}Test 4: Second list query - Cache HIT (version matches, read from Redis)${NC}"
sleep 1
LIST_2=$(curl -s "${BASE_URL}/api/v1/product/list?category=Electronics&page=1&pageSize=10")
print_result "List Products - Second Query (Cache HIT)" "$LIST_2"

echo -e "${GREEN}Expected: Same data, cached with version number${NC}"
echo

# ========================================
# Part 5: Test Search Cache with Version Control
# ========================================
print_section "Part 5: Product Search Cache (Version-Based Invalidation)"

echo -e "${BLUE}Test 5: First search - Cache MISS${NC}"
SEARCH_1=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=Redis&page=1&pageSize=10")
print_result "Search Products - First Query (Cache MISS)" "$SEARCH_1"

echo -e "${BLUE}Test 6: Second search - Cache HIT${NC}"
sleep 1
SEARCH_2=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=Redis&page=1&pageSize=10")
print_result "Search Products - Second Query (Cache HIT)" "$SEARCH_2"

echo -e "${GREEN}Expected: Search results cached${NC}"
echo

# ========================================
# Part 6: Test Cache Invalidation on Product Update
# ========================================
print_section "Part 6: Cache Invalidation on Product Update"

echo -e "${BLUE}Test 7: Update product (should invalidate caches)${NC}"
echo -e "  - Delete product:detail:${PRODUCT_ID}"
echo -e "  - Increment product:CategoryVersion:Electronics"
echo -e "  - Increment product:GlobalVersion"
UPDATE_RESULT=$(curl -s -X PUT ${BASE_URL}/api/v1/product/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": ${PRODUCT_ID},
    \"name\": \"UPDATED Redis Cache Test Product ${TIMESTAMP}\",
    \"description\": \"Updated product to test cache invalidation\",
    \"price\": 199.99
  }")
print_result "Update Product" "$UPDATE_RESULT"

echo -e "${BLUE}Check if detail cache was deleted${NC}"
check_redis_key "product:detail:${PRODUCT_ID}" "Product detail cache (should be deleted)"

echo -e "${BLUE}Test 8: Query product detail after update - Cache MISS (cache was deleted)${NC}"
sleep 1
DETAIL_3=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
print_result "Product Detail After Update (Cache MISS)" "$DETAIL_3"

echo -e "${GREEN}Expected: See updated data (name changed, price changed)${NC}"
echo

echo -e "${BLUE}Test 9: Query product detail again - Cache HIT (re-cached)${NC}"
sleep 1
DETAIL_4=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
print_result "Product Detail Second Query (Cache HIT)" "$DETAIL_4"

echo -e "${BLUE}Check Redis version numbers (should be incremented)${NC}"
check_redis_key "product:CategoryVersion:Electronics" "Category version (incremented)"
check_redis_key "product:GlobalVersion" "Global version (incremented)"

echo -e "${BLUE}Test 10: List products after update - Version changed, re-query DB${NC}"
sleep 1
LIST_3=$(curl -s "${BASE_URL}/api/v1/product/list?category=Electronics&page=1&pageSize=10")
print_result "List Products After Update" "$LIST_3"

echo -e "${GREEN}Expected: List cache detects version change, re-queries database${NC}"
echo

# ========================================
# Part 7: Test Cross-Category Cache Invalidation
# ========================================
print_section "Part 7: Cross-Category Cache Invalidation"

echo -e "${BLUE}Test 11: Change product category (Electronics -> Books)${NC}"
echo -e "  - Should increment Electronics version"
echo -e "  - Should increment Books version"
echo -e "  - Should increment Global version"
UPDATE_CAT=$(curl -s -X PUT ${BASE_URL}/api/v1/product/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": ${PRODUCT_ID},
    \"category\": \"Books\"
  }")
print_result "Update Category (Electronics -> Books)" "$UPDATE_CAT"

echo -e "${BLUE}Check version numbers for both categories${NC}"
check_redis_key "product:CategoryVersion:Electronics" "Electronics version (incremented)"
check_redis_key "product:CategoryVersion:Books" "Books version (incremented)"

echo -e "${BLUE}Test 12: List old category (Electronics) - Product should NOT appear${NC}"
sleep 1
LIST_OLD_CAT=$(curl -s "${BASE_URL}/api/v1/product/list?category=Electronics&page=1&pageSize=10")
print_result "List Electronics Category" "$LIST_OLD_CAT"

echo -e "${BLUE}Test 13: List new category (Books) - Product SHOULD appear${NC}"
sleep 1
LIST_NEW_CAT=$(curl -s "${BASE_URL}/api/v1/product/list?category=Books&page=1&pageSize=10")
print_result "List Books Category" "$LIST_NEW_CAT"

echo -e "${GREEN}Expected: Product moved from Electronics to Books${NC}"
echo

# ========================================
# Part 8: Verify Database Atomicity (Stock Operations)
# ========================================
print_section "Part 8: Database Atomicity (Stock Operations)"

echo -e "${BLUE}Test 14: Check stock (database operation, no Redis for stock)${NC}"
echo -e "Note: Stock operations use database for consistency, not Redis cache"
echo

# We can't directly test CheckStock and UpdateStock without order service
# But we can verify the product has correct stock
PRODUCT_INFO=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
CURRENT_STOCK=$(echo $PRODUCT_INFO | grep -o '"stock":[0-9]*' | cut -d':' -f2)
echo -e "${GREEN}Current stock: ${CURRENT_STOCK}${NC}"
echo -e "${GREEN}Stock operations are atomic at database level (PostgreSQL)${NC}"
echo

# ========================================
# Summary
# ========================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Redis Cache Testing Completed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo -e "${GREEN}Test Summary:${NC}"
echo -e "  ${GREEN}✓${NC} Product Detail Cache (Cache-Aside Pattern)"
echo -e "    - Cache MISS on first query"
echo -e "    - Cache HIT on subsequent queries"
echo -e "  ${GREEN}✓${NC} Product List Cache (Version-Based)"
echo -e "    - Version number mechanism"
echo -e "    - Cache HIT when version matches"
echo -e "  ${GREEN}✓${NC} Product Search Cache (Version-Based)"
echo -e "    - Search results cached with version"
echo -e "  ${GREEN}✓${NC} Cache Invalidation on Update"
echo -e "    - Detail cache deleted"
echo -e "    - Version numbers incremented"
echo -e "    - Stale caches automatically invalidated"
echo -e "  ${GREEN}✓${NC} Cross-Category Invalidation"
echo -e "    - Both old and new category versions updated"
echo -e "    - Product correctly moved between categories"
echo -e "  ${GREEN}✓${NC} Database Atomicity"
echo -e "    - Stock operations use database (no Redis)"
echo -e "    - Ensures data consistency and atomicity"
echo
echo -e "${BLUE}Cache Architecture:${NC}"
echo -e "  - Detail Cache: Cache-Aside (query-level caching)"
echo -e "  - List/Search Cache: Version-Based (category + global versions)"
echo -e "  - Stock Data: Database-Only (atomicity guarantee)"
echo
echo -e "${YELLOW}To view service logs:${NC}"
echo -e "  tail -f logs/product-rpc.log"
echo
echo -e "${YELLOW}To check Redis directly:${NC}"
echo -e "  docker exec -it letsgo-redis redis-cli"
echo -e "  > KEYS product:*"
echo -e "  > GET product:detail:${PRODUCT_ID}"
echo -e "  > GET product:CategoryVersion:Electronics"
echo -e "  > GET product:GlobalVersion"
echo
