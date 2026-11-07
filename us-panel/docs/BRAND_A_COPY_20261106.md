# Marka kararı A — metin uygulama taslağı

**TASLAK — yayınlanmadı.** Canlı site, Shopify, DNS veya Hostinger’a yazılmadı; yalnızca yerel öneri.

- **Tarih:** 06.11.2026 TSİ (Europe/Istanbul)
- **Karar A (Demir, 06.11.2026 08:13 TSİ, KAN.US aktarımı):** Thelema (unicursal heksagram) ana işaret kalır; “yerli ve milli” söylemi metin/ton düzeyinde yumuşatılır; doğrulanmamış iddialar (ör. “Türkiye merkezli”) eklenmez; “homegrown” tarzı sloganlar nötrleştirilir.
- **Referanslar:** `brand/docs/DECISION_THELEMA_VS_YERLI.md`, `TONE.md`, `NAMING.md`, `brand/shopify/PRODUCT_TEMPLATES_TR.md` / `_EN.md`, `prompt/NIHAI_PROMPT_20261106.txt` §7.
- **Tarama yöntemi:** HTTP GET (curl) + WebFetch; kanus.org SPA metinleri `/assets/index-uscommerce.js?v=20261104-usticari` içinden; mağaza `products.json` + Shopify connector **salt okuma** (`get-shop-info`, `search_products`, `get-product`).

---

## 1) Taranan kaynaklar ve erişim durumu

| # | URL / kaynak | Durum | Not |
|---|--------------|-------|-----|
| 1 | https://kanus.org/ | OK (200) | SPA kabuğu; gövde `#root`. Meta + JS i18n okundu. |
| 2 | https://kanus.org/about | OK (200) | Aynı SPA kabuğu; ayrı “Hakkımızda” içeriği **yok** (route muhtemelen istemci tarafı; sunucu HTML’de ayrı sayfa metni yok). |
| 3 | https://kanus.org/hakkimizda | OK (200) | Aynı SPA kabuğu; ayrı hakkımızda metni **yok**. |
| 4 | https://kanus.org/assets/index-uscommerce.js?v=20261104-usticari | OK | TR/EN/DE/RU i18n dizeleri çıkarıldı. |
| 5 | https://shop.kanus.org/ | OK (200) | Ana sayfa hero + tema metinleri. |
| 6 | https://shop.kanus.org/products.json | OK | 7 yayında ürün (storefront). |
| 7 | https://shop.kanus.org/pages.json | OK | Tek sayfa: `contact` / İletişim; `body_html` boş (`null`). |
| 8 | https://shop.kanus.org/pages/about-us | 404 | Hakkımızda sayfası yok. |
| 9 | https://shop.kanus.org/pages/hakkimizda | 404 | Hakkımızda sayfası yok. |
| 10 | https://shop.kanus.org/pages/contact | OK (200) | Başlık “İletişim”; gövde içeriği boş. |
| 11 | Shopify Admin (connector, salt okuma) | OK | Mağaza adı “Mağazam”, domain shop.kanus.org, TRY, TR. 8 ürün (7 ACTIVE + 1 DRAFT). |

**Canlı taramada “yerli / milli / millî / homegrown / Türkiye merkezli / Turkey-based” geçmedi** (kanus.org HTML+JS ve shop HTML/ürün gövdeleri). A kararı bu yüzeylerde çoğunlukla **önleyici hizalama + isim/ton/iddia tutarlılığı** olarak uygulanır; silinecek canlı slogan yok.

---

## 2) Önce / sonra önerileri

Her satırda “Önce” yalnızca okunan metinden; uydurma yok. “Sonra” yayınlanmamış öneridir.

### 2.1 kanus.org — belge meta (HTML `<head>`)

