#!/bin/bash

set -e

cd "$(dirname "$0")"

echo "Starting GoIP service..."
docker-compose up -d

echo "GoIP service started successfully!"
echo ""
echo "Health check: curl http://localhost:8080/api/v1/health"
echo ""
echo "View logs: docker-compose logs -f"
