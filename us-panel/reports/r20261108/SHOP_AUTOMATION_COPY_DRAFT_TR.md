# Otomasyon metinleri — TASLAK / YAYINLANMADI

## 1. Mağaza ana sayfası
Mevcut: “Anahtarlar sipariş sonrası manuel olarak iletilir.”

**Seçenek A:** Ödeme sonrası lisans anahtarınız birkaç dakika içinde otomatik olarak e-postanıza gönderilir. Lisansınızı US uygulamasındaki Hesap ekranından bağlayabilirsiniz. Sorun yaşarsanız: support@kanus.org.

**Seçenek B:** Lisans anahtarınız, ödemenizin ardından birkaç dakika içinde otomatik e-postayla iletilir. US uygulamasında Hesap ekranından lisansınızı bağlayın. Destek: support@kanus.org.

## 2. /us sayfası
Mevcut: “Otomatik hesap açılışı ve lisans teslimi henüz etkin değil…”

**Seçenek A:** Ödemenizin ardından lisans anahtarınız birkaç dakika içinde otomatik olarak e-posta ile gönderilir. Anahtarı US uygulamasının Hesap ekranından hesabınıza bağlayabilirsiniz. Yardım için support@kanus.org.

**Seçenek B:** Lisans teslimi ödeme sonrası otomatik e-postayla, birkaç dakika içinde yapılır. Lisansınızı kullanmak için US uygulamasındaki Hesap ekranından bağlayın. Sorularınız için: support@kanus.org.

## Yayın koşulları ve sınırlar
- Bu ifadeler hedeflenen üretim davranışını anlatan TASLAKTIR; mevcut canlı durumun doğrulanmış vaadi değildir. Hiçbir yerde yayınlanmadı.
- Canlı mağazanın üretim hook'u api.kanus.org ve live webhook'ları henüz cutover edilmedi. Metinler yalnız canlı cutover ve üretim teslim doğrulaması tamamlandıktan sonra yayınlanmalı.
- Testte gerçek SMTP gönderimi allowlist ile sınırlı; #1003 outbox skipped beklenen sonuçtur, gerçek müşteri teslim kanıtı değildir.
- Idempotency Demir kararıyla test edilmedi (08.11.2026 00:52 TSİ); PASS değildir.
- Hesap/panel görünürlüğü UNKNOWN. Metinler otomatik hesap açılışı vaat etmez; lisans bağlama US uygulamasının Hesap ekranında yapılır.
- “Birkaç dakika” zamanlaması üretimde doğrulanmadan yayınlanmamalı; mevcut diğer süre vaatleriyle de tutarlı hale getirilmeli.

## Onay — 00:55 TSİ
Demir: "B - B onaylıyorum". Ana sayfa = Seçenek B, /us = Seçenek B. Henüz yayınlanmadı (canlı tema yazımı bağlayıcıda engelli; yayın kapısı: canlı webhook cutover + üretim e-posta doğrulaması).
- 00:55 TSİ Demir: yayın zamanı = canlı webhook geçişi + e-posta doğrulamasından sonra.
