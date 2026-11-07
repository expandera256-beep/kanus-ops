# Windows EXE kabulü — §10

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
8 Kasım 00:15 TSİ Demir bilgisayarında eski i18n binary: 13 PASS/1 FAIL/1 SKIP; WinPref özel 3 PASS/1 SKIP. Kanıt: /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_WINDOWS_ACCEPTANCE.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Bu denetimde PCye dosya aktarılmadı/komut çalıştırılmadı. GLOBAL yeni DACL test ayrımı/düzeltme bildirmiş; native tekrar kanıtı gelmeden FIXED sayılmadı. Kanıt: /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR7

## Gerçek testler
Mevcut gerçek Windows logu okundu; Linux Go suite ayrıca koşuldu. Windows race/GUI gerçek testleri bu tur yapılmadı. Kanıt: /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| TestWinPrefNormalSave/JunctionBelowRoot/JunctionRootPinned | PASS | /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log:24–32 |
| TestPrefStoreRoundTrip | FAIL | /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log:19–21: perm -rw-rw-rw- |
| TestWinPrefSymlinkFileRejected | BLOCKED / SKIP | /workspace/us-panel-dev/global-work/ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log:28–30: symlink privilege |
| Windows mode bitleri ≠ gerçek ACL güvenliği | PARTIAL | GLOBAL TUR7 os.Stat0666 + miras DACL bulgusu; /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md |
| Yeni Windows DACL binary native tekrar | NOT TESTED | /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR7 |
| Tam güncel .11 EXE GUI kabulü | NOT TESTED | /workspace/us-panel/reports/WINDOWS_NATIVE_SMOKE_20261106.md yalnız eski CLI; /workspace/us-panel/reports/r20261108/evidence/versions.json |


## Kalan eksikler
Kurulum/başlatma/login/lisans/kredi/proje/model/dosya/terminal/build/test/kapat/restart zinciri güncel release EXE ile gerçek Windows’ta ayrı ayrı kanıtlanmalı. Linux suite veya cross-build bu kapıyı kapatmaz. Kanıt: prompt §10; /workspace/us-panel/reports/r20261108/evidence/versions.json

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`). Symlink testi için Developer Mode ya da yükseltilmiş yetki ayrı açık onay olmadan etkinleştirilmez.

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
