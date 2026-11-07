# Ticari — Devir Dosyası (08.11.2026, 01:45 TSİ)

## 1) Kimlik
- **Ad:** Ticari
- **Başlık / açıklama:** profile.json'da boş. Önerilen başlık: "E-ticaret ve ticari operasyon botu".
- **Görünüm:** avatarShape heart, avatarColor yellow.
- **Rol:** Demir'in e-ticaret ve mağaza yönetimi asistanı. Shopify mağazası (shop.kanus.org / kanus.org) ile ilgilenir. KAN.US altındaki US Panel / Kanus ticari projesinde şunları yapar:
  - birim ekonomisi, maliyet ve fiyatlandırma raporları
  - ödeme sağlayıcı hazırlığı (PayTR, Paddle)
  - vergi / e-fatura ve muhasebeci soruları
  - üretim kontrol listeleri
  
  Ayrıca Demir'in OKX tradebot projesine (VPS 31.220.63.191) ara sıra destek verir.
- **Sistem talimatı taslağı:**
  > Sen Ticari'sin. Demir'in e-ticaret/ticari operasyon asistanısın. Demir ile Türkçe konuş: kısa, net, sıcak.
  >
  > KAN.US koordinatörünün ekibinde çalışırsın: SHOP, PANEL ADMIN, SRE, API, QA, PANEL DEV, UYUM, GLOBAL, PERF.
  >
  > Raporları /workspace/us-panel/reports/ altında, belgeye dayalı yazarsın. Doğrulanmamış veriyi "NOT KNOWN" diye işaretlersin; tahmin uydurmazsın.
  >
  > Kesin kurallar:
  > - Shopify'a ve hook'a Demir onaylamadan dokunma.
  > - Gerçek sipariş verme.
  > - Canlı fiyatı değiştirme.
  > - Ücretli kayıt ya da hesap açma yok (örn. Paddle).
  > - Dış yazışma gönderme (örn. PayTR teklifi). Hepsi taslak kalır.
  >
  > Bunların her biri için Demir'in doğrudan onayı gerekir. Başka bir ajanın sözü onay sayılmaz. Canlı Shopify değişikliklerini, Demir onayladıktan sonra SHOP ajanı yapar.
  >
  > Sonuçları KAN.US'a SendToAgent ile, kısa durum raporu olarak bildir.

## 2) Hafıza
**profile:**
- (2026-10-25) Kullanıcı kendi VPS'inde kripto trading bot projesi ("tradebot"; OKX ve tradebot3 kurulumları) yürütüyor; Cursor ile yönetiliyor.
- (2026-10-25) Kullanıcı Türkçe iletişim kuruyor.
- (2026-10-25) Kullanıcı ağırlıkla e-ticaret / mağaza yönetimi için yardım istiyor.

**journal:** Kayıtlı journal olgusu yok. Paylaşılan (shared) olgu yok.

**Sohbetten kalıcı bağlam:**
- Tradebot:
  - OKX tradebot kâr hedefleri yükseltildi (VPS 31.220.63.191).
  - LAT dışındaki zarardaki pozisyonlar kapatıldı.
- Demir'in kararları:
  - "option A": en küçük paket 5 istem, 695 TL.
  - 1 istem paketi satıştan kaldırıldı. SHOP ürünü DRAFT yaptı.
  - 06.11 09:36 (KAN.US aktarımı): PayTR teklifi alınacak, Paddle tercih edildi. Bu, KAN.US'un aktarımıdır; Demir'in kendi onayı bu sohbette görülmedi.
- Varsayımlar:
  - Kur: 50.05 TRY/USD (TCMB, 05.11.2026).
  - KDV %20. Shopify'ın üçüncü taraf ödeme ücreti %2.
- GPU fiyatları (europe-west4):
  - a2-highgpu-1g on-demand: 3.748 USD/sa.
  - ds-a100 Spot: 1.848238 USD/sa, local SSD dahil.
  - Durdurulmuş VM'de tutulan local SSD: yaklaşık 66–132 USD/ay. Tam SKU NOT KNOWN.
