# API-0016 — GLOBAL FX sözleşmesi için API inceleme ve fark belgesi

Tek ortak sözleşme: **GLOBAL'in `global-work/FX_CONTRACT_FOR_US_API.md` belgesi**. Bu belge paralel bir sözleşme değildir. GLOBAL belgesini madde madde inceler ve her madde için **KABUL** ya da **DEĞİŞİKLİK ÖNERİSİ** verir. Handler kodu yazılmadı, US uygulayacak.
Tarih: 2026-11-08 01:2x TSİ. GLOBAL belgesi satırları 01:03 sürümüne göredir. Panel yaması satırları 01:17 sürümüne göredir (imzasız + zorunlu pin); M7/M8 tablosundaki eski `patch:4xx` referansları 01:16 öncesi sürüme aittir.

## Demir'in kesin kararları (girdi)
- Kur kaynağı: TCMB **döviz satış (ForexSelling)**.
- Kur **imzasız**. Panel kuru yalnız lisans sunucusundan HTTPS ile alır. EXE'ye yeni anahtar gömülmez.
- Bu nedenle `X-Kanus-FX-Signature` / HMAC(`US_LICENSE_SECRET`) yolu **REDDEDİLİR**. Simetrik sır panelde dururdu (GLOBAL belgesi :18, :31, :41; panel yaması `global_i18n_08_fx.patch:454-468`). `sig` ve `kid` alanları da olmayacak.

## Madde madde inceleme

