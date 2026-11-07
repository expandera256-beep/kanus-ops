# Accounts-mode gap ve seçilen Go hesap tasarımı — 06.11.2026

Gözlem: 06.11.2026 07:45–07:46 TSİ (Europe/Istanbul). Son karar: **hesap modu, Go sunucusu gateway olarak kalır; Cloud Run kullanılmaz**. Bu dosyanın A–C kısmı implementasyon ÖNCESİ kaynak/önceki test kanıtıdır. PASS kaynak/yerel test kapsamıdır, canlı kabul değildir. Kod veya canlı yapılandırma bu analiz hazırlanırken değiştirilmedi.

## A — EXE accounts gateway sözleşmesi (work2/kiyas/term)

Önemli: releaseAuthority=accounts tek başına yeterli değildir; releaseChannel=commercial, embeddedGatewayURL=https://api-test.kanus.org gerekir. Başarılı login session s.key/accountGateway'e atanır (gui.go:524–553). US_GATEWAY_URL gateway override'ıdır; US_LICENSE_URL eski lisans/CLI adresidir. Commercial normal account çağrısında desktop reserve yapılmaz (gateway.go:82–97). Aşağıdaki auth “S” = Bearer account session; “L” = Bearer us1 lisans. Aynı URL'nin var olması token sözleşmesinin uyduğu anlamına gelmez.

| Method/path | Request ve auth | EXE'nin tükettiği response | Kaynak file:line | Go hook başlangıç | Python work2/kanus | Go eksikliği / ek saat tahmini |
|---|---|---|---|---|---|---|
| POST /account/register | {email,password,license}; auth yok | herhangi2xx; Python201 {ok:true} | account_client.go:13–41, gui.go:524–548 | FAIL yok | PASS commerce_http.py:132–139 | 8–12h: normalized unique email, password KDF, atomic optional license bind, outbox |
| POST /account/login | {email,password,license:""}; auth yok | {session:string32..256} | account_client.go:24–49 | FAIL yok | PASS commerce_http.py:132–140 | 6–10h: KDF verify, HMAC-expiring session, generic failures/IP limiter |
| POST /account/balance | {license:""}; S | {remaining:nonnegative int}; üst limit100000000 | account_client.go:54–85, gui.go:555–583 | FAIL yok | PASS commerce_http.py:141–155 | 4–6h: account→license ownership join, active/expiry/reserved düşümü; saniyeler ayrı |
| POST /account/redeem | {license:us1}; S | {remaining:int} | account_client.go:54–85 | FAIL yok | PASS commerce_http.py:152–155; account_store.py:139–169 | 6–10h: unique key-owner, aynı hesap idempotency, başka hesap reject, outstanding hold guard |
| POST /account/requests | {}; S | {requests:[{request_id,state:reserved/completed/refunded,created:epoch,updated:epoch}]}; max20 | account_client.go:88–138 | FAIL yok | PASS commerce_http.py:151; account_store.py:225 | 4–6h: mevcut reservations/settlements üzerinden scoped tarihçe; ikinci ledger yok |
| POST /account/workers | {}; S | {workers:[{worker,model,phase,provider_status,active,idle_seconds,idle_remaining:int/null,observed_at:epoch,live,model_ready}]}; max10 | worker_client.go:12–56, gui.go:606 | FAIL yok | PASS route commerce_http.py:145–150; controller yoksa503 | 4–8h: güvenli metadata adapter; provider/IAM gerçek durumu BLOCKED, uydurma RUNNING yok |
| POST /v1/chat/completions | S; X-US-Request-ID random32hex, X-Usta-Worker; {model,max_tokens,stream:true,think:false,messages:[{role,content}], optional response_format:{type:json_object}, reasoning_effort/reasoning_budget_tokens/temperature/stop/chat_template_kwargs} | OpenAI SSE choices[].delta.content/reasoning_content, finish_reason/[DONE]; JSON fallback; error; optional X-US-Request-ID eşleşmesi, X-US-Metered:1 | model.go:2057–2066,2219–2294 | BLOCKED route var ama yalnız L kabul eder; S uymaz | PASS commerce_http.py:156–206, metering_stream.py | 6–10h: S→kaynak lisans atomic seçim; aynı reserveModel/settleAmount; upstream'e S/L gönderilmez |
| GET /v1/models | S; body yok | {data:[{id,max_model_len/context_length/context_window,max_output_tokens,meta:{n_ctx}}]}; exact selected model | context_limits.go:72–125 | BLOCKED route var, L auth | PASS commerce_http.py:85–112 | 1–2h: aynı credential resolver (chat ortak işi) |

