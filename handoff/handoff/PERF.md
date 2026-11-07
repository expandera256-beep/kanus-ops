# PERF — Devir Dosyası (2026-11-08 01:45 TSİ)

## 1) Kimlik
- **Ad:** PERF
- **Başlık:** (boş)
- **Açıklama:** KAN.US altında performans ve gerçek proje testi botu. İletişim Türkçe. Görevi /workspace/us-panel/prompt/GLOBAL_PROMPT_20261106.txt §8, §9, §11: tek bölgeden ABD/Avrupa erişim gecikmesi ve ilk token süresi ölçümü; 1/5/10/25/50 kullanıcı yük testleri (p50/p95/p99, hata oranı, kuyruk, kredi tutarlılığı, istek başı maliyet); US Panel ajanına çok dosyalı gerçek SaaS projesi ürettirip derletme ve test etme. Önce mock/yerel harness yazar; gerçek GPU koşusu için KAN.US'tan zaman penceresi ister, GPU'yu kendisi açmaz. Test sunucusu (31.220.63.191) başka iş yükleriyle paylaşımlı ve yüklü; oraya yük testi yapmaz. Prod sunucuya (82.29.184.225) dokunmaz. Yalnız gerçek ölçüm raporlar. Sırları yazdırmaz.
- **Rol:** KAN.US'a bağlı performans/yük testi ve gerçek SaaS proje testi botu. Güncel geçerli prompt: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt (Demir harfiyen uygulanmasını istedi; KAN.US aktardı). Bu prompt altında PERF'in bölümleri: §7 (KANUS SaaS Project Manager gerçek proje testi) ve §11 (1/5/10/25/50 kullanıcı yük testi). İlk prompt: GLOBAL_PROMPT_20261106.txt §8, §9, §11.

### Sistem talimatı taslağı
> Sen PERF'sin; KAN.US'a bağlı performans ve gerçek proje testi botusun. Demir'le Türkçe konuş. Görevin KANUS_PROMPT_20261107.txt §7 ve §11: (a) mock OpenAI-uyumlu SSE upstream + yerel kredi hook'u ile önce yerel harness; (b) 1/5/10/25/50 kullanıcı yük testi: p50/p95/p99, hata oranı, kuyruk, kredi tutarlılığı, istek başı maliyet; (c) US Panel ajanına çok dosyalı gerçek SaaS projesi ürettirip derlet ve test et; (d) tek bölgeden ABD/AB gecikme ve TTFT ölçümü.
> Kurallar: GPU'yu asla kendin açma; gerçek GPU koşusu için KAN.US'tan pencere iste ve GPU kullanmadan önce Demir'in doğrudan onayını al (KAN.US'un onayı yetmez). 31.220.63.191'e yük testi yapma (paylaşımlı). 82.29.184.225 (prod + api-test aynı VPS) yalnız Demir'in doğrudan onayladığı sınırlarla: en çok 20 eşzamanlı, en çok 10 dk, yalnız sentetik hesap, kredi sorguları salt-okunur. Uzak koşularda /v1/chat veya model trafiği gönderme (GPU uyandırır), register çağrısı ve imzalı lisans webhook'u yok. Sentetik hesabı silme (Demir'in kararı). Yalnız gerçek ölçüm raporla; uydurma sayı yok. Sırları asla yazdırma. Kutunun yerel DNS'i 198.18.0.1 (proxy fake-IP) döner; gerçek çözümü DoH (dns.google, cloudflare-dns.com) ile doğrula. GPU boşta kapanma: 10 dk (600 s, policy v5).

## 2) Hafıza
Profile: kayıtlı olgu yok. Paylaşılan hafıza: kayıtlı olgu yok.
Journal:
- (2026-11-08 00:31) Demir, PERF sohbetinde 82.29.184.225 (api-test.kanus.org) üzerinde yük testini doğrudan onayladı: en çok 20 eşzamanlı, en çok 10 dk, yalnız sentetik hesap, kredi kontrolleri salt-okunur; HOST taşımayı teyit ettikten ve sentetik hesap dosyası hazır olduktan sonra başla.
- (2026-11-08) KAN.US bildirdi: Demir GPU idle shutdown'ı 00:20'de 10 dakikaya geri aldı (policy v5, 600 s); api-test.kanus.org artık 82.29.184.225'e çözülüyor; uzak test sınırları: 20 eşzamanlı, 10 dk, sentetik hesap, salt-okunur kredi sorguları.
Bağlam (sohbetten):
- 2026-11-06: Demir idle shutdown'ı 15 dk yapmıştı (sonra 10 dk'ya döndü).
- KAN.US GPU penceresi onayı (yalnız KAN.US'tan, Demir'den doğrudan değil): $3 compute tavanı, 30 dk aktif test, otomatik uzatma yok, KAN'ın cold-start testinden hemen sonra sıcak GPU; .11 deploy'u bekleniyor.
- Prod api.kanus.org ve api-test aynı VPS (82.29.184.225, hostname srv2133566) üzerinde CPU/RAM paylaşıyor; hook servisi Restart=no.
- GPU ds-a100 ≈ $1.84/saat (GPU raporundan).
- Limitler: nginx limit_req IP başına 5r/m burst=2 (/account/login, /account/register); Go tarafı normalize e-posta başına 900 s'de 10 deneme. Kutunun çıkış IP'si döner (140.248.50.x ve 52.x).

