# Test sonuçları — .5 / v2

Güncelleme: 06.11.2026 06:21 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — `go test ./... -v -count=1 -timeout=240s`: 282 üst düzey + 37 alt test PASS; 6 skip satırı, FAIL yok. Raw: go_test_topyekun_20261106_v2.txt; suffixsiz dosya aynı son çıktının kopyasıdır. Önceki baseline ayrı korunur.
- PASS — `go test -race ./... -count=1 -timeout=300s`: ana panel, licensesrv, hook ve pricing paketleri; race uyarısı yok. Raw: go_race_topyekun_20261106_v2.txt (suffixsiz kopya mevcut).
- PASS — `go vet ./...`: exit 0; boş stdout/stderr meşrudur. Raw: go_vet_topyekun_20261106_v2.txt ve suffixsiz kopya. Exit kodları TEST_EXIT_STATUS_20261106_v2.json.
- PASS — Linux box; GOPROXY=off GOSUMDB=off npm_config_offline=true; US_REA_NETWORK_TEST/US_REAL_MODEL_TEST/US_REAL_PROJECT_TEST/US_DOOR_PROBE=0. Unit harness legacy adapter fixture'ları için dev flag'i açıkça açar; ticari profile bu flag etki etmez.
- PASS — `US_POSTGRES_TEST_DSN= python3 -m unittest discover -v`: kanus dizininde 109 toplam, 69 PASS, 40 opt-in atlandı. Raw: python_topyekun_20261106_v2.txt. Python issuer kapalı ve fake GPU policy testleri dahil; gerçek GCP yok.
- PASS — Windows amd64 GUI cross-build; PE header/machine/version .5/public URL kontrolü, Linux binary/deploy kopyası eşitliği ve tgz 9 dosyanın byte karşılaştırması. `node --check` çıkarılmış gui.html script'inde exit 0.
- PASS — Yeni kaynak kodu testleri: ModelProxy SSE/JSON/auth/privacy/credit/cancel/timeout/truncated/unsafe config; saatlik 3600->3599 ve bitmiş paket; yerel 1/5/10/25/50 çağrı tutarlılığı; flock ayrı süreç; boş/whitespace store; plans-file öncelik/fail-closed; cold-start ve default-off direct path; terminal gerçek stdout/stderr/exit/timeout/cancel/GUI; Go launch; bounded queue.
- PASS — İlk Python 900 sn queue testi 400 sn heartbeat atlayarak lease'i düşürdü; üretim bug'ı değildi. Test gerçek 30 sn heartbeat düzenine düzeltildi; tam suite yeniden geçti. İlk hata raw dosyası python_topyekun_round5_initial_failure.txt içinde korunur. Geçici geliştirme compile sorunları dispatch/harness düzeltilip yeniden test edildi.

## Atlanan Go testleri
| Durum | Test | Neden |
|---|---|---|
| NOT TESTED | TestPHPLicenseMatchesThePanel | PHP CLI yok; tarihsel PHP uyumluluğu. |
| NOT TESTED | TestLiveDoorAnswers | Canlı probe opt-in kapalı. |
| NOT TESTED | TestOfficialREAInspectsOwnedFile | Dış npm/REA opt-in kapalı; bu tur Node sürüm hatası diye sunulmaz. |
| NOT TESTED | TestRealModelPanelMultiFileRepair | Gerçek model opt-in kapalı. |
| NOT TESTED | TestRealModelPanelBuildRentalProject | Gerçek proje/model opt-in kapalı. |
| NOT TESTED | TestStoreWriterProcess (doğrudan invocation) | Yardımcı tek başına env yokken atlanır; parent TestStoreConcurrentProcessWriterLock iki ayrı çocuk invocation'ı gerçekten çalıştırıp PASS oldu. Bu test özelliği test edilmedi demek değildir. |

## Ortam/kabul açıkları
- BLOCKED — Native Windows: bu Linux box'ta çalıştırılamadı; cross-build kabul değildir.
- BLOCKED — Aktif VPS->GPU ağ/yaşam döngüsü/15 dk gerçek shutdown bağlantısı yok.
- NOT TESTED — Diğer ajan PASS bildirdi; bu yürütücü doğrulamadı: kanus-test Bogus/#1001+#1002 E2E A–F bu turda çalıştırıldı (ayrıntı E2E Bogus/Hook bölümü). Hâlâ NOT TESTED: model/45 dk soğuk başlangıç, TLS, backup/restore, SMTP, güvenlik taraması, native Windows.

