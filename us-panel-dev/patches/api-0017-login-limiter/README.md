# API-0017 — US için login limiter öneri teslimi

Demir, 08.11.2026 01:28 TSİ (Europe/Istanbul):
> E-posta+IP ve artan bekleme olsun, ayrıca bilinen cihaz/oturum kilitten muaf olsun

## Dosyalar

- `LOGIN_LIMITER_DESIGN_TR.md`: ayrıntılı güvenlik/istemci sözleşmesi, parametre gerekçeleri, F04/F07 QA beklentileri.
- `api-0017-login-limiter-proposal.patch`: yalnız US'in inceleyip birleştireceği 6 licensesrv dosyası; kaynakta merge/deploy yapılmadı.
- `CHANGED_FILES.txt`, `BASE_MANIFEST.sha256`, `RESULT_MANIFEST.sha256`: kopya taban ve sonuç envanteri.
- `REGRESSION_TESTS_FIRST.go.txt`: RED'de çalıştırılan ilk 7 politika testinin gofmt edilmiş kesiti; sonradan eklenen 6 clock/bellek/proxy/iptal testinden ayrı.
- `logs/red.log`: ilk 7 test eski kodda derlendi; 6 FAIL, 1 PASS.
- `logs/green-first.log`: aynı ilk 7 senaryo yeni kodda 7 PASS.
- `logs/green.log`: genişletilmiş 13 politika + 4 güncellenen mevcut test = 17 PASS.
- `logs/qa-before.log`, `logs/qa-after-original.log`: değişmemiş QA risk repro öncesi/sonrası.
- `logs/build.log`, `logs/vet.log`, `logs/apply-check*.log`: komut exit kodları.
- `logs/race-licensesrv.jsonl`, `logs/race-other.jsonl`, `*.stderr`, `*.exit`: gerçek yarış testi çıktıları.
- `TEST_RESULTS_TR.md`: nihai gerçek test tablosu (tamamlanınca yazılır).
- `SHA256SUMS`: teslim dosyaları hash'leri (kendisi dışında).

Çalışma alanı: `/workspace/us-panel-dev/work/api/limiter/`; değişmeyen `base`, değişmiş `term`. Kaynak `/workspace/us-panel-dev/post12/term`, .12-tabanlı post12 kopyasıdır; pinned/v12'ye karşı farklar `logs/base-vs-pinned12.txt` içinde. Limiter ve accounts tabanları saf .12 ile hash eşit. Saf `.12` pinned ağacında da öneri patch `git apply --check` PASS; burada tam testler saf .12'de değil, post12 kopyasında çalıştırıldı.

## US uygulama sınırı

Korunan `/workspace/us-panel/` altına yazılmadı. Korunan licensesrv için yalnız dev kopyasında öneri hazırlandı. `cmd/us-license-hook/main.go`, `model_proxy.go`, ledger ve UYUM kablolamasına dokunulmadı. `.13` ortak deadline/idempotency çalışmasına müdahale yok. US, .13 üzerinde rebase/çakışma çözümü ve PANEL DEV/QA bağımsız kabulünü yapmalı; bu teslim .13 testi veya satış/deploy onayı değildir.

Kopya tabanda kontrol:
```sh
cd /workspace/us-panel-dev/work/api/limiter/base
git apply --check /workspace/us-panel-dev/patches/api-0017-login-limiter/api-0017-login-limiter-proposal.patch
```
Yeni bir **inceleme kopyasında** patch uygulanır. Mevcut `term`e tekrar uygulanmaz. SHA256SUMS için `sha256sum -c SHA256SUMS` kullanılabilir.

## Tekrarlanabilir test komutları

`term` içinde:
```sh
go test ./licensesrv -run 'Test(LoginPolicy|AccountLimiter|AccountIPRateLimit)' -count=1 -v -timeout=5m
go build ./...
go vet ./...
go test -race -json ./licensesrv -count=1 -timeout=20m
go test -race -json $(go list ./... | rg -v '/licensesrv$') -count=1 -timeout=20m
```
Bu iki race komutu birlikte `go list ./...` çıktısındaki tüm paketleri kapsar. Bağımsız komutların exit kodu/logu ayrıdır; tamamlanmayan test PASS sayılmaz.

QA orijinal betiği rapor yolunu `/workspace/us-panel/` altında oluşturduğu için **o konumda çalıştırılmadı**. Betik, `common.py` ve test dosyası içerikleri değiştirilmeden çalışma alanındaki `qa-repro/` dizinine kopyalandı; ROOT doğal olarak o kopyaya çözüldü. Önce `--tree .../base`, sonra `--tree .../term` çalıştırıldı. Hiçbir canlı HTTP isteği, e-posta, hesap veya model/GPU çağrısı yapılmadı.

