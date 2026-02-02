#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:2208"

# Function to check HTTP status code
check_status() {
    local status=$1
    local step=$2
    
    if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
        echo -e "${GREEN}✓ $step - SUCCESS (HTTP $status)${NC}"
        return 0
    else
        echo -e "${RED}✗ $step - FAILED (HTTP $status)${NC}"
        return 1
    fi
}

# Header
echo "================================================"
echo "LoRaWAN Simulator - Basic Flow Test"
echo "================================================"
echo ""

# Step 1: Add Network Server
echo -e "${YELLOW}Step 1: Adding Network Server 'localhost'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers" \
  -H "Content-Type: application/json" \
  -d '{"name":"localhost"}')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Add Network Server"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 2: Add Gateway
echo -e "${YELLOW}Step 2: Adding Gateway 'AABBCCDDEEFF0011'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/gateways" \
  -H "Content-Type: application/json" \
  -d '{"eui":"AABBCCDDEEFF0011", "discoveryUri":"ws://localhost:3001"}')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Add Gateway"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 3: Connect Gateway
echo -e "${YELLOW}Step 3: Connecting Gateway 'AABBCCDDEEFF0011'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Connect Gateway"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 4: Get Gateway
echo -e "${YELLOW}Step 4: Get Gateway 'AABBCCDDEEFF0011'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/network-servers/localhost/gateways/AABBCCDDEEFF0011")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Get Gateway"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Footer
echo "================================================"
echo -e "${GREEN}✓ All steps completed successfully!${NC}"
echo "================================================"
echo ""
