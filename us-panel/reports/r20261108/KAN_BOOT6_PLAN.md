# KAN — BOOT6 planı: rezervasyonla sahiplenilen açılış → watchdog discard kapanışı

**Durum: YALNIZ PLAN / UYGULAMA BEKLİYOR.** Hazırlama: 8 Kasım 2026, 01:10 TSİ (Europe/Istanbul, UTC+3). Bu belge boot, seed, secret erişimi/değişikliği, DB yazması, deploy, trafik/firewall değişikliği veya yeni ücretli kaynak izni değildir. Bu görevde yalnız mevcut yerel kaynaklar okundu, GCP control-plane salt-okuma describe yapıldı ve iki yerel rapor düzenlendi. Hiçbir canlı işlem başlatılmadı; secret/token/DSN alınmadı veya yazdırılmadı.

## 1. Hedef ve doğru atıf

Bir tek sentetik hesabın bir kredilik gerçek `us-commerce /v1/chat/completions` isteği ile ds-a100'ü TERMINATED durumundan açmak; kredi rezervasyonu ve lifecycle sahipliğini doğrulamak; isteğin bitişinden sonra **yalnız bağımsız us-watchdog** tarafından `discardLocalSsd=true` ile kapatıldığını göstermek. BOOT5'teki gcloud start/adopt_running=false eksikliği tekrarlanmayacak. Normal akışta gcloud start, usta-wake `/a100/start`, doğrudan model isteği kullanılmayacak.

**Kaynakta önemli ayrım:** `us-watchdog` servisi açılış yapmaz; reconcile yalnız kapatır. `us-commerce` içindeki aynı lifecycle denetleyicisi `acquire()` → `ensure_ready()` → `GCEProvider.start()` çalıştırır. Bu yüzden açılış audit principal'ı **us-gateway@**, kapanış audit principal'ı **us-watchdog@** olmalıdır. “Watchdog sahipli açılış” burada `managed=true` ile ortak lifecycle kaydının oluşturulmasıdır; start audit'inde us-watchdog@ aranması yanlış kabul ölçütüdür. Talep kelimesi kelimesine watchdog Cloud Run sürecinin start etmesi ise mevcut mimari bunu sağlamaz; kod değişikliği bu planın dışında, BOOT6 öncesi KAN.US bu kapsamı teyit etmeli.

## 2. Başlangıç durumu ve önkoşullar

01:10 TSİ civarı control-plane salt-okuma doğrulaması:
- Proje `project-27082a40-6083-4952-9db`, bölge europe-west4; VM `ds-a100`, zone `europe-west4-b`, instance ID `6979174631840438511`: **TERMINATED**, automaticRestart=false. Son start 8 Kasım 00:28:54,318 TSİ; son stop 00:50:59,721 TSİ. Yeni boot olmadı.
- us-watchdog **%100 `us-watchdog-sre-ready-r1`**; SA `us-watchdog@project-27082a40-6083-4952-9db.iam.gserviceaccount.com`. SRE'ye göre trafik başlangıcı **8 Kasım 01:03:16 TSİ** (ham kaynak 7 Kasım 22:03:16Z), taban 00007-fnc + /ready, rate limit, JSON log; image `gateway:2026-11-08-sre-watchdog-ready`.
- us-commerce **%100 `us-commerce-sre-ready-r1`**; SA `us-gateway@project-27082a40-6083-4952-9db.iam.gserviceaccount.com`.
- Trafik iki serviste de revisionName'e sabit, **latestRevision=false (alan yoksa varsayılan false)**. Yeni deploy kendiliğinden trafik almaz. Bu planda deploy veya trafik değişikliği yok; candidate tag URL'si yerine mevcut %100 servis URL'si kullanılacak.
- Scheduler `us-idle-watchdog`: ENABLED, dakika başı (`* * * * *`, yapılandırma saat dilimi UTC), attemptDeadline=300s, retryCount=1. Elle reconcile/job run yok; doğal Scheduler kullanılacak.
- Watchdog revizyonunun env adlarında US_MAIL_PROVIDER yok: incelenen runtime bu durumda mail sender kurmuyor. Uygulama öncesi bu durum yeniden teyit edilmeli; test mail/outbox üretmeyecek.

