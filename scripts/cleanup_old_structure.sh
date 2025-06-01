#!/bin/bash

# Exit on error
set -e

echo "Cleaning up old directories..."

# Remove old cmd/app directory
if [ -d "cmd/app" ]; then
  echo "Removing cmd/app directory..."
  rm -rf cmd/app
fi

# Remove old internal/handler directory
if [ -d "internal/handler" ]; then
  echo "Removing internal/handler directory..."
  rm -rf internal/handler
fi

# Remove old internal/grpc directory
if [ -d "internal/grpc" ]; then
  echo "Removing internal/grpc directory..."
  rm -rf internal/grpc
fi

# Remove old proto directory
if [ -d "proto" ]; then
  echo "Removing proto directory..."
  rm -rf proto
fi

echo "Cleanup completed successfully!"
