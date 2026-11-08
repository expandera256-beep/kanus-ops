# KAN — Yalnız görsel motoru planı: FLUX.2 klein 4B + schnell yedek

**Durum: PLAN ONLY / UYGULANMADI.** Karar: Demir, 8 Kasım 2026 **01:26 TSİ** (Europe/Istanbul, UTC+3), KAN.US aracılığıyla: şimdilik yalnız görsel; ana motor **FLUX.2 klein 4B**, yedek **FLUX.1 schnell**. Video daha sonra.

## 1. Kapsam ve sabit sınırlar

- Bu dosya yalnız plan ve küçük kamusal metadata/lisans kanıtlarını içerir. GPU açılmadı; VM, GCP, bucket, policy, timer veya canlı servis değiştirilmedi. Ağırlık indirilmedi, GCS'ye yükleme yapılmadı.
- Önceki `KAN_MEDIA_MODELS_PLAN.md` **değiştirilmedi**; video adayları ve ayrıntılı eski araştırma orada referans olarak kalır. Video motoru, video API'si ve video gerekçeli runtime artışı bu dar kapsamdan çıkarıldı.
- Son salt okunur kanıt: `media_research/ds-a100-readonly.json`, **8 Kasım 01:20:46 TSİ**; `ds-a100 = TERMINATED`, `a2-highgpu-1g`, tek `nvidia-tesla-a100`. Bu **A100 40 GB** / 85 GB host RAM'dir; 80 GB değildir. [G]
- **idle=600 sn, max_uptime=1800 sn, daily_runtime=3600 sn** korunur. Görsel pilotu için limit artışı varsayılmıyor. Sağlık/list/poll kontrolü GPU'yu uyandırmaz.
- **BOOT6 halen beklemede**, 8 test açılışından 3 hak kaldı. Model seçimi kararı, test boot veya üretim deploy onayı değildir. Hiçbir boot başlatılmayacak.

## 2. Sabitlenecek HF repo/revision ve bileşen lisansları

### 2.1 Çalışma anında kullanılacak iki tam snapshot

| Rol | Kesin HF repo | Sabit revision (commit SHA) | Lisans |
|---|---|---|---|
| Ana, **distilled 4B** | `black-forest-labs/FLUX.2-klein-4B` | `e7b7dc27f91deacad38e78976d1f2b499d76a294` | Apache-2.0; repo `LICENSE.md` [K], [KL] |
| Yedek | `black-forest-labs/FLUX.1-schnell` | `741f7c3ce8b383c54771c7003378a50191e9efe9` | Apache-2.0; resmi model lisansı BFL `flux/model_licenses/LICENSE-FLUX1-schnell` [S], [SL] |

`main`, kayan tag veya modellerin en son sürümü çalışma anında kullanılmaz. **klein-base-4B, klein-9B, FLUX.1-dev veya FLUX.2-dev denoiser seçilmez.** Özellikle 9B/dev lisansları bu Apache iznine dahil değildir.

Klein pinned `model_index.json`: `Flux2KleinPipeline`, Qwen3 text encoder, `AutoencoderKLFlux2`; `is_distilled=true`. VAE config içindeki `_name_or_path=FLUX.2-dev` yalnız provenance'dır: **VAE'nin ayrı Apache lisansı** resmi BFL dokümanında açıkça yazıyor; dev denoiser lisansını VAE'ye taşımıyoruz. [V2]

**Schnell erişim notu:** HF kamusal metadata (revision, dosya boyutu, LFS SHA256) okunabildi; pinned ham config istekleri **401/gated** döndü. Bunlar erişim açılarak doğrulanmış gibi sunulmadı. Resmi açık model kartı, BFL kaynak/lisans dosyaları ve kamusal manifest kullanıldı. Uygulama öncesi izinli HF erişimi ve gerekiyorsa lisans kabulü Demir'in yönettiği akışta tamamlanmalı; hiçbir token aranmadı/okunmadı ve erişim engeli aşılmadı.

### 2.2 Encoder/VAE ve muhafaza edilecek dosyalar

**Ağırlıklar encoder/VAE dahil yukarıdaki BFL snapshot'larından alınacak.** Aşağıdaki upstream SHA'lar lisans/provenance referanslarıdır; çalışma anında ayrıca upstream'den yeni encoder ağırlığı indirilmesini gerektirmez. Böylece encoder güncellemesiyle pipeline değişmez.

