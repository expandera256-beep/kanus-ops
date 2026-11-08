# US `/v1/image` HTTP sözleşmesi: text-to-image (klein ana, schnell açık seçim)

**Durum: PLAN / TASARIM ONLY. UYGULANMADI.** Canlıya hiçbir şey uygulanmadı. Kod, handler, patch, test, deploy, GPU açılışı, BOOT6 veya HOST işlemi yapılmadı.
**Revizyon geçmişi:**
- **r1:** SHA256 `abef2174b86cee203178319c1184752b5063abcd09499932ff92099500b5a176` (39326 bayt, 508 satır; PLAN/TASARIM ONLY).
- **r2:** Demir’in `t78s286` kart cevapları, 8 Kasım 2026 ~04:34 TSİ (Europe/Istanbul, UTC+3), KAN.US iletti. Yerel tasarım güncellemesi; deploy veya test boot onayı değil. Yalnız bu sözleşme ve hash sidecar güncellendi.

**Tarih:** 8 Kasım 2026, Europe/Istanbul saatiyle (TSİ, UTC+3). Taslak 04:00 TSİ civarında yazıldı.
**Görev:** KAN.US aktarımı: Demir 08.11.2026 03:52 TSİ (KAN.US t68u). `KAN_IMAGE_ENGINE_PLAN.md` §6 ile uyumlu sözleşme.
**Bu belge şunlar değildir:** satış onayı, .13 EXE onayı, ölçülmüş SLA, deploy/test boot izni veya mağaza fiyat değişikliği. r2 ürün kararlarını tasarıma işler; depolama işlemi yapmaz. Model ve/veya boyuta göre kredi oranı kararlıdır; sayısal tarife **PENDING Demir**. 1 görsel = 1 istem kredisi önerisi reddedildi.

Etiketler:
- **SOURCED ESTIMATE:** dış kaynaktan veya plandan alınmış tahmin; yerel ölçüm değil.
- **NOT TESTED:** tasarlandı, hiç çalıştırılmadı.
- **UNKNOWN:** değer bilinmiyor.
- **KARAR GEREKİR:** Demir kararı gerekir.

Bu belgede hiçbir satır PASS değildir.

---

## 0. Kaynaklar (salt okunur; hiçbiri değiştirilmedi)

| Kaynak | SHA256 | Not |
|---|---|---|
| `/workspace/us-panel/reports/r20261108/KAN_IMAGE_ENGINE_PLAN.md` | `6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38` | Sözleşme hedefi §6; ayrıca §1, §2, §5, §7, §8, §9 ve §11 |
| `/workspace/us-panel/reports/KANUS_PRICING_GLOBAL.md` | `0e19b1b455edfd972ab45e58997ea38f5a24d86415f40181e484e74bde1a949a` | İstem paketleri ve TL/istem |
| `us-panel-dev/p6/term/licensesrv/credits.go` | `ebaf335558061a28cb1ebd7880ad1404e8f0117a6bad51af421aacb085ecc275` | r13h kopyasıyla birebir aynı |
| `us-panel-dev/p6/term/licensesrv/model_meter.go` | `5eccff6844639deb38acb9c5b3587bd70927fdbbed3dead3eb0e3a2cbedf679a` | `reserveModel` / `settleAmount`; r13h ile aynı |
| `us-panel-dev/p6/term/licensesrv/backpressure.go` | `f629ad8b0472d127d84f29b7307411d952245483ab2caa08b15371833033930e` | 429/503/504 için `Retry-After` ve kuyruk başlıkları; r13h ile aynı |
| `us-panel-dev/r13h/term/licensesrv/model_proxy.go` | `1f9e473e79368529f0d237f5021f7bcb1f06d20db42506e52fd9f1191a467513` | Yalnız okundu. Admission sırası buradan alındı (R12 entitlement gate). p6 kopyası daha eski. |
| `us-panel-dev/r13h/term/licensesrv/gpu_wake.go` | `e1e73dee3e2fe1ac9b45ebb83729716aba0041125cec552b6779b75c77571dfd` | Yalnız okundu. `writeWarming` 503 gövdesi buradan. |
| `KAN_IMAGE_ENGINE_STAGE.md` (yerel evre notu, salt okunur) ve KAN.US aktarımı (8 Kasım 2026 TSİ) | — | klein private staged bildirildi; ready işareti ve deploy yok. Önceki 0 dosya bilgisi tarihsel kaldı. schnell HF 401 BLOCKED. US bağımsız canlı doğrulama yapmadı. |

**Kaynakta görülemeyen kısım:** `Store.Reserve`, `settleFor`, `authorizedCredential` ve `modelSpendable` gövdeleri ile `Store` / `Reservation` / `Settlement` tipleri, okunan `licensesrv` ağaçlarında **yok** (`store.go` bulunmuyor). Bu belge yalnız çağrı yerlerinden görüleni kullanır:
- `Reserve(key, amount, ttl, now)`
- `settleAmount(key, id, amount, now)`

Bu belge yeni Store/Mail/Plan tipi **tanımlamaz**. Uygulayıcı gerçek imzaları doğrulamalıdır.

---

## 1. Kimlik, yöntem ve kapsam

- **Uç nokta:** `POST /v1/image`, yalnız HTTPS. Bu rotada diğer yöntemler `405 method_not_allowed` döner ve `Allow: POST` başlığı eklenir.
- **Kimlik doğrulama:** Kredi düşen diğer `/v1` rotalarıyla aynıdır: `Authorization: Bearer <lisans anahtarı>`.
  - Global secret yoktur. Anonim üretim yoktur.
  - Kabul edilen kimlik bilgisi kümesi `/v1/chat/completions` model proxy'si ile aynıdır (`authorizedCredential` ve lisans eşlemesi). Yeni kimlik bilgisi türü eklenmez.
  - Başlık yoksa veya `Bearer ` öneki yoksa: `401 license`. Mevcut davranışla aynıdır.
- **Yanıt başlıkları:** Her yanıtta `Cache-Control: no-store` ve `X-US-Metered: 1` bulunur. Model proxy ile aynıdır.
- **İlk kapsam:** yalnız **text-to-image**. Şunlar açıkça **reddedilir**:
  - img2img, düzenleme (editing), inpainting ve mask
  - girdi olarak görsel yükleme (multipart, URL veya base64)
  - LoRA, adapter, kullanıcı checkpoint'i, ControlNet
  - video
  - keyfi çözünürlük, `width` / `height`
  - istemcinin `steps`, `guidance`, `dtype`, `scheduler`, `batch` ve `negative_prompt` göndermesi
  - istemcinin `amount`, `ttl`, `credits` veya `price` göndermesi
- `seed` yalnız tekrar üretilebilirlik içindir. Kaydedilir ve **faturalamayı hiçbir şekilde değiştirmez**.

---

## 2. İstek

### 2.1 Başlıklar

| Başlık | Zorunlu | Kural |
|---|---|---|
| `Authorization: Bearer …` | Evet | §1 |
| `Content-Type: application/json` | Evet | `; charset=utf-8` kabul edilir. Diğer türler, multipart dahil, `415 unsupported_media_type` döner. |
| `Idempotency-Key` | **Evet** | Birincil ve tek idempotency kanalı (§2.4). Yoksa `400 missing_idempotency_key`. |

