# Yerel taslaklar — kısa teslim raporu

**06.11.2026, TSİ (Europe/Istanbul). Yalnız yerel dosyalar; yayın yok.**
Ana prompt bütünüyle tarandı; §7 ve §10 esas alındı. Kod/dağıtım/hesap verisi değiştirilmedi. Secret/anahtar dosyaları okunmadı veya aktarılmadı. Shopify, site, DNS, sunucu veya harici servislere yazılmadı; kimseye mesaj/e-posta gönderilmedi. Web erişimi yalnız kamuya açık resmî mevzuatı okumak içindi.

## Oluşturulan dosyalar
- `/workspace/us-panel/brand/docs/DECISION_THELEMA_VS_YERLI.md`
- `/workspace/us-panel/docs/KULLANIM_KILAVUZU_TR.md`
- `/workspace/us-panel/docs/legal/CAYMA_VE_IADE_POLITIKASI.md`
- `/workspace/us-panel/docs/legal/KULLANIM_SARTLARI.md`
- `/workspace/us-panel/docs/legal/KVKK_AYDINLATMA_METNI.md`
- `/workspace/us-panel/docs/legal/MESAFELI_SATIS_SOZLESMESI.md`
- `/workspace/us-panel/reports/YEREL_TASLAKLAR_RAPORU_20261106_TR.md`

## Güncellenen dosyalar
- `/workspace/us-panel/brand/docs/NAMING.md`
- `/workspace/us-panel/brand/docs/TONE.md`
- `/workspace/us-panel/brand/docs/LOGO.md`
- `/workspace/us-panel/brand/exe/ABOUT_SPLASH_EN.md`
- `/workspace/us-panel/brand/exe/ABOUT_SPLASH_TR.md`
- `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_TR.md`
- `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_EN.md`
- `/workspace/us-panel/reports/KANUS_BRAND_REPORT.md`

## Doldurulan ve korunan doğrulanmış bilgiler
| Alan | Değer | Kaynak |
|---|---|---|
| Destek e-postası | support@kanus.org | `/workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md` mailbox ve SMTP FROM kaydı; TR/EN About yer tutucuları dolduruldu. Yasal bildirim/KVKK yetkisi varsayılmadı. |
| Yerel kaynak sürümü | 2026.11.06.7 | `/workspace/us-panel/work2/kiyas/term/release.go` buildVersion; TR/EN About dolduruldu. Dağıtılan EXE’nin sürüm/derleme kanıtı değildir. |
| İsim, web/mağaza (korundu) | Kanus.org, US Panel; https://kanus.org; https://shop.kanus.org | `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt` §7/§6. Kanus.org’un şirket unvanı olduğu varsayımı NAMING’de düzeltildi. |
| Gerçek ekran/düğmeler | Modeller, Model bağlantısı, Lisansla hesap aç, Hesaba giriş, Kredi yükle, Bakiyeyi göster ve geçmiş/durum düğmeleri | `/workspace/us-panel/work2/kiyas/term/gui.html`, `gui.go`, `pricing.go`; profil koşulu kılavuzda açık. |
| Kayıt/giriş, sahiplik, bakiye ayrımı | Kayıttan sonra ayrıca giriş; mevcut anahtarı bağlama; istem ve saniye ayrı | `/workspace/us-panel/work2/kiyas/term/account_client.go`, `gui.go`, `licensesrv/accounts.go`. Kaynak davranışı; canlı kabul değil. |
| Ölçüm | İstem türünde başarılı çağrı 1 kredi; saatlik başarılı çağrı yukarı saniye; soğuk açılış dahil/kabul kuyruğu hariç | `/workspace/us-panel/work2/kiyas/term/licensesrv/model_meter.go`, `model_proxy.go`. Fiyat taahhüdü değil. |

Fiyat/miktar/geçerlilik içeren genel ürün şablonları belirli SKU’ya bağlanmadı. Değişkenlerin tamamı `[DOLDURULACAK]` biçimine çevrildi; derleme kimliği, telif yılı ve iade URL’si uydurulmadı. Kullanım kılavuzu ve yasal taslaklarda olgusal kaynakların tam yolları ilgili bölümlerde yer alır.

**Önemli:** Önceki .6 çapraz derleme raporları native Windows veya .7 hesap modu kabulü değildir. Yerel kaynak hesabı desteklese de dağıtım/canlı endpoint doğrulaması yapılmadı. Eski GPU açılış süreleri garantiye çevrilmedi. Yerli/milli sloganının tüm model/sunucu menşei veya hukuki şirket merkezi anlamına gelmediği not edildi.

