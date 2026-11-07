# API-0006 — Hash'li API anahtarları + rotasyon

## Ne değişti
- `APIKeyRecord` yalnızca hash saklar; düz `usk_…` anahtar yalnız create/rotate JSON `key` alanında bir kez döner.
- Uçlar (session `usa1` gerekir): `POST/GET /v1/api-keys`, `POST /v1/api-keys/{id}/rotate` (24s grace), `POST .../revoke`.
- `GateAPIKey` / `AuthorizeBearer`: anahtar başına rate + günlük kota (varsayılan 60/dk, 10000/gün); 429 + `Retry-After` + standart zarf.
- `GET /v1/usage`: PII/secret sızdırmaz özet.
- Credits status API key ile hesap bakiyesi döner. Reserve/settle ledger semantiği aynı.
- Liste/yanıtlar secret içermez; store JSON'da plaintext yok.

## Geriye uyumluluk
- Lisans `us1` ve session `usa1` yolları değişmedi.

## Test
- vet PASS, race PASS
- `TestAPIKeyCreateListRotateRevokeAndGrace`, `TestAPIKeyRateLimitRetryAfter`, `TestUsageEndpointNoSecretLeak`

## Rotasyon grace (api-0006 / R2 düzeltme)
- Varsayılan: eski anahtar **24 saat** `grace` durumunda geçerlidir.
- Anında düşürme: `POST .../rotate` gövdesi `{"revoke_old":true}` veya `{"grace_seconds":0}`.