Değişmemiş QA repro sonrası genel FAIL: account-lock güvenlik senaryosu PASS'a döndü, eviction-risk gözlem testi ise artık hedef silinmediği için FAIL verdi. Bu yeni güvenlik regresyonu gibi veya tüm QA PASS gibi sunulmamalı. Eski “fixed window” testinin PASS olması yeni algoritmanın sabit pencere kullandığını kanıtlamaz: assertion yeni quiet-window davranışını ayırt etmiyor. Redeem ve restart sınırlamaları hâlâ açık.

## Açık kabul / kararlar

1. Sayısal başlangıç parametreleri 5-pair/100-global, 1..900 sn, quiet 1800 sn; Demir politika yönünü onayladı, bu sayıları tek tek onayladığı iddia edilmez. US/QA kapasite ve kullanılabilirlik kabulü gerekir.
2. Sayaç kalıcılığı/replica paylaşımı **uygulanmadı**. Restart bütçeyi sıfırlar; çok-replica prod için ortak atomik TTL store gerekli. Yeni altyapı seçimi US/SRE'ye.
3. EXE'nin login sırasında geçerli sessionı Authorization ile göndermesi bu sunucu önerisinden ayrı istemci işi. Süresi dolmuş veya cihaz değiştiren kullanıcı bilinmeyen sınıra tabi.
4. Coarse IP/capacity 429 Retry-After:5 bir polling aralığı; tam pencere bitişi değil. Pair/global gerçek kalan süreyi döndürür.
5. Bilinen yanlış parolayı doğru paroladan ayırt etmek bcrypt gerektirir; geçerli tokenı olan saldırganın CPU tüketimini mevcut semaphore/503 kuyruğu sınırlar, tam ortadan kaldırmaz.
6. Redeem limiti ayrı öneri; bu patch uygulamıyor. Salt .12 ve .13 tam yığın kabulü, nginx gerçek-IP ve canlı TLS doğrulaması bu teslimde test edilmedi.

## US için kısa mesaj taslağı (GÖNDERİLMEDİ)

Alıcı: US (8bee53d9-d6df-45ee-8a5f-22fb2818bc47), iç bot kanalı.

Demir'in 01:28 TSİ tam kararı: “E-posta+IP ve artan bekleme olsun, ayrıca bilinen cihaz/oturum kilitten muaf olsun”. Öneri diff ve Türkçe sözleşme `/workspace/us-panel-dev/patches/api-0017-login-limiter/` altında. E-posta+IPv4 veya IPv6 /64; 5 yanlış sonrası 1,2,4…900 sn; yüksek email global eşik 100. Geçerli aynı hesap usa1 header/cookie + yalnız doğru parola tüm login kilitlerinden muaf; yanlışlar ayrı trusted pair'de sayılıyor. Parola reseti mevcut sessionVersion ile proof'u iptal ediyor. main.go/UYUM/model_proxy/ledger değişmedi. Test sonucu ve hash için TEST_RESULTS_TR.md/SHA256SUMS'a bak. .13 rebase ve bağımsız kabul sende; restart/shared-state, EXE proof taşıması ve redeem açık.

## QA için kısa mesaj taslağı (GÖNDERİLMEDİ)

Alıcı: QA (18c0ec38-b9c0-4f73-b853-9a63cd570fa9), iç bot kanalı.

Demir'in 01:28 TSİ kararı e-posta+IP, artan bekleme, bilinen cihaz/oturum muafiyeti. Yeni beklenti tablosu LOGIN_LIMITER_DESIGN_TR.md içinde. İlk 7 politika testi eski kodda 6 FAIL/1 PASS, yeni kodda 7 PASS; genişletilmiş hedef küme 17 PASS. Eski “doğru parola herkes için kilitliyken429” beklentisi bilinmeyen cihaza ve yeni global eşiğe göre güncellenmeli; bilinen doğru parola200, bilinen yanlış ayrı throttle. Orijinal repro öncesi FAIL; sonrasında account-lock testi PASS, live-target eviction artık olmadığı için eski risk gözlem testi FAIL. Orijinal reproyu değiştirip PASS diye göstermedim; ham loglar teslimde. Restart/replica ve redeem açık; canlı deploy iddiası yok. Bağımsız .13 kabulünüzü rica ederim.