| Bileşen | Kaynak/provenance ve sabit referans | Lisans / saklanacak dosya |
|---|---|---|
| Klein Qwen3 4B text encoder + tokenizer | BFL pinned `text_encoder/`, `tokenizer/`; upstream `Qwen/Qwen3-4B@1cfa9a7208912126459214e8b04321603b3df60c` | **Apache-2.0**; upstream `LICENSE` → pakette `licenses/Qwen3-4B-LICENSE.txt`. [Q] |
| Klein FLUX.2 VAE | BFL pinned `vae/`; resmi `black-forest-labs/flux2@50fe5162777813d869182b139e83b10743caef15`, README “FLUX.2 autoencoder” | **Apache-2.0**, repo `LICENSE.md` + VAE'yi açıkça kapsayan resmi README kopyası. Klein VAE yerine schnell VAE kullanılmaz. [KL], [V2] |
| Schnell CLIP ViT-L/14 text encoder + tokenizer | BFL pinned `text_encoder/`, `tokenizer/`; provenance `openai/clip-vit-large-patch14@32bd64288804d66eefd0ccbe215aa642df71cc41` | **MIT**; resmi `openai/CLIP@d05afc436d78f1c48dc0dbf8e5980a9d471f35f6/LICENSE` → `licenses/CLIP-LICENSE-MIT.txt`. HF repo manifestinde ayrı LICENSE yok; upstream MIT notice korunur. [C] |
| Schnell T5 v1.1 XXL encoder + tokenizer | BFL pinned `text_encoder_2/`, `tokenizer_2/`; provenance `google/t5-v1_1-xxl@3db67ab1af984cf10548a73467f0e5bca2aaaeb2` | **Apache-2.0** (HF metadata + resmi T5 projesi); `google-research/text-to-text-transfer-transformer@42909aa15d1fbd16c69333aa474a00ae8ed9fde4/LICENSE` → `licenses/T5-LICENSE-Apache-2.0.txt`. HF repo manifestinde ayrı LICENSE yok. [T] |
| Schnell FLUX.1 VAE | BFL pinned `vae/`; resmi `black-forest-labs/flux@802fb4713906133fcbd0d8dc5351620ca4773036/README.md` autoencoder açıklaması | **Apache-2.0**, resmi model lisansı ve autoencoder kapsam açıklaması saklanır. [V1], [SL] |

Schnell encoder türleri resmi BFL `flux/src/flux/util.py` içindeki `load_t5` / `load_clip` eşleştirmesiyle çapraz kontrol edildi; gated pinned config ve upstream orijinal ağırlığın byte-byte eşliği **ayrıca doğrulanmış sayılmıyor**. Gerçek runtime byte kaynağı ve hash'i aşağıdaki BFL manifestidir.

Apache LICENSE/NOTICE, MIT copyright/permission ve değişiklik bildirimleri dağıtım paketinde korunmalı. Kaynakta NOTICE varsa alınmalı; yoksa varmış gibi yazılmamalı. Aynı paket içinde tokenizer/konfigürasyon ve lisans provenance dosyaları da sürümlenir. CLIP kartındaki deployment/English sınırlılığı değerlendirme uyarısıdır; MIT yerine gizli bir commercial yasağı diye yorumlanmaz. Çıktıların hak/güvenlik uygunluğu ayrıca değerlendirilir.

Lisans dosyalarının küçük yerel kanıt kopyaları ve SHA256'ları `image_research/` altındadır; bu bir lisans kabulü veya dış paylaşım değildir. Python bağımlılıkları da uygulamada exact lockfile/wheel hash ile sabitlenmeli; repo kartındaki `pip install ... git main` komutu üretimde aynen kullanılmamalı. Bu plan Python/CUDA ortamına kurulum yapmadı.

## 3. Kesin ağırlık manifesti ve boyutlar

**GB=10^9 bayt.** Diffusers bileşen dosyaları seçildi; aynı repo içindeki monolitik kopyalar tekrar sayılmaz. Tam `repo/revision/path/bytes/sha256` ve küçük yardımcı dosya listesi: **`image_research/IMAGE_WEIGHT_MANIFEST.json`**. Hash'ler resmi HF LFS metadata'sından; checkpoint indirilip yeniden hesaplanmış yerel hash değildir. İlk gerçek staging'de hesaplanan hash ile doğrulanması gerekir.

### black-forest-labs/FLUX.2-klein-4B

Revision: `e7b7dc27f91deacad38e78976d1f2b499d76a294`

| Dosya | Bayt | GB | SHA256 |
|---|---:|---:|---|
| `text_encoder/model-00001-of-00002.safetensors` | 4,967,215,360 | 4.967215 | `8c0506e7f4936fa7e26183a4fd8da4e2bdbc5990ba64ae441f965d51228f36ea` |
| `text_encoder/model-00002-of-00002.safetensors` | 3,077,766,632 | 3.077767 | `82f2bd839378541b0557bfabaf37c7d3d637071fdcb73302dedd7cf61162ce07` |
| `transformer/diffusion_pytorch_model.safetensors` | 7,751,109,744 | 7.751110 | `9f29f9edcfdae452a653ffb51a534ca4decd389952c225724ff3b94042612a6e` |
| `vae/diffusion_pytorch_model.safetensors` | 168,120,878 | 0.168121 | `ca70d2202afe6415bdbcb8793ba8cd99fd159cfe6192381504d6c4d3036e0f04` |

**Toplam ağırlık: 15,964,212,614 bayt = 15.964213 GB.** Seçilen tokenizer/config/index/README/lisans yardımcı dosyaları yaklaşık 15.939 MB ekler; manifestte ayrı listelidir.

### black-forest-labs/FLUX.1-schnell

Revision: `741f7c3ce8b383c54771c7003378a50191e9efe9`

