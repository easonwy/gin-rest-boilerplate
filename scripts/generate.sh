#!/bin/bash

set -e

# Create necessary directories
mkdir -p api/proto/user/v1
mkdir -p docs/swagger

# Generate Go code from protobuf
protoc -I api/proto -I third_party \
  --go_out=api/proto \
  --go_opt=paths=source_relative \
  --go-grpc_out=api/proto \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=logtostderr=true:api/proto \
  --grpc-gateway_opt paths=source_relative \
  --openapiv2_out=docs/swagger \
  --openapiv2_opt logtostderr=true \
  --openapiv2_opt generate_unbound_methods=true \
  --openapiv2_opt allow_merge=true \
  --openapiv2_opt merge_file_name=user_service \
  api/proto/user/v1/user.proto

echo "Code generation complete!"
