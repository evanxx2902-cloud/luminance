# Luminance 项目规则

## 规则 1：语言
所有对话、回复、注释和说明均使用**中文**。

## 规则 2：多 Agent 协作规范

### API 契约（单一数据源）
- **API 规范中心**: `api/` 目录是前后端协作的单一数据源
  - `api/proto/v1/`: Protobuf 定义文件（`.proto`），定义 gRPC 接口和消息结构
  - `api/openapi/v1/`: 生成的 OpenAPI/Swagger 规范

### 各 Agent 职责边界

#### 后端 Agent
- **只能修改**: `api/proto/v1/*.proto` 文件
- **必须执行**: 修改 `.proto` 后，运行 `scripts/gen-api.sh` 重新生成代码
- **必须提交**: 同时提交 `.proto` 源文件和生成的 Go 代码
- **禁止**: 直接修改生成的 `*.pb.go` 文件

#### 前端 Agent
- **只能读取**: `api/proto/v1/` 和 `api/openapi/v1/` 下的文件
- **基于生成代码开发**: 使用 `scripts/gen-api.sh` 生成的 TypeScript Client（或通过 OpenAPI 生成）
- **禁止**: 手写 API URL 或请求结构，所有调用必须基于生成的 Client

#### AI Agent
- **基于 proto 理解结构**: 通过读取 `api/proto/v1/` 下的文件理解数据模型
- **与后端对齐**: 向量存储和检索的数据结构必须与 proto 定义的 message 一致

### 代码生成流程
```bash
# 修改 api/proto/v1/*.proto 后，必须运行：
./scripts/gen-api.sh

# 验证生成成功
cd backend && go build ./...

# 提交时包含：
git add api/ backend/api/
git commit -m "api: update DocumentService proto and regenerate"
```

### 数据库迁移流程
```bash
# 创建新迁移
migrate create -ext sql -dir backend/migrations -seq init_schema

# 本地测试迁移
cd backend
go run ./cmd/migrate -command=up -db=business
go run ./cmd/migrate -command=up -db=vector

# 容器启动时会自动运行迁移（见 scripts/start-services.sh）
```
