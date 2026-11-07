# QA — §12 Güvenlik Denetimi ve §13 Test Kuralları

Rapor: 08.11.2026 00:37 TSİ (Europe/Istanbul). Kaynak prompt **`prompt/KANUS_PROMPT_20261107.txt` tamamı okundu** (583 satır). §12`:459–472` tam **14 madde** içerir; eksik bırakılmış ek madde yok. QA kapsamı §12/§13tür; diğer bileşenleri düzeltme/deploy sahipliğini devralmadık.

## Yönetici özeti

**14 gereksinim: PASS8 /FAIL4 /BLOCKED2 /NOT TESTED0.** Dört FAIL satırı **üç ayrı kök nedene** karşılık gelir; F02 hem API token destination-binding hem SSRFyi etkiler. Ek F04 tarihsel IP-only **DESIGN GAP**, FAIL değildir; yeni account limiter kaynakta/yerelde mitigasyon sağlar, canlı etkinliği belirsizdir.

Yalnız yerel otomatik14-case çıktısı: **10 PASS /3 FAIL /1 BLOCKED /0 NOT TESTED** (`qa-raw/all-cases-results.json`). Üst matris token-rowa F02yi uygular ve restore-local-PASSı production kabulü bilinmediği için BLOCKED sayar. Sayılar farklı kapsamlar olduğundan birleştirilmez. Uzak negatif kontrol: **8 PASS**,8 sıralı HTTP, concurrency1,120s timeout. Hiç gerçek hesap/lisans credentialı, gerçek veri/ödeme/mail, signed webhook, VPS write/restart veya GPU/model çağrısı yok.

### Sürüm ve kanıt sınırı

- Güncel Go kaynak `work2/kiyas/term/release.go:11` **2026.11.06.11**. Restore sonrası dist ve `dist/2026.11.06.11/` henüz yok; native artifact veya dağıtılan binary eşliği doğrulanmadı. Eski RELEASE_MANIFEST daha eski sürümleri gösterir; yeni kaynak yerine esas alınmadı.
- Canlı api-test health200, sahte/missing credential reddi gözlendi; bu source/deployed hash eşliğine veya tam üretim kabulüne PASS değildir.
- Source/harness hashleri: `qa-raw/QA_SOURCE_EVIDENCE_SHA256.json`. Çalışma ağacı paylaşılmıştır; pinned release snapshotı değil. Hashler kaynak eşliğini sonradan karşılaştırmak içindir, kodun dağıtıldığı iddiası değil.
- Promptta **PANEL DEV/API/HOST/SRE/UYUM isimlerini bileşenlere atayan bir ownership tablosu yok** (tam metin incelendi). Kullanıcının verdiği örnek bot adlarıyla bileşen-temelli yönlendirme önerildi; koordinatör kesin sahibi doğrulamalı. QA §12/§13 sorumluluğu kullanıcının görev atamasından gelir. UYUMa hukuki uygunluk bulgusu uydurulmadı.

## Tam §12 matrisi

Aşağıdaki Go dosya yollarının kökü aksi belirtilmedikçe `work2/kiyas/term/`, Python `work2/kanus/`; log kökü **L=`reports/r20261108/qa-raw/`**. Komut kısa biçimi `python3 /workspace/us-panel/qa/audit12/run.py <arg>`.