Policy v5 beklenen değerleri: idle=600, boot=600, request=600, lease=60, capacity=10, max_uptime=1800, daily_runtime=3600, daily_starts=6, preserve_local_ssd=false, adopt_running=false, require_watchdog=true. Watchdog discard opt-in `US_ALLOW_DISCARD_LOCAL_SSD=1` ayrıca teyit edilecek. Secret sürümü değiştirilmez. `latest` referansı, her sıcak process'in aynı policyyi yüklediğinin tek başına kanıtı değildir; revizyon/provenance ve güvenli metadata ile SRE doğrulaması kapıdır.

**GO/NO-GO (SRE + KAN):**
1. İki revizyon ve VM kimliği/durumu yukarıdaki gibi; watchdog doğal reconcile 200 ve heartbeat yaşı ≤120 sn. Kontrol için GPU/model/legacy wake health'e gidilmez; Cloud Run /health, /ready veya control-plane/log okunur.
2. SRE yetkili işletim yüzeyinden yalnız worker'a ait güvenli state özetiyle `starts`, `day_seconds`, `run_since`, aktif lease sayısı, op/phase kontrolü sağlar. Bu görevde canlı DB sorgulanmadı; güncel bütçe **BİLİNMİYOR**, otomatik GO yok. `starts < 6`; kabul için `day_seconds+600 <= 3600`; ayrıca tüm testin bütçeye sığması için tercihen **en az 2100 sn kalan runtime**. daily_runtime=3600 korunur; sayaç sıfırlama/bütçe artırma yok.
3. Kodun günlük penceresi `int(now)//86400`: reset TSİ **03:00**'tedir, gece 00:00 değil. BOOT5'in bu penceredeki süre muhasebesi manuel açıldığı için önceden varsayılmaz. Bütçe yetersizse uygun pencere beklenir, limit aşılmaz.
4. Gerçek müşterilere ait aktif/askıda rezervasyon **0** ve test boyunca gerçek trafik olmayacağı işletim olarak teyit edilir. Paylaşılan worker üzerinde gerçek trafik gelirse yalnız sentetik teste aitmiş gibi devam edilmez/VM zorla kapatılmaz; sonuç FAIL/ABORT ve KAN.US'a eskalasyon. Bu izolasyon sağlanamıyorsa prod testi NO-GO; ayrı ortam/yeni kaynak ayrıca Demir onayı ister.
5. Aşağıdaki sentetik seed/session, VM timer müdahalesi, discard acil stop ve maliyet sınırının açık onay kapsamı netleşmiş olmalı. SRE operator erişimi ve KAN SSH erişimi hazır; SSH için firewall/IAM açılmayacak.

## 3. Tam olarak hangi lisans/kredi ve ortam?

### Önerilen tek fixture (henüz oluşturulmadı)

| Alan | Kesin seçim |
|---|---|
| Test etiketi / jti | `kanus-boot6-20261108-01` |
| issued_licenses.order_id | `TEST-BOOT6-20261108-01` (Shopify siparişi değil) |
| Sentetik hesap | `boot6-20261108-01@tests.invalid` (teslimat/e-posta yok) |
| Lisans biçimi | `us1.<body>.<HMAC-SHA256>`; kind=`sub`, prompts=`1`, jti yukarıdaki, exp=seed zamanı+3600 sn |
| Kredi kaynağı | **accounts.quota=1, used=0, base entitlement**; credit_packs eklenmez, ücretli ürün/order/webhook yok |
| İstek ID | `BOOT6-20261108-01` (tek deneme, tekrar kullanılamaz) |
| Model/istek | `deepseek-v4.1-flash`, tek kullanıcı mesajı `Yalnız OK yaz.`, stream=true, max_tokens=8, think=false, reasoning_effort=none |
| Ortam | Mevcut **production us-commerce-sre-ready-r1**, mevcut Cloud SQL PostgreSQL ve mevcut ds-a100; staging/mock değil |

