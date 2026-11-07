# kanus.org release_us_fx canlı yükleme, 08.11.2026 ~00:50 TSİ

Onay: Demir, 00:48 TSİ. Yöntem: Hostinger connector hosting_files_generate-upload-url + TUS (override=true), hesap u346856246, public_html. Başka dosya silinmedi ya da değiştirilmedi; cron/ yüklenmedi.

## Ön kontrol
- sha256sum -c SHA256SUMS: 10/10 OK
- Canlı index yüklemeden önce: a828fa73880192c4cf75c2b37cd5c648be8054c30108c8e84a2bb6c5e866a926
- Yedek: /workspace/us-panel-dev/global-work/site/live_backup_20261108/index.html.release_us (a828fa73…a926)

## Yüklenenler (sırasıyla; canlıdan geri okunup sha eşleşti)
1. public_html/assets/index-uscommerce-i18nfx-5a375f29.js: 252363 B, 5a375f292604694e6f905a9654353700edf9dc4bb595c5df250354048991c5d0
2. public_html/fx.json: 337 B, 9eac8d08317d5e5414efb3194ba12658498833ca4ca26fd2db9bf890147ae935 (TCMB 06.11.2026, USD 50.0724, EUR 55.9609)
3. public_html/index.html: 1596 B, 4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8

Yeni canlı index sha256: 4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8. /, /us, /contact ve /news: 200, yeni bundle'a bağlı. /fx.json: 200, sha eşleşiyor.

## VERIFY_AFTER_UPLOAD.sh
```
PASS / references index-uscommerce-i18nfx-5a375f29.js
PASS /us references index-uscommerce-i18nfx-5a375f29.js
PASS /contact references index-uscommerce-i18nfx-5a375f29.js
PASS /news references index-uscommerce-i18nfx-5a375f29.js
PASS asset sha256
PASS asset content-type
PASS previous asset still served (rollback safe)
PASS fx.json fetch
PASS fx.json content-type
06.11.2026 {'USD': 50.0724, 'EUR': 55.9609}
PASS fx.json schema/age
PASS no live TCMB call in bundle
PASS no 'Yerli Teknoloji'
PASS no 'Küresel Vizyon'
PASS no 'Built in Türkiye'
PASS no 'Designed for the World'
Rollback: rollback/index.html -> public_html/index.html (release_us state); delete public_html/fx.json optional (component hides FX when absent).
ALL PASS
exit=0
```

## Bölüm B (cron): YAPILMADI
Ana ajan durdurdu. fx_cron.php yalnızca okundu ve incelemeden geçti: yalnız public_html/fx.json yazıyor (.tmp üzerinden rename ile), tek ağ çağrısı TCMB today.xml, hatalı yanıtta exit 2 ve dosyaya dokunmuyor. Sunucuya yüklenmedi, cron eklenmedi, çalıştırılmadı.

## Geri dönüş
release_us_fx/rollback/index.html (a828fa73…a926) aynı TUS yoluyla public_html/index.html üzerine yüklenir (DEST=index.html, override=true), ardından sha kontrol edilir. Tam başa dönüş için rollback/index.live_original.html (87b980de…1e34). fx.json ve yeni asset kalabilir.

---
# Bölüm B: kur cron'u (release_us_fx2), 08.11.2026 ~01:15–01:27 TSİ
Onay: Demir, 00:48 TSİ. Eski release_us_fx/cron dosyaları kullanılmadı. release_us_fx2 site paketi yüklenmedi.

## Hash ve inceleme
- _fx/fx_cron.php 3ff10ac51f88c35ef2a35f909553a0f156794ba92716d7ecff521db23b46724e: OK
- _fx/.htaccess 6559fbacdc833562010b31370bc9d2aa39c1b28c8fadb7c08ada5e3fc916caf1: OK
- PHP_SAPI !== 'cli' ise 403 döner ve çıkar: evet.
- Tek kalıcı çıktı public_html/fx.json. Yan dosyalar yalnız _fx/ içinde: fx_cron.lock ve geçici fx.json.tmp.* (rename ile taşınır).
- Ağ çağrısı yalnız TCMB today.xml, yönlendirme kapalı. Not: KANUS_FX_SRC env değişkeni kaynağı değiştirebilir; cron bu değişkeni ayarlamıyor.

## Yükleme (TUS, override=true; sırasıyla)
1. public_html/_fx/.htaccess (237 B)
2. public_html/_fx/fx_cron.php (4157 B)

## Web erişimi
curl -I https://kanus.org/_fx/fx_cron.php: HTTP/2 403, Hostinger'ın standart 403 sayfası (script çalışmadı, script çıktısı yok). _fx/fx_cron.lock ve _fx/.htaccess de 403.

## Sunucu
- PHP 8.3.36 (site ve /usr/bin/php CLI).
- Sunucu saat dilimi UTC (cron ile doğrulandı). Bu yüzden 16:05 TSİ = 13:05 UTC.

## Cron
- uid ad4Ahca1ln
- Zamanlama: 5 13 * * 1-5 (hafta içi 16:05 TSİ)
- Komut: /usr/bin/php /home/u346856246/domains/kanus.org/public_html/_fx/fx_cron.php
- Hesaptaki tek cron bu. Geçici test ve sonda cron'ları (pBLlbnN1pY, QJI7eg1dzr, lavdFrDuwx) silindi.

## İlk çalıştırma
Geçici bir dakikalık cron ile bir kez çalıştırıldı (pBLlbnN1pY). Sonuç: exit=0. fx.json yeniden yazıldı: 200, geçerli JSON, sha256 1bfa293f5747d49fda35a1f6e95a4a63e3ba709603c1f75cbd8d590bb73ad060, tcmb_tarih 06.11.2026 (bülten 2026/209), USD 50.0724, EUR 55.9609, generated_at 2026-11-07T22:17:04Z (01:17 TSİ). Hafta sonu olduğu için TCMB'nin son bülteni hâlâ 06.11.
Not: Geçici cron birkaç dakika her dakika çalıştı (aynı veri, zararsız).
VERIFY tekrar (release_us_fx script'i, ~01:28 TSİ): fx.json kontrolleri PASS (06.11.2026). 4 FAIL, çünkü canlı index artık i18nfx-5a375f29'u değil index-uscommerce-i18nfx-trtl-ee14acd6.js'i yüklüyor. Canlı index sha256 8fd6127ddc97b8830347244826325bde63e02cd5f4299b7e5ba37a582596e21c, release_us_fx3 paketiyle eşleşiyor. Bu index'i ben yüklemedim; geri almadım.

## Geri alma
hosting_cron-jobs_delete uid=ad4Ahca1ln. İstenirse public_html/_fx/ içindeki iki dosya da kaldırılır; fx.json kalabilir.