## Karar — seçenek seçilmedi
- **A:** Thelema korunur, yerli/milli mesajı yumuşatılır/ayrı kurumsal bağlama alınır. **Artı:** mevcut tercih/varlıklar korunur. **Eksi:** sembol çağrışımları ve açıklama ihtiyacı sürer.
- **B:** §7 ile uyumlu yerli/milli odaklı yeni ana işaret; Thelema ikincil/iç kullanım. **Artı:** dış mesaj uyumu artabilir. **Eksi:** yeni tasarım/onay ve çift işaret yönetimi gerekir.
- **Demir’e soru:** Ana marka için A’yı (Thelema korunur, yerli/milli söylem ayrı/yumuşak konumlanır) mı, B’yi (yerli/milli odaklı yeni ana işaret, Thelema ikincil/iç kullanım) mi onaylıyorsun?
- Ayrıntı: `/workspace/us-panel/brand/docs/DECISION_THELEMA_VS_YERLI.md`.

## Resmî kaynaklar
06.11.2026 TSİ tarihinde WebSearch/WebFetch ile içerikleri okunarak doğrulandı:
1. https://www.mevzuat.gov.tr/MevzuatMetin/1.5.6698.htm — KVKK m.4, 5, 7–14; güncel m.9 aktarım rejimi.
2. https://www.kvkk.gov.tr/Icerik/2033/Aydinlatma-Yukumlulugu- — aydınlatma kapsamı ve rızadan bağımsızlığı.
3. https://www.mevzuat.gov.tr/MevzuatMetin/1.5.6502.htm — m.4, 5, 48 ve tüketici başvuru yolları (68, 73, 73/A).
4. https://www.mevzuat.gov.tr/MevzuatMetin/yonetmelik/7.5.20237.htm — ön bilgilendirme 5–8, cayma 9–12, istisna 15/1-ğ ve h, ifa 16, kayıt/ispat 20.

**Hukuki ayrım:** 15/1-ğ anında ifa/teslimi düzenler; ayrıca onay ibaresi içermez. 15/1-h erken ifaya başlanan hizmetlerde tüketicinin önceden onayını gerektirir. Taslaklarda ayrıca ayrı aktif onay/kayıt süreci önerildi; genel “dijital ürün iade edilmez” hükmü kurulmadı. Kullanılmamış kredi/gelecekteki hizmet ayrıca değerlendirilecek. Dört dosyanın ilk satırında zorunlu hukukçu kontrolü notu var. KVKK’da güncel alıcı/ülke/aktarımı sadece tedarikçi markasına bakarak varsaymama ilkesi korundu.

## Kalan `[DOLDURULACAK]` alanları — özet
- Marka/About: derleme kimliği, telif yılı; ürün şablonlarında plan adı, miktar, süre/geçerlilik, teslimat/e-posta akışı, iade URL’si.
- Kılavuz: indirme URL’si, hesap EXE adı/SHA, minimum Windows/donanım/bağımlılıklar, imza/güncelleme/geri dönüş, parola sıfırlama/doğrulama ekranı, sipariş fiyat/miktar/süre/ücret, hesap profiline özgü bekleme ekranı ve iptal düğmesi, garantili süre/zaman aşımı, telefon/destek saatleri.
- Hukuk: satıcı/veri sorumlusu unvanı, adres, VKN/MERSİS, telefon, temsilci, yetkili yasal/KVKK/cayma kanalları; sipariş/alıcı alanları, plan/lisans kapsamı, fiyat/vergi/para birimi/ek ücret, ödeme sağlayıcısı/yöntemi, teslim/ifa ve teknik uyumluluk; ön bilgilendirme/onay/kalıcı nüsha kayıtları ve istisna sınıflandırması; şikâyet/iade ve hesap kapatma/itiraz süreçleri; veri kategorisi/amaç-hukuki sebep eşlemelerinin eksikleri, tedarikçi tüzel kişi/rolleri/ülkeler, fiilî yurt dışı aktarım mekanizması, saklama/imha/güvenlik, eğitim/istem saklama, içerik/çıktı lisansı, yürürlük/değişiklik süreçleri.

## Tam alan envanteri (dosya / satır / bağlam)
Aşağıda `[DOLDURULACAK]` içeren içerik satırları listelenir; bazıları gerçek alan yerine yer tutucu kullanım kuralını anlatır. Farklı satırlardaki aynı yer tutucu tek değer olarak topluca doldurulmamalıdır.

