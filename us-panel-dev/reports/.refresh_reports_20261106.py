from pathlib import Path
from datetime import datetime
from zoneinfo import ZoneInfo
import json,re,collections
r=Path('/workspace/us-panel/reports');root=r.parent
stamp=datetime.now(ZoneInfo('Europe/Istanbul')).strftime('%d.%m.%Y %H:%M TSİ (Europe/Istanbul)')
head=f'\n\nGüncelleme: {stamp}\n\nEtiketler: PASS, FAIL, BLOCKED, NOT TESTED. PASS yalnız bu yürütücünün dosya/test/derleme kanıtına dayanır; diğer ajan bildirimleri bağımsız doğrulanmış sayılmaz.\n\n'
hashes=(r/'BUILD_SHA256_20261106.txt').read_text()
artifact='\n```text\n'+hashes+'```\n'
docs={
'MERGE_20261106_TR':'''# 2026.11.05.4 taban birleştirmesi

## Kaynak ve yöntem
- PASS — Yeni kaynak `source/from-demir-20261106/` -> `work2/` olarak kopyalandı; kaynak üzerinde yerinde düzenleme yapılmadı. `work/` yeni taban geldikten sonra değiştirilmedi; önceki turda yapılan iş orada korundu.
- PASS — Yeni Go panelinin hesap oturumu, build-time commercial gate, request_id, bağlam limitleri, gerçek iptal bağlamı, öneri/onay, güvenli değişiklik ve görev kuyruğu kodları korundu. Eski model.go/gui.html/license.go dosyaları toptan üstüne yazılmadı.
- PASS — `licensesrv/`, `cmd/us-license-hook`, `cmd/us-pricing-gen`, ortak `pricing.go`, 8 canlı product_id ve testleri taşındı. Tam dosya listesi: `MERGE_CHANGED_FILES_20261106.txt`.
- PASS — Yeni tabanın onarım sınıflandırıcısı eski yamadan daha kapsamlı olduğundan korundu; `TestCompletionWordDoesNotMakeNewProjectARepair` regresyonu eklendi ve geçti.
- PASS — Yeni GUI korunarak yalnız fiyat paragrafına `{{US_FIYAT}}`, gui.go'ya renderGUIPage eklendi. Lisans TRY alanı/dinamik fiyatlar ve aylık 200 istem eklendi; yeni hesap geçidi kontrolleri korundu. PHP'nin jti/idempotentlik/güvenlik kodu korunup yalnız aylık kota/fiyat varsayılanı eşitlendi; PHP çalıştırması ayrıca NOT TESTED.

## Eşdeğerler ve tercih
| Durum | Alan | Birleştirme kararı / kanıt |
|---|---|---|
| PASS | Kredi | Yeni `kanus/commerce_http.py` model proxy'si reserve_prompt -> upstream -> finish_reservation yapıyor; bu yol tercih edildi. Go hesap oturumu bu geçide giderken ek VPS rezervasyonu YAPMAZ. `TestAccountGatewayDoesNotDoubleBillVPSLedger` geçti. |
| PASS | Eski lisans/SSH yolu | gateway.go creditGate ve 3 oturum giriş yolu taşındı: modelden önce 1 kredi Reserve, başarıda commit, hata/iptalde release. Eksik kredi servisi ticari lisans yolunda fail-closed. Yeni URL yalnız bu alternatif yol içindir. |
| PASS | İptal/iade | Go hook'ta orders/cancelled + refunds/create, HMAC/topic/shop kontrolleri, refund-id ve order tombstone, miktara göre birim lisans iptali, önce gelen olaylar ve başarısız yazma rollback testleri var. Yeni Python cancelled/full-refund davranışı korunmuştur. |
| PASS | GPU | Yeni inference_lifecycle/GCEProvider/managed_inference mimarisi korundu. Yavaş describe/model_ready sonrasında deadline/cancel yeniden kontrolü eklendi; sahte sağlayıcıyla üç regresyon geçti. SSH alternatifine bounded readiness ve güvenli SSH timeout seçenekleri taşındı. |

## Çözülmeyen dağıtım çatışmaları
- BLOCKED — Go hook JSON ile Python SQL iki bağımsız yetki deposudur. Python AccountStore.register/redeem `issued_licenses` satırı ister; Go hook onu yazmaz. Anahtar biçiminin us1 olması tek başına hesap kaydını sağlamaz. Tek yetkili kayıt/depo veya açık, testli adaptör seçilmeli; aynı sipariş iki tarafta lisanslanmamalı.
- BLOCKED — İstenen yeni EXE commercial kanalında derlendi; hesap/model geçidi bekler. Go hook `/account/*` ve `/v1/chat/completions` sağlamaz. Sadece api-test Go hook kurulumu, yeni EXE için tam backend değildir.
- BLOCKED — Python webhook yolu `/webhooks/shopify`; topicler orders/paid, orders/cancelled, orders/updated üzerinden tam iade. Python refunds/create/kısmi iade uyarlaması yapılmadı. Go hook'taki iade Python hesaplarını iptal etmez.
- BLOCKED — Go katalog product_id/kind/plans ve US_PLANS_JSON kullanır. Python US_CREDIT_PRODUCTS için **variant_id**, TRY, price_minor ve valid_days ister. Test mağazası USD ürünleri Python TRY doğrulamasından doğrudan geçmez; doğrulanmış variant-id ve ayrı test kataloğu/politikası gerekli. Product ID'leri variant ID diye varsaymayın.
- PASS — Go US_PLANS_JSON parser'ı hem test kind şemasını hem canlı plan=credit/monthly/hourly + price_try şemasını kabul edecek şekilde düzeltildi; iki gerçek rapor dosyasıyla test geçti. Python eski aylık ürün varsayılanı 500'den 200'e ve test bakiyesi 199'a eşitlendi; mevcut SQL hesapları migrate edilmedi.
- BLOCKED — Python runtime: SHOPIFY_SHOP, US_COMMERCE_DB, US_CREDIT_PRODUCTS, US_ENV=production, US_WORKERS_CONFIG, US_MODEL_API_TOKEN, US_MODEL_CA_FILE; Go: SHOPIFY_SHOP_DOMAIN, US_STORE_PATH, US_PLANS_JSON, US_SMTP_ENABLE. Bunlar birbirinin yerine geçmez.
- BLOCKED — Cloud Run metadata/workload identity, PostgreSQL, özel model TLS, OIDC korumalı watchdog/Scheduler beklentileri VPS'ye dosya kopyalamakla oluşmaz. Hiçbir Cloud Run/GCP dağıtımı veya canlı çağrı yapılmadı.
- NOT TESTED — Yeni kaynak deploy/README'si geçmiş Cloud Run health 404 ve tamamlanmamış worker geçişi bildiriyor. Bu tarihsel iddiadır; bu tur canlı durum doğrulanmadı. Diğer ajanın “GPU istem-başı çalışıyor” bildirimi de canlı kanıt değildir.

## Sürüm / sonuç
- PASS — buildVersion ve US_DERLE.bat çıktısı `2026.11.06.3`; Windows commercial build gate korunmuş, sırlar gömülmemiştir. Yeni EXE'nin varsayılan public URL'si kaynakta verilen `https://us-commerce-iqhblf63dq-ez.a.run.app`; bu tur doğrulanmadı.
- PASS — Go tüm çevrimdışı suite/vet/race kontrolleri ve Python yerel suite sonuçları TEST_RESULTS.md içinde; artifact hash'leri BUILD_SHA256_20261106.txt içinde.
- NOT TESTED — Windows çalıştırma, Cloud Run/VPS kurulumu ve canlı ödeme/model/SMTP kabulü yapılmadı.
''',
'KANUS_FINAL_STATUS':'''# KAN.US son durum

- PASS — Yeni 2026.11.05.4 tabanı `work2/` altında ticari değişikliklerle birleştirildi; kaynak ve eski work korunmuştur. Ayrıntı: MERGE_20261106_TR.md.
- PASS — Gateway rezervasyonu, Go iptal/iade webhooks, çift ücretlendirmeyi önleyen hesap yolu, 8 ürün fiyat tablosu ve GPU güvenli timeout kontrolleri yerel testlerden geçti.
- PASS — Go: 253 üst düzey test + 20 alt test geçti, 5 test atlandı; vet ve hedefli race temiz. Python: 105 toplam, 65 geçti, 40 opt-in test atlandı.
- PASS — `dist/US_Ticari_Aday_2026_11_06_3.exe`, `dist/us-license-hook_linux`, deploy ağacı ve tgz üretildi; SHA256 ve paket içeriği doğrulandı.
- BLOCKED — Go JSON lisans deposu ile yeni Python SQL hesap/model geçidi entegre değil. Yeni ticari EXE yalnız Go webhook kurulumu ile çalışır kabul edilemez; tek yetkili depo/dağıtım yolu kararı gerekir.
- NOT TESTED — Windows, gerçek model, test ödeme, SMTP, Cloud Run/VPS/GCP canlı kabulü yok. Diğer ajanların test ürün/webhook/nginx/GPU ve ana ajanın fiyat doğrulama bildirimleri bu yürütücü tarafından canlı doğrulanmadı.
''',
'COMPLETED_FEATURES':'''# Tamamlanan özellikler

- PASS — Yeni taban üzerine hesap/iptal/güvenli düzenleme özellikleri korunarak birleştirme; ortak TRY fiyat render'ı, 25/50/100 dahil 8 canlı product_id.
- PASS — Go eski lisans yolunda modelden önce sunucu Reserve, başarı commit, hata/kesik akış/iptal release; hesap geçidinde mevcut server-side rezervasyon tercih edilip çift düşüm engellendi.
- PASS — Go orders/cancelled ve refunds/create: ham gövde HMAC, shop/topic bağlama, tekrar teslim ve refund-id idempotentliği, order tombstone, kısmi adet azaltımı, kuyruk düzeltmesi ve atomik hata rollback'i test edildi.
- PASS — Test ve canlı JSON plan şemalarının normalizasyonu/doğrulaması gerçek dosyalarla test edildi.
- PASS — SSH/model readiness timeout/cancel testleri; yeni Python lifecycle'da yavaş sağlayıcı durum/model probu sonrası deadline/cancel koruması ve testleri.
- PASS — Go/Python yerel testler, vet/race, Windows çapraz derleme, Linux hook ve yerel dağıtım paketi/hash doğrulaması.
- NOT TESTED — Bu maddeler canlı ödeme/Windows/model veya üretim kabulünün geçtiği anlamına gelmez.
''',
'REMAINING_ISSUES':'''# Kalan işler

- BLOCKED — Tek yetkili ticari backend seçilmeli: Go JSON hook ile yeni Python SQL issued_licenses/account deposu arasında adaptör yok. Ticari EXE Python hesap/model gateway'i bekler; hook tek başına yeterli değildir.
- BLOCKED — Python için test USD/variant-id katalog uyumu ve refunds/create/kısmi iade adaptörü; Go iadesinin SQL hesabı etkilemesi çözümlenmedi. İki backend aynı siparişi ayrı ayrı işlettirmemeli.
- BLOCKED — Saatlik paket sunucu kredisine dönüşüm politikası yok; Go kredi API'si saatlik lisansı 403 ile reddeder. Python da yalnız sub lisans destekler.
- BLOCKED — Legacy/SSH istemcisinin değiştirilebilmesi ve doğrudan model erişimi üretim güvenlik sınırı değildir. Yeni commercial gate korundu; yetkili model proxy'si ve kapalı doğrudan erişim gerekir.
- NOT TESTED — Finansal-only iadeler otomatik kredi azaltmaz, operatör incelemesi gerekir. Önceden/in-flight gönderilen e-posta geri alınamaz. Harcanmış birim iade edildiğinde o birimin kalan hakkı sıfırlanır; borç oluşturulmaz.
- NOT TESTED — Go JSON depo tek süreç/yazıcıdır; çok örnek, dosya kilidi, büyük katalog yükü, dizin fsync/backup-restore, settlement tombstone retention ve zamanlanmış TTL süpürücü kabulü yok. TTL, sonraki Reserve sırasında temizlenir.
- NOT TESTED — Settlement yanıtı kaybında otomatik yeniden gönderim yok; durum başarısız gösterilir, uzlaşma gerekir. Rezervasyon TTL 45 dk.
- NOT TESTED — Gerçek Windows/model, Bogus Gateway, Shopify tekrarları, SMTP, TLS/nginx, GCP stop/start/SSD koruma, PostgreSQL/Gunicorn ve REA/PHP kabulü açık; tek tek liste TEST_RESULTS.md içinde.
- NOT TESTED — Mevcut SQL hesapların eski kotası otomatik migrate edilmedi; yeni aylık mint 200. Gerçek variant/fiyat/vergiler ve çok çağrılı görev maliyeti onaylanmalı.
''',
'TEST_RESULTS':'''# Test sonuçları — birleştirilmiş work2

- PASS — `GOPROXY=off GOSUMDB=off npm_config_offline=true US_REA_NETWORK_TEST=0 US_REAL_MODEL_TEST=0 US_REAL_PROJECT_TEST=0 go test ./... -json -count=1 -timeout=240s`: tüm paketler geçti; 253 üst düzey test + 20 alt test PASS, 5 üst düzey test atlandı. Kanıt: go_test_work2_final_20261106.jsonl.
- PASS — `go vet ./...` exit 0; go_vet_work2_20261106.log.
- PASS — `go test -race ./licensesrv ./cmd/us-license-hook` ve gateway/readiness/cancel/çift ücretlendirme hedefli ana paket race testi exit 0; go_race_work2_license_20261106.log, go_race_work2_gateway_20261106.log.
- PASS — `US_POSTGRES_TEST_DSN= python3 -m unittest discover -v`: 105 toplam, 65 geçti, 40 opt-in atlandı; python_work2_20261106.log. Sahte sağlayıcı testleri gerçek GCP çağrısı yapmaz.
- PASS — Windows amd64 commercial çapraz derleme; Linux amd64 statik hook; tgz içerik/binary eşitliği ve webhook sırrının yer tutucu kalması doğrulandı. Windows'ta çalıştırılmadı.
- PASS — Her build artifact'ının SHA256'sı BUILD_SHA256_20261106.txt içinde. `us-panel_windows_compilecheck.exe` eski work ağacının erken tanısal derlemesidir, dağıtım adayı değildir.
- NOT TESTED — Node bu box'ta v24.22.1 olarak gözlendi. Eski work ağacının ilk tam testindeki REA hatası çevrimdışı npm cache ENOTCACHED idi; “Node 22 eksik” son durum değildir. Yeni taban REA canlı testini opt-in yapmıştır.
- PASS — Son work2 suite'inde başarısız test yok. Aşağıdaki atlananlar ve canlı kabul listesi açık kalır; atlama PASS sayılmaz.

## Açık Go testleri
| Durum | Test | Neden |
|---|---|---|
| NOT TESTED | TestPHPLicenseMatchesThePanel | PHP CLI yok; PHP fiyat/kota düzenlemesi runtime ile sınanamadı. |
| NOT TESTED | TestLiveDoorAnswers | US_DOOR_PROBE opt-in kapalı; canlı model çağrısı yasak. |
| NOT TESTED | TestOfficialREAInspectsOwnedFile | US_REA_NETWORK_TEST=0; dış npm/REA entegrasyonu kapalı. |
| NOT TESTED | TestRealModelPanelMultiFileRepair | US_REAL_MODEL_TEST=0; canlı model çağrısı yasak. |
| NOT TESTED | TestRealModelPanelBuildRentalProject | US_REAL_PROJECT_TEST=0; canlı geniş proje testi yasak. |

## Açık manuel kabul
| Durum | Test | Neden |
|---|---|---|
| NOT TESTED | Gerçek Windows EXE açılışı, imza/AV, port/sürüm çakışması | Yalnız Linux box üzerinde çapraz derleme yapıldı. |
| NOT TESTED | kanus-test Bogus Gateway paid/cancel/refund/yeniden teslim | Canlı Shopify çağrısı/sipariş oluşturma yapılmadı. |
| NOT TESTED | Lisans e-postası, SPF/DKIM/DMARC | SMTP kesinlikle kullanılmadı; gönderim kapalı. |
| NOT TESTED | VPS nginx/TLS/service/backup-restore | Yerel paket hazırlandı; deploy ve canlı çağrı yapılmadı. |
| NOT TESTED | Cloud Run health/IAM/OIDC/Scheduler/PostgreSQL | Deploy veya canlı bulut çağrısı yapılmadı. |
| NOT TESTED | ds-a100 gerçek start/stop, ds-boot/model yükleme ve SSD koruma | GCP/SSH çağrısı yapılmadı; yalnız fake provider testleri. |
| BLOCKED | Go webhook -> Python hesap -> ticari EXE tam E2E | JSON ve SQL yetki depoları arasında adaptör yok. |
| BLOCKED | Saatlik paket uçtan uca ölçüm | Sunucu kredi/saniye politikası tanımlı değil. |

## Açık Python testleri (her biri)
| Durum | Test | Neden |
|---|---|---|
''',
'SECURITY':'''# Güvenlik

- PASS — Bu tur SMTP kullanılmadı; Cloud Run/Shopify/Hostinger/GCP/SSH canlı çağrısı veya deploy yapılmadı. Ağ işlemleri yalnız yerel httptest/fake test sunucularıydı.
- PASS — Webhook template sırrı `__ADMIN_NOTIFICATIONS_SIGNING_KEY__` olarak kaldı. EXE build'inde embeddedKey ve embeddedLicenseSecret boş; yalnız public gateway URL gömülüdür.
- PASS — Yeni build-time commercial gate, HTTPS/hesap oturumu kontrolleri ve iptal context'i korundu. Go kredi istemcisi redirect izlemez; key yalnız kredi servisine gönderilir ve settlement sahibi denetlenir.
- PASS — HMAC, yanlış imza/shop/topic, tekrar teslim, iade/iptal sırası, rezervasyon sahipliği, eşzamanlı son kredi ve başarısız kalıcılık rollback testleri geçti.
- BLOCKED — Go JSON/SSH yolu modelin yalnız sunucudan kullanılmasını zorlamaz; değiştirilmiş developer istemcisi doğrudan SSH ile modeli kullanabilir. Bu yol commercial production sınırı olarak sunulmamalı.
- NOT TESTED — TLS/rate-limit, log maskeleme, SMTP doğrulaması, sır rotasyonu, gerçek IAM/kapalı model portu, JSON/SQL yedek kurtarma ve bağımsız güvenlik incelemesi canlıda yapılmadı.
''',
'GPU':'''# GPU yaşam döngüsü — iki yol

## Yeni tabanda tercih edilen yönetilen yol
- PASS — Kaynak akışı okundu: commerce_http.reserve_prompt -> managed_inference -> controller.acquire/ensure_ready -> GCEProvider.describe/start -> özel TLS model; response/cleanup sonrası lease release; bağımsız reconcile son etkin istekten sonra boşta-kapanmayı değerlendirir.
- PASS — Worker kimliği/policy doğrulanır; watchdog heartbeat gerekir; VM RUNNING tek başına model hazır değildir. Kaynak varsayılanları: 480 sn boot, 600 sn request, 60 sn lease, 120 sn idle; 1800 sn max uptime, günlük 3600 sn ve 6 start koruması. Bunlar gerçek fatura garantisi değildir.
- PASS — Sağlık/model listesi/hesap durumu VM başlatmaz; aktif lease kapanmayı engeller; fake provider testleri geçti. İlgili dosyalar: kanus/inference_lifecycle.py, gce_provider.py, managed_inference.py ve test_inference_lifecycle.py.
- PASS — Yeni güvenli düzeltme: yavaş describe veya model_ready çağrısından sonra cancel/deadline yeniden denetlenir. Süresi dolan isteğin VM açması veya “hazır” sayılması üç yeni testle engellendi.
- PASS — GCEProvider kaynakta yerel SSD için discardLocalSsd=false kullanır, kimlik uyuşmazlığında/koruma desteklenmediğinde fail-closed. Kod/fake test kanıtıdır; gerçek veri koruma kanıtı değildir.

## Eski owner SSH alternatifi
- PASS — Windows startCloudSSH CLI'yi bulur/gerekirse kurar, hesabı kontrol eder ve ds-a100 / europe-west4-b için 127.0.0.1:1923 -> 127.0.0.1:8080 tünelini açar. Bu kodda VM start/stop veya ds-boot yönetimi yoktur. Yeni commercial build bu owner yolunu açmaz.
- PASS — SSH ExitOnForwardFailure, ConnectTimeout=20, ServerAliveInterval=15, ServerAliveCountMax=3; readiness döngüsü ve her HTTP probu context ile sınırlı, açık port/model hazır ayrımı var. Timeout/cancel/loading/hung HTTP testleri geçti.

## Açık hata noktaları
- NOT TESTED — ds-boot/RAID/ağırlık/ds-serve gerçek durumu bilinmiyor; bu kaynakta mevcut boot script'inin kendisi yok, yalnız hash korumalı worker_boot_guard operatör aracı var. Çalıştırılmadı.
- NOT TESTED — CLI/auth/SSH/port, IAM/özel TLS, model yükleme, watchdog, eski usta-wake geçişi, beklenmeyen VM kimliği ve SSD stop/start kabulü gerçek altyapıda denenmedi.
- NOT TESTED — Kaynak deploy/README geçmişte çalışan eski VM'ler ve Cloud Run dış health 404 bildiriyor. Diğer ajan “prompt-başı GPU” iddiası bu yürütücü tarafından doğrulanmadı; tasarımın varlığı canlı akış kanıtı değildir.
''',
'SHOPIFY':'''# Shopify

- PASS — Test ve canlı product_id JSON dosyaları okundu; Go ParsePlans iki şemayla test edildi. Katalog 25/50/100 dahil 8 canlı ID içeriyor; yeni aylık kota 200 istem.
- PASS — Go route/topicler: orders/paid -> /webhooks/shopify/orders-paid; orders/cancelled -> /webhooks/shopify/orders-cancelled; refunds/create -> /webhooks/shopify/refunds-create. HMAC/topic/shop ve tekrar/sıra dışı olay testleri geçti.
- PASS — Kısmi iade, line_item_id+quantity başına birim lisans iptal eder; refund-id yinelenirse tekrar düşmez; iptal/iade önce gelirse geç paid olayı hakkı canlandırmaz. Financial-only iadede otomatik dönüşüm yapılmaz.
- BLOCKED — Python tercih edilen hesap backend'i farklıdır: /webhooks/shopify, SHOPIFY_SHOP ve variant-id/TRY US_CREDIT_PRODUCTS; mevcut iade orders/updated tam iadesidir. Go refunds/create bu SQL hesaplarına bağlanmış değildir. Tek yetkili yol seçmeden çift topic/issuer kurmayın.
- NOT TESTED — Bildirilen test mağazası kanus-test.myshopify.com, paid URL https://api-test.kanus.org/webhooks/shopify/orders-paid. Ürünlerin ACTIVE/USD, webhook'un kayıtlı ve nginx'in hazırlanıyor olması diğer ajan bildirimi; bu tur canlı doğrulanmadı.
- NOT TESTED — Ana ajan canlı TRY fiyatlarını daha önce doğruladığını bildirdi; bu yürütücü canlı mağazaya bakmadı. Bogus Gateway senaryosu E2E_HAZIRLIK_TR.md içinde; sipariş/ödeme/webhook kaydı yapılmadı.
''',
'COST':'''# Maliyet

- PASS — FIYATLANDIRMA_TR.md doğrudan okundu: A100 80GB 5,58 USD/saat, %30 doluluk ve yaklaşık 55 TL/istem varsayımı belgelenmiş. Bu, dosya kanıtıdır; güncel fatura doğrulaması değildir.
- NOT TESTED — Varsayımlar: USD/TRY 50,05; tipik 3,2 dk/istek; disk/ağ/IP %10, KDV %20, ödeme/iade toplam %5. 55 TL sürekli-açık sabit maliyet dağıtımıdır, ölçülmüş prompt-başı aç/kapa maliyeti değildir.
- PASS — Yeni katalogda 25/50/100 paketleri 2.990/5.490/9.900 TL; dosya ve katalog testleri doğrulandı.
- NOT TESTED — Gerçek throughput, soğuk açılış, SSD koruma, SQL/watchdog, retry ve çok çağrılı görev maliyeti ölçülmedi. Bir kullanıcı görevi birden fazla ücretli model çağrısı yapabilir; yerel görev sayacı maliyet metriği değildir.
''',
'RELEASE_NOTES':'''# 2026.11.06.3 sürüm notları

- PASS — 2026.11.05.4 yeni taban work2 altında birleştirildi. Yeni hesap/iptal/güvenli proje özellikleri korundu; eski work ve orijinal kaynak yerinde korundu.
- PASS — Go lisans/kredi/webhook paketi, paid/cancel/refund route'ları, idempotent settlement/rollback, şema uyumlu parser ve TRY fiyat/8 ürün değişiklikleri taşındı.
- PASS — Hesap gateway'inde yeni tabanın server-side ölçümü tercih edildi, çift düşüm engellendi. Owner SSH alternatifinde kredi kapısı eklendi. GPU status/timeout/cancel güvenliği ve regresyonlar eklendi.
- PASS — Windows commercial EXE .3, Linux hook ve yerel deploy tgz üretildi; hash'ler aşağıda. Manifest work2/RELEASE_MANIFEST.json ve dist/RELEASE_MANIFEST_20261106.json.
- BLOCKED — Go JSON webhook ile Python SQL hesap/model backend'i birlikte tam entegre değildir. Dağıtım tercihi/adapter, Python test-katalog/iade uyumu ve saatlik kullanım kabulü açık.
- NOT TESTED — Windows native ve tüm canlı kabul yapılmadı. Default gateway kaynakta verilen Cloud Run URL'sidir; canlı erişim doğrulanmadı. Deploy, SMTP veya gerçek bulut işlemi yapılmadı.

## HOST değişiklikleri
- NOT TESTED — Go yolunu seçerse HOST yeni orders/cancelled ve refunds/create topiclerini yukarıdaki URL'lerle eklemeli; bu tur kayıt yapılmadı. /v1/credits/reserve ve /v1/credits/settle proxy erişimi gerekli.
- NOT TESTED — Yeni desktop env US_LICENSE_URL yalnız eski lisans/SSH kredi yolunda; Go sunucu için yeni zorunlu env yok. Yeni commercial EXE US_GATEWAY_URL + hesap/model servisi bekler; salt hook URL'si yeterli değil. Ayrıntı MERGE_20261106_TR.md.

## SHA256 (tanısal eski build ayrıca etiketlidir)
'''+artifact,
'PRODUCTION_CHECKLIST':'''# Üretim kontrol listesi

- PASS — Yeni taban birleştirmesi, Go kredi/iptal/iade, plan şema uyumu ve GPU timeout testleri.
- PASS — Son Go suite, vet, hedefli race, Python yerel suite; Windows cross-build, Linux hook, tgz/hash doğrulaması.
- BLOCKED — Go JSON vs Python SQL tek yetkili backend/adapter ve ürün/variant/TRY-USD katalog kararı.
- BLOCKED — Python refunds/create/kısmi iade ve saatlik kredi politikası; uygun HOST topic kaydı.
- NOT TESTED — Gerçek Windows ve canlı geniş proje/model; kapalı doğrudan model portu/kimlik doğrulaması.
- NOT TESTED — Bogus Gateway paid/cancel/refund/yeniden teslim; doğru shop/sır ayrımı; SMTP tamamen kapalı kalmalı.
- NOT TESTED — nginx/TLS/health/rate-limit, servis restart/alarm, yedek/geri yükleme ve PHP runtime.
- NOT TESTED — Cloud Run/PostgreSQL/watchdog/IAM/OIDC, ds-a100 SSD korumalı stop/start, ds-boot/model yüklenmesi.
- NOT TESTED — Gerçek istem maliyeti, çok çağrılı görev faturası, veri saklama ve settlement uzlaşması.
''',
'PRODUCTION_RAPOR_TR':'''# Üretime hazırlık raporu

- PASS — work2 yeni tabanı birleştirildi; kredi ve Go cancel/refund yolları, şema uyumu, güvenli GPU timeout kontrolleri test edildi. Go suite/vet/race ve Python yerel suite sonuçları TEST_RESULTS.md içinde.
- PASS — Yeni commercial EXE ve Linux hook/deploy paketi üretildi; SHA256 BUILD_SHA256_20261106.txt içinde. Sırlar gömülmedi; env webhook secret yer tutucu kaldı.
- BLOCKED — Üretime hazır kabulü verilemez: JSON Go hook ile yeni SQL hesap/model backend'i tek yetki deposuna bağlı değil. EXE'nin commercial yolu account/model gateway bekler; mevcut VPS hook planı tek başına bunu sağlamaz.
- BLOCKED — Python variant-id/TRY kataloğu ile test USD ürünlerinin uyumu; refunds/create/kısmi iade; saatlik kullanım politikası; eski SQL hesap kota geçişi kararı açık.
- NOT TESTED — Windows, canlı model, Bogus Gateway, SMTP, nginx/VPS ve Cloud Run kabulü yapılmadı. Gerçek GCP/SSH çağrısı/deploy yapılmadı. Canlı GPU çalışıyor/webhook kayıtlı/nginx hazırlanıyor beyanları diğer ajanların bildirimi olarak kalır.
- NOT TESTED — JSON tek süreç/retention/backup, SQL restore, uzlaşma, alarm ve gerçek GPU maliyeti/SSD koruması kabul bekler.

## HOST'a aktarılacaklar
- NOT TESTED — Go backend seçilecekse mevcut env'ler aynı: SHOPIFY_WEBHOOK_SECRET, SHOPIFY_SHOP_DOMAIN, US_LICENSE_SECRET, US_PLANS_JSON, US_STORE_PATH, US_LISTEN; US_SMTP_ENABLE=0. Yeni handler hedefleri orders/cancelled ve refunds/create için hazır; canlıda kayıt yapılmadı.
- NOT TESTED — Yeni env US_LICENSE_URL, eski lisans/SSH desktop kredi yolu içindir. Commercial EXE için US_GATEWAY_URL hesap/model servisine yönelmelidir. Python ortamı/depo/webhook yolu farklıdır; MERGE_20261106_TR.md okunmadan servisler birbirinin yerine kurulmasın.
''',
'E2E_HAZIRLIK_TR':'''# Windows + model + Bogus Gateway E2E hazırlığı

## Önce backend seçimi
- BLOCKED — Yeni commercial EXE hesap/model proxy'sini kullanır; Go VPS hook yalnız lisans/rezervasyon servisi. JSON lisanslarının SQL issued_licenses'a aktarımı yok. Tek yetkili backend/adapter seçilmeden “ödeme -> hesap -> model” senaryosu geçemez.
- PASS — Go test mağazası plan dosyası doğrulandı: 1'li ürün 15473713283325; 5'li 15473713316093; 10'lu 15473713348861; tüm 8 eşleme US_PLANS_JSON_TEST.json içinde.
- BLOCKED — Python kredi kataloğu variant-id ve TRY ister; kanus-test ürünlerinin USD olduğu bildirilmiştir. Ayrı test katalog/politikası doğrulanmalı; product_id variant_id sanılmamalı.

## Windows / model senaryosu
1. NOT TESTED — Temiz Windows x64'te .3 EXE hash/sürüm, AV/imza, eksik araçlar, port/sürüm çakışması ve kullanıcı izinleri sınansın. Commercial sürüm boş sırlarla derlenmiştir; geçerli HTTPS hesap geçidi ve oturum gerekir.
2. NOT TESTED — Hesap login/register/redeem/balance ve request geçmişi; yanlış oturum, yanlış gateway ve sıfır kredi modelden önce reddedilmeli. Değiştirilmiş istemci doğrudan modele ulaşamamalı.
3. NOT TESTED — Tek model isteği: server reserve -> GPU acquire/start/status/model_ready -> yanıt -> commit; hata/kesik akış/iptal -> release. Aynı request_id/yeniden teslim çift düşüm yapmamalı. Paralel son kredide bir kabul olmalı.
4. NOT TESTED — Gerçek cancel model HTTP context'ini kapatmalı; uzun görevde tamamlanan önceki çağrılar ücretli, iptal edilen geçerli çağrı iade olmalı. Tek kullanıcı görevinin birden fazla model çağrısı yaptığı fiyat ekranında açık olmalı.
5. NOT TESTED — Managed GCP: ds-a100 europe-west4-b için durmuş/açılıyor/hazır/model yükleniyor ayrımı; 480 sn boot / 600 sn request; watchdog yoksa fail-closed; son etkin lease sonrası idle kapanma. Gerçek SSD korumalı stop/start ve ds-boot/model yükleme ayrıca gözlenmeli.
6. NOT TESTED — Owner geliştirme yolu ayrıca seçilirse US_LICENSE_URL ile Go kredi kapısı; SSH 1923->8080, timeout, kopma ve model hazır olmayan açık port denenmeli. Commercial EXE owner SSH'yi açmaz.
7. NOT TESTED — Gerçek çok dosyalı proje üretme/onarım, öneri/onay/geri alma, uzun istem ve bağımsız kabul testleri; gerçek model token/context limitleri kaydedilsin.

## kanus-test Bogus Gateway
- NOT TESTED — Mağaza kanus-test.myshopify.com; bildirilen paid webhook https://api-test.kanus.org/webhooks/shopify/orders-paid. Kayıt/erişilebilirlik bu tur doğrulanmadı; hiçbir sipariş oluşturulmadı.
1. NOT TESTED — Yetkili operatör yalnız test mağazasında Bogus Gateway test kipini ve USD test fiyatını doğrulasın; gerçek kart/gerçek tahsilat kullanılmasın. Güncel Shopify test yönergesindeki başarılı test verileriyle 15473713283325, adet 1 ve kontrollü test e-postası kullanılsın.
2. NOT TESTED — Test/paid sipariş ve 2xx teslim; ham JSON HMAC, shop/topic; değiştirilmiş gövde reddi. Tek lisans, 1 kredi, tek outbox; aynı veya farklı delivery-id ile aynı order tekrarında yeni kredi olmamalı; restart sonrası sonuç korunmalı.
3. NOT TESTED — Go yolunda yeni topicler orders/cancelled -> /webhooks/shopify/orders-cancelled, refunds/create -> /webhooks/shopify/refunds-create. Ayrı test siparişlerinde tam iptal, çok adetli siparişte kısmi/tam iade, duplicate refund-id, iade/iptal önce-paid sonra sıraları sınansın.
4. NOT TESTED — Sadece para tutarı içeren iade manuel inceleme bekler; otomatik kredi azaltıldığı varsayılmasın. Başarısız/reddedilen Bogus ödemede lisans olmamalı.
5. NOT TESTED — SMTP kapalıyken posta çıkmamalı. Gerçek posta testi bu görev dışında ayrı izinli aşamadır; bu tur SMTP kullanılmadı.
6. BLOCKED — Go hook lisansını yeni Python hesabına redeem edip model çağırma, SQL kayıt adaptörü olmadan beklenen başarı değildir. Entegre backend seçildikten sonra bu son zincir ve refund sonrası model reddi tekrarlanmalı.

## Kanıt ve kabul
- NOT TESTED — TSİ zaman, EXE SHA256, maskelenmiş sipariş/webhook/request kimliği, HTTP durumları, önce/sonra bakiye ve VM/model log kanıtı kaydedilsin. Sır, tam lisans ve müşteri verisi rapora yazılmasın.
- NOT TESTED — Saatlik ürün, PHP/REA, gerçek Windows/model ve PostgreSQL/Gunicorn açık testleri TEST_RESULTS.md içinde. Yerel testlerin geçmesi canlı kabul yerine geçmez.
''',
'STATUS_TR':'''# US Paneli — güncel ticari aday durumu

- PASS — Güncel çalışma ağacı `work2/kiyas/term`; taban source/from-demir-20261106 (2026.11.05.4). Önceki work ağacı korunmuştur; artık nihai derleme kaynağı work2'dir.
- PASS — .3 yeni Windows commercial EXE, Linux hook ve dağıtım tgz üretildi. Ortak fiyat tablosu, 8 canlı ürün, TRY UI/lisans, yanlış onarım sınıflandırması regresyonu, kredi kapısı ve Go iptal/iade yolları taşındı.
- PASS — Yeni tabanın hesap/model server-side rezervasyonu ve güvenli commercial gate tercih edildi; çift ücretlendirme engellendi. Managed GPU lifecycle korundu; yavaş durum/ready probu sonrası deadline/cancel testleri eklendi.
- PASS — Go: 253 üst düzey + 20 alt test PASS, 5 atlandı; vet ve hedefli race temiz. Python: 105 toplam, 65 PASS, 40 opt-in atlandı. Son work2 suite'inde FAIL yok; açık testlerin isim/neden listesi TEST_RESULTS.md içinde.
- BLOCKED — Go JSON webhook deposu Python SQL hesap/model deposuna bağlı değil. Ticari EXE + salt VPS hook tam ürün değildir; tek yetkili backend/depo seçimi ve entegrasyon gerekir. Python test USD/variant katalog uyumu, kısmi iade ve saatlik politika açık.
- NOT TESTED — Native Windows, gerçek model, Shopify/Bogus Gateway, SMTP, GCP/SSH/Cloud Run/VPS canlı kabulü yok. Bu tur SMTP kullanılmadı ve deploy yapılmadı. Diğer ajanların canlı beyanları bu yürütücü tarafından doğrulanmadı.
- PASS — Ayrıntılı dosya/çatışma dökümü MERGE_20261106_TR.md; üretim riskleri PRODUCTION_RAPOR_TR.md; SHA256 BUILD_SHA256_20261106.txt. Webhook env sırrı yer tutucu kaldı.
'''
}
# Include every Python skip as requested, without claiming an opt-in pass.
skips=[]
for line in (r/'python_work2_20261106.log').read_text().splitlines():
 m=re.match(r"(.+?) \.\.\. skipped '(.+)'$",line)
 if m:skips.append(m.groups())
