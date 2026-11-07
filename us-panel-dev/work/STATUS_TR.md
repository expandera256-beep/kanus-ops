# US Paneli — Ticari Aday Durumu (6 Kasım 2026, 03:40 TSİ)

Kaynak: `source/US_Uzun_Baglam_Kaynak_Taslak.zip` (4 Kasım; daha yeni ZIP yok) → `work/`.
Ağaç: `work/kiyas/term` (Go modülü `kiyas`, ~14.8k satır Go + `gui.html`), `work/kanus/api/license_mint.php` (+test), notlar (`AKIS_DUZELTME_NOTU.txt`, `UZUN_BAGLAM_NOTU.txt`, `PROJE_ONARIM_DURUMU.txt`, `TICARI_ENTEGRASYON_GEREKSINIMLERI.txt`).

## Derleme / test
- Go 1.27.3 (box). `go vet ./...` temiz. ZIP olduğu gibi derleniyordu; derleme hatası yoktu.
- `go test ./...`: tümü geçti, tek istisna `TestOfficialREAInspectsOwnedFile` — box’taki Node 20, `npx rea-agents` için yetersiz (ortam sorunu, kod hatası değil). Windows’ta Node ≥22 ile tekrar denenmeli.
- EXE: `dist/US_Ticari_Aday_2026_11_06_1.exe` (GOOS=windows GOARCH=amd64, `-H windowsgui -s -w`, US_DERLE.bat ile aynı bayraklar). SHA256 `18b5bcf9…5bfbc7`. Gerçek Windows’ta ve canlı modelle denenmedi.
- Not: `uskey` etiketi olmadan derlendi → gömülü lisans sırrı yok; `US_LICENSE_SECRET` ortamda yoksa lisans denetimi kapalı (ticari kip kapalı). Bu bilinçli: sır EXE’ye gömülmemeli.
- Bu ZIP, sohbette anlatılan sonraki Python hesap/kredi sunucusunu (PostgreSQL, Cloud Run, 21 Python testi) İÇERMİYOR. 5 Kasım EXE’sinin (US_Ticari_Aday_2026_11_05_3) kaynağı box’ta yok; bu sürüm ondan geride olabilir. En güncel ZIP’in yüklenmesi gerekiyor.

## Bu turda yapılanlar
1. **“Tamamla” sahte-pozitifi düzeltildi** (`model.go: projectRepairRequest`): “görev tamamlama özelliği olan yeni uygulama oluştur” artık proje onarımı sayılmıyor (önceden boş klasörde dosya bekleyip hiçbir şey üretmiyordu). “Projeyi tamamla ve eksikleri onar”, “mevcut projeyi tamamla” onarım olarak kalıyor. Regresyon testi: `repair_flow_test.go/TestCompletionWordDoesNotMakeNewProjectARepair`.
2. **Yeni `licensesrv` paketi + `cmd/us-license-hook` sunucusu** (yalnız stdlib):
   - `POST /webhooks/shopify/orders-paid`: ham gövde üzerinde `X-Shopify-Hmac-Sha256` doğrulaması (sabit zamanlı), isteğe bağlı `X-Shopify-Shop-Domain` denetimi, yalnız `orders/paid` işlenir, 1 MiB gövde sınırı.
   - İdempotent: `X-Shopify-Webhook-Id` ve sipariş kimliği ikinci kez lisans üretmez.
   - Plan, **product_id** ile eşlenir (`US_PLANS_JSON`), başlık/fiyatla değil. Eşlenmeyen ürün yok sayılır.
   - Lisans anahtarı istemcideki `us1.<gövde>.<imza>` biçimiyle uyumlu (HMAC-SHA256, `US_LICENSE_SECRET`).
   - Kalıcı JSON depo, atomik yazma (geçici dosya + fsync + rename), 0600.
   - E-posta kuyruğu (outbox): üstel geri çekilme, 8 denemede “dead”; SMTP implicit TLS (smtp.hostinger.com:465). **Varsayılan dry-run**: `US_SMTP_ENABLE=1` olmadan e-posta gönderilmez.
   - Sunucu tarafı **kredi rezervasyonu**: Reserve → Settle(commit/release), süresi dolan rezervasyon otomatik iade, yetersiz kredi/iptal/süre dolumu reddi.
   - Testler (4): kötü HMAC reddi, ödeme→lisans→kalıcılık→kuyruk + idempotentlik, kuyruk tekrar/gönderim, kredi rezervasyonu.
   - İkili dosyalar: `dist/us-license-hook_linux`, `dist/us-license-hook_windows.exe`.

## Açık eksikler
| Alan | Durum |
|---|---|
| Shopify webhook | Kod hazır, **kayıtlı değil**. Abonelik yok; HTTPS genel uç nokta yok. `orders/cancelled`, `refunds/create` (iptal/iade → lisans iptali) henüz yok. |
| Lisans e-posta kuyruğu | Kod hazır, dry-run. SMTP kimlik bilgileri, SPF/DKIM/DMARC doğrulaması, gerçek teslim testi yok. |
| Kota modeli | İstemcide sayaç (500 istem / 3600 sn, 30 gün) — kurcalanabilir. Sunucu tarafı kredi rezervasyonu yazıldı ama model kapısına (gateway) **bağlanmadı**; model isteği hâlâ krediden geçmiyor. Fiyat hesap sunucusundan gelmiyor (`licenseDoc.USD` istemcide sabit 200/10). |
| GPU yaşam döngüsü | ds-a100 istem-başı açma / boşta kapatma yok bu ağaçta; `ds-boot` model servisi başarısız (kök neden kayıtlarda yok; KAN tarafında RAID/ağırlık/ds-serve işi sürüyor). İstemci yalnız SSH tüneli (1923→8080) kuruyor. |
| Sahte başarı | “tamamla” sınıflandırması düzeltildi. Kalan riskler: model yokken “hazır” göstergesi, test çalıştırılmadan “tamamlandı” — ayrıca taranmalı. |
| Hesap/oturum | E-posta kodu ile giriş, sunucu oturumu, cihaz yönetimi yok. |
| Doğrulama | Gerçek Windows + canlı model + Shopify test siparişi uçtan uca yapılmadı. |
