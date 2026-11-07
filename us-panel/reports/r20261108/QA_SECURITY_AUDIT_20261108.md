# QA — §12 Güvenlik Denetimi / §13 Test Kuralları

Güncelleme: **08.11.2026 01:00 TSİ (Europe/Istanbul)**. `prompt/KANUS_PROMPT_20261107.txt` tamamı (583 satır) okundu; §12:459–472 **14 madde**. Güncel matris: **11 PASS /1 FAIL /2 BLOCKED /0 NOT TESTED**. Tek açık FAIL kök neden F01dir; F02 fix doğrulandı. Secret Manager ve production restore BLOCKED. GPU bağımlı canlı kontroller ayrıca NOT TESTED, bu14e eklenmez.

Yerel14-case son retest: **12 PASS /1 FAIL /1 BLOCKED** (`qa-raw/current-source-local-retest.log`, `all-cases-results.json`); local restore PASS üst matriste production drill yokluğu nedeniyle BLOCKED. Önceki8P/4F/2B raporu `qa-raw/QA_SECURITY_AUDIT_BEFORE_HOST_REASSESSMENT.md`de korunur; eski başarısız testler silinmedi.

## Sürüm / provenance sınırı
- Paylaşılan `work2/kiyas/term/release.go:11` çalışma sırasında **2026.11.06.12** oldu;40 ürün dosyası önceki audit snapshotından değişmiş (`qa-raw/source-changes-since-audit.json`). Güncel14 yerel case yeniden koşuldu. Önceki full Go/race/vet PASSları eski source snapshotına aittir; yeni .12 için full-suite kabulü diye sunulmaz. Kaynak değişiklikleri QA tarafından yapılmadı.
- `dist/2026.11.06.11/` yeniden mevcut. F02 patchi bağımsız QA scratchına kopyalanmış **manifest205 dosyası doğrulanmış .11** üzerinde test edildi. Mutable work2 PASS tek başına release merge/deploy kanıtı değildir.
- HOST kanıtı `HOST_EVIDENCE_QA_20261108.md:2–13`: **srv2133566 /82.29.184.225**, çalışan test binary SHA `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0` (**pre-.11**); dış TLS/health200 ve proxy140.248.52.x nginx kayıtları. Yeni-host kimliği/rota **HOST-evidenced**, QA root SSH ölçümü değil. Box synthetic DNS198.18.0.1 nedeniyle kendi direct DNS/peer prechecki BLOCKED/0 HTTP; bu authoritative DNS hatası sayılmadı ve gate bypass edilmedi.
- HOSTun00:45 TSİ login/status/429 örnekleri bu workerın probları değildir. Kendi00:30:16–00:30:19 TSİ8 isteğinin tam timestamp+path listesi `QA_QUERY_LOG_REVIEW_20261108.md` sonunda; HOST exact-match sağlayabilir. Yeni remote istek **0**, önceki8 negatif request sıralı/120s timeout; credentiallı completion yok.

## Tam §12 matrisi
Go kökü `work2/kiyas/term/`, Python kökü `work2/kanus/`; L=`reports/r20261108/qa-raw/`. Kısa repro `python3 /workspace/us-panel/qa/audit12/run.py <arg>`. PASSlar yalnız belirtilen kapsam içindir.

