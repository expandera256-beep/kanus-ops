# §12/§13 güvenlik audit harness

Prompt: `prompt/KANUS_PROMPT_20261107.txt` tamamı; §12deki14 madde.

```bash
# Yerel14 madde (F01 açık architectural negative regression nedeniyle exit1 BEKLENİR):
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

Backup shell repro sabit /run,/etc,/opt pathlerini ONLY test kopyasında temp pathlere remap eder; kalan shell mantığı aynı. Eski v1 reproda üç kaynak eksikken exit0 tarihsel FAILdir; güncel runner v2yi test eder, required-store/plans yokken nonzero bekler. Gerçek backup dosyaları açılmaz/silinmez. SQLite backup yeni DB dosyalarıyla; Go snapshot gerçek Store API/atomic JSON state ile, offline ayrı instance reopen/kota/history/lock kontrolüyle yapılır. PostgreSQL/production restore kapsam dışı ve BLOCKEDdir.

Ana rapor: `reports/r20261108/QA_SECURITY_AUDIT_20261108.md`. Findingler: `reports/r20261108/qa-findings/FINDING-*.md`. Ham kanıtlar: `reports/r20261108/qa-raw/`. Product kaynağı QA tarafından değiştirilmedi; yetkilendirilmiş .11 scratch kopyasına owner patchi uygulanarak F02 bağımsız retest PASS oldu, .12 release kabulü owner bekler. F01 OPEN, HOST backup v2 yerel PASS + prod HOST-evidenced. Hiçbir failed regression xfail/PASSa çevrilmedi.


## 00:38 TSİ PROD-host steering
api-test yeni PROD host82.29.184.225tir (HOST aktarımı); sadece test VM sanmayın. Box resolver198.18.0.1 ile serving-IP bağı doğrulanamadı: precheck BLOCKED,0 HTTP; authoritative DNS/TLS peer+health kanıtı HOSTtan beklenir. `precheck.py` fail-closed; başarısız gate başka route ile aşılmaz. Mevcut00:30 negatif logları gate eklenmeden önceki dar kontrol olarak tarihsel korunur, yeni host eşliği iddia edilmez.

**Hiç authenticated chat/completions/model generation isteği gönderilmez**; GPU live checks KAN penceresine kadar NOT TESTED. Sadece invalid/forged/missing/revoked negatif token testleri ve yerel mock pozitifi uygundur. Production ledger/journal root SSH kanıtları BLOCKED; customer data/secret dumps istemeyin. Tam Shopify E2E SHOPa aittir.

## Evidence isolation / explicit source tree (08.11.2026 TSİ)
`python3 qa/audit12/run.py --tree /absolute/go/module --case ssrf`
Her invocation `qa-raw/runs/<timestamp-TSI>-<unique>/` açar; RUN_DIR stdouta yazılır. Go overlayler ve sonuçlar aynı run dizinindedir, eski kanıtlar üzerine yazılmaz. Python synthetic fixtures ve workspace-wide secret taraması QA-local scope kalır; --tree yalnız Go hedef ağacını değiştirir. Güncel backup case HOST-v2 synthetic7 testini koşar; tarihsel v1 failing repro `test_data_security.py DataSecurity.test_backup_missing_store_must_fail -v` altında korunur. Açık F01 exit1 verir; F02 düzeltilmiş kaynakta PASS beklenir.

00:48 sonrası HOST kanıtı ana raporda işlendi: yeni host/proxy rota ve sınırlı journal örneği HOST-provided; direct box peer gate hâlâ fail-closed. Eski blocker paragrafı tarihsel00:38 durumudur.

## R13 local-only suite
`qa/audit12/regress_r13.sh --tree /absolute/Go/source/module`
Commands/platform risk provenance: `reports/r20261108/QA_REGRESSION_R13_COMMANDS.md`. Kernel-unavailable Landlock SKIP is never PASS; native Windows Low IL remains NOT TESTED on Linux. Same-request F04 lock contract versus F07 DoS safety conflict remains explicit. Each suite gets a unique r13-runs directory.
