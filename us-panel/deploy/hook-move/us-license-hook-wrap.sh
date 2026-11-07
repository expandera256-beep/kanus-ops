#!/bin/sh
set -eu
# systemd EnvironmentFile truncates JSON; load plans from file.
if [ -f /opt/us-license-hook-test/plans.json ]; then
  export US_PLANS_JSON="$(cat /opt/us-license-hook-test/plans.json)"
  export US_PLANS_FILE=/opt/us-license-hook-test/plans.json
fi
# 0-byte store crashes OpenStore; missing file is OK. Heal empty file.
STORE="${US_STORE_PATH:-/opt/us-license-hook-test/licenses.json}"
if [ -e "$STORE" ] && [ ! -s "$STORE" ]; then
  printf '%s\n' '{}' > "$STORE"
fi
exec /opt/us-license-hook-test/us-license-hook.real
