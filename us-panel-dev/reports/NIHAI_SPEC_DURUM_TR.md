# Nihai prompt madde matrisi — .5

06.11.2026 06:12 TSİ (Europe/Istanbul)

PASS yalnız yazılan yerel kapsamdır. Fake-provider PASS gerçek GPU kabulü değildir; canlı doğrulama sınırları her satırdadır.

| Madde | Gereksinim | Durum | Kanıt / sınır |
|---|---|---|---|
| §2 ajan | Doğal dil görev sınıflandırma | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Çok dosya action protokolü | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Dosya okuma/değiştirme | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Hata çıktısı/onarım döngüsü | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Durumu kaydetme | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Kesintiden devam | PASS | Korunan intent, file-action, proposal/repair, task queue/history ve checkpoint fixture testleri. Gerçek model kalitesi bu PASS kapsamı değil. |
| §2 ajan | Gerçek modelle mimari oluşturma | NOT TESTED | Wire/control flow var; dış model ve kapsamlı proje üretimi yapılmadı. |
| §2 ajan | Gerçek proje görev aşamaları | NOT TESTED | Wire/control flow var; dış model ve kapsamlı proje üretimi yapılmadı. |
| §2 ajan | Modelün gerçek test oluşturma kalitesi | NOT TESTED | Wire/control flow var; dış model ve kapsamlı proje üretimi yapılmadı. |
| §2 ajan | Bağımlılıkların otomatik kurulması | BLOCKED | Ağ yazması/kurulum bu tur yasak; eksik runtime gerçek hata verir. Kullanıcı açık eylemi olmadan lifecycle script çalışmaz. |
| §2 terminal | Yetkilendirilmiş komutlar | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | Çalışma dizini | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | Gerçek stdout/stderr | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | İşlem iptali | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | Zaman aşımı | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | Gerçek exit code | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | Çıktı ve eşzamanlı işlem sınırı | PASS | Açık kullanıcı JSON argv, shell yok; gerçek Python/GUI, exit7/cancel/timeout testleri; tree kill Linux testli. |
| §2 terminal | CPU/RAM kaynak sınırı | BLOCKED | Container/cgroup/Job sandbox yok. Cwd/env/çıktı/time sınırı OS sandbox değildir; güvenilir proje kodu kullanıcının OS yetkisiyle çalışır. |
| §2 terminal | Kullanıcılar arası güçlü OS izolasyonu | BLOCKED | Container/cgroup/Job sandbox yok. Cwd/env/çıktı/time sınırı OS sandbox değildir; güvenilir proje kodu kullanıcının OS yetkisiyle çalışır. |
| §2 ortam | Python | PASS | Gerçek yerel runner/test fixture; Go launch ve runtime hata denetimi; Windows runtime kabulü değil. |
| §2 ortam | Go | PASS | Gerçek yerel runner/test fixture; Go launch ve runtime hata denetimi; Windows runtime kabulü değil. |
| §2 ortam | Node.js / JavaScript | PASS | Gerçek yerel runner/test fixture; Go launch ve runtime hata denetimi; Windows runtime kabulü değil. |
| §2 ortam | TypeScript | NOT TESTED | Genel metin/dosya editörü korunur; gerçek compiler/browser/DB kabulü çalıştırılmadı. JSON parser/depo testleri ayrı PASS. |
| §2 ortam | HTML/CSS native görünüm | NOT TESTED | Genel metin/dosya editörü korunur; gerçek compiler/browser/DB kabulü çalıştırılmadı. JSON parser/depo testleri ayrı PASS. |
| §2 ortam | YAML semantik | NOT TESTED | Genel metin/dosya editörü korunur; gerçek compiler/browser/DB kabulü çalıştırılmadı. JSON parser/depo testleri ayrı PASS. |
| §2 ortam | SQL/PostgreSQL runtime | NOT TESTED | Genel metin/dosya editörü korunur; gerçek compiler/browser/DB kabulü çalıştırılmadı. JSON parser/depo testleri ayrı PASS. |
| §2 UI | Terminal | PASS | GUI handler/source/fixture kanıtı; gerçek Windows görsel kabulü BLOCKED. Test exit0 tek başına green değildir; unknown/zero/skipped-all doğrulanmadı olur. |
| §2 UI | Projeler/Dosyalar/Kod editörü | PASS | GUI handler/source/fixture kanıtı; gerçek Windows görsel kabulü BLOCKED. Test exit0 tek başına green değildir; unknown/zero/skipped-all doğrulanmadı olur. |
| §2 UI | AI Ajanı/Çalıştır/Testler/Günlük | PASS | GUI handler/source/fixture kanıtı; gerçek Windows görsel kabulü BLOCKED. Test exit0 tek başına green değildir; unknown/zero/skipped-all doğrulanmadı olur. |
| §2 UI | GPU bekleme/readiness | PASS | GUI handler/source/fixture kanıtı; gerçek Windows görsel kabulü BLOCKED. Test exit0 tek başına green değildir; unknown/zero/skipped-all doğrulanmadı olur. |
| §2 UI | Lisans kredileri/Ayarlar | PASS | GUI handler/source/fixture kanıtı; gerçek Windows görsel kabulü BLOCKED. Test exit0 tek başına green değildir; unknown/zero/skipped-all doğrulanmadı olur. |
| §2 UI | Tam hesap girişi/oturum | BLOCKED | Go lisans-scoped Bearer var; Go kullanıcı/parola hesabı yok, Python account formu bilerek kapalı. |
| §3 | Model HTTP API kodu | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | SSE streaming | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Uzun görev | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Çok aşama | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Konuşma bağlamı | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Token bütçesi | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Readiness/reconnect | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | İptal | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Eşzamanlılık/kuyruk | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Kullanım kaydı | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Güvenli dosya uygulama | PASS | Yerel HTTP proxy ve korunan long-context/cancel/proposal testleri. Server reserve/settle; kapasite1/queue32; scoped kullanım; gerçek DeepSeek değildir. |
| §3 | Gerçek DeepSeek / 45 dk cold start | NOT TESTED | Canlı model/GCP/SSH yok. Per-license kayıt var; tam hesap ve filo lease adapter ayrı BLOCKED. |
| §3 | Gerçek GPU kapasitesi/TTFT/model kalitesi | NOT TESTED | Canlı model/GCP/SSH yok. Per-license kayıt var; tam hesap ve filo lease adapter ayrı BLOCKED. |
| §3 | Güvenliği kaldırmadan esneklik | PASS | Auth/path/onay/token bounds korunur, commercial direct bypass kapalı; doğrulanmamış filtresiz/milli pazarlama iddiası eklenmedi. |
| §4.1 | Boşta gereksiz açık kalmama | BLOCKED | 900 sn controller policy testli; aktif Go provider/watchdog adapter yok. |
| §4.2 | İstekte otomatik GPU açılışı | BLOCKED | Go yalnız upstream HTTP çağırır; VM start bağlı değil. |
| §4.3 | Tek GPU start paylaşımı | PASS | Mevcut Python controller 10 parallel fake-provider testi; Go aktif entegrasyonu BLOCKED. |
| §4.4 | Hazır olana kadar güvenli bekleme | PASS | 50 dk client/45 dk server ve heartbeat; gerçek cold-start NOT TESTED. |
| §4.5 | Aktif görevde kapanmama | PASS | Controller lease guard fake-provider testli; Go adapter BLOCKED. |
| §4.6 | Kuyruk varken kapanmama | PASS | Queued lease fake-provider testli; Go proxy kuyruğunu controller lease ile bağlama BLOCKED. |
| §4.7 | Idle yapılandırması | PASS | WorkerPolicy idle_seconds ve config örneği. |
| §4.8 | Varsayılan idle 15 dk | PASS | Demir kararı eski spec 10 dk yerine 900 sn: 899 açık /900 stop fake test. |
| §4.9 | 10 dk boot bekleme | PASS | WorkerPolicy600 ve dev SSH600; .5 HTTP 50 dk. |
| §4.10 | Kontrollü boot retry | PASS | Korunan operation-id/ambiguous-start fixture testleri; Go provider bağlı değil. |
| §4.11 | Gerçek watchdog durumu | NOT TESTED | Reconcile/fake-provider testli; GCP/IAM canlı sorgulanmadı. |
| §4.12 | Least privilege IAM | BLOCKED | Canlı izin değişikliği yok; HOST yetkili kabul gerekli. |
| §4.13 | Eski wake devreden çıkarma | BLOCKED | Canlı servis değişikliği yok; mevcut bağımlılıklar sorgulanmadı. |
| §4.14 | GPU maliyet kaydı | NOT TESTED | Model kullanım saniyesi testli; gerçek provider runtime/fatura uzlaşması yok. |
| §4.15 | Health GPU başlatmaz | PASS | Go read-only health/models ve mevcut controller fake health/wake testleri. |
| §1 | Yeni taban/sürüm | PASS | work2 .5; source/work korunur. ROUND5_CHANGED_FILES_20261106.txt. |
| §5 | Cloud Run 404/deploy | BLOCKED | Karar: canlı Cloud Run issuer yok; canlı çağrı/deploy yasak. Tarihsel 404 gerçek güncel neden sayılmaz. |
| §6 | Shopify/credit/cancel/refund | PASS | Signed Go fixture ve atomic credit/model proxy testleri; gerçek Shopify değil. |
| §6 | Hesap/mail/uzlaşma | BLOCKED | Go lisans-scoped erişim var; hesap, SMTP ve commit sonrası replay/crash uzlaşması eksik. |
| §7 | Türkçe/koyu UI kaynak ve JS | PASS | Dosya ve node syntax check kanıtı; native görünüm/gerçek şirket iddiası NOT TESTED. |
| §8 | Yerel 1/5/10/25/50 proxy yükü | PASS | Fake upstream max active<=1 ve başarı kadar debit; gerçek GPU benchmark değil. |
| §9 | Gerçek kapsamlı AI proje | BLOCKED | Gerçek model/project opt-in kapalı; dış çağrı yapılmadı. |
| §10 | Windows amd64 PE/build/hash | PASS | Cross-build doğrulandı; native Windows BLOCKED. |
| §11 | Maliyet varsayımları | PASS | FIYATLANDIRMA_TR.md dosya kanıtı: 5,58USD/h,%30,55TL; gerçek fatura NOT TESTED. |
| §12 | Test/vet/race/build | PASS | Raw go_*_topyekun_20261106_v2.txt, KANUS_TEST_RESULTS.md. |
| §13 | Canlı işlem sınırları | PASS | VPS/Shopify/GCP/SSH/SMTP/deploy/sipariş/ödeme yok, gerçek sırlar açılmadı. |
| §14–15 | Rapor/açık engel | PASS | İstenen KANUS_* ve smoke/dry-run dosyaları; üretim hazır denmedi. |

- PASS — Son E2E geri bildirim düzeltmeleri: order.name mail/restart/partial refund ve lisanslı GET status kind/remaining/expiry testleri; 282+37 Go PASS. Bunlar canlı deploy iddiası değildir.
