#!/bin/bash

set -e

cd "$(dirname "$0")"

echo "Stopping Redis service..."
docker-compose down

echo "Redis service stopped successfully!"
