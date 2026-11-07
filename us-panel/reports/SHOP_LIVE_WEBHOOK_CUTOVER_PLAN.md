# Canlı webhook kesim planı (PLAN — uygulanmadı)

**Tarih:** 2026-11-06 08:09 Europe/Istanbul (not güncellemesi: 2026-11-06 09:12 Europe/Istanbul)  
**Hedef mağaza:** shop.kanus.org (`19einb-qz.myshopify.com`)  
**Production hook host (hedef):** `https://api.kanus.org` — **şimdilik mevcut değil; HOST hazırlıyor.** Kesim HOST + Demir onayı olmadan yapılmaz.  
**Bu dosya yalnız plandır.** Uygulama **Demir onayı** ister. Sırlar yazılmaz / üretilmez.

## 0. Önkoşullar (onay öncesi)

1. Production VPS’te Go `us-license-hook` `api.kanus.org` arkasında ayakta (`/healthz` → ok). **Durum (p12):** production `api.kanus.org` henüz yok — HOST hazırlıkta.  
2. Nginx/Caddy location’ları (örnek: `deploy/us-license-hook/nginx-location.example.conf`):
   - `POST /webhooks/shopify/orders-paid`
   - `POST /webhooks/shopify/orders-cancelled`
   - `POST /webhooks/shopify/refunds-create`
3. kanus-test webhook’ları **ayrı secret + ayrı shop domain** ile kalır; canlı secret’ı test ile **paylaşılmaz**.  
4. SMTP (`US_SMTP_ENABLE`) ayrı onay; lisans e-postası kapalıysa anahtar yine depolanır ama müşteriye gitmez — go-live’da netleştir.

## 1. Endpoint’ler (canlı)

| Topic | HTTPS URL (production) |
|---|---|
| orders/paid | `https://api.kanus.org/webhooks/shopify/orders-paid` |
| orders/cancelled | `https://api.kanus.org/webhooks/shopify/orders-cancelled` |
| refunds/create | `https://api.kanus.org/webhooks/shopify/refunds-create` |

Kod bağları: `work/kiyas/term/cmd/us-license-hook/main.go` (path→topic).  
Referans test URL’leri: `https://api-test.kanus.org/webhooks/shopify/...` (README_TR) — canlıya kopyalanmaz.

## 2. Ortam (canlı instance — değerler Admin/secret store’da)

| Değişken | Canlı beklenen | Not |
|---|---|---|
| `SHOPIFY_SHOP_DOMAIN` | `19einb-qz.myshopify.com` | Custom domain değil myshopify |
| `SHOPIFY_WEBHOOK_SECRET` | **Yeni**, canlı app’e özel | kanus-test secret’ından **farklı** |
| `US_LICENSE_SECRET` | Production istemciyle aynı | Değiştirilirse eski anahtarlar doğrulanmaz |
| `US_PLANS_FILE` veya `US_PLANS_JSON` | Aşağıdaki canlı product_id eşlemesi | `US_PLANS_FILE` öncelikli; bozuk dosya fail-closed |
| `US_STORE_PATH` | Canlı depo (test deposundan ayrı) | |
| `US_LISTEN` | `127.0.0.1:8787` | Public değil |

**US_CREDIT_PRODUCTS:** Python yolu kapalı; Go tarafında eşdeğer **`US_PLANS_JSON` / `US_PLANS_FILE`** (`product_id` → plan). Python `US_CREDIT_PRODUCTS` kullanılmaz (README_TR).

## 3. Canlı product_id eşlemesi (2026-11-06 salt-okuma)

Hook `line_items[].product_id` (numeric) ile eşler — başlık/fiyat değil.

Önerilen `US_PLANS_JSON` (LivePlansJSON / `licensesrv/pricing.go` Catalog ile uyumlu):

```json
{
  "15428080632036": {"kind":"sub","days":90,"prompts":1,"credits":1},
  "15428080664804": {"kind":"sub","days":90,"prompts":5,"credits":5},
  "15428084170980": {"kind":"sub","days":90,"prompts":10,"credits":10},
  "15432761344228": {"kind":"sub","days":180,"prompts":25,"credits":25},
  "15432761409764": {"kind":"sub","days":180,"prompts":50,"credits":50},
  "15432761442532": {"kind":"sub","days":365,"prompts":100,"credits":100},
  "15424693010660": {"kind":"sub","days":30,"prompts":200,"credits":200},
  "15424693108964": {"kind":"use","days":30,"seconds":3600}
}
```
*(US-CREDIT-1 — DRAFT vitrinde; **Demir kararı: US_PLANS / US_CREDIT_PRODUCTS eşlemesinde kalır.**)*

### Variant GID / numeric (referans)

