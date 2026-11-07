# API-0005 — /v1 sürümleme

## Ne değişti
- `API-Version: 1` yanıt başlığı: `/v1/credits/*`, `/v1/chat*`, `/v1/models`, `/v1/account/*`
- `/v1/account/*` → mevcut AccountHandler (path rewrite); `/account/*` aynen çalışır
- Kredi defteri semantiğine dokunulmadı

## Geriye uyumluluk
- Tüm eski yollar korunur; alias ek katman

## Test
- vet PASS, race PASS
- `TestAPIVersionHeaderOnV1Wrapper`, `TestV1AccountAliasPreservesLegacyAccountPath`, `TestV1AccountAliasAndAPIVersionHeader`
