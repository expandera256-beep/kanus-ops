# PANEL GLOBAL inceleme R4B — final03 + FX08

İnceleme zamanı: 08.11.2026 00:51 TSİ (Europe/Istanbul). Kalıcı kanıt: `/workspace/us-panel-dev/scratch/review-20261108/global-r4b/`.
**03: PASS (Linux teknik inceleme). 08 FX: FAIL (MEDIUM tarih politikası + LOW gösterim sağlamlığı). GLOBAL 01–07 teknik PASS; 08 dahil paket FAIL. Kur kaynağı/ürün entegrasyonu PARTIAL.**
Kaynak/pinned/yama dosyalarına yazılmadı; mesaj/harici yayın/canlı servis işlemi yok; EXE çalıştırılmadı. Test mutasyonu yalnız scratch kopyasında.

## 1. Girdi bütünlüğü ve bağımsız interdiff

- İlk işlem03 SHA256 doğrulaması: **ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e**, mode **0444**, yazma bitleri0, effective_writable=false. Son kontrolde de aynı. `input_03_integrity.json`, `results.json`.
- f66 arşivi SHA **f66d4416bce2b5d61382fb097705104baf6e9d742e085cce9c0dd77933441a8f**. İki bağımsız pinned/v10 kopyasına core→02→eski/yeni03→04→06 uygulandı (her adım check+apply PASS). Python ayrı kanus kopyasına05→07,2/2 PASS. Orijinal yama baytları değişmedi.
- Pinned/v10 manifest başlangıç/son **181/181 PASS**. Pinned/v11 manifest **205/205 PASS**.
- Kendi `own.interdiff`: yalnız **5 dosya / 16 hunk**. GLOBAL'in interdiff'i ile yalnız dosya-başlığı zamanları çıkarılınca **birebir eş** (`interdiff_summary.json`). Dosyalar: pref.go, pref_handles_other_test.go, pref_handles_unix_test.go, pref_handles_windows_test.go, pref_rootcache_test.go.
- Üretim kodu farkı pref.go'da atomik sayımlar ve nil hook; open/close/refcount/OS dosya işlemlerinin mevcut çağrı sırası değişmedi. Atomik maliyet ve nil hook kontrolü eklendi; mutlak “hiç kod çalışması değişmedi” denmedi.

## 2. Sayaç mı, gerçek OS kaynağı mı?

**Bounded tek başına zayıflamıştır; paket bütünü yaygın kök başına sızıntıyı hâlâ yakalıyor.**

- `pref.go:126-129`: rootOpened yalnız başarılı openRoot sonrası artıyor.
- `pref.go:149-155`: r.closed=true, closeRoot, ardından rootClosed++ ve hook. `pref_unix.go:115` ile `pref_windows.go:241` void wrapper'ları OS close hatasını `_ =` ile yutuyor. Dolayısıyla rootClosed başarılı OS kapanış sayısı değil, **kapatma girişimi muhasebesi**; close hata verse de artar. Bu error-yutma f66'da da vardı.
- Bounded (`pref_rootcache_test.go:37-54`) yalnız iç muhasebe ve map boyutunu ölçüyor. `maxRoots=1` önbellek limiti; aktif referansı olan evicted kökleri erken kapatma sınırı değil.
- Scaling (`:61-80`) Linux `/proc/self/fd`, Windows GetProcessHandleCount ölçüyor. Isınma32; 64 ve512 kökte farkı karşılaştırıyor. Tek handle/kök sızıntısı **448**, tolerans32: bu regresyon için makul ayrım; sabit veya seyrek sızıntının yokluğunun matematiksel ispatı değil. Windows gürültü payı gerçek makinede ayrıca değerlendirilmeli.
- ClosedHandleInvalid (`:85-105`) kapanış sonrası **fcntl(F_GETFD)/GetHandleInformation** çağırıyor; sayaç okumuyor. 16 kapanış bekliyor. Helper herhangi bir API hatasını geçersiz sayıyor; daha sıkı test EBADF/ERROR_INVALID_HANDLE doğrulayabilir. Sayısal fd/handle yeniden kullanımı çok eşzamanlı ortamda yalancı pozitif üretebilir; mevcut test seri.
- Hook (`pref.go:58,154-155`, test `:88-94`) unexported ve üretimde nil; mevcut atamalar yalnız seri testte. i18n'de t.Parallel yok, test goroutine'leri join ediliyor; tam race'te yarış görülmedi. Hook atomic/mutex korumalı değil; ileride paralel setter/test eklenirse güvenli değildir. p.mu yalnız tek store'u korur; hook altında yeniden aynı store'u çağırmak kilitlenebilir. Mevcut üretim yolunda setter/callback yok; gerçekleşmiş üretim yarışı/deadlock bulgusu yok.
- Referans, eviction, eşzamanlı Close ve Close sonrası reopen için kendi eski probları final03 üzerinde tekrar3/3 PASS: aktif fd açık, sonrelease EBADF, ikinci Close başka dosyayı kapatmıyor, sonrakiSet/Get yeniden açıyor.

