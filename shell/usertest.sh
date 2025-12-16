#!/bin/bash

# User Service Testing Script
# This script tests all user-related functionality including registration, login, and profile management

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
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
echo -e "${BLUE}User Service Testing${NC}"
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

# 1. Test User Registration
echo -e "${BLUE}1. Testing User Registration...${NC}"
REGISTER_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser_'$(date +%s)'",
    "password": "password123",
    "email": "test_'$(date +%s)'@example.com",
    "phone": "138'$(date +%s | tail -c 9)'"
  }')
print_result "User Registration" "$REGISTER_RESULT"

# Extract token from registration
TOKEN=$(echo $REGISTER_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
USER_ID=$(echo $REGISTER_RESULT | grep -o '"userId":[0-9]*' | cut -d':' -f2)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Registration failed, cannot continue${NC}"
    exit 1
fi

echo -e "${GREEN}Registered with userId: $USER_ID${NC}"
echo

# 2. Test User Login
echo -e "${BLUE}2. Testing User Login...${NC}"
LOGIN_RESULT=$(curl -s -X POST ${BASE_URL}/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob",
    "password": "password123"
  }')
print_result "User Login" "$LOGIN_RESULT"

# Extract token from login
LOGIN_TOKEN=$(echo $LOGIN_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$LOGIN_TOKEN" ]; then
    TOKEN=$LOGIN_TOKEN
    echo -e "${GREEN}Login successful, using new token${NC}"
else
    echo -e "${BLUE}Login failed (may be expected), continuing with registration token${NC}"
fi
echo

# 3. Test Get User Profile
echo -e "${BLUE}3. Testing Get User Profile...${NC}"
PROFILE_RESULT=$(curl -s ${BASE_URL}/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN")
print_result "Get User Profile" "$PROFILE_RESULT"

# 4. Test Update User Profile
echo -e "${BLUE}4. Testing Update User Profile...${NC}"
UPDATE_RESULT=$(curl -s -X PUT ${BASE_URL}/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "updated_'$(date +%s)'@example.com",
    "phone": "139'$(date +%s | tail -c 9)'"
  }')
print_result "Update User Profile" "$UPDATE_RESULT"

# 5. Verify Profile Update
echo -e "${BLUE}5. Verifying Profile Update...${NC}"
VERIFY_RESULT=$(curl -s ${BASE_URL}/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN")
print_result "Verify Updated Profile" "$VERIFY_RESULT"

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${BLUE}========================================${NC}"