### `/workspace/us-panel/brand/docs/NAMING.md`
- Satır 9: | **Kanus.org** | Platform / site marka kimliği; hukuki şirket unvanı değildir (unvan: [DOLDURULACAK]) |

### `/workspace/us-panel/brand/docs/TONE.md`
- Satır 16: 5. **İddia disiplini:** Sayı, sertifika, sıralama → yalnızca doğrulanmış veri veya `[DOLDURULACAK]`.

### `/workspace/us-panel/brand/exe/ABOUT_SPLASH_EN.md`
- Satır 25: Build: [DOLDURULACAK]
- Satır 27: © [DOLDURULACAK] Kanus.org

### `/workspace/us-panel/brand/exe/ABOUT_SPLASH_TR.md`
- Satır 25: Derleme: [DOLDURULACAK]
- Satır 27: © [DOLDURULACAK] Kanus.org

### `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_EN.md`
- Satır 4: Unverified claims avoided; fill `[DOLDURULACAK]` from store data.
- Satır 8: - **Title pattern:** `Kanus AI — [DOLDURULACAK] Model Requests` or `US Panel — [DOLDURULACAK]`
- Satır 10: - **Delivery:** Digital — license / credits via email and panel (`[DOLDURULACAK]`)
- Satır 16: **Title:** `Kanus AI — [DOLDURULACAK] Model Requests`
- Satır 20: [DOLDURULACAK] model-request credits for US Panel on Kanus.org — AI-assisted software development.
- Satır 26: - [DOLDURULACAK] model-request credits
- Satır 28: - Validity: [DOLDURULACAK] days ([DOLDURULACAK])
- Satır 32: 2. Account linking / creation steps are sent by email ([DOLDURULACAK]).
- Satır 38: - Refunds & cancellations: [DOLDURULACAK]
- Satır 49: - Title: `Kanus AI — [DOLDURULACAK] Model Requests | shop.kanus.org`
- Satır 50: - Meta: `[DOLDURULACAK] model requests for US Panel. Homegrown tech for building software.`
- Satır 53: Paket değişkenleri yalnız [DOLDURULACAK] olarak standartlaştırıldı; alanın anlamı başlık, miktar, plan adı, geçerlilik, teslimat kanalı veya iade URL’si bağlamından okunmalıdır. Genel şablon tek bir SKU ile eşleştirilmediğinden farklı planların miktar/fiyat/süreleri buraya taşınmadı. E-posta/panel akışı bir taslaktır; siparişe özgü teslimat taahhüdü değildir. [Karar notu](/workspace/us-panel/brand/docs/DECISION_THELEMA_VS_YERLI.md) beklenir.

### `/workspace/us-panel/brand/shopify/PRODUCT_TEMPLATES_TR.md`
- Satır 4: Doğrulanmamış iddia yok; `[DOLDURULACAK]` alanları mağaza verisiyle doldurulur.
- Satır 8: - **Başlık kalıbı:** `Kanus AI — [DOLDURULACAK] Model İsteği` veya `US Panel — [DOLDURULACAK]`
- Satır 10: - **Teslimat:** Dijital — lisans / kredi e-posta ve panel üzerinden (`[DOLDURULACAK]`)
- Satır 16: **Başlık:** `Kanus AI — [DOLDURULACAK] Model İsteği`
- Satır 20: US Panel’de kullanmak üzere [DOLDURULACAK] model isteği kredisi. Kanus.org üzerinden yapay zekâ destekli yazılım geliştirme.
- Satır 26: - [DOLDURULACAK] adet model isteği kredisi
- Satır 28: - Süre: [DOLDURULACAK] gün ([DOLDURULACAK])
- Satır 32: 2. Hesap eşleştirme / oluşturma adımları e-posta ile iletilir ([DOLDURULACAK]).
- Satır 38: - İade ve iptal: [DOLDURULACAK]
- Satır 49: **Başlık:** `US Panel — [DOLDURULACAK]`
- Satır 53: [DOLDURULACAK]: [DOLDURULACAK] — Kanus.org US Panel için.
- Satır 56: **Uzun açıklama:** Şablon A ile aynı iskelet; `[DOLDURULACAK]` yerine plan alanları.
- Satır 62: - Title: `Kanus AI — [DOLDURULACAK] Model İsteği | shop.kanus.org`
- Satır 63: - Meta: `US Panel için [DOLDURULACAK] model isteği. Yerli teknolojiyle yazılım geliştir.`
- Satır 66: Paket değişkenleri yalnız [DOLDURULACAK] olarak standartlaştırıldı; alanın anlamı başlık, miktar, plan adı, geçerlilik, teslimat kanalı veya iade URL’si bağlamından okunmalıdır. Genel şablon tek bir SKU ile eşleştirilmediğinden farklı planların miktar/fiyat/süreleri buraya taşınmadı. E-posta/panel akışı bir taslaktır; siparişe özgü teslimat taahhüdü değildir. [Karar notu](/workspace/us-panel/brand/docs/DECISION_THELEMA_VS_YERLI.md) beklenir.

