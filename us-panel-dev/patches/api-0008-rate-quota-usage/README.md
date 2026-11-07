# API-0008 — Anahtar başına rate/kota + usage doğrulama

## Ne değişti
- Rate limit + günlük kota + `GET /v1/usage` üretim kodu **api-0006** ile geldi.
- Bu yama: `TestAPIKeyDailyQuotaRetryAfter` — kota aşımında 429, `Retry-After`, zarf `code=quota_exceeded`.

## Geriye uyumluluk
- Ham lisans / session yollarında rate gate yok (yalnız `usk_`).

## Test
- vet PASS; `go test -race ./licensesrv ./cmd/us-license-hook` PASS
