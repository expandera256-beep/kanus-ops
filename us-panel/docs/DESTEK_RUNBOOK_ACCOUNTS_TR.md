# US Panel — Hesap (accounts) modu destek runbook’u

> **TASLAK, yayımlanmadı** — 6 Kasım 2026 (Europe/Istanbul, UTC+3)  
> Yalnızca yerel box dosyası. Canlı wiki / e-posta / Slack yayını yapılmamıştır.  
> Ton: marka kararı A — sakin, net, korkutmadan.  
> Kaynak: ROUND7, ACCOUNTS_DEPLOY_TEST, ACCOUNTS_MODE_GAP, `licensesrv/accounts.go`, `credits.go`, `model_proxy.go`, `reversals.go`, EXE istemcisi.

---

## Destek ilkeleri (zorunlu)

1. **Lisans anahtarı (`us1.…`), oturum token’ı (`usa1.…` / session), parola, `US_SESSION_SECRET`, SMTP veya Shopify webhook secret istemeyin ve kaydetmeyin.**
2. Kullanıcıdan isteyebilecekleriniz: e-posta (hesap adresi), yaklaşık işlem zamanı (TSİ), panel sürüm metni, görünen hata cümlesi, sipariş numarası (Shopify görünen ad — anahtar değil).
3. “Hemen yeniden satın alın” demeyin; önce bakiye, sahiplik ve iptal/iade durumunu netleştirin.
4. Sahte başarı yok: workers yanıtında `live:false` / `provider_status:unknown` “GPU hazır” demek değildir.

Destek kutusu (rapor): `support@kanus.org`.

---

## Mimari özeti (destek için)

| Bileşen | Not |
|---------|-----|
| Panel | Windows EXE, `releaseAuthority=accounts` |
| Geçit | Go **us-license-hook** — test `api-test.kanus.org`, üretim hedefi `api.kanus.org` |
| Lisans üretimi | Shopify `orders/paid` → `us1.…` anahtar (tek ledger) |
| Hesap | E-posta + bcrypt parola; redeem yalnız sahiplik bağlar, kredi kopyalamaz |
| Kullanım | Sunucu reserve / settle; saatlik paketlerde saniye ölçümü |
| İptal / iade | Shopify webhook → lisans `Revoked`; hesap bakiyesi kaynak lisanslardan türetilir |

---

## Senaryo A — Şifre unutma

**Kaynak durumu**

- Sunucu uçları: `POST /account/password-reset/request`, `POST /account/password-reset/confirm` (`licensesrv/accounts.go`).
- İstek e-posta bilinmese bile genel **202** dönebilir (bilgi sızdırmaz).
- Onay: 32 karakter kod + yeni parola (12–72 bayt); başarıda `sessionVersion` artar → eski oturumlar geçersiz.
- **Panel GUI’sinde sıfırlama ekranı yoktur.** Reset/verify GUI kontrolleri NOT TESTED.
- SMTP allowlist / teslimat canlı kabulü her ortamda doğrulanmış sayılmaz.

**Kullanıcıya söyleyin**

> Panelde henüz “parolamı unuttum” ekranı yok. Destek ekibi parolanızı veya anahtarınızı istemez. Hesap e-postanızı ve yaklaşık kayıt zamanınızı yazın; süreç manuel / karar aşamasındadır.

**Destek adımı**

1. E-postayı not alın; parola/anahtar istemeyin.
2. Operasyon ekibine iletin: reset API’si var mı kullanılacak, SMTP allowlist’te mi, yoksa elle geçici süreç mi?
3. Kaynakta “destek panelinden parola yazma” aracı yoktur. Karar bekliyor → açık sorular.

---

## Senaryo B — Yanlış hesaba redeem

**Kaynak durumu**

- `bindLocked`: anahtarın `LicenseOwners` kaydı başka `accountID` ise `errBound` → HTTP **409**.
- Aynı hesaba tekrar redeem **200** (idempotent); kredi çoğalmaz.
- Otomatik “hesaplar arası taşı” yok.

**Kullanıcıya söyleyin**

