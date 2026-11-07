# KANUS güncel sistem keşfi — §2

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
Kaynak `.11` (2026.11.06.11), pinned `.10`; `.6` yalnız eski referans. api-test /healthz=200 düz `ok`; `/api/version`, `/version`, `/v1/version`=404: çalışan hook sürümü **BLOCKED/doğrulanamadı**. Bu uçların yokluğu hookun .6/.10/.11 olduğunu ispatlamaz; panel sürüm endpointi ile hook sürümünü eşitleme. HOST binary/hash kanıtı gerekli. Kanıt: /workspace/us-panel/reports/r20261108/evidence/versions.json; /workspace/us-panel/reports/r20261108/evidence/hook-health.meta; /workspace/us-panel/reports/r20261108/evidence/hook-api-version.meta; /workspace/us-panel/reports/r20261108/evidence/hook-v1-version.meta.

DNS artık **82.29.184.225**, TTL300 (00:21 TSİ); önceki “eski IP” bilgisi güncel değil. DNS değişimi, store aktarımının / kesinti süresinin / GPU tünelinin kabulü değildir. curl remote_ip=198.18.0.1 yerel proxy ara adresidir, origin IP diye kullanılamaz. Kanıt: /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json; /workspace/us-panel/reports/r20261108/evidence/hook-health.meta.

| Bileşen | Durum | Kanıt / sınır |
|---|---|---|
| US Panel ve Go backend | IMPLEMENTED | Kaynak .11; izole test/vet/race sonuçları aşağıda. /workspace/us-panel/reports/r20261108/evidence/versions.json |
| Windows EXE | PARTIAL | Eski i18n binary: 13 PASS/1 FAIL/1 SKIP; yeni düzeltme henüz native retest değil. /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log |
| Yerel panel sunucusu | IMPLEMENTED | guiMux/API yerel testleri var; gerçek Windows GUI oturumu tam kabul yok. /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log |
| AI ajan yürütme | PARTIAL | Mock generate/apply/build/repair doğrulanmış; gerçek kapsamlı SaaS yok. /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md §9 / restore |
| DeepSeek entegrasyonu | PARTIAL | Kimliksiz models=401; gerçek .11 model çağrısı yapılmadı. /workspace/us-panel/reports/r20261108/evidence/hook-models-no-auth.meta; /workspace/us-panel/reports/KANUS_GPU_REPORT.md |
| GPU yaşam döngüsü | PARTIAL | v4/00004 stop bugı kanıtlı; son 00005/v5 etiketi aynı image, gerçek giderim doğrulanmadı. /workspace/us-panel/reports/r20261108/evidence/watchdog-final.json; /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md §00004; /workspace/us-panel/reports/r20261108/evidence/run.json |
| Google Compute Engine | VERIFIED | Bu projede 3 VM: usta-wake RUNNING; ds-a100 ve usta-l4 TERMINATED. /workspace/us-panel/reports/r20261108/evidence/vms.json |
| Cloud Run | PARTIAL | Commerce /health=200, /healthz=404, /readyz=405; watchdog sağlık IAM403. /workspace/us-panel/reports/r20261108/evidence/run.json; /workspace/us-panel/reports/r20261108/evidence/us-commerce-health.json; /workspace/us-panel/reports/r20261108/evidence/us-commerce-healthz.json; /workspace/us-panel/reports/r20261108/evidence/us-commerce-readyz.json |
| Cloud SQL PostgreSQL | PARTIAL | us-commerce-pg RUNNABLE/ALWAYS; doğrudan DSN/migration testi yapılmadı. /workspace/us-panel/reports/r20261108/evidence/sql.json; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md:65 |
| Cloud Scheduler | PARTIAL | europe-west4 us-idle-watchdog ENABLED/OIDC hedefi listede; bu tur reconcile çağrılmadı. /workspace/us-panel/reports/r20261108/evidence/scheduler-west4.json |
| Secret Manager | PARTIAL | Metadata ve v2 disabled/v4 enabled; içerik okunmadı, rotation/reachability kabulü yok. /workspace/us-panel/reports/r20261108/evidence/secrets-metadata.json; /workspace/us-panel/reports/r20261108/evidence/policy-versions.json |
| Shopify mağazası | VERIFIED | Bağlı mağaza Kanus/shop.kanus.org, 7 aktif ürün; otomasyon kabulü değildir. /workspace/us-panel/reports/r20261108/evidence/shopify-info.txt; /workspace/us-panel/reports/r20261108/evidence/shopify-products.txt |
| Webhook sistemi | PARTIAL | HMAC/idempotency/reversal yerel/tarihsel sandbox kanıtı; taşınmış sunucuda yeni sipariş testi yok. /workspace/us-panel/reports/E2E_HOOK_KANUS_TEST_20261106.md; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log |
| Lisans ve kredi sistemi | PARTIAL | Yerel atomiklik/yarış testleri; canlı kimliksiz status401; taşınma sonrası uçtan uca yok. /workspace/us-panel/reports/r20261108/evidence/hook-credit-no-auth.meta; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log |
| E-posta gönderimi | PARTIAL | Tarihsel SMTP testi var; gerçek müşteri/outbox-retry zinciri bu tur NOT TESTED. /workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md |
| KANUS.ORG web sitesi | PARTIAL | HTTP200 SPA; /us paketinde otomatik hesap/lisans etkin değil bildirimi sürüyor. /workspace/us-panel/reports/r20261108/evidence/site-us.meta; /workspace/us-panel/reports/r20261108/evidence/site-us-automation-quotes.txt |
| Çok dilli arayüz | PARTIAL | GLOBAL 415/0/7 yerel raporu; native Windows FAIL/SKIP ve birleşik yayın kabulü açık. /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md; /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log |
| Global API ve SDK | PARTIAL | OpenAPI/SDK yerel öneri paketleri; bütünleşik yayımlanmış API/SDK kabulü yok. /workspace/us-panel/reports/KANUS_API_REPORT.md |


## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_CURRENT_STATE.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Canlı GET/GCP list-describe/Shopify read/BQ SELECT ile kanıtlar sabitlendi. Ürün onarımı ilgili sahip botların işidir. Kanıt: evidence/*.meta.

## Gerçek testler
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| work2-v11 test | PASS | /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.meta; test/subtest olayları: {'pass': 452, 'skip': 7} |
| work2-v11 vet | PASS | /workspace/us-panel/reports/r20261108/evidence/work2-v11-vet.log; /workspace/us-panel/reports/r20261108/evidence/work2-v11-vet.meta; test/subtest olayları: {} |
| work2-v11 race | PASS | /workspace/us-panel/reports/r20261108/evidence/work2-v11-race.log; /workspace/us-panel/reports/r20261108/evidence/work2-v11-race.meta; test/subtest olayları: {'pass': 452, 'skip': 7} |
| pinned-v10-layout test | PASS | /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-test.log; /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-test.meta; test/subtest olayları: {'pass': 386, 'skip': 7} |
| pinned-v10-layout vet | PASS | /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-vet.log; /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-vet.meta; test/subtest olayları: {} |
| pinned-v10-layout race | PASS | /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-race.log; /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-race.meta; test/subtest olayları: {'pass': 386, 'skip': 7} |


## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| Sürüm kaynağı | PASS | /workspace/us-panel/reports/r20261108/evidence/versions.json |
| Çalışan hook sürümü | BLOCKED | /workspace/us-panel/reports/r20261108/evidence/hook-api-version.meta |
| Yeni DNS | PASS | /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json |
| Ticari sürüm | BLOCKED / NO-GO | Aşağıdaki P0 listesi |


