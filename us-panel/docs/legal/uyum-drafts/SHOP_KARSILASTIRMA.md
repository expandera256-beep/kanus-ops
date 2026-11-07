# TASLAK — hukukçu incelemesi gerekir, yayınlanmadı

**Dosya:** SHOP legal final ↔ KANUS_PRIVACY_REPORT karşılaştırma  
**Tarih:** 2026-11-06 Europe/Istanbul  
**Kaynak metinler (DEĞİŞTİRİLMEDİ):** `/workspace/us-panel/docs/legal/final/`  
**Teknik kaynak:** `/workspace/us-panel/reports/KANUS_PRIVACY_REPORT.md`, `LIVE_STORE_AUDIT_20261106.md`  
**Bu belge hukuki tavsiye değildir; SHOP final dosyalarına yazılmamıştır.**

---

## 1. Yöntem

Her madde için: (A) metinde var mı, (B) kodda/altyapıda kanıt var mı, (C) tutarlılık.

---

## 2. Metinde olup kodda / canlıda olmayanlar

| # | Metin iddiası / yer | Teknik gerçek | Tür |
|---|---|---|---|
| 1 | KVKK §5 / Gizlilik: m.11 hakları (öğrenme, silme, düzeltme, itiraz) + 30 gün yanıt | Üretimde export/erase/düzeltme/itiraz API yok; yerel yama henüz merge edilmedi (`KANUS_PRIVACY_REPORT` §5) | **Hak vaadi > kod** |
| 2 | KVKK §5: Başvuru kanalı | `<DOLDUR>`; `support@kanus.org` yetkisi belirsiz | **Metin boş** |
| 3 | Gizlilik / KVKK: saklama süreleri (hesap, sipariş, log, destek) | Tümü `<DOLDUR>`; kodda hesap TTL yok; oturum varsayılan 86400s; yedek 7g/14g metinde yok | **Metin boş + kod kısmi** |
| 4 | KVKK §3 / Gizlilik: yurt dışı aktarım açıklaması | `<DOLDUR>`; teknik envanter raporda var, metne işlenmemiş | **Metin boş** |
| 5 | Kullanım şartları §5: şart kabulü pazarlama/aktarım rızası değildir | Prefs API yalnız yerel yama; canlı checkout/banner rıza UI yok (`LIVE_STORE_AUDIT`) | **İlke metinde var, uygulama yok** |
| 6 | KVKK tablo: destek yazışmaları | Destek kanalı e-posta; ayrı ticket store kanıtı yok | **Belirsiz** |
| 7 | Canlı mağaza: KVKK / iade / şartlar / mesafeli sayfaları | Final taslaklar var; canlıda yok (`LIVE_STORE_AUDIT` §3) | **Taslak ≠ canlı** |

---

## 3. Kodda / altyapıda olup metinde olmayanlar (veya yalnızca `<DOLDUR>`)

| # | Teknik olgu | Kanıt | Final metinde |
|---|---|---|---|
| 1 | **Google Cloud** `europe-west4` (Cloud Run us-commerce/watchdog; GCE `ds-a100` europe-west4-b) | `SRE_AUDIT_20261106.md`; privacy rapor §4 | «Model / bulut» `<DOLDUR>` — **GCP / bölge adı yok** |
| 2 | GCP `us-central1-*` yan VM’ler | SRE audit | **Yok** |
| 3 | Varsayılan model adı **deepseek-v4.1-flash** / `US_MODEL_UPSTREAM` | `cmd/us-license-hook/main.go` | «Model altyapısı» `<DOLDUR>` — **DeepSeek adı yok** |
| 4 | **Hostinger SMTP** (`smtp.hostinger.com:465`, support@…) | `mail.go`, `SMTP_HOSTINGER_20261106.md` | «Barındırma/e-posta ör. Hostinger» geçiyor (KVKK §3) ama tüzel kişi/ülke `<DOLDUR>` |
| 5 | Hostinger **VPS** üzerinde `licenses.json` (e-posta, hash, lisans, outbox) | `PROD_PLAN.md` | Gizlilikte «barındırma» `<DOLDUR>`; VPS/veri dosyası ayrıntısı yok |
| 6 | **IP hız sınırlama** (bellekte ~15 dk; diske yazılmaz) | `accounts.go` rate() | KVKK tabloda IP var; «diske yazılmaz / süre» metinde yok |
| 7 | **Oturum** `usa1` + `US_SESSION_TTL_SECONDS` (86400) | `main.go`, `accounts.go` | «Oturum» satırı var; **TTL rakamı yok** |
| 8 | **Yedekler:** hourly 7g taslak / daily 14g; env 600; at-rest şifreleme yok | `backup-*.sh`, `RESTORE.md` | Saklama/güvenlik `<DOLDUR>` — **yedek süresi ve şifreleme yok** |
| 9 | Outbox gövdesinde lisans anahtarı | `reversals.go` | Metinde yok (güvenlik riski; US botuna redaksiyon) |
| 10 | Shopify storefront çerezleri: `_shopify_essential`, `_shopify_analytics`, `_shopify_marketing`, `localization` | 2026-11-06 `curl -sI https://shop.kanus.org/` | Final gizlilik/KVKK’da **çerez politikası yok** |
| 11 | Shopify CDN ipucu `x-dc: gcp-us-east1` | Aynı HTTP yanıtı | Metinde yok — **[DOĞRULANACAK]** (Shopify/edge) |
| 12 | Yerel prefs anahtarları: `marketingConsent`, `productUpdates`, `cookieAnalytics` | `uyum-preferences` yaması | Metinde çerez/pazarlama ayrımı ilkesi var; **anahtar adları/UI yok** |
| 13 | Privacy export/erase/retention (yalnız yerel kopya) | `work/uyum/r10_privacy` | Metin hakları vaat ediyor; **bu uçlar yayın metnine bağlanmamış** |

