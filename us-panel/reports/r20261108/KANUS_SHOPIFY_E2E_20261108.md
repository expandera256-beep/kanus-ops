# Shopify E2E — yeni host kabulü

Zaman: 2026-11-08 00:28:14 Europe/Istanbul.
İlk okunan kriter: KANUS_SHOPIFY_ACCEPTANCE.md. Canlı mağaza/tema değiştirilmedi; sunucu değişikliği, restart, test siparişi veya ödeme yapılmadı.

| Adım | Sonuç | Kanıt / sınır |
|---|---|---|
| DNS yeni hedef | PASS | Google DoH A=82.29.184.225; kalan TTL=275 saniye (300 kayıt TTL, cache azalır). Önceki gözlem A=31.220.63.191 idi. |
| HTTPS health | PASS | GET https://api-test.kanus.org/healthz → 200, body `ok`, TLS verify=0; Server nginx/1.28.3 (Ubuntu). |
| Origin kimliği | UNKNOWN | DNS yeni IP ile uyumlu; yanıtta benzersiz host/version kimliği yok. Proxy IP 198.18.0.1 origin kanıtı değildir. deploy/host-hardening/PROD_PLAN.md yeni VPS IP’sini kaydeder ama güncel isteğin fiziksel hostunu tek başına kanıtlamaz. Eski sunucudan geldiğine dair kanıt yok. |
| kanus-test erişimi | FAIL / BLOCKED | İzin verilen switch-shop çalıştı; live 19einb-qz.myshopify.com erişimi iptal edildi. Connector: “every Shopify tool now fails until the user signs in to Shopify again and picks the store to use. Do not call another Shopify tool in this turn.” Kullanıcı talimatı gereği duruldu, workaround yok. |
| Test paid sipariş / webhook | BLOCKED | Test mağazası yetkilendirilmedi; sipariş/webhook ID yok. |
| Hesap / lisans / kredi / mail / panel | BLOCKED | Yeni işlem yok; maskeli lisans ID, önce/sonra bakiye veya e-posta message ID kanıtı yok. Health200 kabul sayılmadı. |
| Cancellation / refund / idempotency | BLOCKED | Test çalıştırılmadı; replay/HMAC gönderilmedi. |
| Live connector’a geri dönüş | BLOCKED | Auth düştü; aynı tur başka Shopify çağrısı yapılmadı. |

## Engeller ve devam koşulu
- Shopify connector’da kanus-test için yeniden giriş ve mağaza seçimi gerekiyor. Sonra yalnız test/Bogus gateway doğrulanarak zincir koşulmalı.
- Yeni origin kimliği HOST’un güncel read-only kanıtıyla kesinleştirilmeli.
- Tam zincir PASS değil: SHOP_AUTOMATION_COPY_DRAFT_TR.md oluşturulmadı; otomasyon vaadi taslağı hazırlanmadı.

## Yeniden yetkilendirme sonrası kontrol — 2026-11-08 00:31:37 Europe/Istanbul

| Adım | Sonuç | Yeni kanıt |
|---|---|---|
| Mağaza kimliği | PASS | get-shop-info + GraphQL shop: kanus-test.myshopify.com; Basic App Development, USD. Live mağazaya erişilmedi/değişiklik yapılmadı. |
| Test gateway tarihsel kanıtı | PASS (tarihsel) | #1001 / Order 18997703606525: test=true, paymentGatewayNames=[bogus]. #1002 / 18997716943101: test=false, manual; yeni test için kopyalanmadı. |
| Test ürün eşlemesi | PASS | CREDIT-1 Product 15473713283325 / Variant 71819012604157 / 2.98 USD; diğer 7 test ürün de okundu. |
| Webhook aboneliği | UNKNOWN | webhookSubscriptions(first:100) nodes=[]; connector uygulaması kapsamı. Bu, Admin'den/başka app'ten kayıtlı aboneliklerin yokluğunu kanıtlamaz. Yeni hosta gerçek orders/paid aboneliği doğrulanamadı. |
| API ile test sipariş desteği | PARTIAL | orderCreate şeması test:Boolean ve transactions alanlarını sunuyor. Yalnız browser checkout mümkün olduğu sonucuna varılmadı. |
| Yeni paid sipariş / delivery | BLOCKED / NOT RUN | Önkoşullar eksik: doğrulanmış otomasyon test alıcısı/mailbox; yeni hostta belgelenmiş mevcut SSH bağlantısı ve salt-okuma ledger/mail/panel kanıt yolu; abonelik sahipliği/target doğrulaması. Belirsiz alıcıya otomatik lisans e-postası tetiklenmedi. |
| Hesap / lisans / bakiye / mail / panel | BLOCKED | Yeni order ID/webhook ID/license ID/balance before-after/message ID yok; tarihsel kanıt yeni host PASS sayılmadı. |
| Cancel / refund / replay | BLOCKED / NOT RUN | Tarihsel raporda MCP iptal/iade engelleri var; bunlar yeniden denenmedi veya alternatif yolla aşılmadı. Test HMAC secret okunmadı/üretilmedi. |
| Connector durumu | PASS | kanus-test üzerinde bırakıldı; switch-shop çağrılmadı. |

