#!/usr/bin/env bash
# Run AFTER KAN.US uploads public_html/* to Hostinger public_html/. Read-only checks (curl GET).
# Usage: VERIFY_AFTER_UPLOAD.sh [https://kanus.org]   exit 0 = all PASS
set -u
B="${1:-https://kanus.org}"; ASSET="index-uscommerce-i18n-75f56119.js"; WANT="75f561190fafb99e1e58ddf567539d8c6f785bda2254221face3b8a10d8b9911"; F=0
ok(){ echo "PASS $1"; }; bad(){ echo "FAIL $1"; F=$((F+1)); }
for p in / /us /contact /news; do
  h=$(curl -fsS "$B$p") || { bad "$p fetch"; continue; }
  echo "$h" | grep -q "/assets/$ASSET" && ok "$p references $ASSET" || bad "$p does not reference $ASSET (cache? old index.html?)"
done
got=$(curl -fsS "$B/assets/$ASSET" | sha256sum | cut -d' ' -f1)
[ "$got" = "$WANT" ] && ok "asset sha256" || bad "asset sha256 $got != $WANT"
ct=$(curl -fsSI "$B/assets/$ASSET" | tr -d '\r' | awk -F': ' 'tolower($1)=="content-type"{print $2}')
echo "$ct" | grep -qi javascript && ok "asset content-type $ct" || bad "asset content-type '$ct' (SPA fallback?)"
curl -fsS "$B/assets/index-uscommerce.js?v=20261104-usticari" >/dev/null && ok "old asset still served (rollback safe)" || bad "old asset missing"
curl -fsS "$B/assets/$ASSET" | grep -q '__kDate' && ok "i18n helpers present" || bad "helpers missing"
for s in 'Yerli Teknoloji' 'Küresel Vizyon' 'Built in Türkiye' 'Designed for the World'; do
  curl -fsS "$B/assets/$ASSET" | grep -qi "$s" && bad "forbidden slogan: $s" || ok "no '$s'"
done
echo "Manual: open $B/us, switch TR/EN/DE/RU; /news dates localized; failed form submit shows a localized notice."
echo "Rollback: upload rollback/index.html over public_html/index.html (old asset untouched)."
[ $F -eq 0 ] && echo "ALL PASS" || echo "$F FAIL"; exit $F