| Alan | Kaynak | Önce (alıntı) | Sonra (öneri) | Gerekçe |
|------|--------|---------------|---------------|---------|
| `<title>` | https://kanus.org/ | `KAN.US) — Space for your next idea` | `Kanus — Fikirlerini yazılıma dönüştür` (EN alternatif: `Kanus — Turn your ideas into software`) | Yazım hatası `KAN.US)`; müşteri yüzünde **Kanus** (NAMING); KAN.US dahili etiket. A: menşe sloganı yok. |
| meta description | https://kanus.org/ | `Discover AI applications, explore ideas in your personal workspace, and keep up with the latest marketplace updates.` | TR: `US Panel ile yapay zekâ destekli yazılım geliştirme. Kredini shop.kanus.org’dan al.` / EN: `AI-assisted software development with US Panel. Get credits at shop.kanus.org.` | Ürün odaklı, iddiasız; “marketplace updates” boş katalogla uyumsuz. |

### 2.2 kanus.org — ana sayfa (TR i18n, JS)

Kaynak: `index-uscommerce.js` TR bloğu.

| Anahtar | Önce | Sonra | Gerekçe |
|---------|------|-------|---------|
| homeEyebrow | `Sıradaki şey burada başlar` | `Türkçe arayüz. Net geliştirme.` | A: yerli/milli yumuşak konum; dil/ürün vurgusu, millî slogan değil. |
| homeTitle | `Sıradaki büyük fikrin için bir alan.` | `Fikirlerini yazılıma dönüştür.` | NAMING/TONE nötr slogan. |
| homeIntro | `Yapay zeka uygulamalarını keşfet. Bir düşünceyi yokla. Sıradaki adımını, her seferinde bir istemle at.` | `US Panel’de istemlerini koda ve ürüne çevir. Mağaza: shop.kanus.org.` | Keşif şiirinden ürün/kullanıma. |
| featureTitle | `Daha az sürtünme. Daha çok olasılık.` | `Az sürtünme. Çalışan çıktı.` | Daha somut, abartısız. |
| featureBody | `Boş bir tuval. Senin sözlerin. Bir konuşma başlat.` | `Boş bir tuval. Senin sözlerin. US Panel’de konuşmaya başla.` | Marka adı bağlama. |
| panel (nav) | `Yapay zeka paneli` | `US Panel` | NAMING: ürün adı sabit. |
| openPanel | `Yapay zeka panelini aç` | `US Panel’i aç` | Aynı. |
| appsIntro | `US.exe Windows uygulamasıdır. Aylık abonelik 200 USD, 500 istem. Kullanım 10 USD, 1 saat işlem veya açık panel.` | `US Panel Windows uygulamasıdır. Güncel paket ve fiyatlar shop.kanus.org’dadır; sitedeki örnek rakamlar teklif değildir.` | İsim hizası; canlı mağaza TRY/SKU ile çelişen sabit USD iddiasını yumuşat (doğrulama açık noktası). |
| usIntro | `US.exe indirilir. Lisans shop.kanus.org adresinden alınır. Ödenmiş sipariş anahtarı e-posta ile gönderir. Aylık abonelik 200 USD, 500 istem. Kullanım 10 USD, 1 saat işlem veya açık panel.` | `US Panel’i indir. Lisansı shop.kanus.org’dan al. Ödenmiş siparişte anahtar e-posta ile gelir. Paket içeriği ve fiyat mağaza sayfasındadır.` | Aynı; çelişen fiyatı kaldırma önerisi. |
| usCard | `Ticari erişim. support@kanus.org adresine yazın. 200 USD ile 500 istem, ya da 10 USD ile 1 saat.` | `Ticari erişim için support@kanus.org. Güncel paketler: shop.kanus.org.` | Destek kanalı korunur; doğrulanmamış sabit fiyat iddiası kalkar. |
| usDownload | `US.exe indir` | `US Panel indir` | NAMING. |
| contactIntro | `Pazaryeri veya çalışma alanın hakkında bir sorun mu var? support@kanus.org için bir not bırak.` | `US Panel veya siparişin hakkında soru mu var? support@kanus.org’a not bırak.` | Boş pazaryeri yerine ürün/destek. |
| footerNote | `Biraz merak uzun bir yol alır.` | (koru) veya `Net soru, net yanıt.` | Ton uyumlu; zorunlu değişiklik değil. |

