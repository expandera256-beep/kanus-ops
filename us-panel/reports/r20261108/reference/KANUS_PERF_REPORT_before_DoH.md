# KANUS PERF — r20261108 / §7 + §11

**UZAK KOŞU DURDURULDU — DNS_MISMATCH: box api-test.kanus.org→198.18.0.1 döndürdü; gerekli82.29.184.225 değil. FAZ1 ön koşulda durdu; FAZ2 başlatılmadı. Servis sağlığı/bakiye ölçülmedi.0HTTPS isteği.**

Güncelleme: 2026-11-08T00:41:27.877090+03:00 (TSİ / Europe/Istanbul).

**Genel sonuç: yerel/simülasyon kontrolleri PASS; gerçek model/SaaS/uzak kabul BLOCKED. Üretime hazır sonucu verilmez.**

## Başlangıç, kapsam ve güvenlik
- İstenen yeni prompt bütünüyle okundu: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt; SHA256 3c71bea9f9bc9325d370bd667cc2706e0aec0ecfadae74c8d51db0f2e0963663. PERF kapsamı§7 ve§11. Gerçek kaynak/dağıtım işleri bu scratch ölçüm turunun dışında.
- Kaynak release.go sürümü **2026.11.06.11**. Yerel unmodified Go hook offline derlendi. Kaynak/binary hashleri results/mock_results.json ve login_results.json içinde; model hook source_stable_during_build=True. .11deploy henüz bu tur uzaktan doğrulanmadı.
- İlk yerel alt-turda uzaktrafik0. Sonradan doğrudan onaylı faz girişiminde yalnız api-test.kanus.org için1DNS çözümü yapıldı; beklenenIP çıkmadığından HTTP/TLS/yük/login başlamadı.31.220.63.191 ve82.29.184.225 IPlerine istek/SSH/ping yok; dış IP/bölge sorgusu yok.
- GPU start/stop/status, gerçek model, cloud kiralama veya ücretli kaynak işlemi yok. SMTP kapalı; sentetik geçici ledger/kullanıcılar. Sır/parola/hash/session raporlanmadı; runtime hook logları saklanmadı.

## Promptun tam istediği / önceki prompttan farklar
- **Yeni§7, eski§11:** KANUS SaaS Project Manager; backend API, frontend, PostgreSQL, kullanıcı girişi, RBAC, proje ve görev yönetimi, **dosya yükleme, yönetim paneli**, API dokümantasyonu, Docker, otomatik testler, kurulum kılavuzu. Gerçek uygulama çalışmalı; build hataları ajanla düzeltilmeli; sonunda çalışan uygulama+test kayıtları+kaynak bulunmalı. Yeni bölümde sayısal token/süre/dosya/tur bütçesi yok.
- **Yeni§11, eski§9:**1/5/10/25/50 kullanıcı; p50/p95/p99, **başarı oranı, GPU belleği, kuyruk süresi, hata, kredi tutarlılığı, eşzamanlı görev izolasyonu, toplam maliyet**. Sayısal gecikme/başarı/hata/GPU belleği SLO eşiği yok; ücret doğuran gerçek yük için onay şart.
- Yeni vurgu: eskiGPU kullanımı→GPU belleği; eskiistekbaşımaliyet→toplam maliyet; açık başarı ve görev izolasyonu. Eski proje bölümünde CI/CD açıkça vardı; yenide açık liste maddesi değil, öneri olarak korunabilir. Bölüm numaraları değişti; eski§8 coğrafi ölçüm bu turun bağımsız zorunlu ölçümü değil.
- Prompt§15 rapor formatı: başlangıç durumu, değişen dosyalar, çözüm, gerçek testler, PASS/FAIL/BLOCKED, kalan eksikler, gereken onaylar. Bu rapor ve kabul ekleri bu yapıyı kullanır.
- **Güncel idle:** policy v5,600sn/10dk; Demir8Kasım2026 00:20TSİ kararı. Önceki15dk kararı artık geçersiz; aktif senaryo/kabul dokümanları600sn. Eski ölçüm dosyaları tarihsel kanıt olarak korunur. Canlı GPU ayarı değiştirilmedi. Eski2tur/100dosya/$3 önerileri yeni§7 eşiği olarak gösterilmedi.