| Dosya | Bayt | GB | SHA256 |
|---|---:|---:|---|
| `text_encoder/model.safetensors` | 246,144,352 | 0.246144 | `893d67a23f4693ed42cdab4cbad7fe3e727cf59609c40da28a46b5470f9ed082` |
| `text_encoder_2/model-00001-of-00002.safetensors` | 4,994,582,224 | 4.994582 | `ec87bffd1923e8b2774a6d240c922a41f6143081d52cf83b8fe39e9d838c893e` |
| `text_encoder_2/model-00002-of-00002.safetensors` | 4,530,066,360 | 4.530066 | `a5640855b301fcdbceddfa90ae8066cd9414aff020552a201a255ecf2059da00` |
| `transformer/diffusion_pytorch_model-00001-of-00003.safetensors` | 9,962,580,296 | 9.962580 | `9b633dbe87316385c5b1c262bd4b5a01e3d955170661d63dcec8a01e89c0d820` |
| `transformer/diffusion_pytorch_model-00002-of-00003.safetensors` | 9,949,328,904 | 9.949329 | `58b4434078f0c2567ddc54e3b5cbf39626ab55fbd9d5c22956e183668f535dec` |
| `transformer/diffusion_pytorch_model-00003-of-00003.safetensors` | 3,870,584,832 | 3.870585 | `e2cbc25471ed5186e69a9b51098300cb2f612556453e38a372c851a220ed238d` |
| `vae/diffusion_pytorch_model.safetensors` | 167,666,902 | 0.167667 | `f5b59a26851551b67ae1fe58d32e76486e1e812def4696a4bea97f16604d40a3` |

**Toplam ağırlık: 33,720,953,870 bayt = 33.720954 GB.** Seçilen tokenizer/config/index/README/lisans yardımcı dosyaları yaklaşık 4.973 MB ekler; manifestte ayrı listelidir.

Her snapshot için `model_index.json`, `scheduler/*`, `transformer/config.json`, `vae/config.json`, encoder `config.json` ve shard index JSON'ları; ilgili tokenizer dizinlerinin bütün gerekli dosyaları da alınır. Klein'de tek encoder, schnell'de iki encoder vardır. `from_pretrained` çalışsın diye sadece safetensors indirmek yetmez.

İki modelin ağırlık toplamı **49,685 GB**; küçük dosyalar ve lisanslar buna eklenir. Klein monolitik `flux-2-klein-4b.safetensors`, schnell monolitik `flux1-schnell.safetensors` ve kökteki `ae.safetensors` **bu diffusers paketine ayrıca alınmaz**. Mevcut chat ağırlıkları yaklaşık **256,817 GB**; bunlar silinmez veya dönüştürülmez.

## 4. GCS staging ve soğuk açılış etkisi — yalnız öneri

### 4.1 Onay sonrası düşünülen yerleşim

```text
gs://p27082-ds-istek/media-models/
  black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/
  black-forest-labs/FLUX.1-schnell/741f7c3ce8b383c54771c7003378a50191e9efe9/
```

Her revision altında seçilen diffusers ağaç yapısı + `IMAGE_WEIGHT_MANIFEST.json`, lisanslar/provenance, doğrulama sonucu. Prefixler **oluşturulmadı**. Bucket'ın mevcut model objelerine, ACL/IAM, lifecycle veya lokasyon ayarlarına dokunulmaz. Yeni public obje/URL yok. Yerel indirme ve GCS yükleme/depolama bedeli açık onayla ele alınmalı; önceki görsel model kararı upload onayı değildir.

Önerilen güvenli staging: izinli HF kaynağından exact revision + allowlist → staging alanında her dosya boyutu/SHA256 doğrulama → değişmez revision prefix'ine yükleme → yayın manifestini en son yazma. Kısmi dosya/eksik checksum geçerli model sayılmaz; mevcut prefix overwrite/delete edilmez. Gate/erişim veya hash uyuşmazlığı varsa durulur. İlk staging'in HF indirme süresi bu boot hesabından ayrıdır.

Boot sonrası yerel hedef örneği `/mnt/nvme/media-models/<repo>/<revision>/`. Mevcut chat boot kopyaları değiştirilmeden kalır. Görsel talebi yokken yalnız chat açılır; **medya kopyası/load'u chat readiness kritik yoluna eklenmez**. Ana motor istenince klein kopyalanır/yüklenir; fallback snapshot ancak açık, tanımlanmış fallback koşulunda istenir. Her request'te HF internet indirmesi yok; yükleme yalnız doğrulanmış yerel snapshot'tan ve `local_files_only` mantığıyla yapılmalı.

### 4.2 Mevcut transfer ölçümünden türetilen ek süre

Yerel eski boot arşivi: 256,817 GB paralel transfer yaklaşık195 sn → etkin toplu **1,317 GB/sn**; en büyük tek dosya yaklaşık **0,57 GB/sn**. Yeni ağ ölçümü yapılmadı. Ek kopya aralığı = seçilen GB / (1,317…0,57); küçük dosya overhead'i ayrıca olabilir.

| Durum | Ek GCS kopya | Media load / ilk örnek rezervi | Mevcut432 sn chat boot sonrası ilk medya-ready tahmini |
|---|---:|---:|---:|
| Yalnız klein | **12–28 sn** | **20–120 sn**, ölçülmemiş plan payı | **464–580 sn** (~7:44–9:40) |
| Yalnız schnell | **26–59 sn** | **20–120 sn**, offload daha uzun olabilir | **478–611 sn** (~7:58–10:11) |
| İkisini aynı boot'ta tamamen kopyalama | **38–87 sn** | İki load ayrı; önerilen normal yol değil | Chat boot'una otomatik eklenmez; fallback gecikmesi ayrıca gösterilir |

