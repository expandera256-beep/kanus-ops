# KANUS — KVKK/GDPR teknik boşluk ve gizlilik raporu

**Tarih:** 2026-11-06 Europe/Istanbul  
**Rol:** UYUM (KAN.US veri koruma) — yalnızca okuma + yerel tasarım/yama; canlı silme/yayın/yazma yok  
**Kaynak öncelik:** `/workspace/us-panel-dev/work/api/r10_base` (analiz), çalışma kopyası `/workspace/us-panel-dev/work/uyum/r10_privacy`  
**Prompt:** GLOBAL_PROMPT §7.7, §12 (uyum kısmı), §14, §15; P3 ticari güvenlik  
**Hukuki uyumluluk bu raporla tamamlanmış ilan edilmez.**

---

## 1. Kısa özet

US Panel / us-license-hook kod tabanında hesap, lisans, outbox ve model proxy için **kişisel veri işleniyor**; ancak KVKK m.11 / GDPR erişim-silme-taşınabilirlik-rıza için **üretimde hazır uç noktalar yoktu**. SHOP `docs/legal/final/` taslakları bulundu; çoğu alan `<DOLDUR>`; canlı mağazada KVKK/iade/şartlar sayfaları eksik (`LIVE_STORE_AUDIT_20261106.md`).

Bu turda: boşluk analizi + aktarım envanteri + hak↔kod tablosu yapıldı; ardından **ayrı paket** (`internal/privacy` + `licensesrv/privacy_ops.go`) ile dışa aktarma, anonimleştirme (defter korunur), TTL temizlik ve tercih kaydı **yerel kopyada** kodlandı ve test edildi. Yamalar: `/workspace/us-panel-dev/patches/uyum-*`. PANEL DEV incelemesi beklenir. Canlıya uygulanmadı.

| Alan | Durum | Risk |
|---|---|---|
| Veri dışa aktarma | **Yok** → yama taslağı Var (yerel) | Yüksek |
| Silme / unutulma | **Yok** → anonimleştirme yaması (yerel) | Yüksek |
| Saklama politikası / TTL | **Kısmi** (oturum TTL, yedek 7–14g; uygulama TTL yok) → sweep yaması | Yüksek |
| Kullanıcı izolasyonu | **Kısmi** (LicenseOwners, session) | Orta |
| Gizlilik tercihleri / rıza | **Yok** → prefs yaması (yerel); UI yok | Yüksek |
| Veri aktarım envanteri | **Belirsiz** (teknik ipuçları var, hukuki mekanizma yok) | Yüksek |
| Güvenli yedekleme | **Kısmi** (flock+chmod; şifreleme yok; env yedekte) | Yüksek |
| Loglarda PII/sır | **Kısmi/Problem** (dry-run e-posta log; outbox’ta anahtar) | Yüksek |
| Denetim kayıtları (privacy) | **Yok** | Orta |

---

## 2. Kişisel veri envanteri (teknik)

Her satır: **ne / nerede / süre (kod veya tahmin)**.

| Veri | Nerede | Süre | Kanıt |
|---|---|---|---|
| E-posta, parola bcrypt, hesap id, verify/reset hash | `licenses.json` → `accounts` | Belirsiz (silme API yoktu) | `licensesrv/accounts.go:24-35` |
| Oturum JWT-benzeri (`usa1.…`) | İstemci + sunucu doğrulama; sunucuda kalıcı değil | `US_SESSION_TTL_SECONDS` varsayılan 86400s | `cmd/us-license-hook/main.go:94-98`, `accounts.go:78-88` |
| Lisans anahtarı, sipariş e-postası, orderId, credits | `licenses` map | Plan `ExpiresAt` + sonrası belirsiz | `store.go:29-41`, `webhook.go:162-208` |
| LicenseOwners (hesap↔lisans) | `licenseOwners` | Belirsiz | `store.go:81` |
| Outbox: To, Subject, Body (**lisans anahtarları gövdede**) | `outbox[]` | Gönderilene/dead’e kadar; TTL yoktu | `reversals.go:105-106`, `mail.go` |
| Webhook idempotency id | `webhooks` | Belirsiz | `store.go:82` |
| Rezervasyon / settlement / refund | ilgili map’ler | Belirsiz (mali defter) | `store.go:86-89` |
| IP (hız sınırı) | Bellekte `rates` map; **diske yazılmaz** | ~15 dk pencere | `accounts.go:250-283` |
| Model istem içeriği | Upstream’e proxy; yerel store’da mesaj arşivi **yok** (gözlem) | Upstream politikası **Belirsiz** | `model_proxy.go` |
| SMTP kimlik bilgileri | `/etc/us-license-hook.env` (VPS); yedek 600 | Yedek KEEP_DAYS | `mail.go:16-18`, `PROD_PLAN.md` B5, `backup-*.sh` |
| Gizlilik tercihleri | Yoktu → sidecar `privacy-prefs.json` (yama) | Kullanıcı güncelleyene kadar | `internal/privacy/prefs.go` |