| Madde | Durum | Kaynak file:line | Otomatik repro | Kanıt / sınır |
|---|---|---|---|---|
| Kullanıcı kimlik doğrulaması | **PASS** | `accounts.go:48–91,362–373,453–485` | `--case authentication` | L/authentication.log;remote-negative.json — Signed session expiry/reset/version/verification yerel PASS; uzak fake/missing401. Authenticated yeni-host kabulü yapılmadı. |
| API token güvenliği | **PASS** | `accounts.go:53–91;repair_loop.go:271–281` | `--case api_tokens; --case ssrf` | L/api_tokens.log;ssrf.log;patch-r11-verify/verification-results.json — Forgery denial PASS; F02 patched .11de ve güncel mutable .12 kaynakta redirect-contact/body/Authorization false. Release .12 merge/artifact kabulü ayrı bekler; hook binary bu desktop packageı içermez. |
| Secret Manager | **BLOCKED** | `cmd/us-license-hook/main.go:30–31,57,104,132; kanus/deploy/gcp_apply.py:64–78;gateway.service.yaml:30–47` | `--case secret_manager` | L/secret-manager-source-check.json;secret-scan.json;secret-candidate-classification.json — Env+SecretKeyRef kaynak bindingleri mevcut; taramada confirmed hardcoded real credential yok. Güncel IAM/version rotation/runtime binding kanıtı yok; secret/private/env/data dosyaları okunmadı. |
| Webhook HMAC | **PASS** | `licensesrv/webhook.go:34–45,72–93` | `--case hmac` | L/hmac.log;remote-negative.json — Constant-time HMAC; local bad/missing/reversal shop/topic testleri PASS; remote missing/wrong401. Gerçek webhook sırrı kullanılmadı. |
| SQL injection | **PASS** | `kanus/account_store.py:123–136,162–170;kanus/storage.py:21–35` | `--case sqli` | L/sqli.log;python-tests.log — Gerçek AccountStore/SQLite parameter-binding ile email/session payloadları auth alamadı; accounts tablosu ve quota sağlam. PostgreSQL gerçek DSN/migration ayrıca NOT TESTED. |
| Komut enjeksiyonu | **PASS** | `terminal.go:32+;safe_exec.go:223–258` | `--case command_injection` | L/command_injection.log;panel-audit.log — Terminal shell/meta/python-c argv reddi PASS. İzinli proje kodunun OS erişimi sandbox değildir, F01 ayrıca FAIL. |
| Path traversal | **PASS** | `file_tools.go:44–105,127–138` | `--case path_traversal` | L/path_traversal.log — ../,absolute outside,Windows separator ve symlink escape fileTools üzerinden reddedildi. OS child bypassı F01dir. |
| SSRF | **PASS** | `repair_loop.go:271–281;licensesrv/model_proxy.go:54–66;gpu_pin.go client policy` | `--case ssrf` | L/ssrf.log;patch-r11-verify/verification-results.json — F02 bağımsız .11 kopyasında önce FAIL, patch sonrası PASS; default/caller client guardları PASS. Gerçek metadata/ağ saldırısı yapılmadı. |
| Rate limiting | **PASS** | `accounts.go:257–290,344–360;login_limiter.go:10–16,30–80;accounts.go:362–373,466–468,484` | `--case rate_limit` | L/rate_limit.log;../qa-20261106/ratelimit-evidence.txt — Account10/900s normalized email/local rotating-IP testi PASS. Legacy/nginx IP-only DESIGN GAP (F04), FAIL değil; ek kaynak mitigasyonu var, yeni deploy etkinliği henüz doğrulanmadı. |
| Yetki yükseltme | **PASS** | `accounts.go:53–100,110–118,201–221;credits.go authorization/settle` | `--case privilege_escalation` | L/privilege_escalation.log;go-test.log — Application credential/ownership düzeyinde farklı license reservation settle reddi ve idempotency PASS; root/kernel/container escape audit kabulü değildir. |
| Dosya izolasyonu | **FAIL** | `safe_exec.go:17–19,176–208,223–258;file_tools.go:44–105` | `--case file_isolation` | L/file_isolation.log — F01 güncel .12 ve patched .11de FAIL; aynı-UID sentetik sibling read. EXE/desktop-only paket sınırı Go deps ile doğrulandı, hookta bu kod yok. HOST ayrı uslic/uslic-test UIDs bildiriyor; canlı untrusted panel yürütmesi yok. |
| Hassas loglar | **PASS** | `licensesrv/privacy_mask.go:10–40;mail.go:63–70,97–110;accounts.go:384–387` | `--case sensitive_logs; check_logs.py` | L/sensitive_logs.log;log-leak-check.json;../QA_QUERY_LOG_REVIEW_20261108.md — Yerel masking/redaction PASS; bütün audit alt loglarında token/canary scan PASS. HOSTun7 journal satırı temiz. Ürün credential-query bulunmadı; QA loopback sentetik query bodyye taşındı. Tarihsel uzak leak doğrulanmadı; nginx tam request-target kaydı olmadan mutlak yokluk kanıtlanamaz. |
| Yedekleme | **PASS** | `deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh:8–10,23,28–29,37–43` | `--case backup` | L/backup-v2-local.log;backup.log;backup-v2-hash-comparison.json;../HOST_EVIDENCE_QA_20261108.md:16–20 — v2 local7 PASS, iki sentetik instance coverage/manifest/LAST_SUCCESS; v1 productionda kuruluymuş, F03 v2yle FIXED-PENDING-VERIFY. F05 live-test coverage HOSTa göre mitigated. Production scheduled run/full SHA bağımsız doğrulanmadı; restore BLOCKED. |
| Geri yükleme | **BLOCKED** | `licensesrv/store.go:101–150,153–185;kanus/storage.py:53–58;restore-us-license-hook-store.sh:21–33` | `--case restore` | L/restore.log;go-store-backup-restore.log — Yerel sentetik SQLite+Go Store restore/quota/history/lock PASS. HOST migration copy/hash açılışı bildirdi; planlı production restore drill/RPO/RTO kaydı yok, bakım penceresi bekler. |

