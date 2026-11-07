# KANUS Cloud Run ve Veritabanı Kabulü — 08.11.2026

**Güncel sonuç: us-commerce teknik deploy kabulü PASS (§16); genel ürün kabulü hâlâ tamamlanmadı.** Commerce readiness/rate limit/JSON logging ana ajan tarafından deploy edildi ve doğrulandı. Watchdog yerel yamaları deploy edilmedi; mevcut00007 temiz. Cloud SQL salt-okuma PASS, migration sürüm tablosu yok (FAIL), Shopify kapalı. Eski pencereler tarihsel kanıttır.

İlk rapor: 08.11.2026 00:29:48 TSİ. Son güncelleme: 08.11.2026 01:00:42 TSİ. Proje `project-27082a40-6083-4952-9db`, bölge `europe-west4`.
Ham kanıt dizini: `raw_cloudrun/` (bu rapora göre bağıl). API ham zaman damgaları kaynak biçiminde tutuldu; bu rapordaki zamanlar TSİ (Europe/Istanbul).

## 1. Başlangıç durumu

- Prompt §8, §14, §15; önceki `reports/SRE_CLOUDRUN_20261106.md` tamamı; `sre/cloudrun/image_app/app` ve yerel `us-watchdog-ready.patch` okundu.
- 00:22 TSİ ilk describe: `us-watchdog-00005-z97`, trafik %100, policy v5, fakat eski digest `3e001dbe…4098a`. Eski discard-reddetme açığı bu aşamada devam ediyordu.
- KAN'ın yeni deploy bildiriminin ardından yeniden bağımsız denetim: **`us-watchdog-00006-88h`**, trafik **%100**, Ready=True. Ready tek başına kabul kanıtı sayılmadı.
- us-commerce: `us-commerce-00002-pr2`, public allUsers, SA us-gateway. Shopify entegrasyonu `US_SHOPIFY_ENABLED=false`.
- SQL: `us-commerce-pg`, PostgreSQL 17, RUNNABLE; public IP; authorizedNetworks tanımlı değil; SSL `ENCRYPTED_ONLY`. İki Cloud Run servisi de Cloud SQL bağlantısına sahip.
- Sayfa/servis, IAM, secret, Scheduler, firewall ve veritabanı ayarlarında bu denetim tarafından değişiklik yapılmadı.

## 2. Değiştirilen dosyalar / uygulanan çözüm

İlk audit aşamasında yalnız bu rapor ve raw_cloudrun kanıtları üretildi. Devam göreviyle ayrıca `sre/cloudrun/fixes/` altında yerel kaynak kopyaları, servis patch'leri, testler ve DEPLOY_PLAN.md hazırlandı. KAN release dizini/canlı kaynak değiştirilmedi; deploy yapılmadı.

KAN'ın yaptığı 00006 deploy'u **bu denetimin değişikliği değildir**. Yeni kod ve ortam değişkeni image/digest ve yerel testlerle doğrulandı. Önceden hazırlanmış `/ready` yaması hâlâ yereldir; image runtime'ında bulunmuyor. Mevcut `us-watchdog-ready.patch` iki servise birden readiness ekleyen tamamlanmış bir çözüm değildir; yalnız watchdog taslağıdır.

Güvenlik: GPU/model/chat/completion/wake/start/inference, `/v1/*`, `/a100/*` uçlarına **hiç canlı istek yapılmadı**. ds-a100, VPS, usta-wake erişimi/SSH yok. `/internal/reconcile` çağrılmadı; Scheduler elle çalıştırılmadı. SA impersonation yok. İlk auditte DSN okunmadı; devam görevinde DSN yalnız bellekte kullanılarak proxy üzerinden salt-okuma SQL çalıştırıldı. Kullanıcı/registry token'ları da bellekte kaldı. SQL DDL/INSERT/UPDATE gönderilmedi. Rate-limit burst yapılmadı.

## 3. Kabul özeti

| Kontrol | Sonuç | Gerçek kanıt / sınır |
|---|---|---|
| us-commerce `/health` | PASS | HTTP 200, sabit JSON; kodda GPU/store çağrısı yok |
| `/healthz` kullanılabilirliği | FAIL (platform rezervi) | commerce ve watchdog 404 Google HTML; trace yok, uygulama request logu yok |
| commerce `/ready`, `/readyz` | PASS (/ready); /readyz405 | Yeni canlı /ready200, config/store true; §16 |
| watchdog `/ready`, `/readyz` | FAIL / NOT IMPLEMENTED | Yetkili kullanıcı ile 00005 ve 00006 üzerinde gerçek GET 404 JSON |
| GPU açmayan route/method yönlendirmeleri | PASS (davranış) | Her güvenli route'a GET; POST-only account uçları 405, devre dışı webhook 503 |
| Commerce eksik session reddi | PASS (erişim kontrolü) | SELECT-only `/account/balance`, `/account/requests` POST {} → 400, veri dönmüyor |
| Commerce auth HTTP semantiği | FAIL | Eksik/expired session 401 yerine genel 400 `request rejected`; ayrı kimlik hata sınıfı yok |
| Watchdog kimliksiz negatif test | PASS | `/`, `/health`, `/ready`, `/readyz` 403; `/healthz` IAM'den önce GFE 404 |
| OIDC negatif test | PASS | Rolsüz SA /health ve / →403; kimliksiz403; raw_cloudrun/oidc_negtest_sa_20261108.txt |
| Scheduler OIDC | PASS | Job SA/audience doğru; doğal Scheduler çağrıları 2xx |
| Cloud SQL uygulama bağlantısı | PASS (dolaylı canlı) | Balance/history salt-okuma auth yolu SELECT'e ulaşıp kontrollü 400 döndü; doğal reconcile 200 |
| Proxy üzerinden SELECT 1 / tablo listesi / n_live_tup | PASS | Mevcut kullanıcı IAM ile proxy v2.27.0; readonly=on; sorgular başarılı, proxy kapatıldı |
| Gerçek migration sürümü | FAIL / NOT IMPLEMENTED | Gerçek tablo listesinde migration/version tablosu yok; sürüm uydurulmadı |
| Secret envanteri/sürümler/IAM/ref | PASS (config) | 6 secret, gerekli SA secretAccessor bağlamaları, ref latest; değerler saklanmadı |
| 00006 trafik/image/SA/env | PASS | %100; immutable digest registry manifest ile aynı; env US_ALLOW_DISCARD_LOCAL_SSD=1 |
| Policy latest v5 / idle 600 | PASS (config) | v5 ENABLED; maskeli latest erişiminde idle=600, preserve=false |
| DiscardLocalSsd yeni kod + mock test | PASS | Image'tan çıkarıldı; policy false + env=1 → true, env yok → LifecycleError |
| BOOT5 gerçek otomatik idle stop | PASS (salt-okuma audit) | Tek stop operation, discard=true; TERMINATED. Executor tetiklemedi; SRE_BOOT5_VERIFY.md |
| Watchdog-managed wake/stop / kota akışı | NOT TESTED | BOOT5 manuel açılış; watchdog sahipli akış kullanılmadı |
| VM idle eşik uyumu | PASS (güncel yerel/KAN kaydı) | current/fixed600, aynı SHA; BOOT5 süre659,829sn. Canlı dosya yeniden okunmadı |
| Watchdog 00:10 sonrası hata taraması | PASS (gözlenen pencere) | LifecycleError/refusing/5xx yok; exception'lar maskelenebildiği için 503 de tarandı |
| 00:26 sonrası yeni revizyon reconcile | PASS (kısa pencere) | 00006 request logunda doğal POST reconcile 200 |
| Merkezi platform logları | PASS | severity, httpRequest, insertId, trace, spanId var |
| Uygulama yapılandırılmış hata/request-id logları | Commerce PASS /watchdog bekliyor | Yeni commerce JSON severity/trace/request_id doğrulandı; watchdog yaması yerel |
| Error Reporting kayıt taraması | PASS (gözlem) | ReportedErrorEvent yok; bir ERROR request, bizim devre dışı webhook GET 503 testi |
| Error Reporting uçtan uca teslimatı | NOT TESTED | Yapay exception üretilmedi; kayıt yokluğu entegrasyonun kurulu olduğunu kanıtlamaz |
| Sağlık/okuma endpoint rate limit | Commerce PASS /watchdog yerel | Commerce aday /health15.istekte429, Retry-After1; watchdog patch deploy edilmedi |
| Yerel KAN suite + yerel yeni yamalar | PASS / SKIPPED | 00007 tabanındaki 114 KAN testi + yeni testler; son gerçek sonuç §14'te, 37 opt-in skip aynı |