### `/workspace/us-panel/docs/KULLANIM_KILAVUZU_TR.md`
- Satır 7: 1. Dağıtım dosyasını yalnız doğrulanmış kanaldan al. İndirme URL’si: [DOLDURULACAK]. Hesap moduna uygun EXE dosya adı ve doğrulanmış SHA256: [DOLDURULACAK].
- Satır 13: **Gereksinimler:** Minimum Windows sürümü [DOLDURULACAK]; işlemci/RAM/disk [DOLDURULACAK]; doğrulanmış tarayıcı/çalışma zamanı bağımlılıkları [DOLDURULACAK]. Önceki .6 adayının Windows amd64 çapraz derlemesi raporlanmıştır; bunu .7 veya yerel Windows çalışma testi yerine kullanma. Kod imzası, otomatik güncelleme ve geri dönüş kabulü [DOLDURULACAK]. Güvenlik yazılımlarını devre dışı bırakma.
- Satır 26: E-posta doğrulama ve parola sıfırlama için sunucu yolları vardır; bu panelde doğrulanmış kullanıcı ekranı/bağlantısı [DOLDURULACAK]. “Mail kesin gönderildi” varsayma. Parolanı, oturumunu veya lisans anahtarını destek mesajına koyma.
- Satır 45: - Fiyat/para birimi, satın alınan miktar, geçerlilik ve saatlik ücret: [DOLDURULACAK]. Yerel katalog ile eski/canlı mağaza raporları farklı bağlamlar içerdiğinden bu kılavuz fiyat taahhüdü vermez. Siparişe özgü doğrulanmış bilgiler kullanılmalıdır.
- Satır 50: Model isteğini gönderdikten sonra GPU/model hazırlığı veya kuyruk beklemesi görülebilir. Bu ifade bilgisayarının aşırı ısındığını veya modelin hazır olduğunu kanıtlamaz. Go proxy, upstream beklerken **“GPU isiniyor; model hazirligi dogrulanmadi”** ve **“GPU isiniyor / kuyruk bekleniyor”** heartbeat yorumlarını gönderir. Türkçe görsel ekran ifadesi/hesap profiline özgü eşleşmesi: [DOLDURULACAK].
- Satır 54: - Aktif model işlemini durdurmak için hesabına uygun ekranda doğrulanacak düğme adı: [DOLDURULACAK]. Terminal iptali ile model iptali aynı işlem değildir.
- Satır 55: - Garantili açılış/yanıt süresi ve güncel zaman aşımı: [DOLDURULACAK]. Eski soğuk açılış süreleri ve §4 hedefleri hizmet garantisi değildir.
- Satır 80: **Destek?** `support@kanus.org` — kaynak: `/workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md`. Telefon ve destek saatleri [DOLDURULACAK]. Hata metni, sürüm ve zamanı TSİ olarak paylaş; özel verileri maskele.

