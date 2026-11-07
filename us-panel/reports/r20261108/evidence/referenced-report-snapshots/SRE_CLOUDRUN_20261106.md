# SRE Cloud Run Denetimi: GLOBAL_PROMPT §2 ve §3 izleme (06.11.2026, TSİ)

**Kapsam:**
- Yalnızca salt okuma: describe, IAM get, `logging read`, `bq ls/show`, dry-run, Catalog API, Artifact Registry'den image manifest ve uygulama katmanının indirilmesi.
- **Canlı HTTP isteği atılmadı.** ds-a100'e, Cloud Run sağlık uçlarına ve VPS'e erişilmedi.
- Hiçbir kaynak değiştirilmedi.
- Ham çıktılar `/workspace/us-panel/sre/cloudrun/` altında. Toplama zamanı 09:28–09:31 TSİ.

## 0. Prompt'tan bize düşen kısımlar (GLOBAL_PROMPT_20261106.txt)
- **§2 (satır 55-77):** us-commerce ve us-watchdog için görülen 404'lerin güncel durumda sürüp sürmediği doğrulanacak. Sürüyorsa şunlar yapılacak:
  1. URL'ler
  2. revision ve port
  3. IAM, ingress ve OIDC
  4. yönlendirmeler
  5. `/healthz` ve `/readyz` düzeltmesi
  6. sağlık sorgularının GPU başlatmadığı
  7. Scheduler çağrıları
  8. Cloud SQL
  9. merkezi log
  10. uçtan uca HTTP kabul testleri
  
  Kabul kriteri: "Canlı API erişimi doğrulanmadan sistemi üretime hazır ilan etme."
- **§3 izleme kısmı (satır 91-128):**
  - Zorunlu işlevler: "Sağlayıcı API'sinden gerçek kapanış doğrulaması", "Kapanış hatası uyarısı", "GPU maliyet kaydı", "Harcama limitleri".
  - Kaynak denetimi: us-commerce, us-watchdog, us-idle-watchdog. "Eski sürekli açık sunucuların gerçekten gerekli olup olmadığını belirle. Yeni mimari doğrulanmadan eski kaynakları kapatma."
  - Kabul maddesi 9: "Maliyet kaydını kontrol et."

## 1. Servis tablosu
| | us-commerce | us-watchdog | us-idle-watchdog |
|---|---|---|---|
| Tür | Cloud Run servisi | Cloud Run servisi | **Cloud Scheduler job** (Cloud Run servisi değil; `run services describe` NOT_FOUND) |
| URL | https://us-commerce-iqhblf63dq-ez.a.run.app | https://us-watchdog-iqhblf63dq-ez.a.run.app | hedef: us-watchdog `/internal/reconcile` |
| Revision ve trafik | us-commerce-00002-pr2, %100 latest | us-watchdog-00002-sgr, %100 latest | — |
| Image | `us-runtime/gateway@sha256:3e001dbe…098a` (tag `2026-11-06-health`) | **aynı image** | — |
| Port ve ingress | 8080 http1, ingress **all** | 8080 http1, ingress **all** | — |
| Min/max, CPU/bellek | 0/2, 1 vCPU / 512Mi, concurrency 32, timeout 600 | 0/1, 1 vCPU / 512Mi, concurrency 1, timeout 300 | — |
| cpu-throttling | anotasyon yok, varsayılan = istek bazlı (CPU yalnızca istek sırasında) | aynı | — |
| Servis hesabı | us-gateway@ | us-watchdog@ | OIDC: us-scheduler@ |
| run.invoker | **allUsers** | yalnızca us-scheduler@ | — |
| Startup probe | GET /health:8080 | GET /health:8080 | — |
| Cloud SQL | us-commerce-pg (db-custom-1-3840, ALWAYS) | aynı | — |
| Ready | True | True | ENABLED |