| §12 maddesi | Durum | Yeniden incelenen kaynak (dosya:satır) | Otomatik reproducer | Gerçek kanıt / sınır |
|---|---|---|---|---|
| Kullanıcı kimlik doğrulaması | **PASS** | `accounts.go:48–91,362–373,453–485` | `--case authentication` | L/authentication.log;remote-negative.json: Signed session expiry/reset/version/verification yerel PASS; uzak fake/missing401. Authenticated yeni-host kabulü yapılmadı. |
| API token güvenliği | **FAIL** | `accounts.go:53–91;repair_loop.go:262–274` | `--case api_tokens; --case ssrf` | L/api_tokens.log;ssrf.log: Forgery/invalid-token denial PASS; default repair client307 ile farklı porta synthetic Authorizationı gönderdi (F02). Bu helperın aktif GUI routeuna bağlandığı gösterilmedi. |
| Secret Manager | **BLOCKED** | `cmd/us-license-hook/main.go:30–31,57,104,132; kanus/deploy/gcp_apply.py:64–78;gateway.service.yaml:30–47` | `--case secret_manager` | L/secret-manager-source-check.json;secret-scan.json;secret-candidate-classification.json: Env+SecretKeyRef kaynak bindingleri mevcut; taramada confirmed hardcoded real credential yok. Güncel IAM/version rotation/runtime binding kanıtı yok; secret/private/env/data dosyaları okunmadı. |
| Webhook HMAC | **PASS** | `licensesrv/webhook.go:34–45,72–93` | `--case hmac` | L/hmac.log;remote-negative.json: Constant-time HMAC; local bad/missing/reversal shop/topic testleri PASS; remote missing/wrong401. Gerçek webhook sırrı kullanılmadı. |
| SQL injection | **PASS** | `kanus/account_store.py:123–136,162–170;kanus/storage.py:21–35` | `--case sqli` | L/sqli.log;python-tests.log: Gerçek AccountStore/SQLite parameter-binding ile email/session payloadları auth alamadı; accounts tablosu ve quota sağlam. PostgreSQL gerçek DSN/migration ayrıca NOT TESTED. |
| Komut enjeksiyonu | **PASS** | `terminal.go:32+;safe_exec.go:223–258` | `--case command_injection` | L/command_injection.log;panel-audit.log: Terminal shell/meta/python-c argv reddi PASS. İzinli proje kodunun OS erişimi sandbox değildir, F01 ayrıca FAIL. |
| Path traversal | **PASS** | `file_tools.go:44–105,127–138` | `--case path_traversal` | L/path_traversal.log: ../,absolute outside,Windows separator ve symlink escape fileTools üzerinden reddedildi. OS child bypassı F01dir. |
| SSRF | **FAIL** | `repair_loop.go:262–274; karşı kontrol model_proxy.go:46–58 ve gpu_pin.go:140–173` | `--case ssrf` | L/ssrf.log;panel-audit.log: Local307 sinke POST/project body ve Authorization geçti (F02). Model proxy/GPU clientlarının no-redirect kontrolleri ayrı mevcut; metadata/network saldırısı yapılmadı. |
| Rate limiting | **PASS** | `accounts.go:257–290,344–360;login_limiter.go:10–16,30–80;accounts.go:362–373,466–468,484` | `--case rate_limit` | L/rate_limit.log;../qa-20261106/ratelimit-evidence.txt: Account10/900s normalized email/local rotating-IP testi PASS. Legacy/nginx IP-only DESIGN GAP (F04), FAIL değil; ek kaynak mitigasyonu var, yeni deploy etkinliği henüz doğrulanmadı. |
| Yetki yükseltme | **PASS** | `accounts.go:53–100,110–118,205–218;credits.go:27–44,72–85` | `--case privilege_escalation` | L/privilege_escalation.log;go-test.log: Application credential/ownership düzeyinde farklı license reservation settle reddi ve idempotency PASS; root/kernel/container escape audit kabulü değildir. |
| Dosya izolasyonu | **FAIL** | `safe_exec.go:17–19,176–208,223–258;file_tools.go:44–105` | `--case file_isolation` | L/file_isolation.log: AynıUID child Python testin ürettiği sibling workspace dosyasını okudu (F01). Trusted single-user davranışı biliniyor; multi-user/untrusted izolasyon gereksinimi karşılanmıyor. |
| Hassas loglar | **PASS** | `licensesrv/privacy_mask.go:10–40;mail.go:63–70,97–110;accounts.go:384–387` | `--case sensitive_logs; check_logs.py` | L/sensitive_logs.log;log-leak-check.json: DryRun/fakeSender masking ve GPU redact testleri PASS; raw log token/canary scan PASS. İlk deny-all allowlist fixture hatası düzeltildi; ürün leak değildi. |
| Yedekleme | **FAIL** | `deploy/host-hardening/backup/backup-us-license-hook-hourly.sh:30–43,49;licensesrv/store.go:153–185` | `--case backup` | L/backup.log;go-store-backup-restore.log: DRAFT shell hiç kaynak yokken success0 döndürdü (F03). Local SQLite/Go JSON yedek+restore PASS; bu scriptin productionda kurulu olduğu iddia edilmiyor. |
| Geri yükleme | **BLOCKED** | `licensesrv/store.go:101–150,153–185;kanus/storage.py:53–58;restore-us-license-hook-store.sh:21–33` | `--case restore` | L/restore.log;go-store-backup-restore.log: Sentetik DB ve Go Store offline ayrı instance restore/kota/history/lock PASS. Yeni-host/CloudSQL backup+restore/rpo/rto/installed cron kanıtı yok; üretim içerikleri açılmadı veya restore edilmedi. |

