# KANUS PERF — r20261108 / §7 + §11

**GERÇEK UZAK FAZ1 ve FAZ2 TEMİZ TAMAMLANDI: ikiDoH82.29.184.225, TLS/healthPASS; bakiye10→10; toplam60.236sn;5xx0. Login429beklenen/ayrıraporlandı. Model/GPUçağrısı yok.**

Güncelleme: 2026-11-08T00:48:10.524623+03:00 (TSİ / Europe/Istanbul).

**Genel sonuç: yerel/simülasyon kontrolleri PASS; gerçek model/SaaS kabul BLOCKED; darizinli uzak fazlar CLEAN. Üretime hazır sonucu verilmez.**

## Başlangıç, kapsam ve güvenlik
- İstenen yeni prompt bütünüyle okundu: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt; SHA256 3c71bea9f9bc9325d370bd667cc2706e0aec0ecfadae74c8d51db0f2e0963663. PERF kapsamı§7 ve§11. Gerçek kaynak/dağıtım işleri bu scratch ölçüm turunun dışında.
- Kaynak release.go sürümü **2026.11.06.11**. Yerel unmodified Go hook offline derlendi. Kaynak/binary hashleri results/mock_results.json ve login_results.json içinde; model hook source_stable_during_build=True. .11deploy henüz bu tur uzaktan doğrulanmadı.
- İlk yerel alt-turda uzaktrafik0. İlk yerelDNSkapısı duruşu arşivlendi. Son kullanıcıtalimatıyla çiftDoHönkontrolü ardından domainüzerinden154izinliHTTPS+2DoH yapıldı; aşağıda GERÇEK etiketli.31.220.63.191’e istek yok;IPpin/SSH/ping yok. Dışegress/bölge sorgusu yok.
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

## Uzak FAZ1 + FAZ2 — GERÇEK ölçüm / çift DoH sonrası

Başlangıç: **2026-11-08T00:44:25.575626+03:00 (TSİ / Europe/Istanbul)**. Sonuç **COMPLETED**; durdurma: **yok**.

- **Onay:** Demir8Kasım2026 00:31TSİ doğrudan onayı ve son açık çiftDoH+iki faz talimatı. Sentetik KAN.US token ve sentetik login hesabı kullanıldı; başka QAhesabı kullanılmadı. Hesap dosyası exactmode600, symlink olmayan normal dosya olarak programatik okundu. İçerikler/sessionlar kaydedilmedi; hesap silinmedi.
- **Hedef:** yalnız https://api-test.kanus.org:443; SNI+Host api-test.kanus.org; normal hostname bağlantısı. SabitIP/--resolve/-k yok. İki sağlayıcı da doğrulanmadan hedefHTTPS/credential okuma yok.
- **DoH:** dns.google ve cloudflare-dns.com ayrı ayrı yalnız **82.29.184.225**, TTL**300sn**, DNSstatus0/HTTP200 verdi. Koşunun60.24sn toplamı TTL içinde kaldı. Yerel çözücü **198.18.0.1** gösteriyor; kullanıcının açıkladığı box ağproxy sahteIPsi (RFC2544). Yerel cevabı originIP diye yorumlamıyoruz; orijinal DNS_MISMATCH duruşu tarihsel olarak results/remote_phases_local_dns_stop.* altında korunuyor.
- **TLS:**5/5 health ölçümünde varsayılanCA zinciri+hostname doğrulaması PASS; SAN api-test.kanus.org, TLS1.3. Proxy üzerinden TCPconnect **gerçek uçtan uca originRTT değildir**; TLS/TTFB de proxy yolunu içerir. TLS aşaması yerel SSLcontext/CA hazırlığını da kapsar; safhandshake süresi iddiası yok.
- **HOST beyanı:** hook binarysha2567d72f449…56f0; reports/HOOK_MOVE_PROD_VPS_20261106.md§15. Bu bağımsızbinaryhash/deployedsource ölçümü değildir; marker zorunlu tutulmadı. Restart=no ve prod api.kanus.org ileCPU/RAM paylaşımı nedeniyle kademeli1→5→10→20; her seviyeden öncehealth200, aralarda≥3sn.
- **Sınırlar:** max20 eşzamanlı; global600sn, FAZ1ençok300sn, FAZ2kalanbütçe/ençok300sn. Sağlık/transport hatasında hemenSTOP; ardışık502/503/504 veya beklenmeyenhata>%50STOP;429beklenen, ayrıoranda. Model/chat/GPU/register/webhook/redeem/reserve/settle çağrısı yok;31.220.63.191’e istek yok.

| Faz | Servis sağlığı | Harcanan süre sn | Sonuç |
|---|---|---:|---|
| phase1 | HTTP200 | 38.812 | CLEAN |
| phase2 | HTTP200 | 21.422 | CLEAN |

Bu girişim toplamı **60.233954sn**. Önceki DNS-only duruş dahil muhasebeleştirilen toplam **60.236132sn**, kalan **539.763868sn**. Hedefe **154HTTPS**; ayrıca ikiDoH sorgusu (toplam156HTTPS). Ek uzaktan tekrar/health ölçümü yapılmadı.

