#!/bin/bash
set -e

# Create necessary directories
mkdir -p api/swagger
mkdir -p internal/transport/grpc/auth
mkdir -p internal/transport/grpc/user
mkdir -p internal/transport/http/auth
mkdir -p internal/transport/http/user
mkdir -p cmd/server

# Move proto files
echo "Moving proto files..."
cp -r proto/* api/proto/
# We'll keep the original proto directory until everything is verified

# Move transport layer
echo "Moving transport layer..."
cp -r internal/handler/grpc/auth/* internal/transport/grpc/auth/
cp -r internal/handler/grpc/user/* internal/transport/grpc/user/
cp -r internal/handler/http/auth/* internal/transport/http/auth/
cp -r internal/handler/http/user/* internal/transport/http/user/

# Move grpc server implementation
echo "Moving gRPC server implementation..."
cp internal/grpc/server.go internal/transport/grpc/
# Check if config file exists before copying
if [ -f "internal/grpc/config.go" ]; then
  cp internal/grpc/config.go internal/transport/grpc/
fi

# Move http server implementation
echo "Moving HTTP server implementation..."
if [ -f "internal/http/server.go" ]; then
  cp internal/http/server.go internal/transport/http/
fi
if [ -f "internal/http/router.go" ]; then
  cp internal/http/router.go internal/transport/http/
fi

# Move cmd/app to cmd/server
echo "Moving command structure..."
cp -r cmd/app/* cmd/server/

# Move swagger docs
echo "Moving swagger documentation..."
if [ -d "docs/swagger" ]; then
  cp -r docs/swagger/* api/swagger/
fi

echo "Structure reorganization script completed."
echo "Please review the changes before removing the original directories."
echo "After verification, update import paths in all Go files."
