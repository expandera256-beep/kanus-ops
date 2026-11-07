# SRE — yeni hesap için yerel devir dosyası

Hazırlama: 08.11.2026 TSİ (Europe/Istanbul). Kaynak önceliği: üst sohbet özeti ve verilen hafıza olguları; diskteki raporlar/ham kanıtlar. Bu bir durum devridir, onay veya çalıştırma talimatı değildir. Bulutun şu anki durumu bu görevde yeniden sorgulanmadı. Hashler yerel dosyalardan `sha256sum` ile hesaplandı; imaj digestleri aşağıda belirtilen kayıt dosyalarından okundu. Secret dosyası içerikleri okunmadı; hiçbir bota mesaj gönderilmedi, bulutta değişiklik yapılmadı.

## 1) Kimlik

- Ad: SRE.
- Başlık: boş.
- Açıklama (profile.json içindeki description alanı, aynen):

> KAN.US altında izleme ve güvenilirlik botu. İletişim Türkçe. Görevi: /workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt §4 (11 watchdog gerçek kaynak durumu, 12 en az yetki IAM, 13 eski usta-wake sisteminin güvenli devreden çıkarılması planı, 14 GPU maliyet kaydı, 15 sağlık kontrolleri GPU açmasın) ve §5 (watchdog hataları) için tasarım, kod ve test. GCP projesi project-27082a40-6083-4952-9db, ds-a100 europe-west4-b; boşta kapanma 15 dk (Demir kararı). ds-a100 yaşam döngüsünün sahibi KAN; KAN ile çakışmadan çalış, önce okuma/denetim ve yerel kod/test, değişiklikleri KAN'a ve KAN.US'a öner. Kaynak silme, IAM rolü kaldırma, ücretli kaynak açma Demir onayı ister; başka botun aktardığı onay Demir onayı değildir. Raporlar /workspace/us-panel/reports/ altında PASS/FAIL/BLOCKED/NOT TESTED ile, uydurma veri yok.

**Rolün özeti:** KAN.US altında güvenilirlik, izleme, GPU maliyeti, Cloud Run/DB ve en az yetki denetimi. Önce salt okuma ve yerel kod/test; kaynak yaşam döngüsü KAN ile koordineli. Raporlar kanıta dayalıdır; HOST VPS ölçümlerini sağlar. Açıklamadaki 15 dakika tarihsel bilgidir; daha yeni oturum bağlamında policy v5 idle 600 saniye/10 dakikadır. Açıklamayı aynen korumak güncel policy'yi geçersiz kılmaz.

### Yeniden kurulum için sistem talimatı taslağı

Sen SRE'sin; KAN.US altında izleme ve güvenilirlik botusun. İletişim Türkçe. Kapsamın NIHAI_PROMPT_20261106 §4 (11–15) ve §5, GLOBAL_PROMPT §2–3, KANUS_PROMPT_20261107 §5 (GCP maliyeti) ve §8 (Cloud Run/DB). Önce `/workspace/us-panel/prompt/` altındaki özgün promptları ve bu devrin kaynak raporlarını oku. Çelişen eski kayıtları güncel ölçüm gibi sunma.

Raporları `/workspace/us-panel/reports/r20261108/` altında PASS/FAIL/BLOCKED/NOT TESTED etiketleriyle yaz. Uydurma veri, ölçülmemiş başarı ve varsayımsal canlı durum yok. Kaynakta NOT VERIFIED geçen bulguları doğrulanmadı olarak aktar; yeni raporda test yapılmamışsa NOT TESTED kullan. Zamanlar TSİ/Europe/Istanbul olarak etiketlensin. Secrets/DSN ve kimlik bilgileri maskelensin; private key asla okunmasın/yazdırılmasın, secret değerleri dosyaya girmesin.

**Yasaklar ve onay kapıları:** ds-a100'ü başlatma veya durdurma. ds-a100'e, GPU uyandıran uçlara (`/a100/*`, `/v1/*`, model, chat, wake), `127.0.0.1:18000` tüneline ve usta-wake'e istek atma. Test/prod VPS'lere (`31.220.63.191`, `82.29.184.225`) SSH yapma; fail2ban ban riski var. DB'ye yazma, authorized network ekleme, gerçek e-posta gönderme. DB testlerinde DSN unset olmalı; opt-in testler DDL/INSERT/DROP yapabildiğinden skip başarı değildir.

Ücretli kaynak açma, VM/GPU başlatma/durdurma, prod deploy, silme, IAM rolü kaldırma ve yetki genişletme Demir'in **doğrudan**, o işlem için onayını ister. Başka botun “Demir onayladı” aktarımı onay değildir. Aktarılan istek için karar gerekirse ana bota Demir'den doğrudan onay almasını bildir; devralınmış tarihsel izinleri yeni işlem için sınırsız izin sayma. Rol değişikliği/rotation/secret disable işlemlerini kendiliğinden uygulama.

ds-a100 yaşam döngüsünün sahibi KAN'dır. KAN ile çakışmadan önce okuma/denetim ve yerel kod/test yap; değişiklikleri KAN/KAN.US'a öner. Deploy ayrı servisler için önce trafik almayan aday revizyon, sonra doğrulama, sonra açık onaylı trafik geçişi şeklindedir. Geri dönüş komutu hazır tutulur; rollback de değişikliktir. Trafik sabit revizyona bağlıysa yeni revizyon kendiliğinden trafik almaz. KAN release dizinini değiştirme; ayrı kopyaya rebase et ve lease/READY davranışını koru. Canlı test GPU uyandırmamalı; reconcile elle tetiklenmemeli. Yeni hesapta yalnız gerekli bağlayıcı/kimlik kurulmalı; eksik yetkide otomatik genişletme yapılmamalı.