**Var mı?** İncelenen kaynak/test/raporlarda bu kimliğe ait önceden provision edilmiş üretim lisansı/session kanıtı yok. Canlı DB'de varlık/yokluk sorgulanmadığından “üretimde kesin yok” denmez. Uygulama öncesi SRE yalnız bu tam order_id/jti/email için çakışma/ownership kontrolü yapmalı; varsa seed tekrar edilmez, gerçek veya kökeni belirsiz kayıt ASLA kullanılmaz. Kullanılabilir sentetik fixture doğrulanamazsa yukarıdaki yeni fixture için onay beklenir.

**Kaynak:** release `account_store.py:30–44` imza/exp/prompts denetler; `register():98–121` ayrıca `issued_licenses.status='active'` ister; `login():123–137` session oluşturur. `reserve_prompt():192–199`, `_charge_prompt():237–264` tek krediyi rezerve eder; `finish_reservation():205–223` başarılı isteği completed, başarısızı refunded yapar. `tests/baseline/test_account_store.py:12–26` lisans üretme ve issued row fixture örneği, `test_metered_reservations.py:29–35` kayıt/session örneğidir. Bunlar **geçici SQLite ve test imzalama anahtarı** kullanır; mevcut canlı seed/session değildir. Hazır güvenli prod seed komutu incelenen release'de bulunmadı.

**DB konfigürasyonu:** runtime `US_ENV=production` ve shared PostgreSQL zorunlu; DB secret referansı `us-postgres-dsn`, Cloud SQL instance `us-commerce-pg`, PostgreSQL 17, europe-west4. `US_LICENSE_SECRET` → `us-license-signing`; policy volume → `us-worker-policy`. DB adı/schema/DSN değerini bu görev okumadı; SRE gerçek commerce+watchdog bağlantısının aynı hedef olduğunu secretsiz onaylamalı. Ayrı SQLite/schema fixture mevcut servisin rezervasyonunu authorize etmez; konfigürasyon değiştirilmeden bu gerçek revizyon testi yerine geçmez.

### Seed ve session nasıl hazırlanacak? — yalnız onaydan sonra

**SRE**, hâlihazırda yetkili/sanctioned işletim ortamında, mevcut signing entegrasyonu kullanarak (anahtarı rapora/sohbete/CLI argümanına çıkarmadan) tek lisansı imzalar; yalnız bu fixture için issued_licenses kaydını active yazar; mevcut reviewed AccountStore register/login semantiği ile salt/scrypt hesap ve session üretir. Yeni secret/version, yeni signer, yeni IAM yetkisi veya geçici Cloud Run deploy önerilmez. SRE'nin bunu yapabilen onaylı aracı yoksa erişim etrafından dolaşılmaz; **NO-GO**, araç/seed uygulaması ve kapsamı Demir'e sunulur. Bu belge hazır çalıştırılabilir prod seed aracı varmış gibi davranmaz.

Kayıt/oturum hazırlığı tek fixture'a bağlı dar transaction/işletim işlemi olarak gözden geçirilmeli. Dış HTTP `/account/register` veya `/account/login` körlemesine kullanılmayacak: `check_auth_rate()` kaynakta eski **global auth_attempts DELETE** yapıyor. “Silme yok / gerçek veri değişmesin” koşulu nedeniyle mevcut onayla bu yan etki uygun değil. Test fixture'ını paylaşılan tabloları temizleyen pytest/seed runner ile üretmek de yasak. AccountStore constructor'ının CREATE TABLE IF NOT EXISTS yan etkisi nedeniyle üretimde başlatılan gelişigüzel test script'i değil, SRE'nin önceden incelenmiş dar işletim yolu kullanılmalı.

