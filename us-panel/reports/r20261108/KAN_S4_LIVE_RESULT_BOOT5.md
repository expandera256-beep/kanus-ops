# KAN §4 — Canlı açılış 5/8 sonucu (8 Kasım 2026)

Onay: Demir'in 6 Kasım 8 testlik onayı, 5. açılış (sınır ~30 dk / 1 $). Saatler TSİ (UTC+3); SRE için UTC değerleri parantez içinde.

## Özet
- **Durduran: idle.py (VM'in kendi SA'sı)**, son gerçek istekten **659,8 sn** sonra (600 sn eşik + ds-idle.timer 1 dk periyodu). Watchdog durdurmadı (beklenen: VM gcloud ile açıldı, `managed` değil, adopt_running=false).
- Son durum **TERMINATED**, 4 local SSD `DISK_SAVED_STATE_UNSPECIFIED` (PRESERVED değil → silindi).
- Soğuk açılış **7:12**. RUNNING süresi ≈ **22 dk 05 sn**, tahmini maliyet **≈0,75–0,80 $** (önceki ölçümden ~0,035 $/dk). Toplam harcama ≈1,55 $; 8 onaylı açılıştan 5'i kullanıldı.
- Watchdog reconcile boyunca 25/25 HTTP 200, `LifecycleError` logu yok.

## Açılış yolu seçimi
- Watchdog wake yolu yalnız `ensure_ready()` → `acquire()` içinden çalışır; `authorize(reservation)` gerçek müşteri rezervasyonu (lisans+kredi, us-commerce `/v1/chat/completions`) ister. Box'tan yetkili bir rezervasyon yok; reconcile hiçbir zaman start çağırmaz.
- usta-wake :8443 `/a100/start` yalnız iki VPS /32'ye açık.
- Bu yüzden **gcloud ile açıldı**: `gcloud compute instances start ds-a100 --async` (00:28:36 / 21:28:36Z, op bitişi 21:28:54Z). Sonuç: watchdog state `managed` değil → `_claim('stop')` None döner; watchdog bu VM'i sahiplenmez. Watchdog stop yolu canlıda **test edilemedi** (gerçek wake yolu gerekli).

## Zaman çizelgesi
| Olay | TSİ | UTC |
|---|---|---|
| start çağrısı (demirtumen08, gcloud) | 00:28:36 | 21:28:36.77Z |
| STAGING / RUNNING | 00:28:42 / 00:28:58 | |
| idle.py.600 kuruldu (yedek `idle.py.bak-20261108-idle600`, sha d036c8c3…ae78, py_compile OK) | ~00:29:45 | |
| Model health 200 | 00:35:48 | 21:35:48Z → **soğuk açılış 7:12** |
| idle.py.fixed kuruldu (Demir onayı; yedek `idle.py.bak-20261108-prefixed`, sha d52ff329…b089, py_compile OK) | 00:38:31 | 21:38:31Z |
| Tek gerçek istek `POST /v1/chat/completions` (deepseek-v4.1-flash, 200, ttfb 4,65 sn) | 00:38:44–00:38:48 | 21:38:44.06Z – **21:38:48.73Z** (nginx 21:38:48) |
| Sonrası trafik yok | | |
| **stop** (v1, discardLocalSsd=true, Python-urllib/3.10) | 00:49:48 | **21:49:48.56Z** (+659,8 sn) |
| STOPPING gözlendi | 00:49:58 | |
| Aynı stop operasyonunun tamamlanma kaydı (`operation.last=true`; ikinci çağrı değil) | 00:51:02 | 21:51:02.62Z |
| TERMINATED (lastStopTimestamp) | 00:50:59 | 21:50:59.72Z |

Not: nginx'te 21:35:38/43 ve 21:38:44 `GET /v1/models` satırları var (benim hazırlık/model-adı kontrollerim); idle.py bunları da aktivite sayar, son aktivite yine 21:38:48.

