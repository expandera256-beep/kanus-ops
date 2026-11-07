# Çakışma kontrolü — 06.11.2026 09:43 TSİ, salt okuma curl
Beklenen: paketteki her yol bugün SPA fallback (aynı 1600 B kabuk) dönüyor → sunucuda bu adlarla dosya yok; yükleme üzerine yazmaz.
```
/en/api/ 200 text/html 1600 SPA_FALLBACK
/en/docs/ 200 text/html 1600 SPA_FALLBACK
/en/download/ 200 text/html 1600 SPA_FALLBACK
/en/ 200 text/html 1600 SPA_FALLBACK
/en/pricing/ 200 text/html 1600 SPA_FALLBACK
/en/product/ 200 text/html 1600 SPA_FALLBACK
/en/support/ 200 text/html 1600 SPA_FALLBACK
/robots.txt 200 text/html 1600 SPA_FALLBACK
/sitemap.xml 200 text/html 1600 SPA_FALLBACK
/tr/api/ 200 text/html 1600 SPA_FALLBACK
/tr/docs/ 200 text/html 1600 SPA_FALLBACK
/tr/download/ 200 text/html 1600 SPA_FALLBACK
/tr/ 200 text/html 1600 SPA_FALLBACK
/tr/pricing/ 200 text/html 1600 SPA_FALLBACK
/tr/product/ 200 text/html 1600 SPA_FALLBACK
/tr/support/ 200 text/html 1600 SPA_FALLBACK
```

Paket DIŞI hukuki yollar (yüklenmeyecek; yükleme sonrası da SPA fallback kalmalı):
```
/tr/privacy/ 200 SPA_FALLBACK
/tr/terms/ 200 SPA_FALLBACK
/tr/refund/ 200 SPA_FALLBACK
/en/privacy/ 200 SPA_FALLBACK
/en/terms/ 200 SPA_FALLBACK
/en/refund/ 200 SPA_FALLBACK
```
