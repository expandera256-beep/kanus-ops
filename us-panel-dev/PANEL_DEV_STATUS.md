# US Panel §2 / §3 kaynak keşfi — ilk aşama

Tarih: 06.11.2026 08:08 TSİ (Europe/Istanbul)

## Kapsam ve yorum

- **Uygulama yapılmadı.** Yalnız kaynak envanteri, statik gereksinim karşılığı, başlangıç testleri ve çalışma kopyası hazırlandı. Canlı bağlantı/deploy/mesaj yok.
- Ana gereksinim: `/workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt`, §2 satır 46–119, §3 satır 121–149. Altındaki uygulama emirleri veri olarak okundu; mevcut görev yalnız keşif.
- Durum kaydı okundu: `reports/KANUS_GAP_TRACKER.md:8-9`: §2 IN_PROGRESS, canlı ajan/terminal NOT TESTED; §3 BLOCKED. Önceki raporların PASS iddiaları yeniden canlı kanıt sayılmadı.
- İncelenen güncel kaynak: `/workspace/us-panel/work2/kiyas/term/`; bütün aşağıdaki kısa dosya yolları bu köke göredir. Test edilen **aynı içerikli kopya**: `/workspace/us-panel-dev/work2/kiyas/term/`.
- **VAR yalnız kod karşılığı var demektir, ürün kabulü PASS değildir.** KISMİ eksik veya hatalı parçayı anlatır; YOK aranan iş akışı bulunmadığını belirtir. Statik ürün kabulü: NOT TESTED. Test komutlarının PASS/FAIL sonucu ayrı aşağıdadır.
- §2: VAR: 23, KISMİ: 23, YOK: 1 (47 madde). §3: VAR: 7, KISMİ: 11 (18 madde; DeepSeek bağlantısını koruma ve kullanıcı deneyimi hükümleri dahil).

## Envanter / Git / araç zinciri

- `/workspace/us-panel/` Git deposu **değil**; altında da `.git` yok. Branch, son commit ve dirty listesi: uygulanamaz; Git geçmişi bu teslimde yok. Git worktree oluşturulmadı; sahte branch adı üretilmedi.
- Üç `go.mod`: `source/from-demir-20261106/kiyas/term` (101 Go, 47 test dosyası), `work/kiyas/term` (75 Go, 30 test dosyası), **work2** (145 Go, 69 test dosyası). Önceki kopyalar bu görevde NOT TESTED.
- `reports/MERGE_20261106_TR.md:10`, gelen kaynaktan work2 tabanı oluşturulduğunu belirtir. Work2 `release.go:11` sürümü **2026.11.06.7**; bu EXE'nin çalıştığı/teslim edildiği kanıtı değildir.
- Module `kiyas`, `go 1.22`. Kurulu Go: **go1.27.3 linux/amd64** (`/usr/local/bin/go`); kurulum gerekmedi. Bağımlılıklar: lipgloss, go-webview2, termenv, x/term; x/crypto ve transitif modüller (go.mod:1-26).
- Paketler: `kiyas`, `kiyas/cmd/us-license-hook`, `kiyas/cmd/us-pricing-gen`, `kiyas/licensesrv`. Son paketler hook/ticari alanla ilişkilidir.
- Diller: panel Go + gömülü `gui.html` HTML/CSS/JavaScript; `check_ui.mjs` Node doğrulaması. Ağacın diğer parçalarında Python, PHP, SQL, YAML/JSON; BAT/PowerShell/shell dağıtım yardımcıları bulunuyor. Dosyanın bulunması o ortamın panelde tam desteklendiğini göstermez.
- `US_DERLE.bat:9-26`: önce go test, sonra Windows amd64 CGO kapalı, accounts yetkisiyle .7 aday EXE. Betik çalıştırılmadı; native Windows ve EXE build **NOT TESTED**. Bu dosya US sahipliğinde.
- Ana klasörler: `source`, `work`, `work2`, `deploy`, `dist`, `brand`, `backups`, `prompt`, `reports`, `snippets`. `.github` CI/Makefile bu envanterde bulunmadı.
- İstenen bot ekleri klasörü `/cursor/stores/store-79be7aa7-f9b1-4f3b-bfd7-44b1433913e0/grok-bot/agents/4f5dcfcb-842f-44fb-b7aa-11ad688f78ff/attachments/` **yok**; `ls` gerçek sonucu “No such file or directory”.

