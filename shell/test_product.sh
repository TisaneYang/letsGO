#!/bin/bash

# Product Service Testing Script
# This script tests all product-related functionality including public and admin endpoints

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
echo -e "${BLUE}Product Service Testing${NC}"
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

# ========================================
# Part 1: Public Product Endpoints (No Auth Required)
# ========================================
print_section "Part 1: Testing Public Product Endpoints"

# 1. Test List Products (default pagination)
echo -e "${BLUE}1. Testing List Products (default pagination)...${NC}"
LIST_RESULT=$(curl -s "${BASE_URL}/api/v1/product/list")
print_result "List Products - Default" "$LIST_RESULT"

# Extract first product ID for later tests
PRODUCT_ID=$(echo $LIST_RESULT | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo -e "${GREEN}Found product ID: $PRODUCT_ID${NC}"
echo

# 2. Test List Products with pagination
echo -e "${BLUE}2. Testing List Products with pagination...${NC}"
LIST_PAGE_RESULT=$(curl -s "${BASE_URL}/api/v1/product/list?page=1&pageSize=5")
print_result "List Products - Page 1, Size 5" "$LIST_PAGE_RESULT"

# 3. Test List Products with category filter
echo -e "${BLUE}3. Testing List Products with category filter...${NC}"
LIST_CATEGORY_RESULT=$(curl -s "${BASE_URL}/api/v1/product/list?category=Electronics&page=1&pageSize=10")
print_result "List Products - Electronics Category" "$LIST_CATEGORY_RESULT"

# 4. Test List Products with sorting
echo -e "${BLUE}4. Testing List Products with sorting by price...${NC}"
LIST_SORT_RESULT=$(curl -s "${BASE_URL}/api/v1/product/list?sortBy=price&order=asc&pageSize=5")
print_result "List Products - Sorted by Price ASC" "$LIST_SORT_RESULT"

# 5. Test Get Product Detail
if [ -n "$PRODUCT_ID" ]; then
    echo -e "${BLUE}5. Testing Get Product Detail (ID: $PRODUCT_ID)...${NC}"
    DETAIL_RESULT=$(curl -s "${BASE_URL}/api/v1/product/detail/${PRODUCT_ID}")
    print_result "Get Product Detail" "$DETAIL_RESULT"
else
    echo -e "${RED}5. Skipping Get Product Detail - No product ID available${NC}"
    echo
fi

# 6. Test Search Products
echo -e "${BLUE}6. Testing Search Products...${NC}"
SEARCH_RESULT=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=phone&page=1&pageSize=10")
print_result "Search Products - Keyword: phone" "$SEARCH_RESULT"

# 7. Test Search Products with different keyword
echo -e "${BLUE}7. Testing Search Products with another keyword...${NC}"
SEARCH_RESULT2=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=laptop&page=1&pageSize=10")
print_result "Search Products - Keyword: laptop" "$SEARCH_RESULT2"

# ========================================
# Part 2: Admin Product Endpoints (Auth Required)
# ========================================
print_section "Part 2: Testing Admin Product Endpoints (Requires Auth)"

# First, register a test user to get auth token
echo -e "${BLUE}Registering test user for authentication...${NC}"
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "producttest_'$(date +%s)'",
    "password": "password123",
    "email": "producttest_'$(date +%s)'@example.com",
    "phone": "136'$(date +%s | tail -c 9)'"
  }')
print_result "User Registration for Auth" "$REGISTER_RESULT"

# Extract token
TOKEN=$(echo $REGISTER_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get authentication token, skipping admin tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}Public tests completed!${NC}"
    echo -e "${BLUE}========================================${NC}"
    exit 0
fi

echo -e "${GREEN}Got authentication token${NC}"
echo

# 8. Test Add Product
echo -e "${BLUE}8. Testing Add Product (Admin)...${NC}"
TIMESTAMP=$(date +%s)
CURRENT_DATE=$(date)
ADD_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/product/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Product ${TIMESTAMP}\",
    \"description\": \"This is a test product created by the testing script at ${CURRENT_DATE}\",
    \"price\": 99.99,
    \"stock\": 100,
    \"category\": \"Electronics\",
    \"images\": [
      \"https://example.com/image1.jpg\",
      \"https://example.com/image2.jpg\"
    ],
    \"attributes\": \"{\\\"color\\\": \\\"black\\\", \\\"brand\\\": \\\"TestBrand\\\", \\\"weight\\\": \\\"500g\\\"}\"
  }")
print_result "Add Product" "$ADD_RESULT"

# Extract new product ID
NEW_PRODUCT_ID=$(echo $ADD_RESULT | grep -o '"productId":[0-9]*' | cut -d':' -f2)

