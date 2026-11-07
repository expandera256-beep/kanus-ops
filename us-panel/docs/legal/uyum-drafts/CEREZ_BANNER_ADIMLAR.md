# TASLAK — canlıya uygulanmadı, Demir onayı gerekir

**Dosya:** Shopify Customer Privacy / çerez banner kurulum adımları  
**Tarih:** 2026-11-06 Europe/Istanbul  
**Mağaza (LIVE_STORE_AUDIT):** `https://shop.kanus.org` · myshopify `19einb-qz.myshopify.com`  
**İlgili taslaklar:** `CEREZ_POLITIKASI_TASLAK.md`, `RIZA_METINLERI_TASLAK.md`  
**Tercih anahtarları (uyum-preferences):** `cookieAnalytics`, `marketingConsent`, `productUpdates`  
**Bu belge hukuki tavsiye değildir. Shopify Admin’e / canlı mağazaya bu turda yazılmamıştır.**

---

## 0. Önkoşullar (Demir onayı sonrası)

1. Gizlilik / KVKK metinleri hukukçu + Demir tarafından doldurulmuş olmalı.  
2. Shopify resmi yardım: çerez banner’ı kullanmadan önce **privacy policy yayınlanmış** olmalı.  
   Kaynak: [Configuring customer privacy settings](https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/privacy-settings) («Before you can use a cookie consent banner, you must publish a privacy policy»).  
3. Canlı tema yazımı / Settings değişikliği yalnız Demir onayıyla; bu dosya yalnızca prosedür taslağıdır.

---

## 1. Admin yolu (resmi kaynaktan doğrulanmış)

**Doğrulanan menü yolu (İngilizce Admin UI):**

1. Shopify admin → **Settings** → **Customer privacy**  
2. **Cookie banner** seçin / tıklayın  
3. **Regions and content** altında otomatik veya manuel ayar  
4. Görünüm (appearance), konum (Position), kaydet (Save)  
5. İsteğe bağlı: Position altında **View** ile önizleme  

**Kaynak (2026-11-06 WebFetch):**  
https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/privacy-settings  

Aynı sayfada genel giriş: *From your Shopify admin, go to Settings > Customer privacy* → *Click Cookie banner*.

**Not — yol değişebilir:** Shopify Admin menüleri ve yardım URL’leri zamanla değişir. Yayın günü yukarıdaki help.shopify.com sayfasını yeniden açıp adımları doğrulayın. Türkçe Admin’de etiketler «Ayarlar → Müşteri gizliliği → Çerez bannerı» benzeri olabilir; **kesin TR etiketleri bu turda Admin UI’dan tıklanarak doğrulanmadı** → UI dili için **[DOĞRULANACAK]**.

**Üçüncü parti banner kullanırsanız:** Aynı help sayfasında *Configure a cookie banner using a third-party cookie banner app* — native banner’ı kaldırın (`Remove cookie banner`), uygulama bölgelerini Shopify **consent regions** ile eşleştirin. Aksi halde rıza öncesi non-essential işleme riski.

**Geliştirici API (tema / özel banner):**  
https://shopify.dev/docs/api/customer-privacy  
(`setTrackingConsent`, `analyticsProcessingAllowed`, `marketingAllowed`, `shouldShowBanner`)

---

## 2. Bölge seçimi

### 2.1 Resmi davranış (doğrulandı)

Kaynak: aynı [privacy-settings](https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/privacy-settings) sayfası —

| Durum | Resmi özet |
|---|---|
| Otomatik (yeni mağaza varsayılanı) | Aktif marketlerinizde **UK ve EEA** ziyaretçileri için cookie banner otomatik yapılandırılabilir |
| Yapılandırılmış bölgelerde | Non-essential veri **rıza sonrası**; analitik/pazarlama/kişiselleştirme metrikleri düşebilir (beklenen) |
| Yapılandırılmamış bölgeler (örn. ABD örneği metinde) | Banner varsayılan olarak **aktif değil**; çerez tabanlı toplama için opt-in zorunlu sayılmayan bölgeler için tasarım |
| Manuel | *Edit* (Regions) ile banner’ın gösterileceği bölgeleri seçip Save; yardım metni: banner’ı **herhangi bir bölgede** kullanmayı seçebilirsiniz («you can choose to use the cookie banner in any region») |
| Bölge seçici | Anlama sayfası: ülke/eyalet listeli region picker; seçimler marketlere göre — https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/understanding-customer-privacy-settings |

### 2.2 Öneri (hukukçu + Demir; kesin hüküm değil)

