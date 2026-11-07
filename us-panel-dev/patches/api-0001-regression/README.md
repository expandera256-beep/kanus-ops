# API-0001 — Regresyon temeli (accounts + kredi defteri)

**Tarih:** 06.11.2026 TSİ · **Deploy yok** · **us-panel ağacına yazılmadı**

## Amaç
§7.4–7.6 özelliklerinden önce mevcut `us-license-hook` / `licensesrv` davranışını kilitlemek:
bakiye, reserve/settle/release, yetersiz kredi, idempotent settle, eşzamanlı son kredi, hesap bakiyesi toplama, HTTP durum sözleşmesi.

## Kaynak çalışma kopyası
`/workspace/us-panel-dev/work/api/r10_base/` (baseline: `r10_base` kopyası; orijinal `us-panel/` dokunulmadı)

## Yama
`0001-API-0001-regression-tests-locking-accounts-credit-le.patch`
- Eklenen dosya: `licensesrv/api_regression_baseline_test.go` (6 test)

## Çalıştırılan kontroller (bu turda gerçek)
| Kontrol | Sonuç | Kanıt |
|---|---|---|
| `go vet ./...` | PASS (exit 0) | `logs/go_vet.txt` |
| `go test -race ./... -count=1` | PASS: kiyas, cmd/us-license-hook, licensesrv; pricing-gen no tests | `logs/go_test_race.txt` |
| `TestAPIRegression*` (6) | PASS hepsi | `logs/api_regression_v.txt` |

## Uygulama
```bash
cd /workspace/us-panel-dev/work/api/r10_base
git apply /workspace/us-panel-dev/patches/api-0001-regression/0001-API-0001-regression-tests-locking-accounts-credit-le.patch
go test -race ./licensesrv -run TestAPIRegression -count=1
```
