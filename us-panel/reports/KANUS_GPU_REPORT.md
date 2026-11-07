# GPU ve model akışı — .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Ticari EXE -> Go VPS /v1/chat/completions -> US_MODEL_UPSTREAM. PC SSH butonu/özel model adresi ticari profilde kapalı. VPS'nin 1923 tüneli PC tünelinden ayrıdır.
- PASS — Go proxy global capacity=1 / queue=32; lisans -> admission -> reserve -> upstream -> doğrulanmış bitiş -> settle; read-only health/models kredi düşmez. Yerel 1/5/10/25/50 çağrı testinde upstream etkinlik <=1 ve başarılı çağrı kadar bakiye düşümü doğrulandı.
- PASS — Cold-start istemci 50 dk/sunucu 45 dk; 15 sn SSE heartbeat; gerçek token yokken GPU ısınıyor/kuyruk mesajı, hazır/başarılı denmez. Timeout/cancel gerçek upstream request context'ini keser ve rezervasyonu bırakır.
- PASS — Mevcut inference_lifecycle WorkerPolicy: idle_seconds=900, boot_seconds=600; workers.example.json güncel. Fake-provider 899/900 sn sınırı, aktif/queued lease, son request release, health'in idle resetlememesi testleri geçti. Go dev SSH readiness 10 dk sınırında; bu tur çağrılmadı.
- BLOCKED — Aktif Go model proxy ile GCP provider/start/stop/lease/watchdog entegrasyonu bağlı değil. Python controller kaynakta durur, Cloud Run issuer değildir ve bu pakette dağıtılmadı. VM kapalıysa Go yalnız upstream timeout/hata verir, kendisi uyandırmaz.
- BLOCKED — HOST özel ağ/SSH tüneli ve VM dosya/model hazırlığı; ds-a100 europe-west4-b ve ds-boot/ds-serve gerçek durumu bu tur sorgulanmadı. SSD koruma/IAM/watchdog ve stop akışı canlı kabul ister.
- NOT TESTED — ds-a100, instance-20261030-232245, usta-wake, usta-l4, us-commerce, us-watchdog/us-idle-watchdog güncel durumları; gerçek boot/idle-stop, TTFT/GPU maliyeti ve cold-start testi yok. Önceki ajanın istem-başı çalışma beyanı doğrulanmış sayılmaz.

## KAN GCP canlı test (06.11.2026)

Yazan: GCP canlı test ajanı. Saatler TSİ (UTC+3). Kaynak: gcloud durum takibi (15 sn aralık), ds-a100 journalctl / /var/log/ds-boot.log, usta-wake logu, Cloud Audit Logs. Yukarıdaki satırlar başka ajana aittir; bu bölüm onları değiştirmez.

### Test 1 — Soğuk açılış (usta-wake üzerinden, NVMe boş)
| Adım | Saat | Başlangıçtan |
|---|---|---|
| Uyandırma isteği (`X-Usta-Worker: a100`, GET /v1/models) | 05:54:40 | 0 |
| VM STAGING / RUNNING | 05:54:57 / 05:55:29 | +0:49 |
| ds-boot: 4 boş disk → RAID0 + ext4 + mount | 05:55:23 → 05:55:44 | +1:04 |
| Ağırlık çekimi bitti / BOOT_DONE (257 GB, retry yok) | ~06:07:15 | +12:35 |
| ds-serve: dosya kontrolü → "one time" hazırlık → yükleme | 06:07 → 06:21:08 → 06:22:02 | |
| Model yüklendi (health 200) | 06:22:50 | **+28:10** |
| usta-wake ilk /v1/models yanıtı (200) | ~06:22:50 | +28:10 |
| Chat (stream, max_tokens 32) → "Ankara" | 06:23:57 | ilk bayt 1,04 sn, toplam 7,7 sn |
| Son istek | 06:24:05 | |
| idle.py durdurma (STOPPING / TERMINATED) | 06:38:05 / 06:39:01 | son istekten ~14 dk |
Durduran: VM'in compute varsayılan SA'sı, Python-urllib, discardLocalSsd=true → idle.py. RUNNING süresi ≈ 42,7 dk.

### Test 2 — İkinci soğuk açılış örneği
Uyandırma 06:39:29 → RUNNING 06:40:05 → BOOT_DONE 06:50:13 (çekim ~10 dk) → Strata kontrol 06:50:14–07:03:48 → model yüklendi 07:05:29. **Toplam ≈ 26:00.**

