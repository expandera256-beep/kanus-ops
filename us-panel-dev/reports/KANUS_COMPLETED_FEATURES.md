# Tamamlanan özellikler — .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — §2 terminal: yalnız açık kullanıcı eylemiyle JSON argv; shell yok; seçili proje cwd, env sır süzme, ayrı stdout/stderr, gerçek exit code, process-tree cancel, wall timeout, bounded output ve tek eşzamanlı terminal. TestTerminalGUIActionReturnsRealExitCode ve terminal gerçek Python fixture testleri.
- PASS — §2 Go proje çalıştırma: go.mod + main.go için go run .; Go kurulu değilse gerçek hata. TestGoLaunchDetection. Önceki test/onarım, çok dosya öneri/onay, kalıcı kuyruk/checkpoint ve editör güvenlik regresyonları korundu.
- PASS — §2/3 Go kredi durumu: GET /v1/credits/status; doğrulanmış gözlem zamanı, kredi/saniye birimi, bakiye/rezerve ve son 20 lisans-scoped settlement. Anahtar/e-posta dönmez; okuma depoyu/modeli değiştirmez. TestCreditStatusReadOnlyScopedAndPrivate.
- PASS — A / §3 model proxy: POST /v1/chat ve /v1/chat/completions; lisans/model/body sınırı, server reserve/settle, SSE/JSON, hata/cancel/timeout release, terminal bitişi ancak disk settlement sonrası. Üst servise lisans sızmıyor; upstream token ayrı. TestModelProxy* ve HookRelease uçtan uca yerel HTTP fixture'ı.
- PASS — B saatlik: use=3600 saniye. Kuyruk hariç upstream dispatch'ten itibaren cold-start dahil ceil(wall-second), en az 1 sn; başta timeout/kalan kadar hold, başarıda gerçek saniye, başarısızlıkta tam release. Bitmiş yeni paket yeniden açılmaz; hiç kullanılmamış eski paket kontrollü transaction ile birime çevrilir.
- PASS — C soğuk başlangıç: istemci 50 dk, sunucu varsayılan 2700 sn; 15 sn SSE heartbeat; GPU ısınıyor/kuyruk durumundan yalnız gerçek token geldiğinde akış durumuna geçer. Ticari US_GATEWAY_URL/SSH ayarı geçersiz, dev flag ticari kapıyı aşamaz.
- PASS — D depo: Linux lifetime flock ve Windows LockFileEx kaynakları; tek writer, temp+fsync+rename+Linux directory fsync; parent writer açıkken çocuk süreç reddedilir, close sonrası çocuk yazısı kalıcıdır. OpenStore boş/whitespace kullanılabilirlik ve invalid JSON testleri.
- PASS — GPU policy: WorkerPolicy/default örnekte 900 sn idle, 600 sn boot; 899 sn'de stop yok, 900'de stop; kuyruk/aktif lease koruması fake-provider ile testli. Go SSH dev readiness de 10 dk sınırına çıkarıldı.
- NOT TESTED — Bu maddeler gerçek model, GCP start/stop, Windows veya üretim kabulü değildir. OS CPU/RAM sandbox, hesap girişi ve aktif GPU entegrasyonu kalan işlerdir.

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
