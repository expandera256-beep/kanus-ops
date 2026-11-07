# KANUS PERF — yerel MOCK doğrulama ve gerçek düşük hızlı ağ gözlemi

Güncelleme: 06.11.2026 09:43:50 TSİ (Europe/Istanbul).

**MOCK ≠ gerçek GPU kabulü.** US Panel uygulama kaynakları değiştirilmedi; öneriler yalnız diff. Önceki CloudAgent engeli notu artık bu scratch harness için geçersiz; yerel araçlar yazıldı ve çalıştırıldı.

## Karar ve güvenlik
- **Demir 06.11.2026 09:28 TSİ: idle15 dakika (900sn). Prompttaki10 dakika geçersiz.** SaaS senaryosu, logical-clock testi ve pencere hesabı bu kararı kullanır.
- Yük testlerinin tamamı sayısal127.0.0.1 hedeflere. Test/prod IP adreslerine doğrudan HTTP/TCP/SSH, /etc sır erişimi, GPU start/status/stop veya kiralama yok.
- Son açık talebe göre yalnız https://api-test.kanus.org/healthz için3 GET; her prob bitişinden sonra2sn bekleme. Egress bilgisi için ipinfo.io tek deneme. Token/auth olmadan, redirect/retry/-k/--resolve yok. Gerçek modele hiçbir çağrı yapılmadı.
- Rastgele fixture sırları yalnız geçici süreç/env/store içinde; runtime logları saklanmaz, ledger silinir. Gerçek secret/env dosyaları okunmadı.

## GLOBAL_PROMPT kapsamı
- **§8 (426–444):** mevcut A100 korunur; önce tek bölgeden ABD/Avrupa/diğer hedeflerin API gecikmesi, ilk token, toplam yanıt, ağ maliyeti ve deneyimi ölçülür. Talep olmadan pahalı GPU bölgeleri açılmaz; adaptör/yönlendirme modüler. Sayısal SLO yok.
- **§9 (448–468):** 1/5/10/25/50 kullanıcı; p50/p95/p99, hata, kuyruk, GPU kullanımı, kredi tutarlılığı, istek maliyeti, yanıt bütünlüğü. Kapasite aşımında kontrollü kuyruk/geri basınç. Sayısal kabul eşiği yok.
- **§11 (495–521):** gerçek çok dosyalı SaaS; backend/frontend/PostgreSQL/auth/RBAC/görev yönetimi/REST/test/Docker/CI-CD/dokümantasyon; gerçek build/run/test ve modele hata geri besleme. Sayısal işlem/maliyet sınırı verilmemiş; öneriler aşağıda.

## Kaynak mimarisi ve kredi
- Ticari EXE → Go hook /v1/chat/completions → US_MODEL_UPSTREAM → VPS özel tüneli → ds-a100. Yerel test bu tüneli hiç kullanmaz; mock localhost upstream kullanır.
- İncelenen kaynak: work2/kiyas/term/cmd/us-license-hook/main.go, licensesrv/model_proxy.go/model_meter.go; kurulum sözleşmesi deploy/us-license-hook/README_TR.md. GET /healthz hazır model kanıtı değildir; GET /v1/credits/status ve POST /account/balance kredi görünümüdür. Bearer lisans veya hesap oturumu gerekir.
- Atomik admission→reserve→upstream→settle; hata/iptal release. Sub=request1; hourly=ceil upstream wall-seconds, queue hariç. Bu suite sub/request1 ölçer; **hourly saniye kredisi canlı veya bu Go suite içinde doğrulanmadı**.
- Go konfigürasyon üst sınırları: capacity1..10, queue1..100; fixture ana ölçüm capacity3/queue64/deadline5sn. İlk fixture queue128 konfigürasyonu503 üretti; harness64 ile düzeltildi, nihai dosyalardaki sonuçlar yalnız düzeltilmiş koşudur. Ürün sorunu diye raporlanmaz.
- Mevcut kaynakta başka ajanlar değişiklik yapabilir: JSON Go kaynak öncesi/sonrası hash ve binary hash içerir; nihai derlemede kaynak sabit: True.

