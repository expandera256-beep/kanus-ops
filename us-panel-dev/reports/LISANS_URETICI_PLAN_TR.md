# Seçilen plan (b) — Go tek lisans ve model otoritesi / .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Go us-license-hook VPS 127.0.0.1:8787; nginx/Caddy HTTPS api-test.kanus.org:443 -> Go. Python/Cloud Run issuer yok, Python webhook almaz. JSON US_STORE_PATH tek writer flock ile korunur; PostgreSQL/SQLite aktif dağıtım değildir.
- PASS — .5 EXE varsayılan US_LICENSE_URL=https://api-test.kanus.org ve model de aynı Go proxy. /v1/chat/completions server reserve/settle yapar; desktop imza sırrı veya ön-debit yok. US_GATEWAY_URL/SSH ticari yolda kapalı. .3/.4 yerine .5 gerekir.
- PASS — Env: SHOPIFY_SHOP_DOMAIN (Python SHOPIFY_SHOP değil), SHOPIFY_WEBHOOK_SECRET placeholder, US_LICENSE_SECRET placeholder, US_STORE_PATH, US_LISTEN, US_PLANS_FILE (öncelikli)/US_PLANS_JSON product_id planları, US_SMTP_ENABLE=0. Python variant/TRY US_CREDIT_PRODUCTS bu yolda kullanılmaz.
- PASS — Yeni env: US_MODEL_UPSTREAM=http://127.0.0.1:1923 (VPS tüneli), US_MODEL_NAME=deepseek-v4.1-flash, US_MODEL_TIMEOUT_SECONDS=2700, US_MODEL_CONCURRENCY=1, US_MODEL_QUEUE_LIMIT=32; isteğe bağlı private US_MODEL_UPSTREAM_TOKEN.
- BLOCKED — VPS -> ds-a100 özel erişim/tünel HOST sorumluluğunda, bu tur kurulmadı. Go aktif otomatik start/idle-stop adapter'ı yok; 15 dk policy yalnız mevcut controller/fake-provider seviyesinde testli. Müşteri doğrudan model/SSH erişimi firewall/IAM kabulüyle kapatılmalı.

## HOST route listesi
| Durum | Yöntem/yol | Hedef |
|---|---|---|
| PASS | POST /webhooks/shopify/orders-paid | Go 8787 |
| PASS | POST /webhooks/shopify/orders-cancelled | Go 8787 |
| PASS | POST /webhooks/shopify/refunds-create | Go 8787 |
| PASS | POST /v1/chat ve /v1/chat/completions | Go lisanslı proxy, SSE buffering kapalı, >=50 dk timeout |
| PASS | GET /v1/models | Go lisanslı salt-okuma metadata |
| PASS | GET /v1/credits/status | Go lisanslı bakiye ve scoped kullanım |
| PASS | POST /v1/credits/reserve ve /settle | Go legacy API; .5 kullanmaz, proxy ID dışarı verilmez |
| PASS | GET /healthz | Go salt-okuma health |
| PASS | /account/*, /internal/*, eski /webhooks/shopify | Konfigürasyon örneğinde 404; Python'a proxy yok |

- NOT TESTED — PASS burada route dosyası/unit kanıtıdır; canlı nginx/Caddy/TLS uygulanmadı. Ham HMAC gövdesi ve auth/Shopify başlıkları korunmalı, token loglanmamalı. Exact location örneği deploy/us-license-hook/ içinde.
- NOT TESTED — Tek üretici cutover'da HOST diğer issuer/abonelikleri yetkili bakımda kapatmalı, eski Go sürecini aynı ledger üzerinde yeni süreç başlamadan durdurmalı. .lock dosyasını elde tutulan süreç varken silmeyin; canlı ledger/secrets ezmeyin.
- NOT TESTED — kanus-test topicleri: orders/paid -> https://api-test.kanus.org/webhooks/shopify/orders-paid; orders/cancelled -> https://api-test.kanus.org/webhooks/shopify/orders-cancelled; refunds/create -> https://api-test.kanus.org/webhooks/shopify/refunds-create. Paid mevcutsa çoğaltılmaz.
- NOT TESTED — Bogus Gateway sequence E2E_HAZIRLIK_TR.md: Demir onayı -> test ödeme -> tek Go lisans -> .5 proxy/credit kontrolü -> cancel/refund/replay. SMTP açılmaz; gerçek işlem yapılmadı.
- BLOCKED — Tam hesap/oturum ve otomatik mail teslimatı yok; .5 lisans-scoped erişimdir. Sunucu proxy yazılımı yerel testle tamamlandı; ağ/Windows/GPU canlı kabulü olmadan üretim hazır değildir.