## Uygulanan scratch düzeltmeler / değişen dosyalar
- perf/run_suite.py cleanenv artık **US_GPU_WAKE_DISABLE=1**; güncel.11 unset değerlerde güvenli default waker kurduğundan eski fixture ayarı model configuration503 veriyordu. İlk başarısız kurulum koşusu results/initial_fixture_config_failure/ altında saklandı; GPU-disabled fixture ile doğru koşu tekrarlandı. Bu ürün arızası veya gerçek model başarısı diye gösterilmedi.
- perf/mock_stack.py: yalnız yerel fixture marker echo; perf/login_bench.py: yeni queue header yakalama, tarihsel baseline manifesti. Ürün kaynaklarına yama/değişiklik uygulanmadı.
- Yeni scratch araçları: r20261108_local.py, r20261108_loads.py, isolation_credit_bench.py, remote_bench.py, remote_selftest.py, saas/PROJECT_PROMPT_20261107_TR.md, saas/mock_project_contract.py. Dated sonuçlar, kaynak fixture ve remote_82_config.json bu dizinde.

## §11 — MOCK upstream + gerçek yerel.11 hook
Kontrollü yapılandırma:capacity3/queue64/request5sn; varsayılan prod kapasitesi iddiası değildir.4ardışık batch; mock TTFT20ms,4içerik parçası,150parça/sn. Gerçek token/GPU throughput değil. Küçük örnekte p99 kapasite/SLO kanıtı sayılmaz.

| Kullanıcı | n | p50/p95/p99 ms | TTFT p50 ms | Başarı | Beklenen/gerçek kredi | İzolasyon |
|---:|---:|---|---:|---:|---|---|
| 1 | 4 | 46.88/65.58/68.22 | 22.81 | 100% | 4/4 | PASS |
| 5 | 20 | 50.75/98.51/99.77 | 25.09 | 100% | 20/20 | PASS |
| 10 | 40 | 97.24/190.73/203.73 | 72.45 | 100% | 40/40 | PASS |
| 25 | 100 | 237.86/431.04/456.79 | 214.47 | 100% | 100/100 | PASS |
| 50 | 200 | 477.12/856.50/910.86 | 450.99 | 100% | 200/200 | PASS |

- Ana yük364/364 tam yanıt; kredi364/364; hold0. Beşbağımsız sentetik lisans/benzersiz yanıt markerı ile1/5/10/25/50 düzeylerinde ek91istek: tüm yanıtlar doğru isteğe ait, lisansbaşına düşüm eşleşiyor, hold0. Bu dosya/tenant çalışma alanı izolasyonunun tamamını kanıtlamaz.
- GPU belleği **N/A/MOCK**; gerçek toplam fatura maliyeti **NOT MEASURED**,0diye uydurulmadı. Varsayımsal$0.002/krediyle yalnız ana364isteğin simülasyon toplamı$0.728. Yerel CPU maliyeti fiyatlandırılmadı. GPU API çağrısı0.
- Go model route kuyruk bekleme header’ını dışarı sunmuyor: N/A; Python fixture ölçülen queue dağılımları mock_results.* içinde. Login route yeni header sağlıyor ve aşağıda gerçekten ölçüldü.
- Kapasite1/queue2/deadline1sn, upstream TTFT1.5sn,50istek:47×429+2×504+1×HTTP200 SSEtimeout; tam başarı0, kredi0, hold0.429/504 Retry-After1 doğrulandı; HTTP200 tek başına başarı değil.
- **Güncel.11 strict-DONE:** normal1düşüm; truncate/missingDONE/malformed/SSEerror/HTTPerror/clientcancel0düşüm+hold0. Önceki missingDONE politika açığı bu yerel kaynakta artık gözlenmedi; product_policy_failures boş. Yeni ürün kodunu PERF uygulamadı; mevcut kaynağı test etti.