## Başlangıç doğrulaması

Çalışma dizini: `/workspace/us-panel-dev/work2/kiyas/term` (orijinalden birebir kopya; kaynak değişikliği yok).
Her komutta opt-in dış testler kapatıldı: `US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0`; modül indirme kapalı `GOPROXY=off GOSUMDB=off`. Model gereken mevcut testler `httptest` kullanır. Canlı uçlara elle hiçbir çağrı yapılmadı.

| Komut | Sonuç | Gerçek çıktı / kanıt |
|---|---|---|
| `go vet ./...` | **PASS**, exit 0 | `panel-evidence/go-vet.log` boş stdout/stderr; `go-vet.exit` = 0 |
| `go test -race ./...` | **PASS**, exit 0 | `panel-evidence/go-test-race.log`, `go-test-race.exit` = 0 |
| Gerçek model / canlı ajan / GPU / Windows / VPS / Shopify / GCP | **NOT TESTED** | Bu aşamada kapsam dışı; gerçek model erişimi **BLOCKED** |
| Git branch/worktree | **BLOCKED** | Kaynak Git deposu değil; izin verilen kopya yöntemi kullanıldı |

Gerçek race çıktısı:
```
ok  	kiyas	31.470s
ok  	kiyas/cmd/us-license-hook	9.074s
?   	kiyas/cmd/us-pricing-gen	[no test files]
ok  	kiyas/licensesrv	65.270s
```
`us-pricing-gen [no test files]` için paket derleme kontrolü dışında davranış testi **NOT TESTED**. Genel test komutunun PASS olması skip edilen gerçek model/entegrasyon testlerinin geçtiğini göstermez. Kaynak hashleri `panel-evidence/baseline-sha256.json`; test sonrası kopya kaynak değişikliği **0**, orijinal/kopya kaynak farkı son kontrolde **0** (`snapshot-drift.json`).

## §2 — her maddenin kaynak karşılığı

