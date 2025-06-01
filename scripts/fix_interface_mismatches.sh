#!/bin/bash

# Script to fix interface mismatches between domain interfaces and implementations

echo "Fixing interface mismatches between domain and implementation layers..."

# 1. Update HTTP handlers to use implementation interfaces
echo "Updating HTTP handlers..."

# Fix user HTTP handler
if [ -f internal/transport/http/user/handler.go ]; then
  sed -i '' 's/user\.Service/userService\.UserService/g' internal/transport/http/user/handler.go
  echo "Updated user HTTP handler"
fi

# Fix auth HTTP handler
if [ -f internal/transport/http/auth/handler.go ]; then
  sed -i '' 's/auth\.Service/authService\.AuthService/g' internal/transport/http/auth/handler.go
  echo "Updated auth HTTP handler"
fi

# 2. Update gRPC handlers to use implementation interfaces
echo "Updating gRPC handlers..."

# Fix user gRPC handler
if [ -f internal/transport/grpc/user/handler.go ]; then
  # Update the interface type
  sed -i '' 's/user\.Service/userService\.UserService/g' internal/transport/grpc/user/handler.go
  
  # Update method calls to match implementation
  sed -i '' 's/h\.userService\.Register(/h\.userService\.RegisterUser(/g' internal/transport/grpc/user/handler.go
  sed -i '' 's/h\.userService\.GetByID(/h\.userService\.GetUserByID(/g' internal/transport/grpc/user/handler.go
  sed -i '' 's/h\.userService\.GetByEmail(/h\.userService\.GetUserByEmail(/g' internal/transport/grpc/user/handler.go
  
  echo "Updated user gRPC handler"
fi

# Fix auth gRPC handler
if [ -f internal/transport/grpc/auth/handler.go ]; then
  sed -i '' 's/auth\.Service/authService\.AuthService/g' internal/transport/grpc/auth/handler.go
  echo "Updated auth gRPC handler"
fi

# 3. Update middleware to use implementation interfaces
echo "Updating middleware..."
if [ -f internal/middleware/auth.go ]; then
  sed -i '' 's/auth\.Service/authService\.AuthService/g' internal/middleware/auth.go
  echo "Updated auth middleware"
fi

# 4. Update router to use implementation interfaces
echo "Updating router..."
if [ -f internal/transport/http/router.go ]; then
  sed -i '' 's/auth\.Service/authService\.AuthService/g' internal/transport/http/router.go
  echo "Updated HTTP router"
fi

# 5. Fix imports in all updated files
echo "Fixing imports..."
find internal/transport -type f -name "*.go" | xargs -I{} sed -i '' 's/"github\/yi-tech\/go-user-service\/internal\/domain\/user"/"github\/yi-tech\/go-user-service\/internal\/service\/user"/g' {}
find internal/transport -type f -name "*.go" | xargs -I{} sed -i '' 's/"github\/yi-tech\/go-user-service\/internal\/domain\/auth"/"github\/yi-tech\/go-user-service\/internal\/service\/auth"/g' {}
find internal/middleware -type f -name "*.go" | xargs -I{} sed -i '' 's/"github\/yi-tech\/go-user-service\/internal\/domain\/auth"/"github\/yi-tech\/go-user-service\/internal\/service\/auth"/g' {}

echo "Interface mismatches fixed!"
