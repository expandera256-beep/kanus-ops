# PANEL API İnceleme — 2026-11-06

**İnceleyen:** PANEL DEV (executor)  
**Zaman:** 2026-11-06 09:39 TSİ (Europe/Istanbul)  
**Kaynak rapor:** `/workspace/us-panel/reports/KANUS_API_REPORT.md`  
**Ham kanıt:** `/workspace/us-panel-dev/panel-evidence/api-review/`  
**S2 sonuç:** `/workspace/us-panel-dev/panel-evidence/review/regression/results/api_0936/`

## Taban ağaç

| Aday | `buildVersion` | Seçim |
|---|---|---|
| `/workspace/us-panel/work2/kiyas/term` | `2026.11.06.10` | **Birincil taban** (görev adayı; `/tmp/api-review/base` buradan) |
| `/workspace/us-panel-dev/r11_base` | `2026.11.06.10` | İkincil doğrulama (ReadyGrace’li GPU wake) |
| `/workspace/us-panel-dev/work/api/r10_base` | `2026.11.06.9` | API botunun kendi çalışma kopyası (raporda yazdığı baseline) |

**Not:** Birincil `.10` term ağacında GPU wake / model proxy testleri **yama öncesinde de** kırık (`ReadyGrace` yok; `us-panel` kopyası `r11_base`’ten farklı). Bu FAIL’ler API yamalarına yazılmadı. Temiz `.10` için `r11_base` tercih edilmeli.

## Yama özeti

| Yama | Tür | Dokunulan dosyalar | Üretim kodu |
|---|---|---|---|
| **api-0001-regression** | git patch | `licensesrv/api_regression_baseline_test.go` (**yeni**, 213 satır, 6 test) | Yok |
| **api-0002-openapi** | belge paketi (git patch değil) | `openapi.yaml`, `ENDPOINT_STATUS.md`, `lint.txt`, `README.md` | Yok |

api-0002 incelemede geçici olarak `docs/api/` altına kopyalandı; üretim derlemesine girmez.

---

## 1) Uygulama (dry-run → apply)

| Hedef | api-0001 dry-run | api-0001 apply | Fuzz/offset |
|---|---|---|---|
| term `.10` | EXIT 0 | EXIT 0, temiz | Yok |
| r11_base `.10` | EXIT 0 | EXIT 0, temiz | Yok |

Kanıt: `panel-evidence/api-review/0001_dry_run.txt`, `0001_apply.txt`.

`diff -rq pristine patched` yalnızca şunları gösteriyor:
- `licensesrv/api_regression_baseline_test.go` (yeni)
- `docs/` (OpenAPI kopyası)

---

## 2) Çakışma analizi (US sahipli / son tur dosyaları)

US listesi: `US_DERLE.bat`, `cmd/us-license-hook/**`, `licensesrv/**`, `worker_client.go`, `account_client.go`, `gateway.go`, `release.go`, `commercial_policy.go`, `model_proxy.go`, `safe_exec.go`, `terminal.go`, `run.go`, `file_tools.go`, `task_state.go`, GPU wake.

| Kontrol | Sonuç |
|---|---|
| api-0001 üretim hunk’ı | **Yok** — yalnız yeni test dosyası |
| api-0002 üretim hunk’ı | **Yok** — YAML/Markdown |
| `s2_*` / `r9_*` / `r10_*` yamalarıyla örtüşen üretim satırı | **Yok** (licensesrv üretim `.go` hunk’ı taranmadı / bulunmadı) |
| Sessiz davranış değişikliği (auth, hesap, kredi yolu) | **Yok** — kod değişmedi |
| `licensesrv/` sahipliği | Test dosyası `licensesrv/` altında; entegrasyonda US/API bot koordinasyonu yeterli, çakışma riski düşük |

**Sonuç:** US `.10` / GPU wake / §2 güvenlik düzeltmeleriyle **çakışma yok**.

---

## 3) gofmt / vet / race / Windows

Ortam: `GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0`

