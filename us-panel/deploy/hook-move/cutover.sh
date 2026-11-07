#!/usr/bin/env bash
# HOOK MOVE cutover — api-test.kanus.org → srv2133566 (us-license-hook-test :8788)
# DEFAULT: dry-run only (prints steps, no destructive actions).
# Live cutover requires: CUTOVER_EXECUTE=1 and Demir DNS+stop approval.
set -euo pipefail

DRY=1
if [ "${CUTOVER_EXECUTE:-0}" = "1" ]; then DRY=0; fi

TEST_HOST=31.220.63.191
PROD_HOST=82.29.184.225
TEST_KEY=${TEST_KEY:-/workspace/secrets/hostinger_srv1662550_ed25519}
PROD_KEY=${PROD_KEY:-/workspace/secrets/hostinger_api_prod_ed25519}
SSH_BASE=(-o IdentitiesOnly=yes -o BatchMode=yes -o ConnectTimeout=30 -o StrictHostKeyChecking=accept-new)

log() { printf '[%s] %s\n' "$(date -Iseconds)" "$*"; }
run() {
  if [ "$DRY" -eq 1 ]; then
    log "DRY: $*"
  else
    log "RUN: $*"
    eval "$@"
  fi
}

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
log "prereq check: Demir must have approved DNS cutover + old stop; TTL should be 300 for ≥1h"

# 0) Preflight (always read-only)
log "=== preflight (read-only) ==="
ssh_retry "$PROD_KEY" "$PROD_HOST" 'systemctl is-active us-license-hook-test || true; test -x /opt/us-license-hook-test/us-license-hook.real; nginx -t'
ssh_retry "$TEST_KEY" "$TEST_HOST" 'systemctl is-active us-license-hook; sha256sum /opt/us-license-hook/licenses.json'

log "=== step1 stop old hook (write freeze) ==="
run "ssh_retry \"$TEST_KEY\" \"$TEST_HOST\" 'systemctl stop us-license-hook'"

log "=== step2 final rsync licenses.json (checksum) test→prod ==="
# box-mediated copy pattern
run "tmpdir=\$(mktemp -d /tmp/hook-cutover-XXXX); chmod 700 \"\$tmpdir\"; scp -i \"$TEST_KEY\" \"\${SSH_BASE[*]}\" root@$TEST_HOST:/opt/us-license-hook/licenses.json \"\$tmpdir/licenses.json\"; scp -i \"$PROD_KEY\" \"\${SSH_BASE[*]}\" \"\$tmpdir/licenses.json\" root@$PROD_HOST:/opt/us-license-hook-test/licenses.json; ssh -i \"$PROD_KEY\" \"\${SSH_BASE[*]}\" root@$PROD_HOST 'chown uslic-test:uslic-test /opt/us-license-hook-test/licenses.json; chmod 600 /opt/us-license-hook-test/licenses.json'; rm -rf \"\$tmpdir\""

log "=== step3 start test unit on prod ==="
run "ssh_retry \"$PROD_KEY\" \"$PROD_HOST\" 'systemctl enable --now us-license-hook-test; curl -sf http://127.0.0.1:8788/healthz'"

log "=== step4 DNS api-test A → 82.29.184.225 (TTL 300) — REQUIRES Hostinger API / manual ==="
run "echo 'MANUAL/API: update api-test A to 82.29.184.225; take zone backup first'"

log "=== step5 wait dig @1.1.1.1 and @8.8.8.8 ==="
run "echo 'wait until only 82.29.184.225'"

log "=== step6 TLS: prefer DNS-01 already issued; else webroot certbot on prod ==="
run "ssh_retry \"$PROD_KEY\" \"$PROD_HOST\" 'certbot certonly --webroot -w /var/www/html -d api-test.kanus.org --non-interactive || true; echo add HTTPS server block then nginx -t && reload'"

log "=== step7 smoke ==="
run "curl -sS -m10 -o /dev/null -w '%{http_code}' https://api-test.kanus.org/healthz"
run "curl -sS -m10 -o /dev/null -w '%{http_code}' -X POST https://api-test.kanus.org/webhooks/shopify/orders-paid -H 'Content-Type: application/json' -d '{}'"

log "=== done ==="
log "Old service on test VPS stays stopped but intact for 24h. No file deletes."
[ "$DRY" -eq 1 ] && log "Re-run with CUTOVER_EXECUTE=1 only after Demir approval."