### 2.3 kanus.org — ana sayfa (EN i18n, JS)

| Anahtar | Önce | Sonra | Gerekçe |
|---------|------|-------|---------|
| homeTitle | `Space for your next big idea.` | `Turn your ideas into software.` | A/NAMING EN karşılığı (ayrıca resmî onay varsayılmaz). |
| homeIntro | `Discover AI applications. Explore a thought. Make your next move, one prompt at a time.` | `Build with US Panel. Get credits at shop.kanus.org.` | Ürün odaklı. |
| appsIntro / usIntro / usCard | USD 200 / 500 prompts; USD 10 / 1 hour… | `US Panel for Windows. Current plans and prices are on shop.kanus.org.` | Canlı mağaza ile çelişkiyi azalt. |
| usDownload | `Download US.exe` | `Download US Panel` | NAMING. |

### 2.4 kanus.org — hakkımızda

| Kaynak | Önce | Sonra | Gerekçe |
|--------|------|-------|---------|
| /about, /hakkimizda | **Erişilemedi (ayrı içerik yok):** HTTP 200 ama sunucu HTML yalnızca SPA kabuğu; ayrı hakkımızda metni bulunamadı. JS’te de `about` sayfa gövdesi yok. | Taslak About (yalnız öneri, yayın değil): «Kanus, US Panel ile yapay zekâ destekli yazılım geliştirmeye odaklanır. Türkçe arayüz ve geliştirici odaklı anlatım. Destek: support@kanus.org. Logo: unicursal heksagram (marka geometrisi).» — **“Türkiye merkezli / yerli ve milli” eklenmez.** | A + ABOUT_SPLASH_TR.md ile uyum. |

### 2.5 shop.kanus.org — mağaza kimliği ve ana sayfa

| Alan | Kaynak | Önce | Sonra | Gerekçe |
|------|--------|------|-------|---------|
| Mağaza adı | Shopify `get-shop-info` + HTML title | `Mağazam` | `Kanus` veya `Kanus Mağaza` | Varsayılan Shopify adı; marka kimliği zayıf. |
| Announcement / welcome | https://shop.kanus.org/ | `Welcome to our store` | `Kanus AI erişim paketleri` veya `US Panel kredileri` | TR mağaza; İngilizce karşılama gereksiz. |
| Hero H2 | https://shop.kanus.org/ | `Kanus AI erişim paketleri` | (koru) | Zaten nötr/ürün odaklı; A ile uyumlu. |
| Hero paket 1 | https://shop.kanus.org/ | `Kanus AI – 1 Aylık Erişim / 500 İstem` + `10.000 TL · Bir ay boyunca toplam 500 istek.` | **Ürün kaydıyla hizala:** canlı ürün başlığı `Kanus AI – 1 Aylık Erişim / 200 İstem`, fiyat `18.900,00 TL`, açıklama «200 istem». Öneri metin: ürün kartındaki gerçek değerler (500/10.000 **yazılmasın**). | Hero ile ürün kaydı çelişiyor; A iddia disiplini. |
| Hero paket 2 | https://shop.kanus.org/ | `Kanus AI – 1 Saatlik Kullanım` + `500 TL · Bir saatlik işleme veya panelin açık kalma süresi.` | Ürün kaydı: `1.990,00 TL`. Hero’yu **1.990 TL** (veya “fiyat ürün sayfasında”) yap. | Aynı çelişki. |
| Hero not | https://shop.kanus.org/ | `Her iki paket de tek seferlik satın alımdır; otomatik yenileme yoktur.` | (koru) | Dürüst, A ile uyumlu. |
| Nasıl işler? | https://shop.kanus.org/ | `İhtiyacınıza uygun tek seferlik paketi seçip siparişinizi oluşturun. Satın alma otomatik yenilenmez. API anahtarınız, siparişinizden sonra en geç 10 dakika içinde e-posta ile gönderilir.` | Küçük ton: `İhtiyacına uygun tek seferlik paketi seç. Otomatik yenileme yok. Anahtar, sipariş sonrası e-posta ile iletilir (hedef: 10 dk; gecikmede support@kanus.org).` + isteğe: `Anahtarlar sipariş sonrası manuel iletilebilir.` (mevcut alt satırla tutarlı) | “en geç 10 dk” vaadini abartmadan koru. |
| Alt satır | https://shop.kanus.org/ | `Anahtarlar sipariş sonrası manuel olarak iletilir.` | (koru) | Dürüst. |
| Bülten | https://shop.kanus.org/ | `Join our email list` / `Get exclusive deals and early access to new products.` | `Yeniliklerden haberdar ol` / `Yeni paketler ve duyurular için e-posta bırak.` | TR; “exclusive deals” abartısı yok. |
| Footer | https://shop.kanus.org/ | `© 2026 Mağazam, Shopify tarafından desteklenmektedir` | `© 2026 Kanus` (Shopify credit temada kalabilir) | Marka adı. |

