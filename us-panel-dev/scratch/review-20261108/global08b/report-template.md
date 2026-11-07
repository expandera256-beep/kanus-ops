# PANEL GLOBAL 08B yeniden inceleme — 08.11.2026

**Karar: FAIL (iki LOW işlevsel regresyon).** Önceki MEDIUM FX01 ve LOW FX02 kapandı; bağımsız özgün 39/39 PASS. Panel tarafında bu incelemede yeni HIGH/MEDIUM kanıtlanmadı. HMAC tasarımının sınırı aşağıdadır. Bu bir deploy/ürün onayı değildir.

İnceleme yerel, taşınmamış artefaktlara ve önceki bağımsız problara bağlı olduğu için istenen scratch alanında yapıldı. Kanıt: `/workspace/us-panel-dev/scratch/review-20261108/global08b/`. Kaynak, pinned, dist veya yama dosyalarına yazılmadı. Mesaj gönderilmedi; EXE/Windows binary çalıştırılmadı; gerçek lisans sunucusu, api-test veya canlı servis çağrılmadı. TLS testleri yalnız httptest loopback sunucularında.

GLOBAL'in “Demir kararı, 00:48 TSİ” açıklaması KAN.US üzerinden iletilmiş, doğrudan Demir mesajıyla doğrulanmamıştır; onay sayılmadı. Kod istenen lisans-sunucusu tasarımı üzerinden incelendi.

## Girdi bütünlüğü

İlk SHA256 kontrolü dahil:
- core `e4f10a339c78818ceddbc49e2f687b4ee80afa86a7fce9d407c22bd8536892c7`
- 02 `314fab48b009f06e221ce140758724105b9f5b972c69d0c15c7ad2c9db8d3669`
- 03 `ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e`
- 04 `1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991`
- 06 `83aaca5d4a3aeb8cf7d4ed0e3854d0c3408d1553bb139ddb2eee22f8b5d013c6`
- 08 `9cd86d6bb6aa2a852361ac28e7d745b304647e10bc5cd71680709fca1781d04b`

Taban pinned/v10 manifest 181/181 PASS (`manifest.json`). Scratch kopyada core→02→03→04→06→08 check+apply 6/6 PASS. Bu yığının ayrı scratch kopyasında 05→07 check+apply 2/2 PASS; 05 SHA 7c917564…, 07 SHA 70acc9b0…. Gerçek pinned değiştirilmedi. 08 13 dosyaya dokunuyor; licensesrv, ödeme veya ledger dosyalarına diff yok. PROPOSAL uygulanmadı.

## Bulgular

### LOW FX08B-01 — geçerli büyük yanıt disk önbelleğine sığmıyor
`fx_panel.go:156-157`, `i18n/fx_cache.go:27-29`, `i18n/pref.go:65`.
Fetch 4096 bayta kadar gövdeyi kabul ediyor; cache aynı gövdeyi base64url ve JSON/HMAC zarfıyla büyütüyor, ama WriteRaw limiti yine 4096. Bağımsız geçerli 4096-bayt TLS yanıtı fetch'te PASS; refresh sonrası fxFromCache FAIL. WriteRaw hatası yutuluyor. Bellekte gösterim çalışır, yeniden açılışta offline önbellek kaybolur. Güvenlik kök dışına yazma değil, sözleşme/kalıcılık hatasıdır. Cache zarfı için ayrı sınırlı boyut veya bütün geçerli yanıtların sığacağı biçim gerekiyor; tercihler için genel limit gelişigüzel artırılmamalı.
Kanıt: `own-extra.jsonl` TestOwnCacheMatrix/max-body-cache; `own-followup.jsonl` TestOwn4096Fetch.

### LOW FX08B-02 — eski fiyat yanıtı yeni dili geri ezebiliyor
`gui.html:833-836,845`.
refreshPricing isteklerinde dil/nesil kontrolü veya iptal yok. EN ve ardından DE istenip DE yanıtı önce, EN sonra döndüğünde fiyat EN oluyor. Ana UI DE kalabilir. Deterministik gerçek fonksiyon VM probu: beklenen `DE price`, gerçekleşen `EN old price`, exit1 (`own-ui-race.mjs/.log`). Yanıt uygulanmadan etkin dil/istek nesli doğrulanmalı. Bu XSS değil; textContent kullanılıyor.

## Bağımsız sayısal/tarih probları

