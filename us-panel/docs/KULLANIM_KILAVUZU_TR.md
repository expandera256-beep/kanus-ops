# US Panel — Türkçe kullanım kılavuzu

**Yalnız yerel TASLAK — 06.11.2026, TSİ (Europe/Istanbul).**
Bu belge mevcut yerel kaynak davranışını anlatır; canlı hizmetin veya Windows sürümünün kabul edildiği anlamına gelmez. Ton ve isimler: `/workspace/us-panel/brand/docs/TONE.md`, `/workspace/us-panel/brand/docs/NAMING.md`.

## Marka ve anlatım notu — A kararı
KAN.US’nin ilettiği Demir seçimi, **06.11.2026 08:13 TSİ (Europe/Istanbul)**: Thelema logosu ana işaret kalır; yerli/milli söylemi metin ve ton düzeyinde yumuşatılır. [Karar kaydı](../brand/docs/DECISION_THELEMA_VS_YERLI.md).

Bu kılavuz Türkçe, açık ve ürün odaklı anlatımı korur; şirketin merkezi, modelin veya altyapının menşei hakkında iddia eklemez. Logo seçimi teknik özellik, güvenlik, performans veya hazır olma garantisi değildir. Aşağıdaki kaynaklı kullanım adımları ve bilinmeyen alanlar değişmemiştir.

## 1. EXE’yi hazırlama ve açma
1. Dağıtım dosyasını yalnız doğrulanmış kanaldan al. İndirme URL’si: [DOLDURULACAK]. Hesap moduna uygun EXE dosya adı ve doğrulanmış SHA256: [DOLDURULACAK].
2. Dosyayı yazma iznin olan bir klasöre koy. Bu kaynakta bir kurulum sihirbazının varlığı doğrulanmadı; “İleri / Kur” adımları varsayılmamalıdır.
3. EXE’yi aç. Windows’ta normal açılış panel arayüzünü başlatacak şekilde kodlanmıştır. Yerel panel adresi `http://127.0.0.1:4477`’dir; arayüzü bu bilgisayar üzerinde kullan, dış ağa açma.
4. Eski panel çalışıyorsa önce onu kapat; kaynak kodu aynı portta farklı sürümü sessizce yeni sürüm gibi açmayı engeller.
5. Ekrandaki sürümü kontrol et. Yerel `buildVersion`: **2026.11.06.7**; bu değer eldeki EXE’nin derlendiği/dağıtıldığı sürüm için kanıt değildir.

**Gereksinimler:** Minimum Windows sürümü [DOLDURULACAK]; işlemci/RAM/disk [DOLDURULACAK]; doğrulanmış tarayıcı/çalışma zamanı bağımlılıkları [DOLDURULACAK]. Önceki .6 adayının Windows amd64 çapraz derlemesi raporlanmıştır; bunu .7 veya yerel Windows çalışma testi yerine kullanma. Kod imzası, otomatik güncelleme ve geri dönüş kabulü [DOLDURULACAK]. Güvenlik yazılımlarını devre dışı bırakma.

Kaynaklar: `/workspace/us-panel/work2/kiyas/term/main.go`, `gui.go` (`guiAddr`), `release.go`; `/workspace/us-panel/reports/EXE_SMOKE_CHECKLIST_TR.md`, `/workspace/us-panel/reports/ROUND6_RAPOR_TR.md`. **Gerçek Windows ortamında bu çalışma kapsamında test yapılmadı.**

## 2. Hesap modu: kayıt ve giriş
Bu adımlar hesap formunun açık olduğu derleme içindir. `go-hook` profilinde hesap formu gizlenir; eski anahtar profili ile hesap profili aynı değildir. Yerel kaynakta hesap yolları bulunsa da canlı gateway’in bunları sunduğu burada doğrulanmadı.