## 4. Gerçek komutlar ve HTTP matrisi

Toplama komutları ve tam parametreler `*.meta.json`; çalıştırılabilir testler `collect.py`, `collect_extra.py`, `http_acceptance.py`, `read_live_image.py` altında. HTTP sonuçları yanıt başlıkları ve gövdeleriyle saklandı; Authorization değeri yok.

```sh
python3 raw_cloudrun/collect.py
python3 raw_cloudrun/collect_extra.py
python3 raw_cloudrun/http_acceptance.py
# OAuth/OIDC yalnız kendi kullanıcı kimliği; çıktı dosyaya yazılmaz:
gcloud auth print-identity-token
# Salt-okuma config/IAM/job/log komutlarının exact argv'leri *.meta.json dosyalarında.
gcloud run services describe us-watchdog --project=project-27082a40-6083-4952-9db --region=europe-west4 --format=json
gcloud run revisions describe us-watchdog-00006-88h --project=project-27082a40-6083-4952-9db --region=europe-west4 --format=json
```

| Servis | İstek | Kimlik | HTTP | Kanıt |
|---|---|---|---|---|
| commerce | GET `/health` | none | 200 | `http_results.json` |
| commerce | GET `/healthz` | none | 404 | `http_results.json` |
| commerce | GET `/ready` | none | 405 | `http_results.json` |
| commerce | GET `/readyz` | none | 405 | `http_results.json` |
| commerce | GET `/` | none | 405 | `http_results.json` |
| commerce | GET `/account/register` | none | 405 | `http_results.json` |
| commerce | GET `/account/login` | none | 405 | `http_results.json` |
| commerce | GET `/account/redeem` | none | 405 | `http_results.json` |
| commerce | GET `/account/balance` | none | 405 | `http_results.json` |
| commerce | GET `/account/requests` | none | 405 | `http_results.json` |
| commerce | GET `/account/workers` | none | 405 | `http_results.json` |
| commerce | GET `/webhooks/shopify` | none | 503 | `http_results.json` |
| watchdog | GET `/` | none | 403 | `http_results.json` |
| watchdog | GET `/health` | none | 403 | `http_results.json` |
| watchdog | GET `/healthz` | none | 404 | `http_results.json` |
| watchdog | GET `/ready` | none | 403 | `http_results.json` |
| watchdog | GET `/readyz` | none | 403 | `http_results.json` |
| watchdog | GET `/` | own_user_identity_token | 404 | `http_results.json` |
| watchdog | GET `/health` | own_user_identity_token | 200 | `http_results.json` |
| watchdog | GET `/ready` | own_user_identity_token | 404 | `http_results.json` |
| watchdog | GET `/healthz` | own_user_identity_token | 404 | `http_results.json` |
| watchdog | GET `/readyz` | own_user_identity_token | 404 | `http_results.json` |
| watchdog 00006 | GET `/health` | none | 403 | `http_watchdog_00006.json` |
| watchdog 00006 | GET `/health` | own_user_identity_token | 200 | `http_watchdog_00006.json` |
| watchdog 00006 | GET `/ready` | own_user_identity_token | 404 | `http_watchdog_00006.json` |
| watchdog 00006 | GET `/readyz` | own_user_identity_token | 404 | `http_watchdog_00006.json` |
| commerce | POST `/account/balance` {} | none | 400 | `account_auth_readonly.json` |
| commerce | POST `/account/requests` {} | none | 400 | `account_auth_readonly.json` |

İlk watchdog grubu 00005 üzerinde, `http_watchdog_00006.json` grubu 00006 üzerinde çalıştı. Yeni image'ın runtime ve commerce_http dosyaları test edilen kaynakla aynı SHA256'ya sahip; readiness eksikliği devam ediyor.

### Route tablosu ve GPU güvenlik sınırı