### Yeni dosya
- `gofmt -l licensesrv/api_regression_baseline_test.go` → boş (**temiz**)
- `go vet ./licensesrv` (yamalı term) → EXIT 0

### TestAPIRegression* (asıl yama kapsamı)
| Ağaç | Komut | Sonuç |
|---|---|---|
| term patched | `-race -count=1 -run TestAPIRegression` | **6/6 PASS** (~6.3s) |
| term patched | `-race -count=3 -run TestAPIRegression` | **PASS** |
| r11 patched | `-race -count=3 -run TestAPIRegression` | **PASS** |

### Tam `./...` (bağlam; yama kaynaklı değil)
| Ağaç | Durum | Not |
|---|---|---|
| term pristine | FAIL `kiyas` + `licensesrv` | GPU wake / model_proxy + `TestParsePlansBothExportSchemas` (rapor JSON yolu) |
| term patched | Aynı FAIL seti | API testleri geçiyor; GPU FAIL’leri **önceden var** |
| r11 patched | `ok kiyas`, `ok cmd/us-license-hook`; `licensesrv` fail=1 | Yalnız `TestParsePlansBothExportSchemas` (`../../../../reports/US_PLANS_JSON_TEST.json` yok) — **önceden var**, yama dışı |
| r11 licensesrv json | **pass=78 fail=1 skip=1** | |

### Windows
| Kontrol | Sonuç |
|---|---|
| `GOOS=windows go vet ./...` | EXIT 0 |
| `GOOS=windows go build ./cmd/us-license-hook` | EXIT 0 |
| `GOOS=windows go build .` (panel) | EXIT 0 |

Kanıt: `patched_windows_vet_build.txt`, `patched_api_regression.txt`, `r11_patched_licensesrv.json`.

### Rapor iddiası doğrulama (API bot)
| İddia | Doğrulama |
|---|---|
| `go vet` PASS (kendi r10_base) | Bot logu `logs/go_vet.txt` mevcut; bu incelemede term/r11 vet **PASS** |
| `go test -race ./...` PASS (r10 `.9`) | Bot logu EXIT 0; `.10` term’de GPU kırık → botun `.9` iddiası **çelişmez**, farklı taban |
| 6/6 TestAPIRegression | **Yeniden PASS** (count=1 ve count=3) |
| OpenAPI 20 path | **20/20** kodla birebir |
| Redocly valid, 29 warning | `lint.txt`: `LINT_EXIT:0`, “29 warnings”, “validated” — **doğrulandı** |
| Deploy yok / sır basılmadı | Yamada canlı URL çağrısı yok; testler `httptest` — **uyumlu** |

---

## 4) §2 regresyon paketi

```
bash run_s2_regression.sh /tmp/api-review/patched .../results/api_0936
→ PASS=36 FAIL=0 BLOCKED=0
Kaynak ağacı sha256 manifest: DEĞİŞMEDİ
```

**Sonuç:** API yamaları §2 güvenlik düzeltmelerini **geri almıyor** (allowlist, env sızıntısı, toolexec, symlink, Node `-p`/`--import`, task_state vb. hepsi PASS).

---

## 5) Güvenlik incelemesi

### api-0001 (testler)
- Yalnız `httptest` + mevcut test yardımcıları (`setup`, `post`, `sign("whsec",…)`, `account*`).
- `whsec` / sahte e-posta yalnızca test fixture; üretim sırrı değil.
- Canlı ağ, SMTP, deploy, EXE çalıştırma **yok**.
- Kapsam: bakiye, reserve/release/commit, yetersiz kredi, idempotent settle, 32’li yarışta tek kazanan, hesap aggregate, HTTP 401/402/200, oturumla `/v1/credits/status`.
- **Eksik ama ret nedeni değil:** `use` (saatlik) plan 403, settle 409 gövdesi, rate-limit 429 — sonraki turda genişletilebilir.