## 2. Bulgular (PASS / FAIL / BLOCKED / NOT TESTED)
| # | Madde | Durum | Kanıt |
|---|---|---|---|
| 2.1 | URL, revision, port, ingress, IAM | PASS (config) | `desc_*.json`, `iam_*.json` |
| 2.5a | `/healthz` düzeltmesi kodda | PASS | Deploy edilmiş image'ın uygulama katmanı indirildi (`image_app/app`). `runtime.py:38` ve `commerce_http.py:79`: `PATH_INFO in ('/health','/healthz')`. Yerel `source/from-demir-20261106/kanus` ve `work2` yalnızca `/healthz` içeriyor, yani **deploy edilen sürüm yereldeki kaynak kopyalardan daha yeni ve bu kopyalar güncel değil**. |
| 2.5b | `/health` deploy edilmiş revision'da | PASS (log) | Son 24 saat: us-commerce `/health` 8×200; us-watchdog `/health` 2×200 (token ile) ve 6×403 (token'sız, beklenen). Bunlar KAN ve diğer ajanların curl istekleri. |
| 2.5c | `/healthz` GFE'de rezerve | Bilinen kısıt | 24 saat içinde `/healthz` için **hiç request log kaydı yok**, istek uygulamaya ulaşmıyor (CLOUD_RUN_404_ROOT_CAUSE_20261106.md). |
| 2.5d | `/readyz` | **FAIL (yok)** | Image'da `readyz` route'u yok. Ayrıca adı `z` ile bitiyor, yani GFE'de aynı rezerv sorununa takılır. Öneri: `/ready` (Demir/KAN kararı). |
| 2.6a | Sağlık handler'ı GPU'yu başlatamaz | PASS (kod) | Her iki handler sabit JSON döndürüyor; controller, upstream veya provider çağrılmıyor. Watchdog'da start yolu yok: `reconcile()` yalnızca `stop` claim ediyor (`inference_lifecycle.py:239-258`). Start yalnızca gateway'deki model isteği yolunda (`ensure_ready`). |
| 2.6b | Audit kaydı ile sağlık istekleri karşılaştırması | PASS (gözlem) | ds-a100 audit kayıtları (03.11 17:51 – 06.11 09:26 TSİ): 14 start operasyonu (11 maskeli SA, 3 Demir). 18 sağlık isteğinin hiçbiri bir start'ın ±60 sn içinde değil (`analysis.txt`). Sınır: Cloud Run request logu yalnızca 05.11 23:44 TSİ'den beri var (revision'lar o zaman oluşturulmuş). |
| 2.6c | Sağlık sorgusunun canlı testi | NOT TESTED | Canlı istek yasaktı. |
| 2.7 | Scheduler çağrıları | PASS (log) | Son 24 saatte 584 AttemptStarted / 584 AttemptFinished, hepsi 200. `* * * * *` UTC. Son deneme 06.11 09:28:03 TSİ, `status {}` (hata yok). |
| 2.8 | Cloud SQL bağlantısı | NOT TESTED (dolaylı) | Reconcile, store'a yazıp 200 döndüğü için bağlantı fiilen çalışıyor görünüyor. Doğrudan bağlantı testi yapılmadı. |
| 2.9 | Merkezi log | PASS (kısmi) | Cloud Logging'de request logları var. Gunicorn `--access-logfile /dev/null` (Dockerfile) olduğu için uygulama access logu yok, yalnızca platform logu var. |
| 2.10 | Uçtan uca HTTP kabul testleri | NOT TESTED | Canlı istek yasaktı. |
| 3.x | us-watchdog'un GPU yaşam döngüsündeki fiili rolü | **Pasif** (bkz. §4) | 7 günde ds-a100 stop yapan principal'lar: Demir (6) ve maskeli SA (9). **us-watchdog@ hiç yok.** |

## 3. Scheduler ve OIDC
- Job `projects/…/locations/europe-west4/jobs/us-idle-watchdog`: POST `https://us-watchdog-iqhblf63dq-ez.a.run.app/internal/reconcile`, body `{}`. OIDC SA `us-scheduler@`, **audience `https://us-watchdog-iqhblf63dq-ez.a.run.app`** (servis URL'iyle uyumlu, PASS).
- Sıklık: her dakika (UTC). attemptDeadline 300 sn; retryCount 1, backoff 60 sn. State ENABLED. Son kullanıcı güncellemesi 05.11 23:45 TSİ.
- **Risk:** us-watchdog'un `ingress all` ayarı var ama invoker yalnızca us-scheduler. Kimlik doğrulaması olmayan istekler 403 alıyor (log: `/health` 6×403). Kabul edilebilir.
- **Bulgu:** US_WORKERS_CONFIG `/secrets/policy/workers.json` Secret Manager'dan bağlanıyor ve **okunmadı** (secret değerine erişilmedi). Bu yüzden ds-a100'ün us-watchdog politikasında olup olmadığı NOT VERIFIED.
  - Dolaylı kanıt: ds-a100 son 3 günde defalarca RUNNING durumuna geçti ve us-watchdog onu hiç durdurmadı. Ya politikada değil ya da guard'lar (lease/heartbeat) durdurmayı engelliyor.
  - ds-a100 politikadaysa kodun varsayılan `idle_seconds` değeri 120 sn (`inference_lifecycle.py:30`). Bu, KAN'ın 15 dk'sıyla çelişir. **KAN'ın teyidi gerekiyor.**

## 4. Gereklilik (son 7 gün trafik; log Cloud Run'da 05.11 23:44 TSİ'den beri)
| Servis | Trafik (kaynak/UA) | Bağlı akış | Öneri |
|---|---|---|---|
| us-commerce | Yalnızca curl (8.14.1/8.5.0), `US-readonly-route-audit/1`, 2 adet Python-urllib (account register/login, test) ve 2 tarayıcı favicon isteği. **Go/VPS hook'una ait UA yok**, Shopify webhook'u yok (`US_SHOPIFY_ENABLED=false`). | Lisans hook'u VPS'te. Bu gateway'in üretim trafiği yok. | **Kaldırılabilir adayı.** allUsers invoker ve sürekli açık Cloud SQL nedeniyle saldırı yüzeyi ve maliyet var. Önce VPS hook'unun bu servise hiçbir bağımlılığı olmadığı KAN tarafından doğrulanmalı (Demir ve KAN onayı). |
| us-watchdog | 584 Scheduler POST (200), birkaç curl. | `usModelStop` (ds-a100, instance düzeyinde) bağlaması var ama 7 günde stop yapmadı. Mail delivery döngüsü (`US_MAIL_PROVIDER` tanımlı değil, yani gönderim yok). | **Birleştirilsin:** SRE watchdog'un çalışma yeri olarak kalsın (IAM runbook D adımı). Bu durumda eski reconcile'ın policy'den ds-a100'ü çıkarması ya da `usModelStop` bağlamasının devir sonrası kaldırılması gerekir. SRE watchdog VPS'te çalışacaksa kaldırılabilir. |
| us-idle-watchdog | — | us-watchdog'u dakikada bir tetikliyor. | us-watchdog ile aynı kaderi paylaşmalı. Kaldırılırsa önce job **PAUSE** edilmeli (geri alınabilir). |

## 5. Aylık maliyet
- **Billing export:** `billing_export` altında `gcp_billing_export_v1_017786_B62E9C_E21182` ve `gcp_billing_export_resource_v1_017786_B62E9C_E21182` tabloları **oluşmuş** (08:47:34 TSİ), ikisi de DAY partitioned. **numRows = 0** (09:30 TSİ).
  - Servis bazında maliyet sorgusu (`cloudrun_cost_query.sql`) ve `billing_reconcile` SQL'i gerçek tabloya karşı dry-run edildi: ikisi de **"successfully validated", 0 bayt**. Böylece SQL sözdizimi PASS.
  - Gerçek maliyet: **NOT VERIFIED** (veri yok).
- **Catalog fiyatları** (Cloud Run `152E-C115-5142`, europe-west4, effectiveTime 2026-11-05T08:00Z):
  - Services CPU (Request-based) `4856-B847-F1EB` 0,000024 USD/vCPU-sn
  - Services Memory (Request-based) `02A2-9231-36A6` 0,0000025 USD/GiB-sn
  - Requests `2DA5-55D3-E679`: ilk 2M ücretsiz, sonrası 0,0000004
  - Cloud Scheduler `A5D6-43A2-28A7` Jobs: ilk 93 job-ay biriminde 0, sonrası 0,003225806/count (≈0,10 USD/job-ay)

| Servis | Üst sınır (max instance sürekli istek işliyor) | Gözleme dayalı tahmin (NOT VERIFIED) |
|---|---:|---:|
| us-commerce | 2 × (2.628.000 sn × 0,000024 + 0,5 GiB × 2.628.000 × 0,0000025) = **132,71 USD/ay** | toplam latency 23,39 sn / 7,66 saat → ~2.229 sn/ay → **≈0,06 USD/ay** |
| us-watchdog | 1 × aynı = **66,36 USD/ay** | 606 istek, 181,41 sn / 9,72 saat → ~13.624 sn/ay → **≈0,34 USD/ay** |
| us-idle-watchdog (Scheduler) | 1 job → 0 USD (ilk kademe) | 0 USD |
| Cloud SQL us-commerce-pg (§5a) | 57,96 USD/ay (+IP 8,03 → 65,99) + yedek/PITR NOT VERIFIED | aynı (ALWAYS, kullanımdan bağımsız) |
| **Toplam** | **≈265,06 USD/ay** (IP dahil; yedek/PITR hariç) | **≈66,39 USD/ay** (IP dahil; yedek/PITR hariç) |

- Gözleme dayalı tahminde free tier düşülmedi. Cold start ve instance başlangıç süresi dahil değil. İstek bazlı faturalamada boşta geçen süre ücretlendirilmiyor.
- **En büyük kalem Cloud SQL `us-commerce-pg`** (db-custom-1-3840, ALWAYS). Liste fiyatı §5a'da: ≈57,96 USD/ay (IP hariç), ≈65,99 USD/ay (IP dahil); yedek/PITR NOT VERIFIED. İki servis de buna bağlı. us-commerce kaldırılırsa us-watchdog'un store ihtiyacı yeniden değerlendirilmeli.

## 5a. Cloud SQL us-commerce-pg aylık liste fiyatı (09:4x TSİ, salt-okuma)

Kaynak: `gcloud sql instances describe` (sre/cloudrun/sql_describe.json) ve Catalog servis `9662-B51E-5089` (sre/cloudrun/catalog_cloudsql.json). Tüm SKU'larda effectiveTime 2026-11-05T08:00:00Z (05.11 11:00 TSİ). 730 saat/ay. Liste fiyatı; indirim, kredi ve vergi yok.

Instance: europe-west4-a, POSTGRES_17, db-custom-1-3840 (1 vCPU, 3,75 GiB), ENTERPRISE, ZONAL, PD_SSD 20 GB (autoresize açık, limit 60), PER_USE, activation ALWAYS. Yedek açık (7 adet tutuluyor), PITR açık (7 gün, CLOUD_STORAGE). Public IPv4 açık, private network yok.

| Kalem | SKU | Açıklama | Birim fiyat | Miktar | USD/ay |
|---|---|---|---:|---:|---:|
| vCPU | 002A-FBB4-3C32 | Cloud SQL for PostgreSQL: Zonal - vCPU in Netherlands | 0,0454 /vCPU-saat | 1 × 730 | 33,14 |
| RAM | EC82-E1D4-B318 | Zonal - RAM in Netherlands | 0,0077 /GiB-saat | 3,75 × 730 | 21,08 |
| Depolama | 813B-C170-DD76 | Zonal - Standard storage (SSD) | 0,187 /GiB-ay | 20 | 3,74 |
| IPv4 | 3C46-B5AA-2D9B | Zonal - IP address reservation | 0,011 /saat | 730 | 8,03 (NOT VERIFIED: kullanımdaki IP'ye uygulanıp uygulanmadığı; E959-45F5-15B2 "Free Tier IP address reservation" 0 USD) |
| Yedek | B8D4-7368-AFD4 | Cloud SQL: Backups in Netherlands | 0,088 /GiB-ay | bilinmiyor | **NOT VERIFIED** (yedek boyutu describe'da yok) |
| PITR logları | — | CLOUD_STORAGE'daki transaction log | — | bilinmiyor | **NOT VERIFIED** |
| **Toplam** | | | | | **57,96 (IP hariç) / 65,99 (IP dahil) + yedek/PITR** |

Örnek (doğrulanmamış): yedekler toplam 20 GiB olsaydı +1,76 USD/ay olurdu.

## 6. Demir ve KAN kararları
1. us-commerce kaldırılsın mı (allUsers + Cloud SQL)? Önkoşul: KAN'ın VPS hook bağımlılığının olmadığını teyit etmesi.
2. us-watchdog SRE watchdog'un çalışma yeri olacak mı, yoksa VPS mi? Workers policy'sinde ds-a100 var mı, idle 120 sn mi (KAN)?
3. `/readyz` yerine `/ready` eklensin mi (`z` ile bitmeyen bir yol)?
4. Gunicorn access logu açılsın mı (merkezi log, §2.9)?

## § Deploy sonrası denetim us-watchdog-00003-jdk (08.11)
Denetim zamanı: 08.11.2026 00:10–00:12 TSİ, yalnızca okuma. Ham çıktılar: sre/cloudrun/post_deploy_00003/ (policy içeriği maskelenmiş, `key` alanı [REDACTED]).

| # | Kontrol | Sonuç |
|---|---|---|
| 1a | latestReady = us-watchdog-00003-jdk, trafik %100 | PASS |
| 1b | image digest 00002-sgr ile aynı (sha256:3e001dbe…098a) | PASS |
| 1c | SA us-watchdog@; secret'lar (us-postgres-dsn, us-model-token, policy, ca) `latest` | PASS |
| 1d | ingress=all (IAM korumalı), minScale 0 / maxScale 1 (servis maxScale 3) | PASS (not: ingress internal değil) |
| 1e | IAM: yalnızca us-scheduler@ run.invoker, allUsers yok | PASS |
| 2a | us-worker-policy: v3 enabled, v2 disabled, v1 enabled | PASS |
| 2b | v3 değerleri: idle 900, boot 600, request 600, max_uptime 1800, daily_runtime 3600, daily_starts 6, preserve_local_ssd true, adopt_running false, require_watchdog true | PASS (KAN bildirimiyle aynı) |
| 2c | v1→v3 farkı: yalnızca idle 120→900, boot 480→600 | bilgi |
| 3a | 00003 logları: hata/ValueError/policy parse yok (pencere ~1 dk, 1 istek, 200) | PASS (kısa pencere) |
| 3b | Scheduler us-idle-watchdog → /internal/reconcile: AttemptFinished 200 | PASS (1 deneme) |
| 3c | Audit: 00:00 TSİ sonrası ds-a100 start/stop yok | PASS |
| 4 | ds-a100 TERMINATED; lastStart 06.11 09:51:14, lastStop 06.11 09:53:20 TSİ | PASS |
| 5 | preserve_local_ssd çelişkisi | FAIL (risk) |

### 5. Çelişki (kanıtlı)
- Deploy edilen kod `sre/cloudrun/image_app/app/gce_provider.py:83-91` (KAN kaynağıyla birebir aynı): VM'de local SSD varsa ve `preserve_local_ssd=false` ise `LifecycleError('refusing to discard local SSD')` verir. Değer true ise `discardLocalSsd=false` (beta API) ile durdurur. Yani watchdog SSD'yi **hiçbir zaman silmez**.
- VM içi `idle.py.current:97-124` ise idle ve health-fail durumunda `stop(discard=True)` çağırır, yani SSD'yi siler (Demir'in kararı).
- Risk: watchdog bir stop yaparsa (observe modunda değilse ya da max_uptime/daily kotası tetiklenirse) 4×375 GiB SSD korunur ve kapalıyken ücretlenir (≈66–132 USD/ay senaryosu; Preview özelliği, Spot preemption'da veri yine kaybolur). Sonraki açılışta da idle.py'nin varsaydığı soğuk açılış yerine korunmuş disk gelir. Policy'yi `false` yapmak çözüm değil: o zaman watchdog stop'u reddeder ve VM'i hiç durduramaz. Doğru düzeltme kodda yapılmalı: `discardLocalSsd=true` yolunu açık bir izinle eklemek. Bu karar ve değişiklik KAN'ın/Demir'in.
- Aynı nedenle yerel spend.py `preserve_local_ssd` değerini policy'den değil config'den okumalı; policy değeri true kaldıkça SSD maliyet tahmini çelişkili olur.

NOT TESTED: uzun log penceresi (yalnızca ~1 dk veri vardı), endpoint HTTP davranışı (istek atılmadı).

## § Deploy sonrası denetim us-watchdog-00004-kts (08.11 ~00:13 TSİ, salt okuma)
| # | Kontrol | Sonuç |
|---|---|---|
| 1 | Trafik %100 → us-watchdog-00004-kts (latestRevision) | PASS |
| 2 | us-worker-policy: v4 enabled (07.11 21:10:53Z = 08.11 00:10 TSİ), v3 enabled, v2 disabled, v1 enabled | PASS |
| 3 | latest (v4) ds-a100: idle 900, boot 600, request 600, max_uptime 1800, daily_runtime 3600, daily_starts 6, preserve_local_ssd **false**, adopt_running false, require_watchdog true | PASS (değerler) |
| 4 | Image aynı → `gce_provider.py:88-89`: local SSD varsa ve preserve_local_ssd=false ise `LifecycleError('refusing to discard local SSD')` | **FAIL (risk)** |

Sonuç: v4 ile SSD korunarak durdurma riski kalktı, ama watchdog artık ds-a100'ü (4 local SSD) **hiç durduramaz**. max_uptime 1800 / daily_runtime 3600 / idle kotaları watchdog tarafından uygulanamaz; kota dolduğunda reconcile her dakika LifecycleError üretebilir (log/alarm gürültüsü). Kapatma yalnız VM içi idle.py'ye (discard=True) kalır. Düzeltme kodda: açık izin bayrağıyla `discardLocalSsd=true` stop yolu + test; deploy Demir onayı ister.
NOT TESTED: v4 ile gerçek stop denemesi (VM TERMINATED; tetiklenmedi), uzun log penceresi.