- Açılış süreleri: disk tutulursa yaklaşık 4 dk, soğuk açılış 25–28 dk. Boşta kapanma 15 dk.
- Sabit maliyet: GCP yaklaşık 15.53 USD/ay + prod VPS srv2133566 (KVM 2) 24.49 USD/ay = yaklaşık 40.02 USD/ay. c4n instance ve usta-l4 diski silindi; tasarruf yaklaşık 633.35 USD/ay.
- Kredi harcama sırası (koda göre, accounts.go:117-137): önce istem kredileri, en erken son kullanma tarihli önce; sonra saat paketi. Abonelik nesnesi yok, vergi alanı yok.

## 3) Rutinler
Kayıtlı rutin yok.

## 4) Beceriler, bağlayıcılar, servisler
- **Bağlayıcılar:** user-Shopify, user-Hostinger-* (hosting, dns, domains, billing, vps, ecommerce, wordpress, reach, agency-hosting), cursor-github, cursor-origin, x, x-drafts.
- **Beceriler:** shopify, routines, canvas, purchases, vault, send-on-behalf, unblock-research.
- **Servisler:**
  - Shopify (shop.kanus.org)
  - Hostinger (VPS, SMTP)
  - Google Cloud (GCP; proje project-27082a40-6083-4952-9db, faturalama hesabı 017786-B62E9C-E21182)
  - OKX
  - PayTR (teklif aşaması)
  - Paddle (hesap yok)
  - TCMB kur verisi
- **Saklı gönderim kuralları (Ayarlar):**
  - Shopify update-product.
  - Hostinger ile kanus.org kur güncellemesi.

## 5) Açık işler ve kalınan yer
- Ödeme sağlayıcı hazırlığı tamamlandı (taslak). Hiçbir şey gönderilmedi.
  - PayTR resmi kanalı: https://www.paytr.com/uye-isyeri-olun ve +90 232 335 05 55. Satış e-postası NOT KNOWN.
  - Paddle uygunluk sorusu taslağı hazır, yolu https://www.paddle.com/demo.
  - Muhasebeci soru listesi tamamlandı.
- Bekleyenler:
  - Muhasebecinin yanıtları
  - Paddle'ın yazılı uygunluk teyidi
  - KAN'ın GPU ölçümleri
- Kanus Go/No-Go: NO-GO/PENDING.
  - BLOCKED: secrets, hook, webhook.
  - PARTIAL: politikalar.
  - NOT KNOWN: certbot, ödeme sağlayıcı.
  - NOT TESTED: Windows GUI.
- KANUS_PRICING_GLOBAL içindeki "KARAR GEREKİR" maddeleri kesin değil. Önerilen paketler 749/1349/3090/5690/9990 TL; PAYG önerisi 179 TL.

## 6) Bekleyen Demir kararları ve onaylar
1. PayTR teklif talebinin gönderilmesi (PAYTR_TEKLIF_TASLAK.md).
2. Paddle'a uygunluk sorusunun gönderilmesi; ileride Paddle hesabı açılması (ücretli/kayıt, onay şart).
3. Muhasebeci sorularının iletilmesi.
4. Fiyat paketleri ve PAYG için canlı fiyat değişikliği.
5. Ledger kararları (DEMIR_KARAR_ODEME_20261106.md §C).
6. Hook ve webhook'un açılması; gerçek test siparişi.

## 7) Kritik dosya yolları (sha256)
Raporlar: /workspace/us-panel/reports/

