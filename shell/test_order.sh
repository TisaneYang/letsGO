#!/bin/bash

# Order Service Testing Script
# This script tests all order-related functionality including create, list, detail, cancel, and query

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
echo -e "${BLUE}Order Service Testing${NC}"
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
# Part 1: Setup - Register User and Get Products
# ========================================
print_section "Part 1: Setup - Authentication and Product Data"

# 1. Register a test user to get auth token
echo -e "${BLUE}1. Registering test user for authentication...${NC}"
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ordertest_'$(date +%s)'",
    "password": "password123",
    "email": "ordertest_'$(date +%s)'@example.com",
    "phone": "137'$(date +%s | tail -c 9)'"
  }')
print_result "User Registration" "$REGISTER_RESULT"

# Extract token
TOKEN=$(echo $REGISTER_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get authentication token, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Got authentication token${NC}"
echo

# 2. Get product list to find products for order
echo -e "${BLUE}2. Getting product list for order items...${NC}"
PRODUCT_LIST=$(curl -s "${BASE_URL}/api/v1/product/list?pageSize=5")
print_result "Product List" "$PRODUCT_LIST"

# Extract first two product IDs
PRODUCT_ID_1=$(echo $PRODUCT_LIST | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
PRODUCT_ID_2=$(echo $PRODUCT_LIST | grep -o '"id":[0-9]*' | head -2 | tail -1 | cut -d':' -f2)

if [ -z "$PRODUCT_ID_1" ] || [ -z "$PRODUCT_ID_2" ]; then
    echo -e "${RED}Failed to get product IDs, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Found product IDs: $PRODUCT_ID_1, $PRODUCT_ID_2${NC}"
echo

# ========================================
# Part 2: Order Creation
# ========================================
print_section "Part 2: Testing Order Creation"

# 3. Test Create Order with single item
echo -e "${BLUE}3. Testing Create Order with single item...${NC}"
CREATE_ORDER_1=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID_1},
        \"quantity\": 2
      }
    ],
    \"address\": \"123 Test Street, Test City, Test Province, 100000\",
    \"phone\": \"13800138000\",
    \"remark\": \"Please deliver in the morning\"
  }")
print_result "Create Order - Single Item" "$CREATE_ORDER_1"

# Extract order ID and order number
ORDER_ID_1=$(echo $CREATE_ORDER_1 | grep -o '"orderId":[0-9]*' | cut -d':' -f2)
ORDER_NO_1=$(echo $CREATE_ORDER_1 | grep -o '"orderNo":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ORDER_ID_1" ]; then
    echo -e "${GREEN}Created order with ID: $ORDER_ID_1, Order No: $ORDER_NO_1${NC}"
else
    echo -e "${RED}Failed to create order${NC}"
fi
echo

# 4. Test Create Order with multiple items
echo -e "${BLUE}4. Testing Create Order with multiple items...${NC}"
CREATE_ORDER_2=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID_1},
        \"quantity\": 1
      },
      {
        \"productId\": ${PRODUCT_ID_2},
        \"quantity\": 3
      }
    ],
    \"address\": \"456 Another Street, Another City, Another Province, 200000\",
    \"phone\": \"13900139000\",
    \"remark\": \"Handle with care\"
  }")
print_result "Create Order - Multiple Items" "$CREATE_ORDER_2"