**Gerekli Demir kararı:** Bu production DB'de yalnız yukarıdaki sentetik lisans, 1 ücretsiz test kredisi, hesap ve session oluşturulması; test sonrası sadece aynı fixture'ı expire/inactive yapma işlemleri için açık izin. Başlatma kotasındaki geçmiş izin, yeni DB entitlement yazması için tek başına izin sayılmaz. Yeni müşteri/ödeme/lisans verisi kullanılmayacak.

## 4. Rezervasyonu kim ve nereden yapacak?

**Birincil yürütücü SRE:** mevcut %100 us-commerce servis köküne, hazırlanmış **sentetik account session Bearer** ile tek `POST /v1/chat/completions`; `X-US-Request-ID: BOOT6-20261108-01`. IAM OIDC token tek başına application session değildir. Kimliksiz istek veya model token'ı da müşteri rezervasyonu yerine geçmez. Session güvenli istemci belleği/özel dosya entegrasyonunda tutulur; curl -v/trace, shell history, komut satırı ve JSON loglara yazılmaz. SSE istemci cold-start boyunca bağlı kalmalı; toplam timeout örneğin 1250 sn, otomatik retry kapalı; hata sonrasında kendiliğinden yeni istek/boot yok.

Kaynak `commerce_http.py:156–171` → `store.reserve_prompt()` → `ManagedModelClient.metered()` → `ManagedResponse` acquire/ensure_ready; `lifecycle_config.py:28–46` shared DB authorization=`store.reservation_active`. ensure_ready `managed=true` ve starts sayacını kendisi oluşturur. Reconcile start etmez.

Cloud Run commerce'in incelenen IAM'i public invoker, uygulama session korumalıdır; bu yol usta-wake :8443'ün iki VPS /32 kısıtına **bağlı değildir**. SRE gerçek güncel ingress/auth durumunu salt-okumayla teyit etmeli. Box'tan uygulama erişimi uygun değilse **HOST**, mevcut izinli iki VPS'den biri üzerinden **aynı us-commerce HTTPS API** isteğini çalıştırır; session transferi yalnız güvenli onaylı kanal ile, mesaj gövdesinde değil. Hangi VPS kullanılacağını HOST teyit eder; firewall/allowlist değiştirilmez.

**Yapılmayacak yol:** HOST'un us-license-hook veya usta-wake `/a100/start` çağırması. Bu manuel açılış/başka lisans defteri yoludur, PostgreSQL `metered_requests` rezervasyonunu kanıtlamaz; BOOT5'i tekrar eder. Hostinger hook, Shopify, ödeme sağlayıcı veya gerçek lisans bu teste dahil değildir. Eski wake zorunluymuş gibi ağ erişimi genişletilmez.

## 5. Gerçek veri koruması ve silmeden kapatma

Testte **gerçek müşteri hesabı/lisansı/session'ı, Shopify siparişi, payment provider, gerçek tahsilat/iade veya müşteri maili kullanılmayacak ve değiştirilmeyecek**. Test “production'da ayrı sunucu” değil; **mevcut ortak DB içinde mantıksal izole fixture**'dır. Tablo satırlarında test flag'i yok; exact order_id, jti, email, license_hash, account_id, session hash ve request_id zinciriyle scope sabitlenir. Seçilen fixture id'leri dışında müşteri satırlarının içeriği okunmaz/değiştirilmez; sıfır gerçek trafik kapısı sağlanamıyorsa bu garanti verilemez ve test yapılmaz.

