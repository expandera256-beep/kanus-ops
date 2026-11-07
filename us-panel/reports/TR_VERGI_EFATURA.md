# Türkiye KDV / e-Fatura / e-Arşiv — dijital ürün notu

Güncelleme: 06.11.2026 08:12 TSİ (Europe/Istanbul)

**Yalnız rapor.** Muhasebe/vergi danışmanı onayı olmadan uygulama yok. Shopify uygulaması kurulmadı.

## 1. KDV oranı (dijital ürün / yazılım / e-hizmet)

| Konu | Özet | Kaynak |
|---|---|---|
| Genel oran | Listelerde yoksa **%20** (10.07.2023’ten, 7346 sayılı CB Kararı) | GİB «Güncel KDV Oranları» PDF: [cdn.gib.gov.tr … kdv-oranlari.pdf](https://cdn.gib.gov.tr/api/gibportal-file/file/getFileResources?objectKey=arsiv%2Fyardim-kaynaklar%2Fyararli-bilgiler%2Fkdv-oranlari.pdf) |
| Yazılım / SaaS / model kredisi (TR tüketici) | Tipik olarak **%20**; özel liste kaydı yoksa genel oran | Aynı karar + KDV Genel Uygulama Tebliği (GİB) |
| Katalog fiyatlarımız | 695/… TL **KDV dahil** varsayımı (FIYATLANDIRMA_TR) | İç rapor; vergi danışmanı teyidi **NOT TESTED** |
| Hizmet ihracatı (yurt dışı müşteri) | Koşullu istisna mümkün; kanıt/teslim şartları tebliğde | 3065 KDV Kanunu / Genel Uygulama Tebliği — danışman |

Elektronik ortamda yurt dışı sağlayıcı→TR nihai tüketici için özel mükellefiyet (KDV Gen. Uyg. Tebliği I/C-2.1.2.1.2) bu işletme TR yerleşik satıcı ise farklıdır; senaryoyu muhasebeci ayırmalı.

## 2. e-Fatura / e-Arşiv (hadler — 2026 çerçevesi)

Kaynak omurga: **VUK Genel Tebliği 509** (GİB güncel PDF) ve sektör duyuruları. Rakamlar dönemsel güncellenir — yayın tarihini kontrol edin.

| Konu | Kısa not | Kaynak |
|---|---|---|
| e-Fatura zorunlu geçiş (genel brüt satış) | 2025 hasılatı **≥ 3 milyon TL** → geçiş **01.07.2026** (yaygın özet) | ASMMMO özeti [PDF](https://www.asmmmo.org.tr/userfiles/others/files/Mvzt/Gh/26/06-25-e-Fatura%2C%20e-Ar%C5%9Fiv%20fatura%2C%20e-%C4%B0rsaliye%20Uygulamas%C4%B1na%20ge%C3%A7i%C5%9F%20tarihi%201%20Temmuz%202026%20Olan%20M%C3%BCkellefler.pdf); asıl metin VUK 509 |
| İnternet üzerinden mal/hizmet satışı | Brüt satış **≥ 500 bin TL** (2025) → e-Fatura/e-Arşiv **01.07.2026** | Aynı ASMMMO / 509 özeti |
| e-Arşiv | e-Fatura’ya geçenler e-Arşiv’e de; 2026’dan itibaren tutara bakılmaksızın e-Arşiv düzenleme kuralları sıkılaştı (basit usul vb. istisnalar dönemsel) | GİB e-Arşiv infografik / 509 IV.2; [e_arsiv_fatura.pdf](https://cdn.gib.gov.tr/api/gibportal-file/file/getFileResources?objectKey=arsiv%2Fyardim-kaynaklar%2Finfografikler%2Fpdfs%2Fe_arsiv_fatura.pdf) |
| Mesafeli satış | Mesafeli sözleşmeler (TKHK) + fatura yükümlülüğü ayrı; e-ticaret aracı/platform kuralları | 6502 TKHK; GİB e-belge |

**US Panel için pratik:** Dijital lisans satışı «internet ortamında hizmet satışı» sınıfına girebilir → **500 bin TL** haddi erken tetikleyebilir. Danışman + GİB ebelge.gib.gov.tr teyidi şart.

## 3. Shopify App Store — TR e-Fatura / e-Arşiv seçenekleri (kurulmadı)

| Uygulama | Link | Not |
|---|---|---|
| Paraşüt E-Fatura Entegrasyonu | https://apps.shopify.com/parasut-entegrasyonu-1?locale=tr | Otomatik Paraşüt e-fatura |
| eLogo E-Fatura — EntegreTürk | https://apps.shopify.com/elogo-e-fatura-entegreturk?locale=tr | Logo/eLogo |
| Uyumsoft E-Fatura — EntegreTürk | https://apps.shopify.com/uyumsoft-e-fatura?locale=tr | Uyumsoft / HB e-Faturam |
| Pax: eFatura Entegrasyonu | https://apps.shopify.com/e-fatura-entegrasyonu-1 | e-Arşiv/e-Fatura; kontörlü |

Seçim: entegratör sözleşmesi + GİB mükellefiyet + Shopify plan. **Install yok.**

## 4. Durum

| Madde | Durum | Sahip |
|---|---|---|
| %20 KDV varsayımı dokümante | PASS (GİB oran PDF) | Ticari |
| e-Fatura had kontrolü işletme cirosu | NOT TESTED | Ticari + muhasebe |
| Shopify e-fatura app kurulumu | NOT TESTED | SHOP / Ticari |
| Fatura↔lisans outbox hukuki metin | NOT TESTED | PANEL DEV + hukuk |
