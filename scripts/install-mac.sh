#!/bin/bash

# Determine the architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    BINARY="timesheetz-mac-arm64"
else
    BINARY="timesheetz-mac-amd64"
fi

# Create application directory
APP_DIR="$HOME/Applications/Timesheetz"
mkdir -p "$APP_DIR"

# Copy the binary and config
cp "build/$BINARY" "$APP_DIR/timesheetz"
cp config.json "$APP_DIR/"

# Create LaunchAgent plist
LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"
mkdir -p "$LAUNCH_AGENT_DIR"

cat > "$LAUNCH_AGENT_DIR/com.timesheetz.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.timesheetz</string>
    <key>ProgramArguments</key>
    <array>
        <string>$APP_DIR/timesheetz</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>$APP_DIR/error.log</string>
    <key>StandardOutPath</key>
    <string>$APP_DIR/output.log</string>
</dict>
</plist>
EOF

# Load the LaunchAgent
launchctl load "$LAUNCH_AGENT_DIR/com.timesheetz.plist"

echo "Installation complete! Timesheetz has been installed to $APP_DIR and will start automatically on login." 