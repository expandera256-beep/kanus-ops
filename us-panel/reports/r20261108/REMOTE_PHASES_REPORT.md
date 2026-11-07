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