## Login tekrar ölçümü — yerel.11, iki CPU
- taskset -c0,1 + GOMAXPROCS2; /proc PID ve tümthread Cpus_allowed_list=0-1 başlangıç/bitiş PASS. Cost12 fixture hash metadata’sından;4sentetik hesap.4ısınma hariç her seviyede≥50istek. Benzersiz mantıksal TEST-NET X-Real-IP ile yerel IP rate-limitten bcrypt izole edildi; uzak script bu başlığı taklit etmez.
- Paylaşımlı box/exclusiveCPU değil. Tarihsel.10 binary de yeniden ölçüldü; tüm sonuçlar login_results.*. Farklı gün/source/box yükü nedeniyle tek koşudaki süre farkları tek başına regresyon kanıtı değil.

| Kullanıcı | n | p50/p95/p99 ms | Başarı | Başarılı RPS | Queue p95 ms |
|---:|---:|---|---:|---:|---:|
| 1 | 50 | 226.67/280.94/292.77 | 100% | 4.28 | 0.0 |
| 5 | 50 | 612.04/1195.38/1358.81 | 100% | 5.05 | 794.6 |
| 10 | 50 | 658.54/1103.10/1208.58 | 100% | 8.97 | 861.0 |
| 20 | 60 | 1266.71/2188.82/2398.95 | 100% | 8.00 | 1965.15 |

- Default2bcrypt/10sn beklemede20eşzamanlı503oranı0. Kuyruk gözlenen max2089ms<10sn. Ayrı1sn bekleme tanısı:60istekte{'200': 54, '503': 6};503’lerin tamamı Retry-After5. Bu tanı varsayılan10sn timeout testi değildir.

## Kredi endpoint yükü — model/GPU kullanmayan gerçek yerel read-only ölçüm
GET /v1/credits/status; tüm seviyelerde beklenen/gerçek düşüm0,hold0. Bu modelkullanım maliyeti testi değildir; gelecekteki uzak kredi senaryosuyla aynı route’tur.

| Kullanıcı | n | p50/p95/p99 ms | RPS | Başarı |
|---:|---:|---|---:|---:|
| 1 | 50 | 0.241/0.427/0.596 | 3391.74 | 100% |
| 5 | 50 | 0.58/2.189/2.199 | 4976.64 | 100% |
| 10 | 50 | 1.111/1.605/1.664 | 6511.10 | 100% |
| 20 | 60 | 1.738/2.471/2.486 | 8210.44 | 100% |

## §7 — hazırlık / mock kabul
- S7_PROJECT_PROMPT.md yeni ürünün bütün zorunlu özellikleriyle hazırlanmış ajan görevi; gerçek modele gönderilmedi.
- saas_mock_fixture/ altında gerçekten yazılmış backend.py, test_backend.py, frontend/index.html ve README. Sabit in-memory mock sözleşme; modelin ürettiği gerçek SaaS değildir.
- compileall exit0;8unittest PASS: auth lookup,project/task,RBAC admin,tenant red,upload/download,path traversal,boyut sınırı,upload cross-tenant. Bunlar doğrudan mockservice fonksiyonları; gerçekHTTPbackend/UI/PostgreSQL/Docker kabulü değil.
- **BLOCKED:** gerçek US Panel ajan üretimi, gerçek çalışan frontend/backend+PostgreSQL, Docker build/run, tarayıcı kabulü, ajan hata→düzelt→yeniden test zinciri. Gerekçe:.11deploy teyidi+GPU penceresi bekleniyor. Başarılı mock sonuçları bunların yerine geçirilmedi.

## Uzak fazlar — 8 Kasım00:31 onayı ve güvenli durdurma