GLOBAL'in kopyası yerine önceki kendi `global-r4b/fx-probe/i18n/review_fx_test.go` dosyası ayrı i18n test ağacına alındı; hash `324fe6205173e24f92aeac6945056557f899bf6815a9ab03f548f50e8b5024b0`, kopya birebir. **39 yaprak/39 PASS**, 0 FAIL. Eski 30/9 artık 39/0. Üst testler toplama katılmadı.
Ek **13/13 PASS**: UTC+3 gece yarısı, tam 5 gün, 5 gün+1 saniye, beşinci gün sonu ve altıncı güne geçiş, gelecekteki gün; negatif/sıfır/çok küçük oran; 10^15 kuruş −1/tam/+1; negatif yarım sent.
Önemli: kural 120-saat süre limiti değil takvim günü; 5 gün+1 saniye aynı beşinci günde kabul edilmelidir. `i18n/fx.go:30-34,68-73` bunu sağlıyor. Oran [1,1000], çapraz oran [0.8,1.6], NaN/Inf reddi var. `:104-119` decimal-string big.Rat bölmesi yarım-sent pozitif sınırlarını düzeltmiş. Negatif tutarlar gizleniyor; -0.005 için negatif yuvarlama çıktısı yok. Yorumdaki “half away from zero” negatif destek iddiası diye yorumlanmamalı.

## Fetch, cache ve eşzamanlılık

Kendi `own-extra` **18 PASS/1 FAIL** yaprak; tek FAIL yukarıdaki cache boyutu. `own-followup` **3/3 PASS**. UI ters-sıra probu ayrıca **0/1 FAIL**. Sayılar standart suite'e eklenmedi.
- TLS matrisi **10/10 PASS**: HTTP reddi, 302 hedefi izlenmiyor (hedef hit0), 4097 bayt reddi, imzalı bozuk JSON, eksik/yanlış imza, altıncı gün imzalı replay reddi, geçerli yanıt, güvenilmeyen TLS sertifikası reddi, gerçek 5-sn client timeout.
- TLS client üretimde `fx_panel.go:43`, standart doğrulama; FX kodunda InsecureSkipVerify yok. Testin loopback CA güveni yalnız httptest client transport'u ile sağlandı; ayrıca güvenilmeyen sertifika standart transport'la reddedildi.
- Kaynak `currentFX:181` üzerinden activeGateway(); /api/pricing query'sindeki sahte gateway etkisiz (kendi testim PASS). Ancak mutlak “kullanıcı değiştiremez” iddiası doğru değil: mevcut `gateway.go:24-25` US_GATEWAY_URL ortam ayarıyla değiştirilebilir; normal kullanıcı tarafından düzenlenen process env ile derlenmiş varsayılan override edilebilir. HTTP parametresinden FX kaynağı alınmıyor. Bu yeni bir kullanıcı-girdili URL alanı değildir.
- 100 eşzamanlı currentFX çağrısı +100 ardışık çağrı: tek fetch; başarısız fetch sonrası 100 çağrı da tek fetch (5dk negatif bekleme). Rate/single-flight bu sınırlı problarda PASS, race uyarısı yok. Doğrudan unexported refreshFX çağrılarına single-flight garantisi ileri sürülmedi.
- fx.json geçerli cache kabulü; bozuk, imzasız ve eski imzalı cache reddi PASS. Okurken hem imza hem ParseFX/5-gün kuralı tekrar uygulanır (`fx_panel.go:127-141`); bellekte de her currentFX'te freshness denetlenir (:175-177).
- Dosya symlink'i ve pinned kök açıldıktan sonra kök rename+symlink swap problarında dış sentinel değişmedi (2/2 PASS). Bunlar Linux deterministik kontroller; her olası saldırı interleaving'inin ispatı veya Windows runtime testi değildir.
- Geçerli 5-gün penceresi içindeki eski imzalı snapshot replay'i monoton sürüm kontrolü olmadığı için mümkündür; sözleşme yalnız tarih tazeliği ister, burada “tüm replay engelleniyor” denemez.

## HMAC tasarım görüşü

`license.go:35-39`: sır env veya embeddedLicenseSecret üzerinden istemcide erişilebilir. Binary/process/makine erişimi olan biri sırrı çıkarabilir; yalnız panel arayüzünü açmış olmak tek başına anahtarı uzaktan elde etme kanıtı değildir. Sentetik anahtar kullanılarak sahte 40/44 kuru imzalanıp cache'e kondu ve fxFromCache kabul etti (TestOwnForgedCache PASS: tasarım zafiyetinin gösterimi, güvenlik garantisi PASS değil).
`fx_panel.go:64-77`: FX HMAC girdisinin başında `kanus.fx/1\n` alan ayırıcı var. `license.go:108-115,129-136`: lisans MAC'i aynı anahtarla base64url lisans gövdesi üzerine. Alan ayırıcı ve farklı mesaj dilleri sayesinde bir FX imzasını doğrudan lisans imzasına dönüştüren signing-oracle forgery kanıtı yok; HMAC anahtarını elde etmek ise hem sahte FX hem yerel legacy lisans imzalamayı mümkün kılar. Anahtarın istemcide olması önceden mevcut mimari risktir, “yeni FX doğrudan lisans forgery yarattı” denmedi. FX prefix'i var; ayrı türetilmiş/bağımsız anahtar yok. Ledger'ın tüm yollarının aynı anahtarı kullandığı iddia edilmedi; ödeme/ledger'a FX veri akışı bulunmadı.
Çağıranlar rg ile izlendi (`fx-callers.txt`, `hmac-callers.txt`): üretimde ApproxFX yalnız pricingSummaryFor gösteriminde; FX değeri charge, reserve, settle veya ledger tutarına gitmiyor. `licensesrv/pricing.go:13` eski USDTRYRef=50.05 hâlâ kodda; PriceUSD/us-pricing-gen/legacy lisans metadata hattına ait ve 08 bunu değiştirmiyor. Dolayısıyla “kodun hiçbir yerinde sabit kur yok” YANLIŞ; “FX gösterim fallback'inde sabit kur yok” DOĞRU. Kaynak/sır yokluğunda fiyat paragrafı yalnız TL/TRY; mevcut testlerle doğrulandı.
Öneri: FX için Ed25519; özel anahtar yalnız sunucuda, panelde sürümlü açık anahtar/key-id ve rotasyon; aynı tarih/boyut/TLS doğrulamaları korunmalı. Mevcut lisans sırrını yeni FX güven sınırına taşımamak gerekir. Ed25519 FX özgünlüğünü iyileştirir; mevcut legacy lisans-secret dağıtımını tek başına düzeltmez.

