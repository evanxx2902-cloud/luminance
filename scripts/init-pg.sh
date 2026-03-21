#!/usr/bin/env bash
# Initialize both PostgreSQL 15 instances on first boot.
# Idempotent: skips initialization if data directory already exists.
set -euo pipefail

PG_USER="luminance"
PG_PASS="luminance"

init_instance() {
  local data_dir="$1"
  local port="$2"
  local dbname="$3"
  local extensions="${4:-}"   # optional space-separated extension list

  if [ ! -f "${data_dir}/PG_VERSION" ]; then
    echo "[init-pg] Initializing PostgreSQL 15 at ${data_dir} (port ${port})"
    initdb -D "${data_dir}" --username="${PG_USER}" --pwfile=<(echo "${PG_PASS}")

    # Set the correct port (default in postgresql.conf is #port = 5432)
    sed -i "s/^#port = 5432/port = ${port}/" "${data_dir}/postgresql.conf"

    # Start temporarily for DB/extension setup
    pg_ctl -D "${data_dir}" -l "/data/log/pg-init-${port}.log" start
    sleep 3

    # Create application database
    PGPASSWORD="${PG_PASS}" psql -U "${PG_USER}" -p "${port}" -c "CREATE DATABASE ${dbname};" || true

    # Install requested extensions into the new database
    if [ -n "${extensions}" ]; then
      for ext in ${extensions}; do
        PGPASSWORD="${PG_PASS}" psql -U "${PG_USER}" -p "${port}" -d "${dbname}" \
          -c "CREATE EXTENSION IF NOT EXISTS ${ext};" || true
      done
    fi

    pg_ctl -D "${data_dir}" stop
    echo "[init-pg] Instance at ${data_dir} ready."
  else
    echo "[init-pg] Instance at ${data_dir} already initialized — skipping."
  fi
}

mkdir -p /data/pg-business /data/pg-vector /data/log

# Business DB  — port 5432, no extensions
init_instance /data/pg-business 5432 luminance

# Vector DB — port 5433, pgvector extension
init_instance /data/pg-vector 5433 luminance_vector "vector"