### `/workspace/us-panel/docs/legal/CAYMA_VE_IADE_POLITIKASI.md`
- Satır 6: Bu taslak US Panel için dijital lisans, model isteği kredisi ve süre bazlı hizmet satışlarının tüketici işlemleri bölümünü kapsar. Satıcı/sağlayıcı hukuki unvanı, açık adresi, telefon, VKN/MERSİS ve cayma bildirimi alacak yetkili e-posta/kanal: [DOLDURULACAK]. Genel destek adresi support@kanus.org’dur; yasal bildirim yetkisi [DOLDURULACAK]. Yayınlanmış politika URL’si: [DOLDURULACAK]. Yerel dosya yolu politika URL’si yerine geçmez.
- Satır 18: **Önerilen ek tüketici koruma süreci — henüz uygulanmış değildir:** Ani teslim/erken ifa isteniyorsa, siparişe özgü bilgilendirmeden sonra, ifa başlamadan ayrı, önceden seçilmemiş aktif onay alınmalı; kapsam, onay zamanı, metin sürümü ve ifa/teslim kanıtı saklanmalı, kalıcı nüsha verilmelidir. Bu öneri ğ bendine kanunda olmayan genel bir şart eklendiği iddiası değildir; h bendindeki onay şartından ayrıdır. Onay yoksa erken ifa varsayılmaz; ifanın ertelenmesi/alternatif akış [DOLDURULACAK].
- Satır 21: > “[DOLDURULACAK] adlı hizmetin cayma süresi dolmadan ifasına başlanmasını talep ediyor ve önceden onaylıyorum. Bu onayla hizmet ifasına başlandığında, Yönetmelik m.15/1-h kapsamına giren sözleşme yönünden cayma hakkını kullanamayacağım konusunda bilgilendirildim. Ayıplı veya hiç ifa edilmeyen hizmete ilişkin kanuni haklarım saklıdır.”
- Satır 23: Anında teslim edilen dijital içerik için bilgilendirme metni ayrı hazırlanmalı; tüm paketler için bu hizmet beyanı kopyalanmamalıdır. Plan bazında istisna sınıflandırması, gösterilecek metin ve fiilî onay kayıtları: [DOLDURULACAK]. Genel kullanım şartlarını kabul etmek veya KVKK metnini okumak bu ön onayın yerine geçmez.
- Satır 33: Ödeme altyapısı, talep alma/inceleme süreci, bildirim kanalı ve mali kayıt eşleştirmesi: [DOLDURULACAK]. Tüketici kanunen bedel iadesine hak kazanmışsa yalnız kredi/kupon almaya zorlanamaz. Gönüllü ilave iade avantajı tanımlanmamıştır: [DOLDURULACAK].
- Satır 36: - Kime (satıcı/sağlayıcı unvanı, adres, yetkili e-posta): [DOLDURULACAK]
- Satır 38: - Sipariş/teslim tarihi ve sipariş kimliği: [DOLDURULACAK]
- Satır 39: - Mal/hizmet, miktar ve bedel: [DOLDURULACAK]
- Satır 40: - Tüketici adı soyadı ve adresi: [DOLDURULACAK]
- Satır 41: - Tarih; yalnız kağıt üzerinde gönderimde imza: [DOLDURULACAK]

### `/workspace/us-panel/docs/legal/KULLANIM_SARTLARI.md`
- Satır 6: US Panel, Kanus.org marka kimliğiyle sunulması öngörülen yapay zekâ destekli yazılım geliştirme panelidir. Hizmet sağlayıcı hukuki unvanı, açık adresi, VKN/MERSİS, telefon ve yasal bildirim kanalı: [DOLDURULACAK]. Genel destek support@kanus.org; destek saatleri ve yanıt taahhüdü [DOLDURULACAK]. Bu marka adı şirket unvanı veya Türkiye’de hukuki merkez kanıtı değildir.
- Satır 8: Bu yerel taslak hizmeti etkinleştirmez ve tek başına kullanıcıyı bağlamaz. Yayın/onay sürümü ve yürürlük tarihi [DOLDURULACAK]. Satın alımdaki bedel, süre ve kullanım kapsamı siparişe özgü ön bilgilendirme/sözleşmeyle belirlenir.
- Satır 13: Lisansın cihaz, kullanıcı, ticari kullanım, devir ve yeniden satış kapsamı [DOLDURULACAK]; kodda doğrulanmayan “sınırsız cihaz” veya “devredilemez” kuralı bu taslakla kesinleştirilmez. Güvenlik için hesap kısıtlama koşulları, bildirim, itiraz ve geri açma süreci [DOLDURULACAK]; kazanılmış haklar ve kanuni iadeler sırf hesap kapatıldı diye ortadan kalkmaz.
- Satır 18: Satın alınan planın fiyatı, miktarı, geçerliliği, tüketim kuralları ve varsa yenileme koşulları [DOLDURULACAK]. Bunlar ödeme öncesinde gösterilmeli; mevcut sözleşme tüketici aleyhine tek taraflı değiştirilmemelidir. GPU/model hazırlık veya kuyruk mesajı hazır olma ya da işlem başarısı taahhüdü değildir. Kullanılabilirlik, desteklenen sürüm ve hizmet seviyesi [DOLDURULACAK]. Bu eksiklikler taahhüt edilmiş ifadan kaçınma hakkı vermez.
- Satır 25: Kullanıcı içeriğine ilişkin haklar kullanıcıda veya ilgili hak sahibinde kalır. Hizmet için gerekli içerik işleme yetkisinin kapsamı ve çıktıların lisans/telif durumu [DOLDURULACAK]; münhasır hak veya ihlalsizlik garantisi uydurulmaz. İçeriğin model eğitiminde kullanılması, saklama ve üçüncü taraf lisans koşulları [DOLDURULACAK].
- Satır 28: [KVKK aydınlatma taslağı](/workspace/us-panel/docs/legal/KVKK_AYDINLATMA_METNI.md) ayrı tamamlanmalı ve sunulmalıdır. Bu şartları kabul etmek genel veri işleme/yurt dışı aktarım/pazarlama rızası değildir. Shopify ve Hostinger adlarının yerel kaynakta bulunması bunların hukuki rolü, tüm tedarikçi listesi veya aktarım güvencesinin doğrulandığı anlamına gelmez. Güncel alıcılar, ülkeler, ödeme/barındırma/model hizmeti sözleşmeleri [DOLDURULACAK].
- Satır 36: Şartların yeni sürümleri, bildirim ve kabul mekanizması [DOLDURULACAK]. Değişiklikler geçmiş işlemlere geriye dönük, tüketici aleyhine uygulanamaz. Hesap kapatma, içerik dışa aktarma, kalan kullanım hakkının akıbeti ve veri imha süreci [DOLDURULACAK]; bu süreçler kanuni fesih/iade ve KVKK haklarını daraltamaz.