1. **Modeller → Model bağlantısı** bölümünü aç.
2. **E-posta** ve **Parola** alanlarını doldur. Mevcut Go hesap sunucusu parolayı UTF-8 olarak **12–72 bayt** aralığında denetler; Türkçe karakterler birden fazla bayt olabilir.
3. Varsa satın aldığın anahtarı **Lisans veya kredi paketi anahtarı** alanına gir. Kayıtta anahtar sunucu tarafında isteğe bağlıdır; düğmenin adı **Lisansla hesap aç**’tır.
4. **Lisansla hesap aç** düğmesini kullan. Başarı durumunda **“Hesap açıldı. Giriş yapın.”** görünür; kayıt otomatik giriş yapmaz.
5. Aynı e-posta ve parolayla **Hesaba giriş** düğmesini kullan. Başarı bildirimi **“Hesap girişi yapıldı.”** olur.

E-posta doğrulama ve parola sıfırlama için sunucu yolları vardır; bu panelde doğrulanmış kullanıcı ekranı/bağlantısı [DOLDURULACAK]. “Mail kesin gönderildi” varsayma. Parolanı, oturumunu veya lisans anahtarını destek mesajına koyma.

Kaynaklar: `/workspace/us-panel/work2/kiyas/term/gui.html` (`accountform`), `gui.go` (`account-register`, `account-login`), `pricing.go` (`renderGUIPage`), `commercial_policy.go`, `licensesrv/accounts.go`.

## 3. Lisans anahtarını hesaba bağlama ve bakiye
1. Önce ilgili gateway üzerinde hesabına giriş yap.
2. Anahtarı **Lisans veya kredi paketi anahtarı** alanına gir; **Kredi yükle** düğmesini kullan. Bu düğme yeni ödeme yapmaz, mevcut lisansı hesaba bağlama isteği gönderir.
3. **Bakiyeyi göster** düğmesini kullan. Kaynak arayüz **“Kalan model isteği: … · Saatlik bakiye: … saniye”** biçiminde iki ayrı birim gösterir; bunları toplama.
4. **Model çağrısı ve kredi geçmişi** ile son çağrıların durumunu yenile. `reserved` ayrılmış, `completed` kesinleşmiş, `refunded` serbest bırakılmış çağrı kaydını ifade eder; sonuncusu banka/kart para iadesi demek değildir.
5. **Sunucu durumunu yenile** durum bilgisi ister; bilinmeyen durum “hazır” sayılmaz. Hesap girişi ve durum sorguları model sunucusunu açmak için tasarlanmamıştır.

Aynı anahtar aynı hesaba tekrar bağlandığında ikinci kez kredi üretilmez. Başka hesaba bağlı, süresi bitmiş, iptal edilmiş veya kullanım rezervasyonu bulunan anahtar reddedilebilir. Sorunda yeniden satın almadan önce bakiyeyi ve destek yanıtını kontrol et.

Kaynaklar: `/workspace/us-panel/work2/kiyas/term/gui.html`, `gui.go`, `account_client.go`; `/workspace/us-panel/work2/kiyas/term/licensesrv/accounts.go` (sahiplik, bakiye ve redeem).

## 4. Kredi ve saatlik kullanım
- İstem türü kullanımda mevcut Go proxy başarılı model çağrısı için **1 kredi** kesinleştirir. **Bir kullanıcı görevi birden fazla ücretli model çağrısı yapabilir.** Bir görev her zaman bir kredi değildir.
- Saatlik kullanım saniye bakiyesiyle tutulur. Başarılı çağrıda upstream isteği başladıktan sonraki geçen süre yukarı tam saniyeye yuvarlanır; **soğuk açılış beklemesi dahil**, proxy kabul kuyruğunda geçirilen süre hariçtir. Bu kod davranışıdır; pazarlamadaki “sadece üretim süresi” varsayımıyla karıştırma.
- İşlem öncesi bakiye ayrılır. Mevcut kod hata/iptalde rezervasyonu serbest bırakmayı hedefler; ekranda bekleyen kayıt varsa bakiyeyi ve geçmişi yenile. Bu, ödeme iadesi politikası değildir.
- Fiyat/para birimi, satın alınan miktar, geçerlilik ve saatlik ücret: [DOLDURULACAK]. Yerel katalog ile eski/canlı mağaza raporları farklı bağlamlar içerdiğinden bu kılavuz fiyat taahhüdü vermez. Siparişe özgü doğrulanmış bilgiler kullanılmalıdır.

Kaynaklar: `/workspace/us-panel/work2/kiyas/term/licensesrv/model_meter.go`, `model_proxy.go`; `/workspace/us-panel/work2/kiyas/term/gui.html`. Canlı tahsilat testi bu görevde yapılmadı.

