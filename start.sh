#!/bin/bash

# Twitch Drops Farmer Start Script

echo "Starting Twitch Drops Farmer..."

# Note: No Twitch app setup required! Uses Android app client ID like TDM

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod tidy

# Build the application
echo "Building application..."
go build -o twitchdropsminer

# Check if build succeeded
if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Starting server on http://localhost:8080"
    ./twitchdropsminer
else
    echo "Build failed!"
    exit 1
fi