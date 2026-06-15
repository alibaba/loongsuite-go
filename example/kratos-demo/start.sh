#!/bin/bash

# kratos-demo Service Startup Script
# This script builds the message and weather services and starts the docker containers

set -e  # Exit immediately if a command exits with a non-zero status

echo "Starting kratos-demo Services build..."

# Build message service
echo "Building message service..."
cd ./app/message || { echo "Failed to enter message directory"; exit 1; }
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 otel go build -o ./bin/message ./...
echo "Message service build completed"

# Return to root directory
cd ../../

# Build weather service  
echo "Building weather service..."
cd ./app/weather || { echo "Failed to enter weather directory"; exit 1; }
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 otel go build -o ./bin/weather ./...
echo "Weather service build completed"

# Return to root directory and start docker services
cd ../../
echo "Starting Docker containers..."
docker compose up --build

echo "All services started successfully!"
