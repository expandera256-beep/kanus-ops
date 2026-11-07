# Canlı mağaza denetimi — shop.kanus.org (p12)

**Tarih:** 2026-11-06 09:12 Europe/Istanbul  
**Mağaza:** Mağazam · `shop.kanus.org` · myshopify `19einb-qz.myshopify.com`  
**Mod:** STRICTLY READ-ONLY (mutation yok; kanus-test’e dokunulmadı)  
**Önceki özet:** `SHOP_LIVE_READINESS_20261106.md` (08:09) — bu turda politikalar/ürünler/webhook’lar yeniden okundu.

## Özet tablo

| Madde | Sonuç | Öncelik |
|---|---|---|
| Ödeme sağlayıcı durumu | **UNKNOWN** (API scope yok) | P0 |
| Checkout e-posta zorunlu mu? | **UNKNOWN** (`loginRequiredAtCheckout=false`; e-posta alanı API’de yok) | P0 |
| Politikalar / yasal sayfalar | **PROBLEM** — yalnız gizlilik dolu; iade/şartlar/KVKK/mesafeli/ön bilgilendirme canlıda yok | P0 |
| İletişim sayfası | **PROBLEM** — yayınlı, gövde boş | P1 |
| Ürün fiyatları vs katalog | **OK** (FIYATLANDIRMA / LivePlans) | — |
| `deploy/.../plans.json` product_id | **PROBLEM** — dosya **test** ID’leri; canlı ID’lerle eşleşmiyor | P0 (kesim öncesi) |
| US-CREDIT-1 | **OK (bilinçli)** — DRAFT; Demir kararı: mapping’de kalır | — |
| KDV / vergi | **OK** kısmi (`taxesIncluded=true`); oran UNKNOWN | P1 |
| Webhook listesi + hedef URL | **PROBLEM** — bu app: 0; production `api.kanus.org` henüz yok (HOST) | P0 |
| Müşteri bildirim şablonları | **UNKNOWN** | P1 |

---

## 1. Ödeme sağlayıcı

| Kaynak | Bulgu |
|---|---|
| `paymentSettings.supportedDigitalWallets` | `[]` (gateway listesi değil) |
| `shopifyPaymentsAccount` | Access denied — `read_shopify_payments` / `read_shopify_payments_accounts` yok |

**UNKNOWN — Demir:** Shopify Admin → **Ayarlar → Ödemeler** (`/admin/settings/payments`): hangi sağlayıcı(lar) aktif, test/canlı, 3D Secure, TRY tahsilat.

---

## 2. Checkout e-posta

| Alan | Değer |
|---|---|
| `loginRequiredAtCheckout` | false (zorunlu giriş yok) |
| `customerAccounts` | OPTIONAL |
| Checkout «e-posta zorunlu» | API’de yok |

Lisans hook siparişte e-posta ister; yoksa mail kuyruğu oluşmaz.

**UNKNOWN — Demir:** Admin → **Ayarlar → Checkout** (`/admin/settings/checkout`) → Müşteri iletişim: **e-posta zorunlu** olsun (misafir checkout açık kalabilir).

---

## 3. Politikalar ve yasal sayfalar (yeniden okuma)

### Shopify `shopPolicies`

| type | Canlı | Not |
|---|---|---|
| PRIVACY_POLICY | **Var, dolu** (TR Shopify şablon, 2026-11-02) | «Mağazam» jenerik; KVKK aydınlatma değildir |
| REFUND_POLICY | Yok | |
| TERMS_OF_SERVICE | Yok | |
| SHIPPING_POLICY | Yok | Dijital için isteğe bağlı ama iade/cayma ayrı |
| CONTACT_INFORMATION | Yok | |

### Mağaza sayfaları (`pages`)

| Sayfa | handle | Yayın | Gövde |
|---|---|---|---|
| İletişim | contact | evet | **boş** |

Canlıda ayrıca **KVKK aydınlatma**, **mesafeli satış sözleşmesi**, **ön bilgilendirme formu** sayfası/politikası **görülmedi**.

### Yerel ORG taslakları (çoğaltılmadı)

`docs/legal/` altında hukukçu kontrolü gerektiren taslaklar var — canlıya **kopyalanmadı**:

