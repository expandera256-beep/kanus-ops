# İstem başına dağıtım — kaynaklar kuruldu, canlı kabul eksik

Kullanıcı canlı kurulum için açık onay verdi. 5 Kasım 2026'da `us-commerce-pg` PostgreSQL 17, `us-commerce` ve `us-watchdog` Cloud Run hizmetleri, dakikalık OIDC `us-idle-watchdog` görevi, dar hizmet kimlikleri ve özel A100 TLS dinleyicisi kuruldu. İki Cloud Run hizmeti Ready=True gözlendi; dış `/healthz` kontrolleri 404 döndü. Ardından bağlı Chrome oturumuna erişim kesildi. Yeni onay beklenmiyor; erişim geri geldiğinde önce bu hata ve gerçek Scheduler kalp atışı doğrulanmalıdır. Ayrıntılı durum `US_KAPSAM_VE_FIYAT_RAPORU.md` dosyasındadır.

Geçit URL'si: `https://us-commerce-iqhblf63dq-ez.a.run.app`; özel denetleyici: `https://us-watchdog-iqhblf63dq-ez.a.run.app`. Bunlar henüz dış erişim kabulü geçen hizmetler değildir. İmaj digest'i: `sha256:0da920e46ecd6392e234847ca9e87f3dc4bdb64165deccc51a50a869ade1b70e`. Hiçbir VM güç işlemi yapılmadı; eski sunucuların geçişi tamamlanmadı.

`gcp_apply.py` açık operatör aşamalarıyla sırları, sınırlı PostgreSQL rolünü, Cloud Build imajını ve dağıtımı yönetir; kendisi hiçbir model VM'sini açıp kapatmaz. Cloud Shell çalışma dizini `$HOME/us-live-20261105/release` içindedir. `private/` bu dizinde tutulur ve kaynak paketine/uygulama imajına alınmaz. Yeni bir dizinde çalıştırmadan önce mevcut özel operatör durumu ve sırlar uzlaştırılmalıdır; mevcut sırları körlemesine yenilemeyin.

Cloud Shell'de ilk aktarılan operatör betiği bulunur. Bu kaynak paketindeki son operatör sürümü, başarılı geçişte kullanılan Cloud Build yolunu ve mevcut sırla eşleşme kontrolünü de içerir; yeniden imaj kurmadan önce bu betik ve `cloudbuild.yaml` güncel paketten eşitlenmelidir. Bunlar uygulama container'ına veya EXE'ye giren değişiklikler değildir.

`worker_boot_guard.py`, denetlenmiş açılış betiğindeki otomatik biçimlendirmeyi hash korumasıyla kaldırdı. `worker_tls_setup.py` mevcut nginx'e ayrı özel dinleyici ekledi: 401/200 doğrulandı. `worker_preservation_probe.py capture` yedi model dosyasının boyut/örnek hash ve canary tabanını kaydetti; `verify` gerçek stop/start sonrasında henüz çalıştırılmadı. Bu kontrol tam dosya hash veya yedekten geri yükleme testi değildir.

5 Kasım 2026 denetiminde dört GCE VM bulundu. `ds-a100`, `instance-20261030-232245` (CPU model işçisi) ve `usta-wake` (eski kapı) çalışıyordu; `usta-l4` durmuştu. Dolayısıyla mevcut kurulum tüm sunucuların istem başına açılıp kapandığını sağlamıyor. Bu dosyalar tek başına mevcut bulut yapılandırmasını değiştirmez.

## Tasarım ve sınırlar

Public `us-commerce` Cloud Run hizmeti ve yalnız Cloud Scheduler kimliğinin çağırabildiği `us-watchdog` Cloud Run hizmeti minimum sıfır örnekle yapılandırıldı. PostgreSQL ortak, kalıcı rezervasyon/kredi/işçi durumunu tutar. SQLite geçici Cloud Run diski üzerinde üretim için reddedilir. Onaylanan Cloud SQL tek bölgeli 1 vCPU / 3.75 GiB Enterprise, 20 GB SSD, 60 GB büyüme üst sınırı, silme koruması, günlük yedek ve 7 günlük PITR ile kuruldu; HA değildir. Her altyapı bileşeninin her durumda sıfır maliyetli olduğu iddia edilmez.

Yalnız kredi rezervasyonu bulunan `/v1/chat/completions` çağrısı `instances.start` yapabilir. Sağlık, hesap, model listesi ve panel açılışı işçi açmaz ve boşta kalma süresini yenilemez. Aynı anda en fazla 10 rezervasyon; 8 dakika açılış, toplam 10 dakika istek süresi; 60 saniye yenilenmeyen kullanım kaydı düşer. Son etkin istek bitince 120 saniye boşta kalma eşiği vardır. Dakikalık bağımsız denetleyici nedeniyle kapanışın başlatılması buna yaklaşık bir denetim aralığı ve bulut işlem süresi ekleyebilir.

