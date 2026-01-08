#!/bin/bash

# Payment Service Testing Script
# This script tests all payment-related functionality including create, query, callback, and Kafka events

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
KAFKA_CONTAINER="letsgo-kafka"
KAFKA_TOPICS=(
    "payment.success"
    "payment.failed"
)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Payment Service Testing${NC}"
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

# Function to check Kafka messages for a specific topic
check_kafka_messages() {
    local topic=$1
    local search_key=$2
    local timeout=${3:-10}

    echo -e "${BLUE}Checking Kafka topic: $topic for key: $search_key${NC}"

    # Read messages from Kafka topic
    local messages=$(docker exec $KAFKA_CONTAINER kafka-console-consumer \
        --bootstrap-server localhost:9092 \
        --topic $topic \
        --from-beginning \
        --max-messages 100 \
        --timeout-ms ${timeout}000 2>/dev/null || echo "")

    if [ -z "$messages" ]; then
        echo -e "${YELLOW}No messages found in topic $topic${NC}"
        return 1
    fi

    # Count total messages
    local total_count=$(echo "$messages" | wc -l)
    echo -e "${BLUE}Total messages in topic: $total_count${NC}"

    # Search for the key in messages (case-insensitive)
    local found=$(echo "$messages" | grep -i "$search_key" | wc -l)

    if [ "$found" -gt 0 ]; then
        echo -e "${GREEN}✓ Found $found message(s) containing '$search_key' in topic $topic${NC}"
        echo -e "${BLUE}Message details:${NC}"
        echo "$messages" | grep -i "$search_key" | tail -3 | while read -r line; do
            echo "$line" | python3 -m json.tool 2>/dev/null || echo "$line"
        done
        echo
        return 0
    else
        echo -e "${RED}✗ No messages found containing '$search_key' in topic $topic${NC}"
        echo -e "${YELLOW}Showing last 3 messages from topic:${NC}"
        echo "$messages" | tail -3 | while read -r line; do
            echo "$line" | python3 -m json.tool 2>/dev/null || echo "$line"
        done
        echo
        return 1
    fi
}

# Function to verify Kafka event
verify_kafka_event() {
    local event_type=$1
    local search_key=$2
    local topic=$3

    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}Verifying Kafka Event: $event_type${NC}"
    echo -e "${YELLOW}========================================${NC}"

    # Give Kafka a moment to process the event
    sleep 2

    if check_kafka_messages "$topic" "$search_key" 10; then
        echo -e "${GREEN}✓ Kafka event verification PASSED for $event_type${NC}"
        echo
        return 0
    else
        echo -e "${RED}✗ Kafka event verification FAILED for $event_type${NC}"
        echo -e "${YELLOW}This might be normal if the event was published earlier or Kafka is not configured${NC}"
        echo
        return 1
    fi
}

# Function to print section header
print_section() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo
}

# ========================================
# Part 1: Setup - Register User, Get Products, Create Order
# ========================================
print_section "Part 1: Setup - Authentication, Product, and Order Creation"

# 1. Register a test user to get auth token
echo -e "${BLUE}1. Registering test user for authentication...${NC}"
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "paymenttest_'$(date +%s)'",
    "password": "password123",
    "email": "paymenttest_'$(date +%s)'@example.com",
    "phone": "138'$(date +%s | tail -c 9)'"
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

# Extract first product ID
PRODUCT_ID=$(echo $PRODUCT_LIST | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$PRODUCT_ID" ]; then
    echo -e "${RED}Failed to get product ID, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Found product ID: $PRODUCT_ID${NC}"
echo

# 3. Create an order for payment testing
echo -e "${BLUE}3. Creating an order for payment testing...${NC}"
CREATE_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID},
        \"quantity\": 2
      }
    ],
    \"address\": \"123 Payment Test Street, Test City, 100000\",
    \"phone\": \"13800138000\",
    \"remark\": \"Payment test order\"
  }")
