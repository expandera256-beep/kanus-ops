# api-0004 + api-0012 + grace taşma düzeltmesi (df1ae528) → .12 — licensesrv önerisi (US)
08.11.2026, TSİ. licensesrv US korumalı; bu klasör yalnız öneridir. Deploy yok.

## Neden "ek diff" tam değişikliği içeriyor
US .12'ye commitGrace'i taşımamış: `.12/licensesrv`'de `commitGrace` ya da `COMMIT_GRACE` yok (grep). Eksik olan şey değişikliğin tamamı, bu yüzden diff şunları birleştiriyor:
- `rebased-on-11/api-0004-0012-combined-on-11.diff` (ledger guard'ları + TTL commit anchor)
- `fix-grace-overflow-on-11/fix-grace-overflow-on-11.diff` (sha256 df1ae528…: int64 saniye, en fazla 900, geçersiz/negatifse 600, 0 ise 0, overflow yok; grace boyunca reserved kilitli kalıyor)

## Dosyalar
`licensesrv/store.go`, `ledger_hardening_test.go`, `licensesrv_test.go`, `ttl_commit_anchor_test.go`, `grace_overflow_test.go` (16 uç durum), `ttl_grace_lock_test.go` (`TestTTLGraceKeepsReservedLockedThenReleases`).

## Uygulama sırası
.12 → UYUM v4 (01..04) → **bu diff** → `api-r11/rebased-on-12/…additive…diff`. Diff UYUM'dan bağımsız; UYUM'suz saf .12'ye de temiz uygulanıyor (logs/04_grace_on_pure12.txt).

## sha256
- `api-0004-0012-grace-on-12+uyum-v4.diff`: `4b97a0ad8ffee9b92eca50a66b0efee6ec851e8bc2ce5d09ec3150323d925b71`
- Kaynak fix: `df1ae528483634405cb6b21f3c3a0dd3481db9d2f808b2e10f7991d994c07c28`

## Testler (`work/api/on12/logs/`)
| Kontrol | Sonuç | Log |
|---|---|---|
| .12+UYUM v4 üzerine apply (fuzz=0) | EXIT=0 / EXIT=0 | 02_grace_apply.txt |
| Saf .12 üzerine apply | EXIT=0 | 04_grace_on_pure12.txt |
| Hedefli race (`TestCommitGrace|TestGrace|TestTTL`): 16 uç durum, `TestTTLGraceKeepsReservedLockedThenReleases` dahil 7 test | 7/7 PASS, EXIT=0 | 12_grace_targeted_race.txt |
| go build / go vet (tam yığın) | EXIT=0 / EXIT=0 | 10_, 11_ |
| go test -race ./licensesrv/... | **BLOCKED/NOT TESTED**: 1. koşu dışarıdan öldürüldü (çıktı boş); tek tekrar 57 sn sonra SIGTERM, EXIT=143 (7 PASS, 0 FAIL, sonra `TestAccountSessionAndRawKeyConcurrentLastCredit` sırasında Terminated) | 13b_race_licensesrv.txt, 13b_..._run1_EMPTY_killed.txt |

## Bilinen açık (US .13'te)
- PANEL R3FIX2 HIGH: `accounts.go:769` kapısı grace'i yok sayıyor. Bu diff'te düzeltilmiyor; US ortak deadline fonksiyonunu .13'te getirecek. Kanıt: düzeltme öncesi 4/4 FAIL, logs/20_gate_test_BEFORE.txt.
- `model_meter.go:79` settleAmount commit sınırı ham Expires (grace yok). Ayrıntı `../DEADLINE_CALLSITES_TR.md`.
