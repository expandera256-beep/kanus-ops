# KAN — ds-a100 üzerinde açık görsel/video motorları: YALNIZ PLAN

**Tarih:** 8 Kasım 2026, TSİ (Europe/Istanbul, UTC+3). Dayanak: Demir'in 01:17 TSİ kararı; BOOT6 için 01:13 TSİ'deki bekletme kararı aynen sürer.

## 1. Karar özeti ve uygulama sınırı

- **Önemli düzeltme: mevcut GPU 80 GB değil, tek A100 40 GB.** Salt okunur `gcloud compute instances describe` sonucu `a2-highgpu-1g`, `nvidia-tesla-a100`, count=1, **TERMINATED**. Google'ın resmi tablosunda bu tip 40 GB GPU / 85 GB sistem RAM'idir. `a2-ultragpu-1g` ayrı tiptir: 80 GB GPU / 170 GB RAM. Bu rapor donanım yükseltmesi önermeyi çalıştırma izni saymaz. [G1]
- İlk görsel pilotu: **FLUX.2 klein 4B** (Apache-2.0, daha küçük bellek), karşılaştırma/yedek **FLUX.1 schnell** (Apache-2.0); hafif ve olgun alternatif **SDXL base**. Bunlar öneridir, model seçimi henüz onaylı değildir.
- İlk video pilotu: **Wan2.1 T2V-1.3B / 480p**; 720p ve isteğe bağlı I2V için **Wan2.2 TI2V-5B**. İkisi de Apache-2.0. **LTX 2B 0.9.8 distilled** hız alternatifi, fakat koşullu ticari lisansı nedeniyle ayrı uygunluk kontrolü ister.
- **Sohbet ile eşzamanlı üretim değil, tek GPU üzerinde sırayla yükleme/boşaltma** önerilir. Sohbetin mevcut VRAM tepe değeri eldeki arşivde yoktur; birlikte sığdığı iddia edilmez. 85 GB host RAM de ikinci darboğazdır.
- **200 görsel + 30 video/gün**, Wan tabanlı normal pilotta mevcut **daily_runtime=3600 sn / max_uptime=1800 sn** sınırına sığmaz. Hızlı LTX senaryosu yalnız düşük süre, toplu çalışma ve çok az sohbet altında sığabilir; garanti değildir (§8).
- Bu çalışma sırasında GPU **açılmadı**, GCP/bucket/VM/policy/timer/deploy değiştirilmedi; model ağırlığı indirilmedi veya GCS'ye yazılmadı. Yalnız VM kimliği okundu, yerel kayıtlar ve kamusal kaynaklar incelendi. Box'ta bu plan ve küçük kamusal metadata/benchmark kanıtları oluşturuldu.
- Kalan test hakkı **8'den 3**; **BOOT6 bekliyor**. Hiçbir açılış, fixture, timer değişikliği veya ücretli yeni kaynak bu raporla yetkilendirilmez.

## 2. Kanıt sınıfları — US'nin kredi hesabında karıştırılmamalı

| Etiket | Anlamı |
|---|---|
| **MEASURED-LOCAL / YEREL ÖLÇÜM** | Önceki ds-a100 test kayıtları: sohbet soğuk açılışı 432 sn; kopyalanan bayt ve süreler. Bu araştırmada yeni GPU ölçümü yapılmadı. |
| **OBSERVED-RATE / GÖZLENEN YAKLAŞIK TARİFE** | Önceki test raporundan yaklaşık **0,035 USD/dk = 2,10 USD/saat = 0,000583333 USD/sn**. Kesin SKU faturası veya gelecekte sabit Spot fiyatı değildir. |
| **PUBLISHED-BENCHMARK / YAYINLANMIŞ ÖLÇÜM** | Kaynağın kendi GPU'sundaki sonuç; **ds-a100 üzerinde ölçülmüş değildir**. A100 80 GB, A800, RTX4090 sonuçları 40 GB sonucuyla eşitlenmez. |
| **SOURCED ESTIMATE / KAYNAKLI TAHMİN** | Bu planın, kaynak ölçümlerinden çıkardığı süre, bellek, aktarım ve maliyet hesabı; SLA veya kredi tarifesi değil. |
| **UNKNOWN / DOĞRULANMADI** | Mevcut sohbet VRAM tepesi, model swap p95, pilot kalite oranı, 10 sn video davranışı. Onaylı ölçüm gerekir. |

**Hesap:** örnek marjinal maliyeti = üretim GPU-saniyesi × 0,035/60. Soğuk açılış, model yükleme/swap, boşta bekleme, başarısız üretim, encoding/upload ve sohbet zamanı ayrıca ücretlidir. Saatlik 2,10 USD yalnız yukarıdaki gözlenen yaklaşık oranın türevidir; web benchmark sitelerinin ucuz kiralık GPU fiyatları **kullanılmadı**. Vergi, SQL/Cloud Run, depolama, işlem ve egress dahil değildir.

## 3. Mevcut sohbet/boot ve aktarım tabanı

### 3.1 Eldeki yapılandırma

`/workspace/us-panel/reports/r20261108/KAN_S4_LIVE_RESULT_BOOT5.md` mevcut servis modelini **deepseek-v4.1-flash** ve BOOT5 model-ready süresini **7 dk 12 sn** olarak doğruluyor. `/workspace/us-panel` altında gerçek `/opt/ds/boot.sh` ve `/opt/ds/serve-forever.sh` dosyalarının tam güncel kopyası bulunamadı; kapalı VM'de SSH veya servis sorgusu yapılmadı. Daha ayrıntılı son yerel arşiv `/workspace/lt/coldstart` altında:

- `inspect1a.txt`: `ds-boot.service → /opt/ds/boot.sh`; `ds-serve.service → /opt/ds/serve-forever.sh`, boot birimine bağımlı.
- `boot2_detail.txt`: `/opt/ds/build-dsv41/strata-dsv41 --serve --pack /mnt/nvme/strata/packs/deepseek-v4.1-flash-ist-das-2.00bpw --max-context 8192 --eos 1 --vision off --dspark-conf 0.65 --dspark auto --experts 6`.
- `patches/fetch-weights.py`: IQ2_XXS ana ağırlıklar **138,239 GB**, Engram **110,595 GB**, dspark **7,977 GB**, paylaşılan ek GGUF **0,0058 GB**; toplam **256.816.878.701 bayt = 256,817 GB = 239,179 GiB**. Bu SSD dosya boyutudur, GPU'ya aynı anda yüklendiği anlamına gelmez.
- Arşiv process RSS **62.041.688 KiB ≈59,17 GiB host RAM**. **Bu VRAM değildir.** Kaydedilmiş `nvidia-smi memory.used` / GPU buffer tepesi bulunmadı. Çalışan sohbet için 0/20/30 GB gibi bir VRAM değeri uydurulmadı; bilinmiyor. İlk onaylı ölçümde süreç bazlı GPU bellek, context/KV ve host RSS ölçülmeli.
- 4 yerel SCRATCH SSD, RAID0 `/mnt/nvme`; mevcut discard kapanışında ağırlıklar/yerel cache kaybolur. 80 GB'a geçiş veya kalıcı disk bu planın uygulanmış bir parçası değildir.

