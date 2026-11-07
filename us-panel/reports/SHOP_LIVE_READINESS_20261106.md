# shop.kanus.org canlı mağaza hazırlık denetimi (salt-okuma)

**Tarih:** 2026-11-06 08:09 Europe/Istanbul  
**Mağaza:** Mağazam — `shop.kanus.org` (myshopify: `19einb-qz.myshopify.com`)  
**Bağlayıcı:** user-Shopify MCP · **Mod:** STRICTLY READ-ONLY (mutation yok; kanus-test’e dokunulmadı)

## Özet

| Madde | Sonuç | Kısa bulgu |
|---|---|---|
| Yasal politikalar (refund / privacy / terms / shipping / contact) | **PROBLEM** | Yalnızca **Gizlilik politikası** dolu; iade / hizmet şartları / kargo politikası API’de yok |
| Mağaza sayfaları | **PROBLEM** | `İletişim` sayfası yayınlı ama **gövde boş** |
| Para birimi | **OK** | TRY (presentment yalnız TRY) |
| Vergi (API) | **OK** (kısmi) | `taxesIncluded=true`, `taxShipping=false` — oran/KDV profili API’de yok |
| Ödeme yöntemleri / gateway | **UNKNOWN** | MCP scope yetersiz; Admin’de elle bakılmalı |
| Checkout’ta e-posta zorunluluğu | **UNKNOWN** (login değil) | `loginRequiredAtCheckout=false` (misafir checkout açık); e-posta zorun alanı bu API’de yok |
| Bildirim şablonları | **UNKNOWN** | Admin GraphQL’de bildirim şablonu listesi yok |
| Webhook’lar (bu bağlayıcı) | **PROBLEM** | `webhookSubscriptionsCount=0` — canlı production hook görünmüyor (app-kapsamı uyarısıyla) |

---

## 1. Yasal sayfalar / politikalar

`shop.shopPolicies` dönüşü (tam liste):

| type | Başlık | URL | Gövde |
|---|---|---|---|
| PRIVACY_POLICY | Gizlilik politikası | checkout.shopify.com/.../policies/48729784548.html?locale=tr | **Dolu** (Shopify şablon TR metin, güncelleme 2026-11-02) |
| REFUND_POLICY | — | — | **Yok** (API listesinde yok) |
| TERMS_OF_SERVICE | — | — | **Yok** |
| SHIPPING_POLICY | — | — | **Yok** |
| CONTACT_INFORMATION | — | — | **Yok** |

Mağaza `pages`:

| Sayfa | handle | Yayınlı | Gövde |
|---|---|---|---|
| İletişim | `contact` | evet | **boş** (`body` / `bodySummary` boş) |

**Demir — manuel kontrol:** Shopify Admin → **Ayarlar → Politikalar** (Settings → Policies): İade ve iade politikası, Hizmet şartları, Kargo politikası, İletişim bilgileri oluştur/doldur. Ayrıca **Online Mağaza → Sayfalar → İletişim** gövdesini doldur.

## 2. Ödeme / para birimi / vergi

| Alan | API değeri | Sonuç |
|---|---|---|
| `currencyCode` | TRY | OK |
| `enabledPresentmentCurrencies` | [TRY] | OK |
| `taxesIncluded` | true | OK (fiyatlar KDV dahil varsayımıyla uyumlu) |
| `taxShipping` | false | OK (dijital / kargosuz senaryo için makul) |
| `paymentSettings.supportedDigitalWallets` | [] | Bilgi zayıf — gateway listesi değil |
| `shopifyPaymentsAccount` | Access denied (`read_shopify_payments` / `read_shopify_payments_accounts` scope yok) | UNKNOWN |

**Demir — manuel:** Admin → **Ayarlar → Ödemeler**: hangi sağlayıcılar (Shopify Payments / iyzico / Stripe / manuel vb.), test/canlı mod, 3D Secure. → **Ayarlar → Vergiler**: Türkiye KDV oranı ve ürün vergi ayarları.

