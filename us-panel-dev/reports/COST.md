# Maliyet — .5

Güncelleme: 06.11.2026 07:19 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

Birim ekonomi ayrıntısı: **[UNIT_ECONOMICS.md](./UNIT_ECONOMICS.md)** (GPU oturum senaryoları, Shopify ücretleri, paket marjı, başa baş, tek-istem zarar önerileri).

- PASS — **Resmi** Cloud Billing Catalog API (`6F81-5844-456A`): europe-west4 A100 GPU 2,933908 + A2 core 0,034802×12 + A2 RAM 0,004664×85 = **3,747972 ≈ 3,748 USD/saat** (`a2-highgpu-1g`). SKU: `FAA1-17BD-3A62`, `896E-DB80-1131`, `A78A-A063-7D5A`. Docs: https://docs.cloud.google.com/billing/v1/how-tos/catalog-api . Canlı fatura CSV **NOT TESTED**.
- PASS — FIYATLANDIRMA_TR.md: A100 80GB (a2-ultragpu) ≈5,58 USD/saat dağıtım varsayımı (~55 TL/istem) dosya kaynağı; bu 40GB on-demand oturum modelinden ayrıdır.
- PASS — Oturum senaryo matematiği (20/30 dk soğuk; +15 dk idle): 1,249 / 1,874 / 2,186 / 2,811 USD ≈ 62,53 / 93,79 / 109,43 / 140,69 TL @ USDTRY 50,05. Resmi saatlik değişmediği için tablolar yeniden hesaplanmadı.
- NOT TESTED — Gerçek cold-start / inference / TTFT: **KAN ölçümü bekleniyor** (KANUS_GPU_REPORT.md).
- PASS — Katalog 149/695/1.290/2.990/5.490/9.900 TL; canlı fiyat bu turda değiştirilmedi. Tek-istem zarar önerileri UNIT_ECONOMICS §9 (teklif).
- PASS — kanus-test E2E gerçek Shopify teslimatları dosya kanıtı (maliyet kabulü değil).
- BLOCKED — Go GPU start/idle-stop yokken istem-başı tasarruf iddiası yok. Round 5 proxy canlı; upstream GPU BLOCKED.
