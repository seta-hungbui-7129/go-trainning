#!/bin/bash

# Build the Go application
echo "Building the application..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Build the application
go build -o bin/server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "Build successful! Binary created at: bin/server"
else
    echo "Build failed!"
    exit 1
fi