## 5. “GPU ısınıyor” ve bekleme
Model isteğini gönderdikten sonra GPU/model hazırlığı veya kuyruk beklemesi görülebilir. Bu ifade bilgisayarının aşırı ısındığını veya modelin hazır olduğunu kanıtlamaz. Go proxy, upstream beklerken **“GPU isiniyor; model hazirligi dogrulanmadi”** ve **“GPU isiniyor / kuyruk bekleniyor”** heartbeat yorumlarını gönderir. Türkçe görsel ekran ifadesi/hesap profiline özgü eşleşmesi: [DOLDURULACAK].

- Aynı görevi peş peşe yeniden gönderme; birden fazla çağrı oluşturabilirsin.
- Sonuç veya hata durumunu bekle; bağlantı hatasında hesap, ağ ve bakiye durumunu kontrol et.
- Aktif model işlemini durdurmak için hesabına uygun ekranda doğrulanacak düğme adı: [DOLDURULACAK]. Terminal iptali ile model iptali aynı işlem değildir.
- Garantili açılış/yanıt süresi ve güncel zaman aşımı: [DOLDURULACAK]. Eski soğuk açılış süreleri ve §4 hedefleri hizmet garantisi değildir.
- Saatlik kullanımda başarılı çağrının soğuk açılış beklemesinin de ölçüme dahil olduğunu unutma.

Kaynaklar: `/workspace/us-panel/work2/kiyas/term/licensesrv/model_proxy.go`, `model_meter.go`, `/workspace/us-panel/work2/kiyas/term/model.go`; `/workspace/us-panel/reports/GPU_SOGUK_ACILIS_TR.md` yalnız tarihsel not, `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt` §4 yalnız hedef.

## 6. Güvenli proje kullanımı
Üretilen dosya önerilerini incele; **Değişiklikleri uygula** veya **Reddet** seçeneklerini kullan. **Testleri çalıştır** proje kodu çalıştırabilir. Terminalde **Bu proje kodunu çalıştırmayı onaylıyorum.** onayı gerekir. Komutlar seçili projede işletim sistemi yetkilerinle çalışır; bu bir CPU/RAM güvenlik sandbox’ı değildir. Bağımlılıkların otomatik kurulduğunu varsayma. Önemli dosyalarını yedekle; özel verileri ve sırları model istemlerine ekleme.

Kaynak: `/workspace/us-panel/work2/kiyas/term/gui.html` (proje önerileri ve Terminal).

## 7. SSS
**Hesap formu yok.** Farklı profil/derleme olabilir. Ekrandaki sürümü kaydet; hesap modunu destekleyen doğrulanmış EXE’yi iste. URL değiştirmenin tek başına profil değiştirdiğini varsayma.

**Kayıt oldum ama giriş yapılmadı.** Bu beklenen kaynak davranışı; ayrıca **Hesaba giriş** kullan.

**Kredi yükle yeniden para çeker mi?** Bu panel işlemi mevcut anahtarı bağlama isteğidir; ödeme/checkout değildir.

**Hesap işlemi reddedildi.** Bilgileri ve gateway’i kontrol et. Canlı hesap servisi hazır olmayabilir; art arda denemek yerine destek iste. Anahtarı/parolayı gönderme.

**Bakiye neden düşük görünüyor?** Aktif rezervasyonlar kullanılabilir bakiyeden düşülür; kredi ve saniye ayrı birimlerdir. Geçmişi ve bakiyeyi yenile.

**Hata olunca kartıma para mı iade edilir?** Rezervasyonun serbest bırakılması ile parasal iade ayrıdır; [yerel cayma/iade taslağı](legal/CAYMA_VE_IADE_POLITIKASI.md) henüz yayınlanmış politika değildir.

**EXE üretime hazır mı?** Bu belge böyle bir onay vermez; native Windows ve canlı hesap/model kabulü ayrıca gerekir.

**Destek?** `support@kanus.org` — kaynak: `/workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md`. Telefon ve destek saatleri [DOLDURULACAK]. Hata metni, sürüm ve zamanı TSİ olarak paylaş; özel verileri maskele.
