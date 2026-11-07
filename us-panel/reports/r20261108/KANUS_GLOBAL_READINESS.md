# Global ürün hazırlığı — §9 ve §11

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
TR/EN/DE/RU site paketinde var; bütün ürün/posta/hata akışlarını native+canlı kapsayan kabul yok. /workspace/us-panel/reports/r20261108/evidence/site-us-automation-quotes.txt; /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_GLOBAL_READINESS.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Mevcut GLOBAL/API/UYUM kanıtlarının uygulama/merge/canlı ayrımı korundu; dil menüsü tüm çevirilerin çalıştığına yeterli sayılmadı. /workspace/us-panel/reports/KANUS_API_REPORT.md

## Gerçek testler
Canlı SPA metni; Shopify para birimi TRY; GLOBAL yerel paket/site/race raporu; kendi Go testleri. Gerçek ülkeler arası yük koşulmadı. /workspace/us-panel/reports/r20261108/evidence/shopify-info.txt; /workspace/us-panel/reports/r20261108/evidence/site-assets.meta

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| TR/EN/DE/RU kapsam/parity/placeholder | PARTIAL | /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md TUR6/7 |
| Çok dilli hata/mail/fiyat | PARTIAL | /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md; /workspace/us-panel/reports/r20261108/evidence/shopify-info.txt |
| OpenAPI/Python/JavaScript SDK | IMPLEMENTED (yerel öneri) | /workspace/us-panel/reports/KANUS_API_REPORT.md |
| Org/authz/export/delete/GDPR-KVKK teknik kabul | NOT TESTED (tam birleşik canlı) | /workspace/us-panel/reports/KANUS_API_REPORT.md; /workspace/us-panel-dev/reports/UYUM_BOSLUK_ANALIZI_20261106.md |
| Global performans gerçek 1/5/10/25/50 GPU | NOT TESTED | /workspace/us-panel/reports/KANUS_PERF_REPORT.md |
| Hukuki/vergi/ticari kabul | BLOCKED / ayrı karar | /workspace/us-panel-dev/reports/UYUM_BOSLUK_ANALIZI_20261106.md |


## Kalan eksikler
Slogan/site release ve birleşik patch kabulü PANEL DEV kapısından geçmeli. SDK kaynakları var olmakla yayımlanmış uyumlu ürün sayılmaz. 25/50 yerel mock, gerçek GPU kapasitesi değildir. Kanıt: /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
