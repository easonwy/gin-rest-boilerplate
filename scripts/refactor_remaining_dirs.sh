#!/bin/bash

# Exit on error
set -e

echo "Refactoring remaining directories to follow clean architecture..."

# Create necessary directories if they don't exist
mkdir -p internal/domain/auth
mkdir -p internal/domain/user
mkdir -p internal/repository/auth
mkdir -p internal/repository/user
mkdir -p internal/service/auth
mkdir -p internal/service/user
mkdir -p internal/config
mkdir -p internal/provider

# Move files from internal/auth to their respective directories
if [ -d "internal/auth" ]; then
  echo "Processing internal/auth directory..."
  
  # Move model to domain
  if [ -d "internal/auth/model" ]; then
    echo "Moving auth models to domain..."
    cp -r internal/auth/model/* internal/domain/auth/
  fi
  
  # Move repository to repository layer
  if [ -d "internal/auth/repository" ]; then
    echo "Moving auth repository to repository layer..."
    cp -r internal/auth/repository/* internal/repository/auth/
  fi
  
  # Move service to service layer
  if [ -d "internal/auth/service" ]; then
    echo "Moving auth service to service layer..."
    cp -r internal/auth/service/* internal/service/auth/
  fi
fi

# Move files from internal/user to their respective directories
if [ -d "internal/user" ]; then
  echo "Processing internal/user directory..."
  
  # Move model to domain
  if [ -d "internal/user/model" ]; then
    echo "Moving user models to domain..."
    cp -r internal/user/model/* internal/domain/user/
  fi
  
  # Move repository to repository layer
  if [ -d "internal/user/repository" ]; then
    echo "Moving user repository to repository layer..."
    cp -r internal/user/repository/* internal/repository/user/
  fi
  
  # Move service to service layer
  if [ -d "internal/user/service" ]; then
    echo "Moving user service to service layer..."
    cp -r internal/user/service/* internal/service/user/
  fi
  
  # Move DTO to domain
  if [ -d "internal/user/dto" ]; then
    echo "Moving user DTOs to domain..."
    mkdir -p internal/domain/user/dto
    cp -r internal/user/dto/* internal/domain/user/dto/
  fi
fi

# Update import paths in moved files
echo "Updating import paths in moved files..."

# Function to update import paths in a directory
update_imports() {
  local dir=$1
  local old_path=$2
  local new_path=$3
  
  find "$dir" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/$old_path|\"github.com/yi-tech/go-user-service/$new_path|g" {} \;
}

# Update auth imports
update_imports "internal/domain/auth" "internal/auth/model" "internal/domain/auth"
update_imports "internal/repository/auth" "internal/auth/repository" "internal/repository/auth"
update_imports "internal/repository/auth" "internal/auth/model" "internal/domain/auth"
update_imports "internal/service/auth" "internal/auth/service" "internal/service/auth"
update_imports "internal/service/auth" "internal/auth/repository" "internal/repository/auth"
update_imports "internal/service/auth" "internal/auth/model" "internal/domain/auth"

# Update user imports
update_imports "internal/domain/user" "internal/user/model" "internal/domain/user"
update_imports "internal/domain/user/dto" "internal/user/dto" "internal/domain/user/dto"
update_imports "internal/repository/user" "internal/user/repository" "internal/repository/user"
update_imports "internal/repository/user" "internal/user/model" "internal/domain/user"
update_imports "internal/service/user" "internal/user/service" "internal/service/user"
update_imports "internal/service/user" "internal/user/repository" "internal/repository/user"
update_imports "internal/service/user" "internal/user/model" "internal/domain/user"
update_imports "internal/service/user" "internal/user/dto" "internal/domain/user/dto"

# Update imports in other files
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/auth/model|\"github.com/yi-tech/go-user-service/internal/domain/auth|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/auth/repository|\"github.com/yi-tech/go-user-service/internal/repository/auth|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/auth/service|\"github.com/yi-tech/go-user-service/internal/service/auth|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/user/model|\"github.com/yi-tech/go-user-service/internal/domain/user|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/user/dto|\"github.com/yi-tech/go-user-service/internal/domain/user/dto|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/user/repository|\"github.com/yi-tech/go-user-service/internal/repository/user|g" {} \;
find "internal" -type f -name "*.go" -exec sed -i '' "s|\"github.com/yi-tech/go-user-service/internal/user/service|\"github.com/yi-tech/go-user-service/internal/service/user|g" {} \;

echo "Refactoring completed successfully!"