### Koşullu/yardımcı yollar — normal account debit değildir
- GET /v1/credits/status (gateway.go:167–205; gui.go:310–326): L, response kind/credits/remaining/reserved/available/history/valid/revoked/expiresAt/observedAt. Go PASS, Python FAIL yok. Başarılı account session'da creditGate başlangıçta boş döner; hesap UI balance yerine bu legacy butonu kullanırsa hata alır. Hedef S için de account aggregate gösterimi; 2–4h; saniyeler prompt sayısına eklenmez.
- POST /v1/credits/reserve {}→{id}, POST /v1/credits/settle {id,commit}→{ok:true}: gateway.go:132–164. Go PASS L, Python FAIL public route yok (işlem chat içinde). **Login olmuş accounts normal akışında çağrılmaz**, yanlış/eksik accountGateway + US_LICENSE_URL/legacy license kombosunda devreye girebilir; iki kez ücretlendirme olmaması korunmalı.
- CLI GET /healthz ve GET /v1/credits/status: smoke_cli.go:88–132. Mevcut CLI lisans dosyası L'dir, S değildir. Go health body `ok`; Python health JSON {ok:true,...}, bu nedenle aynı health parser Python'a taşınamaz. Yeni --check-account ayrı JSON credential dosyası + login/balance akışıdır (4–6h).
- GET /api/durum, GET http://127.0.0.1:11434/api/tags (main.go:289–325): development fallback probe; commercial accounts guard:290–291 nedeniyle çağrılmaz; Go/Python accounts gateway API'si değildir. Sırf bu isimler için Go route eklenmemeli.
- GET /v1/models tunnel.go:55 loopback dev-readiness yolu; accounts commercial doğrudan SSH değildir.

### Loopback GUI ve isteğe bağlı genel HTTP (gateway'e eklenmez)
EXE localhost4477 sunucusu: GET / →html (gui.go:134), GET /api/version→{version} (release.go:16; gui.go:143), GET /api/state→guiSnap (gui.html:350/gui.go:142), GET /api/files?offset→dosya sayfası (gui.html:358/gui.go:151), GET /api/source?path,sha256,offset/limit→range JSON (gui.html:738/gui.go:152), POST /api/act→guiSnap/error (gui.html:345/gui.go:153; {op,text,path,email,password,license,request_id,...}), GET /view?path veya ?engine=tuval→yerel proje içerik (gui.go:154,1153). Bunlar Go hook/Python route'ları değildir; mevcut panel sunucusunda PASS; uzak Go'ya proxy edilmesi güvenlik hatasıdır. Kullanıcının agent web-open işlemi arbitrary GET URL (agent.go:167,desk.go:503), cloud SDK archive GET (cloud_install.go:17,50), legacy mağaza simülasyonu (model.go:1271–1286) accounts auth API sözleşmesine dahil değildir; bu analizde yürütülmedi.

PANEL ADMIN'in /account/login404 bildirimi aktarılan, canlı doğrulanmamış bulgudur. Kaynakla Go'da eksiklik doğrulandı. **.6 go-hook** profilinde accountform gizlidir (pricing.go:49–55), account-* action daha HTTP çıkmadan409 olur (gui.go:272–274); dolayısıyla stock .6'nın gizli hesap ekranından404 geldiği iddiası ayrıca çalışan EXE SHA/4477 version/eski pencere üzerinden incelenmeli. .7 accounts build bu ekranı bilinçli açacaktır.

## B — Python/Go parity (implementasyon öncesi)
Kanıt G=reports/go_test_topyekun_20261106_v3.txt (289+37 PASS;6skip; race/vet exit0), P=reports/python_topyekun_20261106_v2.txt (109total/69PASS/40skip; bu tur yeniden çalıştırılmadı). Bunlar bu yürütücünün önceki yerel koşularıdır; live kabul değildir.

