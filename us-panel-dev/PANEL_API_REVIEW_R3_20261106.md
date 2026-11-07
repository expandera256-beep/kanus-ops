# PANEL API İnceleme Tur 3 — 2026-11-06

**İnceleyen:** PANEL DEV (executor)  
**Zaman:** 2026-11-06 10:09 TSİ (Europe/Istanbul)  
**Kaynak:** `/workspace/us-panel/reports/KANUS_API_REPORT.md`  
**Önceki:** `PANEL_API_REVIEW_R2_20261106.md`  
**Kanıt:** `/workspace/us-panel-dev/panel-evidence/api-review/round3/`  
**S2:** `/workspace/us-panel-dev/panel-evidence/review/regression/results/api_r3_1005/`  

> Kaynak/pinned’e yazılmadı; mesaj yok; deploy/EXE yok. Çalışma yalnız `/tmp/api-review-r3/`.

---

## Taban / .11 durumu

| Öğe | Sonuç |
|---|---|
| pinned/v10 manifest | **181/181**, DIFF=0 |
| PANEL’e teslim pinned **v11** | **YOK** |
| `work2/kiyas/term` | `buildVersion=2026.11.06.11` (canlı US ağacı; resmi pin değil) |
| `work/api-r11-term` | `.11` + önceden uygulanmış UYUM/API çalışma kopyası |
| `r11_base` | hâlâ `.10` |

**api-r11 tabanı (tespit):** README + apply doğrulaması → **term `.11` + `uyum-v2` 01→04** üzerine additive `apiv1` + `main.go` 3 satır.  
**Nihai PASS:** `.11` PANEL’e pin/manifest ile teslim edilmeden **verilemez**; aşağıdaki v10 sonuçları ara doğrulama, r11 yığını ara/aday doğrulama.

---

## Apply sırası (önemli bulgu)

API raporundaki sıra `0006→0006b→0008→0004` **pinned v10’da FAIL**.  
Çalışan sıra (**Order B**): `0001→0003→0005→0006→0008→0006b→0004/0012` — hepsi check/apply **0**.

| Hedef | Sonuç |
|---|---|
| pinned v10 + Order B | **temiz** |
| r11_base (.10) | API iddiasıyla uyumlu (bu turda Order B ile doğrulandı) |
| term (.11) + eski monolitik 0003/0006 | bilinçli FAIL (API de yazıyor) |
| term `.11` + uyum-v2 01→04 + **api-r11** | **check/apply 0** |
| pinned v10 + uyum-v2 + api-r11 | **check/apply 0** (ara; nihai değil) |

`api-0004` ve `api-0012` patch dosyaları **byte-aynı** (`SAME_PATCH`).  
**api-0007** klasörü yok (0006’ya birleşmiş iddiası — ayrı yama yok, kabul).

---

## Yama başına sonuç

