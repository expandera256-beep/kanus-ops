# ds-a100 soğuk açılış — analiz ve testler

Güncelleme: 06.11.2026 09:47 TSİ (Europe/Istanbul)
Yazan: GCP cold-start ajanı. Saatler TSİ. Ölçülmeyen her şey **NOT TESTED** veya **TAHMİN**.

Kapsam (Demir onayı): ≤8 boot (~11 USD); A+B öncelik; PD/imaj yalnız gerekirse (`ds-weights-test-*`); silme yok; hook/Hostinger/DNS/Shopify/VPS tüneline dokunulmadı. usta-wake :80 ile uyandırma boot 1–3’te kullanıldı; **bundan sonra test start/stop doğrudan gcloud** (usta-wake firewall daraltılacak).

Kaynaklar: `KANUS_GPU_REPORT.md`, `idle.py.current`, Audit Logs, Catalog JSON, Hyperdisk/PD dokümanları, VM `/opt/ds/*`.

## PASS / FAIL / NOT TESTED

| # | Seçenek | Durum | Ölçülen | Not |
|---|---|---|---|---|
| 0 | Eski soğuk (urllib fetch, marks yok) | PASS (önceki) | 25:13 – 28:10 | KANUS_GPU_REPORT |
| 0b | Korunan SSD | PASS (önceki) | 3:38 – 3:52 | Aylık koruma SKU NOT VERIFIED |
| A | Paralel `gcloud storage cp` (sliced) | **PASS** | fetch duvar **188–195 sn** | Engram 110 GB en yavaş |
| B | Strata `.done` marks + expert-profile | **PASS** | hash atlandı; Strata ~2–3 dk | `marks.tar` + `strata-cache/` |
| A+B | Paralel çekim + marks | **PASS** | **Boot2 ≈7:29, Boot3 ≈7:01** | Hedef &lt;10 dk |
| C1 | pd-balanced ~260 GiB | NOT TESTED | — | ~213 MiB/s → okuma ~20 dk TAHMİN; A+B yeterli |
| C2 | pd-ssd ~260 GiB | NOT TESTED | — | ~365 MiB/s TAHMİN |
| C3 | Hyperdisk Balanced | **FAIL (uygunsuz)** | — | A2 destek listesinde yok |
| C4 | Hyperdisk ML | NOT TESTED | — | A2 destekli; a2-1g max 1800 MiB/s |
| D | Disk imajı | NOT TESTED | — | Oluşturulmadı |
| E | mmap / CUDA graph | NOT TESTED | — | Load zaten ~1–2 dk |
| PERF | 60 dk GPU telemetri penceresi | **BEKLENİYOR** | — | GO gelmedi; VM tutulmadı |

**Boot sayacı: 3 / 8.** Yeni kalıcı disk/imaj: **yok**.  
GPU maliyeti (TAHMİN @ 1,848 USD/sa): boot1 0,25 + boot2 0,24 + boot3 0,21 ≈ **0,71 USD**.

---

## 1) Darboğaz kırılımı (eski soğuk)

| Parça | Süre | Pay |
|---|---|---|
| VM + RAID | ~1 dk | ~4% |
| GCS fetch (eski urllib, 5 worker) | **10–12 dk** | ~40% |
| Strata dosya SHA kontrolü (“complete”) + one-time | **13–14 dk** | ~50% |
| Model load | ~1,5 dk | ~6% |
| **Toplam** | **25–28 dk** | |

Korumalı SSD: fetch=0, kontrol+load ~3–4 dk → “one time” aslında büyük ölçüde **GGUF SHA256**; `.done` işaretleri NVMe’de kalınca atlanıyor.

### Test 1 zaman damgaları (KANUS_GPU_REPORT)
Uyandırma 05:54:40 → RUNNING +0:49 → BOOT_DONE ~+12:35 → model 06:22:50 (**+28:10**).

---

## 2) 70 dk health-fail stop

