# US 2026.11.05.4 — canlı kontrol hizmetleri ve doğrulama durumu

Tarih: 5 Kasım 2026, Europe/Istanbul. Aktif kaynak: `us_checkout/kiyas/term` (Go paneli) ve `us_checkout/kanus` (ticari geçit).

**Kullanıcının açık onayıyla kontrol hizmetleri, kalıcı PostgreSQL ve özel TLS model bağlantısı kuruldu. Bu sürüm üretime hazır değildir: dış uç noktalar ilk kontrolde 404 verdi; ardından bağlı Chrome oturumuna erişim kesildi. Gerçek açılış–istek–boşta kapanış, canlı Shopify/e-posta ve Windows EXE çalışma kabulü tamamlanmadı.**

Yeni EXE'nin varsayılanı oluşturulan HTTPS geçididir: `https://us-commerce-iqhblf63dq-ez.a.run.app`. Ticari derlemede bu adres veya açıkça ayarlanmış başka bir HTTPS hesap geçidi yoksa eski HTTP sunucusuna geri dönüş yapılmaz. Adresin EXE'ye eklenmiş olması erişim ve ticari kabul testinin geçtiği anlamına gelmez.

## Bu onaydan sonra uygulanan kurulum

| Bileşen | Gözlenen sonuç |
| --- | --- |
| Cloud SQL `us-commerce-pg` | PostgreSQL 17 Enterprise, europe-west4, tek bölge/zonal, 1 vCPU / 3.75 GiB; RUNNABLE. 20 GB SSD, otomatik büyüme üst sınırı 60 GB; silme koruması, günlük yedek, 7 günlük PITR açık. HA değildir; geri yükleme denemesi henüz yapılmadı. |
| Veritabanı güvenliği | `ENCRYPTED_ONLY`; genel IP'de doğrudan ağ izin listesi yok. Cloud SQL Auth Proxy uzak TLS/kimlik doğrulaması sağlar. Cloud Run yerel Unix soketi üzerinden bağlanır. Uygulama rolünün superuser/createdb/createrole izinlerinin üçü de false olarak kontrol edildi; yalnız uygulama şeması kullanılır. |
| `us-commerce` | Cloud Run hizmeti oluşturuldu; Ready=True; min=0, max=2. Varsayılan public invoker, uygulamada hesap/oturum ve kota kontrolü. Dış `/healthz` ilk kontrolü Google 404 verdi; neden henüz çözülemedi. |
| `us-watchdog` | Ayrı Cloud Run hizmeti Ready=True; min=0, max=1. IAM politikasında allUsers/allAuthenticatedUsers olmadığı dağıtım betiğinde kontrol edildi; Scheduler hesabına invoker verildi. Bağımsız dış HTTP kontrolü de 404 verdi. |
| Scheduler `us-idle-watchdog` | Dakikalık UTC görev ENABLED; OIDC audience ve ayrı çağıran hesabı gerçek watchdog URL'sine bağlı. Başarılı reconcile/kalp atışı henüz gözlenmedi. |
| A100 izinleri | Ayrı gateway hesabına yalnız instance get/start, watchdog hesabına get/stop özel rolleri **A100 kaynağı üzerinde** bağlandı. Gerçek Cloud Run kimliğiyle güç işlemi kabulü henüz yapılmadı. |
| Sırlar | Model tokenı, lisans imza anahtarı, PostgreSQL DSN, katalog, politika ve CA Secret Manager'a bağlandı; yalnız ilgili uygulama hesapları okuyabilir. Gerçek Shopify anahtarı bulunmadığı için sahte anahtar üretilmedi. Sırlar kaynak paketine/EXE'ye eklenmedi. |
| Model TLS | Mevcut nginx'e ayrı 10.164.0.6:8443 TLS/Bearer dinleyicisi eklendi; anahtarsız 401, doğru anahtarla 200 ve yüklenmiş model doğrulandı. 8443 için yalnız 10.164.0.0/20 kaynak aralığına izin, diğer kaynaklara deny eklendi; diğer servis portları değiştirilmedi. |
| A100 açılış hatası | `/opt/ds/boot.sh` içindeki RAID birleştirmesi başarısız olunca otomatik `mdadm --create` ve `mkfs.ext4` çalıştıran yol kaldırıldı. Artık hata verip durur; model diskini biçimlendirmez. Eski betik özel dizinde yedeklendi; `bash -n` geçti. Yeni hash: `a8da4de772090e82ee6f630325566e1410027e5d409a64e280e9bff576bc6b4f`. |
| Model kopyası ve koruma tabanı | GCS'deki 7 nesne toplam 256.816.878.701 bayt; yerel 7 dosyanın boyutları aynı. İlk/son 4 MiB örnekleri ve 64 bayt SSD canary tabanı kaydedildi. Tam dosya hash eşleşmesi veya stop/start sonrası koruma henüz doğrulanmadı. |
| İmaj | Cloud Build SUCCESS, imaj digest'i `sha256:0da920e46ecd6392e234847ca9e87f3dc4bdb64165deccc51a50a869ade1b70e`; iki hizmet değişmez digest ile kuruldu. İlk Cloud Shell Docker push bağlantı reddi nedeniyle başarısız oldu; Cloud Build ile yayınlandı. Ayrı build hesabı depo yazma, kaynak bucket okuma ve log yazma izinleriyle kullanıldı. |

