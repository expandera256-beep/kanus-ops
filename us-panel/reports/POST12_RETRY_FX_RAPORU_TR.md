# POST12 — transient retry, kredi idempotency ve USD/TRY FX

Tarih: 08.11.2026 01:09 Europe/Istanbul (TSİ).

## Özet — TAMAMLANMIŞ .12 sonrası özellik paketi DEĞİLDİR

- **A genel: NOT DONE / protokol kararı bekliyor.** Mevcut `/settle` protokolüyle güvenli, sınırlı cleanup retry uygulandı; model/repair/reserve için tam idempotency ve transient retry eklenmedi. Kırmızı kabul testleri çift tahsilatı ve kayıp reserve ACK'inden sonra fazla hold'u doğruladı. Yeni protokol alanı gerekince durma talimatı uygulandı; öneri ektedir.
- **B genel: NOT DONE / Ed25519 anahtar önkoşulu doğrulanamadı.** Bu .12 ağacında lisans imzaları Ed25519 değil HMAC-SHA256; gömülü public key yok. Yeni key eklemeden istenen üretim bağlantısı yapılamaz. Sadece yanlış sabit kur ve doğrulanmamış USD gösterimi kaldırıldı; panel **TL-only**. Endpoint/fetcher/imza/cache uygulanmış gibi sunulmuyor.
- Her iki diff **kısmi**, sürüm veya deploy onayı değildir. Protokol/anahtar belirsizlikleri kapatılmadan tamamlanmış kabul edilmemelidir.

## Başlangıç bütünlüğü ve sınırlar

Kaynak `/workspace/us-panel/work2/kiyas/term`, çalışma kopyası `/workspace/us-panel-dev/post12/term`. Başlangıç kopyası `.12/TREE_MANIFEST.sha256` ile **209/209** doğrulandı; ek/eksik dosya yok. Baseline: `/workspace/us-panel-dev/post12/baseline`.

work2, dist ve canlı servisler değiştirilmedi. Release artifact üretilmedi; Linux/Windows derleme kontrolleri `go build ./...` ile çıktı paketlemeden yapıldı. GPU/VPS/Shopify/TCMB'ye canlı çağrı yapılmadı; agent mesajı gönderilmedi. SMTP kapalı tutuldu. Sır dizinleri okunmadı ve wake token işlenmedi.

## §13 — A kök neden ve test sırası

Kök neden notu koddan önce `/workspace/us-panel-dev/post12/evidence/ROOT_CAUSES_TR.md` olarak yazıldı. Sonra testler yazıldı ve kırmızı çalıştırıldı; yalnız mevcut protokolle güvenli düzeltme uygulandı.

| Yol | İnceleme sonucu / risk | Durum |
|---|---|---|
| EXE `model.go:2051–2065`, `:2166–2323` | Door fallback yalnız istek yazılmadan erişim başarısızlığında var; yazılmış ama cevapsız istekte belirsizlik uyarısı var. Her yeni HTTP çağrısı yeni `X-US-Request-ID` üretir (`:2252–2257`), hook bunu idempotency için okumaz. Kör timeout/reset retry ücretli işi tekrarlayabilir. | İnceleme DONE; model transient retry NOT DONE |
| `repair_loop.go:229–304` | `askFix` tek HTTP denemesi yapar, sabit request id/gate yok. Sonraki gerçek repair turu yeni içerik ve yeni ücretli model işidir. Redirect kapatma korunmuştur. Ayrıca helper payload'ında hook'un zorunlu `max_tokens` alanı yok; hook'a doğrudan bağlanırsa 400 alır (bu turda değiştirilmedi). | İnceleme DONE; retry NOT DONE |
| `model.go:265–339`, `continuation.go`, `long_context.go` | `sendOnce` her model turunda ayrı `creditGate.call` çalıştırır. Length continuation, şema düzeltme, özet küçültme ve gerçek repair farklı model işleri; tüm görevi tek id'ye bağlamak eksik ücretlendirme olur. Sadece aynı request'in transient denemeleri id paylaşmalıdır. | DONE (inceleme) |
| Hook `licensesrv/model_proxy.go:157–205` | Her HTTP isteği `reserveModel` ile yeni randID hold açar. Başarıda settle, failure/cancel'da defer ile release; aynı HTTP id tekrarında geçmiş sonuç/replay ilişkisi yok. | Tam idempotency NOT DONE |
| `licensesrv/model_meter.go:11–87` | Yeni reserve her defa randID; settlement aynı id/tutar ile idempotent. HTTP retry kimliğiyle bağlantı yok. | DONE (inceleme) |
| `licensesrv/credits.go:54–85`, `store.go:232–309` | `/reserve` istemci idempotency anahtarı kabul etmez. ACK kaybolunca caller hold id'sini bilmiyor. `expireLocked` ancak sonraki reserve işleminde çalışır, startup/periyodik orphan süpürmesi yok. `/settle` aynı id/karar ile güvenlidir; zıt commit/release reddedilir. | Tam orphan/idempotency NOT DONE |
| Desktop `gateway.go:100–165`, `:234–268` | Eski tek cleanup denemesi, transient hata veya kayıp başarılı ACK'de belirsiz sonuç/hold bırakıyordu. Yeni `settle`: aynı id ve aynı commit ile en fazla 3 deneme; 100/200 ms backoff; toplam 31 sn bağımsız cleanup context. HTTP 4xx tekrar edilmez. Reserve ve model çağrısı **tekrar edilmez**. Başarısız commit için release'e dönülmez. | DONE (dar düzeltme) |

