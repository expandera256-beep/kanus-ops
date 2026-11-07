# SRE BOOT5 bağımsız salt-okuma doğrulaması — 08.11.2026

**BOOT5 otomatik idle kapanışı: PASS. İstenen üç kapı (BOOT5 / idle600 / 00007 temiz): PASS, aşağıdaki kapsam sınırlarıyla.** Canlı watchdog sahipli wake/stop ve usta-wake STOPPING kuyruğu bu açılışta doğrulanmış değildir; bunlar PASS sayılmadı. Bu rapor yeni deploy izni değildir.

Denetim zamanı: 08.11.2026 00:58:57 TSİ. Gözlenen log üst sınırı: 2026-11-08T00:55:02.334751+03:00 (Europe/Istanbul). Kaynak rapor `KAN_S4_LIVE_RESULT_BOOT5.md`; bağımsız API kanıtları `raw_cloudrun/boot5_verify/`.

## Başlangıç / güvenlik / değiştirilen dosyalar

KAN BOOT5'in son istekten yaklaşık660 sn sonra idle.py ile discard edilerek durduğunu bildirdi. Bu denetim Cloud Logging/Compute control-plane describe ile kontrol etti. **GPU/VM/VPS/usta-wake'e doğrudan HTTP/SSH isteği yok; GCP yazması/deploy/başlatma/durdurma yok.** `gcloud compute instances describe` yalnız kullanıcının istediği yönetim API'si salt-okumasıdır. Model isteği atılmadı. Token/DSN alınmadı. Yalnız bu rapor, kabul raporu ve ham kanıt/özet JSON dosyaları yazıldı.

## PASS / FAIL / NOT TESTED tablosu

| Kontrol | Sonuç | Kanıt / sınır |
|---|---|---|
| Stop çağrısı Activity Audit | PASS | operation.first=true, v1.compute.instances.stop |
| Caller / kaynak IP / discard | PASS | VM SA ile aynı caller, 34.178.14.38, discardLocalSsd=true |
| Son istek sonundan stop'a süre | PASS (gözlenen600+60 penceresi) | **659.829246 sn**; son istek zamanı KAN kaynağından |
| idle.py.current / fixed IDLE | PASS (yerel ve KAN kurulum kaydı) | Her ikisi600 sn, SHA256 aynı; canlı VM dosyası yeniden okunmadı |
| Mevcut VM durumu | PASS | **TERMINATED**, lastStop 08.11.2026 00:50:59.721000 TSİ |
| SSD korunmaması | PASS | Audit discard=true;4 SCRATCH disk savedState UNSPECIFIED, PRESERVED değil |
| Watchdog00007 reconcile / hata | PASS | **21/21 HTTP200**;5xx0, ERROR+0, LifecycleError/refusing0 |
| Scheduler genel pencere | PASS | **27/27 HTTP200**, elle çalıştırılmadı |
| 00:34 negatif403 korelasyonu | PASS (zaman/yol/UA) | Parent'ın rolsüz-SA/anon akışına uyumlu3 curl isteği |
| 00:28'in aynı parent testi olduğu iddiası | FAIL (aynı test değil) | 00:32–34'ten önce; SRE'nin ayrı00:28:05 negatif health testiyle eşleşiyor |
| 00:51:02'de ikinci stop API çağrısı iddiası | **FAIL / RAPOR DÜZELTMESİ** | Aynı operation.id'nin operation.last=true tamamlanma kaydı; ikinci çağrı değil |
| Watchdog-managed wake/stop | NOT TESTED | KAN açılışı gcloud ile yaptı; kendi rezervasyonlu wake akışı kullanılmadı |
| STOPPING kuyruğunun canlı işlevi | NOT TESTED | İkinci audit kaydı kuyruğa alınmış start isteği değildir; usta-wake çağrılmadı |
| İlk model200 / soğuk açılış tekrar testi | NOT TESTED (bu denetimde) | KAN raporunda PASS; bu executor GPU/model uçlarını tekrar çağırmadı |

## Gerçek stop kanıtı

Filtre: `cloudaudit.googleapis.com/activity`, gce_instance instance_id6979174631840438511, method compute.instances.stop, BOOT5 penceresi. Exact argv ve zaman sınırları `audit_stops.meta.json`; ham girişler `audit_stops.json`.

