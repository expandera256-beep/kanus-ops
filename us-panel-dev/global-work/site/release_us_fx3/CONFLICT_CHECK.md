# CONFLICT_CHECK — release_us_fx3 (YÜKLENMEDİ)

Taban: canlı durum, curl ile 08.11.2026 ~01:20 TSİ indirildi (salt okuma).
- Canlı `index.html` sha256 `8fd6127ddc97b8830347244826325bde63e02cd5f4299b7e5ba37a582596e21c` → `/assets/index-uscommerce-i18nfx-trtl-ee14acd6.js`
- Canlı bundle sha256 `ee14acd6d255c72612a4cbaa4ae53435cb0d2663a890646b6d302d2b4af42ab7` (SHOP: 5a375f29 + TR usCard/appsIntro/usIntro TL; Demir onayı 01:10 TSİ)
Yüklemeden hemen önce canlı index.html sha256 yukarıdakiyle eşleşmeli; değilse DUR ve yeniden üret.

| Dosya | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/index-uscommerce-i18nfx3-7c513199.js` | YOK | Yeni |
| `public_html/index.html` | VAR — ÜZERİNE YAZAR | Yalnız script src |
- `fx.json` pakette YOK (canlıdaki/cron'un dosyasına dokunulmaz). `cron/` = fx2 ile aynı (ayrı adım, INSTALL_CRON.md).
- Bundle gövdesi canlıyla bayt bayt aynı (BODY_DIFF.txt); TR TL fiyat dizeleri korunur. Yalnız baştaki fx bileşeni değişir.
- Geri alma: `rollback/index.html` (= canlı index-uscommerce-i18nfx-trtl-ee14acd6.js index'i). Canlı asset sunucuda kalır.
- release_us_fx2 DEPRECATED: eski tabana dayanır, yüklenirse TR fiyat dizelerini geri alır.