Tam E2E PASS değil. Otomasyon metni taslağı oluşturulmadı. Devam için HOST yeni origin ve salt-okuma log/ledger yollarını, operatör test alıcısını ve gerçek webhook aboneliklerini doğrulamalı. Hiçbir Shopify mutation, test siparişi, gerçek ödeme, sunucu değişikliği/restart veya müşteri mesajı yapılmadı.

## Yetkili test siparişi denemesi — 2026-11-08 00:33:47 Europe/Istanbul

§15 HOOK_MOVE_PROD_VPS_20261106.md okundu: yeni test instance /opt/us-license-hook-test/licenses.json, port8788; smoke kayıt 990000000001 kabul kanıtı değildir ve dışlandı. HOST ledger/mail modu henüz gelmedi: UNKNOWN. qa-acct adresinin onaylı mailbox/outbox modu §15'te doğrulanmadı; kullanıcı talimatıyla ayrılmış `kanus-e2e-20261108@example.org` kullanıldı.

| Adım | Sonuç | Kanıt |
|---|---|---|
| Connector kimliği | PASS | get-shop-info: kanus-test.myshopify.com; switch-shop yapılmadı. |
| Plan/ürün | PASS | Test plans.json Product15473713316093 → sub,90 gün,5 kredi; Variant71819012636925,13.89USD. |
| Test paid orderCreate | FAIL / BLOCKED | test=true; transaction test=true,gateway=bogus,SALE/SUCCESS,13.89USD; orderCreate=null. Shopify: `Access denied for orderCreate field. Required access: write_orders access scope. Also: This mutation is only accessible to apps authenticated using offline tokens.` Sipariş oluşturulmadı; order ID yok. |
| Gerçek orders/paid teslimi | BLOCKED | Sipariş yok; webhook ID/HTTP200 kanıtı üretilemedi. HOST smoke200 yeni sipariş yerine sayılmadı. |
| Lisans/5 kredi/hesap/panel | UNKNOWN / NOT RUN | Yeni lisans veya bakiye kanıtı yok; HOST ledger bekleniyor. |
| Mail modu/teslimi | UNKNOWN | HOST bekleniyor; example.org gerçek teslim testi değildir. |
| Exact payload idempotency | BLOCKED | Yakalanmış yeni gerçek sipariş payload/header yok; uydurulmuş replay gönderilmedi. Secret okunmadı/yazdırılmadı. |
| İkinci sipariş iptali/üçüncü sipariş iadesi | BLOCKED | orderCreate erişim hatası sonrası alternatif draft/alt kanal ile aşılmadı; yeni sipariş yok. |

Engel: connector orderCreate için write_orders + offline token erişimi sağlamıyor. Operatör kanus-test Bogus checkout ile test siparişi oluşturabilir veya uygun yetkili resmi entegrasyonla çalıştırabilir. Browser'ın tek teknik seçenek olduğu kanıtlanmadı; bu connector ile API denemesi engelli. Canlı mağaza/tema dokunulmadı; sunucu değişikliği/restart yok. Connector kanus-test'te kaldı. Tam zincir PASS olmadığı için SHOP_AUTOMATION_COPY_DRAFT_TR.md hazırlanmadı.