Bunlar boşta chat drain ve cache-restore varsayımıyla yaklaşık değerlerdir; aktif chat'in bitmesini bekleme/compile süreleri dahil değildir. Son boot soğuk açılışı **MEASURED-LOCAL=432 sn**; yeni görsel süreleri **SOURCED ESTIMATE**, yerel test sonucu değil. Aynı boot içinde NVMe snapshot durduğu için sonraki swap'ta tekrar GCS kopyası gerekmez. Discard kapanışından sonra tekrar gerekir.

Mevcut boot timeout600 sn üstüne sessizce çıkılmaz. Klein'in yavaş ucu600'e yakındır; schnell yavaş ucu aşabilir. Chat boot lease'i ve sonradan görsel load deadline'ı aynı şeymiş gibi ele alınmamalı; medya load süresince bounded iş lease'i korunmalı. 15 dk ürün toleransı konuşulmuş olsa da900 sn onaylı policy değildir. İlk pilotta torch.compile/TensorRT derlemesi veya kuantizasyon eklenmez; gerekirse ayrı ölçülür.

## 5. 40 GB A100 üzerinde sequential yükleme ve chat önceliği

**Birlikte resident üretim yapılmaz.** Klein için tam BF16 pipeline/aktivasyon planı yaklaşık **20–24 GB**; resmi “~13 GB” optimize/offload senaryosunu tam-resident garanti saymıyoruz. Schnell tam BF16 pipeline kaynakta **~35–37 GB**, 40 GB'da sınırda; batch1 + gerekirse encoder CPU offload / VAE tiling gerekir. Chat'in kaydedilmiş GPU peak'i yok; arşiv host RSS yaklaşık59 GiB. **Host RAM, VRAM değildir**; CPU offload sırasında85 GB host RAM de izlenmeli.

### 5.1 Tek otorite, tek CUDA sahibi

Önerilen durumlar: `CHAT_READY → CHAT_DRAINING → IMAGE_LOADING → IMAGE_READY/RUNNING → IMAGE_DRAINING → CHAT_RELOADING → CHAT_READY`.

1. **GPU-ownership mutex:** worker supervisor sahipliği tutar; mode switch, chat ve image lease'leri aynı kilit/otorite altında. İki process'in aynı anda CUDA load başlatmasına izin verilmez. Kilit sahibi PID/boot-id/fencing ve crash recovery gerekir; yalnız dosya varlığı yeterli değil.
2. Görsel grubu başlamadan chat kuyruğu kontrol edilir. Aktif chat rezervasyonları **tamamlanana kadar drain**; zorla müşteri yanıtı kesilmez. Yeni chat talebi medya henüz başlamadıysa önceliklidir.
3. Chat servis/process'i kontrollü durdurulur; systemd'nin kendiliğinden yeniden başlatıp GPU'yu kapmasını önleyen supervisor tasarımı gerekir. ds-tunnel'ın ds-serve bağımlılığı dikkate alınır; mevcut restart veya tunnel birimleri bu plan sırasında değiştirilmedi.
4. Ayrı **media process + venv** doğrulanmış yerel snapshot'tan tek pipeline yükler; BF16, batch1. Chat'e ait env/venv değişmez. Health/readiness “chat intentionally drained” durumunu genel VM arızası diye yorumlamamalı.
5. Bir swap'ta birkaç image request işlenerek yükleme maliyeti amorti edilir. Günlük kapasite hesabında **hedef grup50 görsel**, toplam4 gidiş-dönüş varsayıldı; bu50 görsellik bir CUDA batch veya kesilemez rezervasyon değildir. Her request sonunda chat kuyruğu kontrol edilir; chat varsa grup erken kesilir. n=4 request en çok4 **seri** görsel üretir.
6. Chat gelince en geç mevcut kısa request sınırında medya durur. **Chat priority anlık/bedelsiz preemption değildir:** mevcut request ve chat reload gecikmesi olur. US chat SLO'su bunu kabul etmiyorsa grup boyutu düşürülür; bu da günlük kapasiteyi azaltır. Sürekli chat varken200 görsel/gün garanti edilmez.
7. Media process kontrollü sonlandırılır; süre aşarsa yalnız o process/process-group için kontrollü kill fallback tasarlanır. CUDA context'in kalktığı NVML ile doğrulanır. `empty_cache()` başka process'in VRAM'ini temizlemez. Chat ile image process'i aynı anda yüklenmez.
8. Chat, aynı boot'ta mevcut **yerel NVMe chat pack'inden** yeniden yüklenir; GCS'den256 GB baştan çekilmez. Chat health/model-ready ve kısa sentetik chat smoke testinden sonra mode `CHAT_READY`; lease/mutex doğru sırayla bırakılır. Restore başarısızsa yeni görsel kabulü durur, rollback/operatör uyarısı; yeni boot otomatik yok.

**Swap süresi UNKNOWN.** İlk bütçe için bir tam `chat→image→chat` çevrimine **60–180 sn**, baz **120 sn** plan rezervi ayrılır. Bu iki yönü içeren toplamdır; görsel inference ve cold GCS kopyası ayrıca sayılır, yükleme iki defa tekrar eklenmez. Çok büyük Strata pack/cache nedeniyle gerçek süre daha uzun olabilir.