## Çalıştırılan araçlar ve kanıt
```bash
python3 /workspace/us-panel/perf/run_suite.py
python3 /workspace/us-panel/perf/cli_smoke.py
python3 /workspace/us-panel/perf/saas/verify_project.py --mock
python3 /workspace/us-panel/perf/idle_contract.py
python3 -m compileall -q /workspace/us-panel/perf
python3 /workspace/us-panel/perf/latency.py --url https://api-test.kanus.org/healthz --allow-public-health --count 3 --interval 2 --egress
```
- Python fixture kontrolleri PASS; CLI smoke PASS; gerçek Go hook offline build exit0; SaaS sabit fixture compileall+1 unittest PASS; idle logical-clock7/7 PASS. Logical-clock GPU yaşam döngüsü kodu veya gerçek15dk bekleme testi değildir.
- Standard-library Python; pip/npm kurulmadı. Dosyalar: mock_stack.py, run_suite.py, load.py, latency.py, cli_smoke.py, idle_contract.py, README.md; saas/PROJECT_PROMPT_TR.md ve verify_project.py; proposals/*.diff ve kaynak fingerprint; results/*.json/*.md, go-build.log; bin/hook.

## MOCK sonuçları — değiştirilmemiş Go hook + yerel mock upstream
4 ardışık batch/kullanıcı seviyesi; kapalı döngü, kısa smoke. TTFT=ilk delta.content; heartbeat sayılmaz. Sentetik4 içerik parçası, TTFT20ms, rate150parça/sn. Gerçek model token throughput değil. p99 az örnekte SLO kanıtı değildir.

| Kullanıcı | n | p50 ms | p95 ms | p99 ms | TTFT p50 ms | Hata | Beklenen/gerçek kredi |
|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 4 | 46.33 | 46.91 | 46.98 | 22.57 | 0% | 4/4 |
| 5 | 20 | 50.28 | 97.08 | 97.75 | 24.87 | 0% | 20/20 |
| 10 | 40 | 99.38 | 190.73 | 194.17 | 75.24 | 0% | 40/40 |
| 25 | 100 | 238.92 | 403.83 | 434.37 | 215.39 | 0% | 100/100 |
| 50 | 200 | 443.52 | 831.25 | 889.43 | 414.94 | 0% | 200/200 |

Ana yüklerde toplam364/364 tam yanıt, kredi364/364, hold0. Go hook queue süresini dışarı sunmuyor: N/A; farktan uydurulmadı. Python hook ölçülmüş kuyruk dağılımı ham JSON/MD içinde. GPU kullanımı N/A. Mock fiyat varsayımı$0.002/kredi, normal başarılı istek maliyeti$0.002; gerçek GPU maliyeti ölçülmedi. Seeded25% hata konfigürasyonunda50 istekte16 hata (%32), başarı34 ve kredi34/34.

### Kapasite aşımı / back-pressure
- Python hook capacity1/queue2/timeout50ms/50 eşzamanlı: **47×429 +2×503 +1 başarı**; tüm429/503 Retry-After:1; beklenen/gerçek düşüm1/1; hold0.
- Upstream bağımsız capacity1/queue1/timeout30ms/10 istek:8×429+1×503+1başarı.
- Gerçek Go hook capacity1/queue2/request timeout1sn; mock TTFT1.5sn/50 istek: **47×429 +2×504 +1×HTTP200 SSE timeout/error**. Tam başarı0; kredi0/0; hold0. HTTP200 tek başına başarı değildir.
- **Bulgu:** Go429/504 yanıtlarında Retry-After yok; kuyruk gözlemlenebilirliği de yok. Timeout504 mevcut API davranışı,503 diye raporlanmadı.

### Yanıt bütünlüğü / kredi — strict upstream policy
- Go truncate, malformed JSON, SSE error, HTTP503 upstream ve client disconnect: tamamlanma yok; düşüm0; hold0. Normal içerik+stop+DONE:1 kredi.
- **Önemli:** upstream `finish_reason=stop` ardından EOF, `[DONE]` eksik olduğunda Go1 kredi düşüp kendi downstream[DONE] işaretini üretiyor. İstemci açısından stream tam; strict upstream finish+DONE oracle açısından **FAIL/POLICY OPEN**. Hatalar saklanmadı; JSON product_policy_failures=[missing_done].
- Bu davranış stop terminalini yeterli kabul eden ürün tasarımı olabilir. Katı upstream[DONE] şartı isteniyorsa değişiklik gerekir; karar vermeden üretim bug/haksız ücretlendirme kesinleştirilmez.

### Uygulanmayan öneriler
- proposals/01-backpressure-observability.diff:429/504 için Retry-After1 ve X-US-Queue-Wait-Ms;504 korunur. Jitter/backoff istemci tarafında ayrıca gerekir.
- proposals/02-strict-upstream-done-policy.diff:finish_seen+upstream[DONE] gelmeden settle etme; eksik terminal/EOF/timeout release. Politika kararı şart.
- Hiçbiri US koduna uygulanmadı veya derlenip test edilmedi; proposals/README.md regresyon listesi içerir. Kaynak hash dosyasıyla bağlam doğrulanmalı.

## GERÇEK düşük hızlı ağ gözlemi — GPU değil
Zaman: 2026-11-06T09:31:17.942789+03:00 (TSİ).3×HTTP200, TLS doğrulama sonucu0, her response2bayt. Prob başlangıç aralıkları3.126 ve2.875sn (bitişler arası2sn).

| Örnek | DNS ms | TCP connect farkı ms | TLS farkı ms | Health TTFB ms | Toplam ms |
|---:|---:|---:|---:|---:|---:|
| 1 | 1.263 | 0.245 | 409.921 | 1107.959 | 1108.067 |
| 2 | 1.402 | 0.258 | 465.809 | 863.709 | 863.761 |
| 3 | 1.125 | 0.285 | 496.227 | 1095.206 | 1095.261 |

Medyan health TTFB: **1095.206ms**; LLM TTFT ölçülmedi.
- **Ağ sınırlaması:** curl remote_ip=198.18.0.1 (benchmark/ara ağ aralığı). Ölçümler box ağının gözlemidir; TCP~0.25ms gerçek hedef RTT iddiası değildir. TLS/TTFB gerçek origin ve ara ağ bileşenlerini ayıramaz. HTTP200 uygulama/GPU model kabulü sayılmaz.
- **Çıkış bölgesi/IP bilinmiyor:** ipinfo tek çağrı TLS EOF/curl35 ile başarısız; başka ağ rotası/sağlayıcıya geçilmedi. Box timezone Europe/Istanbul konum kanıtı değildir.
- ABD/Avrupa planı: mevcut yetkili US East/West ve Frankfurt/Amsterdam runner/probe konumlarında aynı tek GPU bölgesine aynı TLS/health+izinli stream profili; runner bölgesi belgelensin. Yeni GPU bölgesi açılmaz. Tek box ile farklı URL kullanmak ABD/Avrupa kullanıcı ölçümü olmaz. Egress byte+tarife/billing gerekir; ağ maliyeti henüz NOT TESTED.

## §11 hazırlığı ve sınırlar
- saas/PROJECT_PROMPT_TR.md: tam çok dosyalı görev ve auth/RBAC/tenant izolasyonu/CRUD/PostgreSQL/migration/frontend/backend/Docker/CI/dokümantasyon kabul kriterleri.
- Önerilen sınırlar (onay bekliyor):2 model turu (üretim+1düzeltme),30dk aktif üretim/doğrulama,100dosya, compute bütçesi$3; model input/output token üst sınırı koşudan önce ayrıca sabitlenmeli. Süre yetmezse dur; otomatik uzatma yok.
- verify_project.py --mock sabit mini fixture üzerinde gerçek compileall+unittest çalıştırdı. Bu PostgreSQL/auth/gerçek SaaS kabulü değildir. --project modunda yalnız komut planı; gerçek model çıktısı onaylı disposable sandbox içinde build/run/test edilmeli. EXE.9/.10 sürüm uyumu netleşmeli.

## GPU penceresi — KAN.US onayıyla güncel durum (6 Kasım2026 09:39 TSİ talebi)
- **ONAYLI:** yalnız https://api-test.kanus.org; korunmuş SSD/sıcak GPU; compute tavanı$3; aktif test en çok30dk; otomatik uzatma yok. Ayrı cold-start koşusu yok. Bu onay doğrultusunda önceki90dk cold önerisi iptal.
- **Gerçek koşu YAPILMADI.** Bu tur login ölçümleri yalnız box/localhost; GPU start/stop/status veya dış model çağrısı yok.
- Idle kararı15dk. Planlama için en çok60dk koordinasyon penceresi (mevcut sıcak GPU+30dk aktif+15dk idle+kapanış/pay) yeterli hedef; bu otomatik süre/bütçe uzatımı değildir. reports/KANUS_GPU_REPORT.md satır60–64 varsayımı$1.84/sa ile30dk aktif~$0.92;60dk tüm pencere üst tahmini~$1.84 compute. Güncel fiyat/billing doğrulanmadı; egress/vergi/korunan disk bekleme hariç; $3 toplam fatura tavanı olarak yorumlanmaz.
- Token kaynağı yalnız /workspace/secrets/perf_api_token.txt → PERF_API_TOKEN; değer hiçbir çıktı/log/rapora konmaz. perf/gpu_runner.py --preflight dosyayı güvenli açar, env'e bağlayıp Authorization Request nesnesini ağsız hazırlar; bu sürüm gerçek yürütmeyi bilerek uygulamaz.
- Loader sentetik600-mode fixture ile PASS; eksik dosya için açık hata selftest PASS. Gerçek belirtilen dosya **yok**: preflight exit2, `PERF token file missing: /workspace/secrets/perf_api_token.txt`. Hiç ağ isteği yok. Token dosyası sağlanmadan gerçek koşu hazır değil.

## Gerekenler / açık sorular
- İzinli GPU test endpointi: normal sözleşme POST https://api-test.kanus.org/v1/chat/completions, ancak **bu koşuda yalnız health izni kullanıldı**; gerçek/yük koşusu için KAN.US zaman penceresi ve kapsam teyidi gerekir. Hiçbir model çağrısı şu an plan dışı çalıştırılmaz.
- Ayrılmış test hesabı/lisans → `PERF_API_TOKEN` (harness adı); kredi status/history ve ledger gözlem yetkisi. US_MODEL_UPSTREAM/US_MODEL_UPSTREAM_TOKEN sunucu ayarlarıdır, müşteri auth değildir. US_LICENSE_SECRET/US_SESSION_SECRET/US_GPU_WAKE_TOKEN_FILE değerleri alınmaz/loglanmaz.
- Gerçek SLO/hata bütçesi, örnek sayısı, steady-state/soak süresi, model token bütçesi, hourly kredi kapsamı, GPU utilization+fatura telemetrisi ve EXE sürümü belirlenmeli.
- Strict upstream[DONE] politikası ve429/504 Retry-After değişikliği US sahibiyle kararlaştırılmalı. Ürün kaynaklarında başka ajanların eşzamanlı değişiklikleri nedeniyle diff bağlamı yeniden doğrulanmalı.

## GPU koordinasyonu özeti
KAN.US onayı kaydedildi: endpoint yalnız api-test.kanus.org, sıcak/korunmuş GPU, compute≤$3, aktif≤30dk, idle15dk, otomatik uzatma ve ayrı cold-start yok. Bu tur gerçek koşu yapılmadı; token dosyası eksik. Loader ağsız fixture/selftest ile doğrulandı. Gerçek yürütme açık talimat ve token/telemetri hazır olduğunda ayrıca ele alınacak.

## Yerel login performansı — .10 snapshot ve round11

Başlangıç: 2026-11-06T09:40:55.065495+03:00 (TSİ). Yalnız box/localhost POST /account/login; gerçek kullanıcı verisi, test VPS/prod veya GPU kullanılmadı. perf/login_bench.py tekrar üretilebilir araç; JSON/MD results/login_results.*.

### CPU, sürüm, yöntem
- taskset -c0,1 ile hook başlatıldı; GOMAXPROCS=2. Her iki sürümde başlangıç ve bitiş /proc/<pid>/status Cpus_allowed_list=0-1, sched_getaffinity=[0,1]; tüm Go thread affinity kümeleri0,1. CPU limit doğrulaması PASS. Bu exclusive CPU rezervasyonu değil; diğer ajanlar/paylaşımlı box scheduler jitter yaratabilir.
- .10 aday baseline: önceki PERF koşusundan kalan binary aynen korundu; SHA256 75e92824da67814b8cda3b93e7dc7898244618ee519355cbb1412281bef54e40. Önceki ölçüm hash eşleşmesi PASS; exact commit/release etiketi ayrıca doğrulanmış değil. Eski kaynağı yeniden yazarak karşılaştırma yapılmadı.
- Round11 kaynakta mevcut: US_BCRYPT_CONCURRENCY default2, US_BCRYPT_WAIT_SECONDS default10,503+Retry-After5. Ayrı offline build; source stability PASS. Binary SHA256 dab2348e8f51893a2c61d1888e14ca4a0dea8cb63ec54226a428c7eedaefccfd; accounts.go SHA256 48db78b904f2b49da823f30fc40751c714755a2330def1b0f736366aac6274b4. Tam Go-tree hash JSON’da.
- **bcrypt cost12**, her sürümde geçici fixture DB’deki4 hesabın hash metadata’sından doğrulandı; hash/parola/session loglanmadı. Fixture hesaplar /account/register ile yaratıldı; ledger silindi, SMTP ve GPU wake kapalı.
- Her seviyede4 başarılı warmup hariç en az50 ölçülen istek:1→50,5→50,10→50,20→60. Ardışık eşzamanlı burst; steady-state/open-loop testi değil. Yeni mantıksal TEST-NET X-Real-IP her istekte yalnız güvenilen loopback proxy yolunda kullanıldı; bcrypt testi IP10/15dk rate-limitinden ayrıldı. Ağda başka IP’ye paket gönderilmedi.
- Throughput toplam ve başarılı istek/s ayrı. Tüm yanıt p50/p95 hızlı429 redleri içerir; başarılı-login yüzdelikleri ayrıca verilmiştir. Request latency worker içindeki HTTP isteği başlangıcından cevabın okunmasına kadar; batch throughput dispatch maliyetini de içerir.

### Sonuç tablosu (ms)
| Sürüm | Kullanıcı | n | Tüm p50/p95 ms | Başarılı p50/p95 ms | Hata % | HTTP | Tüm RPS / başarılı RPS |
|---|---:|---:|---|---|---:|---|---|
| baseline_dot10_prior_binary | 1 | 50 | 210.0/220.2 | 210.0/220.2 | 0.0 | {'200': 50} | 4.71/4.71 |
| baseline_dot10_prior_binary | 5 | 50 | 407.8/469.1 | 414.3/473.1 | 20.0 | {'200': 40, '429': 10} | 11.30/9.04 |
| baseline_dot10_prior_binary | 10 | 50 | 66.6/511.4 | 427.0/525.9 | 60.0 | {'429': 30, '200': 20} | 22.12/8.85 |
| baseline_dot10_prior_binary | 20 | 60 | 46.1/422.6 | 420.0/445.0 | 80.0 | {'429': 48, '200': 12} | 45.97/9.19 |
| round11_default2_wait10s | 1 | 50 | 210.3/230.8 | 210.3/230.8 | 0.0 | {'200': 50} | 4.68/4.68 |
| round11_default2_wait10s | 5 | 50 | 507.0/1045.7 | 507.0/1045.7 | 0.0 | {'200': 50} | 5.28/5.28 |
| round11_default2_wait10s | 10 | 50 | 640.7/1087.4 | 640.7/1087.4 | 0.0 | {'200': 50} | 9.26/9.26 |
| round11_default2_wait10s | 20 | 60 | 1225.3/1321.8 | 1225.3/1321.8 | 0.0 | {'200': 60} | 9.21/9.21 |
| round11_overload_wait1s_DIAGNOSTIC_NOT_DEFAULT | 20 | 60 | 1034.5/1076.5 | 943.0/1077.1 | 16.7 | {'503': 10, '200': 50} | 10.35/8.63 |

### Kısa yorum ve round11 davranış kontrolleri
- Tek giriş ~210ms: cost12 bcrypt CPU işi ana gecikme adayı; profil alınmadığından kesin CPU fonksiyon dağılımı iddiası yok. İkiCPU’da yüksek eşzamanlı başarılı throughput yaklaşık9/s civarında. Kuyruk daha çok işi tamamlıyor, tek istek compute hızını arttırmıyor.
- Legacy20 eşzamanlıda48/60 HTTP429 (%80 hata),12 başarı. Tüm-yanıt p50~46ms yanıltıcı biçimde düşük; başarılı-login p50~420ms. Round11 varsayılan20’de60/60 HTTP200, p50~1225ms/p95~1322ms; red yerine kontrollü bekleme.
- Round11 default20’de503 oranı0%; bu profilde Retry-After dalı **NOT_EXERCISED_NO_503**. Default profillerde en yüksek response~1402ms<10sn; gözlenen kuyruk için toplam yanıttan üst sınır kurulabilir, kesin kuyruk milisaniyesi dışarı sunulmadığından N/A. Dolu kuyrukta varsayılan10sn timeout dalı normal ölçümle doğrulanmış sayılmaz.
- Ayrı **US_BCRYPT_WAIT_SECONDS=1** tanısal profili (default10sn değil):20eşzamanlı/60istekte10×503+50×200,503 oranı%16.7;503’lerin%100’ünde Retry-After5;503response max~1065ms (1sn timer+scheduler/HTTPpayı). Round11 geri basınç dalı HTTP üzerinden gerçekten görüldü.

### Tekrar çalıştırma ve sınırlamalar
```bash
python3 /workspace/us-panel/perf/login_bench.py --cpus 0,1 --samples 50 --overload-probe
```
Yeni round geldiğinde ayrı binary hash ve aynı workload ile tekrar karşılaştırılabilir. baseline bin/hook-login-baseline korunmuştur; --baseline ile açıkça bu dosya seçilebilir. Güncel kaynak başka ajanlarca değiştirilebileceğinden hash kontrolü zorunlu. Ölçüm bir kez yapıldı; istatistiksel tekrar/izole makine olmadan küçük performans farklarını regresyon ilan etmeyin.

### Token ve gerçek GPU koşusu durumu
perf/gpu_runner.py --selftest sentetik token yükleme ve missing-file hata kontrolü PASS. --preflight gerçek belirtilen dosyanın olmadığını exit2 ile açıkça bildirdi. Token değeri, hash’i veya credential ortamı yazılmadı.0 dış model/GPU isteği; gerçek koşu yapılmadı. Onaylı$3/30dk/sıcakSSD/no-extension/endpoint-only sınırları GPU bölümünde güncellendi.

Ek kontrol: round11 TestBcryptThirdLoginWaitsThen503 ikiCPU’da PASS (exit0); semafor iki slotu test içinden doldurularak503/Retry-After5 ve slot boşalınca başarılı login doğrulandı. Testin override bekleme değeri1sn; default10sn zaman aşımı kabulü değildir. results/login_semaphore_unit.json kaynak hashlerini ve sonucu içerir.