## 3) Rutinler
Yok.

## 4) Beceriler, bağlayıcılar, servisler
- Beceriler: canvas (kullanılmadı), unblock-research, routines (kullanılmadı).
- Bağlayıcılar: yok (Hostinger/Shopify/GitHub kullanılmadı).
- Servisler: kutu Shell (python3, Go derleme, taskset), DoH (dns.google, cloudflare-dns.com), api-test.kanus.org (82.29.184.225), SendToAgent ile KAN.US, HOST, QA, US, PANEL DEV.

## 5) Açık işler ve kalınan yer
- **Tamamlanan:** Mock §11 (.11 üzerinde) 1/5/10/25/50 kullanıcı p95 = 66/99/191/431/857 ms, 364/364 başarılı, kredi her seviyede tutarlı, 5 lisans izolasyonu geçti, [DONE] eksikse kredi düşülmedi. Yerel login p95 1/5/10/20 = 281/1195/1103/2189 ms (round 11).
- **Uzak koşu (2026-11-08 00:44–00:45 TSİ):** toplam 60,2 s trafik; bakiye 10 → 10; healthz hep 200. Kredi okuma p95 1/5/10/20 = 489/464/448/488 ms (80/80). Başarılı login p95 = 721/911/907/948 ms; login başarı oranı %100/80/40/20, geri kalanı beklenen 429; toplam 17/40, 5xx yok. 429 yanıtlarında Retry-After başlığı yok (yeni bulgu; KAN.US'a bildirildi, US'a iletilmesi opsiyonel). Eski test VPS'iyle karşılaştırılabilir veri yok. HOST sürüm işareti doğrulanamadı (raporda "NOT VERIFIED").
- **Açık:** §7 gerçek SaaS proje testi ve gerçek chat/GPU yükü — BLOCKED: .11 deploy'u (PANEL DEV PASS sonrası) ve KAN.US/KAN'ın GPU penceresi duyurusu bekleniyor. Prompt hazır: reports/r20261108/S7_PROJECT_PROMPT.md. Mock build + 8 sözleşme testi geçti.
- **Açık:** 82.29.184.225'te 25/50 kullanıcı seviyeleri 20 eşzamanlılık tavanı yüzünden BLOCKED (yalnız yerelde koşuldu).
- Uygulanmamış öneriler: perf/proposals/01-backpressure-observability.diff, 02-strict-upstream-done-policy.diff.
- Kontrollü uzak koşu komutu (otomatik yeniden koşma!): `python3 /workspace/us-panel/perf/remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases`

## 6) Bekleyen Demir kararları ve onaylar
- GPU kullanımı (§7 gerçek koşu, $3 / 30 dk penceresi) için Demir'in doğrudan onayı alınmadı; yalnız KAN.US onayı var.
- 82.29.184.225'te 20 üstü eşzamanlılık (25/50) için yeni onay gerekir.
- Sentetik hesapların (sipariş 9000000000505 lisansı ve KAN.US'un 00:33'te açtığı login hesabı) silinmesi Demir'in kararı.
- 82.29.184.225 onayı tek koşu içindi (00:31); yeni uzak koşu için yeniden onay alınmalı.

## 7) Kritik dosya yolları ve sha256
- Promptlar: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt, /workspace/us-panel/prompt/GLOBAL_PROMPT_20261106.txt
- Harness: /workspace/us-panel/perf/ (mock_stack.py, load.py, latency.py, run_suite.py, login_bench.py, gpu_runner.py, idle_contract.py, cli_smoke.py, remote_bench.py, remote_phases.py, …; saas/, results/, bin/, proposals/)
- Raporlar: /workspace/us-panel/reports/KANUS_PERF_REPORT.md (1. tur), /workspace/us-panel/reports/r20261108/KANUS_PERF_REPORT.md (ana), REMOTE_PHASES_REPORT.md, remote_82_config.json, results/remote_phases*.json, results/remote_comparison.json
- HOST taşıma notları: /workspace/us-panel/reports/HOOK_MOVE_PROD_VPS_20261106.md §15 (hook binary sha256 7d72f449…56f0; .11 sonrası 12772d12…b47; ledger /opt/us-license-hook-test/licenses.json; log: journalctl -u us-license-hook-test)
- QA rate-limit kanıtı: reports/qa-20261106/ratelimit-evidence.txt, accounts-security.md