### 2.6 shop.kanus.org — hakkımızda / iletişim

| Kaynak | Önce | Sonra | Gerekçe |
|--------|------|-------|---------|
| /pages/about-us, /pages/hakkimizda | **Erişilemedi:** HTTP 404 (`Page not found`). | Yeni sayfa taslağı (yayın değil): başlık `Hakkımızda`; gövde ABOUT_SPLASH_TR ile aynı çerçeve; menşe sloganı yok. | A. |
| /pages/contact | Başlık `İletişim`; `body_html` = boş | Kısa gövde: `Destek: support@kanus.org. Sipariş ve lisans soruları için bu adresi kullan.` | Boş sayfayı doldurma önerisi. |

### 2.7 Ürün açıklamaları (storefront + Admin okuma)

Ortak kalıp (ACTIVE kredi ürünleri 5/10/25/50/100):

| Ürün (ör.) | Kaynak | Önce (alıntı) | Sonra | Gerekçe |
|------------|--------|---------------|-------|---------|
| Kanus AI — 100 Model İsteği | https://shop.kanus.org/products.json + Admin | `100 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 365 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.` + `İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.` | Teknik gövde **korunabilir**. İsteğe ek satır (menşe yok): `US Panel (Kanus.org) için kullanılır. Fikirlerini yazılıma dönüştür.` | Zaten yerli/milli yok; A ile uyumlu ürün odağı eklenir. Aynı kalıp 5/10/25/50 için süre günleriyle. |
| Kanus AI — 1 Model İsteği | Admin DRAFT | Aynı kalıp, 90 gün, 149 TL | Aynı ek satır; DRAFT kalsın (yayın yok). | — |
| Kanus AI – 1 Saatlik Kullanım | products.json | `kanus.org üzerinden erişilebilen yapay zekâ paneli için bir saatlik kullanım paketi.` + maddeler (1 saat; tek seferlik; 10 dk e-posta; dijital) | `US Panel (kanus.org) için bir saatlik kullanım paketi.` + maddeler korunur. | “yapay zekâ paneli” → US Panel. |
| Kanus AI – 1 Aylık Erişim / 200 İstem | products.json | `… bir aylık API erişim paketi.` + `Bir ay boyunca toplam 200 istem hakkı.` … | Başlıkta `/ 200 İstem` korunur; **handle** `…500-istem` ve ana sayfa “500” metni ayrı düzeltme konusu (açık nokta). Gövde: `US Panel için bir aylık API erişim paketi.` | İsim + tutarlılık. |

**Not:** Ürün gövdelerinde yerli/milli/homegrown yok; A uygulaması çoğunlukla ek slogan ve US Panel adlandırma.

### 2.8 Yerel taslaklarla ilişki (canlı değil)

Zaten A’ya çekilmiş yerel metinler (karşılaştırma; canlıya yazılmadı): `PRODUCT_TEMPLATES_TR.md` sloganı «Fikirlerini yazılıma dönüştür.»; `ABOUT_SPLASH_TR.md` «Türkiye merkezli» yok. Bu dosya onları canlı yüzeylere taşıma **önerisi**dir, yayın emri değildir.

