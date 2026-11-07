# Maliyet raporu — senaryo tablosu (06.11.2026)

Güncelleme: 06.11.2026 09:13 TSİ (Europe/Istanbul)

Salt rapor; canlı değişiklik yok. Bilinmeyenler **NOT KNOWN** (tahmin yok).

## Kaynaklar
- `SRE_WATCHDOG_DESIGN_20261106.md` §4a–4c (Spot **1,848238 USD/sa** çalışırken local SSD dahil; korunmuş SSD dururken **66–132 USD/ay** senaryo bandı, kesin SKU **NOT VERIFIED** / Billing export bekleniyor)
- `sre/billing_catalog_ds-a100_20261106.json`, `KANUS_GPU_REPORT.md` (soğuk ~28 dk; SSD açılış ~3:52)
- `FIXED_COSTS.md` bilinen GCP ara toplam ≈**15,53 USD/ay**
- Prod VPS: Hostinger MCP `vps_virtual-machines_list` + `billing_subscriptions_list` — **srv2133566** / 82.29.184.225, plan **KVM 2**, abonelik `169lkvVaARjZU5YOD`, **24,49 USD/ay** (2449 sent)
- Katalog paketler: min **5 / 695 TL**; 1290 / 2990 / 5490 / 9900 (KDV dahil). `UNIT_ECONOMICS.md` / FIYATLANDIRMA
- USDTRY **50,05** (`FIXED_COSTS.md`)

## Varsayımlar (açık)
1. **Günlük GPU saati** = belirtilen üretken/çalışma saati (1 / 4 / 12 sa/gün × **30 gün**).
2. **Korunmuş SSD (discardLocalSsd=false):** dururken SSD için SRE §4c bandı **+66 … +132 USD/ay** low/high (VM sürekli kapalı varsayımına dayalı senaryo; kısmi çalışma + koruma için kesin tutar **NOT KNOWN**, export sonra). Bu bant **1,848238 içine dahil değildir**.
3. **SSD yok (discard):** soğuk açılış **~28 dk / uyandırma**. Uyandırma varsayımı: **günde 1 uyandırma** (günlük saat tek blokta) → ayda **30 × 28 dk = 14.0 sa** ek faturalanan süre. Birden fazla uyandırma/gün → **NOT KNOWN** (bu tabloda yok).
4. Spot **birincil**; on-demand **3,748 USD/sa üst sınır**. Spot preempt edilebilir; fiyat değişebilir.
5. Sabit taban bu raporda: GCP bilinen **15,53** + prod VPS **24,49** = **40,02 USD/ay**. Test VPS srv1662550 da aktif KVM 2 (**24,49 USD/ay**) — senaryo tabanına **dahil değil** (ayrı satır).
6. Gateway komisyonu, Shopify plan aboneliği, Mail/domain yenileme: maliyet senaryosunda **NOT KNOWN** / sabit tabanda yok (VPS hariç).

## 1. Aylık maliyet (USD) — Spot birincil

| Günlük GPU | GPU sa/ay | GPU Spot | GPU OD (üst) | Korunmuş SSD Spot toplam (low–high) | Korunmuş OD toplam (low–high) | Discard Spot (+soğuk) | Discard OD (+soğuk) |
|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 sa | 30 | 55.45 | 112.44 | **161.47–227.47** | 218.46–284.46 | **121.34** | 204.93 |
| 4 sa | 120 | 221.79 | 449.76 | **327.81–393.81** | 555.78–621.78 | **287.68** | 542.25 |
| 12 sa | 360 | 665.37 | 1349.28 | **771.39–837.39** | 1455.30–1521.30 | **731.26** | 1441.77 |

Taban sabit (GCP 15,53 + prod VPS 24,49) her hücrede dahil. Korunmuş: +66 / +132. Discard: +14 sa soğuk × Spot/OD.

### Aynı tablo — TRY (@50,05)

| Günlük GPU | Korunmuş Spot TRY (low–high) | Discard Spot TRY | Korunmuş OD TRY (low–high) | Discard OD TRY |
|---:|---:|---:|---:|---:|
| 1 sa | **8,081–11,385** | **6,073** | 10,934–14,237 | 10,257 |
| 4 sa | **16,407–19,710** | **14,399** | 27,817–31,120 | 27,140 |
| 12 sa | **38,608–41,911** | **36,600** | 72,838–76,141 | 72,161 |

### Marj notu
Satış adedi bilinmeden yüzde marj hesaplanmaz. Aşağıdaki **başa baş** satırında marj ≈ %0 (sabit+GPU karşılanır). Her ek paket satışı, gateway **NOT KNOWN** olduğu için yalnızca Shopify %2 + KDV net katkısı kadar pozitif marj ekler (`katkı` aşağıda).

## 2. Başa baş — aylık paket satışı

### Formüller
```text
net_KDV_hariç(paket)     = fiyat_KDV_dahil / 1,20
shopify_3p(paket)        = fiyat_KDV_dahil × 0,02     # Basic 3P varsayımı; plan NOT TESTED
gateway(paket)           = NOT KNOWN                   # iyzico/PayTR vb. sözleşme yok
katkı(paket)             = net_KDV_hariç − shopify_3p  # gateway hariç
maliyet_TRY(senaryo)     = maliyet_USD × 50,05
başa_baş_adet            = ceil( maliyet_TRY / katkı )
```

