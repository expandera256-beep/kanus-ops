# Logo — Unicursal heksagram (Thelema)

**Demir talimatı (öncelikli):** Logo = Thelema unicursal heksagram. Geometri sıfırdan üretildi; Wikimedia SVG yalnız yapı/oran referansı (kopyalanmadı).

## Doğru geometri (2026-11-06 düzeltme)

Önceki sürüm düzgün altıgen köşelerini `0→2→4→1→3→5` sırasıyla birleştiriyordu; bu **unicursal heksagram değildir** (üçgen+dörtgen görünümü).

Referans yapı: Wikimedia `Solid_unicursal_hexagram.svg` (Wikipedia *Unicursal hexagram*).

### Köşeler (normalize, merkez 0, SVG y↓, üst `(0,-1)`)

| Köşe | (x, y) |
|------|--------|
| top | `(0, -1)` |
| ur | `(+A, -B)` |
| lr | `(+A, +B)` |
| bottom | `(0, +1)` |
| ll | `(-A, +B)` |
| ul | `(-A, -B)` |

`A = 74.531/90 ≈ 0.828122`, `B = 37.929/90 ≈ 0.421433`  
(Referans SVG oranları; dikey eksene göre **zorlanmış ayna simetrisi**.)

### Path sırası (tek çizgi)

`top → lr → ul → bottom → ur → ll → top`

Makine kontrolü: `docs/geometry.json`; dikey ayna altında kenar kümesi değişmez; üst/alt dikey eksende.

### Gül (Thelema)

Merkezde **5** damla yaprak, 72° aralık (Crowley SVG’deki 5×`drop` düzeninin kendi geometrisi; yaprak path sayısı = 5). Önceki sürümdeki 10+ loblu çiçek kaldırıldı.

## Dosyalar

| Dosya | Açıklama |
|-------|----------|
| `svg/kanus-logo-unicursal-rose.svg` | Tam logo (heksagram + 5 yaprak + daire) |
| `svg/kanus-logo-unicursal-rose-nobg.svg` | Şeffaf zemin |
| `svg/kanus-logo-unicursal.svg` | Yalnız heksagram + daire |
| `svg/kanus-mark-simple.svg` | Küçük boyut / kalın sade mark |
| `svg/kanus-mark-simple-nobg.svg` | Sade, şeffaf |
| `svg/kanus-app-icon.svg` | Uygulama ikonu |
| `svg/kanus-favicon-source.svg` | ≤32px kaynak |
| `svg/kanus-wordmark.svg` | İşaret + KANUS + US PANEL |
| `ico/kanus-us.ico` | 16–256 Windows ICO |
| `png/kanus-brand-preview-strip.png` | Önizleme şeridi |
| `png/preview-rose-256.png` / `preview-hex-256.png` / `preview-simple-128.png` | Hızlı önizleme |
| `docs/geometry.json` | Köşe ve path kaydı |

## Kullanım

- ≥64px / pazarlama: rose varyantı  
- ≤48px / görev çubuğu / EXE: simple → `kanus-us.ico`  
- Açık alan ≈ işaret yüksekliğinin %20’si  

## §7 notu

Thelema işareti ↔ “yerli ve milli” mesaj gerilimi: Demir talimatı öncelikli; metinlerde yerli vurgu korunur, işaret geometrik marka olarak konumlanır.
