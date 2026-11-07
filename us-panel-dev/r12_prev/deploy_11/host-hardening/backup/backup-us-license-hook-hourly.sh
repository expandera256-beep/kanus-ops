#!/bin/bash
# DRAFT — install as /usr/local/sbin/backup-us-license-hook-hourly.sh (mode 755).
# Hourly consistent copy of store + env using flock. 7-day retention.
# Env backups: root:root 600. Do NOT copy env contents off-box.
set -euo pipefail

DEST_DIR=${DEST_DIR:-/var/backups/us-license-hook}
KEEP_DAYS=${KEEP_DAYS:-7}
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
LOCK=/run/us-license-hook-backup.lock
STORE=${US_STORE_PATH:-/opt/us-license-hook/licenses.json}
ENVF=/etc/us-license-hook.env
PLANS=/opt/us-license-hook/plans.json

mkdir -p "$DEST_DIR"
chmod 750 "$DEST_DIR"

exec 9>"$LOCK"
if ! flock -w 30 9; then
  echo "backup: could not acquire flock on $LOCK" >&2
  exit 1
fi

backed=0
copy_one() {
  local src="$1" mode="$2" owner="$3"
  local base dest
  base=$(basename "$src")
  dest="$DEST_DIR/${base}.${STAMP}"
  if [ ! -f "$src" ]; then
    echo "skip missing $src"
    return 0
  fi
  # consistent copy under flock (blocks concurrent backup; store writes use app flock separately)
  install -m "$mode" -o "${owner%:*}" -g "${owner#*:}" "$src" "$dest"
  echo "backed $src -> $dest"
  backed=1
}

copy_one "$STORE" 640 root:root
copy_one "$PLANS" 640 root:root
# secrets: root-only
copy_one "$ENVF" 600 root:root

find "$DEST_DIR" -type f \( \
  -name 'licenses.json.*' -o -name 'plans.json.*' -o -name 'us-license-hook.env.*' \
\) -mtime +"$KEEP_DAYS" -delete 2>/dev/null || true

[ "$backed" = 1 ] || echo "nothing backed up"
