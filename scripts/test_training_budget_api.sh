#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:4001/api"

# Test year
YEAR=2024

echo "Testing Training Budget API..."

# Test GET all entries
echo -e "\n${GREEN}Testing GET all entries...${NC}"
curl -s "${BASE_URL}/training-budget?year=${YEAR}" | jq '.'

# Test POST new entry
echo -e "\n${GREEN}Testing POST new entry...${NC}"
NEW_ENTRY='{
  "date": "2024-03-15",
  "training_name": "API Test Training",
  "hours": 8,
  "cost_without_vat": 1000.00
}'
curl -s -X POST -H "Content-Type: application/json" -d "${NEW_ENTRY}" "${BASE_URL}/training-budget"

# Get the ID of the newly created entry
echo -e "\n${GREEN}Getting the new entry...${NC}"
ENTRY_ID=$(curl -s "${BASE_URL}/training-budget?year=${YEAR}" | jq '.[-1].id')
echo "Created entry ID: ${ENTRY_ID}"

# Test PUT update entry
echo -e "\n${GREEN}Testing PUT update entry...${NC}"
UPDATED_ENTRY='{
  "id": '"${ENTRY_ID}"',
  "date": "2024-03-15",
  "training_name": "Updated API Test Training",
  "hours": 16,
  "cost_without_vat": 2000.00
}'
curl -s -X PUT -H "Content-Type: application/json" -d "${UPDATED_ENTRY}" "${BASE_URL}/training-budget"

# Verify the update
echo -e "\n${GREEN}Verifying the update...${NC}"
curl -s "${BASE_URL}/training-budget?year=${YEAR}" | jq '.'

# Test DELETE entry
echo -e "\n${GREEN}Testing DELETE entry...${NC}"
curl -s -X DELETE "${BASE_URL}/training-budget?id=${ENTRY_ID}"

# Verify the deletion
echo -e "\n${GREEN}Verifying the deletion...${NC}"
curl -s "${BASE_URL}/training-budget?year=${YEAR}" | jq '.'

# Test GET training hours
echo -e "\n${GREEN}Testing GET training hours...${NC}"
curl -s "${BASE_URL}/training-hours?year=${YEAR}" | jq '.'

echo -e "\n${GREEN}All tests completed!${NC}" 