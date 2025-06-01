#!/bin/bash
set -e

echo "Updating import paths to match the new structure..."

# Update proto imports
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/proto/|"github.com/yi-tech/go-user-service/api/proto/|g' {} \;

# Update handler imports for grpc
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/internal/handler/grpc/|"github.com/yi-tech/go-user-service/internal/transport/grpc/|g' {} \;

# Update handler imports for http
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/internal/handler/http/|"github.com/yi-tech/go-user-service/internal/transport/http/|g' {} \;

# Update grpc server imports
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/internal/grpc"|"github.com/yi-tech/go-user-service/internal/transport/grpc"|g' {} \;

# Update http server imports
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/internal/http"|"github.com/yi-tech/go-user-service/internal/transport/http"|g' {} \;

# Update cmd/app imports to cmd/server
find . -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/cmd/app/|"github.com/yi-tech/go-user-service/cmd/server/|g' {} \;

echo "Import paths updated!"
echo "Please review the changes and test the application."
echo "You may need to manually fix some imports if they were not caught by this script."
