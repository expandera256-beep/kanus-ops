# GPU ve model akışı — .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Ticari EXE -> Go VPS /v1/chat/completions -> US_MODEL_UPSTREAM. PC SSH butonu/özel model adresi ticari profilde kapalı. VPS'nin 1923 tüneli PC tünelinden ayrıdır.
- PASS — Go proxy global capacity=1 / queue=32; lisans -> admission -> reserve -> upstream -> doğrulanmış bitiş -> settle; read-only health/models kredi düşmez. Yerel 1/5/10/25/50 çağrı testinde upstream etkinlik <=1 ve başarılı çağrı kadar bakiye düşümü doğrulandı.
- PASS — Cold-start istemci 50 dk/sunucu 45 dk; 15 sn SSE heartbeat; gerçek token yokken GPU ısınıyor/kuyruk mesajı, hazır/başarılı denmez. Timeout/cancel gerçek upstream request context'ini keser ve rezervasyonu bırakır.
- PASS — Mevcut inference_lifecycle WorkerPolicy: idle_seconds=900, boot_seconds=600; workers.example.json güncel. Fake-provider 899/900 sn sınırı, aktif/queued lease, son request release, health'in idle resetlememesi testleri geçti. Go dev SSH readiness 10 dk sınırında; bu tur çağrılmadı.
- BLOCKED — Aktif Go model proxy ile GCP provider/start/stop/lease/watchdog entegrasyonu bağlı değil. Python controller kaynakta durur, Cloud Run issuer değildir ve bu pakette dağıtılmadı. VM kapalıysa Go yalnız upstream timeout/hata verir, kendisi uyandırmaz.
- BLOCKED — HOST özel ağ/SSH tüneli ve VM dosya/model hazırlığı; ds-a100 europe-west4-b ve ds-boot/ds-serve gerçek durumu bu tur sorgulanmadı. SSD koruma/IAM/watchdog ve stop akışı canlı kabul ister.
- NOT TESTED — ds-a100, instance-20261030-232245, usta-wake, usta-l4, us-commerce, us-watchdog/us-idle-watchdog güncel durumları; gerçek boot/idle-stop, TTFT/GPU maliyeti ve cold-start testi yok. Önceki ajanın istem-başı çalışma beyanı doğrulanmış sayılmaz.
