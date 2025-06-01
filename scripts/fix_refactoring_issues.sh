#!/bin/bash

# Exit on error
set -e

echo "Fixing refactoring issues..."

# Create backup directories
mkdir -p backup/internal/repository/auth
mkdir -p backup/internal/repository/user
mkdir -p backup/internal/service/auth
mkdir -p backup/internal/service/user

# Backup files before making changes
cp internal/repository/auth/*.go backup/internal/repository/auth/
cp internal/repository/user/*.go backup/internal/repository/user/
cp internal/service/auth/*.go backup/internal/service/auth/
cp internal/service/user/*.go backup/internal/service/user/

# Step 1: Fix repository/auth directory - keep auth_repository.go and remove repository.go
echo "Fixing repository/auth directory..."
rm internal/repository/auth/repository.go

# Step 2: Fix repository/user directory - keep user_repository.go and remove repository.go
echo "Fixing repository/user directory..."
rm internal/repository/user/repository.go

# Step 3: Fix service/auth directory - keep auth_service.go and remove service.go
echo "Fixing service/auth directory..."
rm internal/service/auth/service.go

# Step 4: Fix service/user directory - keep user_service.go and remove service.go
echo "Fixing service/user directory..."
rm internal/service/user/service.go

# Step 5: Fix import paths in user_repository.go
echo "Fixing import paths in user_repository.go..."
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' internal/repository/user/user_repository.go
sed -i '' 's/model\./user\./g' internal/repository/user/user_repository.go

# Step 6: Fix import paths in auth_repository.go
echo "Fixing import paths in auth_repository.go..."
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/auth\"/g' internal/repository/auth/auth_repository.go
sed -i '' 's/model\./auth\./g' internal/repository/auth/auth_repository.go

# Step 7: Fix import paths in auth_service.go
echo "Fixing import paths in auth_service.go..."
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/auth\"/g' internal/service/auth/auth_service.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/auth\"/g' internal/service/auth/auth_service.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' internal/service/auth/auth_service.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/user\"/g' internal/service/auth/auth_service.go
sed -i '' 's/model\./auth\./g' internal/service/auth/auth_service.go

# Step 8: Fix import paths in user_service.go
echo "Fixing import paths in user_service.go..."
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' internal/service/user/user_service.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/user\"/g' internal/service/user/user_service.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/dto\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\/dto\"/g' internal/service/user/user_service.go
sed -i '' 's/model\./user\./g' internal/service/user/user_service.go

# Step 9: Update wire_gen.go to use the new import paths
echo "Updating wire_gen.go..."
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/auth\"/g' cmd/server/wire/wire_gen.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/auth\"/g' cmd/server/wire/wire_gen.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/auth\"/g' cmd/server/wire/wire_gen.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' cmd/server/wire/wire_gen.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/user\"/g' cmd/server/wire/wire_gen.go
sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/user\"/g' cmd/server/wire/wire_gen.go

# Step 10: Update wire.go to use the new import paths
echo "Updating wire.go..."
if [ -f "cmd/server/wire/wire.go" ]; then
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/auth\"/g' cmd/server/wire/wire.go
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/auth\"/g' cmd/server/wire/wire.go
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/auth\"/g' cmd/server/wire/wire.go
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' cmd/server/wire/wire.go
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/repository\"/\"github.com\/yi-tech\/go-user-service\/internal\/repository\/user\"/g' cmd/server/wire/wire.go
  sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/user\"/g' cmd/server/wire/wire.go
fi

# Step 11: Update transport layer to use the new import paths
echo "Updating transport layer..."
find internal/transport -type f -name "*.go" -exec sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/auth\"/g' {} \;
find internal/transport -type f -name "*.go" -exec sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/auth\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/auth\"/g' {} \;
find internal/transport -type f -name "*.go" -exec sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/model\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\"/g' {} \;
find internal/transport -type f -name "*.go" -exec sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/dto\"/\"github.com\/yi-tech\/go-user-service\/internal\/domain\/user\/dto\"/g' {} \;
find internal/transport -type f -name "*.go" -exec sed -i '' 's/\"github.com\/yi-tech\/go-user-service\/internal\/user\/service\"/\"github.com\/yi-tech\/go-user-service\/internal\/service\/user\"/g' {} \;

echo "Refactoring issues fixed successfully!"
