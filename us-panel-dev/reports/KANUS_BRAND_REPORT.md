# KANUS_BRAND_REPORT

**Tarih:** 2026-11-06 (Europe/Istanbul)  
**Durum:** Yerel dosya üretimi tamam — **yayın yok**  
**Kaynaklar:** `prompt/NIHAI_PROMPT_20261106.txt` §7; Demir talimatı (Thelema unicursal heksagram logo)

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
- TR öncelikli dil, slogan: *Yerli teknolojiyle fikirlerini çalışan yazılıma dönüştür.*
- Koyu lacivert/siyah zemin; mor + mavi neon
- Monospace terminal tipografi yığını
- Mağaza başlık/açıklama şablonları TR+EN (fiyat/canlı dokunulmadı)
- EXE splash/about metinleri
- Doğrulanmamış iddia yok; yer tutucular `{…}` / `[…]`

## §7 ile gerilim / Demir önceliği

| Madde | Not |
|-------|-----|
| Logo = Thelema unicursal heksagram | Demir talimatı §7’deki genel “marka işaretleri”ni somutlaştırır ve **önceliklidir**. |
| “Yerli ve milli” mesaj ↔ Thelema sembolü | Kültürel/marka algısı gerilimi mümkün; raporda açıkça belirtilir. Ürün metinleri yerli vurguyu korur; işaret geometrik logo olarak konumlandı. |
| Minimal Türk bayrağı işareti | Bu turda ayrı bayrak SVG üretilmedi (aşırı yüklememek). İstenirse sonraki tur. |
| Mevcut UI yeşil/lime + bakır (`gui.html` / `ui.go`) | §7 mor/mavi ile uyumsuz. Palette §7 esas; UI hizalaması **karar bekler**, kod değiştirilmedi. |

## Karar bekleyen / uygulanmayan

1. Mevcut panel CSS/lipgloss renklerinin §7 paletine migrasyonu  
2. Minimal TR bayrağı SVG/ikon seti  
3. EN sloganın resmi onayı  
4. `[DESTEK_EPOSTA]`, iade politikası URL, sürüm string’lerinin gerçek değerlerle doldurulması  
5. Wordmark’ta “KANUS” tümü büyük harf vs “Kanus” tipografik tercih (docs’ta KANUS wordmark / Kanus metin ayrımı var)  
6. Thelema işaretinin kamuoyu iletişiminde nasıl anlatılacağı (ezoterik bağlamı öne çıkarmama önerisi docs’ta)

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

