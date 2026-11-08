# Müşteri Windows EXE — 2026-11-08 05:01 TSİ

Demir t32u/t34u: Store/Mail/Plan şimdilik es geçilerek son müşteri Windows EXE’si üretildi.

## Çıktı

- Yol: `/workspace/us-panel-dev/dist/compile-exe-20261108/US_Ticari_Aday_2026_11_06_12.exe`
- SHA256: `450f7bfcf0d6b8fecd51397a6453e92dee6b684b9c4b4f0b774c53d6a53058ee`
- Boyut: 13810688 bayt
- Biçim: Windows PE (MZ), `GOOS=windows GOARCH=amd64 -ldflags=-H=windowsgui`
- `buildVersion`: `2026.11.06.12` (`.13` yok)
- Eski `.12` hash’i (`f951229b…`) bu dosya değildir.

## Ne eklendi (yalnız çalışma kopyası)

Donmuş r13b–h pinlerine yazılmadı. Çalışma ağacı: `work/term`.

- `licensesrv/compile_skip_types.go` SHA256 `75c075c9e228f7310d01d2ad151a8801cd0070d83ab6672ce078d023884f2d34` — Store/Mail/Plan ve çağrı yerlerinden tamamlanan lisans tipleri
- `compile_skip_model.go` SHA256 `13dc6632f46ba4a185e1d1a781120f6231efa4d317157c526a80a277310ac927` — eksik model.go yüzeyi, çağrı yerlerinden
- `gui.go` / `gui.html` mevcut eski ağaçtan alındı (`gui.html` SHA256 `5de828dd8189798dd697a28d9a3be54bebd8bc1b1ec2414d430fdcb4295e8f1b`)

## Dürüst sınırlar

- Store/Mail/Plan ve model.go bildirimleri kurtarılmış orijinal değildir; çağrı yerlerinden derleme için tamamlandı.
- `submit` model döngüsünü çalıştırmaz; görev gönderilmez.
- Lisans üretimi kapalıdır (çift issuer yok). `US_SMTP_ENABLE` değiştirilmedi.
- Yerelde Windows’ta çalıştırılmadı (NOT TESTED).
- Linux panel yolu değildir; ds-a100 / klein paneli üzerinden test için kullanılamaz.
- Siteye yüklenmedi. Satış / `.13` onayı değildir.
- FINDING-01 ve f01 386 FAIL açık kalır.

Derleme günlüğü (başarılı, boş): SHA256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