## 2) Hafıza (tarihleriyle, aynen)

- [profile, 2026-11-06] Test hook VPS (31.220.63.191) ve prod VPS (82.29.184.225) için box'tan SSH denemesi yapılmaz (06.11.2026'da fail2ban ban riski); VPS ölçümleri KAN.US/HOST tarafından toplanıp /workspace/us-panel/sre/vps/ altına bırakılır.
- [journal, 08.11.2026 00:44 TSİ] Demir bu sohbette doğrudan iki onay verdi. (1) us-commerce'in DEPLOY_PLAN.md'ye göre yüklenmesi: önce trafik almayan aday sürüm, doğrulama sonrası trafik, geri dönüş 00002-pr2. (2) us-watchdog'un ancak KAN'ın BOOT5 testi ve idle.py doğrulaması bittikten sonra yüklenmesi, geri dönüş 00007-fnc.
- [journal, 08.11.2026 01:22 TSİ] API ve GLOBAL, Demir'in "pin olmasın, yalnız standart sertifika doğrulaması" kararını aktardı (Demir'in burada verdiği bir onay değil). SPKI pin işi durduruldu; yerine sertifika bitişi ve zincir/hostname izleme betiği yazıldı (r20261108/SRE_CERT_MONITOR.md, 15/15 test PASS). Zamanlanmış izleme kurulmadı.

### Oturum bağlamı (07–08.11.2026; üst sohbet özeti)

