# uyum-erase — Silme/unutulma (anonimleştirme)

## Kapsam
- `Store.AnonymizeAccount`: e-posta → `deleted+<hash>@anon.invalid`, parola çöp hash, oturum version++, verify/reset temizliği
- Lisans `Email` alanları psevödonim; **LicenseOwners / Settlements / Orders / Reservations silinmez**
- Outbox `To` anonimleştirilir; gövdede lisans anahtarları `[REDACTED]`
- Tercihler sidecar’dan silinir
- HTTP: `POST /account/privacy/erase`

## Yamalar
- `0001-privacy-ops-shared.patch` (export ile aynı ops dosyası)
- `0002-privacy-handler.patch`
- Wiring: `../uyum-wiring/`

## Test
`go test ./licensesrv/ -run Privacy` ve `./internal/privacy` — **ok** (2026-11-06).

## Yedek uyarısı
VPS yedeklerindeki `licenses.json.*` eski PII taşıyabilir; unutulma için yedek TTL/imha süreci **HUKUKİ + operasyonel** ayrı iş (PROD_PLAN backup 7g/14g).


---
**NOT (2026-11-06):** Bu paket PANEL FAIL sonrası `patches/uyum-v2/` ile **yerini aldı**. Silinmedi; yeni iş için v2 kullanın.
