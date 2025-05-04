#!/bin/bash

# Create build directory if it doesn't exist
mkdir -p build

# Build for macOS (arm64 and amd64)
echo "Building for macOS..."
GOOS=darwin GOARCH=arm64 go build -o build/timesheetz-mac-arm64 ./cmd/timesheetz
GOOS=darwin GOARCH=amd64 go build -o build/timesheetz-mac-amd64 ./cmd/timesheetz

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/timesheetz-win-amd64.exe ./cmd/timesheetz

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o build/timesheetz-linux-amd64 ./cmd/timesheetz

echo "Build complete! Executables are in the build directory." 