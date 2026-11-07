# US (KAN.US) — Fiyatlandırma yeniden tasarımı (06.11.2026)

Canlı Shopify'a **hiçbir API çağrısı yapılmadı, canlı fiyat değiştirilmedi.** Para birimi TRY (conversation.md: "10.000 TL ve 500 TL olarak görünüyor").

## 1. Kaynak veriler
| Kalem | Değer | Kaynak |
|---|---|---|
| A100 40GB (a2-highgpu-1g), europe-west4, on-demand | 3,748 USD/saat | instances.usage.ai/gcp/compute/a2-highgpu-1g ; devzero.io/instances/gcp/a2-highgpu-1g |
| A100 40GB, europe-west4, spot | 1,740 USD/saat | aynı |
| A100 80GB (a2-ultragpu-1g), europe-west4, on-demand | 5,58 USD/saat (≈4.074 USD/ay) | gcloud-compute.com/a2-ultragpu-1g.html ; tanod.dev/gpu (Billing Catalog, 04.11.2026: 5,536) |
| A100 80GB, europe-west4, spot | 3,07 USD/saat | aynı |
| USD/TRY | 50,05 (TCMB 05.11.2026 döviz satış 50,0489; piyasa 06.11 ≈50,13) | yatirimci.ai/en/forex/usd/history/2026/november ; exa.ai/library/markets/forex/USDTRY |
Bölge: ds-a100 `zone=europe-west4-b` (conversation.md). Resmi sayfa: cloud.google.com/compute/gpus-pricing.

## 2. Maliyet varsayımları (açıkça)
- VM: A100 80GB on-demand, 7/24 açık → **5,58 USD/saat ≈ 279 TL/saat, ≈ 204.000 TL/ay**. (40GB ise 3,75 USD; spot daha ucuz ama kesintili → kredili satışta önerilmez.) Disk/ağ/IP için +%10.
- Model hızı: conversation.md'de ölçülen **5–7 token/sn** üretim; bağlam 8.192 token, yanıt sınırı 2.048 token.
- Tipik istek: ~4–6K token girdi (prefill ~20–30 sn) + ~1.000 token çıktı (~170 sn) → **~3,2 dk GPU**. En kötü: 2.048 çıktı → ~6 dk.
- Tek kullanıcılı (batching yok) kapasite ≈ 19 istem/saat ≈ 13.600 istem/ay (%100 doluluk).
- Doluluk %30 kabul → ayda ~4.100 ücretli istem; sabit maliyet / istem ≈ 4.074×1,1/4.100 ≈ **1,10 USD ≈ 55 TL/istem** (tam dolulukta ≈ 0,33 USD ≈ 17 TL; en kötü uzun istemde ≈ 0,56 USD marjinal).
- Satış fiyatı KDV %20 dahil; ödeme/Shopify komisyonu ≈ %3; iade/fraud ≈ %2.
- Not: panelde tek görev birden çok model çağrısı (=kredi) tüketebilir.