### `/workspace/us-panel/docs/legal/KVKK_AYDINLATMA_METNI.md`
- Satır 9: - Veri sorumlusu gerçek kişi/tüzel kişi adı ve hukuki unvanı: [DOLDURULACAK]
- Satır 10: - Adres, varsa temsilci bilgileri, VKN/MERSİS: [DOLDURULACAK]
- Satır 11: - KVKK başvuru adresi, KEP ve kullanılabilecek elektronik başvuru kanalı: [DOLDURULACAK]
- Satır 13: - Genel destek adresi: support@kanus.org. Bu adresin KVKK başvurusu almaya yetkili kanal olduğu doğrulanmadı; başvuru yetkisi [DOLDURULACAK].
- Satır 20: | Kayıt/giriş formuyla e-posta ve parola; sunucuda parola özeti, hesap/oturum bilgileri | Hesap oluşturma, kimlik doğrulama, erişim güvenliği | Gerekli olduğu ölçüde sözleşmenin kurulması/ifası (m.5/2-c); güvenlik amaçlı meşru menfaat (m.5/2-f) için denge testi [DOLDURULACAK] |
- Satır 21: | Sipariş webhook’u üzerinden sipariş kimliği/adı, müşteri e-postası, ürün/plan ve lisans ilişkilendirme bilgileri | Dijital kullanım hakkını teslim, lisans/kredi yönetimi, iptal/iade mutabakatı | Sözleşmenin kurulması/ifası (m.5/2-c); somut mali kayıt yükümlülüğü ve dayanağı [DOLDURULACAK] |
- Satır 22: | Lisans sahipliği, rezervasyon/kesinleştirme kayıtları, istek kimliği ve zamanları | Bakiye ve kullanım geçmişi, uyuşmazlık incelemesi | Sözleşmenin ifası (m.5/2-c); bir hakkın tesisi/kullanılması/korunması için zorunlu işleme (m.5/2-e), gerekli kapsam [DOLDURULACAK] |
- Satır 23: | Ağ isteğinden IP; hesap işlemi hız sınırlama kaydı | Kötüye kullanım ve yetkisiz erişim önleme | Meşru menfaat (m.5/2-f), gereklilik/denge değerlendirmesi ve kalıcılık [DOLDURULACAK] |
- Satır 24: | Kullanıcının model çağrısına eklediği istemler, mesajlar ve seçilen proje içeriği | İstenen yapay zekâ destekli işlemi gerçekleştirme | İşlemenin gerekliliği ve kullanıcının üçüncü kişi verisini aktarma yetkisi ayrıca değerlendirilmeli; somut veri kategorisi/hukuki sebep [DOLDURULACAK] |
- Satır 25: | Destek başvurusunda kullanıcının ilettiği bilgiler | Talep/şikâyet yönetimi | Gerçek destek süreci, veri kategorileri ve hukuki sebep [DOLDURULACAK] |
- Satır 27: Parola, oturum ve lisans anahtarlarını destek başvurusuna eklemeyin. Özel nitelikli kişisel veri göndermeyin; bu tür veri için doğrulanmış ayrı işleme süreci [DOLDURULACAK]. Kod güvenlik önlemlerinin varlığı tüm hizmetin güvenlik kabulü anlamına gelmez. İstemlerin model eğitimi için kullanılıp kullanılmadığı ve içerik günlükleme/saklama kapsamı [DOLDURULACAK]; “eğitimde kullanılmaz” veya “hiç saklanmaz” taahhüdü verilmez.
- Satır 32: Yerel belgelerde **Shopify** mağaza/sipariş entegrasyonu ve **Hostinger** SMTP hizmeti adı açıkça geçmektedir. Bunlar markadır; sözleşmedeki alıcı tüzel kişi, rolü (veri işleyen/ayrı veri sorumlusu), aktarılan veri kategorileri ve ülke bilgileri doğrulanmış değildir: [DOLDURULACAK]. Shopify’ı ödeme kuruluşu olarak varsaymayın.
- Satır 34: - Ödeme kuruluşu, kart verisi işleme akışı ve alıcı tüzel kişi: [DOLDURULACAK]
- Satır 35: - Barındırma/model altyapısındaki güncel tedarikçi tüzel kişiler, roller ve alıcı listesi: [DOLDURULACAK]
- Satır 36: - E-posta tesliminde aktarılan veri ve sözleşmesel alıcı bilgileri: [DOLDURULACAK]
- Satır 37: - Hukuken yetkili kamu kurumlarına/uyuşmazlık mercilerine zorunlu aktarımın kapsamı ve somut dayanağı: [DOLDURULACAK]
- Satır 39: **Yurt dışı aktarım:** Yerel SRE raporunda `europe-west4-b` bölgesindeki model hesaplama kaynağı kayıtlıdır; bu, Türkiye içi işleme garantisi verilemeyeceğini gösteren teknik işarettir. Ancak kişisel verinin fiilen hangi alıcıya, hangi ülkede, hangi veri kategorileriyle aktarıldığı ve sonraki aktarım zinciri bu çalışma kapsamında doğrulanmadı. Fiilî yurt dışı aktarım envanteri, ülke/alıcı ve hukuki mekanizma: [DOLDURULACAK]. “Aktarım yok” denilemez. Modelin istemesi veya hizmetin kabul edilmesi tek başına aktarım mekanizması değildir.
- Satır 41: Güncel m.9; yeterlilik kararı, uygun güvenceler ve ancak kanundaki şartlarla arızi aktarımlar için farklı yollar düzenler. Uygulanacak mekanizma ve varsa standart sözleşme/bildirim bilgileri [DOLDURULACAK]. Sürekli hizmet aktarımı genel bir rıza kutusuyla otomatik olarak meşrulaştırılamaz.
- Satır 44: Veriler amaç için gerekli veya mevzuatça öngörülen süre kadar saklanır; sebepler kalkınca m.7 kapsamında silme, yok etme veya anonimleştirme değerlendirilir. Hesap, kullanım, istem, destek, e-posta, yedek ve güvenlik kayıtlarının ayrı saklama süreleri ve imha takvimi: [DOLDURULACAK]. Teknik erişim, şifreleme, yedek ve ihlal müdahale prosedürlerinin fiilî kapsamı: [DOLDURULACAK].
- Satır 49: Başvuru kanalı: [DOLDURULACAK]. Kimliği doğrulamak için gerekli ve ölçülü bilgi dışında veri talep edilmemelidir. Başvurular m.13 uyarınca en kısa sürede ve en geç otuz gün içinde, kural olarak ücretsiz sonuçlandırılır; ayrıca maliyet gerektiren işlemlerde Kurul tarifesi uygulanabilir. M.14’teki şartlar ve süreler dahilinde Kurula şikâyet hakkı saklıdır.