- **Onay:** Demir8Kasım2026 00:31TSİ uzak koşuyu doğrudan onayladı. Sonraki açık görevlendirmeye göre FAZ1 şimdi başlatıldı; FAZ2 yalnız FAZ1temiz+health200 ise. Önceki “yalnız hazırlık/çalıştırma” sınırı bu dar izinli fazlar için güncellendi; GPU/model/register/webhook hâlâ yasak.
- **Hedef:** yalnız https://api-test.kanus.org:443; TLS SNI ve Host api-test.kanus.org. Normal DNS; sabitIP/--resolve/-k yok. Her DNS çözümü beklenen82.29.184.225 adresini vermeli; farklı veya ek adres varsa daha socket kurulmadan STOP. IP31.220.63.191’e asla bağlantı yok.
- **HOST beyanı:** hook binarysha2567d72f449…56f0; kaynak reports/HOOK_MOVE_PROD_VPS_20261106.md§15 ve son HOST bildirimi. Bu bağımsız hash ölçümü değil. §15satır389–401, taşıma ve Restart=no bilgisini içerir. Marker yokluğu artık durma koşulu değil; kimlik kanıtı HOSTbeyanı olarak yazılır.
- **Koşu ön koşulları:** HOST taşıma beyanı+boxDNS82doğrulaması; private sentetik lisans token’ı; FAZ2için /workspace/secrets/perf_login_account.json exactmode600, email/password/synthetic:true. Dosya içeriği hiçbir kayda girmez. Bu gerçek girişimde DNS kapısı geçilemediğinden token ve QAhesap dosyası **okunmadı/kullanılmadı**.
- **Sınırlar:** global600sn; FAZ1hard300sn, FAZ2kalanbütçe ve en çok300sn. Max20istek;1→5→10→20. Her seviyeden önce healthz200 ve bakiye; aralarda≥3sn. Sağlık hatası/connectionrefused/transporthatası hemen STOP; arkaarkaya502/503/504 veya beklenmeyenhata>%50 STOP.429beklenen ve ayrı oran; başarılı login yüzdelikleri yalnız200yanıtlardan. Login/register/authspoof/fiyat/deploy/hesap silme FAZ1de yok; register/webhook/model iki fazda da yok.
- **Restart=no/paylaşımlı prod kaynak riski:** hook çökmesi kendiliğinden iyileşmez; api.kanus.org ileCPU/RAM paylaşımı HOSTbeyanı. Başarısız health sonrası bir üstseviyeye geçilmez; yeniden başlatma/onarım yapılmaz.

### Gerçek girişim sonucu

Zaman: 2026-11-08T00:36:59.343712+03:00 (TSİ). DNS sonucu **198.18.0.1**, beklenen82.29.184.225; tekDNSçözümü. DNSölçüm süresi2.134ms; toplamharcananuzaksüre(bu girişimde yalnızDNSönkontrolü) **0.002178sn**. HTTPS/TCP/TLS yükisteği0; bütçekalanı599.997822sn. DNSfallback/publicresolver sorgusu/IPpinleme veya tekrar yok.

| Faz | Servis sağlığı | Harcanan süre | Sonuç |
|---|---|---:|---|
| FAZ1 health5örnek+kredi | NOT TESTED; HTTPbaşlamadı | 0.002178sn (DNSonly) | STOPPED DNS_MISMATCH |
| FAZ2 login | NOT TESTED |0sn | BLOCKED_PHASE1_NOT_CLEAN |

| Faz | Kullanıcı | p50/p95/p99 | Başarı | HTTP dağılımı |429oranı |5xxoranı |
|---|---:|---|---|---|---|---|
| FAZ1 kredi | 1 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ1 kredi | 5 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ1 kredi | 10 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ1 kredi | 20 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ2 login | 1 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ2 login | 5 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ2 login | 10 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |
| FAZ2 login | 20 | NOT RUN | NOT MEASURED |0istek | NOT MEASURED | NOT MEASURED |

Bakiye önce/sonra **NOT MEASURED / NOT MEASURED**;0diye yazılmadı, eşitlik kabulü verilmedi. Sunucu sağlık durumuna PASS/FAIL atanamaz çünkü onaHTTPgönderilmedi. Durdurma nedeni servis arızası değil, box çözümlemesinin zorunlu hedefi doğrulayamamasıdır. HOSTun taşıma beyanını geçersiz ilan etmez.