| Dosya | Konu |
|---|---|
| `docs/legal/CAYMA_VE_IADE_POLITIKASI.md` | Cayma/iade; dijital anında ifa istisnası + açık onay |
| `docs/legal/MESAFELI_SATIS_SOZLESMESI.md` | Mesafeli satış; siparişe özgü alanlar |
| `docs/legal/KVKK_AYDINLATMA_METNI.md` | KVKK aydınlatma |
| `docs/legal/KULLANIM_SARTLARI.md` | Kullanım şartları |

### Gap → Demir aksiyonu + metin stratejisi

| Gap | Öncelik | Demir ne yapsın | Metin |
|---|---|---|---|
| İade / cayma politikası | **P0** | Admin → **Ayarlar → Politikalar** → İade politikası; checkout’ta göster | **Referans:** `docs/legal/CAYMA_VE_IADE_POLITIKASI.md` (taslak; hukukçu). Anında teslim dijital içerikte Mesafeli Sözleşmeler Yönetmeliği cayma istisnası için **açık, önceden seçilmemiş onay** + kayıt şart. |
| Hizmet şartları | **P0** | Politikalar → Hizmet şartları veya Sayfa | **Referans:** `docs/legal/KULLANIM_SARTLARI.md` |
| KVKK aydınlatma | **P0** | Yeni Sayfa (örn. `/pages/kvkk`) + footer link; gizlilik şablonundan ayrı | **Referans:** `docs/legal/KVKK_AYDINLATMA_METNI.md` — `[DOLDURULACAK]` / `[Şirket unvanı]` / MERSİS / adres |
| Mesafeli satış sözleşmesi | **P0** | Checkout ek sözleşme / Sayfa; sipariş başına alanlar | **Referans:** `docs/legal/MESAFELI_SATIS_SOZLESMESI.md` |
| Ön bilgilendirme formu | **P0** | Ödeme öncesi zorunlu bilgilendirme UI (Shopify checkout ekleri / uygulama) | Siparişe özgü; yerel dosya tek başına yeterli değil. İskelet yer tutucular: `[Şirket unvanı]`, `[MERSİS]`, `[adres]`, `[ürün/SKU]`, `[toplam bedel KDV dahil]`, `[ifa süresi]`, `[cayma/istisna onayı]` |
| İletişim sayfası boş | **P1** | Online Mağaza → Sayfalar → İletişim | Kısa iskelet (taslak): «[Şirket unvanı], [adres], [telefon], support@kanus.org / [KEP].» Hukukçu + Demir doldurur. |
| Gizlilik (mevcut) | **P2** | İsteğe bağlı: «Mağazam» → gerçek unvan; KVKK sayfasına çapraz link | Mevcut şablon yeterli sayılmaz; KVKK ayrı |

**Uyarı:** Aşağıdaki kısa iskeletler yalnızca boşluk doldurma yardımıdır; **hukuki tavsiye değildir**, yayın öncesi hukukçu şart.

<details><summary>Ön bilgilendirme — çok kısa iskelet (taslak)</summary>

Satıcı: [Şirket unvanı], [adres], [MERSİS/VKN], [telefon], [e-posta].  
Konu: dijital US Panel kullanım hakkı — [SKU / istem veya süre].  
Bedel: [tutar] TRY (KDV dahil). Fiziksel kargo yok.  
İfa: ödeme sonrası lisans anahtarı e-posta / hesap bağlama.  
Cayma: anında ifa edilen dijital içerik/hizmette Yönetmelik istisnası, **yalnızca bilgilendirme + açık onay** ile; ayıp/ifa edilmeme hakları saklı.  
Ayrıntı: mesafeli sözleşme ve cayma politikası linkleri.

</details>

---

## 4. Ürünler, fiyatlar, plans.json

Kaynak fiyat/kredi: `licensesrv/pricing.go` Catalog + `FIYATLANDIRMA_TR.md`. Canlı varyantlar 09:12’de yeniden okundu.