## Bulgu durumları / routing
Promptta bot ownership tablosu yok; açık kullanıcı ataması dışında bileşen bazlı routing önerisidir.

| Bulgu | Önem / durum | Owner | Dosya | Repro |
|---|---|---|---|---|
|F01 same-UID sibling workspace|HIGH koşullu; **OPEN**|PANEL DEV + HOST/SRE|`qa-findings/FINDING-01.md`|`run.py --case file_isolation`|
|F02 repair redirect|MEDIUM latent/helper; **FIXED**, patched .11 verified; .12 merge/release acceptance pending|PANEL DEV / US|`qa-findings/FINDING-02.md`|`run.py --tree <patched-r11> --case ssrf`|
|F03 v1 all-missing backup exit0|MEDIUM; **FIXED-PENDING-VERIFY**|HOST + SRE|`qa-findings/FINDING-03.md`|`run.py --case backup` (v2); eski v1 repro finding içinde|
|F04 IP-only limiter|MEDIUM **DESIGN GAP, FAIL değil**; source-mitigated/live-unverified|API + HOST|`qa-findings/FINDING-04.md`|`run.py --case rate_limit`|
|F05 live api-test backup exclusion/no drill|HIGH; **MITIGATED-PENDING-VERIFY**, restore drill BLOCKED|**HOST**|`qa-findings/FINDING-05.md`|`python3 qa/audit12/test_backup_v2.py -v`|

### F02 / F01 bağımsız owner-patch doğrulaması
Patch `/workspace/us-panel-dev/patches/qa12/qa12_finding01_02.patch` SHA **c31ac8b8b015184d7911c4eadf0abde6155f41cefa3da24da25fd5badabf40dc**. US cevabı `qa-findings/US_YANIT_FINDING_01_02_TR.md` kaynak/iddia olarak okundu, sonuçları sahiplenilmeden test edildi.
`qa-raw/patch-r11-verify/context.json`: .11 TREE_MANIFEST205 dosya birebir; yalnız QA scratch kopyasına patch uygulandı, US/PANEL DEV kaynaklarına dokunulmadı; orijinal205 dosya sonunda yeniden doğrulandı.
- Aynı QA ssrf testi **önce FAIL true/true/true → patch sonrası PASS false/false/false** (redirect_contacted/authorization_forwarded/project_body_forwarded).
- F01 patched kopyada hâlâ **FAIL**, synthetic_sibling_read=true. Ownerın file-tools read/write traversal/symlink guardı ve env-root guardı **PASS**; bunlar OS izolasyonunu sağlamaz.
- Ownerın repro.sh mantığı da çalıştırıldı; yalnız çıktı yolları QA dizinine yönlendirilen kopya (US kaynak/loglarına yazılmadı), grep eşdeğer rg. SSRF underlying Go exit0, isolation underlying Go exit1. Wrapperın son echo nedeniyle shell exit0 oluşu test PASS sanılmadı.
- Bağımsız `go list -deps -json ./cmd/us-license-hook`: yerel paketler yalnız **kiyas/licensesrv ve kiyas/cmd/us-license-hook**. Kök kiyas package main yok. Linux ve GOOS=windows desktop GoFiles içinde repair_loop.go/safe_exec.go/file_tools.go var. Dolayısıyla **F01/F02 EXE/desktop-only, hook binary etkilenmiyor**; bu diğer hook güvenlik kontrollerini ortadan kaldırmaz. Kanıt `patch-r11-verify/*-deps.json`, `verification-results.json`.

### Backup / restore — HOST kanıtı ve yerel v2
HOST kanıtı`:16` F03ün eski DRAFT scriptle prod kurulumu sırasında geçerli olduğunu doğrular. `:18` saatlik işin yalnız durmuş prod storeuna baktığını, canlı api-test storeunun kapsam dışı olduğunu doğrular: F05 HIGH.
HOST`:20` ve00:48 TSİ aktarımı: v2 kurulu; prod :05/test :07, ayrı backup dizinleri; ilk iki koşum exit0/3dosya, manifest+LAST_SUCCESS; test backup SHA canlı store `a0daf1ea…` ile aynı deniyor. Bunlar HOST-provided, QA bağımsız prod ölçümü değil.
Yerel **7 test PASS**: iki instance ayrı doğru snapshot/manifest/LAST_SUCCESS; missing store/plans ve symlink2, invalid JSON3, unknown arg64, optional env yok0. Sabit yollar/owner sadece test kopyasında temp fixturea remap edildi; gerçek /opt,/etc, env/ledger/backup açılmadı.
Box v2 SHA **d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0** HOST `d77534e5…a0` fragmanıyla tutarlı; **tam remote hash eşitliği bağımsız doğrulanmış değil**. Kaynak cron halen yalnız prod; HOSTun installed-cron iddiasıyla kaynak senkronu bekler. v2nin test argümanı storeu kapsar, yalnız default prod invocation kapsamaz.
Sentetik SQLite backup/restore ve ayrı Go Store snapshot/reopen (499 kota/history/exclusive lock) PASS. HOSTun00:12 TSİ cutover copy/hash/start kaydı planlı restore drill yerine geçmez; **production restore/RPO/RTO BLOCKED**, onaylı bakım penceresi/owner kanıtı gerekir. VPSye dokunulmadı.

