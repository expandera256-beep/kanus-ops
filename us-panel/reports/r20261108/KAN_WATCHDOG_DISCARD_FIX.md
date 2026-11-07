# Watchdog discardLocalSsd düzeltmesi (8 Kasım 2026, KAN)

Onay: Demir, 8 Kasım 00:12 TSİ (KAN.US sohbeti): "v4'te kalsın, KAN watchdog koduna diskleri silerek kapatma yolunu ekleyip yeniden deploy etsin".

- Kaynak: Cloud Build 31f9c0a8 kaynak tarball'ı (çalışan image ile aynı .py dosyaları), /workspace/us-panel/kan/watchdog-fix/
- Değişiklik: gce_provider.py (+10/-3), discard_fix.patch. preserve_local_ssd=true ise davranış aynı (discardLocalSsd=false). preserve_local_ssd=false ve env US_ALLOW_DISCARD_LOCAL_SSD=1 ise beta stop'a discardLocalSsd=true gidiyor. Env yoksa eskisi gibi reddediyor. Local SSD'si olmayan VM'de değişiklik yok.
- Testler: test_gce_discard.py 6/6 PASS (GCE stop mock'lanıyor; istek sorgusunda discardLocalSsd=true doğrulanıyor). Mevcut suite + yeni testler: 104 passed, 37 skipped.
- Build: b2e4027c-96f6-4f62-8122-8e7949c34690 (SA us-builder), image gateway:2026-11-08-discard
- Deploy: us-watchdog-00006-88h, 8 Kasım 00:26 TSİ, %100, Ready. env US_ALLOW_DISCARD_LOCAL_SSD=1. Policy latest = v5 (idle 600, boot 600, preserve_local_ssd false, max_uptime 1800, daily_runtime 3600, daily_starts 6).
- Geri dönüş: gcloud run services update-traffic us-watchdog --region=europe-west4 --to-revisions=us-watchdog-00005-z97=100
- Canlı doğrulama: bekliyor (kalan 4 açılıştan 1'i ile, §4 canlı test planına bağlı).