| SKU | Canlı product_id | Canlı fiyat | Beklenen TRY | credits/saniye | status | Fiyat eşleşmesi |
|---|---|---:|---:|---|---|---|
| US-CREDIT-1 | 15428080632036 | 149 | 149 | 1 kredi / 90g | **DRAFT** | OK |
| US-CREDIT-5 | 15428080664804 | 695 | 695 | 5 / 90g | ACTIVE | OK |
| US-CREDIT-10 | 15428084170980 | 1290 | 1290 | 10 / 90g | ACTIVE | OK |
| US-CREDIT-25 | 15432761344228 | 2990 | 2990 | 25 / 180g | ACTIVE | OK |
| US-CREDIT-50 | 15432761409764 | 5490 | 5490 | 50 / 180g | ACTIVE | OK |
| US-CREDIT-100 | 15432761442532 | 9900 | 9900 | 100 / 365g | ACTIVE | OK |
| US-MONTHLY-200 | 15424693010660 | 18900 | 18900 | 200 / 30g | ACTIVE | OK |
| US-HOUR-1 | 15424693108964 | 1990 | 1990 | 3600 sn | ACTIVE | OK |

**US-CREDIT-1:** Demir kararı — satışta değil (DRAFT); **US_PLANS / kredi eşlemesinde kalır** (eski/manuel sipariş).

### `deploy/us-license-hook/plans.json` uyarısı

Bu dosyadaki ID’ler (**15473713…**) canlı ürün ID’leri **değil** (kanus-test / başka ortam). Canlı kesimde `US_PLANS_FILE` **canlı product_id** ile yazılmalı (`SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md` JSON). Aksi halde paid webhook lisans üretmez.

---

## 5. Vergi / KDV

| Alan | Değer |
|---|---|
| `taxesIncluded` | true (vitrin fiyatı KDV dahil varsayımıyla uyumlu) |
| `taxShipping` | false |
| KDV oranı / vergi koleksiyonu | API’de yok → **UNKNOWN** |

**Demir:** Admin → **Ayarlar → Vergiler** (`/admin/settings/taxes`) — TR %20 (veya geçerli oran) ve dijital ürün vergi ayarı.

---

## 6. Webhook’lar

| Kaynak | Sonuç |
|---|---|
| Bu MCP app `webhookSubscriptions` | **0** (hedef URL yok) |
| Görünürlük | Yalnız bu app’in shop-scoped kayıtları; başka app görünmez |
| Production hook | **`api.kanus.org` henüz yok** — HOST hazırlıyor (KAN.US notu) |

**Demir (HOST hazır olunca):** Admin → **Ayarlar → Bildirimler → Webhooks** ve/veya özel uygulama webhook’ları; hedefler:  
`https://api.kanus.org/webhooks/shopify/orders-paid|orders-cancelled|refunds-create`  
Plan: `SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md` — **uygulanmadı**.

---

## 7. Müşteri bildirim e-posta şablonları

API ile şablon içerikleri **okunamadı**.

**UNKNOWN — Demir:** Admin → **Ayarlar → Bildirimler** (`/admin/settings/notifications`): Sipariş onayı TR mi, aktif mi? Yedek al; lisans bağlama metni (`SHOP_LICENSE_LINK_COPY_TR.md`) onayla ekle.

---

## Yasal metinleri canlıya alma planı (PLAN ONLY — Demir onayı)

1. Hukukçu: `docs/legal/*` taslaklarını `[DOLDURULACAK]` / unvan / MERSİS / adres ile doldurur.  
2. Demir onaylar.  
3. Shopify: Politikalar + Sayfalar (KVKK, mesafeli, ön bilgilendirme linkleri) + footer.  
4. Checkout: cayma istisnası için **açık onay** kutusu / metin sürümü kaydı (dijital anında teslim).  
5. İletişim sayfasını doldur.  
6. Bildirim şablonuna lisans bağlama paragrafı.  
7. Bu adımlarda SHOP botu **yazmaz** ta ki Demir açıkça «canlıya yapıştır» desin.

**Uygulanmadı.**

## Marka ve vitrin bulguları (ORG)

**Kaynak:** `/workspace/us-panel/docs/BRAND_A_COPY_20261106.md` (ORG; yayınlanmamış taslak).  
**Doğrulama:** 2026-11-06 09:17 Europe/Istanbul — salt-okuma: Shopify MCP + `https://shop.kanus.org/` HTML + `products.json`.  
**Kapsam:** Yalnız **öneri**; Shopify mutation / tema yazımı / sayfa oluşturma **yapılmadı**. Her madde: **Demir kararı bekliyor**.

