# Accounts modu — bağımsız güvenlik QA

> **08.11 güncellemesi:** Aşağıdaki tarihsel sonuçlara ek olarak dosya sonundaki restore/limiter bölümüne bakın: IP-only koruma **DESIGN GAP (FAIL değil)**; güncel kaynakta ek account limiter mevcut ve yerel testleri PASS, .11 canlı kabulü bekliyor.

İlk rapor: 06.11.2026 09:24:04 TSİ; son güncelleme: 06.11.2026 09:35:09 TSİ (Europe/Istanbul). İlk gerçek koşum: **06.11.2026 09:14:37–09:21:31 TSİ**. Hedef yalnız `https://api-test.kanus.org`.

**Güncel26 senaryo: PASS20 / FAIL0 / BLOCKED0 / NOT TESTED6.** (Fresh ilk-claim engeli giderildi; eski19/0/1/6 tarihsel toplamdır.) Doğrulanmış yeni güvenlik açığı yok. Bu önceki §8/S12 toplamından ayrı accounts koşumudur. PASS yalnız aşağıdaki dar kapsamdadır; tamamlanmış penetration test/üretim onayı değildir.

## Ön okuma / sözleşme

- `reports/ACCOUNTS_MODE_GAP_20261106.md` bütünü okundu; A–C tarihsel eksikliklerin F bölümünde yerel implementasyonla superseded olduğu ayrıldı.
- `reports/ACCOUNTS_DEPLOY_TEST_20261106.md` mevcut raporu okundu; bu QA bölümü dosya sonuna **eklenir**, önceki içerik değiştirilmez.
- Kaynak `work2/kiyas/term/licensesrv/accounts.go`: session/HMAC/expiry49–87; source lisans sahipliği201–214; IP limiter250–283; parse/limiter/validation294–353; login397–427; authenticated balance/redeem/requests route sözleşmeleri. Kaynak salt okunur incelendi; deploy binary ile bu ağacın birebir eşliği bu koşumda yeniden hashlenmedi.
- Mevcut geçerli `acct-test-license.txt` (sentetik sipariş9000000000303), daha önce hesap1e bağlıydı. Yeni sipariş/webhook üretilmedi. İmza/session/webhook sırları okunmadı veya kullanılmadı.

## İleri gönderilebilir kısa tablo

| Kontrol | Durum | Gerçek sonuç / önemli sınır |
|---|---|---|
| İki QA hesabı login/balance | PASS | Login200;31.362s/31.998s. Hesap1=10 kredi, hesap2=0 |
| Session imza bozma | PASS |401; aynı genel hata gövdesi |
| Geçmiş-exp claim, geçersiz imza | PASS |401; **geçerli imzalı doğal expiry kabulü değildir** |
| Gerçek session expiry | NOT TESTED | Yaklaşık24saat TTL; kısa TTL issuance API yok; VPS TTL değişmedi |
| Cross-account subject/balance/ownership | PASS | A subject→B sahteciliği401; A bearer+B email yine A bakiyesi10; B session Aya bağlı lisansı redeem409 |
| Redeem aynı hesap iki kez | PASS |200/200;10 kredi değişmedi |
| Redeem aynı hesap2-paralel | PASS |200/200;10 kredi değişmedi; **önceden sahiplenilmiş lisans** |
| Fresh/unbound ilk-claim yarışı | PASS | Tek2-paralel burst: hesap1=409, hesap2=200; account1 10→10, account2 0→10. Kaynak10→10; tek owner, tüketim0 |
| Revoked A/B redeem | PASS |409/409; yeni kredi/ownership yok |
| Login enumeration body/status | PASS | Kayıtlı/kayıtsız:401,37byte, aynı SHA256 (iki örnek/sınıf) |
| Login enumeration timing | NOT TESTED | Medyan32.451s/34.697s; n=2/sınıf, yoğun yük nedeniyle fark güvenilir sonuç değildir |
| Register/reset enumeration | NOT TESTED | E-posta oluşturabilecek endpointler hiç çağrılmadı |
| QA yanlış parola | PASS | Dört yanlış QA denemesi401; yalnız belirtilen QA hesabında |
| Hesap-bazlı brute-force lockout | NOT TESTED | Dört yanlış sonrası doğru parola200; tanımlı eşik yok, kaynak IP throttling gösteriyor. Küçük testten lockout yok sonucu çıkarılmaz |
| IP rate-limit429 | NOT TESTED | İlk15 login yanında izinli takipte iki5-paralel baseline burst:400×5 /400×5;429 yok. Koruma **kanıtlanmadı**; daha fazla burst yapılmadı |
| XFF değiştirerek limiter bypass | NOT TESTED | İzinli5-paralel XFF+X-Real-IP spoof burst:400×5; baseline429 yok. Bypass var/yok sonucu çıkarılamaz |
| Kredi/gizlilik | PASS | İlk tur10→10/0→0; fresh tur10→10/0→10 sahiplik artışı tek kez. Kaynak10→10; consumed0/0; leak-check PASS |