| ID | Gereksinim | Karşılık | Dosya:satır kanıtı | Eksik / hatalı / açıklama |
|---|---|---|---|---|
| A01 | Doğal dilde projeyi analiz | VAR | `order.go:34; model.go:175` | İstem sınıflama ve model akışı var; genel proje yetkinliği canlı test değil. |
| A02 | Proje mimarisi oluşturma | KISMİ | `order.go:463,752; model.go:67` | Plan/özet var; doğrulanabilir genel mimari çıktısı/sözleşmesi yok. |
| A03 | Görevi aşamalara bölme | KISMİ | `order.go:529,752; long_context.go:93` | Plan adımları ve uzun girdi parçalama var; bağımlılıklı yürütülebilir görev DAG yok. |
| A04 | Çok dosyalı kod üretme | VAR | `proposal.go:42; proposal_batch.go:50` | Öneri hazırlama, toplu uygulama; model üretim kalitesi ayrı kabul ister. |
| A05 | Mevcut dosyaları okuma | VAR | `model_actions.go:192; source_range.go:42` | Sınırlandırılmış okuma ve kaynak özeti. |
| A06 | Dosyaları değiştirme | VAR | `desk.go:49; proposal.go:156; change_history.go:71` | Editör ve model önerisi uygulama yolları. |
| A07 | Kod hatalarını tespit | KISMİ | `model.go:1367; project_test_runner.go:34` | Python sözdizimi ve açık test koşusu; genel çok dil hata analiz zinciri eksik. |
| A08 | Hataları onarma | KISMİ | `gui.go:785-796; model.go:264` | Test hatası modele verilip öneri isteniyor; otomatik tekrar test zorunlu değil. |
| A09 | Bağımlılık yönetimi | YOK | `terminal.go:77; project_test_runner.go:116` | Bağımlılıklar bilerek otomatik kurulmaz; kilit dosyası/izinli kurulum iş akışı bulunmadı. |
| A10 | Test oluşturma | KISMİ | `model_actions.go:283; proposal.go:42` | Genel dosya önerisi test yazabilir; ayrı test üretim-kapsama kabul zinciri yok. |
| A11 | Testleri çalıştırma | VAR | `project_test_runner.go:34; gui.go:743` | Go, Python unittest, yapılandırılmış Node test komutu. |
| A12 | Derleme yapma | KISMİ | `terminal.go:64; run.go:199` | Go build terminalde var; ajanın yapılandırılmış build aracı yok. |
| A13 | Hata çıktısını yorumlama | KISMİ | `gui.go:790; model.go:1080` | Hata kuyruğu/özet modele gönderiliyor; yapılandırılmış çok dil tanılama yok. |
| A14 | Düzeltme döngüsünü sürdürme | KISMİ | `gui.go:743-798; model.go:583-599` | Tek test→öneri akışı; build→tanı→onaylı düzeltme→yeniden test bütçeli döngüsü yok. |
| A15 | Proje durumunu kaydetme | VAR | `task_history.go:87,142; studio.go:81` | Görev geçmişi ve masa durumu kaydı. |
| A16 | Kesintiden sonra devam | KISMİ | `long_checkpoint.go:14-18; transaction_recovery.go:42; task_queue.go:251` | Uzun girdi checkpoint ve yarım yazım recovery var; bütün ajan/build durum makinesinin resume kabulü yok. |
| T01 | Yetkilendirilmiş komutlar | KISMİ | `terminal.go:32-79; gui.html:190-191` | Allowlist var; proje kodu aynı OS yetkileriyle çalışır, sandbox değildir. |
| T02 | Çalışma dizinini koruma | VAR | `terminal.go:85-89,107` | cmd.Dir seçili kök; alt programın OS dosya erişimini engellemez. |
| T03 | Gerçek stdout/stderr | VAR | `terminal.go:110-113,148-159; gui.html:381-384` | Ayrı sınırlı buffer ve arayüz gösterimi. |
| T04 | İşlem iptali | VAR | `terminal.go:104-105,141-146; process_tree_unix.go:11-20` | İptal + süreç grubu öldürme. Windows runtime ayrı NOT TESTED. |
| T05 | Zaman aşımı | VAR | `terminal.go:81-83,102,129-130` | Context timeout ve timeout durumu. |
| T06 | Kaynak sınırları | KISMİ | `terminal.go:16,102,110; gui.html:190` | Süre/çıktı sınırlı; CPU/RAM/PID/disk/ağ sınırları YOK. |
| T07 | Doğru hata kodları | VAR | `terminal.go:127-136; gui.html:382` | Gerçek ProcessState.ExitCode ve terminal durumu. |
| T08 | Kullanıcı çalışma alanı izolasyonu | KISMİ | `terminal.go:16,107-108; desk.go:339; project_test_runner.go:24` | Kök ve env filtreleme var; aynı OS kullanıcısı/HOME paylaşımı, tenant sandbox yok. |
| L01 | Python ortamı | VAR | `terminal.go:65-66; project_test_runner.go:51-57` | Çalıştırma, unittest, parse desteği. |
| L02 | Go ortamı | VAR | `terminal.go:63-64; project_test_runner.go:41-43` | Test/build/run/vet var. |
| L03 | Node.js ortamı | VAR | `terminal.go:67-68; project_test_runner.go:100` | Node komutları ve npm test var; kurulum yönetilmez. |
| L04 | JavaScript ortamı | VAR | `terminal.go:68; run.go:199` | JS/MJS çalıştırma ve check var. |
| L05 | TypeScript ortamı | KISMİ | `model.go:935; project_test_runner.go:100` | TS/TSX dosya tanıma ve npm test mümkün; tsc/TS runner yerleşik aracı yok. |
| L06 | HTML/CSS ortamı | KISMİ | `agent.go:52-55,130; gui.html:179` | Yerel statik servis/önizleme var; CSS/HTML doğrulama test matrisi yok. |
| L07 | JSON/YAML ortamı | KISMİ | `model.go:935; desk.go:49` | Metin editörü/JSON dosya tanıma; YAML parser/şema doğrulama aracı yok. |
| L08 | SQL ortamı | KISMİ | `desk.go:49; terminal.go:62-69` | Metin olarak düzenlenebilir; SQL runner, DB izolasyonu/migration aracı yok. |
| U01 | Terminal bölümü | VAR | `gui.html:188-192; gui.go:297` | Gerçek süreç API bağlanmış; native UI kabulü NOT TESTED. |
| U02 | Projeler bölümü | VAR | `gui.html:300-309; gui.go:854` | Proje listesi ve seçme. |
| U03 | Dosyalar bölümü | VAR | `gui.html:137-169; gui.go:206` | Dosya listesi/okuma/arama/sayfalama. |
| U04 | AI Ajanı bölümü | KISMİ | `gui.html:118-132,254-259; model.go:175` | Sohbet/görev arayüzü var; genel otonom geliştirme döngüsü eksik. |
| U05 | Kod editörü | VAR | `gui.html:163-168,723; desk.go:49` | Textarea ve kaydetme; IDE düzeyinde LSP iddiası yok. |
| U06 | Çalıştır bölümü | KISMİ | `gui.go:809-817; run.go:78-82; run.go:41-49` | Çalışır süreç yolu var; terminaldeki env filtrelemesini kullanmıyor (HATALI güvenlik farkı). |
| U07 | Testler bölümü | VAR | `gui.html:176-177; gui.go:743` | Gerçek test raporu ve iptal. |
| U08 | İşlem günlüğü | VAR | `gui.html:115-116; journal.go:16` | Gerçek günlük ve redaksiyon. |
| U09 | GPU/model durumu | KISMİ | `gui.html:202; worker_client.go:59-119` | Hesap/model durumu var; gerçek GPU kapasitesi/tahsis durumu unknown olabilir. |
| U10 | Hesap/kredi bilgileri | VAR | `gui.html:199-202; gui.go:1076` | Sunucudan yenilenen durum; canlı doğrulama yapılmadı, US sahipliği. |
| U11 | Ayarlar bölümü | KISMİ | `gui.html:206-218` | Gelişmiş bağlantı ayarları var; bütünleşik terminal/çalışma alanı/kaynak politikası ayarları eksik. |
| R01 | Sahte başarı mesajı olmaması | KISMİ | `model.go:556,599; gui.go:817` | Çoğu sonuç gerçek; başarısız startProgram sonucunda da “Program çalıştırıldı” günlüğü yazılır (HATALI ifade). |
| R02 | Çalışmayan teste başarı yazmama | VAR | `project_test_runner.go:83-95,139-152` | Sıfır test ve kanıtsız exit=0 başarı sayılmaz. |
| R03 | Model hazır değilken hazır göstermeme | KISMİ | `worker_client.go:87-95,114-116` | Accounts hazır bilgisi var; legacy hook yolunda boş status da Awake kabul ediliyor. Canlı hazır kanıtı değil. |
| R04 | Sonuçların gerçek duruma bağlı olması | KISMİ | `terminal.go:127-136; gui.go:817; worker_client.go:94` | Terminal gerçek; çalıştır günlüğü ve legacy hazır yorumu düzeltilmeli. |

