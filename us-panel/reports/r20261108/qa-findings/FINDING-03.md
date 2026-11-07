# FINDING-03 — Backup taslağı hiçbir dosya yedeklenmediğinde success dönüyor

- Önem: **MEDIUM — operasyonel fail-open, eski kurulu v1**.
- Durum: **FIXED-PENDING-VERIFY** — v2 yerel retest PASS; production kurulum HOST kanıtına dayanır, bağımsız tam SHA/cron doğrulaması bekler.
- Yönlendirme: **HOST** birincil, **SRE** backup monitoring/restore kabulü. Bileşene göre önerilen routing; promptta bot tablosu yok.
- Kaynak: `deploy/host-hardening/backup/backup-us-license-hook-hourly.sh:30–32` eksik her kaynağı success ile atlar;`:40–43` store/plans/env çağrıları;`:49` nothing-backed-up echo yine exit0. Script`:2` DRAFTtır.

## Repro — yalnız sandboxa uyarlanmış kopya
```bash
python3 /workspace/us-panel/qa/audit12/test_data_security.py DataSecurity.test_backup_missing_store_must_fail -v
```
`qa/audit12/test_data_security.py:DataSecurity.test_backup_missing_store_must_fail` orijinal shell mantığını test kopyasında çalıştırır. Sadece sabit LOCK/ENVF/PLANS yolları temp fixture yollarına remap edilir; STORE/DEST_DIR zaten env override destekler. Üç input da eksik. /etc,/opt,/run gerçek dosyaları, üretim/env/backuplar okunmaz; VPS işlemi yok.

Beklenen: Zorunlu store yok veya başarılı backup sayısı0 ise nonzero exit ve alarm sinyali.
Gerçek: `exit_code=0`, `nothing_backed_up=true`; assertion FAIL/exit1. Tarihsel kanıt `qa-raw/data-security.log:1–18` ve `qa-raw/pre-host-source-retest/`; güncel `backup.log` v2 retestidir.

## Kök neden / etki
Eksik kritik inputlar soft-skip edilir ve son echo başarılı çıkışı maskeler. Cron/monitor exit statüsüne bakıyorsa backupsız koşumu başarılı sanabilir. HOSTun sonradan sağladığı `HOST_EVIDENCE_QA_20261108.md:16` eski DRAFTın aynı hash ile productionda kurulu olduğunu doğrular; bulgu yalnız taslak kapsamından çıkmıştır. Gerçek backup verisi QA tarafından açılmadı.

## Önerilen fix / §13
Required STORE missing/non-regular ise fail; optional PLANS/env politikasını açık tanımla;0 başarılı backup nonzero; JSON/DB integrity ve checksum manifest; son-success/age alarmı. Root-only izinler ve retention glob sınırını koru. Önceki gerçek snapshotları silme/değiştirme. QA product/deploy taslağını değiştirmedi. HOST/SRE fix sonrası aynı reproducer PASS ve mevcut sentetik SQLite+Go store restore regresyonları PASS olmalı. Production restore sadece onaylı bakım penceresinde owner tarafından; bu görevde **BLOCKED**.

## 08.11.2026 00:48 TSİ sonrası v2 retest
`HOST_EVIDENCE_QA_20261108.md:20` v2nin `/usr/local/sbin/backup-us-license-hook-hourly.sh` olarak kurulduğunu bildiriyor. QA yalnız **yerel** `test_backup_v2.py -v` çalıştırdı: **7 PASS**, raw `qa-raw/backup-v2-local.log`. Missing-store/plans ve symlink nonzero2, invalid JSON3; iki sentetik instance normal0, manifest+LAST_SUCCESS doğrulandı. v2 kaynak `deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh:23,28–29,37–43` eski fail-openı kapatır.

Güncel retest komutu:
```bash
python3 /workspace/us-panel/qa/audit12/run.py --case backup
```
Box tam SHA `d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0`; HOST `d77534e5…a0` fragmanıyla tutarlı, tam remote hash eşitliği henüz bağımsız kanıtlanmadı (`qa-raw/backup-v2-hash-comparison.json`). HOSTa göre iki ilk backup exit0/3dosya, manifest+LAST_SUCCESS var. Eski v1 dosya/repro tarihsel olarak korunur; runner artık v2yi test eder. Production restore drill ayrı **BLOCKED**, F05e bakınız. Ürün/deploy değişikliği QA tarafından yapılmadı.
