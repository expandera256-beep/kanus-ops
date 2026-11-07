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
- NOT TESTED — Gerçek Shopify Bogus Gateway, ödeme/iade, model/45 dk soğuk başlangıç, TLS, backup/restore, SMTP ve güvenlik taraması yapılmadı. Yerel 50 istek testi gerçek GPU benchmark değildir.

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

## E2E geri bildirimi — yerel kod düzeltmesi (canlı deploy değildir)

- PASS — Lisans e-postası subject/body order.name (ör. #1001) kullanır. Numeric order.id idempotency anahtarı olarak kaldı; outbox OrderName kalıcıdır, restart + partial-refund mail yenilemesinde korunur. TestLicenseMailUsesOrderNameAndKeepsIDIdempotency geçti; SMTP gönderilmedi.
- PASS — EXE gateway.go status() HTTP GET /v1/credits/status çağırır. Yeni hook lisanslı GET sunar; remaining/credits, kind, expiresAt, reserved/available, unit ve scoped history döner. POST status 405; geçersiz/eksik lisans 401; route ve privacy testleri geçti. nginx/Caddy listeleri bu yolu içerir.
- NOT TESTED — E2E'nin eski deployed binary'deki 404 ve numeric mail bulguları diğer ajan kanıtıdır; bu yürütücü remote doğrulama yapmadı. Yeni paket deploy edilmedi. Eski outbox'ta orderName yoksa gerçek #1001 adı otomatik tahmin edilmez; eski pending kayıtlar gerekirse operatör incelemesi ister.

## Round 6 — 06.11.2026 07:20 TSİ (Europe/Istanbul)
- PASS — Yerel fake SSE HTTP yükü:50 ayrı lisans, queue32+active1;33 success/17 beklenen429, unexpected0; p50=104.970ms,p95=334.002ms. Tek kredi50 paralel reserve:1 success/49 beklenen402; negatif/double-spend yok,p50=5.084ms,p95=6.136ms. Expired/revoked401;1.1sn hourly2sn debit. Raw LOAD_TEST_20261106.txt (mock, gerçek GPU performansı değil).
- PASS — --version/--selftest/--check-license file/--check-server erken GUI'siz dispatch; log-first Windows console fallback, maskeli key, GET status/health ve redirect engeli unit testleri. Gerçek Linux version/selftest main çağrısı geçti.
- PASS — Go289+37,6 skip,0 fail; full race ve vet exit0; v3 raw ve .6 cross-build güncel. EXE SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327.
- PASS — Hook production kodu/binary/tgz değişmedi, yeniden build/deploy yapılmadı. Lisanslı model proxy ve Go issuer mimarisi korundu.
- BLOCKED — Native Windows ve VPS model upstream127.0.0.1:18000 tunnel; gerçek GPU start/idle-stop entegrasyonu. Gerçek rental E2E env unset temiz skip; canlı model çalıştırılmadı.
- NOT TESTED — VPS Round5 deploy bilgisi aktarıldı, bu yürütücü doğrulamadı; mevcut KAN.US E2E kayıtları korunmuştur. Tam komutlar, metrikler, exit kodları ve açık testler ROUND6_RAPOR_TR.md.

## Round 7 — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- PASS — Go hesap modu yerelde tamamlandı: register/login/balance/redeem/requests/workers ve verification/reset API'leri; bcrypt12, HMAC-expiry session, IP limit, aynı locked atomic JSON. us1 tek hesaba bağlanır; aynı hesap tekrar200, başka hesap409. Kredi kopyası/ikinci ledger YOK; raw key ve session aynı license/reservation/settlement'ı tüketir.
- PASS — Yerel fake SSE E2E register→login→redeem→balance→chat reserve/settle; hata/cancel release, hourly ceil saniye, karma session/raw son-kredi yarışı, refund ile bakiye azalması, eski JSON/restart testleri. SMTP allowlist fake sender testleri; gerçek SMTP kullanılmadı.
- PASS — Go 304 üst + 37 alt PASS, 6 beklenen skip,0 FAIL; full race/vet exit0. Ham çıktılar go_test/go_race/go_vet_topyekun_20261106_v4.txt. Linux gerçek headless version/selftest exit0; native Windows değildir.
- PASS — §8 tekrar (LOAD_TEST_20261106.txt son ROUND7):50lisans33HTTP200+17beklenen429; p50=102.374ms,p95=366.031ms; tek-kredi50reserve1HTTP200+49beklenen402,p50=3.491ms,p95=4.458ms; negatif/double-spend0,unexpected0. Expired/revoked401;1.1sn hourly→2sn. Bunlar localhost/fake SSE, GPU kapasite ölçümü değil.
- PASS — .7 accounts EXE ve yeni hook/tgz yerelde üretildi; PE amd64 GUI,9 arşiv dosyası kaynakla byte-eşit, binary kopyaları eşit, webhook secret template placeholder. Hashler BUILD_SHA256_20261106_v4.txt.
- BLOCKED — Native Windows, canlı accounts route kabulü (HOST yetkili deploy yapmadı), VPS18000 model tüneli, gerçek DeepSeek/proje kabulü ve Go GPU start/15dk idle-stop entegrasyonu.
- NOT TESTED — Gerçek SMTP/posta kutusu sahipliği, auth bağımsız güvenlik/vuln audit, Windows AV/code-sign. Önceki KAN.US/HOST canlı E2E girdileri korunmuştur; bu yürütücü canlı doğrulamadı. Ayrıntı ROUND7_RAPOR_TR.md.