---

## 4. Çelişen / gerilimli süreler ve haklar

| Konu | Metin | Kod / ops | Gerilim |
|---|---|---|---|
| Saklama | «gerekli süre» + dört `<DOLDUR>` | Oturum 1 gün; yedek 7–14 gün; hesap süresiz; sweep yaması varsayılan 90 gün (henüz prod değil) | Süreler **karşılaştırılamaz** — metin boş, kod parçalı |
| Silme hakkı | m.11 silme | Üretimde yok; yama = **anonimleştirme**, yok etme değil; yedeklerde PII kalır | Hak vaadi ile teknik yöntem **uyumsuz** olabilir (**HUKUKİ**) |
| Düzeltme | m.11 | E-posta değiştirme API yok | Vaadi karşılıksız |
| Pazarlama rızası ayrılığı | Şart §5 + KVKK giriş | Canlı UI yok; Shopify `_shopify_marketing` çerezi **önceden** set ediliyor (2026-11-06 gözlem) | Banner/rıza olmadan marketing çerezi → **yüksek tutarsızlık adayı** |
| «Aktarım yok» yasağı | KVKK §3 doğru uyarı | Teknik aktarım var | Metin doğru yönde; envanter doldurulmalı |
| Ödeme sağlayıcı | Gizlilik/KVKK `<DOLDUR>` | Canlı audit **UNKNOWN** | Bilinmeyen bilinmeyen |

---

## 5. Dosya bazında özet

| Final dosya | Privacy raporla hiza | Ana boşluk |
|---|---|---|
| `KVKK_AYDINLATMA_METNI.md` | Kategoriler kısmen uyumlu | Alıcı/ülke/süre/hak kanalı `<DOLDUR>`; GCP/DeepSeek adı yok |
| `GIZLILIK_POLITIKASI.md` | Çok kısa | Çerez yok; aktarım/yedek/TTL yok |
| `KULLANIM_SARTLARI.md` | Rıza birleştirmeme iyi | Eğitim/saklama `<DOLDUR>` |
| `CAYMA_*` / `MESAFELI_*` / `ON_BILGILENDIRME_*` | Tüketici; privacy yan | Açık onay `<DOLDUR>`; KVKK ile çapraz |
| `ILETISIM.md` | — | Kimlik `<DOLDUR>` |
| Canlı Shopify politikalar | — | Final metinler **yapıştırılmamış** |

---

## 6. En önemli 5 tutarsızlık (yönetici özeti)

1. **Hak vaadi vs API:** Metin m.11 + 30 gün; üretimde erişim/silme/düzeltme yok.  
2. **Marketing çerezi vs rıza ilkesi:** Şartlar «pazarlama rızası birleştirilmez» der; storefront `_shopify_marketing` rıza UI’siz set (gözlem).  
3. **Aktarım envanteri metne işlenmemiş:** GCP europe-west4, DeepSeek, Hostinger SMTP/VPS kodda; metinde `<DOLDUR>`.  
4. **Saklama/yedek süreleri:** Metin boş; kodda 1g oturum / 7–14g yedek / hesapsız süresiz — çelişki riski.  
5. **Silme = anonimleştirme + yedek kalıntısı:** Metin «silme/yok etme»; teknik tasarım psevödonim + yedek TTL.