print_result "Create Order" "$CREATE_ORDER"

# Extract order ID and total amount
ORDER_ID=$(echo $CREATE_ORDER | grep -o '"orderId":[0-9]*' | cut -d':' -f2)
ORDER_NO=$(echo $CREATE_ORDER | grep -o '"orderNo":"[^"]*"' | cut -d'"' -f4)
TOTAL_AMOUNT=$(echo $CREATE_ORDER | grep -o '"totalAmount":[0-9.]*' | cut -d':' -f2)

if [ -z "$ORDER_ID" ] || [ -z "$TOTAL_AMOUNT" ]; then
    echo -e "${RED}Failed to create order, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Created order with ID: $ORDER_ID, Order No: $ORDER_NO, Amount: $TOTAL_AMOUNT${NC}"
echo

# ========================================
# Part 2: Payment Creation
# ========================================
print_section "Part 2: Testing Payment Creation"

# 4. Test Create Payment with Alipay
echo -e "${BLUE}4. Testing Create Payment with Alipay (type=1)...${NC}"
CREATE_PAYMENT=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"orderId\": ${ORDER_ID},
    \"paymentType\": 1,
    \"amount\": ${TOTAL_AMOUNT}
  }")
print_result "Create Payment - Alipay" "$CREATE_PAYMENT"

# Extract payment info
PAYMENT_ID=$(echo $CREATE_PAYMENT | grep -o '"paymentId":[0-9]*' | cut -d':' -f2)
PAYMENT_NO=$(echo $CREATE_PAYMENT | grep -o '"paymentNo":"[^"]*"' | cut -d'"' -f4)
PAY_URL=$(echo $CREATE_PAYMENT | grep -o '"payUrl":"[^"]*"' | cut -d'"' -f4)

if [ -n "$PAYMENT_ID" ] && [ -n "$PAYMENT_NO" ]; then
    echo -e "${GREEN}Created payment with ID: $PAYMENT_ID, Payment No: $PAYMENT_NO${NC}"
    echo -e "${GREEN}Payment URL: $PAY_URL${NC}"
else
    echo -e "${RED}Failed to create payment${NC}"
fi
echo

# 5. Test Create Payment for same order (should return existing payment)
echo -e "${BLUE}5. Testing Create Payment for same order (idempotency)...${NC}"
CREATE_PAYMENT_AGAIN=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"orderId\": ${ORDER_ID},
    \"paymentType\": 1,
    \"amount\": ${TOTAL_AMOUNT}
  }")
print_result "Create Payment - Idempotency Test" "$CREATE_PAYMENT_AGAIN"

# ========================================
# Part 3: Payment Query Operations
# ========================================
print_section "Part 3: Testing Payment Query Operations"