| Faz | Kullanıcı | n | Başarılı p50/p95/p99 ms | Başarı | HTTP dağılımı |429 |5xx | RPS toplam/başarılı |
|---|---:|---:|---|---:|---|---:|---:|---|
| phase1 | 1 | 20 | 437.754/488.575/509.286 | 100% | {"200": 20} | 0% | 0% | 2.21/2.21 |
| phase1 | 5 | 20 | 443.053/464.048/499.239 | 100% | {"200": 20} | 0% | 0% | 10.59/10.59 |
| phase1 | 10 | 20 | 441.420/447.621/448.032 | 100% | {"200": 20} | 0% | 0% | 21.98/21.98 |
| phase1 | 20 | 20 | 466.848/488.474/488.494 | 100% | {"200": 20} | 0% | 0% | 40.61/40.61 |
| phase2 | 1 | 5 | 695.783/720.947/723.542 | 100% | {"200": 5} | 0% | 0% | 1.43/1.43 |
| phase2 | 5 | 5 | 904.467/911.285/911.799 | 80% | {"429": 1, "200": 4} | 20% | 0% | 5.47/4.38 |
| phase2 | 10 | 10 | 898.912/906.934/907.186 | 40% | {"429": 6, "200": 4} | 60% | 0% | 10.96/4.38 |
| phase2 | 20 | 20 | 944.785/948.311/948.446 | 20% | {"429": 16, "200": 4} | 80% | 0% | 21.05/4.21 |

**Bakiye önce/sonra: 10 → 10 kredi; düşüm0.** Her seviyeden önce/sonra kontrol edildi; herhangi bir farkta koşu duracaktı. Saltokunur yükün80/80yanıtında remaining10/reserved0. FAZ2sonrası da10. Hiçbirfazda5xx/transport/healthhatası yok; durdurma koşulu oluşmadı.
- FAZ1:80/80HTTP200,429%0,5xx%0. FAZ2:40istek →17HTTP200+23HTTP429; toplam başarı%42.5, beklenen429%57.5,5xx%0; kabul edilen200+429%100. Login429yanıtlarında Retry-After görünmedi. Hangi limiterin ürettiği body/log tutulmadığından kesinleştirilemez.
- **Örnek sınırlılığı:** kredi her seviyede20istek; login5/5/10/20istek, başarılı örnek5/4/4/4. Bu kısa ve bütçeli güvenlik koşusu; uzakbcryptkapasite/SLO/p99 istatistik gücü kanıtı değil. Yerel≥50örnek benchmark şartı önceki yerel koşuda karşılandı; uzaklogin için aynı iddia yok. Queueheader uzak yanıtlarda yok: N/A,0ms değil.

### Healthz 5 gerçek zamanlama örneği
Python monotonic+socket+defaultTLS ölçümü; curl verisi değildir. DNS yerel proxy çözüm süresi; TCP DNS hariç connect; TLS bağlam hazırlığı+handshake; TTFB istek başlangıcından ilkbyte’a kümülatif. TTFT/model ölçülmedi.

| Örnek | Yerel DNS ms | ProxyTCPconnect ms | TLS aşaması ms | KümülatifTTFB ms | HTTP/TLS |
|---:|---:|---:|---:|---:|---|
| 1 | 1.426 | 0.413 | 421.613 | 564.151 | 200/verified |
| 2 | 0.877 | 0.127 | 292.967 | 435.363 | 200/verified |
| 3 | 1.422 | 0.159 | 337.678 | 480.274 | 200/verified |
| 4 | 1.316 | 0.250 | 295.712 | 439.072 | 200/verified |
| 5 | 1.749 | 0.450 | 336.244 | 478.604 | 200/verified |

p50: DNS1.422ms; proxyTCP0.250ms; TLS336.244ms; TTFB478.604ms. Box çıkış bölgesi bu tur bağımsız doğrulanmadı; ABD/Avrupa ölçümü yapıldı denmez. ABD veABayrıyetkili worker/VM üzerinde aynı protokol ve küçük bütçeyle ölçülmeli; bölge/egress/TLSproxy kayıt edilmeli.

### Yerel / uzak / eski testVPS karşılaştırması

| Senaryo | Kullanıcı | Yerel p50/p95/p99 ms | Uzak başarılı p50/p95/p99 ms | Eski testVPS |
|---|---:|---|---|---|
| credit_status | 1 | 0.241/0.427/0.596 | 437.754/488.575/509.286 | N/A |
| credit_status | 5 | 0.580/2.189/2.199 | 443.053/464.048/499.239 | N/A |
| credit_status | 10 | 1.111/1.605/1.664 | 441.420/447.621/448.032 | N/A |
| credit_status | 20 | 1.738/2.471/2.486 | 466.848/488.474/488.494 | N/A |
| login | 1 | 226.671/280.938/292.766 | 695.783/720.947/723.542 | N/A |
| login | 5 | 612.044/1195.376/1358.811 | 904.467/911.285/911.799 | N/A |
| login | 10 | 658.540/1103.099/1208.577 | 898.912/906.934/907.186 | N/A |
| login | 20 | 1266.715/2188.818/2398.955 | 944.785/948.311/948.446 | N/A |