### Paket katkı (gateway hariç)

| Paket (istem) | Fiyat KDV dahil | Net /1,20 | Shopify %2 | Katkı (gateway hariç) |
|---:|---:|---:|---:|---:|
| 5 | 695 | 579.17 | 13.90 | **565.27** |
| 10 | 1290 | 1,075.00 | 25.80 | **1,049.20** |
| 25 | 2990 | 2,491.67 | 59.80 | **2,431.87** |
| 50 | 5490 | 4,575.00 | 109.80 | **4,465.20** |
| 100 | 9900 | 8,250.00 | 198.00 | **8,052.00** |

### Başa baş adet — Spot birincil

| Senaryo | 5 / 695 | 10 / 1290 | 25 / 2990 | 50 / 5490 | 100 / 9900 |
|---|---:|---:|---:|---:|---:|
| 1 sa/gün korunmuş Spot | 15–21 | 8–11 | 4–5 | 2–3 | 2–2 |
| 1 sa/gün discard Spot | 11 | 6 | 3 | 2 | 1 |
| 4 sa/gün korunmuş Spot | 30–35 | 16–19 | 7–9 | 4–5 | 3–3 |
| 4 sa/gün discard Spot | 26 | 14 | 6 | 4 | 2 |
| 12 sa/gün korunmuş Spot | 69–75 | 37–40 | 16–18 | 9–10 | 5–6 |
| 12 sa/gün discard Spot | 65 | 35 | 16 | 9 | 5 |

### Başa baş adet — on-demand üst sınır

| Senaryo | 5 / 695 | 10 / 1290 | 25 / 2990 | 50 / 5490 | 100 / 9900 |
|---|---:|---:|---:|---:|---:|
| 1 sa/gün korunmuş OD | 20–26 | 11–14 | 5–6 | 3–4 | 2–2 |
| 1 sa/gün discard OD | 19 | 10 | 5 | 3 | 2 |
| 4 sa/gün korunmuş OD | 50–56 | 27–30 | 12–13 | 7–7 | 4–4 |
| 4 sa/gün discard OD | 49 | 26 | 12 | 7 | 4 |
| 12 sa/gün korunmuş OD | 129–135 | 70–73 | 30–32 | 17–18 | 10–10 |
| 12 sa/gün discard OD | 128 | 69 | 30 | 17 | 9 |

## 3. Sabit kalem dökümü (bu rapor)

| Kalem | USD/ay | Kanıt |
|---|---:|---|
| GCP bilinen (wake+boot+snapshot) | 15,53 | `FIXED_COSTS.md` |
| Prod VPS srv2133566 KVM 2 | **24,49** | Hostinger billing sub `169lkvVaARjZU5YOD` |
| **Senaryo tabanı** | **40,02** | |
| Test VPS srv1662550 KVM 2 (hâlâ aktif) | 24,49 | aynı MCP; senaryoya dahil değil |
| Korunmuş local SSD (dururken) | 66–132 | SRE §4c senaryo; kesin **NOT KNOWN** |
| Hostinger Mail / domain / Shopify plan | NOT KNOWN | abonelik listesinde Mail trial / .ORG var; aylık senaryoya sokulmadı |
| Gateway komisyonu | NOT KNOWN | |
| Dış IP / egress | NOT KNOWN | |

## 4. NOT KNOWN listesi
- Korunmuş local SSD dururken uygulanan gerçek SKU/birim fiyat (SRE: NOT VERIFIED)
- Kısmi çalışma gününde korunmuş SSD’nin orantılı ücreti
- Günde 1’den fazla GPU uyandırma sayısı / cold overhead
- Gateway (iyzico/PayTR vb.) komisyon oranı ve sabit ücret
- Shopify plan abonelik ücreti (Basic/Grow…) ve mağaza ülkesi doğrulaması
- Hostinger Mail / domain’in bu ürün P&L’ine tahsisi
- On-demand’ın gerçekten kullanılacağı durum (Spot birincil)
- Idle 15 dk’nın günlük saat varsayımına eklenmesi (bu tabloda günlük saat = çalışma saati; idle ayrıca eklenmedi → idle eklemek istenirse **NOT KNOWN** veya ayrı model)
- Billing export satırlarıyla mutabakat (ilk veri ≤24 sa)

## 5. Risk notları
- Spot preempt → her preempt soğuk ~28 dk (koruma Preview; preempt’te local SSD kaybı, SRE §4c).
- `discardLocalSsd=false` Preview; Demir/KAN kararı açık.
- Canlı/ücretli sipariş ve irreversible cutover: **Demir onayı**.

## Ek — PAYG paylaşılan uyandırma (06.11.2026 09:35 TSİ (Europe/Istanbul))
KAN ölçümü: korunmuş disk wake ≈**4 dk**, soğuk **25–28 dk**; +15 dk idle. Spot uyandırma maliyeti ≈ **29,3 TL** (SSD) / **61,7–66,3 TL** (cold). Ayrıntı ve PAYG **179 TL** önerisi: `KANUS_PRICING_GLOBAL.md` §2.0.
