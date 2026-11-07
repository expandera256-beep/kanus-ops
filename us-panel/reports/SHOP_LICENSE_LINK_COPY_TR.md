# Sipariş sonrası «Lisans anahtarını hesabına bağla» — taslak metin (TR)

**Tarih:** 2026-11-06 09:12 Europe/Istanbul (güncellendi)  
**Durum:** Yalnız kopya taslağı — Shopify’a / panele **kurulmadı**.  
**Kaynak akışlar:** `POST /account/redeem` (lisans → hesap sahipliği; kredi kopyalanmaz), `POST /account/register|login`, panel hesap modu (`deploy/us-license-hook/README_TR.md`, `ACCOUNTS_MODE_GAP_20261106.md`). Hook e-posta konusu örneği: «KAN.US US lisans anahtarınız».

**Yer tutucular**
- `{{LICENSE_KEY}}` — `us1.…` lisans anahtarı  
- `{{ORDER_NAME}}` — örn. `#1003`  
- `{{SUPPORT_EMAIL}}` — örn. `support@kanus.org`  
- `{{SHOP_URL}}` — `https://shop.kanus.org`

**Hesap bağlama yeri:** Web URL değil — **US uygulamasında Hesap ekranı** (`POST /account/redeem`). Exact ekran/buton adları US’nin EXE **.7** teslimatına bağlıdır; şimdilik genel ifade: «Hesap ekranı → Lisansı bağla / Redeem».

---

## A) Sipariş durumu sayfası bloğu (order status / thank-you)

**Başlık:** Lisans anahtarını hesabına bağla

**Gövde:**

Siparişin `{{ORDER_NAME}}` için dijital lisansın hazır.

1. E-postana gelen **KAN.US US lisans anahtarı**nı aç (konu: lisans anahtarın). Anahtar `us1.` ile başlar.  
2. **US uygulamasında Hesap ekranı**nda oturum aç veya kayıt ol (EXE .7 hesap modu).  
3. **Lisansı bağla / Redeem** alanına anahtarı yapıştır ve onayla (kesin buton adı .7 UI ile netleşir).  
4. Bakiye ekranında istem veya süre hakkının göründüğünü kontrol et.

Aynı hesabına ikinci kez bağlamak güvenlidir (yineleme). Anahtarı başka bir hesaba bağlamak reddedilir. Anahtarı kimseyle paylaşma.

Sorun olursa: `{{SUPPORT_EMAIL}}` — sipariş numaran `{{ORDER_NAME}}` ile yaz.

**Kısa alternatif (dar alan):**

Lisansın e-postanda. **US uygulamasında Hesap ekranı** → giriş → Lisansı bağla / Redeem. Destek: `{{SUPPORT_EMAIL}}`.

---

## B) Sipariş onay e-postası bölümü (Order confirmation — ek paragraf)

**Önerilen ek başlık:** Lisansını hesabına bağla

Merhaba,

`{{ORDER_NAME}}` ödemen alındı. Dijital paketinin lisans anahtarı ayrıca **«KAN.US US lisans anahtarınız»** konulu e-postada (veya aynı iletide) gönderilir.

Hesap modunda kullanımı için:

1. **US uygulamasında Hesap ekranı**ndan kayıt ol veya giriş yap.  
2. Lisans anahtarını (**Lisansı bağla / Redeem**; kesin etiket EXE .7) yapıştır.  
3. Bakiyeni kontrol et; ardından aynı uygulamadan modele bağlan.

Anahtar hesabına bağlanınca haklar o oturumdan düşülür; ham anahtar da aynı bakiyeyi paylaşır — iki kez kredi üretilmez.

Mağaza: `{{SHOP_URL}}`  
Destek: `{{SUPPORT_EMAIL}}`

---

## Operasyon notları (kurulum değil)

- Shopify Admin → Bildirimler → Sipariş onayı şablonuna B metni **Demir onayıyla** eklenir.  
- Order status / checkout extensibility veya Additional scripts / thank-you ekleri A için ayrı onay.  
- SMTP kapalıysa (`US_SMTP_ENABLE=0`) müşteri anahtarı görmez; önce mail çıkışı açılmalı veya status sayfasında güvenli gösterim tasarlanmalı (**açık soru**).  
- `loginRequiredAtCheckout=false`; e-posta zorunluluğu Admin’den doğrulanmalı (aksi halde lisans maili gidecek adres yok).
