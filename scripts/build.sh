#!/bin/bash

# Exit on error
set -e

# Create build directory
mkdir -p build

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

echo "Build complete! Binaries are in the build directory." 