### Kırmızı kanıtlar

- `A_before.log`: kayıp başarılı settlement ACK → `ok=false, settle=1, credits=1, commits=1`; geçici release 503 → `reserved=1`. **2 FAIL**, düzeltmeden sonra iki test de PASS.
- `A_protocol_repro.log`: aynı `X-US-Request-ID` ile iki chat → `credits=498` (istenen 499); aynı id ile iki reserve → `reserved=2` (istenen 1). **2 FAIL**; bunlar **çözülmedi**.
- Bilerek kırmızı kalan iki kabul testi `//go:build post12repro` etiketiyle ayrı tutuldu. Default test PASS tam idempotency PASS anlamına gelmez. Çalıştırma: `go test -tags post12repro -run '^TestPost12Repro' ./licensesrv`.

Öneri: `term/POST12_A_PROTOCOL_PROPOSAL_TR.md` (A diff içindedir). Chat'teki mevcut başlığın idempotency semantiği, reserve'e istemci request-id alanı eklenmesi, principal+body-hash bağlama, kalıcı pending/terminal/replay kayıtları, orphan sweep ve EXE/repair 3-deneme politikası birlikte onaylanmalı. Bu protokol değişikliği uygulanmadı.

## §13 — B kök neden, güvenli kısım ve eksikler

### Anahtar varsayımı kaynakla uyuşmuyor

- `license.go:103–139`: `hmac.New(sha256.New, []byte(secret))`; `embeddedLicenseSecret` HMAC alanı.
- `licensesrv/store.go:337–350`: `MintKey` HMAC-SHA256 kullanıyor.
- `licensesrv/webhook.go:17–24`: Config'te yalnız string `LicenseSecret` var.
- `cmd/us-license-hook/main.go:30–32`: `US_LICENSE_SECRET` bu Config'e veriliyor.
- Tüm Go kaynaklarında Ed25519 lisans imzalayıcısı veya gömülü Ed25519 public key bulunmadı. `embedded_license.go` legacy secret alanını boş bırakıyor.

HMAC secret Ed25519 private key değildir. Ondan seed türetilmedi, key yaratılmadı, EXE'ye yeni key gömülmedi ve sözleşme HMAC'a çevrilmedi. Devam için gizli key değeri değil, mevcut Ed25519 imzalayıcı/public-key çözümleyicisinin doğru **kaynak sürümü/dosya ve simge konumu** gerekiyor. Böyle bir sürüm yoksa algoritma/key migrasyonu ayrıca kararlaştırılmalı.

### GLOBAL patch08 birleştirme analizi

