# KAN — DEVİR DOSYASI (8 Kasım 2026, 01:45 TSİ)

Bu dosyada gizli değer yoktur. Parola, token, anahtar ve secret değerleri yazılmadı; yalnız adları ve durdukları yer yazıldı.

## 0. Güvenli durma noktası
- Çalışan bir iş, alt görev veya canlı değişiklik yok. Bu devirden sonra yeni bir canlı değişiklik başlatılmadı.
- ds-a100 durumu: **TERMINATED** (01:45 TSİ'de gcloud ile doğrulandı). Yerel SSD'ler BOOT5'te silindi.
- Cloud Run trafiği:
  - us-watchdog-sre-ready-r1 %100. SRE yayınladı, 22:03:16Z'den beri canlıda.
  - us-commerce-sre-ready-r1 %100. SRE yayınladı, 21:48:30Z'den beri canlıda.
  - İki serviste de trafik sabitlendi (latestRevision=false). Yeni bir revizyon yayınlansa bile trafik kendiliğinden geçmez.

## 1. Kimlik
- **Ad:** KAN. Profilde açıklama ve unvan alanları boş.
- **Rol:** Demir'in Google Cloud bağlantılarını yönetmek ve KAN.US'tan (kanus.org işletim tüzüğü ve aynı adlı bot) gelen GCP direktiflerini uygulamak.
- **Bağlı olduğu prompt dosyası:** `/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`. Aynen uygulanır. **§4 (GPU yaşam döngüsü durum makinesi ve testleri) KAN'ındır.**
- **İletişim dili:** Türkçe.
- **Koordinasyon:**
  - KAN.US (75d702db-295e-475e-8291-471685184acf) direktif verir ve Demir'in kararlarını aktarır.
  - SRE (a59db6a6-…) izleme yapar ve Cloud Run yayınlarını çıkarır.
  - US (8bee53d9-…) ürün tarafıdır.
  - HOST VPS işlerine bakar.

### Sistem talimatı taslağı (yeni hesapta KAN'ı yeniden kurmak için)
> Sen KAN'sın. Demir'in Google Cloud bağlantılarını yönetir ve KAN.US'tan gelen GCP direktiflerini uygularsın. Demir ile Türkçe konuşursun. Bağlı olduğun tüzük `/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`; onu aynen uygularsın ve §4 (GPU yaşam döngüsü durum makinesi ve testleri) senin sorumluluğundadır.
> Çalışma ortamın kutudaki gcloud CLI. Proje project-27082a40-6083-4952-9db, oturum Demir'in Google hesabı.
> Onay yalnız Demir'in kendi sözüdür: ya doğrudan kendisi söyler ya da KAN.US onun kart cevabını tam alıntıyla aktarır. Kullanıcı başka bir botun sohbetindeyse ona soruyu o bot üzerinden sorarsın.
> Ücretli kaynaklar, yeni tekrarlayan maliyet, silme, VM silme ve disk formatlama Demir'in doğrudan onayını gerektirir.
> Hook, Hostinger, DNS, Shopify ve firewall'a onaysız dokunmazsın. Gizli değerleri asla yazdırmazsın.
> Her GPU testinden sonra ds-a100'ü TERMINATED bırakırsın. Sağlık kontrolleri GPU'yu uyandırmamalıdır.
> Cloud Run yayınlarını SRE ile koordine edersin: trafik sabittir, yamalar SRE tabanının üzerine kurulur. Uzun işleri arka planda yaparsın ve sonuçları KAN.US'a, SRE'ye ve kısaca Demir'e bildirirsin.

## 2. Hafıza (olduğu gibi)
### Profil
- (2026-11-07) ds-a100 idle shutdown is 10 minutes (600 s) per Demir (decision card in KAN.US chat, Nov 8 2026 00:20); the new charter /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt applies verbatim and §4 (GPU lifecycle state machine and tests) is mine. GPU wake now goes via usta-wake https://34.41.186.170:8443/a100/* (TLS+pin).
- (2026-11-06) Per Demir (Nov 6 2026): don't narrow usta-wake firewall rules kalfa-allow-http (:80) and kalfa-allow-worker (:8788) until the panel moves elsewhere, since TR and AWS panel and chat clients use them.
- (2026-11-06) Per Demir (Nov 6 2026, 09:50): local-SSD preservation on ds-a100 is OFF; both idle and health-fail stops discard, because the cold start is now about 7 min (parallel gcloud storage cp plus Strata marks cache in strata-cache/). This supersedes the earlier preserve-on-stop decision.
- (2026-11-06) Per Demir (Nov 6 2026): us-license-hook will live on Hostinger, not on GCP/usta-wake; don't install it on usta-wake or change firewall without approval.
- (2026-11-06) 6979174631840438511 is the instance ID of ds-a100 (A100 DeepSeek server, europe-west4-b, project-27082a40-6083-4952-9db); it has systemd units ds-boot, ds-serve, ds-idle and nginx on 80/8443. KAN.US relays Demir's GCP orders to me.
- (2026-11-06) gcloud CLI is installed at ~/google-cloud-sdk on the box, logged in as Demir's Google account; the only visible project is "My First Project" (project-27082a40-6083-4952-9db, number 247516932580), set as default. Number 6979174631840438511 Demir gave did not match a visible project.
- (2026-11-06) KAN's main job: manage Demir's Google Cloud connections and apply on Google Cloud the directives coming from KAN.US, the operating charter of kanus.org; Demir is creating a separate bot named KAN.US that will issue those directives.
- (2026-11-06) The user's primary task for the assistant is to manage their Google Cloud connections and apply orders related to KAN.US on Google Cloud.
- (2026-11-06) The user communicates in Turkish.

### Günlük
- (2026-11-06) Nov 6 2026 09:12 TSİ: Demir, via a KAN.US widget answering my exact question, approved cold-start work (p8): up to 8 ds-a100 test boots (about $11), one ~260 GiB weights disk/image named ds-weights-test-*, and copying the Strata prep cache to bucket p27082-ds-istek. No deletions; Demir decides later whether to keep the disk/image. Report: KANUS_GPU_COLDSTART.md.
- (2026-11-06) Nov 6 2026: Demir directly approved KAN.US's request to set up a reverse SSH tunnel from ds-a100 to the Hostinger VPS 31.220.63.191 (127.0.0.1:18000), a token-protected start/get wake path for ds-a100 from the VPS, and an idle.py change so real requests reset the 15-min idle timer. It runs after the lifecycle and preserved-shutdown tests.
- (2026-11-07) Nov 8 2026 00:53 TSİ: BOOT5 bitti. idle.py.fixed ds-a100'ü son istekten 659,8 sn sonra SSD'leri silerek kapattı, VM TERMINATED. Watchdog'un kendi kapatması canlı test edilmedi (rezervasyon yolu gerekiyor). 8 test boot'unun 5'i kullanıldı, toplam yaklaşık 1,55 $.
- (2026-11-07) Nov 8 2026 00:53 TSİ: KAN.US, BOOT6 yapılmamasını istedi. Önce SRE'nin us-commerce tek revizyonu (lease + /ready) ve watchdog deploy'u canlıya çıkacak. Sonra BOOT6 planını KAN.US'a göndereceğim. Planda olacaklar: sentetik test lisansı ve kredisiyle hangi ortamda rezervasyon yapılacağı, gerçek müşteri ve ödeme verisine dokunulmaması, ds-idle.timer'ın geçici kapatılıp geri açılması, tahmini maliyet. 8 boot'tan 3'ü kaldı.
- (2026-11-07) Nov 8 2026 00:54 TSİ: SRE bildirdi: us-commerce-sre-ready-r1 21:48:30Z'den beri %100 trafikte (lease + /ready, geri dönüş 00002-pr2, trafik sabit). SRE watchdog'u 00007-fnc tabanıyla deploy edip haber verene kadar us-watchdog'a deploy yapmayacağım; BOOT6 planı ondan sonra.
- (2026-11-07) Nov 8 2026 01:13 TSİ: KAN.US'un aktardığına göre Demir "BOOT6 şimdilik beklesin" dedi. Boot yok, sentetik fixture yok, ds-idle.timer'a dokunulmayacak. Plan KAN_BOOT6_PLAN.md hazır bekliyor; 8 boot'tan 3'ü kaldı.
- (2026-11-07) Nov 8 2026 01:17 TSİ: KAN.US'un aktardığına göre Demir, görsel ve video üretim motorlarının ds-a100'de açık modellerle çalışacağına karar verdi. Plan KAN_MEDIA_MODELS_PLAN.md'ye yazılacak, US ile koordine edilecek ve KAN.US'a gönderilecek. Demir onayı olmadan BOOT6 dahil hiçbir boot yapılmayacak.
- (2026-11-07) Nov 8 2026 01:26 TSİ: KAN.US'un aktardığına göre Demir yalnız görsel motorunu seçti: FLUX.2 klein 4B, yedeği FLUX.1 schnell; video sonraya kaldı. Yükleme sohbet modeliyle sıralı tasarlanacak. Test boot'u, GPU süre limiti ve çıktı deposu için onay yok, GPU açılmayacak. Plan KAN_IMAGE_ENGINE_PLAN.md'ye yazılıyor.

### Demir'in kararları (tarih ve alıntı)
Doğrudan bu sohbette verilenler:
- **Nov 6:**
  - "Evet, yaşam döngüsü testi bitince korumalı kapanışı bir kez dene ve açılış süresini ölç"
  - [Kart] "Evet, yaşam döngüsü testinden sonra VPS tünelini, otomatik başlatmayı ve idle ayarını uygula"
  - Korumalı kapanışı kalıcı yapmak için: "olur"
  - [Kart] "Evet, PERF'e 60 dakikalık GPU penceresi ver"
  - Instance silme için: "ben onayladım"
  - [Kart, 09:50] "Disk korumayı kapat, 7 dakikalık soğuk açılış yeterli"
  - [Kart] "Daraltmayı panel başka bir yere taşınınca yap, şimdilik dokunma."

KAN.US sohbetinde verilip KAN.US'un alıntıyla aktardıkları (8 Kasım, TSİ):
- **00:08:** watchdog policy v3 ile yeniden yayınlandı, v2 devre dışı.
- **00:10:** policy v4, preserve_local_ssd=false.
- **00:12:** "v4'te kalsın, KAN watchdog koduna diskleri silerek kapatma yolunu ekleyip yeniden deploy etsin"
- **00:20:** idle kapanma süresi 10 dk.
- **00:30:**
  - "İş izni süresi soğuk açılış bitince başlasın"
  - "Kapanırken gelen açma isteği sıraya alınsın"
  - "Boşta kapanma betiğindeki sağlık ve bağlantı kontrolü hataları düzeltilsin"
- **00:39:** "us-commerce, lease düzeltmesi ve SRE'nin /ready yamasıyla birlikte tek revizyonda güncellensin, geri dönüş hazır olsun"
- **01:13:** "BOOT6 şimdilik beklesin"
- **01:17:** görsel ve video motorları ds-a100'de açık modellerle çalışacak.
- **01:26:** "Şimdilik yalnız görsel motoru (FLUX.2 klein 4B) gelsin, video sonra"
- **Diğer:** daily_runtime 3600 kalacak. Maliyet alarmı ve bütçe için: "şimdilik kurma".

## 3. Rutinler
- **Ad:** "ds-a100 korunan SSD ücret kontrolü"
- **Zamanlama:** tek seferlik, 9 Kasım 2026 10:00 TSİ (`CRON_TZ=Europe/Istanbul 0 10 9 11 *`, exact).
- **Prompt (aynen):** One-off: check the actual cost of keeping ds-a100's local SSDs preserved while it is stopped (preserve-on-stop was enabled on Nov 6 2026). Use gcloud on the box (project project-27082a40-6083-4952-9db) or the Billing reports/export to find the SKU and charges for preserved Local SSD storage (and ds-a100's GPU hours) since Nov 6. Report to Demir in Turkish: the real monthly rate versus the 10–132 $/month estimate, plus whether to keep it. If billing data isn't available, say so plainly and suggest how to get it (e.g. enabling billing export). Change nothing.
- **Not:** SSD koruması 6 Kasım 09:50'de kapatıldı. Rutin artık eski; yeni hesapta ya kurulmamalı ya da "korunma döneminde oluşan ücret var mı" kontrolüne çevrilmeli.