Bu tablo gereksinim düzeyinde grupludur;26 ham senaryo sayısı bootstrap/account bazlı alt kontrolleri de içerir. Makine-okunur güncel toplam `accounts-security-final-results.json` içindedir.

## Kanıtlar ve yeniden çalıştırma

Ham dosya kökü **L = `/workspace/us-panel/reports/qa-20261106/`**.

| Kanıt | Alıntı / gerçek satırlar |
|---|---|
| L/accounts-security.log:1–13 | `valid_login_1 PASS`31362.373ms; `valid_login_2 PASS`31998.250ms; masked session_length188; remaining10/0; expiry remaining86335s |
| :14–22 | signature tamper401; expired-claims tamper401; cross-subject tamper401 |
| :23–30 | A token+B email balance200 döner ve A10 korunur; A history karşılaştırması200/200 |
| :33–45 | Aynı owner repeat200/200,2-paralel200/200; diğer hesap409 |
| :46 | Fresh first-claim race BLOCKED; eldeki lisans zaten owner1e bağlı |
| :47–58 | Revoked lisans A/B409; before/after hesap bakiyeleri10/10 ve0/0; tüketim0 |
| :59–69 | Registered/unregistered dört401; body37byte; SHA256 `6832c24dc64d8e6926835a9c50921423f40ac9ec4fc502d83a9ce741649b3648` aynı |
| :70–78 | Ek yanlış QA parolaları401 (46.618s/46.474s); dört toplam yanlış sonrası doğru login200 (38.311s) |
| :79–88 | Üç kısa-invalid-parola probe400/400/400;429 yok; ana koşum19 PASS/0 FAIL/1 BLOCKED/6 NOT TESTED; leak-check PASS |
| L/accounts-xff.log:1–5 | Son3 ucuz probe: kontrol400, XFF400/400;37byte ve aynı hash; baseline429 yok |
| L/accounts-security-events.json | Her istek seq/path/http/ms/TSİtimestamp; **gövde/Authorization/session/password yok** |
| L/accounts-security-results.json | İlk ana koşumun değiştirilmeyen ham senaryo sonuçları |
| L/accounts-security-final-results.json | Güncel26 gereksinim senaryosu20/0/0/6; ilk-claim ve limiter/XFF kanıtları yerinde güncellendi, bootstrap/repeat alt kontroller iki kere sayılmadı |
| L/accounts-credential-leak-check.json | Runtime bilinen parolalar, verilen lisanslar ve bütün üretilen sessionlarla tarama; credential_hit_files=[] PASS |

Gerçek komutlar:
```bash
python3 -m py_compile /workspace/us-panel/qa/accounts/run.py
python3 /workspace/us-panel/qa/accounts/run.py > /workspace/us-panel/reports/qa-20261106/accounts-security.log 2>&1
python3 /workspace/us-panel/qa/accounts/xff_probe.py > /workspace/us-panel/reports/qa-20261106/accounts-xff.log 2>&1
```
Tekrar çalıştırma rehberi: `qa/accounts/README.md`. **İlk turda ek burst yapılmadı; aşağıdaki açık izinli takipte toplam üç5li burst yapıldı. Başka burst veya fresh yarışını yeniden koşmayın.**

## İlk koşumun tarihsel trafik / güvenlik sınırları (09:14–09:21 TSİ)

