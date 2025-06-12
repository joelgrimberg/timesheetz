#!/bin/bash

# Exit on error
set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Timesheetz service update...${NC}"

# Step 1: Build the new version
echo -e "\n${GREEN}Building new version...${NC}"
"$PROJECT_ROOT/scripts/build.sh"

# Step 2: Stop the current service
echo -e "\n${GREEN}Stopping current service...${NC}"
if launchctl list | grep -q "com.timesheetz"; then
    launchctl unload ~/Library/LaunchAgents/com.timesheetz.plist
    echo "Service stopped"
else
    echo "Service was not running"
fi

# Step 3: Copy the new binary
echo -e "\n${GREEN}Installing new binary...${NC}"
sudo cp "$PROJECT_ROOT/bin/timesheet" /usr/local/bin/timesheetz
echo "Binary installed"

# Step 4: Start the service
echo -e "\n${GREEN}Starting service...${NC}"
launchctl load ~/Library/LaunchAgents/com.timesheetz.plist
echo "Service started"

# Step 5: Verify the service is running
echo -e "\n${GREEN}Verifying service status...${NC}"
if launchctl list | grep -q "com.timesheetz"; then
    echo -e "${GREEN}Service is running successfully!${NC}"
else
    echo -e "${RED}Service failed to start!${NC}"
    exit 1
fi

# Step 6: Test the API health check
echo -e "\n${GREEN}Testing API health check...${NC}"
max_retries=5
retry_count=0
while [ $retry_count -lt $max_retries ]; do
    if curl -s http://localhost:8080/health | grep -q '"status":"ok"'; then
        echo -e "${GREEN}API is healthy!${NC}"
        break
    fi
    retry_count=$((retry_count + 1))
    if [ $retry_count -lt $max_retries ]; then
        echo "Waiting for API to start... (attempt $retry_count of $max_retries)"
        sleep 2
    else
        echo -e "${RED}API health check failed after $max_retries attempts!${NC}"
        exit 1
    fi
done

echo -e "\n${GREEN}Update completed successfully!${NC}" 