## Korumalı kapanış testi
- Kota (europe-west4, DISKS_TOTAL_GB = "Persistent disk standard GB"): limit 2048, kullanım 0 → ≥1500 GB, test yapıldı.
- Öncesi: /opt/ds/idle.py → idle.py.bak-preserve-test yedeği, ds-idle.timer geçici durduruldu.
- `gcloud compute instances stop ds-a100 --discard-local-ssd=false` (GA gcloud'da var): 07:06:12 → TERMINATED 07:10:18 (komut 07:10:38'de döndü) → **durdurma ≈ 4 dk 26 sn** (normal discard durdurma ≈ 1 dk).
- Sonra usta-wake ile açılış:
| Adım | Saat | Başlangıçtan |
|---|---|---|
| Uyandırma isteği | 07:10:45 | 0 |
| RUNNING | 07:11:22 | +0:37 |
| ds-boot: mevcut RAID **assemble** edildi (create/mkfs yok), fetch tüm 7 dosyada "have" (indirme yok) | 07:11:32 → 07:11:37 | 5 sn |
| Strata dosya kontrolü | 07:11:43 → 07:13:00 | 1:17 |
| "one time" hazırlık + yükleme | 07:13:00 → 07:14:37 | 1:37 |
| Model yüklendi / usta-wake ilk yanıt (ttfb 230 sn) | 07:14:35–07:14:37 | **+3:52** |
| Chat (stream, 32 token) | 07:14:53 | ilk bayt 1,11 sn, toplam 7,7 sn |
- Karşılaştırma: soğuk 26–28 dk ↔ korunmuş SSD ile **~4 dk**. Korunmuş veri Spot preemption sırasında kaybolabilir (Google dokümanı) ve korunan alan durduğu süre boyunca ücretlenir.
- Sonrası geri alındı: ds-idle.timer enabled+active, idle.py yedekle birebir aynı (cmp), discardLocalSsd=true korunuyor. Bu açılış idle ile discard edilerek kapanacak (aşağıda onay).
- Faturalama: korunmuş depolama kalemi Billing'de henüz görünmüyor — DOĞRULANMADI.

### Bulgular (testler sırasında)
- usta-wake chat yanıtını tamponluyor: başlık 1 sn'de, tüm token'lar ~7,7 sn'de tek seferde geliyor (urllib `read(8192)`); gerçek SSE akışı yok.
- `Connection: close` göndermeyen istemcide usta-wake→ds-a100 bağlantısı açık kalabiliyor (Test 1'de /v1/models curl'ü bitmedi); açık :80 bağlantısı idle.py'nin "istemci var" saymasına yol açar → VM kapanmayabilir.
- idle.py `clients()` ":80 " eşleşmesi metadata sunucusuna (169.254.169.254:80) giden giden bağlantıları da sayıyor — yanlış "aktivite" riski.

### Maliyet (TAHMİN)
- Fiyat: a2-highgpu-1g Spot europe-west4 ≈ **$1,74/saat** (DevZero ve Usage AI, Billing Catalog tabanlı; gcloud-compute.com $1,83). Kaynaklar: https://www.devzero.io/instances/gcp/a2-highgpu-1g , https://instances.usage.ai/gcp/compute/region/europe-west4 , https://cloud.google.com/compute/gpus-pricing (Spot fiyatları değişkendir).
- 4×375 GB local SSD Spot: ≈ $0,10/saat varsayıldı (TAHMİN, doğrulanmadı). Toplam ≈ $1,84/saat.
- Test 1: 42,7 dk ≈ $1,31. Test 2: 26,2 dk ≈ $0,80. Test 3: ~20 dk ≈ $0,61 (kapanış saati aşağıda). Korunmuş depolama ~4–5 dk × 1500 GB ≈ <$0,01.
- Bir soğuk açılış (~27 dk) ≈ $0,83 + 15 dk idle kuyruğu ≈ $0,46 → boş bir uyandırma ≈ **$1,3**.
- Test 3 (korumalı SSD açılışı) sonrası: son istek 07:15:00. İlk stop beklenirken VPS tüneli kuruldu (aşağıda); tünelden gelen 07:29:18 isteği idle sayacını yeniledi.

## VPS tüneli ve uyandırma
Durum: KURULDU (07:25 TSİ). Soğuk TERMINATED durumundan uçtan uca VPS testi bu bölümün sonunda.