- `GET /health`, `GET /healthz`: iki uygulamada da sabit JSON; controller/provider, lease/heartbeat/store çağırmaz.
- `GET /v1/models`: authenticated model metadata yolu. **NOT TESTED (açık /v1/* yasağı ve model/GPU teması riski)**. Kaynakta `ManagedModelClient.models` start yapmayan metadata yolu olsa da canlı çağrılmadı.
- `POST /v1/chat/completions`: reservation → managed inference → ensure_ready → provider.start olabilir. **NOT TESTED (GPU riski)**; GET dahi yapılmadı.
- `POST /account/register`, `/account/login`: auth_attempts ve hesap/session yazabilir. Yalnız GET 405 test edildi; **POST NOT TESTED (yazma yasağı)**.
- `POST /account/redeem`: lisans/kredi yazar; yalnız GET 405 test edildi; **POST NOT TESTED**.
- `POST /account/balance`, `/account/requests`: yalnız SELECT; GET method testi + kimliksiz POST {} reddi çalıştı. Geçerli müşteri/session kullanılmadı.
- `POST /account/workers`: session kontrolünden sonra live provider.describe/model_ready yapabilir; **POST NOT TESTED (VM'e erişim yasağı)**, GET 405 yalnız method gate'e ulaşır.
- `/webhooks/shopify`: flag=false ise method kontrolünden önce 503. GET 503 doğrulandı; POST/webhook yazması yapılmadı. Bu 503, Cloud SQL ya da route kaybı değil, bilinçli kapalı Shopify entegrasyonudur.
- Watchdog `POST /internal/reconcile`: controller.reconcile ve olası stop/mail delivery tetikler. **Elle hiçbir yöntemle çağrılmadı**; yalnız doğal Scheduler logları okundu. Watchdog'da `/`, `/ready`, `/readyz` route'u yoktur.
- Gateway bilinmeyen GET'leri 404 yerine 405 döndürür; method gate route eşleştirmesinden önce olduğundan `/ready` eksikliği maskelenir. Watchdog bilinmeyen GET'i 404 JSON döndürür.

### 404 kök nedeni — iki farklı durum

1. **`/healthz`: GFE/Cloud Run platform rezervi.** Her iki serviste 404 HTML; normal `/health` yanıtındaki trace başlığı yok; gözlenen log penceresinde `/healthz` request kaydı 0. Kodda healthz olsa da istek handler'a ulaşmıyor. Google resmî dokümanı *bazı* z-sonlu yolların ayrıldığını belirtir; **bütün z-sonlu yolların engellendiği doğru değildir**. Burada `/readyz` commerce uygulamasına ulaşıp 405, yetkili watchdog çağrısında 404 JSON döndü.
2. **`/ready`, `/readyz`: route uygulanmamış.** Image runtime.py sağlık ve reconcile dışında endpoint tanımlamıyor. Commerce de readiness tanımlamıyor. Watchdog `/ready` 404 JSON + trace; GFE healthz HTML 404'ünden farklıdır. Düzeltme: `/health` korunmalı, GPU açmayan gerçek `/ready` bağımlılık kontrolü eklenmeli; yalnız status sabit 200 olmamalı.

Resmî kaynak (bu denetimde açıldı): https://cloud.google.com/run/docs/known-issues (Reserved URL paths).

## 5. Kimlik doğrulama / OIDC / veritabanı

- us-commerce servis IAM: allUsers roles/run.invoker. Uygulama session denetimi ayrıdır. Kimliksiz SELECT-only uçlar HTTP 400 ile veri vermedi. Geçerli hesap login/üretim müşteri akışı bu salt-okuma testine dahil değil.
- us-watchdog servis IAM: yalnız us-scheduler@ için run.invoker. Ancak bu **proje Owner'ını dışlamaz**. Aktif kullanıcı projede roles/owner; kendi token'ıyla `/health` 200 alması beklenen IAM inheritance ile açıklanıyor. `/` ve `/ready` uygulama 404'ü döndü. Bu nedenle istenen “kendi kullanıcı token'ım da 403 almalı” negatif testinin önkoşulu sağlanmıyor; bunu PASS diye yazmadım. IAM'i daraltmadım, SA impersonate etmedim. Ayrı yetkisiz kullanıcı testi için onaylı gerçek nonprivileged kimlik gerekir.
- Scheduler: job us-idle-watchdog ENABLED, `* * * * *` UTC, deadline 300 sn, retryCount 1. POST servis `/internal/reconcile`, audience tam servis kökü, OIDC SA us-scheduler@. Audience/SA tanımı ve bu job'un gerçek 200 bitiş logları birlikte OIDC kabul kanıtıdır; biz reconcile çalıştırmadık.
- Cloud SQL uygulama tarafı: `/account/balance`, `/account/requests` store SELECT'inden sonra session bulunamadı → kontrollü ValueError → 400; connection RuntimeError olsaydı bu handler onu yakalayıp 400'e çevirmiyor, 5xx olurdu. Bu, iki auth read yolunda canlı DB kullanımına dolaylı kanıttır; **SELECT 1 ya da migration tablosu sonucu değildir**.
- Runtime configured_store PostgreSQL zorunlu kılıyor. Doğal reconcile store.view/observe/watchdog_beat kullanıyor ve 200. Uygulama loglarında connection failure/traceback yok; Cloud SQL son-24saat WARNING+ log sayısı **0**.
- Doğrudan mevcut SQL public IP:5432 bağlantısı TimeoutError (5 sn). Authorized network yok, yeni IP eklenmedi; Cloud SQL Auth Proxy/driver bu kutuda kurulu değildi. DSN'yi almak ağ engelini çözmeyeceğinden **secret DSN okunmadı**. Bu ilk doğrudan-IP denemesinde sorgular BLOCKED idi; §13'teki açık kullanıcı talebiyle proxy bağlantısı bu engeli kaldırdı. Migration tablosu bulunmadığından migration kabulü FAIL oldu. Yeni ağ yetkisi açılmadı veya `gcloud sql connect` ile otomatik authorizedNetwork değişikliği denenmedi.
- Image'da AccountStore ve LifecycleStore başlangıçta CREATE TABLE IF NOT EXISTS kullanıyor; versioned migration runner/schema_migrations görülmedi. Canlı DB sürümü hakkında çıkarım yapılmadı; şema uyumu ve migration kabulü açık kalıyor.
- Resmî bağlantı kaynağı (bu denetimde açıldı): https://docs.cloud.google.com/sql/docs/postgres/connection-options . Gerekirse izinli resmî Cloud SQL connector üzerinden yalnız okuma oturumu planlanmalı; ağ erişimi genişletmek önerilmiyor. Bu denetimde kurulmadı.

## 6. Secret Manager

Değerleri raporlanmayan 6 secret; sürümler, secret-level IAM ve proje IAM okundu. Proje Owner'ı idari erişim taşır; aşağıdaki SA'lar runtime erişimleridir. Public/allUsers secret binding yok.

| Secret | Sürümler | secretAccessor SA | Servis ref |
|---|---|---|---|
| us-credit-products | 1:ENABLED | us-gateway | latest |
| us-license-signing | 1:ENABLED | us-gateway | latest |
| us-model-ca | 1:ENABLED | us-gateway, us-watchdog | latest |
| us-model-token | 1:ENABLED | us-gateway, us-watchdog | latest |
| us-postgres-dsn | 1:ENABLED | us-gateway, us-watchdog | latest |
| us-worker-policy | 5:ENABLED, 4:ENABLED, 3:ENABLED, 2:DISABLED, 1:ENABLED | us-gateway, us-watchdog | latest |

- Commerce env: us-postgres-dsn, us-license-signing, us-model-token, us-credit-products → latest; volume policy ve ca → latest.
- Watchdog 00006 env: us-postgres-dsn ve us-model-token → latest; volume us-worker-policy ve us-model-ca → latest. SA us-watchdog@.
- Policy `versions describe latest` v5; v2 DISABLED, v1/v3/v4/v5 ENABLED. `policy_latest_redacted.json` yalnız izinli config alanlarını tutar, worker key ve upstream maskelidir. Üst JSON `version=1` **policy dosya şema sürümüdür**, Secret Manager v5 ile karıştırılmamalıdır.
- `latest` floating referansı, process'in yüklediği anın değeriyle her zaman aynı demek değildir; 00006 v5 etiketiyle güncel deploy edilmiş olsa da in-memory policy'yi güvenli HTTP introspection endpoint'iyle okuyamadık. Runtime'ın policyyi yüklerken aldığı config'in direct kanıtı yok; latest içeriği bağımsız okundu.

```json
{
  "version": 1,
  "workers": [
    {
      "key": "[REDACTED]",
      "model": "deepseek-v4.1-flash",
      "project": "project-27082a40-6083-4952-9db",
      "zone": "europe-west4-b",
      "instance": "ds-a100",
      "instance_id": "6979174631840438511",
      "upstream": "[REDACTED]",
      "idle_seconds": 600,
      "boot_seconds": 600,
      "request_seconds": 600,
      "lease_seconds": 60,
      "capacity": 10,
      "max_uptime_seconds": 1800,
      "daily_runtime_seconds": 3600,
      "daily_starts": 6,
      "preserve_local_ssd": false,
      "adopt_running": false,
      "require_watchdog": true
    }
  ]
}
```

## 7. Watchdog 00006 bağımsız image ve test doğrulaması

- Son service describe: %100 trafik us-watchdog-00006-88h, latestReady aynı; SA `us-watchdog@project-27082a40-6083-4952-9db.iam.gserviceaccount.com`.
- Tag: `gateway:2026-11-08-discard`. Revision imageDigest:
  `sha256:5bf7bc8e276d907f65ae0f274b6de2a9853b5b32b73ed0f59bf9cb0f66392122`.
- Registry OCI Distribution API'sinden manifest + küçük üst uygulama katmanı **salt okuma** indirildi (`read_live_image.py`). Docker/crane kurulu değildi; image çalıştırılmadı. Manifest hash'i deployed revision imageDigest ile **birebir aynı**. Blob hash'i de doğrulandı. `live_gce_provider.py`, `live_runtime.py`, `live_commerce_http.py` doğrudan bu image'tan çıkarıldı; KAN'ın test edilen kaynak kopyalarıyla byte-for-byte aynı (`source_image_comparison.json`). Bu nedenle **image içeriği bağımsız doğrulandı**; yalnız KAN'ın patch beyanına dayanılmıyor.
- Env `US_ALLOW_DISCARD_LOCAL_SSD=1`; preserve_local_ssd=false policy ile `params['discardLocalSsd']='true'`, API beta stop. Env opt-in yoksa `LifecycleError('refusing to discard local SSD')`, POST yapılmıyor. Preserve=true ise false (koruma), local SSD yoksa normal v1 stop. Opt-in ayrıca true/yes değerlerini de kabul eder; mevcut deploy 1 kullanıyor.
- İncelenen patch: `reports/r20261108/discard_fix.patch`; test: `test_gce_discard.py`. **6/6 discard testi full suite içinde geçti.** Mock transport gerçek GCE'ye çıkmaz; testi çalıştırmak gerçek GPU stop değildir.

Gerçek test komutu:
```sh
cd /workspace/us-panel/kan/watchdog-fix/t
env -u US_POSTGRES_TEST_DSN python3 -m pytest -q -rs /workspace/us-panel/kan/watchdog-fix/t
```
Sonuç: **104 passed, 37 skipped in 7.45s** (`discard_suite_tests.txt`). 37 skip = **36 opt-in gerçek PostgreSQL sözleşme testi + 1 PostgreSQL/Gunicorn runtime testi**. PostgreSQL testleri CREATE SCHEMA/INSERT/DROP da yapacağından bu salt-okuma kabulünde DSN özellikle unset tutuldu. Bu skip'ler başarı sayılmaz ve canlı Cloud SQL migration kabulünü karşılamaz.

### 10 dakika kuralı / VM içi idle.py

Policy latest v5 `idle_seconds=600` = **10 dakika**, prompt kuralı config düzeyinde PASS. Buna karşılık `sre/watchdog/idle.py.current:14` **IDLE = 15 * 60 = 900 saniye**. `idle600.diff` 10*60'a yerel düzeltme öneriyor; bunun VM'e uygulandığına dair canlı okuma yapılmadı. Dolayısıyla **yerel snapshot ile watchdog uyumsuz**. VM'e/SSH'ye erişim yasağı nedeniyle güncel VM dosyasının kesin durumu NOT TESTED. KAN'ın izinli denetiminde runtime eşiğini doğrulaması gerekiyor; bu rapor snapshot'ı canlı durum diye sunmaz.

## 8. Loglar / Scheduler / hata takibi

Tam 24 saat penceresi: **07.11.2026 00:27:56 TSİ – 08.11.2026 00:27:56 TSİ**. API freshness ile asc sıralamanın eski kayıtları içerebildiği görüldüğünden nihai toplamada **açık timestamp alt/üst sınırları** kullanıldı. Dosyalar limitlerine ulaşmadı; sonuçlar truncation göstermiyor. Log ingestion gecikmesi son birkaç saniyeyi dışarıda bırakabilir.

- Scheduler AttemptStarted **1439**, AttemptFinished **1439**, 2xx **1439** → gözlenen biten denemelerde **100.00%**. Tam pencere sonundaki henüz yazılmamış son execution bu paydaya eklenmedi. Job elle çalıştırılmadı.
- 00:10 TSİ sonrası tam toplanmış watchdog penceresi: **68 kayıt**, son kayıt **08.11.2026 00:27:02 TSİ**. LifecycleError/refusing-to-discard: **0**; HTTP 5xx: **0**.
- 00:26 TSİ sonrası: **14 watchdog kaydı**; LifecycleError/refusing: **0**; 5xx **0**. 00006 özelinde doğal reconcile **1 adet**, kod dağılımı **{'200': 1}**. Scheduler 00:26 sonrası bitiş **2**, durumlar **{'200': 2}**. Pencere kısadır; canlı stop kanıtı değildir.
- Önemli nüans: lifecycle/runtime exception'ları generic hata mesajlarına çevirebilir; logda literal LifecycleError olmaması tek başına sağlamlık kanıtı olmaz. HTTP 503/5xx ayrıca arandı ve watchdog'da bulunmadı.
- Platform logu severity/httpRequest/trace/spanId/insertId içeriyor; acceptance UA ile istekler eşleştirildi (`acceptance_request_logs.json`). `/healthz` uygulama request log kaydı **0**. Uygulama logu Gunicorn düz metin, platformda DEFAULT; tam JSON structured logging ve her app yanıtında request-id bulunmuyor. Completion response request-id kodu var ama GPU yasağı nedeniyle canlı test edilmedi.
- ERROR+ / ReportedErrorEvent taraması: `error_reporting_logs.json` **1 kayıt**: kendi `GET /webhooks/shopify` testimizin 503 ERROR request logu. `ReportedErrorEvent` tipi yok; Error Reporting grup/notification teslimatı doğrulanmış değildir. Sentetik üretim exception gönderilmedi.

## 9. İstek sınırlandırma

Commerce kaynakta yalnız login/register için `check_auth_rate` çağırıyor; IP ve email scope başına **10 / 300 sn**. Sayaç SQL'e yazdığı için o test yapılmadı. Aşım ValueError → HTTP 400; 429/Retry-After değil.

`/health`, balance/history salt-okuma yollarında limiter middleware yok; Cloud Run maxScale/concurrency ayarı istek/saniye rate limit değildir. **FAIL / NOT IMPLEMENTED**; kullanıcının koşulu gereği burst testi başlamadan koddan eksiklik belirlendi. 60 istek/2–3 sn yük üretilmedi. Normal acceptance istekleri düşük sayıda, ardışık çalıştı.

## 10. Kalan eksikler ve gereken onaylar

1. **P0 readiness:** commerce ana ajan tarafından deploy edildi, gerçek /ready200 (§16). Watchdog readiness yaması yerel; bu executor deploy yapmadı.
2. **P0 idle tutarlılığı:** policy v5 600 ile VM içi gerçek idle eşiğini KAN doğrulasın; snapshot 900. Canlı VM değişikliği/kapanma deneyi bu görevin kapsamı dışında ve açık onay gerektirir.
3. **P1 migration:** proxy ile salt-okuma SQL kanıtı tamamlandı (§13), fakat migration tablosu yok. Versioned migration tasarımı hazırlanmalı; üretim DDL/migration için ayrıca onay.
4. **P1 güvenli rate limit ve auth semantiği:** commerce429/Retry-After canlı kabulü tamam; shared instance bucket/webhook etkisi işletim riski. Watchdog yaması yerel, auth400 semantiği değişmedi.
5. **P1 gözlemlenebilirlik:** yerel structured JSON/request-id/trace ve redaction testleri tamamlandı; canlı doğrulama ve Error Reporting teslimat kabulü açık. Üretimde hata enjekte edilmedi.
6. Owner token negatif testi yetkisiz kimlik değildir. Ayrı onaylı rolsüz SA negatif testi dış kanıtta tamamlandı (§12); bu executor IAM değiştirmedi/SA impersonate etmedi.
7. Discard düzeltmesi **kod/config/mock düzeyinde PASS**; gerçek kota tetikleme ve sağlayıcı kapanışı **NOT TESTED**. Gerçek GPU açma/kapama onayı bu görev kapsamında yoktur. Geri dönüş 00005-z97 eski stop-reddetme açığını geri getirir; burada rollback yapılmadı.

**Genel hüküm:** gerçek HTTP, image ve doğal Scheduler kanıtı toplandı. Ready=True ile tamam kabul edilmedi. Canlı servis/kaynak ayarı değiştirilmeden eksikler açıkça FAIL/BLOCKED/NOT TESTED bırakıldı.

## 11. Son uzatılmış watchdog log penceresi

Son toplama üst sınırı **08.11.2026 00:30:10 TSİ**, başlangıç 08.11.2026 **00:10:00 TSİ**. `watchdog_post_deploy_final_window.json` ve `.meta.json` tam sorgu/kanıtı taşır; yukarıdaki önceki snapshot istatistiklerini zaman olarak genişletir. Toplam **75 kayıt**, son log **00:30:02 TSİ**. LifecycleError / refusing-to-discard **0**; watchdog HTTP 5xx **0**. 00006 özelinde doğal Scheduler reconcile **4**, HTTP dağılımı **{'200': 4}**. İstekleri biz tetiklemedik. GPU stop testi hâlâ NOT TESTED.

## 12. Güncel OIDC negatif test ve idari temizlik

- **Kimliksiz GET /health →403 PASS** (`http_watchdog_00006.json`). Owner token'ın 200 alması idari IAM mirasıdır; negatif test başarısızlığı sayılmaz.
- **Rolsüz test SA negatif testi: PASS.** Parent'ın doğrudan kullanıcı onayıyla yürüttüğü test kanıtı `raw_cloudrun/oidc_negtest_sa_20261108.txt` okundu: SA token'ıyla `/health` ve `/` 403, kimliksiz `/health` de403. Kayıt zamanı 08.11.2026 00:32:45 TSİ. Bu executor SA oluşturmadı ve impersonation yapmadı.
- **SA `sre-oidc-negtest` rolsüz duruyor ve silinmedi; silme kararı Demir'in.** Geçici TokenCreator bağlaması kaldırıldı; servis hesabı politikası boş. `oidc_negtest_sa_final_iam.json` bağımsız salt-okuma get-iam-policy sonucu bindings içermiyor. Kaynak test kanıtı proje rollerini `[]` olarak bildiriyor.

## 13. Gerçek Cloud SQL salt-okuma kabulü — proxy ile engel çözüldü

Açık devam talebiyle Cloud SQL Admin API üzerinden **resmî Cloud SQL Auth Proxy v2.27.0** kullanıldı; public IP PRIMARY var. Yalnız private IP olsaydı kutunun VPC rotası olmadığı için durulacaktı. Binary resmî Google Storage URL'inden indirildi; SHA256 `39ed2abcaa073c3b4714585e36a904e2b1d92218c5e07c18224fb73ef4bedde5` eşleşti.

- Proxy `--gcloud-auth`, mevcut kullanıcı IAM; local listener `127.0.0.1:15438`; **yeni SA/rol/authorizedNetwork/IP yok**. `--enable-builtin-metrics=false`; Cloud Monitoring/Trace telemetry açılmadı. Cloud SQL Auth Proxy'nin geçici bağlantı sertifikası dışında kaynak/ayar yazması yapılmadı.
- us-postgres-dsn latest yalnız subprocess çıktısından **bellekte** parse edildi; DB kullanıcı/şifre DSN dosyaya/rapora/log'a yazılmadı. psql parolası yalnız geçici child process ortamında kullanıldı; komut satırına eklenmedi. Proxy TLS tüneli nedeniyle yerel loopback psql SSL kapalıdır; uzak bağlantı TLS ile proxy tarafından korunur.
- PGOPTIONS ile başlangıçtan read-only, her sorgu oturumunda ayrıca `SET default_transaction_read_only=on`; `SHOW default_transaction_read_only` → **on**.
- **SELECT 1 →1 PASS**; **SELECT version() → PostgreSQL17.11 PASS**.
- Kullanıcı şema tablo listesi **12 tablo**, şema **us_app**. Aşağıda pg_stat_user_tables.n_live_tup değerleri var; **yaklaşık istatistiktir, kesin COUNT(*) sonucu değildir**. Müşteri satır içerikleri okunmadı.
- Migration tablosu aranabilecek gerçek tablo listesi okundu: schema_migrations/alembic_version/django_migrations/migrations/schema_version/flyway_schema_history yok. Diğer tablolar arasında da migration sürüm tablosu görülmedi. **Migration sürümü FAIL / NOT IMPLEMENTED**, değer uydurulmadı.
- Proxy **kapatıldı** (`proxy_stopped=true`). İlk startup denemesinde script yanlış readiness log cümlesini bekledi ve timeout sonrası proxy'yi kapattı; `Listening on` algısı düzeltildi, aynı resmî proxy yolu tekrar çalıştırılarak kabul tamamlandı. Auth/ağ yetkisi değiştirilmedi.

Kanıtlar `raw_cloudrun/db_proxy/`: `instance.json`, `read_only_probe.py`, `db_readonly_result.json`, `proxy.log`. Sorgu SQL'leri, psql returncode=0 ve maskeli çıktılar JSON'da. Oturum zamanı **08.11.2026 00:33:14 TSİ**; DB metadata dışında veri okunmadı. Hiçbir INSERT/UPDATE/DDL çalıştırılmadı.

| Tablo (us_app) | n_live_tup (yaklaşık) |
|---|---:|
| accounts | 1 |
| auth_attempts | 2 |
| credit_packs | 0 |
| delivery_jobs | 0 |
| inference_events | 0 |
| inference_watchdog | 1 |
| inference_workers | 1 |
| issued_licenses | 1 |
| metered_requests | 0 |
| sessions | 1 |
| usage | 0 |
| webhook_orders | 0 |

## 14. Yerel yamalar ve 00007 rebase — deploy yok

Talep edilen işlem bir repository checkout/PR değil, image'tan çıkarılmış yerel kaynak kopyalarına geri alınabilir yama/testtir. KAN'ın `/kan/watchdog-fix/release-20261108/` dizini değiştirilmedi; kopya **`sre/cloudrun/fixes/base-00007/`**, birleşik yerel sonuç **`app-00007/`**.

### Ayrı ve temiz servis patch'leri

- **`/workspace/us-panel/sre/cloudrun/fixes/us-watchdog/local-acceptance.patch`**
- **`/workspace/us-panel/sre/cloudrun/fixes/us-commerce/local-acceptance.patch`**

Her biri ayrı temiz base-00007 kopyasına uygulanır; **birbirinin üzerine uygulanmaz**. Watchdog patch commerce route/gateway factory'sini değiştirmez; commerce patch watchdog factory'sini değiştirmez. Yardımcı modüller self-contained olarak her patch'te bulunur. Birleşik alternatif `combined.patch`; sıralı alternatif readiness.patch → rate-limit.patch → structured-logging.patch. `patch_apply_check.json` üç yolun temiz uygulanmasını ve beklenen dosyalarla birebir eşleşmesini doğrular.

1. **/ready:** commerce'in eski405 kökü genel HTTP method gate'in GET'i route eşleşmesinden önce reddetmesiydi; GET /ready bu gate'den önce eklendi. Watchdog'da eksik route eklendi. Config + yalnız SELECT1; başarı200, bağımlılık hatası503; hata/DSN metni döndürülmez. Sağlık/model/provider/heartbeat/lease/reconcile/mail yan etkisi yok.
2. **Rate limit:** instance-local, thread-safe, sabit üç token bucket; health2/sn burst10, read5/sn burst30, genelAPI10/sn burst60. 429 + Retry-After. Bounded memory; concurrency/refill/clock rollback testleri var. Yalnız IAM-private watchdog POST /internal/reconcile muaf; sahte Scheduler UA/Authorization genel API'yi muaf yapmaz. Cloud Run IAM hâlâ OIDC sınırıdır.
3. **JSON log:** severity/message/httpRequest, doğrulanmış project+trace context'ten logging.googleapis.com/trace ve sunucuda üretilen request_id. Headers/body/query/Authorization/DSN/token/exception metni loglanmaz; bilinmeyen path maskelenir. 429 WARNING, hata ERROR; middleware/stream kapanışı ve logger arızası test edildi.
4. **Lease korunumu:** KAN ready_window.patch incelendi. READY anında bir kez lease deadline rebase edilir; managed response ready_monotonic taşır, commerce600sn penceresi READY'den sayılır. /ready eklemesi health branch civarında, streaming penceresi değişikliği farklı bölgede; çatışma çözülürken KAN'ın streaming AST'si birebir korundu. inference_lifecycle.py, managed_inference.py ve gce_provider.py byte-eşitlik testleri var. KAN'ın10 READY/lease davranış testi aynı yeni suite içinde koşuyor.
5. **Bu yerel test aşamasında canlı durum değişmedi:** sonradan commerce ana ajan tarafından deploy edildi ve /ready/429/JSON log kabulü PASS oldu (§16). Watchdog yaması hâlâ yerel. Auth400 ve migration açık; idle600 BOOT5 takibinde doğrulandı.

**Son birleşik sonuç: 142 passed / 37 skipped (8.36 sn)** = KAN114 + yeni28 test. Servis patch'leri ayrı temiz tabanda da koşuldu:

| Yerel ağaç | Sonuç | Kanıt |
|---|---|---|
| app-00007 birleşik | 142 passed, 37 skipped | fixes/combined_tests_00007.txt |
| us-watchdog bağımsız patch | 138 passed, 37 skipped, 4 deselected | fixes/us-watchdog/tests_00007.txt |
| us-commerce bağımsız patch | 138 passed, 37 skipped, 4 deselected | fixes/us-commerce/tests_00007.txt |

Her servis koşusunda diğer servise özgü4 yeni test seçilmedi; KAN'ın114 baseline testi korunur. 37 skip =36 gerçek PostgreSQL contract +1 PostgreSQL/Gunicorn; DDL içerdiklerinden readonly kabulde açılmadı. Yerel /ready ve 429 çıktıları fake-store/WSGI testleri, JSON log/redaction testleri, 100 fake Scheduler reconcile muafiyet testi ve lease AST/byte preservation testleri tamamlandı. Gerçek Cloud Run job/reconcile tetiklenmedi.

Test dizinleri `tests/baseline` ve `tests/new`; module_provenance_00007.json hangi app modüllerinin import edildiğini kaydeder. Opt-in DB DDL testleri için US_POSTGRES_TEST_DSN unset; 37 skip aynı nedenden kalır.

## 15. Yeni us-watchdog-00007-fnc salt-okuma denetimi ve yayın planı

- Trafik **%100 00007-fnc**, Ready=True; SA us-watchdog@. Image tag **gateway:2026-11-08-readywindow**, immutable digest **sha256:f8a85dbcdee57c0d21639feed0b722beac1ab4b9d94781a06a9d1fb85011e1a6**.
- KAN release tarball SHA256 **80d16639d3c3440518e4db519a98feb0a7f44f31cb6bf8aaea7ace03aaba5584** doğrulandı. Registry manifest hash'i deployed revision digest ile aynı. Image'tan bağımsız çıkarılan commerce_http.py, runtime.py, gce_provider.py, inference_lifecycle.py ve managed_inference.py **release kopyasıyla birebir eşleşiyor**.
- Env US_ALLOW_DISCARD_LOCAL_SSD=1; US_APP_FACTORY runtime:watchdog_app(); DB/model token secret ref latest, policy ve ca volume ref latest. Gerekli SA/secret bağlamaları korunmuş. Secret değerleri loglanmadı.
- 08.11.2026 **00:34–00:39:33 TSİ** penceresinde9 kayıt; 5xx0, LifecycleError/refusing0; 00007 doğal reconcile **6/6 HTTP200**. Kısa pencere, gerçek GPU stop kanıtı değil. Kanıt `raw_cloudrun/revision00007/`.
- **00:39 tarihsel durum:** STOPPING fix usta-wake'te; bu executor erişmedi. O anki raporda idle.py uygulanmamıştı. **BOOT5 takibinde** KAN idle600 kurulum kaydı, yerel eşit hash, gerçek idle stop ve00007 temizliği doğrulandı; istenen üç kapı PASS. Daha geniş canlı STOPPING-kuyruk akışı NOT TESTED kalır; `SRE_BOOT5_VERIFY.md`.
- **DEPLOY_PLAN.md** us-watchdog/us-commerce için ayrı satırlar, ayrı build/deploy komutları (çalıştırılmadı), /ready/429/JSON log doğrulama adımları ve update-traffic rollback komutları içerir. Watchdog rollback **00007-fnc**, commerce rollback **00002-pr2**.
- Tarihsel00:37 snapshot: us-commerce **00002-pr2**, image digest **sha256:3e001dbec0a7d033ac105366bf282a59e4babff84e60413a33dbc2a7c474098a**. **Bu00:37 snapshot'ından sonra ana ajan commerce'i yeni tabana taşıdı; lease düzeltmesi artık ulaştı (§16).** Bu executor build/deploy yapmadı.

## 16. us-commerce deploy — GERÇEK SONUÇ PASS (ana ajan uyguladı)

Son doğrulama: **08.11.2026 01:00:15 TSİ**. Önceki bu bölümdeki “başlatılmadı” durumu yalnız executor'ın ilk hazırlık durumu idi; ana ajan deploy'u tamamladıktan sonra ham kanıtlar ve bağımsız salt-okuma API sonuçlarıyla güncellendi. **Bu executor yeni bir build/deploy/traffic/IAM değişikliği yapmadı; VM/VPS/GPU uçlarını çağırmadı.**

| Kontrol | Sonuç | Gerçek kanıt |
|---|---|---|
| Önceki revizyon | PASS | us-commerce-00002-pr2, ilk describe %100 |
| Temiz context / yerel regresyon | PASS |138 passed /37 skipped /4 diğer-servis yeni testi deselected; context_tests.txt |
| Başarılı regional build | PASS | **dda56d30-c325-47d9-8689-c21cb9a63583**, europe-west4, SUCCESS |
| Builder/staging | PASS | us-builder SA; source bucket **us-build-247516932580-euw4**; build API ile bağımsız doğrulandı |
| Digest / yeni revizyon | PASS | **us-commerce-sre-ready-r1**, sha256:2416ec71eb4e16c3da4c1d9866cb3c7393ce8819efd043e8b1f28bf9a8bfdbab |
| Aday önce trafiksiz | PASS |candidate_deploy.log:0%; tag URL oluşturuldu |
| SA/env/secret/CloudSQL runtime spec | PASS |spec_old.json==spec_new.json; bağımsız eski/yeni **revision spec** image hariç birebir aynı |
| IAM | PASS |allUsers invoker bağlaması öncekiyle aynı; yeni yetki eklenmemiş |
| Aday /health | PASS |200, shopify_webhooks=false |
| Aday /ready | PASS |200, checks.config=true /checks.store=true |
| 429 /Retry-After | PASS |Yalnız health testi;15.istekte429, Retry-After=1, JSON error;60 sınırının altında duruldu |
| JSON log / sır taraması | PASS (gözlenen kayıtlar) |candidate_logs19 kayıt; yapılandırılmış stdout ve trace/httpRequest; JWT/Bearer/DSN örüntüsü yok |
| Trafik | PASS |**08.11.2026 00:48:30 TSİ** sonrası %100 yeni revizyon, latest alias değil **sabit revision pin** |
| Canlı /health ve /ready | PASS |live_checks.txt:iki endpoint200, ready check'leri true |
|5+dk hata gözlemi | PASS (düşük trafik) |Ana ajan293 sn pencere; bağımsız log uzatması **416,012 sn /6dk56sn**,8 HTTP200,5xx0,ERROR+0 |
| Webhook /ledger canlı yazma testi | NOT TESTED (bilinçli) |Gerçek POST yok; yalnız kaynak/route incelemesi yapıldı |
| Model/wake/GPU doğrulaması | NOT TESTED (yasak) |/v1/*,/a100/*,wake/start çağrılmadı |

### Build, digest ve spec ayrıntısı

- Build başlangıcı **08.11.2026 00:46:47.599635 TSİ**, bitişi **08.11.2026 00:47:09.057773 TSİ**; serviceAccount `projects/project-27082a40-6083-4952-9db/serviceAccounts/us-builder@project-27082a40-6083-4952-9db.iam.gserviceaccount.com`. Image tag **gateway:2026-11-08-sre-commerce-ready**, deployed immutable digest build sonucu ve revision describe ile aynı.
- İlk **global bölge denemesinin us-builder GCS izni yüzünden başarısız olduğu ana ajan tarafından bildirildi**. Başarılı retry'nin europe-west4 ve us-build staging bucket'ı kullandığı hem build.log hem Cloud Build API ile doğrulandı. Global son-build listesi boş; ilk başarısız denemenin ham hata metni teslim edilen kanıt setinde bulunmadığından o hata ayrıntısı bağımsız doğrulanmış sayılmadı. Bu, başarılı retry'yi etkilemiyor.
- Eski/new normalize spec eşit. Ayrıca `commerce_old_revision.json` ve `commerce_revision.json` **revision-to-revision runtime spec**, yalnız image alanı çıkarıldığında tam eşit; `commerce_revision_to_revision_spec.diff` boş. Service template ile materialized revision doğrudan kıyaslanırsa container name gateway-1 eklenmiş görünür; eski revision'da da aynı ad olduğundan bu gerçek ayar değişikliği değildir. SA, env, secret mounts, CloudSQL/VPC/scale anotasyonları korunmuş.
- Invoker bindings eşitliği bağımsız kontrol sonucu: **True**. Env/IAM/secret/Scheduler değişikliği bu denetim tarafından yapılmadı.

### Aday ve canlı kabul / logların anlamı

Aday tag URL: `https://sre-candidate-commerce---us-commerce-iqhblf63dq-ez.a.run.app`.
Canlı URL: `https://us-commerce-iqhblf63dq-ez.a.run.app`.

`/health` → `{"ok":true,"shopify_webhooks":false}`; `/ready` → `{"ready":true,"checks":{"config":true,"store":true}}`. `/readyz`405 olarak kalır; kabul edilen yol `/ready`'dir. Bu audit yeni canlı HTTP isteği atmadı; ana ajan ham HTTP sonuçları ve mevcut runtime/log/API kanıtı kullanıldı.

Cloud Logging özel JSON anahtarlarını severity/trace/httpRequest üst LogEntry alanlarına taşır; dolayısıyla jsonPayload'da yalnız message/request_id görülmesi bir schema hatası değildir. Candidate429 WARNING ve request-id/trace kaydı var. Sır taraması actual secret değerlerini okuyarak değil, kod allowlist'i ve JWT/Bearer/DSN örüntülerini tarayarak yapıldı; HTTP Authorization/body/query dump yok.

**Rate limit:** instance başına, kategori bazında bütün istemcilerin paylaştığı token bucket; client-IP bazlı veya instances-arası global kota değildir. Logdaki remoteIp **169.254.x** platform proxy adresi; son kullanıcı kimliği olarak yorumlanmamalı. Webhook genel API bucket'ındadır; yoğun/istenmeyen genel API trafiği webhook'a da429 döndürebilir. Bu paylaşımlı limit etkisi kalan işletim riski olarak kaydedildi; burada politika değiştirilmedi.

**Shopify flag:** shopify_webhooks=false, eski00002'de de false idi (env spec eşit); yeni deploy'un getirdiği bir regresyon değil. Gerçek Shopify/ledger uçtan uca kabulü PASS ilan edilmez. `account_store.py` ve `shopify_ledger.py` kaynakları eski image kopyasıyla byte-eşit; webhook handler flag kapalıyken503 yolunda. Gerçek webhook/ledger/model POST gönderilmedi.

### Trafik ve gözlem penceresi

- Mevcut describe **%100 us-commerce-sre-ready-r1** gösteriyor. `latestRevision=false` (alanın atlandığı gösterimlerde varsayılanfalse); trafik belirli revision'a sabitlenmiş durumda. Sonradan latestReady değişmesi trafiği otomatik taşımaz. Candidate tag korunuyor.
- `post_promote_monitor.txt` ana ajan penceresi **00:48:30–00:53:23 TSİ =4dk53sn**, yani tam5dk değil. Eksik7sn'yi gizlememek için salt-okuma Cloud Logging sorgusuyla pencere **00:55:26 TSİ**'ye uzatıldı: **6dk56,012sn**; HTTP200=8,5xx0,ERROR+0. Kaynak `boot5_verify/commerce_post_promote_logs.json` ve `.meta.json`. Bu düşük trafik hacminde sağlıklı gözlemdir; yük testi/SLA değildir. Geri dönüş gerekmedi ve yapılmadı.

### Geri dönüş / kaynaklar / kalan açıklar

Geri dönüş (çalıştırılmadı):
```sh
gcloud run services update-traffic us-commerce --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-commerce-00002-pr2=100
```

Önceki revizyon **us-commerce-00002-pr2**; yeni **us-commerce-sre-ready-r1**. Rollback readiness/rate/JSON log ve READY-window lease düzeltmesini de geri alır; secret latest içeriğini otomatik geri almaz.

Ham ana ajan kanıtları `raw_cloudrun/us_commerce_deploy/`: build.log,digest.txt,candidate_deploy.log,spec_old/new.json,candidate_checks.txt,candidate_logs.json,promote.log,live_checks.txt,post_promote_monitor.txt. Bağımsız doğrulama `raw_cloudrun/boot5_verify/commerce_*` dosyaları. Yerel önceki “başlatılmadı” status kaydı tarihsel olarak korundu, güncel durum gerçek sonuçla yenilendi.

**Sonuç: us-commerce teknik deploy kabulü PASS.** KAN'ın READY-window/lease değişikliği artık bu servisin yeni image'ına ulaşmıştır. Bu, Shopifyfalse/migration yokluğu/geçerli müşteri-model akışı gibi daha geniş ürün kabul açıklarını kapatmaz. us-watchdog hâlâ00007'de; ona dokunulmadı. BOOT5 ayrıntılı bağımsız sonuç `SRE_BOOT5_VERIFY.md` içinde.

## us-watchdog deploy (08.11.2026, Demir doğrudan onayı 00:44 TSİ; ön koşullar SRE_BOOT5_VERIFY.md'de PASS)

| Kontrol | Sonuç | Kanıt |
|---|---|---|
| Önceki: us-watchdog-00007-fnc %100 | PASS | raw_cloudrun/us_watchdog_deploy/ |
| Temiz context (base-00007 + us-watchdog/local-acceptance.patch), yerel test | PASS: 138 passed / 37 skipped. 4 commerce testi bu context'te kapsam dışı (önceki koşudaki "deselected" ile aynı) | local_tests.txt |
| Build 1a5bfe03 (europe-west4), image gateway:2026-11-08-sre-watchdog-ready, digest sha256:5ad490386738a1006f667497c2891cd392b4af5febac59a8018c0bd8a3c8a64a | PASS | build.log, digest.txt |
| Aday us-watchdog-sre-ready-r1, trafiksiz; spec (SA/env/volume/annotation) 00007-fnc ile aynı | PASS | spec_*.json |
| Aday: kimliksiz /health 403; yetkili /health 200; /ready 200 (config, store true) | PASS | candidate_checks.txt |
| Aday: 429, 18. istekte geldi; Retry-After 1; JSON gövde; log severity WARNING | PASS | candidate_checks.txt, post_promote_monitor.txt |
| JSON log (severity, httpRequest, trace) var, secret izi yok | PASS | candidate_logs.json |
| Trafik %100 sre-ready-r1, 22:03:16Z | PASS | promote.log |
| Trafik sonrası ~5 dk: 8/8 Scheduler reconcile 200, ERROR 0, LifecycleError/refusing 0; ds-a100 TERMINATED (GPU açılmadı) | PASS | post_promote_monitor.txt |
| Trafik artık sabit (latestRevision=false) | NOT | — |

Geri dönüş: `gcloud run services update-traffic us-watchdog --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-watchdog-00007-fnc=100`