30 dakika çalışma sınırı için yeni işler önceden reddedilir; etkin kullanım yoksa denetleyici süre sınırında da kapatır. Günlük 1 saat tahmini çalışma ve 6 açılış koruması varsayılandır; bunlar fatura tutarı veya kesin maliyet tavanı değildir. Denetleyici iki dakika doğrulanamazsa yeni model istekleri engellenir. IAM veya sağlayıcı arızası sırasında kapanış garanti edilemez; 503 ve `inference_events` için operatör alarmı gerekir. Çalışırken kontrol dışı açılmış VM otomatik sahiplenilmez veya kapatılmaz.

## Mevcut A100 için zorunlu geçiş

`ds-a100` üzerinde toplam 1.500 GB yerel SSD var. Konuk işletim sisteminden kapatma, `discardLocalSsd=true`, disk silme veya yerel SSD verisini atan doğal maksimum çalışma süresi ayarı kullanılmaz. Adaptör yalnız beta API ile `discardLocalSsd=false` gönderir; korunma desteklenmezse kapanışı reddeder. Korunan verinin depolama maliyeti sürer. Spot kesintileri için bunun yedek yerine geçtiği varsayılmaz.

Üretim öncesi disk kullanımı, model dosyalarının konumu ve geri yükleme yöntemi denetlenmeli; yedek/persistan model kopyası doğrulanmalı. Test trafiği boşaltıldıktan sonra veri koruyarak gerçek stop/start çevrimi ve modelin yeniden yüklenmesi izlenmelidir. Çalışan A100'ü otomatik devralan bayrak üretim yapılandırmasında reddedilir. Kaynak kimlikleri eşleşmeden hiçbir güç işlemi yapılmaz.

## Hazır dosyalar

| Dosya | İşlev |
| --- | --- |
| `workers.example.json` | Denetlenmiş A100 kimliği; özel TLS adresi doldurulmalı. Diğer işçiler ancak kimlik, model ve disk politikaları ayrı doğrulandıktan sonra eklenir. |
| `gateway.service.yaml` | 0–2 örnek, 32 HTTP eşzamanlılığı, 600 saniye istek sınırı; gizli değerler Secret Manager referansları. |
| `watchdog.service.yaml` | 0–1 örnek, tek denetim çağrısı; ayrı hizmet hesabı. **allUsers invoker verilmez.** |
| `worker.Caddyfile` | İşçinin özel IP'sinde 8443 TLS ve Bearer doğrulaması; loopback 8080 modeline vekil. Sertifika, özel anahtar ve token dışarıdan sağlanır. |

Yer tutucular değiştirilmeden dağıtım yapılmamalı. Özel TLS sertifikası gerçek DNS adıyla eşleşmeli; CA dosyası iki kontrol hizmetine verilmelidir. Modelin doğrudan genel internete açık portu olmamalı; güvenlik duvarı yalnız kontrol hizmetlerinin özel ağından 8443'e izin vermeli. Doğrudan VPC çıkışının bölge/alt ağ değerleri kurulumda doğrulanmalı. Gerçek model anahtarı ve Google yetkileri EXE'ye yazılmaz.

## Kimlik ve hizmet izinleri

Geçide yalnız izinli işçilerde `compute.instances.get` ve `compute.instances.start`; denetleyiciye `compute.instances.get` ve `compute.instances.stop` verilir. `list`, `create`, `delete`, IAM değiştirme veya Editor/Owner rolü gerekmez. Özel rol/koşullu bağlama gerçek kaynak üzerinde izin testiyle doğrulanmalı; SSD koruma API'si 403 verirse geniş rol verilmez. Gözlenen varsayılan VM hizmet hesabının okuma ve A100 testIamPermissions isteği 403 dönmüştür; üretim kimliği olarak çalıştığı varsayılmamalı.

Her hizmet yalnız ihtiyaç duyduğu Secret Manager kayıtlarını okuyabilir. Scheduler için ayrı hesap, sadece `us-watchdog` hizmetinde `roles/run.invoker` alır. Bulut platformunun IAM doğrulaması iç uç noktayı korur; bu uç nokta kimliksiz bir ters vekile taşınmamalı. Kurulu PostgreSQL `ENCRYPTED_ONLY` kullanır. Cloud SQL Auth Proxy sunucu kimliğini doğrular ve uzak taşıma TLS'ini sağlar; uygulama Cloud Run'ın yerel Unix soketine bağlanır. Sınırlı `us_runtime` rolü yalnız `us_app` şemasını kullanır. Gerçek yedekten geri yükleme henüz denenmedi.

## Uygulama ve kalan kabul sırası

1. Var olan müşteri verisinin yedeğini al; test verisini ayrı PostgreSQL şemasında tut. İmajı Dockerfile'dan oluştur, tarat, değişmez digest ile iki YAML dosyasına yaz.
2. Ayrı hizmet hesapları, dar izinler, sırlar, CA ve özel model TLS bağlantısını hazırla. Sırları komut satırına veya günlük dosyasına yazma. İki hizmeti `europe-west4` bölgesinde min=0 olarak kur.
3. `us-watchdog` için genel invoker bağlaması olmadığını doğrula. Scheduler kimliğine invoker ver. Dakikalık, çakışması sınırlı ve OIDC korumalı görev hazırla:

