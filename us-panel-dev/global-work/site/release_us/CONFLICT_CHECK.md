# CONFLICT_CHECK — kanus.org/us i18n paketi (YÜKLENMEDİ)

Kaynak: canlı dosyalar (curl, 08.11.2026 ~00:22 TSİ). SPA kaynağı box'ta yok; paket canlı derlenmiş bundle'ın yerel kopyası üzerine 6 noktasal değişiklik + yardımcılar.
- Canlı `index.html` sha256 `87b980dec6020cb8071b33aeef441fa80b146352a7761612065f393588841e34` (1600 B, last-modified 03.11.2026 23:22 GMT)
- Canlı `/assets/index-uscommerce.js` sha256 `04bc662ab6232da8fa1a7a06397cfac65998600644d2ee0b2551183ceaa61e59`

| Paket dosyası | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/index-uscommerce-i18n-75f56119.js` | YOK (yeni ad, içerik hash'li) | Yeni dosya, üzerine yazma yok |
| `public_html/index.html` | VAR — **ÜZERİNE YAZAR** | Tek fark script src satırı; SPA fallback (.htaccess) tüm yollar için bu dosyayı sunar → tüm site yeni bundle'ı yükler |

- `.htaccess`, `api/`, `robots.txt`, `sitemap.xml`, eski asset ve CSS'e dokunulmaz. Eski bundle sunucuda kalır.
- Yüklemeden hemen önce canlı index.html sha256'sı yukarıdakiyle eşleşmeli; eşleşmiyorsa biri siteyi değiştirmiştir → DUR, paketi yeniden üret (`python3 site/us/build_us.py` güncel canlı dosyalarla).
- Önceki statik /tr–/en paketi (`site/release/`) ile dosya çakışması yok (ayrı yollar). İkisi birlikte yüklenebilir.
- Geri alma: `rollback/index.html` (canlı orijinal) → public_html/index.html.
