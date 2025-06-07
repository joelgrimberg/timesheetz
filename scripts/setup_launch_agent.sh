#!/bin/bash

# Get the absolute path of the timesheet executable
TIMESHEET_PATH="$(pwd)/timesheet"

if [ ! -f "$TIMESHEET_PATH" ]; then
    echo "Error: timesheet executable not found at $TIMESHEET_PATH"
    exit 1
fi

# Create the LaunchAgents directory if it doesn't exist
mkdir -p ~/Library/LaunchAgents

# Create the plist file
cat > ~/Library/LaunchAgents/com.timesheetz.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.timesheetz</string>
    <key>ProgramArguments</key>
    <array>
        <string>${TIMESHEET_PATH}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>~/Library/Logs/timesheetz.err</string>
    <key>StandardOutPath</key>
    <string>~/Library/Logs/timesheetz.out</string>
</dict>
</plist>
EOF

# Set the correct permissions
chmod 644 ~/Library/LaunchAgents/com.timesheetz.plist

# Unload if already loaded
launchctl bootout gui/$UID ~/Library/LaunchAgents/com.timesheetz.plist 2>/dev/null || true

# Load the launch agent using bootstrap
launchctl bootstrap gui/$UID ~/Library/LaunchAgents/com.timesheetz.plist

echo "Launch Agent has been set up successfully!"
echo "The application will now start automatically when you log in."
echo "Logs will be written to ~/Library/Logs/timesheetz.out and ~/Library/Logs/timesheetz.err" 