`idle.py.current`: `FAIL_STOP=70*60`; `not health_ok()` ve `uptime()>FAIL_STOP` → `stop(discard=True)`. Idle yolu: hazır + 15 dk sessizlik → `stop(discard=False)`.

Audit (son 7 gün): discard=true/false stop’lar ya `demirtumen08@gmail.com` (gcloud) ya VM’den `Python-urllib` (idle.py). **70 dk health-fail tetiklenmiş kanıt yok** (gözlem). Marj: soğuk ~28 dk (eski) / ~7 dk (yeni) vs 70 dk → **≥42 dk pay**.

Bu turda us-watchdog’dan gelen stop **görülmedi** (aşağı §6).

---

## 3) Seçenekler (özet maliyet)

| Seçenek | Aylık ek | Beklenen soğuk | Durum |
|---|---|---|---|
| A paralel GCS | 0 | fetch ~3 dk | PASS ölçüldü |
| B marks cache | GCS ~71 KiB ≈ 0 | Strata ~2–3 dk | PASS |
| A+B | ≈0 | **~7 dk** | **PASS — bırakılan boot yolu** |
| pd-balanced 260 GiB | ~28,6 USD (0,11×260 liste) | PD okuma sınırlı | atlandı |
| pd-ssd 260 GiB | ~48,6 USD | daha hızlı PD | atlandı |
| Hyperdisk Balanced | — | A2 yok | FAIL |
| Hyperdisk ML | fiyat NOT TESTED (Catalog 429 riski) | teorik hızlı | atlandı |
| İmaj | ~14 USD TAHMİN (0,055×260) | NOT TESTED | atlandı |
| Korunan SSD (ref, artık kapalı) | 66–132 USD senaryo | 3:38–3:52 | idle discard’a alındı 2026-11-06 |

---

## 4) Uygulanan boot yolu (kalıcı bırakıldı)

Yedekler: `/opt/ds/*.bak-20261106-coldstart`.

1. **`fetch-weights.py`**: `gcloud storage cp` + sliced env; 7 dosya paralel. Ölçülen duvar: Engram **188–195 sn** (≈10,5 Gbps etkin TAHMİN).
2. **`stamp-weights.py`**: mtime stamp + `/opt/ds/marks.tar` extract + `expert-profile.bin` kopyası.
3. **`/opt/ds/marks.tar`**: 7/7 `.done` (eksik shard4 üretildi).
4. **Bucket** `gs://p27082-ds-istek/strata-cache/` (marks, expert-profile, pack-source; ~71 KiB).
5. **`idle.py`**: değiştirilmedi bu turda — idle **preserve**, health-fail **discard** (önceki kalıcı politika).

PD/imaj **oluşturulmadı** (A+B hedefi karşıladı).

---

## 5) Test günlüğü

### Boot 1 — sıcak keşif (usta-wake)
| | TSİ |
|---|---|
| Uyandırma | 09:15:13 |
| RUNNING | 09:15:38 |
| ready ttfb | **222,6 sn ≈ 3:43** |
| discard=true stop → TERMINATED | 09:23:52 → 09:25:12 |
| GPU TAHMİN | ≈0,25 USD |

Keşif: `.done` = `sha256 <hex>`; `marks.tar` boot’ta açılmıyordu; shard4 `.done` eksikti.

### Boot 2 — soğuk A+B (usta-wake) — **PASS ≈7:29**
| | TSİ |
|---|---|
| Uyandırma | 09:25:29 |
| boot.sh | 09:26:10 UTC+0 log → 09:26:10 |
| Engram copy | 195,0 sn |
| BOOT_DONE / MARKS_RESTORED | ~09:29:25–09:30 |
| health=200 | **09:32:58** (**+7:29**) |
| discard=true stop | 09:33:47 → 09:35:03 |
| GPU TAHMİN | ≈0,24 USD |