### Salt okunurluk kaynak kanıtı
- cmd/us-license-hook/main.go:83 → GET/v1/credits/status CreditHandler;:109 healthz yalnızok döner.
- licensesrv/credits.go:23 yalnızGETstatus;:95–96read sweepletmez/workeruyandırmaz;:118–147CreditStatus mu+Licenseokuması veçıktı oluşturur, credits/reserved/settlement üzerinde yazma yok.
- Kullanılan lisans token’us1prefixi gerektirir; modelgateway,creditsreserve/settle,accountredeem,register,webhook kodallowlistdışında. İmzalı webhook gönderilmez; licenses.json canlı test verisi değiştirilmez.

### QA rate-limit kaynakları ve yorum sınırı
- reports/qa-20261106/ratelimit-evidence.txt:10,20–24,43–64: nginx/login ve/register IPbaşına **5r/m,burst2,nodelay**; gerçekpeerIPanahtarı; X-Real-IP overwrite.
- reports/qa-20261106/accounts-security.md:140–160: bunlar tarihsel HOSTnginx kanıtı, yeni canlıconfig doğrulaması değil. Go güncel kaynakta normalizeemailbaşına **10başarısızdeneme/900sn**, başarıylareset; eskiGoIP10/15dk limiti ayrı. Bu kaynak canlıbinary7d72f449…ile birebir aynı kabul edilmedi.
- AynıQA:148box çıkışı140.248.50.x değiştiği için tekIPbucketı dolmayabiliyor; HOSTsonbildirimi140.248.50.x ile52.x arasında dönme. Bu tur egresslookup yapılmadı.429görülmemesi limiteryok veya spoofbypass kanıtı olmaz; login tekhesapaccountlimiterinden de etkilenebilir.

### Yerel yeniden doğrulama
- Domain/SNI/Host normalHTTPS/noIPpin;25/50,21concurrency,601sn, tüketim yolları/register/webhook reddi;429ayrıexpected;600izinlihesap okuma/644ret/missingclearerror; mockDNS82accept/198ve31reject; yerelpeak20 ve0.25sSIGALRMexit124 PASS. GerçekDNSbu testlerde yapılmadı.
- Sağlık503→fazların durması, bakiyedüşmesi→loginbaşlamaması, yalnıztemizFAZ1sonrasıFAZ2geçişi yerel kontrolakışı simülasyonunda PASS. Gerçek token/QAokuma ve gerçek ağ0.
- Idlepolicyv5 logicalclock7/7PASS;599snoidle/600sidle,active/queuedstopengelleri. Bu canlıGPUkapama kabulü değil.
- Kanıt: results/remote_safety_tests.*,remote_phase_unit.*,idle_policy_v5.*,remote_phases.*. Önceki yerelmockölçüm tabloları aynen korunur.

### Komutlar
Ağsızplan ve yereltest:
```bash
python3 /workspace/us-panel/perf/remote_bench.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --dry-run
python3 /workspace/us-panel/perf/remote_selftest.py
python3 /workspace/us-panel/perf/remote_phase_unit.py
```
Onaylı girişimde kullanılan komut (DNSuyuşmazlığıyla durdu; **otomatik tekrar yapılmadı**):
```bash
python3 /workspace/us-panel/perf/remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases
```
DNSönkoşulu giderilmeden tekrar/--resolve/pin ile geçiş yapılmamalı. QAhesap dosyası hazır HOSTbeyanı olarak kaydedildi; DNSkapısı yüzünden varlığı/içeriği bu görevde okunmadı.

## Güncel kabul durumu
Yerel/simülasyon PASS. GerçekFAZ1 STOPPED DNS_MISMATCH;FAZ2 BLOCKED. Gerçek§7model/SaaS,GPUbelleği/fatura ve.11deploykabulü BLOCKED. Uzak25/50izin20nedeniyleBLOCKED. Önceki “10/15dkkararıaçık” sorusu kapanmıştır:8Kasım00:20TSİpolicyv5=600sn.

Son yerel doğrulama: remote_safety_tests 22/22 PASS; remote_phase_unit7/7PASS;idlepolicyv5 7/7PASS;compileallPASS. Gerçek uzak girişim yalnız1DNS/0HTTPS olarak kaldı; yeniden deneme yapılmadı.