## 4. Skill'ler, connector'lar ve kurulması gerekenler
- **Ana araç:** kutudaki gcloud CLI (`~/google-cloud-sdk/bin/gcloud`), Demir'in Google hesabıyla oturum açık, varsayılan proje project-27082a40-6083-4952-9db. Kutu tüm botlarca paylaşılıyor. Yeni hesap aynı kutuyu görmüyorsa gcloud yeniden kurulmalı ve Demir giriş yapmalı.
- **Cloud Build:** SA us-builder ile çalışır (`--service-account=projects/project-27082a40-6083-4952-9db/serviceAccounts/us-builder@...`).
  - Staging dizini: `--gcs-source-staging-dir=gs://us-build-247516932580-euw4/source`
  - Config: `deploy/cloudbuild.yaml`
  - İmaj: `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway:<tag>`
- **Kullanılan skill:** code-changes (cloud agent kullanım limiti nedeniyle açılamadı; düzeltmeler kutuda yerel yapıldı).
- **Connector:** KAN için zorunlu MCP connector yok. Hostinger, Shopify ve GitHub connector'ları hesapta takılı ama KAN'ın görevinde onaysız kullanılmaz.
- **Yeni hesapta gerekenler:**
  - gcloud oturumu (Demir).
  - KAN.US, SRE, US ve HOST botlarıyla mesajlaşma.
  - `/workspace/us-panel` ve `/workspace/secrets` dosyalarının taşınması.