## Rate limit: per-IP / per-account ayrımı
`accounts.go:257–290`: eski15min generic limiter **group+IP** (`:258` RemoteAddr; yalnız loopback peer için X-Real-IP `:262–264`; key`:268`). Login group10 limit `:344–360`; accounta göre değildir. Nginx **5r/m burst2 $binary_remote_addr** (`reports/qa-20261106/ratelimit-evidence.txt`), rotating140.248.50.x kaynaklarda kovalar dolmaz: **DESIGN GAP**, FAIL değil;16 satır15×400/1×200/0×429 bütün32 denemeye genellenmedi.
Yeni `login_limiter.go:10–16,30–80` normalized account/email bazlı10/900s; `accounts.go:369–370,467,484` block/fail/reset. Aynı account/farklı11IP yerel10×401,11.429 PASS. Kaynak mitigasyonu canlı pre-.11 binaryde var diye iddia edilmedi.

## Secret Manager / hassas log / URL
Kaynak/env/SecretKeyRef loading var (`main.go:30–31,57,104,132`; `gateway.service.yaml:30–47`; `gcp_apply.py:64–78`). Gerçek secret/private/key/env/customer DB/backupların içi okunmadı; deploy modülü import/execute edilmedi. Son workspace kaynak/config taraması **16,322 dosya,4 aday**:3 synthetic QA password fixtureı (biri kopya) ve cryptography parserının35-byte header markerı, gerçek private-key payloadı değil. **Confirmed hardcoded gerçek secret yok**; heuristic scope, tam DLP değil. Güncel IAM/rotation/provisioning metadata yok: BLOCKED.
Yerel raw log token/provider/private-key/canary taraması alt dizinler dahil **87 log PASS/hit0** (`log-leak-check-post-patch-20261108.json`; önceki scanler ayrıca korunur). HOSTun7 journal satırında secret/email yok; tüm retention kabulü değil. Nginx query string loglar.
Ürün panel/desktop/hook/Python kodunda credential-query gönderen istek bulunmadı. QAda bulunan license query yalnız loopback synthetic local fixturedaydı; JSON bodyye taşındı, merkezi decoded-name/token query guardı eklendi (otomatik PASS). Saklanan iki başarılı eski remote A/B koşumu zaten GET body kullanıyor. **Gerçek keyin nginx loguna sızdığı doğrulanmadı**; arşivlenmemiş geçmişi kesin dışlamak için maskeli nginx target-count gerekir. Eski QA-A/order301 ve QA-B/order302 zaten06.11.2026 08:23:49/08:23:46 TSİ revoked; mevcut303/404 key exposure kanıtı yok. Ayrıntı, maskeli tarihsel inceleme penceresi ve exact probe listesi **`QA_QUERY_LOG_REVIEW_20261108.md`**.

## §13 test kuralları / harness sağlığı
Kök neden → negative repro → owner fix → aynı repro retest zinciri F02de bağımsız doğrulandı; F01 gerçek FAIL tutuldu, xfail yapılmadı. QA yalnız harness/rapor değiştirir; product patch yalnız yetkilendirilmiş kendi scratch kopyasına uygulandı, merge/commit/push/deploy yok.
- Güncel14-case yerel retest:12P/1F/1B. Sec8 **21 PASS/1 NOT TESTED**, e2e9 mock **3 PASS/2 NOT TESTED**; gerçek GPU/model değil.
- İlk sec8 fixture yeni default GPU waker nedeniyle503 verdi; log token_unreadable/status0 ve `gpu_wake.go:242` token kontrolü HTTP Do öncesi, ağ çağrısı gerçekleşmedi. Fixture `US_GPU_WAKE_DISABLE=1` ile düzeltildi; ilk hata ayrı korundu (`sec8-query-fix-initial-waker-fixture/`). Mock batchlerin actual concurrency en fazla2; n50 concurrency50 anlamına gelmez.
- Önceki snapshotta `go test ./...`, `go test -race ./... -timeout=12m`, `go vet ./...` PASS; race ilk240s bcrypt timeout kanıtı da korundu. Python109 test/40skip/exit0. Bunlar yeni .12 shared-tree full regression yerine geçmez; ownerın yeni release full-suite kanıtı beklenir.
- `qa/audit12/run.py --tree <Go-module>` artık her koşumda **benzersiz TSİ timestamp dizini `qa-raw/runs/…`** açar; sonuç/overlay/static-scan önceki kanıtı ezmez. RUN_DIR yazılır. Python fixture/shared scan QA-local kalır, Go hedefi --tree ile seçilir. Ayrı snapshot/initial fail kayıtları korunur.
- SMTP disable/fakeSender/local-only, gerçek allowlisted adreslere mail yok; webhook gerçek sırrı yok; customer/live-store yok; remote rate/load burst yok. Authenticated `/v1/chat` veya herhangi completions çağrısı api-testte **YASAK**; KAN canlı penceresi dışında GPU-dependent kontroller NOT TESTED. Shopify E2E SHOPa ait: `KANUS_SHOPIFY_E2E_20261108.md`, tekrar edilmedi.