- issued_licenses hariç Shopify webhook_orders/delivery_jobs oluşturulmaz; webhook endpoint'ine istek yok; credit_products/credit_packs ve gerçek entitlement'lar değişmez. Watchdog sender yokluğu yeniden doğrulanır.
- Başarılı tek istek: sentetik base quota 1 → kalan 0; metered_requests completed. Hata: uygulamanın kendi idempotent finish_reservation(false) yolu aynı sentetik krediyi refunded yapar; belirsiz sonuç manuel incelemede kalır, otomatik replay yok.
- Testten sonra SRE, önceden onaylı dar scope ile **yalnız fixture** issued_licenses.status'ını inactive, accounts.expires'ı geçmiş, o hesaba ait sessions.expires'ı geçmiş yapar. Lisansın doğal exp=1 saat güvenlik katmanıdır; session'ın varsayılan 24 saat ömrü nedeniyle session expiry ayrıca gerekir.
- usage, metered_requests, inference_events ve fixture satırları denetim için korunur; DELETE/DROP/TRUNCATE yok. Kaynakta hazır “test cleanup” API yok; expiry/inactive güncellemesi de ayrı açık DB yazma onayı kapsamındadır. Lisans token'ı DB issued_licenses.token alanında tutulduğu için sır sayılır, rapora alınmaz.

## 6. ds-idle.timer izolasyonu ve geri açma

Aşağıdaki komutlar **şimdi çalıştırılmadı**. KAN yalnız rezervasyonun açtığı VM RUNNING/SSH erişilebilir olduktan hemen sonra (hedef ilk 90 sn, READY beklenmeden), mevcut yetki/ağ ile uygular. Öncesinde `is-enabled` sonucu **enabled** olmalı; masked/disabled ise test başlamaz. VM'i önceden açıp hazırlamak yok. Timer durdurulamazsa isteği iptal et, testi FAIL/ABORT say; idle.py kaynaklı stop'u watchdog PASS yazma.

```sh
# Mevcut izinli KAN terminalinde; yalnız onaylı BOOT6 sırasında:
~/google-cloud-sdk/bin/gcloud compute ssh ds-a100 \
  --project=project-27082a40-6083-4952-9db --zone=europe-west4-b \
  --command='sudo systemctl is-enabled ds-idle.timer; sudo systemctl is-active ds-idle.timer; sudo systemctl show ds-idle.timer -p UnitFileState'
# Üstte enabled teyidinden sonra:
~/google-cloud-sdk/bin/gcloud compute ssh ds-a100 \
  --project=project-27082a40-6083-4952-9db --zone=europe-west4-b \
  --command='sudo systemctl stop ds-idle.timer ds-idle.service && sudo systemctl mask --runtime ds-idle.timer ds-idle.service && sudo systemctl is-enabled ds-idle.timer; sudo systemctl is-active ds-idle.timer ds-idle.service'
```

`is-enabled` masked ve `is-active` inactive çıktılarının exit kodu nonzero olabilir; mask/stop başarısı ve journal ayrı doğrulanır, kör `&&` ile “hata oldu” yorumu yapılmaz. Hem timer hem service runtime mask edilir; o anda çalışan idle.py kalmadığı kontrol edilir. `/opt/ds/idle.py` içeriği/eşik değiştirilmez. **Kalıcı disable kullanılmaz**: test sonunda kapalı VM'e SSH mümkün olmadığı için disabled halde bırakmamak amacıyla mevcut persistent enabled symlink korunur, maske yalnız `/run` içindedir.

VM hâlâ RUNNING iken test iptalinde/erken toparlanmada aşağıdaki geri açma uygulanır; bunun ardından idle.py kapanırsa watchdog testi PASS değildir:
```sh
~/google-cloud-sdk/bin/gcloud compute ssh ds-a100 \
  --project=project-27082a40-6083-4952-9db --zone=europe-west4-b \
  --command='sudo systemctl unmask --runtime ds-idle.service ds-idle.timer && sudo systemctl enable --now ds-idle.timer && sudo systemctl is-enabled ds-idle.timer && sudo systemctl is-active ds-idle.timer'
```

