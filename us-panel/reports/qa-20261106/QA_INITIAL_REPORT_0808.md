# QA Test Raporu — §8 / §9 / §12

Tarih: 06.11.2026 08:08 TSİ (Europe/Istanbul).

## Sonuç ve teslimler
- İlk teslim: `S12_TEST_MATRIX_20261106.md` oluşturuldu; §12 için 14 zorunlu testin tamamı önceki dosya/alıntılarla eşlendi (10 PASS yalnız yerel/fixture kapsam; 4 NOT TESTED).
- §8: mevcut mock HTTP/Go yük testleri ve api-test üzerinde düşük hacimli negatif kontroller gerçekten koşuldu. Tam güvenlik kabulü değildir.
- §9: yeni harness ve yeni mock-model proje üretim doğrulaması BLOCKED; mevcut gerçek model/rental harness temiz SKIP. Yeni mock E2E çalışmış gibi raporlanmadı.
- Ürün kaynağı değiştirilmedi. qa/ altında yalnız derleme çıktıları vardır; yeni harness kodu yoktur. Yeni harness gereksinimi tamamlanmadı.

## Araç engeli ve kapsam dürüstlüğü
Bu worker code-changes skillini yükledi: kapsamlı kaynak incelemesi ve yeni test kodu CloudAgent aracına devredilmelidir. GetDynamicTools(pattern=CloudAgent) sonucu matches=[] döndü. Bu nedenle kendi başıma ürün kaynaklarını inceleyip yeni harness yazmadım; mevcut test komutlarını çalıştırdım, raporları okudum ve güvenli uzak negatif kontroller yaptım. Parent, CloudAgent erişimli bir kodlama yürütücüsüne qa/ harness geliştirmesini devretmelidir. Bu araç eksikliği model tünelinden ayrıdır.
work2/source/deploy kapsamlı kod incelemesi ve satır bazlı yeni güvenlik açığı analizi yapılmadı. Aşağıdaki test adları gerçek koşum çıktısıdır; kaynak inceleme iddiası değildir.

## Bu tur kanıt tablosu
Ham loglar için kök: `/workspace/us-panel/reports/qa-20261106/` (aşağıda L). Her satırın PASS kapsamı açıkça belirtilmiştir.