### 3.2 Soğuk açılışa ek süre nasıl hesaplandı?

`boot2_detail.txt` 7 paralel dosyanın sonuncusunun **195,0 sn** içinde bittiğini gösterir: toplam / son bitiş ≈ **1,317 GB/sn etkin toplu throughput**. En büyük tek dosya 110,595 GB /195 ≈ **0,567 GB/sn**; önceki diğer koşuda 188,4 sn. Başlangıç farkları nedeniyle bunlar yaklaşık aktarım göstergeleridir, BOOT5'in 432 saniyesinin tamamı transfer değildir.

Tablolarda ek kopyalama = **dosya GB / (1,317 … 0,57 GB/sn)**. Bu bir ağ garantisi veya ilave tüm modelleri eşzamanlı indirirken aynı hızı koruma iddiası değildir; küçük dosya/kimlik doğrulama overhead'i ve yükleme/ilk kernel dahil değildir. Kaynak HF→GCS ilk staging süresi dahil değil; GCS'de medya ağırlıkları varmış gibi davranılmadı.

**Tahmini ilk medya-ready:** mevcut 432 sn + seçilen modelin ek kopyası + **20–120 sn yükleme/ilk örnek rezervi** (küçük modeller için ölçülmemiş plan payı). Örnek: klein ~464–580 sn; schnell ~478–611 sn; Wan 5B ~478–612 sn. Sohbeti baştan yükleyip sonra boşaltan mevcut yolun korunması varsayımıdır. Derleme/kuantizasyon/yoğun offload daha uzun sürebilir. Kritik boot yoluna tüm adaylar eklenmemeli; sadece istenen motor lazy-load edilmelidir.

US'nin “15 dk cold warmup” varsayımı **onaylı policy değildir**: eldeki WorkerPolicy boot timeout üst doğrulaması 600 sn'dir. 900 sn ürün toleransı mevcut 600 sn kontrol sınırını kendiliğinden yükseltmez. Öncelik <600 sn ready; 900 sn gerekiyorsa timeout/validasyon değişikliği ayrıca incelenip onaylanmalı.

## 4. Resmi lisans kontrolü

“Açık ağırlık” ile koşulsuz açık kaynak aynı şey değildir. Kod lisansı, ana ağırlık lisansı ve text encoder/LoRA/upscaler lisansları ayrı tutulmalı. Aşağıdaki metinler bu araştırmada açıldı; lisans kabulü, kayıt veya sözleşme yapılmadı. Bu teknik değerlendirme hukuki görüş değildir.

| Aday / sürüm | Ticari self-host durumu ve dikkat noktası | Resmi kaynak |
|---|---|---|
| **FLUX.2 klein 4B distilled** | **Apache-2.0**, ticari kullanım mümkün. 9B ve dev varyantına genellenmez; 9B non-commercial. NOTICE/attribution ve bileşen lisansları korunur. | [F2], [F2L] |
| **FLUX.1 schnell** | **Apache-2.0**; kişisel/bilimsel/ticari kullanım model kartında açıkça yazılı. | [F1] |
| **FLUX.1 dev** | **ELENDİ:** Non-Commercial License; ayrı anlaşma olmadan ticari üretim servisi için uygun değil. Çıktının ticari kullanımına izin verilmesi, modeli ücretli üründe çalıştırmaya izin demek değildir. | [FD] |
| **SDXL 1.0 base** | **CreativeML Open RAIL++-M**: ticari/SaaS mümkün; Attachment A kullanım yasakları ve son kullanıcı sözleşmesine aktarılacak kısıtlar var. Apache değil; SD3.5'in gelir eşiği bu ayrı SDXL lisansına uygulanmaz. Kartın araştırma amaçlı niyet ifadesi de uygunluk değerlendirmesinde not edilmeli. | [SX], [SXL] |
| **SD3.5 Medium** | **KOŞULLU:** Stability Community License. <1 milyon USD toplam yıllık gelir için ücretsiz ticari kullanım; ticari kullanım için kayıt, lisans/NOTICE ve “Powered by Stability AI” görünürlüğü; bağlı kuruluş geliri de önemlidir. Eşik civarı/aşımında Enterprise lisans teyidi gerekir. T5, AUP ve başka temel model eğitimi kısıtları ayrıca okunmalı. | [S3], [S3L] |
| **Wan2.1 1.3B / 14B**, **Wan2.2 TI2V-5B / A14B** | **Apache-2.0** model ağırlıkları; ticari self-host için uygun ana adaylar. Model kartı kullanım sorumluluğu ve zarar/illegal içerik uyarılarını ayrıca koyuyor. | [W1], [W2] |
| **LTX-Video 2B/13B v0.9.6–0.9.8** | **KOŞULLU:** LTXV Open Weights License 0.X, Apache değil. Yıllık gelir **en az 10 milyon USD** olan işletme ücretli Commercial Use Agreement almalı. Daha küçük işletme ticari kullanabilir; dağıtım/kullanım kısıtları, izinsiz impersonation yasağı ve içeriğin makine üretimi olduğunun açıklanması gerekir. | [LTX], [LTXL] |
| **LTX-Video 2B 0.9.5** | Ayrı **Open RAIL-M** metni açıkça ticari kullanıma izin verir; kullanım kısıtları/AI açıklaması var. Yeni sürümün 10M koşulunu eski sürüme otomatik taşıma; eski benchmark'ı yeni checkpoint ölçümü sayma. Metindeki “Permitted Purpose” atfı gibi belirsizlikler hukuki kontrol konusu. | [LTX95L] |
| **CogVideoX-2B** | **Apache-2.0**, transformer ve VAE dahil; T5 bileşeni ayrıca kayıt altına alınır. | [C2] |
| **CogVideoX-5B** | **KOŞULLU**, 2B ile aynı lisans değil: kayıt + basic commercial license gerekir; aylık **1 milyon ziyaret** üstünde ilave ticari lisans görüşmesi gerekir. Kayıt yapılmadı; pilot önceliği değil. | [C5L] |
| **HunyuanVideo (ilk 13B sürüm)** | **BU ORTAM İÇİN ELENDİ:** Tencent lisansı AB, Birleşik Krallık ve Güney Kore'yi “Territory” dışında bırakıyor; bu yerlerde kullanım/çıktı dahil kapsamlı kısıtlar var. `europe-west4-b` Hollanda/AB olduğundan mevcut hostta kullanma. 100M MAU ek ticari eşik ve harici LLaVA/Llama text encoder lisansları da var. Başka Hunyuan sürümleri ayrıca incelenmeden bu hüküm değiştirilmez. | [HYL], [HY] |
| **Mochi-1-preview** | **Apache-2.0**; ticari kullanıma uygun ana lisans. “Research preview” kalite olgunluğu etiketi, kendi başına non-commercial lisans değildir. | [M1] |