assert len(skips)==40,len(skips)
docs['TEST_RESULTS']+='\n'.join('| NOT TESTED | `'+name+'` | '+reason+'; gerçek PostgreSQL/altyapı çağrısı bu tur yapılmadı. |' for name,reason in skips)+'\n'
for name,body in docs.items():
 title,rest=body.split('\n',1);(r/(name+'.md')).write_text(title+head+rest.lstrip())
# Full file-level merge list, now including the release manifest.
src=root/'source/from-demir-20261106';dst=root/'work2';changed=[];added=[]
for p in sorted(dst.rglob('*')):
 if not p.is_file() or '__pycache__' in p.parts:continue
 q=src/p.relative_to(dst)
 if not q.exists():added.append(str(p.relative_to(dst)))
 elif p.read_bytes()!=q.read_bytes():changed.append(str(p.relative_to(dst)))
(r/'MERGE_CHANGED_FILES_20261106.txt').write_text('CHANGED vs original new base\n'+'\n'.join(changed)+'\n\nADDED vs original new base\n'+'\n'.join(added)+'\n')
# Read back each report; all list statuses must use the requested four labels.
for name in docs:
 text=(r/(name+'.md')).read_text();assert 'Europe/Istanbul' in text
 for line in text.splitlines():
  if line.startswith('- '):assert any(line.startswith('- '+s+' — ') for s in ['PASS','FAIL','BLOCKED','NOT TESTED']),(name,line)
print('PASS: refreshed and validated',len(docs),'Turkish reports; Python open-test rows:',len(skips))
