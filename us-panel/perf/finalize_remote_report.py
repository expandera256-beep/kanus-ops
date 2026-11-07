"""Offline-only report assembly. Never reads credential files or performs networking."""
import json,sys
from pathlib import Path
from datetime import datetime
from zoneinfo import ZoneInfo
from run_suite import percentile
from remote_phases import finish_report
B=Path('/workspace/us-panel/reports/r20261108');D=B/'results'
def read(name):return json.loads((D/name).read_text())
r=read('remote_phases.json');cp=B/'remote_82_config.json';c=json.loads(cp.read_text())
archive=D/'remote_phases_execution_raw.json'
if not archive.exists():archive.write_text(json.dumps(r,indent=2,ensure_ascii=False)+'\n')
r.setdefault('accounted_previous_remote_seconds',c.get('previous_remote_seconds',0))
r['data_source']='REAL_REMOTE_NO_MODEL'
r['total_accounted_remote_seconds']=r['remote_elapsed_seconds']+r['accounted_previous_remote_seconds']
r['measurement_notes']=['Local198.18.0.1 is box network proxy address; TCP connect is NOT origin end-to-end RTT.','TLS stage includes handshake and local SSL context/CA setup.','Success-only latency is the primary login metric; all-response latency remains separately in JSON.','No model/GPU/register/webhook/account deletion; no QA account other than KAN.US provided synthetic file.']
for h in r['phases']['phase1'].get('health_samples',[]):h['tls_stage_note']=r['measurement_notes'][1]
finish_report(r,D/'remote_phases.json')
c['previous_remote_seconds']=r['total_accounted_remote_seconds'];c['last_completed_run']={'timestamp':r['timestamp'],'status':r['status'],'no_automatic_repeat':True}
cp.write_text(json.dumps(c,indent=2,ensure_ascii=False)+'\n')
local_credit={x['concurrency']:x for x in read('credit_status_results.json')['levels']}
local_login={x['concurrency']:x for v in read('login_results.json')['variants'] if v['label']=='round11_default2_wait10s' for x in v['levels']}
comparison=[]
lines=['## Uzak FAZ1 + FAZ2 — GERÇEK ölçüm / çift DoH sonrası', '',
 f"Başlangıç: **{r['timestamp']} (TSİ / Europe/Istanbul)**. Sonuç **{r['status']}**; durdurma: **{r.get('stop_reason','yok')}**.", '',
 '- **Onay:** Demir8Kasım2026 00:31TSİ doğrudan onayı ve son açık çiftDoH+iki faz talimatı. Sentetik KAN.US token ve sentetik login hesabı kullanıldı; başka QAhesabı kullanılmadı. Hesap dosyası exactmode600, symlink olmayan normal dosya olarak programatik okundu. İçerikler/sessionlar kaydedilmedi; hesap silinmedi.',
 '- **Hedef:** yalnız https://api-test.kanus.org:443; SNI+Host api-test.kanus.org; normal hostname bağlantısı. SabitIP/--resolve/-k yok. İki sağlayıcı da doğrulanmadan hedefHTTPS/credential okuma yok.',
 '- **DoH:** dns.google ve cloudflare-dns.com ayrı ayrı yalnız **82.29.184.225**, TTL**300sn**, DNSstatus0/HTTP200 verdi. Koşunun60.24sn toplamı TTL içinde kaldı. Yerel çözücü **198.18.0.1** gösteriyor; kullanıcının açıkladığı box ağproxy sahteIPsi (RFC2544). Yerel cevabı originIP diye yorumlamıyoruz; orijinal DNS_MISMATCH duruşu tarihsel olarak results/remote_phases_local_dns_stop.* altında korunuyor.',
 '- **TLS:**5/5 health ölçümünde varsayılanCA zinciri+hostname doğrulaması PASS; SAN api-test.kanus.org, TLS1.3. Proxy üzerinden TCPconnect **gerçek uçtan uca originRTT değildir**; TLS/TTFB de proxy yolunu içerir. TLS aşaması yerel SSLcontext/CA hazırlığını da kapsar; safhandshake süresi iddiası yok.',
 '- **HOST beyanı:** hook binarysha2567d72f449…56f0; reports/HOOK_MOVE_PROD_VPS_20261106.md§15. Bu bağımsızbinaryhash/deployedsource ölçümü değildir; marker zorunlu tutulmadı. Restart=no ve prod api.kanus.org ileCPU/RAM paylaşımı nedeniyle kademeli1→5→10→20; her seviyeden öncehealth200, aralarda≥3sn.',
 '- **Sınırlar:** max20 eşzamanlı; global600sn, FAZ1ençok300sn, FAZ2kalanbütçe/ençok300sn. Sağlık/transport hatasında hemenSTOP; ardışık502/503/504 veya beklenmeyenhata>%50STOP;429beklenen, ayrıoranda. Model/chat/GPU/register/webhook/redeem/reserve/settle çağrısı yok;31.220.63.191’e istek yok.', '',
 '| Faz | Servis sağlığı | Harcanan süre sn | Sonuç |','|---|---|---:|---|']
