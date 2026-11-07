# release_us — kanus.org/us çok dil düzeltmesi (YEREL, YÜKLENMEDİ)
Yükleme KAN.US'te ve Demir onayına bağlı. DNS/Shopify/GitHub değişikliği yok.
Değişiklikler (canlı bundle üzerinde): tarih gösterimi dile göre Intl.DateTimeFormat (tr-TR/en-US/de-DE/ru-RU); iletişim, forum konusu ve yükleme formlarında başarısız gönderimde 4 dilde hata mesajı (önce sessizce düşüyordu). Metin sözlüklerine dokunulmadı (TR/EN/DE/RU 126 anahtar zaten tam).
Fiyat: canlı sitede fiyat gösterilmiyor (payments rotası /us'a yönleniyor); döviz gösterimi eklenmedi, kur kaynağı yok.
Yükleme: public_html/ içeriğini sunucu public_html/ köküne. Sonra `./VERIFY_AFTER_UPLOAD.sh`.
Yeniden üretim: `python3 site/us/build_us.py`.