- Toplam33 HTTP isteği; bunların15i login. Maksimum eşzamanlılık2 yalnız idempotent redeem yarışı; bcrypt loginler **sıralı** koştu.
- Socket timeout120s; ilk ağ hatası/5xx durumunda tüm sonraki istekleri durduracak fail-closed harness. Bu tur timeout veya5xx olmadı. Deneme sayıları büyütülmedi; son3 XFF isteği baştan belirlenen15-login tavanı içinde kaldı.
- İlk12-login ana koşumu sonrası3 kısa-invalid-parola XFF kontrolü yapıldı; bunlar bcrypti tetiklemeden validation400 verdi.15 sınırından sonra duruldu.
- Login/reset/verification/register dışındaki hesap işlemleri için yalnız balance/requests/aynı-owner redeem ve revoked redeem kullanıldı. **Register, password-reset, verification, model/workers, gerçek e-posta, sipariş, admin, SSH/VPS dosya/deploy/restart işlemleri yok.** Hesap/parola değişmedi.
- Son balance ölçümü redeem fazının sonunda09:16:00–09:16:01 TSİ: hesap1 remaining10,reserved0; hesap2 remaining0,reserved0. Daha sonraki istekler yalnız login/invalid token/validation türündedir; model/credit dispatch yok, bu QA kaynaklı tüketim0/0.
- Runtime dosyaları yalnız belirtilen QA credential/lisans dosyaları; tokenlar sadeceRAMda, disk veya parent çıktısına kaydedilmedi. Kaynak rapordaki session-secret dosyasına erişilmedi.

## Session expiry / account isolation sınırları

Sessionlar gerçek login ile geldi. İmzalı payload metadata içinden kalan süre86335s gözlendi. Gerçek sonlanmalar07.11.2026 **09:14:37 ve09:15:09 TSİ**. Bu kadar beklenmedi. Payload exp geçmişe taşınıp eski imza bırakılarak401 elde edildi; bu sadece claims/imza tamper kontrolüdür. Doğal expiryyi kanıtlamak için imzası geçerli, kısa ömürlü QA sessionı veya düşükTTLli izole test ortamı gerekir; sır okumak/yeniden imzalamak veya yoğun VPS yapılandırmasını değiştirmek yapılmadı.

Cross-account pozitif kontrolünde hesap bakiyeleri farklıydı (10 vs0), dolayısıyla email alanı üzerinden hedef hesap değişseydi anlaşılırdı. A tokenı+B emaili A bakiyesini verdi; B tokenının A lisansına redeem girişimi409. Subject claiminin B hesabına taşınması401. History yanıtları boştu; iki boş yanıtın eşliği **dolu başka kullanıcı resource/history fixtureı üzerinde tam IDOR kabulü değildir**. Uzak dosya/proje endpointleri bu APIye ait değil; bunlara yeni hedef uydurulmadı.

## Timing ve limiter yorumları — aşırı iddia yok

Yanlış parola timingleri:
- Kayıtlı QA:33.101s ve31.802s (medyan32.451s).
- Kayıtsız:34.798s ve34.597s (medyan34.697s).
- Daha sonraki yanlış QA loginleri46.618s ve46.474s; doğru login38.311s oldu. Bu değişim yüksek yük/gürültünün ciddi olduğunu gösterir.2 örnek/sınıf yaklaşık2.246s farkı side-channel kanıtı değildir; aynı şekilde timing güvenliğini de kanıtlamaz. Ağ ve sunucu scheduling/KDF süreleri ayrıştırılmadı.

Kaynak `accounts.go:250–283` limiteri group+IP,15dk pencere, login limiti10 olarak kuruyor (`:321–337`). Login body validationı `:341–344`ten önce limiter var. **Bu koşumda15 login yapılmasına rağmen429 görülmemesi dikkat gerektirir**, fakat VPSnin gördüğü gerçek kaynak-IP bucketları, olası proxy egress değişimleri, eşzamanlı deploy/restart veya çalışan binary/konfigürasyon bu worker tarafından doğrulanmadı. Bu sebeple rate-limitin çalıştığına PASS verilmedi; kesin bozukluğa FAIL de verilmedi. Aynı-IP/aynı-bucket kontrollü yeniden test veya yetkili operatörün salt-okunur limiter/proxy log kanıtı gerekir.

XFF iki RFC-documentation IP değeriyle değiştirildi; tümü400 ve aynı genel hata. Reddedilmiş429 baselineı yokken bunun bypass kanıtı olamayacağı açıkça bırakıldı. Kaynak XFFyi okumuyor; loopbackten gelen `X-Real-IP`ye güveniyor (`:255–258`). Bu yüzden proxy header overwrite kabulü ayrıca doğrulanmalı, bu source bilgisi canlı config kanıtı değildir.