## Kalan ihtiyaçlar
1. PANEL DEV + HOST/SRE: F01 trusted-single-user vs multi-user sandbox/UID/mount politikası. F02nin .12 release merge/artifact kabulü; source ve hook/EXE kapsamı karıştırılmamalı.
2. HOST: tam v2 installed SHA, iki installed cron satırı, gerçek scheduled LAST_SUCCESS/manifest doğrulama metadata; source cron senkronu; F05 restore drill için onaylı bakım penceresi ve RPO/RTO kaydı. Secret/customer backup içerikleri istenmiyor.
3. HOST: yukarıda bağlı exact00:30 TSİ probe korelasyonu ve ihtiyaten eski08:00–08:24 TSİ06.11 nginx credential-query maskeli sayısı (raw key yok).
4. SRE/HOST: güncel Secret Manager IAM/version/rotation ve env-provisioning permissions metadata. API/HOST: deploy hash ve account-based limiter canlı kabulü. Yeni .12 full regression/artifact acceptance owner kanıtı.

```bash
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case ssrf
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case file_isolation
python3 /workspace/us-panel/qa/audit12/run.py --case backup
python3 /workspace/us-panel/qa/audit12/check_logs.py
```
F01 komutu beklenen exit1dir. Uzak kontroller otomatik tekrar edilmez. Makine matrisi `qa-raw/FINAL_QA_AUDIT_MATRIX.json`, hashler önceki ve güncel ayrı dosyalardadır.

## Kapanış güncellemesi — 08.11.2026 01:05 TSİ

Kaynak: HOST_EVIDENCE_QA_20261108.md, "QA kapanış kanıtı" bölümü (HOST'un salt okuma kanıtı; QA'nın sunucuya erişimi yok).

- **FINDING-03 → FIXED.** HOST'un bildirdiği prod sha256 `d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0`, box'taki `deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh` ile QA tarafından birebir doğrulandı (sha256sum, 01:05 TSİ). Yerel v2 repro 7 PASS. cron: prod :05, test :07; LAST_SUCCESS prod 00:48:05, test 00:48:06 TSİ (HOST kanıtı).
- **FINDING-05 → MITIGATED (backup kapsamı).** api-test store saatlik yedekte (HOST kanıtı). Restore drill: **BLOCKED**, bakım penceresi için Demir onayı bekleniyor (HOST soruyor).
- **nginx query log korelasyonu → PASS (kapsam sınırlı).** access.log + access.log.1'de query'de hassas parametre/token biçimli değer: 0 satır; 8 QA probe'unun 8'i eşleşti; query kullanan tek istek sentetik command senaryosu. Sınırlama: eski VPS (srv1662550) cutover öncesi logları taranmadı; eşleştirme saat+path+status ile (Host alanı yok).

- **Restore drill:** HOST üzerinden iletilen Demir kararı (01:05 TSİ): "Restore drill'i şimdilik yapma". Madde **BLOCKED** kalır; bakım penceresi açılmayacak. Yerel sentetik restore PASS ayrı kanıttır, prod restore drill yerine geçmez.


## 08.11.2026 01:17 TSİ (Europe/Istanbul) — SRE / API / PANEL DEV girdileri: tarihçeyi koruyan güncelleme

Bu bölüm önceki kararları silmez; güncel yeniden değerlendirmeyi ekler. **Güncel14-madde matris: 10 PASS /3 FAIL /1 BLOCKED /0 NOT TESTED.** Secret Manager BLOCKED→FAIL (F06), rate limiting PASS→FAIL (yeni F07); F01 FAIL ve production restore BLOCKED sürer. F04ün tarihsel IP-only **DESIGN GAP, FAIL değil** kararı değişmedi. Makine matrisi **`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/UPDATED_QA_AUDIT_MATRIX.json`**; önceki FINAL matrix/history korunur. Alt test PASS sayıları güvenlik-policy FAILlarını gizlemez.

