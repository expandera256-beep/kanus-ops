# Paddle onboarding notu (global MoR)

Güncelleme: 06.11.2026 09:38 TSİ (Europe/Istanbul) · Salt belge · Hesap açılmadı
Demir kararı 09:36: global satış **Paddle** (`DEMIR_KARAR_ODEME_20261106.md`).

## 1. Ürün uygunluğu (dijital / AI kredi) — dikkat

Kaynak AUP (son güncelleme sayfada 13 Apr 2026): [what-am-i-not-allowed-to-sell-on-paddle](https://www.paddle.com/help/start/intro-to-paddle/what-am-i-not-allowed-to-sell-on-paddle).

- Paddle hedefi: yazılım / B2B SaaS / consumer software / games. Fiziksel mal veya yazılımla bağlantısız insan hizmeti uymuyor.
- **Önemli risk — «stored value»:** AUP yasakları arasında *«Virtual currency or stored value, including but not limited to store credit, gift cards, vouchers»* var. Ön ödemeli **istem kredisi paketi** buna çekilebilir. Konumlandırma önerisi (henüz Paddle onayı yok): «SaaS / yazılım paneli erişim hakkı» (süreli lisans), salt cüzdan/hediye çeki dili kullanma. **Paddle’a başvuru öncesi yazılı uygunluk teyidi şart — şimdilik NOT KNOWN / risk açık.**
- **Üretken AI / içerik üretimi:** AUP’de deepfake, face swap, izinsiz likeness, insanları otomatik sınıflandırma yasak/kısıtlı. Model sohbet-onarım paneli farklı konumlanmalı; ek due diligence beklenir.
- Domain onayı: live checkout yalnız onaylı domain ([go-live checklist](https://developer.paddle.com/build/go-live-checklist)).
- **Kesin ürün onayı:** Paddle risk incelemesi sonrası; hesap açılmadı.

## 2. Onay süreci / süre

| Aşama | Ne | Süre |
|---|---|---|
| Kayıt | Live hesap + e-posta doğrulama | Dakikalar |
| Kimlik / işletme | Owner ID, adres; Sumsub / ek belge | **NOT KNOWN** (iş/ülkeye göre; [identity](https://www.paddle.com/help/start/account-verification/what-is-identity-verification), [verification](https://www.paddle.com/help/start/account-verification)) |
| Website approval | Checkout domain’leri AUP kontrolü | **NOT KNOWN** (paralel yürüyebilir; [checklist](https://developer.paddle.com/build/go-live-checklist)) |
| Katalog + webhook | Product/price, notification URL | Operatör |
| Sandbox → live | Ayrı anahtarlar, domain live’da zorunlu onay | — |

Takvim taahhüdü resmi sayfada sabit gün olarak yok → süre **NOT KNOWN**.

## 3. Belgeler / doğrulama

Tipik: işletme bilgileri, sahibi kimlik, adres kanıtı; uyuşmazlıkta ek belge ([account verification](https://www.paddle.com/help/start/account-verification)). Türkiye’ye özel evrak listesi sayfada tek satırda yok → başvuru sırasında netleşir (**NOT KNOWN** şimdi).

## 4. TR banka payout

- Aylık payout; eşik min **$100**; wire veya Payoneer ([when/how paid](https://www.paddle.com/help/manage/get-paid/when-and-how-do-i-get-paid)).
- Payout settings: IBAN / BIC vb. ([payout settings](https://www.paddle.com/help/manage/get-paid/how-do-i-set-up-my-payout-settings)).
- TR şirket onboarding: [paddle.com/billing/turkey](https://www.paddle.com/billing/turkey).
- TR bankaya fiilen ödeme + para birimi / SWIFT ücreti: yardım genel; **bu turda TR özel teyit NOT KNOWN**.

Ücret yayın: **%5 + 0,50 USD**/checkout ([pricing](https://www.paddle.com/pricing)).

## 5. Shopify ile yan yana

| | Shopify+PayTR (TR) | Paddle (global) |
|---|---|---|
| Checkout | shop.kanus.org Shopify | Paddle overlay / hosteli checkout (onaylı domain, örn. kanus.org/en) |
| MoR | Hayır (satıcı) | Evet (Paddle) |
| Lisans teslimi | Mevcut: Shopify webhook → `us-license-hook` | **Öneri:** Paddle notification → aynı hook’ta yeni provider adaptörü → `credit_grant` / redeem; Shopify `orders/paid` yolundan ayrı `payment_ref` |
| Çift tahsilat | Aynı sepeti iki yerde satma | Ürün/coğrafya ayır (TR Shopify, global Paddle) |

Paddle alan adları sandbox’ta **doğrulanmadı** (`TICARI_YANITLAR` notu).

## 6. Resmi linkler

- AUP (ne satılmaz): https://www.paddle.com/help/start/intro-to-paddle/what-am-i-not-allowed-to-sell-on-paddle
- MSA / AUP URL: https://www.paddle.com/legal/terms (AUP: paddle.com/support/aup/)
- Pricing: https://www.paddle.com/pricing
- TR: https://www.paddle.com/billing/turkey
- Verification: https://www.paddle.com/help/start/account-verification
- Payout: https://www.paddle.com/help/manage/get-paid/when-and-how-do-i-get-paid
- Go-live: https://developer.paddle.com/build/go-live-checklist

## 7. Sonraki adımlar (Demir onayı sonrası)

1. Sandbox hesap + örnek kredi ürünü  
2. Domain submit (kanus.org / checkout path)  
3. Webhook → hook tasarımı (API yaması; ledger kararları §C)  
4. Live verification + payout IBAN  
5. Mali müşavir: MoR settlement (`MALI_MUSAVIR_SORULARI.md`)

## Ön başvuru uygunluk sorgusu

İngilizce, **gönderilmemiş** taslak: [PADDLE_UYGUNLUK_SORU_TASLAK.md](./PADDLE_UYGUNLUK_SORU_TASLAK.md). Resmi satış kanalı https://www.paddle.com/demo; hesap açılmadı. Stored-value kapsamı, TR banka payout ve inceleme süresi yazılı teyit bekliyor.
