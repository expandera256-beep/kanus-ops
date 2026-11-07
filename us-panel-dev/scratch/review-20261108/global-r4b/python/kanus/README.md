# Kanus ticari geçidi — yerel kurulum ve doğrulama

Bu sürüm bir geliştirme sürümüdür. Gerçek ödeme, gerçek webhook teslimi, e-posta ve Windows uygulaması birlikte doğrulanmadan üretime hazır sayılmaz. `commercial_server.py` yalnızca `127.0.0.1` üzerinde dinler. Testlerde kullanılan siparişler ve alıcılar sentetiktir.

## Testler

Bu klasörde `python3 -m unittest discover -v` çalıştırın. Go paneli için `../kiyas/term` içinde `go test -race ./...`, `go vet ./...` ve `node check_ui.mjs` kullanın. Python, Go ve Node.js'in kurulu olması gerekir; bağımlılıklar veya GPU kaynakları otomatik satın alınmaz.

Var olan yerel modele karşı isteğe bağlı test:

```sh
python3 smoke_real_model.py http://127.0.0.1:8080 deepseek-v4.1-flash
```

Bu komut geçici veritabanında sahte imzalı sipariş, hesap ve tek kredi oluşturur; gerçek model isteği yapar, krediyi doğrular. Shopify'a istek veya gerçek e-posta göndermez. Panelin gerçek model ile iki dosyalı onarım testi `kiyas/term/real_model_e2e_test.go` içindedir; yalnız `US_REAL_MODEL_TEST=1`, `US_REAL_MODEL_URL` ve `US_REAL_MODEL_NAME` açıkça ayarlandığında çalışır.

## Servis yapılandırması

| Değişken | Anlamı |
| --- | --- |
| `US_COMMERCE_DB` | Yerelde mutlak özel SQLite yolu; üretimde TLS doğrulamalı ortak PostgreSQL DSN. Mevcut müşteri veritabanını testlerde kullanmayın. |
| `SHOPIFY_WEBHOOK_SECRET` | Webhook'u gönderen Shopify uygulamasının HMAC sırrı; Hostinger anahtarı değildir. |
| `US_LICENSE_SECRET` | Lisans üretimi ve doğrulaması için aynı gizli değer; kaynak koda yazmayın. |
| `SHOPIFY_SHOP` | Webhook başlığındaki gerçek, kanonik mağaza alan adı. |
| `US_MODEL_UPSTREAM` | Mevcut modelin loopback HTTP kökü, ör. `http://127.0.0.1:8080`. |
| `US_CREDIT_PRODUCTS` | Doğrulanmış varyant/fiyat izin listesi JSON'u; örnek dosya `credit_products.example.json`. |
| `US_COMMERCE_PORT` | İsteğe bağlı yerel port; varsayılan `8888`. |

Değerler süreç ortamına güvenle verildikten sonra `python3 commercial_server.py` çalıştırılır. Üretim için TLS, ters vekil, dış erişim sınırları, yedekler ve servis gözetimi ayrıca kurulmalıdır. Model sunucusunu müşterilere doğrudan açmayın. Bu dosyanın okunması veya örnek yapılandırmanın bulunması dağıtım yapmaz.

Masaüstünde geçit adresini HTTPS adresine ayarlayın, gerçek model adını girin ve lisansla hesap açıp giriş yapın. Oturum anahtarı geçide bağlıdır. Farklı geçide taşınmaz. `GET /v1/models` kimlik doğrulaması ister, gerçek üst sunucu bilgisini sınırlandırarak döndürür ve kredi kullanmaz.

## Webhook, lisans ve e-posta

`POST /webhooks/shopify` için `orders/paid`, `orders/cancelled` ve `orders/updated` konuları desteklenir. Ham gövde HMAC'i ve mağaza alanı doğrulanır. Yeni paketler yalnız izin listesindeki tam fiyatlarla işlenir; indirimler ve kısmi iadeler otomatik işlenmez. Taslak Shopify ürünleri satışa çıkmış sayılmaz.

