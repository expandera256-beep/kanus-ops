# release_us_fx3 (YEREL, YÜKLENMEDİ)
Canlı ee14acd6 + düzeltilmiş fx bileşeni (R4B FX01/FX02: TSİ takvimi bugün-5..bugün; kuruş/BigInt yarımda sıfırdan uzağa).
Sitedeki fiyat metinleri (usCard/appsIntro/usIntro; TR'de TL, EN/DE/RU'da TRY) düz sözlük metni; `data-kanus-try` bağlantısı yok, bileşen onlara dokunmaz — uydurma bağlama yapılmadı.
Yeniden üretim: `python3 site/us/build_us_fx3.py`.