### 1) Ana sayfa hero fiyat / istem ≠ ürün kayıtları — **P0** · Demir kararı bekliyor

| | Hero (canlı HTML alıntı) | Ürün kaydı (doğrulandı) |
|---|---|---|
| Aylık paket | `Kanus AI – 1 Aylık Erişim / **500 İstem**` · `**10.000 TL** · … **500** istek` | Başlık `… / **200 İstem**` · fiyat **18.900,00 TRY** · SKU `US-MONTHLY-200` · id `15424693010660` |
| Saatlik paket | `Kanus AI – 1 Saatlik Kullanım` · `**500 TL** · …` | Fiyat **1.990,00 TRY** · SKU `US-HOUR-1` · id `15424693108964` |
| Hero H2 | `Kanus AI erişim paketleri` | (ürünlerle uyumlu; korunabilir) |

**Nerede duruyor:** Tema ana sayfa şablonu — HTML’de `shopify-section-template--28319750783204__section_*` altında rich-text / özel metin bloğu (ürün kartı değil; sabit kopya). Tema id (storefront): `189837082852`.

**Önerilen düzeltme (Admin):** Online Mağaza → Temalar → **Özelleştir** → Ana sayfa → ilgili metin/rich-text bölümü. Metni ürün kayıtlarıyla hizala, örn.:
- Aylık: `Kanus AI – 1 Aylık Erişim / 200 İstem` — `18.900 TL · Bir ay boyunca toplam 200 istem.`
- Saatlik: `Kanus AI – 1 Saatlik Kullanım` — `1.990 TL · Bir saat işlem veya açık panel.`
- Alt not (`tek seferlik / otomatik yenileme yok`) korunabilir.

---

### 2) Handle «500» ama başlık «200» — **P1** · Demir kararı bekliyor

| Alan | Değer |
|---|---|
| Product ID | `gid://shopify/Product/15424693010660` (numeric `15424693010660`) |
| Başlık | `Kanus AI – 1 Aylık Erişim / 200 İstem` |
| Handle (canlı) | `kanus-ai-1-aylik-erisim-500-istem` (URL’de `500-i̇stem`) |
| Storefront URL | `https://shop.kanus.org/products/kanus-ai-1-aylik-erisim-500-i%CC%87stem` |

**Önerilen handle:** `kanus-ai-1-aylik-erisim-200-istem`  
**Admin:** Ürünler → ilgili ürün → Arama motoru listesi / handle (URL ve handle).  
**Yönlendirme:** Eski handle’a gelen linkler kırılmasın diye Shopify Admin → Online Mağaza → **Gezinme / URL yönlendirmeleri** (veya Ürün kaydetme sırasında otomatik yönlendirme seçeneği açıksa kullan):  
`/products/kanus-ai-1-aylik-erisim-500-istem` → `/products/kanus-ai-1-aylik-erisim-200-istem` (Unicode noktalı İ varyantını da kontrol et).

---

### 3) Mağaza adı hâlâ «Mağazam» — **P1** · Demir kararı bekliyor

| Kaynak | Değer |
|---|---|
| `get-shop-info` / storefront `<title>` / footer | `Mağazam` |

**Önerilen ad (ORG / brand):** `Kanus` veya `Kanus Mağaza` (NAMING: müşteri yüzünde Kanus; KAN.US dahili).  
**Admin yolu:** **Ayarlar → Mağaza ayrıntıları** (`/admin/settings/general`) → Mağaza adı.  
Footer «© 2026 Mağazam…» tema/footer metninden de güncellenebilir (Özelleştir → Footer) — ayrı Demir onayı.

---

### 4) Hakkımızda sayfası yok — **P1** · Demir kararı bekliyor

| Kontrol | Sonuç |
|---|---|
| Admin `pages` | Yalnız `contact` (İletişim), gövde boş |
| `https://shop.kanus.org/pages/hakkimizda` | **404** |
| `https://shop.kanus.org/pages/about-us` | **404** |