| Konu | Python | Go | Kanıt / önemli sınır |
|---|---|---|---|
| Model proxy | PASS | PASS L; BLOCKED S | commerce_http.py:156; model_gateway.py:10 synchronous adapter, HTTP stream ayrı; G TestModelProxySSEChargeAndNoLicenseLeak, P test_webhook_delivery_account_and_metered_model |
| reserve/settle/release | PASS | PASS | account_store.py:192–263; commerce_http.py:162–205; G TestModelProxyFailuresRelease, P test_connect_failure_is_not_charged/test_truncated_stream_refunds_and_complete_stream_charges; Go/Python ayrı ayrı mevcut, birlikte çalıştırmak ikinci ledger doğurur |
| Hourly per-second | FAIL | PASS | Python verify_license yalnız kind=sub (account_store.py:40); paid yalnız sub üretir(shopify_ledger.py:94); Go TestModelProxyHourlySecondsAndLegacyUpgrade, TestRound6ExpiredRevokedHourlySSE |
| SSE/warming | PASS; BLOCKED45dk | PASS yerel; NOT TESTED gerçek GPU | managed_inference.py:63–75 heartbeat; commerce_http.py:184 request600s; inference_lifecycle.py:48 max600; deploy/gateway.service.yaml:18 timeout600; Go45dk server/50dk client ve heartbeat testli. Python tam taşımada45dk yeniden tasarlanmalı |
| GET /v1/credits/status | FAIL yok | PASS L | Python POST /account/balance eşdeğer prompt balance, expiry/kind/seconds sözleşmesi değil; G TestCreditStatusIsGETAndRequiresLicense |
| Cancel/full refund | PASS | PASS | Python orders/cancelled ve orders/updated financial_status=refunded (commerce_http.py:122–126); P test_paid_order_register_use_and_cancel/test_full_refund_denies_gateway_and_partial_does_not_revoke; Go üçtopic ayrı path |
| Partial refund / refunds/create | FAIL | PASS quantity; NOT TESTED finansal-only | Python refunds/create yok, partial bilinçli revoke etmez; Go TestPartialRefundIdempotentAndOutOfOrder/TestFinancialOnlyRefundDoesNotInventCreditConversion |
| Webhook idempotency | PASS sınırlı | PASS | Python order_id+body_hash çatışması; değişmiş paid raw body aynıorder reject(P test_same_order_changed_body_rejected). Go delivery-id/order-id/refund-id+tombstone tests; depolar ayrı olursa global exactly-once YOK |
| HMAC/raw body/shop | PASS | PASS | shopify_ledger.py:39,107,131; commerce_http.py:120; G TestRejectsBadHMAC/TestReversalHMACShopTopicAndInvalidPayload; P test_bad_shop_and_signature_never_send |
| Store locking/atomicity | PASS SQLite; NOT TESTED productionPG | PASS single writer | storage.py:21–33 PG advisory transaction lock; runtime.py:15–19 production sharedPG şartı;40PG/Gunicorn tests skip. Go lifetime flock+fsync/rename; G TestStoreConcurrentProcessWriterLock; JSON autoscaling2replica desteklemez |
| Queue/concurrency | PASS farklı sözleşme | PASS | Python account_store.py:249 global max10 reserved, worker capacity<=10 + buffer8;32queue/clean429 parity yok. Go active1+queue32, overflow429; G Round6 50lisans33success17reject |
| Tam hesap/hash/session | PASS local | FAIL başlangıçta yok | Python scrypt+salt/session hash+24h(account_store.py:98–137); Go baseline account routes/schema yok |

Python production runtime **Go sole issuer** kararını kodda enforce eder: runtime.py:25–27 US_SHOPIFY_ENABLED=true ise startup fail. Sadece Cloud Run env değiştirerek Python issuer açılmaz. Katalog da farklıdır: Python US_CREDIT_PRODUCTS variant_id/TRY/price_minor; Go US_PLANS_FILE/US_PLANS_JSON product_id. kanus-test USD ve8product planları otomatik eşdeğer değildir. Python SHOPIFY_SHOP vs Go SHOPIFY_SHOP_DOMAIN. DB Python US_COMMERCE_DB (production PostgreSQL; local SQLite), Go US_STORE_PATH lockedJSON. Go lisansı Python issued_licenses'a güvenilir import edilmeden imza doğru olsa bile register reddeder(account_store.py:108–110).

## C — İzinli public Cloud Run GET probe (07:45:15–07:45:17 TSİ)
URL https://us-commerce-247516932580.europe-west4.run.app. Auth/body yok, redirect takip edilmedi; POST, credentials, control-plane/GCP/SMTP yok. Bu GET'ler son “Cloud Run yok” kararından ÖNCE tamamlandı; sonra tekrar sorgulanmadı.