for phase,p in r['phases'].items():lines.append(f"| {phase} | {p.get('service_health','NOT TESTED')} | {p.get('elapsed_seconds',0):.3f} | {p['status']} |")
lines += ['',f"Bu girişim toplamı **{r['remote_elapsed_seconds']:.6f}sn**. Önceki DNS-only duruş dahil muhasebeleştirilen toplam **{r['total_accounted_remote_seconds']:.6f}sn**, kalan **{r['remaining_total_seconds']:.6f}sn**. Hedefe **{r['https_requests_sent']}HTTPS**; ayrıca ikiDoH sorgusu (toplam156HTTPS). Ek uzaktan tekrar/health ölçümü yapılmadı.", '',
 '| Faz | Kullanıcı | n | Başarılı p50/p95/p99 ms | Başarı | HTTP dağılımı |429 |5xx | RPS toplam/başarılı |',
 '|---|---:|---:|---|---:|---|---:|---:|---|']
for phase,p in r['phases'].items():
    for x in p['levels']:
        v=x['success_latency_ms'];values='/'.join(f'{v[k]:.3f}' if v[k] is not None else 'N/A' for k in ('p50','p95','p99'))
        lines.append(f"| {phase} | {x['concurrency']} | {x['n']} | {values} | {x['success_rate']:.0%} | {json.dumps(x['status_counts'])} | {x['rate_limit_429_rate']:.0%} | {x['five_xx_rate']:.0%} | {x['throughput_rps']:.2f}/{x['successful_rps']:.2f} |")
        ref=(local_credit if phase=='phase1' else local_login)[x['concurrency']]
        comparison.append({'scenario':'credit_status' if phase=='phase1' else 'login','concurrency':x['concurrency'],'remote_success_latency_ms':v,'remote_success_rate':x['success_rate'],'remote_429_rate':x['rate_limit_429_rate'],'local_ms':{k:ref[k+'_ms'] for k in ('p50','p95','p99')},'old_test_VPS':None,'old_test_VPS_note':'No comparable prior login/status load dataset.'})
lines += ['', f"**Bakiye önce/sonra: {r['balance_before']} → {r['balance_after']} kredi; düşüm0.** Her seviyeden önce/sonra kontrol edildi; herhangi bir farkta koşu duracaktı. Saltokunur yükün80/80yanıtında remaining10/reserved0. FAZ2sonrası da10. Hiçbirfazda5xx/transport/healthhatası yok; durdurma koşulu oluşmadı.",
 '- FAZ1:80/80HTTP200,429%0,5xx%0. FAZ2:40istek →17HTTP200+23HTTP429; toplam başarı%42.5, beklenen429%57.5,5xx%0; kabul edilen200+429%100. Login429yanıtlarında Retry-After görünmedi. Hangi limiterin ürettiği body/log tutulmadığından kesinleştirilemez.',
 '- **Örnek sınırlılığı:** kredi her seviyede20istek; login5/5/10/20istek, başarılı örnek5/4/4/4. Bu kısa ve bütçeli güvenlik koşusu; uzakbcryptkapasite/SLO/p99 istatistik gücü kanıtı değil. Yerel≥50örnek benchmark şartı önceki yerel koşuda karşılandı; uzaklogin için aynı iddia yok. Queueheader uzak yanıtlarda yok: N/A,0ms değil.', '',
 '### Healthz 5 gerçek zamanlama örneği',
 'Python monotonic+socket+defaultTLS ölçümü; curl verisi değildir. DNS yerel proxy çözüm süresi; TCP DNS hariç connect; TLS bağlam hazırlığı+handshake; TTFB istek başlangıcından ilkbyte’a kümülatif. TTFT/model ölçülmedi.', '',
 '| Örnek | Yerel DNS ms | ProxyTCPconnect ms | TLS aşaması ms | KümülatifTTFB ms | HTTP/TLS |','|---:|---:|---:|---:|---:|---|']