| # | GLOBAL maddesi (dosya:satır) | Karar | Gerekçe / kanıt |
|---|---|---|---|
| M1 | `GET /v1/fx` + HEAD, kimlik doğrulama yok, gövde ≤4096 (FX_CONTRACT:12) | **KABUL** | Rota çakışması yok, ayrıntı S1'de |
| M2 | `X-Kanus-FX-Signature` HMAC (FX_CONTRACT:18,23) | **DEĞİŞİKLİK: KALDIR** | Demir kararı imzasız. Başlık hiç gönderilmez |
| M3 | Gövde `kanus.fx/1` (FX_CONTRACT:20) | **KABUL + ek alan** | Ayrıntı S2'de. `rate_type:"forex_selling"` eklenir, sayılar number kalır |
| M4 | `rates` = EUR+USD, [1,1000], çapraz [0,8;1,6] (FX_CONTRACT:22) | **KABUL** | Panel aynı kuralları uyguluyor (`global_i18n_08_fx.patch:33,77-79`) |
| M5 | Veri yok/geçersiz/**eski** → 503 `{"error":"fx_unavailable"}` + no-store (FX_CONTRACT:24) | **DEĞİŞİKLİK** | Ayrıntı S4'te. Bayatlık 503 sebebi olmaz, zarf api-0003 ile uyumlu hale gelir |
| M6 | Tazelik today−5…today TSİ, gelecek yok (FX_CONTRACT:26-27) | **KABUL (panel uygular)** | Ayrıntı S6'da. Sunucu bayat kuru reddetmez |
| M7 | Panel: `activeGateway()`+`/v1/fx`, yalnız https, yönlendirme izlenmez, 5 sn, ≤4096 (FX_CONTRACT:30) | **KABUL** | `patch:433` `CheckRedirect→ErrUseLastResponse`, 3xx 200 olmadığı için hata (`patch:499-501`), `patch:486-487` https zorunlu. Ek öneri: yanıt `resp.Request.URL.Host` yapılandırılmış host'a eşit olmalı (savunma derinliği) |
| M8 | Panel imza doğrulama, önbellekte `{body,sig}` (FX_CONTRACT:31-32) | **DEĞİŞİKLİK** | İmza kalkar. Önbellek `{body}` olur, okunurken şema ve tazelik yeniden doğrulanır. Liste aşağıda (GLOBAL için panel değişiklik listesi) |
| M9 | Yenileme max-age 5 dk–6 sa, hata sonrası 5 dk; yuvarlama; ödeme TL (FX_CONTRACT:33-35) | **KABUL** | — |
| M10 | Sunucu verisi: cron `fx_snapshot.py` → `/var/lib/us/fx.json`, handler ağa çıkmaz (FX_CONTRACT:38) | **DEĞİŞİKLİK** | Ayrıntı S5'te. Sunucu kuru kendisi çeker |
| M11 | Güvenlik notu HMAC/Ed25519 (FX_CONTRACT:40-41) | **DEĞİŞİKLİK** | Metin şöyle olmalı: "İmzasız; bütünlük TLS + host doğrulaması; etki yalnız yaklaşık gösterim" |
| M12 | Örnek handler `global_fx_licensesrv_PROPOSAL.patch` | **DEĞİŞİKLİK (referans olarak)** | HMAC (`PROPOSAL:31-36,63`), dosya okuma, bayatta 503 (`:54-57`) ve düz metin hata (`http.Error`, `:40,56`) çıkarılmalı. Hata zarfı `WriteAPIError` ile yazılmalı |

## GLOBAL'in 7 sorusuna yanıt

### S1 — `GET /v1/fx` r11 router/auth ile çakışıyor mu? Kimlik doğrulama zorunlu mu?
**Çakışma yok, kimlik doğrulama zorunlu değil.** Koddan kanıt:
- `/v1` için catch-all ya da global auth middleware yok. Router düz `http.ServeMux` ve tam yollar kullanıyor: `.12 ağacı post12/term/cmd/us-license-hook/main.go:81-83` (`/v1/credits/*`), `:96` (`/v1/models`), `:98-99` (`/v1/chat*`), `:108` (`/account/`), `:109` (`/healthz`). `work2/kiyas/term` (.11) aynı yapıda: `main.go:81-103`.
- r11 apiv1 yalnız `/v1/account/` alt ağacını kaydediyor: `patches/api-r11/0001-…patch:117-122` (`MountAccountAlias`), `main.go` kaydı `:181`. Rebase sonrası: `rebased-on-11/api-r11-additive-on-11+api+uyum-v4.diff:94-95,154`. `WithAPIVersion` yalnız başlık ekliyor, auth yapmıyor (`api-r11/0001:99-116`; api-0005 `licensesrv/api_version.go`, patch `:109`).
- api-0006 `/v1/api-keys`, `/v1/api-keys/`, `/v1/usage` (`api-0006…patch:27-29`). Bunlar `/v1/fx`'i kapsamıyor.
- Auth her handler içinde yapılıyor (`Authorization: Bearer` okuma: `api-0006…patch:310,463,771`; `api-0003…patch:172,239,359`), yani mux düzeyinde değil.
- UYUM v4 yalnız `/account/privacy/*` ekliyor (`uyum-v4/03-handler/0001-handler.patch:54-59`).
- Not: `mux.Handle("/v1/fx", …)` sonda `/` olmadan kaydedilmeli. `/v1/fx/` alt ağacı açılmamalı (önceki `/v1/fx/usd-try` taslağı bırakıldı). Yerleşim: UYUM `privacy.Mount` ve `MountAccountAlias`'tan sonra, `/healthz`'den önce. Böylece KAN.US merge sırasına (.11 → UYUM → API → GLOBAL) uyar.

### S2 — Alanlar, string ondalık, rate_type
- **Öneri: `rates` değerleri `kanus.fx/1`'de JSON number KALSIN.** Panel string'i açıkça reddediyor: `global_i18n_08_fx.patch:81` (`r[0] == '"'` → hata). İsim veya tip değişikliği dayatılmıyor. Kesinlik açısından risk yok: TCMB 4 ondalık yayımlıyor ve panel `big.Rat` ile bölüyor (`patch:~95-110`). Sunucu sayıyı TCMB metnindeki ondalık haliyle, yeniden yuvarlamadan yazmalı (örneğin `json.Number`). String ondalık ancak ileride `kanus.fx/2`'de düşünülebilir.
- **`rate_type: "forex_selling"` EKLENSİN.** Bu ek alan geriye uyumlu: panel bilinmeyen alanları yok sayıyor (`patch:59-64` struct unmarshal).
- **`date` = TCMB yayın tarihi** (Europe/Istanbul, YYYY-MM-DD). Bu alan benim taslağımdaki `rate_date`'in karşılığıdır, isim değiştirilmiyor. `generated_at` = sunucunun TCMB'den çektiği an, RFC3339. Benim taslağımdaki `fetched_at`'in karşılığıdır, `+03:00` ofsetiyle yazılması önerilir.
- TR/EN karışıklığı (yalnız not): `tcmb_tarih`, `bulten_no` Türkçe; diğer alanlar İngilizce. Değiştirilmesi önerilmiyor.
- `source` serbest metin kalabilir. Makine için `rate_type` + `source_url` yeterli.

### S3 — EUR ve USD: **KABUL.**

### S4 — Hata: 503 + no-store, api-0003 zarfı, "eski veride 503" farkı
- api-0003 zarfı `{error, code, message, request_id}` ve `X-Request-Id` başlığıdır (`api-0003…patch:47-52,79-92`). Burada `error` = message, legacy alan.
- **Öneri:** 503 gövdesi `{"error":"fx_unavailable","code":"fx_unavailable","message":"FX rate unavailable","request_id":"…"}`, `Cache-Control: no-store`. `error` alanı GLOBAL'in `{"error":"fx_unavailable"}` beklentisiyle bayt düzeyinde değil ama alan değeri düzeyinde geriye uyumlu. Panel zaten yalnız status ≠ 200 kontrol ediyor (`patch:499-501`). Bu, api-0003'teki `error=message` geleneğinden bilinçli bir sapmadır. FX için `error` alanına kod yazılır, bu OpenAPI'de belgelendi.
- **Fark:** GLOBAL "eski veride 503" diyor. API önerisine göre **sunucu 503'ü yalnız hiç geçerli kur yoksa döner** (ilk açılış ve hiç başarılı çekim yok, ya da kalıcı son-geçerli dosya bozuk). Bayat ama geçerli kur 200 ile sunulur. Gerekçe: TCMB hafta sonu ve bayramlarda yayın yapmaz, 9 günlük bayram tatilinde sunucu 503'e düşerdi. Ayrıca tazelik kararı tek yerde (panel, S6) verilmeli. Panel 5 günden eski kurda yine yalnız TL gösterir. Sonuç kullanıcı için aynı, ama sunucu "hiç veri yok" ile "bayat veri" durumlarını ayırt edebilir hale gelir.
- 405 (GET/HEAD dışı) ve 429 (IP limiti, `Retry-After`) de aynı zarfla döner.

### S5 — Kur kaynağı: sunucu TCMB'yi kendisi çeksin — **KABUL (API önerisi)**
- Sunucu `https://www.tcmb.gov.tr/kurlar/today.xml` adresini kendisi çeker: https, 10 sn zaman aşımı, ≤64 KB gövde, yönlendirmede yalnız aynı host. Açılışta bir kez, sonra hafta içi 16:05 TSİ civarında çeker. Başarısızlıkta 16:05'ten sonra sınırlı aralıklarla yeniden dener (örneğin 15/30/60 dk, gün sonuna kadar). Hafta sonu ve tatilde de denemesi zararsızdır, yeni bülten yoksa son kur korunur.
- Parse hatası, eksik EUR/USD, [1,1000] dışı, çapraz oran [0,8;1,6] dışı, ya da önceki geçerli kurdan daha eski veya gelecek tarih durumunda **son geçerli kur korunur**. Son geçerli kur atomik olarak kalıcılaştırılır (temp+rename, örneğin `/var/lib/us/fx.json`) ve açılışta yüklenir.
- `kanus.org/fx.json` bağımlılığı yok. Tek doğruluk kaynağı bu sunucu ucu olur.
- **fx.json üreticisi (kutuda bulundu):** GLOBAL'in site aracı `global-work/site/release_us_fx2/cron/fx_snapshot.py` (ve eski `release_us_fx/cron/fx_snapshot.py`). README'ye göre KAN.US hPanel'de hafta içi 16:05 cron kuruyor (`release_us_fx2/README_RELEASE.md:10`). Tutarlılık önerisi: site cron'u da TCMB yerine `https://<lisans-sunucusu>/v1/fx` adresinden beslensin. Gövde aynı şema olduğu için site ile panel aynı kuru gösterir. KAN.US/GLOBAL kararı.

### S6 — Tazelik: **KABUL, panel uygular.** `i18n.FXFresh` (`patch:~20-25`) today−5…today, TSİ. Sunucu tazelik nedeniyle reddetmez (S4).

### S7 — SPKI pin: **KABUL (varsayılan pin yok). KARAR GEREKİR** (pin istenirse).
Varsayılan olarak pin yok. **Mevcut panel yaması (01:17) pini zorunlu kılıyor, pin boşsa kur hiç çekilmiyor (`patch:433-438,517-518,590`). Değişiklik önerisi aşağıda.** Bütünlük Go'nun sistem kök CA doğrulamasına, https zorunluluğuna, yapılandırılmış host'a ve yönlendirme reddine dayanır. Pin, Let's Encrypt rotasyonunda paneli kırar. Etkisi yalnız TL'ye düşüş olurdu, yine de gereksiz kırılganlık yaratır. Pin istenirse en az 2 SPKI gerekir (ara CA ya da yaprak + yedek anahtar), değerleri HOST/SRE vermeli. EXE'ye yeni sabit gömüleceği için **Demir onayı gerekir**.

## Panel tarafı kontrolleri (KABUL)
Panel kuru yalnız yapılandırılmış lisans sunucusunun https host'undan alır. http adresi ve yönlendirme reddedilir (`patch:433,486-487`). Aralık ve çapraz oran kontrolü panelde de yapılır (`patch:33,77-79`). 01:17 sürümünde `licenseSecret()` önkoşulu kalkmış. Yerine zorunlu SPKI pin gelmiş (`patch:590`), bu S7 ile çelişiyor.

## GLOBAL için panel değişiklik listesi (`global_i18n_08_fx.patch`, değiştirilmedi, yalnız liste)
**Güncelleme 01:17 TSİ:** GLOBAL yamayı yeniden yazmış. HMAC, `X-Kanus-FX-Signature`, `fxMAC`/`fxVerify`, `licenseSecret()` önkoşulu ve önbellekteki `sig` kaldırılmış. `grep -i 'hmac|licenseSecret|X-Kanus-FX-Signature'` sonucu 0 eşleşme. **Kaldırılacak imza doğrulaması kalmadı.** Yukarıdaki M8 maddesi böylece karşılanmış oldu. Kalan farklar:
1. **ZORUNLU SPKI pin — S7 ile çelişiyor.** `:433-438` `fxPinsSPKI = ""` (build-time ldflags), `:456-468` `fxVerifyPin`, `:475` `VerifyConnection`, `:517-518` ve `:530-531` pin yoksa/tutmazsa hata, `:590` pin yoksa arka plan yenileme hiç başlamaz. Sonuç: pin değeri sağlanmadıkça panel kuru **hiç** çekmez, her zaman yalnız TL gösterir. Öneri: pin opsiyonel olsun. Boş pin = yalnız standart zincir + hostname doğrulaması (`:590`'daki ve `:517`'deki `len(fxPins())>0` şartları kaldırılsın, `fxVerifyPin` boş listede `nil` dönsün). Pin isteniyorsa KARAR GEREKİR (Demir), değerler HOST/SRE'den, en az 2 pin. Testler `:875-897` (`TestFXPinRequiredAndEnforced`) buna göre "pin boşsa çekilir, pin verilmiş ve tutmuyorsa reddedilir" olarak bölünmeli.
2. `:81` — string kur reddi. kanus.fx/1'de number kaldığı için **değişiklik gerekmiyor** (S2).
3. (Opsiyonel) Yanıtın son URL host'u yapılandırılmış host'a eşit olmalı. Yönlendirme zaten izlenmiyor (`:479`), 3xx ise `:533`'te reddediliyor. Yalnız savunma derinliği.
4. `rate_type` alanını göstermek gerekmiyor. Bilinmeyen alan yok sayılıyor.
5. GLOBAL belgesi (`FX_CONTRACT_FOR_US_API.md`, mtime 01:03) hâlâ `:18` HMAC başlığını ve `:9,:23,:31-32,:41` imza ifadelerini içeriyor. Panel yamasıyla eşitlenmeli, 503 maddesi (`:24`) S4'e göre güncellenmeli.

## US için uygulama kontrol listesi (licensesrv, .12 sonrası)
- [ ] `licensesrv/fx.go`: `FXHandler` (GET/HEAD, diğerleri 405 + zarf). İmza yok, `US_LICENSE_SECRET` kullanılmaz.
- [ ] Gövde `kanus.fx/1`: `schema, source, source_url, rate_type:"forex_selling", tcmb_tarih, bulten_no, date, base:"TRY", rates{EUR,USD} (number, TCMB ondalığı), generated_at (RFC3339 +03:00)`. ≤4096 bayt. Önceden serileştirilmiş baytlar sunulur.
- [ ] 200: `Content-Type: application/json`, `Cache-Control: public, max-age=3600`.
- [ ] 503 yalnız hiç geçerli kur yoksa: `WriteAPIError(…, 503, "fx_unavailable", …)` ve `error` alanı = `"fx_unavailable"`, `Cache-Control: no-store`.
- [ ] Bayat kur reddedilmez. Tazelik panelin işi.
- [ ] Fetcher: TCMB today.xml https, 10 sn zaman aşımı, gövde sınırı, açılış + hafta içi ~16:05 TSİ + sınırlı yeniden deneme. `kanus.org` bağımlılığı yok.
- [ ] Doğrulama: EUR+USD var, sonlu, [1,1000], çapraz [0,8;1,6], tarih ≤ bugün (TSİ) ve ≥ son geçerli. Aksi halde son geçerli korunur ve loglanır (kur değeri değil, yalnız neden).
- [ ] Son geçerli kur atomik kalıcı (temp+rename), açılışta yüklenir.
- [ ] IP başına rate limit: `accounts.go:257-268`'deki IP çözümleme (loopback arkasında `X-Real-IP`) yeniden kullanılır. Aşımda 429 + `Retry-After` + zarf.
- [ ] `cmd/us-license-hook/main.go`: `mux.Handle("/v1/fx", WithAPIVersion(fx))`, sonda `/` yok. UYUM ve `MountAccountAlias`'tan sonra, `/healthz`'den önce.
- [ ] Testler: 200 şema, HEAD gövdesiz, 405, 503 (hiç kur yok) zarf+no-store, bayat kur 200, parse/aralık hatasında son geçerli korunur, gelecek tarih reddi, kalıcılık restart, 429, ağsız sahte TCMB sunucusu.
- [ ] OpenAPI: `patches/api-0016-fx-contract/openapi.yaml` (api-0009 + FX). Redocly NOT TESTED.

## Dosyalar
- `openapi-fx.yaml` — yalnız FX parçası (paths + FxSnapshot)
- `openapi.yaml` — api-0009 + FX (tag `fx`, sürüm 2026.11.08.16)
- `lint.txt` — Redocly NOT TESTED (yerel paket yok)

## GLOBAL'e kısa yanıt (taslak)
API → GLOBAL: Senin `FX_CONTRACT_FOR_US_API.md` tek sözleşme. İnceleme: `patches/api-0016-fx-contract/FX_CONTRACT_UNIFIED_TR.md`.
(1) `/v1/fx` hiçbir rotayla çakışmıyor. `/v1` catch-all ya da mux auth yok, auth handler içinde. Kimlik doğrulama yok. Rota `"/v1/fx"` (sonda `/` yok).
(2) Rates number kalsın (panelin `:81` string reddi bozulmasın). `rate_type:"forex_selling"` ek alan olarak eklensin. İsimler aynı kalıyor.
(3) EUR+USD KABUL.
(4) 503 + no-store KABUL, gövde api-0003 zarfı: `error:"fx_unavailable", code, message, request_id`. 503 yalnız hiç geçerli kur yoksa dönsün, bayat kur 200 ile sunulsun.
(5) Sunucu TCMB today.xml'i kendisi çeksin (hafta içi ~16:05 TSİ + yeniden deneme, son geçerli kuru korusun). Site `fx_snapshot.py` de `/v1/fx`'ten beslenmeli.
(6) Tazelik KABUL, panel uygular.
(7) 01:17 yamanda pin ZORUNLU: pin boşsa kur hiç çekilmiyor (`:517`, `:590`). Öneri: pin opsiyonel olsun, boşsa standart TLS + host. Pin istenirse en az 2 pin, HOST/SRE değerleriyle, Demir onayıyla (KARAR GEREKİR).
HMAC kalıntısı panel yamasında yok. Belgendeki `:18/:23/:31-32/:41` imza satırları ve `:24` 503 maddesi güncellenmeli.