İşlem dizini Cloud Shell'de `$HOME/us-live-20261105` altındadır. Özel anahtarlar/operatör parolaları bu dizinin özel bölümündedir; bu raporda ve dağıtım paketinde bulunmaz. Ön hazırlık, SQL, imaj, TLS ve dağıtım günlükleri burada korunur. Canlı erişim geri geldiğinde önce 404 ve Scheduler hatası incelenmeli; sağlıklı geçit/kalp atışı doğrulanmadan VM devralma yapılmamalıdır.

**Bu çalışmada hiçbir VM açma/kapatma işlemi yapılmadı.** A100 ve eski CPU/uyandırma sunucularının önceki çalışma durumu sürdü; otomatik devralma açılmadı ve eski kapının güç yolları henüz değiştirilmedi. Bu nedenle bütün sunucuların yalnız ücretli istemle çalıştığı iddiası hâlâ doğrulanmış değildir. PostgreSQL, yedekler, diskler ve kontrol hizmetlerinin taban maliyeti sürer; model VM'lerinin istemle yönetilmesi bütün altyapının sıfır maliyetli olacağı anlamına gelmez.

## Doğrulanan sonuç

Gerçek `deepseek-v4.1-flash` modeli, Linux üzerinde gerçek panel HTTP kuyruğu ve dosya inceleme/uygulama akışı üzerinden sekiz dosyalı AI kiralama DEMO projesi oluşturdu. Güvenlik, SQLite veri katmanı, HTTP API, sunucu, HTML, JavaScript, testler ve README model tarafından üretildi. Uygulama kodu test aracı tarafından hazır şablon olarak yazılmadı.

Bu sonuç önceki `.3` sürümündeki gerçek model koşusudur; `.4` sürümünde yeni canlı ticari geçitten tekrar edilmedi.

İlk denemeler başarısız oldu; aşağıdaki panel hataları ve modelin sunucu hataları giderilerek devam edildi. Kabul edilen dosyalar açık test checkpoint’iyle korundu; son PASS bütün dosyaların tek çağrıda veya ilk denemede üretildiği anlamına gelmez. Son gerçek model testinin sonucu **PASS, 113,968 saniye, çıkış 0**. Test: `TestRealModelPanelBuildRentalProject`.

Bağımsız kabul testi modelin çalışma alanının dışında tutuldu. Gerçek yerel HTTP istekleriyle şunları doğruladı:

- PBKDF2 parola özeti, benzersiz salt ve yanlış parolanın reddi.
- Kayıt, giriş, oturum zorunluluğu ve parolanın yanıtta açığa çıkmaması.
- Görev ekleme/tamamlama ve kullanıcılar arası okuma/değiştirme izolasyonu.
- Model kataloğu, 10 eşzamanlı aynı request_id için tek kredi düşümü.
- Store yeniden oluşturulduğunda bakiye kalıcılığı; bakiye tükenince HTTP 402.
- Ana sayfa ve JavaScript dosyası için HTTP 200; dizin dışına erişim için HTTP 404.
- Modelin ürettiği yedi Python testinin panelden çalıştırılması; ardından bağımsız kabul testinin tekrar geçmesi. Kaynaklar indirilip yerelde Python 3.12 ile bağımsız kabul ve yedi proje testi yeniden geçti; JavaScript sözdizimi denetimi de geçti.

