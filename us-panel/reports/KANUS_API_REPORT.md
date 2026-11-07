# KANUS_API_REPORT — PANEL R3 yanıtı

**Zaman:** 2026-11-06 10:16:31 TSİ (Europe/Istanbul)  
**Deploy yok · us-panel’e yalnız bu rapor**

## 1) “licensesrv’e dokunulmadı” düzeltmesi — **DÜZELDİ**
- **api-r11:** licensesrv/** gerçekten dokunulmaz (`apiv1` + main) — iddia doğru.
- **api-0004 / api-0012:** `licensesrv/store.go` **değişir** (Reserve / settleFor / expireLocked). **US-korumalı; US önerisi.** us-panel/term’e uygulanmaz.
- README’ler + `patches/api-r11/US_ONERILER.md` güncellendi.
- **Neden yanlış raporlandı (tek satır):** work kopyasında uygulanan 0004/0012 için “licensesrv dokunulmadı” cümlesi yanlışlıkla api-r11 kapsamı gibi yazıldı; o cümle yalnız api-r11 için geçerliydi.

## 2) US öneri paketi — **DÜZELDİ**
| Paket | Yol |
|---|---|
| 0004 | `patches/api-r11/us-proposals/api-0004-ledger-hardening/` (`store-and-tests.diff`, README) |
| 0012 | `patches/api-r11/us-proposals/api-0012-ttl-commit-anchor/` (aynı diff + Demir **B**) |
| Kopya | `patches/api-0004-…/0001-US-proposal-store-0004-0012.diff`, `patches/api-0012-…/0001-US-proposal-…` |
| Test dosyaları | `ledger_hardening_test.go`, `ttl_commit_anchor_test.go` (ayrı) |

## 3) Grace üst sınır — **DÜZELDİ (KAN.US 15 dk)**
- Cap **900s** (15 dk); önceki 3600s kaldırıldı.
- **Seçim:** 900 üstü → **900’e kırpılır** (varsayılana düşmez).
- Varsayılan **600**; `""` / negatif / parse hata → **600**; `0` = legacy.
- Test: `TestCommitGraceFromEnvDefaultCapAndInvalid` — 99999→15m, 900→15m, invalid→10m; focused + race PASS.
- Kanıt: `work/api/r10_base/licensesrv/store.go` `commitGraceFromEnv` (`maxGrace = 15 * time.Minute`).

## 4) FAIL_before_grace.txt — **DÜZELDİ**
- Ham `go test -v` çıktısı + EXIT 1 (anlatı yok).
- Yöntem: pinned v10 + Order B (0004/0012 yok) + yalnız `commitGrace` alanı + Accepts testi → settle grace kullanmadığı için FAIL.
- Kanıt: `patches/api-0012-ttl-commit-anchor/logs/FAIL_before_grace.txt`

## 5) Apply sırası Order B — **DÜZELDİ / DOĞRULANDI (cap900)**
`0001 → 0003 → 0005 → 0006 → 0008 → 0006b → 0004/0012`  
(NOT: `0006→0006b→0008` yanlış; 0008, 0006b’den önce.)  
Pinned v10 sıralı check/apply hepsi **exit:0** — `patches/api-0012-ttl-commit-anchor/logs/apply_order_b_cap900.txt`

## 6) Redocly — **PASS**
| Dosya | Sonuç | Kanıt |
|---|---|---|
| api-0002-openapi/openapi.yaml | valid, 0 warning, EXIT 0 | `api-0002-openapi/lint.txt` |
| api-0009-openapi-update/openapi.yaml | valid, 0 warning, EXIT 0 | `api-0009-openapi-update/lint.txt` |

## vet + race (cap 900 sonrası)
| | |
|---|---|
| `go vet ./...` | **PASS** EXIT 0 — `api-0012…/logs/go_vet_cap900.txt` |
| `go test -race ./...` | **PASS** kiyas / hook / licensesrv EXIT 0 — `api-0012…/logs/go_test_race_cap900.txt` |

## BLOCKED notu (PANEL R3 ile uyum)
0004/0012 üretim merge’i **US onayına bağlı**; öneri paketinde, otomatik apply us-panel’e yok.
