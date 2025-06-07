#!/bin/bash

# Exit on error
set -e

# Create build and bin directories
mkdir -p build
mkdir -p bin

# Build for different platforms
echo "Building for different platforms..."

# macOS (darwin)
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o build/timesheet-darwin-amd64 ./cmd/timesheet
GOOS=darwin GOARCH=arm64 go build -o build/timesheet-darwin-arm64 ./cmd/timesheet

# Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o build/timesheet-linux-amd64 ./cmd/timesheet
GOOS=linux GOARCH=arm64 go build -o build/timesheet-linux-arm64 ./cmd/timesheet

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/timesheet-windows-amd64.exe ./cmd/timesheet

# Create checksums
echo "Creating checksums..."
cd build
shasum -a 256 * > checksums.txt
cd ..

# Copy the appropriate binary to bin directory based on current platform
echo "Creating local binary..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if [[ $(uname -m) == "arm64" ]]; then
        cp build/timesheet-darwin-arm64 bin/timesheet
    else
        cp build/timesheet-darwin-amd64 bin/timesheet
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    if [[ $(uname -m) == "aarch64" ]]; then
        cp build/timesheet-linux-arm64 bin/timesheet
    else
        cp build/timesheet-linux-amd64 bin/timesheet
    fi
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    # Windows
    cp build/timesheet-windows-amd64.exe bin/timesheet.exe
fi

# Make the binary executable
chmod +x bin/timesheet*

echo "Build complete! Binaries are in the build directory and local binary is in bin directory." 