**Tahmin:** Model GPU (`ds-a100`, `europe-west4-b`) üzerinde istem loglanıyor olabilir — kodda US Panel store’a yazılmıyor; **doğrulanmadı**.

---

## 3. Alan bazlı boşluk analizi

### 3.1 Veri dışa aktarma — **Yok** (önce) / yama hazır
- **Kanıt (önce):** Bilinen hesap rotaları export içermiyor: `accounts.go:304`.
- **Risk:** Yüksek (KVKK m.11 öğrenme/bilgi; GDPR Art.15/20 teknik karşılık yok).
- **Çözüm (yerel yama):** `ExportPrivacy` + `POST /account/privacy/export` — `patches/uyum-export/`.

### 3.2 Silme talebi — **Yok** (önce) / anonimleştirme yaması
- **Kanıt:** Delete/erase route yok (`accounts.go:304`).
- **Risk:** Yüksek. Yedeklerdeki kopyalar ayrıca kalır (`RESTORE.md`, hourly 7g / daily 14g).
- **Çözüm:** `AnonymizeAccount` — defter satırları korunur, PII psevödonim. `patches/uyum-erase/`.  
- **HUKUKİ İNCELEME GEREKİR:** Anonimleştirme vs yok etme; mali saklama süreleri; yedek imha prosedürü.

### 3.3 Saklama politikası — **Kısmi**
- Var: oturum TTL; yedek `KEEP_DAYS` 7 (hourly taslak) / 14 (canlı daily — `RESTORE.md:13-14`); journald `SystemMaxUse=500M` (`PROD_PLAN.md:82`).
- Yok: uygulama içi outbox/webhook/account TTL job (önce).
- **Çözüm:** `RetentionSweep` + admin `retention-run`. `patches/uyum-retention/`.
- **HUKUKİ İNCELEME GEREKİR:** hesap/sipariş/log/destek süreleri (`KVKK_AYDINLATMA` §4 `<DOLDUR>`).

### 3.4 Kullanıcı veri izolasyonu — **Kısmi**
- Var: oturum HMAC; bakiye/geçmiş `LicenseOwners` ile hesap id’ye bağlı (`accounts.go:567-604`, `618-646`); model `authorizedCredential`.
- Masaüstü scope_guard (proje yolu) ayrı güvenlik katmanı — çok kiracılı bulut FS izolasyonu bu hook’ta yok.
- Risk: Orta. Tek JSON store; dosya düzeyi tenant ayrımı yok (PROD_PLAN staging/prod store ayrımı planı var).

### 3.5 Gizlilik tercihleri / onay — **Yok** (önce)
- Pazarlama/çerez rızası kodda yok; kayıt = sözleşme ifası varsayımı.
- Canlı checkout açık onay UI yok (`LIVE_STORE_AUDIT`).
- **Çözüm:** prefs sidecar + API. UI/Shopify **ayrı**. `patches/uyum-preferences/`.

### 3.6 Veri aktarım envanteri — **Belirsiz** → §4 tablo
### 3.7 Güvenli yedekleme — **Kısmi**
- Var: flock, mode 640/600, hourly taslak, restore runbook (`deploy/host-hardening/backup/`).
- Yok: yedek şifreleme (at-rest), yedekten PII silme/erase senkronu, offsite politikası.
- Risk: Yüksek (env yedekte sır; licenses.json’ta e-posta+anahtar+hash).

### 3.8 Loglarda kişisel veri / sır sızıntısı — **Problem (kısmi)**
| Bulgu | Kanıt | Not |
|---|---|---|
| Dry-run gönderici e-posta + konu loglar | `mail.go:66-68` | Üretimde SMTP açıksa dry-run kullanılmaz; staging riski |
| Outbox Body’de ham lisans anahtarı | `reversals.go:105-106` | Store + yedek = sır benzeri |
| Webhook hata logu topic+err | `webhook.go:126` | E-posta yazılmıyor (iyi) |
| Model upstream token header’da | `model_proxy.go:163-164` | Log’a yazılmıyor (gözlem) |
| `.env` / Secret dosyaları | `SMTP_HOSTINGER` raporunda şifre yolu anıldı | Rapora sır **yazılmadı**; varlığı: Hostinger SMTP secret dosyası / VPS env |

