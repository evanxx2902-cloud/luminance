#!/usr/bin/env bash
# Start all services via monit.  Called by entrypoint.sh (child of tini).
set -euo pipefail

echo "[start] Initializing persistent data directories..."
bash /opt/luminance/scripts/init-pg.sh

echo "[start] Handing off to monit..."
exec /usr/bin/monit -Ic /etc/monit/monitrc