**Hedef:** Online Mağaza → **Sayfalar** → yeni sayfa, örn. handle `hakkimizda`, başlık `Hakkımızda`; menüye ekle.

**Taslak gövde (TR, yayın değil; menşe sloganı yok — karar A):**

> Kanus, **US Panel** ile yapay zekâ destekli yazılım geliştirmeye odaklanır. Türkçe arayüz ve geliştirici odaklı anlatım. Krediler ve lisans paketleri: https://shop.kanus.org. Destek: support@kanus.org.  
> Sağlayıcı unvan / adres / MERSİS: `<DOLDUR>`.  
> Logo: unicursal heksagram (marka geometrisi; ideolojik vaat değildir).

«Türkiye merkezli / yerli ve milli» **eklenmez** (ORG A kararı).

---

### Özet öncelik

| # | Madde | Öncelik | Durum |
|---|---|---|---|
| 1 | Hero 500/10.000/500 TL → 200/18.900 / 1.990 | **P0** | Demir kararı bekliyor |
| 2 | Handle `…500-istem` → `…200-istem` + redirect | **P1** | Demir kararı bekliyor |
| 3 | Mağaza adı `Mağazam` → `Kanus` / `Kanus Mağaza` | **P1** | Demir kararı bekliyor |
| 4 | Hakkımızda sayfası oluştur | **P1** | Demir kararı bekliyor |

## Uygulanan vitrin düzeltmeleri

**Zaman:** 2026-11-06 09:23 Europe/Istanbul  
**Mağaza:** shop.kanus.org (canlı) — yalnızca Demir onaylı 3 madde. Hakkımızda / fiyatlar / diğer ürünler / diğer tema içeriği dokunulmadı.

### PASS/FAIL tablosu

| # | Madde | Sonuç | Kanıt |
|---|---|---|---|
| 1 | Ana sayfa hero fiyat/istem metni | **FAIL** | Yedek: `/workspace/us-panel/backups/shop_live_theme_hero_20261106_092220.json`. `themeFilesUpsert` MAIN tema (Horizon id 189837082852) için connector güvenlik politikası reddetti: "This mutation targets the live (published) theme. Theme file writes against the live storefront are blocked. Duplicate the theme in Shopify admin, edit the draft, and a merchant can publish it manually." (blocked=true, category=live_theme). Canlı HTML hâlâ: 500 İstem / 10.000 TL / 500 TL. Hedef (uygulanamadı): 200 İstem / 18.900 TL / 1.990 TL. Admin: Online Mağaza → Temalar → Özelleştir → section_4baPEj / text_T3U9q4. |
| 2 | Handle → kanus-ai-1-aylik-erisim-200-istem + 301 | **PASS** | Eski handle: `kanus-ai-1-aylik-erisim-500-i̇stem` (i+U+0307). productUpdate + redirectNewHandle=true → yeni handle `kanus-ai-1-aylik-erisim-200-istem`. Redirects: `/products/kanus-ai-1-aylik-erisim-500-i%cc%87stem` → `/products/kanus-ai-1-aylik-erisim-200-istem` (auto id 610001715428); ASCII `/products/kanus-ai-1-aylik-erisim-500-istem` → aynı (urlRedirectCreate id 610001748196). curl: unicode eski 301; ASCII eski 301; yeni URL 200. |
| 3 | Mağaza adı → Kanus | **FAIL** | Connector’da shopUpdate / mağaza adı yazma aracı yok (schema search boş; dedicated tool yok). Workaround yok. get-shop-info read-back: hâlâ **Mağazam**. Admin: Ayarlar → Mağaza ayrıntıları (`/admin/settings/general`). |

### Notlar
- Hero için düzeltilmiş templates/index.json yedekten üretildi; yalnızca text_T3U9q4 fiyat/istem string’i. Featured-product eski handle referansına dokunulmadı (Demir kısıtı); handle değişince o blok boşa düşebilir — ayrı onay gerekir.
- Hakkımızda sayfası bilinçli uygulanmadı.


## Ana sayfa ürün vurgusu teşhisi

**Zaman:** 2026-11-06 09:40 Europe/Istanbul  
**Kapsam:** Salt-okuma (shop.kanus.org / Horizon MAIN). Mutation yok.  
**Semptom:** İki «Ürün vurgusu / Featured product» bölümünde yalnızca başlık (erişilebilirlik: «Öne çıkan ürün»); ürün kartı görünmüyor.