Dört yanlış QA girişinden sonra doğru parola200 döndü. Kaynakta hesap-bazlı failure counter/lock alanı yok (`Account :24–35`, rate key `:261` IP bazlı); bu tasarım riski dağıtık brute-force açısından incelenmeli. Ancak bu küçük canlı test belirlenmemiş bir account-lock eşik politikasını tüketmedi; hesap-bazlı lockout kabulü NOT TESTEDdir. IP throttling ile hesap lockout aynı şey değildir.

## Engeller / gerekenler

1. **GİDERİLDİ — fresh first-claim:** Sağlanan dosyayla canlı status kontrolü sonrası tek2-paralel yarış PASS. Hesap2 owner oldu; hesap1 tekrar409. Ana tabloda artık BLOCKED yok. Anahtar artık unbound değildir; tekrar first-claim testi için kullanılmamalı. Revoke yapılmadı; koordinatör sonradan revoke edebilir.
2. **NOT TESTED — gerçek expiry:** Geçerli imzalı kısaTTL QA tokenı / izole düşükTTL test ortamı veya mevcut gerçek sessionın doğal sonlanmasından sonra manuel test. Prod/VPS TTL değişikliği bu QA görevinin parçası değildir.
3. **NOT TESTED — rate/XFF/lockout:** KaynakIP bucketı ve uygulanacak hesap lockout politikası netleştirilmeli; yük düşükken kontrollü küçük test veya operatörün salt-okunur proxy/limiter logları gerekli. Daha fazla yük üretmedik.
4. **NOT TESTED — mail-capable enumeration/timing:** Mail kapalı/deny-all izole fixture ve daha sakin, kontrollü timing ölçümü gerekir. Gerçek adreslere mail tetiklenmedi.

**FAIL yok; yeni güvenlik bugı/başarısızlık için dosya-satır iddiası uydurulmadı.** Dikkat gerektiren canlı gözlem429un görülmemesi ve account-lock politikasının belirsizliği; bunlar yukarıdaki sınırlamalarla aktarılır.


## İzinli takip — burst/XFF ve fresh ilk-claim (06.11.2026 09:34:49 TSİ (Europe/Istanbul))

Bu bölüm önceki turdaki engel/ölçümleri günceller; geçmiş ham loglar korunmuştur. **Güncel26 senaryo:20 PASS /0 FAIL /0 BLOCKED /6 NOT TESTED.** Sekiz fresh alt kontrol ayrı dosyada8 PASSdır;26 senaryoya tekrar eklenmez. Üç burst dosyasının her biri1 PASS (invalid parolayla session üretilmedi),1 NOT TESTED (limiter/bypass kanıtlanmadı) içerir; bunlar da aynı gereksinimlerin tekrarlarıdır.

### Nginx429 hipotezi — kontrollü15 ek istek, sonra duruldu

| Burst (06.11.2026 TSİ) | Paralel / sonuç | Cooldown | Yorum |
|---|---|---|---|
| Baseline09:27:45–09:27:46 |5 /400×5,429=0 | Önceki turdan sonra | Client start spread2.070ms |
| Tek izinli baseline tekrarı09:29:08–09:29:10 |5 /400×5,429=0 |81.500s | Client start spread0.921ms |
| XFF+X-Real-IP spoof09:30:34–09:30:35 |5 /400×5,429=0 |83.134s | Client start spread1.111ms; her istekte farklı TEST-NET IPv4 |

Kanıtlar: `accounts-burst-baseline.log`, `accounts-burst-baseline-repeat.log`, `accounts-burst-spoof.log` ve aynı adlı `-results.json` dosyaları (L altında). İstekler ucuz **kısa-invalid parola** kullanır;400 body validation sonucu,401 yanlış-geçerli-uzunluk-parola sonucu değildir. Bu üç burst bcryptli brute-force/lockout testi sayılmaz. Nginx body-agnostic frontend limit hipotezini hedefledik; gerçekte limit zone/location ve sunucunun gördüğü aynıIP bucketı bağımsız doğrulanmadı. **Validation400ün tek başına nginx sayacını bypass ettiği iddia edilmiyor.** Baseline429 üretilmediği için XFF bypassına PASS/FAIL verilmedi.5xx/timeout yok; daha fazla/escalating burst yok.

### Fresh dosya — sipariş9000000000404 etiketi,10 kredi

