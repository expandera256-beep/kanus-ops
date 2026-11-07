# KANUS — Global fiyatlandırma ve ödeme araştırması

Güncelleme: 06.11.2026 09:38 TSİ (Europe/Istanbul)

Salt rapor (GLOBAL_PROMPT §6 fiyatlandırma, §7.3 uluslararası ödeme, §7.4 SaaS paketleri). **Canlı fiyat değişmedi.** Doğrulanamayan = **NOT KNOWN**.

## 0. Kaynaklar
- Canlı vitrin: `https://shop.kanus.org/products.json` (06.11.2026 ~09:27 TSİ)
- Admin salt okuma: user-Shopify `search_products` (ACTIVE+DRAFT)
- GPU: `COST_REPORT_20261106.md`, Spot **1,848238 USD/sa**, OD **3,748**; oturum cold+15 / SSD+15
- USDTRY senaryo: **50,05** (`COST_REPORT` / `FIXED_COSTS`)
- FX çapraz: open.er-api.com **06.11.2026 00:02:31 UTC** (= **03:02 TSİ**): USDTRY **50,076241**, EUR/USD **0,89048**, GBP/USD **0,742915**
- Shopify Payments ülkeleri: https://help.shopify.com/en/manual/payments/shopify-payments/supported-countries (Türkiye **yok**)
- Stripe: https://stripe.com/global
- PayPal TR: https://www.paypal.com/tr/home (2016’dan beri faaliyet yok)
- Paddle pricing: https://www.paddle.com/pricing — Turkey MoR: https://www.paddle.com/billing/turkey
- Lemon Squeezy ülkeler: https://docs.lemonsqueezy.com/help/getting-started/supported-countries — ücret: https://docs.lemonsqueezy.com/help/getting-started/fees
- iyzico yayın oran: https://www.iyzico.com/en/open-source (%4,29 + 0,25 TL, aylık >20.000 TL satış notu)
- PayTR: resmi “%2,19’dan başlayan” (sözleşme **NOT KNOWN**)

---

## 1. Canlı fiyat denetimi (shop.kanus.org)

| SKU | Durum | Fiyat (KDV dahil TRY) | İstem | TRY / istem |
|---|---|---:|---:|---:|
| US-CREDIT-1 | **DRAFT** | 149 | 1 | **149.00** |
| US-CREDIT-5 | **ACTIVE** | 695 | 5 | **139.00** |
| US-CREDIT-10 | **ACTIVE** | 1290 | 10 | **129.00** |
| US-CREDIT-25 | **ACTIVE** | 2990 | 25 | **119.60** |
| US-CREDIT-50 | **ACTIVE** | 5490 | 50 | **109.80** |
| US-CREDIT-100 | **ACTIVE** | 9900 | 100 | **99.00** |
| US-MONTHLY-200 | ACTIVE | 18900 | 200 | **94,50** |
| US-HOUR-1 | ACTIVE | 1990 | (1 saat) | n/a |

### Anomali kontrolü (çoklu paket > tekli paket birim fiyatı)
- Referans tekli: DRAFT **US-CREDIT-1 = 149 TL/istem**.
- Tüm ACTIVE kredi paketlerinde birim fiyat **149’un altında** ve paket büyüdükçe **monoton azalıyor** (139 → 129 → 119,60 → 109,80 → 99).
- **Anomali yok** (çoklu paket tekilden pahalı değil).
- Aylık 200 istem **94,50 TL/istem** — 100’lük paketten (~99) biraz daha ucuz; abonelik/hacim indirimi olarak tutarlı, anomali sayılmaz.
- **Not:** 1’lik paket DRAFT (07:26 Demir/SHOP); vitrinde satılmıyor. Birim ekonomide cold Spot oturumu tek isteğe yazılırsa 149 TL marjı ince kalır (aşağı §2).

---

## 2. Fiyat önerisi (yalnız teklif — canlıya uygulanmadı)

### 2.0 PAYG 360 TL neden ~3× paket birim fiyatıydı?

**Evet — önceki 360 TL/istem, her isteğin kendi soğuk açılışını (+15 dk idle) tek başına taşıdığı varsayımına dayanıyordu** (N=1, cold≈28 dk+15 → Spot ≈66,29 TL GPU; m=%70 → ≈360 TL).