| GET path | Durum | HTTP / ms / body |
|---|---|---|
| /healthz | FAIL |404 /251.6; Google-style HTML Error404; Content-Type text/html; uygulama health JSON'u değil |
| / | PASS probe, NOT TESTED auth |405 /402.5; application/json {error:method} |
| /account/login | NOT TESTED POST |405 /337.9; {error:method} |
| /account/balance | NOT TESTED POST |405 /386.4; {error:method} |
| /v1/models | PASS unauth rejection |400 /507.5; {error:request rejected}; model metadata alınmadı |
| /v1/credits/status | FAIL sözleşme |405 /436.4; {error:method} |
| /v1/chat/completions | NOT TESTED POST |405 /379.2; {error:method} |

405 genel method gate'inden de gelir; /account/login POST gerçekten var/çalışıyor kanıtı değildir. Diğer path'lerde uygulama JSON'u görülmesi “servis tamamen yok” iddiasını çürütür; health404 kök nedeni revision/ingress/path ayrımı kaynak/GET ile belirlenemez. Deploy/revision/IAM okunmadı. Bu URL üretim hazır/aktif issuer sayılmaz.

## D — Seçenek, efor, risk, tek-issuer

1. **BLOCKED / karar dışı: full Python Cloud Run.** Tahmini96–160 mühendis-saat + canlı kabul: issuer-policy değişikliği,8ürün variant/currency katalog eşlemesi, Go JSON→PG entitlement/reservation/settlement migration, hourly16–24h, partial refund12–20h,45dk timeout/watchdog12–20h, CLI health/status/client rebuild4–8h, rollback/audit. Hazır account API'si bu kayıpları gizlememeli. Çift lisans guard: önce eski Go webhook tüketimini durdur, consistent ledger/watermark aktar, pending/replay reconcile, tek Python webhook setini sonra aç; iki issuer aynı anda açık olamaz. Bu plan uygulanmaz, Cloud Run çağrısı/deploy'u yok.
2. **PASS karar / BLOCKED implementasyon öncesi: accounts + Go gateway.** Mevcut webhook/model/hourly/refund tek ledger korunur. A tablosu endpoint işleri ortak temel dahil toplam **56–88 mühendis-saat**; yeni confirm/reset/outbox/allowlist10–16h, migration/security/Windows/E2E8–16h ile **74–120h** planlama bandı (paralelleştirme/toplamlar birebir toplama değildir). Risk: hesap sahipliği, mixed seconds/prompts, reset/session invalidate, brute-force, mevcut JSON boyutu ve tek writer. Çift lisans guard: Shopify üçtopic yalnız Go; Python issuer kapalı; account redeem yalnız ownership link yazar, yeni license/credit üretmez. Aynı us1 farklıhesaba409; aynıhesaba200 idempotent. Raw key ve session aynı kaynaktan reserve eder. Refund/cancel mevcut license satırını değiştirir; account balance anında türetilir.

**Öneri/karar: seçenek2.** Login404 nginx rewrite ile çözülmez; hashlenmiş parolalar, session auth, ownership ve kaynak lisansa atomik dispatch gerekir. .6 authority build-time go-hook olduğu için yalnız URL değiştirerek accounts olmaz; .7 releaseAuthority=accounts, embeddedGatewayURL=https://api-test.kanus.org ile rebuild gerekir; production aynı linker option api.kanus.org olabilir, deployment yapılmaz. US_LICENSE_URL de CLI için Go adresinde kalmalı; admin/dev bypass ticari model yolunu açmamalı.

## E — Seçilen Go hesap tasarımı / uygulanacak sözleşme