## §3 — her maddenin kaynak karşılığı

| ID | Gereksinim | Karşılık | Dosya:satır kanıtı | Eksik / hatalı / açıklama |
|---|---|---|---|---|
| M00 | Mevcut DeepSeek bağlantısını koruma | VAR | `context_limits.go:51-68; model.go:2214-2217` | Kaynakta model seçimi korunmuş; bu görevde kod değiştirilmedi. |
| M01 | Gerçek model API bağlantısı | KISMİ | `model.go:2172-2299` | Gerçek HTTP istemcisi var; gerçek model uç bağlantısı BLOCKED/kapsam dışı, mock ile doğrulanmalı. |
| M02 | SSE streaming | VAR | `model.go:2292-2294,2361` | SSE okuyucu/progress; canlı akış değil birim kapsamı. |
| M03 | Uzun görev desteği | KISMİ | `long_context.go:93; long_checkpoint.go:11-18` | Parçalı özetleme/checkpoint var; tam proje yaşam döngüsü değil. |
| M04 | Çok aşamalı model çağrısı | VAR | `model.go:319-425; long_context.go:214` | Okuma/yeniden istem/özet birleştirme aşamaları. |
| M05 | Konuşma bağlamı yönetimi | VAR | `journal.go:254,278; model.go:800` | Sohbet belleği ve sınırlı bağlam. |
| M06 | Token sınırı kontrolü | KISMİ | `context_limits.go:13-15,29-48,72; model.go:2183-2188` | Çıkış token limiti var; giriş byte bütçesi, gerçek tokenizer sayımı değil. |
| M07 | Model hazır olma denetimi | KISMİ | `worker_client.go:59-119; licensesrv/model_proxy.go:413` | Model listesi/status denetimi; gerçek GPU readiness uçtan uca NOT TESTED. |
| M08 | Yeniden bağlanma | KISMİ | `model.go:2057; model.go:2278-2284; continuation.go:58` | Uç listesi/yanıt devamı var; SSE kesintisinde güvenli replay ve idempotent resume garantisi yok. |
| M09 | İptal işlemleri | VAR | `model.go:206,2197; gui.go:440` | Context iptal yolu var. |
| M10 | Eşzamanlı istek yönetimi | VAR | `model_admission.go:47; licensesrv/model_proxy.go:27-43,90-97` | Yerel admission ve sunucu kapasite kanalları var. |
| M11 | GPU kapasitesine göre kuyruk | KISMİ | `model_admission.go:11-12,28; licensesrv/model_proxy.go:27-43` | Sabit yapılandırılan kapasite var; GPU telemetrisinden dinamik türetme yok. |
| M12 | Kullanıcı bazlı kullanım kayıtları | VAR | `account_client.go:101-144; licensesrv/accounts.go:554,614-626` | Accounts/istek-kredi geçmişi mevcut; US alanı, canlı kabul değil. |
| M13 | Yanıtların güvenli dosya değişiklikleri olarak uygulanması | KISMİ | `proposal.go:42,156; proposal_batch.go:22,50; change_history.go:30,71` | Okuma kanıtı, hash, önizleme/geri alma var; OS düzeyi yarış/izolasyon sertleştirmesi ayrıca gerekli. |
| E01 | Yetenekleri gereksiz uygulama kısıtlarıyla daraltmama | KISMİ | `model.go:2223,2234-2242; context_limits.go:29-48` | think=false, reasoning/stop/output varsayılanları var; model desteğine göre yapılandırılabilirlik ve kalite karşılaştırması gerekli. |
| E02 | Açık, esnek, teknik olarak güçlü deneyim | KISMİ | `gui.html:206-218; model_actions.go:34` | Özel uç/model ve yapılandırılmış dosya eylemi var; bağımlılık/build/otomatik onarım eksikleri sürüyor. |
| E03 | Yetki, güvenlik, veri koruma ve kötüye kullanım önleme korunmalı | KISMİ | `gui.go:127; project_test_runner.go:24; run.go:41-49; model_admission.go:47` | Panel guard/env/admission var; Run env mirası ve OS/tenant sandbox eksikliği önemli. |
| E04 | “Filtresiz” pazarlama gerçeği temsil etmeli; güvenlik kaldırılmamalı | KISMİ | `gui.html:190,223; model.go:2223-2242` | Teknik sınırlamalar bazı yerlerde açıklanıyor; gerçek model davranışı/pazarlama uyumu NOT TESTED. |

