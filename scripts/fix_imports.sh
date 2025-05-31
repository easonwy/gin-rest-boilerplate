#!/bin/bash

# Fix import paths in generated files
find api/proto -type f -name "*.go" -exec sed -i '' 's|"github.com/yi-tech/go-user-service/api/proto/|"github.com/yi-tech/go-user-service/api/proto/|g' {} \;

echo "Import paths fixed!"
