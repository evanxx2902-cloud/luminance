#!/usr/bin/env bash
# entrypoint.sh — PID 1 child handed to tini.
# tini → entrypoint.sh → start-services.sh → monit
set -euo pipefail
exec bash /opt/luminance/scripts/start-services.sh