### Bileşenler
- **VPS (31.220.63.191, srv1662550):**
  - Kullanıcı `dstunnel`: sistem kullanıcısı, kabuk `/usr/sbin/nologin`, parola kilitli.
  - `/home/dstunnel/.ssh/authorized_keys`: tek anahtar (ds-a100'de üretildi), `restrict,port-forwarding,permitlisten="127.0.0.1:18000",command="/bin/false"`.
  - `/etc/ssh/sshd_config` sonuna `Match User dstunnel` bloğu eklendi: AllowTcpForwarding remote, GatewayPorts no, PermitListen 127.0.0.1:18000, PermitTTY no. Global AllowTcpForwarding **no** olarak kaldı. Yedek: `/etc/ssh/sshd_config.bak-20261106-dstunnel`.
  - Uyandırma token'ı: `/etc/us-model/wake.token` (root:root 600). Değer raporda yok.
- **ds-a100:**
  - Anahtar: `/root/.ssh/ds_tunnel_ed25519`.
  - VPS host anahtarı sabitlendi: `/root/.ssh/known_hosts_vps` (ED25519 SHA256:3is0zR97gEtY4eYhq1VYXfLckmiysGdPo5obhxR0/Qk; iki ayrı ağdan aynı görüldü).
  - Unit `/etc/systemd/system/ds-tunnel.service` (enabled): After/Requires/PartOf ds-serve; ExecStartPre `127.0.0.1/v1/models` hazır olana kadar bekler (≤60 dk); `ssh -N -R 127.0.0.1:18000:127.0.0.1:80`, ServerAliveInterval=30, ExitOnForwardFailure=yes, StrictHostKeyChecking=yes; Restart=always.
  - Hedef port 80: nginx → Strata 8080, OpenAI uyumlu /v1. nginx'e yönlendirme, istekleri access.log'a düşürür (idle için gerekli).
  - autossh yerine ssh + systemd Restart kullanıldı; paket kurulumu yok, davranış aynı.
- **Güvenlik duvarı:** değişiklik YOK. GCP egress varsayılan olarak açık; ds-a100 → 31.220.63.191:22 test edildi.
- **usta-wake (34.41.186.170, port 80):**
  - Yeni uç noktalar: `POST /a100/start` ve `GET /a100/status`. Yalnızca ds-a100'ü başlatır/sorgular; mevcut GPU tek-seferde kuralı korunur (stop_other_gpus).
  - Gerekli: `Authorization: Bearer <token>` ve kaynak IP 31.220.63.191; aksi halde 403.
  - Token: `/etc/usta-wake/a100.token` (root 600). Yedek: `/opt/usta-wake.py.bak-20261106-a100api`.
  - `/a100/start` yanıtı: 202 `start_requested`, ya da 200 `none` (VM zaten açıksa/açılıyorsa). `/a100/status` yanıtı: `{"instance": ..., "model_ready": bool}`.
  - Mevcut /health davranışı değişmedi; port 80 /health hâlâ seçili worker'ı uyandırır (önceki bulgu).
- **usta-wake SA:** compute varsayılan SA, özel rol `ustaWake` = instances.get/start/stop/resume/suspend, proje geneli. Bu, start+get'ten geniştir ve **ds-a100 da aynı SA'yı kullanır**. IAM değiştirilmedi (onay gerekli).

### idle.py değişikliği (yedek: /opt/ds/idle.py.bak-20261106-tunnel)
- Yeni: nginx access.log'da son kontrolden beri `/v1/...` (ve /chat, /models, /api) isteği varsa aktivite yenilenir. Ofset `/run/ds-a100-nginx-offset` dosyasında tutulur; log döndürülürse ofset sıfırlanır.
- `clients()` artık yalnızca YEREL portu 80/8080 olan bağlantıları sayar. Eski sürüm metadata sunucusuna (169.254.169.254:80) giden bağlantıları da sayıyordu; 07:23:14'te bu yüzden yanlış bir aktivite görüldü.
- Sessiz tünel (yalnızca SSH keepalive) yerel 80/8080 bağlantısı açmaz ve nginx'e log yazmaz → GPU'yu açık tutmaz.
- 15 dk idle, 70 dk health-fail stop ve discardLocalSsd=true değişmedi.
- Doğrulama: VPS'ten 07:29:18 tarihli `GET /v1/models` (tünel) → activity 07:30:04'te yenilendi.

### Canlı doğrulama (VM açıkken)
- VPS: `127.0.0.1:18000` yalnızca loopback'te LISTEN (dış arayüzde yok).
- VPS → `GET http://34.41.186.170/a100/status` (token ile) → 200 `{"instance":"RUNNING","model_ready":true}`.
- Yanlış token / izinsiz IP → 403.
- VPS → `GET http://127.0.0.1:18000/v1/models` → 200, deepseek-v4.1-flash loaded.

### Geri alma
1. ds-a100: `systemctl disable --now ds-tunnel; rm /etc/systemd/system/ds-tunnel.service; systemctl daemon-reload; cp -p /opt/ds/idle.py.bak-20261106-tunnel /opt/ds/idle.py`.
2. usta-wake: `cp -p /opt/usta-wake.py.bak-20261106-a100api /opt/usta-wake.py; systemctl restart usta-wake; rm /etc/usta-wake/a100.token`.
3. VPS: `cp -p /etc/ssh/sshd_config.bak-20261106-dstunnel /etc/ssh/sshd_config && sshd -t && systemctl reload ssh; userdel -r dstunnel; rm /etc/us-model/wake.token`.

### Yeni idle.py ile kapanış (Test 3 sonu)
- Son model isteği 07:29:18 (VPS tüneli). Tünel bağlı ama sessizdi. idle.py stop 07:45:54 (discardLocalSsd=true), STOPPING 07:46:05, **TERMINATED 07:47:09**. Son istekten ~16,5 dk sonra kapandı; sessiz tünel GPU'yu açık tutmadı.
- VPS'te 127.0.0.1:18000 dinleyicisi VM kapanınca düştü (LISTEN sayısı 0).
- Test 3 RUNNING süresi: 07:11:22–07:46:05 ≈ 34,7 dk ≈ $1,06 (TAHMİN). Kurulum için açık tutuldu.

### Soğuk uçtan uca test (VPS'ten) — başladı 07:50:17 TSİ
Uyandırma: VPS `POST /a100/start` → 202 (**07:50:18 TSİ** / 04:50:18 UTC). Kaynak: usta-wake journal. `scheduling.automaticRestart=false` (Spot).

| Adım | Saat (TSİ) | Başlangıçtan |
|---|---|---|
| Uyandırma (`POST /a100/start` → 202) | 07:50:18 | 0 |
| VM RUNNING | 07:50:35 | +0:17 |
| ds-boot: 4 boş disk → RAID0 + ext4 (önceki discard) | 07:50:52 | +0:34 |
| BOOT_DONE (257 GB çekim) | ~08:00 | ~+9:40 |
| Model ready (`model_ready=true`) | **08:15:31** | **+25:13** |
| Tünel LISTEN `127.0.0.1:18000` | 08:16:02 | +25:44 |
| VPS `GET …/v1/models` → 200 | 08:16:39 | +26:21 (0,50 sn) |
| VPS chat max_tokens 16 → 200 | 08:16:42 | ttfb≈3,0 sn |
| Son istek | 08:16:45 | |
| idle.py STOPPING (discardLocalSsd=**false**) | 08:33:45 | son istekten ~17 dk |
| TERMINATED | 08:36:07 | durdurma ≈ **2 dk 22 sn** |
| VPS 18000 LISTEN | 0 (düştü) | |

Sonuç: **PASS**. RUNNING ≈ 07:50:35–08:33:45 ≈ 43,2 dk ≈ **$1,32** (TAHMİN @ $1,84/sa). idle.py bu turda kalıcı korumalı moda alındı (aşağıda).

### Kalıcı korumalı kapanış
Uygulandı 08:10 TSİ civarı (VM açıkken). Yedek: `/opt/ds/idle.py.bak-20261106-prepreserve`. Kutu kopyası: `/workspace/us-panel/sre/watchdog/idle.py.current`.

- 15 dk idle stop → `stop?discardLocalSsd=false` (koru).
- 70 dk health-fail stop → `discardLocalSsd=true` (temiz yeniden çekim).
- `scheduling.automaticRestart=false` (Spot); koruma yine çalıştı.

**Geri alma:** `cp -p /opt/ds/idle.py.bak-20261106-prepreserve /opt/ds/idle.py` (ds-idle.timer zaten oneshot/python dosyasını her dk çalıştırır; unit restart gerekmez).

#### Hızlı korumalı uyanış doğrulaması (soğuk E2E kapanışından sonra)
| Adım | Saat (TSİ) | Başlangıçtan |
|---|---|---|
| VPS `POST /a100/start` → 202 | 08:37:09 | 0 |
| RUNNING | 08:38:13 | +1:04 |
| boot: tüm ağırlıklar **have** (indirme yok; RAID assemble) | 08:37:50Z boot | |
| model_ready | **08:40:47** | **+3:38** |
| Tünel LISTEN | 08:41:49 | +4:40 |
| VPS `/v1/models` → 200 | 08:42:24 | 0,40 sn |
| idle preserve STOPPING | 08:59:07 | son istekten ~16,7 dk |
| TERMINATED | 09:02:10 | durdurma ≈ **3 dk 03 sn** |
| VPS 18000 LISTEN | 0 | |

Sonuç hızlı korumalı uyanış: **PASS**. RUNNING ≈ 08:38:13–08:59:07 ≈ 20,9 dk ≈ **$0,64** (TAHMİN @ $1,84/sa). automaticRestart=false; korunan veri ikinci stop sonrası da ayaktaydı (bir sonraki açılışta doğrulanabilir; bu turda ikinci stop sonrası yeniden açılmadı).

Korunan veri: **evet** (7/7 `have`, blank create yok). automaticRestart=false iken preserve stop→start sonrası veri ayakta.

boot.sh (kod okuması): assemble başarısızsa yalnız *tüm* diskler blank ise RAID create+mkfs; aksi halde exit 1 → Spot preempt sonrası wipe’ta temiz yeniden kurulum yolu duruyor.

