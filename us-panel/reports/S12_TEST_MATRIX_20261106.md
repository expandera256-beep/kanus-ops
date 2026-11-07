# §12 Zorunlu Test Matrisi — 06.11.2026

İlk teslim08:06 TSİ; yeni scratch harness kanıtlarıyla güncelleme08:26 TSİ (Europe/Istanbul).
**14 zorunlu test: PASS10 / FAIL0 / BLOCKED0 / NOT TESTED4.** Kanıt olmayan kabul NOT TESTED bırakıldı; engeller ayrı belirtilir. PASS yalnız açıkça belirtilen yerel/mock kapsamdır, üretim kabulü değildir.

Ham kanıt kökü **L = `/workspace/us-panel/reports/qa-20261106/`**. Önceki raporların envanteri ilk sürümde korundu: L/S12_INITIAL_MATRIX_0806.md.

| §12 zorunlu testi | Durum | Yeni gerçek kanıt (dosya + alıntı) | Ortam / eksik / önceki kanıt |
|---|---|---|---|
| go test ./... | PASS | L/go-test.log:561 `ok kiyas 9.720s`; :573 hook0.685s; :814 licensesrv8.104s; :815 exit0 | Linux;301 üst düzey+37 alt PASS,6 SKIP. Önceki reports/go_test_topyekun_20261106_v4.txt |
| go test -race ./... | PASS | L/go-race.log:1–5 `ok kiyas 31.849s`, hook8.967s,licensesrv65.833s; exit_code=0 | Yerel; önceki go_race_topyekun_20261106_v3.txt |
| go vet ./... | PASS | L/go-vet.log:1 `exit_code=0` | Önceki go_vet_topyekun_20261106_v2.txt/TEST_EXIT_STATUS_20261106_v2.json |
| Derleme testleri | PASS | L/build.log:3 `linux_build_exit=0`; :5 `windows_cross_build_exit=0`; L/e2e9-panel-driver.log:3 `BUILD py_compile PASS` | Linux hook,Windows amd64 cross-build ve model-fixture üretimi Python proje; native Windows değil |
| API entegrasyon testleri | PASS | L/sec8-local-harness.log:1–10 register201/login200, cross-user redeem409/settle409/owner-release200; L/e2e9-independent-acceptance.log:1 `INDEPENDENT_HTTP_ACCEPTANCE PASS health/auth/create/list/cross-user/validation` | Gerçek localhost HTTP hook/üretilmiş servis. Uzak health200. Önceki python_topyekun_20261106_v2.txt:11 |
| PostgreSQL migration testleri | NOT TESTED | L/python-test.log:66–105 `skipped opt-in real PostgreSQL contract` | Geçici DB/DSN ve migration ileri/geri/tekrar kabulü yok; SKIP PASS sayılmaz |
| Shopify webhook testleri | PASS | L/go-test.log:605 `TestRejectsBadHMAC PASS`; :607 paid idempotent PASS; :788 reversalHMAC PASS. L/sec8-remote-final.log:5–6 missing/wrongHMAC401 | Yerel sentetik webhook; uzakta yalnız yanlış/eksik imza, gerçek secret yok. Önceki KANUS_TEST_RESULTS.md:85–87 |
| Kredi yarış durumu testleri | PASS | L/go-test.log:717–718 `http=map[200:1 402:49]`, `negative_balance=0 double_spend=0`; L/sec8-local-harness.log:17–21 `credit_consistency:true` | Yerel gerçek ledger/HTTP + uzak gerçek QA A/B: L/sec8-remote-credit.log:5–11 ownership409, A duplicate commit10→9, B release10→10,9 reserve200/1 reserve402; tüm holdlar release200. Önceki LOAD_TEST_20261106.txt:225–226 |
| SSE streaming testleri | PASS | L/go-test.log:730 `TestModelProxySSEChargeAndNoLicenseLeak PASS`; :739–746 truncated/error/length/timeout/cancel PASS; yeni hook yükü1/5/10/25/50 PASS | Mock upstream; gerçek GPU değil. Önceki LOAD_TEST_20261106.txt:171–172 |
| GPU yaşam döngüsü testleri | PASS | L/python-test.log:25 active lease ... ok; :36 `test_ten_parallel_controllers_issue_one_cloud_start ... ok`; :21–44 fake provider PASS | Yalnız mock provider. Gerçek start/load/idle-stop NOT TESTED; tünel/lifecycle operatör kabulü gerekli. Önceki python_topyekun_20261106_v2.txt |
| Windows EXE testleri | NOT TESTED | Bu workerın native yürütme kanıtı yok; L/build.log:5 yalnız cross-build0 | Windows makinesi gerekli. Parent/native bildirimleri bu workerın bağımsız ham kabulü sayılmadı |
| Güvenlik taramaları | NOT TESTED | Bağımsız dependency/SAST taraması yok. Hedefli güvenlik regressionı L/sec8-panel-boundaries.log:2/5 PASS; L/sec8-local-harness.log:15 gerçek account limiter429 | Harness/negatif auth testleri tam güvenlik taraması değildir; govulncheck kurulu değil. Önceki KANUS_SECURITY_REPORT.md:11–12 sandbox/audit açık |
| Çok kullanıcılı yük testleri | PASS | L/sec8-local-harness.log:17–21:1/5/10/25/50 PASS;50deHTTP20014/42936; ttft_p50_ms31.317,p95_total_ms42.762; unexpected_error_rate0.0 | Yerel gerçek hook + mock SSE; TTL/TTFT/toplam ölçümleri mock kapsam. Gerçek GPU/model/kuyruk/maliyet kabulü yok; uzak12 probe concurrency2,429 gözlenmedi |
| Büyük proje üretim testi | NOT TESTED | L/go-test.log:417 `SKIP TestRealModelPanelBuildRentalProject`. Yeni küçük proje E2E: L/e2e9-panel-driver.log:52 `E2E PANEL ... PASS`; bağımsız kabul PASS | Küçük mock REST projesi büyük §9 frontend/PostgreSQL/Docker/CI-CD platformu değildir. Gerçek model çağrısı yapılmadı |

