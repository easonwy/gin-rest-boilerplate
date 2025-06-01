#!/bin/bash

# Exit on error
set -e

# Install protoc-gen-go and protoc-gen-grpc-gateway if not already installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Create temporary directory for third-party proto files
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download required Google API proto files
mkdir -p $TMP_DIR/google/api
curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > $TMP_DIR/google/api/annotations.proto
curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > $TMP_DIR/google/api/http.proto

# Generate user proto (original)
protoc -I. -I$TMP_DIR \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/user/user.proto

# Generate auth proto (original)
protoc -I. -I$TMP_DIR \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/auth/auth.proto

# Generate user proto v1 with gRPC-Gateway
protoc -I. -I$TMP_DIR \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  --openapiv2_out=./api/swagger --openapiv2_opt=logtostderr=true \
  api/proto/user/v1/user.proto

# Generate auth proto v1 with gRPC-Gateway
protoc -I. -I$TMP_DIR \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  --openapiv2_out=./api/swagger --openapiv2_opt=logtostderr=true \
  api/proto/auth/v1/auth.proto

echo "Proto generation completed successfully"
