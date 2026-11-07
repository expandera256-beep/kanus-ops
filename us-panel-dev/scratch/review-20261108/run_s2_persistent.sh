#!/usr/bin/env bash
# S2 saldırı regresyon koşucusu (Linux). Kaynak ağacını DEĞİŞTİRMEZ.
# Kullanım: run_s2_regression.sh [term_ağacı] [sonuç_klasörü]
#   term_ağacı varsayılan: /workspace/us-panel-dev/work2/kiyas/term
# Çıkış kodu: 0 = tüm senaryolar PASS, 1 = en az bir FAIL/BLOCKED, 2 = kurulum hatası.
set -uo pipefail
HERE="/workspace/us-panel-dev/panel-evidence/review/regression"
SRC="$(cd "${1:-/workspace/us-panel-dev/work2/kiyas/term}" && pwd)" || { echo "ağaç yok"; exit 2; }
STAMP="$(date +%Y%m%d_%H%M%S)"
OUT="${2:-$HERE/results/$STAMP}"
mkdir -p "$OUT"
WORK="$(mktemp -d /workspace/us-panel-dev/scratch/review-20261108/s2reg.XXXXXX)"
DST="$WORK/term"
export GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0

# Ana ağacın değişmediğini kanıtlamak için öncesi/sonrası manifest.
manifest() { (cd "$SRC" && find . -type f ! -name '*.exe' -print0 | sort -z | xargs -0 sha256sum) ; }
manifest > "$OUT/src_manifest_before.txt"

cp -a "$SRC" "$DST"
cp "$HERE"/poc/*_test.go "$DST"/
echo "kaynak=$SRC kopya=$DST tarih=$(date '+%F %T %Z') go=$(go version)" > "$OUT/meta.txt"
[ -f "$SRC/release.go" ] && grep -h 'buildVersion' "$SRC/release.go" >> "$OUT/meta.txt"

# Senaryo kimlikleri ve hangi PoC dosyasında oldukları.
declare -A FILE_OF=(
 [H1-RUN]=http [H1-TEST]=http [H2-TOOLEXEC]=http [H3-RUN]=http [H3-SAFECMD]=internal
 [M1-SETSID-TIMEOUT]=http [M1-SETSID-CANCEL]=http [M2-TASKSTATE]=internal [M3-HARDLINK]=internal [M4-SYMLINK-RACE]=internal
 [L1-ADS]=internal [L1-ADS-DATA]=internal [L1-DRIVE-REL]=internal [L1-RESERVED-CON]=internal [L1-RESERVED-NUL]=internal
 [L1-RESERVED-AUX]=internal [L1-LONGPATH]=internal [L1-UNC]=internal [L1-DEVICE]=internal [L2-TESTS-RLIMIT]=http
 [N-EXEC]=http [N-TOOLEXEC-SPLIT]=http [N-OVERLAY]=http [N-LDFLAGS]=http [N-GCFLAGS]=http [N-MODFILE]=http
 [N-PY-C]=http [N-PY-C-AFTER]=http [N-PY-M-OTHER]=http [N-NODE-E]=http [N-NODE-EVAL]=http [N-NODE-P]=http
 [N-NODE-IMPORT]=http [N-NPM]=http [N-GOFLAGS-ENV]=http [N-CC-ENV]=http
)
ORDER="H1-RUN H1-TEST H2-TOOLEXEC H3-RUN H3-SAFECMD M1-SETSID-TIMEOUT M1-SETSID-CANCEL M2-TASKSTATE M3-HARDLINK M4-SYMLINK-RACE L1-ADS L1-ADS-DATA L1-DRIVE-REL L1-RESERVED-CON L1-RESERVED-NUL L1-RESERVED-AUX L1-LONGPATH L1-UNC L1-DEVICE L2-TESTS-RLIMIT N-EXEC N-TOOLEXEC-SPLIT N-OVERLAY N-LDFLAGS N-GCFLAGS N-MODFILE N-GOFLAGS-ENV N-CC-ENV N-PY-C N-PY-C-AFTER N-PY-M-OTHER N-NODE-E N-NODE-EVAL N-NODE-P N-NODE-IMPORT N-NPM"

# Derleme uyumu: derlenmeyen PoC dosyasını çıkar (API .9'da değişmiş olabilir).
REMOVED=""
for attempt in 1 2 3; do
  if (cd "$DST" && go vet . > "$OUT/compile_$attempt.txt" 2>&1); then break; fi
  bad=$(grep -oE 's2reg_[a-z]+_test\.go' "$OUT/compile_$attempt.txt" | sort -u)
  if [ -z "$bad" ]; then echo "PoC dışı derleme hatası, bkz compile_$attempt.txt"; exit 2; fi
  for f in $bad; do
    if [ "$f" = "s2reg_helpers_test.go" ]; then echo "helpers derlenmedi: adapter'ı güncelleyin (compile_$attempt.txt)"; exit 2; fi
    rm -f "$DST/$f"; REMOVED="$REMOVED ${f#s2reg_}"
  done
done

(cd "$DST" && go test -race -count=1 -v -timeout 20m -run '^TestS2Reg_' . > "$OUT/raw_test.log" 2>&1)
echo "go_test_exit=$?" >> "$OUT/meta.txt"

# Ana ağaç değişmedi mi?
manifest > "$OUT/src_manifest_after.txt"
if cmp -s "$OUT/src_manifest_before.txt" "$OUT/src_manifest_after.txt"; then SRCOK="DEĞİŞMEDİ"; else SRCOK="DEĞİŞTİ!"; fi

fails=0
{
  echo "# S2 regresyon sonucu — $STAMP"
  echo
  echo "- Kaynak: \`$SRC\` ($(grep -o '"[0-9.]*"' "$OUT/meta.txt" | head -1))"
  echo "- Kaynak ağacı koşu sonrası: **$SRCOK** (sha256 manifest karşılaştırması)"
  [ -n "$REMOVED" ] && echo "- Derlenmediği için çıkarılan PoC dosyaları:$REMOVED (senaryoları BLOCKED)"
  echo
  echo "| Senaryo | Sonuç | Ayrıntı |"
  echo "|---|---|---|"
  for id in $ORDER; do
    line=$(grep -h "S2REG|$id|" "$OUT/raw_test.log" | tail -1 | sed 's/.*S2REG|//')
    if [ -z "$line" ]; then
      res=BLOCKED; det="sonuç satırı yok (dosya: s2reg_${FILE_OF[$id]}_test.go${REMOVED:+; çıkarılanlar:$REMOVED})"
    else
      res=$(echo "$line" | cut -d'|' -f2); det=$(echo "$line" | cut -d'|' -f3- | sed 's/|/\\|/g')
    fi
    [ "$res" != PASS ] && fails=$((fails+1))
    echo "| $id | $res | $det |"
  done
  echo
} > "$OUT/table.md"
p=$(grep -c '| PASS |' "$OUT/table.md"); f=$(grep -c '| FAIL |' "$OUT/table.md"); b=$(grep -c '| BLOCKED |' "$OUT/table.md")
echo "Toplam: PASS=$p FAIL=$f BLOCKED=$b" | tee -a "$OUT/table.md"
cat "$OUT/table.md"
# Geçici kopyayı temizle; olası yetim sleep 47/48 süreçlerini öldür.
pkill -f '^sleep 4[78]$' 2>/dev/null || true
echo "Kalıcı kopya: $WORK"
echo "Sonuçlar: $OUT"
[ "$fails" -eq 0 ]