- Yalnız koordinatörün `/workspace/secrets/acct-qa-fresh-license.txt` dosyası runtime okundu. Sipariş numarası koordinatör metadata etiketidir; status API order ID yayımlamaz.
- **Yarış öncesi09:31:50 TSİ:** canlı `GET /v1/credits/status`200; harness `valid=true`, `kind=sub`, remaining10,reserved0 doğrulamadan ilerlemez. Kaynak `credits.go:132` sözleşmesinde `valid = !revoked && now.Before(expiresAt)`; ham kayıtta `revoked` alanı ayrıca tutulmadı. Dolayısıyla önkoşul salt koordinatör iddiasına dayanmadı.
- Gerçek loginler sıralı200:45.105s ve51.111s. Farklı session subjectleri doğrulandı; tokenlar sadece RAMda. Yarış öncesi bakiyeler hesap1=10,hesap2=0.
- **09:33:29–09:33:30 TSİ tek2-paralel burst:** hesap1 **409**, hesap2 **200**. Yarış sonrası bakiyeler10 ve10; delta **[0,+10]**. Tek owner/tek quota görünürlüğü; çift grant yok.
- Yarıştan sonra sıralı winner-repeat200,loser-repeat409; son bakiyeler yine10/10,reserved0. Kaynak lisans10→10,reserved0. Mevcut source kredilerinin ownership bağlantısı kuruldu; yeni mint veya model harcaması yok, **tüketim0**.
- Fresh tur **14 HTTP istek**, bunlardan2 sıralı login; yalnız iki ilk redeem paraleldir. Koşum09:31:50–09:33:38 TSİ. Önceki spoof bitişiyle ilk status arasında74s+ cooldown. Tüm tur toplamı62 HTTP/32 login; default max2, yalnız açık izinli üç login burstünde max5. Network error/5xx/retry yok.
- Kanıt: `accounts-fresh-claim.log:1–23`, `accounts-fresh-claim-results.json`, `accounts-fresh-credential-leak-check.json` (**PASS; hit=[]**). Ham fresh sonuç **8 PASS /0 FAIL /0 BLOCKED /0 NOT TESTED**.
- **Anahtar artık hesap2ye bağlı. Yarış tekrarlanmadı ve tekrar edilmeyecek.** Koordinatör sonradan revoke edebilir; bu worker revoke etmedi.

Gerçek komut (bir kez çalıştı; tekrar koşmayın):
```bash
python3 /workspace/us-panel/qa/accounts/fresh_claim.py > /workspace/us-panel/reports/qa-20261106/accounts-fresh-claim.log 2>&1
```

Kalan6 NOT TESTED: doğal geçerli-imzalı session expiry; enumeration timing; register/reset enumeration; hesap-bazlı lockout; IP429 enforcement; XFF bypass kabulü. Yeni doğrulanmış FAIL yok. Önceki durumu koruyan snapshot: `accounts-security-pre-fresh-results.json`; güncel özet: `accounts-security-final-results.json`.


## Restore / limiter kanıt güncellemesi — 08.11.2026 00:13 TSİ (Europe/Istanbul)

**Bu tur uzak ağ/DNS/login/yük testi yapılmadı.** Aşağıdaki nginx bilgisi HOSTun daha önce aldığı salt-okunur kanıttan; yeni canlı konfigürasyon doğrulaması değildir.

### DESIGN GAP — IP-only koruma, FAIL değil

`reports/qa-20261106/ratelimit-evidence.txt:8–36,43–78` aktif nginx `uslic_account` limitinin **5r/m,burst=2,nodelay**, anahtarının **$binary_remote_addr** olduğunu gösterir. `/account/login` exact location gerçekten limitlidir; location ordering bugı yok. X-Real-IP upstreamte `$remote_addr` ile overwrite edilir; XFF gerçek kaynak-IP değildir.

Aynı dosya`:89–110,136–149`: **06.11.2026 09:26–09:31 TSİ** penceresinde16 login satırı:15×400,1×200,0×429. QA15 invalid probe çoğunlukla farklı **140.248.50.x gerçek bağlantı adreslerinden** geldi; tek200 başka IP/Go clientıydı. Bu pencere tüm QA aşamalarının32 login toplamı değildir. Kaynak IPler değiştiğinden aynıIP bucketı dolmadı; bu, önceki429 gözlenmemesini açıklar. **IP tabanlı sınır dağıtık/rotating-IP giriş denemelerini aynı hesapta biriktirmez: DESIGN GAP; uygulanmayan nginx limiter/işlevsel FAIL olarak sınıflandırılmaz.** XFF spoofing bypassı bu kanıtla gösterilmiş değildir; gerçek source-IP değişimi ile header spoof farklıdır. SabitIP enforcementının bağımsız canlı kabulü hâlâ NOT TESTED.

