#!/bin/bash

# Find all .go files and update the import paths
find . -type f -name "*.go" -exec sed -i '' 's|github\.com/tapas/go-user-service|github\.com/yi-tech/go-user-service|g' {} \;

echo "Import paths updated!"
