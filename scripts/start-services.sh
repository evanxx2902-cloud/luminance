#!/usr/bin/env bash
# Start all services via monit.  Called by entrypoint.sh (child of tini).
set -euo pipefail

echo "[start] Initializing persistent data directories..."
bash /opt/luminance/scripts/init-pg.sh

# 启动临时 PostgreSQL 实例用于迁移（如果 monit 还没启动 PG）
echo "[start] Starting temporary PostgreSQL for migrations..."
/usr/local/bin/pg_ctl -D /data/pg-business -l /data/log/pg-business-migrate.log start -w -t 30 || true
/usr/local/bin/pg_ctl -D /data/pg-vector -l /data/log/pg-vector-migrate.log start -w -t 30 || true

sleep 2

# 运行数据库迁移
echo "[start] Running database migrations..."
cd /opt/luminance

# 业务库迁移
/opt/luminance/bin/luminance-migrate -command=up -db=business || true

# 向量库迁移
/opt/luminance/bin/luminance-migrate -command=up -db=vector || true

# 停止临时 PG 实例（让 monit 接管）
echo "[start] Stopping temporary PostgreSQL instances..."
/usr/local/bin/pg_ctl -D /data/pg-business stop || true
/usr/local/bin/pg_ctl -D /data/pg-vector stop || true

echo "[start] Handing off to monit..."
exec /usr/bin/monit -Ic /etc/monit/monitrc