### Kaynak bulgusu — eski IP limiteri ile yeni account limiteri AYRI

Güncel restore edilmiş ağaçta `work2/kiyas/term/licensesrv/accounts.go:257–290` eski Go15dakika limiteridir: RemoteAddr`:258–260`, yalnız loopback proxyden X-Real-IP güveni`:262–265`, **key=group+":"+ip `:268`**,15dakika pencere`:272`, limit10/login grubu`:344–351`, çağrı/reddetme`:358–360`. **Bu limiter hesap değil IP bazlıdır.** Email body bu anahtara katılmaz.

Ancak güncel kaynakta **ek hesap-bazlı koruma da var**; eski rapordaki “account counter yok” cümlesi bu yeni kaynak için geçerli değildir:
- `licensesrv/login_limiter.go:10–16,30–35,40–56,61–80`: normalized e-mail başına başarısız giriş sayacı; varsayılan10/900s; başarılı giriş reset; pencere dolunca açılır.
- `accounts.go:362–373`: normalize email sonrası account lock/429+Retry-After;`:466–468` yanlış parola sayılır;`:484` başarılı girişte reset.
- Invalid kısa parola`:364–367`de400 alır, account fail countera ulaşmaz. Eski15×400 probe yeni account-limiter testi DEĞİLDİR.
- Mevcut `TestAccountLimiterRotatingIPs` yerelde PASS: kayıtlı/kayıtsız aynı hesabın farklı IPlerden ilk10 yanlış girişi401,11.429; casing/whitespace normalize, başka hesap etkilenmez. Reset/window ve memory-bound testleri de PASS. Kanıt `reports/qa-20261108/account-limiter-local.log:1–15`; test kaynağı `login_limiter_test.go:25–61,64–99`.

**Bu yeni source/local PASS, eski HOST binarysinin veya henüz gelmemiş .11 deployunun account-limiter içerdiğine kanıt değildir.** Eski HOST kanıtındaki hash `7d72f449…`; yeni .11 çalışan binary/hash doğrulaması HOSTtan beklenecek. IP-only design gapı giderecek kaynak değişikliği mevcut; canlı giderim .11 sonrası kabul edilmelidir.

### Restore sonrası yerel harness sağlığı

- `qa/build/` yeniden oluşturuldu; `qa/common.py` hook build parentını, `qa/sec8/run.py` log/build klasörlerini artık kendisi oluşturur. Restore öncesi /tmp veya dist çıktısına bağımlılık yok. Ürün kaynağı değiştirilmedi.
- §8 yerel: **21 PASS /0 FAIL /0 BLOCKED /1 NOT TESTED**; gerçek loopback hook build, hesap/credit/isolation, mock SSE1/5/10/25/50 ve panel overlay sınırları. Kanıt `reports/qa-20261108/sec8-local.log`, `hook-build.log`, `sec8-panel-boundaries.log`.
- §9 mock: **3 PASS /0 FAIL /0 BLOCKED /2 NOT TESTED**; gerçek panel generate/build/failing-test/repair,3 mock upstream çağrısı, bağımsız HTTP kabul. Kanıt `reports/qa-20261108/e2e9-mock.log`, `e2e9-panel-driver.log`, `e2e9-independent-acceptance.log`.
- Üç account limiter yerel testi ayrıca PASS. Bu yeni dar yerel koşumlar, tarihsel uzaktaki20/0/0/6 accounts veya64/0/0/7 §8–§12 toplamlarına karıştırılmadı.
- `qa-license-a/b` dosyalarının kayıp olması bu local/mock koşumları etkilemez; uzakta revoked regresyonu için yeni coordinator-provided revoked fixture gerekir. `dist/2026.11.06.11/` henüz yok; gerçek release/Windows kabulü yapılmadı.

Plan: `reports/QA_PLAN_R11_20261108.md`. Regresyon: `python3 /workspace/us-panel/qa/regression/run.py --dry-run`. **Yalnız dry-run doğrulandı**; DNS/network/credential-read/subprocess guardları altında hem explicit hem default dry-run PASS. Kanıt `reports/qa-20261108/r11-dry-run.json`, `r11-dry-run-validation.json`. GPU/gerçek model .11 deploy sonrasına ertelendi.
