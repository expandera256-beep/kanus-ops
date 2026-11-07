# uyum-retention — Saklama TTL temizlik

## Kapsam
- `Store.RetentionSweep(keep)`: eski webhook idempotency, gönderilmiş/dead outbox, süresi dolmuş verify/reset hash
- Varsayılan keep: 90 gün (`US_PRIVACY_RETENTION_DAYS`)
- HTTP: `POST /account/privacy/retention-run` — `US_PRIVACY_ADMIN_TOKEN` Bearer zorunlu
- Ledger (orders/licenses/settlements) temizlenmez

## Yamalar
- `0001-privacy-ops-shared.patch`
- `0002-privacy-handler.patch` + `../uyum-wiring/`

## Test
RetentionSweep birim testi `TestPrivacyExportAnonymizeRetention` içinde — **ok**.


---
**NOT (2026-11-06):** Bu paket PANEL FAIL sonrası `patches/uyum-v2/` ile **yerini aldı**. Silinmedi; yeni iş için v2 kullanın.