## Öncelikli eksik / hatalı bulgular

1. **HATALI / NOT TESTED güvenlik:** `run.go:41-49,78-82` Çalıştır yolu tüm ebeveyn ortamını çocuk sürece miras veriyor. Terminal/test yolunun `projectTestEnv` filtresiyle tutarsız. Gizli değer sızıntısı test edilmedi/oluşturulmadı; statik veri akışı riski var.
2. **EKSİK / NOT TESTED:** allowlist ve cwd sandbox değildir. İzinli Python/Go/Node proje kodu OS yetkisiyle kök dışına erişebilir. CPU/RAM/PID/disk/ağ/tenant izolasyonu yok (`terminal.go:16`, `gui.html:190`). “Isolation” adlı test terminal çıktı ayrımını test eder; OS izolasyonu kanıtı değildir (`terminal_test.go:24-58`).
3. **EKSİK / NOT TESTED:** `gui.go:785-796` test hatası için bir model onarım isteği üretir; öneri uygulanınca build/test tekrarını bütçe, iterasyon ve iptal denetimiyle sürdüren genel ajan döngüsü yok.
4. **HATALI ifade / NOT TESTED:** `gui.go:813-817` başarısız başlatmada dahi “Program çalıştırıldı” günlüğü yazar. Gerçek sonuç koduna bağlı mesaj yapılmalı.
5. **KISMİ / NOT TESTED:** `worker_client.go:94` legacy hook yolunda eksik model status alanı da Awake sayılır; “bilinmiyor” ile “hazır” ayrılmalı. Accounts yolunda `ModelReady` kullanılır, canlı GPU kanıtı sayılmaz.
6. **EKSİK:** bağımlılık kurulumu/sürüm kilitleme/izin ve geri alma akışı yok. TS/YAML/SQL dosya düzenleme, çalıştırma ve doğrulama ile karıştırılmamalı.
7. **KISMİ:** checkpoint uzun girdi özetlerini; recovery dosya transaction'ını kapsıyor. Genel build/test/onarım sürecinin kalıcı ve idempotent yeniden başlatması henüz kabul edilmedi.