## PROPOSAL ve sözleşme (yalnız okundu)

`FX_CONTRACT_FOR_US_API.md:18,26-35,40-44` wire HMAC, takvim, fallback ve cache zarfını açıklar; 4096-gövde/4096-cache çelişkisi giderilmeli. Kaynak metadata'sı parser'da doğrulanmıyor; imza atanın gerçekten TCMB ForexSelling ürettiği ayrı sunucu sorumluluğu.
`global_fx_licensesrv_PROPOSAL.patch:46,57-58`: dosya tümüyle okunup sonradan limitleniyor; bounded read önerilir. Yalnız rates haritasının uzunluğu kontrol ediliyor, EUR/USD anahtarları, pozitif/range/cross doğrulaması yok; yanlış veri imzalanabilir (panel reddeder). Sunucu sözleşmesiyle eş doğrulama eklenmeli. Rota kaydı ve veri üretimi yok; uygulanmadı, derlenip çalıştırılmadı; sunucu kabulü NOT TESTED. HMAC prefix panelle eş.

## i18n / UI

TR/EN/DE/RU'da 5 price anahtarı ve placeholder kümeleri aynı (`i18n-check.json`). Fiyat özeti katalog TL değerleriyle dört dile taşınmış. `/api/pricing?lang=` salt metin JSON; render html.EscapeString, canlı yenileme textContent. applyLocale sonrası endpoint çağrısı var. Eski UI harness 4/4 PASS; check_ui PASS. Fakat yeni async ters yanıt sırası kapsanmıyordu; FX08B-02 açık. Native çeviri redaksiyonu NOT TESTED.

## Standart kontroller

{{RESULTS}}

Ortam: GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0. US_PLANS_JSON_* fixture'ları önceki yerel fixture kopyasından tree/reports altına alındı; geçici US_TASK_STATE_KEY_FILE scratch'te, kişisel anahtar kullanılmadı. TMPDIR kalıcı scratch; /tmp kullanılmadı. §2 aynı 36 senaryolu önceki kalıcı runner ile, yalnız scratch ağaçlarında.

## Sınırlar / kısa taslaklar — GÖNDERİLMEDİ

Gerçek /v1/fx, canlı sunucu, api-test, Windows runtime, macOS, tam .12+UYUM+API birleşik ürün kabulü NOT TESTED. 05→07 için burada yalnız apply-check ve kendi ek FX probları var; tam mail dalı suite kabulü ileri sürülmedi.

GLOBAL taslak: “08 9cd86d6b yeniden inceleme FAIL (iki LOW). Eski FX01/02 kapandı: kendi özgün39 vaka39/39, ek sınır13/13 PASS. fx_panel.go:156-157 + fx_cache.go:27-29: geçerli4096-byte yanıt base64/JSON cache zarfıyla4KB limitini aşıyor, sessizce persist olmuyor. gui.html:833-836: EN→DE isteklerinin ters yanıt sırası fiyatı eski EN'e çeviriyor. HMAC prefix var, fakat istemci sırrı sahte FX üretir; doğrudan FX→lisans imza dönüşümü kanıtlanmadı. Ed25519 öneriyorum. Sunucu PROPOSAL'da rate anahtar/range/cross doğrulaması eksik; uygulamayın. Tam kanıt PANEL_GLOBAL_REVIEW_08B_20261108.md.”

KAN.US taslak: “GLOBAL08 eski MEDIUM tarih ve LOW rounding düzeltmeleri kendi39/39 ve ek13/13 ile PASS; yama bütünü iki yeni LOW nedeniyle FAIL:4KB cache zarfı taşması ve ters sıralı fiyat-dil yanıtı. HMAC yerel sır sahibinin FX imzalamasını engellemez; Ed25519 önerisi var, ödeme/ledger'a FX akışı bulmadım. Lisans-sunucusu kaynak kararı yalnız sizin üzerinden iletilmiş, doğrudan Demir onayı doğrulanmadı. Bu rapor deploy/ürün onayı değildir.”
