# Bogus Gateway / Windows / model E2E hazırlığı — .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Yerel fixture testleri paid/cancel/refund/replay, proxy auth/reserve/settle/SSE/seconds ve restart/flock davranışını doğrular; gerçek sipariş değildir.
- BLOCKED — HOST önce Go proxy + VPS->ds-a100 özel erişimini yetkili ortamda kurup doğrulamalı. .5 istemcisi PC SSH tüneli kullanmaz; otomatik VM aç/kapa entegrasyonu ayrıca açık iştir.
- NOT TESTED — Aşağıdaki adımlar dry-run planıdır. Gerçek sipariş, ödeme veya iade ancak Demir'in yeni açık onayıyla yapılır; bu rapor onay değildir. SMTP kapalı kalır.

1. NOT TESTED — kanus-test.myshopify.com test/Bogus Gateway kipini yetkili operatör doğrular. US_PLANS_FILE test product_id kataloğunu kullanır; 1 kredi ürün 15473713283325, saatlik ürün 15473713512701 (3600 sn). Gerçek kart yok, test e-postası operatör kontrolünde.
2. NOT TESTED — Üç webhook: orders/paid -> https://api-test.kanus.org/webhooks/shopify/orders-paid; orders/cancelled -> https://api-test.kanus.org/webhooks/shopify/orders-cancelled; refunds/create -> https://api-test.kanus.org/webhooks/shopify/refunds-create. Mevcut paid kaydını çoğaltmayın; Python/Cloud Run issuer/webhook aboneliği olmamalı.
3. NOT TESTED — Onay sonrası başarılı Bogus test siparişi -> tek order/lisans/kredi/outbox. Aynı/farklı delivery-id tekrarları, restart ve başarısız test ödeme lisans üretmemesi. SMTP açılmaz; anahtar özel operatör erişimiyle alınır ve loglanmaz.
4. NOT TESTED — Windows .5, US_LICENSE_URL=https://api-test.kanus.org; Go lisansı kaydedilir, kredi status okunur. Model yalnız POST /v1/chat/completions üzerinden; eski account/SSH/US_GATEWAY_URL yolu kullanılmaz. Bilinmeyen/iptal lisans upstream'e ulaşmamalı.
5. NOT TESTED — sub: başarı 1 kredi; hata/timeout/cancel/kesik SSE 0. Hourly: upstream dispatch'ten itibaren queue hariç, cold-start dahil ceil saniye; unused hold bırakılır. Aynı request'in kör tekrarını yapmayın; commit-sonrası teslim kaybı manuel uzlaşma ister.
6. NOT TESTED — GPU cold-start + 45 dk bekleme/heartbeat ve iptal; “hazır” gerçek metadata/token gelmeden gösterilmemeli. Aktif Go GPU lifecycle adapter tamamlandıktan sonra aktif/kuyruk varken stop yok, son görevden 899 sn sonra açık/900 sn sonrası stop ve SSD korunması ayrıca canlı kabul edilmeli.
7. NOT TESTED — Ayrı çok adetli siparişte kısmi/tam refund, duplicate refund-id, cancelled ve ters sıra; iptal key'i modelden önce reddedilmeli. Finansal-only iade otomatik çözüm sayılmaz.
8. NOT TESTED — TSİ zaman, .5 SHA256, maskelenmiş order/request kimlikleri, HTTP/SSE sonuçları, önce/sonra bakiye, reserved ve birim kaydedilir. Secret/license/e-posta loglanmaz.

## Son adaya ek kontrol
- NOT TESTED — Lisans mail kuyruğu subject/body Shopify order.name (#1001 gibi) olmalı; numeric id idempotency içindir. SMTP kapalı; eski pending kayıtta ad yoksa yeni deploy adı kendiliğinden geriye doldurmaz.
- NOT TESTED — EXE GET /v1/credits/status yanıtında kind, remaining/available, expiresAt, unit doğrulanmalı; POST status beklenmez. Eski hook'ta 404 diğer ajan tarafından raporlandı; yetkili deploy sonrası yeni aday ayrıca kabul edilmeli.
- NOT TESTED — 8 ürün requiresShipping=false olduğu parent tarafından bildirildi; bu yürütücü doğrulamadı/değiştirmedi. Sipariş/ödeme/iade adımları için ayrı Demir onayı gerekir; önceki başka ajan E2E raporu bu yeni adaya canlı kabul sayılmaz.

## Round6 proxy-only proje E2E
BLOCKED — Tunnel127.0.0.1:18000 ve ayrı canlı test onayı bekleniyor; bu tur yalnız local fake SSE kullanıldı. TestRealModelPanelBuildRentalProject US_REAL_MODEL_E2E=1 + US_PROXY_URL + US_TEST_LICENSE=<private file path> ister; tam komut ROUND6_RAPOR_TR.md. Env unset SKIP. Bogus sipariş/ödeme/iade yine ayrı Demir onayı gerektirir; bu tur yapılmadı.