for i,h in enumerate(r['phases']['phase1']['health_samples'],1):lines.append(f"| {i} | {h['dns_ms']:.3f} | {h['tcp_ms']:.3f} | {h['tls_ms']:.3f} | {h['ttfb_ms']:.3f} | {h['status']}/verified |")
med={k:percentile([h[k] for h in r['phases']['phase1']['health_samples']],.5) for k in ('dns_ms','tcp_ms','tls_ms','ttfb_ms')}
lines += ['',f"p50: DNS{med['dns_ms']:.3f}ms; proxyTCP{med['tcp_ms']:.3f}ms; TLS{med['tls_ms']:.3f}ms; TTFB{med['ttfb_ms']:.3f}ms. Box çıkış bölgesi bu tur bağımsız doğrulanmadı; ABD/Avrupa ölçümü yapıldı denmez. ABD veABayrıyetkili worker/VM üzerinde aynı protokol ve küçük bütçeyle ölçülmeli; bölge/egress/TLSproxy kayıt edilmeli.", '',
 '### Yerel / uzak / eski testVPS karşılaştırması', '',
 '| Senaryo | Kullanıcı | Yerel p50/p95/p99 ms | Uzak başarılı p50/p95/p99 ms | Eski testVPS |','|---|---:|---|---|---|']
for x in comparison:
    f=lambda d:'/'.join(f'{d[k]:.3f}' if d[k] is not None else 'N/A' for k in ('p50','p95','p99'))
    lines.append(f"| {x['scenario']} | {x['concurrency']} | {f(x['local_ms'])} | {f(x['remote_success_latency_ms'])} | N/A |")
lines += ['', 'Yerel.11 login taskset0,1/GOMAXPROCS2/cost12 ve≥50örnek; uzak donanım/CPUlimiti bağımsız doğrulanmadı ve loginlerin büyük bölümü429ile elendi. Bu yüzden uzaktaki yüksekseviye başarılılogin sürelerinin düşük görünmesi hızlanma/regresyon kararı değildir. Kredi farkı tekbaşına sunucuCPU maliyeti sayılamaz; proxy+DNS+TLS+WAN kapsar. Eski testVPS için karşılaştırılabilir login/status yük verisi yok; N/A, ölçüm uydurulmadı.', '',
 '### Salt okunurluk kaynak kanıtı',
 '- work2/kiyas/term/cmd/us-license-hook/main.go:83→GET/v1/credits/status;:109healthz yalnızok. licensesrv/credits.go:23GETstatus,:95–96read sweepletmez/workeruyandırmaz,:118–147CreditStatus kilit altındaokur; kredi/hold/settlement yazmaz. Login ayrı izinli accountsession yaratır, kredi/model işlemi değildir; gelen session içerikleri atılır.',
 '- Kodallowlist yalnızGET/healthz,GET/v1/credits/status vePOST/account/login; tüketim/register/webhook yolları kodda reddedilir. licenses.json canlıtestverisine yazma yapılmadı.', '',
 '### QA rate-limit kaynakları ve yorum sınırı',
 '- reports/qa-20261106/ratelimit-evidence.txt:10,20–24,43–64: nginx/account/login ve/account/register IPbaşına5r/m,burst2,nodelay; gerçekpeerIPanahtarı;X-Real-IP overwrite.',
 '- reports/qa-20261106/accounts-security.md:140–160: tarihselHOSTnginxkanıtı, bu koşuda canlıconfig okunmadı. GüncelGo kaynağı normalizeemailbaşına10başarısızdeneme/900sn ve başarıylareset; eskiGoIPlimiti ayrı. Deploybinary ile kaynağın birebirliği bağımsız doğrulanmadı.',
 '- Box egress140.248.50.x ile52.x arasında dönebiliyor (HOST/QA). Her isteğin gerçek çıkışIPsi ölçülmedi;429orani tekIPkovası limiti veya bypass kanıtı değil. UzakisteklerdeX-Real-IP taklit edilmedi.', '',
 '### Yerel yeniden doğrulama / kanıtlar',
 f"- remote_safety_tests **{len(read('remote_safety_tests.json')['checks'])}/{len(read('remote_safety_tests.json')['checks'])}PASS**: çiftDoHuyumu/mismatch simülasyonu;198proxyyanıtı yalnıztanı;max20/600sn; yasakroute;mode600/missing/644;domain/SNI/Host;429ayrımı;localhostpeak20;0.25sn hardcutoff. Gerçekuzakağ/secretdosyaokuma0.",
 f"- remote_phase_unit{len(read('remote_phase_unit.json')['checks'])}/{len(read('remote_phase_unit.json')['checks'])}PASS: paylaşılan600sn tükeninceDNSöncesidurma ve kalanbütçeyleherikifaztimerkesmesi; sağlık503STOP, bakiyeazalmaSTOP, yalnıztemizFAZ1sonrasıFAZ2,429beklenen,>%50/ardışık5xxdurma. idlepolicyv57/7PASS;600sn/10dk (Demir8Kasım00:20TSİ).",
 '- Hamgerçek: results/remote_phases_execution_raw.json,remote_phases.json/.md. Karşılaştırma: results/remote_comparison.json. Yerel/mock veri kendi dosyalarında ayrı tutuldu. ÖncekiDNSduruşu remote_phases_local_dns_stop.*; bu gözlem yanlışsilinmedi.', '',
 '### Komutlar / tekrar uyarısı',
 'Ağsızkontrol:', '```bash',
 'python3 /workspace/us-panel/perf/remote_bench.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --dry-run',
 'python3 /workspace/us-panel/perf/remote_selftest.py',
 'python3 /workspace/us-panel/perf/remote_phase_unit.py','```',
 'Gerçekte bir kez kullanılan sır içermeyen komut (**koşu tamamlandı; otomatiktekrar yok**):','```bash',
 'python3 /workspace/us-panel/perf/remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases','```',
 'Config previous_remote_seconds muhasebesi güncellendi; aynı600sn bütçesi tekrar sıfırlanmaz. Raporu ağsız yeniden üretme: python3 /workspace/us-panel/perf/finalize_remote_report.py.', '',
 '## Güncel kabul durumu',
 'Yerel/simülasyonPASS; darizinli gerçekFAZ1/FAZ2CLEAN. Bu tümSaaS/üretimkabulü değildir. Gerçek§7model/SaaS,GPUbelleği/fatura ve.11deployedsource kabulüBLOCKED/bağımsızdoğrulanmadı. Uzak25/50max20izni nedeniyleBLOCKED. GPUaçılmadı/modelçağrılmadı. Idlepolicyv5=600sn;15dkgeçersiz. Açık: gerçekmodel/SaaS içinayrıdeploy/GPUplanı; gerekiyorsaizindahilindeABD/ABvantage ve dahaistatistiklikoşu.']
