#!/bin/bash

# Simple Kafka Event Test
# Tests if order creation events are properly published to Kafka

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Unset proxy
unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY

BASE_URL="http://localhost:8888"
KAFKA_CONTAINER="letsgo-kafka"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Kafka Event Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# 1. Register user
echo -e "${BLUE}1. Registering test user...${NC}"
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "kafkatest_'$(date +%s)'",
    "password": "password123",
    "email": "kafkatest_'$(date +%s)'@example.com",
    "phone": "138'$(date +%s | tail -c 9)'"
  }')

TOKEN=$(echo $REGISTER_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}✓ User registered, got token${NC}"
echo

# 2. Get product
echo -e "${BLUE}2. Getting product...${NC}"
PRODUCT_LIST=$(curl -s "${BASE_URL}/api/v1/product/list?pageSize=1")
PRODUCT_ID=$(echo $PRODUCT_LIST | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo -e "${GREEN}✓ Got product ID: $PRODUCT_ID${NC}"
echo

# 3. Create order
echo -e "${BLUE}3. Creating order...${NC}"
CREATE_ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"items\": [{\"productId\": ${PRODUCT_ID}, \"quantity\": 1}],
    \"address\": \"Test Address\",
    \"phone\": \"13800138000\"
  }")

ORDER_NO=$(echo $CREATE_ORDER | grep -o '"orderNo":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}✓ Order created: $ORDER_NO${NC}"
echo "$CREATE_ORDER" | python3 -m json.tool
echo

# 4. Wait for Kafka event
echo -e "${BLUE}4. Waiting for Kafka event (3 seconds)...${NC}"
sleep 3

# 5. Check Kafka
echo -e "${BLUE}5. Checking Kafka for order event...${NC}"
MESSAGES=$(docker exec $KAFKA_CONTAINER kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic order.created \
    --from-beginning \
    --max-messages 100 \
    --timeout-ms 5000 2>/dev/null || echo "")

if echo "$MESSAGES" | grep -q "$ORDER_NO"; then
    echo -e "${GREEN}✓✓✓ SUCCESS! Found order event in Kafka${NC}"
    echo -e "${BLUE}Event details:${NC}"
    echo "$MESSAGES" | grep "$ORDER_NO" | python3 -m json.tool
else
    echo -e "${RED}✗✗✗ FAILED! Order event not found in Kafka${NC}"
    echo -e "${YELLOW}Last 3 messages in topic:${NC}"
    echo "$MESSAGES" | tail -3
fi

echo
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Test Complete${NC}"
echo -e "${BLUE}========================================${NC}"
