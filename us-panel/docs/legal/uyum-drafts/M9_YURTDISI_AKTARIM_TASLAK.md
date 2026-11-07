# TASLAK — hukukçu incelemesi gerekir, yayınlanmadı

**Konu:** 6698 sayılı KVKK m.9 — yurt dışı aktarım maddesi (aydınlatma / gizlilik eki taslağı)  
**Tarih:** 2026-11-06 Europe/Istanbul  
**Kesin hukuki hüküm verilmez.** Aşağıdakiler teknik envantere dayalı **seçenekler**dir.  
**Kaynak:** `KANUS_PRIVACY_REPORT.md` §4, `SRE_AUDIT_20261106.md`, `SMTP_HOSTINGER_20261106.md`, `LIVE_STORE_AUDIT_20261106.md`, us-license-hook kaynakları.

---

## Uyarı

Bu metin yayınlanmak üzere değildir. Kurul’un güncel m.9 uygulaması, standart sözleşme metinleri ve bildirim usulleri **hukukçu + Demir** tarafından doğrulanmadan Shopify’a veya siteye konmamalıdır. «Aktarım yoktur» ifadesi kullanılmaz.

---

## 1. Taslak madde metni (yerleştirme: KVKK aydınlatma §3 / Gizlilik «Paylaşım» eki)

### Yurt dışına aktarım

Hizmetin sunulması için kişisel verileriniz, aşağıdaki kategorilerdeki alıcılara **amaçla sınırlı** olarak aktarılabilir veya bu alıcıların sistemlerinde işlenebilir. Aktarımın Türkiye’den yurt dışına yapılıp yapılmadığı, alıcının yerleşik olduğu ülke/bölge ve kullandığı alt işleyenler **doğrulanmış envantere** göre güncellenir.

**Şu an için kesin ülke listesi ve tek bir m.9 dayanağı seçilmemiştir.** Aşağıdaki tablo teknik gözlemleri özetler; boş alanlar `[DOĞRULANACAK]` veya `<DOLDUR>` bırakılmıştır.

---

## 2. Alıcı envanteri (taslak)

| Alıcı kategorisi / marka | Rol (taslak) | Aktarılan veri kategorileri (teknik) | Amaç | Ülke / bölge | Kanıt |
|---|---|---|---|---|---|
| **Shopify** (shop.kanus.org / myshopify) | E-ticaret platformu; ayrı sorumluluk/işleyen ayrımı `<DOLDUR>` | Sipariş kimliği/adı, müşteri e-postası, ürün/plan, ödeme süreci meta verisi | Sipariş, checkout, mağaza | Shopify / CDN: storefront yanıtında `x-dc: gcp-us-east1` görüldü — **Shopify’ın veri yerleşimi [DOĞRULANACAK]** (bu başlık tek başına Google’a aktarım kanıtı sayılmaz) | LIVE_STORE_AUDIT; 2026-11-06 HTTP gözlemi |
| **Ödeme sağlayıcı** | `<DOLDUR>` tüzel kişi | Kart / ödeme verileri (çoğunlukla sağlayıcıda) | Tahsilat | `[DOĞRULANACAK]` — canlı audit’te sağlayıcı **UNKNOWN** | LIVE_STORE_AUDIT §1 |
| **Hostinger** (VPS barındırma + SMTP) | Barındırma / e-posta iletimi; tüzel kişi/rol `<DOLDUR>` | `licenses.json` içeriği (e-posta, parola özeti, lisans, outbox); SMTP ile To/konu/gövde | API hook barındırma, lisans e-postası | VPS: staging IP kaydı yerel raporlarda; SMTP `smtp.hostinger.com` — **ülke [DOĞRULANACAK]** | PROD_PLAN, SMTP_HOSTINGER, mail.go |
| **Google Cloud (GCP)** | Bulut altyapısı / GPU / Cloud Run | Model istem/yanıt (proxy); yönetilen ticaret/gateway trafiği (kapsam [DOĞRULANACAK]) | Çıkarım, watchdog, commerce | Teknik: **europe-west4** (Cloud Run, ds-a100 europe-west4-b); ayrıca us-central1 VM’ler (operasyonel) | SRE_AUDIT; kanus/deploy/README |
| **Model sağlayıcı / upstream** | Varsayılan yapılandırmada model adı `deepseek-v4.1-flash`; fiilî `US_MODEL_UPSTREAM` | Modele gönderilen istem ve proje içeriği | AI işlemi | Upstream URL/ülke **[DOĞRULANACAK]**; DeepSeek tüzel kişi/DPA `<DOLDUR>` | main.go varsayılan model adı |
| Yetkili kamu mercileri | Yasal zorunluluk | Talebe özgü | Mevzuat | TR / ilgili | — |

---

