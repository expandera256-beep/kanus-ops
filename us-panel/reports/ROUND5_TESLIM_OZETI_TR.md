# .5 teslim özeti — 06.11.2026 06:22 TSİ (Europe/Istanbul)

PASS yalnız yerel test/derleme/kaynak kanıtıdır; diğer ajanın canlı E2E kayıtları bu yürütücü tarafından doğrulanmadı ve korunmuştur.

| Spec / ek görev | Durum | Sonuç |
|---|---|---|
| §2 terminal/Go run/UI | PASS | Gerçek stdout/stderr, exit code, timeout/tree cancel, explicit argv, bounded output; Go/Python/Node fixtures. |
| §2 test/repair | PASS | Boş/skip-all/unknown çıktıda fake green yok; gerçek test kanıtı aranır. Korunan action/checkpoint/repair testleri geçti. |
| §2 OS CPU/RAM ve kullanıcı izolasyonu | BLOCKED | OS sandbox/container/cgroup/Windows Job limiti uygulanmadı; cwd/env/time sınırı bunun yerine geçmez. |
| §3 model/queue/stream/credit | PASS | Go lisanslı proxy + SSE/JSON, reserve-before-upstream, persisted terminal settle, hata/cancel/timeout release; bounded client/server queue. |
| Hourly use | PASS | 3600 saniye; queue hariç upstream wall-time ceil saniye, cold-start dahil; hata release, ambiguous legacy migration reddi. |
| İstemci/commercial | PASS | Go hook tek issuer/model authority; commercial direct SSH kapalı, dev flag commercial bypass sağlamaz. Warming/50 dk client +45 dk server wait. |
| Store | PASS | Boş store ve US_PLANS_FILE korundu; flock/LockFileEx build, fsync+rename, Linux ayrı-süreç tek-writer testi. |
| §4 idle 15 dk | PASS | Python policy 900 sn, boot600;899/900 ve active/queue guard fake-provider testleri. |
| Aktif Go GPU yönetimi | BLOCKED | Go proxy GCP/controller lease/watchdog'a bağlanmadı; VPS->GPU özel erişim kurulmadı. |
| E2E mail/status geri bildirimi | PASS | order.name subject/body, idempotency order.id, outbox restart/refund; EXE GET status ve licensed remaining/kind/expiry testleri. |
| Gerçek DeepSeek/Windows | BLOCKED | Canlı çağrı yasak; Linux box native Windows çalıştırmaz. Gerçek kalite/cold-start ve GPU maliyeti NOT TESTED. |

## Test kanıtı
- PASS — Go282 üst düzey+37 alt,6 skip satırı; full race ./..., vet, Windows/Linux cross-build exit0. Raw go_*_topyekun_20261106_v2.txt ve suffixsiz eş kopyalar; TEST_EXIT_STATUS_20261106_v2.json.
- PASS — Önceki değişmeyen Python raw:109 total,69 PASS,40 opt-in skip. GUI Node syntax PASS.
- NOT TESTED — Go skip listesi: TestPHPLicenseMatchesThePanel (PHP yok), TestLiveDoorAnswers (opt-in kapalı), TestOfficialREAInspectsOwnedFile (npm/network opt-in kapalı), iki TestRealModelPanel* (canlı model kapalı). TestStoreWriterProcess tek başına skip; parent iki ayrı process ile kilidi gerçekten test etti. 40 Python PostgreSQL/Gunicorn skip isimleri KANUS_TEST_RESULTS.md'de. FAIL kalan test yok.
- BLOCKED — Native Windows; VPS tunnel/IAM/firewall; active GPU start/idle-stop/watchdog; CPU/RAM/OS sandbox; Go tam hesap; SMTP teslimatı.
- NOT TESTED — TLS/proxy config runtime/restore, canlı model45dk, gerçek GPU yükü, güvenlik taraması, crash-after-commit replay/uzlaşma, finansal-only refund; eski pending outbox ad migration.

## HOST — yetkili sonraki adım için, bu tur uygulanmadı
Go127.0.0.1:8787 yolları:
- POST /webhooks/shopify/orders-paid; /webhooks/shopify/orders-cancelled; /webhooks/shopify/refunds-create
- POST /v1/chat; /v1/chat/completions (EXE)
- GET /v1/models; /v1/credits/status; /healthz
- POST /v1/credits/reserve; /v1/credits/settle (legacy; .5 model proxy kendi reserve eder)
Python account/internal/eski webhook yolları404. nginx/Caddy model buffering kapalı,50 dk timeout; raw HMAC/Authorization korunur. Üç Shopify topic api-test.kanus.org üzerinde aynı webhook path'lerine gider; var olan paid duplicate kaydı eklenmez.

Yeni env: US_MODEL_UPSTREAM; US_MODEL_NAME; US_MODEL_TIMEOUT_SECONDS=2700; US_MODEL_CONCURRENCY=1; US_MODEL_QUEUE_LIMIT=32; optional US_MODEL_UPSTREAM_TOKEN. Mevcut gerçek adlar SHOPIFY_SHOP_DOMAIN, SHOPIFY_WEBHOOK_SECRET, US_LICENSE_SECRET, US_PLANS_FILE (US_PLANS_JSON'dan önce), US_STORE_PATH, US_LISTEN; US_SMTP_ENABLE=0. Python SHOPIFY_SHOP/US_CREDIT_PRODUCTS kullanılmaz. Secret şablon placeholder kaldı.

BLOCKED — VPS upstream'e özel ağ veya VPS-origin SSH tüneli gerekir (örnek http://127.0.0.1:1923). Desktop tüneli yetmez. Hiçbir GCP/SSH/VPS/Shopify/SMTP/live çağrı yapılmadı; private upstream ve VM lifecycle kurulmadan GPU açılır iddiası yok.

## Artifact SHA256
```text
1ef0f70429953a79de7ea95f21b028f9fb0ad22412e9b399d792bc388423f863  /workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_5.exe
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/dist/us-license-hook_linux
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/deploy/us-license-hook/us-license-hook
0e599d2877087eb6efaf97f34fa711ad3263a18c272897d89cc52af88a6fa39a  /workspace/us-panel/deploy/us-license-hook_hostinger.tgz
```
Kaynak değişiklik listesi ROUND5_CHANGED_FILES_20261106.txt;73 satır spec matrisi NIHAI_SPEC_DURUM_TR.md; smoke EXE_SMOKE_CHECKLIST_TR.md; dry-run E2E_HAZIRLIK_TR.md. STATUS_TR ve istenen raporlar güncel; KAN.US E2E appendices korunmuştur.