**Öneri:** dry-run’da e-posta maskele; outbox body’de anahtar yerine short-ref; log scrubber.

### 3.9 Denetim kayıtları — **Yok**
Privacy export/erase için audit trail yok (yama da eklemedi). **Öneri:** sonraki yama `privacy-audit.jsonl`.

---

## 4. Veri aktarım envanteri

| Sistem | Veri türü | Bölge/ülke (teknik ipucu) | Amaç | Kanıt | Hukuki mekanizma |
|---|---|---|---|---|---|
| Shopify (shop.kanus.org / 19einb-qz) | Sipariş, müşteri e-posta, product_id | Shopify altyapısı (çoğunlukla yurt dışı — **tahmin**) | Sipariş/webhook | `webhook.go`, `LIVE_STORE_AUDIT` | `<DOLDUR>` / **HUKUKİ İNCELEME** |
| Ödeme sağlayıcı | Kart/ödeme | **UNKNOWN** (API scope yok) | Tahsilat | `LIVE_STORE_AUDIT` §1 | `<DOLDUR>` |
| Hostinger VPS + SMTP | licenses.json, e-posta teslimi (To, lisans gövdesi) | VPS IP 31.220.63.191 (staging); SMTP `smtp.hostinger.com` | Barındırma / mail | `PROD_PLAN.md`, `SMTP_HOSTINGER_20261106.md`, `mail.go:16` | `<DOLDUR>` |
| GCP Cloud Run `us-commerce` / `us-watchdog` | Ticari/gateway trafik (kapsam doğrulanmalı) | **europe-west4** | Yönetilen çıkarım / watchdog | `SRE_AUDIT_20261106.md:22-24`, `kanus/deploy/README.md:49` | m.9 `<DOLDUR>` |
| GCP GCE `ds-a100` | Model istem/yanıt (proxy üzerinden) | **europe-west4-b** | GPU çıkarım | `SRE_AUDIT:18`, `main.go:126` default model adı | `<DOLDUR>` |
| Diğer GCP VM’ler | Operasyonel | us-central1-* | Eski/yan hizmetler | `SRE_AUDIT:19-21` | Ayrı değerlendirme |
| Model adı / upstream | `deepseek-v4.1-flash` varsayılan; `US_MODEL_UPSTREAM` | Upstream URL ortama bağlı — **Belirsiz ülke** | LLM | `main.go:126` | DeepSeek/sağlayıcı DPA **HUKUKİ İNCELEME** |
| US masaüstü istemci | Oturum, lisans, proje dosyaları (yerel) | Kullanıcı cihazı | UX | `account_client.go` (kiyas) | Yerel işleme |

«Aktarım yoktur» denilemez (aydınlatma taslağıyla uyumlu).

---

## 5. Hak ↔ metin ↔ kod tutarlılık tablosu

Kaynak haklar: `docs/legal/final/KVKK_AYDINLATMA_METNI.md` §5 / ORG taslak §5 (m.11); `GIZLILIK_POLITIKASI.md` «Haklarınız».

| Hak | Metin | Kod kanıtı (önce → sonra yerel) | Durum |
|---|---|---|---|
| İşlenip işlenmediğini öğrenme | KVKK §5 | Export yoktu → `ExportPrivacy` / `/account/privacy/export` | **Kısmi** (yama yerel; canlı yok; başvuru kanalı `<DOLDUR>`) |
| Bilgi isteme / amaç | KVKK §2–5 | Aydınlatma taslak; kodda amaç metni yok | **Belirsiz** — metin + hukuk |
| Yurt içi/dışı alıcıları bilme | KVKK §3; Gizlilik «Paylaşım» | Envanter teknik; metinde `<DOLDUR>` | **Yok** (tamamlanmamış metin) |
| Düzeltme | m.11 | E-posta değiştirme API yok | **Yok** |
| Silme / yok etme (m.7) | m.11; Gizlilik | Yoktu → anonimleştirme yaması; yok etme/yedek yok | **Kısmi** |
| Düzeltme/silmenin alıcılara bildirimi | m.11 | Shopify/Hostinger’a bildirim otomasyonu yok | **Yok** |
| Taşınabilirlik (GDPR 20) | Gizlilik dolaylı | Export JSON yaması | **Kısmi** |
| İtiraz / otomatik analize itiraz | m.11 | Endpoint yok | **Yok** |
| Zararın giderilmesi | m.11 | Hukuki süreç | **HUKUKİ** (kod değil) |
| Pazarlama rızasını ayırma | KVKK giriş: birleştirilmez | Prefs API yaması; checkout UI yok | **Kısmi** |
| Saklama süresi şeffaflığı | KVKK §4 | Metin `<DOLDUR>`; sweep varsayılan 90g kod | **Belirsiz** |
| Başvuru 30 gün | KVKK §5 | Operasyonel SLA kodda yok | **Yok** |

