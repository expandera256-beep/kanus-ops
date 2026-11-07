#!/bin/bash
# DRAFT — restore store (and optionally env) from /var/backups/us-license-hook.
# REQUIRES maintenance window: stop us-license-hook briefly. Not for unattended use.
# Usage:
#   restore-us-license-hook-store.sh licenses.json.YYYYMMDDThhmmssZ
#   restore-us-license-hook-store.sh --env us-license-hook.env.YYYYMMDDThhmmssZ
set -euo pipefail

DEST_DIR=/var/backups/us-license-hook
STORE=/opt/us-license-hook/licenses.json
ENVF=/etc/us-license-hook.env

if [ "${1:-}" = "--env" ]; then
  src="$DEST_DIR/${2:?need env backup basename}"
  [ -f "$src" ] || { echo "missing $src"; exit 1; }
  echo "Will restore ENV from $src -> $ENVF (mode 600 root:uslic or root:root per policy)"
  echo "STOP: run manually after systemctl stop us-license-hook"
  exit 2
fi

base=${1:?need licenses.json.STAMP basename}
src="$DEST_DIR/$base"
[ -f "$src" ] || { echo "missing $src"; exit 1; }

echo "Planned restore: $src -> $STORE"
echo "Steps (operator):"
echo "  1) systemctl stop us-license-hook"
echo "  2) cp -a $STORE ${STORE}.pre-restore.\$(date -u +%Y%m%dT%H%M%SZ)  # if exists"
echo "  3) install -m 640 -o uslic -g uslic $src $STORE"
echo "  4) systemctl start us-license-hook"
echo "  5) curl -s http://127.0.0.1:8787/healthz"
echo "This script does NOT perform the restore (safety)."
exit 0
