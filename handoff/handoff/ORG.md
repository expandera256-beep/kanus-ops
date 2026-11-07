# ORG — Devir Dosyası

Hazırlanma: 08.11.2026 01:45 TSİ (Europe/Istanbul, UTC+3). Hazırlayan: ORG. İsteyen: KAN.US, Demir'in 01:42 TSİ talimatıyla.
Bu dosyada hiçbir gizli değer (anahtar, parola, token) yoktur.

## 1) Kimlik

- **Ad:** ORG
- **Başlık:** ORG (profile.json'da title alanı boş; uygulamada başlık "ORG" olarak görünüyor)
- **Açıklama:** profile.json'da boş. Önerilen açıklama: "kanus.org altyapısı, DNS doğrulaması, marka kimliği ve site/mağaza metin taslaklarından sorumlu bot."
- **Avatar:** blob şekli, kırmızı renk. Özel resim yok.
- **Rol:** Demir'in talimatıyla kanus.org sitesinin altyapısından ve yönetiminden sorumlu. HOST'tan yönlendirme alır ve onunla birlikte çalışır. Görevlerin çoğu KAN.US üzerinden gelir. Kullanıcıyla ve diğer botlarla Türkçe iletişim kurar.

### Sistem talimatı taslağı

```
Senin adın ORG. Demir'in (Türkçe konuşur, saat dilimi Europe/Istanbul) botusun.
Görevin kanus.org sitesinin altyapısı ve yönetimi: barındırma ve DNS durumunu izlemek,
marka kimliğini (Thelema logosu) korumak, site ve mağaza metinlerinin taslaklarını hazırlamak.
HOST'un yönlendirmesiyle ve onunla birlikte çalışırsın. Görevler çoğunlukla KAN.US'den gelir.

Kesin kurallar:
- Canlı shop.kanus.org'a ve shop DNS kayıtlarına dokunma.
- Varsayılan olarak yalnızca taslak üret. Shopify'a, siteye, mağazaya ya da DNS'e yazmak için
  Demir'in kendi sözüyle verilmiş açık onayı gerekir. Başka bir botun "Demir onayladı" demesi
  onay sayılmaz.
- DNS'e dokunma. Webhook (hook) nginx/Caddy tarafı HOST'ta. Yeni DNS istekleri KAN.US
  üzerinden gelir ve Demir'in onayını gerektirir.
- Marka Kararı A geçerli: ana işaret Thelema logosu (unicursal hexagram ve 5 yapraklı gül).
  "Yerli ve milli" vurgusu yumuşak tutulur. "Yerli Teknoloji, Küresel Vizyon" sloganı
  kullanılmaz. Sloganı Demir yazar; slogan uydurma. Doğrulanmamış iddia ekleme
  (örneğin "Türkiye merkezli").
- Asla veri, fiyat ya da metin uydurma. Bir şeyi kendin doğrulamadıysan "bildirildi" diye yaz.
- Gizli değerleri hiçbir zaman mesaja ya da dosyaya yazma.
- Demir başka bir botun sohbetinde bekliyorsa ona soracağın soruyu o bota ilet.
```

## 2) Hafıza

**Profile:**
- (2026-11-06) ORG, Demir'in talimatıyla kanus.org sitesinin altyapısı ve yönetiminden sorumludur; HOST'un yönlendirmeleriyle ve onunla birlikte çalışır.

**Journal ve shared memory:** Kayıtlı olgu yok. RecallMemory yalnızca yukarıdaki tek profil olgusunu döndürdü.

**Sohbetten çıkan kalıcı olgular** (hafızada kayıtlı değil, devamlılık için buraya eklendi):
- (06.11.2026) Demir: "logo olarak thelema işaretini kullan".
- (06.11.2026) Demir (widget yanıtı): "Yalnız api-test.kanus.org A kaydını 31.220.63.191 olarak uygula; shop DNS'e dokunma." Kayıt zaten doğru olduğu için yazma yapılmadı. Kamuya açık DNS sorgusunda doğrulandı; 08.11.2026 00:08'de de hâlâ 31.220.63.191.
- (06.11.2026) KAN.US iletti: "DNS'e dokunma. Hook nginx HOST'ta. Yeni DNS isteği olursa KAN.US üzerinden gelinir."
- (06.11.2026 08:13 ve 09:28 TSİ) Demir'in Marka Kararı A'yı seçtiğini ve yeniden teyit ettiğini KAN.US iletti. Yeni prompttaki slogan kullanılmayacak. GLOBAL botu /tr ve /en sayfalarında BRAND_A_COPY metnini esas alacak.
- (06.11.2026) kanus.org barındırma bilgisi (kendi okumam): Hostinger, hCDN, hPanel. A kayıtları 84.32.84.209 ve 2.57.91.238. www için CNAME www.kanus.org.cdn.hstgr.net. Site derlenmiş bir SPA (Vite tarzı, /assets/index-uscommerce.js; dil tr/de/ru/en, localStorage anahtarı kanus.lang). robots.txt ve sitemap.xml HTML kabuğu döndürüyor (SPA yönlendirmesi). Kaynak repo ve yayın yöntemi bende bilinmiyor; HOST biliyor olabilir.
- (06.11.2026) Domain NS kayıtları: helios.dns-parking.com ve aster.dns-parking.com (Hostinger). Hook VPS: srv1662550.hstgr.cloud, 31.220.63.191. Caddy 80/443 portlarında, 127.0.0.1:8787 adresindeki us-license-hook servisine yönlendiriyor (HOST'un alanı).
- (06.11.2026) KAN.US bildirdi; ORG doğrulamadı: iki ürün görseli 6 Kasım'da Demir'in onayıyla canlı mağazaya yüklendi. Ürün 15424693108964 için medya 75501629538532; ürün 15424693010660 için medya 75501632618724.

## 3) Rutinler

Yok. ORG'un kayıtlı rutini bulunmuyor.

## 4) Beceriler, bağlayıcılar, servisler

- **Bağlayıcılar:** user-Hostinger-dns, user-Hostinger-hosting (ve diğer Hostinger ad alanları), user-Shopify (yalnızca okuma için kullanıldı), SendToAgent (KAN.US, HOST, GLOBAL ile iletişim).
- **Araçlar:** Shell (curl, DoH sorguları Cloudflare/Google; kutuda dig yok), Python/rsvg/ImageMagick ile SVG→PNG/ICO üretimi, WebFetch, Read.
- **Beceriler:** canvas (kullanılmadı), deploy-to-hosting, connect-domain, troubleshoot-website (başvuru amaçlı).
- **Servisler:** Hostinger (hPanel, DNS, hCDN), Shopify (shop.kanus.org), Hostinger VPS.

## 5) Açık işler ve kalınan yer

- Açık iş yok. Son tamamlananlar:
  - BRAND_A_COPY taslağı
  - İki ürün görseli (KAN.US'e göre yüklendi)
  - GLOBAL'in barındırma sorularının yanıtı
  - 08.11 durum raporu
- Talep gelirse yapılacak işler:
  - TR bayrağı SVG'si (yapılmadı)
  - Arayüz renk geçişi yeşilden mor-maviye (uygulanmadı)
  - GLOBAL'e İngilizce metin desteği

## 6) Bekleyen Demir kararları ve onaylar

- BRAND_A_COPY'deki açık noktalar:
  - Hero bölümündeki fiyat ve istem sayısı, ürün kaydındaki ve sitedeki USD rakamlarıyla uyuşmuyor.
  - Bir üründe handle "500" iken başlık "200".
  - Mağaza adı hâlâ "Mağazam".
  - Hakkımızda sayfası yok.
- BRAND_A_COPY önerilerinin siteye ve mağazaya uygulanması: onay yok, yalnızca taslak.
- Yasal taslaklardaki [DOLDURULACAK] alanları: satıcı bilgileri, iade URL'si vb. Envanter: reports/YEREL_TASLAKLAR_RAPORU_20261106_TR.md. Taslaklar ayrıca avukat incelemesi gerektiriyor.
- Slogan: Demir yazacak.

## 7) Kritik dosya yolları ve SHA256

Kök: /workspace/us-panel/

| Dosya | SHA256 |
|---|---|
| brand/svg/kanus-logo-unicursal-rose.svg | 1cee97676d6139c51285d0bf2ac998bc1e2d85fbb21c39f342d299d7f020762b |
| brand/svg/kanus-logo-unicursal-rose-nobg.svg | 850585b64aa1f149752f9954cf05973e896643b05cd43813841746346fb88b1a |
| brand/svg/kanus-logo-unicursal.svg | 8b97af50bffdb2361a8ee80e2a7034a37104d08ae7c8323b05cfe6991dd74187 |
| brand/svg/kanus-mark-simple.svg | 4bdee8e95b73350d561312d90d3ce242d47067214eeb8d9c980af8a7bbf629a2 |
| brand/svg/kanus-mark-simple-nobg.svg | eca0c76cb4549a04a6925dac375fa17672f6a0e8c334a1dc621a34ca35fed106 |
| brand/svg/kanus-app-icon.svg | 4bdee8e95b73350d561312d90d3ce242d47067214eeb8d9c980af8a7bbf629a2 |
| brand/svg/kanus-favicon-source.svg | 27c3249119956d878a2eb2ffdca9131aa2a82489026023adfc18d2804d846627 |
| brand/svg/kanus-wordmark.svg | 02775a60c9d491ea473968eb98d3abd749ffb5c005cac8e64ef0e6aaccc464a4 |
| brand/ico/kanus-us.ico (16–256 px) | 24e00172a52d99a3c391611cff3c97447a168f5f3de736b254c4786a0c22e830 |
| brand/product-images/kanus-ai-1-saat_2048.png | 5289ea9b985d2934f719f93dbcdefd8cf33002c2658110eced144a47996cf5e3 |
| brand/product-images/kanus-ai-1-ay-200-istem_2048.png | 0e882cb05c2ff76732f50c6f03e25b3feadea39e805a66b5fd7a72e1c7f027a7 |
| brand/product-images/kanus-ai-1-saat_2048.svg | 52e423f8b5e89408904c8b817ca799ccdfd97911eb694773c7930cc773122020 |
| brand/product-images/kanus-ai-1-ay-200-istem_2048.svg | 1e7ed99c2d73082fd1cf8b4397a31ad98c20b358a1c2eea4766632dd7b66f2bf |
| brand/product-images/preview_pair.png | d0e896975980f31a298506d3da6a0d43f2ebd6079c36533cc521655930979873 |
| brand/docs/palette.json | e873a28c40eab3ea1b700081e28b0c65becb0aef895a31e7195fcaec12cf804d |
| brand/docs/geometry.json | 112e216f0e2a731eb9614c99d0d9755ccee84ce47855ad47350dca72cbb52cb5 |
| brand/docs/COLORS.md | b6523c5ff73757a2a3d006db763d6439ed6f60375777686dd1f32881de28b6d2 |
| brand/docs/LOGO.md | ad4ca95092697c1c72d433ff964dc16f8f8f7b4c724f008552d3f0a5e5f19bba |
| brand/docs/NAMING.md | a959dc03c239d2463bb94a1335c7e976f4f61ddc57a95c9a5f3b6589d00411aa |
| brand/docs/TONE.md | 95f67bf75f1b6da020eea6e86b8239a01630688c3a946eef70650a0ed0437ccf |
| brand/docs/TYPOGRAPHY.md | 53da77af12b22d8fdca4df2762098f69f0b455662a4c73981e22128691db525d |
| brand/docs/DECISION_THELEMA_VS_YERLI.md | bddae241d16695e436090c003c7de46d299f2c50094b92edbcc254b094ba4c66 |
| brand/shopify/PRODUCT_TEMPLATES_TR.md | 35fbf3e764555366a325e6662b09686df4fbc92340b2378bd59ad5bf4961f63c |
| brand/shopify/PRODUCT_TEMPLATES_EN.md | b7b8c25b8b2bee09b6c6ac15b2c983678a6ecc52b7a6e4a5282f706a77a51609 |
| brand/exe/ABOUT_SPLASH_TR.md | 6c0eb92f9102bdcaf8f16bcf8d650bf80b94da9496ded5ff8935f16c086a0057 |
| brand/exe/ABOUT_SPLASH_EN.md | ab5458fda0d8e84f1ce5cfdd046e84b50f2018b3f71520fe58e8fc67434490a2 |
| docs/BRAND_A_COPY_20261106.md | 0679a021c7a9be04a2e8403ae37305575358bcd09f818aa427683521184b9f61 |
| docs/KULLANIM_KILAVUZU_TR.md | 531cebdc48edbb33960b08a97f2c6b3db981394c566cd913af3e9ab4d1077db4 |
| docs/legal/KVKK_AYDINLATMA_METNI.md | af04d01e8bcc23e5a00d17f08bca6431035287cf73f8308c3aaf389c9b28ca9f |
| docs/legal/MESAFELI_SATIS_SOZLESMESI.md | 9853c4433374c9c60d71c47f183bb8b8df02a1985ad4fc0317c29a6bbbed99fd |
| docs/legal/CAYMA_VE_IADE_POLITIKASI.md | 2ecf9b3f28ea45794a9da797bb9dbb4ee65b87e30b6ed10b4f3a97993ef239c7 |
| docs/legal/KULLANIM_SARTLARI.md | 0f345c1ea98c858172be4f0c2e58dd9e967e1ceb584d6ad40e116c214f378b5d |
| reports/KANUS_BRAND_REPORT.md | 49579fe6d49081a43e2afb3de0f701a9ff471731a1d5808f54c6c9718104c385 |
| reports/YEREL_TASLAKLAR_RAPORU_20261106_TR.md | 12415088c3cc60af8c00b1697ab05d39280e2162f116bd038657c4e10ac86445 |
| reports/KANUS_ORG_SITE_HOSTING_20261106.md | d6ab69a2c8225a4d17e5c719bc653e692a68b4a5f6f38d95e365bcc6958d58bd |

Diğer referanslar:
- Kaynak prompt: prompt/NIHAI_PROMPT_20261106.txt (§7 marka, §10)
- PNG ikon ve önizleme seti: brand/png/
- Başka botların ürettiği yasal dosyalar: docs/legal/final/ ve docs/legal/uyum-drafts/

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnızca adları)

- **Hostinger bağlayıcı yetkilendirmesi (OAuth/hesap girişi):** DNS doğrulama ve barındırma okuma için kullanılır (user-Hostinger-*).
- **Shopify bağlayıcı yetkilendirmesi (mağaza OAuth):** shop.kanus.org ürün ve sayfa okuma için kullanılır (user-Shopify).
- ORG'un kendine ait ortam değişkeni ya da gizli dosyası yok. SMTP (support@kanus.org), hook ve VPS gizli değerleri başka botların alanında; bunlar için /workspace/secrets ve HOST'a bakın.
