Taslaktır, yayınlanmadan önce hukukçu kontrolü gerekir

# US Panel — KVKK aydınlatma metni taslağı

**Yerel taslak; veri envanteri ve alıcı listesi tamamlanmadan yayınlanamaz.**
Bu metin, 6698 sayılı Kişisel Verilerin Korunması Kanunu kapsamında US Panel ve Kanus.org ile ilişkili veri işleme faaliyetlerini açıklamak üzere hazırlanmıştır. Aydınlatma açık rıza yerine geçmez; sözleşme kabulü ile pazarlama rızası birleştirilmemelidir.

## 1. Veri sorumlusu ve iletişim
- Veri sorumlusu gerçek kişi/tüzel kişi adı ve hukuki unvanı: [DOLDURULACAK]
- Adres, varsa temsilci bilgileri, VKN/MERSİS: [DOLDURULACAK]
- KVKK başvuru adresi, KEP ve kullanılabilecek elektronik başvuru kanalı: [DOLDURULACAK]
- Marka/platform: Kanus.org; ürün: US Panel. Bu marka adları hukuki veri sorumlusu unvanı değildir.
- Genel destek adresi: support@kanus.org. Bu adresin KVKK başvurusu almaya yetkili kanal olduğu doğrulanmadı; başvuru yetkisi [DOLDURULACAK].

## 2. Veriler, yöntem, amaç ve hukuki sebep
Aşağıdaki tablo yerel koddan çıkarılmış taslak envanterdir; canlıda gerçekten toplanan kapsam veri sorumlusu tarafından doğrulanmalıdır. Gereksiz veri toplanmaz; amaçla bağlantılı ve ölçülü işleme esastır (m.4).

| Veri / elde edilme yöntemi | İşleme amacı | Hukuki sebep taslağı / yayın engeli |
|---|---|---|
| Kayıt/giriş formuyla e-posta ve parola; sunucuda parola özeti, hesap/oturum bilgileri | Hesap oluşturma, kimlik doğrulama, erişim güvenliği | Gerekli olduğu ölçüde sözleşmenin kurulması/ifası (m.5/2-c); güvenlik amaçlı meşru menfaat (m.5/2-f) için denge testi [DOLDURULACAK] |
| Sipariş webhook’u üzerinden sipariş kimliği/adı, müşteri e-postası, ürün/plan ve lisans ilişkilendirme bilgileri | Dijital kullanım hakkını teslim, lisans/kredi yönetimi, iptal/iade mutabakatı | Sözleşmenin kurulması/ifası (m.5/2-c); somut mali kayıt yükümlülüğü ve dayanağı [DOLDURULACAK] |
| Lisans sahipliği, rezervasyon/kesinleştirme kayıtları, istek kimliği ve zamanları | Bakiye ve kullanım geçmişi, uyuşmazlık incelemesi | Sözleşmenin ifası (m.5/2-c); bir hakkın tesisi/kullanılması/korunması için zorunlu işleme (m.5/2-e), gerekli kapsam [DOLDURULACAK] |
| Ağ isteğinden IP; hesap işlemi hız sınırlama kaydı | Kötüye kullanım ve yetkisiz erişim önleme | Meşru menfaat (m.5/2-f), gereklilik/denge değerlendirmesi ve kalıcılık [DOLDURULACAK] |
| Kullanıcının model çağrısına eklediği istemler, mesajlar ve seçilen proje içeriği | İstenen yapay zekâ destekli işlemi gerçekleştirme | İşlemenin gerekliliği ve kullanıcının üçüncü kişi verisini aktarma yetkisi ayrıca değerlendirilmeli; somut veri kategorisi/hukuki sebep [DOLDURULACAK] |
| Destek başvurusunda kullanıcının ilettiği bilgiler | Talep/şikâyet yönetimi | Gerçek destek süreci, veri kategorileri ve hukuki sebep [DOLDURULACAK] |

Parola, oturum ve lisans anahtarlarını destek başvurusuna eklemeyin. Özel nitelikli kişisel veri göndermeyin; bu tür veri için doğrulanmış ayrı işleme süreci [DOLDURULACAK]. Kod güvenlik önlemlerinin varlığı tüm hizmetin güvenlik kabulü anlamına gelmez. İstemlerin model eğitimi için kullanılıp kullanılmadığı ve içerik günlükleme/saklama kapsamı [DOLDURULACAK]; “eğitimde kullanılmaz” veya “hiç saklanmaz” taahhüdü verilmez.

## 3. Alıcılar ve aktarım
Yalnız amaç için gerekli veri, uygun hukuki şartlarla yetkili alıcılarla paylaşılmalıdır. Yurt içi aktarım m.8, yurt dışı aktarım m.9 kapsamında ayrı değerlendirilir.

Yerel belgelerde **Shopify** mağaza/sipariş entegrasyonu ve **Hostinger** SMTP hizmeti adı açıkça geçmektedir. Bunlar markadır; sözleşmedeki alıcı tüzel kişi, rolü (veri işleyen/ayrı veri sorumlusu), aktarılan veri kategorileri ve ülke bilgileri doğrulanmış değildir: [DOLDURULACAK]. Shopify’ı ödeme kuruluşu olarak varsaymayın.

