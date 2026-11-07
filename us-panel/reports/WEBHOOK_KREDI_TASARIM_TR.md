# Shopify Webhook + Kredi Tasarımı (taslak, canlıya yazılmadı)

## Akış
Shopify `orders/paid` → HTTPS (ör. `https://api.kanus.org/webhooks/shopify/orders-paid`, ters vekil → `us-license-hook` 127.0.0.1:8787)
→ HMAC doğrula → idempotentlik (webhook-id + order-id) → product_id → plan → lisans anahtarı üret → depoya yaz (atomik) → outbox’a e-posta → arka planda SMTP (Hostinger 465) ile gönder, hatada geri çekilme.
Hata durumunda 500 döner, Shopify yeniden dener; tekrar teslim güvenli.

## Ortam değişkenleri (sırlar yalnız ortamda / gizli yöneticide)
| Değişken | Açıklama | Durum |
|---|---|---|
| `SHOPIFY_WEBHOOK_SECRET` | Uygulamanın client secret’ı (webhook HMAC anahtarı) | EKSİK |
| `SHOPIFY_SHOP_DOMAIN` | `*.myshopify.com` alanı | EKSİK (shop.kanus.org’un myshopify adı) |
| `US_LICENSE_SECRET` | Lisans imza sırrı; istemcinin doğrulama sırrıyla aynı | EKSİK — üretilmeli |
| `US_PLANS_JSON` | `{"<product_id>":{"kind":"sub","days":30,"prompts":500,"credits":500}, ...}` | EKSİK — 5 aktif ürünün id’leri ve planları Demir’den |
| `US_STORE_PATH` | Depo dosyası | varsayılan `us-license-store.json` |
| `US_SMTP_HOST/PORT/USER/PASS/FROM` | Hostinger posta (support@kanus.org) | EKSİK |
| `US_SMTP_ENABLE=1` | Gerçek gönderimi açar | kapalı (dry-run) |

## Kredi modeli
- Lisans başına `credits` (plan’dan). İstek kabulünde `Reserve(key, n, ttl)` atomik ayırır; model yanıtı tamamlanınca `Settle(id, true)` düşer; hata/iptal/kesik akışta `Settle(id, false)` iade; ttl dolunca otomatik iade. Böylece yarım yanıt ücretlendirilmez.
- Önerilen: 1 istem = 1 kredi (abonelik) / 1 sn = 1 kredi (kullanım paketi). Fiyat ve kredi oranı hesap sunucusundan okunmalı, istemciye sabitlenmemeli.
- Sonraki adım: gateway (`gateway.go`) model çağrısını sunucu tarafı kredi kapısından geçirmek; istemci sayacı yalnız gösterge olmalı.

## Eksik webhook’lar
`orders/cancelled`, `refunds/create` → lisansı `Revoked`; `app/uninstalled` gerekirse. Tek depo yerine çok örnekli çalışmada PostgreSQL’e taşınmalı (şu an tek süreç, mutex + dosya).

## Kayıt (Demir onayı gerekli — yapılmadı)
Shopify Admin’de `orders/paid` aboneliği (JSON, yukarıdaki HTTPS adres). Önce geliştirme mağazası veya test siparişi ile denenmeli.