İncelenen dosya: `/workspace/us-panel-dev/patches/global_i18n_08_fx.patch`. `.12`'de `i18n`, `Localizer`, `PrefStore`, `langPref` altyapısı yok. Patch08 `/v1/fx`, HMAC, EUR+USD ve `kanus.fx/1` bekliyor; yeni Ed25519 `/v1/fx/usd-try` ile mantıksal çakışıyor. Tam patch körlemesine uygulanmadı, paralel formatter yaratılmadı. GLOBAL'in Türkçe `price.summary` **TL-only fallback** metni kullanıldı. Tam formatter/FX logic merge **NOT DONE**. 5 günlük yaş doğrulamasında GLOBAL'in fazladan +24h toleransı yeni şartla uyuşturulmalı.

### Uygulanan fallback

Testler önce yazıldı: `B_before.log` **3 FAIL** (sabit kurla USD metni, legacy USD-only fiyat ve yeni lisansa uydurma USD tutarı). Düzeltmeden sonra `B_after.log` **3 PASS**.

- `licensesrv/pricing.go`: sabit `USDTRYRef` ve buna bağlı `PriceUSD` kaldırıldı; `PriceText` yalnız TL.
- `pricing.go`: GLOBAL'den TL-only paragraf; doğrulanmamış eski USD-only fiyat notu gösterilmez.
- `license.go`: mint işleminde doğrulanmamış USD tutarı üretilmez; mevcut `usd` JSON alanı geriye uyum için 0 kalır. HMAC algoritması ve token biçimi değişmez.
- `cmd/us-pricing-gen/main.go`: sabit `usdtry_ref` ve hesaplanmış USD çıktı kolonu kaldırıldı; TRY fiyatları/planları değişmedi. Bu geliştirici aracının çıktı şemasındaki değişiklik tüketicilerce incelenmeli.
- Çalışma ağacında `50.05`, `50,05`, `USDTRYRef`, `PriceUSD` grep sonucu **0 eşleşme**.

| B maddesi | Durum |
|---|---|
| Sabit kurun tüm kullanım yerlerinden çıkarılması, rate yokken TL-only | DONE |
| GET `/v1/fx/usd-try`, per-IP rate limit, Cache-Control | NOT DONE |
| TCMB ForexSelling today.xml HTTPS+timeout, startup+günlük fetch | NOT DONE |
| Last-good atomic kalıcı cache, başarısız fetch/parse/sanity'de koruma | NOT DONE |
| EXISTING Ed25519 private key ile imza ve gömülü public key ile doğrulama | NOT DONE (anahtar önkoşulu kaynakta yok) |
| Panel fetch/cache, her kullanımda 5 günlük yaş doğrulaması | NOT DONE |
| Signature round-trip, tamper, wrong prefix, >5 gün, 503, last-good ve rate-limit testleri | NOT TESTED (uygulanmış endpoint yok) |

Ayrıntı: `term/POST12_B_FX_BLOCKER_TR.md` (B diff içindedir).

## Test / regresyon sonuçları

| Kontrol | Sonuç | Sayı / açıklama |
|---|---|---|
| Başlangıç manifest | PASS | 209/209; ek/eksik yok |
| Orijinal work2 son manifest | PASS | 209/209, .12 ile aynı |
| gofmt (değişen Go dosyaları) | PASS | 8 dosya, gofmt-changed.log boş |
| Linux go vet ./... | PASS | exit 0 |
| Linux go build ./... | PASS | exit 0; release çıktı paketi yok |
| Windows/amd64 go vet ./... | PASS | exit 0 |
| Windows/amd64 go build ./... | PASS | exit 0; gerçek Windows çalıştırması değil |
| go test -count=1 ./... | PASS | 489 PASS / 0 FAIL / 7 SKIP (alt testler dahil); üst seviye 411 PASS / 7 SKIP |
| go test -race -count=1 ./... | PASS | 489 PASS / 0 FAIL / 7 SKIP (alt testler dahil); üst seviye 411 PASS / 7 SKIP |
| Yeni default post12 testleri | PASS | 6 üst seviye, alt testlerle 8 PASS |
| A protokol kabul reproları | FAIL (açık) | 2/2 beklenen güvenlik kabul testi kırmızı; düzeltme sonrası da tekrarlandı (A_protocol_repro_after.log) |
| S2 security regression | PASS | 36 PASS / 0 FAIL / 0 BLOCKED |
| Python lifecycle (kopya + socket network guard) | PASS | 24/24; fake provider, canlı çağrı yok |
| Diff A → B uygulama + son ağaç karşılaştırması | PASS | fuzz=0, dosyalar birebir aynı |
| FX imza/fetch/rate-limit/yaş kabul testleri | NOT TESTED | Uygulama Ed25519 önkoşulunda durdu |

