# US önerileri — API (KAN.US)

Birleştirme: **US .11 → UYUM wiring → API → GLOBAL i18n**.

## api-r11 ekleyici (licensesrv/** DOKUNULMAZ)
`0001-API-r11-additive-*.patch` — yalnız `apiv1/` + `main.go` minimal (privacy.Mount sonrası).

## US-korumalı — onay + elle merge gerekir

### api-0004 ledger hardening + api-0012 TTL commit (Demir B)
- **Dosya:** `licensesrv/store.go` (**değişir** — Reserve / settleFor / expireLocked).
- **Öneri paketi:** `us-proposals/api-0004-ledger-hardening/`, `us-proposals/api-0012-ttl-commit-anchor/`.
- API botu bunları us-panel/term’e **uygulamaz**; yanlış “dokunulmadı” iddiası R3’te düzeltildi.
- Yanlış rapor nedeni (tek satır): work kopyasında uygulandığı için genel “licensesrv’e dokunulmadı” cümlesi 0004/0012’ye de taşındı; o iddia yalnız **api-r11** için doğruydu.

### Diğer (önceki monolitik)
- api-0003 zarf: accounts/credits/model_proxy.go
- api-0006 keys: store APIKeys alanları + credits usk_ + accounts statusByID

## Apply sırası (Order B — doğrulanmış)
`0001 → 0003 → 0005 → 0006 → 0008 → 0006b → (US onayıyla) 0004/0012 store diff`

- `US_CREDIT_COMMIT_GRACE_SECONDS`: varsayılan **600**; cap **900s** (15 dk, KAN.US); 900 üstü → **900’e kırpılır**; geçersiz/negatif/parse → 600; `0` = legacy.