# 6. Test Query Payment by Order ID
if [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}6. Testing Query Payment by Order ID (${ORDER_ID})...${NC}"
    QUERY_PAYMENT=$(curl -s "${BASE_URL}/api/v1/payment/query/${ORDER_ID}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Query Payment by Order ID" "$QUERY_PAYMENT"
else
    echo -e "${RED}6. Skipping Query Payment - No order ID available${NC}"
    echo
fi

# ========================================
# Part 4: Payment Callback (Simulating Payment Gateway)
# ========================================
print_section "Part 4: Testing Payment Callback (Simulating Payment Success)"

# 7. Test Payment Callback - Success
if [ -n "$PAYMENT_NO" ] && [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}7. Testing Payment Callback - Success...${NC}"
    echo -e "${YELLOW}Simulating payment gateway callback (Alipay/WeChat)${NC}"
    PAYMENT_CALLBACK=$(curl -s -X POST ${BASE_URL}/api/v1/payment/callback \
      -H "Content-Type: application/json" \
      -d "{
        \"paymentNo\": \"${PAYMENT_NO}\",
        \"orderId\": ${ORDER_ID},
        \"status\": 2,
        \"amount\": ${TOTAL_AMOUNT},
        \"tradeNo\": \"MOCK_TRADE_$(date +%s)\"
      }")
    print_result "Payment Callback - Success" "$PAYMENT_CALLBACK"

    # Verify Kafka event for payment success
    verify_kafka_event "Payment Success" "$PAYMENT_NO" "payment.success"
else
    echo -e "${RED}7. Skipping Payment Callback - No payment info available${NC}"
    echo
fi

# 8. Test Payment Callback - Idempotency (calling again should succeed)
if [ -n "$PAYMENT_NO" ] && [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}8. Testing Payment Callback - Idempotency...${NC}"
    PAYMENT_CALLBACK_AGAIN=$(curl -s -X POST ${BASE_URL}/api/v1/payment/callback \
      -H "Content-Type: application/json" \
      -d "{
        \"paymentNo\": \"${PAYMENT_NO}\",
        \"orderId\": ${ORDER_ID},
        \"status\": 2,
        \"amount\": ${TOTAL_AMOUNT},
        \"tradeNo\": \"MOCK_TRADE_$(date +%s)\"
      }")
    print_result "Payment Callback - Idempotency Test" "$PAYMENT_CALLBACK_AGAIN"
else
    echo -e "${RED}8. Skipping Idempotency Test - No payment info available${NC}"
    echo
fi

# ========================================
# Part 5: Verify Payment and Order Status
# ========================================
print_section "Part 5: Verifying Payment and Order Status After Callback"

# 9. Verify payment status changed to success
if [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}9. Verifying payment status after callback...${NC}"
    VERIFY_PAYMENT=$(curl -s "${BASE_URL}/api/v1/payment/query/${ORDER_ID}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Verify Payment Status" "$VERIFY_PAYMENT"

    # Check if status is 2 (success)
    PAYMENT_STATUS=$(echo $VERIFY_PAYMENT | grep -o '"status":[0-9]*' | cut -d':' -f2)
    if [ "$PAYMENT_STATUS" = "2" ]; then
        echo -e "${GREEN}✓ Payment status is SUCCESS (2)${NC}"
    else
        echo -e "${RED}✗ Payment status is not SUCCESS, got: $PAYMENT_STATUS${NC}"
    fi
    echo
else
    echo -e "${RED}9. Skipping Payment Verification - No order ID available${NC}"
    echo
fi

# 10. Verify order status changed to paid
if [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}10. Verifying order status after payment...${NC}"
    VERIFY_ORDER=$(curl -s "${BASE_URL}/api/v1/order/detail/${ORDER_ID}" \
      -H "Authorization: Bearer $TOKEN")
    print_result "Verify Order Status" "$VERIFY_ORDER"

    # Check if status is 2 (paid)
    ORDER_STATUS=$(echo $VERIFY_ORDER | grep -o '"status":[0-9]*' | head -1 | cut -d':' -f2)
    if [ "$ORDER_STATUS" = "2" ]; then
        echo -e "${GREEN}✓ Order status is PAID (2)${NC}"
    else
        echo -e "${RED}✗ Order status is not PAID, got: $ORDER_STATUS${NC}"
    fi
    echo
else
    echo -e "${RED}10. Skipping Order Verification - No order ID available${NC}"
    echo
fi

# ========================================
# Part 6: Testing Payment Failure Scenario
# ========================================
print_section "Part 6: Testing Payment Failure Scenario"

# 11. Create another order for failure test
echo -e "${BLUE}11. Creating another order for failure test...${NC}"
CREATE_ORDER_2=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [
      {
        \"productId\": ${PRODUCT_ID},
        \"quantity\": 1
      }
    ],
    \"address\": \"456 Failure Test Street, Test City, 200000\",
    \"phone\": \"13900139000\",
    \"remark\": \"Payment failure test order\"
  }")
print_result "Create Order for Failure Test" "$CREATE_ORDER_2"

ORDER_ID_2=$(echo $CREATE_ORDER_2 | grep -o '"orderId":[0-9]*' | cut -d':' -f2)
TOTAL_AMOUNT_2=$(echo $CREATE_ORDER_2 | grep -o '"totalAmount":[0-9.]*' | cut -d':' -f2)

if [ -n "$ORDER_ID_2" ]; then
    echo -e "${GREEN}Created order with ID: $ORDER_ID_2${NC}"

    # 12. Create payment for second order
    echo -e "${BLUE}12. Creating payment for second order...${NC}"
    CREATE_PAYMENT_2=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"orderId\": ${ORDER_ID_2},
        \"paymentType\": 2,
        \"amount\": ${TOTAL_AMOUNT_2}
      }")
    print_result "Create Payment for Failure Test" "$CREATE_PAYMENT_2"

    PAYMENT_NO_2=$(echo $CREATE_PAYMENT_2 | grep -o '"paymentNo":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$PAYMENT_NO_2" ]; then
        # 13. Test Payment Callback - Failure
        echo -e "${BLUE}13. Testing Payment Callback - Failure...${NC}"
        PAYMENT_CALLBACK_FAIL=$(curl -s -X POST ${BASE_URL}/api/v1/payment/callback \
          -H "Content-Type: application/json" \
          -d "{
            \"paymentNo\": \"${PAYMENT_NO_2}\",
            \"orderId\": ${ORDER_ID_2},
            \"status\": 3,
            \"amount\": ${TOTAL_AMOUNT_2},
            \"tradeNo\": \"MOCK_FAIL_$(date +%s)\"
          }")
        print_result "Payment Callback - Failure" "$PAYMENT_CALLBACK_FAIL"

        # Verify Kafka event for payment failure
        verify_kafka_event "Payment Failed" "$PAYMENT_NO_2" "payment.failed"
    fi
else
    echo -e "${RED}11-13. Skipping Failure Test - Failed to create order${NC}"
    echo
fi

# ========================================
# Part 7: Edge Cases and Error Handling
# ========================================
print_section "Part 7: Testing Edge Cases and Error Handling"

# 14. Test Create Payment without authentication
echo -e "${BLUE}14. Testing Create Payment without authentication...${NC}"
NO_AUTH_PAYMENT=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
  -H "Content-Type: application/json" \
  -d "{
    \"orderId\": ${ORDER_ID},
    \"paymentType\": 1,
    \"amount\": 100.00
  }")