### 2.2 Gövde (JSON nesnesi, **katı ayrıştırma**)

| Alan | Tür | Varsayılan | Doğrulama |
|---|---|---|---|
| `model` | string | `"klein"` (alan yoksa) | Allowlist yalnız `"klein"` ve `"schnell"`; büyük/küçük harf duyarlı, tam eşleşme. Bilinmeyen değer, boş string, `null` veya tam HF repo adı `400 invalid_model` döner. **klein'e sessizce yeniden eşleme yok.** `schnell` yalnız istemci bu değeri açıkça gönderirse kullanılır. |
| `prompt` | string | — (zorunlu) | Geçerli UTF-8. Baştaki ve sondaki boşluk kırpıldıktan sonra boş olmamalı. NUL ve `\n`, `\t` dışındaki C0 kontrol karakterleri `400 invalid_prompt` döner. Üst sınır için §2.3'e bakın. |
| `size` | string | `"1024x1024"` | Yalnız `"1024x1024"`, `"1344x768"`, `"768x1344"`. Ayraç tek ve kanoniktir: küçük ASCII `x`. `1024×1024` (Unicode ×), `1024*1024`, `1024X1024`, `1024 x 1024`, `[1024,1024]` ve diğer her biçim `400 invalid_size` döner. |
| `n` | integer | `1` | `1..4`. Kesirli değer, string, 0, 5+ veya `null` `400 invalid_n` döner. n × rate(model, size) kullanılabilir krediyi aşıyorsa n sessizce küçültülmez; `402` döner (§5). |
| `seed` | integer | Sunucu rastgele seçer | `0..4294967295`. Görsel i'nin seed değeri `(seed + i) mod 2^32` olur (i=0..n-1). Seed'ler yanıtta ve usage kaydında yer alır. Faturalamaya etkisi yoktur. |

**Bilinmeyen alanlar:** **REDDEDİLİR**, sessizce atılmaz. Bilinmeyen bir anahtar `400 unknown_field` döner ve `error.field` alanı doldurulur.

Aşağıdaki bilinen ama yasak alanlar `400 unsupported_parameter` döner. Bu, model proxy'nin `unsupported generation override` davranışıyla aynı ruhtadır:
- `steps`, `num_inference_steps`, `guidance`, `guidance_scale`, `true_cfg_scale`, `dtype`
- `width`, `height`, `batch`, `batch_size`, `scheduler`
- `image`, `init_image`, `mask`, `strength`
- `lora`, `adapter`, `checkpoint`
- `negative_prompt`, `response_format`, `quality`, `style`
- `amount`, `ttl`, `credits`, `price`, `user`
- `idempotency_key` (gövdede gönderilirse)
- `fallback`, `allow_fallback`

Alan adları büyük/küçük harf duyarlıdır. Yinelenen JSON anahtarı `400 invalid_json` döner. Bu yüzden ayrıştırıcı yinelenen anahtarları tespit etmelidir. JSON dışında kalan bayt (trailing data) de `400 invalid_json` döner.

### 2.3 Gövde ve prompt sınırları

- **Taşıma sınırı (öneri):** gövde en çok **32 KiB**. Aşılırsa `413 body_too_large`. Bu yalnız taşıma koruyucusudur; token sınırı değildir. Bu değerin her geçerli prompt'u kapsayıp kapsamadığı **NOT TESTED**.
- **Prompt token sınırı:** **doğrulanacak.**
  - Sınır pinned model config'inden okunacak: klein Qwen3 tokenizer ve pipeline `max_sequence_length`; schnell CLIP + T5 tokenizer ve pipeline sınırı.
  - Türkçe kalite testi gerekir. Gizli dış çeviri API'si yoktur.
  - Bu belge **bir token sayısı uydurmaz**. Değer **UNKNOWN / NOT TESTED**.
- **Sınır ölçülünce kural:** sınırı aşan prompt `413 prompt_too_long` döner. Sessiz kırpma (truncation) yapılmaz.
- **schnell CLIP kanalı:** schnell'in çift encoder'ının (CLIP 77 token kanalı ve T5 kanalı) iç kırpma davranışı **UNKNOWN**. Pinned config gated olduğu için okunamadı (plan §2.1). Uygulamadan önce doğrulanıp belgelenmelidir.
- Sınır bilinene kadar 413 `prompt_too_long` yalnız katalogda ayrılmış bir koddur.

### 2.4 Idempotency (tek mantıksal iş = tek anahtar)

- **Birincil kanal:** `Idempotency-Key` başlığı.
  - Gövde alanı `idempotency_key` **reddedilir** (`400 unsupported_parameter`). Böylece başlık ve gövde arasında split-brain oluşmaz.
  - Biçim: 16–128 karakter, `[A-Za-z0-9._:-]`. Uymayan anahtar `400 invalid_idempotency_key` döner.
- **Kapsam:** `(lisans kimliği, Idempotency-Key)` çifti. Farklı lisanslar aynı anahtarı kullanabilir; birbirini görmezler.
- **Parmak izi:** normalize edilmiş `{model, prompt, size, n, seed}` kümesinin sunucu tarafında anahtarlı HMAC-SHA256 değeri.
  - Varsayılan değerler önce doldurulur. `seed` yoksa parmak izine "yok" olarak girer; sunucunun seçtiği seed ilk yürütmede kayda bağlanır.
  - Prompt düz metin olarak kaydedilmez ve loglanmaz.
- **Davranış:**

| Durum | Yanıt | Yeni iş, boot veya swap | Ek kredi |
|---|---|---|---|
| Anahtar yeni | Normal akış (§4), 200 veya 202 | Admission kurallarına göre | n × rate(model, size); tarife PENDING |
| Aynı anahtar, aynı parmak izi, **iş sürüyor** | `202 Accepted`, aynı iş kimliği + güncel durum ve yoklama yolu (§9) | **Hayır** | Hayır |
| Aynı anahtar, aynı parmak izi, **terminal** (başarılı veya kısmi) | Saklanan 200 yanıtı aynen döner; `Idempotent-Replay: true` eklenir | **Hayır** | **Hayır** |
| Aynı anahtar, aynı parmak izi, **terminal hata, 0 çıktı** | Saklanan hata yanıtı döner; `Idempotent-Replay: true` eklenir | **Hayır** | Hayır (zaten serbest bırakıldı) |
| Aynı anahtar, **farklı parmak izi** | `409 idempotency_key_reused` | Hayır | Hayır |

- **Rezervasyondan önceki retler** anahtarı **bağlamaz**: 400, 401, 402, 403, 413, 415, 429, 503, 504. Aynı anahtarla daha sonra tekrar deneme meşrudur ve işi başlatabilir.
  - Anahtar, rezervasyonla **aynı atomik işlemde** bağlanır.
  - Eşzamanlı iki aynı-anahtar isteği tek uçuşa (single-flight) indirgenir.
- **İstemci timeout'u ve kör retry güvenlidir:** aynı anahtar ikinci rezervasyon, ikinci boot, ikinci swap veya ikinci ücret üretmez (plan §6).
- **Kayıt ömrü:** idempotency kaydı en az rezervasyon TTL'i (45 dk), kabul edilmiş işin ömrü ve 24 saatlik çıktı erişim penceresi boyunca korunmalıdır; terminal replay yeni üretim başlatmaz. Daha uzun ledger saklama politikası **PENDING** (§11).
  - Bu belge otomatik silme veya lifecycle **kurmaz**.
  - Kayıt yalnız ledger meta verisidir: HMAC, durum, sayılar ve referanslar. Prompt metni içermez.

