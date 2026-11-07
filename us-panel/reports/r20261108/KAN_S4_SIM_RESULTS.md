# KAN §4 — Dry-run / simülasyon sonuçları (r20261108)

Tarih: 8 Kasım 2026 ~00:30 TSİ. GPU açılmadı; kaynak tarball değiştirilmedi (geçici kopya `/tmp/wdsim`).
Test dosyaları ve tam çıktılar: `r20261108/sim/` (sim_lifecycle.py/.out, sim_idle.py/.out, sim_wake.py/.out).

**Özet:** 41 simülasyon testi koşuldu, 41/41 test PASS. Bunlardan 2'si (W19, W20) bilinen hatayı **beklenen FAIL** olarak belgeler (gereksinim açısından FAIL). Ek 4 salt-okuma gcloud kontrolü PASS.

## 1. Watchdog yaşam döngüsü (mock GCE provider, sahte saat, SQLite) — 23 test
Komut: `cd /tmp/wdsim && python3 sim_lifecycle.py` → `Ran 23 tests … OK`
Policy: us-worker-policy v5 alanları (idle 600, boot 600, request 600, lease 60, capacity 10, max_uptime 1800, daily_runtime 3600, daily_starts 6, preserve_local_ssd false).
Mock `FakeGCE.stop()` gerçek `gce_provider.py` kuralını taklit eder; W20/W21 gerçek `GCEProvider`'ı sahte transport ile çağırır.

