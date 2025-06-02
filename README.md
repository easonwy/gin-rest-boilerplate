# Go 用户授权微服务

## 项目目标

构建一个基于 Gin 框架的 Go 用户授权微服务，为其他微服务及前端提供用户管理和系统鉴权功能。

## 功能模块

1.  **用户管理**
    *   用户注册
    *   用户登录
    *   用户信息查询
    *   用户信息更新
    *   用户删除

2.  **系统鉴权**
    *   基于 Redis + JWT Token 的认证
    *   支持 Refresh Token 机制
    *   统一接口返回结构

## 技术选型

*   **Web 框架**: Gin
*   **数据库**: PostgreSQL
*   **缓存**: Redis
*   **认证**: JWT
*   **依赖注入**: Wire
*   **配置管理**: Viper 或其他
*   **日志**: Zap
*   **单元测试**: Go testing

## 项目结构

项目采用领域驱动设计 (DDD) 和清洁架构 (Clean Architecture) 原则，同时支持 HTTP 和 gRPC 接口。

```
├── api/
│   ├── proto/           # Protocol Buffers 定义
│   │   ├── auth/        # 认证服务 Proto 文件
│   │   │   └── v1/      # v1 版本 API 定义
│   │   └── user/        # 用户服务 Proto 文件
│   │       └── v1/      # v1 版本 API 定义
│   └── swagger/         # Swagger/OpenAPI 规范
├── cmd/
│   └── server/          # 应用程序入口点
│       ├── main.go
│       └── wire/        # 依赖注入配置
├── configs/             # 配置文件
├── internal/            # 私有应用程序代码
│   ├── domain/          # 领域模型和业务逻辑
│   │   ├── auth/        # 认证领域模型
│   │   └── user/        # 用户领域模型
│   ├── repository/      # 数据访问层
│   │   ├── auth/        # 认证数据仓储
│   │   └── user/        # 用户数据仓储
│   ├── service/         # 业务逻辑实现
│   │   ├── auth/        # 认证服务实现
│   │   └── user/        # 用户服务实现
│   ├── transport/       # 传输层
│   │   ├── grpc/        # gRPC 处理器
│   │   │   ├── auth/    # 认证 gRPC 处理器
│   │   │   └── user/    # 用户 gRPC 处理器
│   │   └── http/        # HTTP 处理器 (Gin)
│   │       ├── auth/    # 认证 HTTP 处理器
│   │       └── user/    # 用户 HTTP 处理器
│   ├── middleware/      # 共享中间件
│   ├── config/          # 配置加载和管理
│   └── provider/        # 依赖提供者 (数据库、Redis 等)
├── pkg/                 # 可被其他服务使用的公共库
├── migrations/          # 数据库迁移
├── scripts/             # 实用脚本
│   ├── generate_proto.sh          # 生成 Proto 文件脚本
│   ├── reorganize_structure.sh    # 重组项目结构脚本
│   └── update_structure_imports.sh # 更新导入路径脚本
├── third_party/         # 外部依赖
├── go.mod
├── go.sum
├── Makefile             # 构建自动化
├── Dockerfile
└── README.md
```

## 架构说明

本项目采用领域驱动设计 (DDD) 和清洁架构 (Clean Architecture) 原则，主要分为以下几层：

1. **领域层 (Domain Layer)**：定义核心业务模型、接口和业务规则
2. **服务层 (Service Layer)**：实现领域层定义的业务逻辑
3. **仓储层 (Repository Layer)**：负责数据持久化
4. **传输层 (Transport Layer)**：处理 HTTP 和 gRPC 请求

### 依赖注入

项目使用 Google Wire 进行依赖注入，确保各组件之间的松耦合。依赖注入配置位于 `cmd/server/wire/wire.go`。

### 多协议支持

项目同时支持 HTTP (RESTful API) 和 gRPC 协议：

- **HTTP**：使用 Gin 框架实现
- **gRPC**：使用标准 gRPC 库实现，并通过 grpc-gateway 提供 HTTP/JSON 代理

## 已实现功能

1. **用户管理**
   - 用户注册
   - 用户信息查询
   - 用户信息更新
   - 用户删除

2. **认证系统**
   - 基于 JWT 的认证
   - Refresh Token 机制
   - Redis 会话管理
   - 令牌验证

## 项目重组说明

项目已经完成了结构重组，采用了更清晰的领域驱动设计和清洁架构。主要变化如下：

1. **API 合约管理**
   - 将 Protocol Buffers 定义移至 `/api/proto` 目录
   - 按版本组织 API，添加 v1 版本目录
   - Swagger 文档移至 `/api/swagger` 目录

2. **应用入口点**
   - 将应用入口点从 `/cmd/app` 移至 `/cmd/server`
   - 依赖注入配置位于 `/cmd/server/wire`

3. **传输层分离**
   - 将 HTTP 和 gRPC 处理器移至 `/internal/transport/http` 和 `/internal/transport/grpc`
   - 清晰区分不同协议的处理逻辑

4. **领域模型分离**
   - 将领域模型移至 `/internal/domain`
   - 按功能模块组织领域模型

### 开发者指南

#### 生成 Protocol Buffers 代码

使用以下命令生成 Protocol Buffers 和 gRPC-Gateway 代码：

```bash
./scripts/generate_proto.sh
```

生成的代码将位于相应的 proto 目录中，Swagger 文档将生成在 `/api/swagger` 目录中。

#### 验证 gRPC API

本项目提供了多种方式验证 gRPC API：