### 2.5 Örnek

```http
POST /v1/image HTTP/1.1
Host: <US API host>
Authorization: Bearer <lisans-anahtarı>
Content-Type: application/json
Idempotency-Key: 7f3c2a1e-img-20261108-0001

{"model":"klein","prompt":"İstanbul'da yağmurlu bir akşam, sulu boya","size":"1344x768","n":2}
```

---

## 3. Sunucuya sabit parametreler (istemci değiştiremez)

Değerler plandan alınan **başlangıç önerisidir**. Hepsi **NOT TESTED**.

| Alias | Sunucu kimliği (HF repo @ pinned revision) | Lisans | Adım | Guidance | dtype | CUDA batch |
|---|---|---|---|---|---|---|
| `klein` (varsayılan) | `black-forest-labs/FLUX.2-klein-4B` @ `e7b7dc27f91deacad38e78976d1f2b499d76a294` (distilled 4B) | Apache-2.0 | 4 | 1 | BF16 | 1 |
| `schnell` (yalnız açık seçim) | `black-forest-labs/FLUX.1-schnell` @ `741f7c3ce8b383c54771c7003378a50191e9efe9` | Apache-2.0 | 4 | 0 | BF16 | 1 |

- **n=4, en çok 4 seri görsel demektir.** CUDA batch-4 değildir.
- Runtime yalnız pinned revision'ı doğrulanmış yerel snapshot'tan yükler. Niyet `local_files_only` modudur. `main` ve kayan tag kullanılmaz. Request başına HF indirmesi yapılmaz.
- Şu modeller **yasaktır:** klein-base-4B, klein-9B, FLUX.1-dev ve FLUX.2-dev denoiser. Allowlist'te yer almazlar ve hiçbir alias bunlara eşlenmez.
- **Çıktı biçimi:** başlangıç önerisi `image/png`, kayıpsız. Encode süresi ve boyutu **NOT TESTED**. Biçim değişikliği ayrı bir karardır.
- **Evre notu (KAN.US, 8 Kasım 2026 TSİ; yerel evre raporu da okundu):** klein ağırlıkları onaylı mevcut önekte private staged (24 nesne, 15.980.183.023 bayt); ready işareti yok; deploy yok. US bu evreyi bağımsız canlı doğrulamadı.
  - Önceki “klein staging BLOCKED, 0 dosya” ifadesi tarihsel bilgidir; güncel aktarım değildir.
  - Ready işareti olmadan canlı servis ready sayılmaz. Ready yok + deploy yok ⇒ canlı istek hâlâ `503 not_ready`, `reason=model_not_staged` veya eşdeğer servis-hazır-değil reddi olmalıdır (tasarım; canlı istek yapılmadı).
  - schnell HF gated: HTTP 401 **BLOCKED**; indirme, token arama veya aşma yok.
  - Diğer modele görünür veya otomatik fallback yok.


---

## 4. İşleme sırası, senkron/asenkron ve admission (plan §6, kritik)

### 4.1 Kesin sıra

1. Yöntem, `Content-Type`, gövde sınırı, katı JSON ve alan doğrulaması (400, 405, 413, 415). GPU'ya veya control-plane'e dokunulmaz.
2. Kimlik (401).
3. **Entitlement gate:** r13h model proxy R12 sırasıyla readiness ve wake'ten önce gelir. Lisans aktif değilse 403; `Kind == "use"` için geçici `403 plan_kind_not_supported` (§5.3). Kredi kontrolü salt okunurdur: gerekli tutar `n × rate(model, size)`; tarife **PENDING Demir**, sayısal ücret hesaplanamaz ve bu belge uygulama başlatmaz. GPU'ya dokunulmaz.
4. **Idempotency araması** (§2.4). Mevcut iş aynı kimlikle döner; ek rezervasyon, boot/swap veya ücret yoktur. Replay'in mevcut işe erişimi yeni iş için bakiye kontrolüyle engellenmemelidir; uygulayıcı gate ile bu ayrımı korumalıdır.
5. **Admission:** supervisor/control-plane salt okunur durumundan snapshot/readiness, chat önceliği, kuyruk, günlük runtime, max_uptime ve boot yetkisi kontrol edilir. Bu okuma wake/idle etkisi yapmaz. `NotReadyTTL` negatif önbelleği ve singleflight mevcut yaklaşımı korunur; kabul edilebilir cold/swap talebi terminal retle karıştırılmaz (§4.3).
6. **Rezervasyon ve iş kabulü:** sunucu `amount = n × rate(model, size)`, `TTL = ReservationTTL = 45 dk` belirler; anahtar aynı atomik işlemde bağlanır. Tarife yokken sayısal rezerv/üretim uygulanmaz. Yarışta kredi yetmezse 402, lisans geçersizse 403, ledger hatasında `503 reservation_failed`. Rezervasyon kimliği istemciye dönmez.
7. **Başlatma:** kabul edilmiş POST cold boot veya gerekli swap'ı tek mantıksal iş için en çok bir kez başlatabilir. Ertelenen işte 202 + yoklama; pipeline ve slot hazırken senkron tamamlanabilen işte 200 yolu. Boot/swap başlatılamazsa terminal hata kaydı ve rezervin serbest bırakılması gerekir; kör retry yeni tetikleme yaratmaz. **Hepsi tasarım; şu an hiçbir tetikleme yoktur.**
8. **Çalıştırma:** n görsel seri üretilir. Encode ve persist tamamlanınca görsel üretilmiş sayılır. Deadline/iptal kontrolleri görseller arasında yapılır; sequential GPU sahipliği korunur (§4.2).
9. **Tek settlement:** tamamlanan çıktıların kredi tutarı `produced × rate(model, size)` ile `settleAmount(key, id, amount, now)` semantiğinde kapatılır; rezervin kullanılmayan kısmı serbest kalır. 0 çıktı → release. Rate sayısal olarak belirlenmemiştir.
10. Terminal 200 veya hata idempotency kaydına yazılır; async iş sonucu aynı lisansla yoklanır (§9). 202 kabul yanıtı terminal sonuç değildir.

### 4.2 Warm senkron yol ve async iş süresi