## 3. Checkout iletişim / e-posta

| Alan | Değer | Yorum |
|---|---|---|
| `customerAccounts` | OPTIONAL | Hesap isteğe bağlı |
| `customerAccountsV2.customerAccountsVersion` | NEW_CUSTOMER_ACCOUNTS | Yeni müşteri hesapları |
| `loginRequiredAtCheckout` | **false** | Checkout’ta zorunlu giriş yok |
| `loginLinksVisibleOnStorefrontAndCheckout` | true | Login linkleri görünür |
| Hesap URL | `https://shopify.com/85573468388/account` | Shopify hesap portalı |
| Checkout e-posta zorunlu mu? | **API’de yok** | UNKNOWN |

`checkoutAndAccountsConfigurations` → Access denied.

Lisans hook siparişte `email` / `contact_email` / `customer.email` ister; e-posta yoksa lisans maili kuyruğa alınmaz. Bu yüzden checkout’ta e-posta zorunluluğu kritik.

**Demir — manuel:** Admin → **Ayarlar → Checkout** (Checkout): «Müşteri iletişim» / e-posta zorunlu seçenekleri; misafir checkout açık kalsın ama **e-posta zorunlu** olsun.

## 4. Bildirim şablonları

Admin API / bu MCP ile sipariş onayı, kargo, iptal şablon içerikleri **okunamadı** (ilgili query yok / scope yok).

**Demir — manuel:** Admin → **Ayarlar → Bildirimler**: Sipariş onayı (Order confirmation) şablonunun aktif ve Türkçe olduğundan emin ol; lisans bağlama metni eklenmeden önce mevcut metni yedekle.

## 5. Webhook’lar

| Kaynak | Sonuç |
|---|---|
| `webhookSubscriptions` (bu MCP app) | **0 kayıt** |
| `webhookSubscriptionsCount` | **0** |

**Görünürlük sınırı:** Bu sorgu yalnız **bağlı uygulamanın** shop-scoped aboneliklerini döner. Başka custom app / Admin’den elle kayıtlı webhook’lar burada görünmez. Görev notuyla uyumlu: canlıda henüz production hook beklenmiyor; yine de Admin’de doğrulanmalı.

**Demir — manuel:** Admin → **Ayarlar → Bildirimler → Webhooks** ve/veya ilgili **Özel uygulama → API kimlik bilgileri → Webhooks**: `orders/paid`, `orders/cancelled`, `refunds/create` var mı, URL’ler nereye gidiyor?

## 6. Ürün durumu (bağlam; bu turda değiştirilmedi)

| SKU | Product ID (numeric) | Variant ID | status | Fiyat TRY |
|---|---|---|---|---:|
| US-MONTHLY-200 | 15424693010660 | 68094993465572 | ACTIVE | 18900 |
| US-HOUR-1 | 15424693108964 | 68094994317540 | ACTIVE | 1990 |
| US-CREDIT-1 | 15428080632036 | 68119764304100 | **DRAFT** | 149 |
| US-CREDIT-5 | 15428080664804 | 68119764336868 | ACTIVE | 695 |
| US-CREDIT-10 | 15428084170980 | 68119790026980 | ACTIVE | 1290 |
| US-CREDIT-25 | 15432761344228 | 68150471360740 | ACTIVE | 2990 |
| US-CREDIT-50 | 15432761409764 | 68150471426276 | ACTIVE | 5490 |
| US-CREDIT-100 | 15432761442532 | 68150471459044 | ACTIVE | 9900 |

## Sonuç

Canlı vitrin fiyat/SKU tarafı önceki düzeltmelerden sonra büyük ölçüde hazır; **yasal boşluklar (iade/şartlar/kargo + boş iletişim sayfası)**, **ödeme/bildirim bilinmeyenleri** ve **canlı webhook yokluğu** go-live öncesi açık işler. Ayrıntılı kesim planı: `SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md`.