- İlk event: **08.11.2026 00:49:48.559246 TSİ**, `operation.first=true`.
- Method **v1.compute.instances.stop**, API v1.
- principalEmail/principalSubject: **247516932580-compute@developer.gserviceaccount.com**. `vm_disks_service_accounts.json` VM'in mevcut SA yapılandırmasında aynı hesabı gösteriyor; karşılaştırma redaction placeholder üzerinden yapılmadı.
- Delegation firstPartyPrincipal: `service-247516932580@compute-system.iam.gserviceaccount.com`.
- callerIp **34.178.14.38**; callerSuppliedUserAgent **Python-urllib/3.10,gzip(gfe)**.
- request.discardLocalSsd **true**. Status hata alanı boş; stop operasyonu başarılı tamamlanmış.
- Bu kimlik, IP, v1 URL ve Python UA, yerel idle.py'nin metadata-token + v1 stop(discard=true) koduyla ve KAN'ın canlı kaydıyla uyumludur. Cloud Audit process adı tutmaz; **idle.py atfı bu kanıtların birleşimidir**, audit doğrudan Python dosya adını yazmıyor. Çağıran us-watchdog@ değildir. Kaynak IP'nin o andaki VM dış IP'si olduğuna ilişkin geçmiş eşleme KAN raporundandır; audit IP'yi bağımsız doğrular.

### Önemli düzeltme: iki audit kaydı ≠ iki stop çağrısı

| Event | TSİ | İşaret | operation.id |
|---|---|---|---|
| Başlatma | 08.11.2026 00:49:48.559246 TSİ | first=true | operation-1794088188584-65fb6b52d2419-9b0266cc-c2399d50 |
| Tamamlanma | 08.11.2026 00:51:02.623274 TSİ | last=true | **aynı operation.id** |

Yalnız **1 farklı stop operation.id** ve1 operation.first kaydı var. Son kayıtta request.discardLocalSsd yok; bu normal tamamlanma logudur. Dolayısıyla KAN raporunun “STOPPING sırasında ikinci stop çağrısı” yorumu bu audit kanıtıyla desteklenmiyor. lastStopTimestamp **08.11.2026 00:50:59.721000 TSİ**, operation.last kaydı bundan yaklaşık2,90 sn sonra; bu ikinci çağrı çelişkisi değil, tamamlanma kaydının zamanlamasıdır. Başka bir ikinci çağrı olduğunu ispatlayan ayrı operation.first/id bu filtrede yok.

## Süre hesabı / neden yaklaşık660 sn?

Son gerçek request **bitişi** KAN raporunda **08.11.2026 00:38:48.730000 TSİ**. Stop event'i **08.11.2026 00:49:48.559246 TSİ**:

**00:49:48.559246 −00:38:48.730000 =659.829246 sn =10 dk59.829246 sn**.

İstek başlangıcından ölçülse ~664,50 sn çıkar; idle hesabında aktif istek süresini yanlışlıkla eklememek için bitiş alındı. Stop →TERMINATED **71.161754 sn**. Bu kapanış operasyonu süresi600 sn idle eşiğine dahil değildir.

`idle.py.current` davranışı:
- nginx yeni model satırı veya aktif model bağlantısı görülürse **o kontrol anında** ACTIVITY mtime touch edilir (log satırının eski timestamp'i activity'ye yazılmaz).
- READY sonrası `age() > IDLE`, IDLE=600 ile stop(discard=True) çağrılır.
- `ds-idle.timer`1 dakikalık periyot bilgisi KAN'ın kurulum kaydındandır; bu denetimde VM systemd dosyası yeniden okunmadı.

Bu nedenle gerçek isteğin bir sonraki tick'te görülmesi ve eşik kontrolünün tick'e oturması gecikme yaratır; gözlenen **600+59.829 sn** bu örnek için beklenen kontrol-periyodu etkisiyle uyumludur. Ancak kod activity'yi **tespit anında** yenilediğinden,600 sn'yi tüm koşullarda gerçek istek bitişinden itibaren sert bir üst sınır diye sunmak doğru değildir; activity-tespit fazı ve strict `>`/scheduler jitter ek faz gecikmesi yaratabilir. Bu tek ölçüm genel bir SLA/p99 üst sınırı kanıtlamaz.

## idle.py sürüm ve hash doğrulaması

Her iki dosyada satır14: `IDLE = 10 * 60` = **600 sn**.

| Yerel dosya | SHA256 |
|---|---|
| sre/watchdog/idle.py.current | d52ff329d4812b99903ace85e831ed467956f1d7d2261aa8c5b4ec1d91fcb089 |
| sre/watchdog/idle.py.fixed | d52ff329d4812b99903ace85e831ed467956f1d7d2261aa8c5b4ec1d91fcb089 |

