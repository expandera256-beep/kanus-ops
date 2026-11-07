# uyum-export — Veri dışa aktarma (erişim/taşınabilirlik)

**Tarih:** 2026-11-06 Europe/Istanbul  
**Çalışma kopyası:** `/workspace/us-panel-dev/work/uyum/r10_privacy`  
**Kaynak ağaç (dokunulmadı):** `/workspace/us-panel-dev/work/api/r10_base`

## Kapsam
- `licensesrv.Store.ExportPrivacy` — oturum hesabının e-posta, lisans özeti, istek geçmişi dışa aktarımı
- HTTP: `POST /account/privacy/export` (Bearer oturum)
- Hesap/takım dosyalarına (`accounts.go` vb.) dokunulmadı; ayrı paket + `privacy_ops.go`

## Yamalar
1. `0001-privacy-ops-export-erase-retention.patch` — store ops (export/erase/retention aynı dosya)
2. `0002-privacy-ops-test.patch`
3. `0003-privacy-handler.patch` — HTTP yüzey (preferences paketiyle ortak)
4. Wiring: `../uyum-wiring/0001-privacy-mux-wiring.patch`

## Test
```bash
cd /workspace/us-panel-dev/work/uyum/r10_privacy
go test ./licensesrv/ -count=1 -run Privacy
go test ./internal/privacy/ -count=1
go test ./cmd/us-license-hook/ -count=1
```
**Sonuç (2026-11-06):** üç paket de `ok`.

## Uygulama
1. PANEL DEV incelemesi  
2. `privacy_ops.go` + test + `internal/privacy` + wiring sırasıyla uygula  
3. Canlıya alma: Demir onayı (§14); bu yama yalnız yerel.


---
**NOT (2026-11-06):** Bu paket PANEL FAIL sonrası `patches/uyum-v2/` ile **yerini aldı**. Silinmedi; yeni iş için v2 kullanın.