Paket birim fiyatları (95–139 TL) ise **aynı uyandırmada birden fazla istem amortismanı** varsayar. Bu yüzden 360, paket merdivenine göre aşırı yüksek görünür; metodolojik olarak “her istek = soğuk uyandırma”dır, tipik paylaşılan oturum değildir.

### Formüller (paylaşılan uyandırma)

```text
uyandırma_dk = wake_dk + idle_dk
  wake_dk (KAN ölçümü): korunmuş disk ≈ 4 dk; soğuk 25–28 dk
  idle_dk = 15
GPU_uyandırma_TRY = saatlik_USD × (uyandırma_dk/60) × 50,05
GPU_istem_TRY     = GPU_uyandırma_TRY / N   # N = aynı uyandırmadaki istem
P = (GPU_istem_TRY + 0,25) / ((1−m)/1,20 − 0,0629)
  # 0,0629 = Shopify %2 + iyzico %4,29 (yayın); gateway sözleşme NOT KNOWN
```

Spot birincil **1,848238**; OD üst **3,748**.

| Senaryo | Uyandırma (wake+idle) | Spot TRY/uyandırma | OD TRY/uyandırma |
|---|---:|---:|---:|
| Korunmuş ~4 dk + 15 | 19 dk | **29.29** | 59.40 |
| Soğuk 25 + 15 | 40 dk | **61.67** | 125.06 |
| Soğuk 28 + 15 | 43 dk | **66.29** | 134.44 |

### 2.0b GPU/istem ve P (hedef marj) — Spot birincil

| N / uyandırma | GPU Spot (SSD 4′) | P Spot m=%70 | P Spot m=%50 | GPU Spot (cold 28′) | P cold m=%70 | P cold m=%50 | OD GPU (cold 28′) | P OD m=%50 |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 29.29 | **160** | 85 | 66.29 | **360** | 190 | 134.44 | 385 |
| 3 | 9.76 | **55** | 30 | 22.10 | **120** | 65 | 44.81 | 130 |
| 5 | 5.86 | **35** | 20 | 13.26 | **75** | 40 | 26.89 | 80 |
| 10 | 2.93 | **20** | 10 | 6.63 | **40** | 20 | 13.44 | 40 |
| 20 | 1.46 | **10** | 5 | 3.31 | **20** | 15 | 6.72 | 20 |

**Paket merdiveni (canlı birim TRY):** 5→139 · 10→129 · 25→119,6 · 50→109,8 · 100→99 · aylık200→94,5.

### 2.0c Önerilen PAYG list fiyatı (paketlerle tutarlı)

| Seçenek | TRY/istem | Mantık |
|---|---:|---|
| ~~Eski 360~~ | 360 | N=1 cold — paket merdivenini bozar; **liste fiyatı olarak kullanma** |
| Matemetik SSD N=1 m=%70 | 160 | Çalışan GPU + korunmuş açılış; 5’li paket biriminden (139) biraz yüksek |
| **Öneri liste: 179** | **179** | 5’li birim (139) ve 10’lu (129) üzerinde kalır → paketi baltalamaz; N≥3 SSD’de bol marj; N=1 cold’ta zararı sınırlar (tam karşılamaz) |
| Cold N=1 m=%70 tavan | 360 | Yalnız uyarı / “tek istek soğuk” maliyet etiketi; satış fiyatı değil |

**Öneri:** PAYG **179 TL/istem** (KDV dahil). Beklenen davranış: istekler çalışan GPU’ya düşer (N≈3–10). Tek-istek soğuk senaryoyu ürün/UX ile caydırmak (pakete yönlendir veya kısa bekleme birleştirme). Canlı fiyat **değiştirilmedi**.

### Formüller (genel — önceki)

```text
GPU_oturum_TRY = saatlik_USD × (dakika/60) × USDTRY
GPU_istem_TRY  = GPU_oturum_TRY / N
fee_oran = 0,02 + gateway; fee_sabit = gateway sabit
P = (GPU_istem_TRY + fee_sabit) / ((1−m)/1,20 − fee_oran)
```