## Güncelleme 00:46 TSİ — Bogus test siparişi #1003 (checkout #35ZNXD8RW)
Kaynak: tarayıcı checkout (SHOP) + HOST'un maskeli sunucu dökümü (srv2133566, binary sha256 7d72f449…56f0).
| Adım | Sonuç | Kanıt |
|---|---|---|
| Test siparişi (Bogus) | PASS | #1003, id 19002359972093, 1 Model İsteği x1, 2.98 USD, kargo yok |
| orders/paid → api-test.kanus.org (82.29.184.225) | PASS | nginx 00:37:46 TSİ POST /webhooks/shopify/orders-paid 200, Shopify-Captain-Hook, webhook-id d3e85ed8-…5630 |
| HMAC doğrulama | PASS (dolaylı) | 200 + kayıt oluştu; imzasız istek 401 |
| Lisans basımı | PASS | tek lisans us1.…pLm8, productId 15473713283325, 1 kredi, 90 gün, revoked=false |
| Kredi | PASS | kalan 1, reserved 0 |
| E-posta | PASS (outbox) | outbox satır 14, skipped: recipient not on SMTP allowlist (beklenen) |
| Hesap / panelde görünme | UNKNOWN | hesap bağlama US uygulamasında yapılıyor; panel kontrolü yapılmadı |
| Tekrar teslim (idempotency) | NOT RUN | Shopify tekrar göndermedi; yakalanmış payload yok |
| İptal / iade | NOT RUN | bağlayıcıda write_orders yok; Admin üzerinden yapılmalı |
Not: journalctl'de istek başına log yok (binary yazmıyor). Outbox satır 13: 00:33:46'da p***@kanus.org hesap kaydı (skipped) — kaynağı belirsiz, SHOP oluşturmadı.

## Güncelleme 00:48 TSİ — #1003 iptal + iade (Demir onayı)
- kanus-test Admin'den iptal edildi, 2.98 USD tam iade (Bogus), müşteri bildirimi kapalı. Doğrulama 00:48:05 TSİ.
- Sunucu tarafı (orders/cancelled, refunds/create, lisans revoke) HOST dökümü bekleniyor.
- Outbox satır 13 (00:33, p***@kanus.org): KAN.US açtı — PERF login yük testi için lisanssız sentetik hesap, beklenen kayıt.

## Güncelleme 00:52 TSİ — İptal/iade sunucu doğrulaması (HOST dökümü, srv2133566)
| Adım | Sonuç | Kanıt |
|---|---|---|
| orders/cancelled | PASS | nginx 00:48:02 POST /webhooks/shopify/orders-cancelled 200, tek istek |
| refunds/create | PASS | nginx 00:48:02 POST /webhooks/shopify/refunds-create 200, tek istek |
| Lisans revoke | PASS | us1.…pLm8 revoked=true, kredi 1→0, cancelled[19002359972093]=true |
| İade kaydı | PASS | refunds[1116251390205] orderId 19002359972093, line 50913839546621 x1 |
| Çift lisans yok | PASS | orders[19002359972093] tek lisans |
| Outbox | PASS | yeni mail yok; satır 14 dead=true, lastErr "sipariş iptal/iade" |
| Idempotency (aynı webhook tekrar) | NOT RUN | Shopify tekrar göndermedi; doğal tekrar yok |
Webhook id'leri (store): 02202401-…c34a, 65ff2c20-…17df (21:48:02Z).

## Nihai özet — 2026-11-08 00:53:12 Europe/Istanbul

Aşağıdaki sonuçlar KAN.US/Demir'in 00:52 TSİ bildirimine dayanır; bu ajan tarafından yeni sunucu/API doğrulaması yapılmadı. Önceki connector orderCreate erişim hatası tarihsel olarak geçerlidir; #1003'ün bu ajan tarafından oluşturulduğu iddia edilmez.

| Adım | Nihai durum | Bildirilen kanıt / sınır |
|---|---|---|
| Paid sipariş | PASS | #1003 → webhook HTTP200 → lisans üretildi, 1 kredi. Numeric order ID, webhook ID ve maskeli lisans ID bu bildirimde verilmedi. |
| Outbox | PASS (beklenen atlama) | Allowlist nedeniyle skipped; gerçek müşteriye SMTP teslimi PASS değildir, message ID yok. |
| Cancel + refund | PASS | Her iki webhook HTTP200; lisans revoked; kredi 1→0. Ayrı ayrı reversal etkisi bu bildirimden ayrıştırılamaz. |
| Idempotency | SKIPPED | Demir kararı, 2026-11-08 00:52 TSİ (Europe/Istanbul). Test edilmedi; PASS sayılmaz. |
| Hesap/panel görünürlüğü | UNKNOWN | Lisans bağlama US uygulamasının Hesap ekranında yapılır; otomatik hesap açılışı veya panel görünürlüğü doğrulanmadı. |

Bu kapsam sınırlı test kabulüdür; tam üretim E2E PASS değildir. Canlı api.kanus.org hook'u ve live webhook cutover henüz yapılmadı; test SMTP allowlist'li; idempotency test edilmedi. Metinler yalnız taslak olarak SHOP_AUTOMATION_COPY_DRAFT_TR.md dosyasına yazıldı; canlı yayın/cutover onayı değildir. Shopify mutation veya yayın yapılmadı.