## Kalan eksikler
18 bileşenin tümünde kapsam sınırı tabloda. `/workspace/kan` ve `/workspace/sre` bu box kesitinde yok; mevcut kanıt `/workspace/us-panel/sre` ve raporlarda. `dist/` 00:23 TSİ kesitinde yok; round11-final altındaki build dosyaları tek başına release teslimi değildir. Kanıt: /workspace/us-panel/reports/r20261108/evidence/versions.json; /workspace/us-panel/reports/round11-final-20261108.

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## §3–§12 öncelikli eksik/kırık işler
| Öncelik | Bölüm | İş | Önerilen sahip | Kanıt |
|---|---|---|---|---|
| P0 | §3 | Test siparişi→hesap→lisans/kredi→e-posta→panel→gerçek model→ledger zincirini taşınan sunucuda tamamla | SHOP + HOST + US + QA | /workspace/us-panel/reports/E2E_HOOK_KANUS_TEST_20261106.md; /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json |
| P0 | §4 | SSD discard stop kodu ve yeni 00005/v5 etiketini açıklığa kavuştur; aynı image sonrası gerçek giderimi, VM kapanışını onaylı pencerede doğrula | KAN + SRE | /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md §00004; /workspace/us-panel/reports/r20261108/evidence/watchdog-final.json |
| P0 | §8 | Cloud Run /healthz404 /readyz405 yönlendirme ve OIDC+DB gerçek API kabulünü kapat; Ready=True yeterli değil | KAN + SRE + QA | /workspace/us-panel/reports/r20261108/evidence/us-commerce-healthz.json; /workspace/us-panel/reports/r20261108/evidence/us-commerce-readyz.json |
| P0 | §3/§4/§8 | Çalışan hook sürüm/hash ve .11 deploy, yeni IP GPU bağlantısı/rollback/store tutarlılığını HOST kanıtlasın | HOST + US + PANEL DEV | /workspace/us-panel/reports/r20261108/evidence/hook-api-version.meta; /workspace/us-panel/reports/r20261108/evidence/versions.json; /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json |
| P1 | §6 | Gerçek modelle dosya/bağlam/continue/atomic apply/build/test/repair kabulü; mock başarı gerçek model yerine sayılmasın | US + QA | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md §9 |
| P1 | §7 | KANUS SaaS Project Manager tam frontend/backend/PostgreSQL/RBAC/upload/Docker projesini üretip çalıştır | US + QA + PERF | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md:59; /workspace/us-panel/reports/KANUS_PERF_REPORT.md §11 |
| P1 | §10 | Windows i18n DACL düzeltmesini yeniden native koş; symlink SKIPini kapat; tam GUI/lisans/model/terminal/restart kabulü | GLOBAL + QA + US | /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR7; /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log |
| P2 | §5 | Billing gecikmesini ve kredileri ayır, kaynak/istek/tenant maliyetini işlemlerle ilişkilendir; bütçe alarmlarını öner | SRE + KAN + PERF | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json; /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| P2 | §9 | GLOBAL/UYUM/API yamalarını tek sabit .11 ağaçta inceleyip doğrula; SDK/org/authz/export/delete teknik kabulü | GLOBAL + API + UYUM + PANEL DEV | /workspace/us-panel/reports/KANUS_API_REPORT.md; /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md |
| P2 | §11 | Önce onaylı ≤20 eşzamanlı/≤10 dk/sentetik yeni-IP yük testi; 25/50 gerçek yük ayrı onay gerektirir | PERF + QA | /workspace/us-panel/reports/KANUS_PERF_REPORT.md; üst konuşma Demir 8 Kasım onayı |
| P3 | §12 | Bağımsız SAST/dependency/DAST; tenant OS izolasyonu, secret rotation, yeni host backup/restore ve log redaksiyonu kabulü | SRE + QA + HOST + US | /workspace/us-panel/reports/KANUS_SECURITY_REPORT.md; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md:71 |

Boşta kapanma: Demir 2026-11-08 00:20 TSİ'de 10 dk'ya (600 sn) çekti; çelişki kapandı. KAN 00:26'da us-watchdog-00006-88h (policy v5, idle 600, discardLocalSsd düzeltmesi) deploy etti; canlı kapanış doğrulaması bekliyor. Kanıt: KAN_WATCHDOG_DISCARD_FIX.md.

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
