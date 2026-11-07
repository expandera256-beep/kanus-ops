# Demir karar notu — ödeme sağlayıcıları

Güncelleme: 06.11.2026 09:38 TSİ (Europe/Istanbul) · Salt belge · Canlı kesim / gönderim yok

## KAN.US bildirimi: Demir kararı 09:36

- **TR:** PayTR’den **yazılı teklif** alınacak (kesin oran+valör teklifle; iyzico yedek).
- **Global:** Satışlar **Paddle** (MoR) üzerinden.
- Kaynak: KAN.US sözlü/bildirim 06.11.2026 **09:36 TSİ**. Bu ajan hesap açmadı / e-posta göndermedi.
- Taslaklar: `PAYTR_TEKLIF_TASLAK.md` (gönderilmedi), `PADDLE_ONBOARDING.md`, `MALI_MUSAVIR_SORULARI.md`.

## A) Türkiye: iyzico vs PayTR (Shopify 3P %2 varsayımı — plan NOT TESTED)

| | **iyzico** | **PayTR** |
|---|---|---|
| Yayın komisyon | **%4,29 + 0,25 TL** (BSMV dahil; >20k TL/ay notu) | **%2,19’dan başlayan** (7g valör); kesin oran **teklif / NOT KNOWN** |
| Shopify 3P | +%2 Basic | +%2 |
| Peşin efektif (yayın) | ~%6,29 + 0,25 | ~%4,19 (taban; gerçek ↑ olabilir) |
| Payout | ~3 hafta blokaj + Çarşamba (sözleşmeye bağlı) | Valör teklifte |
| Taksit | Banka vade farkı | Teklifte |
| Shopify | Resmi uygulama | App / entegrasyon |
| Kaynak | [iyzico fiyat](https://www.iyzico.com/fiyatlandirma) | [PayTR sanal POS](https://www.paytr.com/paytr-sanal-pos) |

**Net (TRY), peşin, Shopify %2 + PayTR taban %2,19:** 695 → iyz 651 / PayTR 666 (+15); 1290 → +27; 2990 → +63; 5490 → +116; 9900 → +208.

**Durum A:** Demir 09:36 → **PayTR yazılı teklif**. Teklif kötüyse iyzico’ya dönülür.

## B) Global: Paddle vs Lemon Squeezy

| | **Paddle** (seçildi) | Lemon Squeezy |
|---|---|---|
| Ücret | **%5 + 0,50 USD** | %5+0,50 + intl +%1,5 + … |
| MoR | Evet | Evet |
| TR | [billing/turkey](https://www.paddle.com/billing/turkey) | Banka payout listesinde Turkey |
| Kaynak | [pricing](https://www.paddle.com/pricing) | [fees](https://docs.lemonsqueezy.com/help/getting-started/fees) |

**Durum B:** Demir 09:36 → **Paddle**. MoR vergi/e-belge → mali müşavir (`MALI_MUSAVIR_SORULARI.md`).

## C) Ledger — KARAR GEREKİR (Demir onayı yok; API `TICARI_YANITLAR.md`)

1. `renewal_mode`: replace mi stack mi; `rollover_cap`
2. PostgreSQL’e geçiş var mı / ne zaman (VPS PG+yedek)
3. Dunning e-posta metin sahibi; MoR e-postaları açık mı (biz göndermeyelim mi)
4. Enterprise admin grant API (`POST /v1/admin/credit-grants`) gerekli mi
5. Shopify TR KDV hesap/beyan + e-Fatura sorumluluğu (muhasebe)
6. Tüketim öncelik önerisi (abonelik dönemi→rollover→paket→admin→saat) teyidi — bugünkü kod: istem kredisi→saat, erken biten önce
7. *(Tasarım ekleri, aynı karar kümesi)* `past_due` hard/soft-block; kısmi iade oranı; kullanılmış kredide borç vs sıfırda dur; eşzamanlı reserve kotası; saat paketi tercih başlığı

## Karar kutusu

- [x] A: **PayTR yazılı teklif** (Demir 09:36) — taslak hazır, **gönderilmedi**
- [x] B: **Paddle** (Demir 09:36) — onboarding notu hazır, hesap açılmadı
- [ ] PAYG 179 TL liste — şimdilik yok / sonra: ________
- [ ] §C ledger maddeleri: ________
