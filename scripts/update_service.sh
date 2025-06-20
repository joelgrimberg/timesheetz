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

# Output folder names for the viewer
echo -e "\n${GREEN}Folder Locations:${NC}"
echo "Script Directory: $SCRIPT_DIR"
echo "Project Root: $PROJECT_ROOT"
echo "Build Directory: $PROJECT_ROOT/build"
echo "Local Binary Directory: $PROJECT_ROOT/bin"
echo "Installed Binary: /usr/local/bin/timesheetz"
echo "LaunchAgent Plist: ~/Library/LaunchAgents/com.timesheetz.plist"

# Step 1: Build the new version
echo -e "\n${GREEN}Building new version...${NC}"
"$PROJECT_ROOT/scripts/build.sh"

# Step 2: Stop the current service
echo -e "\n${GREEN}Stopping current service...${NC}"
if launchctl list | grep -q "com.timesheetz"; then
    launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/com.timesheetz.plist
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

# Step 5.5: Check that the running version matches the built version
echo -e "\n${GREEN}Checking binary version...${NC}"
BUILT_VERSION=$("$PROJECT_ROOT/bin/timesheet" --version)
INSTALLED_VERSION=$("/usr/local/bin/timesheetz" --version)

if [ "$BUILT_VERSION" != "$INSTALLED_VERSION" ]; then
    echo -e "${RED}Version mismatch!${NC}"
    echo "Built version:     $BUILT_VERSION"
    echo "Installed version: $INSTALLED_VERSION"
    exit 1
else
    echo -e "${GREEN}Version matches!${NC}"
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