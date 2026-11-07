# KANUS_BRAND_REPORT

**Tarih:** 2026-11-06 (Europe/Istanbul)  
**Durum:** Yerel dosya üretimi tamam — **yayın yok**  
**Kaynaklar:** `prompt/NIHAI_PROMPT_20261106.txt` §7; Demir talimatı (Thelema unicursal heksagram logo); KAN.US’nin ilettiği 06.11.2026 08:13 TSİ (Europe/Istanbul) Demir seçimi: A.

## Özet

KAN.US / US Panel için marka kimliği paketi `/workspace/us-panel/brand/` altında oluşturuldu. Logo, Demir’in talimatıyla **sıfırdan çizilmiş unicursal heksagram** (Thelema; gül varyantı + küçük boyut sade varyant). §7’deki koyu lacivert/siyah zemin ve mor/mavi neon vurgular uygulandı. Shopify / site / DNS / sunucuya yazılmadı.

## Üretilen dosyalar

### docs/
- `brand/docs/NAMING.md`
- `brand/docs/COLORS.md`
- `brand/docs/palette.json`
- `brand/docs/TYPOGRAPHY.md`
- `brand/docs/TONE.md`
- `brand/docs/LOGO.md`
- `brand/README.md`

### svg/
- `brand/svg/kanus-logo-unicursal-rose.svg`
- `brand/svg/kanus-logo-unicursal-rose-nobg.svg`
- `brand/svg/kanus-logo-unicursal.svg`
- `brand/svg/kanus-mark-simple.svg`
- `brand/svg/kanus-mark-simple-nobg.svg`
- `brand/svg/kanus-app-icon.svg`
- `brand/svg/kanus-favicon-source.svg`
- `brand/svg/kanus-wordmark.svg`

### png/
- `brand/png/kanus-logo-unicursal-rose-512.png`
- `brand/png/kanus-logo-unicursal-512.png`
- `brand/png/kanus-mark-simple-512.png`
- `brand/png/kanus-app-icon-512.png`
- `brand/png/kanus-wordmark-640.png`
- `brand/png/kanus-brand-preview-strip.png`
- `brand/png/preview-rose-256.png`
- `brand/png/preview-simple-128.png`
- `brand/png/kanus-icon-{16,24,32,48,64,128,256,512}.png`
- `brand/png/kanus-ico-src-{16,24,32,48,64,128,256}.png`

### ico/
- `brand/ico/kanus-us.ico` — boyutlar: 16, 24, 32, 48, 64, 128, 256

### shopify/ & exe/
- `brand/shopify/PRODUCT_TEMPLATES_TR.md`
- `brand/shopify/PRODUCT_TEMPLATES_EN.md`
- `brand/exe/ABOUT_SPLASH_TR.md`
- `brand/exe/ABOUT_SPLASH_EN.md`

## Doğrulama

| Kontrol | Sonuç |
|---------|--------|
| SVG XML (ElementTree parse) | PASS — 8/8 |
| ICO çoklu boyut (Pillow IcoFile) | PASS — {(16,16),(24,24),(32,32),(48,48),(64,64),(128,128),(256,256)} |
| PNG önizleme (rose + simple) | PASS — üretildi ve görsel incelendi |
| Canlı yayın / Shopify / DNS | Yapılmadı (bilinçli) |
| Font dosyası indirme | Yapılmadı (sistem font yığınları) |

## §7 uygulananlar

- Kanus.org / US Panel isim çerçevesi
- TR öncelikli dil, slogan: *Fikirlerini yazılıma dönüştür.*
- Koyu lacivert/siyah zemin; mor + mavi neon
- Monospace terminal tipografi yığını
- Mağaza başlık/açıklama şablonları TR+EN (fiyat/canlı dokunulmadı)
- EXE splash/about metinleri
- Açık yer tutucular `[DOLDURULACAK]` olarak standartlaştırıldı; marka hedefleri hukuki/menşe kanıtı değildir.

## §7 ile gerilim / Demir önceliği

| Madde | Not |
|-------|-----|
| Logo = Thelema unicursal heksagram | Mevcut varlıklar Demir talimatıyla korunmuştur; KAN.US’nin ilettiği Demir seçimiyle A kararı kaydedildi: Thelema ana işaret kalır. |
| “Yerli ve milli” mesaj ↔ Thelema sembolü | Algısal gerilim mümkün; etki araştırmayla doğrulanmış değildir. [Karar kaydı](../brand/docs/DECISION_THELEMA_VS_YERLI.md): A seçildi; yerli/milli söylemi metin ve tonda yumuşatıldı. |
| Minimal Türk bayrağı işareti | Bu turda ayrı bayrak SVG üretilmedi (aşırı yüklememek). İstenirse sonraki tur. |
| Mevcut UI yeşil/lime + bakır (`gui.html` / `ui.go`) | §7 mor/mavi ile uyumsuz. Palette §7 esas; UI hizalaması **karar bekler**, kod değiştirilmedi. |

