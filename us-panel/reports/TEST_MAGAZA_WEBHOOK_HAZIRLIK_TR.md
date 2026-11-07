# Test mağazası + orders/paid webhook hazırlığı (canlıya YAZILMADI)

## Canlı mağaza (salt okuma ile alındı)
- Ad: Mağazam, myshopify: 19einb-qz.myshopify.com, birincil alan: shop.kanus.org

## US_PLANS_JSON taslağı (canlı product_id'ler; test mağazasında ID'ler farklı olacak)
| Ürün | product_id | Fiyat | SKU | Önerilen plan |
|---|---|---|---|---|
| Kanus AI – 1 Aylık Erişim / 200 İstem | 15424693010660 | 10000.00 | - | monthly, 200 kredi, 30 gün |
| Kanus AI – 1 Saatlik Kullanım | 15424693108964 | 500.00 | - | hourly, 1 saat |
| Kanus AI — 1 Model İsteği | 15428080632036 | 250.00 | US-CREDIT-1 | 1 kredi |
| Kanus AI — 5 Model İsteği | 15428080664804 | 1400.00 | US-CREDIT-5 | 5 kredi |
| Kanus AI — 10 Model İsteği | 15428084170980 | 2600.00 | US-CREDIT-10 | 10 kredi |
Not: Fiyatlar mağaza para biriminde; istemcideki sabit "200 / 10 USD" ile uyumsuz, Demir doğrulamalı.

## Test mağazası seçenekleri
1. Shopify Partner hesabı + development store (önerilen; ücretsiz, test ödemesi "Bogus Gateway"). Demir'in Partner hesabı açması/giriş yapması gerekir.
2. Connector'daki switch-shop ile test mağazasına geçiş (dev store oluşturulduktan sonra; bağlantıyı canlıdan koparır, yeniden yetki gerekir).
3. Canlı mağazada test modu: REDDEDİLDİ (Demir: canlıya yazma yok).

## Test mağazası bağlanınca gerekenler
- HTTPS URL: örn. https://api-test.kanus.org/webhooks/shopify/orders-paid (ters vekil -> 127.0.0.1:8787, us-license-hook) veya geçici tünel.
- Secret: test uygulamasının client secret'ı (US_SHOPIFY_WEBHOOK_SECRET), env'de; asla koda gömülmez.
- US_LICENSE_SECRET: yeni üretilecek, istemciyle aynı.
- product_id eşlemesi: test mağazasında 5 ürün kopyalanınca yeni ID'lerle yukarıdaki tablo güncellenir.
- Kayıt: webhookSubscriptionCreate(topic: ORDERS_PAID) — yalnızca TEST mağazasında; canlı için ayrıca Demir onayı.
