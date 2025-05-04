#!/bin/bash

# Create application directory
APP_DIR="$HOME/.local/share/timesheetz"
mkdir -p "$APP_DIR"

# Copy the binary and config
cp "build/timesheetz-linux-amd64" "$APP_DIR/timesheetz"
cp config.json "$APP_DIR/"

# Create systemd user service
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
mkdir -p "$SYSTEMD_USER_DIR"

cat > "$SYSTEMD_USER_DIR/timesheetz.service" << EOF
[Unit]
Description=Timesheetz Service
After=network.target

[Service]
Type=simple
ExecStart=$APP_DIR/timesheetz
WorkingDirectory=$APP_DIR
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
EOF

# Enable and start the service
systemctl --user enable timesheetz.service
systemctl --user start timesheetz.service

echo "Installation complete! Timesheetz has been installed to $APP_DIR and will start automatically on login." 