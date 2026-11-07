# Shopify ürün şablonları (TR)

**Yayınlanmadı.** Yalnız taslak. Fiyat, stok, canlı ürün ID’si buradan değiştirilmez.
Doğrulanmamış iddia yok; `{…}` alanları mağaza verisiyle doldurulur.

## Ortak yapı

- **Başlık kalıbı:** `Kanus AI — {N} Model İsteği` veya `US Panel — {plan_adı}`
- **Satıcı notu (kısa):** Yerli teknolojiyle fikirlerini çalışan yazılıma dönüştür.
- **Teslimat:** Dijital — lisans / kredi e-posta ve panel üzerinden (`[TESLIMAT_KANALI]`)

---

### Şablon A — Model isteği kredisi

**Başlık:** `Kanus AI — {N} Model İsteği`

**Kısa açıklama:**
```
US Panel’de kullanmak üzere {N} model isteği kredisi. Kanus.org üzerinden yapay zekâ destekli yazılım geliştirme.
```

**Uzun açıklama:**
```
## Bu paket ne sağlar?
- {N} adet model isteği kredisi
- Kullanım: US Panel (Kanus.org)
- Süre: {GECERLILIK_GUN} gün ([PLAN_VERISI])

## Nasıl çalışır?
1. Satın alım sonrası sipariş işlenir.
2. Hesap eşleştirme / oluşturma adımları e-posta ile iletilir ([EPOSTA_AKISI]).
3. US Panel’de giriş yapıp kredini kullanırsın.

## Önemli
- Bir kullanıcı görevi birden fazla ücretli model çağrısı yapabilir.
- GPU / model hazır olana kadar istek kuyruğa alınabilir.
- İade ve iptal: [IADE_POLITIKASI_URL]

Slogan: Yerli teknolojiyle fikirlerini çalışan yazılıma dönüştür.
```

**Etiketler (öneri):** `us-panel`, `kredi`, `kanus`, `dijital`

---

### Şablon B — Aylık / süre planı

**Başlık:** `US Panel — {PLAN_ADI}`

**Kısa açıklama:**
```
{PLAN_ADI}: {KREDI_VEYA_SURE} — Kanus.org US Panel için.
```

**Uzun açıklama:** Şablon A ile aynı iskelet; `{N}` yerine plan alanları.

---

### SEO başlık / meta (TR)

- Title: `Kanus AI — {N} Model İsteği | shop.kanus.org`
- Meta: `US Panel için {N} model isteği. Yerli teknolojiyle yazılım geliştir.`