print_result "Create Payment - No Auth" "$NO_AUTH_PAYMENT"

# 15. Test Create Payment with invalid order ID
echo -e "${BLUE}15. Testing Create Payment with invalid order ID...${NC}"
INVALID_ORDER_PAYMENT=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "orderId": 999999999,
    "paymentType": 1,
    "amount": 100.00
  }')
print_result "Create Payment - Invalid Order" "$INVALID_ORDER_PAYMENT"

# 16. Test Create Payment with invalid amount
echo -e "${BLUE}16. Testing Create Payment with invalid amount...${NC}"
INVALID_AMOUNT_PAYMENT=$(curl -s -X POST ${BASE_URL}/api/v1/payment/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"orderId\": ${ORDER_ID},
    \"paymentType\": 1,
    \"amount\": 0
  }")
print_result "Create Payment - Invalid Amount" "$INVALID_AMOUNT_PAYMENT"

# 17. Test Query Payment with invalid order ID
echo -e "${BLUE}17. Testing Query Payment with invalid order ID...${NC}"
INVALID_QUERY=$(curl -s "${BASE_URL}/api/v1/payment/query/999999999" \
  -H "Authorization: Bearer $TOKEN")
print_result "Query Payment - Invalid Order ID" "$INVALID_QUERY"

# 18. Test Payment Callback with invalid payment number
echo -e "${BLUE}18. Testing Payment Callback with invalid payment number...${NC}"
INVALID_CALLBACK=$(curl -s -X POST ${BASE_URL}/api/v1/payment/callback \
  -H "Content-Type: application/json" \
  -d '{
    "paymentNo": "INVALID_PAYMENT_NO",
    "orderId": 1,
    "status": 2,
    "amount": 100.00,
    "tradeNo": "INVALID_TRADE"
  }')
