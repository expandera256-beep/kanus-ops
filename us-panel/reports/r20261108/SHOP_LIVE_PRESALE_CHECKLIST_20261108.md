# shop.kanus.org — satış öncesi salt-okuma kontrolü

**Kontrol:** 8 Kasım 2026, 01:20–01:24 TSİ (Europe/Istanbul). **Karar: SATIŞA HAZIR DEĞİL.**

Yalnız herkese açık HTTPS GET ve mevcut yerel raporlar okundu. Shopify connector kullanılmadı; test mağazası bağlantısı, auth, tema, ürün, ödeme, webhook ve sunucu ayarları değiştirilmedi. Sepete ürün eklenmedi, checkout oluşturulmadı, sipariş verilmedi. Yalnız bu rapor ve yerel GET kanıt dosyaları üretildi.

## Kontrol listesi

| Kontrol | Durum | Kanıt / eksik | Sorumlu / sonraki adım |
|---|---|---|---|
| Ödeme sağlayıcısı canlı ve aktif | UNKNOWN | `/cart` 200 fakat boş; `/checkouts` 404. Ana sayfa/sepet HTML içinde ödeme ikonları veya doğrulanabilir aktif hızlandırılmış ödeme düğmesi yok. Genel PaymentButton scripti sağlayıcı kanıtı değildir. 404 checkout kapalı demek değildir. | Demir: Admin → Ödemeler; aktif sağlayıcı, canlı/test modu, TRY tahsilatını doğrula. KAN.US kabul kapısı. |
| Checkout e-posta zorunlu | UNKNOWN | Checkout açılmadı. `/pages/contact` formunda e-posta zorunlu olması checkout kanıtı değildir. 6 Kasım raporu yalnız misafir checkout açık bilgisini veriyor. | Demir: Admin → Checkout → müşteri iletişim yöntemini e-posta zorunlu olacak şekilde kontrol et. |
| Ürün fiyatları / yayın | UNKNOWN (kısmi PASS) | Aylık 18.900 / 200 istem, saatlik 1.990, 10 kredi 1.290 ve 100 kredi 9.900 TL ana sayfa/sepet HTML ile uyumlu. 5/25/50 kredi güncel doğrulanamadı; ürün JSON ve koleksiyon GET'leri HTTP429. Aşağıdaki tabloya bak. | Demir Admin; SHOP ancak connector canlıya ayrı yetkilendirildiğinde salt-okuma teyit. Bu görevde switch/auth yok. |
| 1 kredi satıştan gizli / DRAFT | UNKNOWN (tarihsel PASS) | 6 Kasım raporunda DRAFT, Online Store isPublished=false. Bugünkü ana sayfada görünmüyor ama bu tek başına doğrudan URL/checkout üzerinden alınamadığını kanıtlamaz. | Demir: ürünün DRAFT ve tüm kanallarda satış dışı olduğunu teyit et; silme yok, mapping korunur. |
| Dijital ürünlerde requiresShipping=false | UNKNOWN (tarihsel PASS) | 6 Kasımda 8/8 false. Bugünkü aylık/saatlik JSON-LD açıklaması fiziksel gönderim yok diyor; Admin varyant alanının güncel kanıtı değildir. JSON uçları 429. | Demir / daha sonra canlı bağlantılı SHOP. |
| TRY para birimi | PASS | Ana sayfa `Shopify.currency.active=TRY`; aylık/saatlik JSON-LD priceCurrency=TRY. | İş yok. |
| KDV dahil gösterim / vergi oranı | UNKNOWN | 6 Kasım `taxesIncluded=true`, taxable=true; bugünkü ana sayfa fiyat satırlarında açık KDV dahil ifadesi yok. Oran, vergi profili ve checkout toplamı doğrulanmadı. | Demir + mali müşavir: Vergiler ve ürün vergileri; KAN.US tema metni. |
| Production API / canlı lisans teslimi | FAIL | Google DNS bugün api.kanus.org A=82.29.184.225; **DNS artık var**. Ancak GET `https://api.kanus.org/healthz` HTTP502 nginx/1.28.3. Önceki “alan adı yok” bilgisi güncel değil; production sağlık kabulü yok. | HOST: production servis/upstream/TLS sağlık ve canlı plan eşlemesini doğrula; KAN.US koordine eder. |
| Canlı webhook hedefleri | FAIL (kesim yapılmadı); güncel Admin URL'leri UNKNOWN | Kesim planı uygulanmamış; E2E nihai raporu canlı cutover yok diyor. Verilen operasyon bağlamı canlı bildirimlerin test sunucusuna gittiğini belirtiyor; bugünkü Admin listesi okunmadığından bunu yeni teknik teyit gibi sunmuyoruz. MCP geçmiş 0 sonucu yalnız uygulama kapsamlıdır. | HOST + Demir: orders/paid, orders/cancelled, refunds/create hedeflerini Admin'de doğrula; ayrı canlı secret/depo ve plans.live.json; yalnız onaylı geçiş. |
| Gerçek müşteri lisans e-postası | FAIL (üretim kabulü) | Test #1003 outbox kaydı allowlist nedeniyle skipped. Bu gerçek SMTP teslimi PASS değildir. Hesap/panel görünürlüğü UNKNOWN. | HOST SMTP + KAN.US uçtan uca kabul; yalnız onaylı alıcıyla doğrulama. Idempotency Demir kararıyla SKIPPED; yeniden test talebi yok. |
| İade/cayma ve hizmet şartları | FAIL | `/policies/refund-policy` ve `/policies/terms-of-service` HTTP404. Footer yalnız gizlilik politikasına bağlanıyor. | ORG/UYUM hukuk incelemesi + Demir satıcı bilgileri/onayı; Demir Admin'de yayınlar veya sonra yetkili SHOP. |
| Gizlilik / KVKK | FAIL (KVKK kapsamı); gizlilik sayfası erişimi PASS | `/policies/privacy-policy` 200, 6 Kasım tarihli genel Shopify metni. Ayrı KVKK, aktarım/saklama ve veri sorumlusu kapsamı doğrulanmış değil. Gizlilik metni iletişim bilgileri politikasına yönlendiriyor ama o URL 404. | ORG/UYUM + Demir: gerçek veri envanteri, aktarım mekanizması ve başvuru kanalı; hukukçu kabulü. |
| Mesafeli satış / ön bilgilendirme | FAIL (yayın/erişim kanıtı yok) | Footer'da bağlantı yok; 6 Kasım sayfa listesinde yok; final klasörü yalnız yerel taslak. Siparişe özgü ön bilgilendirme ve dijital erken ifa onay kaydı doğrulanmadı. Tüm olası özel URL'ler taranmadı. | ORG/UYUM + Demir + KAN.US: metinleri doldur, ödeme öncesi görünürlük/onay akışını değerlendir. |
| Satıcı kimliği / iletişim | FAIL | `/policies/contact-information` 404; `/pages/contact` 200 ama yalnız Ad/E-posta/Telefon/Yorum formu; satıcı unvan/adres/VKN-MERSİS bilgisi görünmüyor. | Demir gerçek şirket bilgilerini verir, ORG/UYUM kontrol eder; Admin'de yayın. |
| Kargo/dijital teslim politikası | FAIL (sayfa eksik) | `/policies/shipping-policy` 404. Dijital ürün için fiziksel kargo politikası şart kabul edilmedi; teslim süresi/kanalı ve iade koşulları yine açıklanmalı. | ORG/UYUM + Demir. |
| Çerez banner'ı / rıza | UNKNOWN | Ana sayfa/sepet statik HTML'de cookie/consent banner elemanı ve doğrulanmış Customer Privacy banner ayarı yok. Shopify analytics ve genel buyer-consent scripti var; bunlar çerez rızası kanıtı değil. JS/konum koşullu görünüm test edilmedi. | Demir: Admin → Customer privacy; ORG/UYUM kapsamı belirler; QA farklı bölge/temiz tarayıcıda kabul/red ve izleyici davranışını doğrular. |
| Alan adı / TLS / parola duvarı | PASS (kamu erişimi) | Ana sayfa ve sepet HTTPS üzerinden içerik dönüyor; curl sertifika hatası yok, parola sayfası değil. | Tam güvenlik taraması sayılmaz. |
| Marka / eski hero fiyatları | PASS | Başlık ve marka Kanus. Hero 200 istem / 18.900 TL ve 1.990 TL; eski 500 istem / 10.000 TL metni artık görünen ana sayfada yok. | Eski raporların bu maddeleri güncel engel sayılmamalı. |
| Featured ürün blokları | PASS (statik görünüm) | Aylık ve saatlik “Öne çıkan ürün”, fiyat ve Ekle düğmeleri HTML'de mevcut; JSON-LD InStock. | Etkileşimli satın alma testi yapılmadı. |
| Müşteriye teslim vaadi | FAIL (tutarsız) | Ana sayfada hem “API anahtarınız ... en geç 10 dakika içinde e-posta ile gönderilir” hem “Anahtarlar sipariş sonrası manuel olarak iletilir” var. Otomatik production teslim kabulü yok. | KAN.US + Demir: mevcut vaadi operasyonla tutarlı hale getirme değerlendirmesi. **Onaylı B/B metinleri ancak canlı webhook geçişi + gerçek e-posta doğrulaması sonrası yayınlanacak.** Bu görevde yayın yok. |