## US sahipliği — dokunulmayacak alanlar

Aşağıdaki yollar her üç kaynak kopyasında da US alanı olarak korunmalı; bu görev hiçbirini değiştirmedi:
- `cmd/us-license-hook/**` (özellikle `main.go`, `main_test.go`, `accounts_routes_test.go`).
- `licensesrv/**`: accounts, account_workers, model_proxy/model_meter, credits, store/lock, webhook, mail, pricing/plans/reversals ve testleri.
- EXE/release: `US_DERLE.bat`, `release.go`, `embedded_key.go`, `embedded_license.go`, `pricing.go`, `cmd/us-pricing-gen/**`, Windows kaynakları `*_windows.go`, `check_ui.mjs`, `us.ico`, `*.syso` varsa.
- Hesap/release entegrasyonunun ortak dosyaları: `account_client.go`, `worker_client.go`, `commercial_policy.go`, `license.go`, `gateway.go`, `tunnel*.go`, `cloud_install.go`, `cloud_space.go`, `main.go`, `go.mod`, `go.sum`, ilgili account/hook/smoke testleri. Bu liste sahiplik çakışmasını önlemek için ihtiyatlıdır, hiçbir dosyanın devri varsayılmaz.
- `/workspace/us-panel/dist/**`, `/workspace/us-panel/deploy/**`, mevcut EXE'ler ve build raporları. Orijinal kaynakta tek yeni dosya bu talep edilen gap raporudur.
- `gui.go`, `gui.html` hesap/terminal ekranlarını bir arada içeriyor: yalnız ayrı kopyada panel kapsamı değişebilir; account/release bloklarını değiştirmek ayrıca koordinasyon ister. Ortak dosyayı bütünüyle geri kopyalamak yasak; yalnız gözden geçirilmiş dar patch aktarımı önerilir.

## Öncelikli uygulama planı (onay sonrası, küçük ve test edilebilir)