### Boot 3 — soğuk doğrulama (usta-wake) — **PASS ≈7:01**
| | TSİ |
|---|---|
| Uyandırma | 09:35:15 |
| RUNNING | 09:35:40 |
| Engram copy | 188,4 sn |
| BOOT_DONE | ~09:39:20 |
| health=200 | **09:42:16** (**+7:01**) |
| /v1/models loaded; 7 `.done` + expert OK | 09:42 |
| preserve stop → TERMINATED | 09:42:37 → **09:46:18** |
| GPU TAHMİN | ≈0,21 USD |

### PERF penceresi
Demir onayı: 60 dk / ≤3 USD, GO mesajı şart. **GO gelmedi** → VM tutulmadı; son durum TERMINATED (korumalı NVMe).

### Sonraki testler
Start/stop: `gcloud compute instances start/stop ds-a100 --zone=europe-west4-b [--discard-local-ssd=…]`. Hazırlık: SSH ile `curl 127.0.0.1:8080/v1/models`.

---

## 6) us-watchdog / us-worker-policy

### Gözlem (test sırasında)
Audit’de ds-a100 start: usta-wake `Python-urllib/3.11`; stop: `demirtumen08` gcloud veya idle.py `Python-urllib/3.10`. **us-watchdog principal’ı görülmedi.**

Neden (kaynak `inference_lifecycle.py`): `adopt_running=false` iken `managed` olmayan RUNNING VM için stop claim **return None** — dışarıdan açılan VM’i kapatmaz. `max_uptime_seconds` / `daily_runtime_seconds` aşıldığında `limit_due` ile stop **eder** (yalnızca uyarı değil), ama yalnız **managed** worker’da (veya adopt_running true ise). `run_since` start anında set → boot süresi uptime’a dahil. `daily_runtime_seconds` gün içi `day_seconds` birikimi.

### Secret güncellemesi (testler bitince, onaylı)
- Diff latest→proposed: **yalnız** `idle_seconds` 120→**900**, `boot_seconds` 480→**3000**.
- `gcloud secrets versions add` → **version 2** enabled. **v1 enabled, silinmedi.**
- `latest` = idle 900 / boot 3000; `v1` = 120 / 480.

### us-watchdog okuma
- Env: `US_WORKERS_CONFIG=/secrets/policy/workers.json`
- Volume: secret `us-worker-policy`, item **`key: latest`** → `workers.json`
- Revizyon: `us-watchdog-00002-sgr`. Cloud Run secret volume “latest” genelde **revizyon oluşturulurken** çözülür; yeni secret sürümü **otomatik yüklenmeyebilir**.
- **Yeniden deploy yapılmadı** (talimat). Watchdog’un çalışan instance’ının hâlâ v1 okuyup okumadığı canlı doğrulanamadı (endpoint 403, kimlik token’sız).

### Uyarı — boot_seconds=3000 vs kod üst sınırı 600
Yerel `inference_lifecycle.py` doğrulaması: `boot_seconds` ∈ **[1, 600]**. Proposed **3000** bu kontrole takılır. Yeni bir Cloud Run revizyonu policy’yi `WorkerPolicy(...)` ile yüklerse **ValueError** riski. SRE yamasındaki 3600 üst sınırı deploy edilmemiş görünüyor. Secret v2 yazıldı; yükleme/redeploy **yapılmadı**.

---

## 7) Tutulan kaynaklar ve aylık maliyet

| Kaynak | Durum | Aylık |
|---|---|---|
| `gs://p27082-ds-istek/strata-cache/` | 10 nesne, ~71 KiB | ≈ **0 USD** |
| `/opt/ds` boot yolu A+B | boot disk’te | 0 ek |
| `ds-weights-test-*` disk/imaj | **oluşturulmadı** | 0 |
| Korunan local SSD (VM kapalı) | **kapatıldı** 09:53 discard stop; PRESERVED yok | 0 (koruma ücreti kesildi) |
| us-worker-policy v2 | enabled | secret ücreti ihmal |

---

## Sonuç

