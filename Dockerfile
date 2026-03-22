# =============================================================================
# Stage 1: Build Go binary
# =============================================================================
FROM golang:1.24-alpine AS go-builder

# Install protoc and dependencies for code generation
RUN apk add --no-cache protobuf protobuf-dev git curl

# Install Go protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

WORKDIR /build

# Copy proto files
COPY api/ ./api/

# Download Kratos third_party proto files
RUN mkdir -p third_party && \
    cd third_party && \
    go mod init temp && \
    go get github.com/go-kratos/kratos/v2@v2.9.2 && \
    cp -r $(go env GOPATH)/pkg/mod/github.com/go-kratos/kratos/v2@v2.9.2/third_party/* . || true

# Copy go.mod/go.sum and download dependencies
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /build/backend
RUN go mod download

# Create output directory for generated code
RUN mkdir -p api/luminance/v1

# Generate Go code from proto
WORKDIR /build
RUN protoc \
    --proto_path=./api/proto/v1 \
    --proto_path=./third_party \
    --proto_path=/usr/include \
    --go_out=paths=source_relative:./backend/api/luminance/v1 \
    --go-grpc_out=paths=source_relative:./backend/api/luminance/v1 \
    --grpc-gateway_out=paths=source_relative:./backend/api/luminance/v1 \
    --grpc-gateway_opt=generate_unbound_methods=true \
    ./api/proto/v1/*.proto

# Copy the rest of backend source code
COPY backend/ ./backend/

# Build binaries
WORKDIR /build/backend
RUN CGO_ENABLED=0 GOOS=linux go build -o luminance-api ./cmd/luminance-api
RUN CGO_ENABLED=0 GOOS=linux go build -o luminance-migrate ./cmd/migrate


# =============================================================================
# Stage 2: Build React frontend
# =============================================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ .
RUN npm run build
# Output: /build/frontend/dist/  (vite default outDir)


# =============================================================================
# Stage 3: Runtime — AlmaLinux 8 (RHEL-compatible, long-term support)
# =============================================================================
FROM almalinux:8

# Enable PowerTools (CRB) for additional packages
RUN dnf install -y 'dnf-command(config-manager)' && \
    dnf config-manager --set-enabled powertools && \
    dnf clean all

# ── PostgreSQL 15 via official PGDG yum repo ──────────────────────────────
# PGDG provides PG 15 for EL8.
# Note: Skip EPEL for now due to network issues in container
# Disable built-in PostgreSQL module to allow PGDG packages
RUN dnf install -y \
      https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm && \
    dnf -qy module disable postgresql && \
    dnf install -y \
      nginx \
      postgresql15-server \
      postgresql15-contrib \
      sudo \
      curl \
      redis && \
    dnf clean all

# ── monit (process supervisor) ────────────────────────────────────────────
# monit is in EPEL, install from GitHub release instead
RUN curl -fsSL -o /tmp/monit.tar.gz \
      https://mmonit.com/monit/dist/binary/5.33.0/monit-5.33.0-linux-x64.tar.gz && \
    tar -xzf /tmp/monit.tar.gz -C /tmp && \
    cp /tmp/monit-5.33.0/bin/monit /usr/local/bin/monit && \
    chmod +x /usr/local/bin/monit && \
    rm -rf /tmp/monit*

# ── tini (PID 1 init system) ─────────────────────────────────────────────
# tini is not in AlmaLinux base repos, install from GitHub release
RUN curl -fsSL -o /usr/local/bin/tini \
      https://github.com/krallin/tini/releases/download/v0.19.0/tini-amd64 && \
    chmod +x /usr/local/bin/tini

# Symlink PG 15 binaries to a PATH location used by scripts and monit
RUN ln -s /usr/pgsql-15/bin/initdb   /usr/local/bin/initdb   && \
    ln -s /usr/pgsql-15/bin/pg_ctl   /usr/local/bin/pg_ctl   && \
    ln -s /usr/pgsql-15/bin/psql     /usr/local/bin/psql     && \
    ln -s /usr/pgsql-15/bin/postgres /usr/local/bin/postgres

# ── Python 3.9 via AlmaLinux 8 AppStream ────────────────────────────────
# AlmaLinux 8 provides python39 package directly in AppStream.
RUN dnf install -y python39 python39-pip && \
    dnf clean all

# Symlink python3.9 / pip3.9 so scripts can use python3/pip3 unqualified
RUN ln -sf /usr/bin/python3.9 /usr/local/bin/python3 && \
    ln -sf /usr/bin/pip3.9    /usr/local/bin/pip3

# ── pgvector — build from source against PG 15 headers ───────────────────
RUN dnf install -y postgresql15-devel make gcc git && \
    git clone --branch v0.7.0 https://github.com/pgvector/pgvector.git /tmp/pgvector && \
    cd /tmp/pgvector && \
      PG_CONFIG=/usr/pgsql-15/bin/pg_config make && \
      PG_CONFIG=/usr/pgsql-15/bin/pg_config make install && \
    rm -rf /tmp/pgvector && \
    dnf remove -y git make gcc postgresql15-devel && \
    dnf clean all

# ── Runtime directory layout ──────────────────────────────────────────────
RUN mkdir -p \
      /opt/luminance/bin \
      /opt/luminance/web \
      /opt/luminance/ai \
      /opt/luminance/configs \
      /opt/luminance/scripts \
      /data/log \
      /data/pg-business \
      /data/pg-vector \
      /data/redis \
      /etc/monit/conf.d \
      /etc/nginx/conf.d

# ── Copy compiled artifacts from build stages ────────────────────────────
COPY --from=go-builder       /build/backend/luminance-api     /opt/luminance/bin/luminance-api
COPY --from=go-builder       /build/backend/luminance-migrate /opt/luminance/bin/luminance-migrate
COPY --from=frontend-builder /build/frontend/dist/             /opt/luminance/web/

# ── Python AI service ─────────────────────────────────────────────────────
COPY ai/requirements.txt /opt/luminance/ai/
RUN pip3 install --no-cache-dir -r /opt/luminance/ai/requirements.txt
COPY ai/ /opt/luminance/ai/

# ── Config and script files ───────────────────────────────────────────────
COPY configs/config.yaml /opt/luminance/configs/config.yaml
COPY nginx/nginx.conf             /etc/nginx/nginx.conf
COPY nginx/conf.d/                /etc/nginx/conf.d/
COPY monit/monitrc                /etc/monit/monitrc
COPY monit/conf.d/                /etc/monit/conf.d/
COPY scripts/                     /opt/luminance/scripts/
COPY backend/migrations/          /opt/luminance/backend/migrations/

RUN chmod +x /opt/luminance/scripts/*.sh && \
    chmod 600 /etc/monit/monitrc

# Convenience symlink so entrypoint is at a clean path
RUN ln -s /opt/luminance/scripts/entrypoint.sh /entrypoint.sh

# ── Persistent data volume ────────────────────────────────────────────────
VOLUME ["/data"]

EXPOSE 80 2812

# tini as PID 1 for clean signal handling and zombie reaping
ENTRYPOINT ["/usr/local/bin/tini", "--", "/entrypoint.sh"]