| Yama | Sonuç | Not |
|---|---|---|
| api-0002 / 0009 OpenAPI | **PASS** | R2 ORTA/DÜŞÜK kapalı; alias’lar genişlemiş; Redocly R2’de 0 (bu turda yeniden lint yapılmadı → NOT TESTED lint) |
| api-0003 | **PASS** | (önceki tur; bu yığında apply OK) |
| api-0005 | **PASS** (v10 yolu) | `.11` için yerine **api-r11** |
| api-0006 + **0006b** | **PASS** | `revoke_old` / `grace_seconds=0` → eski anahtar **hemen 401** (PoC) |
| api-0008 | **PASS** | |
| **api-0004** | **BLOCKED** | `licensesrv/store.go` değiştiriyor; KAN.US “licensesrv/** dokunma” + US onayı şart |
| **api-0012** | **BLOCKED** | aynı dosya + **davranış değişikliği**; US taşıması şart; teknik PoC’ler yeşil |
| api-0010 SDK Py | **PASS** | R2 YÜKSEK/ORTA kapandı; pytest **11/11**; mutasyon FAIL |
| api-0011 SDK TS | **PASS** | tsc PASS; node **9/9**; mutasyon FAIL |
| **api-r11** | **PASS** (ara) | licensesrv/** / model_proxy.go **dokunulmadı**; `apiv1`+hook; race apiv1+hook PASS. Nihai: `.11` pin sonrası |

---

## R2 bulgu kapanışları (kendi PoC)

| # | Bulgu | Durum | Kanıt |
|---|---|---|---|
| 1 | 429 POST reserve retry | **KAPANDI** | Py/TS: `reserve_server_posts=1` |
| 2 | Retry-After sınırsız | **KAPANDI** | cap 60s → `retry_after_capped`, sleep yok |
| 3 | Rotate 24s zorunlu grace | **KAPANDI** | `revoke_old` / `grace_seconds=0` → 401 anında |
| 5 | https | **KAPANDI** | `http://example.com` reddi; **localhost / 127.0.0.1 / ::1** istisna; `allow_insecure` bayrağı |
| 6 | SSE buffer | **KAPANDI** | 1MiB / 64KiB satır + test |
| 7 | KanusError raw | **KAPANDI** | `repr`/`toString` redacted; secret yok |
| 8 | gofmt | **KAPANDI** | yamalı ağaçta `gofmt -l` boş |
| 9 | OpenAPI aliases | **KAPANDI** (kısmen→tam) | 0009’da login/balance/redeem/requests/workers |
| 4 | US/UYUM çakışma | **Strateji OK** | api-r11 additive; monolitik 0005/0006 term’de hâlâ FAIL |

---

## licensesrv dokunuşları ve defter SHA

### Kim dokunuyor?

| Yama | licensesrv/** | model_proxy.go | hook main.go |
|---|---|---|---|
| api-r11 | **YOK** (yalnız yeni `apiv1/` + main 3 satır) | YOK | Evet (minimal) |
| api-0004 / 0012 | **EVET `store.go`** + testler | YOK | YOK |
| api-0003/0006 (monolitik) | EVET (accounts/credits/store/…) | EVET (0003) | EVET |

API’nin “api-r11 licensesrv’e dokunulmadı” iddiası **doğru**.  
“0004/0012 licensesrv’e dokunulmadı” **yanlış** — dokunuyor → **BLOCKED (US onayı)**.

### SHA (pristine v10 `store.go` vs Order-B yamalı)

| Fonksiyon | Aynı mı? |
|---|---|
| `Settle` (ince sarmalayıcı) | **AYNI** |
| `Reserve` | **DEĞİŞTİ** (`applyReservedDelta`) |
| `settleFor` | **DEĞİŞTİ** (deadline + reserved guard + credits&lt;0 geri alma) |
| `expireLocked` | **DEĞİŞTİ** (deadline + clamp) |

Kanıt: `round3/ledger_sha.txt`.

### api-0004 “davranışı değiştirmiyor”

- **Happy path** Reserve→Settle: `+=`/`-=` yerine `applyReservedDelta`; geçerli miktarlarda aynı sonuç (`TestLedgerHardeningHappyReserveSettleUnchanged` PASS).
- **Üretimde farklı sonuç:** corrupt/negatif Reserved veya commit sonrası Credits&lt;0 → artık **hata** (önceden sessiz negatif mümkün). Panic yok.
- api-0001 regresyon + defter testleri yamalı ağaçta **yeşil** (tam race 411/0/7).

### api-0012 davranış

- Commit, `Expires + commitGrace` içinde kabul (varsayılan **600s**).
- Grace sonrası `expireLocked` reserved serbest, **debit yok**; geç commit reddi → çift düşüm yok.
- PoC’ler: expire/commit yarışı debit≤1; grace’te reserved tutuluyor (yeni reserve yok); env `""/0/-5/abc/30`; reopen sonrası grace içi commit; grace sonrası red.
- `US_CREDIT_COMMIT_GRACE_SECONDS` **üst sınır yok** (çok büyük değer reserved’ı uzun kilitler) — **ORTA**.
- `FAIL_before_grace.txt`: anlatı notu (ham `go test` dökümü değil). Bağımsız doğrulama: grace’siz ağaçta Expires+30s commit → `rezervasyon süresi doldu` (**PASS olarak “red beklenir”**); yamalıda `AcceptsCommitWithinGrace` **PASS**.

---

## Test sayıları

| Kontrol | Sonuç |
|---|---|
| `go vet ./...` | PASS |
| `GOOS=windows` vet + build panel/hook | PASS |
| `go test -race ./...` (v10+Order B) | **pass=411 fail=0 skip=7** (taban 386/0/7; +API testleri) |
| Yeni testler `-count=3` | PASS (Rotate×2, Ledger×3, TTL×4 + diğerleri) |
| S2 regresyon | **36/36** |
| pytest SDK | **11/11** (+ mutasyon FAIL) |
| tsc / node SDK | PASS / **9/9** (+ mutasyon FAIL) |
| api-r11: `go test -race ./apiv1 ./cmd/us-license-hook` | PASS |

---

## Yeni / kalan bulgular

### BLOCKED / politika
1. **api-0004 / api-0012** → `licensesrv/store.go` — US onayı olmadan birleştirilmemeli (`store.go:242–370` civarı).

### ORTA
2. **Apply sıra dokümantasyonu yanlış** (0006b, 0008’den önce isteniyor; gerçekte 0008 sonrası).
3. **`US_CREDIT_COMMIT_GRACE_SECONDS` üst sınır yok** (`commitGraceFromEnv`, `store.go:244–253`) — aşırı değer reserved kilidi.
4. api-0012 grace sırasında reserved başka harcamaya **açılmıyor** (doğrulandı; tasarım notu — bakiye “görünür kullanılabilir” değil).

### DÜŞÜK
5. api-0006 ilk patch hâlâ trailing whitespace uyarısı (0006b gofmt ile temizlemiş olmalı; Order B sonrası gofmt temiz).
6. `FAIL_before_grace.txt` ham test logu değil.

### Çakışmalar
- **GLOBAL × api-r11:** dosya örtüşmesi **yok**.
- **UYUM v2 × api-r11:** aynı `main.go`; UYUM önce → API sonra **çakışmasız** (doğrulandı).
- Monolitik 0005/0006 × UYUM/term: hâlâ çakışır (bilinçli; api-r11 kullan).

---

## NOT TESTED

| Madde | Neden |
|---|---|
| Resmi pinned **v11** + manifest | PANEL’e teslim yok |
| Nihai `.11→UYUM→API→GLOBAL` tek yığında tam `go test -race ./...` | `.11` pin yok; ara r11stack’te yalnız apiv1+hook |
| Windows Job Object / gerçek EXE | Linux kutu |
| Canlı deploy / api-test | yasak |
| Redocly yeniden lint (0009) | bu turda koşulmadı (R2’de 0 idi) |
| Saat geri alma NTP senaryosu (sistem saati) | yalnız mantıksal `now` parametresi PoC |

---

## Sonuç

- **SDK R2 YÜKSEK’leri kapandı → 0010/0011 PASS.**  
- **0006b rotate anında iptal PASS.**  
- **0004/0012 teknik olarak tutarlı ve testli; ancak licensesrv dokunduğu için BLOCKED (US).**  
- **api-r11** licensesrv’e dokunmadan `.11`+UYUM üzerine oturuyor → ara PASS; **nihai PASS `.11` pin sonrası.**  
- Apply için **Order B** kullanılmalı.