Ölçüm: monoton saatle drain bekleme, chat stop, VRAM boşalma, image load, ilk örnek warmup, image terminate/VRAM serbest, chat local reload ve chat-ready ayrı span. İlk cache-cold swap ile aynı boot'taki tekrarlar ayrı raporlanır; p50/p95 için küçük ilk testin örnek sayısı yetmez, p95 garantisi verilmez. Host RSS, page fault, NVMe okuma ve NVML memory.used ile reload darboğazı bulunur.

## 6. US `/v1/image` sözleşmesine uyum

- İlk kapsam **text-to-image**; img2img/editing, LoRA, kullanıcı checkpoint'i, video ve arbitrary çözünürlük yok.
- `model` allowlist: ana klein / açık şekilde seçilen yedek schnell. `size` yalnız **1024×1024,1344×768,768×1344**; `n=1..4`. Boyutlar benzer piksel sayılı olsa da her biri ayrıca ölçülür. Sabit server adım/dtype; klein4-adım/guidance1, schnell4-adım/guidance0 başlangıç önerisi. Prompt token sınırı modelin pinned config'iyle doğrulanır; Türkçe kalite testi gerekir, gizli dış çeviri API'si yok.
- **Kısa senkron iş uygundur:** slot ve image pipeline hazırsa200 yanıtında çıktı referansları; n=4 seri, gerçek batch4 değil. Video benzeri ayrı async altyapı ilk kapsam için zorunlu değildir.
- GPU sıcak ama chat yüklüyse image **warm değildir**: load/swap gerekir. Queued/cold/load gecikmesi kullanıcıya açıkça gösterilmeli. Uzun kuyrukta kısa endpoint'i belirsizce açık tutmak yerine US mevcut admission/Retry-After veya kuyruk protokolünü belirler; async'e geçilecekse yeni ürün kararıdır. Client timeout'unun ardından kör retry yeni iş/boot/ücret üretmez.
- **Warm≤30 sn hedef:** image pipeline yüklü ve GPU slot hazırken, n=4 dahil kabul→encode→yanıt p95 ölçülmeli. Klein2–6 sn/görsel →8–24 sn inference; kalan6 sn encode/çıktı işi için dar olabilir. Kaynak sayılar hedefi makul yapar ama garanti etmez. Sadece sampler süresini≤30 gösterip API yanıtını dışarıda bırakmak doğru değildir. Kuyruk/cold/swap dahil≤30 sn taahhüdü bu mimariyle verilemez; US bunu ürün metninde ayırmalı.
- Tek idempotency key + reservation; gerçek motor/revision/seed ve usage kaydı. Fallback şeffaf: aynı id altında terminal hata/başarılı örnek sayısı biliniyorsa, bounded tek fallback; otomatik sonsuz tekrar veya ikinci kredi yok. Klein OOM durumunda daha büyük schnell'i “bellek kurtarıcısı” diye denemek yanlış; önce süreç temizliği, kalan runtime ve açık fallback kuralı gerekir.
- Reserve→run→encode/output tamamlanması→tek settlement; client disconnect/cancel/OOM için US refund/partial success kuralını netleştirir. Gerçek müşteri/lisans/ödeme verisi testte kullanılmaz.

## 7. GPU-saniye ve maliyet — SOURCED ESTIMATE

**Yeni yerel görsel ölçümü YOK.** Önceki ds-a100 raporundan yaklaşık0,035 USD/dk = **2,10 USD/saat =0,000583333 USD/sn**. Bu **OBSERVED APPROXIMATE RATE**, kesin SKU faturası veya sabit Spot fiyatı değil. Kaynak sitelerin ucuz rental tarifeleri kullanılmadı.

| Motor | Dayanak | SOURCED ESTIMATE / görsel | Yalnız inference maliyeti |
|---|---|---:|---:|
| klein4B,1024²,4-adım | Resmi hızlı4-adım modeli; önceki araştırmada indekslenen A100~1,216 sn topluluk ölçümünün README'si canlı doğrulanamadı, fiyatlama ankrajı yapılmadı | **2–6 GPU-sn**, baz **3 sn** | **0,00117–0,00350 USD**, baz **0,00175 USD** |
| schnell,1024²,4-adım | Yayınlanmış üçüncü taraf A10040GB SXM4 batch1 BF16 **2,11 sn** [B]; ds-a100 ölçümü değil | **2–5 GPU-sn**, kaynak baz **2,11 sn** | **0,00117–0,00292 USD**;2,11 sn için **0,00123 USD** |

Yatay/dikey çözünürlük için başlangıçta aynı geniş aralık ayrılır; farklı tepe/latency ölçülmeden kesin eşitlik yok. n=4 inference yaklaşık4 kat; ortak prompt encoding optimizasyonu pilotta varsayılmaz.

Kredi hesabında ayrı kalemler: inference; slot/load/swap; cold copy/boot; idle tail; encode/output; başarısızlık ve Spot retry. **120 sn swap =0,07 USD/grup**:50 görsele yayılırsa+0,0014 USD/görsel; her request'te swap yapılırsa ucuz görüntü çıkarımı bile pahalı/çok yavaş olur. **432 sn cold=~0,252 USD**,600 sn idle=~0,350 USD. Depolama/egress/Cloud Run/SQL/vergi ayrıca; burada kredi fiyatı veya kâr marjı belirlenmedi.