- `200` senkron yol için iki koşul birlikte gerekir: istenen image pipeline **yüklü** (`IMAGE_READY`) ve GPU slotu **hazır**; iş senkron tamamlanabiliyorsa çıktı referanslarıyla 200 döner. Warm ≤30 sn yalnız hedef, **ÖLÇÜLMEDİ / NOT TESTED** (§4.4).
- **GPU sıcak ve chat yüklü olması image'ın sıcak olduğu anlamına gelmez.** `CHAT_READY → CHAT_DRAINING → IMAGE_LOADING → IMAGE_READY/RUNNING → IMAGE_DRAINING → CHAT_RELOADING → CHAT_READY` sırası ve tek CUDA sahibi korunur (plan §5.1). klein↔schnell açık model değişimi de swap'tır, fallback değildir.
- Aktif chat tamamlanana kadar drain edilir; medya başlamadan yeni/bekleyen chat önceliklidir. Her kısa image request sonunda chat kuyruğu kontrol edilir. Kabul edilen n=4 request deadline/hata yoksa seri tamamlanır; chat için zorla ortadan kesme eklenmez.
- **Çalıştırma deadline'ı önerisi: 90 sn, NOT TESTED.** Warm yolda kabul→yanıt; async yolda pipeline+slot hazır olup çalıştırma başladığından terminal encode/persist'e kadar. Cold/swap/kuyruk beklemesine 90 sn uygulanmaz. Bu beklemenin bounded üst sınırı **PENDING / NOT TESTED**; rezerv TTL'i, boot/load sınırları ve 1800/3600 limitleri aşılamaz. Sonsuz kuyruk/lease yoktur.
- Sonraki görsel deadline'a sığmıyorsa başlatılmaz; teknik kısmi sonuç §5.4. Bu, kısmi iade ürün metni değildir.
- Senkron istemci timeout önerisi en az 120 sn (**NOT TESTED**); timeout'ta aynı `Idempotency-Key` ile retry. 202 sonrasında aynı lisansla yoklama yapılır; HTTP bağlantısını açık tutmak gerekmez.

### 4.3 Cold, swap, kuyruk ve kota: 202 kabul / ret ayrımı

- **Demir kararı:** `POST /v1/image` cold ise boot, swap gerekiyorsa chat→image swap başlatabilir ve **202 Accepted** döner. Bu API ürün kararıdır; **UYGULANMADI**, ds-a100/BOOT6 veya herhangi bir GPU/VM şu anda açılmayacak, canlı tetikleme yapılmayacaktır.
- Kabul edilen cold, swap, load/drain veya kuyruk işi HTTP bağlantısında belirsizce bekletilmez; iş kimliği, status ve yoklama ipucuyla 202 döner (§9). Eski “cold/swap → GPU uyandırmadan varsayılan 503” kaldırılmıştır.
- **Chat priority korunur:** aktif/bekleyen chat image çalışmasını öne almaz; kabul edilebilir talep `queued` kalabilir, swap chat drain/öncelik kuralları izin verdiğinde başlar. Kuyruğa kabul mümkün değilse aşağıdaki retler kullanılır. Kuyruk/ETA ürün metni **PENDING, sonra yazılacak**; burada ETA vaadi yoktur.
- `daily_runtime=3600`, `max_uptime=1800`, `idle=600` değişmez. Günlük runtime tükenmiş/yetersizse `503 daily_runtime`: boot/swap başlatılmaz. Reload/output/kapanış payı ve max_uptime sınırı da geçerlidir; aktif lease limitleri uzatmaz.
- Boot'un çözmeyeceği `model_not_staged` veya `engine_unavailable` durumunda 503 kalır ve boot başlatılmaz. Salt ürün kararı BOOT6/test boot beklemesini veya otorite yetki sınırlarını kaldırmaz.
- Kabul öncesi 429/503/504 retlerinde kredi ayrılmaz, anahtar bağlanmaz ve wake/idle etkisi yoktur. Kabul edilmiş işin sonraki hatasında §5.4 uygulanır.