## Ürün/fiyat ayrıntısı

| Ürün | Onaylı TRY (KDV dahil) | Bugünkü kanıt | Sonuç |
|---|---:|---|---|
| 5 kredi | 695 | 6 Kasım aktif/fiyat doğru; bugün tüm katalog doğrulanamadı | UNKNOWN |
| 10 kredi | 1290 | Ana sayfa arama kartı + sepet önerisi 1,290.00TL | PASS fiyat; checkout uygunluğu UNKNOWN |
| 25 kredi | 2990 | 6 Kasım aktif/fiyat doğru; güncel JSON yok | UNKNOWN |
| 50 kredi | 5490 | 6 Kasım aktif/fiyat doğru; güncel JSON yok | UNKNOWN |
| 100 kredi | 9900 | Ana sayfa arama kartı + sepet önerisi 9,900.00TL | PASS fiyat; checkout uygunluğu UNKNOWN |
| Aylık / 200 istem | 18900 | `/` JSON-LD Offer 18900.00 TRY, InStock; handle kanus-ai-1-aylik-erisim-200-istem | PASS fiyat/ilan |
| Saatlik | 1990 | `/` JSON-LD Offer 1990.00 TRY, InStock; handle kanus-ai-1-saatlik-kullanim | PASS fiyat/ilan |
| 1 kredi | 149; satış dışı | Son yetkili rapor DRAFT; bugün tam katalog alınamadı | UNKNOWN güncel; tarihsel PASS |

