#!/bin/bash

# Run the application
echo "Starting the SETA Training application..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Check if database is running
if ! nc -z localhost 5432; then
    echo "Database is not running. Please start it first with: ./scripts/start-db.sh"
    exit 1
fi

# Run the application
go run cmd/server/main.go
