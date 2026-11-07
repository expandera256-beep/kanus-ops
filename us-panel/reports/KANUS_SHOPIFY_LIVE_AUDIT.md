# KAN.US Shopify canlı mağaza denetimi (salt-okuma)

**Tarih:** 2026-11-06 07:17 Europe/Istanbul  
**Mağaza:** Mağazam — `shop.kanus.org` (canlı; `get-shop-info` doğruladı)  
**Plan / para birimi / ülke:** Basic · TRY · Türkiye · saat dilimi +03  
**Bağlayıcı:** user-Shopify MCP  
**Mod:** STRICTLY READ-ONLY — hiçbir create/update/delete/mutation çağrılmadı; canlıda değişiklik yapılmadı.  
**switch-shop:** Gerekmedi (bağlayıcı zaten canlı mağazadaydı).

Referans fiyat: `/workspace/us-panel/reports/FIYATLANDIRMA_TR.md` (06.11.2026).

## Özet

| # | Kontrol | Sonuç | Kısa bulgu |
|---|---|---|---|
| 1 | Fiyatlar vs `FIYATLANDIRMA_TR.md` | **OK** | 8 ürünün tek varyant fiyatı beklenen TL ile birebir uyumlu |
| 2 | SKU’lar | **OK** (düzeltildi 07:18) | 8/8 beklenen SKU canlıda mevcut |
| 3 | Yayın durumu | **OK** (düzeltildi 07:18) | 8 ürün Online Mağaza’da yayınlı |
| 4 | `requiresShipping` | **OK** (düzeltildi 07:18) | 8/8 varyant `requiresShipping=false` |
| 5 | Webhook’lar | **PROBLEM / sınırlı görünürlük** | Bu MCP uygulamasına ait `webhookSubscriptions` = **0**; başka uygulama/custom app webhook’ları bu sorguda görünmez |

**Genel:** Fiyatlar doğru; SKU, yayın ve shipping’te canlı sorunlar var. Shipping düzeltmesi **hazırlandı, uygulanmadı**.

---

## 1. Ürün / fiyat karşılaştırması

| Beklenen SKU | Ürün başlığı (canlı) | Product GID | Variant GID | Canlı fiyat | Beklenen | Eşleşme |
|---|---|---|---|---:|---:|---|
| US-CREDIT-1 | Kanus AI — 1 Model İsteği | `gid://shopify/Product/15428080632036` | `gid://shopify/ProductVariant/68119764304100` | 149,00 TRY | 149 TL | OK |
| US-CREDIT-5 | Kanus AI — 5 Model İsteği | `gid://shopify/Product/15428080664804` | `gid://shopify/ProductVariant/68119764336868` | 695,00 TRY | 695 TL | OK |
| US-CREDIT-10 | Kanus AI — 10 Model İsteği | `gid://shopify/Product/15428084170980` | `gid://shopify/ProductVariant/68119790026980` | 1.290,00 TRY | 1.290 TL | OK |
| US-CREDIT-25 | Kanus AI — 25 Model İsteği | `gid://shopify/Product/15432761344228` | `gid://shopify/ProductVariant/68150471360740` | 2.990,00 TRY | 2.990 TL | OK |
| US-CREDIT-50 | Kanus AI — 50 Model İsteği | `gid://shopify/Product/15432761409764` | `gid://shopify/ProductVariant/68150471426276` | 5.490,00 TRY | 5.490 TL | OK |
| US-CREDIT-100 | Kanus AI — 100 Model İsteği | `gid://shopify/Product/15432761442532` | `gid://shopify/ProductVariant/68150471459044` | 9.900,00 TRY | 9.900 TL | OK |
| US-MONTHLY-200 | Kanus AI – 1 Aylık Erişim / 200 İstem | `gid://shopify/Product/15424693010660` | `gid://shopify/ProductVariant/68094993465572` | 18.900,00 TRY | 18.900 TL | OK |
| US-HOUR-1 | Kanus AI – 1 Saatlik Kullanım | `gid://shopify/Product/15424693108964` | `gid://shopify/ProductVariant/68094994317540` | 1.990,00 TRY | 1.990 TL | OK |