> Bu anahtar başka bir hesaba bağlı görünüyor. Panel ikinci hesaba aktaramaz. Anahtarı paylaşmadan sipariş bilgisi ve iki e-posta adresini (doğru / yanlış) yazın.

**Destek adımı**

1. Hangi e-postanın “doğru” olduğunu teyit edin (sipariş e-postası ile çapraz kontrol — anahtarsız).
2. Ledger’da sahiplik değişimi için onaylı operasyon prosedürü **kaynakta tanımlı değil** → karar bekliyor.
3. Kullanıcıya yeniden satın alma önermeden önce operasyon yanıtını bekleyin.

---

## Senaryo C — İptal ve iade sonrası kredi

**Kaynak durumu** (`reversals.go`)

- `orders/cancelled`: sipariş tombstone + ilgili lisanslar `revokeLocked`.
- `refunds/create`: satır miktarı kadar birim lisans iptal; finansal-only iade kredi dönüşümü **uydurulmaz**.
- Hesap bakiyesi bağlı, iptal edilmemiş, süresi dolmamış lisansların kullanılabilir toplamıdır; iptal edilen lisans bakiyeye katkı yapmaz.
- Model geçmişindeki `refunded` durumu çağrı rezervasyonunun serbest bırakılmasıdır; kart iadesi değildir.

**Kullanıcıya söyleyin**

> Mağaza iptali veya iadesi sonrası ilgili paket kullanıma kapanır. Panel bakiyesi buna göre düşer veya sıfırlanır. Kart iadesi Shopify/ödeme sürecindedir; panel geçmişindeki “Kredi iade edildi” satırı banka iadesi demek değildir.

**Destek adımı**

1. Sipariş numarası / zaman (anahtarsız).
2. Operasyondan webhook’un işlenip işlenmediğini ve hangi lisansların `Revoked` olduğunu sorun (destekçi secret/anahtar görmesin).
3. Kullanıcıya bakiyeyi yenilemesini söyleyin. Kısmi finansal iade ↔ kredi oranı için ürün kararı gerekebilir.

---

## Sık HTTP kodları (accounts / model)

Sunucu hesap hatalarında gövde genelde `{"error":"account request rejected"}` şeklindedir; kod ayırt edicidir. EXE çoğu durumda genel Türkçe cümleye çevirir.

### 401 — Yetkisiz / oturum veya kimlik reddi

| | |
|--|--|
| **Anlamı** | Yanlış e-posta/parola; geçersiz veya süresi dolmuş oturum; Bearer yok/bozuk; ham lisans geçersiz. |
| **Kodda nereden** | Login: `accounts.go` (~411, ~418). Oturumlu yollar: ~519. Kredi durumu session: `credits.go` ~30–35. Model proxy lisans: `model_proxy.go` ~54. Deploy test: yanlış parola / sahte token → 401. |
| **Kullanıcıya** | “E-posta ve parolayı kontrol edin; gerekirse çıkıp yeniden giriş yapın. Anahtar veya parolayı destek mesajına yazmayın.” |
| **Destek** | Çok deneme 429’a yol açabilir; bekletın. Reset GUI yok. Session secret veya token istemeyin. |

### 403 — Yasak / politika reddi

| | |
|--|--|
| **Anlamı** | İstek kimlikli olabilir ama işlem bu kimlik/plan için izinli değil. Örn. saatlik saniye ölçümü için legacy desktop reserve yolu kapalı; rezervasyon geçersiz / artık aktif değil; kredi rezervasyonu `ErrInvalid`. |
| **Kodda nereden** | `credits.go` ~56 (hourly → “server model proxy” gerekir); ~63. `model_proxy.go` ~115, ~122. |
| **Kullanıcıya** | “Bu işlem şu an hesabınız veya paket türünüz için uygun değil. Giriş yaptığınızdan ve bakiyenizin yeterli olduğundan emin olun; sorun sürerse sürüm ve saat bilgisini yazın.” |
| **Destek** | Accounts normal akışında EXE desktop reserve çağırmaz. Kullanıcı eski lisans URL karışımı kullanıyorsa gateway’i netleştirin. Anahtar istemeyin. |

