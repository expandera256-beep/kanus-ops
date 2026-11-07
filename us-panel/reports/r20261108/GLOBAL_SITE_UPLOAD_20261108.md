# kanus.org i18n paketi (release_us) canlı yükleme, 08.11.2026 ~00:45 TSİ

Onay: Demir, 00:40 TSİ ("upload the package to kanus.org's main page; the language picker can apply site-wide, keep the rollback ready").
Yöntem: Hostinger connector hosting_files_generate-upload-url + TUS (override=true), hesap u346856246, kök public_html. Başka hiçbir dosya silinmedi ya da değiştirilmedi.

## Ön kontrol
- sha256sum -c SHA256SUMS: 6/6 OK
- Yüklemeden önce canlı kök index.html: 87b980dec6020cb8071b33aeef441fa80b146352a7761612065f393588841e34 (rollback/index.html ile aynı)
- Yedek: /workspace/us-panel-dev/global-work/site/live_backup_20261108/index.html.live_root (aynı sha)

## Yüklenenler (sırasıyla)
1. public_html/assets/index-uscommerce-i18n-75f56119.js: 250021 B, sha256 75f561190fafb99e1e58ddf567539d8c6f785bda2254221face3b8a10d8b9911 (canlıdan geri okundu, eşleşti)
2. public_html/index.html: 1594 B, sha256 a828fa73880192c4cf75c2b37cd5c648be8054c30108c8e84a2bb6c5e866a926

## Yeni canlı index sha256
a828fa73880192c4cf75c2b37cd5c648be8054c30108c8e84a2bb6c5e866a926 (/, /us, /contact ve /news hepsi 200, hepsi yeni bundle'a bağlı; JS 200)

## VERIFY_AFTER_UPLOAD.sh (exit 0)
```
PASS / references index-uscommerce-i18n-75f56119.js
PASS /us references index-uscommerce-i18n-75f56119.js
PASS /contact references index-uscommerce-i18n-75f56119.js
PASS /news references index-uscommerce-i18n-75f56119.js
PASS asset sha256
PASS asset content-type application/x-javascript
PASS old asset still served (rollback safe)
PASS i18n helpers present
PASS no 'Yerli Teknoloji'
PASS no 'Küresel Vizyon'
PASS no 'Built in Türkiye'
PASS no 'Designed for the World'
Manual: open https://kanus.org/us, switch TR/EN/DE/RU; /news dates localized; failed form submit shows a localized notice.
Rollback: upload rollback/index.html over public_html/index.html (old asset untouched).
ALL PASS
```
(Not: "curl: (23)" satırı, grep -q'nun erken kapanmasından gelen zararsız bir uyarı; ilgili kontrol PASS.)

## Geri dönüş
Yöntem: release_us/rollback/index.html dosyasını aynı TUS yoluyla public_html/index.html üzerine yükle (DEST=index.html, override=true). Ardından canlı sha'nın 87b980de…1e34 olduğunu kontrol et. Eski asset index-uscommerce.js yerinde duruyor; yeni asset kalabilir.