## Durduranın kanıtı (Cloud Audit Logs)
- `v1.compute.instances.stop`, callerIp 34.178.14.38 (ds-a100'ün geçici dış IP'si), user agent `Python-urllib/3.10`, `serviceAccountDelegationInfo.firstPartyPrincipal = service-247516932580@compute-system.iam.gserviceaccount.com` (metadata token → VM'in varsayılan SA'sı; principalEmail logda REDACTED), `request.discardLocalSsd=true`, API v1. Watchdog `beta` API ve kendi SA'sıyla çağırırdı → **idle.py**.
- Watchdog (Cloud Run) loglarında bu aralıkta stop yok; reconcile 25× 200.

## idle.py kurulumu
- `ds-idle.service` her tick'te `python3 /opt/ds/idle.py` çalıştırır (OnUnitActiveSec=1min) → yeni dosya bir sonraki tick'te etkin; ayrıca restart gerekmedi.
- Son VM dosyası: `idle.py.fixed` (sha256 d52ff329d4812b99903ace85e831ed467956f1d7d2261aa8c5b4ec1d91fcb089). Box `sre/watchdog/idle.py.current` aynı sha'ya güncellendi; eski box kopyaları: `idle.py.current.bak-20261108-pre600` (dbbfeb5d…, eski 15 dk), `idle.py.current.bak-20261108-prefixed` (d036c8c3…, 600).
- VM yedekleri: `/opt/ds/idle.py.bak-20261108-idle600` (eski 15 dk, dbbfeb5d…), `/opt/ds/idle.py.bak-20261108-prefixed` (600, d036c8c3…).
- Bu açılışta durduran sürüm idle.py.fixed (600 sn yolu; health/ss düzeltme yolları tetiklenmedi).

## Test sırasında değişen bileşenler
- us-watchdog: açılış 00006-88h altında başladı, **00:34'te 00007-fnc**'ye geçti (discard kodu + lease-from-ready). 00007 reconcile hep 200. Revizyon geçişi testi etkilemedi: VM managed olmadığı için iki revizyon da stop denemedi.
- usta-wake 00:36:33'te STOPPING-kuyruk düzeltmesiyle yeniden başladı; bu açılışta usta-wake yolu kullanılmadı, etkisi yok.
- 21:28 ve 21:34 civarında us-watchdog'a 403 (kimliksiz) istekler var; Scheduler değil (reconcile hepsi 200), muhtemelen elle/denetim probları.

## Sonuç
| Hedef | Sonuç |
|---|---|
| Watchdog sahipliğinde açılış | **NOT TESTED** — gerçek rezervasyonlu wake yolu box'tan yok; gcloud ile açıldı |
| idle 600 kurulumu | PASS (önce .600, sonra onaylı .fixed) |
| Soğuk açılış | PASS 7:12 (<10 dk) |
| İlk yanıt | PASS 200, ttfb 4,65 sn |
| 10 dk etkinliksizlik → otomatik kapanış | **PASS** 659,8 sn (≤ 600+60) |
| Kapanış API'den doğrulama + SSD silindi | PASS TERMINATED, savedState UNSPECIFIED |
| Watchdog stop (discard) | **NOT TESTED** (managed değil) |
| Watchdog sağlığı sonrası | PASS reconcile 200 (21:50:02Z, 21:51:03Z…), LifecycleError yok |

## SRE düzeltme notu — 8 Kasım 2026
SRE’nin `SRE_BOOT5_VERIFY.md` bağımsız doğrulamasına göre **8 Kasım 00:51:02,623 TSİ** kaydı (ham log: 7 Kasım 21:51:02.623274Z), 00:49:48 TSİ’de başlatılan **aynı `operation.id`** için `operation.last=true` tamamlanma kaydıdır; ikinci stop çağrısı değildir. Tek farklı stop operasyonu ve tek `operation.first=true` bulunmuştur. Yukarıdaki zaman çizelgesinin önceki “ikinci stop” yorumu düzeltilmiştir.
