#!/bin/bash

set -e

cd "$(dirname "$0")"

echo "Starting Redis service..."
docker-compose up -d

echo "Redis service started successfully!"
echo ""
echo "Test connection: docker exec goip-redis redis-cli ping"
echo ""
echo "View logs: docker-compose logs -f"