- Soğuk açılış **korunan SSD olmadan ~7 dk** (A+B). Hedef &lt;10 dk **PASS**.
- 4 boot, ≈**0,78 USD** GPU. 4 boot hakkı kaldı.
- PD/Hyperdisk/imaj gerekmedi.
- idle.py: idle **discard** / health-fail discard (2026-11-06).
- PERF GO bekleniyor; gelirse bir sonraki açık oturumda veya ayrı gcloud start ile 60 dk telemetri.
- us-worker-policy **v2** eklendi; watchdog redeploy yok; boot_seconds=3000 kod limiti ile çelişebilir.

### Disk koruma kapatıldı (2026-11-06, TSİ)

Demir kararı: 7 dk soğuk açılış yeterli → local SSD preserve **kapatıldı**.

| Adım | Zaman (TSİ) |
|---|---|
| gcloud start ds-a100 | 09:51:15 |
| `/opt/ds/idle.py` yedek + düzenleme + `py_compile` | 09:51:49–09:52:06 |
| gcloud stop `--discard-local-ssd=true` | 09:52:13 → 09:53:21 |
| VM durumu | **TERMINATED**; local-ssd-0..3 `savedState` artık **PRESERVED değil** (`DISK_SAVED_STATE_UNSPECIFIED`) |

**idle.py değişikliği** (yalnız 2 satır; yedek: `/opt/ds/idle.py.bak-20261106-nopreserve` ve box `sre/watchdog/idle.py.bak-20261106-nopreserve`):
- Yorum: idle preserve → her iki yol discard
- Idle yolu: `stop(discard=False)` → `stop(discard=True)` (health-fail zaten `True`)
- Kopya: `/workspace/us-panel/sre/watchdog/idle.py.current`

**Maliyet (bu boot):** ~2,1 dk × 1,848 USD/sa ≈ **0,065 USD** TAHMİN. Onaylı 8 boot’tan 4. kullanıldı (~0,71+0,07 ≈ **0,78 USD** kümülatif); **4 boot hakkı kaldı**.

**Geri alma:** VM açıkken `sudo cp /opt/ds/idle.py.bak-20261106-nopreserve /opt/ds/idle.py` (veya box yedeğinden), sonra bir sonraki idle stop’ta preserve yeniden açılır. Korunan SSD ücretini yeniden başlatmak için ayrıca `stop --discard-local-ssd=false` gerekir.

### Watchdog discard yolu

- Kontrol: 2026-11-08 00:13:13 TSİ (Europe/Istanbul). Bu çalışma yalnız salt-okuma ön kontrolünde durdu; kod, policy, env veya deploy değiştirilmedi.
- Çalışan revizyon: `us-watchdog-00004-kts`, trafik %100. Image: `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway@sha256:3e001dbec0a7d033ac105366bf282a59e4babff84e60413a33dbc2a7c474098a`; tag `2026-11-06-health`.
- Provenance: europe-west4 Cloud Build `31f9c0a8-8c46-4e9a-aefb-255eb8e46bac` SUCCESS; docker build --pull -t IMAGE .; kaynak `gs://us-build-247516932580-euw4/source/1793913254.263987-634563de94f346b7a5308e1af618a6e2.tgz`, generation `1793913254427846`. Yerel kaynakla birebir eşleşme henüz doğrulanmadı.
- Engelleyici: code-changes kuralı kod inceleme/değişiklik/test işini Cursor CloudAgent'a devretmeyi gerektiriyor; bu executor'da CloudAgent aracı yok (tool discovery ve plugin kontrolü yapıldı). Tahmini kod deploy edilmedi.
- Test: çalıştırılmadı; diff: yok; yeni image/revizyon: yok. ds-a100 `TERMINATED`, açılmadı. Mevcut revizyonun son 10 dakikalık severity>=ERROR log sorgusu boş döndü; deploy sonrası gözlem değildir.
- Geri dönüş komutu (çalıştırılmadı): `~/google-cloud-sdk/bin/gcloud run services update-traffic us-watchdog --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-watchdog-00004-kts=100`
