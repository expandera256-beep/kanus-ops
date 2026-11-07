# TASLAK — hukukçu incelemesi gerekir, yayınlanmadı

**Konu:** Çerez politikası taslağı (shop.kanus.org + US hesap/API ilişkisi)  
**Tarih:** 2026-11-06 Europe/Istanbul  
**Hukuki tavsiye değildir. Canlı mağazaya konmamıştır.**  
**Tercih anahtarları (uyum-preferences yaması):** `cookieAnalytics`, `marketingConsent`, `productUpdates`

---

## 1. Veri sorumlusu

Unvan `<DOLDUR>` · Adres `<DOLDUR>` · VKN/MERSİS `<DOLDUR>` · Destek support@kanus.org · KVKK başvuru `<DOLDUR>`.

---

## 2. Çerez nedir?

Çerezler, tarayıcınızda saklanan küçük metin dosyalarıdır. Benzer teknolojiler (yerel depolama, piksel) bu politikada «çerez» kapsamında anılabilir.

---

## 3. Kategori tablosu

| Kategori | Amaç | Rıza | uyum-preferences anahtarı | Örnek |
|---|---|---|---|---|
| **Zorunlu / gerekli** | Oturum güvenliği, sepet/checkout temel işlev, yerelleştirme, güvenlik | Genelde aydınlatma ile; pazarlama rızasına bağlanmaz | (ayrı anahtar yok — her zaman açık varsayımı hukukçu onaylamalı) | Aşağıda «görülen» essential / localization |
| **İşlevsel** | Tercih hatırlama, dil, oturum kolaylığı (zorunlu olmayan) | Açık rıza veya meşru menfaat — `<DOLDUR>` | `productUpdates` ile karıştırılmaz; işlevsel UI tercihi ayrı tutulmalı | [DOĞRULANACAK] — shop’ta ayrı işlevsel çerez envanteri yok |
| **Analitik** | Trafik / kullanım ölçümü | Açık rıza (önerilen varsayılan: kapalı) | **`cookieAnalytics`** | `_shopify_analytics` (görüldü) |
| **Pazarlama** | Reklam / yeniden pazarlama / ölçüm | Açık rıza (önerilen varsayılan: kapalı); hizmet şartına bağlanmaz | **`marketingConsent`** | `_shopify_marketing` (görüldü) |

**Ürün güncellemeleri e-postası** çerez değildir; hesap tercihi **`productUpdates`** ile tutulur (e-posta kanalı).

---

## 4. Kanıtlı envanter — shop.kanus.org (2026-11-06 Europe/Istanbul)

**Yöntem:** Salt okuma `curl -sI https://shop.kanus.org/` (kimlik doğrulama yok, mutasyon yok).  
**LIVE_STORE_AUDIT** içinde çerez envanteri yoktu; bu gözlem UYUM tarafından eklendi.

| Çerez adı | Max-Age / expires (yanıt) | HttpOnly / Secure / SameSite | Kategori (taslak sınıflama) | Durum |
|---|---|---|---|---|
| `localization` | ~1 yıl (expires 2027-11-06) | SameSite=Lax | Zorunlu / işlevsel (dil-bölge) — **hukukçu sınıfı kesinleştirsin** | **Görüldü** |
| `_shopify_essential` | Max-Age=31536000 | HttpOnly; Secure; SameSite=Lax; Priority=High | Zorunlu (Shopify platform) | **Görüldü** |
| `_shopify_analytics` | Max-Age=31536000 | HttpOnly; Secure; SameSite=Lax | Analitik | **Görüldü** |
| `_shopify_marketing` | Max-Age=31536000 | HttpOnly; Secure; SameSite=Lax | Pazarlama | **Görüldü** |

**Ek HTTP olgusu (aktarım/edge, çerez değil):** `x-dc: gcp-us-east1`, `server: cloudflare` — Shopify/CDN yolu; ülkesel sonuç **[DOĞRULANACAK]**.

### Görülmeyen / tahmin

| Madde | Not |
|---|---|
| `_ga`, `_fbp`, özel Kanus piksel | Bu turda Set-Cookie’de **görülmedi** — yok demek değil; tema/app sonradan ekleyebilir (**tahmin/izleme**) |
| `cart`, `secure_customer_sig`, eski `_shopify_y` / `_shopify_s` | Klasik Shopify isimleri; bu yanıtta **görülmedi** (Shopify’ın yeni essential/analytics/marketing set’i kullanılıyor olabilir) — **tahmin** |
| Üçüncü taraf reklam çerezleri | HTML/script envanteri bu turda yapılmadı — **[DOĞRULANACAK]** |

---

## 5. US Panel / api.kanus.org (us-license-hook)

| Teknoloji | Kanıt | Not |
|---|---|---|
| HTTP `Set-Cookie` | licensesrv / main.go’da Set-Cookie **yok** (kod taraması) | Oturum **Bearer** (`usa1…`); tarayıcı çerezi değil |
| IP rate limit | Bellekte ~15 dk | Çerez değil; KVKK aydınlatmada IP satırı var |

Masaüstü uygulama yerel depolaması çerez politikası kapsamı dışında bırakılabilir; ayrı bildirim `<DOLDUR>`.

---

## 6. Yönetim ve saklama

- Tarayıcı ayarlarından çerez silinebilir; zorunlu çerezler silinirse mağaza bozulabilir.  
- Shopify tarafı saklama Max-Age ≈ 365 gün (gözlem); Kanus’un silme talebi Shopify çerezlerini doğrudan silmez — **HUKUKİ / operasyon**.  
- Tercih kaydı (hesaplı kullanıcı): `privacy-prefs.json` / API alanları `cookieAnalytics`, `marketingConsent` (yama; canlıda yok).

---

## 7. Güncelleme

Politika sürümü / yürürlük: `<DOLDUR>`.