---

## 3) Ses ve ton rehberi (A uygulaması)

### Yapılacaklar
- Türkçe arayüz ve hata mesajı; geliştiriciye saygılı, sakin, net.
- Ürün adı: **US Panel**; marka: **Kanus** / **Kanus.org**; mağaza: **shop.kanus.org**.
- Ana işaret Thelema unicursal heksagram — metinde “geometrik marka işareti”; dini/ideolojik vaat yok.
- Yumuşak konum: “Türkçe anlatım”, “Türkiye’de geliştiriciler için” **yalnızca ihtiyaç ve doğrulanmış bağlamda**, abartısız.
- Fiyat, süre, istem sayısı: yalnızca mağaza kaydı veya `[DOLDURULACAK]`.
- Slogan (TR): **Fikirlerini yazılıma dönüştür.** EN taslak: **Turn your ideas into software.**

### Kaçınılacaklar
- Slogan olarak **“yerli ve milli”**, **“homegrown”**, **“Türkiye’nin 1 numarası”**, doğrulanmamış **“Türkiye merkezli”**.
- Model/sunucu/menşe iddiası; sahte müşteri sayısı; rakip kötüleme.
- Thelema’yı ideolojik pazarlama sloganına çevirmek.
- Müşteri yüzünde **KAN.US**’yi ürün adı gibi öne çıkarmak.
- Site hero’sunda mağaza kaydıyla çelişen fiyat/istem.

### Örnek ifadeler

| Bağlam | Tercih | Kaçın |
|--------|--------|-------|
| Konum | Türkçe arayüz ve dokümantasyon | Yerli ve milli yazılım gücü |
| Ürün | Yapay zekâ destekli yazılım geliştirme paneli | Filtresiz = güvenlik yok ima |
| CTA | Kredini al, fikirlerini koda dönüştür | Homegrown AI for Turkey |
| About | Destek: support@kanus.org | Türkiye merkezli şirketiz (doğrulanmadan) |

### Kelime listesi

| Tercih | Dikkatli / sınırlı | Yasak (doğrulama yoksa) |
|--------|--------------------|-------------------------|
| Kanus, US Panel, kredi, istem, lisans, dijital erişim | Türkçe, Türkiye’de kullanım | yerli ve milli (slogan), millî üstünlük, homegrown, Türkiye merkezli, #1 |
| net, dürüst, kuyruk, hazır değil | Thelema / unicursal (yalnız marka işareti olarak) | dinî vaat, ezoterik ideoloji |

---

## 4) Açık noktalar

1. **Fiyat/istem tutarsızlığı:** Ana sayfa hero (500 istem / 10.000 TL; saatlik 500 TL) ≠ ürün kayıtları (200 istem / 18.900 TL; saatlik 1.990 TL) ≠ kanus.org JS (200 USD / 500 istem; 10 USD / saat). Hangisi doğru — Demir/HOST onayı gerekir; bu taslak uydurmaz.
2. **Handle:** `kanus-ai-1-aylik-erisim-500-istem` başlıkta 200 istem diyor.
3. **Hakkımızda** hem sitede hem mağazada yok.
4. **İletişim** sayfası boş gövde.
5. **Mağaza adı** hâlâ “Mağazam”.
6. A kararı **yayın izni değildir**; sloganın birebir onayı ve menşe doğrulaması ayrı işlerdir.
7. Logo/SVG değişikliği bu dosyanın kapsamında değildir.

---

## 5) Özet sayaç

- Taranan kaynak: 11 (yukarıdaki tablo).
- Önce/sonra satırı: **~28** metin birimi (meta, i18n anahtarları, shop hero, ürün örnekleri, about/contact boşlukları).
- Canlıda bulunan yerli/milli/homegrown satırı: **0** (silinecek canlı slogan yok; hizalama + çelişki düzeltme önerileri var).
- Yayın / Shopify yazımı / DNS: **yapılmadı**.