1. **Güvenli terminal sözleşmesini sertleştir.** Mevcut çalışan allowlist/stdout/stderr/timeout/iptali yeniden kullan, yeniden yazma. Alt komut bayraklarını türlenmiş politikayla doğrula; Çalıştır/test/terminal için aynı temiz env ve kök kurallarını kullan. Kabul: shell/flag/path kaçışı reddi, stdout/stderr + exit 7, timeout, iptal/çocuk süreç ölümü ve gizli env taşınmaması testleri PASS. OS sandbox kararı verilmeden çok kullanıcılı güvenlik PASS denmesin.
2. **Kök dışına çıkmayan dosya araçları.** Mevcut safePath/sourceSnapshot/proposal hash/transaction altyapısını kullan; okuma-yazma-silme kapsamı, symlink/rename yarışları, boyut sınırı ve atomik yazım sözleşmesi ekle. Kabul: `..`, absolute/symlink, kök değişimi, eski hash, kısmi yazım rollback ve iki çalışma alanı testleri PASS; US dosyaları değişmemeli.
3. **Build→hata→düzeltme döngüsü.** Yapılandırılmış komut sonucu (argv/cwd/exit/stdout/stderr/durum), açık kullanıcı onayı, iterasyon/süre/çağrı bütçesi. İlk kabul Go fixture: ilk build başarısız → httptest OpenAI/SSE önerisi → izinli patch → yeniden build/test PASS; ayrıca tekrarlayan hata/bütçe/iptal/bozuk SSE durma testleri. Gerçek model kullanılmayacak.
4. **Kalıcı durum ve kesintiden devam.** Sürümlemeli task state, atomik checkpoint, çalışma alanı/model/policy/hash bağlama ve işlem kimlikleri; mevcut `.us-tasks` kayıtlarıyla uyum/migration. Kabul: alt süreç öldürülüp aynı görev yeniden açıldığında dosya değişikliği çift uygulanmaz, tamamlanmış aşama tekrarlanmaz, belirsiz istek otomatik ücretli replay edilmez; bozuk kayıt korunur.
5. **Panel ve kullanıcı model deneyimi kabulü.** Gerçek sonuç etiketleri, bilinmeyen/hazır ayrımı, görev durumu/resume ve yapılandırılabilir model limitleri. Mock ile API/GUI regresyonları + vet/race PASS; no-test/no-model durumunda başarı iddiası yok. Native Windows, gerçek model ve canlı entegrasyonlar ayrı NOT TESTED kalır; US .7 çalışmasına karışılmaz.

## Engeller ve kararlar

- **Onay kapısı:** ebeveyn uygulamayı onaylamadan kod değişikliği başlanmayacak.
- **BLOCKED:** Git geçmişi/remote yok; branch/worktree yok. Demir kanonik tabanın bu work2 kopyası olduğunu ve ileride Git başlangıcı mı yoksa bilinen repo geçmişi mi kullanılacağını kararlaştırmalı. Bu görev Git init/commit yapmadı.
- **Karar:** hedef güven modeli güvenilir tek kullanıcı desktop mı, yoksa güvenilmeyen çok kullanıcılı hizmet mi? İkincisi container/OS sandbox ve ağ politikası gerektirir; mevcut allowlist yeterli değil.
- **Karar:** CPU/RAM/PID/disk/ağ limitleri, izinli bağımlılık kaynakları ve ajan süre/iterasyon/model çağrı bütçeleri sayısal olarak belirlenmeli.
- **Koordinasyon:** US hook/accounts/.7 dosya listesi kilitli; gui/run gibi paylaşılan alanların patch devri ve sıra/sürüm mutabakatı gerekli.
- **BLOCKED/NOT TESTED:** gerçek model ve Windows kabulü için sonradan ayrı test ortamı gerekir. Mevcut kapsamda VPS 31.220.63.191, Shopify veya GCP bağlantısı/deploy **yapılmadı**.
- Kod-ajanı devri için aranan CloudAgent aracı bu alt görevin araç kataloğunda bulunmadı. Bu keşifte yalnız zaten box'ta bulunan dosyalar okundu ve yerel test/rapor üretildi; kod uygulaması yapılmadı. Uygulama onayında ebeveynin uygun kod ajanına devretmesi gerekir.

## Kesintiden devam kaydı

- Aşama: **DISCOVERY_COMPLETE / WAITING_FOR_IMPLEMENTATION_APPROVAL**.
- Çalışma alanı: `/workspace/us-panel-dev/`; aktif modül `/workspace/us-panel-dev/work2/kiyas/term/`; branch **yok (Git değil)**.
- Kaynak kod değişikliği: **0**. Hook/EXE değişikliği: **0**. Canlı işlem: **0**.
- Rapor: `/workspace/us-panel/reports/PANEL_S2_GAP_20261106.md`; tam kopyası ve uygulama planı: `/workspace/us-panel-dev/PANEL_DEV_STATUS.md`.
- Başlarken bu dosya + `panel-evidence/baseline-sha256.json` + `snapshot-drift.json` okunmalı; US'nin son kaynak hashleriyle tekrar karşılaştırılmalı. Fark varsa otomatik overwrite/senkronizasyon yapılmamalı.
- Sıradaki eylem: uygulama onayı ve sandbox/bütçe/sahiplik kararları geldikten sonra yalnız 1. adım için mock/fixture testleri ve dar değişiklik planı; canlı modele veya dağıtıma geçiş yok.