| Gereksinim | Durum | Kanıt / sınır |
|---|---|---|
| §8 1/5/10/25/50 eşzamanlı kullanıcı | PASS | L/go-test.log:757–768: TestModelProxyLocalMultiuserAdmissionAndCreditConsistency tüm seviyeler PASS; yerel mock |
| §8 İlk token süresi | NOT TESTED | Ayrı TTFT metriği bu koşumda yok |
| §8 Toplam yanıt süresi | PASS | L/go-test.log:663: n=50 p50_ms=101.188 p95_ms=326.975; mock; 429 süreleri dahil |
| §8 Kuyruk bekleme süresi | NOT TESTED | Ayrı enqueue/dequeue zamanı ölçülmedi |
| §8 GPU açılış süresi | BLOCKED | Gerçek model tüneli/lifecycle kabulü gerekli; yalnız mock provider koşuldu |
| §8 Model yükleme süresi | BLOCKED | Gerçek model hazır olma ölçümü gerekli |
| §8 Hata oranı | PASS | L/go-test.log:663–664: HTTP200=33 HTTP429=17; unexpected_errors=0; beklenen kapasite reddi ayrı |
| §8 Kredi tutarlılığı | PASS | L/go-test.log:717–718: 1 success/49 HTTP402; negative_balance=0 double_spend=0 |
| §8 Sunucu maliyeti | NOT TESTED | Gerçek faturalama/tüketim verisi toplanmadı |
| §8 Kapasite üstü güvenli kuyruk | PASS | L/go-test.log:664: peak_admitted=33 active_capacity=1; :307 bounded queue cancel/release PASS |
| §8 Dosya/oturum/veri izolasyonu tam kabul | BLOCKED | Ayrı iki geçerli kullanıcı/lisans ve izole workspace gerekiyor; yerel symlink/terminal testleri tam tenant sandbox kanıtı değil |
| §8 Uzak path traversal | NOT TESTED | L/remote-security.log:44–68: query401, encoded URL400; sadece yetkisiz kapı kontrolü; yetkili dosya handlerı denenmedi |
| §8 Uzak command injection | NOT TESTED | L/remote-security.log:70–79: harmless echo query401; bu endpoint komut çalıştırmaz; gerçek terminal handlerı için kanıt değil |
| §8 Uzak cross-user isolation | BLOCKED | İki ayrı geçerli test kimliği verilmedi; kimlik uydurulmadı |
| §8 Uzak lisans sahteciliği negatif kontrol | PASS | L/remote-security.log:33–42: literal us1.qa.invalid ile models401; geçerli imza sırrı kullanılmadı |
| §8 Uzak token sahteciliği negatif kontrol | PASS | L/remote-security.log:22–31: literal qa-invalid-token ile status401 |
| §8 Uzak HMAC eksik/yanlış | PASS | L/remote-security.log:81–99: body{} ve eksik/yanlış imza HTTP401 unauthorized; webhook secret okunmadı/kullanılmadı |
| §8 Uzak rate-limit enforcement | NOT TESTED | L/remote-rate-limit.log:3–14: 12 istek concurrency2; tümü401; 429 gözlenmedi, bu düşük hızda limit bozuk sonucu çıkarılamaz |
| §9 qa/ yeni gated harness | BLOCKED | CloudAgent aracı mevcut değil; yeni harness yazılmadı |
| §9 Mock ile küçük gerçek proje üret/build/test | NOT TESTED | Mevcut file-action ve runner testleri PASS; bunlar yeni uçtan uca REST projesi üretimi değildir |
| §9 Gerçek model üretim/build/test/repair döngüsü | BLOCKED | L/go-test.log:414/417 gerçek model testleri SKIP; tünel+geçerli test lisansı+işlem/maliyet bütçesi gerekli |
| §9 Frontend/backend/PostgreSQL/auth/authz/proje/görev/API/test/Docker/CI-CD/dokümantasyon | NOT TESTED | Rental harness adı farklıdır; hiçbir küçük proje tam §9 kapsamlı kabul yerine konmadı |
| §12 go test ./... | PASS | L/go-test.log:815 exit_code=0; 301 üst düzey PASS,37 alt PASS,6 SKIP; yerel |
| §12 go test -race ./... | PASS | L/go-race.log:1–5: kiyas31.849s hook8.967s licensesrv65.833s; exit_code=0 |
| §12 go vet ./... | PASS | L/go-vet.log:1 exit_code=0 |
| §12 Derleme | PASS | L/build.log:3/5 linux_build_exit=0 windows_cross_build_exit=0; artifactlar qa/hook ve qa/US-QA.exe; native Windows değil |
| §12 API entegrasyon | PASS | L/go-test.log:601 CreditHTTPAuthorizationAndIdempotentSettlement PASS; L/python-test.log:11 yerel commerce HTTP PASS; uzak health200 |
| §12 PostgreSQL migration | BLOCKED | L/python-test.log:66–105 real PostgreSQL opt-in SKIP; geçici DB/DSN ve migration senaryosu gerekli |
| §12 Shopify webhook | PASS | L/go-test.log:605 RejectsBadHMAC; :607 PaidOrder...Idempotent; :788 ReversalHMAC PASS; fixture; Shopifya yeni yazma yok |
| §12 Kredi yarış durumu | PASS | L/go-test.log:591/603 son kredi yarış testleri PASS ve :717–718 50 parallel reserve |
| §12 SSE streaming | PASS | L/go-test.log:730 SSEChargeAndNoLicenseLeak; :739–746 truncated/error/length/timeout/cancel PASS; mock |
| §12 GPU yaşam döngüsü | PASS | L/python-test.log:21–44 fake provider yaşam döngüsü testleri PASS; gerçek GPU kabulü değil |
| §12 Windows EXE native | BLOCKED | Linux box; Windows makinesinde gerçek yürütme yapılmadı; build sonucu bunun yerine geçmez |
| §12 Güvenlik taramaları | NOT TESTED | govulncheck komutu mevcut değil; bağımsız dependency/SAST taraması koşulmadı |
| §12 Çok kullanıcılı yük | PASS | L/go-test.log:665/719/726/763; yalnız yerel HTTP mock, uzak ağır yük yok |
| §12 Büyük proje üretim | BLOCKED | §9 gerçek E2E koşulmadı; SKIP başarı sayılmaz |

