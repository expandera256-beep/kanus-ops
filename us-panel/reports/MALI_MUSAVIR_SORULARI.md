# Mali müşavir — kısa soru listesi (KDV / e-belge)

08.11.2026 00:09 TSİ (Europe/Istanbul) · Salt liste · Gönderilmedi

Kaynak bağlam: TR Shopify+PayTR (Demir 09:36 teklif); global Paddle MoR; ledger vergi alanları yok (`TICARI_YANITLAR` S8).

1. **Shopify TR peşin satış:** KDV’yi kim hesaplar/beyan eder (Shopify vergi mi, biz mi)? e-Arşiv / e-Fatura zorunluluğu ve alıcı tipi?
2. **PayTR komisyon faturası:** BSMV/KDV indirimi; iade sonrası komisyon muhasebesi?
3. **Paddle MoR:** Müşteri faturasını Paddle kesince biz Paddle’a ne belge keseriz? Settlement KDV’siz hizmet ihracı (KDVK 11/1-a, 12/2) sayılır mı, yoksa farklı mı? (**özelge/MM**)
4. **Hizmet ihracı koşulları:** SaaS/kredi’de «yurt dışında faydalanma»; TR kullanıcılı global planda istisna riski?
5. **Ön ödemeli kredi:** Gelir satışta mı kullanımda mı tahakkuk eder; kullanılmayan/süresi dolan krediler ne zaman gelirleşir? Ertelenmiş gelir ve iade yükümlülüğü nasıl izlenir?
6. **İade / chargeback:** KDV düzeltme ve e-belge iptal/iade usulü (Shopify vs Paddle)?
7. **Stopaj / kurumlar:** Paddle veya yabancı kartlı tahsilatta stopaj veya ek yükümlülük var mı?
8. **Para birimi:** Paddle USD/EUR settlement’ın defter kuru ve KDV matrahı?

9. **Ledger vergi alanları (API taslağı):** `gross_minor`, `tax_minor`, `net_minor`, `currency`, `tax_collector` yeterli mi? MoR brüt satış, komisyon, vergi ve net payout ayrı mı kaydedilmeli; hangi belge/dönem ile mutabakat yapılmalı? Mevcut Go hook’ta bu alanlar yok; tasarım muhasebe onayı bekliyor.

Mevzuat ankrajı (genel): [3065 KDV Kanunu](https://www.mevzuat.gov.tr/mevzuat?MevzuatNo=3065&MevzuatTur=1&MevzuatTertip=5) md. 11/1-a, 12/2. MoR özel uygulaması **NOT KNOWN** — MM yanıtı şart.