Tüm adaylarda insan tasviri/izin, telif/marka, yasadışı cinsel içerik ve kişisel veri için ürün güvenlik kontrolleri gerekir. Apache lisansı eğitim verisi veya bütün çıktılar için hak garantisi sağlamaz.

## 5. Dosya boyutu, VRAM ve mevcut 40 GB'a sığma

**GB = 10^9 bayt** indirme boyutu; VRAM rakamları kaynağın GB gösterimiyle yaklaşık verilir. Aktivasyon/attention/VAE/KV ve allocator yüzünden dosya GB'sı VRAM tepesine eşit değildir. Batch=1; 4 görsel tek GPU batch=4 değil **4 seri örnek** olarak planlanır.

Boyutlar resmi HF `/api/models/<repo>?blobs=true` dosya manifestlerinden toplandı; aynı modelin monolitik/diffusers/FP16/FP32 kopyaları birlikte sayılmadı. Ham metadata ve revision SHA'ları `media_research/*.json` dosyalarında. Bunlar küçük metadata indirmeleridir; ağırlık indirilmedi.

| Aday | Seçilmiş ağırlık seti, GB (encoder/VAE dahil) | Tek motor VRAM / 40 GB durumu | 80 GB olsaydı | GCS ek kopya, sn |
|---|---:|---|---|---:|
| **FLUX.2 klein 4B** | **15,964**: transformer 7,751 + Qwen text encoder 8,045 + VAE 0,168 | Resmi ~13 GB optimize/offload iddiası; tüm BF16 dosyalar/aktivasyonlar için **~20–24 GB plan tahmini**. 40 GB'da tek motor makul. 13 GB'ı tam resident BF16 garantisi sayma. | Rahat tek motor; sohbetle birliktelik yine ölçüm ister. | **12–28** |
| **FLUX.1 schnell** | **33,721**: DiT 23,782 + T5 9,525 + CLIP 0,246 + VAE 0,168 | Yayınlanmış tam BF16 pipeline **~35–37 GB**; 40 GB'da sınırda ama A100 40 GB benchmark mevcut. VAE tiling / gerekirse text encoder CPU offload; batch1. Sohbetle birlikte hedeflenmez. | Tek motor rahat. | **26–59** |
| **SDXL base FP16** | **6,938** (refiner hariç) | **~8–12 GB SOURCED ESTIMATE**, uygulama/boyuta göre; 40 GB'da tek motor rahat. Refiner ayrı yük ve boyut getirir. | Rahat. | **5–12** |
| **SD3.5 Medium BF16** | **16,268** tam diffusers seti; T5 dahil | **~20–26 GB plan tahmini**, T5 offload ile daha az. İnternetteki “6,4 GB” tek denoiser/optimize tepeyi tam resident pipeline diye kullanma. 40 GB'da tek motor makul. | Rahat. | **12–29** |
| **Wan2.1 T2V-1.3B** | **17,546** resmi set: denoiser 5,676 (dosya dtype'ı daha büyük olabilir), UMT5 11,362, VAE 0,508 | Resmi **8,19 GB optimize yol**; tam resident başka diffusers ölçümlerinde ~20–21 GB. 40 GB'da tek motor uygun; T5 CPU seçeneği var. 480p önerilir. | Rahat. | **13–31** |
| **Wan2.1 T2V-14B** | **69,024** resmi checkpoint seti; çalışma dtype BF16'a çevrilince dosyadan daha küçük | Tam pipeline **~50+ GB kaynaklı tahmin**; 40 GB'da doğrudan önerilmez. Blok/CPU offload-kuantizasyon ayrı doğrulama ister. | 80 GB sınıfı makul, uzun süre riski var. | **52–121** |
| **Wan2.2 TI2V-5B** | **34,180** resmi set: denoiser 19,999, UMT5 11,362, VAE 2,819; sadece BF16 staging yapılırsa daha az olabilir, yapılmadı | Resmi offload yolu **24 GB sınıfı**, RTX4090 benchmark **22,9 GB**; kısa tam resident diffusers ~35,5 GB. 40 GB'da **T5 CPU + offload + VAE tiling** öner; 121 kare 720p tepe ölçülmeli. | Offload kaldırılarak hızlanabilir; 40 GB ile eşit varsayma. | **26–60** |
| **Wan2.2 T2V-A14B** | **126,178** resmi büyük set: iki expert + UMT5/VAE. “14B aktif” tüm modelin 14B olduğu anlamına gelmez. | Resmi A100/A800 tablo: **480p 41,3 GB**, **720p 59,8 GB** → 40 GB'a standart yolda **SIĞMAZ**. | Bu referans yollar sığabilir; süre çok uzun. | **96–221** |
| **LTX 2B 0.9.8 distilled** | **25,390**: monolitik 6,341 + yayımlanan F32 T5 19,049. T5 BF16-only paket önceden hazırlanırsa ~15,9 GB; henüz yapılmadı. | **~20–30 GB plan tahmini** (eski 2B benchmark ~20 GB); CPU encoder/tiling ile 40 GB'da makul. 0.9.8 için kesin peak yok. | Rahat tek motor. | **19–45**; BF16 paket olursa ~12–28 |
| **LTX 13B 0.9.8 distilled** | **47,628**: monolitik28,579 + F32 T5 19,049; upscaler isteğe bağlı +0,505 | **~40–50+ GB tam resident tahmin**; 40 GB için text encoder/katman offload gerekir, yeni kalite/hız ölçümü olmadan pilot değil. FP8 dosyası15,694 GB olsa da A100'de native FP8 hızlanması yoktur. | Tek motor daha uygun. | **36–84** |
| **CogVideoX-2B** | **13,775**: transformer3,388 + T5 9,525 + VAE0,862 | Resmi SAT FP16 **18 GB**; diffusers agresif offload4 GB'dan başlayan örnek var, aynı performans iddiası değil. 40 GB'da uygun. | Rahat. | **10–24** |
| **CogVideoX-5B** | **21,528** | Resmi SAT BF16 **26 GB**; diffusers offload daha düşük. 40 GB'da tek motor uygun; lisans koşulu var. | Rahat. | **16–38** |
| **Mochi-1 BF16** | **30,500**: BF16 DiT20,055 + BF16 VAE0,920 + FP16 T5 9,525 | Resmi native harness **~60 GB**; diffusers BF16 + offload/tiling **22 GB** örneği yalnız84 kare. 40 GB'da ancak optimize yol koşullu; 5 sn121/151 kare yeniden ölçüm ister. | Native harness mümkün. | **23–54** |
| **HunyuanVideo 13B** | **26,628 çekirdek**; harici LLaVA/Llama8B + CLIP ile **~43–45 GB toplam tahmin**, tam toplam doğrulanmadı | Resmi129 kare 720p **60 GB**, 544p **45 GB**; 40 GB standart yola uygun değil; **AB lisansı ayrıca engel**. | Teknik olarak80 GB sınıfı, hukuki engel çözülmez. | Çekirdek20–47; toplam~33–79 (tahmin) |

Bütün adayları bootstrap'a eklemek toplam boyutu/ilk yanıtı gereksiz artırır. Klein+Wan1.3B birlikte disk seti ~33,51 GB → ilave kopya ~25–59 sn; bu **eşzamanlı VRAM resident** demek değildir. Paylaşılan T5 adlı dosyalar aynı tokenizer/revision olmadığı sürece körlemesine tekilleştirilmez. GCS staging, checksum ve lisans kopyalarının saklanması daha sonraki onaylı uygulama adımıdır.

## 6. Üretim GPU-saniyesi ve maliyet — hiçbiri yeni yerel medya ölçümü değil

**Benchmarklar ısınmış batch1 içindir.** Kaynaklardaki CPU/compile/offload yöntemi değişince hız çok değişir. Model swap ve ilk yükleme aşağıdaki marjinal bedellere dahil değildir. Video “5 sn içerik” ile “5 GPU-sn” karıştırılmamalı.

| Motor / örnek | Yayınlanmış dayanak | ds-a100 için SOURCED ESTIMATE, GPU-sn/örnek | Yalnız üretim USD/örnek |
|---|---|---:|---:|
| **klein4B, 1024²,4 adım** | Resmi model kartı hızlı4-adım; indekslenen topluluk A100 raporu1,216 sn / compile0,896 sn, ancak canlı repo README'si bu araştırmada doğrulanamadı. Bu dar sayı fiyatlama dayanağı yapılmadı. | **2–6** ihtiyatlı plan aralığı; baz3 | **0,00117–0,00350**; baz0,00175 |
| **schnell,1024²,4 adım** | [BF1] A10040GB SXM4 **2,11 sn**, tam BF16 pipeline; bizim makine ölçümü değil | **2–5**; benchmark hesabı2,11 | **0,00117–0,00292**; 2,11 için **0,00123** |
| **SDXL,1024²,30 adım** | [BSX] resmi A100 baseline **3,704 sn**, TensorRT2,742 sn. Pipeline/version/refiner ayrıntıları aynı olmadan eşit performans vaat edilmez. | **4–10** ilk diffusers planı | **0,00233–0,00583** |
| **SD3.5 Medium,1024²,30–40 adım** | [BS3] resmi H100 TensorRT30 adım **2,497 sn**; **A100 değil** | **6–15**, düşük güvenli tahmin, A100 ölçümü bulunmadı | **0,00350–0,00875** |
| **Wan2.1 1.3B,832×480,~5 sn81 kare16fps** | [W1] RTX4090~240 sn; [BTW] **A800** baseline~175 sn / TeaCache88 sn. Arama özetinin175 sn'yi A100 diye göstermesi yanlıştı; ham kaynak A800 diyor. | **180–300** baz240; A10040 yerel ölçüm yok. TeaCache hızı kalite testi olmadan kullanılmaz. | **0,105–0,175**; baz **0,140** |
| **Wan2.2 5B,1280×704,~5 sn121 kare24fps** | [BW5] A10040 SXM4 **0,691 kare/sn**, PCIe0,544; deney **49 kare/30 adım**, 5 sn değil. [W2] resmi4090 5 sn T2V534,7sn /22,9GB | 121/kare-hızı → **175–223** lineer ölçekli **tahmin**; uzun attention/offload nedeniyle bütçeleme **200–540**. Baz350 ihtiyatlı. | Lineer **0,102–0,130**; ihtiyatlı **0,117–0,315**; baz **0,204** |
| **Wan2.2 A14B,480p/720p** | [W2T] resmi A100/A800 1GPU T2V **785,7/2735,7 sn**; I2V810/2810,9sn; **41,3/59,8GB** | Bu donanımda standart yol **NO-GO**. 80GB referans sonucu, 40GB tahmini sayılmaz. | Referans T2V **0,458/1,596**; kullanılacak fiyat değil |
| **LTX2B0.9.8 distilled,~5 sn480p121 kare24fps** | [BLTX] eski **0.9.5** 97kare704×48030 adım: A10040 SXM4 **8,422 kare/sn** (~11,52sn), PCIe7,67. 121kareye~14,4–15,8sn. Site “distilled” diyor ama ağırlık0.9.5; **0.9.8 ölçümü değil**. | **20–60** 0.9.8 için düşük güvenli plan payı; baz40. 720p için **40–180** çok kaba tahmin, kalite/upscaler maliyeti dahil değil. | 480p **0,0117–0,0350**; baz **0,0233**;720p0,0233–0,105 |
| **LTX13B0.9.8 distilled,~5 sn720p** | Aynı A100 koşulunda doğrulanmış benchmark bulunmadı;2B sayısı taşınmaz | **60–300** sadece kapasite planlama zarfı; offload ve kaliteye çok duyarlı, fiyatlandırma için kullanılmaz | **0,035–0,175**, çok düşük güven |
| **CogVideoX2B,720×480,6sn49kare8fps,50 adım** | [C2] resmi **tek A100~90sn** | **90–180**, referans90; doğal çıktı **8fps**, ürün16–24fps hedefini doğrudan sağlamaz. 6sn klipten5 sn kesme mümkün; daha yüksek fps yazmak süreyi kısaltır, gerçek kare üretmez. Interpolation ayrı motor/maliyet/kalite testidir. | **0,0525–0,105** |
| **CogVideoX5B,aynı6sn8fps** | [C2] resmi tekA100~180sn | **180–360**; lisans ve8fps engelleri nedeniyle yedek | **0,105–0,210** |
| **Mochi1,~5 sn480p** | [BM] A10080GB **0,20 kare/sn**; referans harness60GB. 121kare24fps'e605 sn; bu40GB optimize ölçüm değildir | **600–1200+**, düşük güvenli bütçe;40GB offload daha yavaş olabilir | **0,350–0,700+**;10dk hedefi güvenilmez |
| **Wan2.1 14B,720p~5 sn** | [BTW] tek **A800** baseline~55dk; A10040 değil | 40GB standart yolNO-GO; agresif offload benchmark olmadan sayı verilmez | 55dk kaynak sonucu üzerinden **~1,925**;40GB fiyatı değil |
| **HunyuanVideo13B,720p129kare50 adım** | Resmi hafıza60GB; tekA10080GB topluluk raporu~47dk [BHY], farklı eski kod koşulu | **ELENDİ** (AB+VRAM); benchmark mevcut30 dk max_uptime'ı da aşar | Referans47dk **~1,645**; uygulanacak fiyat değil |

**Ürün varsayımlarına etkiler:**
- 1024² ile1344×768/768×1344 piksel sayısı yakın (~%1,6 daha az) ama attention şekli/tiling nedeniyle aynı p95 garanti edilmez; üçünün ayrı ölçümü gerekir.
- 4 görsel seri üretim: schnell tipik8–20sn / klein8–24sn +encoding. **Warm<=30sn** yalnız model yüklüyse, queue ve swap hariç makul hedeftir; chat önceliği/kuyruk bunu uzatabilir. 4'lü gerçek batch VRAM testi olmadan açılmaz.
- Video async **<=10dk** için Wan1.3B480p ve Wan5B720p olası; model yükleme/kuyruk/soğuk açılış ürün cevabında ayrı görünür. p95 ve kötü promptlar bilinmiyor.
- 10 sn video,5 sn'nin fiyatını/süresini otomatik2 ile çarpmak değildir: frame attention maliyeti ve peak bellek süperlineer artabilir. Wan5B241kare lineer tahmin349–443sn olsa da offload ile600sn üstü/OOM olabilir; **10 sn ilk pilotta kapalı**. LTX model kartı<257kare önerir;241kare yine ölçüm ister. Wan1.3B uzun video kalite ve hafıza ayrı testtir.

## 7. Sohbetle birlikte bellek ve gerekli çalışma düzeni

### 7.1 Neden başlangıçta sequential?

Tam schnell35–37GB bırakacağı2–5GB ile mevcut büyük sohbetin GPU/KV ihtiyacının karşılanacağını kanıtlamaz. Klein~20–24GB ve SDXL~8–12GB daha fazla alan bıraksa da sohbet peak'i **bilinmiyor**. Wan5B~35,5GB kısa klipte bile sınırda. Ayrıca sohbet arşiv RSS~59GiB + Wan5B kaynak host RAM~38GB >85GB nominal host kapasitesi; GPU'dan CPU'ya taşıma tek başına çözüm değildir.

**Öneri:** GPU-ownership mutex, bir anda tek üretim işlemi; chat önceliği; medya için ayrı process/venv. Chat süreçleri yalnız aktif rezervasyonlar bitince güvenli drain edilip kapatılır; medya pipeline yüklenir. Medya bitince process sonlandırılarak CUDA context/VRAM geri alınır; chat yerel NVMe ağırlıklarından tekrar yüklenip health doğrulanır. Sadece `torch.cuda.empty_cache()` başka süreçteki chat belleğini boşaltmaz. Chat'in bütün CPU durumunu RAM'de tutmak da varsayılmaz.

Mevcut systemd `Restart=on-failure` ve ds-tunnel'ın ds-serve bağımlılığı nedeniyle “chat process'i öldür, medya çalışsın” basitçe yapılamaz: yeniden başlatma yarışı, tünelin kapanması ve `/health`'in yanlış negatif olması riski var. İleride onaylı supervisor/mode state (`chat_ready`, `media_loading`, `media_running`, `chat_reloading`) ve gateway route/mode readiness gerekir. Hiçbir servis şimdi değiştirilmedi.

Swap bedeli **UNKNOWN**; her gidiş-dönüş için **60–180sn plan rezervi**, ölçülmüş sonuç değil. DeepSeek256GB tamamını yeniden GCS'den çekmek gerekmez (aynı boot NVMe duruyorsa), ama Strata yeniden yükleme/cache yine pahalı olabilir. Her tek resimde swap yapılmaz; sınırlı medya grupları halinde amorti edilir. Grup başlatmadan önce chat kuyruğu kontrol edilir; sürekli chat varsa medya gecikir ve async ETA açıkça gösterilir.

**Chat önceliği ≠ anlık preemption:** halen çalışan5dk video anında bedelsiz kesilip devam ettirilemez. Destekli sampler adımında cooperative cancel veya mevcut job bitişini bekleme seçenekleri US ile seçilmeli; chat gecikmesi bu sınıra bağlıdır. Uzun video sırasında aynı GPU'da sıfır gecikmeli chat taahhüdü verilmez. Adil kuyruk/aging gerekir, fakat chat SLO'su değiştirilmeden media açlığına çözüm bir ürün kararıdır.

### 7.2 US/SRE ile tasarım sözleşmesi (henüz onaylı değil)

US'den parent üzerinden gelen varsayımlar dikkate alındı: `/v1/image` kısa iş; `/v1/video` job-id, polling/cancel;200 görsel+30 video/gün;24saat çıktı tutma. Bu oturumdan başka ajana doğrudan mesaj gönderilmedi; parent planı US/KAN.US'a iletmeli.

- Tek dayanıklı job kaydı: kullanıcıya bağlı yetki, idempotency key, model/revision, seed, çözünürlük, kare/fps/adım, durum, gpu_elapsed_s, load/swap_s, sonuç digest ve hata/cancel nedeni. Prompt/medya/secret içerikleri JSON loglarına yazılmamalı.
- `queued → admitted/loading → running → encoding/uploading → succeeded/failed/cancelled`; queued işler GPU'yu sonsuza kadar uyanık tutmaz. Kuyruk, model batch planı ve günlük runtime admission aynı otoriteyle yönetilir.
- Kredi rezervasyonu üretim boyunca sürer, HTTP202 yanıtıyla serbest bırakılmaz. Tek charge/settlement, failure/cancel/refund kuralları US tarafından tanımlanır; retry iki kez boot veya kredi tüketmemeli. Kredi tarifesi bu rapor tarafından belirlenmedi.
- İstek kabulünden toplam sonuca kadar queue/cold/load/inference süreleri ayrı gösterilir. İmaj kısa-job timeout'unda client kopsa da işin sahipliği belirsiz bırakılmaz; gerektiğindejob-id dönüşü gerekir.
- Giriş/çıktı boyutu, MIME, codec, seed sınırı, yetkili indirme, kısa ömürlü signed URL, SSRF koruması, telif/insan izni ve moderasyon tasarımı gerekir.
- Çıktı yalnız yerel SSD'de tutulursa discard kapanışında24saat hedefi sağlanmaz. **Onaylı dayanıklı obje deposuna yükleme tamamlanmadan başarı bildirilmemeli.** Yeni bucket/ACL/DNS değişikliği yapılmadı.24saat saklama/otomatik silme şu an **öneri**, mevcut “silme yok” kuralını geçmez: silme/retention mekanizması için açık kapsamlı onay gerekir; şimdilik otomatik silme kurulmaz.

### 7.3 idle.py ve bağımsız watchdog için zorunlu değişiklik taslağı

Eldeki `/workspace/us-panel/sre/watchdog/idle.py.fixed` yalnız80/8080 TCP bağlantıları ve nginx `/v1/` isteklerini aktivite sayıyor; health başarısızlığını hazır olduktan sonra normal idle'a bırakıyor. Asyncvideo202 sonra socket kapanınca **GPU işi çalışırken idle kapanışı riski** var. Tersine `/v1/video/<id>` polling nginx'ten geçerse geniş `GET /v1/` eşleşmesi her poll'da idle'ı uzatabilir. **İkisi de düzeltilmeden medya canlıya çıkmamalı.**

Önerilen, henüz uygulanmamış davranış:
1. **Local güvenilir busy lease:** üretim/decoder/upload çalışırken süreç kimliği+boot-id+iş kimliği, kısa TTL ve heartbeat (örn.20sn) içeren atomik yerel durum. `/run` boot'ta temizlenir. idle.py bunu busy sayar; sadece dosya varlığına körü körüne güvenmez. PID/boot-id/TTL/progress doğrular; worker ölünce TTL dolup kapanış engeli kalkar. Yetkisiz client busy işaretleyemez.
2. **Watchdog durable job lease:** mevcut reservation_active / acquire/heartbeat/release yolu media job tipini desteklemeli. İstemciden bağımsız worker heartbeat'i örn.20sn, mevcut60sn lease'i renew eder; bounded absolute deadline korunur. Cold-ready sonrası execution deadline doğru başlar; queue süresiyle GPU çalışma hakkı karıştırılmaz.
3. Polling, `/health`, `/ready`, model listesi ve kontrol paneli yalnız kayıt okur; **wake ve idle refresh yapmaz**. Gerçek admission ve iş progress'i ayrı olaydır. nginx genişGET eşleşmesi allowlist yerine gerçek iş olayına çekilir; hazır olma yalnız chat health'e bağımlı kalmaz, medya modu geçerli iş varken sağlık hatası diye stop edilmez.
4. Cancel/timeout/OOM/Spotpreemption durumunda job state, tek kredi sonucu, lease release ve son gerçek etkinlik saati atomik tutarlı olmalı. Terminal duruma gelmeden yalnız HTTPcevabı bitti diye release edilmez. Poll/cancel yeni boot tetiklemez.
5. **max_uptime=1800 ve daily_runtime=3600 bypass edilmez.** Yeni iş için kalan runtime>=yükleme+konservatif üretim+flush+en az30sn kapanış payı kontrolü. Küçük GPU heartbeat'i sonsuza kadar sert sınırı uzatamaz; drain + bounded deadline gerekir. Eldeki lifecycle `_claim('stop')` aktif lease varken beklediğinden sınıra güvenli uyum admission/deadline ile de test edilmelidir.
6. ds-idle.timer devrede kalır; medya için sürekli kapatma önerilmez. Canlı testte geçici kapatma gerekiyorsa Demir'den ayrıca açık izin alınmalı. Manuel discard stop fallback'i, disk silme ve test sınırları da ayrı onay kapsamıdır.

## 8. Günlük hacim / mevcut runtime sınırına uyum

**US hacmi bir tahmindir, Demir onayı değil.** Aşağıdaki “200 görsel”,200 ayrı üretim demektir;200istek×4=800görsel ise görsel süresi4kat olur. Hesapta örnek olarak karşılaştırması en açık schnell benchmark bazını kullandım: **200×2,11=422GPU-sn**. Klein plan baz3sn kullanılırsa bu600sn olur (+178sn). Bunlar chat dışı sürelerdir.

### 8.1 Salt üretim alt sınırları

| 30 video/gün seçimi |200 görsel+30 video üretim toplamı|3600sn'ye ilişkin sonuç|
|---|---:|---|
| Wan2.1 1.3B480p,180–300sn | **5.822–9.422sn**,1,62–2,62saat | **Üretim tek başına sığmaz.** |
| Wan2.2 5B720p,175–223sn lineer iyimser tahmin | **5.672–7.112sn**,1,58–1,98saat | **Üretim tek başına sığmaz.** Gerçek ihtiyatlı200–540sn zarfında6.422–16.622sn. |
| Cog2B6sn8fps,90sn kaynak tabanı | **3.122sn**,52,0dk | Sadece üretim yakındır; **16–24fps ürünü sağlamaz**.1800sn cap nedeniyle>=2 boot →+864sn cold ile **>=3.986sn**,idle/swap/chat yokken bile sığmaz. |
| LTX2B480p,20–60sn düşük güvenli tahmin | **1.022–2.222sn**,17,0–37,0dk | **Koşullu:** düşük aralık sığabilir; üst aralık ve chat garantisiz. LTXlisansı ve0.9.8ölçümü kapısı var. |

### 8.2 Boot/idle/swap ve oturum sınırı

- Bir normal oturumda cold432+idle600=**1032sn** sabit kayıp örneği;1800sn içinde faydalı üretime **768sn** kalır (swap/transfer/stop payı daha düşürür). İki tam böyle oturum3600sn'de **1536sn** faydalı iş sağlar.
- **600sn idle her durumda zorunlu sabit süre değildir:** max_uptime'a dayalı drain/stop boşta kuyruğun tamamını beklemeyebilir. Yoğun, kesintisiz plan için iyimser alt sınır `T >= üretim + N×432 + swap/ek_copy + son_idle`; sert1800sn oturum ve admission payı ayrıca kontrol edilir. Kaynağın rate'inde kullanılmayan saat değil VM'in açık kaldığı süre ödenir.
- İki normal oturum ve toplam240sn swap/ek-yükleme rezerviyle LTX düşük senaryo: **1022+864+1200+240=3326sn**,~**1,94USD/gün**;3600'e274sn kalır, chat için hemen hiç marj yok. Üst senaryo **4526sn**,~2,64USD, sığmaz.600sn idle yalnız sonoturumda sayılırsa bu toplamlar600sn azalır; yine ölçüm ve burst düzeni belirleyici.
- Wan5B'nin **10 sn** işleri tek1800sn pencerede cold+swap+600sn tail ile çok daralır; birden çok video request'i sınırsız kabul edilemez. Aynı anda30 video kuyruğuna girerse son işin beklemesi birkaçsaat olabilir: “tekjobGPU<=10dk” ile “kuyruk dahil herkes<=10dk” aynıSLA değildir.
- Günlük talep gün içine dağılmışsa cold/idle tekrarları artar.15dk warmup toleransını herboota uygularsak bir30 dkoturumun üretim bütçesi daha da azalır. Günlükruntime chat ile ortak; burada chat yükü0alındığı için bütün hesaplar medya lehine iyimserdir.

### 8.3 Onaysız uygulanmayacak kapasite önerisi

**Mevcut policy aynen kalır:** idle 600, max_uptime 1800, daily_runtime 3600 sn. Bu sınırla küçük, sentetik pilot ve günlük **ölçülen kalan saniyeye göre quota/admission** önerilir; 200+30 hizmet taahhüdü değil. Kredi yeterli diye runtime kontrolü geçilmez.

US gerçekten 200 görsel + 30 video ve 480p/720p Wan hizmeti istiyorsa, önce yerel test sonucu üzerinden Demir'e seçenek sunulmalı:

- **Örnek öneri, onay değil:** `daily_runtime=10800 sn (3 saat)`, `max_uptime=3600 sn (1 saat)`; idle 600 aynı. Ancak bu her aday için yeterli değildir. Wan1.3B video başına 240 sn olursa üretim 7622 sn; 3 cold boot 1296 sn + toplam 360 sn swap + 3×600 sn idle ile **11078 sn** gerekir. Bu örnekte 3 saat bile yetersizdir; daha çok oturum ek cold overhead getirir.
- Wan5B **200 sn** doğrulanırsa 6422+1296+360+1800=**9878 sn ≈2,74 saat**, yaklaşık **5,76 USD/gün** (chat/retry hariç). 3 saat yalnız bu iyimser doğrulanmış durumda makul paydır. Video başına 350 sn bazında toplam **14378 sn ≈4 saat** olur; 3 saat uygun değildir. Ölçülen p95 ile 4–5 saat günlük limit veya azaltılmış hacim ayrıca karar ister.
- 10800 sn günlük limit kaba üst maliyet **6,30 USD/gün**, 5 saat **10,50 USD/gün**: yalnız 0,035 USD/dk varsayımı; satın alma, fatura garantisi veya otomatik bütçe değil. Chat yükü ve Spot retry ayrıca boyutlandırılır. 1800 sn oturum sınırı korunacaksa daha çok boot/cold ve daily_starts kontrolü gerekir; son canlı daily_starts burada doğrulanmadı.
- Daha uzun uptime yeni recurring maliyet riskidir; **Demir'in açık onayı olmadan policy değişmez**. 80 GB VM, yeni GPU, kalıcı disk veya bucket oluşturulmaz. Şimdilik alarm, bütçe veya rutin de kurulmaz.

## 9. Aşamalı uygulama ve onay kapıları

**Aşama 0 — Şimdi tamamlanan plan:** 40 GB doğrulaması, lisans taraması, dosya boyutu ve kapasite hesabı. GPU kapalı; BOOT6 beklemede.

**Aşama 1 — US + SRE tasarımı, GPU açmadan:**
1. US, model allowlist/kalite, klip fps/çözünürlük, kredi/metering, cancel/refund ve chat önceliği sözleşmesini netleştirir. Özellikle Cog'un 8 fps çıktısı ürün hedefinden ayrılır; 10 sn klip kapalı başlar.
2. SRE, **canlı sre-ready tabanının** /ready, rate-limit ve JSON log yamalarını koruyacak supervisor/mutex, mode-readiness, busy-lease ve watchdog lease tasarımını hazırlar. Bağımsız deploy yok; mevcut rollback kuralı korunur.
3. Lisans dosyaları + revision SHA + seçilen ağırlık manifesti (bütün repo snapshot'ı değil), bileşen izinleri ve artifact hash planı hazırlanır. Gated lisans kabulleri/kayıtlar gerekirse Demir'e sunulur; bu çalışmada yapılmadı.
4. Sağlık kontrolü/polling uyandırmaz. Kuyruk, lease crash/TTL, cancel, günlük limit ve mode-switch testleri gerekir. Bu araştırmada kod değiştirilmedi.

**Aşama 2 — Demir'in açık onayına sunulacak tek pilot paketi:**
- Mevcut 3 test açılışı hakkının birinin medya pilotu için kullanılıp kullanılmayacağı; **BOOT6 bekletmesinin ayrıca kaldırılması gerekir**. Eski boot bütçesi otomatik medya testi onayı değildir.
- Hangi 2–3 checkpoint; staging/GCS yazma/depolama maliyeti; VM geçici process değişiklikleri ve geri dönüş. Sentetik fixture varsa ayrı izin; gerçek müşteri ve ödeme verisi yok.
- İlk hedef: klein veya schnell batch=1; 1024² ve yatay/dikey iki boyut; 4 seri görsel. Wan1.3B 480p 5 sn ve Wan5B 720p 5 sn birer kısa örnek. Zaman bütçesine sığmazsa hepsini tek boot'a sıkıştırma; ek boot otomatik yok.
- Test başına önceden onaylı süre/maliyet: 30 dk pencerenin yalnız süre hesabı **~1,05 USD**; sağlayıcı tarafından enforce edilen maliyet cap'i değil. Aynı gün kalan runtime ve başlatma limiti önceden kontrol edilir. Önceki test planındaki 35 dk / 1,30 USD genişlemesi onaylı sayılmaz.
- Chat drain/restore, idle timer devrede, busy heartbeat testi; kapanış/discard fallback ve disk silme kapsamı dahil onay gerekir. Gerçek trafik güvenle ayrıştırılamıyorsa NO-GO.

**Aşama 3 — Ölçüm kabul kriterleri; sadece onaydan sonra:**
- Gerçek GPU model/memory.total; her motorun weight-load ve warmup süresi; peak allocated/reserved/NVML VRAM, host RSS, swap I/O; kopya bayt/sn; model load ve swap p50/p95; çözünürlük, kare, adım ve seed kaydı.
- Üretim/load/queue/upload süreleri ayrı; örnek başı maliyet ve başarısızlık payı. Spot kesintisi, cancel/refund, TTL ve idempotency doğrulaması. 202 sonrası socket kapalıyken busy lease ile kapanmama; polling kesilse işin sürmesi; boşta 600 sn sonrası kapanma; health/poll'un uyandırmaması.
- Chat modelinin başarılı yeniden yüklenmesi; gerçek chat isteğinin önceliği ve gecikme ölçümü; NVMe'den reload kanıtı. US kredi hesabı ölçümlerden sonra kesinleşir.
- Her pilot sonrası onaylı discard yolu ile **TERMINATED** doğrulaması; boot sayacı, maliyet ve rollback raporu. Sonuç başarısızsa 10 sn / 720p kapsam genişletme yok.

**Aşama 4 — Dar canlı pilot:** yalnız onaylı model ve quota; dayanıklı çıktı; 24 saat retention/silme için ayrı izin. Policy 3600/1800 varsayılanı ölçümlere ve yazılı onaya dek korunur.

## 10. Kaynaklar ve kanıt dosyaları

Resmi lisans/model dokümanları kapasite benchmark bloglarından üstündür. Kaynaklar 8 Kasım 2026 TSİ bu çalışmada okundu; web sayfaları değişebilir, uygulamada revision pinlenmeli.

- **[G1]** Google resmi GPU/makine tablosu: https://docs.cloud.google.com/compute/docs/gpus
- **[F2]** Resmi klein 4B kartı: https://huggingface.co/black-forest-labs/FLUX.2-klein-4B
- **[F2L]** Resmi Apache lisansı: https://huggingface.co/black-forest-labs/FLUX.2-klein-4B/resolve/main/LICENSE.md
- **[F1]** Resmi schnell kartı: https://huggingface.co/black-forest-labs/FLUX.1-schnell
- **[FD]** Resmi dev lisansı: https://huggingface.co/black-forest-labs/FLUX.1-dev/resolve/main/LICENSE.md
- **[SX]** SDXL kartı: https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0
- **[SXL]** SDXL lisansı: https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0/resolve/main/LICENSE.md
- **[S3]** SD3.5 kartı: https://huggingface.co/stabilityai/stable-diffusion-3.5-medium
- **[S3L]** SD3.5 resmi lisans metni: https://huggingface.co/stabilityai/stable-diffusion-3.5-medium/resolve/main/LICENSE.md
- **[W1]** Wan2.1 resmi kartı: https://huggingface.co/Wan-AI/Wan2.1-T2V-1.3B ; resmi README: https://github.com/Wan-Video/Wan2.1
- **[W2]** Wan2.2 resmi README/lisans/komutlar: https://raw.githubusercontent.com/Wan-Video/Wan2.2/main/README.md
- **[W2T]** Resmi A100/A800 benchmark görseli: https://raw.githubusercontent.com/Wan-Video/Wan2.2/main/assets/comp_effic.png (yerel kopya: `media_research/wan22_comp_effic.png`).
- **[LTX]** Resmi kart/sürüm bağlantıları: https://huggingface.co/Lightricks/LTX-Video
- **[LTXL]** 0.9.6+ lisansı: https://huggingface.co/Lightricks/LTX-Video/resolve/main/LTX-Video-Open-Weights-License-0.X.txt
- **[LTX95L]** 0.9.5 lisansı: https://huggingface.co/Lightricks/LTX-Video/resolve/main/ltx-video-2b-v0.9.5.license.txt
- **[C2]** Cog2B resmi kart, A100 90/180 sn ve VRAM tablosu: https://huggingface.co/zai-org/CogVideoX-2b
- **[C5L]** Cog5B lisansı: https://huggingface.co/zai-org/CogVideoX-5b/resolve/main/LICENSE
- **[M1]** Mochi resmi kart, 60 GB / 42 GB / 22 GB yolları: https://huggingface.co/genmo/mochi-1-preview
- **[HY]** Hunyuan resmi kart/VRAM: https://raw.githubusercontent.com/Tencent-Hunyuan/HunyuanVideo/main/README.md ; ayrı text encoder'lar: https://raw.githubusercontent.com/Tencent-Hunyuan/HunyuanVideo/main/ckpts/README.md
- **[HYL]** Tencent AB/UK/Kore lisansı: https://raw.githubusercontent.com/Tencent-Hunyuan/HunyuanVideo/main/LICENSE.txt
- **[BF1]** Üçüncü taraf schnell A100 40 GB benchmark'ı: https://gpubattle.com/guides/flux-1-schnell-gpu-benchmarks
- **[BSX]** Resmi Stability/NVIDIA SDXL A100 benchmark'ı: https://huggingface.co/stabilityai/stable-diffusion-xl-1.0-tensorrt ; ek PyTorch A100 80 GB çalışması: https://pytorch.org/blog/accelerating-generative-ai-3/
- **[BS3]** SD3.5 **H100** benchmark'ı: https://huggingface.co/stabilityai/stable-diffusion-3.5-medium-tensorrt
- **[BTW]** TeaCache **A800** ölçümleri: https://raw.githubusercontent.com/ali-vilab/TeaCache/main/TeaCache4Wan2.1/README.md
- **[BW5]** Üçüncü taraf Wan5B 49 kare A100 benchmark'ı: https://gpubattle.com/guides/wan-2-2-5b-gpu-benchmarks
- **[BLTX]** Üçüncü taraf LTX 0.9.5 A100 benchmark'ı: https://gpubattle.com/guides/ltx-video-gpu-benchmarks
- **[BM]** Mochi A100 80 GB tablosu: https://gpubattle.com/rankings/text_to_video
- **[BHY]** Eski A100 80 GB topluluk raporu, 47 dk: https://github.com/Tencent-Hunyuan/HunyuanVideo/issues/2 (arama indeksi dayanağı; yüksek güvenli fiyatlama değil).
- Klein A100 indeks raporu (README canlı doğrulanamadı): https://github.com/Gladiator07/flux2-klein

### Yerel kanıtlar
- `media_research/ds-a100-readonly.json`: salt okunur VM kimliği/durum; son okuma **8 Kasım 01:20:46 TSİ**, TERMINATED.
- `media_research/*__*.json`: resmi HF metadata, `sha`, seçilen dosya boyutları. Küçük metadata; hiçbir checkpoint indirilmedi.
- `/workspace/lt/coldstart/inspect1a.txt`, `boot2_detail.txt`, `boot3_now2.txt`, `patches/fetch-weights.py`: eski üniteler, çalıştırma parametreleri ve copy ölçümleri.
- `KAN_S4_LIVE_RESULT_BOOT5.md`, `SRE_BOOT5_VERIFY.md`: son sohbet açılışı ve idle/discard kanıtı.
- `/workspace/us-panel/sre/watchdog/idle.py.fixed`: async busy event içermeyen mevcut yerel idle betiğinin box kopyası.
- `/workspace/us-panel/work2/kanus/inference_lifecycle.py`: mevcut lease/admission/stop semantiğinin yerel kopyası; SRE'nin canlı ek yamalarının tam yerine geçmez.

**Sonuç:** 40 GB'lık mevcut makinede açık medya motoru tek tek çalıştırılabilir; esas darboğaz sadece VRAM değil, chat reload/host RAM, günlük 1 saat policy ve async iş sahipliğidir. Bütçe ve BOOT6 bekletmesi değişmedi; bu plan uygulama başlatmaz.