```
3c71bea9f9bc9325d370bd667cc2706e0aec0ecfadae74c8d51db0f2e0963663  prompt/KANUS_PROMPT_20261107.txt
e1cef8aee3d78119ff48e9084db4ec7611b7447ebb7def58b938f94f1970cf6f  prompt/GLOBAL_PROMPT_20261106.txt
fb633b7c2c9faf16ffeb8a3370e78d10e382995c585adeb39f6a6935f1bcc658  reports/KANUS_PERF_REPORT.md
95abb4ee04f70d365275239e76fd58d56344425973a7b7211c3c3e306e19b888  reports/r20261108/KANUS_PERF_REPORT.md
6b8ba8b3bcda2963d04da4136a22e092179ae3d0a28e4f9dd8a040e93cc0cd0b  perf/cli_smoke.py
89d5d3e0d4a3191b67d63d5f14603c3a01c440a772df869235d5656f1fb9f7cd  perf/finalize_remote_report.py
0efc139813300d2af6ad160a233e7ba6597b551e48023a976b191dfc8b9728db  perf/gpu_runner.py
286f480d5fe0f946f7fb6c26730c098a292a055f9e6f68948d769d1f4bb167c4  perf/idle_contract.py
c31464a800680ab26881d9b35e33484aa153189ca13bdd7715d018502ce9dc51  perf/isolation_credit_bench.py
4f8283b9cab7b60d90c30936a75254a8cb2948fce3cd76cbbbfd6e98eb99a8e5  perf/latency.py
fae585fd6b6c48b8bc5efa7ce59dc0e53cebdb8cd0ac9c7ad671081fa6f7503b  perf/load.py
79b09e10f005a18b6751ddd121c2b479300fb00f31fa9858d842f1fab2aeed9a  perf/login_bench.py
7c680a71d2e73a1b7f7d3e8fcb8da0278ed96f874fc7ee8a3ee3024e82bc0ce7  perf/mock_stack.py
34ef97836407d2a1d705c480cd005c6f1f80f8cf99a40c5adfa846e9ea6bb29a  perf/r20261108_loads.py
edb1ef80e4b03c91beee039cdeb0d68839341389be5b348cc4f99b2a29548710  perf/r20261108_local.py
538a45f19dd71b51fb934036b17d113c96a2f8e96e102875a066312d09eda245  perf/remote_bench.py
0af7919e16469fb310ff7f9511284ebb4fde40af778ac514824cf6f0bddaa293  perf/remote_phase_unit.py
1bd0030c8afa3f536c3543e6c63f81f8b0f83954ffd7496bf61a89224884c2d4  perf/remote_phases.py
0677209adf158690122c310296c70fea30a9fc90522ae824c84a23a409dd9df2  perf/remote_selftest.py
427db082bf114b16e682fa09417a27bd4b570f3612f7428ddfa9ed2490ec36e9  perf/run_suite.py
4af6197ffd8064db85a84af73d97fb08eca58372951e7295d69bff610d0b15f3  remote_82_config.json
72cd0dc48dc40652c6b322c89ad4bcd7580e9ed671b8cfc52c478f08e7b5c38c  S7_PROJECT_PROMPT.md
e1ec249af91ea49942dbab19bf7559841f2787a836b0db3dc423251211730aa6  REMOTE_PHASES_REPORT.md
800f5d2e32b728da6b2f723552f5269d3bc4e8b44d3e7e712af6a48dea9b1b19  EVIDENCE_SHA256.json
4c477c06fd7a7bf4ec5456fede77827daecfaf9133adc0138db3cc8b16a14b68  /workspace/us-panel/perf/proposals/01-backpressure-observability.diff
6e65f159d755fb8fe0b5578938d20b85f9a509c0aa6fe70f1d196ba55ad40891  /workspace/us-panel/perf/proposals/02-strict-upstream-done-policy.diff
87c38dae2b36f82df700c9cdc9cb94285068d04e413c36c02d6bf11eb61351a0  /workspace/us-panel/perf/README.md
```
(Son blok yolları reports/r20261108/ dizinine görelidir.)

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)
- **PERF_API_TOKEN** — /workspace/secrets/perf_api_token.txt (mod 600): sentetik lisans (sipariş 9000000000505, 10 kredi); remote_phases.py / remote_bench.py kredi bakiye/durum sorgularında kullanılır.
- **Sentetik login hesabı (e-posta + parola)** — /workspace/secrets/perf_login_account.json (mod 600): KAN.US'un 00:33'te açtığı, lisanssız hesap; login_bench / remote_phases login fazında kullanılır.
Not: kutu tüm botlarca paylaşıldığından aynı kutu kullanılırsa dosyalar yerinde kalır; yeni kutuda yeniden oluşturulmalı. Değerleri asla yazdırma.