1. **使用 gRPC-Gateway REST 接口**

   启动服务后，可以通过 HTTP 请求访问 gRPC-Gateway 提供的 REST 接口：

   ```bash
   # 例如，获取用户信息
   curl -X GET "http://localhost:8081/api/v1/users/1" -H "Authorization: Bearer YOUR_TOKEN"
   ```

   gRPC-Gateway 监听在配置的 `grpc.port + 1` 端口上（默认为 8081）。

#### 使用 Makefile

本项目提供了全面的 Makefile 来简化开发、测试和部署流程。使用 `make help` 查看所有可用命令。

##### 构建与运行

```bash
# 构建服务（包含版本信息）
make build

# 构建并运行服务
make run

# 清理构建产物
make clean
```

##### 测试与代码质量

```bash
# 运行所有测试
make test

# 生成测试覆盖率报告
make test-coverage

# 运行代码格式化
make fmt

# 运行代码静态分析
make vet

# 运行代码质量检查
make lint

# 生成测试用 Mock 对象
make mocks
```

##### 依赖注入

```bash
# 重新生成 Wire 依赖注入代码
make wire
```

##### Protobuf 与 API 文档

```bash
# 安装 Protobuf 相关工具
make proto-install

# 生成 Protobuf 代码
make proto-gen

# 生成 Swagger 文档
make proto-swagger
```

##### Docker 支持

```bash
# 构建 Docker 镜像
make docker-build

# 运行 Docker 容器
make docker-run
```

##### 数据库迁移

```bash
# 创建新的迁移文件
make migrate-create name=migration_name

# 执行数据库迁移
make migrate-up

# 回滚数据库迁移
make migrate-down
```

##### 开发环境设置

```bash
# 安装所有开发依赖
make dev-deps
```

2. **使用 gRPCurl 工具**

   安装 gRPCurl 工具：

   ```bash
   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
   ```

   列出所有可用服务：

   ```bash
   grpcurl -plaintext localhost:8080 list
   ```

   输出示例：
   ```
   user.v1.UserService
   auth.v1.AuthService
   grpc.reflection.v1alpha.ServerReflection
   ```

   查看服务方法：

   ```bash
   grpcurl -plaintext localhost:8080 list user.v1.UserService
   ```

   输出示例：
   ```
   user.v1.UserService.CreateUser
   user.v1.UserService.GetUser
   user.v1.UserService.UpdateUser
   user.v1.UserService.DeleteUser
   user.v1.UserService.GetUserByEmail
   ```

   调用方法示例及响应：

   **用户注册示例**：
   ```bash
   grpcurl -plaintext -d '{
     "email": "test@example.com",
     "password": "Password123!",
     "full_name": "Test User"
   }' localhost:8080 user.v1.UserService/CreateUser
   ```

   成功响应：
   ```json
   {
     "user": {
       "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
       "email": "test@example.com",
       "fullName": "Test User",
       "createdAt": "2025-06-02T00:24:15Z",
       "updatedAt": "2025-06-02T00:24:15Z"
     }
   }
   ```

   **用户登录示例**：
   ```bash
   grpcurl -plaintext -d '{
     "email": "test@example.com",
     "password": "Password123!"
   }' localhost:8080 auth.v1.AuthService/Login
   ```

   成功响应：
   ```json
   {
     "tokens": {
       "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
       "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
     },
     "user": {
       "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
       "email": "test@example.com",
       "fullName": "Test User"
     }
   }
   ```

   **获取用户信息示例**：
   ```bash
   grpcurl -plaintext -d '{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d479"}' \
     -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' \
     localhost:8080 user.v1.UserService/GetUser
   ```

   成功响应：
   ```json
   {
     "user": {
       "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
       "email": "test@example.com",
       "fullName": "Test User",
       "createdAt": "2025-06-02T00:24:15Z",
       "updatedAt": "2025-06-02T00:24:15Z"
     }
   }
   ```

   **更新用户信息示例**：
   ```bash
   grpcurl -plaintext -d '{
     "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
     "fullName": "Updated User Name"
   }' \
     -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' \
     localhost:8080 user.v1.UserService/UpdateUser
   ```

   成功响应：
   ```json
   {
     "user": {
       "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
       "email": "test@example.com",
       "fullName": "Updated User Name",
       "createdAt": "2025-06-02T00:24:15Z",
       "updatedAt": "2025-06-02T00:25:30Z"
     }
   }
   ```

   **刷新令牌示例**：
   ```bash
   grpcurl -plaintext -d '{
     "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   }' localhost:8080 auth.v1.AuthService/RefreshToken
   ```

   成功响应：
   ```json
   {
     "tokens": {
       "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
       "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
     }
   }
   ```

3. **使用 Swagger UI**

   可以通过 Swagger UI 浏览和测试 API：

   ```bash
   # 安装 swagger-ui 工具
   npm install -g swagger-ui-cli
   
   # 启动 Swagger UI
   swagger-ui-cli serve ./api/swagger/user_service.swagger.json
   ```

   然后在浏览器中访问 http://localhost:8080 查看和测试 API。

4. **使用 Postman**

   可以导入 Swagger 文档到 Postman：
   - 打开 Postman，点击 "Import" > "Import File"  
   - 选择 `./api/swagger/user_service.swagger.json`
   - Postman 将自动创建所有 API 端点的请求集合

#### 构建和运行

使用 Makefile 构建和运行应用：

```bash
# 构建应用
make build

# 运行应用
make run

# 构建并运行
make build run
```

#### 添加新功能

添加新功能时，请遵循以下步骤：

1. 在 `/api/proto` 中定义 API 合约
2. 在 `/internal/domain` 中定义领域模型
3. 在 `/internal/repository` 中实现数据访问层
4. 在 `/internal/service` 中实现业务逻辑
5. 在 `/internal/transport` 中实现 HTTP 和 gRPC 处理器
6. 在 `/cmd/server/wire` 中更新依赖注入配置