---

## 6. SHOP legal taslak tutarlılığı

**Bulundu:** `/workspace/us-panel/docs/legal/final/` (7 metin + README). ORG taslaklar `docs/legal/*.md` korunmuş.

### 6.1 Canlı mağaza vs taslak
`LIVE_STORE_AUDIT_20261106.md`: canlıda yalnız jenerik PRIVACY_POLICY; iade/şartlar/KVKK/mesafeli/ön bilgilendirme **yok**; iletişim sayfası boş.

### 6.2 Metin vaadi vs kod (çelişki / boşluk)
| Metin iddiası | Kod gerçeği |
|---|---|
| m.11 hakları kullanılabilir | Başvuru API/kanal yok; export/erase yalnız yerel yama |
| Saklama süreleri | `<DOLDUR>`; kodda hesap TTL yok |
| Yurt dışı aktarım açıklanacak | Teknik europe-west4 + Deepseek varsayılan; metin boş |
| Pazarlama ayrı rıza | Prefs yaması var; mağaza UI yok; kayıt formu rıza istemiyor |
| Güvenlik önlemleri | TLS/HMAC/bcrypt var; yedek şifreleme yok; outbox’ta anahtar |
| Veri sorumlusu kimliği | Tüm final metinlerde unvan/adres/VKN `<DOLDUR>` |

### 6.3 `<DOLDUR>` özeti (final)
Ortak: Unvan, Adres, VKN/MERSİS, Telefon, KEP, yasal/KVKK başvuru kanalı, ödeme kuruluşu, barındırma/model aktarım, saklama süreleri, eğitimde kullanım, m.5/2-f denge, checkout açık onay metni. Ayrıntılı liste: `docs/legal/final/README.md` «`<DOLDUR>` alanları».

---

## 7. Tasarım (uygulanan yerel yama)

```
internal/privacy/          # HTTP + prefs sidecar (accounts.go’ya dokunulmaz)
  prefs.go                 # marketing/product/cookie flags
  handler.go               # export|erase|preferences|retention-run
licensesrv/privacy_ops.go  # ExportPrivacy, AnonymizeAccount, RetentionSweep
cmd/us-license-hook/main.go # mux wiring only
```

**Anonimleştirme kuralları:** defter (orders, settlements, reservations, license keys, owners) kalır; e-posta/parola/outbox PII/psevödonim; oturum version invalidation.

**Ortam:** `US_PRIVACY_PREFS_PATH`, `US_PRIVACY_RETENTION_DAYS` (varsayılan 90), `US_PRIVACY_ADMIN_TOKEN`.

---

## 8. Yamalar ve test sonuçları

| Dizin | Kapsam |
|---|---|
| `patches/uyum-export/` | Dışa aktarma ops + test + handler |
| `patches/uyum-erase/` | Anonimleştirme (shared ops) |
| `patches/uyum-retention/` | TTL sweep |
| `patches/uyum-preferences/` | Prefs store + handler test |
| `patches/uyum-wiring/` | main.go mux |

**Test komutları / sonuç (2026-11-06 Europe/Istanbul, çalışma kopyası):**
```
go test ./licensesrv/ -count=1 -run Privacy   → ok
go test ./internal/privacy/ -count=1          → ok
go test ./cmd/us-license-hook/ -count=1       → ok
```
Kaynak ağaç `/work/api/r10_base` **değiştirilmedi**.

---

## 9. Önceliklendirilmiş yama / sonraki iş planı (PANEL DEV)

| Öncelik | İş | Not |
|---|---|---|
| P0 | `uyum-wiring` + `uyum-export` + `uyum-erase` incele/merge | API bot hesap dosyalarıyla çakışma yok |
| P0 | Outbox body’den ham anahtar kaldırma / dry-run e-posta maskeleme | Ayrı güvenlik yaması |
| P0 | SHOP legal `<DOLDUR>` + hukukçu + canlı sayfa (SHOP bot) | Bu rapor kodu değil |
| P1 | `uyum-retention` + cron (HOST) | Admin token secret manager |
| P1 | `uyum-preferences` + checkout/çerez UI | SHOP |
| P1 | Yedek şifreleme + erase sonrası yedek politikası | HOST + hukuk |
| P2 | Privacy audit log | |
| P2 | Düzeltme (e-posta değişimi) API | |
| P2 | Aktarım DPA / SCC envanteri doldurma | Hukuk |