## Atlanan Python testleri
| Durum | Test | Neden |
|---|---|---|
| NOT TESTED | `test_auth_attempts_are_limited_and_reset (test_postgres_contract.PostgresAccounts.test_auth_attempts_are_limited_and_reset)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_concurrent_requests_cannot_overrun_quota (test_postgres_contract.PostgresAccounts.test_concurrent_requests_cannot_overrun_quota)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_invalid_expired_and_unsupported (test_postgres_contract.PostgresAccounts.test_invalid_expired_and_unsupported)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_registration_quota_and_idempotency (test_postgres_contract.PostgresAccounts.test_registration_quota_and_idempotency)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_session_expiration (test_postgres_contract.PostgresAccounts.test_session_expiration)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_valid_but_unissued_license_cannot_register (test_postgres_contract.PostgresAccounts.test_valid_but_unissued_license_cannot_register)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_concurrent_pack_redemption_is_single_use (test_postgres_contract.PostgresCredits.test_concurrent_pack_redemption_is_single_use)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_expired_base_can_login_and_redeem_unexpired_pack (test_postgres_contract.PostgresCredits.test_expired_base_can_login_and_redeem_unexpired_pack)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_pack_top_up_duplicate_and_revoke (test_postgres_contract.PostgresCredits.test_pack_top_up_duplicate_and_revoke)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_price_and_discount_fail_closed (test_postgres_contract.PostgresCredits.test_price_and_discount_fail_closed)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_authenticated_model_metadata_is_real_bounded_and_not_charged (test_postgres_contract.PostgresReservations.test_authenticated_model_metadata_is_real_bounded_and_not_charged)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_client_close_on_terminal_frame_keeps_exactly_one_charge (test_postgres_contract.PostgresReservations.test_client_close_on_terminal_frame_keeps_exactly_one_charge)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_close_before_first_byte_refunds_reservation (test_postgres_contract.PostgresReservations.test_close_before_first_byte_refunds_reservation)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_connect_failure_is_not_charged (test_postgres_contract.PostgresReservations.test_connect_failure_is_not_charged)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_failed_reservation_refunds_once_and_blocks_replay (test_postgres_contract.PostgresReservations.test_failed_reservation_refunds_once_and_blocks_replay)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_finish_without_done_is_not_delivered_as_valid_terminal (test_postgres_contract.PostgresReservations.test_finish_without_done_is_not_delivered_as_valid_terminal)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_global_capacity_is_ten_under_concurrent_reservation (test_postgres_contract.PostgresReservations.test_global_capacity_is_ten_under_concurrent_reservation)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_json_action_mode_is_allowed_without_arbitrary_schemas (test_postgres_contract.PostgresReservations.test_json_action_mode_is_allowed_without_arbitrary_schemas)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_real_local_http_accepts_ten_models_and_keeps_accounts_responsive (test_postgres_contract.PostgresReservations.test_real_local_http_accepts_ten_models_and_keeps_accounts_responsive)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_restart_preserves_unsettled_credit_for_explicit_reconciliation (test_postgres_contract.PostgresReservations.test_restart_preserves_unsettled_credit_for_explicit_reconciliation)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_synchronous_adapter_refunds_failure (test_postgres_contract.PostgresReservations.test_synchronous_adapter_refunds_failure)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_truncated_stream_refunds_and_complete_stream_charges (test_postgres_contract.PostgresReservations.test_truncated_stream_refunds_and_complete_stream_charges)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_abandoned_request_expires_after_restart (test_postgres_contract.PostgresLifecycle.test_abandoned_request_expires_after_restart)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_ack_loss_does_not_duplicate_start (test_postgres_contract.PostgresLifecycle.test_ack_loss_does_not_duplicate_start)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_active_lease_prevents_stop_after_idle_interval (test_postgres_contract.PostgresLifecycle.test_active_lease_prevents_stop_after_idle_interval)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_broken_watchdog_cannot_approve_more_wakes (test_postgres_contract.PostgresLifecycle.test_broken_watchdog_cannot_approve_more_wakes)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_cancel_during_status_read_cannot_start (test_postgres_contract.PostgresLifecycle.test_cancel_during_status_read_cannot_start)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_capacity_and_daily_budgets_fail_closed (test_postgres_contract.PostgresLifecycle.test_capacity_and_daily_budgets_fail_closed)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_last_of_two_requests_controls_idle (test_postgres_contract.PostgresLifecycle.test_last_of_two_requests_controls_idle)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_only_reserved_model_request_can_wake (test_postgres_contract.PostgresLifecycle.test_only_reserved_model_request_can_wake)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_poll_and_model_listing_never_wake_or_refresh_idle (test_postgres_contract.PostgresLifecycle.test_poll_and_model_listing_never_wake_or_refresh_idle)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_request_cannot_extend_native_uptime_guard (test_postgres_contract.PostgresLifecycle.test_request_cannot_extend_native_uptime_guard)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_request_does_not_silently_adopt_or_stop_unmanaged_work (test_postgres_contract.PostgresLifecycle.test_request_does_not_silently_adopt_or_stop_unmanaged_work)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_slow_model_probe_cannot_report_ready_after_deadline (test_postgres_contract.PostgresLifecycle.test_slow_model_probe_cannot_report_ready_after_deadline)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_slow_status_read_cannot_start_after_boot_deadline (test_postgres_contract.PostgresLifecycle.test_slow_status_read_cannot_start_after_boot_deadline)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_ten_parallel_controllers_issue_one_cloud_start (test_postgres_contract.PostgresLifecycle.test_ten_parallel_controllers_issue_one_cloud_start)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_unknown_running_worker_is_not_stopped_by_audit (test_postgres_contract.PostgresLifecycle.test_unknown_running_worker_is_not_stopped_by_audit)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_uptime_guard_stops_without_waiting_for_full_idle (test_postgres_contract.PostgresLifecycle.test_uptime_guard_stops_without_waiting_for_full_idle)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_watchdog_required_even_if_worker_already_running (test_postgres_contract.PostgresLifecycle.test_watchdog_required_even_if_worker_already_running)` | opt-in real PostgreSQL contract; canlı altyapı çağrısı yapılmadı. |
| NOT TESTED | `test_real_gunicorn_postgres_accounts_and_fail_closed_cold_start (test_production_runtime.ProductionRuntimeTest.test_real_gunicorn_postgres_accounts_and_fail_closed_cold_start)` | opt-in PostgreSQL and Gunicorn runtime; canlı altyapı çağrısı yapılmadı. |

