# Hook watchdog — yerel/offline taslak
Hiçbir ağ istemcisi, daemon veya deploy kodu yok. `Metrics` verisini dışarıdan alır; gerçek VPS/toplayıcı entegrasyonu NOT TESTED. `check_health(fake_probe, timeout)` yalnız enjekte edilen callback'i çağırır; callback timeout'u kendi uygulamalı. Bu sürümde gerçek HTTP probe uygulanmadı/çağrılmadı.

Dosyalar: hook_watchdog.py, config.example.json, tests/test_hook_watchdog.py; gerçek çıktılar TEST_OUTPUT_20261106.txt ve REGRESSION_OUTPUT_20261106.txt.

Kullanım (bağlantısız entegrasyon API'si):
```python
from hook_watchdog import Config, Watchdog, Metrics, Health, Login
cfg = Config.from_file('config.example.json')
monitor = Watchdog(cfg, clock=fake_clock)
# metrics timestamp'ı aynı saat tabanında olmalı; değerler önceden toplanmış olmalı.
result = monitor.evaluate(metrics)
```
Varsayılan outbox yerel `hook_alarm_outbox.jsonl`; üst dizin önceden mevcut olmalı. Örnek API çağrısı çalıştırılmadı, tests yalnız pytest tmp_path outbox'ına yazdı. Mevcut `../watchdog/watchdog.py:Alarm.notify` doğrudan yeniden kullanılıyor; modül değiştirilmedi. E-posta gönderme kodu yok; mevcut sınıfın `[ds-watchdog] us-license-hook hook.*` konu öneki korunuyor. Her kayıtta TSİ ISO zamanı, sent=false, send_enabled=false.

Varsayılanlar: 30 sn örnekleme; healthz status=200 ve süre <=2 sn; 3 ardışık ihlal; 3 başarı recovery; 900 sn cooldown; telemetri 90 sn sonra stale. Login 300 sn kayan pencere, en az 20 gözlem, nearest-rank p95 >2 sn warning / >5 sn critical. Üç değerlendirme ardışık olmalı; bunlar 30 sn aralıklı, örtüşen kayan pencerelerdir (3 bağımsız 5-dk pencere değildir).

Load1/nCPU >2 warning / >4 critical; CPU PSI some avg60 >20% / >50%; memory PSI full avg60 >1%; IO PSI full avg60 >10%. Basınç alarmlarında ayrıca 300 sn kesintisiz süre şartı var. Eşik eşitliği ihlal değil. None/az örnek/bozuk veri unknown; 3 gözlemde ayrı telemetry warning; unknown aktif olayı iyileşti saymaz. 90 sn'den büyük veri boşluğu ardışıklığı sıfırlar. Aynı timestamp veya 30 sn'den sık değerlendirme sayacı ilerletmez.

Login zaman aşımı gözlemleri p95'te alt sınır olarak tutulur; failed/timeouts sayaçları alarm detayındadır. Her olay ayrı state tutar. İlk olayda kayıt; devam ederse 900 sn sonra reminder; warning→critical yükselmesi cooldown'u aşar. Recovery kaydı 3 başarıdan sonra hemen çıkar; yeniden aynı olay cooldown içindeyse mail bastırılır ama degraded=true kalır. Outbox hatası log + degraded, bir sonraki uygun örnekte tekrar denenir. State bellektedir: süreç yeniden başlarsa cooldown sıfırlanır. Kalıcı state ve gerçek log/HTTP collector ileride entegrasyon işidir.

Testler:
```sh
python3 -m pytest -q /workspace/us-panel/sre/hook_watchdog/tests
python3 -m pytest -q /workspace/us-panel/sre/watchdog/tests /workspace/us-panel/sre/hook_watchdog/tests
```
Sonuç: hook 59 passed, 0 failed; toplam regression 125 passed, 0 failed. Tüm hook testlerinde socket/create_connection yasaklanmıştır. İlk iki başarısız fixture (1970 saat dilimi ve yanlış test path'i) düzeltilmiştir; ilk çıktı TEST_OUTPUT_INITIAL_20261106.txt içinde korunur.
