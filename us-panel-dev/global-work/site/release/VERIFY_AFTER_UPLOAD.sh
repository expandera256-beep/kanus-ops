#!/usr/bin/env bash
# Yükleme SONRASI salt-okuma doğrulaması (kanus.org). Hiçbir şey yazmaz.
# Kullanım: ./VERIFY_AFTER_UPLOAD.sh [https://kanus.org]
set -u; B=${1:-https://kanus.org}; fail=0
DIR=$(cd "$(dirname "$0")" && pwd)
pass(){ echo "PASS $*"; }; bad(){ echo "FAIL $*"; fail=1; }
# 1) Paketteki her dosya sunucudan birebir aynı geliyor mu (SHA256)?
while read -r sum path; do
  got=$(curl -sS -m 20 "$B/$path" | sha256sum | cut -d' ' -f1)
  [ "$got" = "$sum" ] && pass "/$path sha256" || bad "/$path sha256 farklı (SPA fallback ya da eski sürüm?)"
done < "$DIR/SHA256SUMS.txt"
# 2) Dizin URL'leri (DirectoryIndex) canonical ile sunuluyor mu?
for f in $(grep 'index.html$' "$DIR/FILES.txt"); do p="/${f%index.html}"
  curl -sS -m 20 "$B$p" | grep -qF "<link rel=\"canonical\" href=\"https://kanus.org$p\">" && pass "$p canonical" || bad "$p canonical"; done
curl -sS -m 20 "$B/robots.txt" | grep -qF "Sitemap: https://kanus.org/sitemap.xml" && pass robots || bad robots
curl -sS -m 20 "$B/sitemap.xml" | python3 -c 'import sys,xml.etree.ElementTree as E;E.fromstring(sys.stdin.read())' 2>/dev/null && pass "sitemap XML parse" || bad "sitemap XML parse"
# 3) Hukuki sayfalar paketle GELMEDİ: SPA fallback dönmeli, statik sayfa olmamalı; sitemap'te olmamalı.
for l in tr en; do for p in privacy terms refund; do
  body=$(curl -sS -m 20 "$B/$l/$p/")
  if grep -qF 'index-uscommerce.js' <<<"$body" && ! grep -qF 'rel="canonical" href="https://kanus.org/'"$l/$p/" <<<"$body"; then pass "/$l/$p/ paket dışı (SPA)"; else bad "/$l/$p/ beklenmeyen statik içerik"; fi
done; done
curl -sS -m 20 "$B/sitemap.xml" | grep -qE '/(privacy|terms|refund)/' && bad "sitemap hukuki yol içeriyor" || pass "sitemap hukuki yol yok"
# 4) SPA bozulmadı
curl -sS -m 20 "$B/" | grep -qF 'index-uscommerce.js' && pass "/ SPA" || bad "/ SPA"
curl -sS -m 20 "$B/olmayan-spa-rotasi" | grep -qF 'index-uscommerce.js' && pass "SPA fallback" || bad "SPA fallback"
exit $fail