## 8. Günlük200 görsel, mevcut3600/1800 sınırına sığar mı?

**Evet, toplu ve sınırlı swap varsayımında makul biçimde sığması beklenir; bu koşullu kapasite planıdır, chat dahil günlük garanti değildir.**200 görsel=200üretilen dosya;200 istek×4=800 görsel değildir.200 görselin gün içine tek tek yayılıp her seferinde boot/swap yaptırılması sığmaz.

Konservatif normal oturum örneği: **günde2 oturum, her birinde100 görsel; oturum başına2 medya grubu, yani toplam4 tam swap.** Her oturum için600 sn tam idle tail de ayrıldı; hard uptime drain daha erken kapatırsa daha az kullanılabilir. Günlük chat bütçesi aynı3600 sn havuzundadır.

| Kalem | Klein günlük baz | Schnell günlük baz | Kanıt niteliği |
|---|---:|---:|---|
|200 görsel inference |200×3=**600 sn** |200×2,11=**422sn** | SOURCED ESTIMATE / yayınlanmış kaynak baz |
|4 tam swap çevrimi |4×120=**480sn** |**480sn** | UNKNOWN için plan rezervi |
|2cold chat boot |2×432=**864sn** |**864sn** | Eski MEASURED-LOCAL, geleceğe garanti değil |
|2GCS media copy |2×28=**56sn** |2×59=**118sn** | Eski throughput'tan SOURCED ESTIMATE |
|2idle tail |2×600=**1200sn** |**1200sn** | Mevcut policy; dolu tail bütçesi |
|Encode/yanıt/output |200×0,5=**100sn** |**100sn** | Ölçülmemiş kaba plan payı; storage gecikmesi uzatabilir |
|Stop/drain güvenlik payı |2×30=**60sn** |**60sn** | Plan rezervi; gerçek stop gecikmesi daha uzun olabilir |
|**Toplam** |**3360sn=56dk** |**3244sn≈54,1dk** | **3600 sn altında** |
|**Oturum başına** |**1680sn=28dk** |**1622sn≈27dk** | **1800 sn altında** |
|**Günlük kalan pay** |**240sn** |**356sn** | Chat/ek retry için çok küçük |

Yaklaşık toplam VM zaman maliyeti baz klein **1,96USD/gün**, schnell **1,89USD/gün**; tüm3600 sn kullanılırsa yaklaşık **2,10USD/gün**. Bunlar chat/ek kaynak maliyetini içeren yeni fatura teklifi değildir; tabloya ayrılmış VM sürelerinin fiyatıdır.

**Duyarlılık ve admission:**
- Klein inference üst6sn olursa+600 sn → **3960sn**, sığmaz. Swap180 sn olursa baz toplam+240sn → **3600 sn**, hiç chat/retry marjı kalmaz. Her iki yavaşlık birlikte olursa kota azaltılır; limit otomatik yükseltilmez.
- 4 yerine 8 tam swap çevrimi (örn.25 görsellik gruplar veya sık chat kesmesi),120 sn bazda+480sn → klein **3840sn**; sığmaz. Chat önceliği grup erken kesiyorsa kapasite hesabı gerçek swap sayısıyla güncellenmeli.
- Image load maliyeti120 sn swap içinde sayıldı, tekrar eklenmedi. İlk warmup/yeniden derleme bu rezervi aşarsa tablo geçersizdir. Output0,5sn/görsel ve stop30sn yalnız rezerv; önceki BOOT5 auditinde stop operasyonunun tamamlanması daha uzun sürebilmişti. Gerçek muhasebe/süre ve kapanış gecikmesi ölçülmeli, marj daraltılmamalı.
- Her yeni grupta `kalan günlük runtime`, `kalan max_uptime`, mevcut chat rezervasyonları ve **reload+output+kapanış payı** hesaplanır. Kalan süre yetmiyorsa admission durur/ertelenir;200hedefi uğruna iş süresi aşılamaz. Kontrol mevcut SQL lifecycle günlük sayacını kullanır; sayaç sıfırlanmaz.
- İkioturum toplamı chat için yalnız birkaçdk bırakır. Yoğun chat varsa görsel kotası düşürülür veya talep başka güne kalır. Bu dar image-only plan için şimdilik **policy artışı önerilmiyor**; önce gerçek ölçüm. Boot sayısı limiti/günlük kalan açılışlar ayrıca kontrol edilmeli.

## 9. Kısa image işi için idle.py / watchdog busy lease

**Senkron kısa iş async video kadar uzun sürmez; bu, busy-lease gerekmiyor demek değildir.** HTTP açıkken mevcut80/8080 socket kontrolü yardımcı olabilir ama image process başka portta çalışabilir, socket kopabilir ve swap/load müşteri socket'i olmadan sürebilir.

