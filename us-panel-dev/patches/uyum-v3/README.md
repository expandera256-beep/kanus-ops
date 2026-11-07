> **v4 yerini aldı**: R3 düzeltmeleri `/workspace/us-panel-dev/patches/uyum-v4/` altında; v3 yamaları silinmedi. (08.11.2026 Europe/Istanbul)

# UYUM privacy patches — v3 (revoke-on-erase + PANEL R2 kapanış)

**Tarih:** 2026-11-06 Europe/Istanbul  
**Kararlar:** Demir (KAN.US): anonimleştirmede lisans iptali; US: lisans PII scrub + hesap bağı kaldır.  
**Taban:** `/workspace/us-panel-dev/pinned/v10` (dokunulmadı; çalışma `work/uyum/v10_privacy_v3`)  
**Önceki:** `patches/uyum-v2/` — **v3 yerini aldı**; v2 silinmedi.

## Uygulama sırası (temiz v10)

| Adım | Dizin | US |
|---|---|---|
| 1 | `01-preferences/` (iface, prefs, open helpers, tests) | Hayır |
| 2 | `02-ops/` | **Evet** |
| 3 | `03-handler/` | Hayır |
| 4 | `04-hook-us/` | **Evet** |

```bash
cd /path/to/v10-copy
for p in $(ls ../patches/uyum-v3/0{1,2,3,4}-*/*.patch | sort); do
  patch -p1 --batch --fuzz=0 < "$p"
done
```

### Birleştirme / .11 notu

1. Önerilen sıra: **US .11 → UYUM v3 → API rebase**.  
2. **`r11_base`** (box): v10 ile main/privacy açısından aynı; v3 apply **PASS** (fuzz=0).  
3. **`work/api-r11-term`**: hareketli .11 + **eski UYUM (v2) zaten gömülü**. Tam v3 «add file» yamaları `already exists` ile FAIL. `main.go` hunk’ları fuzz=0 FAIL; fuzz=3 ile ~+1 satır offset’te tutuyor ama **KeepDays** ve v3 ops/handler içeriği için **elle rebase / dosya değiştirme** gerekir.  
4. **API-0005:** `/account/` ile `/healthz` arası mux bölgesi — UYUM Mount ile çakışır; API yamasını UYUM sonrası yeniden üret / sıraya koy (`patches/api-r11` check notu).  
5. Nihai sabit .11 manifest yoksa entegrasyon **NOT FINAL**; v10 doğrulaması geçerlidir.

## U-xx / R2-xx → düzeltme → test

| ID | Bulgu | Düzeltme | Test |
|---|---|---|---|
| U-01 | prefs’siz derleme | sıralı yamalar; Prefs nil OK | birikimli apply |
| U-02 | outbox PII | HMAC; kesin mail ilişkisi | Export + cross-account negatif |
| U-03 | tek POST erase | request→confirm | `TestEraseTwoStep*` |
| U-04 | ham key / silme sonrası erişim | mask; **revoke** + credential false | revoke assert |
| U-05 | rate/Origin | 15dk; Host | origin 403 |
| U-06 | Bearer | zorunlu | 401 |
| U-07 | US dosyaları | 02+04 ayrı | US öneri |
| U-08 | Prefs.Delete | 500 | handler |
| **R2-01** | Body∋OrderName çapraz hesap | yalnız To==email \|\| OrderID∈hesap | `TestPrivacyEraseDoesNotTouchUnrelatedOutbox` |
| **R2-02** | Subject / OrderID PII | Subject scrub + OrderID eşlemesi | aynı |
| **R2-03** | global challenge race | per-Store `sync.Map`+mu | `TestPrivacyEraseChallengeMultiStoreNoRace` |
| **R2-04** | symlink/FIFO prefs | Lstat regular + unix `O_NOFOLLOW` + LimitReader | Symlink + `TestPrefStoreRefusesFIFO` |
| **R2-05** | lost update / disk rollback | Update tek kilit; save hata→rollback | 200 concurrent + rollback test |
| **R2-06** | erken challenge; kısmi erase | challenge save sonrası; prefs önce Delete | `TestPrivacyEraseChallengeSurvivesFailedSave` |
| **R2-07** | trailing JSON | ikinci Decode EOF | `TestPreferencesTrailingJunkRejected` |
| **R2-08** | Origin scheme | yalnız https (+ allowlist localhost http) | `TestOriginSchemeHTTPSOnly` |
| **R2-09** | KeepDays env | Mount KeepDays ← `US_PRIVACY_RETENTION_DAYS` | `TestPrivacyRetentionDaysMounted` |

## Erase (v3)

- `revokeLocked` ile aktif lisans iptali + rezervasyon release; `privacy_erase:<key>` Settlement.  
- Lisans e-posta scrub; **LicenseOwners silinir** (Orders/Settlements/Refunds kalır).  
- İade yok — **HUKUKİ/TİCARİ İNCELEME GEREKİR**.  
- Uyarı: «aktif lisanslarınız iptal edilecek».  
- İdempotent + `transactionLocked` atomikliği.

## HMAC (HOST — ajan üretmez/yazdırmaz)

- VPS env **mod 600**; box: `/workspace/secrets/`. Değer burada yok.

## US’e öneri

| Dosya | Fonksiyon/ne | Neden | Test |
|---|---|---|---|
| `licensesrv/privacy_ops.go` | `anonymizeLocked`→`revokeLocked`; `mailBelongsToAccount`; `eraseBook` | Demir revoke; R2-01..03/06 | `go test ./licensesrv -run 'Privacy\|Erase'` |
| `licensesrv/privacy_ops_test.go` | revoke/release/DeepEqual/idempotent/cross-account/multi-store/save-retry | kanıt | `-race` |
| `cmd/us-license-hook/main.go` | Mount + KeepDays | R2-09 | `TestPrivacyRetentionDaysMounted` |
| `cmd/us-license-hook/main_test.go` | retention | R2-09 | aynı |

`accounts.go` / `store.go` / `reversals.go` değişmedi.

## Doğrulama (2026-11-06, apply ağacı)

```
go vet ./licensesrv ./internal/privacy ./cmd/us-license-hook     → OK
GOOS=windows go vet … / go build ./cmd/us-license-hook           → OK
go test -race (licensesrv Privacy|Erase|Retention, privacy, hook) → ok
Birikimli patch 01→04 (v10)                                      → PASS
r11_base apply                                                   → PASS
api-r11-term tam apply                                           → FAIL (önceki UYUM var; rebase)
```

## Açık hukuki/ticari

1. Lisans iptali otomatik para iadesi yapmaz (mesafeli sözleşme / kullanılmamış süre).  
2. Onay ekranı metni ürün onayı.  
3. Sabit .11 + API-0005 rebase.
