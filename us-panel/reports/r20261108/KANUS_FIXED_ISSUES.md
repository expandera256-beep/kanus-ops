# Düzeltilenler ve kapanmayanlar

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
Kaynak `.11` (2026.11.06.11), pinned `.10`; `.6` yalnız eski referans. api-test /healthz=200 düz `ok`; `/api/version`, `/version`, `/v1/version`=404: çalışan hook sürümü **BLOCKED/doğrulanamadı**. Bu uçların yokluğu hookun .6/.10/.11 olduğunu ispatlamaz; panel sürüm endpointi ile hook sürümünü eşitleme. HOST binary/hash kanıtı gerekli. Kanıt: /workspace/us-panel/reports/r20261108/evidence/versions.json; /workspace/us-panel/reports/r20261108/evidence/hook-health.meta; /workspace/us-panel/reports/r20261108/evidence/hook-api-version.meta; /workspace/us-panel/reports/r20261108/evidence/hook-v1-version.meta.

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_FIXED_ISSUES.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Bu yürütücü ürün kodu düzeltmedi; keşif yanlış/eski varsayımları ve izole test fixture yerleşimini düzeltti. Ekipteki kod değişiklikleri native/canlı kabul gelmeden kapanmış sayılmadı. Kanıt: aşağıdaki tablo.

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
| Kaynak gerçek sürümü .6 değil .11 | PASS (tespit; ürün düzeltmesi değil) | /workspace/us-panel/reports/r20261108/evidence/versions.json |
| “DNS hâlâ eski IP” bilgisi | PASS (gözlem güncellendi) | /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json |
| “Billing export boş” bilgisi | PASS (gözlem güncellendi) | /workspace/us-panel/reports/r20261108/evidence/billing-standard-metadata.json; /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Pinned v10 izole test ilk FAIL | PASS (test fixture düzeni) | İlk copy yanlış dizin: ../../../../reports fixture eksik. Orijinal iki fixture uygun dizine kopyalandı; ürün kodu değişmedi. /workspace/us-panel/reports/r20261108/evidence/pinned-v10-test.log; /workspace/us-panel/reports/r20261108/evidence/pinned-v10-layout-test.meta |
| BQ ilk alias hatası | PASS (yerel sorgu düzeltmesi) | /workspace/us-panel/reports/r20261108/evidence/billing-dry-run.json; /workspace/us-panel/reports/r20261108/evidence/billing-dry-run-v2.meta |
| Watchdog discard sorunu | FAIL (00004); yeni 00005 kabulü açık | /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md §00004; /workspace/us-panel/reports/r20261108/evidence/watchdog-final.json |
| Windows i18n düzeltmesi | PARTIAL / native tekrar açık | /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log; /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR7 |


## Kalan eksikler
Watchdog ve Windows native, tam Shopify/GPU/model kabulü açık. Ayrıntı: KANUS_CURRENT_STATE.md P0/P1 listesi.

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