Bu proje gerçek ödeme veya model hizmeti satmaz; istem yanıtları açıkça DEMO’dur. Projeyi oluşturan model bağlantısı gerçektir. Projenin kendi istem uç noktası gerçek AI sağlayıcısına bağlı değildir. Bu test Windows’ta yeni EXE testi veya canlı Shopify ödeme testi değildir.

## Eklenen ve onarılan davranışlar

| Alan | Son davranış |
| --- | --- |
| Bağlam ve dosya okuma | Gerçek model bağlamı metadata’dan alınır; gözlenen model 8.192 bağlam bildirdi. Sayfalı dizinler, güvenli alt dizin listeleme, UTF-8 bayt/satır parçaları, hash ve devam ofseti kullanılır. |
| Büyük satır aralığı | Bütçeyi aşınca tüm okuma durmaz; `truncated`, `next_offset`, `eof` ile sınırlı parça döner. Gerçek JSON boyutu ölçülür; gerekirse daha küçük parça denenir. |
| Okuma/yama kanıtı | Aynı hash’e ait bitişik okuma parçaları birleştirilebilir. Okunmamış boşluk veya farklı sürümler birleştirilmez. Tam okunmamış dosya bütünüyle yazılmaz; gözlenen benzersiz bölüm yamalanabilir. |
| İstek ayrımı | Yeni proje şartnamesindeki “eksik oturum” ifadesi onarım isteği sanılmaz. İlk okuma aşaması, görevde belirtilen dosya ve satır aralığını istemeye yönlendirilir. |
| İşlem protokolü | Dosya değiştiren görevlerde JSON işlemleri kullanılır. Hatalı işlem şeması için en fazla bir ek düzeltme çağrısı yapılır ve kayda geçer. Bu ek çağrı da ticari kullanımda ayrı model çağrısıdır. |
| Çok dosyalı uygulama | Sürüm ve yol ön kontrolü, özel yedek, işlem günlüğü, çakışmayı koruyan geri alma ve yeniden başlatmada kurtarma sürer. Python varsa önerilen bütün Python dosyaları ilk yazmadan önce izole AST denetiminden geçer. |
| Görevler | Kalıcı kuyruk, payload/istek eşleştirmesi, tek işçi, geçmiş, inceleme, iptal ve açık devam korunur. Testler gerçek süreç, süre ve çıkış sonucuyla raporlanır. |
| Ticari EXE | Derleme bayrağıyla HTTPS hesap geçidi ve oturum zorunludur. Yerel tünel, gömülü anahtar veya ortam değişkeni bu kontrolü kapatmaz. Model/sunucu yönetim anahtarları EXE’ye konmadı. |
| Panel açılışı | Hiçbir sürüm açılışta SSH başlatmaz. SSH düğmesi yalnız açık geliştirici modu içindir. Doğrulanmamış 15/70 dakika kapanış vaatleri kaldırıldı. |
| Hesap görünümü | Sunucu durumu hesabın geçidinden alınır; modelin hazır olması VM’nin RUNNING durumundan ayrı değerlendirilir. Ticari kullanımda yerel saat/abonelik sayacı ayrıca kredi düşmez; model çağrısı başına kredi sonucu geçitte tutulur. |
| SQL kalıcılığı | Yerelde SQLite; çok süreçli üretim için ortak PostgreSQL adaptörü. İşlemler kilitlenir; DDL başlangıcı seri uygulanır. Sayısal bakiye JSON’a int olarak döner, SQLite rowid bağımlılığı kaldırıldı. |

Çok dosyalı uygulama işletim sistemi düzeyinde tek bir dosya sistemi işlemi değildir: dış süreçler ara durumu görebilir. Başarısızlıkta geri alma ve çökme günlüğü vardır; dış editör çakışması korunur. Büyük metinlerin modelle özetlenmesinde anlamsal bilgi kaybı hâlâ mümkündür; modelin gerçek bağlam sınırı büyütülmüş veya sınırsız yapılmış değildir.

## İstem başına sunucu yönetimi

`inference_lifecycle.py`, `managed_inference.py`, `gce_provider.py`, `lifecycle_config.py` ve `runtime.py` eklendi.

