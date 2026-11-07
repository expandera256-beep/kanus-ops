# Birim ekonomi — GPU oturumu, Shopify ücretleri, paket marjı

Güncelleme: 06.11.2026 08:24 TSİ (Europe/Istanbul)

PASS yalnız dosya/fiyat kaynağı okumasıdır; canlı fatura veya canlı GCP/Shopify tahsilatı değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.
Ölçülen süreler için `KANUS_GPU_REPORT.md` beklenir; şu an ölçüm yok → **KAN ölçümü bekleniyor**.

Bağlı raporlar: [COST.md](./COST.md) · [KANUS_COST_REPORT.md](./KANUS_COST_REPORT.md) · [FIYATLANDIRMA_TR.md](./FIYATLANDIRMA_TR.md)

---

## 1. GPU saatlik fiyat (a2-highgpu-1g, europe-west4)

**Resmi doğrulama (06.11.2026 07:19 TSİ): PASS** — Cloud Billing Catalog API `services/6F81-5844-456A` (Compute Engine), service account ile `GET .../skus` (public anahtarsız çağrı 403). Dokümantasyon: [Get publicly available pricing](https://docs.cloud.google.com/billing/v1/how-tos/catalog-api). HTML fiyat sayfaları WebFetch’te zaman aşımına düştü; SKU birim fiyatları API’den alındı.

| Bileşen | SKU ID | Açıklama (Catalog) | Birim fiyat | a2-highgpu-1g payı |
|---|---|---|---:|---:|
| GPU | `FAA1-17BD-3A62` | Nvidia Tesla A100 GPU running in Netherlands | **2,933908 USD/saat** | ×1 = 2,933908 |
| vCPU | `896E-DB80-1131` | A2 Instance Core running in Netherlands | **0,034802 USD/saat** | ×12 = 0,417624 |
| RAM | `A78A-A063-7D5A` | A2 Instance Ram running in Netherlands | **0,004664 USD/GiBy·saat** | ×85 = 0,396440 |
| **Toplam on-demand** | — | `a2-highgpu-1g` europe-west4 | — | **3,747972 ≈ 3,748 USD/saat** |

Çapraz kontrol (özet siteler, artık yalnızca uyum): tanod.dev/gpu, instances.usage.ai, devzero → 3,748 USD/sa. **Beklenen ds-a100 oranı: Spot 1,848238 USD/sa** (aşağıdaki Spot tablosu + JSON/§4a); on-demand 3,748 **üst sınır**.

Resmi ürün sayfaları (makine ailesi; bölgesel SKU HTML’de sabit değil): [GPU pricing](https://cloud.google.com/compute/gpus-pricing) · [VM instance pricing](https://cloud.google.com/compute/vm-instance-pricing).

Formül:

```text
saatlik_USD = GPU + (vCPU_birim × 12) + (RAM_birim × 85)
            = 2,933908 + 0,417624 + 0,396440 = 3,747972 ≈ 3,748
oturum_maliyeti_USD = 3,748 × (dakika / 60)
oturum_maliyeti_TRY = oturum_maliyeti_USD × USDTRY
istek_başı_USD     = oturum_maliyeti_USD / oturumdaki_ücretli_istek_sayısı
```

Canlı proje faturası / fatura CSV teyidi hâlâ **NOT TESTED** (Catalog list fiyatı ≠ indirimli fatura).

**Not:** Billing BigQuery export **PASS** — **KAN.US bildirimi 08:21 TSİ** (standard + detailed, dataset `billing_export`, proje `project-27082a40-6083-4952-9db`). **İlk veri ≤24 sa bekleniyor.**

### ds-a100 beklenen Spot oranı (Catalog doğrulandı)

Kaynak: `sre/billing_catalog_ds-a100_20261106.json` (sorgu 06.11.2026 **08:20 TSİ**) + `SRE_WATCHDOG_DESIGN_20261106.md` §4a. effectiveTime 2026-11-05T08:00:00Z (= **05.11 11:00 TSİ**).

| Bileşen | SKU | Birim | Adet | USD/sa |
|---|---|---:|---:|---:|
| Spot A2 vCPU | E540-EA28-2318 | 0,01616 /h | 12 | 0,193920 |
| Spot A2 RAM | D847-3FF3-B717 | 0,002165 /GiBy.h | 85 | 0,184025 |
| Spot A100 | 9BBC-6855-1C9A | 1,3618 /h | 1 | 1,361800 |
| Spot Local SSD | 8D0B-EA58-9E35 | 0,0528 /GiBy.mo | 1500 GiB → ÷730 | 0,108493 |
| **Toplam Spot (local SSD dahil)** | | | | **1,848238** |

Hariç: 50 GB boot PD (dururken de ücret; FIXED_COSTS), dış IP, egress. **Spot preempt edilebilir; fiyat değişebilir.** On-demand **3,748 USD/sa** üst sınır / karşılaştırma olarak korunur.

Export hazır olunca Spot/on-demand list fiyatı fatura satırıyla çaprazlanacak.

---

## 2. Kur (USD/TRY)

| Değer | Kaynak / tarih |
|---|---|
| **50,05 TRY/USD** (hesaplarda kullanılan) | TCMB döviz satış 05.11.2026 ≈ 50,0489 — [yatirimci.ai … november 2026](https://yatirimci.ai/en/forex/usd/history/2026/november); FIYATLANDIRMA_TR.md ile aynı |
| Piyasa 06.11.2026 ≈ 50,16 | [ozmihenk.com](https://ozmihenk.com/en/) (06.11.2026 ~04:00 TSİ) — hassasiyet için not; tablolar 50,05 ile |

---

## 3. Soğuk açılış + idle senaryoları (faturalanan dakika)

**Spot sütunları:** §10 (oturum TRY Spot soğuk 66,29 / SSD 29,09). On-demand aşağıdaki tabloda üst sınır.

Inference / TTFT / gerçek cold-start süresi: **KAN ölçümü bekleniyor** (`KANUS_GPU_REPORT.md` — 06.11.2026 06:10 sürümünde ölçüm yok; yalnız yerel policy/fake test PASS).

| Senaryo | Faturalanan | USD | TRY @50,05 |
|---|---:|---:|---:|
| Soğuk açılış 20 dk | 20 dk | 1,249 | 62,53 |
| Soğuk açılış 30 dk | 30 dk | 1,874 | 93,79 |
| 20 dk soğuk + 15 dk idle | 35 dk | 2,186 | 109,43 |
| 30 dk soğuk + 15 dk idle | 45 dk | 2,811 | 140,69 |
| Inference süresi (tek istek) | — | **KAN ölçümü bekleniyor** | **KAN ölçümü bekleniyor** |
| Gerçek cold-start / TTFT | — | **KAN ölçümü bekleniyor** | **KAN ölçümü bekleniyor** |

Disk / dış IP / egress / Persistent Disk bu tabloya **dahil değil** (NOT TESTED; FIYATLANDIRMA’da +%10 tampon önerisi var).

---

## 4. İstek başı GPU maliyeti (oturum N isteğe paylaştırılırsa)

`istek_başı = oturum_maliyeti / N`. N = bir GPU oturumunda faturalanan ücretli istek sayısı (ölçüm yok → senaryo tablosu).

| N istek / oturum | 20 dk soğuk | 30 dk soğuk | 20+15 idle | 30+15 idle |
|---:|---:|---:|---:|---:|
| 1 | 62,53 TL | 93,79 TL | 109,43 TL | 140,69 TL |
| 5 | 12,51 TL | 18,76 TL | 21,89 TL | 28,14 TL |
| 10 | 6,25 TL | 9,38 TL | 10,94 TL | 14,07 TL |
| 25 | 2,50 TL | 3,75 TL | 4,38 TL | 5,63 TL |
| 50 | 1,25 TL | 1,88 TL | 2,19 TL | 2,81 TL |
| 100 | 0,63 TL | 0,94 TL | 1,09 TL | 1,41 TL |

Karşılaştırma: FIYATLANDIRMA_TR.md’deki ~55 TL/istem, **7/24 açık A100 80GB + %30 doluluk** dağıtımıdır; bu tablo **istem-başı cold-start oturum amortismanıdır** — aynı şey değildir.

---

## 5. Shopify işlem ücretleri

Kaynaklar: [shopify.com/pricing](https://www.shopify.com/pricing), [Third-party transaction fees](https://help.shopify.com/en/manual/your-account/manage-billing/billing-charges/types-of-charges/third-party-charges/third-party-transaction-fees), [Shopify Payments fees](https://help.shopify.com/en/manual/payments/shopify-payments/onboarding/cost-of-shopify-payments), [Supported countries](https://help.shopify.com/en/manual/payments/shopify-payments/supported-countries).

| Durum | Ücret |
|---|---|
| Shopify Payments (ABD Basic örnek) | Online kart ~%2,9 + 0,30 USD; üçüncü taraf işlem ücreti **yok** |
| Üçüncü taraf ağ geçidi + Shopify işlem ücreti | Basic **%2**, Grow **%1**, Advanced **%0,6**, Plus **%0,2** — ağ geçidinin kendi komisyonuna **ek** |
| **Türkiye uyarısı** | Shopify Payments desteklenen ülkeler listesinde **Türkiye yok** → TR işletmesi tipik olarak üçüncü taraf (ör. iyzico) + Shopify third-party fee kullanır. Mağaza ülkesi/planı bu turda admin’den **NOT TESTED**. |

Bu rapordaki sayısal örnekler: **Basic %2 Shopify 3P**; gateway komisyonu ayrı satırda **varsayım %2,9 + sabit** (gerçek iyzico/PayTR sözleşmesi **NOT TESTED** / **KAN ölçümü bekleniyor**).

Formül (Shopify 3P):

```text
shopify_3p = (ürün − indirim + vergi + kargo) × plan_oranı
```

Bogus / test siparişlerinde Shopify transaction fee uygulanmaz (yardım merkezi: test / paid-marked draft).

---

## 6. Katalog paket fiyatları ve brüt marj

Kaynak fiyatlar (KDV dahil TRY): FIYATLANDIRMA_TR.md / önceki COST — **149 / 695 / 1.290 / 2.990 / 5.490 / 9.900 TL** (1/5/10/25/50/100 istem). Canlı Shopify fiyatı bu turda değiştirilmedi (**PASS** dosya kanıtı).

```text
net_KDV_hariç     = fiyat_KDV_dahil / 1,20
shopify_3p_TL     = fiyat_KDV_dahil × 0,02          # Basic varsayımı
gateway_TL        = fiyat_KDV_dahil × gateway_oran   # varsayım 0,029; sözleşme yoksa KAN ölçümü bekleniyor
ödeme_toplam      = shopify_3p_TL + gateway_TL
GPU_paket         = istek_başı_GPU × paket_istem_sayısı   # seçilen senaryo
katkı_KDV_hariç   = net_KDV_hariç − ödeme_toplam − GPU_paket
brüt_marj_%       = katkı_KDV_hariç / net_KDV_hariç × 100
```

### 6.1 Ödeme sonrası nakit (Shopify %2; gateway hariç)

| Paket | Fiyat (KDV dahil) | Net /1,20 | Shopify %2 | Kalan nakit |
|---|---:|---:|---:|---:|
| 1 | 149 | 124,17 | 2,98 | 146,02 |
| 5 | 695 | 579,17 | 13,90 | 681,10 |
| 10 | 1.290 | 1.075,00 | 25,80 | 1.264,20 |
| 25 | 2.990 | 2.491,67 | 59,80 | 2.930,20 |
| 50 | 5.490 | 4.575,00 | 109,80 | 5.380,20 |
| 100 | 9.900 | 8.250,00 | 198,00 | 9.702,00 |

### 6.2 Örnek brüt marj — senaryo «30 dk soğuk + 15 dk idle», oturumda N= paket boyutu kadar istek (kötümser: tek satış = tek oturum)

GPU paket maliyeti = (2,811 USD × 50,05) = **140,69 TL** oturum; N=paket boyutu ise istek başı = 140,69/N, paket GPU = 140,69 (tüm oturum tek pakete yazılır).

| Paket | Net | Ödeme (%2+%2,9≈%4,9) | GPU (kötümser oturum) | Katkı | Brüt marj |
|---|---:|---:|---:|---:|---:|
| 1 | 124,17 | 7,30 | 140,69 | −23,82 | **negatif** |
| 5 | 579,17 | 34,06 | 140,69 | 404,42 | ~%70 |
| 10 | 1.075,00 | 63,21 | 140,69 | 871,10 | ~%81 |
| 25 | 2.491,67 | 146,51 | 140,69 | 2.204,47 | ~%88 |
| 50 | 4.575,00 | 269,01 | 140,69 | 4.165,30 | ~%91 |
| 100 | 8.250,00 | 485,10 | 140,69 | 7.624,21 | ~%92 |

**Yorum:** Tek istemlik pakette cold-start’ı tek satışa yazmak marjı öldürür; 5+ paketlerde veya oturumda birden fazla istek amorti edilirse marj düzelir. Gerçek N ve inference: **KAN ölçümü bekleniyor**.

### 6.3 İyimser amortisman — aynı 45 dk oturumda 25 ücretli istek (istek başı 5,63 TL)

Ödeme yine %4,9 varsayımı; GPU_paket = 5,63 × istem.

| Paket | GPU_paket | Katkı (net−ödeme−GPU) | Brüt marj |
|---|---:|---:|---:|
| 1 | 5,63 | 111,24 | ~%90 |
| 5 | 28,14 | 516,97 | ~%89 |
| 10 | 56,28 | 955,51 | ~%89 |
| 25 | 140,69 | 2.204,47 | ~%88 |
| 50 | 281,38 | 4.024,61 | ~%88 |
| 100 | 562,75 | 7.202,15 | ~%87 |

---

## 7. Başa baş tabloları

### 7.1 GPU oturum maliyetini karşılamak için gereken paket satışı

`gerekli_satış = oturum_TRY / (fiyat × (1 − 0,02))` (yalnız Shopify %2; gateway ve KDV nakitte ayrı muhasebe).

| Paket \ Senaryo | 20 dk | 30 dk | 20+15 | 30+15 |
|---|---:|---:|---:|---:|
| 1 (149 TL) | 0,43 | 0,64 | 0,75 | 0,96 |
| 5 (695) | 0,09 | 0,14 | 0,16 | 0,21 |
| 10 (1.290) | 0,05 | 0,07 | 0,09 | 0,11 |
| 25 (2.990) | 0,02 | 0,03 | 0,04 | 0,05 |
| 50 (5.490) | 0,01 | 0,02 | 0,02 | 0,03 |
| 100 (9.900) | 0,01 | 0,01 | 0,01 | 0,01 |

Yorum: Tek paket satışı (5+), cold+idle GPU oturumunu nakit olarak kolayca karşılar; asıl risk **tek istem + soğuk açılış** ve **düşük doluluk / 7/24 açık VM**.

### 7.2 Sabit aylık maliyet başa baş (bilgi; canlı fatura NOT TESTED)

7/24 `a2-highgpu-1g` on-demand ≈ 3,748 × 730 ≈ **2.736 USD/ay** ≈ **136.900 TL/ay** @50,05 (disk/IP hariç).

```text
başa_baş_adet_paket_P = aylık_GPU_TRY / (fiyat_P × (1 − shopify_oran) / 1,20 − …)
```

Örnek yalnız GPU + Shopify %2, KDV hariç katkı ≈ net − %2×fiyat (gateway=0):

| Paket | Yaklaşık aylık satış adedi (yalnız GPU karşılamak) |
|---|---:|
| 25 (2.990) | ≈ 136.900 / (2.491,67 − 59,80) ≈ **56** |
| 50 (5.490) | ≈ 136.900 / (4.575 − 109,80) ≈ **31** |
| 100 (9.900) | ≈ 136.900 / (8.250 − 198) ≈ **17** |

Idle-stop çalışırsa aylık sabit düşer; Go proxy→GCP start/stop hâlâ **BLOCKED** (KANUS_GPU_REPORT / FINAL_STATUS).

---

## 8. Durum özeti

| Madde | Durum |
|---|---|
| a2-highgpu-1g europe-west4 3,748 USD/sa | **PASS — resmi Catalog API** (SKU FAA1/896E/A78A; GPU+vCPU+RAM) |
| Canlı GCP fatura CSV / indirimli fatura teyidi | NOT TESTED |
| Cold 20/30 + idle 15 senaryo matematiği | PASS (formül); süreler varsayım |
| Gerçek cold-start / inference / TTFT | **KAN ölçümü bekleniyor** |
| Shopify plan + TR gateway sözleşmesi | NOT TESTED |
| Katalog 2990/5490/9900 (ve 149/695/1290) | PASS (FIYATLANDIRMA dosya) |
| Go idle-stop ile gerçek tasarruf | BLOCKED |

---

## 9. Öneri — tek istem paketinde zarar riski (**yalnız teklif; canlı fiyat değişmedi**)

Bağlam: §6.2 kötümser senaryo (30 dk soğuk + 15 dk idle = **140,69 TL** GPU, ödeme ≈%4,9). Tek satış **149 TL** paketine tüm oturum yazılırsa:

```text
net = 149/1,20 = 124,17
ödeme ≈ 149×0,049 = 7,30
katkı = 124,17 − 7,30 − 140,69 = −23,82 TL  → brüt marj ≈ −%19
```

### Seçenek A — Minimum paket boyutu (5 istem / 695 TL)

Canlı katalogda 5’li paket zaten var; tek-istem SKU’yu satışı kapatmak veya «önerilmez» işaretlemek teklifi.

```text
net = 695/1,20 = 579,17
ödeme ≈ 695×0,049 = 34,06
katkı = 579,17 − 34,06 − 140,69 = 404,42 TL  → brüt marj ≈ %70
```

| Etki | Değer |
|---|---|
| Tek-istem marjı (kötümser) | −%19 → **kaldırılır / satılmaz** |
| 5’li paket, 1 oturum/satış | **~%70** |
| Operasyon | Fiyat API’siz; mağaza/ürün görünürlüğü (Demir onayı) |

### Seçenek B — Warm window / istek biriktirme (GPU açıkken N istek)

Idle-stop 15 dk içinde aynı oturuma birden fazla ücretli istek almak; GPU maliyetini N’e böl.

```text
GPU_istek = 140,69 / N
katkı_1 = 124,17 − 7,30 − (140,69/N)
marj_1  = katkı_1 / 124,17
```

| N (oturumdaki ücretli istek) | GPU/istek | Katkı (149 TL paket) | Brüt marj |
|---:|---:|---:|---:|
| 1 (bugünkü risk) | 140,69 | −23,82 | **−%19** |
| 5 | 28,14 | 88,73 | **~%71** |
| 10 | 14,07 | 102,80 | **~%83** |

Önkoşul: Go→GCP start/idle-stop **BLOCKED** değil; kuyruk/capacity=1 ile batch mümkün. Ölçülen N: **KAN ölçümü bekleniyor**.

### Seçenek C — Tek istem fiyat artışı (teklif; uygulanmadı)

Başa baş (katkı≥0, %4,9 ödeme, 140,69 TL GPU):

```text
fiyat × (1/1,20 − 0,049) ≥ 140,69
fiyat × 0,784333 ≥ 140,69  →  fiyat ≥ ≈179 TL
```

Hedef ~%50 brüt (aynı kötümser oturum):

```text
# 0,5×(fiyat/1,20) = fiyat/1,20 − 0,049×fiyat − 140,69
fiyat × (0,416667 − 0,049) = 140,69  → fiyat ≥ ≈383 TL (~390 TL yuvarlak)
```

| Teklif fiyat | Tam oturum marjı | N=5 warm marjı |
|---:|---:|---:|
| 149 (mevcut) | −%19 | ~%71 |
| 249 (ara zam teklifi) | ~%26 | ~%81 |
| 390 (~%50 hedef) | ~%51 | (daha yüksek) |

**Canlı Shopify fiyatı bu turda değiştirilmedi.** Seçim Demir + KAN.US onayı ister.

### Kısa karşılaştırma

| Seçenek | Marj etkisi (kötümser 45 dk) | Bağımlılık |
|---|---|---|
| A Min paket 5 | Tek-istem riski yok; 5’li ~%70 | Ürün politikası |
| B Warm N≥5 | 149 TL ile ~%71+ | GPU lifecycle PASS |
| C Zam ≥179 / ~390 | Başa baş / ~%50 | Fiyat onayı |



## 10. Kâr modeli — on-demand üst sınır + Spot beklenen (min paket 5)

Güncelleme: 06.11.2026 08:24 TSİ (Europe/Istanbul). Min paket **5 / 695 TL**. On-demand üst sınır **3,748 USD/sa**. Beklenen ds-a100: **Spot 1,848238 USD/sa** (local SSD dahil; JSON+§4a **doğrulandı**). Idle **+15 dk**. Cold ~28 dk / SSD ~3:52: **SRE (`KANUS_GPU_REPORT`), KAN teyidi bekleniyor**. Spot preempt riski var; fiyat değişebilir.

```text
katkı_istek = 109,02 TL
oturum_TRY(rate) = rate × USDTRY × dakika/60
soğuk 43 dk:  on-demand 134,44 TL · Spot 66,29 TL
SSD   ≈18,87 dk: on-demand 58,99 TL · Spot 29,09 TL
kar_1oturum = N×109,02 − oturum_TRY
```

### Oturum maliyetleri

| Senaryo | dk | On-demand TRY | Spot TRY |
|---|---:|---:|---:|
| Soğuk ~28 + 15 idle | 43 | **134,44** | **66,29** |
| SSD ~3:52 + 15 idle | ≈18,87 | **58,99** | **29,09** |

### Günlük kâr — 1 uyandırma / gün

| N | Katkı | OD soğuk | Spot soğuk | OD SSD | Spot SSD |
|---:|---:|---:|---:|---:|---:|
| 10 | 1.090 | **956** | **1.024** | **1.031** | **1.061** |
| 50 | 5.451 | **5.317** | **5.385** | **5.392** | **5.422** |
| 200 | 21.804 | **21.670** | **21.738** | **21.745** | **21.775** |

(Değerler yuvarlak TL.) Boot 50 GB PD ≈5 USD/ay sabit (`FIXED_COSTS`); oturum tablosuna dahil değil. Local SSD **çalışırken** Spot satırında; korunmuş SSD saklama ücreti DOĞRULANMADI / boş.

---

## 11. Karar (paket) (06.11.2026 07:22 TSİ) — Demir: **Seçenek A seçildi**
- Minimum paket **5 istem (695 TL)**; 1 istemlik 149 TL paket satıştan kaldırılacak. Beklenen kötümser marj ~%70.
- Durum: **KARAR VERİLDİ / UYGULANMADI** — canlı Shopify ürünü/varyantı ve `US_CREDIT_PRODUCTS` allowlist henüz değişmedi (KAN.US webhook/hook işiyle koordinasyon gerekli).
- **Güncelleme 07:26 TSİ:** Demir onayıyla SHOP, shop.kanus.org'daki `Kanus AI — 1 Model İsteği` (gid://shopify/Product/15428080632036) ürününü **DRAFT** yaptı (isPublished=false; silinmedi, fiyat 149 TRY aynı). kanus-test, `US_CREDIT_PRODUCTS` ve plans.json değişmedi (eski siparişlerin iade/iptal eşlemesi için). Kaynak: SHOP raporu, KANUS_SHOPIFY_LIVE_AUDIT.md.