# Extract second order ID and order number
ORDER_ID_2=$(echo $CREATE_ORDER_2 | grep -o '"orderId":[0-9]*' | cut -d':' -f2)
ORDER_NO_2=$(echo $CREATE_ORDER_2 | grep -o '"orderNo":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ORDER_ID_2" ]; then
    echo -e "${GREEN}Created order with ID: $ORDER_ID_2, Order No: $ORDER_NO_2${NC}"
else
    echo -e "${RED}Failed to create second order${NC}"
fi
echo

# ========================================
# Part 3: Order Query Operations
# ========================================
print_section "Part 3: Testing Order Query Operations"

# 5. Test Get Order Detail
if [ -n "$ORDER_ID_1" ]; then
    echo -e "${BLUE}5. Testing Get Order Detail (ID: $ORDER_ID_1)...${NC}"
    ORDER_DETAIL=$(curl -s "${BASE_URL}/api/v1/order/detail/${ORDER_ID_1}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Get Order Detail" "$ORDER_DETAIL"
else
    echo -e "${RED}5. Skipping Get Order Detail - No order ID available${NC}"
    echo
fi

# 6. Test Query Order by Order Number
if [ -n "$ORDER_NO_1" ]; then
    echo -e "${BLUE}6. Testing Query Order by Order Number (${ORDER_NO_1})...${NC}"
    QUERY_BY_NO=$(curl -s "${BASE_URL}/api/v1/order/query/${ORDER_NO_1}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Query Order by Number" "$QUERY_BY_NO"
else
    echo -e "${RED}6. Skipping Query by Order Number - No order number available${NC}"
    echo
fi

# 7. Test List Orders (default pagination)
echo -e "${BLUE}7. Testing List Orders (default pagination)...${NC}"
ORDER_LIST=$(curl -s "${BASE_URL}/api/v1/order/list" \
  -H "Authorization: Bearer $TOKEN")
print_result "List Orders - Default" "$ORDER_LIST"

# 8. Test List Orders with pagination
echo -e "${BLUE}8. Testing List Orders with pagination...${NC}"
ORDER_LIST_PAGE=$(curl -s "${BASE_URL}/api/v1/order/list?page=1&pageSize=5" \
  -H "Authorization: Bearer $TOKEN")
print_result "List Orders - Page 1, Size 5" "$ORDER_LIST_PAGE"

# 9. Test List Orders with status filter (status=1 means pending)
echo -e "${BLUE}9. Testing List Orders with status filter (Pending)...${NC}"
ORDER_LIST_STATUS=$(curl -s "${BASE_URL}/api/v1/order/list?status=1&page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN")
print_result "List Orders - Status: Pending" "$ORDER_LIST_STATUS"

# ========================================
# Part 4: Order Cancellation
# ========================================
print_section "Part 4: Testing Order Cancellation"

# 10. Test Cancel Order
if [ -n "$ORDER_ID_2" ]; then
    echo -e "${BLUE}10. Testing Cancel Order (ID: $ORDER_ID_2)...${NC}"
    CANCEL_RESULT=$(curl -s -X PUT "${BASE_URL}/api/v1/order/cancel/${ORDER_ID_2}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Cancel Order" "$CANCEL_RESULT"

    # 11. Verify order cancellation
    echo -e "${BLUE}11. Verifying order cancellation...${NC}"
    VERIFY_CANCEL=$(curl -s "${BASE_URL}/api/v1/order/detail/${ORDER_ID_2}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Verify Cancelled Order" "$VERIFY_CANCEL"
else
    echo -e "${RED}10-11. Skipping Cancel Order tests - No order ID available${NC}"
    echo
fi

# ========================================
# Part 5: Edge Cases and Error Handling
# ========================================
print_section "Part 5: Testing Edge Cases and Error Handling"

# 12. Test Create Order without authentication
echo -e "${BLUE}12. Testing Create Order without authentication...${NC}"
NO_AUTH_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID_1},
        \"quantity\": 1
      }
    ],
    \"address\": \"Test Address\",
    \"phone\": \"13800138000\"
  }")
print_result "Create Order - No Auth" "$NO_AUTH_ORDER"

# 13. Test Create Order with invalid product ID
echo -e "${BLUE}13. Testing Create Order with invalid product ID...${NC}"
INVALID_PRODUCT_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "productId": 999999999,
        "quantity": 1
      }
    ],
    "address": "Test Address for Invalid Product",
    "phone": "13800138000"
  }')
print_result "Create Order - Invalid Product" "$INVALID_PRODUCT_ORDER"

# 14. Test Create Order with empty items
echo -e "${BLUE}14. Testing Create Order with empty items...${NC}"
EMPTY_ITEMS_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [],
    "address": "Test Address",
    "phone": "13800138000"
  }')
print_result "Create Order - Empty Items" "$EMPTY_ITEMS_ORDER"

# 15. Test Create Order with invalid phone number
echo -e "${BLUE}15. Testing Create Order with invalid phone number...${NC}"
INVALID_PHONE_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID_1},
        \"quantity\": 1
      }
    ],
    \"address\": \"Test Address\",
    \"phone\": \"123\"
  }")
print_result "Create Order - Invalid Phone" "$INVALID_PHONE_ORDER"

# 16. Test Get Order Detail with invalid ID
echo -e "${BLUE}16. Testing Get Order Detail with invalid ID...${NC}"
INVALID_ORDER_DETAIL=$(curl -s "${BASE_URL}/api/v1/order/detail/999999999" \
  -H "Authorization: Bearer $TOKEN")
print_result "Get Order Detail - Invalid ID" "$INVALID_ORDER_DETAIL"

# 17. Test Query Order by invalid order number
echo -e "${BLUE}17. Testing Query Order by invalid order number...${NC}"
INVALID_ORDER_NO=$(curl -s "${BASE_URL}/api/v1/order/query/INVALID123456" \
  -H "Authorization: Bearer $TOKEN")
print_result "Query Order - Invalid Number" "$INVALID_ORDER_NO"

# 18. Test Cancel already cancelled order
if [ -n "$ORDER_ID_2" ]; then
    echo -e "${BLUE}18. Testing Cancel already cancelled order...${NC}"
    DOUBLE_CANCEL=$(curl -s -X PUT "${BASE_URL}/api/v1/order/cancel/${ORDER_ID_2}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Cancel Order - Already Cancelled" "$DOUBLE_CANCEL"
else
    echo -e "${RED}18. Skipping double cancel test - No order ID available${NC}"
    echo
fi

# ========================================
# Summary
# ========================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}All Order Service Tests Completed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo -e "${GREEN}Tests Executed:${NC}"
echo -e "  - Order Creation: Single item, Multiple items"
echo -e "  - Order Query: Detail, List, Query by Number"
echo -e "  - Order Cancellation: Cancel and Verify"
echo -e "  - Edge Cases: Invalid IDs, Missing Auth, Invalid Data"
echo
echo -e "${BLUE}Note: Check the response codes and messages above to verify each test${NC}"
