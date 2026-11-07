# FINDING-07 — Account-login lock DoS ve limiter dayanıklılığı

Tarih: 08.11.2026 01:15 TSİ (Europe/Istanbul). **Owner: API. Önem: MEDIUM. Durum: OPEN / FAIL (§12 rate limiting).** F04 tarihsel IP-only DESIGN GAP ayrı kalır.

## Kaynak / kapsam
API cevabı `/workspace/us-panel-dev/patches/api-r11/fix-grace-overflow-on-11/README.md:47–66` kaynak incelemesiyle karşılaştırıldı. QA .11 limiter dosyalarını `dist/2026.11.06.11/TREE_MANIFEST.sha256` ile birebir doğruladı.
- `licensesrv/login_limiter.go:10–16,30–37,40–76,78–99`: normalized email için process-local failed-login map; varsayılan10/900s sabit pencere,10000 kova. Env limit1–1000/window10–86400.
- `accounts.go:48–52,257–290,344–374,453–484`: IP+email katmanları; account block bcryptten önce; bilinmeyen email dummy bcrypt+fail; başarılı login reset.

## MEDIUM repro / beklenen-gerçek
Saldırgan hiçbir hesap credentialı olmadan hedef email için **10 başarısız >=12 karakterli parola** gönderir (farklı sentetik IPler). 11. istekte **doğru parola da429**, Retry-After var. Aynısı bilinmeyen emaile sayılır; sayılma enumeration farkı yaratmaz ancak kayıtlı kurbanın yeni loginini engeller. Mevcut sessionların tümünün kilitlendiği veya veri erişimi gösterilmedi.
Beklenen güvenlik özelliği: unauthenticated failure spray meşru kullanıcının doğrulanmış girişini tek başına tamamen engellememeli; brute-force koruması da korunmalı.
Gerçek: doğru-password login429, sabit pencere ilk failuredan15dk; kilitliyken blocked çağrısı süreyi uzatmaz, fakat pencere sonrasında saldırı tekrarlanabilir.

```bash
python3 /workspace/us-panel/qa/audit12/repro_limiter_risks.py \
 --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term
# İşlevsel limiter sözleşmesi ayrı PASS:
python3 /workspace/us-panel/qa/audit12/run.py \
 --tree /workspace/us-panel/qa/build/qa-r11-patch-6ij93zmg/term --case rate_limit
```
`qa/audit12/limiter_risks_test.go:TestQAAuditLimiterAccountLockDoS` **bilerek gerçek FAIL**, doğru-password200 beklerken429 gördü; xfail yok. Diğer4 davranış reprosu PASS=mevcut riskin doğrulanması, güvenli olduğu anlamına gelmez. Kanıt `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/r11-limiter-risks-wrapper-reassessed.log` ve onun benzersiz RUN_DIRi. Tamamen local/in-process synthetic; canlı IP/burst/hesap/mail/model yok.

## LOW notları ve bağımsız düzeltme
| Konu | Yerel sonuç / kaynak |
|---|---|
|Volatile/not-shared sayaç|Yeni accountLimiter nesnesi boş, iki instance paylaşmıyor; gerçek servis restartı yapılmadı. login_limiter.go:17–35; FixedWindowAndVolatileState PASS.|
|10000-kova eviction spray|10000 diğer sentetik email insertinden sonra en eski hedef kilidi kaldırıldı; default max10000 değiştirilmedi. login_limiter.go:66–70,84–99; EvictionSpray PASS. Doğrudan yerel data-structure testi; canlı saldırı maliyeti ölçülmedi.|
|IP4096 kapasite fail-closed|4096 aktif bucket fixtureında yeni IP429; accounts.go:271–280; IPCapacityAndRetryAfter PASS. Doldurma live yükü yapılmadı.|
|IP429 Retry-After iddiası **çürütüldü**|API README:59 “yok” diyor; gerçek .11 response **Retry-After:5**. accounts.go:317 withBackpressure ve backpressure.go:29–37 tüm429a default5 ekliyor. **Eksik header findingi açılmadı.** LOW sınırlama:5sn gerçek IP15dk pencere kalan süresi değil. İlk harness yanlış beklenti ayrı korundu.|
|register/reset/verify|Yalnız IP katmanı: register5, reset(request+confirm ortak)5, verify10;15dk. Account limiter login-only. accounts.go:344–360,369–374.|
|redeem|Session gerekir; rate group boş. Local12 authenticated invalid-license isteğinde0×429 ve0 yeni rate bucket; hiçbir lisans/kredi değişmedi. accounts.go:344–360,571+; RedeemHasNoRateBucket PASS.|

Parola<12/>72 veya geçersiz email400 ve account-fail sayılmaz (`accounts.go:363–366`); eski15×400 kanıtı bu limiteri test etmez. Pure .11 `/v1/account` alias içermez (`cmd/us-license-hook/main.go:108`); API additive stack aliasında aynı handler/kova olması ayrı kaynak iddiasıdır, pure .11de endpoint var diye sunulmadı.

## Önerilen fix / §13
API, generic account lock yerine account+source/risk tabanlı throttling, step-up/trusted-session recovery ve meşru kullanıcı için DoS-dirençli politika belirlesin; bcrypt concurrency/brute-force korumasını kaldırmasın. Multi-instance/restart tutarlılığı ve state eviction hedef politikası belirlenmeli; IP kapasite koruması ile yüksek-cardinality DoS dengelenmeli; Retry-After gerçek bekleme semantiğine uymalı; redeem ve diğer auth endpoint bütçeleri gerekçelendirilmeli. QA ürün kodu değiştirmedi. Fix sonrası aynı safety regression PASS olmalı, mevcut10/900s correctness testleri korunmalı.
