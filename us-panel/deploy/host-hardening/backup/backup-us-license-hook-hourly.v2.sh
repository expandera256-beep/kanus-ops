#!/bin/bash
# DRAFT v2 (2026-11-08) — FINDING-03 fix. NOT INSTALLED. Needs Demir approval to install.
# Usage: backup-us-license-hook-hourly.sh [prod|test]   (default prod)
# Fails (exit 2) if the required store is missing/non-regular or nothing was backed up.
# Verifies JSON integrity + sha256 manifest per run. 7-day retention.
set -euo pipefail
INST=${1:-prod}
case "$INST" in
  prod) BASE=/opt/us-license-hook; ENVF=/etc/us-license-hook.env; NAME=us-license-hook ;;
  test) BASE=/opt/us-license-hook-test; ENVF=/etc/us-license-hook-test.env; NAME=us-license-hook-test ;;
  *) echo "backup: unknown instance $INST" >&2; exit 64 ;;
esac
DEST_DIR=${DEST_DIR:-/var/backups/$NAME}
KEEP_DAYS=${KEEP_DAYS:-7}
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
LOCK=/run/$NAME-backup.lock
STORE=$BASE/licenses.json
PLANS=$BASE/plans.json
umask 077
mkdir -p "$DEST_DIR"; chmod 750 "$DEST_DIR"
exec 9>"$LOCK"
flock -w 30 9 || { echo "backup[$INST]: FAIL could not acquire $LOCK" >&2; exit 1; }
if [ ! -f "$STORE" ] || [ -L "$STORE" ]; then echo "backup[$INST]: FAIL required store missing/non-regular: $STORE" >&2; exit 2; fi
backed=0; MAN="$DEST_DIR/manifest.$STAMP.sha256"
copy_one() { # src mode required(0/1)
  local src="$1" mode="$2" req="$3" dest
  dest="$DEST_DIR/$(basename "$src").$STAMP"
  if [ ! -f "$src" ]; then
    if [ "$req" = 1 ]; then echo "backup[$INST]: FAIL missing required $src" >&2; exit 2; fi
    echo "backup[$INST]: WARN optional missing $src" >&2; return 0
  fi
  install -m "$mode" -o root -g root "$src" "$dest"
  (cd "$DEST_DIR" && sha256sum "$(basename "$dest")") >> "$MAN"
  echo "backed $src -> $dest"; backed=$((backed+1))
}
copy_one "$STORE" 640 1
python3 -c 'import json,sys; json.load(open(sys.argv[1]))' "$DEST_DIR/licenses.json.$STAMP" || { echo "backup[$INST]: FAIL store copy is not valid JSON" >&2; exit 3; }
copy_one "$PLANS" 640 1
copy_one "$ENVF" 600 0
(cd "$DEST_DIR" && sha256sum -c --quiet "$(basename "$MAN")") || { echo "backup[$INST]: FAIL manifest verify" >&2; exit 3; }
find "$DEST_DIR" -maxdepth 1 -type f \( -name 'licenses.json.*' -o -name 'plans.json.*' -o -name '*.env.*' -o -name 'manifest.*.sha256' \) -mtime +"$KEEP_DAYS" -delete
[ "$backed" -ge 2 ] || { echo "backup[$INST]: FAIL only $backed files backed up" >&2; exit 2; }
date -u +%FT%TZ > "$DEST_DIR/LAST_SUCCESS"
echo "backup[$INST]: OK $backed files"