### `/workspace/us-panel/docs/legal/MESAFELI_SATIS_SOZLESMESI.md`
- Satır 8: **Satıcı/sağlayıcı:** Hukuki unvan/ad soyad [DOLDURULACAK]; açık adres [DOLDURULACAK]; VKN/MERSİS [DOLDURULACAK]; telefon [DOLDURULACAK]; yetkili sözleşme/şikâyet e-postası [DOLDURULACAK]. Marka: Kanus.org; ürün: US Panel. Genel destek: support@kanus.org; bu adresin yasal bildirim yetkisi [DOLDURULACAK]. Mağaza: https://shop.kanus.org.
- Satır 10: **Alıcı:** Ad soyad/unvan [DOLDURULACAK]; adres [DOLDURULACAK]; e-posta [DOLDURULACAK]; gerekli iletişim bilgisi [DOLDURULACAK]. Sipariş numarası ve sözleşme tarihi [DOLDURULACAK]. Gereksiz kimlik bilgisi toplanmamalıdır.
- Satır 19: | Ürün/plan ve SKU | [DOLDURULACAK] |
- Satır 20: | Lisans kapsamı, kullanım hakkı/devredilebilirlik ve cihaz sınırı | [DOLDURULACAK] |
- Satır 21: | Miktar ve birim (model isteği veya saniye/saat) | [DOLDURULACAK] |
- Satır 22: | Geçerlilik başlangıcı, bitişi, varsa yenileme/fesih | [DOLDURULACAK] |
- Satır 23: | Vergiler dahil toplam bedel / para birimi / vergi ayrımı | [DOLDURULACAK] |
- Satır 24: | Ek ücretler ve varsa hesaplama yöntemi | [DOLDURULACAK] |
- Satır 25: | Ödeme aracı, tahsilat yöntemi ve ödeme sağlayıcısı | [DOLDURULACAK] |
- Satır 26: | Teslim/ifa kanalı ve taahhüt edilen süre | [DOLDURULACAK] |
- Satır 27: | EXE sürümü, indirme URL’si, sistem/uyumluluk gereksinimleri | [DOLDURULACAK] |
- Satır 28: | Teknik korumalar / çevrim içi bağlantı koşulları | [DOLDURULACAK] |
- Satır 33: Belirli sipariş için onay metni, ödeme yükümlülüğü bildiren düğme, kalıcı nüsha teslim yöntemi ve kayıt kanıtı: [DOLDURULACAK]. Alıcının önceden açık onayı olmadan ek bedel tahsil edilmez. Sonradan yayımlanan fiyatlar mevcut sözleşmeyi tüketici aleyhine değiştirmez. Yenileme veya tekrarlayan ödeme bu taslakla varsayılmaz.
- Satır 38: Mevcut Go proxy kodunda başarılı istem türü çağrı 1 krediyle; başarılı saatlik çağrı upstream isteği başladıktan sonraki süre yukarı saniyeye yuvarlanarak kesinleştirilir. **Soğuk açılış beklemesi ölçüme dahil; proxy kabul kuyruğu süresi hariçtir.** Çağrı hatası/iptalinde kullanım rezervasyonu serbest bırakılır; bu teknik davranış para iadesi anlamına gelmez. Siparişe uygulanacak ölçüm usulü ve ön bilgilendirmede gösterildiğinin kanıtı: [DOLDURULACAK]. Bu açıklama fiyat veya süre garantisi vermez.
- Satır 45: İfa öncesi ayrı bilgilendirme ve aktif ön onay kaydı alınması bu taslağın yayın koşuludur. Kullanılmamış kredi veya gelecekte sunulacak hizmete, yalnız lisans kodu teslimi gerekçesiyle toplu “iade yok” hükmü uygulanamaz. Somut planın hukuki niteliği ve istisna kapsamı hukukçu tarafından belirlenmelidir: [DOLDURULACAK]. Ayıplı/eksik ifaya ilişkin zorunlu haklar hiçbir durumda ortadan kalkmaz.
- Satır 50: Şikâyet kanalı ve çözüm yöntemi: [DOLDURULACAK]. Tüketiciler uyuşmazlığın değeri/görev kurallarına göre tüketici hakem heyetine veya dava şartı arabuluculuğun uygulanabildiği hallerde önce arabulucuya, ardından tüketici mahkemesine başvurabilir. 6502 m.68, 73 ve 73/A kapsamındaki haklar korunur; yıllık parasal sınır burada doğrulanmadığından tutar belirtilmemiştir. Tek bir şehir mahkemesi zorunlu kılınmaz.
- Satır 53: Veri sorumlusu ve işleme/aktarım bilgileri [KVKK taslağında](/workspace/us-panel/docs/legal/KVKK_AYDINLATMA_METNI.md) tamamlanmalıdır; bu sözleşme genel açık rıza değildir. Sözleşme sürümü ve alıcının onay kaydı [DOLDURULACAK]. Yönetmelik m.20 kapsamındaki işlem bilgi/belgeleri için üç yıllık saklama yükümlülüğü, tüm kişisel verilerin aynı süreyle tutulacağı anlamına gelmez.

## Yerel dosya kontrolü
PASS — Dört hukuk dosyasının zorunlu ilk satırı, 18 metin dosyasının yerel bağlantıları ve marka metinlerindeki eski yer tutucular denetlendi. İlk kontrolde rapora aktarılan dört göreli bağlantı bozuk bulundu; raporda mutlak yerel hedeflerle düzeltildi, yeniden kontrol PASS. Uygulama/Windows/canlı hizmet testi veya görsel varlık yeniden üretimi yapılmadı.