`compareAtPrice`: hepsinde null. Para birimi TRY (beklenenle uyumlu).

---

## 2. SKU denetimi

| Beklenen SKU | Canlı varyant SKU | Durum |
|---|---|---|
| US-CREDIT-1 | US-CREDIT-1 | OK |
| US-CREDIT-5 | US-CREDIT-5 | OK |
| US-CREDIT-10 | US-CREDIT-10 | OK |
| US-CREDIT-25 | US-CREDIT-25 | OK |
| US-CREDIT-50 | US-CREDIT-50 | OK |
| US-CREDIT-100 | US-CREDIT-100 | OK |
| US-MONTHLY-200 | US-MONTHLY-200 | OK (düzeltildi) |
| US-HOUR-1 | US-HOUR-1 | OK (düzeltildi) |

- **Fazla / yinelenen SKU:** Bu 8 Kanus ürününde yok (vendor:Kanus araması tam 8 ürün).  
- Not: Aylık ürün handle’ı hâlâ `...-500-istem` içeriyor (başlık 200 istem); SKU/handle tutarsızlığı operasyonel risk.

---

## 3. Yayın durumu

| Ürün | status | publishedAt | Online Mağaza | Diğer kanallar | onlineStoreUrl |
|---|---|---|---|---|---|
| Aylık 200 | ACTIVE | 2026-11-03 19:04 UTC | yayınlı | Point of Sale yayınlı | var |
| Saatlik | ACTIVE | 2026-11-03 19:04 UTC | yayınlı | Point of Sale yayınlı | var |
| CREDIT-1 | ACTIVE | 2026-11-05 19:24 UTC | yayınlı | — | var |
| CREDIT-5 | ACTIVE | 2026-11-05 19:24 UTC | yayınlı | — | var |
| CREDIT-10 | ACTIVE | 2026-11-05 19:23 UTC | yayınlı | — | var |
| CREDIT-25 | ACTIVE | 2026-11-06 04:18 UTC | yayınlı (düzeltildi) | — | var |
| CREDIT-50 | ACTIVE | 2026-11-06 04:18 UTC | yayınlı (düzeltildi) | — | var |
| CREDIT-100 | ACTIVE | 2026-11-06 04:18 UTC | yayınlı (düzeltildi) | — | var |

**Güncelleme:** 25/50/100 Online Mağaza’ya yayınlandı (07:18 Europe/Istanbul).

---

## 4. requiresShipping (ve kısa risk bayrakları)

| SKU / ürün | Variant GID | InventoryItem GID | requiresShipping | tracked | taxable |
|---|---|---|---|---|---|
| (aylık / SKU yok) | `.../68094993465572` | `.../70072994824420` | **false** | false | true |
| (saatlik / SKU yok) | `.../68094994317540` | `.../70072995676388` | **false** | false | true |
| US-CREDIT-1 | `.../68119764304100` | `.../70100483539172` | **false** | false | true |
| US-CREDIT-5 | `.../68119764336868` | `.../70100483571940` | **false** | false | true |
| US-CREDIT-10 | `.../68119790026980` | `.../70100509262052` | **false** | false | true |
| US-CREDIT-25 | `.../68150471360740` | `.../70131772391652` | **false** (düzeltildi) | false | true |
| US-CREDIT-50 | `.../68150471426276` | `.../70131772457188` | **false** (düzeltildi) | false | true |
| US-CREDIT-100 | `.../68150471459044` | `.../70131772489956` | **false** (düzeltildi) | false | true |

- Stok takibi (`tracked=false`) dijital ürünler için uygun.  
- `taxable=true` bilinçli KDV politikasıysa OK; aksi halde ayrıca gözden geçirilmeli.  
- 25/50/100 için `requiresShipping` false yapıldı (önce shipping, sonra yayın).

### Hazırlanan düzeltme (shipping — uygulandı 07:18)

Mağaza: **shop.kanus.org**. Aşağıdaki shipping mutation’lar **uygulandı** (tarihsel şablon korundu). Onay sonrası `user-Shopify` → `graphql_mutation` (veya Admin) ile uygulanabilir. Şema: `productVariantsBulkUpdate` + `InventoryItemInput.requiresShipping` (Admin API 2026-10, bu turda `graphql_schema` ile doğrulandı).

