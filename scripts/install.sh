#!/bin/bash

# Exit on error
set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if timesheet executable exists
if [ ! -f "$PROJECT_ROOT/bin/timesheet" ]; then
    echo "Error: timesheet executable not found in $PROJECT_ROOT/bin/"
    echo "Please build the project first using: ./scripts/build.sh"
    exit 1
fi

# Create LaunchAgents directory if it doesn't exist
mkdir -p ~/Library/LaunchAgents

# Create the plist file
cat > ~/Library/LaunchAgents/com.timesheetz.plist << EOL
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.timesheetz</string>
    <key>ProgramArguments</key>
    <array>
        <string>$PROJECT_ROOT/bin/timesheet</string>
        <string>--no-tui</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/Library/Logs/timesheetz.out</string>
    <key>StandardErrorPath</key>
    <string>$HOME/Library/Logs/timesheetz.err</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
EOL

# Create Logs directory if it doesn't exist
mkdir -p ~/Library/Logs

# Load the Launch Agent
launchctl unload ~/Library/LaunchAgents/com.timesheetz.plist 2>/dev/null || true
launchctl load ~/Library/LaunchAgents/com.timesheetz.plist

echo "Installation complete! Timesheetz will now start automatically at login."
echo "To verify it's running, check: launchctl list | grep timesheetz"
echo "To view logs, check: ~/Library/Logs/timesheetz.out"
echo ""
echo "To stop auto-starting: launchctl unload ~/Library/LaunchAgents/com.timesheetz.plist"
echo "To remove the Launch Agent: rm ~/Library/LaunchAgents/com.timesheetz.plist" 