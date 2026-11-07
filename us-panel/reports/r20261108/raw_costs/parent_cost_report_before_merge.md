# Google Cloud gerçek maliyet denetimi — §5

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
Eski SRE raporunda export boştu; şimdi standard 12.738 ve detailed 14.202 satır var. Bunlar 00:21 TSİ metadata kesiti; export gecikebilir. Kanıt: /workspace/us-panel/reports/r20261108/evidence/billing-standard-metadata.json; /workspace/us-panel/reports/r20261108/evidence/billing-resource-metadata.json

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_CLOUD_COSTS.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Salt SELECT, standart tablo (detailed ile birleştirip çift sayılmadı); krediler ayrı. İlk sorgunun ROWS alias sözdizimi hatası düzeltildi; ürün hatası değil. Dry-run 610.622 bayt; query max billed 20.000.000 bayt. Kanıt: /workspace/us-panel/reports/r20261108/evidence/billing-dry-run.json; /workspace/us-panel/reports/r20261108/evidence/billing-dry-run-v2.json; /workspace/us-panel/reports/r20261108/evidence/billing-query.sql

## Gerçek testler
gcloud VM/disks/snapshots/sql/run list; bq ls/show/query. Gerçek export hesap para birimi TRY; katalog USD tahminleriyle doğrudan karıştırılmadı. Kanıt: /workspace/us-panel/reports/r20261108/evidence/billing-costs.meta; /workspace/us-panel/reports/r20261108/evidence/billing-sku.meta

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| Kaynak envanteri | PASS | /workspace/us-panel/reports/r20261108/evidence/vms.json |
| Gün/hizmet/SKU export toplamları | PASS (mevcut kesit) | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json; /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Tam gün 7 Kasım + 8 Kasım / kesin fatura | BLOCKED | Export gecikmesi; /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Kullanıcı/istek başına gerçek maliyet | NOT TESTED | Tam eşlenmiş request-id, billable model kullanım sayacı/tenant ölçümü yok; /workspace/us-panel/reports/KANUS_PERF_REPORT.md |
| Gerçek aylık öngörü | BLOCKED | GPU aktif görev profili ve net kredi sürekliliği yok; /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |


## Kalan eksikler
Net krediler hizmetler arasında brütü aşabiliyor: düzeltme/kredi dağılımı ayrı hizmet netlerini negatif yapıyor; bu “GPU para kazandırıyor” demek değildir. 7 Kasım hizmetlerin son saatleri farklı, tam gün harcama değildir. Kullanım hacmi bilinmeden aylık ve kullanıcı/istek maliyeti uydurulmadı. Kanıt: billing-costs.json. Cloud SQL ALWAYS/PITR, persistent disk/snapshot ve usta-wake kapalı GPUdan bağımsız kalemler. Kanıt: sql.json/disks.json/snapshots.json/vms.json.

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## VM ve çalışma süresi kapsamı
| VM | Durum | Tür | Son start/stop TSİ (toplam runtime değildir) | Kanıt |
|---|---|---|---|---|
| usta-l4 | TERMINATED | g2-standard-4 | lastStartTimestamp=02.11 21:26:32 TSİ; lastStopTimestamp=03.11 03:09:32 TSİ | /workspace/us-panel/reports/r20261108/evidence/vms.json |
| usta-wake | RUNNING | e2-micro | lastStartTimestamp=31.10 08:14:41 TSİ | /workspace/us-panel/reports/r20261108/evidence/vms.json |
| ds-a100 | TERMINATED | a2-highgpu-1g | lastStartTimestamp=06.11 09:51:14 TSİ; lastStopTimestamp=06.11 09:53:20 TSİ | /workspace/us-panel/reports/r20261108/evidence/vms.json |

Açık VM=1; çalışan GPU=0; envanterde kapalı GPUlar 1×A100 ve 1×L4. instance-20261030-232245 listede yok; tarihi C4N SKUları bugünkü çalışan VM anlamına gelmez. Kanıt: vms.json/billing-sku.json.

## Gerçek günlük kesit
| Gün TSİ | Brüt TRY | Krediler TRY | Net TRY | Kanıt |
|---|---|---|---|---|
| 2026-11-01 | 1098.263875 | -1098.263881 | -0.000006 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-02 | 1545.879558 | -1545.879557 | 0.000001 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-03 | 1836.904029 | -1836.904039 | -0.000010 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-04 | 2274.261310 | -2274.261319 | -0.000009 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-05 | 3294.520610 | -3294.520637 | -0.000027 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-06 | 1042.204191 | -1042.204198 | -0.000007 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| 2026-11-07 | 87.927740 | -87.927735 | 0.000005 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |

