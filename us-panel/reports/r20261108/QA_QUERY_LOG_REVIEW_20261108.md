# QA — URL query / nginx hassas veri incelemesi (08.11.2026 TSİ)

## Sonuç
**Gerçek lisans/token/parola/e-posta için nginx URL sızıntısı doğrulanmadı.** Güncel ürün istemcilerinde credential-query gönderen bir istek bulunmadı; ürün bulgusu açılmadı. QAda bulunan tek açık license-query gönderimi **local() içindeki loopback/sentetik** senaryoydu; api-test değildi. Bu kullanım düzeltildi; merkezi request guard eklendi. Arşivlenmemiş bütün tarihsel sürümler veya nginx request-targetları elimizde olmadığı için “hiçbir geçmiş uzak istek asla query kullanmadı” garantisi verilemez.

## Ürün kaynak incelemesi
- `work2/kiyas/term/account_client.go:43–55,91–98`: email/password/license JSON body, session Authorization header.
- Panel license/model/gateway/worker istemcileri: `qa-raw/url-query-client-review.txt`, `url-query-review-source.txt`, `url-query-code-index.json`. GUI URLSearchParams alanları path/offset/hash/line; auth credentialı değil. Gateway ve model proxy upstream queryyi reddeder (`gateway.go:41`, `licensesrv/model_proxy.go`, NewModelProxy URL kontrolü).
- Python `work2/kanus/gce_provider.py` query parametreleri requestId/discardLocalSsd; credential değil. `deploy/gcp_apply.py:147–148` eşleşmesi Secret Managera kaydedilen PostgreSQL DSN authority password encodingidir; HTTP queryye credential gönderen istek değildir, DSN querysinde yalnız Cloud SQL socket hostu vardır. Modül import/execute edilmedi.
- Güncel `work2/kiyas/term`, `work2/kanus`, tarihsel teslim `source/from-demir-20261106` ve QA lexical scan: `qa-raw/url-query-postfix-code-check.json`. Guard negatif testlerinin sentetik URL metinleri ağ isteği değildir. Bu kaynak/heuristic kapsamıdır, bütün olası runtime veri akışlarının formal ispatı değildir.

## QA düzeltmesi ve log incelemesi
Eski `qa/sec8/run.py` local fixtureı `/v1/credits/status?license=<local B>` kullanıyordu; base `local_hook()` → `127.0.0.1`, lisanslar local sentetik 8101/8102 siparişlerindendi. Gerçek QA-A/B veya api-test bu kod yolunda kullanılmadı; production nginx bu loopback isteğini görmez.
Güncel `qa/sec8/run.py:99–100`: B lisansı **GET JSON body**, A sessionı Authorization header. Remote senaryo zaten `:57–64` GET body kullanır. `qa/common.py:14–18` decoded/casefold hassas query alan adlarını ve us1./usa1. token-shaped queryleri istek yapılmadan reddeder; hata URL/değer yazmaz.
```bash
python3 /workspace/us-panel/qa/audit12/test_url_query_guard.py -v
python3 /workspace/us-panel/qa/sec8/run.py --logs /workspace/us-panel/reports/r20261108/qa-raw/sec8-query-fix
```
Guard PASS (`qa-raw/url-query-guard.log`). Sec8 retesti **21 PASS/1 NOT TESTED**, e2e9 mock **3 PASS/2 NOT TESTED** (`sec8-query-fix.log`, `e2e9-fixture-retest.log`); ilk fixture-waker hatası ayrı korunmuştur. Yalnız loopback/mock; gerçek model/GPU çağrısı yok. GPU waker fixtureda açıkça kapalı; batchler en fazla2 request paralel.

HOST `HOST_EVIDENCE_QA_20261108.md:22–29`: journal 7 satırında secret/email görünmüyor; bu HOST-provided sınırlı örnek, tüm retention kapsamı kanıtı değil. Nginx query string logladığı için body/header seçimi önemlidir.
Tarihsel kayıt taraması `qa-raw/url-query-historical-log-scan.json`: saklanan QA loglarında credential-query/full-token hit0. `reports/qa-20261106/credential-leak-check.json`:45 dosya, hit0. Loglar her request-targetı tutmadığından yokluk tek başına kesin delil değildir.
Saklanan **iki başarılı** remote A/B koşumu (`sec8-remote-authenticated-final.log:14`, `sec8-harness.log:40`) açıkça “A bearer+B credential in GET body; credential URL/log içine yazılmaz” yazar. Diğer eski koşumlar `sec8-remote-final.log:11`, `sec8-initial-fixture-error.log:32` eksik A/B dosyası nedeniyle BLOCKED; cross-license isteği yok.

### Geçmiş belirsizliği kapatmak için HOSTtan istenecek yalnız maskeli kanıt
- Doğrulanmış uzak key leak **yok**; ihtiyat amaçlı eski nginxte **06.11.2026 08:00–08:24 TSİ** `/v1/credits/status` credential-query **sayısı** ve maskeli zaman/path/parametre-adları. Bu konservatif inceleme penceresidir; kesin eski gönderim zamanı veya ihlal iddiası değildir. Raw query/key değeri paylaşılmamalı.
- Olası tarihsel kapsam yalnız eski QA-A **us1.*** /order9000000000301 ve QA-B **us1.*** /order9000000000302; elimizde gerçek keyler yok. `revoke-qa-licenses-20261106.json`a göre B **06.11.2026 08:23:46 TSİ**, A **08:23:49 TSİ** revoke edildi; sonraki statüler valid=false/revoked=true. Yeniden aktif edilmeleri istenmiyor.
- Mevcut account lisansları order303/404ün URLye çıktığına dair kanıt **yok**. Sırf belirsizlik nedeniyle aktif lisansları sızdı diye etiketlemedik veya revoke/rotate etmedik.

## HOST için kendi önceki yeni-host probe korelasyon listem
Tüm zamanlar **08.11.2026 TSİ (Europe/Istanbul)**, response-completion timestamp; nginx request-start kaydı biraz önce olabilir. Ham sonuç `qa-raw/remote-negative.json`.

| Tamamlanma TSİ | Method | Path (secret yok) | HTTP |
|---|---|---|---|
|00:30:16.585481|GET|`/healthz`|200|
|00:30:17.013999|GET|`/v1/credits/status`|401|
|00:30:17.582162|GET|`/v1/models`|401 (fake token)|
|00:30:18.013258|POST|`/account/balance`|401 (fake session)|
|00:30:18.458577|POST|`/webhooks/shopify/orders-paid`|401 missing HMAC|
|00:30:18.942606|POST|`/webhooks/shopify/orders-paid`|401 wrong HMAC|
|00:30:19.430943|GET|`/v1/%2e%2e/%2e%2e/qa-synthetic`|400|
|00:30:19.911923|GET|`/v1/credits/status?command=qa%3Becho%20synthetic`|401|

00:38:36.206323 TSİ precheck yalnız DNS, **0 HTTP**. HOSTun 00:45 TSİ login200/status200/login429 örnekleri bu workerın probları **değil**. Lütfen yukarıdaki00:30 hedeflere redacted exact-match kanıtı veriniz; yeni istek üretmedik. Bu tur uzak ek request0.
