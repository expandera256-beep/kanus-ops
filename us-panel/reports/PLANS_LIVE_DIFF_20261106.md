# plans.live.json fark özeti

**Tarih:** 2026-11-06 09:15 Europe/Istanbul
**Kaynak test dosyası:** `/workspace/us-panel/deploy/us-license-hook/plans.json` (host-hardening altında `plans.json` yoktu)
**Üretilen:** `/workspace/us-panel/deploy/host-hardening/plans.live.json`
**Durum:** Yerel dosya; **deploy edilmedi**, servis yeniden başlatılmadı, Shopify’a yazılmadı.

## Biçim

Anahtarlar Shopify **product_id** (numeric string). Hook `line_items[].product_id` ile eşler; variant_id plans.json anahtarı değildir (aşağıda referans için listelendi).

## SKU / ID eşlemesi

| SKU | Test product_id | Canlı product_id | Canlı variant_id | status (canlı) |
|---|---|---|---|---|
| US-CREDIT-1 | `15473713283325` | `15428080632036` | `68119764304100` | DRAFT |
| US-CREDIT-5 | `15473713316093` | `15428080664804` | `68119764336868` | ACTIVE |
| US-CREDIT-10 | `15473713348861` | `15428084170980` | `68119790026980` | ACTIVE |
| US-CREDIT-25 | `15473713381629` | `15432761344228` | `68150471360740` | ACTIVE |
| US-CREDIT-50 | `15473713414397` | `15432761409764` | `68150471426276` | ACTIVE |
| US-CREDIT-100 | `15473713447165` | `15432761442532` | `68150471459044` | ACTIVE |
| US-MONTHLY-200 | `15473713479933` | `15424693010660` | `68094993465572` | ACTIVE |
| US-HOUR-1 | `15473713512701` | `15424693108964` | `68094994317540` | ACTIVE |

## Plan alanları (kind / days / prompts / credits / seconds)

| SKU | Test plan | Canlı plans.live.json | Fark |
|---|---|---|---|
| US-CREDIT-1 | `{"kind": "sub", "days": 90, "prompts": 1, "credits": 1}` | `{"kind": "sub", "days": 90, "prompts": 1, "credits": 1}` | aynı |
| US-CREDIT-5 | `{"kind": "sub", "days": 90, "prompts": 5, "credits": 5}` | `{"kind": "sub", "days": 90, "prompts": 5, "credits": 5}` | aynı |
| US-CREDIT-10 | `{"kind": "sub", "days": 90, "prompts": 10, "credits": 10}` | `{"kind": "sub", "days": 90, "prompts": 10, "credits": 10}` | aynı |
| US-CREDIT-25 | `{"kind": "sub", "days": 180, "prompts": 25, "credits": 25}` | `{"kind": "sub", "days": 180, "prompts": 25, "credits": 25}` | aynı |
| US-CREDIT-50 | `{"kind": "sub", "days": 180, "prompts": 50, "credits": 50}` | `{"kind": "sub", "days": 180, "prompts": 50, "credits": 50}` | aynı |
| US-CREDIT-100 | `{"kind": "sub", "days": 365, "prompts": 100, "credits": 100}` | `{"kind": "sub", "days": 365, "prompts": 100, "credits": 100}` | aynı |
| US-MONTHLY-200 | `{"kind": "sub", "days": 30, "prompts": 200, "credits": 200}` | `{"kind": "sub", "days": 30, "prompts": 200, "credits": 200}` | aynı |
| US-HOUR-1 | `{"kind": "use", "days": 30, "seconds": 3600}` | `{"kind": "use", "days": 30, "seconds": 3600}` | aynı |

## Fiyat (plans.json içinde yok; canlı vitrin vs katalog)

| SKU | Canlı fiyat TRY | Katalog (FIYATLANDIRMA) |
|---|---:|---:|
| US-CREDIT-1 | 149 | 149 |
| US-CREDIT-5 | 695 | 695 |
| US-CREDIT-10 | 1290 | 1290 |
| US-CREDIT-25 | 2990 | 2990 |
| US-CREDIT-50 | 5490 | 5490 |
| US-CREDIT-100 | 9900 | 9900 |
| US-MONTHLY-200 | 18900 | 18900 |
| US-HOUR-1 | 1990 | 1990 |

## Notlar

- **US-CREDIT-1** DRAFT; Demir kararıyla mapping’de **tutuldu**.
- Test `plans.json` değiştirilmedi.
- `deploy/host-hardening/plans.json` mevcut değildi; karşılaştırma `deploy/us-license-hook/plans.json` ile yapıldı.
- Production’a kopyalamadan önce HOST/Demir onayı gerekir.
