# GPU yaşam döngüsü kabulü — §4

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
Kaynak `.11` (2026.11.06.11), pinned `.10`; `.6` yalnız eski referans. api-test /healthz=200 düz `ok`; `/api/version`, `/version`, `/v1/version`=404: çalışan hook sürümü **BLOCKED/doğrulanamadı**. Bu uçların yokluğu hookun .6/.10/.11 olduğunu ispatlamaz; panel sürüm endpointi ile hook sürümünü eşitleme. HOST binary/hash kanıtı gerekli. Kanıt: /workspace/us-panel/reports/r20261108/evidence/versions.json; /workspace/us-panel/reports/r20261108/evidence/hook-health.meta; /workspace/us-panel/reports/r20261108/evidence/hook-api-version.meta; /workspace/us-panel/reports/r20261108/evidence/hook-v1-version.meta.

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_GPU_ACCEPTANCE.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
VM veya GPU açılıp kapatılmadı; bozuk watchdog yolu mevcut image/config ve SRE incelemesiyle işaretlendi. Kanıt: /workspace/us-panel/reports/r20261108/evidence/run.json; /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md

## Gerçek testler
gcloud compute instances list; run services list/describe; secrets versions list metadata. Salt HTTP sağlık kontrolleri dışında worker :80/health çağrılmadı (tarihsel uyandırma riski). Kanıt: /workspace/us-panel/reports/r20261108/evidence/vms.meta; /workspace/us-panel/reports/r20261108/evidence/run.meta; /workspace/us-panel/reports/SRE_USTA_WAKE_HTTP_RISK_20261106.md

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| ds-a100 / usta-l4 / usta-wake | PASS (envanter) | /workspace/us-panel/reports/r20261108/evidence/vms.json |
| instance-20261030-232245 | MISSING (erişilebilen tek proje envanteri) | /workspace/us-panel/reports/r20261108/evidence/projects.json; /workspace/us-panel/reports/r20261108/evidence/vms.json |
| us-watchdog-00004-kts / v4 ilk kesit; son 00005-z97/v5 etiketi | PASS (config gözlemi) | /workspace/us-panel/reports/r20261108/evidence/run.json; /workspace/us-panel/reports/r20261108/evidence/watchdog-final.json; /workspace/us-panel/reports/r20261108/evidence/policy-versions.json |
| SSD silerek watchdog stop | FAIL (00004); 00005 giderim BLOCKED | /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md §00004; /workspace/us-panel/reports/r20261108/evidence/watchdog-final.json aynı image digest; yeni policy içeriği okunmadı |
| Boşta kapanma 10 dk (Demir 08.11 00:20) | PARTIAL | policy v5 idle 600, us-watchdog-00006-88h (00:26); idle.py 600 bir sonraki onaylı açılışta uygulanacak; canlı kapanış testi bekliyor |
| Cold boot / model / first response tarihçe | PARTIAL | /workspace/us-panel/reports/KANUS_GPU_REPORT.md (eski 25dk/korumalı 4dk); yeni ~7dk KAN bildirimi bağımsız ham kanıtla teyit edilmedi |
| Concurrent wake / cancel / active lease / queue / stop verify / restart / recover | NOT TESTED (yeni canlı .11) | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md; /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log yalnız yerel |
| Sağlık GPU uyandırmasın | PARTIAL | Bu tur yalnız Cloud Run health GET; öncesi VM kapalı; aktif sağlık→audit nedensel testi yapılmadı. /workspace/us-panel/reports/r20261108/evidence/vms.json; /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md |


## Kalan eksikler
00:25 TSİ son salt-okunur kesitte us-watchdog-00005-z97 Ready/%100, worker-policy etiketi v5 oldu; image digest hâlâ 3e001dbec0a7d033ac105366bf282a59e4babff84e60413a33dbc2a7c474098a. Bu config değişimi tek başına eski koddaki stop reddini düzeltme kanıtı değildir. Yeni policy değeri/gerçek stop bu tur okunmadı/denenmedi; KAN açıklaması ve SRE bağımsız yeni-image/stop kabulü gerekli. Kanıt: evidence/watchdog-final.json ve .meta. KAN düzeltmesi tamam varsayılmadı. Sonraki kabul gerçek discard API, max_uptime/daily quota ve aktif görev korunmasını göstermeli; 7dk startup tek başına tüm makine kabulü değildir. Kanıt: /workspace/us-panel/reports/SRE_CLOUDRUN_20261106.md; /workspace/us-panel/reports/KANUS_PERF_REPORT.md

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