| Hedef | Öneri (taslak) | Shopify’da |
|---|---|---|
| **AB/EEA + UK (+ genelde CH ile birlikte anılır)** | GDPR benzeri opt-in; otomatik ayar veya manuel EEA/UK seçimi | Otomatik UK/EEA (aktif market varsa) veya Regions → ilgili ülkeler |
| **Türkiye (KVKK)** | Ana müşteri kitlesi TR ise banner’ın TR’de de görünmesi güçlü aday | Regions’ta **Türkiye seçilebilir mi?** resmi yardım metninde ülke adı tek tek listelenmiyor → **[DOĞRULANACAK]** Admin’de *Edit → Regions* listesinde `Turkey` / `Türkiye` aranmalı |
| **Diğer / küresel** | Rıza tutarlılığı için geniş bölge veya «mümkün olan tüm bölgeler» | Help: any region; market + hukukçu |

### 2.3 Türkiye listede yoksa / doğrulanamazsa — alternatifler

1. **Admin’de mümkün olan en geniş bölge seti** (help: any region) — Demir + hukukçu onayıyla.  
2. **Üçüncü parti CMP** + Shopify consent regions eşlemesi (help third-party bölümü).  
3. **Tema / Customer Privacy API** tabanlı özel banner: `shouldShowBanner` / `setTrackingConsent({ analytics, marketing, preferences })` — https://shopify.dev/docs/api/customer-privacy  
4. Hesaplı kullanıcılar için ayrıca `uyum-preferences` API (`cookieAnalytics`, `marketingConsent`) — mağaza ziyaretçisi çerez rızasının yerine geçmez.

**KVKK için Shopify’ın «otomatik TR» varsayımı yoktur** (otomatik metin UK/EEA odaklı). TR’yi manuel eklemek veya alternatif 1–3 gerekir.

---

## 3. Banner metni taslağı (TR / EN)

Shopify Content düzenleyicide yapıştırılacak iskelet. **Kabul et / Reddet eşit görsel ağırlıkta**; **Tercihler** ayrı. Önceden işaretli kutu yok. Çerez politikası linki yer tutucu.

### 3.1 Türkçe

**Başlık:** Çerez tercihleri  

**Metin:**  
`shop.kanus.org` zorunlu çerezler kullanır (ör. mağaza işlevi). Analitik ve pazarlama çerezleri **yalnızca onayınızla** kullanılır. Ayrıntı: [Çerez politikası — URL_DOLDUR] · [KVKK aydınlatma — URL_DOLDUR].

**Düğmeler:**  
- **Reddet** — yalnız zorunlu  
- **Tercihler** — kategori paneli  
- **Kabul et** — analitik + pazarlama açık  

**Tercihler paneli:**

| Shopify / UI kategori | Açıklama (kısa) | uyum-preferences / API eşlemesi |
|---|---|---|
| Zorunlu (essential) | Mağaza çalışması; kapatılamaz | (her zaman; anahtar yok) |
| Tercih / işlevsel (preferences) | Dil, para birimi vb. hatırlama | Shopify `preferences`; Kanus hesabında ayrı alan yok — **karıştırma:** `productUpdates` e-posta tercihidir, çerez değil |
| Analitik | Kullanım ölçümü | **`cookieAnalytics`** ↔ Shopify `analytics` |
| Pazarlama | Reklam / yeniden pazarlama ölçümü | **`marketingConsent`** ↔ Shopify `marketing` |

**Kabul et** → `analytics=true`, `marketing=true` (`productUpdates` otomatik açılmaz).  
**Reddet** → `analytics=false`, `marketing=false`.

### 3.2 English (banner copy only)

**Title:** Cookie preferences  

**Body:**  
`shop.kanus.org` uses essential cookies to run the store. Analytics and marketing cookies are used **only if you accept**. Details: [Cookie policy — URL_PLACEHOLDER] · [Privacy / KVKK notice — URL_PLACEHOLDER].

**Buttons (equal weight):** **Reject** | **Preferences** | **Accept**  

**Preferences:** Essential (always on) · Preferences/functional · Analytics (`cookieAnalytics`) · Marketing (`marketingConsent`). Product update emails (`productUpdates`) are **not** a cookie toggle on this banner unless you add a separate account setting.

Shopify’ın kendi şablon metni varsa hukukçu onu Kanus metniyle karşılaştırıp seçsin; Network Intelligence açıksa help’teki ek dil gereksinimleri **[DOĞRULANACAK]**.

---

## 4. Tema / uygulama / piksel kontrol listesi

Amaç: analitik/pazarlama piksellerinin **rıza olmadan** çalışmaması.