### 1. SRE — Secret Manager artık FAIL; FINDING-06 (SRE, MEDIUM)
Kaynaklar: `SRE_SECRETS_IAM_METADATA.md:4–46,61–87` + `raw_secrets_iam/secrets_list.json`, `ver_*.json`, `iam_*.json`, `run_services.json`, `instances.json`, `role_ustaWake.json`, `sas.json`, `project_iam.json`. Bunlar **SRE-delivered evidence-by-report/raw metadata**, QA bağımsız GCP ölçümü değil. QA yerel JSON consistency scriptiyle iddiaları çapraz doğruladı; yalnız token pathine lstat yaptı.
- 6 secret; rotation metadata yok. Credential secretların v1 create zamanları05.11.2026 23:36–23:39 TSİ; bu manual rotation hiç yapılmadığı ispatı değil, rotation yönetişim kanıtı yokluğu. Policy v5 create08.11.2026 00:21:40 TSİ, v1/v3/v4 ENABLED; policy credential değil.
- Secret-level accessor yalnız us-gateway/us-watchdog; public/project-wide accessor olmadığı olumlu kanıt korundu. Cloud Run latest bindingleri hem delivered rawda hem yerel gateway YAML:30–47/watchdog YAML:29–40ta görüldü. Repo template ile canlı config eşitliği iddia edilmez.
- ds-a100/usta-wake default compute SA + cloud-platform scope; effective custom ustaWake yetkileri compute get/start/stop/suspend/resume ile sınırlı. Scope tek başına secret read/Owner yetkisi değildir. sre-oidc-negtest enabled/rolsüz; us-builder ve watchdog DB ihtiyacı NOT VERIFIED.
- **F06 önemleri:** rotation yönetişimi MEDIUM; latest/eski policy versions LOW; VM identity/scope MEDIUM; unused SA LOW; eski token artığı MEDIUM (geçerlilik NOT VERIFIED). Path **`/workspace/secrets/usta-wake/a100.token.old`**, regular0600 uid/gid1000 yerelde stat ile doğrulandı; **içerik/değer/hash okunmadı**, rotate/delete/chmod yapılmadı, exposure gösterilmedi.
- **HOST NOT VERIFIED:** canlı `/etc/us-license-hook*.env`, wake.token/pin mode/owner ve unit User/Group erişimi. Docs600/640 çelişkisi measured failure değildir; metadata-only teyit istenir.
- Sayı düzeltmesi: SRE aktarımı5P/5F/2NV; kaynak raporun açık tablosu **6P/4F/2NV**, eski-token riski ayrı satır değil. Bu alt kontrol sayıları QA14 ile birleştirilmez.
Kanıt **`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/sre-metadata-verification.json`**, stat `old-token-stat.json`. Yeni **`qa-findings/FINDING-06.md`** repro/öneri/owner ayrımını içerir.

### 2. API — .11 limiter kapsamı ve FINDING-07 (API, MEDIUM)
Kaynak `/workspace/us-panel-dev/patches/api-r11/fix-grace-overflow-on-11/README.md:47–66`. .11 copyde4 limiter/account kaynak dosyası TREE_MANIFEST ile eşit. Yerel kaynak: `login_limiter.go:10–107`; `accounts.go:48–52,257–290,344–374,467,484`.
**`run.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case rate_limit` PASS**: >=12 karakter sentetik parolalar, farklı IPlerde aynı normalized email10wrong401→11.429+Retry-After; doğru parola kilitliyken429; unknown email aynı; reset/sabit pencere/memory testleri PASS. Pure .11de alias yok (`main.go:108` yalnız /account/); API additive alias handler-sharing iddiasını pure .11de var saymadık.
Kapsam: account limiter **login-only**, ToLower+TrimSpace+ParseAddress exact;10failure/900s sabit pencere, env1–1000/10–86400, başarı reset. register5/reset5/verify10 yalnız IP15dk; redeem rate group boş, session gerekiyor. Short/invalid400 account failure sayılmaz.
Yeni **F07 MEDIUM**: credential bilmeyen biri10yanlış parola ile bilinen emailin doğru-password loginini15dk engelleyebiliyor; mevcut session erişimi bütünüyle kesildiği iddiası yok. Safety regression doğru-password200 beklerken429 ile **gerçek FAIL**; xfail yapılmadı.
Yerel ek reprolar: sabit pencere/unknown bucket/yeni-instance boş state; default10000-kova eviction spray hedef sayacını kaldırdı;4096 IP bucket doluyken yeni IP429;12 authenticated invalid-license redeem girişimi0×429/0 yeni bucket ve state mutation yok. Bunlar **4 davranış PASS /1 security FAIL**. Gerçek process restartı, multi-host veya canlı spray yapılmadı.
**Önemli API iddia düzeltmesi:** “IP429 Retry-After yok” **yanlış**: saf .11 response **Retry-After:5**, `accounts.go:317` → `backpressure.go:29–37` wrapper. Eksik-header findingi açılmadı; LOW not yalnız5snin gerçek15dk IP-window kalan süresi olmaması. İlk harness yanlış beklentisi ayrı korunup düzeltildi.
Kanıt **`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r11-rate-limit.log`**, **`r11-limiter-risks-wrapper-reassessed.log`**; `qa-findings/FINDING-04.md` dated scope update; yeni **`qa-findings/FINDING-07.md`**.