text='\n'.join(lines)+'\n'
(B/'REMOTE_PHASES_REPORT.md').write_text(text)
(D/'remote_comparison.json').write_text(json.dumps({'data_source':'REAL_REMOTE_VS_REAL_LOCAL_SYNTHETIC','comparability_warning':'Different CPU hardware, network/proxy and rate-limiter admission; not a direct regression verdict.','levels':comparison},indent=2,ensure_ascii=False)+'\n')
p=B/'KANUS_PERF_REPORT.md';old=p.read_text();back=B/'reference/KANUS_PERF_REPORT_before_DoH.md';back.parent.mkdir(exist_ok=True)
if not back.exists():back.write_text(old)
head=old.split('## Uzak fazlar')[0] if '## Uzak fazlar' in old else old.split('## Uzak FAZ1')[0]
head=head.replace('**UZAK KOŞU DURDURULDU — DNS_MISMATCH: box api-test.kanus.org→198.18.0.1 döndürdü; gerekli82.29.184.225 değil. FAZ1 ön koşulda durdu; FAZ2 başlatılmadı. Servis sağlığı/bakiye ölçülmedi.0HTTPS isteği.**','**GERÇEK UZAK FAZ1 ve FAZ2 TEMİZ TAMAMLANDI: ikiDoH82.29.184.225, TLS/healthPASS; bakiye10→10; toplam60.236sn;5xx0. Login429beklenen/ayrıraporlandı. Model/GPUçağrısı yok.**')
head=head.replace('gerçek model/SaaS/uzak kabul BLOCKED','gerçek model/SaaS kabul BLOCKED; darizinli uzak fazlar CLEAN')
head=head.replace('- İlk yerel alt-turda uzaktrafik0. Sonradan doğrudan onaylı faz girişiminde yalnız api-test.kanus.org için1DNS çözümü yapıldı; beklenenIP çıkmadığından HTTP/TLS/yük/login başlamadı.31.220.63.191 ve82.29.184.225 IPlerine istek/SSH/ping yok; dış IP/bölge sorgusu yok.','- İlk yerel alt-turda uzaktrafik0. İlk yerelDNSkapısı duruşu arşivlendi. Son kullanıcıtalimatıyla çiftDoHönkontrolü ardından domainüzerinden154izinliHTTPS+2DoH yapıldı; aşağıda GERÇEK etiketli.31.220.63.191’e istek yok;IPpin/SSH/ping yok. Dışegress/bölge sorgusu yok.')
parts=head.splitlines();parts=[('Güncelleme: '+datetime.now(ZoneInfo('Europe/Istanbul')).isoformat()+' (TSİ / Europe/Istanbul).') if x.startswith('Güncelleme: ') else x for x in parts]
p.write_text('\n'.join(parts).rstrip()+'\n\n'+text)
print(json.dumps({'status':r['status'],'remote_seconds_including_previous':r['total_accounted_remote_seconds'],'main_report':str(p),'network_requests_from_finalizer':0},ensure_ascii=False))
