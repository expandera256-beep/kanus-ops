# §12 Zorunlu Test Matrisi — 06.11.2026

Hazırlanma: 06.11.2026 08:06 TSİ (Europe/Istanbul).
Bu ilk teslim mevcut kanıt envanteridir; bu worker tarafından yeniden koşulmuş test sonucu değildir. PASS yalnız belirtilen yerel/mock kapsam içindir. Kanıtsız gerçek ortam kabulü NOT TESTED olarak bırakıldı.

| §12 testi | Durum | Mevcut kanıt ve gerçek alıntı | Sınır / eksik |
|---|---|---|---|
| go test ./... | PASS | reports/go_test_topyekun_20261106_v4.txt:561: `ok kiyas 9.991s`; :571 hook 0.010s; :812 licensesrv 8.236s | Gerçek model testleri :414/:417 SKIP |
| go test -race ./... | PASS | reports/go_race_topyekun_20261106_v3.txt:1: `ok kiyas 33.192s`; :2 hook 1.027s; :4 licensesrv 6.366s | Yerel; canlı yük kabulü değil |
| go vet ./... | PASS | reports/KANUS_TEST_RESULTS.md:9: `go vet ./...: exit 0`; ham: reports/go_vet_topyekun_20261106_v2.txt; exit: reports/TEST_EXIT_STATUS_20261106_v2.json | Raporlanmış önceki koşum |
| Derleme testleri | PASS | reports/KANUS_TEST_RESULTS.md:12: `Windows amd64 GUI cross-build; PE header/machine/version ...` | Native Windows çalıştırması değil |
| API entegrasyon testleri | PASS | reports/python_topyekun_20261106_v2.txt:11: `test_webhook_delivery_account_and_metered_model ... ok`; reports/KANUS_TEST_RESULTS.md:126: healthz 200 / imzasız webhook 401 | Yerel fixture + önceki uzak rapor; panel tam akış değil |
| PostgreSQL migration testleri | NOT TESTED | reports/KANUS_TEST_RESULTS.md:34–73: real PostgreSQL contract opt-in atlandı | Gerçek geçici PostgreSQL + migration ileri/geri/tekrar kanıtı yok |
| Shopify webhook testleri | PASS | reports/KANUS_TEST_RESULTS.md:85–87: aynı/yeni id HTTP200, tek lisans; kötü HMAC401 | Önceki kanus-test raporu; yeni sipariş/ödeme yok |
| Kredi yarış durumu testleri | PASS | reports/LOAD_TEST_20261106.txt:225–226: `http=map[200:1 402:49]`; `negative_balance=0 double_spend=0` | Yerel sentetik lisanslar |
| SSE streaming testleri | PASS | reports/LOAD_TEST_20261106.txt:171–172: `http=map[200:33 429:17]`; `unexpected_errors=0` | Yerel fake SSE; gerçek GPU değil |
| GPU yaşam döngüsü testleri | PASS | reports/python_topyekun_20261106_v2.txt:36: `test_ten_parallel_controllers_issue_one_cloud_start ... ok`; :25 active lease ... ok | Yalnız fake provider; gerçek start/load/stop NOT TESTED. :21 default idle900; prompt §4 ister600, ayrıca kabul farkı |
| Windows EXE testleri | NOT TESTED | reports/KANUS_TEST_RESULTS.md:130: `Native Windows EXE .5 NOT TESTED`; gap tracker :16 yalnız parent bildirimi | Bu worker için bağımsız native ham kanıt yok; cross-build ayrı |
| Güvenlik taramaları | NOT TESTED | reports/KANUS_SECURITY_REPORT.md:12: `bağımsız security scan/topic-body replay audit açık` | SAST/dependency scan + gerçek güçlü kullanıcı izolasyonu kanıtı yok |
| Çok kullanıcılı yük testleri | PASS | reports/LOAD_TEST_20261106.txt:171–172: `n=50 p50_ms=102.374 p95_ms=366.031`; overflow17 unexpected0 | Mock; §8 gerçek TTFT/GPU/maliyet/kuyruk ölçümleri NOT TESTED |
| Büyük proje üretim testi | NOT TESTED | reports/go_test_topyekun_20261106_v4.txt:417: `--- SKIP: TestRealModelPanelBuildRentalProject`; reports/E2E_HAZIRLIK_TR.md:26: tunnel bekleniyor | §9 proje türü rental değil: çok kullanıcılı proje yönetimi; mock büyük proje kabulü sayılmaz |

Özet (14 satır): PASS 10; FAIL 0; BLOCKED 0; NOT TESTED 4. Bunlar yalnız mevcut kanıt kapsamı; bu tur yeniden test sayısı değildir.

Yeni worker çalışmaları: QA_TEST_RAPORU_20261106.md ve qa-20261106/ ham logları.
