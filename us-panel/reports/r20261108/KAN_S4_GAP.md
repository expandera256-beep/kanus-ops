# KAN §4 — GPU durum makinesi ve test boşluk matrisi (r20261108)

Yazan: KAN executor, 8 Kasım 2026 ~00:35 TSİ. Kaynak: `prompt/KANUS_PROMPT_20261107.txt` §4 (satır 139–216).
İncelenen: `/opt/ds/idle.py` (box kopyası `sre/watchdog/idle.py.current`, öneri `idle.py.600`), `reports/KANUS_GPU_REPORT.md`, `reports/KANUS_GPU_COLDSTART.md`,
watchdog kaynak `kan/watchdog-fix/src.tgz` (sha256 19241465…9e77; `inference_lifecycle.py`, `gce_provider.py`, `runtime.py`), `kan/usta-wake/usta-wake.py.new` (:8443 wake yolu).
Canlı durum (salt okuma, 00:25 TSİ): ds-a100 TERMINATED, Spot, automaticRestart=false, 4× SCRATCH SSD `DISK_SAVED_STATE_UNSPECIFIED`; Cloud Scheduler `us-idle-watchdog` her dk → us-watchdog-00005-z97 `/internal/reconcile` 200.

Etiketler: **CANLI** = önceki canlı testle kanıtlı · **SİM** = dry-run/simülasyonla kapsanabilir (bu tur koşuldu, bkz. KAN_S4_SIM_RESULTS.md) · **GPU** = canlı GPU açılışı gerekir · **EKSİK** = kodda/altyapıda yok.

## A. Durum makinesi (§4: STOPPED→STARTING→BOOTING→LOADING_MODEL→READY→BUSY→IDLE→STOPPING→STOPPED; FAILED/RECOVERING)

