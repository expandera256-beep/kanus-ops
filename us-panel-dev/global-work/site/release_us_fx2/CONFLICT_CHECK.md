# CONFLICT_CHECK — release_us_fx2 (YÜKLENMEDİ)

Ön koşul: `release_us_fx` canlıda (08.11 ~00:59 TSİ curl ile görüldü: index → i18nfx-5a375f29). Yüklemeden hemen önce canlı `public_html/index.html` sha256 = `4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8` (release_us_fx index.html) olmalı. Değilse DUR.
- Hâlâ orijinal canlı index (`87b980dec6020cb8071b33aeef441fa80b146352a7761612065f393588841e34`) duruyorsa: önce release_us'i yükleyin ya da `public_html/assets/index-uscommerce-i18nfx-5a375f29.js` dosyasını da sunucuya koyun (bu paket o asset'i silmez ama referans vermez; geri alma için gerekli).

| Dosya | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/index-uscommerce-i18nfx2-dc1e3536.js` | YOK | Yeni dosya |
| `public_html/index.html` | VAR — ÜZERİNE YAZAR | Yalnız script src değişir |
| `public_html/fx.json` | VAR — üzerine yazar (cron kuruluysa yüklemeyin; cron'un dosyası daha günceldir) | TCMB 06.11.2026 anlık görüntüsü |

- `cron/upload_public_html/_fx/` (fx_cron.php + .htaccess) ayrı adımda public_html/_fx/'e yüklenir (YOK; yeni klasör). `cron/` altındaki diğer dosyalar yüklenmez.
- Ödeme/ledger/api, .htaccess, eski asset'lere dokunulmaz.
- Geri alma: `rollback/index.html` (release_us_fx durumu); `rollback/index.release_us.html`; tam başa dönüş `rollback/index.live_original.html`. Önceki asset `index-uscommerce-i18nfx-5a375f29.js` sunucuda kalır.