## Gerçek hizmet kesiti
| Hizmet (Kasım mevcut export) | Brüt TRY | Krediler TRY | Net TRY | Kanıt |
|---|---|---|---|---|
| Artifact Registry | 0.000162 | -0.000161 | 0.000001 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| BigQuery | 0.000000 | 0.000000 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Cloud Build | 0.000000 | 0.000000 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Cloud Logging | 0.000000 | 0.000000 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Cloud Run | 1.184544 | -1.184544 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Cloud SQL | 160.608657 | -151.325945 | 9.282712 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Cloud Storage | 34.481125 | -33.869515 | 0.611610 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Compute Engine | 10908.261089 | -10918.155465 | -9.894376 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Deep Learning VM | 0.000000 | 0.000000 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Networking | 56.164467 | -56.164467 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| Secret Manager | 0.000000 | 0.000000 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |
| VM Manager | 19.261269 | -19.261269 | 0.000000 | /workspace/us-panel/reports/r20261108/evidence/billing-costs.json |

## Disk, snapshot ve ağ çıkışı
| Hizmet | SKU (6–7 Kasım mevcut) | Brüt TRY | Net TRY | Kanıt |
|---|---|---|---|---|
| Compute Engine | Multi-regional Snapshot upload within North America | 53.061886 | 0.142769 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | SSD backed Local Storage attached to Spot Preemptible VMs in Netherlands | 29.295002 | 0.113923 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Cloud Storage | Standard Storage Netherlands | 15.87127 | 0.50331 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Balanced PD Capacity in Netherlands | 15.153078 | 0.886643 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Storage PD Snapshot in US | 14.0135 | 0.926048 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Balanced PD Capacity | 8.497119 | 0.189916 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Hyperdisk Balanced Capacity in Iowa | 8.259569 | 0.041318 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | VM state: preserved local SSD in Netherlands | 6.197053 | 0.05093 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Hyperdisk Balanced Throughput in Iowa | 0.344144 | 0.001722 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Hyperdisk Balanced IOPS in Iowa | 0.172068 | 8.63E-4 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Internet Data Transfer Out from Netherlands to Americas | 0.008946 | 5.6E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Cloud SQL | Network Internet Data Transfer Out from Netherlands to Americas | 0.007607 | 5.4E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Internet Data Transfer Out from Netherlands to EMEA | 0.00245 | 1.0E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Inter Region Data Transfer Out from Netherlands to Americas | 0.002206 | 1.4E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Cloud SQL | Network Internet Data Transfer Out from Netherlands to EMEA | 0.001577 | 1.03E-4 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Data Transfer Out via Carrier Peering Network - Americas Based | 9.17E-4 | 3.8E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Cloud SQL | Network Inter Connect Data Transfer Out EMEA | 2.93E-4 | 2.3E-5 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Internet Data Transfer Out from Netherlands to Apac(Excluding Korea and Indonesia) | 4.1E-5 | 0.0 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Internet Data Transfer Out from Americas to Australia | 4.0E-6 | 0.0 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Cloud Run | Cloud Run Network Internet Data Transfer Out Intercontinental (Excl Oceania, Africa and China) | 2.0E-6 | 0.0 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |
| Compute Engine | Network Internet Data Transfer Out from Netherlands to South America | 1.0E-6 | 0.0 | /workspace/us-panel/reports/r20261108/evidence/billing-sku.json |

## Güvenli tasarruf planı / bütçe önerisi (uygulanmadı)
1. Watchdog discard düzeltmesi ve alarmını önce tamamla; preserved SSD SKU’su gerçekte oluşmuş. Kanıt: billing-sku.json.
2. us-commerce bağımlılığı/Cloud SQL store gereksinimini KAN incelemeden hiçbir şeyi kapatma; Cloud SQL kalıcı maliyeti ayrı. Kanıt: sql.json ve SRE_CLOUDRUN_20261106.md §5a.
3. Kullanılmayan disk/snapshotları sahipleri etiketlesin; silme yok. Kanıt: disks.json/snapshots.json.
4. Öneri: hizmet başına brüt/net bütçe ve %50/%80/%100 uyarı, günlük brüt sapma alarmı; eşik tutarını Demir seçmeli. Bu tur bütçe alarm varlığı doğrulanmadı ve oluşturulmadı. Kanıt: billing-costs.json; prompt §5.

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