## Yeni §9 küçük mock E2E (büyük proje satırının yerine geçmez)
Panel task queue → OpenAI model proxy → review → atomic apply → py_compile →4 gerçek loopback HTTP unittest → kasıtlı health hatası → gerçek hata çıktısının modele geri verilmesi → read_file → SHA256 patch_file → değişmeyen4 test PASS → panel project-tests PASS → modele verilmeyen bağımsız HTTP kabulü PASS. Mock upstream3 çağrı. L/e2e9-panel-driver.log:4–32 ilk failure, :34–51 repair sonrası PASS; L/e2e9-independent-acceptance.log:1 PASS.

## Komutlar
```bash
python3 /workspace/us-panel/qa/sec8/run.py
python3 /workspace/us-panel/qa/sec8/run.py --remote-only
python3 /workspace/us-panel/qa/e2e9/run.py
```
Go/Python tam komutlar, ortam, ilk fixture hatası/düzeltme/yeniden test ve kalan engeller: `QA_TEST_RAPORU_20261106.md`; yeniden çalıştırma: `qa/README.md`.

## Uzak authenticated güncelleme — lisans engeli çözüldü
İki private QA lisansı artık mevcut ve yalnız runtime okundu. A↔B gerçek reservation settle409; A tek commit ve duplicate settle sonrası10→9, B reserve/release sonrası10→10. A kalan kotasına10 request/concurrency2 yarışta9 HTTP200/1 HTTP402;9 geçici holdun tamamı release200; son reserved A0/B0. **Kullanım: A1 kredi, B0.** Kanıt L/sec8-remote-credit.log:1–11. A sipariş9000000000301, B9000000000302; anahtarlar hiçbir log/rapora yazılmadı (credential-leak-check.json PASS).

Geçerli A bearer ile traversal/command query HTTP200 ve scoped alanlar değişmedi (L/sec8-remote-authenticated-final.log:12–13). Bu yalnız status yüzeyi; uzak file/terminal endpointi bulunmadığından tam OS/dosya izolasyonu kabulü değildir. Önceki missing-license BLOCKED kaldırıldı.

Uzak429 ucuz testle gözlenmedi: status12/concurrency2→401; account/login12/concurrency2→404; ayar/servis değişikliği/yük artırımı yapılmadı. Yerel account limiter429 kanıtı ayrı PASS.

## Revoke sonrası canlı A/B kabulü — PASS

`revoke-qa-licenses-20261106.json` dosyasının varlığı ve JSON yapısı doğrulandı, kimlik/order/status alanları tutarlı; hash `e5644c731e3b1f479d8109d879c0910403fe08feb0ac6a9560fbbca9f89eeea0`. Dosya operatör bildirimi olarak tutuldu, körlemesine kabul edilmedi; iki private lisansla bağımsız canlı API denetimleri yapıldı.

| Kontrol | A | B | Kanıt |
|---|---|---|---|
| Status sözleşmesi | PASS:200 valid=false/revoked=true | PASS:200 valid=false/revoked=true | L/post-revoke-A-B.log:2/10 |
| /v1/chat ve /v1/chat/completions | PASS:401/401 | PASS:401/401 | :3–4/11–12 |
| /v1/models | PASS:401 | PASS:401 | :5/13 |
| /v1/credits/reserve | PASS:403 | PASS:403 | :6/14 |
| Nonexistent-ID settle/release | PASS:409/409 resource reddi | PASS:409/409 resource reddi | :7/15; mevcut-ID auth kanıtı değildir |
| Önce/sonra kredi hareketi | PASS:0 ilave tüketim | PASS:0 ilave tüketim | :8/16; tüm remaining/credits/available/reserved0 sabit |
| Mevcut settlement-ID replayi | NOT TESTED | NOT TESTED | Önceden bütün holdlar kapatıldı, eski IDler tutulmadı |

Gerçek komut: `python3 /workspace/us-panel/qa/sec8/post_revoke.py --after-revoke --license both`. Alt koşum15 PASS/0 FAIL/0 BLOCKED/2 NOT TESTED; önceki pending B satırı superseded. **Genel güncel harness64 PASS/0 FAIL/0 BLOCKED/7 NOT TESTED.** §12nin14 zorunlu satırı10 PASS/4 NOT TESTED olarak kalır (yerel/mock kapsam sınırlamaları değişmedi).

Test tüketimi A1/B0; revoke kalan A9/B10 hakkını iptal etti, bu ek kullanım değildir. Post-revoke ilave tüketim A0/B0. Model/proxy401 erken auth reddi ve sabit ledger kanıttır; bağımsız upstream çağrı sayacı ölçülmedi. Bu worker revoke/admin/VPS yazma/restart yapmadı. Anahtarlar runtime dışında paylaşılmadı.