HTTP429 nedeniyle `/products.json?limit=250`, `/collections/all` ve aylık/saatlik `.js` sonuçları kullanılabilir ürün kanıtı değildir. Bunlar yeniden denenmedi veya kimlik/başlık değiştirerek aşılmadı. Doğrudan ürün HTML sayfaları ve tüm kredi handle'ları ayrıca ziyaret edilmedi; fiyatların tamamı bugün teyit edildi denilemez.

## Yayın öncesi engeller — sıralı

1. **Production lisans/e-posta zinciri yok:** api.kanus.org health502; canlı webhook kesimi, doğru canlı product_id eşlemesi, ayrı depo ve gerçek SMTP teslim kabulü tamamlanmalı (HOST + Demir + KAN.US).
2. **Ödeme ve zorunlu e-posta UNKNOWN:** canlı ödeme sağlayıcı ve checkout iletişim ayarı Admin'de kanıtlanmalı (Demir).
3. **Yasal metinler ve satıcı kimliği eksik:** iade/şartlar/iletişim 404; KVKK, mesafeli satış, ön bilgilendirme/onay akışı ve gerçek şirket alanları hazır değil (ORG/UYUM + Demir).
4. **Teslim vaatleri tutarsız:** 10 dakika vaadi üretim kabulünden önce güvenceye alınamaz. B/B metinlerini Demir'in şartı gerçekleşmeden yayınlama (KAN.US).
5. **Tam katalog, KDV, dijital shipping ve 1 kredi gizliliği güncel teyit bekliyor; çerez rızası davranışı UNKNOWN** (Demir/SHOP + QA/UYUM).

## Kaynaklar ve kanıt zinciri

Bugünkü URL yanıtları: aynı klasörde `presale-evidence/`; her GET'in HTTP kodu, etkili URL ve hata çıktısı `fetch-status.json`. Ana sayfa `home.html`, sepet `cart`, politikalar `policies_*`, iletişim `pages_contact`, DNS `dns-api.json`, production sağlık `api-healthz`. Ana sayfa ve ürün liste GET'i ilk adımda kaydedildi; liste gövdesi `local_rate_limited`.

Okunan tarihsel dosyalar (`/workspace/us-panel/reports/`):
- LIVE_STORE_AUDIT_20261106.md
- KANUS_SHOPIFY_LIVE_AUDIT.md (uygulama sonrası ve 1 kredi DRAFT ekleri dahil; erken bölümlerde kalan eski durum notları güncel durum sanılmadı)
- SHOP_LIVE_READINESS_20261106.md
- SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md (PLAN; uygulanmış işlem değil)
- PLANS_LIVE_DIFF_20261106.md (plans.live.json yerel, deploy kanıtı değil)
- r20261108/KANUS_SHOPIFY_E2E_20261108.md (test kabulü üretim kabulü değildir)
- r20261108/SHOP_AUTOMATION_COPY_DRAFT_TR.md (B/B koşullu yayın kararı görev bağlamından)

`/workspace/us-panel/docs/legal/final/README.md`: unvan, adres, VKN/MERSİS, telefon, KEP/yasal başvuru kanalları, ödeme kuruluşu, veri aktarım envanteri, ifa süresi ve onay kayıt yöntemi dahil çok sayıda `<DOLDUR>` var. Dosya adındaki “final” yayın veya hukukçu kabulü anlamına gelmiyor. Bu denetim hukuki uygunluk sertifikası değildir.

Araç notu: Shopify skill okundu; paketinin log_feedback.mjs yardımcı dosyası belirtilen konumda bulunamadı (MODULE_NOT_FOUND). Hiçbir Shopify API/CLI işlemi yapılmadı.