- Yalnız geçerli kredi rezervasyonu bulunan model isteği işçi açabilir. Hesap, sağlık, model listesi, sunucu durumu ve panel açılışı açılış yapmaz, boşta kalma süresini yenilemez.
- Açılış niyeti ve kullanım kayıtları ortak veritabanında tutulur. Eşzamanlı denetleyiciler için sürüm kontrolü vardır; eski bir bulut durum yanıtı ikinci açılış üretemez. Yanıtı belirsiz güç işlemi aynı operation ID ile izlenir.
- Varsayılan: en fazla 10 etkin rezervasyon, 480 saniye açılış, toplam 600 saniye istek, 60 saniye yenilenmeyen kullanım kaydının düşmesi, son etkin istekten sonra 120 saniye boşta kalma eşiği.
- Yeni işe kabulde 1.800 saniye çalışma sınırı ve günlük 3.600 saniye tahmini çalışma/6 açılış sınırı kontrol edilir. Bunlar fatura tutarı tavanı değildir.
- Ayrı denetleyici iki dakika içinde doğrulanamazsa yeni işler engellenir. Sağlık sorgusu denetleyici onayı sayılmaz. Hatalı denetim, yeni açılışları onaylayan kalp atışı üretmez.
- Etkin istek varken normal boşta kapanışı yapılmaz. Son istek, iptal veya bağlantı kesilmesi kullanım kaydını bırakır. Kayıt bırakma işlemi bir kez yapılır; veritabanı arızasında süreli kayıt kendiliğinden geçersiz olur.
- Gerçek model hazır olmadan kullanıcı istemi modele gönderilmez. Soğuk açılışta gönderilen SSE yorumları model cevabı veya ücretlendirme başarısı sayılmaz. Başarısız/eksik akış krediyi bir kez iade eder.
- Üretimde yalnız özel TLS model adresi, model kimlik doğrulaması ve doğrulanmış VM kimliği kullanılır. Sağlayıcının otomatik yeniden başlatması kapalı olmalıdır. Kontrol dışında çalıştırılmış VM kendiliğinden devralınmaz.
- Yerel SSD verisi asla discard edilmez. Koruyarak durdurma başarısız olursa veri atan alternatif denenmez.

Kodda bu senaryolar test edilmiştir. **Gerçek sağlayıcı stop/start çevrimi, canlı boşta kapanışı ve bütün VM’lerin yeni politika altında olması doğrulanmamıştır.**

## Canlı altyapı denetimi ve geçiş engelleri

Proje: `project-27082a40-6083-4952-9db`.

| VM | Bölge | Gözlenen durum / rol |
| --- | --- | --- |
| ds-a100 | europe-west4-b | RUNNING; gerçek model, A100 40 GB; 4 yerel SSD toplam 1.500 GB |
| instance-20261030-232245 | us-central1-c | RUNNING; usta.service CPU kodlama işçisi |
| usta-wake | us-central1-c | RUNNING; eski HTTP/uyandırma kapısı |
| usta-l4 | us-central1-a | TERMINATED |

Üç VM’nin açık olması ve eski kapının `/health`, `/v1/`, `/static/` gibi yolları da uyandırma adayı sayması, “bütün sunucular yalnız istem başına açılıyor” koşulunun eski canlı kurulumda sağlanmadığını gösterir. Eski kodun `ensure` işlevi diğer GPU’ları durdurma yolu da içerir. Yukarıdaki yeni kaynaklar ve dar IAM kuruldu; VM güç durumu değiştirilmedi ve müşteri siparişi/ödeme kaydı oluşturulmadı.

A100’de doğal maksimum çalışma süresi ayarı yoktu; otomatik yeniden başlatma kapalıydı. Yerel SSD nedeniyle veri atan kapatma veya guest shutdown kullanılmadı. Varsayılan VM kimliğiyle bulut okuma ve belirli A100 için izin kontrolü HTTP 403 verdi. Yeni dar izinli hizmet kimliği doğrulanmalıdır.

Dağıtımın kaynakları, Dockerfile, sabit bağımlılıklar, iki YAML, Cloud Build ve operatör betikleri `kanus/deploy` altındadır. Platformda iki hizmet Ready=True görüldü; bu işaret gerçek ticari HTTP/soğuk açılış kabulü yerine geçmez.

