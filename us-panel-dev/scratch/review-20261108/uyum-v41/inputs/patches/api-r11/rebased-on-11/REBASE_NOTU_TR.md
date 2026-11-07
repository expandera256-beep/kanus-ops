# API 0004 + 0012 → .11 rebase notu (08.11.2026, Europe/Istanbul)

- Taban: `/workspace/us-panel/work2/kiyas/term` (.11; değiştirilmedi). Çalışma kopyası: `/workspace/us-panel-dev/rebase-api/term`.
- api-0004 ve api-0012 `store-and-tests.diff` dosyaları byte-byte aynı (cmp); bir kez uygulandı.
- Yama: `api-0004-0012-combined-on-11.diff` (yalnız `licensesrv/`; .11 ağacına karşı; `patch -p1`). sha256 `9c8a4c45553d64ee60f2562aeac2f2020fb70ad90e5c4e49d0812415d10751a4`.

## Çakışma: licensesrv/store.go hunk #3 (OpenStore)
Öneri, bağlam olarak `UsageEvents` nil-init bloğunu bekliyordu; .11'de bu blok yok, onun yerine `Accounts` ve `LicenseOwners` nil-init blokları var. Çözüm: .11 init blokları olduğu gibi korundu, tek satır `s.commitGrace = commitGraceFromEnv()` `good = true` öncesine (LicenseOwners init'inden sonra) eklendi. Diğer hunk'lar −8 satır ofsetle temiz.
Davranış: US_CREDIT_COMMIT_GRACE_SECONDS varsayılan 600, üst sınır 900 (öneri kodu değişmedi).

## Testler
- gofmt: değişen dosyalar temiz (ağaçtaki 30 dosyalık gofmt listesi .11'de de aynı, önceden var).
- go vet ./...: temiz.
- go test ./...: 403 PASS, 0 FAIL (kiyas, cmd/us-license-hook, licensesrv ok).
- go test -race ./...: hepsi ok.
- S2 regresyon: PASS=36 FAIL=0 BLOCKED=0 (results/20261108_004123).

## UYUM v3 dry-run (.11 + API rebase üzerine)
`patch --fuzz=0`, sıra 01-preferences → 02-ops → 03-handler → 04-hook-us: **TEMİZ** (yalnız 04/0001 hunk #2 +6 ofset). Uygulanmış kopyada go vet temiz, licensesrv + us-license-hook testleri ok.
Not: uyum-v3/README "v4 yerini aldı" diyor (`patches/uyum-v4/`); hangisinin birleşeceği kararı açık.
Önerilen uygulama sırası: .11 → API 0004/0012 (bu yama) → UYUM 01→02→03→04 (ya da UYUM önce; ikisi de temiz, ama UYUM README'si UYUM→API sırasını öneriyor).
Taban doğrulandı: r12_prev/term_11 == dist .11 TREE_MANIFEST (sha256sum -c temiz). .12'ye rebase gerekecek.

## GÜNCELLEME — UYUM v4 ve main.go:121 çakışması (08.11.2026 ~00:55 Europe/Istanbul)
- UYUM v3 yerine **v4** dry-run yapıldı (v4, v3'ün yerini aldı). Taban: `r12_prev/term_11` (dist .11 TREE_MANIFEST ile sha256sum -c temiz).
- Sıra ve sonuç (git apply --check, sonra apply):
  1. .11
  2. `api-0004-0012-combined-on-11.diff`: TEMİZ (fuzz=0)
  3. UYUM v4: 01-preferences (7) → 02-ops (4, `licensesrv/store.go` dahil) → 03-handler (3) → 04-hook-us (2): 16/16 TEMİZ, API store yaması ile çakışma yok
  4. API r11 additive (`0001-API-r11-additive-…patch`): orijinali `cmd/us-license-hook/main.go:121` hunk'ında FAIL veriyor. Hunk eski UYUM mount satırlarını (`EraseEnabled/AdminToken`) arıyor, oysa v4 Mount bloğu farklı (prefStore, keepDays). **Çözüm:** `apiv1` dosyaları olduğu gibi uygulandı; `main.go`'da `"kiyas/apiv1"` importu eklendi ve `apiv1.MountAccountAlias(mux, accounts)` + NOTE satırı, v4 `privacy.Mount(...)` bloğundan sonra, `/healthz`'tan önce elle eklendi. UYUM Mount'una dokunulmadı.
     Yeniden üretilmiş yama: `api-r11-additive-on-11+api+uyum-v4.diff`
- Tüm yığın tek diff olarak (.11'e karşı): `FULL_STACK_11+api+uyum-v4+additive.diff`
- Yığın testleri: go vet ./... temiz. licensesrv, cmd/us-license-hook, apiv1 ve internal/privacy paketlerinde go test ve -race ok. Kök `kiyas` paketi bu yığında koşulmadı; ona dokunan bir yama yok, rebase-api ağacında testi ve -race'i geçti.
- Hepsi .11'e karşı; work2'de hazırlanan .12'ye rebase gerekecek.