### api-0002 (OpenAPI 3.1)
- **Yeni HTTP ucu eklemiyor**; mevcut 20 yolu belgeliyor. Kodda olmayan uç **yok**.
- `security: []` doğru yerlerde: `/healthz`, Shopify webhooks, register/login, verify-email, verification/request, password-reset/*.
- Korumalı uçlar global `BearerAuth` miras alıyor: credits, models, chat, account balance/redeem/requests/workers.
- Webhook’larda HMAC header parametreleri zorunlu; `security: []` + HMAC dokümantasyonu tutarlı.

#### Bulgular

| Seviye | Bulgu | Kanıt |
|---|---|---|
| **ORTA (belge)** | `BearerAuth` şeması “lisans veya oturum” diyor; `/v1/credits/settle` oturum kabulünü ima ediyor. Kodda `usa1` yalnızca **status** için özel-cased; reserve/settle `License(key)` ile ham lisans bekliyor (`credits.go:28–46`, `52–72`). Reserve özeti “ham lisans” diyor (iyi); settle açıklaması eksik. | `openapi.yaml` components.BearerAuth + `/v1/credits/settle`; `licensesrv/credits.go:28-46` |
| **DÜŞÜK (belge)** | `ChatRequest` şemasında `max_tokens` zorunlu değil / bounds yok; kod `max_tokens` 1..8192 ve `messages` 1..256 zorunlu kılıyor; `n`/`best_of` reddediliyor. | `openapi.yaml` ChatRequest; `model_proxy.go:78-96` |
| **DÜŞÜK** | `info.version: 0.7.0-round7` panel `2026.11.06.10` ile uyumsuz (bilinçli Round7 yüzeyi notu var). | `openapi.yaml:4` |
| **DÜŞÜK** | 29 Redocly uyarısı (`operation-operationId` vb.); şema geçerli. Raporla uyumlu. | `lint.txt` |
| **BİLGİ** | Standart hata zarfı yok — raporun PARTIAL iddiası doğru. | OpenAPI description + kod `http.Error` düz metin |

Sır/token loglaması: OpenAPI ve yama içinde gerçek üretim sırrı yok. Testlerdeki `whsec` mevcut suite ile aynı fixture.

---

## Karar tablosu

| Yama | Sonuç | Öneri |
|---|---|---|
| **api-0001-regression** | **PASS** | **Kabul** — üretim dokunulmadı; 6/6×3 yeşil; çakışma yok; S2 36/36 |
| **api-0002-openapi** | **PASS** (belge bulgularıyla) | **Düzeltmeyle kabul** — path/auth genel olarak doğru; settle için “yalnız ham lisans” notu + ChatRequest `max_tokens`/bounds eklenmeli. Bloklayıcı değil. |

### Entegrasyon notu
- Birincil term `.10` (`us-panel/work2/kiyas/term`) GPU testleri kırık; API yamaları bunu bozmaz/düzeltmez.
- Entegrasyon hedefi olarak **`r11_base` (.10, ReadyGrace)** veya GPU düzeltmesi sonrası term tercih edilmeli.
- `TestParsePlansBothExportSchemas` için `reports/US_PLANS_JSON_TEST.json` yolu ağaç düzenine bağlı — API yaması dışı.

### NOT TESTED
- Canlı `api-test.kanus.org` / SMTP / gerçek Shopify HMAC
- EXE çalıştırma (bilinçli)
- Windows Job Object / ADS (S2 planında ayrı; bu tur Linux S2 36/36)

---

## Özet (tek paragraf)

API-0001 temiz uygulanıyor, yalnız regresyon testi ekliyor, 6 test race altında üç kez geçiyor ve §2 güvenlik paketini bozmuyor → **Kabul**. API-0002 kod yazmıyor; 20 uç kodla eşleşiyor, auth sınıflandırması büyük ölçüde doğru; settle’da oturum ima eden Bearer metni ve ChatRequest bounds eksikliği düzeltilerek **kabul** edilmeli. Taban olarak kullanılan `us-panel` `.10` term GPU testlerinde önceden kırık; temiz doğrulama `r11_base` üzerinde yapıldı.