### 409 — Çakışma

| | |
|--|--|
| **Anlamı** | E-posta zaten kayıtlı; lisans başka hesaba ait / bağlanamaz; settle çakışması. |
| **Kodda nereden** | Register e-posta: `accounts.go` ~390. Redeem sahiplik: ~541 + `bindLocked` ~203–211. Settle: `credits.go` ~83. Deploy: ikinci hesap aynı lisans → 409. |
| **Kullanıcıya** | “Bu e-posta veya lisans başka bir kayıtla ilişkili. Yeniden satın almadan destekle iletişime geçin.” |
| **Destek** | Senaryo B. Settle 409 ise yarım kalmış çağrı olabilir; bakiyeyi yenilemesini söyleyin. |

### 429 — Çok fazla istek

| | |
|--|--|
| **Anlamı** | IP başına login/register/reset kotası; bcrypt eşzamanlılık kapısı; model kuyruğu dolu; nginx login/register `5r/m` (test VPS). |
| **Kodda nereden** | `accounts.go` ~336, ~350. Model kuyruk: `model_proxy.go` ~97. Deploy: paralel login burst → 429. |
| **Kullanıcıya** | “Bir süre bekleyip tek deneme yapın. Giriş sunucuda yavaş olabilir; peş peşe tıklamayın.” |
| **Destek** | Servis çökmesi sanmayın. VPS yükü yüksekse login 20–30 sn sürebilir; 504 görürse nginx timeout operasyon konusudur. |

### 503 — Servis hazır değil / geçici yok

| | |
|--|--|
| **Anlamı** | Hesap işleyici hazır değil (ör. session secret yetersiz); dahili kod üretimi/token hatası; workers/metadata yok; model upstream yapılandırılmamış. |
| **Kodda nereden** | `accounts.go` ~300 (ready değil), ~363–368, ~424, ~525. `account_workers.go` ~44. `model_proxy.go` ~58, ~415. |
| **Kullanıcıya** | “Hesap veya model servisi şu an yanıt veremiyor. Birkaç dakika sonra yeniden deneyin; sürmesi halinde saati ve gördüğünüz metni yazın.” |
| **Destek** | Operasyona healthz / hook durumu sorun. Secret değerini sohbette istemeyin veya yapıştırmayın. |

### (Ek) 402 — Yetersiz kredi

Accounts chat yolunda rezervasyon reddinde görülebilir (`model_proxy.go` ~112; load test). Kullanıcıya: bakiyeyi yenileyin; saatlik saniye ile prompt kredisini karıştırmayın.

### (Ek) 400 — Geçersiz istek

Kısa/uzun parola, bozuk e-posta, boş redeem anahtarı (`accounts.go` ~343, ~533). Kullanıcıya alanları ve 12+ karakter parolayı hatırlatın.

---

## Hızlı teşhis listesi

1. Kullanıcı accounts EXE mi kullanıyor? (Eski go-hook profilinde hesap formu gizlidir.)
2. Geçit adresi test mi üretim mi? Cloud Run adresi kullanılmamalı.
3. Giriş var mı? Redeem/balance giriş ister.
4. Görünen metin mi, HTTP kodu mu? (EXE çoğu zaman kodu göstermez.)
5. Shopify iptal/iade yakın zamanda mı oldu?
6. Soğuk açılış / GPU metni mi — bu bakiye hatası değildir; 4–28 dk bekleyiş normal olabilir.

---

## Açık sorular

1. Destek için onaylı “parola sıfırlama” işletim prosedürü (API mi, manuel mi, kim çalıştırır?).
2. Yanlış hesap sahipliğini düzeltme: ledger’da owner değiştirme yetkisi ve denetim kaydı.
3. Müşteri SMTP allowlist politikası (üretimde kimlere mail gider?).
4. 504 / yavaş login için kullanıcıya resmi bekleme süresi metni.
5. Destek saatleri, telefon, iç escalation kanalı.
6. Kart iadesi ile lisans revoke arasındaki SLA ve kullanıcı bilgilendirme şablonu (hukuk taslağı ayrı).

