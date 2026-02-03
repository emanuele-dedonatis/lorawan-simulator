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

# Check if "clean" parameter is provided
if [ "$1" = "clean" ]; then
    # Header
    echo "================================================"
    echo "LoRaWAN Simulator - Cleanup Flow"
    echo "================================================"
    echo ""

    # Step 1: Delete Device
    echo -e "${YELLOW}Step 1: Deleting Device '0011223344556677'...${NC}"
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/network-servers/localhost/devices/0011223344556677")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo "Response: $body"
    check_status "$http_code" "Delete Device"
    echo ""

    # Step 2: Disconnect Gateway
    echo -e "${YELLOW}Step 2: Disconnecting Gateway 'AABBCCDDEEFF0011'...${NC}"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/gateways/AABBCCDDEEFF0011/disconnect")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo "Response: $body"
    check_status "$http_code" "Disconnect Gateway"
    echo ""

    # Step 3: Delete Gateway
    echo -e "${YELLOW}Step 3: Deleting Gateway 'AABBCCDDEEFF0011'...${NC}"
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/network-servers/localhost/gateways/AABBCCDDEEFF0011")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo "Response: $body"
    check_status "$http_code" "Delete Gateway"
    echo ""

    # Step 4: Delete Network Server
    echo -e "${YELLOW}Step 4: Deleting Network Server 'localhost'...${NC}"
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/network-servers/localhost")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo "Response: $body"
    check_status "$http_code" "Delete Network Server"
    echo ""

    # Footer
    echo "================================================"
    echo -e "${GREEN}✓ Cleanup completed!${NC}"
    echo "================================================"
    echo ""
    
    exit 0
fi

# Header
echo "================================================"
echo "LoRaWAN Simulator - Basic Flow Test"
echo "================================================"
echo ""

# Step 0: Health Check
echo -e "${YELLOW}Step 0: Checking API Health...${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/health")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Health Check"
if [ $? -ne 0 ]; then
    echo -e "${RED}API is not responding. Please ensure the simulator is running.${NC}"
    exit 1
fi
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

# Step 5: Add Device
echo -e "${YELLOW}Step 5: Adding Device '0011223344556677'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/devices" \
  -H "Content-Type: application/json" \
  -d '{"deveui":"0011223344556677","joineui":"0011223344556677","appkey":"00112233445566770011223344556677","devnonce":0}')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Add Device"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 6: Get Device
echo -e "${YELLOW}Step 6: Get Device '0011223344556677'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/network-servers/localhost/devices/0011223344556677")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Get Device"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 7: Send Device Join Request
echo -e "${YELLOW}Step 7: Send Device '0011223344556677' Join Request...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/devices/0011223344556677/join")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Send Device Join Request"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Step 8: Get Device
echo -e "${YELLOW}Step 8: Get Device '0011223344556677'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/network-servers/localhost/devices/0011223344556677")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Get Device"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Wait for join to complete
sleep 2

# Step 9: Send Uplink Data
echo -e "${YELLOW}Step 9: Sending Uplink Data from Device '0011223344556677'...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/network-servers/localhost/devices/0011223344556677/uplink")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "Response: $body"
check_status "$http_code" "Send Uplink Data"
if [ $? -ne 0 ]; then
    exit 1
fi
echo ""

# Footer
echo "================================================"
echo -e "${GREEN}✓ All steps completed successfully!${NC}"
echo "================================================"
echo ""