## Karar bekleyen / uygulanmayan

1. Mevcut panel CSS/lipgloss renklerinin §7 paletine migrasyonu  
2. Minimal TR bayrağı SVG/ikon seti  
3. Yeni nötr sloganın ve EN karşılığının birebir resmî metin onayı (A seçimi konumlandırma kararıdır)  
4. Destek: `support@kanus.org`; kaynak sürümü: `2026.11.06.7`. İade politikası URL’si: [DOLDURULACAK]; derleme kimliği/telif yılı: [DOLDURULACAK]. Kaynak ve sınır aşağıdadır.  
5. Wordmark’ta “KANUS” tümü büyük harf vs “Kanus” tipografik tercih (docs’ta KANUS wordmark / Kanus metin ayrımı var)  
6. A kapsamında Thelema ana işaret kalır; ürün iletişiminde dinî/ideolojik vaat veya teknoloji menşei iddiası kurulmaz. Ayrıntılı dış iletişim metni taslak olarak kalır.

## Bilinçli yapılmayanlar

- Shopify ürün güncellemesi / canlı mağaza  
- Site, DNS, sunucu, harici servis değişikliği  
- Başka marka taklidi  
- Müşteri sayısı / sertifika / fiyat iddiası uydurma

## Geometri düzeltmesi (2026-11-06)

Önceki unicursal path hatalıydı (düzgün altıgen `0→2→4→1→3→5`). Wikimedia *Solid_unicursal_hexagram.svg* oranlarıyla yeniden çizildi (kopya değil; kendi koordinatlar + zorunlu dikey ayna simetrisi).

- Path: `top → lr → ul → bottom → ur → ll → top`
- `A≈0.828122`, `B≈0.421433` (normalize, üst `(0,-1)`)
- Referans path ile max köşe sapması: **&lt; 0.01 px** (aynı ölçekte)
- Geometri-only L↔R maske uyumsuzluğu: **&lt; 0.03%** piksel (AA)
- Gül: **5** yaprak (önceki 10+ lob düzeltildi)
- ICO yeniden üretildi: 16,24,32,48,64,128,256 — PASS
- Önizleme şeridi: `/workspace/us-panel/brand/png/kanus-brand-preview-strip.png`
- Ek önizleme: `preview-rose-256.png`, `preview-hex-256.png`, `preview-simple-128.png`
- Ayrıntı: `brand/docs/LOGO.md`, `brand/docs/geometry.json`

Yayın yok.


## Yerel taslak güncellemesi — 06.11.2026, TSİ (Europe/Istanbul)

- Karar notu: [Thelema ve yerli/milli A/B](../brand/docs/DECISION_THELEMA_VS_YERLI.md). **A kararı 08:13 TSİ’de KAN.US üzerinden iletildi; logo/geometri/görsel dosyalar değiştirilmedi.**
- Kullanım kılavuzu: [Türkçe taslak](../docs/KULLANIM_KILAVUZU_TR.md).
- Hukuk taslakları: [KVKK](../docs/legal/KVKK_AYDINLATMA_METNI.md), [mesafeli satış](../docs/legal/MESAFELI_SATIS_SOZLESMESI.md), [cayma/iade](../docs/legal/CAYMA_VE_IADE_POLITIKASI.md), [kullanım şartları](../docs/legal/KULLANIM_SARTLARI.md).
- Tam değişiklik ve eksik alan listesi: [yerel teslim raporu](YEREL_TASLAKLAR_RAPORU_20261106_TR.md).

### Kaynağı bulunan alanlar
| Alan | Değer | Kaynak / sınır |
|---|---|---|
| Destek e-postası (TR+EN About) | support@kanus.org | `/workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md`, mailbox/SMTP FROM kaydı; canlı test yapılmadı; KVKK/cayma bildirimi alma yetkisi doğrulanmadı. |
| Sürüm (TR+EN About) | 2026.11.06.7 | `/workspace/us-panel/work2/kiyas/term/release.go`, `buildVersion`; yalnız kaynak sürümü, dağıtılan EXE veya üretim sürümü değildir. |
| Marka/ürün ve web/mağaza | Kanus.org / US Panel; https://kanus.org / https://shop.kanus.org | `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt` §7, §6; mevcut alanlar korundu. Hukuki şirket unvanı değildir. |