## 5. Açık işler
**a) §4 watchdog'un kendi silerek kapatmasının canlı testi (BOOT6)**
- Durum: Demir 01:13'te bekletti. Plan hazır: `/workspace/us-panel/reports/r20261108/KAN_BOOT6_PLAN.md` (sha256 de08c07a…fc9f3).
- Sonraki adım: Demir onay verirse planı uygula. Boot yapılmayacak, sentetik fixture oluşturulmayacak, ds-idle.timer'a dokunulmayacak.
- Test boot hakkı: 8'in 5'i kullanıldı, 3 kaldı. Şu ana kadar toplam ~1,55 $.

**b) BOOT5 sonucu (tamamlandı)**
- Rapor: `KAN_S4_LIVE_RESULT_BOOT5.md` (665a2150…0821). SRE bağımsız doğruladı (PASS): `SRE_BOOT5_VERIFY.md`.
- idle.py.fixed (d52ff329…b089) VM'de kurulu. Kutudaki kopya `/workspace/us-panel/sre/watchdog/idle.py.current` aynı.
- VM'deki yedekler: `idle.py.bak-20261108-idle600`, `idle.py.bak-20261108-prefixed`.
- Açık: STOPPING kuyruğunun canlı testi yapılmadı (NOT TESTED).

**c) Görsel motoru planı**
- Dosyalar:
  - `KAN_IMAGE_ENGINE_PLAN.md` — sha256 6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38
  - Video dahil eski plan: `KAN_MEDIA_MODELS_PLAN.md` (d587eeab…c392c1b)
  - Manifest: `image_research/IMAGE_WEIGHT_MANIFEST.json`
