#!/bin/bash

# Script to build sslcheck binaries for multiple platforms
# Usage: ./build.sh [version]

VERSION=${1:-"0.1.0"}
BINARY_NAME="sslcheck"
BUILD_DIR="./dist"

# Create build directory if it doesn't exist
mkdir -p $BUILD_DIR

echo "Building $BINARY_NAME version $VERSION..."

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/${BINARY_NAME}_${VERSION}_linux_amd64" -ldflags="-s -w" .
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o "$BUILD_DIR/${BINARY_NAME}_${VERSION}_linux_arm64" -ldflags="-s -w" .

# Build for macOS (amd64)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/${BINARY_NAME}_${VERSION}_darwin_amd64" -ldflags="-s -w" .
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/${BINARY_NAME}_${VERSION}_darwin_arm64" -ldflags="-s -w" .

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/${BINARY_NAME}_${VERSION}_windows_amd64.exe" -ldflags="-s -w" .

# Create archives for each binary
echo "Creating archives..."
cd $BUILD_DIR || exit

# Check if zip command is available
if command -v zip >/dev/null 2>&1; then
  # Use zip if available
  for file in *; do
    if [ -f "$file" ]; then
      zip "${file}.zip" "$file"
      echo "Created ${file}.zip"
    fi
  done
else
  # Use tar if zip is not available
  echo "zip command not found, using tar instead"
  for file in *; do
    if [ -f "$file" ]; then
      tar -czf "${file}.tar.gz" "$file"
      echo "Created ${file}.tar.gz"
    fi
  done
  echo "Note: To create .zip files, please install the zip utility:"
  echo "  - On Ubuntu/Debian: sudo apt-get install zip"
  echo "  - On CentOS/RHEL: sudo yum install zip"
  echo "  - On macOS: brew install zip"
fi

cd .. || exit

echo "Build complete! Binaries are available in the $BUILD_DIR directory."