Eski kapıyı güvenli geçişten sonra güç işlemi yapamaz hale getiren `legacy_wake_migration.py` hazırdır. Varsayılan dry-run’dır; beklenen hash, özel yedek ve açık apply gerekir. Canlı kaynak üzerinde bellekte dönüşüm doğrulandı; dosya ve servis değiştirilmedi. Uygulama sırası ve kabul adımları `kanus/deploy/README.md` içindedir.

Kullanıcının açık canlı kurulum onayı alınmıştır; aynı kapsam için yeni onay beklenmez. Kalan engel onay değil, bağlı tarayıcı erişimi ve henüz geçmeyen canlı kabul kontrolleridir. Sağlayıcı/IAM arızasında kapanış garanti edilemez. Duran VM diskleri, korunan SSD verisi, IP’ler ve kontrol hizmetleri ayrıca ücret doğurabilir.

## Ticaret ve güncel fiyatlar

Bağlı mağaza: `shop.kanus.org`; webhook için doğrulanan kanonik mağaza: `19einb-qz.myshopify.com`. Bu oturumda ürünler yalnız okundu. Çalışma sırasında yapılan mağaza değişiklikleri korunarak yerel katalog eşitlendi.

| Paket | Gözlenen fiyat | İstek başına | Varyant | Statü |
| --- | ---: | ---: | --- | --- |
| 1 model isteği | 250 TL | 250 TL | 68119764304100 | ACTIVE |
| 5 model isteği | 1.400 TL | 280 TL | 68119764336868 | ACTIVE |
| 10 model isteği | 2.600 TL | 260 TL | 68119790026980 | ACTIVE |

Kodlar 90 gün; bir kredi bir model API çağrısı. Tek görev okuma, devam ve düzeltme nedeniyle birden fazla kredi kullanabilir. Geçit 32 KiB metin ve 2.048 çıktı tokenı sınırını uygular; gerçek model bağlamı ayrıca geçerlidir. Beş/onluk paketler mevcut mağaza fiyatlarıyla tekli alıma göre birim indirim sağlamıyor; fiyatlar burada değiştirilmedi.

Ürün açıklamalarında test taslağı ifadesi hâlâ var. ACTIVE statüsü satış/lisans zincirinin doğrulandığı anlamına gelmez. Yeni fiyata göre gelen imzalı siparişin eski fiyat listesinden reddedilmemesi için `credit_products.example.json` güncellendi. Gerçek ödeme veya sipariş oluşturulmadı.

Shopify HMAC/imzalı webhook, tekrar önleme, iptal/tam iade, lisans, teslim kuyruğu, hesap, kredi yükleme ve kota davranışları yerel testlerde vardır. Önceki sürümün gerçek modelle sentetik imzalı sipariş→hesap→SSE→tek kredi→kota engeli testi de geçmişti. **Gerçek Shopify siparişi→canlı webhook→destek posta kutusundan lisans→yeni Windows hesap girişi zinciri bu sürümde doğrulanmadı.** Canlı geçitte `US_SHOPIFY_ENABLED=false`; gerçek imza anahtarı bağlanana kadar webhook 503 verir. Sağlık cevabı bu durumu açıkça bildirir; sağlık sorgusu model açmaz. Anahtarı eksik Shopify etkinleştirmesi üretimde reddedilir.

Verilen Hostinger anahtarıyla yalnız kimlik/mailbox keşfi denemesi yapıldı; `/api/v1/me` HTTP 403 döndü. Bu sonuçla gönderim yetkisi veya support@kanus.org posta kutusu kimliği doğrulanmış sayılmadı. Gerçek e-posta gönderilmedi; anahtar kaynak veya EXE’ye eklenmedi. Sonucu belirsiz e-postalar ve ani süreç ölümü sonrası sonucu belirsiz rezervasyonlar operatör uzlaşması ister; körlemesine yeniden gönderilmez.

## Son test ve derleme kanıtı

