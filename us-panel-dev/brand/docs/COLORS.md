# Renk paleti

§7: koyu lacivert/siyah zemin; mor ve mavi neon vurgular.

| Token | Hex | Kullanım | Kontrast notu |
|-------|-----|----------|---------------|
| `ink` | `#05070f` | Ana zemin, splash | — |
| `navy` | `#0a0e1a` | Sayfa / panel zemin | — |
| `navy_mid` | `#12182b` | Kart / rail | — |
| `panel` | `#161e33` | Yükseltilmiş yüzey | — |
| `line` | `#2a3555` | Kenarlık, ayırıcı | Düşük kontrast; metin için kullanma |
| `muted` | `#8b95b2` | İkincil metin | `ink` üzerinde ~AA büyük metin; gövde için dikkat |
| `paper` | `#e8eefc` | Birincil metin (koyu tema) | `ink` / `navy` üzerinde yüksek kontrast (hedef AAA) |
| `blue` | `#3b82f6` | Bağlantı, bilgi | `ink` üzerinde ikon/vurgu |
| `blue_neon` | `#60a5fa` | Neon vurgu, logo gradient | İnce çizgi/glow |
| `blue_deep` | `#1d4ed8` | Hover / derin glow | — |
| `purple` | `#a855f7` | Neon vurgu, logo | — |
| `purple_neon` | `#c084fc` | Logo stroke açık uç | — |
| `purple_deep` | `#7c3aed` | Glow merkezi | — |
| `cyan` | `#22d3ee` | GPU/model “hazır” vb. durum (opsiyonel) | — |
| `success` | `#34d399` | Başarı | Metin için koyu zemin |
| `warn` | `#fbbf24` | Uyarı | — |
| `danger` | `#f87171` | Hata | — |
| `flag_red` | `#E30A17` | Yalnız minimal TR bayrağı işareti | §7: minimal kullanım |
| `flag_white` | `#FFFFFF` | Bayrak beyazı | — |

## Logo gradient

Unicursal stroke: `purple_neon` → `blue_neon` → `purple` (köşegen).

## Mevcut UI ile fark (karar bekler)

`work2/kiyas/term` arayüzünde yeşil/lime (`#d6f25c`, `#b6eb79`) ve bakır (`#e7a06a`) vurgular var. §7 mor/mavi neon önceliklidir; yeşilin aşamalı hizalanması ürün kararı bekler. Bu paket §7 paletini tanımlar; canlı UI’ya uygulamadı.
