# §12/§13 güvenlik audit harness

Prompt: `prompt/KANUS_PROMPT_20261107.txt` tamamı; §12deki14 madde.

```bash
# Yerel14 madde (3 açık negative regression nedeniyle exit1 BEKLENİR):
python3 /workspace/us-panel/qa/audit12/run.py
# Tek finding/konu:
python3 /workspace/us-panel/qa/audit12/run.py --case file_isolation
python3 /workspace/us-panel/qa/audit12/run.py --case ssrf
python3 /workspace/us-panel/qa/audit12/run.py --case backup
python3 /workspace/us-panel/qa/audit12/run.py --case rate_limit
# Synthetic SQLite + Go JSON store backup/restore:
python3 /workspace/us-panel/qa/audit12/run.py --case restore
# Metadata/path-only hardcoded-secret scan (private/key/data files NOT read):
python3 /workspace/us-panel/qa/audit12/scan_secrets.py
# Own raw logs only, path-only leak hits:
python3 /workspace/us-panel/qa/audit12/check_logs.py
# PROD VPS: requires new-host DNS/TLS/health precheck first; up to1 health+8 sequential negatives,120s timeout:
python3 /workspace/us-panel/qa/audit12/run.py --remote
```

Go test dosyaları yalnız `-overlay` ile qa/dan enjekte edilir; ürün dosyaları değişmez. Command/path/env/SSRF/isolation reproları yalnız loopback ve yeni sentetik tmp dosyaları kullanır. SSRF testi metadata servisine veya gerçek internal servise gitmez; iki yerel httptest server kullanır. Isolation testi başka bir gerçek kullanıcı verisine değil aynı testin oluşturduğu iki synthetic workspacee gider. Gerçek webhook/session sırları veya QA hesabının gerçek credential dosyaları okunmaz; yerel dummy sırlar test substrateıdır. SMTP gerçek senderı kullanılmaz; privacy testi yalnız DryRunSender/fakeSenderla fixture allowlisti kullanır. Gerçek e-posta yok.

Backup shell repro sabit /run,/etc,/opt pathlerini ONLY test kopyasında temp pathlere remap eder; kalan shell mantığı aynı. Store dahil üç kaynak da eksikken exit0 dönmesi beklenen-açık FAILdir. Gerçek backup dosyaları açılmaz/silinmez. SQLite backup yeni DB dosyalarıyla; Go snapshot gerçek Store API/atomic JSON state ile, offline ayrı instance reopen/kota/history/lock kontrolüyle yapılır. PostgreSQL/production restore kapsam dışı ve BLOCKEDdir.

Ana rapor: `reports/r20261108/QA_SECURITY_AUDIT_20261108.md`. Findingler: `reports/r20261108/qa-findings/FINDING-*.md`. Ham kanıtlar: `reports/r20261108/qa-raw/`. Product fix QA tarafından uygulanmadı: §13te root cause+reproducer+regression tamamlandı, fix/retest ownera açıkça devredilmek üzere bırakıldı. Hiçbir failed regression xfail/PASSa çevrilmedi.


## 00:38 TSİ PROD-host steering
api-test yeni PROD host82.29.184.225tir (HOST aktarımı); sadece test VM sanmayın. Box resolver198.18.0.1 ile serving-IP bağı doğrulanamadı: precheck BLOCKED,0 HTTP; authoritative DNS/TLS peer+health kanıtı HOSTtan beklenir. `precheck.py` fail-closed; başarısız gate başka route ile aşılmaz. Mevcut00:30 negatif logları gate eklenmeden önceki dar kontrol olarak tarihsel korunur, yeni host eşliği iddia edilmez.

**Hiç authenticated chat/completions/model generation isteği gönderilmez**; GPU live checks KAN penceresine kadar NOT TESTED. Sadece invalid/forged/missing/revoked negatif token testleri ve yerel mock pozitifi uygundur. Production ledger/journal root SSH kanıtları BLOCKED; customer data/secret dumps istemeyin. Tam Shopify E2E SHOPa aittir.