Her maddede kaynak ve çalıştırılabilir test vardır; Secret Manager static testinin kaynak kontrolleri true olsa da üretim IAM kontrolü olmadığı için PASS ilan edilmedi. Production restore BLOCKED satırı, sentetik restore testlerini koşmadığımız anlamına gelmez.

## Bulgu listesi / routing

| Bulgu | Önem | Sorumlu bot önerisi | Dosya | Repro |
|---|---|---|---|---|
| F01 — same-UID child sibling workspace read | HIGH, koşullu multi-user/untrusted-code | PANEL DEV; HOST/SRE sandbox | `qa-findings/FINDING-01.md` | `run.py --case file_isolation` |
| F02 — repair helper307/project body+Authorization redirect | MEDIUM, latent/helper | PANEL DEV | `qa-findings/FINDING-02.md` | `run.py --case ssrf` |
| F03 — DRAFT backup all-missing input exit0 | MEDIUM, operasyonel/DRAFT | HOST;SRE | `qa-findings/FINDING-03.md` | `run.py --case backup` |
| F04 — IP-only rate limit | MEDIUM DESIGN GAP,source-mitigated/live-unverified; FAIL değil | API;HOST | `qa-findings/FINDING-04.md` | `run.py --case rate_limit` |

**Kapsamı aşan iddia yok:** F01 gerçek başka kullanıcının/production müşterinin dosyasını okumaz; iki workspacei test yaratır. F02 mevcut GUI/public-route reachabilitysi gösterilemeyen repair helperın default istemcisidir; diğer model proxy/GPU no-redirect layerleri mevcut. F03 deploy edilmemiş olabilecek DRAFT scriptte üretilmiştir; production backupın bozukluğu kanıtlanmadı. Bulgu dosyalarında beklenen/gerçek, kök neden, file:line, önerilen fix ve §13 handoff ayrı yazılmıştır.

## §13 uygulaması / test kanıtı

1. Kök nedenler source review ile saptandı.
2. Her açık bulgu için `qa/` altında otomatik negative regression oluşturuldu.
3. **Ürün fixleri QA tarafından uygulanmadı.** Önceki görev ayrımı uyarınca product/deploy dosyaları ownerlara ait; bu atama bulgu/repro/fix önerisi/routing istiyor. Prompt §13 adım3–5in product-fix+fix sonrası retest bölümü owner bekliyor; tamamlandı gösterilmedi.
4. Testler gerçekten çalıştı; üç açık negative regression FAIL olarak korundu, xfail yapılmadı.
5. Genel regresyonlar güncel kaynakta çalıştı:
   - `go test ./... -count=1 -timeout=180s`: **PASS**, L/go-test.log exit0.
   - `go test -race ./... -count=1 -timeout=12m`: **PASS**, L/go-race-generous.log exit0; paket süreleri57.114s/9.092s/225.294s. İlk240s bütçeli race koşumu bcrypt enstrümantasyonunda timeout oldu (L/go-race.log), race bugı olarak sayılmadı; daha geniş bütçeli gerçek tekrarın kanıtı korundu.
   - `go vet ./...`: **PASS**, L/go-vet.log exit0.
   - Python unittest discover: **109 test,40 SKIP,exit0** (L/python-tests.log). SKIPler gerçek PostgreSQL/GPU/model opt-inleri içerir; tamamlanmış kabul sayılmaz.