| HTTP | `reason` | Anlamı | `Retry-After` önerisi (hepsi **NOT TESTED**, ürün ETA'sı değil) |
|---|---|---|---|
| 503 | `boot_pending` | Otorite boot yetkisi vermiyor; BOOT6 beklemesi veya açılış hakkı yok | 300 |
| 503 | `chat_priority` | Chat önceliği nedeniyle image kuyruğuna kabul de mümkün değil | 30 |
| 503 | `chat_reloading` | Restore sürüyor ve güvenli kuyruğa kabul mümkün değil | 60 |
| 503 | `daily_runtime` | 3600 sn günlük havuz image/reload/kapanış payına yetmiyor | Sıfırlanmaya kalan süre biliniyorsa o değer, yoksa 3600 |
| 503 | `max_uptime` | 1800 sn oturum sınırında kalan süre yetmiyor | Bilinmiyorsa 600 |
| 503 | `model_not_staged` | Doğrulanmış/yayınlanmış ready snapshot yok; klein ready işareti yok, schnell BLOCKED | 3600 |
| 503 | `engine_unavailable` | Upstream yapılandırılmamış veya supervisor yanıt vermiyor | 30 |
| 503 | `reservation_failed` | Ledger rezervasyonu başarısız | 5 |
| 429 | `queue_full` | Kuyruk veya lisans başına eşzamanlı iş sınırı (öneri: 1) dolu | 1 |
| 504 | `queue_timeout` | Kabul öncesi sınırlı slot beklemesi aşıldı; rezerv yapılmadı | 5 |

- 429/503/504 için `Retry-After`; mevcut backpressure stilindeki `X-US-Queue-Depth` / `X-US-Queue-Limit` korunur. Sayısal öneriler **NOT TESTED**.
- `(lisans, Idempotency-Key)` başına en çok bir image demand; tekrar POST aynı işi döner. Health/list/poll GPU'yu uyandırmaz ve idle'ı tazelemez. 202 polling sinyali boot talebi değildir.

### 4.4 Warm ≤30 sn hedefi: yalnız hedef, garanti değil

- **Kapsam:** yalnız image pipeline yüklü **ve** GPU slotu hazırken geçerlidir. Süre kabulden başlar; encode, persist ve yanıt dahildir. Yalnız sampler süresi sayılmaz. n=4 dahildir.
- **Hesap:** klein görsel başına 2–6 sn (**SOURCED ESTIMATE**). n=4 için 8–24 sn inference eder. Encode, çıktı ve yanıt için kalan ~6 sn dar olabilir. Kaynak sayılar hedefi makul gösterir ama **garanti etmez**. Ölçüm **NOT TESTED**; p95 ölçülmedi.
- **Kuyruk, cold veya swap dahil ≤30 sn taahhüdü bu mimariyle verilemez.**
- Demir warm ile cold/swap'ın ürün metninde ayrı anlatılmasını onayladı. Warm hedef, pipeline+slot hazır kabul→encode/persist→yanıt süresidir; cold/swap bekleme bu hedefe dahil değildir.
- Kuyruk/ETA ve kısmi iade ürün cümleleri **PENDING, sonra yazılacak**. Bu revizyon bunları yazmaz; süre/ücret garantisi vermez.


### 4.5 Kontrol-plane ve idle etkisi

- Şu sorgular GPU'yu uyandırmaz, idle süresini tazelemez ve busy lease almaz:
  - `GET /v1/credits/status`
  - `/health`, `/ready`
  - model listesi
  - `GET /v1/image/jobs/{id}` iş yoklaması ve `GET /v1/image/outputs/{ref}` yetkili indirme
  - `POST /v1/image` admission retleri (4xx, 429, 503, 504)
- Mevcut geniş nginx `GET /v1/` aktivite eşleşmesi kontrol sorgularını iş saymamalıdır (plan §9).
- Kabul edilmiş POST boot/swap başlatabilir; salt HTTP kabulü veya tekrar yoklama tek başına idle yenileme gerekçesi değildir. Idle yalnız **gerçek kabul edilmiş iş ve ilerleme** ile tazelenir: supervisor busy lease'i ve heartbeat (plan §9). HTTP access log'u bu amaçla kullanılmaz.
- `max_uptime` (1800) ve `daily_runtime` (3600) aktif lease nedeniyle ötelenmez.

---

## 5. Kredi, rezervasyon ve settlement (mevcut ledger)

### 5.1 Birim ve akış

- Yeni para birimi veya yeni birim yoktur. Mevcut **istem kredisi** kullanılır: `Kind "sub"` (kredi paketleri ve aylık plan). `CreditStatus.unit` değeri `"credits"` kalır.
- **Akış:** reserve, run, encode ve persist tamamlanması, ardından **tek settlement**.
- Tutar ve TTL'i **sunucu belirler.** İstemcinin `amount` veya `ttl` alanı `400 unsupported_parameter` döner. Bugünkü `/v1/credits/*` davranışıyla aynıdır.
- Mevcut `GET /v1/credits/status`, `POST /v1/credits/reserve` ve `POST /v1/credits/settle` kredi yüzeyi yeniden kullanılır; yeni para/ledger endpoint'i yoktur. Sunucuya ait image rezervasyonu sunucu içi mevcut ledger semantiğiyle yönetilir, HTTP settle istemcisine açılmaz.
- **Rezervasyon:** `amount = n × rate(model, size)`; `rate` tablosu **PENDING Demir**, sayısal çarpan yoktur. TTL **45 dk** (`ReservationTTL`, mevcut değer). Async bekleme/çalışma bu TTL'e sığmazsa sonsuz uzatma veya yeni kredi yok; terminal tutarlılık tasarımı doğrulanmalıdır (**NOT TESTED**).
  - Mevcut HTTP reserve'in sabit tutar davranışı image oranı değildir; image tutarını istemci belirlemez. Bu revizyonda reserve/settle çağrısı veya kredi kesme yapılmaz.

- **Settlement:** `model_meter.go` içindeki `settleAmount(key, id, amount, now)` semantiği kullanılır.
  - `0 ≤ amount ≤ reserved`.
  - `amount > 0` commit eder; tutulan rezervasyonun tamamı `Reserved` değerinden düşülür, yalnız `amount` `Credits` değerinden düşülür.
  - `amount = 0` release eder.
  - Aynı tutarla tekrar çağrı idempotenttir. Farklı tutar `ErrInvalid` döner ve settle çakışması (409) olarak ele alınır.
- **Rezervasyon kimliği istemciye ASLA dönmez.**
  - **Risk:** mevcut `POST /v1/credits/settle` lisans sahibine `{id, commit:false}` ile kendi rezervasyonunu serbest bırakma imkânı verir. Sunucunun sahip olduğu image rezervasyonu istemci tarafından release edilirse görsel ücretsiz alınabilir.
  - **Gereklilik:** image ve model-proxy rezervasyonları HTTP settle yoluyla settle edilememeli. Uygulama tasarımı ayrıca incelenmelidir.
  - Bu belge Store tipi tanımlamadığı için mekanizma uygulayıcıya bırakılmıştır. Durum: **NOT TESTED / UYGULANMADI**.

### 5.2 Kredi oranı: farklı oran kararı, sayısal tarife PENDING

- Demir, `t78s286`: “Hayır, model veya boyuta göre farklı kredi oranı olsun.” **1 üretilen görsel = 1 mevcut istem kredisi REDDEDİLDİ.**
- Oran model ve/veya `size`'a göre değişir: `rate(model, size)`. Sayısal tabloyu Demir ayrıca belirleyecek; **PENDING**. Burada TL, istem sayısı veya USD çarpanı türetilmez; eski 1:1 gerekçesi ve seçenek tablosu geçerli değildir.
- Sunucu rezervi `n × rate(model, size)`, teknik settlement'ı `produced × rate(model, size)` olarak belirler; kalan tutar serbest kalır. Sayısal tarife olmadan bu formüller uygulama veya ücretlendirme yetkisi değildir. Tamsayı ledger ile tarifenin gösterimi uygulama öncesi doğrulanmalıdır; ölçek/birim uydurulmaz.
- Boyutların süre/maliyet eşitliği **kanıtlanmadı / NOT TESTED**; mevcut plan maliyet tahminleri fiyat veya kredi tablosu değildir. Mağaza/canlı fiyat değiştirilmez.

### 5.3 Saatlik `use` planları

- Demir: “Saatlik (use) planlar ileride image ile eşlensin.” Eşleneceği kararlıdır; nasıl eşleneceği **PENDING**. Saniye→görsel dönüşümü veya sayısal oran tanımlanmaz.
- Şimdilik `Kind == "use"` → **`403 plan_kind_not_supported`**: geçici/beklemede durumudur, kalıcı ürün yasağı değildir. Rezervasyon ve wake yoktur. Mevcut HTTP reserve'in saatlik saniye ölçümünü server model proxy'ye bırakan 403 davranışıyla uyumludur.
- `Kind == "sub"` (paket + aylık istem) mevcut kredi yoludur (§5.1).

### 5.4 Kısmi başarı, iptal, OOM ve hata matrisi

`produced` = encode ve persist işlemleri **tamamlanmış** görsel sayısıdır. Persist başarısızsa görsel üretilmiş sayılmaz ve ücretlendirilmez.

| Durum | Settlement | Yanıt | Not |
|---|---|---|---|
| n/n başarılı | `settleAmount(n × rate(model, size))` commit | 200, `status:"succeeded"` | |
| k/n (0<k<n): deadline, motor hatası veya OOM | `settleAmount(k × rate(model, size))` commit; (n−k) × rate serbest | 200, `status:"partial"`, `partial_reason` | Aynı anahtarla replay ikinci ücret üretmez ve eksik görselleri tamamlamaz. Eksikler için yeni anahtarla yeni istek gerekir. |
| 0/n: motor hatası veya OOM | `settleAmount(0)` release | 502 `engine_error` veya `engine_oom` | Otomatik schnell denenmez (§6) |
| 0/n: persist hatası | release | 500 `persist_failed` | |
| Senkron yol: iş başlamadan istemci kopar | release | — (istemci yok; yanıt idempotency kaydına yazılır) | **NOT TESTED** |
| Senkron yol: bir görsel çalışırken istemci kopar | Çalışan görsel bitirilir (başlamış sampler boşa harcanmaz). Kalan görseller başlatılmaz. `settleAmount(tamamlanan × rate(model, size))` çağrılır; 0 ise release. | Kayda 200 `partial`, `partial_reason:"client_disconnected"` yazılır. Aynı anahtarla retry bu kaydı döner. | **NOT TESTED** |
| Çıktılar persist edildikten sonra, yanıt yazılırken istemci kopar | Tamamlananlar commit edilir (iş yapıldı) | Kayıtta 200; retry aynı referansları döner | **NOT TESTED** |
| Settle çakışması (ör. TTL aşımı, lisans iptali) | — | 409 `settlement_conflict`; çıktı referansları **dönmez** | Mevcut settle 409 ile tutarlı. Operatör incelemesi gerekir. **NOT TESTED** |
| Sunucu çökmesi veya rezervasyon TTL aşımı | Mevcut TTL ve süpürme davranışıyla serbest kalır | — | Ledger tutarlılığı ve crash recovery **NOT TESTED** |

Async 202 ile kabul edilmiş iş, POST veya poll bağlantısının kapanmasıyla iptal edilmez; aynı iş kaydı sürer. Yeni iptal endpoint’i tanımlanmaz. Teknik kısmi settlement `0..reserved` aralığında kalır; **kısmi iade ürün metni PENDING, sonra yazılacak**.

İkinci kredi hiçbir durumda düşülmez. Bir mantıksal iş için tek settlement yapılır.

### 5.5 Usage kaydı (rezervasyona bağlı)

- **Kaydedilecek alanlar:**
  - lisans (iç kimlik), idempotency HMAC, iş kimliği
  - model alias, HF repo, pinned revision
  - size, `n_requested`, `n_produced`
  - görsel başına seed'ler
  - `fallback.used` ve deneme sayısı, `partial_reason`
  - rezerve / commit / release edilen kredi
  - zaman aralıkları (admission, inference, encode/persist, toplam), boot-id ve GPU-owner fencing kimliği
- **Prompt metni ve görsel içeriği loglanmaz.** Yalnız uzunluk ve HMAC tutulur.
- **Mevcut durum:** `Settlement` kaydı yalnız `{CreatedAt, Key, Commit, Amount, SettledAt}` alanlarını taşır. Bu alanlar için yeni bir kalıcı kayıt gerekir. Biçimi bu belge **tanımlamaz** (yeni Store tipi uydurulmaz); ayrı tasarım ve onayla eklenir.
- `GET /v1/credits/status` içindeki `history[].amount` image settlement'larını tamamlanan çıktıların kredi tutarı olarak gösterir; görsel sayısı k ile eşit varsayılmaz (rate tablosu PENDING).
  - `committed` ve `released` alanları tutar değil, settlement **sayısıdır** (mevcut kod).
  - Geçmişte `source:"image"` gibi bir alan ileride eklenebilecek opsiyonel bir değişikliktir.

---

## 6. Fallback: görünür/otomatik model fallback kapalı

- **Varsayılan: otomatik model değişimi YOK.**
  - klein isteği klein ile, schnell isteği schnell ile çalışır.
  - Demir kullanıcıya görünür klein→schnell fallback açılmasını reddetti. Görünür veya otomatik fallback **kapalı kalır**. `allow_fallback` / `fallback` istek alanları `400 unsupported_parameter` döner. İstemcinin ayrı istekte açık `model:"schnell"` seçimi korunur; HF 401 BLOCKED aşılmaz.
- **Aynı model içinde sınırlı tek kurtarma:**
  - Aynı idempotency kimliği altında terminal hata ve o ana kadar başarılı görsel sayısı biliniyorsa, kalan görseller için en çok **bir** yeniden deneme yapılabilir.
  - Ön koşullar: süreç temizliği yapılmış ve VRAM serbest kaldığı NVML ile doğrulanmış olmalı; kalan iş deadline'ı, `max_uptime` ve `daily_runtime` yetmeli.
  - Bu deneme de başarısız olursa terminal sonuç döner.
  - Otomatik sonsuz tekrar ve ikinci kredi yoktur.
- **klein OOM:** daha büyük schnell'i "bellek kurtarıcısı" olarak denemek **yanlıştır** ve yapılmaz. schnell'in BF16 pipeline'ı ~35–37 GB'tır (plan §5). Sıra şöyledir:
  1. Süreç temizliği
  2. Kalan runtime kontrolü
  3. Model değiştirmeden kısmi sonuç veya 502 `engine_oom`; schnell otomatik denenmez
- **Şeffaflık:** yanıtta ve usage kaydında şunlar bulunur:
  - `model` = **gerçekten çalışan** motor (alias, repo, revision)
  - `fallback.used`, `fallback.attempts`, `fallback.reason`

---

## 7. Yanıt: `200 OK`

```json
{
  "object": "image.generation",
  "id": "imgjob_<sunucu-rastgele>",
  "idempotency_key": "7f3c2a1e-img-20261108-0001",
  "status": "succeeded",
  "partial_reason": null,
  "model": {
    "alias": "klein",
    "repo": "black-forest-labs/FLUX.2-klein-4B",
    "revision": "e7b7dc27f91deacad38e78976d1f2b499d76a294"
  },
  "params": { "steps": 4, "guidance": 1, "dtype": "bf16", "cuda_batch": 1 },
  "size": "1344x768",
  "n_requested": 2,
  "n_produced": 2,
  "seed": 123456789,
  "outputs": [
    { "index": 0, "ref": "imgout_<opak>", "content_type": "image/png",
      "width": 1344, "height": 768, "bytes": 0, "sha256": "<hex>", "seed": 123456789 },
    { "index": 1, "ref": "imgout_<opak>", "content_type": "image/png",
      "width": 1344, "height": 768, "bytes": 0, "sha256": "<hex>", "seed": 123456790 }
  ],
  "fallback": { "used": false, "attempts": 0, "reason": null },
  "credits": { "unit": "credits", "reserved": "<sunucunun belirlediği tutar; tarife PENDING>", "committed": "<tamamlanan çıktıların tutarı; tarife PENDING>", "released": 0 },
  "timing_ms": { "admission": 0, "inference_total": 0, "encode_persist": 0, "total": 0 },
  "retention": { "policy": "authorized_24h" }
}
```

- Örnek şematiktir: `credits` içindeki metinler sayısal oran uydurmamak için yer tutucudur, gerçek protokolde bu alanlar mevcut ledger ile uyumlu sayısal tutardır. `bytes` ve `timing_ms` sıfırları ölçüm değildir. `retention.policy = "authorized_24h"` §7.1 politikasını belirtir.
- `status` değeri `succeeded` ya da `partial`'dır. `partial` için `partial_reason` şunlardan biridir: `deadline`, `engine_error`, `engine_oom`, `client_disconnected`, `chat_priority`. `chat_priority` yalnız ileride istek ortasında kesme eklenirse kullanılır.
- `outputs` **referans + content-type** döner. Ana sözleşme büyük base64 gövdesi **değildir**.
  - `ref` tahmin edilemez ve opaktır.
  - Public URL, public ACL ve paylaşım linki **yoktur**.
  - İndirme ve 24 saat erişim politikası kararlıdır (§7.1); fiziksel yeni depolama yolu tanımlanmaz. Depolama veya indirme bu revizyonda uygulanmadı.

- `credits.reserved`, `committed` ve `released` bu çağrıya aittir. Rezervasyon kimliği **dönmez** (§5.1).
- `timing_ms` yalnız süre içerir. Prompt, seed dışı içerik ve dahili hata metni içermez.
- **Başlıklar:**
  - `Content-Type: application/json`
  - `Cache-Control: no-store`
  - `X-US-Metered: 1`
  - replay ise `Idempotent-Replay: true`
- Senkron 200 warm hazır yolu ifade eder. Async iş terminal sonucunun 200 ile okunması geçmişte cold/swap beklemesi olmadığı anlamına gelmez (§9).

### 7.1 Çıktı erişimi ve saklama: 24 saat, yetkili indirme

- Demir kararı: çıktılar **yetkili mevcut önek (yol uydurulmadı)** içinde persist tamamlanmasından itibaren **24 saat** erişilebilir tutulur. Plan/evre notundaki `media-models/` model ağırlığı önekidir; çıktı alanı olarak yeniden yorumlanmaz. Kaynaklar yetkili bir çıktı öneğini adlandırmadığı için burada yeni GCS/depolama yolu yazılmaz.
- `GET /v1/image/outputs/{ref}`: aynı Bearer lisans ve iş/çıktı sahipliği doğrulamasıyla yetkili indirme; kısa ömürlü, yetkilendirilmiş imzalı URL de kullanılabilir. Public ACL, public URL veya herkese açık paylaşım linki **yoktur**. Opak `ref` yetkilendirme yerine geçmez.
- 24 saat **erişim politikasıdır**, otomatik silme talimatı değildir. Sonrasında erişim taahhüt edilmez. **Otomatik silme, sweeper veya lifecycle işi yok**; silme mekanizması bu sözleşmede tanımlanmaz.
- İndirme de HTTPS, `Cache-Control: no-store`, `X-US-Metered: 1` kullanır; yeni kredi kesmez, GPU uyandırmaz, idle yenilemez. Bu rota yalnız tasarım, **UYGULANMADI**.

---

## 8. Hata kataloğu

**Gövde biçimi:** `{"error":{"code":"…","message":"<kısa Türkçe>","field":null,"reason":null,"retry_after":null},"credits":{"reserved":0,"committed":0,"released":0}}`

**Not:** mevcut `/v1/credits/*` ve model proxy yanıtlarının çoğu düz metin `http.Error` kullanır. Yalnız warming yanıtı JSON'dır. Image için JSON hata gövdesi bilinçli bir **farktır**; HTTP durum kodları mevcut anlamlarıyla aynı tutulmuştur. Dahili nedenler (ör. token okunamadı) istemciye gösterilmez; mevcut `writeWarming` ilkesiyle aynıdır.

| HTTP | `code` | Anlamı | Kredi |
|---|---|---|---|
| 400 | `invalid_json`, `unknown_field`, `unsupported_parameter`, `invalid_model`, `invalid_prompt`, `invalid_size`, `invalid_n`, `invalid_seed`, `missing_idempotency_key`, `invalid_idempotency_key` | Doğrulama hatası | Ayrılmaz |
| 401 | `license` / `session` | Kimlik yok, geçersiz veya Bearer değil | Ayrılmaz |
| 402 | `insufficient_credit` | Kullanılabilir kredi < n × rate(model, size) (n küçültülmez; tarife PENDING) | Ayrılmaz |
| 403 | `license_invalid`, `plan_kind_not_supported`, `reservation_inactive` | Lisans iptal edilmiş, süresi dolmuş veya geçersiz; `use` planı geçici eşleme beklemesinde; rezervasyon artık aktif değil | Ayrılmaz veya serbest |
| 405 | `method_not_allowed` | İlgili rota dışı yöntem (`/v1/image` için `Allow: POST`; yardımcı rotalarda `Allow: GET`) | — |
| 409 | `idempotency_key_reused` | Aynı anahtar, farklı istek | — |
| 409 | `settlement_conflict` | Settle tutarsızlığı (mevcut settle 409) | §5.4 |
| 413 | `body_too_large` | Gövde 32 KiB'ı aşıyor (öneri) | Ayrılmaz |
| 413 | `prompt_too_long` | Ölçülmüş token sınırı aşıldı. **Sınır ölçülene kadar ayrılmış koddur.** | Ayrılmaz |
| 415 | `unsupported_media_type` | JSON dışı içerik veya multipart | Ayrılmaz |
| 429 | `queue_full` | Kuyruk veya lisans başına eşzamanlılık sınırı (+`Retry-After`, kuyruk başlıkları) | Ayrılmaz |
| 503 | `not_ready` + `reason` (§4.3) | boot_pending, chat_priority (kuyruğa kabul edilemiyor), chat_reloading (kuyruğa kabul edilemiyor), daily_runtime, max_uptime, model_not_staged, engine_unavailable, reservation_failed; kabul edilen cold/swap/load/kuyruk işi 202 (§9) (+`Retry-After`, `X-US-Image-State`) | Ayrılmaz |
| 504 | `queue_timeout` | Slot için sınırlı bekleme aşıldı (+`Retry-After`) | Ayrılmaz |
| 500 | `internal_error`, `persist_failed` | Sunucu veya kalıcı yazma hatası, 0 çıktı | Serbest |
| 502 | `engine_error`, `engine_oom` | Motor hatası, 0 çıktı | Serbest |

`X-US-Image-State` (öneri) şunlardan biri olur: `CHAT_READY`, `CHAT_DRAINING`, `IMAGE_LOADING`, `IMAGE_READY`, `IMAGE_DRAINING`, `CHAT_RELOADING`, `COLD`. Değer salt okunur kaynaktan gelir ve GPU'yu uyandırmaz.

---

## 9. GET yardımcıları

### 9.1 Async kabul ve iş yoklaması (NOT TESTED / UYGULANMADI)

- `POST /v1/image` ertelenen işi kabul edince **202 Accepted** döner. Şematik gövde:

```json
{"id":"imgjob_<sunucu-rastgele>","status":"queued"}
```

- Başlıklar: `Location: /v1/image/jobs/{id}`, `Retry-After` (sunucu yoklama aralığı; ETA vaadi değil), `Cache-Control: no-store`, `X-US-Metered: 1`. Sayısal yoklama süresi bu revizyonda uydurulmadı.
- `GET /v1/image/jobs/{id}` aynı Bearer lisans ve iş sahipliğiyle okunur. İş kimliği opak/tahmin edilemez; başka lisansın işi veya rezervasyon kimliği açığa çıkarılmaz.
- İş durumları: `queued` (kuyruk), `starting` (boot/swap/load), `running`, `succeeded`, `failed`. Bunlar iş durumlarıdır; GPU sahipliği durumları §4.2'de ayrı kalır.
- Devam eden işin GET yanıtı 200 ile aynı `id` + güncel `status` ve `Retry-After` döner. Tekrar POST aynı anahtarla 202 + aynı işi döner; yeni rezerv/boot/swap yoktur.
- Başarılı terminal yoklama 200 ile §7'deki **aynı çıktı gövdesini** döner: `outputs`, kredi `reserved/committed/released`, `timing_ms`, `retention`; rezervasyon kimliği yok. Teknik kısmi başarıda sonuç gövdesinin `status:"partial"` değeri korunur (iş yaşam döngüsünde terminal başarı); yeni alan eklenmez.
- Başarısız terminal iş aynı `id`, `status:"failed"` ve §8'deki `error` + `credits` biçimiyle 200 durum okumasında raporlanır; bu 200 işin başarılı olduğu anlamına gelmez. Senkron POST terminal hatasında §8 HTTP hata kodu korunur.
- Yoklama **GPU'yu uyandırmaz, swap/boot başlatmaz, idle süresini yenilemez, busy lease/heartbeat üretmez**, yeni kredi kesmez. Health/list/poll salt kontroldür; yalnız kabul edilmiş POST image demand başlatabilir.
- 24 saatlik erişim sonrasında terminal replay yeni üretim başlatmaz; çıktı erişimi taahhüt edilmez (§7.1). Ayrıntılı kuyruk/ETA ürün metni **PENDING, sonra yazılacak**.

### 9.2 Diğer yardımcılar

- Yetkili indirme: `GET /v1/image/outputs/{ref}` (§7.1), aynı Bearer lisans. İstek gövdesi veya yeni üretim alanı eklenmez.
- `GET /v1/credits/status` bakiye okuması olarak kalır. Değişmez, GPU'yu uyandırmaz ve rezervasyon süpürmez (mevcut yorum: "A read neither sweeps reservations nor wakes a worker").
- `GET /v1/models` için ileride opsiyonel bir ek önerilir:
  - Satır biçimi: `type:"image"`, `id:"klein"` / `"schnell"`, `repo`, `revision`, `sizes`, `n_max:4`.
  - Statik allowlist'ten üretilir. GPU'yu uyandırmaz ve idle'ı tazelemez.
  - **İlk kapsamda listeleme yok.** Allowlist yalnız bu belgede ve `POST /v1/image` doğrulamasında tanımlıdır.
  - Mevcut model proxy'nin upstream `GET /v1/models` readiness probunu image listesiyle karıştırmamak gerekir.

---

## 10. Uygulama öncesi doğrulama listesi (hepsi NOT TESTED)

- Katı JSON ve alan matrisi: bilinmeyen alan, yasak alan, Unicode `×`, `n` sınırları, model allowlist, sessiz yeniden eşleme olmaması.
- Idempotency: eşzamanlı aynı anahtar, timeout sonrası kör retry, farklı payload (409), replay'de ikinci rezervasyon, boot veya swap olmaması.
- Ledger: n × rate rezerve, k × rate commit, (n−k) × rate serbest (sayısal tarife PENDING); settle idempotensi; HTTP `/v1/credits/settle` üzerinden image rezervasyonunun serbest bırakılamaması; TTL aşımı.
- Admission: kabul öncesi retler ve health/list/poll/indirmede wake/idle olmaması; kabul edilen POST cold/swap için tek demand, 202 ve aynı işin yoklanması; 3600/1800 sınırları; `Retry-After` ve kuyruk başlıkları; kontrol sorgularının iş sayılmaması.
- Async 202, terminal 200/hata, yetkili 24 saat indirme ve otomatik silme olmaması. Disconnect, OOM ve persist hatası matrisi (§5.4). Otomatik schnell denenmemesi.
- Warm uçtan uca süre: üç boyut × n=1 ve n=4 (yalnız onaylı bir test boot'unda; şu an **yok**).
- Prompt token sınırı ve Türkçe kalite (pinned tokenizer'lar; gizli dış çeviri yok).
- Prompt, secret veya müşteri verisi loglanmaz. Testte gerçek müşteri, lisans veya ödeme verisi kullanılmaz.

---

## 11. Yanıtlanmış kararlar ve kalan PENDING

Kaynak: Demir'in KAN.US sohbetinde `t78s286` kartına kendi sözleriyle cevapları; 8 Kasım 2026 ~04:34 TSİ (Europe/Istanbul, UTC+3), KAN.US aktardı. Yerel tasarım revizyonu; canlı uygulama izni değil.

1. **Kredi:** “Hayır, model veya boyuta göre farklı kredi oranı olsun.” 1:1 reddedildi; `rate(model, size)` farklılaşır. **Sayısal kredi tablosu PENDING Demir** (§5.2).
2. **Saatlik use:** “Saatlik (use) planlar ileride image ile eşlensin.” Geçici 403, kalıcı ret değil. **Eşlemenin nasıl olacağı PENDING**, dönüşüm uydurulmadı (§5.3).
3. **Async:** “Evet, async 202 ve iş yoklaması eklensin.” 202 + `GET /v1/image/jobs/{id}` tasarlandı; warm senkron 200 korunur (§4, §9).
4. **Saklama:** “Çıktılar yetkili prefix'te 24 saat tutulsun; imzalı/yetkili indirme olsun; public link ve otomatik silme olmasın.” 24 saat erişim, yetkili mevcut önek, opak ref; yeni fiziksel çıktı yolu uydurulmadı (§7.1).
5. **Fallback:** “Hayır, kullanıcıya görünür klein→schnell fallback kapalı kalsın.” Görünür/otomatik fallback kapalı, açık schnell model seçimi korunur; HF 401 BLOCKED aşılmaz (§6).
6. **Ürün metni:** “Warm ve cold/swap ürün metninde ayrı anlatılsın; kuyruk/ETA ve kısmi iade metni sonra yazılsın.” Warm/cold ayrımı işlendi; **kuyruk/ETA ve kısmi iade ürün metni PENDING, sonra yazılacak** (§4.4, §5.4).
7. **Boot/swap:** “Evet, POST /v1/image cold/swap durumunda boot veya swap başlatabilsin.” Kabul edilen POST için tek demand + 202; health/list/poll inert; daily_runtime=3600 ve max_uptime=1800 değişmez. **Şu anda GPU/VM/ds-a100/BOOT6 açılmaz; UYGULANMADI** (§4.3).

**Diğer açık noktalar:** İçerik güvenliği/moderasyon ayrı ürün kararıdır, kapsam dışı ve **PENDING**. Uygulama öncesi ölçüm/ledger/Store entegrasyonu, bounded async bekleme, yetkili fiziksel çıktı öneğinin teyidi ve uzun ledger kayıt ömrü doğrulanmalıdır; bunlar yapılmış sayılmaz. Plan dosyası salt okunur kaldı; r2 kararlarıyla eski planın async/fallback ifadeleri çelişirse bu sözleşmede r2 kararı uygulanacak tasarım olarak esas alınır, plan yeniden yazılmaz.


---

## 12. Etiket özeti

| Konu | Etiket |
|---|---|
| klein 2–6 GPU-sn/görsel (baz 3 sn); schnell 2–5 sn (kaynak baz 2,11 sn) | SOURCED ESTIMATE |
| Kredi oranı | Model ve/veya size’a göre farklı; sayısal tarife PENDING Demir, 1:1 reddedildi |
| İlk media-ready 464–611 sn | SOURCED ESTIMATE |
| Swap süresi (plan rezervi 60–180, baz 120 sn) | UNKNOWN |
| Warm ≤30 sn uçtan uca (n=4 dahil), p95 | NOT TESTED |
| 4 adım, guidance 1/0, BF16, batch 1 | NOT TESTED (başlangıç önerisi) |
| Prompt token sınırı; schnell CLIP/T5 kırpma davranışı | UNKNOWN / NOT TESTED |
| 32 KiB gövde, 90 sn iş deadline'ı, Retry-After değerleri, lisans başına 1 eşzamanlı iş | NOT TESTED (öneri) |
| PNG encode süresi ve boyutu; boyutların süre/maliyet eşitliği | NOT TESTED |
| Disconnect, OOM, persist, crash yolları; HTTP settle koruması | NOT TESTED |
| klein staging | KAN.US: private staged, 24 nesne / 15.980.183.023 bayt; ready ve deploy yok; US bağımsız canlı doğrulamadı |
| schnell HF erişimi | HTTP 401 BLOCKED, aşılmadı |
| Async 202 / poll / yetkili 24 saat indirme / POST boot-swap yeteneği | Ürün kararı tasarıma işlendi; NOT TESTED / UYGULANMADI |
| Günlük 200 görsel | Koşullu kapasite planı; chat dahil garanti değil (plan §8) |

**Son durum:** PLAN / TASARIM ONLY, UYGULANMADI. Kod, test, deploy, GPU, BOOT6 ve HOST işlemi yok. Canlı fiyat ve mağaza değişmedi. Satış veya .13 onayı değildir.
