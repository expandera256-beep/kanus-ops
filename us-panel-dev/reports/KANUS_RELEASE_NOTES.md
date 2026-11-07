# Sürüm notları — 2026.11.06.5

Güncelleme: 06.11.2026 06:21 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — .4 desktop-reservation/tünel modeli yerine Go server-model-proxy otoritesi. .5 varsayılan model/lisans adresi api-test.kanus.org; müşteri lisansı sadece Go'ya, upstream token ayrı. Ticari SSH veya US_GATEWAY_URL ile doğrudan bypass yok.
- PASS — Yeni yollar POST /v1/chat, /v1/chat/completions; GET /v1/models ve /v1/credits/status. Mevcut üç Shopify topic ve legacy reserve/settle korunur. HOST model streaming route timeout/buffering ayarlarını güncellemeli.
- PASS — Yeni env: US_MODEL_UPSTREAM, US_MODEL_UPSTREAM_TOKEN (isteğe bağlı), US_MODEL_NAME, US_MODEL_TIMEOUT_SECONDS=2700, US_MODEL_CONCURRENCY=1, US_MODEL_QUEUE_LIMIT=32. US_PLANS_FILE korunup öncelik/hata testleri eklendi. Dev-only US_DEV_DIRECT_MODEL=1 default kapalı; ticari sürümü açmaz.
- PASS — Saatlik gerçek saniye metering; server terminal-before-settlement engeli; cold-start heartbeat; 50 dk desktop timeout; terminal/Go launch/credit summary/local queue; lifetime writer lock ve empty-store düzeltmesi.
- PASS — Konfigürasyonda 15 dk idle ve 10 dk boot; fake-provider sınır testleri. Bu aktif GCP entegrasyonu tamamlandı demek değildir.
- PASS — Test/vet/race/derleme/JS syntax geçti. .3/.4 dosyaları silinmedi; .5 yeni adaydır. Yeni hook paketi yalnız yerelde üretildi, deploy edilmedi.
- BLOCKED — Windows native ve VPS özel upstream; aktif GPU start/stop/watchdog entegrasyonu açık. Üretim hazır onayı yok.

```text
1ef0f70429953a79de7ea95f21b028f9fb0ad22412e9b399d792bc388423f863  /workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_5.exe
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/dist/us-license-hook_linux
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/deploy/us-license-hook/us-license-hook
0e599d2877087eb6efaf97f34fa711ad3263a18c272897d89cc52af88a6fa39a  /workspace/us-panel/deploy/us-license-hook_hostinger.tgz
```

- PASS — Test runner artık Go/Python/Node gerçek test kanıtı ister; boş/skip-all/yalnız exit0 çıktısını başarı saymaz. Yeni gerçek Go+Node fixture testi geçti. Proxy yavaş istemci yazmalarına 20 sn write deadline ve flush hata kontrolü uygular; regresyon testi geçti.

## E2E geri bildirimi — yerel kod düzeltmesi (canlı deploy değildir)

- PASS — Lisans e-postası subject/body order.name (ör. #1001) kullanır. Numeric order.id idempotency anahtarı olarak kaldı; outbox OrderName kalıcıdır, restart + partial-refund mail yenilemesinde korunur. TestLicenseMailUsesOrderNameAndKeepsIDIdempotency geçti; SMTP gönderilmedi.
- PASS — EXE gateway.go status() HTTP GET /v1/credits/status çağırır. Yeni hook lisanslı GET sunar; remaining/credits, kind, expiresAt, reserved/available, unit ve scoped history döner. POST status 405; geçersiz/eksik lisans 401; route ve privacy testleri geçti. nginx/Caddy listeleri bu yolu içerir.
- NOT TESTED — E2E'nin eski deployed binary'deki 404 ve numeric mail bulguları diğer ajan kanıtıdır; bu yürütücü remote doğrulama yapmadı. Yeni paket deploy edilmedi. Eski outbox'ta orderName yoksa gerçek #1001 adı otomatik tahmin edilmez; eski pending kayıtlar gerekirse operatör incelemesi ister.

## Round 6 — 06.11.2026 07:20 TSİ (Europe/Istanbul)
- PASS — Yerel fake SSE HTTP yükü:50 ayrı lisans, queue32+active1;33 success/17 beklenen429, unexpected0; p50=104.970ms,p95=334.002ms. Tek kredi50 paralel reserve:1 success/49 beklenen402; negatif/double-spend yok,p50=5.084ms,p95=6.136ms. Expired/revoked401;1.1sn hourly2sn debit. Raw LOAD_TEST_20261106.txt (mock, gerçek GPU performansı değil).
- PASS — --version/--selftest/--check-license file/--check-server erken GUI'siz dispatch; log-first Windows console fallback, maskeli key, GET status/health ve redirect engeli unit testleri. Gerçek Linux version/selftest main çağrısı geçti.
- PASS — Go289+37,6 skip,0 fail; full race ve vet exit0; v3 raw ve .6 cross-build güncel. EXE SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327.
- PASS — Hook production kodu/binary/tgz değişmedi, yeniden build/deploy yapılmadı. Lisanslı model proxy ve Go issuer mimarisi korundu.
- BLOCKED — Native Windows ve VPS model upstream127.0.0.1:18000 tunnel; gerçek GPU start/idle-stop entegrasyonu. Gerçek rental E2E env unset temiz skip; canlı model çalıştırılmadı.
- NOT TESTED — VPS Round5 deploy bilgisi aktarıldı, bu yürütücü doğrulamadı; mevcut KAN.US E2E kayıtları korunmuştur. Tam komutlar, metrikler, exit kodları ve açık testler ROUND6_RAPOR_TR.md.