**Başarılı otomatik stop sonrası restorasyonun sınırı:** VM TERMINATED iken “timer active” olamaz ve SSH ile yeniden başlatılamaz. Runtime mask power-cycle'da silinir; persistent enabled hali değişmediğinden **bir sonraki zaten onaylı açılışta timer otomatik etkinleşir**. Ek boot sırf timer kontrolü için yapılmaz. Rapor, “enabled kalıcı durumu korundu; runtime mask sonraki boot'ta kalkacak; canlı active doğrulaması sonraki açılışa ertelendi” diye açık yazar, yapılmamış `enable --now` iddia etmez. Demir/KAN.US, bu runtime-only yöntemi kabul etmeden test GO verilmez. Başarılı stop öncesi timer'ı yeniden başlatmak ölçümü kirleteceği için yapılmaz.

## 7. Zaman çizelgesi, maliyet ve sert durdurma

- **T0:** sentetik rezervasyon/ilk lifecycle start isteği. Bir tek boot; onaylı 8'in **6.'sı**, kalan 3'ten **1** tüketir; sonra 2 kalır. Spot kapasite hatası/başarısız start da ayrıca kaydedilir; otomatik yeni başlatma denemesi yapılmaz.
- T0+~7:12: beklenen READY; boot üst sınırı 600 sn. Lease/request penceresi READY sonrası başlamalı; heartbeat ve lease release izlenir.
- Tek kısa SSE yanıtı tamamen tüketilir, terminal framing ve completed kaydı doğrulanır. İstek biter bitmez başka model/models/wake çağrısı yok. Lifecycle last_activity/release esas alınır; nginx hit'i ile watchdog idle saati karıştırılmaz.
- Beklenen stop API: son lease release/last_activity + **600 sn + dakika başı reconcile fazı (tipik 0–60 sn + jitter)**. Ardından stop operasyonunun bitmesi yaklaşık 1–2 dk. Toplam beklenen **19–21 dk** (gecikmelerle 22 dk). SLA değil; gerçek ölçüm raporlanır.
- Model READY 600 sn içinde gelmezse veya idle eşiğinden sonra iki reconcile boyunca stop yoksa test FAIL olarak işaretlenir, istemci kapatılır; ek model isteğiyle süre uzatılmaz. Hata halinde watchdog doğal olarak kapatabiliyorsa kimliği/auditi izlenir; son güvenlik sınırı beklenmeden gerekiyorsa elle stop yapılır ve otomatik stop FAIL kalır.
- **max_uptime=1800** politika güvenlik sınırı; aktif lease varken `_claim(stop)` durdurmayı erteleyebilir, bağımsız sert saat olarak tek başına güvenilmez. Başlangıçtan **30 dk** sonra hâlâ RUNNING ise (ve gerçek müşteri bulunmadığı teyitliyse) rezervasyon istemcisi iptal edilir, KAN fallback stop'u hemen uygular. **35 dk içinde TERMINATED** hedef/hard operasyonel sınır. API/Spot/platform gecikmesi maliyeti fiziksel olarak garanti edilemez; 32. dk hâlâ durmuyorsa acil eskalasyon, yeni start yok.

```sh
# Yalnız test onayı kapsamındaki acil stop; VM/disk silme değildir.
# Mevcut provider operation STOPPING ise duplicate çağrı üretmeden completion izlenir.
~/google-cloud-sdk/bin/gcloud compute instances stop ds-a100 \
  --project=project-27082a40-6083-4952-9db --zone=europe-west4-b \
  --discard-local-ssd=true --async
~/google-cloud-sdk/bin/gcloud compute instances describe ds-a100 \
  --project=project-27082a40-6083-4952-9db --zone=europe-west4-b \
  --format='value(status)'
```

Bu fallback **yalnız yerel SCRATCH SSD içeriğini** discard eder, persistent boot disk/VM silmez. Önceden Demir'in bu test/fallback discard kapsamını teyidi gerekir; çalıştırıldığı anda watchdog otomatik kapanış hedefi FAIL sayılır. Çalışan VM'de önceki bölümdeki timer restore komutu güvenle yapılabiliyorsa uygulanır; restore/SSH arızası acil stop'u geciktirmez. Runtime-only maske kalıcı enabled durumunu bozmamıştır.