Dosyalar byte-eşit; hash KAN'ın VM'e kurduğunu bildirdiği final hash ile aynı. Önceki900 sn snapshot artık güncel dosya değil. Bu denetim canlı VM içeriğini SSH ile tekrar hashlemedi; canlı kurulum bağlantısı KAN'ın kaydıdır. Kapanışın gözlenen yaklaşık660 sn süresi600 ayarıyla tutarlıdır;900 ile tutarlı değildir. health-failure/ss-error özel yolları bu tek testte sınanmamış kalır.

## Watchdog / OIDC403 incelemesi

Watchdog log penceresi **08.11.2026 00:28:00–00:55:02 TSİ**. 00006 açılış kısmında,00007 **00:34** sonrası.00007 doğal POST /internal/reconcile **21/21 HTTP200**; tüm pencere27 reconcile, Scheduler bitiş27/27 HTTP200.00007 5xx0, ERROR+0; LifecycleError/refusing taraması0. `watchdog_logs.json`, `scheduler_logs.json`, `boot5_summary.json`.

| HTTP403 zamanı | Yol | User-Agent | Revizyon |
|---|---|---|---|
| 08.11.2026 00:28:05.507731 TSİ | /health | KANUS-SRE-readonly-acceptance/20261108 | us-watchdog-00006-88h |
| 08.11.2026 00:34:04.590119 TSİ | /health | curl/8.14.1 | us-watchdog-00007-fnc |
| 08.11.2026 00:34:04.831577 TSİ | / | curl/8.14.1 | us-watchdog-00007-fnc |
| 08.11.2026 00:34:05.041474 TSİ | /health | curl/8.14.1 | us-watchdog-00007-fnc |

- **00:28:05.507731** kaydı, `http_watchdog_00006.json` içindeki SRE kimliksiz /health testinin00:28:05.819630 cevap kaydıyla ve aynı özel UA ile eşleşiyor. Parent'ın00:32–34 penceresine ait değildir.
- **00:34:04.590119 /00:34:04.831577 /00:34:05.041474** üçlüsü `/health`, `/`, `/health`, curl UA: `oidc_negtest_sa_20261108.txt` rolsüz SA iki isteği +kimliksiz health dizisiyle uyumlu. Kanıt dosyasının00:32:45 başlığı test başlangıcı; token8.denemede alınmış, gerçek HTTP kayıtları00:34'te. Request logu tek başına SA principal'ını göstermediği için SA eşlemesi kaynak test kaydıyla birlikte yapılır.
- **Hiçbiri Scheduler reconcile hatası değildir**; bu pencerede reconcile403 yok.

## Mevcut kaynak durumu

Control-plane describe: ds-a100 **TERMINATED**. lastStart **08.11.2026 00:28:54.318000 TSİ**; lastStop **08.11.2026 00:50:59.721000 TSİ**.4×375 GiB local SSD savedState **DISK_SAVED_STATE_UNSPECIFIED**, PRESERVED değil; discard=true audit'i ile birlikte disk koruma yapılmadığı doğrulanır. Persistent boot disk ayrı kaynaktır ve silindiği iddia edilmez.

## Deploy ön koşulları / kalan eksikler / onaylar

| İstenen üç koşul | Durum |
|---|---|
| BOOT5 otomatik idle kapanışı PASS | **Karşılandı** — bağımsız stop/terminate audit kanıtı |
| idle.py600 | **Karşılandı** — güncel yerel iki hash aynı, KAN canlı kurulum kaydı aynıhash |
| us-watchdog00007 temiz | **Karşılandı** —21/21 reconcile200, hata yok |

**Bu üç koşul EVET.** Fakat daha geniş `fixes/DEPLOY_PLAN.md` içinde ayrıca beklenen **canlı STOPPING-kuyruk doğrulaması** bu BOOT5 ile kapanmış sayılamaz; watcher-owned reserve/wake/stop/discard zinciri de **NOT TESTED**. BOOT5 manuel gcloud açılışıdır, adopt_running=false akışı nedeniyle watchdog'un VM'i durdurmaması hata sayılmaz. Yeni watchdog deploy'u ancak ana ajanın kapsam/onay kararıyla; **bu executor deploy yapmadı ve yeni izin çıkarmıyor**.

Uygulanan çözüm yalnız rapor düzeltmesidir: ikinci-stop yanlış yorumu ve403 kaynak eşlemesi netleştirildi. Hiçbir canlı kaynak ayarı değişmedi.
