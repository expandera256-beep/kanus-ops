# Sabit maliyetler — envanter

Güncelleme: 06.11.2026 09:13 TSİ (Europe/Istanbul)

Salt rapor. Bilinmeyen tutarlar **boş**. USDTRY 50,05.

| Kalem | Durum | Bölge / spec | Saatlik / aylık | Not |
|---|---|---|---|---|
| Hostinger VPS (test) | aktif | srv1662550 · KVM 2 | **24,49 USD/ay** | Hostinger billing sub `AzZRHIVJKcZem2ycn` |
| Hostinger VPS (**prod**) | aktif | **srv2133566** · 82.29.184.225 · KVM 2 | **24,49 USD/ay** | sub `169lkvVaARjZU5YOD`; `PROD_PLAN.md` |
| Hostinger Mail | aktif | support@kanus.org | **(boş)** | SMTP PASS |
| Domain kanus.org | aktif | Hostinger | **(boş)** | |
| Shopify plan | — | — | **(boş)** | NOT TESTED |
| GCP `usta-wake` | RUNNING | us-central1-c · e2-micro · 10 GB | ≈**6,11 USD/ay** | |
| GCP `instance-20261030-232245` | SİLİNDİ | — | **0** | `GCP_CLEANUP_20261106.md` |
| GCP `usta-l4` disk | SİLİNDİ | — | **0** | snapshot ≈4,41 USD/ay (aşağı) |
| GCP `ds-a100` GPU (çalışırken) | Spot beklenen | europe-west4-b · a2-highgpu-1g + 4×375 GiB local SSD | **1,848238 USD/sa Spot** (Catalog JSON); üst sınır on-demand **3,748 USD/sa** | Sabit toplamda yok (idle-stop). `sre/billing_catalog_ds-a100_20261106.json` + `SRE_WATCHDOG_DESIGN_20261106.md` §4a |
| GCP `ds-a100` boot disk 50 GB | sabit (VM dururken de) | europe-west4-b · **pd-balanced** (`sre/disks.txt`) | **≈5,00 USD/ay** (50 × $0,000136986/GiB-sa × 730) | [disks-image-pricing](https://cloud.google.com/compute/disks-image-pricing); `GCP_CLEANUP_20261106.md` §5. Bölgesel 0,11 USD/GiB-ay listesi ≈5,50 olabilir |
| Snapshot saklama | READY ×2 | ~53 GiB | **≈4,42 USD/ay** | `GCP_CLEANUP_20261106.md` |
| **Billing BigQuery export** | **PASS (bildirim)** | dataset `billing_export`, proje `project-27082a40-6083-4952-9db` | — | **KAN.US bildirimi 08:21 TSİ**: standard + detailed export etkin. Dataset önceden: `GCP_CLEANUP_20261106.md` §3. **İlk veri ≤24 sa bekleniyor.** Console/API bu turda doğrulanmadı. |

### Toplamlar (bilinen)

| Küme | USD/ay |
|---|---:|
| usta-wake e2-micro | ~6,11 |
| ds-a100 50 GB pd-balanced boot | ~5,00 |
| Snapshot saklama | ~4,42 |
| **GCP bilinen ara** | **≈15,53** |
| Prod VPS srv2133566 KVM 2 | 24,49 |
| Test VPS srv1662550 KVM 2 | 24,49 |
| Shopify plan / Mail tahsisi / domain | **NOT KNOWN** (veya abonelikte ayrı; P&L tahsisi yok) |
| **GCP + prod VPS (COST_REPORT tabanı)** | **≈40,02 USD/ay** |
| **GCP + her iki VPS** | **≈64,51 USD/ay** |

ds-a100 Spot saatlik maliyeti sabit toplama **dahil değil** (yalnız çalışırken).