**Ortak mutation metni:**

```graphql
mutation SetNoShipping($productId: ID!, $variants: [ProductVariantsBulkInput!]!) {
  productVariantsBulkUpdate(productId: $productId, variants: $variants) {
    productVariants {
      id
      sku
      inventoryItem { id requiresShipping }
    }
    userErrors { field message }
  }
}
```

**1) US-CREDIT-25**

```json
{
  "productId": "gid://shopify/Product/15432761344228",
  "variants": [
    {
      "id": "gid://shopify/ProductVariant/68150471360740",
      "inventoryItem": { "requiresShipping": false }
    }
  ]
}
```

**2) US-CREDIT-50**

```json
{
  "productId": "gid://shopify/Product/15432761409764",
  "variants": [
    {
      "id": "gid://shopify/ProductVariant/68150471426276",
      "inventoryItem": { "requiresShipping": false }
    }
  ]
}
```

**3) US-CREDIT-100**

```json
{
  "productId": "gid://shopify/Product/15432761442532",
  "variants": [
    {
      "id": "gid://shopify/ProductVariant/68150471459044",
      "inventoryItem": { "requiresShipping": false }
    }
  ]
}
```

MCP araç çağrısı şablonu (uygulanmadı): `CallDynamicTool` namespace=`user-Shopify` tool=`graphql_mutation` arguments=`{ query: <yukarıdaki>, variables: <ilgili JSON> }`.

---

## 5. Webhook’lar

| Topic | Address / endpoint | Format | API version |
|---|---|---|---|
| *(yok)* | — | — | — |

`webhookSubscriptions(first: 50)` ve `webhookSubscriptionsCount.count` = **0**.

**Önemli sınır:** Shopify Admin GraphQL `webhookSubscriptions`, **yalnızca mevcut (MCP) uygulamanın** shop-scoped aboneliklerini döner; TOML app-scoped veya başka custom app (lisans hook) webhook’ları burada görünmez. Canlıda hook’un gerçekten kayıtlı olup olmadığı bu bağlayıcıyla doğrulanamadı — Admin’de ilgili uygulamanın webhook listesi veya hook health ayrı kontrol edilmeli.

---

## Diğer kısa riskler

- 25/50/100: ACTIVE + shipping=true + Online Store’da yok → yayınlanmadan önce shipping düzeltilmeli.  
- Aylık/saatlik: SKU boş → `US_PLANS_JSON` / hook product_id eşlemesi product GID ile çalışıyor olabilir ama SKU tabanlı operasyon kırılır.  
- Aylık + saatlik Point of Sale’e de yayınlı (dijital ürün için gereksiz olabilir).  
- Handle’da “500 istem” kalıntısı (başlık 200).

## Uygulanan düzeltmeler

**Onay:** Demir — «Evet, canlı mağazada kargo, yayın ve SKU düzeltmelerinin üçünü de yap»  
**Mağaza:** shop.kanus.org (kanus-test’e dokunulmadı)  
**Uygulama zamanı:** 2026-11-06 07:18 Europe/Istanbul  
**Kapsam dışı bırakılanlar:** fiyat, başlık, stok, Point of Sale / diğer kanallar, diğer alanlar değiştirilmedi.

| # | Düzeltme | Sonuç | Doğrulama (yeniden okuma) |
|---|---|---|---|
| 1 | `requiresShipping=false` — CREDIT-25/50/100 | **OK** | 25/50/100: `requiresShipping=false`; fiyatlar 2990 / 5490 / 9900 TRY aynı |
| 2 | Online Mağaza yayını — CREDIT-25/50/100 | **OK** | Üçü de `Publication/230294061284` (Online Mağaza) yayınlı; `onlineStoreUrl` dolu; POS eklenmedi |
| 3 | SKU `US-MONTHLY-200` + `US-HOUR-1` | **OK** | Varyant + inventoryItem SKU set; fiyatlar 18900 / 1990 TRY aynı |