print_result "Payment Callback - Invalid Payment No" "$INVALID_CALLBACK"

# 19. Test Payment Callback with amount mismatch
if [ -n "$PAYMENT_NO" ] && [ -n "$ORDER_ID" ]; then
    echo -e "${BLUE}19. Testing Payment Callback with amount mismatch...${NC}"
    AMOUNT_MISMATCH=$(curl -s -X POST ${BASE_URL}/api/v1/payment/callback \
      -H "Content-Type: application/json" \
      -d "{
        \"paymentNo\": \"${PAYMENT_NO}\",
        \"orderId\": ${ORDER_ID},
        \"status\": 2,
        \"amount\": 999999.99,
        \"tradeNo\": \"MISMATCH_TRADE\"
      }")
    print_result "Payment Callback - Amount Mismatch" "$AMOUNT_MISMATCH"
else
    echo -e "${RED}19. Skipping Amount Mismatch Test - No payment info available${NC}"
    echo
fi

# ========================================
# Summary
# ========================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}All Payment Service Tests Completed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo -e "${GREEN}Tests Executed:${NC}"
echo -e "  - Payment Creation: Alipay, WeChat, Idempotency"
echo -e "  - Payment Query: By Order ID"
echo -e "  - Payment Callback: Success, Failure, Idempotency"
echo -e "  - Order Status Update: Verify payment triggers order status change"
echo -e "  - Edge Cases: Invalid IDs, Missing Auth, Invalid Data, Amount Mismatch"
echo -e "  - Kafka Events: Payment Success, Payment Failed"
echo
echo -e "${BLUE}Note: Check the response codes and messages above to verify each test${NC}"
echo

# ========================================
# Part 8: Kafka Event Summary
# ========================================
print_section "Part 8: Kafka Event Summary - Payment Topics"

echo -e "${BLUE}Listing all Kafka topics...${NC}"
TOPICS=$(docker exec $KAFKA_CONTAINER kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null | grep "payment\." || echo "")

if [ -n "$TOPICS" ]; then
    echo -e "${GREEN}Available payment-related topics:${NC}"
    echo "$TOPICS"
    echo

    # Show message count for each payment topic
    for topic in $TOPICS; do
        echo -e "${BLUE}Topic: $topic${NC}"
        MESSAGE_COUNT=$(docker exec $KAFKA_CONTAINER kafka-run-class kafka.tools.GetOffsetShell \
            --broker-list localhost:9092 \
            --topic $topic 2>/dev/null | awk -F ":" '{sum += $3} END {print sum}')

        if [ -n "$MESSAGE_COUNT" ] && [ "$MESSAGE_COUNT" != "0" ]; then
            echo -e "${GREEN}  Total messages: $MESSAGE_COUNT${NC}"

            # Show last 3 messages from this topic
            echo -e "${YELLOW}  Last 3 messages:${NC}"
            docker exec $KAFKA_CONTAINER kafka-console-consumer \
                --bootstrap-server localhost:9092 \
                --topic $topic \
                --from-beginning \
                --max-messages $MESSAGE_COUNT \
                --timeout-ms 5000 2>/dev/null | tail -3 | while read -r line; do
                echo "    $line" | python3 -m json.tool 2>/dev/null || echo "    $line"
            done
        else
            echo -e "${YELLOW}  No messages in this topic${NC}"
        fi
        echo
    done
else
    echo -e "${YELLOW}No payment-related topics found in Kafka${NC}"
    echo -e "${YELLOW}This might indicate that Kafka is not properly configured or no events have been published yet${NC}"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Kafka Event Verification Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