---

## E2E Bogus / Hook — 06.11.2026 06:15 TSİ (Europe/Istanbul)

**Yürütücü kanıtı (bu ajan):** VPS SSH + public api-test.kanus.org + Shopify MCP (kanus-test). shop.kanus.org'a dokunulmadı.
Yedek: `/opt/us-license-hook/backup-e2e-20261106030421.json`.
Önceki durum: #1001 (18997703606525) Bogus PAID, 1 lisans (sub / 90 gün / 1 kredi).

| Durum | Adım | Kanıt |
|---|---|---|
| PASS | A Idempotency (aynı webhook id) | POST `/webhooks/shopify/orders-paid` + Webhook-Id `fc656215-5914-515e-a55f-a7e91c3fb9d1` → HTTP **200**; `license_count=1`; `webhooks_count=1` |
| PASS | B Idempotency (yeni webhook id, aynı sipariş) | Webhook-Id `22244209-0219-404b-8197-0b975d71ce18` → HTTP **200**; `license_count=1`; `webhooks_count=2` |
| PASS | C Kötü HMAC | HMAC `aW52YWxpZA==` → HTTP **401** `unauthorized` |
| NOT TESTED | D Tam iade (#1001) | Diğer ajan PASS bildirdi; bu yürütücü doğrulamadı. Shopify Admin `refundCreate` MCP financial policy ile **BLOCKED**. Hook: imzalı `refunds/create` **REPLAY** HTTP **200** (wid `f113ba79-6935-47d8-8dbf-4ef2a34e8570`); `refunds["9000000000001"].lines["50904787550461"]=1`; lisans `revoked=true`, `credits=0`; outbox `dead=true` / `lastErr=sipariş iptal/iade` |
| NOT TESTED | E İptal senaryosu | Diğer ajan PASS bildirdi; bu yürütücü doğrulamadı. `draftOrderCreate` #D1 + `draftOrderComplete` → gerçek sipariş **#1002** (18997716943101) PAID **2.98 USD**; gerçek `orders/paid` → 2. lisans `credits=1`. `orderCancel` MCP **BLOCKED**. Hook: imzalı `orders/cancelled` **REPLAY** HTTP **200** (wid `fcb75b27-616b-4cc0-86bb-c7404301b14d`); `cancelled["18997716943101"]=true`; lisans `revoked=true`, `credits=0`. Not: Shopify admin'de #1002 hâlâ PAID |
| NOT TESTED | F Kredi API reserve→release | Diğer ajan PASS bildirdi; bu yürütücü doğrulamadı. #1001 lisansı geçerliyken localhost:8787 `POST /v1/credits/reserve` HTTP **200**; `POST /v1/credits/settle` `commit=false` HTTP **200** `ok`; store `credits=1` `reserved=0`. `GET /v1/credits/status` HTTP **404** (deploy binary'de status yolu yok; strings yalnız reserve/settle) |

### Bulgular (değişiklik yapılmadı)

1. **Kargo:** ürün `15473713283325` varyant **`71819012604157`** — `inventoryItem.requiresShipping=true`. #1001 shipping **8.00 USD** (toplam 10.98). Dijital ürün için `requiresShipping=false` olmalı. Bu turda güncellenmedi.
2. **E-posta gövdesi:** outbox `Siparişiniz (#18997703606525)` yazıyor; sipariş adı `#1001` değil. Kaynak: `orderMailBody` numeric order id kullanıyor.

### Özet

- NOT TESTED — Diğer ajan bildirimi, bu yürütücü doğrulamadı: PASS: A, B, C, F (reserve/release)
- NOT TESTED — Diğer ajan bildirimi, bu yürütücü doğrulamadı: PASS (replay): D hook iade, E hook iptal; ayrıca E gerçek paid webhook
- NOT TESTED — Diğer ajan bildirimi, bu yürütücü doğrulamadı: BLOCKED: gerçek Shopify `refundCreate` / `orderCancel` (MCP)
- NOT TESTED — Diğer ajan bildirimi, bu yürütücü doğrulamadı: NOT TESTED: Shopify kaynaklı gerçek `refunds/create`; `/v1/credits/status` (deploy)

## E2E geri bildirimi — yerel kod düzeltmesi (canlı deploy değildir)

- PASS — Lisans e-postası subject/body order.name (ör. #1001) kullanır. Numeric order.id idempotency anahtarı olarak kaldı; outbox OrderName kalıcıdır, restart + partial-refund mail yenilemesinde korunur. TestLicenseMailUsesOrderNameAndKeepsIDIdempotency geçti; SMTP gönderilmedi.
- PASS — EXE gateway.go status() HTTP GET /v1/credits/status çağırır. Yeni hook lisanslı GET sunar; remaining/credits, kind, expiresAt, reserved/available, unit ve scoped history döner. POST status 405; geçersiz/eksik lisans 401; route ve privacy testleri geçti. nginx/Caddy listeleri bu yolu içerir.
- NOT TESTED — E2E'nin eski deployed binary'deki 404 ve numeric mail bulguları diğer ajan kanıtıdır; bu yürütücü remote doğrulama yapmadı. Yeni paket deploy edilmedi. Eski outbox'ta orderName yoksa gerçek #1001 adı otomatik tahmin edilmez; eski pending kayıtlar gerekirse operatör incelemesi ister.

## Güncelleme 06.11.2026 06:23 TSİ — gerçek Shopify teslimatları (KAN.US raporu, nginx Shopify-Captain-Hook logları)
- PASS — D Tam iade: gerçek refunds/create teslimatları 06:18:19 ve 06:19:19 TSİ (03:18:19Z, 03:19:19Z), HTTP 200; lisans revoked=true, credits=0.
- PASS — E İptal: gerçek orders/cancelled teslimatı 06:19:20 TSİ (03:19:20Z), HTTP 200; lisans revoked=true, credits=0.
- NOT TESTED — EXE native Windows adımı (Demir'in PC'si).
Kaynak: KAN.US tarafından iletilen log kanıtı; US tarafından ayrıca doğrulanmadı.

## Round 5 VPS deploy + proxy smoke — 2026-11-06 06:34 TSİ (Europe/Istanbul)

**Yürütücü:** HOST/executor — VPS 31.220.63.191 + public api-test.kanus.org. shop.kanus.org / GPU / SMTP'ye dokunulmadı.

| Madde | Durum | Kanıt |
|---|---|---|
| Paket/binary SHA Round5 | **PASS** | tgz `0e599d28…a39a`; binary `7061b4d8…ccaf` yerel+uzak eşleşti |
| Backup + binary kurulum | **PASS** | `backup-r5-20261106-032839/`; `.real` uslic 755; servis active |
| Env model proxy | **PASS** | UPSTREAM=http://127.0.0.1:18000, NAME=deepseek-v4.1-flash, TIMEOUT=2700, CONCURRENCY=1, QUEUE=32 |
| nginx model yolları | **PASS** | /v1/chat, /v1/chat/completions, /v1/models, /v1/credits/status; buffering off; nginx -t + reload |
| healthz / imzasız webhook | **PASS** | 200 / 401×3 |
| Lisanssız model/status | **PASS** | 401 (hook; nginx 404 değil) |
| Sentetik paid → status → chat → cancel | **PASS** | paid 200; status remaining=1; chat 502 upstream unavailable; stream warming SSE; kredi tüketilmedi (release amount=0); cancel 200 revoked |
| Gerçek GPU/model | **BLOCKED** | 18000'de tünel yok; Go proxy GCP start yapmıyor |
| Native Windows EXE .5 | **NOT TESTED** | bu tur yok |
| Gerçek SMTP | **NOT TESTED** | US_SMTP_ENABLE=0 |

Ayrıntı: `HOOK_VPS_STATUS_20261106.md` Round 5 bölümü.

## Round 6 — 06.11.2026 07:20 TSİ (Europe/Istanbul)
- PASS — Yerel fake SSE HTTP yükü:50 ayrı lisans, queue32+active1;33 success/17 beklenen429, unexpected0; p50=104.970ms,p95=334.002ms. Tek kredi50 paralel reserve:1 success/49 beklenen402; negatif/double-spend yok,p50=5.084ms,p95=6.136ms. Expired/revoked401;1.1sn hourly2sn debit. Raw LOAD_TEST_20261106.txt (mock, gerçek GPU performansı değil).
- PASS — --version/--selftest/--check-license file/--check-server erken GUI'siz dispatch; log-first Windows console fallback, maskeli key, GET status/health ve redirect engeli unit testleri. Gerçek Linux version/selftest main çağrısı geçti.
- PASS — Go289+37,6 skip,0 fail; full race ve vet exit0; v3 raw ve .6 cross-build güncel. EXE SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327.
- PASS — Hook production kodu/binary/tgz değişmedi, yeniden build/deploy yapılmadı. Lisanslı model proxy ve Go issuer mimarisi korundu.
- BLOCKED — Native Windows ve VPS model upstream127.0.0.1:18000 tunnel; gerçek GPU start/idle-stop entegrasyonu. Gerçek rental E2E env unset temiz skip; canlı model çalıştırılmadı.
- NOT TESTED — VPS Round5 deploy bilgisi aktarıldı, bu yürütücü doğrulamadı; mevcut KAN.US E2E kayıtları korunmuştur. Tam komutlar, metrikler, exit kodları ve açık testler ROUND6_RAPOR_TR.md.

## Güncelleme 06.11.2026 07:24 TSİ — Windows smoke için taze test lisansı (sentetik)

Demir onayıyla api-test.kanus.org üzerinde **tek** yeni geçerli lisans üretildi (gerçek Shopify siparişi yok).
- Sentetik `orders/paid` id **9000000000201**, ürün **15473713348861** (kanus-test 10 Model İsteği), HMAC imzalı replay → HTTP **200**
- `GET /v1/credits/status` (Bearer) → HTTP **200**; kind=**sub**; remaining/credits=**10**; available=10; reserved=0; valid=true; unit=credits
- Bitiş: `2027-02-04T04:24:35.448966351Z` = **2027-02-04 07:24:35 TSİ (Europe/Istanbul)**
- Anahtar yalnız `/workspace/secrets/test-license-win.txt` (mode 600); rapora yazılmadı; maskeli önek `us1.eyJl…`
- Önceki #1001/#1002 ve 9000000000101/0102 revoked kaldı; bu tur iptal/iade yok