Yerel.11 login taskset0,1/GOMAXPROCS2/cost12 ve≥50örnek; uzak donanım/CPUlimiti bağımsız doğrulanmadı ve loginlerin büyük bölümü429ile elendi. Bu yüzden uzaktaki yüksekseviye başarılılogin sürelerinin düşük görünmesi hızlanma/regresyon kararı değildir. Kredi farkı tekbaşına sunucuCPU maliyeti sayılamaz; proxy+DNS+TLS+WAN kapsar. Eski testVPS için karşılaştırılabilir login/status yük verisi yok; N/A, ölçüm uydurulmadı.

### Salt okunurluk kaynak kanıtı
- work2/kiyas/term/cmd/us-license-hook/main.go:83→GET/v1/credits/status;:109healthz yalnızok. licensesrv/credits.go:23GETstatus,:95–96read sweepletmez/workeruyandırmaz,:118–147CreditStatus kilit altındaokur; kredi/hold/settlement yazmaz. Login ayrı izinli accountsession yaratır, kredi/model işlemi değildir; gelen session içerikleri atılır.
- Kodallowlist yalnızGET/healthz,GET/v1/credits/status vePOST/account/login; tüketim/register/webhook yolları kodda reddedilir. licenses.json canlıtestverisine yazma yapılmadı.

### QA rate-limit kaynakları ve yorum sınırı
- reports/qa-20261106/ratelimit-evidence.txt:10,20–24,43–64: nginx/account/login ve/account/register IPbaşına5r/m,burst2,nodelay; gerçekpeerIPanahtarı;X-Real-IP overwrite.
- reports/qa-20261106/accounts-security.md:140–160: tarihselHOSTnginxkanıtı, bu koşuda canlıconfig okunmadı. GüncelGo kaynağı normalizeemailbaşına10başarısızdeneme/900sn ve başarıylareset; eskiGoIPlimiti ayrı. Deploybinary ile kaynağın birebirliği bağımsız doğrulanmadı.
- Box egress140.248.50.x ile52.x arasında dönebiliyor (HOST/QA). Her isteğin gerçek çıkışIPsi ölçülmedi;429orani tekIPkovası limiti veya bypass kanıtı değil. UzakisteklerdeX-Real-IP taklit edilmedi.

### Yerel yeniden doğrulama / kanıtlar
- remote_safety_tests **23/23PASS**: çiftDoHuyumu/mismatch simülasyonu;198proxyyanıtı yalnıztanı;max20/600sn; yasakroute;mode600/missing/644;domain/SNI/Host;429ayrımı;localhostpeak20;0.25sn hardcutoff. Gerçekuzakağ/secretdosyaokuma0.
- remote_phase_unit9/9PASS: paylaşılan600sn tükeninceDNSöncesidurma ve kalanbütçeyleherikifaztimerkesmesi; sağlık503STOP, bakiyeazalmaSTOP, yalnıztemizFAZ1sonrasıFAZ2,429beklenen,>%50/ardışık5xxdurma. idlepolicyv57/7PASS;600sn/10dk (Demir8Kasım00:20TSİ).
- Hamgerçek: results/remote_phases_execution_raw.json,remote_phases.json/.md. Karşılaştırma: results/remote_comparison.json. Yerel/mock veri kendi dosyalarında ayrı tutuldu. ÖncekiDNSduruşu remote_phases_local_dns_stop.*; bu gözlem yanlışsilinmedi.

### Komutlar / tekrar uyarısı
Ağsızkontrol:
```bash
python3 /workspace/us-panel/perf/remote_bench.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --dry-run
python3 /workspace/us-panel/perf/remote_selftest.py
python3 /workspace/us-panel/perf/remote_phase_unit.py
```
Gerçekte bir kez kullanılan sır içermeyen komut (**koşu tamamlandı; otomatiktekrar yok**):
```bash
python3 /workspace/us-panel/perf/remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases
```
Config previous_remote_seconds muhasebesi güncellendi; aynı600sn bütçesi tekrar sıfırlanmaz. Raporu ağsız yeniden üretme: python3 /workspace/us-panel/perf/finalize_remote_report.py.

## Güncel kabul durumu
Yerel/simülasyonPASS; darizinli gerçekFAZ1/FAZ2CLEAN. Bu tümSaaS/üretimkabulü değildir. Gerçek§7model/SaaS,GPUbelleği/fatura ve.11deployedsource kabulüBLOCKED/bağımsızdoğrulanmadı. Uzak25/50max20izni nedeniyleBLOCKED. GPUaçılmadı/modelçağrılmadı. Idlepolicyv5=600sn;15dkgeçersiz. Açık: gerçekmodel/SaaS içinayrıdeploy/GPUplanı; gerekiyorsaizindahilindeABD/ABvantage ve dahaistatistiklikoşu.
