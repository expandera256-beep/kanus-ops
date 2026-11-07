# Hook watchdog taslağı — 06.11.2026 TSİ
Durum: YEREL TASLAK UYGULANDI — hook testleri 59/59 PASS; mevcut watchdog ile toplam 125/125 PASS. Dizin git repository değil (git rev-parse: not a git repository); önceki CloudAgent BLOCKED yorumu kapsam hatasıydı. VPS/ağ/deploy/gerçek gönderim yok.

## Girdi ve alarm kuralları (öneri; henüz SLO onayı değil)
- Her 30 sn healthz, 2 sn timeout; art arda 3 başarısızlık kritik, 3 başarı recovery. GPU/model/chat endpoint çağrılmaz. Bu görevde healthz isteği de yapılmadı.
- Login süreleri sadece mevcut erişim/application loglarından; nginx upstream_response_time veya uygulama süre metriği gerekir. Varsayılan access log'da süre yoksa missing_data, sıfır gecikme kabul edilmez. Parola/session/token/email loglanmaz.
- 5 dk kayan pencere, en az 20 login; nearest-rank p95 >2 sn uyarı, >5 sn kritik; 3 pencere sürsün. Başarısız/timeout login'ler ayrı sayaç ve timeout alt sınırı olarak raporlansın; sessizce ayıklanmasın. Örnek azsa insufficient_samples; sentetik login yok.
- load1/nCPU >2 veya CPU PSI some avg60 >20%: 5 dk uyarı; >4 veya PSI >50%: 5 dk kritik. load tek başına CPU atfı değildir (D-state I/O dahil).
- Memory PSI full avg60 >1%, IO PSI full avg60 >10%: 5 dk uyarı; swap-in/out sürekliliği ayrıca ölçülür. Swap doluluğu tek başına aktif thrashing kanıtı değildir.
- Metrik 90 sn eskiyse ayrı stale/missing_data; önceki değerle PASS üretilmez.
- Her alarm türü kendi durumunu tutar; healthz başarısı login/CPU olayını temizlemez. İlk olayda bir outbox kaydı, 15 dk dedup/cooldown, recovery kaydı. Disk/outbox hatası stderr ve degraded durumudur.

## Yeniden kullanım
`../watchdog/watchdog.py:Alarm.notify(where, err, detail)` ve `Config` yeniden kullanılabilir; hook için bağımsız instance='us-license-hook', ayrı outbox yolu, gönderim kapalı. Mevcut Alarm.notify her çağrıda kayıt üretir, dedup sağlamaz: hook tarafında olay durum makinesi şart. failure/success tek sayaçlı olduğu için farklı alarm türleri arasında paylaşılmamalı. Subject mevcut sınıfta ds-watchdog öneki taşır; hook konusu için uyarlama gerekir.

## Outbox taslağı (henüz gerçek outbox'a yazılmadı)
from: support@kanus.org
to: demirtumen08@gmail.com
subject: [hook-watchdog] srv1662550 login_p95 critical
body: TSİ zaman; pencere başlangıç/bitiş; örnek sayısı; p95 ve eşik; load/nCPU; PSI; telemetry age; olay kimliği; log kanıtı. Gizli/kullanıcı verisi yok.
sent=false, send_enabled=false. SMTP/harici gönderim yok.

## Sahte metrik test kabul listesi — PASS
Healthz 2 hata alarm yok/3 hata tek alarm; recovery; p95 nearest-rank sınırları; düşük örnek; timeout dahil; load 2/4 eşiği ve 5 dk süre; PSI; stale veri; bağımsız alarm türleri; cooldown/dedup; outbox hata/degraded; gönderim kapalı; hiçbir gerçek ağ isteği veya VPS süreci/config değişikliği yok. tests/test_hook_watchdog.py içinde 59 test geçti; gerçek çıktı TEST_OUTPUT_20261106.txt. Mevcut watchdog ile birlikte 125 test geçti (REGRESSION_OUTPUT_20261106.txt). Canlı entegrasyon NOT TESTED.

Uygulama ayrıntıları ve semantik: README.md. Varsayılan örnekleme 30 sn; login N değerlendirmesi örtüşen kayan pencerelerdir. Alarm subject mevcut sınıfın ds-watchdog önekini korur. Gerçek outbox değil, test tmp_path dosyaları kullanıldı.