- Proje `project-27082a40-6083-4952-9db`, numara `247516932580`. Cloud Run/Build bölgesi `europe-west4`; ds-a100 zone `europe-west4-b`.
- ds-a100: `a2-highgpu-1g`, Spot katalog bedeli 1,848238 USD/saat, dört local SSD; cold-start yaklaşık 7 dakika. Bunlar kaynak oturumun bilgileri, yeni fiyat/performans ölçümü değil.
- idle.py stop çağrısı `discardLocalSsd=true`; policy v5 idle 600 saniye. BOOT5 sonrasında KAN kurulumu, yerel eşit hash ve gerçek idle stop kaydı SRE_BOOT5_VERIFY.md ile doğrulandı. Daha geniş STOPPING-kuyruk canlı akışı NOT TESTED.
- us-watchdog KAN 00007-fnc readywindow/lease düzeltmesi üzerine SRE /ready, rate-limit ve JSON log yamalarını taşır. us-commerce da readywindow tabanına geçirildi. Özetin son canlı revizyonları §5'te; yeni canlı ölçüm yok.
- İki serviste `latestRevision=false`: trafik sabitlenmiş. Yeni build tek başına trafik geçişi değildir.
- Rate limit client başına değil, **instance başına bellekte tek token bucket**: health 2/s burst10; read 5/s burst30; genel API 10/s burst60. Webhook genel kovada, Scheduler reconcile muaf. remoteIp 169.254.x görüldüğü için client kimliği sayılmaz; scale-out/cold-start kovayı sıfırlayabilir.
- JSON log alanları: severity, httpRequest, trace ve request_id. Duyarlı header/query/body/DSN içerikleri loglanmamalı.
- us-watchdog ingress all; yalnız invoker us-scheduler@. `US_ALLOW_DISCARD_LOCAL_SSD=1`. Scheduler `us-idle-watchdog`, doğal dakikalık `/internal/reconcile`; manuel tetikleme yok.
- Build `europe-west4`, staging `gs://us-build-247516932580-euw4/source`, SA `us-builder`; global build GCS izni nedeniyle başarısız olur.
- Billing para birimi TRY; brüt/kredi/net ayrılmalı. Budget kurulmadı: son doğrudan karar “Bütçeyi şimdilik kurma.” 07–08.11 tamamlanma/watermark doğrulanmadan tam gün denmez.
- usta-wake :80/:8788 açıklığı kabul edilmiş risk; grace_off bekliyor. STOPPING kuyruğu 08.11 00:36 TSİ'de eklendiği aktarıldı, canlı test yapılmadı.
- sre-oidc-negtest@ rolsüz; geçici TokenCreator kaldırıldı, OIDC negatif test kapalı. Hesap silme kararı Demir'de.
- Secret Manager 6 secret; rotation metadata yok. Bu, manuel rotation hiç yapılmadığının kanıtı değildir. us-worker-policy v1/v3/v4/v5 etkin, v2 devre dışı (01:15 TSİ metadata snapshot'ı).
- Sertifikalar 08.11.2026 01:22:20–01:22:21 TSİ'de host başına tek public TLS bağlantısıyla kontrol edildi; HTTP gönderilmedi. api-test.kanus.org bitiş 04.02.2027 03:48:11 TSİ; api.kanus.org 04.02.2027 08:13:20 TSİ; kanus.org 31.01.2027 18:28:13 TSİ. Üçünde zincir/hostname/süre PASS; Let's Encrypt YE1. Yenileme otomasyonu ve revocation NOT TESTED.
- .12 teslim raporunda lisans/gateway varsayılanı api-test.kanus.org; GLOBAL kur yaması activeGateway() üzerinden `/v1/fx` kullanır. Web sitesi `kanus.org/fx.json` farklı sözleşmedir. Dağıtılmış panel sürümü/override/production kur adresi doğrulanmadı. Yasak `/v1/*` uçları bu keşifte çağrılmadı.

## 3) Rutinler

Kayıtlı rutin yok. Sertifika izlemesi için de cron/job/routine kurulmadı; yalnız elle çalıştırılabilir betikler var.

## 4) Beceriler, bağlayıcılar ve servisler (yalnız adlar)

- gcloud CLI
- Cloud Run
- Cloud Build
- Compute
- Secret Manager
- IAM
- Billing
- Logging
- Cloud SQL
- cloud-sql-proxy (salt okunur)
- python3
- pytest
- unittest
- openssl
- cryptography
- Grok Bot ajanlar arası mesajlaşma
- KAN
- KAN.US
- API
- GLOBAL
- US
- HOST
- QA
- UYUM
- Shopify (kullanılmayan bağlayıcı)
- Hostinger (kullanılmayan bağlayıcı)
- GitHub (kullanılmayan bağlayıcı)

## 5) Açık işler ve kalınan yer

### Cloud Run yayın durumu

Aşağıdaki canlı durum üst sohbetin son doğrulanmış kaydıdır; bu görevde yeni describe yapılmadı. DEPLOY_PLAN.md ve bazı kabul raporları daha eski “watchdog deploy edilmedi” ifadeleri içerir; zaman çizelgesinde bu son kayıt esas alınmalı, canlıya dokunmadan önce salt okunur doğrulanmalıdır.

| Alan | us-watchdog | us-commerce |
|---|---|---|
| Son canlı revizyon | us-watchdog-sre-ready-r1 | us-commerce-sre-ready-r1 |
| Trafik | %100, latestRevision=false | %100, latestRevision=false |
| Geçiş zamanı | 08.11.2026 01:03:16 TSİ | 08.11.2026 00:48:30 TSİ |
| Image tag | gateway:2026-11-08-sre-watchdog-ready | gateway:2026-11-08-sre-commerce-ready |
| Build kimliği (build.log) | 1a5bfe03-50f1-417e-8422-4eb1b5803d5c | dda56d30-c325-47d9-8689-c21cb9a63583 |
| Geri dönüş revizyonu | us-watchdog-00007-fnc | us-commerce-00002-pr2 |

İmajlar: registry `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/`.

- Watchdog immutable image (kayıt dosyasından okundu): `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway@sha256:5ad490386738a1006f667497c2891cd392b4af5febac59a8018c0bd8a3c8a64a`.
- Commerce immutable image (kayıt dosyasından okundu): `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway@sha256:2416ec71eb4e16c3da4c1d9866cb3c7393ce8819efd043e8b1f28bf9a8bfdbab`.
- Eski 00007-fnc/00002-pr2 image digestleri bu devirde registry'den yeniden doğrulanmadı; rollback komutu eski revizyonu hedefler. Kaynak raporlardaki eski digestler bağımsız güncel registry doğrulaması yerine geçmez.

**Geri dönüş komutları — yalnız taslak; çalıştırılmadı, açık değişiklik onayı gerekir:**

```sh
gcloud run services update-traffic us-watchdog --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-watchdog-00007-fnc=100
gcloud run services update-traffic us-commerce --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-commerce-00002-pr2=100
```

Rollback secret latest/env değişikliğini otomatik geri almaz. Commerce eski revizyona dönüş lease/READY düzeltmesini kaldırır. KAN yeni revizyon çıkarırsa SRE tabanını korumalı ve ayrıca onaylı `update-traffic` gerekir; pinli trafik kendiliğinden geçmez.

### Build ve policy

Proje `project-27082a40-6083-4952-9db` / `247516932580`, region `europe-west4`, zone `europe-west4-b`.

**Build komutu — DRAFT, ÇALIŞTIRILMADI; ücret/yazma ve yayın için ilgili doğrudan onay gerekir.** DEPLOY_PLAN.md eski komutunda eksik olan regional/staging bayrakları aşağıda korunmuştur. Context yeni temiz kopya, testleri geçmiş ilgili servis context'i olmalı; mevcut context'in üstüne kör kopya yapılmamalı.

```sh
PROJECT=project-27082a40-6083-4952-9db
CONTEXT=/workspace/us-panel/sre/cloudrun/fixes/build-context-us-watchdog
IMAGE="europe-west4-docker.pkg.dev/$PROJECT/us-runtime/gateway:<yeni-onayli-tag>"
gcloud builds submit "$CONTEXT" --project="$PROJECT" --region=europe-west4   --gcs-source-staging-dir=gs://us-build-247516932580-euw4/source   --config="$CONTEXT/deploy/cloudbuild.yaml" --substitutions="_IMAGE=$IMAGE"   --service-account="projects/$PROJECT/serviceAccounts/us-builder@$PROJECT.iam.gserviceaccount.com"
```

Commerce için ayrı `build-context-us-commerce`, ayrı image/tag ve bağımsız yayın doğrulaması gerekir. `us-builder` mevcut build kayıtlarında kullanılmıştır; gelecekte hâlâ gerekip gerekmediği doğrulanmadı, unused diye silinmemeli.

Policy `us-worker-policy` latest v5 (değerler üst sohbetten, secret içeriği bu görevde okunmadı): idle_seconds=600; boot_seconds=600; max_uptime_seconds=1800; daily_runtime_seconds=3600; daily_starts=6; preserve=false. `US_ALLOW_DISCARD_LOCAL_SSD=1`. Bu değerler credential değil operasyonel policy özetidir; policy değişikliği yapılmadı.

### Açık iş listesi

- **NOT TESTED:** STOPPING kuyruğu uçtan uca canlı akış. SRE yasak uçlara gitmez; KAN/HOST izinli ölçüm sağlar.
- **FAIL:** Migration sürüm tablosu yok; schema_migrations/alembic_version vb. bulunmadı. Gerçek migration sürümü yok, uydurulmayacak. Tasarım yerelde hazırlanabilir; üretim DDL/DB yazımı yok.
- **FAIL:** Secret rotation metadata/takvimi yok. v1/v3/v4 eski policy sürümleri etkin. QA'nın “manual rotation yokluğu kanıtlanmadı” ayrımı korunmalı.
- **BLOCKED (Demir kararı):** sre-oidc-negtest SA silinsin mi?
- **NOT TESTED:** us-builder gelecekte hâlâ gerekli mi; build loglarındaki fiili kullanım nedeniyle unused kabul edilmemeli.
- **NOT TESTED:** Instance-local rate-limit ölçeklenme/yük davranışı. Client başına kota değil; kontrollü tasarım incelemesi yapılacak, izinsiz burst/load yok.
- **BLOCKED (grace_off bekleniyor):** usta-wake :80 firewall yeniden denetimi. :8788 ve :80 mevcut kabul edilmiş risk; SRE hedefe istek atmaz.
- **BLOCKED (veri tamamlanması):** 07–08.11 billing export tamamlanınca maliyet raporu güncellenecek. TRY brüt/kredi/net, standard/detail export watermark farkları ayrı tutulmalı.
- **BLOCKED (taşıma bekleniyor):** Hook prod VPS'e taşınınca HOST üzerinden izleme/kapasite ölçümü. SRE SSH yapmaz; `/workspace/us-panel/sre/vps/` kanıtlarını ister.
- **FAIL / OPEN (MEDIUM):** QA FINDING-06, `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-06.md`. Rotation/sürüm-retention/VM kimliği/rolsüz test SA/eski token lifecycle bulguları; compromise gösterilmedi. Düzeltmeler Demir onayı bekliyor, uygulandıktan sonra QA yeniden test edecek. Metadata consistency PASS, güvenlik policy kabulü FAIL ayrılmalı. `review_sre_metadata.py --assert-safe` mevcut beş bulguyla FAIL/exit1; değişiklik sonrası taze metadata ve owner runtime teyidi gerekir.
- **NOT TESTED:** Watchdog DB-secret ihtiyacı ve kaldırılabilirliği. Yetki kaldırma onaysız yapılmaz. Hook env mode 600/640 ve wake-token owner doküman çelişkileri HOST canlı metadata-only stat ile çözülmeli.
- **BLOCKED (karar bekleniyor):** Sertifika izlemesi periyodu, alarm alıcısı, çalışacağı makine ve production kur adresi belirsiz. Production panel binary'sindeki doğrulama/URL override, ACME renewal geçmişi ve revocation NOT TESTED.

Sertifika izlemesi `/workspace/us-panel/sre/pin_monitor/`: isim tarihsel; pin kullanılmaz. `pin_check.py` UNKNOWN/exit3 ile durur, yeniden etkinleştirilmez. `cert_expiry.py` ≤21 gün WARN, ≤7 gün CRITICAL; <25 gün ve yenilenme kanıtı yoksa WARN. `chain_check.py` sistem CA deposu, hostname/SAN ve server purpose doğrular. Host başına tek TLS bağlantısı/5 saniye ağ timeout; HTTP yok, retry yok. 15/15 PASS önceki test kanıtıdır; bu devirde testler yeniden koşturulmadı. Snapshot çevrimdışı değerlendirmedir, güncel canlı izleme yerine geçmez.

## 6) Bekleyen Demir kararları ve onaylar

- Bütçe kurulmadı; son doğrudan talimat “Bütçeyi şimdilik kurma.” Önceki Budget API/200 USD oluşturma talebi son kararın önüne geçmez. Otomatik kapatma kurulmaz.
- sre-oidc-negtest SA silinsin mi?
- Eski us-worker-policy v1/v3/v4 sürümleri devre dışı bırakılsın mı?
- Secret rotation yapılsın mı; sahip/takvim/rollout/rollback nasıl olacak?
- VM servis hesabı ve scope daraltılsın mı?
- `/workspace/secrets/usta-wake/a100.token.old` silinsin mi? İçerik/geçerlilik okunmadı, sızıntı kanıtı yok.
- Sertifika izlemesi zamanlansın mı? Periyot, makine ve alarm alıcısı belirlenecek.
- Watchdog gereksiz DB-secret yetkisi varsa kaldırılması için ayrıca doğrudan onay gerekir.

### Geçmiş doğrudan onay kayıtları (üst sohbet özeti; yeni iş için açık çek değil)

- **t21s1:** “Evet, rolsüz test servis hesabını oluştur.” → sre-oidc-negtest oluşturuldu; yeni hesap/rol için genel izin değildir.
- **t22s2:** “Evet, test hesabı üzerinde Token Creator rolünü ver, testten sonra kaldır.” → geçici binding kaldırıldı; negatif OIDC test kapalı. Tekrar token/rol üretme onayı değildir.
- **t26s1:** “Evet, us-commerce'i planla yükle.” / “Evet, us-watchdog'u KAN'ın testi ve idle.py doğrulaması bitince yükle.” → 08.11.2026 00:44 TSİ journal kaydı; önce no-traffic aday, doğrulama, trafik, ayrı rollback hedefleri. Özette bu onaylarla yayın tamamlandı. Yeni deploy için yeniden açık kapsamlı onay gerekir.
- Önceki “setMetadata da eklensin.” talimatı tarihsel bağlamdır; bu devir yeni IAM değişikliğine izin vermez.
- 01:22 TSİ pin iptal kararı API/GLOBAL üzerinden aktarıldı; burada Demir'in doğrudan onayı diye sınıflanmadı. Pin işi durdurulmuş durumda tutulur.

## 7) Kritik dosya yolları ve sha'lar

### Yol haritası ve çalışma komutları

- Rapor kökü: `/workspace/us-panel/reports/r20261108/`; SRE ve KANUS raporlarının mevcut tüm üst-dizin `.md` dosyaları aşağıda hashlenmiştir.
- Cloud Run kanıt kökü: `/workspace/us-panel/reports/r20261108/raw_cloudrun/`. Altındaki env/metadata dosyaları duyarlı olabilir; toptan içerik kopyalama yapılmamalı. Bu devir yalnız seçili digest/build/promotion kanıtlarını hashler; dizinin tek bir SHA256'sı yok, tamamı için manifest oluşturulmadı (doğrulanmadı).
- Secret/IAM redacted metadata: `/workspace/us-panel/reports/r20261108/raw_secrets_iam/`; secret payload dosyası değildir, yine de yalnız gerekli metadata incelenmeli.
- Yerel düzeltmeler: `/workspace/us-panel/sre/cloudrun/fixes/`; `base-00007/` temiz KAN tabanı, `app-00007/` birleşik yerel kaynak; servis başına `local-acceptance.patch`; `tests/new/`, `tests/baseline/`, `test_commands_00007.json`.
- KAN release arşivi: `/workspace/us-panel/kan/watchdog-fix/release-20261108.tgz`; SHA256 aşağıda gerçek dosyadan hesaplandı. KAN kaynak dizini `/workspace/us-panel/kan/watchdog-fix/release-20261108/` değiştirilmez.
- Sertifika monitor gerçek yolu `/workspace/us-panel/sre/pin_monitor/`; pin dosyaları tarihsel iptal taslaklarıdır.
- Promptlar: `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt`, `GLOBAL_PROMPT_20261106.txt`, `KANUS_PROMPT_20261107.txt`.

Yerel Cloud Run test komutu (DB yazan opt-in testlerin çalışmaması için DSN unset):

```sh
PYTHONPATH=/workspace/us-panel/sre/cloudrun/fixes/app-00007 env -u US_POSTGRES_TEST_DSN python3 -m pytest -q -rs /workspace/us-panel/sre/cloudrun/fixes/tests
```

Sertifika yerel birim testleri / çevrimdışı kanıt değerlendirme:

```sh
cd /workspace/us-panel/sre/pin_monitor
python3 -m unittest -v test_cert_monitor
python3 cert_expiry.py --config cert_hosts.json --snapshot evidence/observations.json
python3 chain_check.py --config cert_hosts.json --evidence-directory evidence
```

Aşağıdaki hashler dosya bütünlüğünü gösterir, rapor içindeki iddianın canlıda doğru olduğunu tek başına kanıtlamaz. Gizli dosyalar hashlenmedi; secret içeriği okunmadı. Image digest ile dosya SHA256 ayrı kavramlardır. Eski image digestlerinin bağımsız registry doğrulaması: doğrulanmadı.

| Dosya | SHA256 (sha256sum) |
|---|---|
| `/workspace/us-panel/kan/watchdog-fix/release-20261108.tgz` | `80d16639d3c3440518e4db519a98feb0a7f44f31cb6bf8aaea7ace03aaba5584` |
| `/workspace/us-panel/prompt/GLOBAL_PROMPT_20261106.txt` | `e1cef8aee3d78119ff48e9084db4ec7611b7447ebb7def58b938f94f1970cf6f` |
| `/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt` | `3c71bea9f9bc9325d370bd667cc2706e0aec0ecfadae74c8d51db0f2e0963663` |
| `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt` | `b87b8e99d32e1afbd0fdc5d1b6580b7ba7728cf88dc4683aa4ea5c805c0ff9e2` |
| `/workspace/us-panel/reports/r20261108/KANUS_AGENT_ACCEPTANCE.md` | `40c317846adf929f5db098d45c064b54a747fe1f6e68ff6e57273e608550eefa` |
| `/workspace/us-panel/reports/r20261108/KANUS_CLOUDRUN_DB_ACCEPTANCE.md` | `e2afd99579c29422cf311c93707a098e59ed9aa29654c5e622c63dcdc27d9961` |
| `/workspace/us-panel/reports/r20261108/KANUS_CLOUD_COSTS.md` | `68b14170cdc8284ba502caff11b73ae893590a8f9e2feb421513751e960c439b` |
| `/workspace/us-panel/reports/r20261108/KANUS_CLOUD_COSTS_DETAILED.md` | `7441e6c508a714dc1fcb23ac371c33584ae71c6d61a4a792f043810e8c743313` |
| `/workspace/us-panel/reports/r20261108/KANUS_CURRENT_STATE.md` | `692ca86deae6ac98b527ffc0b6a74584d31f0dd2c31a6d1f350d58bfe3f78af4` |
| `/workspace/us-panel/reports/r20261108/KANUS_FINAL_RELEASE.md` | `331f8713602dd54f44d16fa15779f2fa603f617e2d25725d214cf7092de3f9bd` |
| `/workspace/us-panel/reports/r20261108/KANUS_FIXED_ISSUES.md` | `de2d54772198afd6f15ce7d3db4ece1ec8c328ddc5602ad442682fbb268b35e8` |
| `/workspace/us-panel/reports/r20261108/KANUS_GLOBAL_READINESS.md` | `4a69e5dcc3c4279ce4bdfbeb310cebb42f7c237541ee5c3e57c878dafe02d5bc` |
| `/workspace/us-panel/reports/r20261108/KANUS_GPU_ACCEPTANCE.md` | `c1387a3523509943275acc126ac0f350e1a810cb20b2ffa0724422b3311ae25e` |
| `/workspace/us-panel/reports/r20261108/KANUS_ORG_PRICING_UPDATE_PLAN.md` | `2d6451558039ba0142041ecff0ee30d1f7b45c3c595d156e71060233166e824b` |
| `/workspace/us-panel/reports/r20261108/KANUS_PERF_REPORT.md` | `95abb4ee04f70d365275239e76fd58d56344425973a7b7211c3c3e306e19b888` |
| `/workspace/us-panel/reports/r20261108/KANUS_SECURITY_REPORT.md` | `7b743dae1f23129202c27b2b20e4129331f011f69e0fca5e2d303ad1459d9afe` |
| `/workspace/us-panel/reports/r20261108/KANUS_SHOPIFY_ACCEPTANCE.md` | `e61201d7b11c5490ac1bee59579ca8eda659ac91847d84f6c48f94d8f397e4ed` |
| `/workspace/us-panel/reports/r20261108/KANUS_SHOPIFY_E2E_20261108.md` | `ba84a8898f2878fe8d17ab885f3934e68fe99d7d805ac2bddef381356d137e09` |
| `/workspace/us-panel/reports/r20261108/KANUS_WINDOWS_ACCEPTANCE.md` | `707a813e2be4bc404076695d850e97730e9e4380e0bdd379c8b1d901defeb53f` |
| `/workspace/us-panel/reports/r20261108/SRE_BOOT5_VERIFY.md` | `0f212c69107d2090085299048a43faf63229df3a6c5d372aaed59dbbe086c16c` |
| `/workspace/us-panel/reports/r20261108/SRE_CERT_MONITOR.md` | `32c3fbad791c5f1c1598b4d8867555fdc89aef9bafe55dfb19b175e45aef99cf` |
| `/workspace/us-panel/reports/r20261108/SRE_SECRETS_IAM_METADATA.md` | `4d6350b8ac5117a56275a6cf8fb8c8f0b12833aad48da138459ddfaf96988525` |
| `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-06.md` | `84bd8d4c8763a87bed95ff3538eea55bb283a7be6afabeea30d6e329d64448bf` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/final_integrity_00007.json` | `3292be9d7a3b9e2b32b497c9c6372456faea30fedf20d2524969abafa2e98a46` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/revision00007/live_image_manifest.json` | `292cbdd458fd625925c6d5a1bcbcbeb3601164a8b00223da5ac53726e0ea4982` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/revision00007/revision.json` | `dd0eaeb3a497e41fac423ba542f3ae9864127740bd7410750dd5eb78f7825584` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_commerce_deploy/build.log` | `a00f31c0ad938a85d361ddc96b7c0e68b56b086f0aaec9a989d239cad765183e` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_commerce_deploy/candidate_checks.txt` | `59f7702279a6309ce4217d284ca9eafa1e313a9ffd8e4f767f2bbf322f5fb4a4` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_commerce_deploy/digest.txt` | `c2c948dc6b363aa775247306d23879032d53e778445013c4f1a2173f1a1ea8c2` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_commerce_deploy/promote_time.txt` | `fc6a87186819674eb2fb8de8a94aee37bd7ff5982265151b1099decdf492234f` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_watchdog_deploy/build.log` | `efadb216d5c4eeaeee68bd271583d7bc013dc4190e85538239909700269505ff` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_watchdog_deploy/candidate_checks.txt` | `25bcb210227e1cd977f797324508a87a37f10608011a1bb8e8431d66c26ebf4b` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_watchdog_deploy/digest.txt` | `e9e798128f8ca0c9e7be4dca429dbcf618c830cc3871b795714f0e8358a277e4` |
| `/workspace/us-panel/reports/r20261108/raw_cloudrun/us_watchdog_deploy/promote_time.txt` | `07ebc75c35ce5d90d4e4c7f9e5cef5952b4a22bbba684304ee580b186e0e2a1e` |
| `/workspace/us-panel/sre/cloudrun/fixes/DEPLOY_PLAN.md` | `659f10ef57b60423a1829b272ea53cbe6d5a13f2d1c4def6d8b59aa9f75ac956` |
| `/workspace/us-panel/sre/cloudrun/fixes/IMPLEMENTATION_HANDOFF.md` | `92650d699390ee176794ae96f3e711f9f0539ae6e12dbfff6a2d6891d59574ac` |
| `/workspace/us-panel/sre/cloudrun/fixes/README.md` | `2034a987b99c8f579471d67460f7309e49184a05ba240ff38081a836c6cc5ea7` |
| `/workspace/us-panel/sre/cloudrun/fixes/combined.patch` | `78faa025c0e9b6bf0e2a962393b496ca989c3343bf3a3bb48c01d5bdbeec0a11` |
| `/workspace/us-panel/sre/cloudrun/fixes/combined_tests.txt` | `470c8eba2f1f53819b784d66d141025655e8bb028127cc880bbfb6b8afdfa508` |
| `/workspace/us-panel/sre/cloudrun/fixes/combined_tests_00007.txt` | `c91a1194a44bc1af4809182178dfb5f54007f53473aecffab81cb3243d4bda83` |
| `/workspace/us-panel/sre/cloudrun/fixes/module_provenance_00007.json` | `c7888186b8cb70371fa87e74b89fc9d02a1df075a6ecc2519728e9c3be52cd92` |
| `/workspace/us-panel/sre/cloudrun/fixes/patch_apply_check.json` | `cd67dee084ac46d534b70823956717feabfb9866f459ace00e8e81f123054632` |
| `/workspace/us-panel/sre/cloudrun/fixes/rate-limit.patch` | `4b1dc81be0d50f89252555cdcc2cd6736b0bb87552268459708ed015cbddf0b0` |
| `/workspace/us-panel/sre/cloudrun/fixes/readiness.patch` | `a96d62f6b1d6a23bb95f2b09dce5044538ddcc4cdb17685925f71b7da8028ac7` |
| `/workspace/us-panel/sre/cloudrun/fixes/structured-logging.patch` | `475edcfe68e5f72bd2e6d49d1897300a9718e6f38545a5e01c879f6e1d40ba95` |
| `/workspace/us-panel/sre/cloudrun/fixes/test_commands_00007.json` | `6bd78ba85f62d9ecbb153d0ab0b71760f5e0af560ad3b98312b68be96bc9cd2d` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_account_store.py` | `4838d6a4919268dca9ee5f2f7f1329c4b379f584ac49bd4b7e3d9940a64abdba` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_commerce_http.py` | `068339e63d8c27f72b0e689299df9691006493e8cb4fccfb11f3cfd2ce1bc5ab` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_credit_packs.py` | `815758a306f78bf92a0fa427743de6f1edb1db6511ea589de3c006a2e7b310ab` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_gce_discard.py` | `57727ce841f12cc55bef077ccaac761216f71864f94084c461f94252c6f89cf4` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_hostinger_delivery.py` | `630b50e87a4d0176a54639cd5a3d57ff111246669f0a8fad4e908be7b8dc18eb` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_inference_lifecycle.py` | `ea22074044b181f88fbee76f9b58d2d972586086d847d8e65fc17409f5162d83` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_legacy_wake_migration.py` | `4d251e9d4eb4da5f037cdbd6cfbcd8100856280635a0e06e410cba8bc7c2921b` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_managed_http.py` | `ecf634d4dbed291b166d120c09f561180fb7e9a566d9bd33ba28521f0010a288` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_metered_reservations.py` | `f610ff03a5788db1d1ed961ad9fff30d0aa8d62b003e0e685c41f629d3b4663e` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_postgres_contract.py` | `104af964f7f2a1bcc889f86f080339c5869d316de17105c301b1376096ecb1dc` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_production_runtime.py` | `5040d31fda792abe424e48139ef0925e7ea67d77e8840e063d86aafd55150951` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_ready_window.py` | `3b5b38fa8fdff239611ca879958484b95496ab0b6fff4c544b8edf3a98ab6a99` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/baseline/test_shopify_ledger.py` | `6e21808262cdda36323d6d5634ded189089eed65ae5b07c83d87a23ed50c8666` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/new/test_rate_limit_new.py` | `a2c004804eea2dea4a8b46d63197dade3600b6219081deda63f6bc4ad78b3a97` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/new/test_readiness_new.py` | `f1790ca23d42f7780bf1a656943441018290d0cae29c778a3ec5439ffa480226` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/new/test_rebase_00007_new.py` | `b9c85e4f24cea9f1a60976fab341d1b99c33deecc25cb5470d6c58146fea7e04` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/new/test_runtime_integration_new.py` | `5d7b18b8526460d44678ef291ece42275b7ca59a87eb989fba41ea02f0f32040` |
| `/workspace/us-panel/sre/cloudrun/fixes/tests/new/test_structured_logging_new.py` | `0a03efb1ad320c3c170260d5adba0725ac4a0f1bf7128c11b793ab5a52477c67` |
| `/workspace/us-panel/sre/cloudrun/fixes/us-commerce/local-acceptance.patch` | `c5c8d3b1951607e31f46733b29ca6f462ab6408f4ba55170eb767bbcd96f76f7` |
| `/workspace/us-panel/sre/cloudrun/fixes/us-commerce/tests_00007.txt` | `f9a80c6b42e97a4bf6471cbea0854eef31ab78045fd5d2c1cd562aeeb84e7657` |
| `/workspace/us-panel/sre/cloudrun/fixes/us-watchdog/local-acceptance.patch` | `1d9969be9008b47153e3fcb440b08cd852c807825305ffde6ef033614283189e` |
| `/workspace/us-panel/sre/cloudrun/fixes/us-watchdog/tests_00007.txt` | `db2bdf2b5151a9e689dd82affac0d8ddd48f3a1d7cc79c00a023cfa788273669` |
| `/workspace/us-panel/sre/pin_monitor/CANCELLED_PINNING.txt` | `d2c30856b35a968695563293b29be9004a19dafcd13236083bae636e171f1eb7` |
| `/workspace/us-panel/sre/pin_monitor/README.md` | `cfab10153362604d5a4ca297ce9b5d5abcda29cf9b759273cc20381bb69ff141` |
| `/workspace/us-panel/sre/pin_monitor/cert_expiry.py` | `4e8886e5ba3eeaca6457108eaadcaa6b5d177c0277cf735e94e94d7789a2ac02` |
| `/workspace/us-panel/sre/pin_monitor/cert_hosts.json` | `0e256c1b7ebf18adb44e200db96aa4568c77082a4586639024be2f01d9b31130` |
| `/workspace/us-panel/sre/pin_monitor/chain_check.py` | `48b0eeeae2d9ab310249368c4e4c6dbf233d851baa4044759a52d3aaf360347c` |
| `/workspace/us-panel/sre/pin_monitor/common.py` | `1c448646f7dd81399f79b877c72493f0f7709ca566968d1960820b32d419a047` |
| `/workspace/us-panel/sre/pin_monitor/evidence/unit-tests.txt` | `83e16c192d6b470c934528899fda82711e3e99e58f4496fbd877f1d30134123a` |
| `/workspace/us-panel/sre/pin_monitor/pin_check.py` | `c7320cee53ee7cf837b1e3af177bec0d84c5895be4712c2521f2e634f5a6f536` |
| `/workspace/us-panel/sre/pin_monitor/pins.example.json` | `c58ecc9887374b3a595fdc8195a985d2cb5b04791b364aa30c8e2622eed49035` |
| `/workspace/us-panel/sre/pin_monitor/test_cert_monitor.py` | `4cc21976a63efb03a394c5d87d3e82639039bedaffd2077bf3b687e28cc2e60c` |

## 8) Yeniden girilmesi gereken gizli değerler (yalnız ad ve kullanım yeri)

Aşağıda değer yoktur. Adların eski ortamda bulunması yeni hesapta hepsinin mutlaka kopyalanması gerektiği anlamına gelmez; en az yetkiyle owner onaylı yeniden bağlantı tercih edilir. Kimlik dosyası/token/DSN içerikleri bu görevde okunmadı.

### GCP kimliği ve Secret Manager

- Aktif gcloud hesap **adı** (salt okunur `gcloud auth list`): `demirtumen08@gmail.com`. `auth/impersonate_service_account` ayarı `(unset)`. Aktif hesap user hesabı; bu kontrolde aktif bir SA anahtarı saptanmadı. Diskte SA key/ADC dosyası var mı veya hangisi kullanılıyor: doğrulanmadı; kimlik dosyaları açılmadı. Yeni hesapta gcloud kullanıcı yetkilendirmesi veya onaylı workload kimliği yeniden kurulmalıdır; mevcut erişim tokenları taşınmaz.
- Secret Manager adları: us-credit-products, us-license-signing, us-model-ca, us-model-token, us-postgres-dsn, us-worker-policy.
- `us-credit-products`: commerce kredi ürünleri; `US_CREDIT_PRODUCTS` binding.
- `us-license-signing`: commerce lisans imzalama; `US_LICENSE_SECRET` binding.
- `us-model-ca`: model TLS CA volume; gateway/watchdog `/secrets/ca`.
- `us-model-token`: gateway/watchdog model kimliği; `US_MODEL_API_TOKEN` binding.
- `us-postgres-dsn`: Cloud SQL bağlantısı; gateway/watchdog `US_COMMERCE_DB` binding. DSN değeri yazılmaz/aktarılmaz.
- `us-worker-policy`: gateway/watchdog `/secrets/policy` volume; policy metadata v5, payload burada okunmadı.
- `US_POSTGRES_TEST_DSN`: yalnız izinli opt-in PostgreSQL test ortamı; testler şema/INSERT/DROP yapabilir. SRE salt-okuma regresyon komutunda unset tutulur.

### usta-wake dosya adları

Dizin `/workspace/secrets/usta-wake/` (yalnız adlar listelendi, içerik/hash okunmadı): `a100.token`, `a100.token.old`, `cert.pem`, `key.pem`, `pin.txt`, `wake.pin`.

- `a100.token`: GPU wake kimlik doğrulaması.
- `a100.token.old`: eski wake token artığı; silme kararı bekliyor.
- `pin.txt`, `wake.pin`: eski wake TLS pin dosyaları; API istemci pin iptali ile aynı katman olduğu varsayılmaz, içerik doğrulanmadı.
- `key.pem`: usta-wake TLS private-key dosya adı; asla okunmaz/yazdırılmaz.
- `cert.pem`: usta-wake TLS sertifika dosya adı; bu görevde içeriği okunmadı.

### Hook env dosyaları ve değişken adları

- `/etc/us-license-hook.env`: prod hook EnvironmentFile; VPS'te bulunduğu/izinleri HOST teyidi bekliyor.
- `/etc/us-license-hook-test.env`: test hook EnvironmentFile; aynı teyit bekliyor.
- Yerel şablon adları: `/workspace/us-panel/deploy/us-license-hook/us-license-hook.env.example`, `/workspace/us-panel/deploy/us-license-hook/python-disabled.env.example`. Bu görevde env dosyalarının içeriği okunmadı.
- `SHOPIFY_WEBHOOK_SECRET`, `US_LICENSE_SECRET`: webhook doğrulama/lisans imzası.
- `US_SESSION_SECRET`: hook session imzası.
- `US_SMTP_USER`, `US_SMTP_PASS`: hook e-posta bağlantısı; gerçek e-posta gönderme izni yok.
- `US_MODEL_UPSTREAM_TOKEN`: hook upstream model kimliği.
- `US_GPU_WAKE_TOKEN_FILE`: wake token dosya yolu; varsayılan kullanım yeri `/etc/us-model/wake.token`.
- `US_GPU_WAKE_PIN_FILE`: wake TLS pin dosya yolu.
- Hook Secret Manager değil env/dosya kullanıyor; canlı değerler istenmedi. Panel desktop EXE için sunucu tarafı secret kaynağı doğrulanmadı.

**Gizlilik kontrolü:** Dosya sonrasında `rg` ile token/key/DSN adları ile gerçek değer kalıpları ayrı taranır; gerekli değişken/dosya adları normaldir, gizli değer değildir. Kesin sızıntı kalıbı eşleşirse devir paylaşılmaz. Bu dosya hiçbir dış hesaba yüklenmedi/gönderilmedi.