**Maliyet:** önceki boots ölçümü ~0,035 $/dk; 7:12 cold start ~0,25 $, beklenen 19–22 dk **~0,67–0,77 $**; 30 dk ~1,05 $, 35 dk ~1,225 $. **Plan üst sınırı 35 dk / ~1,30 $ ek test tahmini**. Bu $1,30 sağlayıcı tarafından enforce edilen bütçe değil; mevcut SQL/Cloud Run/transfer gibi marjinal kalemler küçük ek fark yaratabilir. Sağlayıcı faturası sonradan doğrulanır. Önceki boot başına ~30 dk/1 $ sınırına göre 35 dk/$1,30 genişlemesi Demir'in açık kararı olmadan varsayılmaz; istenirse daha düşük süreyle erken fallback seçilir. Yeni kaynak, yeni rutin/alarm/bütçe kurulmaz; günlük 3600 sn değişmez.

## 8. Kanıt ve PASS / FAIL ölçütleri

SRE control-plane/audit/Cloud Run loglarını, KAN VM timer journal'ını toplar; tüm rapor saatleri TSİ, ham zamanlar istenirse artefaktta korunur. Token/prompt/response/müşteri verisi loglara alınmaz; sadece sentetik test kimliği ve güvenli state/event metadata.

**PASS için hepsi gerekli:**
1. Başlangıç TERMINATED; request_id `BOOT6-20261108-01` sentetik account'a bağlı tek metered reservation; kaynak base, tek kredi. lifecycle request_admitted → start_requested → model_ready; worker managed=true, tek start intent/operation id. Doğrudan gcloud/legacy start yok.
2. Start audit `v1.compute.instances.start`, beklenen us-gateway SA; Compute requestId lifecycle start_requested detail ile eşleşir. READY ≤600 sn, kısa valid SSE başarıyla biter, metered completed, quota kullanılan=1/kalan=0; lease kalmaz.
3. Stop'a kadar `ds-idle.timer` VE `ds-idle.service` inactive/masked, idle journal'ında stop çalıştırması yok. Son release sonrası idle ≥600; normal reconcile gecikmesi ayrıca ölçülür.
4. **Kapanışın ilk audit kaydı** `beta.compute.instances.stop` (local SSD yolu), principalEmail `us-watchdog@project-27082a40-6083-4952-9db.iam.gserviceaccount.com`, **request.discardLocalSsd=true**. Lifecycle inference_events `stop_requested` detail UUID ile API requestId/operation korelasyonu; Cloud Run **us-watchdog-sre-ready-r1** doğal reconcile çağrısı ve zaman eşleştirmesi. Uygulama JSON access logu tek başına provider stop'u ispatlamaz; audit + event + timer izolasyonu birlikte gerekir. CallerIp/user-agent yalnız yardımcıdır, tek başına SA atfı yerine geçmez.
5. operation.first ve operation.last **aynı operation.id** ile tek stop operasyonunun başlangıç/bitişi olarak sayılır; completion ikinci stop değildir (BOOT5 düzeltmesi). Hata status yok; VM TERMINATED; 4 SCRATCH disk PRESERVED değil, audit discard=true. Persistent boot disk aynen kalır.
6. Gerçek müşteri/ödeme/Shopify verisi işlemi yok, kalıcı timer enabled durumu korunmuş, sentetik fixture expiry/inactive kanıtı ve tüketilen boot/maliyet yazılmış; limitler ihlal edilmemiş.

**FAIL/ABORT:** gcloud/usta-wake ile açılış, yanlış principal ile stop, idle.py stop'u, manuel fallback, Spot preemption/harici stop, discarded=false/belirsiz, budget/heartbeat/auth hatası, READY timeout, gerçek müşteri trafiği, timer izolasyonu kurulamayışı, audit/event eşleştirmesinin eksikliği. Böyle durumda eksik hedef PASS yazılmaz; kredi/VM güvenliği sağlanır, 7. boot kendiliğinden kullanılmaz. STOPPING sırasında ikinci start/kuyruk testi bu boot'un kapsamı değildir; ayrı olarak NOT TESTED kalır.