6. Kaydedilen değişiklikler sadece `qa/audit12/` ve audit rapor/kanıt dosyalarıdır; commit/push/ticket/message/deploy yok.

### Harness-only düzeltmeler (ürün güvenlik bulgusu değil)

- İlk sensitive-log hedefli koşum `cleanenv` deny-all SMTP allowlist nedeniyle sendera hiç ulaşmadı ve testte FAIL verdi. Asıl koşul `len(logs)==0` idi; raw credential leak yoktu. Harness yalnız DryRunSender/fakeSender kullanılan sentetik fixturea allowlist verdi; gerçek sender/SMTP kapalı. Tekrar **PASS** (`sensitive_logs.log`); ilk kanıt `sensitive_logs-initial-fixture-error.log` korunur. Yanlış pozitif security finding açılmadı.
- Go backup fixture ilk denemede tarihsel Nov6 mint zamanı ile bugünün Settle zamanını karıştırıp local reservation expiryye takıldı. Yalnız test now=time.Now ile düzeltildi; yeni sentetik test **PASS**, ilk hata `go-store-backup-restore-initial-clock-fixture.log` korundu.

## Secret Manager / hardcoded secret / hassas veri

`scan_secrets.py` /workspace kaynak/config kod dosyalarını taradı; yalnız path/line/rule çıktı. `/workspace/secrets`, private/key/env dosyaları, müşteri DB/backupları, arşivler, symlinkler içerik olarak taranmadı; var olan prod backup/customer data açılmadı. Bu sınırlar açıkça `secret-scan.json`da kayıtlıdır; “her gizli dosyanın içi kontrol edildi” iddiası yok.

Taramadaki adaylar: iki QA synthetic password fixtureı ve third-party cryptography parserında35-byte tek satırlık PRIVATE KEY header markerı. Sonuncusu AST constant metadata ile doğrulandı (`secret-candidate-classification.json`), private-key payloadı değil. **Confirmed hardcoded real secret bulunmadı**; bu heuristic/source-scope sonucu tam DLP garantisi değildir. Secret directory/file metadata stat-only toplandı; public `.pub` veya dizinlerin644/755i private-key leak olarak raporlanmadı. Değerler raporda yok.

Go hook sırları envden yükler (`main.go:30,57,104,132`); native Secret Manager entegrasyonu bu hook dosyasında yok, HOST provisioningine bağlı. Cloud Run kaynak YAML `secretKeyRef`+secret volume bindingleri mevcut; GCP operator kodu generated/private state ve scoped secretAccessor bindinglerini kullanır. **Kod deploy modülü import/execute edilmedi**, secret version access/IAM write yapılmadı. Güncel service-account least privilege, rotation/version policy, Secret Manager erişim audit logu ve yeni host env-provisioning permission kanıtı SRE/HOSTtan gelmeden production Secret Manager kabulü BLOCKEDdir.

Kendi ham loglarımızda token shape/provider-token/private-key header ve known synthetic canary taraması **PASS**, hit listesi boş (`log-leak-check.json`). Değer çıkarılmaz; bu unknown tüm secret türlerinin yokluğunu matematiksel olarak kanıtlamaz. Hiç gerçek credential test sürecine yüklenmedi.

## Backup/restore güvenlik sınırı