E-posta ayrı `python3 mail_worker.py` süreciyle gönderilir. SMTP için `US_MAIL_PROVIDER=smtp` ve `US_SMTP_HOST`, `US_SMTP_PORT`, `US_SMTP_USER`, `US_SMTP_PASSWORD`, `US_SMTP_FROM` gerekir. Hostinger adaptörü için `US_MAIL_PROVIDER=hostinger_api`, `US_HOSTINGER_MAIL_TOKEN` ve gerçek `US_HOSTINGER_MAILBOX_ID` gerekir. Bir e-posta adresi tek başına posta kutusu kimliği değildir. Teslimi belirsiz gönderimler inceleme bekler; körlemesine tekrar gönderilmez.

## Kredi ve kesinti yönetimi

Bir kredi bir model API çağrısıdır. Bir masaüstü görevi okuma, devam ve onarım için birden fazla kredi kullanabilir. Geçit en fazla 32 KiB mesaj metni ve 2.048 çıktı tokenı kabul eder; modelin gerçek bağlamı ayrıca geçerlidir. Aynı anda en fazla 10 model rezervasyonu açılır; yerel HTTP sunucusu hesap isteklerini ayrıca karşılar. Bu, 10 gerçek GPU işinin performans garantisi değildir.

Başarısız veya eksik akıştaki rezervasyon bir kez iade edilir. Başarı, son akış parçası kullanıcıya verilmeden önce kaydedilir. Aynı istek kimliği yeniden çalıştırılmaz. Sunucu aniden kapanırsa sonucu belirsiz rezervasyonlar otomatik oynatılmaz veya iade edilmez; yerel operatör incelemesi gerekir:

```sh
python3 metering_admin.py /mutlak/yol/commerce.sqlite
python3 metering_admin.py /mutlak/yol/commerce.sqlite --refund REZERVASYON_KIMLIGI
```

İkinci komut yalnız inceleme sonrasında seçilen bekleyen rezervasyonu bir kez iade eder; tamamlanmış çağrıyı değiştirmez. Bu araç herkese açık bir yönetim API'si değildir. Süre kontrolü akış okumaları arasında yapılır; kesin kaynak izolasyonu, işletim sistemi bellek sınırları ve katı 10 dakika hedefi bu sürümde garanti edilmez.


## 2026.11.05.3 — istem başına işçi yönetimi

`inference_lifecycle.py` kalıcı kullanım kayıtları ve açılış niyeti, yarış koruması, süre/başlatma sınırları ve bağımsız boşta kalma denetimi sağlar. `managed_inference.py` kredi rezervasyonu doğrulanmadan VM açmaz; soğuk açılış sırasında akış yorumları gönderir. Bunlar model cevabı sayılmaz. `GET /v1/models` ve kimlik doğrulamalı `POST /account/workers` kapalı işçiyi uyandırmaz. Kapanış yalnız etkin kullanım yokken yapılır; yerel SSD atılmaz.

Üretim için Gunicorn factory'leri, min=0 Cloud Run tanımları ve uygulama sırası [deploy/README.md](deploy/README.md) içindedir. Hazırlanmış dosyalar canlı dağıtım sayılmaz. A100 yerel SSD koruması, ayrı IAM kimliği, bağımsız Scheduler, özel model TLS'i ve gerçek stop/start çevrimi üretim geçişinde doğrulanmalıdır. Yeni ticari EXE'de hesap/HTTPS geçidi zorunludur; ortam değişkeni veya yerel tünel bu denetimi kapatmaz. Geliştirici sürümü yerel test için korunur; hiçbir sürüm panel açılışında SSH başlatmaz.

Gerçek PostgreSQL sözleşmeleri ayrı test veritabanında `US_POSTGRES_TEST_DSN` ile `python3 -m unittest test_postgres_contract test_production_runtime -v` kullanılarak çalışır. Gerçek modelin panelden kapsamlı proje üretme testi `US_REAL_PROJECT_TEST=1` ile `kiyas/term/real_project_build_test.go` içindedir; kendi bağımsız HTTP kabul testi modelin çalışma alanının dışındadır.

5 Kasım 2026 canlı mağaza kontrolünde 1/5/10 istek ürünleri ACTIVE ve fiyatları 250/1.400/2.600 TL olarak görüldü; yerel katalog buna eşitlendi. Ürün açıklamalarında test taslağı ifadesi hâlâ var. Statünün ACTIVE olması ödeme/lisans zincirinin doğrulandığı anlamına gelmez. Bir kredi bir model API çağrısıdır; tek görev okuma, devam ve protokol düzeltmesiyle birden fazla çağrı kullanabilir.
