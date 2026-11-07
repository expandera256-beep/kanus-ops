#!/usr/bin/env bash
# Rollback api-test.kanus.org to test VPS 31.220.63.191
# DEFAULT: dry-run. Live: ROLLBACK_EXECUTE=1
set -euo pipefail
DRY=1
if [ "${ROLLBACK_EXECUTE:-0}" = "1" ]; then DRY=0; fi
TEST_HOST=31.220.63.191
PROD_HOST=82.29.184.225
TEST_KEY=${TEST_KEY:-/workspace/secrets/hostinger_srv1662550_ed25519}
PROD_KEY=${PROD_KEY:-/workspace/secrets/hostinger_api_prod_ed25519}
SSH_BASE=(-o IdentitiesOnly=yes -o BatchMode=yes -o ConnectTimeout=30 -o StrictHostKeyChecking=accept-new)
log() { printf '[%s] %s\n' "$(date -Iseconds)" "$*"; }
run() { if [ "$DRY" -eq 1 ]; then log "DRY: $*"; else log "RUN: $*"; eval "$@"; fi; }
ssh_retry() {
  local key="$1" host="$2"; shift 2
  local a=0
  while [ $a -lt 6 ]; do
    a=$((a+1))
    if ssh -i "$key" "${SSH_BASE[@]}" root@"$host" "$@"; then return 0; fi
    log "ssh retry $a → $host"; sleep 4
  done
  return 1
}
log "mode=$([ "$DRY" -eq 1 ] && echo DRY-RUN || echo EXECUTE)"
log "=== DNS api-test A → 31.220.63.191 ==="
run "echo 'MANUAL/API: point api-test A back to 31.220.63.191'"
log "=== start old hook on test VPS ==="
run "ssh_retry \"$TEST_KEY\" \"$TEST_HOST\" 'systemctl start us-license-hook; curl -sf http://127.0.0.1:8787/healthz'"
log "=== stop test instance on prod (do not delete) ==="
run "ssh_retry \"$PROD_KEY\" \"$PROD_HOST\" 'systemctl stop us-license-hook-test; systemctl disable us-license-hook-test'"
log "=== smoke old endpoint ==="
run "curl -sS -m15 -o /dev/null -w '%{http_code}' https://api-test.kanus.org/healthz"
log "done — if prod wrote new ledger rows during cutover, merge manually"
[ "$DRY" -eq 1 ] && log "Re-run with ROLLBACK_EXECUTE=1 to apply."