7 SKIP opt-in/canlı veya platform koşullu mevcut testlerdir. `cmd/us-pricing-gen` paketinde test dosyası olmaması ayrıca package-skip olarak görünür; 7 test-skip sayısına dahil değildir. Test kayıtlarında toplam 3 testli paket PASS, 1 test içermeyen paket vardır.

Loglar: `/workspace/us-panel-dev/post12/evidence/`. Önce-kırmızı logları ile default regresyon ve bilerek kırmızı protokol kabul testleri ayrı tutuldu. Windows sonuçları çapraz derleme/vet'tir; gerçek Windows çalıştırması **NOT TESTED**.

## Çıktılar / uygulama doğrulaması

1. `/workspace/us-panel-dev/patches/post12/A_retry_idempotency.diff`
   SHA256: `22576f8979deae9469b219b32c4de5c121f44e2ccffac4321cde2f8adca316b2`
2. `/workspace/us-panel-dev/patches/post12/B_fx_usd_try.diff`
   SHA256: `4513d98983c1116b445573a259e73051127a880a00c0d743202f2b50d485306d`

İki diff sırasıyla .12 baseline kopyasına `patch --batch --fuzz=0 -p1` ile uygulandı; son ağaç `post12/term` ile dosya bazında aynı (`patch-tree-compare.log` boş). A 4, B 6 dosya değiştirir. Bu kısmi yamalar release artifact değildir.

## API için OpenAPI tarzı sözleşme taslağı — UYGULANMADI

Bu parça API'nin .12 sonrasındaki planına içindir; sunucuda bu route henüz yoktur. API 0003 / apiv1 error-envelope kodu .12'de yok; yakın patch'teki `error`, `code`, `message`, `request_id` additive biçimi esas alınmalıdır.

```yaml
paths:
  /v1/fx/usd-try:
    get:
      summary: TCMB USD döviz satış kuru (bilgi amaçlı)
      security: []
      description: >-
        Ed25519 imzalı yayın. Sunucu bayat yayın tarihini reddetmez.
        Panel rate_date 5 günden eskiyse TL-only gösterir.
        İmzalanan UTF-8 baytlar: "kanus-fx-v1\n" +
        rate + "|" + source + "|" + rate_date + "|" + fetched_at.
        Anahtar mevcut lisans anahtarıdır; yeni anahtar eklenmez.
      responses:
        '200':
          description: Son geçerli kur
          headers:
            Cache-Control:
              schema: {type: string, const: 'public, max-age=3600'}
          content:
            application/json:
              schema:
                type: object
                required: [rate, source, rate_date, fetched_at, kid, sig]
                properties:
                  rate: {type: string, description: '1 USD başına TRY; ondalık string'}
                  source: {type: string, const: TCMB}
                  rate_date: {type: string, format: date, description: 'TCMB yayın tarihi, Europe/Istanbul'}
                  fetched_at: {type: string, format: date-time}
                  kid: {type: string, description: 'Mevcut güvenilen lisans anahtarı kimliği'}
                  sig: {type: string, description: 'Ed25519 imza; taşıma kodlaması API ile kesinleşmeli'}
        '503':
          description: Doğrulanabilir kur yok
          content:
            application/json:
              example:
                error: FX rate unavailable
                code: fx_unavailable
                message: FX rate unavailable
                request_id: opaque-generated-id
        '429':
          description: IP başına sınır; Retry-After ile
```

## Devam için gerekenler

1. A için `/reserve` idempotency alanı ve chat replay/terminal sözleşmesi onayı (yeni alan eklenmedi).
2. B için EXISTING Ed25519 key kaynağının doğru sürümü/simgeleri; böyle bir kaynak yoksa tasarım kararı. Key değerlerini mesajda paylaşmayın.
3. GLOBAL 01..07 altyapısıyla patch08'in yeni Ed25519 sözleşmesine rebase sırası.
4. İki yama da inceleme gerektirir; bu işte başka ajanlara bildirim yapılmadı.