| # | Kontrol | Kaynak / not | Durum kutusu |
|---|---|---|---|
| 1 | Settings → Customer privacy → Cookie banner bölgelerinde TR (+ EEA/UK) yapılandırıldı mı? | help privacy-settings | ☐ |
| 2 | Privacy policy yayında mı? | Banner önkoşulu | ☐ |
| 3 | **Customer events / Web pixels:** her piksel için gerekli amaçlar (`analytics` / `marketing` / `preferences` / `sale_of_data`) doğru bildirildi mi? | https://shopify.dev/docs/apps/build/marketing-analytics/build-web-pixels · Pixel Privacy | ☐ |
| 4 | Google / Meta vb. app pikselleri Customer Privacy ile entegre mi; Consent Mode v2 (EEA) gerekiyorsa açık mı? | understanding-customer-privacy-settings (Google Consent Mode v2 notu) | ☐ **[DOĞRULANACAK]** mağazada hangi pikseller yüklü |
| 5 | Tema `theme.liquid` / custom HTML’de doğrudan `gtag` / Meta Pixel / hard-coded script var mı? Varsa rıza kapısı yoksa **kaldır veya API’ye bağla** | Manuel envanter | ☐ |
| 6 | Üçüncü parti CMP kullanılıyorsa native Shopify banner kaldırıldı mı ve bölgeler eşleşiyor mu? | help third-party bölümü | ☐ N/A veya yapıldı |
| 7 | Checkout’ta banner konumu isteniyor mu? | Position ayarı; checkout’ta kapalıysa footer «Cookie preferences» | ☐ |
| 8 | Shopify Network Intelligence açık mı? Açıksa ek aydınlatma / rıza yükümlülükleri | help Network Intelligence requirements | ☐ **[DOĞRULANACAK]** |
| 9 | Custom kod asla Shopify çerezlerini elle okuyup yazmamalı; yalnız Customer Privacy API | shopify.dev customer-privacy caution | ☐ |
| 10 | Kanus hesabı prefs (`cookieAnalytics` / `marketingConsent`) mağaza ziyareti rızasının yerine geçmez; ayrı kanal | uyum-preferences | ☐ bilindi |

---

## 5. Yayın sonrası doğrulama

### 5.1 Mevcut durum (banner uygulanmadan) — 6 Kasım 2026 kanıtı

Salt okuma `curl` (rıza çerezi yok; kimlik yok):

```bash
curl -sI -L --max-redirs 2 -A 'KanusUyumBannerDraft/1.0' \
  'https://shop.kanus.org/' | tr -d '\r' | grep -i '^set-cookie:'
```

**Gözlem (2026-11-06 ~09:40 Europe/Istanbul, bu tur tekrarı):** yanıtta şunlar **var**:

- `localization=TR`
- `_shopify_essential=…`
- `_shopify_analytics=…`
- `_shopify_marketing=…`

Bu, yapılandırılmış consent bölgesi dışında (veya banner kapalıyken) Shopify’ın non-essential çerezleri varsayılan verebildiği davranışla uyumlu (help: banner yapılandırılmamış bölgelerde varsayılan olarak aktif değil). **Hedef durum değildir.**

### 5.2 Hedef durum (yapılandırılmış bölgede, rıza yok)

Aynı isteği **rıza/consent çerezi göndermeden** yapın. Yapılandırılmış bölge + Shopify enforcement sonrası beklenen:

| Çerez | Beklenti (taslak doğrulama) |
|---|---|
| `_shopify_essential` / gerekli | Olabilir (zorunlu) |
| `localization` | Olabilir (sınıf hukukçu) |
| `_shopify_analytics` | **Olmamalı** (rıza yokken) |
| `_shopify_marketing` | **Olmamalı** (rıza yokken) |

Topluluk/Shopify notu: consent gereken bölgelerde bu çerezler rıza öncesi yüklenmemeli (forum + help impact bölümü). Kesin çerez adları değişebilir → yayın günü Set-Cookie listesini yeniden kaydedin.

### 5.3 Örnek komutlar

```bash
# 1) Temiz istek — Cookie başlığı YOK
curl -sI -L --max-redirs 2 -A 'KanusPrivacyCheck/1.0' \
  'https://shop.kanus.org/' \
  | tr -d '\r' | tee /tmp/shop-headers-no-consent.txt \
  | grep -iE '^(HTTP/|set-cookie:)'

# 2) Analitik/pazarlama sızdı mı?
grep -iE 'set-cookie:.*_shopify_(analytics|marketing)' /tmp/shop-headers-no-consent.txt \
  && echo 'KALDI: non-essential Set-Cookie var' \
  || echo 'GEÇTI: analytics/marketing Set-Cookie yok (bu kontrol)'

# 3) Zorunlu kalıntı (bilgi)
grep -iE 'set-cookie:.*(_shopify_essential|localization=)' /tmp/shop-headers-no-consent.txt || true
```