## Çalıştırılan komutlar / ortam
Linux box, mevcut eşzamanlı düzenlenen work2 kaynak ağacı; sabit commit snapshot alınmadı, bu yüzden sonraki source/binary değişimiyle birebir aynı artefakt iddiası yok.
Go çalışma dizini `/workspace/us-panel/work2/kiyas/term`:
```bash
env GOPROXY=off GOSUMDB=off US_REAL_MODEL_E2E=0 US_REAL_MODEL_TEST=0 US_REAL_PROJECT_TEST=0 US_REA_NETWORK_TEST=0 US_DOOR_PROBE=0 go test ./... -v -count=1 -timeout=240s
# Aynı env ile:
go test -race ./... -count=1 -timeout=300s
env GOPROXY=off GOSUMDB=off go vet ./...
env GOPROXY=off GOSUMDB=off go build -o /workspace/us-panel/qa/hook ./cmd/us-license-hook
env GOPROXY=off GOSUMDB=off GOOS=windows GOARCH=amd64 go build -o /workspace/us-panel/qa/US-QA.exe .
cd /workspace/us-panel/work2/kanus
env US_POSTGRES_TEST_DSN= US_REAL_MODEL_E2E=0 US_REAL_MODEL_TEST=0 python3 -m unittest discover -v
```
Python: Ran109, OK skipped40, yani69 test çalıştı; L/python-test.log:112–115.
Uzak curl komutları ve sentinel URL/header değerleri L/remote-security.log içinde CMD başlıklarında kayıtlıdır. TLS varsayılan doğrulama açık, redirect takibi yok, her çağrı max15s. Toplam negatif smoke10 istek ve ek düşük hacimli12 istek; concurrency en fazla2. Rate-limit için üretim ayarları değiştirilmedi ve trafik artırılmadı.

## Hata / düzeltme / yeniden test
Bu tur FAIL veya doğrulanmış yeni güvenlik açığı yok. Kaynak düzeltmesi yapılmadı; dolayısıyla düzeltme-sonrası test iddiası yok. 401/400 saldırı önleme başarısını yalnız bu negatif girişler kapsamında gösterir, authenticated traversal/injection güvenliği anlamına gelmez.
Bilinen mimari risk: reports/KANUS_SECURITY_REPORT.md:11 güçlü OS/container kaynak ve kullanıcı izolasyonu yok diye işaretli; bu worker tarafından kaynak/satır doğrulaması yapılmadı. Yeni bir dosya/satır güvenlik bulgusu uydurulmadı.
Gereksinim farkı: prompt §4 idle varsayılan600sn isterken Python test adı test_default_idle_exactly_900_and_boot_600 geçiyor (L/python-test.log:21). Bu §8–§12 test başarısızlığı değildir; koordinatör kabul politikasını netleştirmeli.

## Engelleri açmak için
1. CloudAgent erişimli kodlama yürütücüsü: yalnız qa/ altında gated harness ve mock REST proje üret/build/repair testini geliştirsin; ürün kaynaklarını değiştirmesin.
2. KAN.US: iki ayrı geçerli kullanıcı/lisansı private dosyalarda sağlasın; bu worker mevcut secret/lisans dosyalarını okumadı. Lisanslar yalnız kanus-test; webhook secret istemiyoruz.
3. KAN.US yetkili VPS operatörü: 127.0.0.1:18000 tüneli ve gerçek model readiness doğrulasın; bu worker VPS yazmadı/restart etmedi. Loopback box ile VPS farklıdır; box18000 kontrolü VPS tünel kanıtı sayılmaz.
4. Gerçek §9 maliyet/işlem sınırı, geçici PostgreSQL, native Windows ve bağımsız security scanner erişimi gerekli.

Mevcut (yeni yazılmamış) gerçek rental harness komutu; bu tur ÇALIŞTIRILMADI:
```bash
cd /workspace/us-panel/work2/kiyas/term
GOPROXY=off GOSUMDB=off US_REAL_MODEL_E2E=1 US_PROXY_URL=https://api-test.kanus.org US_TEST_LICENSE=/absolute/private/test-license.txt US_REAL_MODEL_NAME=deepseek-v4.1-flash go test . -run "^TestRealModelPanelBuildRentalProject$" -count=1 -v -timeout=12h
```
Bu mevcut harnessin env sözleşmesi reports/ROUND6_RAPOR_TR.md:17–28 kaynağından alınmıştır; bu worker kaynak incelemedi ve mock ile bu E2Eyi doğrulamadı. Tam §9 proje yönetimi kabulü için genişletilmesi gerekir.

Canlı mağaza/ödeme/sipariş/müşteri e-posta/GCP/VPS değişikliği yapılmadı.

## Sayısal özet
Yukarıdaki 36 gereksinim satırı: PASS18 / FAIL0 / BLOCKED9 / NOT TESTED9. Bu sayılar üst düzey gereksinim kapsamıdır; Go301+37 ve Python69 test sayılarıyla karıştırılmamalı.
Teslim bazında: §12 matris PASS (mevcut kanıt envanteri hazır); §8 kısmi PASS (negatif auth/HMAC + yerel mock yük), tam kabul BLOCKED/NOT TESTED; §9 yeni harness BLOCKED ve mock proje E2E NOT TESTED.