### 3. PANEL DEV — gerçek .12 source/artifact bulundu; iki bağımsız retest
Kaynak `/workspace/us-panel-dev/PANEL_S2_REVIEW_R12_20261108.md:14–34,161–175`. `dist/2026.11.06.12/` mevcut. Pinned/v12den **yalnız QA scratchına** kopyalandı; kaynak manifest **209/209** ve artifact SHA256SUMS **5/5** yerelde bağımsız doğrulandı. Manifest SHA `915cea46addb8680f7ad7907480d482e8145dfcf5200dd92e9da09640cf2bf38`. Binaryler çalıştırılmadı.
- `run.py --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term --case ssrf`: **PASS**, redirect_contacted/authorization_forwarded/project_body_forwarded **false/false/false**. F02 .12 source/artifact entegrasyonu artık doğrulandı; önceki merge-pending notunun bu kısmı aşılmıştır. Native EXE/live deploy kabulü değildir.
- Aynı tree `--case file_isolation`: **FAIL / synthetic_sibling_read=true**. F01 HIGH koşullu OPEN, safe_exec.go:17–19,176–258. Guardlar OS sandbox değildir; hook-only paket kapsamı değişmedi.
- PANEL DEV ayrıca hook/Windows EXE için genel **FAIL** veriyor; entitlement/cancel-wake/status-contract ve Windows path/junction TOCTOU bulguları kendi raporunda. Bunları bu iki QA testinin PASSı ile kapatmadık; bağımsız native/full regression yapılmadı, deploy PASS çıkarılmaz. Tarihsel path-traversal PASS yalnız test edilmiş yerel confinement kapsamıdır, Windows race kabulü değildir.
Kanıt **`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/targeted-retest-summary.json`**, `r12-artifact-sha-verification.json`, `r12-ssrf.log`, `r12-file-isolation.log`; F01/F02 findinglerine dated retest eklendi.

### Güncel durum / ihtiyaçlar
| Madde | Güncel durum | Gerekçe / owner |
|---|---|---|
|Secret Manager|**FAIL**|F06 SRE; canlı env permissions ayrı HOST NOT VERIFIED|
|Rate limiting|**FAIL**|F07 API account-lock DoS; işlevsel10/900s PASS, F04 DESIGN GAP tarihsel karar değişmedi|
|Dosya izolasyonu|**FAIL**|F01 PANEL DEV+HOST/SRE, .12de tekrar|
|Geri yükleme|**BLOCKED**|Production restore drill/onaylı bakım penceresi halen yok|
|Diğer10 madde|**PASS**, önceki dar kapsamlarıyla|F02 .12 retest PASS; tüm platform/deploy kabulü değil|

Reprolar:
```bash
python3 /workspace/us-panel/qa/audit12/review_sre_metadata.py --evidence /workspace/us-panel/reports/r20261108/raw_secrets_iam --output /workspace/us-panel/reports/r20261108/qa-raw/sre-recheck-$(date +%Y%m%dT%H%M%S%N).json
python3 /workspace/us-panel/qa/audit12/repro_limiter_risks.py --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term --case ssrf
python3 /workspace/us-panel/qa/audit12/run.py --tree /workspace/us-panel/qa/build/qa-r12-odheei69/term --case file_isolation
```
F07/F01 safety reproları beklenen açık FAIL/exit1. SRE metadata assertion PASSı güvenlik-policy PASSı değildir. SRE rotation/version/identity kararları ve API DoS fixi; HOST gerçek mode/owner/readability ve restore-drill metadata; PANEL DEV diğer .12 blockersı beklenir. Hiçbir IAM/secret/dosya silme, VPS write/restart veya cloud komutu uygulanmadı. **Bu güncellemede remote request0**; local testler sıralı/max2 sınırı içinde, model/GPU/mail çağrısı yok. Önceki kanıtlar ve rapor tarihçesi aynen korunmuştur.