## 9. Rollback — yalnız SRE, gerekirse açık onay/kapsamla

Önce VM/rezervasyon güvenle sonlandırılır ve durum kanıtlanır; sonra gerekiyorsa SRE aşağıdaki önceden belirtilmiş hedefe trafik döndürür. Bu belge rollback'i şimdi çalıştırma izni değildir:

```sh
~/google-cloud-sdk/bin/gcloud run services update-traffic us-watchdog \
  --project=project-27082a40-6083-4952-9db --region=europe-west4 \
  --to-revisions=us-watchdog-00007-fnc=100
~/google-cloud-sdk/bin/gcloud run services update-traffic us-commerce \
  --project=project-27082a40-6083-4952-9db --region=europe-west4 \
  --to-revisions=us-commerce-00002-pr2=100
```

Pinned trafik nedeniyle yalnız image/revision deploy'u rollback değildir; update-traffic gerekir. Rollback secret latest'i veya DB state'ini geri almaz. Watchdog rollback /ready/rate/JSON yamalarını; commerce rollback ayrıca READY-window lease düzeltmesini kaldırır. İki servise kendi başına deploy yok; sonraki herhangi bir revizyon SRE'nin yamalı tabanı üzerine kurulmalıdır.

## 10. Açık kararlar / devir

- **Demir:** exact sentetik fixture ve 1 ücretsiz base kredi için prod DB seed/session + fixture-only expiry/inactive yazma onayı; silme yok.
- **Demir:** bir kalan boot'un kullanımı, timer'ın runtime-only geçici maskesi ve sonraki açılışta otomatik restorasyon yöntemi, watchdog başarısızlığında local SSD discard fallback; önceki 30 dk/$1 çerçevesinden 35 dk/~$1,30'a genişlemenin kabulü veya daha dar fallback sınırı.
- **SRE:** mevcut onaylı seed/signer/operator yolunun gerçekten mevcut olması, aynı DB hedefi, fixture çakışmaması, güncel runtime/start bütçesi/heartbeat, prod izolasyonu ve ingress/auth teyidi. Hiçbiri bu görevde DB veya secret okuyarak varsayılmadı.
- **Yürütücüler:** SRE seed + us-commerce tek rezervasyon + audit/log; KAN VM runtime-mask/restore/fallback ve final TERMINATED; HOST ancak mevcut izinli VPS'den commerce API istemcisi gerekirse. KAN.US GO ve koordinasyon; onay yerine geçmez.

## Kaynak/provenance

- `/workspace/us-panel/kan/watchdog-fix/release-20261108/{account_store,commerce_http,managed_inference,inference_lifecycle,lifecycle_config,runtime,gce_provider}.py`
- `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/{test_account_store,test_metered_reservations,test_production_runtime}.py` (yalnız kaynak okundu, prod test çalıştırılmadı).
- `/workspace/us-panel/sre/cloudrun/fixes/build-context-us-watchdog/runtime.py`, `us-watchdog/local-acceptance.patch`, `DEPLOY_PLAN.md`.
- `KANUS_CLOUDRUN_DB_ACCEPTANCE.md` (§5–6 DB/auth/secret adları ve son SRE revizyon ekleri); eski `sre/cloudrun/desc_*.json` yalnız tarihsel config, güncel trafik için kullanılmadı.
- `KAN_S4_LIVE_RESULT_BOOT5.md`, `SRE_BOOT5_VERIFY.md`: 659,829 sn idle kapanış ve aynı-operation completion düzeltmesi.
- Bu görevde yapılan salt-okuma gcloud describe: VM, iki Cloud Run servisinin güvenli trafik/SA alanları, Scheduler config ve yalnız iki revizyonun env **adları**. Runtime bütçe/fixture canlı varlığı bu incelemenin kapsamı dışında kaldı.