- **Tek ledger**: existing licenses/reservations/settlements/webhooks/orders/refunds/outbox aynen kalır. Yeni accounts[emailNormalized] {id,passwordHash,sessionVersion,createdAt,emailVerified,verificationHash/expiry,resetHash/expiry}; licenseOwners[us1]→accountID. Kalan değer account'a kopyalanmaz. Sadece valid/nonrevoked source license Credits-Reserved toplamı; sub prompts ve use seconds AYRI birimler. Legacy JSON yeni alan yokken otomatik boş map ile açılır; mevcut sayılar/key'ler/expiry değişmez. Read-only map defaults ayrı test.
- POST /account/register {email,password,license?}: bcrypt cost12 (UTF-8 byte12..72), unique normalize email; optional existingus1 bind aynı transaction'da; welcome/verification outbox aynıcommit.201 {ok:true}. Kayıt aktif posta erişimi kanıtı sayılmaz; verification ayrı code endpoint'i, SMTP default kapalı olduğu için login'i gizlice “mail gönderildi” diye engellemez.
- POST /account/login {email,password}: constant-cost unknown-user comparison, generic401; HMAC-SHA256 signed usa1.body.sig {accountID,sessionVersion,exp,nonce}; secret yalnız US_SESSION_SECRET >=32byte env, dosyada yok. TTL24h configurable; reset sessionVersion increment bütün eski token'ları geçersizleştirir.200 {session}; token plaintext persist edilmez.
- POST /account/redeem {license}: S; yalnız mevcut us1/active/expiry, reserved varsa409; owner unique; sameaccount repeat200 idempotent, otheraccount409. Credit transfer/clone YOK. POST /account/balance {}: {remaining:availablePrompts,seconds_remaining:availableSeconds,kind:account,expiresAt,...}.
- Model/metadata auth: S veya L. Session queueden sonra tekrar doğrulanır; store mutex içinde uygun kaynak lisans seçilir, aynı reserveModel kaydına realKey yazılır. Önce sub dahaerken expiry, sonra hourly; tekrequest birkaçlicense'a bölünmez. Proxy charge/release seçilen realKey'de kalır; refund/reset yarışları fail-closed. Rawkey hesabın da gördüğü aynıbakiyeyi tüketir, ikinci ledger yok.
- GET /v1/credits/status S aggregate; account balance/CLI saniyeyi prompt'a eklemez. POST /account/requests existing reservations/settlements'tan scoped max20. POST /account/workers güvenli upstream metadata adapter; provider_status unknown/livefalse, gerçek GCP yönetimi yok; upstream yoksa503, sahte ready yok.
- Per-IP register/login/reset limitleri (ör.15dk10login,5register,5reset) + bounded in-memory buckets + toplam bcrypt concurrency gate. Forwarded IP yalnız loopback güvenilir proxyden kabul edilir; HOST forwarded header'ı overwrite etmeli,8787public olmamalı. Restart'ın limit penceresini sıfırladığı belgelenir; nginx dışrate-limit ek savunmadır.
- POST /account/verify-email {email,code}; POST /account/password-reset/request {email}→generic202; POST /account/password-reset/confirm {email,code,password}. Cryptorandom single-use code hash+expiry aynıatomicJSON; reset yeni bcrypt +sessionVersion birliktecommit. Reset/verification kodları yalnızoutbox body'de, response/log'da yok. US_SMTP_ENABLE=0 kalır. US_SMTP_ALLOWLIST set ise listede olmayan alıcı sender'a HİÇ verilmez; skip ayrı terminal queue state, yeni güvenlik testi.
- YeniHOST env: US_SESSION_SECRET placeholder, US_SESSION_TTL_SECONDS=86400, US_SMTP_ALLOWLIST comma-separated exactnormalized emails (optional; set-empty liste denyall olmalı). Nginx **/account/** aynıGo127.0.0.1:8787 upstream'e; mevcutchat/models/credits ve3webhook unchanged. HTTPbodylimit, TLS, authheader ve clientIPoverwrite; hiçbir /account/ pathPython/CloudRun'a yönlenmez.
- --check-account <credentials.json> veya US_ACCOUNT_CREDENTIALS_FILE; {email,password}; gerçek key/parola envinline/log'a konmaz. Login→balance; session yalnızRAM; masked output, nonzeroonfailure. Userenv inline isteyenopsiyon yerine dosya-only güvenli biçim tercih edilir.

Bu tasarım dosyası sonrası yerel implementasyon/test yapılacak; Windows native, gerçek SMTP/tunnel/GPU/production kabulü ayrı BLOCKED kalır. Bu analizin yazılması canlı route'ların hazır olduğu anlamına gelmez.

## F — Round7 yerel implementasyon sonucu — 06.11.2026 08:20 TSİ (Europe/Istanbul)
**PASS — Yukarıdaki A–C başlangıç analizi tarihsel kaldı. Seçenek2 yerel kaynak/test/binary düzeyinde uygulanmıştır; Cloud Run yok, deployment yapılmadı.** A tablosundaki “Go yok/S uymaz” hücreleri artık yerel Go için geçerli değildir. Canlı login404'ün giderildiği henüz iddia edilmez.

Aşağıdaki tüm Go route'ları PASS yerel unit/production-mux testi; auth S=HMAC session. Hata body {error:string}, yanlış method405; eksik/placeholder sessionsecret503. Dosya kökü work2/kiyas/term. Eforlar geliştirme+review planlama bandıdır, bu agent'ın geçen süresi değildir; ortak altyapı nedeniyle toplanmaz. CLI4–6h, model S dispatch6–10h, models1–2h, status2–4h A'dakiyle ortak; toplam tasarım/tüm kabul74–120mühendis-saat (baseline tahmin).

| Method/path | Request→response | Auth | Güncel kaynak | Tahmin |
|---|---|---|---|---|
| POST /account/register | {email,password,license?} →201{ok:true} | yok | licensesrv/accounts.go:355–396 | 8–12h |
| POST /account/login | {email,password,license?} →200{session} | yok | accounts.go:397–427 | 6–10h |
| POST /account/balance | {} →{remaining,seconds_remaining,reserved,reserved_seconds,kind:account,expiresAt} | S | accounts.go:512–556; AccountBalance:153 | 4–6h |
| POST /account/redeem | {license} →aynı balance; same200/other409 | S | accounts.go:531–556 | 6–10h |
| POST /account/requests | {} →{requests:[{request_id,state,created,updated}]} max20 | S | accounts.go:548–552,567–605 | 4–6h |
| POST /account/workers | {} →{workers:[{worker,model,phase,provider_status,active,idle_seconds,idle_remaining,observed_at,live,model_ready}]} | S | accounts.go:522–529; account_workers.go:26 | 4–8h |
| POST /account/verification/request | {email} →generic202{ok:true} | yok | accounts.go:428–461 | 2–3h |
| POST /account/verify-email | {email,code} →200{ok:true} | tek-kullanım code | accounts.go:462–511 | 2–3h |
| POST /account/password-reset/request | {email} →generic202{ok:true} | yok | accounts.go:428–461 | 2–3h |
| POST /account/password-reset/confirm | {email,code,password} →200{ok:true} | tek-kullanım code | accounts.go:462–511 | 4–7h |

- PASS — EXE güncel callsite: account_client.go:13–98 register/login/balance/redeem; :101 requests; worker_client.go:29 workers; gui.go:524–606 account işlemleri. Normal login accounts akışı desktop reserve/settle YAPMAZ (gateway.go:82–97); server aynı source ledger'ı reserve eder.
- PASS — POST /v1/chat ve /v1/chat/completions,GET /v1/models S veya L; GET /v1/credits/status S aggregate veya L eski sözleşme (credits.go:28). Chat source seçimi accounts.go:98–142 + model_meter.go:11–59; settle aynı source key. POST reserve/settle L-only legacy kaldı; EXE .7 normal accounts çağırmaz.
- PASS — HMACsession accounts.go:49–87; bcrypt12/min12max72; accounts+licenseOwners yeni maps; gerçek kredi yalnız mevcut licenses/reservations/settlements. Key bir owner, tekrar aynı account idempotent; partial refund source düşüşü aggregate'a yansır. Eski JSON test: accounts_test.go:177; source race:389; SSE E2E:216; unique owner55; expiry/reset103; SMTP allowlist345.
- PASS — Parity B güncellemesi: Go account/hash/session, S proxy/status/models artık PASS yerel. Diğer Go HMAC/idempotency/refund/locking/queue/hourly/SSE regresyonları full v4 PASS. Python değişmedi/yeniden test edilmedi; C probe tarihsel ve son no-Cloud-Run kararından önceydi; sonrasında Cloud Run çağrısı yok.
- PASS — .7 accounts commercial default https://api-test.kanus.org +deepseek-v4.1-flash; production build-time api.kanus.org. --check-account file/env-filepath, masked balance; Windows GUI subsystem log-first. Eski US_GATEWAY_URL override varsa yalnız Go HTTPS kalmalı, Cloud Run'a gitmemeli.
- PASS — Yerel test çıktısı:304üst+37alt PASS/6skip, full race/vet0; fake SMTP,50lisans yük tekrarı ve gerçek Linux headless version/selftest. Dosya hashleri BUILD_SHA256_20261106_v4.txt.
- BLOCKED — HOST /account/ nginx prefix/env + hook yetkili deploy, native Windows, upstream tunnel/GPU controller. Mail delivery/posta sahipliği ve bağımsız security acceptance NOT TESTED. Reset/verify GUI, persistent rate-limit/logout/refresh, tek çağrıda multi-pack hourly pooling yok. Ayrıntılı sınırlar ROUND7_RAPOR_TR.md.

- NOT TESTED — Aggregate expiresAt en geç source expiry bilgisidir; tek ortak paket süresi değildir, lisansların tekil expiry kuralları geçerlidir. Eski Python hesapları varsa otomatik migration yapılmadı; mevcut Go us1 redeem desteklenir.