```sh
gcloud scheduler jobs create http us-idle-watchdog \
  --location=europe-west4 --schedule='* * * * *' --time-zone=UTC \
  --uri="$US_WATCHDOG_URL/internal/reconcile" --http-method=POST \
  --message-body='{}' --oidc-service-account-email="$US_SCHEDULER_SA" \
  --oidc-token-audience="$US_WATCHDOG_URL" --attempt-deadline=300s \
  --max-retry-attempts=1 --min-backoff=60s --max-backoff=60s
```

4. Denetim çağrısının 200 ve güncel kalp atışı verdiğini doğrula. Bildirim/alarm kur. Tüm işçiler durmuşken sağlık/hesap/model listesi sorgularını tekrarla; hiçbir `instances.start` olmamalı.
5. Ayrı test hesabında tek rezervasyon → tek açılış → gerçek model çıktısı → tek kredi düşümü → son istek sonrası kapanış çevrimini gözle. 10 eşzamanlı istek, aynı request_id, bağlantı kesilmesi, süreç yeniden başlatma, kota tükenmesi, IAM reddi ve durmuş Scheduler senaryolarını uygula. Eksik akış iadesini ve SSD verisinin korunduğunu doğrula.
6. Shopify uygulamasının gerçek webhook sırrıyla `orders/paid`, iptal ve iade konularını HTTPS geçidine bağla. Gerçek mağazadan test siparişi, webhook tekrar teslimi, tek lisans üretimi, destek posta kutusundan tek e-posta ve masaüstü hesap girişini doğrula. Doğrulama bitmeden satışa hazır kabul edilmez. Bu çalışmada yeniden denetlenen ürünler kullanıcı tarafındaki değişiklikle ACTIVE ve 250/1.400/2.600 TL durumundaydı; bunlar değiştirilmedi. Güncel fiyatlar `credit_products.example.json` dosyasında bulunur.
7. Yerel Windows'ta yeni imzalı/sürüm denetimli EXE'yi aç; eski süreçle karışmadığını doğrula. Geniş proje oluşturma ve bağımsız kabul testlerini gerçek ticari geçitten tekrarla. Sonra model trafiğini yeni geçide taşı ve eski `usta-wake` güç uç noktalarını kapat. CPU/L4/A100 rol geçişi bitmeden eski kapıyı durdurmak istemleri keser. `legacy_wake_migration.py` eski ensure işlevine güç işlemi yapmayan bir koruma hazırlar; varsayılanı dry-run, uygulaması beklenen SHA-256 ile ve özel yedekle yapılır. Denetlenen canlı sürümde /health, /v1/models ve /static/ yolları da uyandırma adayıydı. Geçiş koruması canlı kaynak üzerinde bellekte doğrulandı; hiçbir dosya veya servis değiştirilmedi. Canlı önceki hash: `23a5e85511b92818747175cba2063ebf49d5faa15910bdb34f0b85acdaf76d69`; koruma sonrası beklenen hash: `b227c9aeb7faee31ba306275eef0ad022dad6553b3b028b67512f9fe7b9ad37b`.

## Çalıştırma ve doğrulama

```sh
pip install -r requirements.txt
python3 -m unittest discover -v
US_POSTGRES_TEST_DSN='postgresql://TEST_DATABASE_ONLY' python3 -m unittest test_postgres_contract -v
gunicorn --bind 127.0.0.1:8888 --workers 1 --worker-class gthread --threads 32 --timeout 630 'runtime:gateway_app()'
```

Üretim factory'si `US_ENV=production`, PostgreSQL, özel TLS modeli, yetkili policy, katalog ve en az 32 karakter lisans/model tokenı yoksa açılmaz. Shopify varsayılan olarak gerçek imza anahtarı zorunlu olacak şekilde açıktır. Mevcut dağıtımda gerçek Shopify imza anahtarı bağlı olmadığı için **açıkça** `US_SHOPIFY_ENABLED=false` ayarlandı: webhook 503 verir; `/healthz` kapalı durumu bildirir ve hiçbir model veya kredi işlemi başlatmaz. Gerçek anahtar bağlanıp canlı teslim doğrulanmadan açılmamalı. E-posta gönderimi için denetleyici ortamına mevcut `US_MAIL_PROVIDER` ve ilgili sırlar eklenebilir. Sonucu belirsiz e-posta gönderimi `review` durumunda kalır. Ani süreç ölümüyle sonucu belirsiz kredi rezervasyonları otomatik tekrar edilmez; `metering_admin.py` ile kontrollü uzlaşma sürer.

Kaynakların kurulması Windows veya canlı ticari kabulün geçtiği anlamına gelmez. Kullanıcının onayı mevcut; eksik gerçek sırların bağlanması, erişimin geri gelmesi ve yukarıdaki canlı kabul sonuçları gereklidir.
