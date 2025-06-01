#!/bin/bash

# Exit on error
set -e

echo "Fixing duplicate type declarations..."

# Create a backup directory
mkdir -p backup/internal/domain/user

# Backup the conflicting files
cp internal/domain/user/model.go backup/internal/domain/user/
cp internal/domain/user/user.go backup/internal/domain/user/

# Rename the User type in model.go to UserProfile
echo "Renaming User to UserProfile in model.go..."
sed -i '' 's/type User struct/type UserProfile struct/g' internal/domain/user/model.go
sed -i '' 's/func NewUser/func NewUserProfile/g' internal/domain/user/model.go
sed -i '' 's/\*User/\*UserProfile/g' internal/domain/user/model.go
sed -i '' 's/User{/UserProfile{/g' internal/domain/user/model.go

# Update references to User in other files
echo "Updating references to User in other files..."
find internal -type f -name "*.go" -not -path "internal/domain/user/model.go" -not -path "internal/domain/user/user.go" -exec sed -i '' 's/domain\/user\.\*User/domain\/user\.\*UserProfile/g' {} \;
find internal -type f -name "*.go" -not -path "internal/domain/user/model.go" -not -path "internal/domain/user/user.go" -exec sed -i '' 's/domain\/user\.User/domain\/user\.UserProfile/g' {} \;

echo "Duplicate type declarations fixed!"
