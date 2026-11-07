# Üretim kontrol listesi — .5

Güncelleme: 06.11.2026 07:14 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme veya adlı kanıt dosyasıdır; canlı mağaza (shop.kanus.org) kabulü veya Demir onayı değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Go tek issuer/model proxy, server-side metering, SSE hata/cancel/timeout, status privacy, flock/empty-store/plans-file ve tam Go suite/vet/race; Python policy fixture'ları. Kanıt: KANUS_TEST_RESULTS.md.
- PASS — Windows cross-build/PE ve Linux/deploy/tgz/hash; env placeholders ve SMTP=0; no Cloud Run issuer. Kanıt: KANUS_TEST_RESULTS.md / BUILD_SHA256_*.
- PASS — kanus-test E2E (Bogus/hook): idempotency A/B, kötü HMAC C, kredi reserve→release F. Kanıt: E2E_HOOK_KANUS_TEST_20261106.md.
- PASS — kanus-test gerçek Shopify `refunds/create` ve `orders/cancelled` teslimatları (lisans revoke). Kanıt: KANUS_TEST_RESULTS.md «Güncelleme 06.11.2026 06:23 TSİ»; E2E_HOOK önceki PASS* replay ile uyumlu.
- PASS — Round 5 VPS deploy + sunucu taraflı model proxy yolları (nginx /v1/chat|/completions|/models|/credits/status; sentetik paid→status→chat 502 upstream; kredi düşmedi). Kanıt: HOOK_VPS_STATUS_20261106.md Round 5; KANUS_TEST_RESULTS.md Round 5 bölümü.
- BLOCKED — Native Windows smoke; gerçek GPU/model (18000 tünel yok); Go proxy→GCP start/15 dk idle-stop/watchdog.
- BLOCKED — Güçlü OS CPU/RAM sandbox ve tam hesap/oturum katmanı.
- BLOCKED — SMTP otomatik teslimatı bilerek kapalı (US_SMTP_ENABLE=0).
- BLOCKED — Shopify Admin MCP `refundCreate` / `orderCancel` financial policy (hook tarafı ayrı PASS).
- NOT TESTED — Canlı shop.kanus.org; gerçek ücretli müşteri siparişi (Demir onayı bekler); 45 dk cold-start ölçümü; TLS/backup-restore tam kabul; security scan; maliyet/fatura kabulü; `/v1/credits/status` eski binary 404 tarihsel — Round 5’te status yolu eklendi (HOOK_VPS_STATUS).
- NOT TESTED — Operatör tek writer için eski servisi yetkili bakımda durdurup yeni sürümü başlatmalı; gerçek sır/ledger ezilmemeli. Bu rapor deploy yetkisi veya Go-Live onayı değildir.