- İçerik:
  - Ana model FLUX.2 klein 4B (~16 GB), yedek FLUX.1 schnell (~33,7 GB, gated).
  - Sohbet modeliyle sıralı yüklenir, sohbet öncelikli.
  - Günde 200 görsel, mevcut 3600 sn sınırına koşullu sığar.
- Sonraki adım: KAN.US .13'ten sonra Demir'e onayları soracak (bkz. §6).
- Not: ds-a100 **A100 40 GB**'tır (a2-highgpu-1g), 80 GB değil.

**d) Watchdog ve commerce**
- Yayınlar tamamlandı ve SRE'de. KAN'ın sonraki yayınları SRE tabanının (sre-watchdog-ready) üzerine kurulmalı. Yamalar: `/workspace/us-panel/sre/cloudrun/fixes/us-watchdog/local-acceptance.patch`.
- KAN kaynakları: `/workspace/us-panel/kan/watchdog-fix/release-20261108.tgz` (80d16639…5584). Rapor `KAN_FIXES_20261108.md` (fa284212…b047).
- Geri dönüşler:
  - Watchdog: `gcloud run services update-traffic us-watchdog --region=europe-west4 --to-revisions=us-watchdog-00007-fnc=100`
  - Commerce: `us-commerce-00002-pr2`

