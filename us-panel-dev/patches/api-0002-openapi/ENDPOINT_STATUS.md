# Mevcut API uç sınıflandırması (06.11.2026 TSİ)

Kaynak: `work/api/r10_base` (r10_base kopyası) + Round7 raporları. VERIFIED = bu turda veya Round7 yerel test kanıtı.

| Uç | Sınıf | Not |
|---|---|---|
| GET /healthz | VERIFIED | main.go; smoke/route testleri |
| POST /webhooks/shopify/orders-paid | VERIFIED | HMAC + mint + idempotent |
| POST /webhooks/shopify/orders-cancelled | VERIFIED | revoke + tombstone |
| POST /webhooks/shopify/refunds-create | VERIFIED | kısmi iade ledger |
| GET /v1/credits/status | VERIFIED | lisans + usa1 hesap; salt okuma |
| POST /v1/credits/reserve | VERIFIED | sub=1 kredi; use→403 |
| POST /v1/credits/settle | VERIFIED | idempotent commit; conflict 409 |
| GET /v1/models | IMPLEMENTED | proxy metadata; upstream gerekir |
| POST /v1/chat | VERIFIED | SSE/JSON + reserve/settle (yerel fake upstream) |
| POST /v1/chat/completions | VERIFIED | aynı handler |
| POST /account/register | VERIFIED | bcrypt12, rate, bind |
| POST /account/login | VERIFIED | HMAC session |
| POST /account/balance | VERIFIED | aggregate |
| POST /account/redeem | VERIFIED | tek sahip / yarış |
| POST /account/requests | VERIFIED | son 20 |
| POST /account/workers | IMPLEMENTED | live=false metadata |
| POST /account/verify-email | VERIFIED | outbox kod; SMTP canlı NOT TESTED |
| POST /account/verification/request | VERIFIED | enumeration-safe |
| POST /account/password-reset/request | VERIFIED | enumeration-safe |
| POST /account/password-reset/confirm | VERIFIED | sessionVersion++ |

## §7.5 / §7.6 henüz yok (MISSING / PARTIAL)

| Özellik | Sınıf |
|---|---|
| Sürümlü REST (/v1 dışında contract versioning) | PARTIAL (`/v1/*` var; OpenAPI/SDK/anahtar yok) |
| OpenAPI yayın + lint CI | PARTIAL (bu tur belge üretildi; CI yok) |
| API anahtarları + hash + rotasyon/iptal | MISSING |
| Anahtar başına rate limit / kota | PARTIAL (account IP limiter var; API key yok) |
| Standart hata gövdesi | PARTIAL (karışık text/json) |
| Kullanım analitiği uçları | MISSING |
| Python / TS SDK + SSE mock test | MISSING |
| Takım / org / RBAC | MISSING |
| MFA (TOTP) | MISSING |
| Takım kredileri | MISSING |
| Denetim kaydı (audit log) | MISSING |
| Abonelik yenileme/iptal/fatura akışı (SaaS paket motoru) | PARTIAL (Shopify sub/use plan + webhook; self-serve yenileme yok) |
