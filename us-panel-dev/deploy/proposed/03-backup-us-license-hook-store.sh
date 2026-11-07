#!/bin/bash
# UYGULANMADI — taslak. Sunucuya kopyalanıp çalıştırılmadı.
# Hedef: /usr/local/sbin/backup-us-license-hook-store.sh
# Cron:  /etc/cron.d/us-license-hook-store  →  15 3 * * * root /usr/local/sbin/backup-us-license-hook-store.sh
# Servis restart GEREKMEZ. Geri alma: cron dosyasını ve scripti sil; yedekler /var/backups altında kalır.

set -euo pipefail
SRC="${US_STORE_PATH:-/opt/us-license-hook/licenses.json}"
DEST_DIR=/var/backups/us-license-hook
KEEP_DAYS=14
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
mkdir -p "$DEST_DIR"
chmod 750 "$DEST_DIR"
if [ ! -f "$SRC" ]; then
  echo "store missing: $SRC" >&2
  exit 1
fi
install -m 640 -o root -g root "$SRC" "$DEST_DIR/licenses.json.${STAMP}"
# prune
find "$DEST_DIR" -type f -name 'licenses.json.*' -mtime +"$KEEP_DAYS" -delete
echo "backed up -> $DEST_DIR/licenses.json.${STAMP}"