- **Yerel idle.py:** güvenilir supervisor'dan boot-id/PID/fencing/iş-id + kısa TTL/progress heartbeat ile `loading/running/encoding/chat_reloading` busy bilgisi. Örn.20sn heartbeat; stale dosya sonsuz busy sayılmaz. Sadece terminal outcome ve güvenli chat restore sonrası release. Ready bilgisi yalnız chat health'ine bağımlı kalmaz; intentional mode switch arıza sayılmaz.600 sn idle eşiği değiştirilmez.
- **Watchdog:** image reservation aynı shared store/lifecycle authorization üzerinden acquire→heartbeat→release ile sürer. HTTP200/202 dönmesi değil, iş/çıktı tamamlanması ve GPU owner'ın düzgün bırakılması esas alınır. Mevcut60sn lease'e örn.20sn heartbeat; absolute deadline bounded. Çok kısa image için tek senkron request içinde bu yapılabilir, yeni async job service şart değil.
- **Idle refresh kaynağı:** gerçek kabul edilmiş iş/progress. `/health`, `/ready`, model listesi ve salt durum sorgusu GPU'yu uyandırmaz/idle'ı tazelemez. Mevcut geniş nginx `GET /v1/` aktivite eşleşmesi salt kontrolü iş saymamalı; kontrol-plane yetkili state'i kullanılmalı.
- `max_uptime1800`/`daily_runtime3600` aktif lease ile sonsuza dek ötelenmez. Admission; image grup süresi+chat reload+stop marjını önceden ayırır. Crash/OOM/client disconnect sonrası TTL/terminal kayıt ve tek kredi sonucu tutarlı olur. Timer sürekli kapatılmaz; gelecekte test için kapatma gerekirse ayrıca izin.
- SRE'nin canlı `sre-ready` tabanındaki /ready, rate-limit ve JSON log yamaları korunmalı. Değişiklikler tek incelenmiş paket/rollback ile sunulur; bu plan kod yazmadı veya deploy etmedi.

## 10. Gelecekte onaylanmış tek test boot'u için sıra

**Şimdi test yok.** BOOT6 onay gelene kadar bekler; sonraki testin BOOT6 olup olmadığı da Demir'in kararıdır. Kalan3hak içinde kendiliğinden ek açılış veya retry yapılmaz.

1. **GPU açmadan:** US API/admission/fallback sözleşmesi; SRE supervisor/lease/mutex + CPU testleri; exact dependency lockfile; lisans ve manifest doğrulama planı. Staging/upload ancak ayrı açık onayla. Manuel GPU boot veya gerçek müşteri rezervasyonu yapılmaz.
2. Onay gelirse günlük runtime/starts bakiyesi, sentetik izolasyon, rollback ve ücret/süre üst sınırı teyit edilir. Öneri mevcutmax_uptime1800 sn aşılmadan bir sınırlı açılış; yaklaşık30dk×0,035=**1,05USD**, sağlayıcı enforce edilen cap değil. Bu öneri test onayı yerine geçmez.
3. GPU identity/memory.total ve mevcut chat VRAM/host RSS kaydı; chat hazırken kısa sentetik chat doğrulaması. Model weights/copy throughput ve boot432 sn referansı karşılaştırılır; health probe boot tetiklemez.
4. Chat drain→klein load→1024² n1 ve n4, yatay/dikey n1; mümkünse tekrar n4.4adım, BF16, batch1 sabit. İlk sonuç ile warmed sonuç ayrı; API uçtan uca süre, sampler süresi, text encoding, VAE decode, çıktı encoding/persist ayrı ölçülür. Prompt/secret içerikleri loga yazılmaz.
5. Peak VRAM (NVML+allocated/reserved), host RSS/pagefault/NVMe, GCS bayt/sn; n4 sırasında sürekli istek concurrency'si test edilmeden açılmaz. Klein kendi başına OOM ise fallback büyük olduğu için otomatik deneme yapılmaz.
6. İmajlar çalışırken bir sentetik chat admission: mevcut kısa request sınırında grup erken kesme; medya process'i terminate→VRAM serbest→chat'i yerel NVMe'den reload→health ve kısa chat yanıtı. Her span ve toplam **tam swap** ölçülür.
7. Kalan süre uygunsa, **önceden onaylı** schnell erişim/staging ile aynı sınırlı koşullar; tüm matrisi bitirmek için süre uzatılmaz, ikinci boot yok. Schnell n4 asla batch4 değildir. Teste yetişmeyen fallback matrisi “NOT TESTED” kalır.
8. Busy lease, client disconnect, küçük timeout/OOM simülasyonu ve health-readiness mantığı; gerçek uzun/hatalı iş başlatarak süre yakma yok. Crash/TTL cleanup testlerinin çoğu önceden box/CPU testlerinde yapılır. Bu adımlar idle timer'ı kapatma izni vermez.
9. Chat restore + durum/bütçe kaydı; onaylı kapanış/discard yolu ile **TERMINATED** teyidi. Yerel SSD silme/manual fallback kapsamı önceden açık olmalı; ilk geniş planın izni varsayılmaz. Boots/maliyet ve eksik testler raporlanır. Başarısız test yeni boot'a otomatik dönmez.

Küçük ilk örneklemden p95 kesinleşmez. Geçiş kapıları: üç boyut/n1/n4 warm≤30sn hedefi, güvenli chat restore, measured swap≤bütçe, hiçbir false idle stop, polling'in wake/idle etkisi olmaması ve 1800/3600 admission. Sağlanmazsa eşzamanlı residency denenmez; n/kota/grup/SLA yeniden değerlendirilir.

## 11. Demir'in kararını gerektiren açık noktalar