| SKU | Product GID | product_id | Variant GID | variant_id | status |
|---|---|---|---|---|---|
| US-CREDIT-1 | gid://shopify/Product/15428080632036 | 15428080632036 | …/68119764304100 | 68119764304100 | **DRAFT** |
| US-CREDIT-5 | …/15428080664804 | 15428080664804 | …/68119764336868 | 68119764336868 | ACTIVE |
| US-CREDIT-10 | …/15428084170980 | 15428084170980 | …/68119790026980 | 68119790026980 | ACTIVE |
| US-CREDIT-25 | …/15432761344228 | 15432761344228 | …/68150471360740 | 68150471360740 | ACTIVE |
| US-CREDIT-50 | …/15432761409764 | 15432761409764 | …/68150471426276 | 68150471426276 | ACTIVE |
| US-CREDIT-100 | …/15432761442532 | 15432761442532 | …/68150471459044 | 68150471459044 | ACTIVE |
| US-MONTHLY-200 | …/15424693010660 | 15424693010660 | …/68094993465572 | 68094993465572 | ACTIVE |
| US-HOUR-1 | …/15424693108964 | 15424693108964 | …/68094994317540 | 68094994317540 | ACTIVE |

### US-CREDIT-1 (DRAFT) — karar kilitli

- Vitrinde satışta değil (DRAFT).  
- **Demir kararı:** `US_PLANS_JSON` / kredi ürün eşlemesinde (**US_CREDIT_PRODUCTS** Go karşılığı) **kalır** — eski/manuel siparişler lisans üretebilir; yeni vitrin satışı yok.  

## 4. HMAC ve güvenlik

1. Shopify Admin’de canlı custom app veya webhook kaydı → signing secret → yalnız sunucu env.  
2. Hook raw body üzerinde `X-Shopify-Hmac-Sha256` doğrular (sabit zamanlı); bozuk HMAC → 401.  
3. İsteğe bağlı `X-Shopify-Shop-Domain` = `19einb-qz.myshopify.com`.  
4. Path başına topic bağları: yanlış topic → 400.  
5. Proxy raw body’yi bozmamalı (nginx buffering dikkat).

## 5. Idempotency

- `X-Shopify-Webhook-Id` + order id / refund id depoda.  
- Tekrar `orders/paid` → ikinci lisans yok.  
- `orders/cancelled` tombstone; geç `paid` kredi vermez.  
- `refunds/create` kısmi/tam iade mevcut davranışa göre (kod: `reversals.go`).  
- Test ve canlı **ayrı store dosyası** — ID çakışması / yanlış revoke riski.

## 6. Shopify’da kayıt adımları (onay sonrası)

1. Canlı shop’ta webhook app/secret hazırla (test app’ini paylaşma).  
2. Üç abonelik ekle (JSON format, yukarıdaki URL’ler).  
3. Env’e secret + shop domain + plans yaz; hook’u yeniden başlat.  
4. Admin → Webhooks’ta delivery log’u aç.

*Not:* MCP `publishableUnpublish` benzeri güvenlik politikaları webhook create’i de engelleyebilir; kayıt büyük olasılıkla **Shopify Admin UI** veya yetkili custom app üzerinden yapılacak.

## 7. Test siparişi checklist

1. ACTIVE bir paket (örn. US-CREDIT-5 veya ucuz test SKU) — **DRAFT CREDIT-1 ile test etme**.  
2. Gerçek/ödeme sağlayıcı test modu veya düşük tutarlı sipariş (Demir onayı).  
3. Beklenen: `orders/paid` 200; depoda 1 lisans; (SMTP açıksa) mail; idempotent ikinci delivery.  
4. İptal veya iade testi (Admin policy izin veriyorsa); lisans revoke / kısmi davranış.  
5. Yanlış secret ile 401; yanlış shop domain reddi.

## 8. İzleme

- Hook log: HMAC fail oranı, 5xx, duplicate delivery.  
- `/healthz` uptime.  
- Outbox mail kuyruğu (SMTP kapalıysa birikim).  
- İlk 24s: her paid siparişte lisans sayısı = eşleşen satır adedi.

## 9. Geri alma (rollback)

1. Shopify’da üç webhook’u **devre dışı bırak / sil** (canlı teslimat dursun).  
2. Hook process’i durdurma isteğe bağlı; durursa beklemede kalan Shopify retry’leri 5xx yer.  
3. Yanlış üretilmiş lisans: revoke + müşteri iletişimi (manuel).  
4. kanus-test ortamına dokunma.  
5. Plans/secret geri alma notunu operasyon günlüğüne yaz.

## 10. Onay kapısı

**Uygulanmaz** ta ki Demir açıkça şunu onaylasın: canlı `shop.kanus.org` için üç webhook’un `api.kanus.org` path’lerine bağlanması, ayrı canlı secret, plans dosyasının yazılması ve (varsa) CREDIT-1 mapping kararı.