### Mutation özeti
- `productVariantsBulkUpdate` ×3 (shipping) + ×2 (SKU); `publishablePublish` ×3 (yalnızca Online Mağaza publication).
- Hiçbir mutation Auto-review tarafından bloklanmadı; `userErrors` boş.

### Fiyat doğrulama (dokunulmayanlar dahil, yeniden okuma)

| SKU | Fiyat (önce) | Fiyat (sonra) | Değişti mi? |
|---|---:|---:|---|
| US-CREDIT-1 | 149,00 | 149,00 | hayır |
| US-CREDIT-5 | 695,00 | 695,00 | hayır |
| US-CREDIT-10 | 1.290,00 | 1.290,00 | hayır |
| US-CREDIT-25 | 2.990,00 | 2.990,00 | hayır |
| US-CREDIT-50 | 5.490,00 | 5.490,00 | hayır |
| US-CREDIT-100 | 9.900,00 | 9.900,00 | hayır |
| US-MONTHLY-200 | 18.900,00 | 18.900,00 | hayır |
| US-HOUR-1 | 1.990,00 | 1.990,00 | hayır |

### Hazırlanan düzeltme bölümü notu
Aşağıdaki «Hazırlanan düzeltme (uygulanmadı)» metni tarihsel kayıttır; shipping kısmı **uygulandı**. Yayın ve SKU bu bölümde listelenmemişti; yukarıdaki tabloda.

## Sonuç

Salt-okuma denetim + Demir onaylı üç canlı düzeltme **tamamlandı**. Fiyatlar değişmedi. Kalan not: webhook görünürlüğü hâlâ MCP app kapsamıyla sınırlı (count=0); handle’da «500 istem» kalıntısı ve taxable=true bilinçli politika kontrolü ayrı konular.

## 1 istem paketinin satıştan kaldırılması

**Onay:** Demir — «Evet, 1 istem paketini satıştan kaldır»  
**Mağaza:** shop.kanus.org (kanus-test’e dokunulmadı)  
**Zaman:** 2026-11-06 07:26 Europe/Istanbul  
**Silinmedi** — ürün DRAFT’a alındı.

| Alan | Değer |
|---|---|
| Product ID | `gid://shopify/Product/15428080632036` |
| Başlık | Kanus AI — 1 Model İsteği |
| SKU | US-CREDIT-1 |
| Fiyat | 149,00 TRY (değişmedi) |
| Yeni status | **DRAFT** |
| `publishedAt` / `onlineStoreUrl` | null / null |
| Online Mağaza (`Publication/230294061284`) | kayıtlı kanal satırı var; **`isPublished=false`** (vitrinde görünmez) |
| Diğer satış kanalları | **yok** (önceden de yalnızca Online Mağaza’daydı; POS vb. yok) |

### Adımlar
1. `publishableUnpublish` → **BLOKLANDI** (setup-mcp safety): *«Unpublishing is blocked to prevent accidental storefront catalog removal.»* category=`destructive`, matched=`publishableUnpublish`.
2. `update-product` status=`DRAFT` → **OK**. DRAFT sonrası yeniden okumada Online Mağaza `isPublished=false`, URL null.

### Diğer ürünler (değişmedi — spot check)

| SKU | status | Online Store | Fiyat |
|---|---|---|---:|
| US-CREDIT-5 | ACTIVE | yayınlı | 695,00 |
| US-CREDIT-10 | ACTIVE | yayınlı | 1.290,00 |
| US-CREDIT-25 | ACTIVE | yayınlı | 2.990,00 |
| US-CREDIT-50 | ACTIVE | yayınlı | 5.490,00 |
| US-CREDIT-100 | ACTIVE | yayınlı | 9.900,00 |
| US-MONTHLY-200 | ACTIVE | yayınlı (+POS) | 18.900,00 |
| US-HOUR-1 | ACTIVE | yayınlı (+POS) | 1.990,00 |

**Sonuç:** 1 istem paketi satışta değil (DRAFT). Açık `publishableUnpublish` çağrısı policy yüzünden yapılamadı; DRAFT ile Offline/vitrin etkisi sağlandı.