- SQLite AccountStore: sqlite backup API → ayrı yeni DB → checksum/integrity_check → credential/session/quota roundtrip. Kaynak DBdeki sentetik charge sonrası restored DBnin önceki10 kredisini koruduğu doğrulandı.
- Go hook Store: temp synthetic order/license, local1-credit settle, offline JSON snapshot, ayrı pathte OpenStore/reopen →499 quota/history/validity ve exclusive store lock korundu. Gerçek lisans veya müşteri verisi yok; blob/key stdouta basılmadı.
- Backup shell: yalnız test kopyasının fixed filesystem pathleri temp fixturea yönlendirildi; production script/deployment değiştirilmedi. All-missing input başarı hatası gerçek shell mantığından geliyor.
- Gerçek üretim backups için sadece tarihsel `HOOK_MOVE_PROD_VPS_20261106.md:75–77`, `deploy/host-hardening/backup/RESTORE.md:13–14` okundu. Bunlar eski cron/backup iddiasıdır; yeni82.29.184.225 host/CloudSQL için backup age, hashes,retention,encryption/offsite,RPO/RTO veya restore drill kanıtı değildir. **Production restore/backup kabulü BLOCKED**; HOST/SRE read-only kanıt sağlamalı. Gerçek backup içeriği boxa kopyalanmadı, secrets/env dosyası açılmadı.

## Uzak kontrolün tam sınırı

08.11.2026 **00:30:16–00:30:19 TSİ**: health200;missing-auth401;fake-license401;fake-session401;missing-HMAC401;wrong-HMAC401;encoded traversal route400;command-query auth gate401.8 request,tek thread,max1,120s timeout,no retries; rate/load burst yok. POSTlar yalnız geçersiz token veya boş-body/yanlış-signature denialdır; register/reset/verification/valid webhook/model çağrısı yok. Path/command auth-gate probe tam OS izolasyonu kanıtı sayılmadı.

Bu tur uzak limiter11-deneme testi veya gerçek account credential kullanımı yapılmadı. Önceki HOST nginx/effective-source-IP kanıtı F04te uygun tarih/kapsamıyla yeniden değerlendirildi;32 QA loginin tamamı16 satırlık dar pencereden çıkarılmadı.

## Kalan ihtiyaçlar / sonraki adım

1. **PANEL DEV** F01 multi-user OS boundary politikası ve F02 default redirect fixini ele almalı; source-only/helper reachabilityyi doğrulamalı. **HOST/SRE** isolated UID/container/mount politikasını belirlemeli. QA fixes-applied diye raporlamadı.
2. **HOST/SRE** F03 required-input/zero-success backup exitini düzeltmeli; yeni host ve CloudSQL için read-only backup inventory,son başarı/hash/izin/retention,alarm ve sentetik restore drill kanıtı sağlamalı. Prod restore/yıkıcı işlem bu görevin parçası değil.
3. **SRE/HOST** current secret bindings/IAM/rotation/provisioning permission kanıtı (metadata-only, değer yok).
4. **API/HOST** .11 deployed hash+account-limiter enabled attestation ve kontrollü gerçek farklı-source-IP11.login kabulü; kaynak mitigation zaten yerelde PASS. Daha fazla public yük burada üretilmedi.
5. Koordinatör routing sahiplerini doğrulamalı: verilen promptta bot ownership tablosu yok. Hiç dış mesaj/ticket gönderilmedi.
6. Artifact/native Windows, gerçek PostgreSQL, GPU/gerçek model ve production backup restore bu raporun dar yerel PASSlarından kabul edilmiş sayılmaz; ilgili owner/fixture/izin kanıtları gerekir.

## Tekrar koşum
```bash
python3 /workspace/us-panel/qa/audit12/run.py
python3 /workspace/us-panel/qa/audit12/run.py --case file_isolation
python3 /workspace/us-panel/qa/audit12/run.py --case ssrf
python3 /workspace/us-panel/qa/audit12/run.py --case backup
python3 /workspace/us-panel/qa/audit12/run.py --case restore
python3 /workspace/us-panel/qa/audit12/check_logs.py
```
Tam local audit açık3 negatif regression nedeniyle exit1 verir; bunu PASSa çevirmek için ürün fixleri ve owner sonrası yeniden kanıt gerekir. Uzak probe komutu READMEde ayrı opt-indir, otomatik tekrar önerilmez.