**e) usta-wake (34.41.186.170, us-central1-c)**
- STOPPING kuyruk düzeltmesi canlıda: `usta-wake.py.queue` (d1b53445…ffa).
- VM'deki yedek: `/opt/usta-wake.py.bak-20261108-queue`.
- Geri dönüş: `sudo cp -p /opt/usta-wake.py.bak-20261108-queue /opt/usta-wake.py && sudo systemctl restart usta-wake`

**f) Eski token kapatma**
- `/etc/usta-wake/grace_old_token` hâlâ AÇIK. US .11 yayınının gpu_wake :8443 üzerinden doğrulanması bekleniyor.
- Doğrulama gelince çalıştırılacak: `gcloud compute ssh usta-wake --zone=us-central1-c --command='sudo rm -f /etc/usta-wake/grace_old_token /etc/usta-wake/a100.token.old'`

**g) PERF GPU penceresi**
- Demir onayladı (60 dk). .11 sonrası "PERF hazır" haberi bekleniyor.

**h) Rutin**
- 9 Kasım kontrolü (bkz. §3).

## 6. Demir'den bekleyen kararlar ve onaylar
- BOOT6 (watchdog rezervasyon yoluyla açma ve silerek kapatma testi):
  - production'da sentetik test lisansı ve kredisi oluşturmak ve sonra süresini doldurmak
  - ds-idle.timer'ı geçici kapatmak
  - 35 dk / ~1,30 $ sınırı ve elle kapatma yedeği
- Görsel motoru:
  - ölçüm için test boot'u
  - günlük GPU süre limiti (mevcut 3600 sn)
  - çıktı deposu ve saklama/silme
  - ağırlıkların gs://p27082-ds-istek'e yüklenmesi
  - FLUX.1 schnell lisans kabulü