**Varsayımlar:** Shopify Basic %2 (**plan NOT TESTED**); iyzico yayın %4,29+0,25; PayTR sözleşme **NOT KNOWN**.

### 2.1 Matematik taban (amortismanlı) — Spot SSD, hedef marj *(önceki; referans)*

| N | GPU/istem (SSD Spot ~19 dk) | P min m=%70 | P min m=%50 | Canlı paket P |
|---:|---:|---:|---:|---:|
| 1 | 29.29 | 160 | 85 | 149 DRAFT |
| 5 | 5.86 | 35 | 20 | 695 |
| 10 | 2.93 | 20 | 10 | 1290 |
| 25 | 1.17 | 10 | 5 | 2990 |
| 50 | 0.59 | 5 | 5 | 5490 |
| 100 | 0.29 | 5 | 5 | 9900 |

### 2.2 Kötümser — cold 28′+15, Spot, m=%70 *(N=1 = eski 360)*

| N | GPU/istem | P min m=%70 | Not |
|---:|---:|---:|---|
| 1 | 66.29 | **360** | Eski PAYG; liste için uygun değil |
| 5 | 13.26 | 75 | Paket 695 hâlâ yüksek marj |
| 10 | 6.63 | 40 | |
| 20 | 3.31 | 20 | |

### 2.3 Önerilen ticari tablo (güncellenmiş headline) — uygulanmadı

| Model / paket | Öneri TRY | TRY/istem | Not |
|---|---:|---:|---|
| **PAYG** | **179** | 179 | Paket merdiveni ile uyumlu; §2.0c |
| Kredi 5–100 | Canlı 695…9900 **korunabilir** | 139…99 | Zaten yüksek marj (amortismanlı) |
| İsteğe bağlı hafif zam | 749 / 1349 / … | — | Önceki tur; zorunlu değil |
| 1’lik DRAFT 149 | DRAFT kalsın | 149 | PAYG 179’dan ucuz kalırsa çelişki → 1’lik açma |

**Özet:** PAYG **179**; min paket **5**; 149 DRAFT açılmasın. Canlı fiyatlar aynı.

---

## 3. Uluslararası ödeme sağlayıcıları (TR merkezli şirket)

