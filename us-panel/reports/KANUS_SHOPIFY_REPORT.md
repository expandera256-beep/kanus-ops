# Shopify — .5 Go tek issuer

Güncelleme: 06.11.2026 07:14 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme veya adlı E2E kanıt dosyasıdır; shop.kanus.org veya ücretli canlı lansman kabulü değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Üç HMAC/topic/shop handler, idempotent paid/cancel/refund/tombstone/kısmi adet iadesi ve atomik kredi testleri geçti. Go product_id eşler; US_PLANS_FILE önceliği ve test/canlı plan parser uyumu korunur. Kanıt: KANUS_TEST_RESULTS.md (Go suite).
- PASS — Model proxy sub için 1 kredi, use için saniye ölçer; Go status kullanım özeti sunar. Python/Cloud Run issuer yok; account login ayrı bir özellik olarak eksiktir.
- PASS — kanus-test E2E (06.11.2026 06:15 TSİ): orders/paid idempotency A/B, kötü HMAC C, #1002 gerçek paid→lisans, kredi API F. Kanıt: E2E_HOOK_KANUS_TEST_20261106.md.
- PASS — kanus-test gerçek Shopify teslimatları: `refunds/create` (06:18–06:19 TSİ) ve `orders/cancelled` (06:19:20 TSİ) HTTP 200; lisans revoked. Kanıt: KANUS_TEST_RESULTS.md «Güncelleme 06.11.2026 06:23 TSİ» (nginx Shopify-Captain-Hook logları; KAN.US).
- PASS — Round 5 VPS: model proxy + `/v1/credits/status` nginx yolları canlı; lisanslı status 200. Kanıt: HOOK_VPS_STATUS_20261106.md Round 5. Gerçek GPU upstream hâlâ BLOCKED.
- BLOCKED — SMTP kapalı; otomatik mail teslimatı yapılmadı. Shopify Admin MCP `refundCreate` / `orderCancel` financial policy.
- BLOCKED — Native Windows EXE smoke (Demir PC).
- NOT TESTED — shop.kanus.org; gerçek ücretli müşteri siparişi (Demir onayı bekler). E2E_HAZIRLIK_TR.md dry-run çerçevesi geçerlidir.
- NOT TESTED — Mağaza Shopify planı / TR üçüncü taraf gateway sözleşmesi (birim ekonomi varsayımları UNIT_ECONOMICS.md).

## E2E (kanus-test) özet — kanıt dosyaları

| Durum | Madde | Kanıt |
|---|---|---|
| PASS | orders/paid idempotency A/B | E2E_HOOK_KANUS_TEST_20261106.md |
| PASS | HMAC reddi | aynı |
| PASS | orders/paid gerçek (#1002, 2.98 USD Bogus) | aynı + KANUS_TEST_RESULTS.md |
| PASS | refunds/create gerçek teslimat | KANUS_TEST_RESULTS.md 06:23 |
| PASS | orders/cancelled gerçek teslimat | KANUS_TEST_RESULTS.md 06:23 |
| PASS | kredi reserve→settle(commit=false) | E2E_HOOK… |
| PASS* | iade/iptal imzalı REPLAY (Admin MCP BLOCKED iken) | E2E_HOOK… PASS* |
| BLOCKED | Admin refundCreate / orderCancel MCP | E2E_HOOK… |
| NOT TESTED | Native Windows; shop.kanus.org; ücretli Go-Live | — |

### Ürün / mail bulguları (değişiklik yapılmadı)

- Varyant kargo `requiresShipping` tarihsel E2E bulgusu vs sonraki FYI: rapor only; bu tur ürün güncellenmedi.
- Lisans mail order.name düzeltmesi yerel kodda PASS; SMTP gönderilmedi.

## E2E geri bildirimi — yerel kod (canlı deploy iddiası değil; Round 5 ayrı)

- PASS — Lisans e-postası subject/body order.name; TestLicenseMailUsesOrderNameAndKeepsIDIdempotency. SMTP yok.
- PASS — EXE/hook GET /v1/credits/status; Round 5’te VPS’te yol doğrulandı (HOOK_VPS_STATUS).

## Round7 — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- PASS — Shopify issuer Go olarak kaldı; accounts redeem yeni lisans üretmez, ownership link yazar. HMAC/idempotency/refund/cancel/order.name regresyonları v4 suite geçti. Üç webhook URL değişmedi.
- NOT TESTED — Bu tur Shopify/VPS çağrısı yok; önceki KAN.US canlı E2E kayıtları aynen korundu, bağımsız yeniden doğrulanmadı.
