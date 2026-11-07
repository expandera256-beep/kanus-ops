# ds-a100 soğuk açılış notu (panel uçtan uca test için)
Güncelleme: 6 Kasım 2026, 05:30 (UTC+3)

- Kapı: usta-wake, http://34.41.186.170. /health ile hazır olup olmadığı görülür, GPU'yu açmaz.
- GPU'yu uyandırmak için: /v1/... isteğine `X-Usta-Worker: a100` header'ı ya da `?worker=a100` eklenir. Bu, ds-a100 kapalıysa onu başlatır.
- Soğuk açılış sırası:
  1. VM başlıyor: ~1 dk
  2. boot.sh 4 boş NVMe ile RAID0 kurup ext4 ile biçimlendiriyor: ~1 dk
  3. fetch-weights ~257 GB'ı p27082-ds-istek bucket'ından çekiyor: ~10 dk. Kopmada kaldığı yerden devam ediyor, dosya başına 6 deneme var.
  4. ds-serve (Strata) dosya kontrolü ve model yükleme: ~8 dk
  Toplam: ~20-30 dk. usta-wake isteği en fazla 50 dk bekletiyor; health 70 dk'da hâlâ başarısızsa VM'i durduruyor.
- Panel tarafı: ilk istekte zaman aşımı ≥ 45 dk olmalı ya da "GPU ısınıyor" durumu gösterilmeli.
- Boşta kapanma: 15 dk trafik olmazsa idle.py VM'i durduruyor ve NVMe siliniyor; bir sonraki açılış yine soğuk.
- Spot VM olduğu için Google tarafından kesilebilir; kesilirse sonraki istek yeniden soğuk açılış başlatır.
- Doğrudan 34.7.143.62 adresi firewall ile kasıtlı olarak kapalı; erişim yalnızca usta-wake üzerinden.