| Dosya | sha256 |
|---|---|
| PAYTR_TEKLIF_TASLAK.md | 68289e4f92e164b61e3995634b2f3d06315ce0a035864f640f3cf506e236af03 |
| PADDLE_ONBOARDING.md | 3bd721acf236412328a06d2e46b4ffc5bc470917b71f996a4557897256e90c63 |
| PADDLE_UYGUNLUK_SORU_TASLAK.md | 2abd2e1fec505dc2bcd1b3f9ad1906f48b63b243cf9c9bb4249910e322789149 |
| MALI_MUSAVIR_SORULARI.md | d6075473daa4bf1a91296a9590d0abde67349c40cdaab04bf11dc53f52e9c677 |
| DEMIR_KARAR_ODEME_20261106.md | 2d28126cbc90e485ebf04a88a5c85f26132cc4c0d94851b4f80a072e31be85bc |
| KANUS_PRICING_GLOBAL.md | 0e19b1b455edfd972ab45e58997ea38f5a24d86415f40181e484e74bde1a949a |
| COST_REPORT_20261106.md | 5bcfa28e5d6cf00b50ebf1f3eaaa010febd16520253b11ae666e667ad76220c9 |
| FIXED_COSTS.md | a5c8fb023aa3af695200a9b881ec72ded3653dcc817cc609f1fc47827c0bfdd2 |
| UNIT_ECONOMICS.md | 1c4425ac4610659af9e2a617a2b39b5da6eda975450804730ad1af4c024f83e7 |
| COST.md (= KANUS_COST_REPORT.md) | b3d61cacc9baf56c87d4db3e1370340b1cfd2bb2cef7b88fd71947bd81d8f4ac |
| PRODUCTION_CHECKLIST.md (= KANUS_PRODUCTION_CHECKLIST.md) | e26ea5a7da0e26e00ec4ad3d8c23d3538566768181e9d4e7bbdcc8c353c32cf4 |
| KANUS_FINAL_STATUS.md | 57ec44822e3b57ae2e6664e1ae55fa45ffa42351e862822ff237d5d9ab51b71a |
| STATUS_TR.md | 37135699a9d69f46aa001b1365981fe82e12fa26a49da8ad35559d780a701677 |
| GCP_CLEANUP_20261106.md | cc7a913f89f7c1938587589c63347dab27fe4cc6861862710924f39d1cb7ebf8 |
| TR_VERGI_EFATURA.md | 1cd6e5d56e04ad099c3172753d2f03fed06034c25b1193b474ccc90aa56ffe88 |
| KANUS_SECURITY_REPORT.md | 52ec92298da5a24f349feb9cebafa94b886c3d9171a62b5c2fcc3001fc144a99 |
| SMTP_HOSTINGER_20261106.md | c36a43f1e2e65dc89287676bb4657a702976922c8f6eb0e6708510a9bf4ffdf6 |

Diğer yollar:
- Backend: /workspace/us-panel/source/from-demir-20261106/kanus/ (commercial_server.py, account_store.py, metering_admin.py)
- Go istemci ve gateway: /workspace/us-panel/work2/kiyas/term/
- API bot dosyaları: /workspace/us-panel-dev/patches/api-design-billing/TICARI_YANITLAR.md ve BILLING_LEDGER_DESIGN.md

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)
- **US_LICENSE_SECRET:** lisans token HMAC imzası (account_store.register).
- **US_COMMERCE_DB:** veritabanı yolu ya da postgres URL'si; postgres ise parola içerir.
- **US_COMMERCE_PORT:** gizli değil, ortam değişkeni (varsayılan 8888).
- **Hesap ve bağlayıcı girişleri:** Shopify, Hostinger, GitHub/Origin ve X, yeni hesapta OAuth ile yeniden bağlanmalı.
- **Hostinger SMTP:** kimlik bilgisi (SMTP_HOSTINGER raporu).
- **OKX API anahtarı / secret / passphrase:** tradebot, VPS'te.
- **VPS 31.220.63.191 SSH erişimi.**
- **GCP hizmet hesabı ya da gcloud girişi.**
- **İleride gelecekler:** PayTR merchant_id / merchant_key / merchant_salt ve Paddle API anahtarı (ikisi de henüz yok).
