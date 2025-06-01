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
*   **日志**: Zap 或 Logrus
*   **单元测试**: Go testing
*   **项目结构**: 参考 `gin-pathway` <mcurl name="gin-pathway" url="https://github.com/vespeng/gin-pathway"></mcurl>

## 开发计划 (初步)

**迭代 1: 项目骨架搭建与用户管理基础功能**

*   搭建项目基本目录结构 (参考 `gin-pathway`)
*   集成 Gin 框架
*   配置多环境支持
*   实现用户注册功能
*   编写用户注册的单元测试

**迭代 2: 系统鉴权功能**

*   集成 Redis
*   实现基于 Redis + JWT 的登录认证
*   实现 Refresh Token 机制
*   设计统一接口返回结构
*   编写认证功能的单元测试

**迭代 3: 用户管理进阶功能与优化**

*   实现用户信息查询、更新、删除功能
*   完善统一接口返回结构
*   集成依赖注入 (Wire)
*   编写用户管理进阶功能的单元测试
*   代码重构与优化

**迭代 4: 文档与部署**

*   编写 API 文档
*   编写部署文档
*   持续集成/持续部署 (CI/CD) 配置 (可选)

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

## 后续优化

* 集成其他微服务组件 (如服务发现、熔断等)
* 更完善的错误处理和日志记录
* 性能调优
* 添加单元测试和集成测试
* 自动化部署流程