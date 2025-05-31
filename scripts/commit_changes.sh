#!/bin/bash

# Add all modified files
git add .

# Commit with a descriptive message
git commit -m "feat: Implement gRPC server and Docker configuration

- Add gRPC server implementation with HTTP gateway
- Fix dependency injection issues in wire.go
- Update config.go to properly load application settings
- Expose gRPC ports (50051, 50052) in Dockerfile
- Create docker-compose.yml with PostgreSQL and Redis services
- Add health checks for all services
- Configure proper networking between services"

echo "Changes committed successfully!"
