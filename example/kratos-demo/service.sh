#!/bin/bash

# kratos-demo - Service Startup Script
# This script starts both weather and message services in the container

set -e  # Exit on error

echo "Starting kratos-demo services..."

# Start weather service (port 8080)
echo "Starting weather service on port 8080..."
/app/weather &

# Start message service with nohup (port 8081)
echo "Starting message service on port 8081..."
nohup /app/message

echo "All services started successfully!"