| ID | Test | Sonuç |
|---|---|---|
| W01 | v5 policy WorkerPolicy doğrulaması | PASS |
| W02 | reconcile/snapshot (health) GPU başlatmaz | PASS |
| W03 | rezervasyonsuz istek uyandıramaz | PASS |
| W04 | soğuk açılış: 1 start, phase stopped→starting→running, ≥420 sn bekleme | PASS |
| W05 | 5 eşzamanlı istek → tek start | PASS |
| W06 | kapasite 10 dolu → 11. istek `worker queue is full` | PASS |
| W07 | iptal → `request cancelled`, release lease'i siler | PASS |
| W08 | idle 599 sn → stop yok (preserve=true) | PASS |
| W09 | idle 600 sn → stop, sonra phase=stopped (preserve=true) | PASS |
| W10 | aktif lease (heartbeat) varken idle dolsa bile stop yok | PASS |
| W11 | snapshot/health idle sayacını sıfırlamaz | PASS |
| W12 | stop sonrası tekrar açılış (starts=2) | PASS |
| W13 | ardışık kısa işlerle 1800 sn → `worker draining for uptime limit`, sonra stop | PASS |
| W14 | daily_starts=6 sonrası 7. start reddi (runtime bütçesi gevşetilerek) | PASS |
| W14b | v5 bütçesiyle 3 soğuk döngüden sonra `daily runtime budget exhausted` | PASS (bulgu: süre bütçesi önce bağlıyor) |
| W15 | watchdog heartbeat 120 sn+ eski → wake reddi, start yok | PASS |
| W16 | yeni controller/revizyon aynı DB'den çalışan VM'i devralıp idle'da durdurur | PASS |
| W17 | çalışırken Spot preemption (TERMINATED) → phase=stopped, run_since=0; yeni istek yeniden açar | PASS |
| W18 | model hiç hazır olmazsa 600 sn'de `request lease expired`, stop yok | PASS |
| W19 | **v5 (preserve=false) idle stop → reconcile hata, VM RUNNING kalır** | PASS = **beklenen FAIL** |
| W20 | **gerçek GCEProvider.stop: local_ssd=True → `refusing to discard local SSD`, POST yok** | PASS = **beklenen FAIL** |
| W21 | preserve=true → `beta/…/stop?requestId=op1&discardLocalSsd=false` | PASS |
| W22 | başarısız reconcile heartbeat yazmaz → watchdog stale (yeni wake'ler reddedilir) | PASS |

Çıktı alıntıları:
```
W13 drained: worker draining for uptime limit | stop at uptime limit, stops= 1
W14b cold cycles before refusal: 3 | daily runtime budget exhausted
W18 request lease expired after 600 s
W19 reconcile -> [{'worker': 'ds-a100', 'error': 'reconciliation failed'}] | VM stays RUNNING
W20 local_ssd= True error= refusing to discard local SSD POSTs= []
W21 beta/projects/project-27082a40-6083-4952-9db/zones/europe-west4-b/instances/ds-a100/stop?requestId=op1&discardLocalSsd=false
```
**Beklenen FAIL nedeni:** `gce_provider.py` `stop()`: `if vm['local_ssd']: if not p.preserve_local_ssd: raise LifecycleError('refusing to discard local SSD')`. ds-a100'de 4 SCRATCH disk var, v4/v5'te preserve_local_ssd=false → watchdog idle/kota stop'u hiç POST etmez; `reconcile()` hatayı yutar, 503 döner, heartbeat yazmaz (W22). Düzeltme (discardLocalSsd=true yolu) Demir kararını bekliyor; kod/image değiştirilmedi.
Sim artefaktı notu: W05'te paylaşılan sahte saat thread açlığı yaratmasın diye ready_delay=20 sn kullanıldı; W10/W14'te tek bir alan test içinde override edildi (secret değil).

## 2. idle.py (sahte nginx log, sahte `ss`, sahte uptime/health, mtime ile saat) — 14 test
Komut: `cd /tmp/idlesim && python3 -m py_compile idle600.py && python3 sim_idle.py` → `Ran 14 tests … OK`

| ID | Test | Sonuç |
|---|---|---|
| I01 | idle.py.600: IDLE=600, FAIL_STOP=4200 | PASS |
| I02 | ilk sağlıklı tick READY+ACTIVITY yazar, stop yok | PASS |
| I03 | 599 sn sessizlik → stop yok | PASS |
| I04 | 601 sn → `stop(discard=True)` | PASS |
| I05 | `POST /v1/chat/completions` log satırı sayacı sıfırlar | PASS |
| I06 | `GET /health` sayacı sıfırlamaz | PASS |
| I07 | gelen :80 bağlantısı meşgul sayılır | PASS |
| I08 | metadata'ya giden (peer :80) bağlantı sayılmaz | PASS |
| I09 | hazır değil + uptime 4000 → stop yok | PASS |
| I10 | health-fail + uptime 4201 → stop(discard=True) | PASS |
| I11 | log rotation (boyut küçülür) sonrası yeni istek görülür | PASS |
| I12 | eski idle.py.current 601 sn'de durdurmaz (15 dk) — fark doğrulandı | PASS |
| I13 | `ss` hata → meşgul sayılır (stop yok) | PASS (risk notu) |
| I14 | READY sonrası health düşerse 70 dk dolmadan stop yok | PASS (risk notu) |

## 3. usta-wake wake yolu (usta-wake.py.new kopyası, mock instance_status/power) — 4 test
Komut: `cd /tmp/wakesim && python3 sim_wake.py` → `Ran 4 tests … OK`
U01 power_verb eşlemesi (TERMINATED→start, STOPPING→'') PASS · U02 8 eşzamanlı `/a100/start` → 1 power çağrısı PASS · U03 STOPPING sırasında start no-op PASS (boşluk: kuyruk yok) · U04 a100 `discard_local_ssd=True` PASS.

## 4. gcloud salt-okuma (dry-run yerine; compute stop/start için resmi --dry-run yok)
| ID | Komut | Sonuç |
|---|---|---|
| G1 | `gcloud compute instances describe ds-a100` | PASS: TERMINATED, SPOT, automaticRestart=False, instanceTerminationAction=STOP, 4× SCRATCH `DISK_SAVED_STATE_UNSPECIFIED` |
| G2 | `gcloud scheduler jobs list --location europe-west4` | PASS: us-idle-watchdog `* * * * *` ENABLED → /internal/reconcile |
| G3 | `gcloud logging read … us-watchdog reconcile` | PASS: 00:22:03 TSİ 200 `us-watchdog-00005-z97` (VM kapalı olduğu için stop yolu denenmedi) |
| G4 | `gcloud compute instances stop --help` | PASS: `--discard-local-ssd` mevcut |
| G5 | `gcloud monitoring policies list` / `billing budgets list` | Boş çıktı → maliyet alarmı **EKSİK** (gap) |
