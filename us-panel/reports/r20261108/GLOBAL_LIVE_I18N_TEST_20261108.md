# GLOBAL — Canlı kanus.org i18n testi (salt okuma) — 8 Kasım 2026

Tur 1 (~00:45 TSİ): Tüm POST'lar engellendi. hCDN tarayıcı kontrolü geçilemedi ve 16 hücre BLOCKED kaldı.
Tur 2 (~00:46–00:49 TSİ): Tek istisna olarak aynı origin'e giden POST /hcdn-cgi/jschallenge-validate isteğine izin verildi. Normal headless Chrome kullanıldı; stealth, cookie ya da token kopyalama yok. Cookie oturumda tutuldu. Diğer tüm GET/HEAD dışı istekler Fetch ile engellendi ve loglandı. Hiçbir şey yüklenmedi; site/Hostinger değiştirilmedi.

## Hangi paket canlı
- Tarayıcı testi (tur 2) sırasında canlı paket release_us'tı: index.html a828fa73880192c4…a926 ve asset index-uscommerce-i18n-75f56119.js (75f56119…9911), ikisi de curl ile PASS.
- Testten sonra (~00:50 TSİ) yeniden kontrol edildi. Canlıda artık **release_us_fx** var:
  - index.html 4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8 → PASS
  - asset index-uscommerce-i18nfx-5a375f29.js 5a375f29…c5d0 → PASS; cache-control: public, max-age=604800
  - /fx.json 200 9eac8d08317d5e54…e935 → PASS (TCMB 06.11.2026, bülten 2026/209)
- Aşağıdaki 16 hücrelik matris **release_us** içindir. release_us_fx için tarayıcı matrisi NOT TESTED. Döviz görünümü de NOT TESTED, çünkü sitede fiyat öğesi yok.

## Ağ logu (tur 2, tüm GET dışı istekler)
- POST /hcdn-cgi/jschallenge-validate ×2: izin verildi. İlk deneme 405, ardından kontrol geçildi; oturumun geri kalanında tekrar tetiklenmedi.
- POST /api/store.php?resource=contact ×4: ENGELLENDİ (form testi, her dilde 1). Sunucuya ulaşmadı.

## Matris (release_us, localStorage kanus.lang)
| Sayfa | TR | EN | DE | RU |
|---|---|---|---|---|
| / | PASS | PASS | PASS | PASS |
| /us | PASS | PASS | PASS | PASS |
| /contact | PASS | PASS | PASS | PASS |
| /news | PASS | PASS | PASS | PASS |

Her hücre için:
- `<html lang>` seçilen dile eşit.
- Başlık, menü ve gövde metni seçilen dilde. Başka dilde kalmış belirgin metin görülmedi; dil seçici adları ve "Forum", "US", "Journal" gibi markalar/eşdeğer kelimeler hariç.
- Uygulama konsol hatası ve 4xx/5xx asset yok.
- Ham ISO tarih yok.
- Yasak sloganlar render edilmiş DOM'da yok.
- Ekran görüntüleri: tur_live_20261108/{tr,en,de,ru}_{home,us,contact,news}.png ve click_*.png.

Notlar:
- Konsoldaki tek "hata"lar şunlar: TR / sayfasındaki ilk açılıştaki challenge 403/405 kayıtları (kontrol sayfası, uygulama değil) ve /contact'ta bizim engellediğimiz POST'un net::ERR_FAILED kaydı.
- /contact formu: POST engellendi ve hata yolu tetiklendi. Bildirim seçilen dilde çıktı:
  - TR "Sunucuya ulaşılamadı. Bağlantını kontrol edip tekrar dene."
  - EN "Could not reach the server. Check your connection and try again."
  - DE "Der Server ist nicht erreichbar. Prüfe deine Verbindung und versuche es erneut."
  - RU "Не удалось связаться с сервером. Проверьте подключение и повторите попытку."
  → PASS. Sunucunun döndürdüğü hata (4xx/5xx) yolu NOT TESTED.
- Tarih biçimi: 4 sayfada da render edilmiş tarih yok, yalnız "© 2026" var. /news'te tarihli kayıt görünmüyor. Ham ISO yok (PASS), ama Intl biçimi canlıda NOT TESTED.
- Dil seçici tıklaması (/us): TR→EN→DE→RU→TR. Her adımda `<html lang>`, localStorage kanus.lang ve metin değişti → PASS.

## Yasak sloganlar
Statik HTML/JS'te (release_us) ve 16 render edilmiş sayfada yok → PASS. release_us_fx'in render edilmiş DOM'u NOT TESTED.

## Öneri
Geri dönüş gerekmiyor. release_us canlıda tüm kontrollerden geçti. release_us_fx için aynı matris ayrıca koşulmalı.

---
# release_us_fx bölümü (~00:51–00:54 TSİ)
Kurallar aynı. Yalnız POST /hcdn-cgi/jschallenge-validate isteğine izin verildi; diğer GET dışı istekler engellendi ve loglandı. Script: tur_live_20261108/live_fx.mjs. Ham çıktı: live_results_fx.json. Ekran görüntüleri: fx_*.png (20 adet).

## Hash'ler (curl, ?ts ile) — önce ve sonra aynı → PASS
- index.html 4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8 → canlı paket release_us_fx
- index-uscommerce-i18nfx-5a375f29.js 5a375f292604694e6f905a9654353700edf9dc4bb595c5df250354048991c5d0
- fx.json 9eac8d08317d5e5414efb3194ba12658498833ca4ca26fd2db9bf890147ae935

## Ağ logu
- POST jschallenge-validate ×2: izin verildi (ilk deneme 405, ardından kontrol geçildi).
- POST /api/store.php?resource=contact ×4: ENGELLENDİ.
- /fx.json tarayıcı tarafından yüklendi (33 GET) → PASS. i18nfx asset'i yüklendi.
- tcmb.gov.tr'ye 0 istek gitti → PASS.

## Matris (release_us_fx)
| Sayfa | TR | EN | DE | RU |
|---|---|---|---|---|
| / | PASS | PASS | PASS | PASS |
| /us | PASS | PASS | PASS | PASS |
| /contact | PASS | PASS | PASS | PASS |
| /news | PASS | PASS | PASS | PASS |

- Her hücrede `<html lang>` seçilen dile eşit ve metin seçilen dilde.
- Uygulama konsol hatası, 404 asset ve ham ISO tarih yok. Konsoldaki kayıtlar yalnız challenge 403/405 ve engellenen POST'un ERR_FAILED kaydı.
- Dil seçici tıklaması TR→EN→DE→RU→TR: `<html lang>`, kanus.lang ve metin her adımda değişti → PASS.
- /contact hata yolu (POST engelli): bildirim 4 dilde doğru → PASS. Sunucu 4xx/5xx yolu NOT TESTED.
- Yasak sloganlar render edilmiş 16 sayfada yok → PASS.
- Intl tarih biçimi: sayfada tarih öğesi yok → NOT TESTED.
- Döviz görünümü: sitede fiyat öğesi yok → NOT TESTED.

## Öneri
Geri dönüş gerekmiyor.
