# API-PLAN — §7.4 / §7.5 / §7.6 sıralı iş planı

**Tarih:** 06.11.2026 TSİ · **Deploy yok** · Çalışma kopyası: `/workspace/us-panel-dev/work/api/`  
**Öncelik:** geriye uyumluluk + kredi defteri güvenliği önce. US `.9` güvenlik düzeltmeleri `us-panel/` altında; API yamaları yalnız `patches/api-*/`.

## Mevcut temel (kilitle)

- Accounts: register/login/balance/redeem/requests/workers/verify/reset — bcrypt12 + HMAC `usa1` session + IP rate.
- Ledger: `Reserve` / `settleFor` / `settleAmount` / `expireLocked` — atomik JSON (fsync+rename), tek writer flock.
- Plan kinds: `sub` (prompt/kredi) ve `use` (saatlik saniye; model proxy).
- Kanıt: API-0001 regresyon (6 yeni test) + Round7 suite; `go vet`/`go test -race` PASS bu turda.

## Sıra (yamalar)

### P0 — Defter ve sözleşme koruması (önce)
1. **api-0001-regression/** ✅ (bu tur) — ledger/accounts regresyon kilidi.
2. **api-0002-openapi/** ✅ (bu tur) — mevcut yüzey OpenAPI 3.1 + sınıflandırma.
3. **api-0003-error-envelope/** — standart hata gövdesi `{error:{code,message,request_id}}`; mevcut text yanıtlarla dual-mode (Accept/query) veya yalnız yeni `/v1` alt yolları. Risk: istemci kırılması → feature flag / yalnız yeni uçlar.
4. **api-0004-ledger-invariants-hardening/** — yalnız test+dokümantasyon veya minör guard: reserved asla negatif olmasın assert; expire+status tutarlılığı. Üretim düşüm yoluna dokunmadan önce ek race testleri.

### P1 — §7.5 Global API (sürümleme + anahtarlar)
5. **api-0005-v1-versioning/** — `/v1` sözleşmesini sabitle; `OpenAPI` `info.version` + `Sunset`/`Deprecation` başlık politikası; geriye uyumlu alias.
6. **api-0006-api-keys/** — hash'li API anahtarları (prefix+bcrypt/argon2 veya HMAC-SHA256 pepper); oluştur/listele/iptal; **düz anahtar yalnız bir kez döner**. Session ile yan yana: `Authorization: Bearer usk_…`.
7. **api-0007-key-rotation/** — çoklu aktif anahtar, grace, iptal; audit satırı.
8. **api-0008-per-key-ratelimit-quota/** — anahtar başına rate + günlük/aylık kota; 429 + `Retry-After`; mevcut IP account limiter korunur.
9. **api-0009-usage-analytics/** — `/v1/usage` (özet) + audit/settlement türevi; PII sızdırmaz.
10. **api-0010-sdk-python/** + **api-0011-sdk-typescript/** — OpenAPI’den veya elle ince istemci; SSE mock sunucuya karşı test; canlı GPU yok.

### P2 — §7.6 Global kullanıcı sistemi
11. **api-0012-org-team-schema/** — org/team/member şema + migration (JSON ledger → veya Postgres; karar: önce JSON tek-writer kısıtını belgeleyip Postgres’e geçiş tasarımı).
12. **api-0013-rbac/** — roller: owner/admin/member/billing/readonly; uç bazlı guard.
13. **api-0014-totp-mfa/** — TOTP enroll/verify; recovery codes; login step-up; session claim `amr`.
14. **api-0015-team-credits/** — org ortak ledger; üyeler harcar, owner/billing yükler; mevcut tek-lisans sahipliği ile uyum (redeem → org vault).
15. **api-0016-audit-log/** — append-only olaylar (auth, redeem, key, billing); imza/hash zinciri opsiyonel.

### P3 — §7.4 SaaS paket motoru
16. **api-0017-billing-models/** — Pay-as-you-go / Credit Pack / Subscription durum makinesi; Shopify webhook’larla eşleme; yenileme, iptal, kota reset, proration kuralları.
17. **api-0018-enterprise-pricing/** — özel fiyatlandırma mimarisi (override table); canlı ödeme sağlayıcı ekleme YALNIZ doğrulanmış sağlayıcılarla (§7.3).

## Riskler
| Risk | Azaltma |
|---|---|
| Çift harcama / race | P0 testler + mevcut `transactionLocked`; yeni özellik reserve yolunu paylaşır |
| JSON ledger ölçeği / multi-writer | Tek flock; org/team öncesi Postgres tasarımı zorunlu |
| Geriye uyumsuz hata gövdesi | Dual-mode veya yalnız yeni anahtarlı istemciler |
| API anahtarı sızıntısı | Hash-at-rest; loglarda prefix; rotasyon |
| US `.9` ile çakışma | `us-panel/` yazma yok; periyodik rebase work kopyası |
| Canlı SMTP/GPU/Shopify | Bu planda deploy yok; mock/yerel |

## İlk kritik kod yaması (sonraki tur)
**api-0003-error-envelope** veya **api-0006-api-keys** iskeleti — yalnız regresyon yeşil kaldıktan sonra; bu turda özellik kodu yazılmadı.

## Çıkış kriterleri (her yama)
1. Regresyon testleri yeşil (`go vet`, `go test -race`).
2. OpenAPI güncellemesi.
3. `patches/api-NNNN-*/` altında format-patch + README + logs.
4. `KANUS_API_REPORT.md` PASS/FAIL güncellemesi.
5. Sır/env değeri basılmaz; deploy yok.
