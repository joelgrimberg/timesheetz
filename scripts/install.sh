#!/bin/bash

# Exit on error
set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to Go architecture
case $ARCH in
    "x86_64")
        ARCH="amd64"
        ;;
    "arm64")
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS to Go OS
case $OS in
    "darwin")
        OS="darwin"
        ;;
    "linux")
        OS="linux"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Create installation directory
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Download and install the appropriate binary
BINARY_NAME="timesheet-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "Installing Timesheet for $OS/$ARCH..."
echo "Binary will be installed to: $INSTALL_DIR"

# Copy the binary to the installation directory
cp "build/$BINARY_NAME" "$INSTALL_DIR/timesheet"
chmod +x "$INSTALL_DIR/timesheet"

echo "Installation complete!"
echo "You can now run 'timesheet' from anywhere in your terminal."
echo "To get started, run: timesheet --help" 