#!/usr/bin/env bash
# 生成 API 代码脚本
# 根据 api/proto/v1/*.proto 生成 Go (Kratos) 和 TypeScript (OpenAPI) 代码

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_DIR="${PROJECT_ROOT}/api/proto/v1"
OPENAPI_DIR="${PROJECT_ROOT}/api/openapi/v1"
BACKEND_API_DIR="${PROJECT_ROOT}/backend/api/luminance/v1"

# Kratos third_party proto 路径（包含 google/api/annotations.proto）
KRATOS_THIRD_PARTY="$(go env GOPATH)/pkg/mod/github.com/go-kratos/kratos/v2@v2.9.2/third_party"

echo "=== Generating API code ==="
echo "Proto dir: ${PROTO_DIR}"

# 确保输出目录存在
mkdir -p "${OPENAPI_DIR}"
mkdir -p "${BACKEND_API_DIR}"

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install it first."
    echo "  macOS: brew install protobuf"
    echo "  Linux: apt-get install -y protobuf-compiler"
    exit 1
fi

# 检查 Go 插件是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# 生成 Go 代码
echo ""
echo "=== Generating Go code (Kratos) ==="
cd "${PROJECT_ROOT}/backend"

protoc \
    --proto_path="${PROTO_DIR}" \
    --proto_path="${KRATOS_THIRD_PARTY}" \
    --go_out=paths=source_relative:"${BACKEND_API_DIR}" \
    --go-grpc_out=paths=source_relative:"${BACKEND_API_DIR}" \
    "${PROTO_DIR}"/*.proto

echo "Go code generated at: ${BACKEND_API_DIR}"

# 生成 gRPC-Gateway 代码
if command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo ""
    echo "=== Generating gRPC-Gateway code ==="
    protoc \
        --proto_path="${PROTO_DIR}" \
        --proto_path="${KRATOS_THIRD_PARTY}" \
        --grpc-gateway_out=paths=source_relative:"${BACKEND_API_DIR}" \
        --grpc-gateway_opt=generate_unbound_methods=true \
        "${PROTO_DIR}"/*.proto
    echo "Gateway code generated at: ${BACKEND_API_DIR}"
else
    echo ""
    echo "Skipping gRPC-Gateway generation (protoc-gen-grpc-gateway not installed)"
    echo "To install: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"
fi

# 生成 OpenAPI (使用 grpc-gateway 的 protoc-gen-openapiv2)
if command -v protoc-gen-openapiv2 &> /dev/null; then
    echo ""
    echo "=== Generating OpenAPI spec ==="
    protoc \
        --proto_path="${PROTO_DIR}" \
        --proto_path="${KRATOS_THIRD_PARTY}" \
        --openapiv2_out="${OPENAPI_DIR}" \
        --openapiv2_opt=logtostderr=true \
        "${PROTO_DIR}"/*.proto
    echo "OpenAPI spec generated at: ${OPENAPI_DIR}"
else
    echo ""
    echo "Skipping OpenAPI generation (protoc-gen-openapiv2 not installed)"
    echo "To install: go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-openapiv2@latest"
fi

echo ""
echo "=== API generation complete ==="
echo ""
echo "Next steps:"
echo "  1. Review generated code"
echo "  2. Run 'go build ./...' to verify"
echo "  3. Commit both proto and generated files"
