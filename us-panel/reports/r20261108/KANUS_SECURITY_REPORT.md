# Güvenlik kabulü — §12

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
Tarihsel güvenlik PASSları sürüm/host değişince otomatik taşınamaz. /workspace/us-panel/reports/KANUS_SECURITY_REPORT.md

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_SECURITY_REPORT.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Secret değerleri/credential dosyaları açılmadı; yalnız metadata okundu. Testler sentetik yerel verilerle, gerçek model kapalı. /workspace/us-panel/reports/r20261108/evidence/secrets-metadata.meta; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.meta

## Gerçek testler
Salt GET koruma kontrolleri; Go suite/vet/race; QA güvenlik ve Windows log incelemesi. SQL injection için gerçek DB/DAST koşulmadı. /workspace/us-panel/reports/r20261108/evidence/hook-credit-no-auth.meta; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| Auth/token/HMAC/SQLi/command/path/SSRF/privilege/file boundaries | PARTIAL | Hedefli yerel test; kapsamlı bağımsız audit değil. /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md |
| Kimliksiz model/status | PASS (dar canlı negatif) | /workspace/us-panel/reports/r20261108/evidence/hook-models-no-auth.meta; /workspace/us-panel/reports/r20261108/evidence/hook-credit-no-auth.meta |
| Hesap bazlı rotating-IP limiter | IMPLEMENTED; canlı NOT TESTED | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md §restore; /workspace/us-panel/reports/qa-20261108/account-limiter-local.log |
| Windows tercih DACL | PARTIAL / FAIL eski test | /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log; /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR7 |
| Secret metadata/versiyon | PASS (metadata); rotation NOT TESTED | /workspace/us-panel/reports/r20261108/evidence/policy-versions.json; /workspace/us-panel/reports/r20261108/evidence/secrets-metadata.json |
| SAST/dependency/DAST bağımsız tarama | NOT TESTED | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md:71 |
| Güçlü OS tenant izolasyonu | BLOCKED | /workspace/us-panel/reports/KANUS_SECURITY_REPORT.md:11 |
| Yedek/restore yeni host | NOT TESTED | Eski host checksum restore var; yeni host devri kabulü yok. /workspace/us-panel/reports/KANUS_SECURITY_REPORT.md §2; /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json |
| Hassas log redaksiyonu | PARTIAL | E-posta maskeli yerel test logları; bütün canlı loglar taranmadı. /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log |


## Kalan eksikler
Cloud SQL migration/restore ve yeni host backup bağımsız kabulü, broad secret/log taraması, container/OS izolasyonu, govulncheck/SAST/DAST sonuçları açık. Sahipleri değişikliği hazır saysa bile kanıtsız PASS yok. Kanıt: QA_TEST_RAPORU_20261106.md; KANUS_SECURITY_REPORT.md; Windows logu.

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
