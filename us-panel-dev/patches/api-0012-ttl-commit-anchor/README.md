# API-0012 — TTL commit anchor (**US-korumalı öneri**, Demir B)

## Önemli
`licensesrv/store.go` **değişir**. US onayı şart. us-panel’e yazılmaz.  
Öneri: `patches/api-r11/us-proposals/api-0012-ttl-commit-anchor/`.

## Grace
- Default **600s**; cap **900s** (15m); 900 üstü → 900'e kırp; invalid/neg/parse → 600; `0` = legacy.
- Env: `US_CREDIT_COMMIT_GRACE_SECONDS`.

## Apply sırası (Order B)
`… → 0006 → 0008 → 0006b → 0004/0012`

## Kanıt
- FAIL (yamasız settle): `logs/FAIL_before_grace.txt` (ham `go test`, EXIT 1)
- PASS: work kopyası race/vet logları