# Uygulama görevi alındı — 06.11.2026 08:09 TSİ (Europe/Istanbul)

Önceki WAITING_FOR_IMPLEMENTATION_APPROVAL kaydı tarihsel keşif durumudur; uygulama talebi artık alındı. Güncel durum: **BLOCKED_REQUIRED_CODING_AGENT_UNAVAILABLE**.

Zorunlu kod-ajanı devir aracı `CloudAgent` bu yürütücünün araç kataloğunda bulunmuyor. `GetDynamicTools` ile CloudAgent/cloud_agent/coding/agent-launch arandı (eşleşme yok); kurulu connector durumları ve eklenti kataloğu da kontrol edildi, kullanılabilir CloudAgent aracı bulunamadı. Alternatif CLI/API/browser üzerinden bu zorunlu devri aşma girişimi yapılmadı. Kod uygulaması başlanmadı.

| Adım | Durum | Yeni test / vet / race | Yama |
|---|---|---|---|
| 1 Terminal + Çalıştır güvenliği ve Linux kaynak sınırları | BLOCKED | NOT TESTED | Üretilmedi |
| 2 Güvenli ve atomik dosya araçları | BLOCKED | NOT TESTED | Üretilmedi |
| 3 Mock OpenAI derleme→onarım döngüsü | BLOCKED | NOT TESTED | Üretilmedi |
| 4 Kalıcı görev/resume ve legacy hazır durumu | BLOCKED | NOT TESTED | Üretilmedi |

- Değişen uygulama dosyası: **0**. Bu devam turunda yalnız bu durum kaydı ve `panel-evidence/implementation-blocker.json` yazıldı.
- Önceki başlangıç vet/race PASS kanıtları korunuyor; bunlar yapılmamış uygulama adımları için PASS değildir.
- Windows kaynak sınırı/çalışması: **NOT TESTED**.
- Orijinal `/workspace/us-panel/` ve US hook/EXE dosyalarına bu devam turunda dokunulmadı. Canlı bağlantı/deploy/mesaj yok.
- Sahiplik çakışması: talepteki `worker_client.go` legacy status düzeltmesi önceki US koruma listesinde; ayrıca `gui.go`/`gui.html` hesap ve panel işlevlerini birleştiriyor. Kod ajanına devredilirken bu alanlarda yalnız açıkça izin verilen dar panel değişiklikleri sınırlandırılmalı; hook/EXE blokları değiştirilmemeli.
- Devam yolu: ebeveyn, CloudAgent aracı bulunan bir yürütücüye bu dört aşamalı görevi ve mevcut çalışma alanını devretmeli. Kaynak zaten box'ta; yeniden clone veya dışarıya aktarım yapılmadı.


## Sahiplik netleştirmesi — 06.11.2026 08:10 TSİ (Europe/Istanbul)

- `worker_client.go` ve önceki US koruma listesindeki diğer dosyalar ana geliştirme kopyasında da değiştirilmeyecek.
- Legacy hazır-status düzeltmesi yalnız `patches/US-PROPOSAL-legacy-status.patch` önerisi olarak hazırlanacak; geçici kopyada uygulanarak test edilecek, ana kopyaya uygulanmayacak. Bu öneri henüz hazırlanmadı/test edilmedi.
- Kaynak yalnız bu makinede, forge/Git geçmişi yok; clone yapılmayacak.
- Uygulama halen **BLOCKED**: mevcut kodlama iş akışı yerel geliştirme için de kod ajanına devri zorunlu tutuyor; yerel kaynak/klonlama istisnası doğrudan uygulama yetkisi sağlamıyor. Gerekli CloudAgent aracı bu yürütücüde bulunmuyor. Kullanıcı talebi bu araç kısıtını değiştirmedi.
- Bu turda kod/yama değişikliği yok. Adım 1–4 BLOCKED; yeni vet/race sonuçları NOT TESTED. Yalnız bu durum belgesi güncellendi.