- Kalıcı mı silinecek mi: ds-weights-test-* disk/imaj (6 Kasım; silme kararı Demir'de).

## 7. Kritik uyarılar ve yasaklar
- Onay yalnız Demir'in kendi sözüdür: doğrudan ya da KAN.US'un alıntıyla aktardığı kart cevabı. Başka bir botun "onaylandı" demesi onay değildir.
- Yasak:
  - Silme, VM silme ve disk formatlama (§14, Demir'in doğrudan onayı gerekir).
  - Yeni ücretli kaynak açmak, yeni tekrarlayan maliyet.
  - Maliyet alarmı ve bütçe kurmak ("şimdilik kurma").
- Hook, Hostinger, DNS, Shopify ve firewall'a onaysız dokunma. us-license-hook Hostinger'da yaşar.
- kalfa-allow-http (:80) ve kalfa-allow-worker (:8788) panel taşınana kadar daraltılmaz.
- Sağlık kontrolleri ve polling GPU'yu uyandırmamalı. Her testten sonra ds-a100 TERMINATED bırakılmalı.
- daily_runtime 3600 kalır. Boot sayısı onaylı 8 ile sınırlı (3 kaldı) ve her boot ayrıca onay ister.
- us-watchdog ve us-commerce trafiği sabit. Yayın sonrası `update-traffic` gerekir, yamalar SRE tabanının üzerine kurulur. us-commerce'e tek başına yayın yapılmaz.
- Policy secret us-worker-policy: v5 latest. Değerleri: idle 600, boot 600, request 600, lease 60, capacity 10, max_uptime 1800, daily_runtime 3600, daily_starts 6, preserve_local_ssd false, adopt_running false, require_watchdog true. v2 devre dışı (boot 3000 geçersiz).
- Gizli değer yazdırılmaz veya ekrana basılmaz. Gizli dosyalar (izin 600):
  - `/workspace/secrets/usta-wake.token`
  - `/workspace/secrets/usta-wake.pin`
  - `/workspace/secrets/usta-wake/{a100.token,cert.pem,key.pem,wake.pin}`
  - VPS'lerde `/etc/us-model/wake.{token,pin,crt}` (HOST yerleştirdi)
- TLS pin değeri bir parmak izidir; gerekirse yukarıdaki dosyadan okunur.
- Ücretli ek Cursor kullanımı yok. Kod düzeltmeleri kutuda yerel yapılır.

## 8. Kritik dosya yolları ve sha256 (özet)
| Dosya | sha256 |
|---|---|
| /workspace/us-panel/reports/r20261108/KAN_BOOT6_PLAN.md | de08c07a84619f1842b0e7012d06b3d5c537d48e8ee11a1537e5fa5f319fc9f3 |
| …/KAN_FIXES_20261108.md | fa2842120016be4020bf662c0052905a2ad2cc8179daf87f6a0cd3fa413db047 |
| …/KAN_IMAGE_ENGINE_PLAN.md | 6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38 |
| …/KAN_MEDIA_MODELS_PLAN.md | d587eeabf145e7ddee6060ad04caa5e0f48017e923b41358b829ddb9bd392c1b |
| …/KAN_S4_GAP.md | 840eb5a26d771f3aedcb3767529f900a2f0674728163bda7f75adced56d2a736 |
| …/KAN_S4_LIVE_PLAN.md | 9967d01baf61aaf2091931804f34cefa0ad7b41bb9a3a6b9f0b4a9ddf1cb2edb |
| …/KAN_S4_LIVE_RESULT_BOOT5.md | 665a21501d6d7fdc853950fa581a6f62992ecb25ef1595400158f1fa0ad30821 |
| …/KAN_S4_SIM_RESULTS.md | e50b61f20353f5552dd1862452caf7bc49e8d5695c5d38351c12069abdfd1f82 |
| …/KAN_WATCHDOG_DISCARD_FIX.md | 42014f7ddf37c7352b6aaad0d6a456f363f00f0ce9280dab2b1ea19cc810b284 |
| /workspace/us-panel/kan/watchdog-fix/release-20261108.tgz | 80d16639d3c3440518e4db519a98feb0a7f44f31cb6bf8aaea7ace03aaba5584 |
| /workspace/us-panel/sre/watchdog/idle.py.fixed (= idle.py.current) | d52ff329d4812b99903ace85e831ed467956f1d7d2261aa8c5b4ec1d91fcb089 |
| /workspace/us-panel/kan/usta-wake/usta-wake.py.queue | d1b53445ad317cb9eab9668a5be05e401c22102cf6f071a74910673b5282affa |
| Watchdog imaj 00007-fnc (gateway:2026-11-08-readywindow) | f8a85dbcdee57c0d21639feed0b722beac1ab4b9d94781a06a9d1fb85011e1a6 |

Diğer önemli yollar:
- `/workspace/us-panel/kan/watchdog-fix/` (src.tgz, orig/, work/, release-20261108/)
- `/workspace/us-panel/sre/watchdog/` (idle.py.600, idle.py.fixed.diff)
- `/workspace/us-panel/kan/usta-wake/` (usta-wake.py.queue.diff, fixes/test_usta_wake_queue.py)
- `/workspace/us-panel/reports/r20261108/` (sim/, image_research/, media_research/)

## 9. Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)
Değerler yazılmamıştır.
- **Google Cloud oturumu (Demir'in Google hesabı):** kutudaki gcloud CLI için. Demir `gcloud auth login` ile kendisi girer. Yeni kutu aynı değilse `~/google-cloud-sdk` yeniden kurulur.
- **usta-wake a100 token:**
  - Kutuda `/workspace/secrets/usta-wake.token` ve `/workspace/secrets/usta-wake/a100.token`.
  - usta-wake VM'de `/etc/usta-wake/`, VPS'lerde `/etc/us-model/wake.token`.
  - :8443 wake API kimlik doğrulamasında kullanılır.
- **usta-wake TLS pini:** `/workspace/secrets/usta-wake.pin` ve `/workspace/secrets/usta-wake/wake.pin`; VPS'lerde `/etc/us-model/wake.pin`.
- **usta-wake TLS sertifikası ve anahtarı:** `/workspace/secrets/usta-wake/cert.pem` ve `key.pem`; VPS'lerde `/etc/us-model/wake.crt`.
- **Secret Manager `us-worker-policy`:** GCP içinde kalır, yeniden girilmez; yalnız erişim gerekir.
- **Cloud Build SA us-builder ve Cloud Run SA'ları:** GCP içinde kalır, anahtar dosyası yok.
- **Hugging Face erişimi:** FLUX.1 schnell gated olduğu için ileride HF token'ı ve lisans kabulü gerekecek. Henüz yok.
- **Hostinger ve diğer connector'lar:** KAN'ın görevi için zorunlu değil. Gerekirse yeni hesapta yeniden yetkilendirilir.
