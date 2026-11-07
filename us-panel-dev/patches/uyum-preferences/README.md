# uyum-preferences — Gizlilik tercihleri / açık rıza kaydı

## Kapsam
- Sidecar: `privacy-prefs.json` (varsayılan store yanında; `US_PRIVACY_PREFS_PATH`)
- Alanlar: marketingConsent, productUpdates, cookieAnalytics (varsayılan false)
- HTTP: `POST /account/privacy/preferences`
- Kayıt/giriş ile pazarlama rızası birleştirilmez (kodda ayrı endpoint)

## Yamalar
- `0001-prefs-store.patch`
- `0002-privacy-handler.patch`
- `0003-privacy-handler-test.patch`
- Wiring: `../uyum-wiring/`

## Test
`go test ./internal/privacy -count=1` — **ok**.

## Hukuk
Çerez banner / Shopify checkout açık rıza UI’si bu yamada yok — **HUKUKİ İNCELEME GEREKİR** + SHOP UI işi.


---
**NOT (2026-11-06):** Bu paket PANEL FAIL sonrası `patches/uyum-v2/` ile **yerini aldı**. Silinmedi; yeni iş için v2 kullanın.
