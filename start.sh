#!/bin/bash

# Twitch Drops Farmer Start Script

echo "Starting Twitch Drops Farmer..."

# Note: No Twitch app setup required! Uses Android app client ID like TDM

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.24 or higher."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Node.js is not installed. Please install Node.js 18 or higher."
    exit 1
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build frontend
echo "Building frontend..."
cd web

# Install npm dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Build the frontend
echo "Building Vue.js frontend..."
npm run build

# Check if frontend build succeeded
if [ $? -eq 0 ]; then
    echo "Frontend build successful!"
    cd ..
else
    echo "Frontend build failed!"
    exit 1
fi

# Build the Go application
echo "Building Go application..."
go run .

# Check if Go build succeeded
if [ $? -eq 0 ]; then
    echo "Go build successful!"
    echo "Starting server on http://localhost:8080"
    ./twitchdropsfarmer
else
    echo "Go build failed!"
    exit 1
fi