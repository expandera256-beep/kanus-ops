# FINDING-05 — Canlı api-test store saatlik backup kapsamı dışında; restore drill kaydı yok

- Önem: **HIGH**. Sorumlu bot: **HOST** (kullanıcı tarafından açıkça atandı); restore kabulünde SRE.
- Güncel durum: **MITIGATED-PENDING-VERIFY**. Canlı test store kapsamı HOSTa göre 08.11.2026 **00:48 TSİ** itibarıyla düzeltildi. Üretim restore drill **BLOCKED**; bulgu tamamen kapatılmadı.

## Kanıt / kök neden
`reports/r20261108/HOST_EVIDENCE_QA_20261108.md:16–20`: eski saatlik :05 işi yalnız durdurulmuş `/opt/us-license-hook` üretim storeunu yedekliyordu; canlı `/opt/us-license-hook-test/licenses.json` kapsam dışıydı. Elle cutover kopyaları düzenli backup değildir. Bu prod durumu QA SSH ölçümü değil, HOSTun salt-okuma kanıtıdır.
Kaynak eski script `deploy/host-hardening/backup/backup-us-license-hook-hourly.sh:9–16,40–43`; kaynak cron `deploy/host-hardening/backup/us-license-hook-hourly.cron:5` halen yalnız prod işini içerir. Kaynak cron ile HOSTun kurulduğunu bildirdiği cron farklı; owner kaynak/runbooku senkronlamalıdır.

Beklenen: Her canlı instance için ayrı, zamanlanmış, bütünlüğü doğrulanan snapshot ve ölçülmüş restore/RPO/RTO kaydı.
Önceki gerçek: Yalnız durmuş prod storeuna saatlik backup; canlı api-test için düzenli iş yok, planlı drill kaydı yok.

## Yerel otomatik repro / fix retesti
```bash
python3 /workspace/us-panel/qa/audit12/test_backup_v2.py -v
# Aynı güncel v2 testlerini audit üzerinden:
python3 /workspace/us-panel/qa/audit12/run.py --case backup
```
`qa/audit12/test_backup_v2.py` gerçek v2 shell mantığını yalnız sentetik temp store/plans/env ile çalıştırır; sabit yollar ve root ownership test kopyasında yerelleştirilir. Prod/test fixtureları ayrı içerik ve hedeflerle doğrulanır, manifest checksumları ve LAST_SUCCESS kontrol edilir. **7 test PASS**; normal iki instance exit0, eksik store/plans veya symlink store2, bozuk JSON3, bilinmeyen arg64, opsiyonel env yokken0. `qa-raw/backup-v2-local.log`, `backup.log`. Gerçek backup/env/ledger açılmadı, VPSye gidilmedi.

v2 `backup-us-license-hook-hourly.v2.sh:8–10` prod/test yollarını ayırır; `:23,28–29,37–43` required input, JSON, manifest ve başarı işaretini kontrol eder. Script test storeunu **test argümanıyla kapsar**; varsayılan prod invocation tek başına kapsamaz. Schedulerın iki ayrı işi kurması şarttır.

## HOSTun düzeltme kanıtı / kalan kabul
HOST kanıtı`:20` ve 00:48 TSİ koordinatör aktarımı: v2 kurulu, prod :05 → `/var/backups/us-license-hook`; test :07 → `/var/backups/us-license-hook-test`; test log `/var/log/us-license-hook-test-backup-hourly.log`. İlk iki koşum exit0/3 dosya, manifest+LAST_SUCCESS; test snapshot SHA canlı store SHA `a0daf1ea…` ile aynı deniyor. QA bunu bağımsız ölçmedi.
Box v2 SHA256 **d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0**, HOSTun `d77534e5…a0` fragmanıyla tutarlı. Tam remote SHA verilmediği için tam eşitlik bağımsız doğrulanmış değildir.

Gereken: HOSTtan sadece tam kurulu script SHA, kurulu iki cron satırı, scheduled-run zamanı/exit/manifest doğrulama/LAST_SUCCESS metadata (değer/secret veya müşteri backupı yok); kaynak cron güncellemesi; onaylı bakım penceresinde ownerın restore drill/RPO/RTO kanıtı. QA VPS write/restart/restore yapmaz. Yerel restore PASS, production drill yerine geçmez.
