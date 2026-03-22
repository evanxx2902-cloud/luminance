# =============================================================================
# Stage 1: Build Go binary
# =============================================================================
FROM golang:1.22-alpine AS go-builder

WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
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
# Stage 3: Runtime — CentOS 7
# =============================================================================
FROM centos:7

# ── PostgreSQL 15 via official PGDG yum repo ──────────────────────────────
# CentOS 7 default repos provide PG 9.2 — far too old for pgvector support.
# PGDG provides PG 15 for EL7.
RUN yum install -y \
      https://download.postgresql.org/pub/repos/yum/reporpms/EL-7-x86_64/pgdg-redhat-repo-latest.noarch.rpm \
      epel-release && \
    yum install -y \
      tini \
      monit \
      nginx \
      postgresql15-server \
      postgresql15-contrib \
      sudo \
      curl && \
    yum clean all

# Symlink PG 15 binaries to a PATH location used by scripts and monit
RUN ln -s /usr/pgsql-15/bin/initdb   /usr/local/bin/initdb   && \
    ln -s /usr/pgsql-15/bin/pg_ctl   /usr/local/bin/pg_ctl   && \
    ln -s /usr/pgsql-15/bin/psql     /usr/local/bin/psql     && \
    ln -s /usr/pgsql-15/bin/postgres /usr/local/bin/postgres

# ── Python 3.9 via Software Collections (SCL) ────────────────────────────
# CentOS 7 ships Python 3.6; sentence-transformers + fastapi require >= 3.8.
RUN yum install -y centos-release-scl && \
    yum install -y rh-python39 rh-python39-python-pip && \
    yum clean all

# Symlink python3.9 / pip3.9 so scripts can use python3/pip3 unqualified
RUN ln -sf /opt/rh/rh-python39/root/usr/bin/python3.9 /usr/local/bin/python3 && \
    ln -sf /opt/rh/rh-python39/root/usr/bin/pip3.9    /usr/local/bin/pip3

# ── pgvector — build from source against PG 15 headers ───────────────────
RUN yum install -y postgresql15-devel make gcc git && \
    git clone --branch v0.7.0 https://github.com/pgvector/pgvector.git /tmp/pgvector && \
    cd /tmp/pgvector && \
      PG_CONFIG=/usr/pgsql-15/bin/pg_config make && \
      PG_CONFIG=/usr/pgsql-15/bin/pg_config make install && \
    rm -rf /tmp/pgvector && \
    yum remove -y git make gcc postgresql15-devel && \
    yum clean all

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
ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
