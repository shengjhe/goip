#!/bin/bash

set -e

cd "$(dirname "$0")"

echo "Stopping GoIP service..."
docker-compose down

echo "GoIP service stopped successfully!"