## 3. KVKK m.9 — hukuki dayanak **seçenekleri** (kesin seçim yok)

Hukukçu, her alıcı için (veya alıcı grubu için) aşağıdaki yollardan birini veya birleşimini değerlendirmelidir. **Bu taslak herhangi birini tercih edilmiş saymaz.**

| Seçenek | Özet (taslak dil) | Ne zaman gündeme gelebilir (örnek soru) |
|---|---|---|
| **A. Yeterlilik kararı** | Aktarımın yapıldığı ülke/uluslararası kuruluş hakkında Cumhurbaşkanınca belirlenmiş / Kurul’ca ilan edilmiş yeterlilik varsa | Hedef ülke yeterlilik listesinde mi? `[DOĞRULANACAK]` |
| **B. Uygun güvenceler — standart sözleşme + Kurul’a bildirim** | Taraflar arasında Kurul’ca ilan edilen standart sözleşmenin imzası ve mevzuattaki bildirim | Shopify / Hostinger / Google / model DPA veya SCC benzeri sözleşme var mı? `<DOLDUR>` |
| **C. Bağlayıcı şirket kuralları (BCR)** | Alıcının onaylı BCR’si | Çok uluslu grup içi aktarım mı? Genelde mağaza senaryosunda ikincil |
| **D. Arızi aktarımda açık rıza** | Kanunî şartları sağlayan **arızi** aktarımlarda, aydınlatılmış açık rıza | Sürekli hizmet altyapısı (barındırma, model, e-ticaret) için «genel rıza kutusu» ile meşrulaştırma **risklidir**; hukukçu değerlendirmeli |
| **E. Kanundaki diğer istisnalar** | Sözleşmenin ifası vb. (güncel m.9 metnine göre) | Yalnızca kanunun izin verdiği dar haller; sürekli platform aktarımı için tek başına yeterli olmayabilir — **HUKUKİ İNCELEME** |

**Not (2024 sonrası m.9):** Eski «yalnız açık rıza» yaklaşımı esas alınmamalıdır; güncel resmi metin ve Kurul rehberleri yayın günü yeniden okunmalıdır.

---

## 4. GDPR Chapter V notu (AB kullanıcıları için — uygulanabilirlik `[DOĞRULANACAK]`)

AB/AEA yerleşikleri söz konusuysa (GDPR uygulanabilirliği hukukçuya bırakılır) Chapter V gündeme gelir:

| Araç | Not |
|---|---|
| **Standart sözleşme maddeleri (SCC)** | Avrupa Komisyonu SCC’leri; ithalatçı/ihracatçı rolleri ve transfer impact assessment |
| **AB–ABD Data Privacy Framework (DPF)** | ABD’deki sertifikalı ithalatçılar için; Shopify/Google vb. DPF statüsü `[DOĞRULANACAK]` |
| **BCR / istisnalar (Art. 49)** | Arızi ve dar; sürekli aktarımda dikkat |
| **UK / diğer** | Ayrı rejimler — `<DOLDUR>` |

KVKK m.9 ile GDPR Chapter V **otomatik özdeş değildir**; çift rejim gerekebilir.

---

## 5. Aydınlatmaya eklenecek kısa paragraf (örnek iskelet — doldurulmadan yayınlanmaz)

> Kişisel verileriniz; sipariş ve hesap süreçleri için Shopify’a; e-posta iletimi ve barındırma için Hostinger’a; model çıkarımı ve bulut bileşenleri için Google Cloud üzerinde çalışan altyapıya ve yapılandırılmış model upstream’ine; ödeme için `<DOLDUR>` ödeme kuruluşuna aktarılabilir. Teknik bölge kayıtlarımız arasında Google Cloud **europe-west4** bulunmaktadır. Diğer ülkeler ve alt işleyenler: `[DOĞRULANACAK]`. Aktarımın hukuki dayanağı: `<DOLDUR — A/B/C/D/E seçimi hukukçu>`. Ayrıntılı tablo: [bu belge / KVKK ek sayfası].

---

## 6. Hukukçuya sorular

1. Her satır için işleyen mi, ayrı veri sorumlusu mu?  
2. Shopify ve Hostinger için imzalı DPA / standart sözleşme / Kurul bildirimi var mı?  
3. europe-west4 kişisel veri aktarımı sayılır mı (TR’den erişim modeli)?  
4. DeepSeek / upstream için ülke ve sözleşme?  
5. Ödeme sağlayıcı kim, kart verisi Kanus’a geliyor mu?  
6. Storefront `gcp-us-east1` başlığı Shopify aktarım envanterine nasıl yazılmalı?  
7. AB kullanıcıları için SCC + DPF kombinasyonu gerekli mi?  
8. Sürekli altyapı aktarımında açık rıza kullanılsın mı, kullanılmasın mı?

