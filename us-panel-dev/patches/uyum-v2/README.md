# UYUM privacy patches — v2 (PANEL DEV FAIL düzeltmeleri)

> **Not:** `patches/uyum-v3/` **v2 yerini aldı** (Demir: anonimleştirmede lisans iptali). Bu v2 paketi silinmedi.


**Tarih:** 2026-11-06 Europe/Istanbul  
**Taban:** `/workspace/us-panel-dev/pinned/v10` (değiştirilmedi; çalışma kopyası `work/uyum/v10_privacy_v2`)  
**Eski paketler:** `patches/uyum-{wiring,export,erase,retention,preferences}/` — **v2 yerini aldı**; silinmedi.

## Zorunlu uygulama sırası

| Adım | Dizin | İçerik | US onayı |
|---|---|---|---|
| 1 | `01-preferences/` | `iface.go`, `prefs.go`, test | Hayır |
| 2 | `02-ops/` | `licensesrv/privacy_ops*.go` | **Evet** (`licensesrv/**`) |
| 3 | `03-handler/` | `handler.go`, test | Hayır |
| 4 | `04-hook-us/` | `cmd/us-license-hook/main.go` ince mount | **Evet** (korumalı main) |

Her adımda birikimli derleme doğrulandı (`patch -p1` v10 kökünden). `02-ops` tek başına da `go test ./licensesrv -run Privacy|Erase|Retention` PASS.

**Wiring preferences’a runtime bağımlı değil:** `US_PRIVACY_PREFS_PATH` yoksa / `OpenPrefStoreFunc==nil` ise prefs nil; export çalışır. Erase varsayılan **kapalı** (`US_PRIVACY_ERASE_ENABLED=1` ile açılır).

```bash
cd /path/to/v10   # pinned/v10 kopyası
for p in \
  ../patches/uyum-v2/01-preferences/*.patch \
  ../patches/uyum-v2/02-ops/*.patch \
  ../patches/uyum-v2/03-handler/*.patch \
  ../patches/uyum-v2/04-hook-us/*.patch
do patch -p1 --batch --fuzz=0 < "$p"; done
```

## U-01 … U-08 tablo

| ID | Bulgu (PANEL) | Düzeltme | Test kanıtı |
|---|---|---|---|
| **U-01** | wiring+export prefs’siz derlenmiyor | Sıralı yamalar; `PreferenceStore` + `OpenPrefStoreFunc`; Mount prefs’siz export | Birikimli apply PASS; 02 standalone PASS |
| **U-02** | Outbox PII; SHA256(id) takma ad | HMAC-SHA256(`US_PRIVACY_HMAC_KEY`); Account.ID rotasyonu; OrderName/LastErr/To temiz; Body anonim metin; LicenseOwners remap | `TestPrivacyExportMaskAnonymizeLedgerAndResidue` |
| **U-03** | Tek POST erase | `erase/request` → outbox kodu → `erase/confirm`+parola; bayrak kapalıyken route yok | `TestEraseTwoStep*`, HTTP erase 404 |
| **U-04** | Export ham anahtar | `keyMasked` son 4; anonimleştirme sonrası anahtar geçerli (**karar**) | mask assert; `authorizedCredential` true |
| **U-05** | Rate/Origin yok | accounts ile aynı 15dk rate; Origin Host/allowlist; loopback X-Real-IP | origin 403 |
| **U-06** | Bearer zorunlu değil | prefix yoksa 401 | ham token 401 |
| **U-07** | Korumalı dosyalar | `02-ops` + `04-hook-us` ayrı | US notu aşağıda |
| **U-08** | Prefs.Delete yutuluyor | Delete→500; disk/HMAC 5xx; auth 401; MaxBytesReader 400 | handler yolları |

## Ortam (sırları commit/log etme)

| Değişken | Rol |
|---|---|
| `US_PRIVACY_HMAC_KEY` | ≥32 karakter; yoksa erase/anonymize reddeder |
| `US_PRIVACY_ERASE_ENABLED` | `1` → erase route |
| `US_PRIVACY_PREFS_PATH` | Opsiyonel |
| `US_PRIVACY_ADMIN_TOKEN` | retention-run |
| `US_PRIVACY_ALLOWED_ORIGINS` | Opsiyonel Origin listesi |
| `US_PRIVACY_RETENTION_DAYS` | Varsayılan 90 |

HMAC: VPS env 600 veya Secret Manager — **HOST/US kararı**.

## Komut özetleri (2026-11-06)

```
go vet ./licensesrv ./internal/privacy ./cmd/us-license-hook     → OK
GOOS=windows go vet … / go build ./cmd/us-license-hook           → OK
go test -race (licensesrv Privacy|Erase|Retention, privacy, hook) → ok
Birikimli patch 01→04 + 02 standalone                            → PASS
```

## Açık karar noktaları

1. **U-04:** Anonimleştirme sonrası lisans anahtarı geçerli kalır (ödenmiş erişim). İptal ayrı karar.  
2. **U-07:** `02-ops` + `04-hook-us` için US onayı.  
3. **HMAC anahtarı** nerede tutulur.  
4. Erase kodu outbox/SMTP ile (SMTP kapalıysa kuyruk); testler dry-run.

## US entegrasyon notu (`04-hook-us` + `02-ops`)

- **main.go:** import `kiyas/internal/privacy`; `/account/` sonrası `privacy.Mount(...)` (~15 satır). AccountHandler route tablosuna dokunulmaz.  
- **licensesrv:** yalnızca yeni `privacy_ops.go` + test; `accounts.go`/`store.go`/`webhook.go` değişmedi.  
- **Test:** `go test ./cmd/us-license-hook/`; export için `Authorization: Bearer <session>`.  
- **Özellik bayrakları:** erase ve prefs path env ile; varsayılan güvenli (erase off).