## Koordinatör güncellemesi sonrası PROD-host güvenlik gatei (08.11.2026 00:39 TSİ (Europe/Istanbul))

- HOSTa göre api-test **08.11.2026 00:13 TSİ** itibarıyla **82.29.184.225 / srv2133566 PROD VPS**ye taşındı. Bu, “sadece test VM” varsayımıyla yük üretmeyi haklı çıkarmaz. Bu auditte8 sıralı negatif probedan başka public HTTP yok; auth/model state değişikliği yok.
- HOSTun bildirdiği çalışan binary **pre-.11**, hash `7d72f449…56f0`; beklenen .11 hash `12772d12…b47`. Bunlar bu workerın yeni hostta ölçtüğü hashler **değil**, koordinatörün aktarımıdır. Source .11 bulguları dağıtılan pre-.11 binaryye doğrudan atfedilmez.
- **DNS/IP/TLS/health precheck denendi:** `qa-raw/new-host-precheck.json`,08.11.2026 **00:38:36 TSİ**. Box resolver **198.18.0.1** döndürdü;82.29.184.225i bağımsız gösteremedi. Bu adres proxy/synthetic DNS olabilir; authoritative public DNSin yanlış olduğuna FAIL verilmez. **Precheck BLOCKED**, actual peer/serving-IP bağı doğrulanmadı. Fail-closed gate nedeniyle HTTP yapılmadı; curl-to-IP, başka proxy/route veya retry ile aşılmadı. HOST bağımsız authoritative DNS+TLS peer IP+health read-only kanıtı sağlamalı.
- `qa/audit12/precheck.py` artık gelecekteki `run.py --remote` koşumunun zorunlu önkoşuludur; gate PASS olmadan8 negatif probe başlamaz. Mock resolver guard testi **PASS,0 HTTP** (`precheck-guard-test.json`). Bu guard sonradan eklendi: önceki00:30 probe kanıtları saklandı ve “yeni-IP doğrulandı” diye yeniden etiketlenmedi. Gelecekte cap1 anonim health precheck+8 negatif=9 public HTTP; bu tur ek HTTP0.
- **Authenticated /v1/chat veya herhangi model/completions requesti YASAK.** Yeni/gerçek GPU/model bağımlı kontroller **NOT TESTED — yalnız KANın canlı penceresi**. Bu auditte `/v1/models` yalnız invalid token denialdı; hiç geçerli model tokenı veya authenticated completion gönderilmedi. Tüm pozitif model/credit tests loopback/synthetic fixturedır.
- HOSTun `/opt/us-license-hook-test/licenses.json` ledger/credits/outbox ve `journalctl -u us-license-hook-test` bilgisi aktarılmıştır, bağımsız ölçülmedi. Root SSH erişimi yok; erişim genişletilmedi. Production backup/restore ve actual journal-sensitive-log kabulü **BLOCKED**. HOSTtan sadece stats/hash/age/cron/izin ve **sentetik QAya filtrelenmiş, credential/recipient içermeyen redacted** kanıt istenir; raw müşteri ledgerı veya secret/env dumpı boxa getirilmemeli.
- Gerçek SMTP allowlisted adreslere mail tetiklenmedi; remote register/reset/verification/delivery çağrısı yok. Privacy testleri sadece local DryRunSender/fakeSenderdır. Allowlistteki gerçek adresler test fixtureı yapılmadı.
- **Shopify uçtan uca kabul SHOPun işi**: `reports/r20261108/KANUS_SHOPIFY_E2E_20261108.md` referansı. Bu worker tam Shopify E2Eyi tekrar etmedi veya o raporun sonucunu ölçmeden sahiplenmedi; yalnız §12 HMAC/authorization sentetik güvenlik yüzeyini doğruladı.

14-madde toplamı değişmedi: **8 PASS /4 FAIL /2 BLOCKED**; ayrıca yeni-host serving-IP precheck BLOCKED, production journal review BLOCKED, GPU-dependent live checks NOT TESTED. Bu ek operasyonel kontroller14 maddelik yerel case toplamına eklenmedi.