## 3. Önerilen tablo (KDV dahil TL, USD @50,05)
| SKU | Ürün | Fiyat | Birim/istem | ≈USD | Geçerlilik | Brüt marj* |
|---|---|---:|---:|---:|---|---:|
| US-CREDIT-1 | 1 istem | 149 TL | 149,00 | 2,98 | 90 gün | ~%54 |
| US-CREDIT-5 | 5 istem | 695 TL | 139,00 | 13,89 | 90 gün | ~%51 |
| US-CREDIT-10 | 10 istem | 1.290 TL | 129,00 | 25,77 | 90 gün | ~%47 |
| US-CREDIT-25 | 25 istem (yeni) | 2.990 TL | 119,60 | 59,74 | 180 gün | ~%43 |
| US-CREDIT-50 | 50 istem (yeni) | 5.490 TL | 109,80 | 109,69 | 180 gün | ~%38 |
| US-CREDIT-100 | 100 istem (yeni) | 9.900 TL | 99,00 | 197,80 | 365 gün | ~%31 |
| US-MONTHLY-200 | Aylık, 200 istem / 30 gün | 18.900 TL | 94,50 | 377,62 | 30 gün | ~%28 (kullanılmayan kotayla daha yüksek) |
| US-HOUR-1 | 1 saat kullanım | 1.990 TL | (~105/istem @19 istem/sa) | 39,76 | 30 gün | ~%82 saat başı GPU'ya göre |
*Marj = (KDV hariç fiyat − %5 ödeme/iade − 55 TL/istem) / KDV hariç fiyat; %30 doluluk varsayımıyla. Saatlik: 1.658 TL KDV hariç vs 279 TL/sa GPU (ama müşteri GPU'yu tekeline alır; kuyruk etkisi fiyatlanmıştır).
- Hedef: tek istemde ≥%50, en büyük pakette ≥%30 brüt marj. Birim fiyat monoton azalan (testle zorunlu: `licensesrv/pricing_test.go`).
- Giderilen tutarsızlıklar: eski 10'lu paket istem başına 260 TL, 5'li paket 280 TL, tekli 250 TL idi; yani 5'li paket tekliden pahalıydı. Aylık 10.000 TL / 200 = 50 TL/istem ve saatlik 500 TL (≈26 TL/istem) tahmini maliyetin (~55 TL) altındaydı. İstemci "500 istem / 200 USD / 10 USD" yazıyordu, oysa mağazadaki ürün 200 istem.
- Kur riski: fiyatlar TL; USD/TRY %5 artarsa tabloyu güncelleyin (tek yer: `licensesrv/pricing.go`, `USDTRYRef`).

## 4. Kod değişiklikleri
- `licensesrv/pricing.go` (YENİ): tek fiyat tablosu `Catalog`, `LivePlansJSON()`, `TRYText`.
- `pricing.go` (YENİ): panel metni `{{US_FIYAT}}` yer tutucusu katalogdan üretilir; `gui.go` sayfayı `renderGUIPage` ile sunar; `gui.html` sabit "200 USD/500 istem/10 USD" kaldırıldı.
- `license.go`: `mintLicense` katalogdan (200 istem/30 gün, 3600 sn); `licenseDoc.TRY` eklendi (eski `usd` alanı geri uyumluluk için kaldı); "N istem doldu" ve açıklama metni dinamik.
- `cmd/us-license-hook/main.go`: `US_PLANS_JSON` boşsa varsayılan `licensesrv.LivePlansJSON()`.
- `cmd/us-pricing-gen` (YENİ): test mağazası dosyalarını üretir.
- Testler: `license_test.go` güncellendi + `TestPricingTextFromCatalog`; `licensesrv/pricing_test.go` (monotonluk, plan JSON).

### US_PLANS_JSON taslağı (canlı product_id; test mağazasında ID değişir)
```json
{"15424693010660":{"kind":"sub","days":30,"prompts":200,"credits":200},"15424693108964":{"kind":"use","days":30,"seconds":3600},"15428080632036":{"kind":"sub","days":90,"prompts":1,"credits":1},"15428080664804":{"kind":"sub","days":90,"prompts":5,"credits":5},"15428084170980":{"kind":"sub","days":90,"prompts":10,"credits":10}}
```
25/50/100'lük paketlerin canlıda ürünü yok; test mağazasında oluşturulunca ID eklenmeli.

## 5. Test mağazası dosyaları
`reports/TEST_MAGAZA_URUNLER_FIYAT.csv` (Shopify ürün içe aktarma biçimi, status=draft, kargo yok) ve `.json`.

## 6. Derleme
`go build ./...`, `go vet ./...` geçti. `go test ./...`: tümü geçti; tek istisna önceden bilinen `TestOfficialREAInspectsOwnedFile` (box'ta Node 20, ortam sorunu). EXE: `dist/US_Ticari_Aday_2026_11_06_2.exe` (windows/amd64, -H windowsgui -s -w).

## 7. Onay bekleyenler
Canlı fiyat güncellemesi, yeni 3 ürünün canlıda açılması, KDV/fatura durumu, gerçek doluluk ölçümü (ilk ay sonrası fiyat revizyonu).