### Bağımsız mutasyon — gerçek kapatma kaldırıldı

`mutant-no-close/i18n/pref.go` içinde yalnız **closeRoot(r.h)** çıkarıldı; r.closed/sayaç/hook korundu. Orijinal yama değiştirilmedi.

| Test | Final03 | Kapatmasız mutant |
|---|---|---|
| Bounded | PASS | **PASS — tek başına sızıntıyı kaçırıyor** |
| HandleScaling | PASS,64:+0 /512:+0 | **FAIL**,64:+64 /512:+512 |
| ClosedHandleInvalid | PASS | **FAIL**,16 handle hâlâ geçerli |
| Kendi /proc64-kök probu | PASS,7→8 | **FAIL**,1304→1368 (+64) |

Seçilmiş koşu final7 PASS/0 FAIL, mutant4 PASS/**3 FAIL**, exit0→1 (`probe.log`, `mutant-no-close.log`). Bağımsız ek yaşam-döngüsü koşusu3/3 PASS; kendi fd probunun izole gözlemi **6→7** (`own-lifecycle.log`). Bunlar tam suite toplamına eklenmedi.

**LOW R4B-01 — test oracle/muhasebe açıklığı:** `i18n/pref.go:149-155; pref_unix.go:115; pref_windows.go:241; pref_rootcache_test.go:37-54`. rootClosed başarılı OS close garantisi değildir; Bounded tek başına PASS vererek gerçek sızıntıyı gizleyebilir (mutasyonla doğrulandı). Yeni iki OS testi kaldırılmamalı; isim/yorum girişim muhasebesi olarak netleştirilmeli veya close sonucu ayrıca izlenmeli. Bu koşuda gerçek final kodda sızıntı bulunmadı; merge engeli HIGH/MEDIUM yok.

## 3. PANEL DEV'in kendi kontrolleri — 08 olmadan

Ortam: GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0. US_TASK_STATE_KEY_FILE scratch içi; kişisel ~/.config anahtarı kullanılmadı. Fixture'lar önceki kalıcı `scratch/reports/US_PLANS_JSON_*` kopyasından çözüldü. TMPDIR kalıcı scratch/tmp.

| Kontrol | Sonuç |
|---|---|
| go build ./...; go vet ./... | PASS,0/0 |
| GOOS=windows vet/build | PASS,0/0 |
| Windows i18n test-c (yalnız derleme) | PASS,0; EXE çalıştırılmadı |
| gofmt -l bütün ağaç | Tabanın aynı30 dosyası; yeni biçimsiz dosya0 |
| Değişen Go dosyaları gofmt | PASS,boş çıktı |
| go test -race -count=1 -p=1 -json ./... | **422 PASS /0 FAIL /7 SKIP**,exit0,DATA RACE0 |
| §2 | **36 PASS /0 FAIL /0 BLOCKED**,exit0 |
| Node check_ui / UI mock | PASS /4/4 PASS |
| Python discover |79 PASS /0 FAIL /41 SKIP (120) |
| test_i18n_mail,US_GO_LOCALES finaltree |11/11 PASS |

Ham çıktılar `checks/`, `*.log`, `*.exit`, `results.json`. §2'nin kalıcı uyarlaması yalnız HERE/WORK/temizleme satırlarını değiştirir;36 senaryo/assertion aynı (`s2-runner.interdiff`). Kaynak manifesti §2 sonrası değişmedi.

## 4. v11 ve teslim sınırı

Pinned/v11 scratch snapshot'ına **core→02→final03→04→06** sırayla check+apply **5/5 PASS** (`v11-apply-check.log`); gerçek pinned'e yazılmadı. Bu .11+UYUM+API birleşik ürün kabulü değildir; başka executor'ün incelemesi burada tekrarlanmadı.

`/workspace/us-panel/dist/2026.11.06.11/` **VAR**; yalnız liste: SHA256SUMS, TESLIM_RAPORU_TR.md, TREE_MANIFEST.sha256, US_Ticari_Aday_2026_11_06_11.exe, us-license-hook_hostinger.tgz, us-license-hook_linux. Teslime dokunulmadı/çalıştırılmadı.

## 5. Harici Windows r4 — KAN.US/GLOBAL tarafından koşulmuş, PANEL DEV log incelemesi

**PANEL DEV'in kendi Windows koşusu: NOT TESTED.** Salt-okunur ham log incelemesi: global-work/ci/windows-results-DESKTOP-48D4KOL-20261108-r4/. Kanıt kopyası `windows-r4-evidence/`.

- binary-tests.log: **24 RUN /23 PASS /0 FAIL /1 SKIP**,sonuç PASS. Bounded,Scaling,ClosedHandleInvalid,EvictionWhileInUse,ConcurrentSwitch;normalSave,junction-root,junction-below,DACL,hardlink testleri PASS. TekSKIP TestWinPrefSymlinkFileRejected; environment.txt yetki yok/DeveloperMode kapalı,bunla tutarlı. 24 test adı `review.json` içinde.
- **İddia düzeltmesi:** hamlog `:26` Scaling64 **+6**,512 **+2** diyor; aktarılan +0 değil. d512-d64=-4≤32; test yine PASS.
- Mevcut CI binary'si SHA **166d869a5dad7e21985a5b910f894c4283af1799244653fce3f0c97104f974fe**; SHA256SUMS ve Tur9 ile eş. EXE çalıştırılmadan buildinfo ve nm okundu; final03'te yeni rootOpened/rootClosed/rootCloseHook ve iki yeni OS test sembolü var. Final03 ile **uyumlu**; ancak tam kaynak→binary build manifesti/reproducible-build kaydı yok, salt hash tek başına eksiksiz provenance ispatı değil.
- Ham iki dosyada **çalıştırılan binary'nin actual/expected hash stdout'u ve gerçek process exit kaydı yok**. run_windows_tests.ps1 hash doğrular; binary PASS için native exit0 bekler;1SKIP nedeniyle runner **RESULT exit=2(PARTIAL)** vermelidir (README ve script:76-78). Bu **beklenen/türetilen** sonuç, gözlenmiş exit kaydı değil. Tam konsoltranskripti olmadan hash→bu koşu ve gerçek exit bağı **BLOCKED/kanıt eksik**. Test logunda FAIL yokluğu PASS; provenance/exit için koşulsuz onay verilmedi.

**LOW R4B-W01 — harici kanıt eksikliği:** r4/environment.txt:1; binary-tests.log:26,51; run_windows_tests.ps1:96-110. Hash doğrulama stdout'u, build kaynak manifesti ve RESULT/OS exit transkripti arşivlenmeli. Bildirilen +0 yerine ham +2 kullanılmalı. Windows race'siz binary sonucu, Windows race PASS veya symlink testi PASS diye genişletilmez.

## 6. Değişmeyen sınırlar

04/06/07 önceki R4 ile aynıSHA; DE/RU katalog/mail kodunun önceki statik değerlendirmesi geçerli, UI/Python tekrarlandı. Gerçek SMTP çokdil entegrasyonu/HTML renderer, native dil redaksiyonu, macOS, canlı servis/deploy NOT TESTED. Yerel test fixture'ları gerçek fiyat/kur kaynağı diye sunulmadı. Önceki f66 raporu tarihsel; ba final03'ün NOT TESTED engeli bu Linux incelemesiyle kalktı.

## 7. Ek inceleme — 08 FX (TAMAMLANDI)

### Bütünlük / kapsam / PARTIAL

- 08 SHA **6ac3588bd36e76be008ea0e5d7180bbade76b15718cb9b27f592164668e5f363** başlangıç/son eş. Bağımsız v10 ve v11 yığınlarına check+apply PASS.
- Önce/sonra byte karşılaştırması yalnız **i18n/fx.go ve i18n/fx_test.go** ekliyor. Mevcut pricing.go, ödeme, ledger veya başka mevcut dosyada değişiklik **yok** (`fx-apply.log`). Var olan koda çağrı/entegrasyon eklenmemiş; fetch/kur alma yok. İki yığının bütün i18n kaynakları byte-eş.
- **PARTIAL:** Kur sağlayıcısı ve canlı/ürün entegrasyonu Demir'in kararını bekliyor. Kod notları TCMB/CBRT ForexSelling varsayıyor; bu, onaylanmış/verisi doğrulanmış canlı kaynak anlamına gelmez. `source`/`tcmb_tarih` metadata'sı ParseFX'te doğrulanmıyor; mevcut metin TCMB diye sabit. Farklı kaynak seçilirse metin/provenance sözleşmesi de değiştirilmelidir. Canlı TCMB/başka servis çağrısı yapılmadı.
- Kendi testlerinde kullanılan sayılar **sentetik, deterministik test fixture'larıdır; gerçek/güncel döviz kuru iddiası değildir**.

### ParseFX sözleşmesi

- Şema/base değerleri kanus.fx/1 ve TRY, YYYY-MM-DD biçimi+gerçek takvim tarihi; tam iki rates anahtarı ve EUR/USD zorunlu. Bilinmeyen üst alanlar (kaynak metadata'sı dahil) yok sayılır; strict/duplicate-key reddeden decoder değildir.
- Sayısal/string/null/eksik/negatif/sıfır kur denetimleri doğru; JSON NaN/Inf ve taşan1e309 reddediliyor. Kabul edilen kur aralığı [1,1000]; sonlu oranlardan EUR/USD **[0,8;1,6]** dahil, hemen dışı reddediliyor. Bölme0 yok.
- `json.Unmarshal` trailing ikinci JSON'u reddediyor; trailing boşluk serbest. Boyut **1..4096 byte**;4096 PASS,4097 reject. Clock `now time.Time` ile enjekte edilebilir;2030 tarihli fixture kendi enjekte saatiyle PASS. Ambient time.Now çağrısı yok.
- **Tarih kuralı FAIL:** tarih time.Parse ile UTC; kabul edilen yaş penceresi **[-36saat,+144saat]**. FXMaxAge5gün olsa da fazladan24saat ve36saat gelecek toleransı var. Bu, “gelecek tarih yok /5günden eski yok” şartını sağlamıyor.

### MEDIUM R4B-FX01 — gelecek / altıncı gün kabul ediliyor

**i18n/fx.go:46-52 (özellikle50)**. Kendi4 negatif vaka FAIL:

| Enjekte now (TSİ) | Snapshot date | Sonuç |
|---|---|---|
|08.11.2026 03:00 TSİ|2026-11-02|6gün eski olmasına rağmen kabul|
|08.11.2026 00:00 TSİ|2026-11-02|yerel takvimde6gün eski;141saat yaşla kabul|
|08.11.2026 15:00 TSİ|2026-11-09|gelecek tarih kabul|
|08.11.2026 15:00 TSİ|2026-11-10|36saat gelecekteki tarih kabul|

5.gün tam sınırı,5.takvim gününün TSİ gün sonu,bugünün TSİ00:01 saati; bunlar pozitif vakalarda PASS. Yalnız toleransı sıfırlamak yerine **bülten gününün zaman dilimi ve takvim-günü politikası** açıkça belirlenmeli: bugünden sonraki gün reddedilsin,bugün−5 dahil olsun,bugün−6 reddedilsin. UTC midnight ile yerel tarihi karıştıran geçici36saat toleransı çözüm değil.

### ApproxFX / dört dil / yuvarlama

- tr/en/de/ru ayrı yaklaşık tutar,ödemenin TL olması ve tarih/kaynak notları var. Teslim TestFXApprox4Langs dört dili gerçek çıktı karşılaştırmasıyla PASS. EN sembol yerleşimi $önce,diğerleri sonra; tarih locale'e göre. Yerel anadil redaksiyonu NOT TESTED.
- NegatifTRY/eksik snapshot gizlenir; bu yalnız gösterim fonksiyonu,ödeme/ledger'a bağlı değil. Güncelliği yeniden kontrol etmez; caller ParseFX'in geçerli snapshot'ını sağlamalıdır.

**LOW R4B-FX02 — float ara işlem ve int64 dönüşüm sağlamlığı:** `i18n/fx.go:84-94 (özellikle85-89)`.

1. ParseFX'ten geçen EUR=USD=2 fixture'ında29kuruş:0,145 birim→0,15 beklenirken **0,14** gösteriliyor. EUR=USD=6 ve21kuruş:0,035→0,04 yerine **0,03**. `/100`, `/rate`, `*100` ek float ara yuvarlaması yarım-sent değerini aşağı itiyor. İki bağımsız negatif test FAIL.
2. Geçerli parse edilmiş1/1 snapshot ve MaxInt64 TRYminor: pozitif miktar float→int64 taşmasıyla **negatif** gösterime dönüyor. Bu uç sağlamlık vakası gerçek kullanıcı fiyatı iddiası değil.
3. Elle kurulmuş NaN/Inf snapshot gizlenmiyor; NaN negatif devasa değer,Inf sıfır gösterimi üretiyor. **ParseFX bu ikisini zaten reddeder**; bunlar exported FXSnapshot/ApproxFX sınırı için defansif test, parser bypass iddiası değil.

En azından gereksiz/100×100 ara işlemi kaldırılmalı; yuvarlama politikası/decimal doğruluk ve int64 aralık kontrolü tanımlanmalı. Snapshot helper'da sonlu/geçerli kontrolü eklenebilir. LOW,çünkü yalnız yaklaşık gösterim,mevcut ürün/ödeme entegrasyonu yok.

### FX test matrisi (teslim testleri ve kendi negatif probları ayrı)

| Kontrol | v10+01–08 | v11+01–08 |
|---|---|---|
| check+apply08 |PASS|PASS|
| go vet ./... |PASS,exit0|PASS,exit0|
| gofmt -l yeni iki dosya |PASS,boş|PASS,boş|
| Bütün ağaç gofmt |tabanın aynı30dosyası,yeni0|tabanın aynı30dosyası,yeni0|
| tam go test -race -count=1 -p=1 -json ./... |**425 PASS/0 FAIL/7 SKIP**,exit0|**492 PASS/0 FAIL/7 SKIP**,exit0|
| DATA RACE |0|0|

FX'e özgü teslim3 testi PASS; bu, eksik sınır vakalarını kapsadıkları anlamına gelmez. Kendi `fx-probe` (v10) **39 yaprakvaka:30 PASS/9 FAIL**,exit1: tarih4,rounding2,sağlamlık3 FAIL. Üç üst testin ayrıca FAIL kaydı yaprak sayılarına katılmadı. Ham kanıt `fx-own-probes.jsonl`, `fx-own-summary.json`, yeni probun Go kaynağı `fx-probe/i18n/review_fx_test.go`. Tam suite toplamlara bu negatif testler gizlenerek karıştırılmadı. Tam v11+08 sonucu yalnız bu GLOBAL yığınına aittir; başka executor'ün ham .11/US/API/UYUM kabulünün yerine geçmez.

**08 kararı: FAIL; MEDIUM FX01 düzelmeden kabul önermiyorum. Kur kaynağı/bağlama kararı ayrıca PARTIAL.**

## 8. Nihai kayıt ve kısa mesaj taslağı (GÖNDERİLMEDİ)

Son kontrol: 08.11.2026 00:57 TSİ. 03/08 SHA hâlâ aynı; pinned/v10 **181/181**,pinned/v11 **205/205**. Kaynak/yama/teslime dokunulmadı. Bu raporun reports/ kopyası aynı baytlardır; özet GLOBAL satırı güncellendi.

**GLOBAL'e taslak:** “Final03 ba8845c5 Linux teknik PASS: tam race422/0/7,§2 36/36; gerçek close'u kaldıran mutasyonda Scaling/ClosedHandleInvalid ve kendi fd probum FAIL oldu. Bounded tek başına mutantı kaçırıyor; OS testleri korunmalı.08(6ac3588b) FAIL: fx.go:46-52 gelecek/6.gün tarihini kabul ediyor;:84-94 yarım-sent yuvarlama ve uç taşma var. v10/v11+08 tam race425/0/7 ve492/0/7 temiz; kendi39FX vakası30/9. Windows harici log23/0/1,Scaling+6/+2(+0 değil); actual-hash/build-manifest ve RESULT/exit konsol kanıtını arşivleyin. Kaynak kararı PARTIAL,henüz bağlamayın.”
