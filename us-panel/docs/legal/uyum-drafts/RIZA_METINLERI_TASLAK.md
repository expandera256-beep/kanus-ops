# TASLAK — hukukçu incelemesi gerekir, yayınlanmadı

**Konu:** Çerez bannerı, açık rıza metinleri, rıza geri alma  
**Tarih:** 2026-11-06 Europe/Istanbul  
**İlkeler:** Açık rıza, hizmet şartı / satın alma kutusuna **bağlanmaz**; pazarlama ile yurt dışı aktarım rızası **ayrı** tutulur.  
**Tercih anahtarları:** `marketingConsent`, `productUpdates`, `cookieAnalytics` (`uyum-preferences`).

---

## 1. Çerez banner metni (eşit ağırlıklı eylemler)

**Başlık:** Çerez tercihleri  

**Gövde (kısa):**  
Shop.kanus.org, mağazanın çalışması için zorunlu çerezler kullanır. Analitik ve pazarlama çerezleri yalnızca siz onaylarsanız kullanılır. Ayrıntı: Çerez politikası · KVKK aydınlatma.

**Düğmeler (görsel eşit ağırlık — hukukçu/UX onayı):**  
1. **Reddet** — Yalnız zorunlu çerezler (`cookieAnalytics=false`, `marketingConsent=false`)  
2. **Tercihler** — Kategori bazlı panel  
3. **Kabul et** — Analitik + pazarlama açık (`cookieAnalytics=true`, `marketingConsent=true`); productUpdates **otomatik açılmaz**

**Tercihler paneli:**

| Anahtar | Etiket (TR) | Varsayılan |
|---|---|---|
| (zorunlu) | Zorunlu çerezler | Her zaman açık (kapatılamaz) |
| `cookieAnalytics` | Analitik çerezler | Kapalı |
| `marketingConsent` | Pazarlama çerezleri / ölçümü | Kapalı |
| `productUpdates` | Ürün / hizmet e-posta güncellemeleri (çerez değil) | Kapalı |

Kaydet → tercihler saklanır; banner kapanır.  
**Önceden işaretli kutu yok.**

---

## 2. Açık rıza metni — pazarlama (ayrı)

**Amaç alanı:** E-posta / elektronik ileti ile kampanya, indirim, duyuru (`marketingConsent`).

**Metin iskeleti:**  
☐ Kanus / US Panel hakkında kampanya ve tanıtım iletileri almak için **açık rızamı** veriyorum. Bu rıza satın alma veya hesap açma şartı değildir. İstediğim zaman geri alabilirim: `<DOLDUR başvuru/tercih linki>`.

**Bağlanmaz:** Kullanım şartları onay kutusu, mesafeli satış / cayma onayları, ödeme düğmesi.

---

## 3. Açık rıza metni — analitik çerez (`cookieAnalytics`)

☐ Mağaza deneyimini ölçmek için analitik çerezlerin kullanılmasına izin veriyorum. Ayrıntılar Çerez politikasındadır. İstediğim zaman banner / Tercihler’den kapatabilirim.

---

## 4. Açık rıza metni — yurt dışı aktarım (yalnız hukukçu «D / arızi rıza» seçerse)

**Uyarı:** Sürekli barındırma / Shopify / model altyapısı için m.9’da açık rıza **tek başına uygun olmayabilir**. Aşağıdaki metin **yalnızca** hukukçunun arızi/istisnai rıza seçeneğini işaretlediği haller içindir; A/B/C dayanakları seçilmişse bu kutu **kullanılmamalıdır**.

☐ KVKK aydınlatma metninde belirtilen yurt dışı aktarımlar hakkında bilgilendirildim; **arızi** aktarım kapsamındaki açık rızamı veriyorum. Rızam hizmet şartının kabulünden bağımsızdır. Geri alma: `<DOLDUR>`.

Ayrı kutular: (1) pazarlama, (2) aktarım — **tek kutuda birleştirilmez**.

---

## 5. Ürün güncellemeleri (`productUpdates`)

☐ Ürün, güvenlik ve hizmet güncellemeleri hakkında bilgilendirme e-postası almak istiyorum. (Pazarlama kampanyasından ayrıdır; `marketingConsent` ile birleştirilmez.)

---

## 6. Rıza geri alma

- **Çerez:** Banner → Tercihler veya tarayıcı çerez silme; `cookieAnalytics` / `marketingConsent` false.  
- **Pazarlama / productUpdates:** Hesap gizlilik tercihleri API (`POST /account/privacy/preferences` — yama) veya e-posta «abonelikten çık» / KVKK başvurusu `<DOLDUR>`.  
- **Aktarım rızası (kullanıldıysa):** Yazılı başvuru `<DOLDUR>`; altyapısal aktarımın durması teknik olarak mümkün olmayabilir — aydınlatmada belirtilmeli (**HUKUKİ**).  
- Geri alma, rıza öncesi hukuka uygun işlemi geçmişe dönük tek başına hukuka aykırı kılmaz (genel ilke; kesin hüküm değil).

---

## 7. Teknik eşleme

| UI eylemi | JSON alanları |
|---|---|
| Reddet | `cookieAnalytics:false`, `marketingConsent:false` (`productUpdates` değişmez) |
| Kabul et | `cookieAnalytics:true`, `marketingConsent:true` |
| Yalnız analitik | `cookieAnalytics:true`, `marketingConsent:false` |
| Ürün e-postası | `productUpdates:true/false` |
| source alanı | örn. `"shop-banner"`, `"account-settings"` |

Canlı Shopify’da `_shopify_marketing` / `_shopify_analytics` şu an rıza UI’siz set edilebiliyor (2026-11-06 gözlem) — banner uygulanana kadar **tutarsızlık** devam eder (`SHOP_KARSILASTIRMA.md`).

---

## 8. Hukukçuya sorular

1. Reddet’te Shopify `_shopify_marketing` / `_shopify_analytics` teknik olarak engellenebilir mi (Customer Privacy API / theme)?  
2. `localization` zorunlu mu işlevsel mi?  
3. AB ziyaretçileri için GDPR consent mode gerekir mi?  
4. Aktarım rızası kutusu hiç kullanılacak mı?