**TR IP / AB IP notu:** Shopify bölgeyi ziyaretçi konumuna göre uygular (`getRegion` API). Box/VPS egress IP’si TR, AB veya başka görünebilir; `localization=TR` dil ipucu verir ama **geo kanıtı değildir**.  

- TR’den doğrulama: Türkiye egress veya tarayıcı + VPN kapalı.  
- AB doğrulama: EEA egress veya Shopify region preview / VPN (AB ülkesi) — **Demir laboratuvarı**.  
- Aynı curl farklı IP’den farklı Set-Cookie üretebilir; her senaryoyu ayrı kaydedin.

**Tarayıcı (Reddet sonrası):**

1. Gizli pencere → `https://shop.kanus.org/`  
2. Banner → **Reddet**  
3. DevTools → Application → Cookies: `_shopify_analytics` / `_shopify_marketing` olmamalı (veya rıza=hayır ile uyumlu)  
4. Tercihler’den analitik aç → ilgili çerez/piksel gelebilir  
5. Footer **Cookie preferences** ile geri açılış  

### 5.4 Geçti / kaldı kriterleri

| Test | Geçti | Kaldı |
|---|---|---|
| A — Yapılandırılmış bölge, rıza yok, curl | `_shopify_analytics` ve `_shopify_marketing` **yok** | Bu ikisinden biri Set-Cookie’de |
| B — Tarayıcı Reddet | Analitik/pazarlama pikselleri ağda ateşlenmez (DevTools) | GA/Meta/Shopify marketing pixel rıza öncesi istek |
| C — Kabul et | İlgili çerez/piksel beklenen şekilde çalışır | Hiçbir ölçüm yok (yanlış region / piksel kapalı) — ayrı teşhis |
| D — Önkoşul | Privacy policy yayında; banner regions kaydedilmiş | Policy yok / regions boş |
| E — Mevcut durum (6 Kas) | Referans taban: A **şu an KALDI** (kanıt §5.1) | — |

---

## 6. Geri alma adımları (operasyon)

### 6.1 Ziyaretçi

- Banner → Tercihler → analitik/pazarlama kapat  
- Footer **Cookie preferences** (checkout’ta banner kapalıysa help bunu söyler)  
- Tarayıcı çerezlerini sil / site verisini temizle  

### 6.2 Demir — native banner’ı kapatma / geri alma (Admin)

1. Settings → Customer privacy → Cookie banner  
2. Üçüncü partiye geçiliyorsa help’teki **Remove cookie banner**  
3. Veya Regions’tan bölgeleri kaldır / automated settings’i bilinçli ayarla  
4. Save  
5. §5 curl ile `_shopify_analytics` / `_shopify_marketing` davranışını yeniden ölç (geri alma sonrası non-essential’ın geri gelmesi beklenen olabilir)

### 6.3 Piksel / tema

- Customer events’te şüpheli pikseli disconnect  
- Tema custom script’lerini geri al (git/theme backup)  
- Network Intelligence değişikliği irreversible app uninstall’lara yol açabilir — help’teki uyarıyı oku; **Demir onayı**

### 6.4 Kanus hesap prefs

- `POST /account/privacy/preferences` ile `cookieAnalytics` / `marketingConsent` false (yama merge edildiyse)  
- Mağaza Shopify çerezlerini tek başına silmez

---

## 7. Açık [DOĞRULANACAK] özeti

1. TR Admin UI etiketlerinin birebir çevirisi  
2. Regions listesinde **Turkey/Türkiye** seçilebilirliği (Admin’de tıklanarak)  
3. Mağazada yüklü Customer events / Google / Meta piksel envanteri  
4. Shopify Network Intelligence açık/kapalı  
5. Yayın sonrası gerçek geo (TR vs EEA IP) ile Set-Cookie farkı  
6. `localization` çerezinin zorunlu vs işlevsel hukuki sınıfı  
7. Network Intelligence açıksa zorunlu ek banner/policy dili  

---

## 8. Kaynak listesi (bu turda açılan)

- https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/privacy-settings  
- https://help.shopify.com/en/manual/privacy-and-security/privacy/customer-privacy-settings/understanding-customer-privacy-settings  
- https://shopify.dev/docs/api/customer-privacy  
- https://shopify.dev/docs/apps/build/marketing-analytics/build-web-pixels  
- https://shopify.dev/docs/api/web-pixels-api/pixel-privacy  
- Canlı kanıt: `LIVE_STORE_AUDIT_20261106.md` (domain); UYUM curl 2026-11-06 Set-Cookie  

**Sonuç:** Bu dosya uygulanmış sayılmaz. Demir onayı + (önerilir) hukukçu metin onayı olmadan Admin’de Save yapılmamalıdır.