### Ürün / yayın / stok (kanıt)

| Ürün | ID | Handle | status | Online Mağaza | publishedAt (UTC) | availableForSale | tracked | qty | inventoryPolicy |
|---|---|---|---|---|---|---|---|---|---|
| 1 Saatlik | 15424693108964 | `kanus-ai-1-saatlik-kullanim` | ACTIVE | isPublished=true (Publication 230294061284) | 2026-11-03T19:04:14Z | **true** | false | 0 | DENY |
| 1 Aylık / 200 İstem | 15424693010660 | `kanus-ai-1-aylik-erisim-200-istem` | ACTIVE | isPublished=true | 2026-11-03T19:04:13Z | **true** | false | 0 | DENY |

Storefront: `/products/<handle>.js` → `available: true` (her ikisi); `.json` 200; ürün sayfaları erişilebilir. Stok takibi kapalı → qty=0 + DENY satışa engel değil.

### Tema `templates/index.json` (MAIN Horizon 189837082852)

| Section | settings.product (handle) | Eski `…500-i̇stem`? | Kritik blok durumu |
|---|---|---|---|
| `featured_product_bP9LGi` (saatlik) | `kanus-ai-1-saatlik-kullanim` | yok | **`blocks.featured-product.disabled: true`** (tüm ürün kartı içeriği kapalı). Media açık ama üründe görsel yok → placeholder SVG. |
| `featured_product_39embW` (aylık) | `kanus-ai-1-aylik-erisim-200-istem` (güncel) | **yok** (eski 500 handle kalmamış) | Media + title + price + gallery hepsi **`disabled: true`**. `_featured-product` açık → boş `<product-card>` kabuğu (doğru product id 15424693010660) ama başlık/fiyat/galeri yok. |

Liquid (`sections/featured-product.liquid`): `assign product = section.settings.product`; `product == blank` / `product.available` / ödeme sağlayıcısı **kontrolü yok**. Kart `content_for` ile `_featured-product` + media bloklarından geliyor; **disabled bloklar render edilmez**.  
`snippets/product-card.liquid`: yalnızca `settings.quick_add and product.available` quick-add için; kartı gizlemez.  
Ödeme: featured liquid ödeme koşulu içermiyor; eksik ödeme kartı gizleme kök nedeni değil (shopifyPaymentsAccount scope yok; homepage’te PaymentButton script var).

### Canlı HTML

- Saatlik section: `product-card` **yok**; yalnızca `visually-hidden` «Öne çıkan ürün» + media placeholder.
- Aylık section: `product-card` var, `data-product-id="15424693010660"`, aria-label doğru; içerikte fiyat/başlık metni yok (`has price False`).
- Ürünlerde `featured_image` / media sayısı = **0** (görsel yok — ikincil görünürlük sorunu).

### Kök neden

**Tema editöründe Featured product bloklarının (ve aylıkta title/price/gallery/media’nın) `disabled: true` olması.** Ürünler ACTIVE, Online Store’da yayınlı, `availableForSale=true`, handle referansları güncel (aylık zaten `…200-istem`). Handle kırığı / yayın / stok / ödeme sağlayıcısı **değil**.

### Önerilen düzeltme (Admin — connector live tema yazamaz)

1. **Online Mağaza → Temalar → Özelleştir → Ana sayfa**  
2. Saatlik «Ürün vurgusu» (`featured_product_bP9LGi`): ürün seçimini doğrula (`Kanus AI – 1 Saatlik…`); **Product** bloğunu yeniden etkinleştir (disabled kaldır).  
3. Aylık «Ürün vurgusu» (`featured_product_39embW`): ürün = `…200 İstem`; **Title / Price / Gallery / Media** bloklarını etkinleştir.  
4. İsteğe bağlı: her iki ürüne görsel ekle (şu an media=0 → galeri boş kalır).  
5. Kaydet / yayınla. Connector ile `themeFilesUpsert` MAIN’de engelli — tema düzeltmesi **yalnızca Theme editor**.

**Durum:** Demir kararı bekliyor (mutation yok).