| Kontrol | Sonuç |
| --- | --- |
| Güncel `go test -race ./...` | PASS; 36,694 sn. Yeni ticari HTTPS varsayılanı/HTTP geri dönüşünün engellenmesi dahil. Önceki sürümün ayrıntılı JSON sayımı 222 PASS, 5 opt-in SKIP idi. |
| `go vet ./...` | Çıkış 0 |
| `node check_ui.mjs` | Gömülü JavaScript sözdizimi ve statik DOM bağları PASS; görsel tarayıcı testi değildir |
| `python3 -m unittest discover -q` | 99 test toplandı: 62 PASS, opt-in PostgreSQL/Gunicorn 37 SKIP; 4,549 sn |
| Gerçek PostgreSQL 17.11 + Gunicorn 26.2.0 | Ayrı şemalarda `test_postgres_contract test_production_runtime`: 37 PASS, 0 FAIL; 7,698 sn. Shopify gerçek anahtarı olmadan açıkça kapalı factory, sağlık, webhook 503, hesap ve iade de doğrulandı. |
| Gerçek model + panel kuyruğu + proje kabulü | PASS, son devam/onarım koşusu 113,968 sn; oluşan sekiz dosya ve bağımsız kabul çıktısı proje ZIP’inde |
| Windows x64 GUI çapraz derleme | İki EXE üretildi; PE makine x86-64 ve GUI alt sistemi doğrulandı; yeni sürüm Windows’ta çalıştırılmadı |

Atlanan Go kontrolleri: PHP yorumlayıcısı gerektiren lisans eşleşmesi, açıkça etkinleştirilmeyen genel canlı model testi, önceki iki dosyalı gerçek model onarımı, kapsamlı gerçek model testi (ayrıca uzak makinede çalıştırıldı) ve dış npm erişimi isteyen resmi REA entegrasyonu. REA’nın ağ erişimi ortam politikasıyla engellendiği için bu kontrol varsayılan yerel regresyonda açık opt-in oldu; yerel kapsam/gizli değer izolasyon testleri çalıştı.

Go derleyicisi resmi arşiv SHA-256’sı doğrulanmış Go 1.27.3’tür. Bağımlılıklar `gunicorn==26.2.0`, `psycopg[binary]==3.3.8` olarak sabitlendi. PostgreSQL test verisi geçici, ayrı şemalardaydı; müşteri verisi kullanılmadı.

Ticari EXE hesap/HTTPS şartını derleme bayrağıyla zorlar. Geliştirici EXE yerel gerçek model testine izin verir. İkisinde de açılışta otomatik SSH yoktur. EXE’ler bu ortamda derlenmiştir; Authenticode imzalama ve yerel Windows çalışma testi yapılmamıştır.

Windows kaynak derleme betiği `kiyas/term/US_DERLE.bat` geliştirici test EXE'si üretir; `US_DERLE.bat ticari` ticari adayı ayrı adla ve aynı HTTPS varsayılanıyla derler. Kaynak ZIP'inde git metadata'sı gerektirmemesi için `-buildvcs=false` kullanılır. Bu betik Windows üzerinde henüz çalıştırılmadı.

## Kalan üretim kabulü

1. Oluşturulan kontrol hizmetlerinin dış 404 sorununun çözümü; gerçek Scheduler/reconcile kalp atışı, Cloud Run IAM/model erişimi, alarm ve yedekten geri yükleme kabulü.
2. A100 yerel SSD verisini koruyan gerçek stop/start, modelin soğuk yüklenmesi, yetkili istek ve son istekten sonra gerçek kapanışın gözlenmesi. CPU/L4 ve eski kapının da yeni politikaya geçirilmesi.
3. Gerçek mağaza test siparişi, uygulama HMAC’i, webhook tekrarı, tek lisans/e-posta, iptal/iade ve gerçek hesap/kota zinciri.
4. Yeni ticari EXE’nin gerçek Windows’ta açılması; gerçek ticari model geçidinden aynı geniş proje ve kesinti/iptal senaryolarının tekrar edilmesi.

Bu dört kabul tamamlanmadan ticari üretim sürümü etiketi verilmemelidir.

## Derleme dosyaları

- `US_Ticari_Aday_2026_11_05_4.exe`: 14597632 bayt; SHA-256 `fd16074c40bdc1ab434cb4473f6238cf233511792e1e8fdc945f2640a3493129`.
- `US_Gorev_Kota_Test.exe`: 14597632 bayt; SHA-256 `db911b11597237a6202c7cfd91fd51e96d56da7214e503a26f11d51fe9803edc`.
