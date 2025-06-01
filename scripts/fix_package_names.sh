#!/bin/bash

# Exit on error
set -e

echo "Fixing package names in refactored files..."

# Fix package names in domain/auth
echo "Fixing package names in domain/auth..."
find internal/domain/auth -type f -name "*.go" -exec sed -i '' 's/^package model$/package auth/g' {} \;

# Fix package names in domain/user
echo "Fixing package names in domain/user..."
find internal/domain/user -type f -name "*.go" -exec sed -i '' 's/^package model$/package user/g' {} \;
find internal/domain/user -type f -name "*.go" -exec sed -i '' 's/^package dto$/package user/g' {} \;

# Fix package names in repository/auth
echo "Fixing package names in repository/auth..."
find internal/repository/auth -type f -name "*.go" -exec sed -i '' 's/^package repository$/package auth/g' {} \;

# Fix package names in repository/user
echo "Fixing package names in repository/user..."
find internal/repository/user -type f -name "*.go" -exec sed -i '' 's/^package repository$/package user/g' {} \;

# Fix package names in service/auth
echo "Fixing package names in service/auth..."
find internal/service/auth -type f -name "*.go" -exec sed -i '' 's/^package service$/package auth/g' {} \;

# Fix package names in service/user
echo "Fixing package names in service/user..."
find internal/service/user -type f -name "*.go" -exec sed -i '' 's/^package service$/package user/g' {} \;

echo "Package names fixed successfully!"