1. **Test boot:** BOOT6 bekletmesi kaldırılsın mı; kalan3test hakkından1'i bu image smoke için kullanılsın mı? Şu an hayır/beklemede.
2. **GPU test süresi ve tahmini harcama:** örn. mevcut1800 sn sınırında sınırlı test ve yaklaşık1,05USD; günlük kalan runtime daha azsa daha kısa. Hiçbir limit artırımı varsayılmıyor.
3. **Model staging ve çıktı saklama:** ağırlıkların mevcutbucket'a yazılması/depolama bedeli; image çıktılarının hangi yetkili dayanıklı prefix'te ve ne kadar süre tutulacağı. Çıktı yalnız localSSD'de kalırsa discard sonrası kaybolur. US'nin24 saat önerisi henüz storage/retention/silme izni değildir. “Silme yok” kuralı nedeniyle otomatik24 saat silme/lifecycle kurulmamalı; ayrı açık izin gerekir. Public ACL/link paylaşımı yok.
4. **Ürün işletimi:** chat önceliği nedeniyle image grupları kesildiğinde kuyruk/ETA; yüklenmiş warm ile swap/cold ayrımı; fallback görünürlüğü ve partial refund. Bunları US tasarlar, maddi kapsam/SLA değişimi Demir'e sunulur.

**Şimdiki sonuç:** seçilmiş iki açık model için paket/manifest ve sequential tasarım hazır. Toplu image-only200 görsel/gün mevcut limitlere **baz senaryoda sığar**; sık swap, yoğun chat veya yavaş uçta kota düşürülür. GPU kapalı kalır; canlı işlem başlatılmaz.

## 12. Kaynaklar ve yerel kanıtlar

8 Kasım 2026 TSİ araştırması. HF revision'ları metadata'dan sabitlendi; lisans/provenance sayfaları bu turda resmi kaynaklardan okundu. HF gated dosyalara erişim açılmadı.

- [G] Google A10040/80GB tipleri: https://docs.cloud.google.com/compute/docs/gpus
- [K] Pinned klein kartı: https://huggingface.co/black-forest-labs/FLUX.2-klein-4B/resolve/e7b7dc27f91deacad38e78976d1f2b499d76a294/README.md
- [KL] Pinned klein LICENSE: https://huggingface.co/black-forest-labs/FLUX.2-klein-4B/resolve/e7b7dc27f91deacad38e78976d1f2b499d76a294/LICENSE.md
- [S] Resmi schnell kartı: https://huggingface.co/black-forest-labs/FLUX.1-schnell ; pinned manifest API: https://huggingface.co/api/models/black-forest-labs/FLUX.1-schnell/revision/741f7c3ce8b383c54771c7003378a50191e9efe9?blobs=true
- [SL] Pinned resmi schnell MODEL lisansı: https://raw.githubusercontent.com/black-forest-labs/flux/802fb4713906133fcbd0d8dc5351620ca4773036/model_licenses/LICENSE-FLUX1-schnell
- [V1] Resmi FLUX.1 VAE Apache açıklaması: https://raw.githubusercontent.com/black-forest-labs/flux/802fb4713906133fcbd0d8dc5351620ca4773036/README.md
- [V2] Resmi FLUX.2 VAE Apache açıklaması: https://raw.githubusercontent.com/black-forest-labs/flux2/50fe5162777813d869182b139e83b10743caef15/README.md
- [Q] Pinned Qwen upstream LICENSE: https://huggingface.co/Qwen/Qwen3-4B/resolve/1cfa9a7208912126459214e8b04321603b3df60c/LICENSE
- [T] T5v1.1: https://huggingface.co/google/t5-v1_1-xxl ; resmi pinned Apache LICENSE: https://raw.githubusercontent.com/google-research/text-to-text-transfer-transformer/42909aa15d1fbd16c69333aa474a00ae8ed9fde4/LICENSE
- [C] CLIP kartı: https://huggingface.co/openai/clip-vit-large-patch14 ; resmi pinned MIT LICENSE: https://raw.githubusercontent.com/openai/CLIP/d05afc436d78f1c48dc0dbf8e5980a9d471f35f6/LICENSE
- BFL encoder eşleştirmesi: https://raw.githubusercontent.com/black-forest-labs/flux/main/src/flux/util.py (`load_t5`, `load_clip`; kod okunmuş, çalıştırılmamıştır).
- [B] A10040GB schnell warm benchmark: https://gpubattle.com/guides/flux-1-schnell-gpu-benchmarks
- Yerel manifest: `image_research/IMAGE_WEIGHT_MANIFEST.json`; lisans kaynakları/hash'leri: `image_research/official-license-sources.json`; küçük lisans/README kopyaları aynı dizinde.
- Ana/fallback metadata ve revision kanıtı: `media_research/black-forest-labs__FLUX.2-klein-4B.json`, `media_research/black-forest-labs__FLUX.1-schnell.json`; upstream provenance metadata: `media_research/*-provenance.json`.
- Chat boot/transfer kanıtı: `KAN_S4_LIVE_RESULT_BOOT5.md`, `SRE_BOOT5_VERIFY.md`; `/workspace/lt/coldstart/boot2_detail.txt`, `patches/fetch-weights.py`. Mevcut idle/lifecycle box kopyaları önceki medya planında listeli.

**Dokunulmayan önceki medya planının SHA256:** `d587eeabf145e7ddee6060ad04caa5e0f48017e923b41358b829ddb9bd392c1b`.