| Sağlayıcı | TR şirket kullanabilir mi? | Para birimleri | Ücret (yayın) | TR’ye ödeme (payout) | Shopify entegrasyonu | MoR / KDV | Kanıt |
|---|---|---|---|---|---|---|---|
| **Shopify Payments** | **Hayır** (ülke listesinde TR yok) | n/a | n/a | n/a | Yerleşik (uygun ülkede) | Hayır (Shopify/Stripe altyapı; satıcı satıcı) | [supported-countries](https://help.shopify.com/en/manual/payments/shopify-payments/supported-countries) — TR yok |
| **Stripe** | **Doğrudan hayır** (global listede TR yok) | Desteklenen ülkede hesap para birimleri | Ülkeye göre (TR hesabı yok) | Desteklenen ülke bankası | Shopify Payments / Stripe app (uygun ülkede) | Hayır | [stripe.com/global](https://stripe.com/global); TR yok |
| **PayPal** | **Hayır** (TR faaliyet 2016’da durdu) | n/a | n/a | TR hesapla alım/gönderim yok | n/a | Hayır | [paypal.com/tr](https://www.paypal.com/tr/home) |
| **Paddle** | **Evet** (TR MoR sayfası + satıcı onboarding iddiası) | Çoklu; müşteri ülkesine göre | **%5 + 0,50 USD** / checkout (yayın) | Aylık banka/Payoneer (**kesin valör NOT KNOWN**) | Doğrudan checkout; Shopify native Payments değil (ayrı checkout / app **NOT KNOWN** ayrıntı) | **Evet MoR** — vergi/VAT toplar | [paddle.com/pricing](https://www.paddle.com/pricing), [paddle.com/billing/turkey](https://www.paddle.com/billing/turkey) |
| **Lemon Squeezy** | **Evet** (banka payout listesinde **Turkey**) | Çoklu | **%5 + 0,50 USD** + intl +%1,5 + PayPal +%1,5 + abonelik +%0,5 | Banka (TR listede); PayPal payout TR’de PayPal kapalı → banka tercih | LS checkout; Shopify **NOT KNOWN** / sınırlı | **Evet MoR** — vergi dahil platform fee anlatımı | [supported-countries](https://docs.lemonsqueezy.com/help/getting-started/supported-countries), [fees](https://docs.lemonsqueezy.com/help/getting-started/fees) |
| **iyzico** | **Evet** (TCMB lisanslı TR) | TRY (+ multi-currency yıllık ücret yayın: 99 TL — doğrula) | **%4,29 + 0,25 TL** (yayın, >20k TL/ay notu) | TR banka (valör sözleşmeye bağlı) | **Resmi Shopify uygulaması** (üçüncü taraf) | Hayır — satıcı MoR; TR KDV satıcıda | [iyzico open-source/fees](https://www.iyzico.com/en/open-source) |
| **PayTR** | **Evet** (TR ödeme kuruluşu) | TRY | **%2,19’dan başlayan** (7g valör); gerçek oran **NOT KNOWN** (teklif) | TR banka, valör teklifle | Shopify app (mağaza ülkesi/para birimine bağlı) | Hayır | paytr.com (ikincil özetler); sözleşme şart |

### Shopify 3P işlem ücreti (TR için kritik)
Shopify Payments yokken üçüncü taraf gateway → plana göre ek ücret: Basic **%2**, Grow %1, Advanced %0,6, Plus %0,2 ([Shopify third-party fees](https://help.shopify.com/en/manual/your-account/manage-billing/billing-charges/types-of-charges/third-party-charges/third-party-transaction-fees)). Mağaza planı bu turda Admin’den **NOT TESTED**.

### Uygunluk özeti
- **TR içi kart:** iyzico veya PayTR + Shopify 3P ücreti.
- **Uluslararası kart / VAT MoR:** Paddle veya Lemon Squeezy (Shopify dışı veya hibrit checkout).
- **Shopify Payments / Stripe / PayPal (TR tüzel):** uygun değil (resmi sayfalar).
- Yabancı şirket kurup Stripe açmak hukuki/vergi kararı → bu raporda **önerilmez / NOT KNOWN uygunluk**.

---

## 4. SaaS fiyat modelleri (§7.4)

### 4.1 Pay-as-you-go
| | TRY | Not |
|---|---:|---|
| **Liste öneri** | **179** | Paylaşılan GPU; paket merdiveni üstü |
| SSD N=1 m=%70 taban | 160 | |
| Cold N=1 m=%70 (eski) | 360 | Liste değil; maliyet etiketi |

Ön ödeme cüzdanı veya faturalı kullanım; her başarılı ücretli istem düşümü (rezervasyon+settle).

### 4.2 Credit packs (öne çıkan)
§2.3 tablosu. Min paket **5**. DRAFT 1’lik satışa açılmasın.

### 4.3 Subscription tiers

| Katman | Kota (istem/ay) | Öneri TRY/ay | USD | EUR | GBP | Not |
|---|---:|---:|---:|---:|---:|---|
| Starter | 50 | 5690 | 113.69 | 101.24 | 84.46 | ~50’lik paket |
| Pro | 200 | 18900 | 377.62 | 336.27 | 280.54 | Canlı US-MONTHLY-200 |
| Business | 500 | 44900 | 897.10 | 798.85 | 666.47 | Öneri; hacim |

Yenileme, iptal, kota sıfırlama → §5 defter soruları. MoR (Paddle/LS) abonelik için uygun; Shopify Subscriptions TR gateway ile **kısmi / NOT KNOWN**.

### 4.4 Enterprise
- Özel kota, SLA, özel model, faturalı (PO), SSO, veri bölgesi.
- Fiyat: **NOT KNOWN** (satış görüşmesi); taban ≥ Business + GPU reserved kapasite maliyeti (`COST_REPORT` 4–12 sa/gün senaryoları).
- Sözleşme + e-Fatura TR; uluslararası için MoR veya yerel fatura **Demir/hukuk**.

### 4.5 Önerilen model (öneri)
1. **TR vitrin (Shopify):** Credit packs (min 5) + mevcut saatlik/aylık; ödeme **iyzico veya PayTR**.
2. **Global web/API:** **Paddle veya Lemon Squeezy** MoR checkout (USD/EUR/GBP), kredi defterine webhook ile yükleme.
3. PAYG cüzdan ileri faz.
4. Enterprise satış-destekli.

---

## 5. Abonelik defteri tasarımı — API botu ile koordinasyon

Kaynak API yanıtı: `us-panel-dev/patches/api-design-billing/TICARI_YANITLAR.md` + `BILLING_LEDGER_DESIGN.md` (06.11.2026). Bu bölüm **kararlaştırılmış politika değildir**; **KARAR GEREKİR** satırları Demir onayı olmadan nihai sayılmaz.

### 5.6 Kota sırası (düzeltme)

~~Eski ifade: «abonelik önce»~~ — **doğrulanamadı / çizildi.**

**Kod olgusu:** Önce **istem kredisi** (`Kind "sub"`: kredi paketleri + aylık plan aynı tür), sonra **saat paketi** (`Kind "use"`); aynı türde **en erken biten** önce (`licensesrv/accounts.go:117-137`). Abonelik ile paket arasında ayrı öncelik **yok**.

**Öneri (henüz karar değil):** priority alanı — abonelik dönemi → rollover → paket → admin → saat. **KARAR GEREKİR (Demir onayı yok).**

### 5.8 API yanıtları (8 soru)

Biçim: **Olgu** (dosya:satır) · **Öneri** · gerekiyorsa **KARAR GEREKİR (Demir onayı yok)**.

**1. Go hook’ta subscription nesnesi var mı?**
- **Olgu:** Yok. Yalnız tek seferlik `License` grant (`licensesrv/store.go:29-42`, üretim `licensesrv/webhook.go:176-196`). `Plan.Kind "sub"` abonelik değil, istem-kredisi türü (`licensesrv/store.go:22`); US-MONTHLY-200 de 30 günlük tek seferlik grant (`licensesrv/pricing.go:42`). Webhook yalnız `orders/paid|cancelled`, `refunds/create` (`licensesrv/webhook.go:93-124`); yenileme olayı yok. Paddle/LS kodu yok.
- **Öneri:** `subscription` varlığı + dönem başına `credit_grant` (tasarım §2–4).

**2. US_PLANS_JSON yenilemeyi nasıl temsil edecek?**
- **Olgu:** Temsil etmiyor: `product_id → {kind, days, prompts, seconds, credits}` (`licensesrv/plans.go:11-60`, `licensesrv/store.go:21-27`); `US_PLANS_FILE` öncelikli (`cmd/us-license-hook/main.go:37-47`); boşsa `LivePlansJSON()` (`licensesrv/pricing.go:55-65`). Recurring/interval **yok**.
- **Öneri:** İsteğe bağlı geriye uyumlu alanlar: `recurring`, `interval`, `renewal_mode`, `rollover_cap`, `max_concurrent`, `provider_refs`.
- **KARAR GEREKİR (Demir onayı yok):** `renewal_mode` (replace|stack) ve `rollover_cap` değerleri.

**3. Shopify + Paddle müşterisi hangi anahtarla tek account_id’de birleşir?**
- **Olgu:** Shopify customer id **saklanmıyor**; lisansa sipariş e-postası yazılıyor (`licensesrv/webhook.go:164-173,192`). Bağlama yalnız `/account/redeem` (`licensesrv/accounts.go:201-214,531-545`); otomatik e-posta eşleşmesi yok. E-posta normalize (`licensesrv/accounts.go:44`).
- **Öneri:** İç `account_id` + `account_provider_link`; otomatik bağ yalnız **doğrulanmış** e-posta eşleşmesinde; aksi halde redeem / checkout `account_id` custom_data.

**4. Ledger PostgreSQL mi, JSON mu?**
- **Olgu:** JSON store (flock + atomik rename) (`licensesrv/store.go:101-186`), `US_STORE_PATH` (`cmd/us-license-hook/main.go:50`). Go’da PostgreSQL kodu **yok**.
- **Öneri:** Önce JSON’a gölge ledger; sonra PG geçiş yolu (tasarım §2.1).
- **KARAR GEREKİR (Demir onayı yok):** PG’ye geçilecek mi, ne zaman (VPS PG/yedek sorumluluğu dahil).

**5. Rezervasyon TTL ile period_end?**
- **Olgu:** HTTP kredi TTL 45 dk (`licensesrv/credits.go:19`); model proxy `US_MODEL_TIMEOUT_SECONDS` (vars. 2700)+5 dk (`licensesrv/model_meter.go:56`, `main.go:114`). Reserve/commit `now < ExpiresAt` ister (`licensesrv/store.go:240,281`); süresi dolmuş rezervasyon commit edilemez (`276-278`) → release (`302-312`). Bitişe yakın uzun çağrı bitişten sonra commit edilemez → kredi düşmez.
- **Öneri:** Commit’i rezervasyon açılış anındaki geçerliliğe bağla; yeni reserve için `expires_at - now >= TTL`.

**6. Yenileme başarısız / kart reddi e-posta şablonları kimde?**
- **Olgu:** Bu şablonlar yok. Mevcut sabit metin: kayıt/sıfırlama/doğrulama (`licensesrv/accounts.go:384,448,452`), sipariş (`licensesrv/reversals.go:102-107`); Outbox+SMTP (`licensesrv/mail.go`).
- **Öneri:** MoR dunning varsa biz göndermeyelim; Shopify/manual için Outbox’a iki şablon (metin Ticari, kod API).
- **KARAR GEREKİR (Demir onayı yok):** Metin sahibi; MoR e-postalarının açık olup olmayacağı.

**7. Faturasız kurumsal admin grant API gerekli mi?**
- **Olgu:** Yok; uçlar webhook, `/v1/credits/*`, `/account/*`, model proxy (`main.go:71-105`, `accounts.go:304`). Tek kredi kaynağı Shopify `orders/paid`.
- **Öneri:** Enterprise satılacaksa `POST /v1/admin/credit-grants` (loopback/mTLS, Idempotency-Key, üst sınır, audit) — tasarım §11.
- **KARAR GEREKİR (Demir onayı yok):** Gerekli mi.

**8. KDV / brüt-net ledger?**
- **Olgu:** Vergi alanı yok; «Fiyatlar KDV dahil TL» (`licensesrv/pricing.go:9-11`); webhook tutar/vergi okumuyor (`licensesrv/webhook.go:48-63`).
- **Öneri:** Grant’e `gross_minor, tax_minor, net_minor, currency, tax_collector`; değerler sağlayıcı payload’ından; kredi bakiyesi parasal değil.
- **KARAR GEREKİR (Demir / muhasebe onayı yok):** Shopify TR KDV hesap/beyan ve e-Fatura sorumluluğu.

**Notlar (kesinleşmemiş / sınır):** Paddle/LS/Shopify abonelik alan adları sandbox’ta **doğrulanmadı**. Provider status eşlemesi öneri tablosudur (`BILLING_LEDGER_DESIGN.md` §3.2).

### 5.1–5.5 (önceki gereksinim taslağı — hâlâ tasarım)

Varlıklar / state machine / yenileme / iade / webhook maddeleri API tasarımında (`BILLING_LEDGER_DESIGN.md`) genişletildi; uygulaması Demir kararları + ayrı yamalara bağlı.


---

## 6. NOT KNOWN
- Canlı Shopify plan adı / 3P oranının hesaptaki değeri
- iyzico/PayTR **sözleşme** oranı (yayın ≠ sözleşme); PayTR kesin oran
- Gateway sabit ücretleri, taksit, BSMV yansıması muhasebe
- Paddle/LS başvuru onayı, payout valörü, yasaklı kategori (AI/API) sonucu
- Lemon Squeezy / Paddle ↔ Shopify sipariş senkronunun resmi yolu
- FX’in mağaza fiyatına bağlanma politikası (günlük kur vs sabit)
- Idle+cold’un KAN teyitli dakika değerleri
- Korunmuş SSD dururken gerçek SKU maliyeti

## 7. Demir onayı gerekenler
- Ledger **KARAR GEREKİR** maddeleri: `DEMIR_KARAR_ODEME_20261106.md` §C
- Her türlü canlı fiyat değişikliği
- Ödeme sağlayıcı başvurusu / production key
- MoR checkout’un shop.kanus.org ile yan yana açılması
- 1’lik paket politikası (DRAFT kalsın)

