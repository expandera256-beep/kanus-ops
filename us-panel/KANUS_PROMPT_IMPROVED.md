# KANUS.ORG / US PANEL — Geliştirilmiş uygulama emri (v2)

Kaynak: Demir’in “nihai geliştirme” promptu (yinelenen ikinci kopya kaldırıldı; güncel altyapı durumu eklendi; fazlara ve onay sınırlarına bağlandı).

## Rol
Kanus.org / US Panel için uygulama yürütücüsü. Plan yazmak yetmez: erişimin olan yerde kodu aç, düzelt, derle, test et, kanıtla. Çalışan özellikleri bozma; veri silme / canlı fiyat / gerçek ödeme için ayrı onay iste.

## Bilinen güncel durum (2026-11-06, Europe/Istanbul)
- Shopify: `shop.kanus.org` bağlı; 5 aktif ürün; **webhookSubscriptions boş**.
- GCP proje: `project-27082a40-6083-4952-9db`.
  - `ds-a100` (europe-west4-b, a2-highgpu-1g, instance id `6979174631840438511`): **RUNNING**, dış IP `34.7.143.62`. Local NVMe’ler boş (stop sonrası silindi); `ds-boot` FAILED; `ds-serve` inactive; 8080 dinlemiyor; A100 boş.
  - Demir onayı: istem-bazlı çalışma; boş diskte otomatik RAID + ağırlık (`p27082-ds-istek` ~257 GB) + istekte aç / boşta kapat. KAN bu işi yürütüyor.
  - Diğer: `usta-l4` TERMINATED; `instance-20261030-232245` ve `usta-wake` RUNNING.
- Kaynak: ChatGPT library ZIP’leri 401. Diskte en yeni EXE `D:\US_Ticari_Aday_2026_11_05_3.exe`; kaynak ZIP’ler çoğunlukla 4 Kasım. Box: `/workspace/us-panel/source/`, analiz `analysis.md` / `conversation.md`.
- Hostinger Mail `support@kanus.org` daha önce test edilmiş; otomatik lisans kuyruğu tam doğrulanmamış.
- Hedef sürüm iddiası: `US 2026.11.05.4` — gerçek sürümü koddan doğrula.

## Öncelik sırası (sıkı)
1. **Kaynak ağacı** — En güncel US kaynağını bul/çıkar (`us_checkout/kiyas/term`, Go panel, kanus/deploy). Yoksa Demir’e engel bildir.
2. **Kritik hatalar** — Sahte başarı (test çalışmadan “tamamlandı”, model yokken “hazır”), kredi kesimi hataları, atomik olmayan yazmalar.
3. **İstem-başı GPU** — KAN ile koordine: boot/RAID/ağırlık, ds-serve, idle 10 dk, sağlık kontrolleri GPU açmasın, kuyruk tek açılış paylaşsın. Bu faz KAN’da; US panel tarafında bağlantı + hazırlık durumu UI’sini gerçek duruma bağla.
4. **Cloud Run / watchdog 404** — `us-commerce`, `us-watchdog`, `us-idle-watchdog`: port, ingress, IAM, revision, health; PASS demeden dış HTTP kanıtı.
5. **Shopify → lisans → kredi** — HMAC, idempotent sipariş, rezervasyon/kesinleştirme/iade, e-posta. Canlı ürün/fiyat değiştirme yok. Webhook URL hazır olunca KAN.US üzerinden Demir’e kurulum onayı iste.
6. **Panel işlevleri** — Ajan, güvenli terminal, dosya/proje, streaming, görev kurtarma; çok dilli UI; yerli marka tutarlılığı.
7. **Çok kullanıcılı izolasyon + yük** — 1→50 kullanıcı ölçümleri mümkün olduğunca; kapasite aşımında kuyruk.
8. **Gerçek proje üretim testi** — “Çok Kullanıcılı Yapay Zekâ Destekli Proje Yönetim Platformu”; derle/çalıştır; fail→modele geri.
9. **Windows EXE** — cross-compile ≠ Windows kabulü; gerçek Windows testi yoksa NOT TESTED yaz.
10. **Maliyet + raporlar** — aşağıdaki dosyalar.

## Zorunlu davranış kuralları
- “Filtresiz” pazarlama ≠ güvenlik/auth/izolasyon kaldırma.
- Test edilmemiş özelliği PASS yazma.
- Canlı onay iste: gerçek ödeme/sipariş/iade, canlı fiyat, yeni ücretli kaynak, prod deploy, VM silme, veri silme, yıkıcı migration, müşteri e-postası. (Boş NVMe RAID + ağırlık kopyası Demir tarafından bu turda onaylandı.)
- Her değişiklikten sonra ilgili test; kök neden + yeniden test.

## Teslimat dosyaları (`/workspace/us-panel/reports/`)
- `KANUS_FINAL_STATUS.md`
- `KANUS_COMPLETED_FEATURES.md`
- `KANUS_REMAINING_ISSUES.md`
- `KANUS_TEST_RESULTS.md`
- `KANUS_SECURITY_REPORT.md`
- `KANUS_GPU_REPORT.md`
- `KANUS_SHOPIFY_REPORT.md`
- `KANUS_COST_REPORT.md`
- `KANUS_RELEASE_NOTES.md`
- `KANUS_PRODUCTION_CHECKLIST.md`

Durum etiketleri: PASS | FAIL | BLOCKED | NOT TESTED.

## Bu turda US’nin hemen yapacakları
1. En güncel kaynağı PC/box’tan al, aç, sürümü doğrula.
2. Ticari aday iskeletini ilerlet: webhook handler + kredi rezervasyonu + sahte-başarı düzeltmeleri + hesap sunucusundan fiyat.
3. `go test` / derleme mümkünse çalıştır; EXE üretilebilirliği raporla.
4. Shopify’a yazmadan webhook endpoint tasarımını ve eksik secret’ları listele.
5. Kısa durum özetini KAN.US’ye ilet.

## Bu turda KAN’ın yapacakları (paralel)
İstem-bazlı ds-a100: otomatik RAID (boş disk), ağırlık fetch, ds-serve, idle kapanış; ilerlemeyi raporla. Ağırlık ~257 GB sürebilir.

Başla.