### Kalan yer tutucular ve iddia sınırı
- About derleme kimliği ve telif yılı: [DOLDURULACAK]. Belge tarihi telif yılı kanıtı sayılmadı.
- Ürün şablonları: model isteği miktarı, plan adı, kredi/süre, geçerlilik günü, plan verisi, siparişe özgü teslimat/e-posta akışı, iade politikası URL’si: [DOLDURULACAK]. Şablonlar tek bir SKU ile eşleşmedi; kaynakta birden fazla katalog/plan bağlamı bulunduğundan rastgele miktar/fiyat taşınmadı.
- Fiyat, saatlik ücret, hukuki satıcı/merkez bilgileri uydurulmadı. §7’deki önceki yerli/milli söylemi A kararıyla yumuşatıldı; müşteri metinlerinde menşe iddiasız ürün anlatımı kullanılır. Doğrulanmamış “Türkiye merkezli” tanımı About metinlerinden kaldırıldı. Şirket veya teknoloji menşei hakkında yeni bir iddia eklenmedi.
- Önceki doğrulama tablosu önceki marka üretim turuna aittir; bu tur görsel doğrulama, EXE çalıştırma veya canlı test yapılmadı. Dört hukuk metni hukukçu kontrolü bekler.
- Hiçbir Shopify/site/DNS/sunucu/harici servis yazımı, mesaj/e-posta veya yayın yapılmadı.

## A kararı uygulaması — 06.11.2026 08:13 TSİ (Europe/Istanbul)
Kaynak: KAN.US’nin ilettiği Demir seçimi: “A: Thelema logosu ana işaret kalsın, yerli ve milli söylemi metin ve ton düzeyinde yumuşatılsın.” [Karar kaydı](../brand/docs/DECISION_THELEMA_VS_YERLI.md).

- Thelema ana işareti ve beş yapraklı gül varyantı korunur; B uygulanmaz.
- TR metinlerde yerli/milli slogan yerine “Fikirlerini yazılıma dönüştür.”; EN taslaklarda “Turn your ideas into software.” kullanıldı. Bunlar konumlandırmayı uygulayan metin taslaklarıdır; ayrıca birebir onaylandıkları ileri sürülmez.
- About tanımı menşe yerine “Yapay zekâ destekli yazılım geliştirme paneli.” / “AI-assisted software development panel.” olarak düzenlendi. Türkçe ve geliştirici odağı korunur; şirket/altyapı kökeni veya millî üstünlük iddiası eklenmez.
- Kılavuza karar/ton notu eklendi; kullanım adımları, kaynaklar ve bütün mevcut yer tutucular korundu.
- Önceki teslim raporu önceki turun tarihsel kaydıdır; bu yeni karar için bu bölüm ve karar notu esas alınmalıdır. Yasal metinler bu tur değiştirilmedi.

### Bu tur güncellenen dosyalar
- `/workspace/us-panel/reports/KANUS_BRAND_REPORT.md`
- `/workspace/us-panel/brand/docs/NAMING.md`
- `/workspace/us-panel/brand/docs/TONE.md`
- `/workspace/us-panel/brand/docs/LOGO.md`
- `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_TR.md`
- `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_EN.md`
- `/workspace/us-panel/brand/exe/ABOUT_SPLASH_TR.md`
- `/workspace/us-panel/brand/exe/ABOUT_SPLASH_EN.md`
- `/workspace/us-panel/docs/KULLANIM_KILAVUZU_TR.md`
- `/workspace/us-panel/brand/docs/DECISION_THELEMA_VS_YERLI.md`

### Bu tur kontrol kapsamı
Metin/bağlantı, karar tutarlılığı, yer tutucu korunumu ve görsel/yasal dosyaların SHA256 değişmezliği kontrol edilir. Sonuç aşağıdaki kontrol kaydında belirtilir. Önceki SVG/ICO üretim testleri bu tur yeniden yapılmış sayılmaz. Yayın, harici yazım ve mesaj/e-posta yok.

**Kontrol sonucu: PASS.** 10 hedef metin dosyası güncellendi; tüm mevcut `[DOLDURULACAK]` sayıları dosya bazında korundu, yerel bağlantılar geçerli, eski aktif A/B bekleme ifadeleri kaldırıldı. 38 korunan görsel/geometri/yasal dosyanın SHA256 değeri değişmedi. Uygulama veya canlı test yapılmadı.