- Ödeme kuruluşu, kart verisi işleme akışı ve alıcı tüzel kişi: [DOLDURULACAK]
- Barındırma/model altyapısındaki güncel tedarikçi tüzel kişiler, roller ve alıcı listesi: [DOLDURULACAK]
- E-posta tesliminde aktarılan veri ve sözleşmesel alıcı bilgileri: [DOLDURULACAK]
- Hukuken yetkili kamu kurumlarına/uyuşmazlık mercilerine zorunlu aktarımın kapsamı ve somut dayanağı: [DOLDURULACAK]

**Yurt dışı aktarım:** Yerel SRE raporunda `europe-west4-b` bölgesindeki model hesaplama kaynağı kayıtlıdır; bu, Türkiye içi işleme garantisi verilemeyeceğini gösteren teknik işarettir. Ancak kişisel verinin fiilen hangi alıcıya, hangi ülkede, hangi veri kategorileriyle aktarıldığı ve sonraki aktarım zinciri bu çalışma kapsamında doğrulanmadı. Fiilî yurt dışı aktarım envanteri, ülke/alıcı ve hukuki mekanizma: [DOLDURULACAK]. “Aktarım yok” denilemez. Modelin istemesi veya hizmetin kabul edilmesi tek başına aktarım mekanizması değildir.

Güncel m.9; yeterlilik kararı, uygun güvenceler ve ancak kanundaki şartlarla arızi aktarımlar için farklı yollar düzenler. Uygulanacak mekanizma ve varsa standart sözleşme/bildirim bilgileri [DOLDURULACAK]. Sürekli hizmet aktarımı genel bir rıza kutusuyla otomatik olarak meşrulaştırılamaz.

## 4. Saklama ve güvenlik
Veriler amaç için gerekli veya mevzuatça öngörülen süre kadar saklanır; sebepler kalkınca m.7 kapsamında silme, yok etme veya anonimleştirme değerlendirilir. Hesap, kullanım, istem, destek, e-posta, yedek ve güvenlik kayıtlarının ayrı saklama süreleri ve imha takvimi: [DOLDURULACAK]. Teknik erişim, şifreleme, yedek ve ihlal müdahale prosedürlerinin fiilî kapsamı: [DOLDURULACAK].

## 5. Haklar ve başvuru
Kanun m.11 uyarınca; verilerinizin işlenip işlenmediğini öğrenme, işlenmişse bilgi isteme, amaç ve amaca uygunluğu öğrenme, yurt içi/yurt dışı alıcıları bilme, eksik/yanlış verilerin düzeltilmesini isteme, m.7 şartlarında silme/yok etme, düzeltme/silmenin alıcılara bildirilmesini isteme, münhasıran otomatik analizle aleyhinize sonuç doğmasına itiraz ve hukuka aykırı işleme nedeniyle zararın giderilmesini talep haklarınız vardır.

Başvuru kanalı: [DOLDURULACAK]. Kimliği doğrulamak için gerekli ve ölçülü bilgi dışında veri talep edilmemelidir. Başvurular m.13 uyarınca en kısa sürede ve en geç otuz gün içinde, kural olarak ücretsiz sonuçlandırılır; ayrıca maliyet gerektiren işlemlerde Kurul tarifesi uygulanabilir. M.14’teki şartlar ve süreler dahilinde Kurula şikâyet hakkı saklıdır.

## Yerel olgusal kaynaklar
- `/workspace/us-panel/work2/kiyas/term/licensesrv/accounts.go`: hesap, parola özeti, oturum, lisans sahipliği, IP hız sınırlama.
- `/workspace/us-panel/work2/kiyas/term/account_client.go`: hesap formları ve gateway’e iletim.
- `/workspace/us-panel/work2/kiyas/term/licensesrv/model_proxy.go`, `model_meter.go`: model iletimi ve kullanım kayıtları.
- `/workspace/us-panel/reports/KANUS_RELEASE_NOTES.md`: sipariş adı/kimliği ve outbox açıklaması.
- `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt` §6: Shopify hedef entegrasyonu; hedefler canlı kanıt sayılmaz.
- `/workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md`: destek posta kutusu ve Hostinger SMTP kaydı.
- `/workspace/us-panel/reports/SRE_AUDIT_20261106.md`: model kaynağının teknik bölge kaydı; tam veri aktarım envanteri değildir.

## Resmî kaynaklar
06.11.2026, TSİ (Europe/Istanbul), WebSearch/WebFetch doğrulaması:
- 6698 sayılı Kanun, m.4, 5, 7, 8, 9, 10, 11, 12, 13, 14: https://www.mevzuat.gov.tr/MevzuatMetin/1.5.6698.htm
- KVKK — Aydınlatma Yükümlülüğü: https://www.kvkk.gov.tr/Icerik/2033/Aydinlatma-Yukumlulugu-
M.9’un 2024 değişikliği işlenmiş güncel metni kullanıldı; eski “yalnız açık rıza” yaklaşımı esas alınmadı. Yayın günü mevzuat ve fiilî envanter yeniden doğrulanmalıdır.
