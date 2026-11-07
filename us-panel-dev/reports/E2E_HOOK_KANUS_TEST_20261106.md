# E2E Hook kanus-test — kanıt (06.11.2026 06:15 TSİ Europe/Istanbul)

 Bogus / Hook — 06.11.2026 06:15 TSİ (Europe/Istanbul)

**Yürütücü kanıtı (bu ajan):** VPS SSH + public api-test.kanus.org + Shopify MCP (kanus-test). shop.kanus.org'a dokunulmadı.
Yedek: `/opt/us-license-hook/backup-e2e-20261106030421.json`.
Önceki durum: #1001 (18997703606525) Bogus PAID, 1 lisans (sub / 90 gün / 1 kredi).

| Durum | Adım | Kanıt |
|---|---|---|
| PASS | A Idempotency (aynı webhook id) | POST `/webhooks/shopify/orders-paid` + Webhook-Id `fc656215-5914-515e-a55f-a7e91c3fb9d1` → HTTP **200**; `license_count=1`; `webhooks_count=1` |
| PASS | B Idempotency (yeni webhook id, aynı sipariş) | Webhook-Id `22244209-0219-404b-8197-0b975d71ce18` → HTTP **200**; `license_count=1`; `webhooks_count=2` |
| PASS | C Kötü HMAC | HMAC `aW52YWxpZA==` → HTTP **401** `unauthorized` |
| PASS* | D Tam iade (#1001) | Shopify Admin `refundCreate` MCP financial policy ile **BLOCKED**. Hook: imzalı `refunds/create` **REPLAY** HTTP **200** (wid `f113ba79-6935-47d8-8dbf-4ef2a34e8570`); `refunds["9000000000001"].lines["50904787550461"]=1`; lisans `revoked=true`, `credits=0`; outbox `dead=true` / `lastErr=sipariş iptal/iade` |
| PASS* | E İptal senaryosu | `draftOrderCreate` #D1 + `draftOrderComplete` → gerçek sipariş **#1002** (18997716943101) PAID **2.98 USD**; gerçek `orders/paid` → 2. lisans `credits=1`. `orderCancel` MCP **BLOCKED**. Hook: imzalı `orders/cancelled` **REPLAY** HTTP **200** (wid `fcb75b27-616b-4cc0-86bb-c7404301b14d`); `cancelled["18997716943101"]=true`; lisans `revoked=true`, `credits=0`. Not: Shopify admin'de #1002 hâlâ PAID |
| PASS* | F Kredi API reserve→release | #1001 lisansı geçerliyken localhost:8787 `POST /v1/credits/reserve` HTTP **200**; `POST /v1/credits/settle` `commit=false` HTTP **200** `ok`; store `credits=1` `reserved=0`. `GET /v1/credits/status` HTTP **404** (deploy binary'de status yolu yok; strings yalnız reserve/settle) |

### Bulgular (değişiklik yapılmadı)

1. **Kargo:** ürün `15473713283325` varyant **`71819012604157`** — `inventoryItem.requiresShipping=true`. #1001 shipping **8.00 USD** (toplam 10.98). Dijital ürün için `requiresShipping=false` olmalı. Bu turda güncellenmedi.
2. **E-posta gövdesi:** outbox `Siparişiniz (#18997703606525)` yazıyor; sipariş adı `#1001` değil. Kaynak: `orderMailBody` numeric order id kullanıyor.

### Özet

- PASS: A, B, C, F (reserve/release)
- PASS (replay): D hook iade, E hook iptal; ayrıca E gerçek paid webhook
- BLOCKED: gerçek Shopify `refundCreate` / `orderCancel` (MCP)
- NOT TESTED: Shopify kaynaklı gerçek `refunds/create`; `/v1/credits/status` (deploy)