if [ -n "$NEW_PRODUCT_ID" ]; then
    echo -e "${GREEN}Created product with ID: $NEW_PRODUCT_ID${NC}"
    echo

    # 9. Verify the new product was created
    echo -e "${BLUE}9. Verifying newly created product...${NC}"
    VERIFY_NEW_RESULT=$(curl -s "${BASE_URL}/api/v1/product/detail/${NEW_PRODUCT_ID}")
    print_result "Verify New Product" "$VERIFY_NEW_RESULT"

    # 10. Test Update Product
    echo -e "${BLUE}10. Testing Update Product (Admin)...${NC}"
    UPDATE_RESULT=$(curl -s -X PUT ${BASE_URL}/api/v1/product/update \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"id\": ${NEW_PRODUCT_ID},
        \"name\": \"Updated Test Product ${TIMESTAMP}\",
        \"description\": \"This product has been updated by the testing script\",
        \"price\": 149.99,
        \"stock\": 150,
        \"category\": \"Electronics\",
        \"images\": [
          \"https://example.com/updated1.jpg\",
          \"https://example.com/updated2.jpg\",
          \"https://example.com/updated3.jpg\"
        ],
        \"attributes\": \"{\\\"color\\\": \\\"silver\\\", \\\"brand\\\": \\\"TestBrand\\\", \\\"weight\\\": \\\"450g\\\", \\\"warranty\\\": \\\"2 years\\\"}\"
      }")
    print_result "Update Product" "$UPDATE_RESULT"

    # 11. Verify the product update
    echo -e "${BLUE}11. Verifying product update...${NC}"
    VERIFY_UPDATE_RESULT=$(curl -s "${BASE_URL}/api/v1/product/detail/${NEW_PRODUCT_ID}")
    print_result "Verify Updated Product" "$VERIFY_UPDATE_RESULT"

    # 12. Test searching for the new product
    echo -e "${BLUE}12. Testing search for newly created product...${NC}"
    SEARCH_NEW_RESULT=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=Updated%20Test%20Product")
    print_result "Search for New Product" "$SEARCH_NEW_RESULT"
else
    echo -e "${RED}Failed to create product, skipping update tests${NC}"
    echo
fi

# ========================================
# Part 3: Edge Cases and Error Handling
# ========================================
print_section "Part 3: Testing Edge Cases and Error Handling"

# 13. Test invalid product ID
echo -e "${BLUE}13. Testing Get Product with invalid ID...${NC}"
INVALID_RESULT=$(curl -s "${BASE_URL}/api/v1/product/detail/999999999")
print_result "Get Non-existent Product" "$INVALID_RESULT"

# 14. Test search with empty keyword
echo -e "${BLUE}14. Testing Search with empty keyword...${NC}"
EMPTY_SEARCH_RESULT=$(curl -s "${BASE_URL}/api/v1/product/search?keyword=")
print_result "Search with Empty Keyword" "$EMPTY_SEARCH_RESULT"

# 15. Test add product without authentication
echo -e "${BLUE}15. Testing Add Product without authentication...${NC}"
NO_AUTH_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/product/add \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Unauthorized Product",
    "description": "This should fail",
    "price": 50.00,
    "stock": 10,
    "category": "Test",
    "images": ["https://example.com/test.jpg"]
  }')
print_result "Add Product - No Auth" "$NO_AUTH_RESULT"

# 16. Test add product with invalid data
echo -e "${BLUE}16. Testing Add Product with invalid price...${NC}"
INVALID_ADD_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/product/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Product",
    "description": "This has negative price",
    "price": -10.00,
    "stock": 10,
    "category": "Test",
    "images": ["https://example.com/test.jpg"]
  }')
print_result "Add Product - Invalid Price" "$INVALID_ADD_RESULT"

# 17. Test list products with invalid pagination
echo -e "${BLUE}17. Testing List Products with invalid page number...${NC}"
INVALID_PAGE_RESULT=$(curl -s "${BASE_URL}/api/v1/product/list?page=0&pageSize=10")
print_result "List Products - Invalid Page" "$INVALID_PAGE_RESULT"

# ========================================
# Summary
# ========================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}All Product Service Tests Completed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo -e "${GREEN}Tests Executed:${NC}"
echo -e "  - Public Endpoints: List, Detail, Search"
echo -e "  - Admin Endpoints: Add, Update"
echo -e "  - Edge Cases: Invalid IDs, Missing Auth, Invalid Data"
echo
echo -e "${BLUE}Note: Check the response codes and messages above to verify each test${NC}"