---

## 10. Hukuki inceleme gereken noktalar

1. Veri sorumlusu kimliği ve KVKK başvuru kanalı (`support@kanus.org` yetkisi).  
2. Yurt dışı aktarım m.9 mekanizması: Shopify, Hostinger, GCP europe-west4, DeepSeek/upstream.  
3. Ödeme kuruluşu ve kart verisi rolü (Shopify Payments / alternatif — canlı UNKNOWN).  
4. Anonimleştirmenin m.7 / mali saklama ile yeterliliği; yedek imha.  
5. Meşru menfaat denge testleri (IP rate limit, güvenlik logları).  
6. Model istemlerinin eğitimde kullanımı ve saklanması.  
7. Dijital anında ifa için checkout açık onay metni ve kayıt yöntemi.  
8. Aydınlatma/gizlilik final metinlerinin yayın öncesi onayı.  
9. GDPR uygulanabilirliği (AB kullanıcıları) ve temsilci.  
10. Unutulma talebinde Shopify müşteri kaydı silme süreci.
11. **Kullanılmamış süre iadesi (yeni, 2026-11-06 kararı sonrası):** Demir'in kararıyla (KAN.US aracılığıyla, 09:58) anonimleştirmede aktif lisanslar iptal ediliyor (uyum-v3); otomatik para iadesi yapılmıyor (`refundTriggered:false`). Mesafeli sözleşme / tüketici mevzuatı açısından kullanılmamış süre iadesi gerekip gerekmediği ve kullanıcıya önceden yapılacak bilgilendirme metni — **HUKUKİ/TİCARİ İNCELEME GEREKİR**.
12. **Şirket bilgileri (açık):** Veri sorumlusu unvanı, MERSİS/vergi no, adres, KEP, VERBİS kayıt durumu ve KVKK başvuru kanalı Demir'den bekleniyor; SHOP final metinlerindeki `<DOLDUR>` alanları bunlara bağlı.
13. **Çerez banner'ı (açık):** Canlı mağaza rıza olmadan `_shopify_analytics` ve `_shopify_marketing` set ediyor (2026-11-06 curl). Shopify Customer Privacy ayarı Demir onayı bekliyor (`docs/legal/uyum-drafts/CEREZ_BANNER_ADIMLAR.md`); Türkiye bölge seçimi, `localization` çerezinin sınıfı ve banner metni hukukçu onayı istiyor.
14. **.14 görsel/video üretimi (yeni):** CSAM olayında kanıt saklama ve bildirim; AB YZ Yasası m.50 (filigran/C2PA) kapsamı ve güncel uygulama durumu; anlık yüz tespitinin biyometrik veri sayılıp sayılmadığı; model lisanslarının ticari kullanım/bölge kısıtları ve OpenRAIL aktarımı; moderasyon logu ve legal hold süreleri. Ayrıntı: `reports/UYUM_ICERIK_GUVENLIGI_14_20261108.md`.

_Son güncelleme: 2026-11-08 01:30 (UTC+3)._

---

## 11. GLOBAL_PROMPT hizası

- §7.7 maddeleri tarandı; eksikler yama ile teknik olarak ele alındı (hukuki tamamlanma yok).  
- §12: kullanıcı izolasyonu kısmi; yedekleme kısmi; log sızıntısı işaretlendi.  
- §14: canlı silme/yayın/e-posta **yapılmadı**.  
- §15 / P3: bu rapor `KANUS_PRIVACY_REPORT.md` olarak üretildi.

---

## 12. Referans dosyalar

- `docs/legal/final/*`, `reports/LIVE_STORE_AUDIT_20261106.md`
- `deploy/host-hardening/PROD_PLAN.md`, `backup/*`
- `work/api/r10_base/licensesrv/{accounts,store,mail,webhook,model_proxy,reversals}.go`
- `reports/SRE_AUDIT_20261106.md`, `SMTP_HOSTINGER_20261106.md`
- Çalışma kopyası: `us-panel-dev/work/uyum/r10_privacy`
- Yamalar: `us-panel-dev/patches/uyum-*`