### 08.11.2026 01:19 TSİ (Europe/Istanbul) — Son negatif-test ve kaynak koruma eki
F06 için aynı metadata ile `review_sre_metadata.py **--assert-safe**` ayrıca koşuldu: **5 policy riski, gerçek FAIL/exit1** (`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/sre-policy-negative-final.json`/.log). Önceki consistency PASSı güvenlik PASSı değildir. Repro:
```bash
python3 /workspace/us-panel/qa/audit12/review_sre_metadata.py --assert-safe --evidence /workspace/us-panel/reports/r20261108/raw_secrets_iam --output /workspace/us-panel/reports/r20261108/qa-raw/sre-safety-$(date +%Y%m%dT%H%M%S%N).json
```
Güncel nihai makine matrisi **`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/UPDATED_QA_AUDIT_MATRIX_FINAL.json`**: **10PASS/3FAIL/1BLOCKED**, değişmedi. Önceki tarihli matrix de korunur. QA .12 scratch209/209, PANEL DEV pinned/v12 209/209 ve US term.orig/.11 205/205 final hash eşliği doğrulandı; kaynaklara yazılmadı (`source-trees-final-check.json`). Şimdiye kadarki100 log taraması hit0; final tarama aynı tarihli kanıt dizinindedir. Tüm ham sonuçlar ayrı run dizinlerinde; model/GPU/remote request0.


## 08.11.2026 01:26 TSİ (Europe/Istanbul) — F01 platform risk kararı ve R13 regression hazırlığı
**F01 Windows: ACCEPTED-RISK — Demir'in08.11.2026 01:15 TSİ kararı, KAN.US üzerinden iletildi; QA doğrudan teyit etmedi. Sürüm notuna product/release owner yazacak. Linux: OPEN, .13te Landlock ile doğrulanacak.** Kaynak: koordinatörün01:20 TSİ steering aktarımı. KAN.USun “.13 Windows writes blocked via Low IL” iddiası **to be verified**, native Windows **NOT TESTED**. Yazma kısıtı okuma izolasyonu değildir; başka Windows TOCTOU veya F07 risk kabulü sayılmaz.
Tek komut hazır: **`qa/audit12/regress_r13.sh --tree <Go-source-tree>`**, native Python entry `regress_r13.py`; detay/exit kodları **`QA_REGRESSION_R13_COMMANDS.md`**. F02, F04, F07, Linux-Landlock sibling boundary ve runnable native Windows Low-IL write probe içerir. ABI unsupported/blocked ise SKIP+reason; inside control/expected test-run olmadan PASS yok.
Gerçek baseline doğrulaması manifest .12 ve F02-yamalı .11de **2PASS/2FAIL/1NOT TESTED**, exit1: F02/F04 PASS; F07 doğru-password429 DoS FAIL; F01 Linux sibling read FAIL. **Kernel6.12.113+ Landlock ABI6** destekli ama bu kaynaklar enforcement sağlamıyor. Windows fixture cross-compile PASS, runtime NOT TESTED; kernel skip guard6 unit test PASS. Kanıt `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r13-baseline-validation.json` ve ilgili loglar. .13 uygulama retesti henüz yapılmadı; hiçbir failure xfail yapılmadı.
F04 doğru-password429 sözleşmesi ile F07 aynı davranışın DoS riskini kapatma beklentisi aynı requestte birlikte PASS olamaz; API politika/recovery kararı gerekir. QA bunu gizlemedi. **14-madde audit sayıları10PASS/3FAIL/1BLOCKED olarak kalır** (F01 Linux OPEN). Remote/model/GPU/mail/VPS işlem0, local case parallelism1; önceki rapor tarihçesi korunur.

R13 platform durumlarını içeren son makine matrisi: `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/UPDATED_QA_AUDIT_MATRIX_R13_PREP.json`. JSONL dahil son **121 log taraması PASS/hit0** (`reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r13-final-log-leak-check.json`).

## Login kilidi politika kararı — 08.11.2026 01:28 TSİ

Kaynak: API botunun aktarımı (QA doğrudan teyit etmedi). Aktarılan Demir metni: "E-posta+IP ve artan bekleme olsun, ayrıca bilinen cihaz/oturum kilitten muaf olsun".
- FINDING-04 beklentisi güncellenecek: kilitteyken doğru parola 429 yalnız aynı e-posta+IP ve bilinmeyen cihaz için.
- FINDING-07 kapanış kriteri: başka IP'den veya bilinen cihazdan doğru parolayla giriş başarılı olmalı; artan bekleme doğrulanmalı.
- .13 regresyon beklentileri API'nin `patches/api-0017-login-limiter/` tasarım belgesi gelince ayarlanacak; o zamana kadar regress_r13.sh değiştirilmedi.
