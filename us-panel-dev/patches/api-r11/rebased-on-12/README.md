# API r11 additive → .12 rebase (main.go:121 çakışması çözümü)
08.11.2026, TSİ. Yerel öneri; deploy yok, merge US'te.

## Ne değişti
- Çakışma: orijinal `0001-API-r11-additive-…patch` 2. hunk'ı (`cmd/us-license-hook/main.go:121`) UYUM v3 mount satırlarını (`EraseEnabled/AdminToken`, KeepDays yok) bağlam olarak arıyor. UYUM v4'te `privacy.Mount` bloğu farklı (prefStore, prefsReady, KeepDays) → `patch --fuzz=0` FAIL (logs: `work/api/on12/logs/03_additive_dryrun.txt`).
- Çözüm: hunk v4 bloğuna göre yeniden üretildi. UYUM satırlarına dokunulmadı. Eklenenler: `"kiyas/apiv1"` importu ve `apiv1.MountAccountAlias(mux, accounts)` + NOTE yorumu.
- Ek bir yorum satırı: `/v1/fx` için ayrılan yer (UYUM `privacy.Mount` ve apiv1 alias sonrası, `/healthz` öncesi). Kod yok.
- Sonuç sırası (main.go): `/account/` → UYUM privacy.Mount → apiv1 alias (`/v1/account/`) → [/v1/fx yeri] → `/healthz`.
- Dosyalar: `apiv1/{README.md,errors.go,version.go,version_test.go}` (yeni), `cmd/us-license-hook/main.go`. licensesrv'ye dokunmaz.

## Uygulama sırası (`patch -p1 --fuzz=0`)
1. .12 tabanı: `/workspace/us-panel/work2/kiyas/term`
2. UYUM v4: `patches/uyum-v4/01..04` (16/16 temiz)
3. `patches/api-0012-ttl-commit-anchor/addendum-on-12/api-0004-0012-grace-on-12+uyum-v4.diff`
4. `api-r11-additive-on-12+uyum-v4+grace.diff` (bu klasör)

## sha256
- `api-r11-additive-on-12+uyum-v4+grace.diff`: `d715136065bf587d7b32eff3d04fb690c9ac8203998bb785cfa4fd6a7130e85a`

## Testler (tam yığın, taze kopya `work/api/on12/verify`; loglar `work/api/on12/logs/`)
| Kontrol | Sonuç | Log |
|---|---|---|
| Taze uygulama (UYUM 16 + grace + additive) | EXIT=0, ağaç çalışma ağacıyla birebir | 05_fresh_apply.txt |
| go build ./... | EXIT=0 | 10_go_build.txt |
| go vet ./... | EXIT=0 | 11_go_vet.txt |
| go test -race (kiyas, apiv1, cmd/us-license-hook, internal/privacy) | hepsi ok, EXIT=0 | 13a_race_non_licensesrv.txt |
| go test -race ./licensesrv/... | **BLOCKED/NOT TESTED**: 1. koşu dışarıdan öldürüldü (çıktı boş); tek tekrar 57 sn sonra SIGTERM, EXIT=143 (7 PASS, 0 FAIL, sonra `TestAccountSessionAndRawKeyConcurrentLastCredit` sırasında Terminated) | 13b_race_licensesrv.txt, 13b_..._run1_EMPTY_killed.txt |
| GLOBAL 08b ile birlikte uygulama | NOT TESTED | — |