| Durum/geçiş | Mevcut karşılık | Kapsama | Kanıt / not |
|---|---|---|---|
| STOPPED | watchdog `phase=stopped` (TERMINATED gözlemi) | CANLI + SİM | GPU_REPORT: TERMINATED 06:39:01, 07:47:09; W02, W12 |
| STOPPED→STARTING | `_claim('start')` (`phase=starting`), usta-wake `/a100/start` 202 | CANLI (usta-wake) + SİM (watchdog) | COLDSTART Boot2/3 usta-wake start; W04. Watchdog yolu canlıda hiç başlatmadı (audit'te us-watchdog principal yok) |
| BOOTING | ayrı durum YOK; provider `staging/provisioning` → `phase=staging` | CANLI (zaman ölçümü) / **EKSİK (ad)** | ds-boot RAID+fetch süreleri COLDSTART'ta; durum adı olarak raporlanmıyor |
| LOADING_MODEL | ayrı durum YOK; `phase=running` + `model_ready()` false | CANLI (ölçüm) / **EKSİK (ad)** | ttfb 222–230 sn; kullanıcıya “GPU ısınıyor” (Go proxy, GPU_REPORT satır 9) |
| READY | `model_ready` olayı; idle.py `/run/ds-a100-ready` | CANLI + SİM | GPU_REPORT 06:23:57 ilk yanıt; W04, I02 |
| BUSY | aktif lease (watchdog) / nginx model isteği veya :80/:8080 bağlantı (idle.py) | SİM + kısmen CANLI | W10, I05, I07; canlıda tünel isteği idle sayacını yeniledi (GPU_REPORT 07:29:18) |
| IDLE | `last_activity` + `idle_remaining` (snapshot); idle.py ACTIVITY mtime | CANLI(15 dk) + SİM(10 dk) | GPU_REPORT ~14 dk / ~16,5 dk sonra stop; W08/W09, I03/I04. **10 dk canlı test yok → GPU** |
| IDLE→STOPPING | idle.py `stop(discard=True)` ✔; watchdog `_claim('stop')` ✘ (v5'te reddediliyor) | CANLI(idle.py) / **FAIL(watchdog)** | W19/W20: `refusing to discard local SSD` (gce_provider.py stop) |
| STOPPING→STOPPED, API ile doğrulama | `_observe` TERMINATED → op temizlenir; testlerde gcloud polling | CANLI + SİM | GPU_REPORT discard stop ≈1 dk; W09 |
| FAILED | `reconcile_error` olayı, `LifecycleError`; idle.py 70 dk health-fail stop | SİM (watchdog), GPU (health-fail) | W18, W19; I10. 70 dk yolu canlıda tetiklenmedi (COLDSTART §2) |
| RECOVERING | ayrı durum YOK; yeni revizyon aynı DB'den devralır, op `until` süresi dolunca yeniden claim | SİM | W16; W17 (preemption sonrası yeni istek yeniden açar) |

## B. §4 zorunlu kurallar

| Kural | Kapsama | Kanıt / boşluk |
|---|---|---|
| İstem yoksa GPU kapalı | CANLI + SİM | ds-a100 TERMINATED; W02 (reconcile/snapshot başlatmaz) |
| İlk istem açılışı tetikler | CANLI (usta-wake) + SİM | Boot2/3; W04, U02 |
| Eşzamanlı istemler aynı açılışı paylaşır | SİM | W05 (5 istek→1 start), U02 (8 /a100/start→1 power). **Canlı yok → GPU** |
| Model hazır olana kadar güvenli kuyruk | SİM + yerel | W04/W06 (kapasite 10); GPU_REPORT satır 8–9 (Go proxy kuyruk 32, 15 sn SSE heartbeat) |
| Kullanıcıya gerçek açılış durumu | kısmen | usta-wake `/a100/status` instance+model_ready; BOOTING/LOADING ayrımı **EKSİK** |
| Aktif görev varken kapanmaz | SİM | W10, I07. Not: lease en fazla request_seconds=600 sn yaşar (W18) |
| Son işten sonra 10 dk bekle | SİM (idle.py.600 + v5) | I03/I04, W08/W09. **Canlı 10 dk → GPU** |
| Süre sonunda kapat | CANLI (idle.py 15 dk) / FAIL (watchdog v5) | W19 beklenen FAIL |
| Kapanış Compute API'den doğrulanır | CANLI | testlerde gcloud describe TERMINATED |
| Sağlık kontrolleri uyandırmaz | SİM | W02, W11, I06; runtime.py `/health` lease uzatmaz |
| Kuyruktaki işler etkinliksiz sayılmaz | SİM | W10 (lease varken stop yok) |
| Watchdog başarısızlığında maliyet alarmı | **EKSİK** | Monitoring alert policy listesi boş; budget görünmedi. Watchdog hatada yalnız `reconcile_error` olayı + 503 yazıyor (W19, W22) |
| 10 dk'ya kadar bekleme + doğru ilerleme | kısmen | boot_seconds=600 ≥ soğuk ~7 dk; ilerleme yüzdesi/aşama yok |

## C. §4 test listesi

| Test | Durum | Kanıt |
|---|---|---|
| Soğuk başlangıç | CANLI PASS | COLDSTART Boot2 ≈7:29, Boot3 ≈7:01 (usta-wake yolu); SİM W04 |
| Model yükleme | CANLI PASS | Strata ~2–3 dk, ttfb 222–230 sn |
| İlk yanıt | CANLI PASS | GPU_REPORT: ilk bayt 1,04 sn / toplam 7,7 sn |
| Eşzamanlı istek | SİM PASS · **GPU gerekli** | W05, U02 |
| Görev iptali | SİM PASS · yerel PASS · GPU önerilir | W07; GPU_REPORT satır 9 (Go: cancel upstream context'i keser) |
| 10 dk etkinliksizlik | SİM PASS · **GPU gerekli** | I03/I04, W08/W09; idle.py.600 VM'de değil |
| Otomatik kapanış | CANLI PASS (15 dk, discard) · SİM | GPU_REPORT 06:39:01, 07:47:09 |
| Tekrar açılış | CANLI PASS (Boot2→Boot3) · SİM | W12 |
| Watchdog kurtarma | SİM PASS (devralma) · **FAIL (v5'te stop yolu)** · GPU gerekli | W16 PASS; W19/W20 beklenen FAIL |
| Beklenmeyen kesinti / Spot preemption | SİM PASS · **GPU gerekli** | W17; canlı preemption kanıtı yok. `simulate-maintenance-event` ile denenebilir |
| (ek) Health-fail 70 dk | SİM PASS · canlı tetiklenmedi | I10, I14 |
| (ek) Kota: max_uptime / günlük süre / günlük açılış | SİM PASS (preserve=true ile) · v5'te FAIL | W13, W14, W14b |

## Sayılar
- Durum/geçiş (11): CANLI kanıtlı 6 (en az kısmen), yalnız SİM 3, ayrı durum adı EKSİK 3 (BOOTING, LOADING_MODEL, RECOVERING), FAIL 1 (watchdog stop).
- Kurallar (13): CANLI 4, SİM 6, kısmen 2, EKSİK 1 (maliyet alarmı).
- Testler (10 + 2 ek): CANLI PASS 5 (soğuk, yükleme, ilk yanıt, otomatik kapanış, tekrar açılış); canlı GPU gereken 4 (eşzamanlı, 10 dk idle, watchdog kurtarma/stop, preemption); iptal yerel+SİM.

## Ana boşluklar
1. **Watchdog v5 kota/idle stop çalışmıyor** — `gce_provider.py` `stop()`: local SSD + `preserve_local_ssd=false` → `LifecycleError('refusing to discard local SSD')`, POST yok; reconcile 503 + `reconcile_error`, heartbeat yazılmaz → 120 sn sonra yeni wake'ler de `external watchdog not verified` ile reddedilir (W22). Tek kapatıcı idle.py. Kod düzeltmesi Demir kararını bekliyor.
2. **Maliyet alarmı yok** (alert policy / budget görünmedi).
3. **10 dk idle canlıda yok**; idle.py.600 hazır, VM'e uygulanmadı.
4. **Günlük süre bütçesi açılış sayısından önce bağlıyor**: 3600 sn ile ~3 soğuk döngü (≈7 dk açılış + 10 dk idle) sonra `daily runtime budget exhausted` (W14b). daily_starts=6 pratikte ulaşılmaz.
5. **Lease tavanı 600 sn**: tek istek acquire'dan 600 sn sonra düşer; soğuk açılış 7 dk ise ilk isteğe ~3 dk kalır (W18).
6. **STOPPING sırasında gelen wake** usta-wake'te no-op (U03); istemci yeniden denemeli, kuyruk yok.
7. idle.py: health düşerse (READY sonrası) 70 dk'ya kadar idle stop yok (I14); `ss` hata verirse sürekli “meşgul” (I13) → maliyet riski.
8. BOOTING/LOADING_MODEL/RECOVERING adları ve ilerleme göstergesi yok.

## Güncelleme — Canlı açılış 5 (8 Kasım 00:28–00:51 TSİ, bkz. KAN_S4_LIVE_RESULT_BOOT5.md)
| Satır | Yeni durum |
|---|---|
| 10 dk etkinliksizlik (test/kural) | **CANLI PASS** — idle.py, son istek 21:38:48.73Z → stop 21:49:48.56Z (659,8 sn) |
| Otomatik kapanış / IDLE→STOPPING (idle.py) | CANLI PASS, discardLocalSsd=true, VM SA |
| STOPPING→STOPPED + API doğrulama | CANLI PASS, TERMINATED 21:50:59Z, SSD savedState UNSPECIFIED |
| Soğuk başlangıç / ilk yanıt | CANLI PASS (tekrar): 7:12, ttfb 4,65 sn |
| Watchdog stop (discard, 00006/00007) | Hâlâ **canlı test edilmedi** — VM gcloud ile açıldı, managed değil; gerçek rezervasyonlu wake yolu gerekli |
| idle.py ss/health riskleri (boşluk 7) | Kodda düzeltildi (idle.py.fixed VM'de); canlıda tetiklenmedi |
| Kalan onaylı açılış | 3 |
