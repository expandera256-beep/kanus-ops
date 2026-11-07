# CONFLICT_CHECK — release_us_fx (YÜKLENMEDİ)

Ön koşul: `release_us` önce yüklenmiş olmalı. Yüklemeden hemen önce canlı `public_html/index.html` sha256 = `a828fa73880192c4cf75c2b37cd5c648be8054c30108c8e84a2bb6c5e866a926` (release_us index.html) olmalı. Değilse DUR.
- Hâlâ orijinal canlı index (`87b980dec6020cb8071b33aeef441fa80b146352a7761612065f393588841e34`) duruyorsa: önce release_us'i yükleyin ya da `public_html/assets/index-uscommerce-i18n-75f56119.js` dosyasını da sunucuya koyun (bu paket o asset'i silmez ama referans vermez; geri alma için gerekli).

| Dosya | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/index-uscommerce-i18nfx-5a375f29.js` | YOK | Yeni dosya |
| `public_html/index.html` | VAR — ÜZERİNE YAZAR | Yalnız script src değişir |
| `public_html/fx.json` | YOK | Yeni; TCMB 06.11.2026 anlık görüntüsü. .htaccess `!-f` koşulu gerçek dosyayı sunar |

- `cron/upload_public_html/_fx/` (fx_cron.php + .htaccess) ayrı adımda public_html/_fx/'e yüklenir (YOK; yeni klasör). `cron/` altındaki diğer dosyalar yüklenmez.
- Ödeme/ledger/api, .htaccess, eski asset'lere dokunulmaz.
- Geri alma: `rollback/index.html` (release_us durumu). Tam başa dönüş: `rollback/index.live_original.html`.
