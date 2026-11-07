# UYUM v4.1 — PANEL DEV R4 iki LOW bulgu

08.11.2026 (Europe/Istanbul). **Yerel yama önerisi; canlıya uygulanmadı, yayın/merge/e-posta yok. Hukuki uyum onayı değildir.**

- Taban: `/workspace/us-panel-dev/pinned/v10` (dokunulmaz).
- Çalışma: `/workspace/us-panel-dev/work/uyum/v10_privacy_v4_1`.
- Temiz apply kopyası: `work/uyum/v10_privacy_v4_1_apply`.
- Kanıt: `/workspace/us-panel-dev/scratch/uyum-v4.1-evidence/`.
- Önceki v4 korunmuştur; v4.1 yerini alır. Bu set **v4 üzerine değil temiz v10/.11 kaynak tabanına** uygulanır.

## Uygulama ve US sahipliği

18 patch, dört klasör: `01-preferences → 02-ops → 03-handler → 04-hook-us`, klasör içinde dosya adı sırası.

```bash
cd /path/to/clean-source-copy
for d in /workspace/us-panel-dev/patches/uyum-v4.1/0[1-4]-*; do
  for p in "$d"/*.patch; do patch -p1 --batch --fuzz=0 < "$p" || exit 1; done
done
```

01/02'nin kodu v4 ile aynıdır. 03 HTTP politika/availability kodu ve testlerini; 04 US kancasını ve yeni route testini içerir. **02-ops ve 04-hook-us US alanıdır; kaynağa entegre edilmedi.** 04 yalnız mevcut prefs açılış/mount bölgesine availability iletimi, sabit uyarı ve ayrı readiness rotası ekler. API botunun dosyalarına dokunulmadı.

## R4 bulgu → düzeltme → test

| Bulgu | Düzeltme | Test |
|---|---|---|
| LOW-R4-01: exact allowlist geçersiz/boş port kabul ediyor | Tek `parsePrivacyOrigin` hem config yüklemede hem gelen Origin'de kullanılır. Scheme/host/port/yapı doğrulanmadan same-host veya allowlist kontrolü yok. Geçersiz config girdisi atılır ve **ham değer olmadan** indeksli sabit uyarı loglanır. | `TestV41OriginParserAndAllowlist` (1/65535 kabul; 0/65536/boş/negatif/non-numeric/taşan port ve bozuk host/yapı reddi), `TestV41AllowlistLoadLogsNoRawValueAndIsSnapshot`; bağımsız `TestReviewInvalidOriginPort` |
| LOW-R4-02: prefs açılış arızası erase404, healthz200 ile gizleniyor | `PrivacyUnavailable` ayrı durumdur: hata halinde erase request/confirm **kayıtlı kalır**, export/preferences ile beraber sabit503+Retry-After60+no-store verir; hiçbir erase/cleanup çalışmaz. `/readyz` degraded503, `/healthz` mevcut200 ve tam `ok` gövdesini korur. | `TestV41PrefsFailureRoutesReadinessAndLiveness`, `TestV41UnavailableRoutesAndDisabledMount`, `TestV41ReadinessAvailableOptionalAndDisabled` |

### Origin sözleşmesi

- `https` veya yalnız tam localhost/loopback IP için `http`; localhost HTTP **ayrıca exact allowlist** gerektirir. HTTPS aynı Host veya exact allowlist kaydı ile kabul edilir.
- DNS host ASCII etiketlerden oluşur (1..63 karakter/etiket, toplam en çok253; baş/son tire yok); IP adresleri parse edilir; IPv6 köşeli parantezli olmalıdır. IDN için punycode kullanılır. Zone-id, trailing-dot ve bozuk numeric host kabul edilmez. DNS çözümlemesi yapılmaz.
- Açık port canonical onluk1..65535 olmalıdır: boş port, +/−, non-numeric, leading-zero (`00080`) ve aşırı büyük değerler reddedilir. Port yoksa HTTPS/localhost HTTP varsayılan biçimleri geçerlidir. Varsayılan portlar otomatik eşdeğer sayılmaz: `https://x` ve `https://x:443` ayrı exact değerlerdir.
- Userinfo/path (son `/` dahil), query (`?` dahil), fragment (`#` dahil), opaque URL yoktur. Allowlist scheme/host/port kontrolünü bypass edemez.
- `Mount`/`NewHandlerFromEnv` startup'ta allowlist'i doğrulayıp değişmez snapshot alır. Config değişikliği yeni handler/restart gerektirir; her istekte hatalı config yeniden loglanmaz. Eski origin testindeki env değişikliğinden sonra handler yeniden oluşturuldu; geçerli localhost HTTP desteği kaldırılmadı. Standalone `originOK` politika yardımcısı aynı yükleyici/parser'ı kullanır.
- Boş Origin (non-browser) mevcut davranışla kabul edilir; bearer/parola/onay kontrolleri devam eder. `null` veya bozuk Origin reddedilir.

### Arıza, export ve readiness sözleşmesi

Configured prefs açılamazsa, POST export/preferences/erase-request/erase-confirm yanıtı:

```http
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
Retry-After: 60
Cache-Control: no-store

{"error":"privacy unavailable"}
```

Ham dosya yolu, env değeri ve hata ayrıntısı istemciye/loga yazılmaz. Configured prefs mevcutken export onun consent verilerini de içermelidir; arıza halinde sessizce kısmi export vermek yerine aynı503 döner. Arızada status bearer doğrulamasından önce sabittir; kişisel veri verilmez. Method/Origin güvenlik kontrolleri bundan önce uygulanır. Ledger retention endpoint'i bu sidecar arızasıyla kapatılmaz; admin auth gerektirir.

**Seçim:** Healthz yanıtını JSON'a çevirmek mevcut `200 + ok` gövdesi kullanan SRE/izleyicileri bozabilir; repo bunu değiştirmeye yetki veren bir sözleşme içermiyor. Bu yüzden `/healthz` **liveness** olarak birebir korunur; yeni `/readyz` ayrı privacy-capability readiness sunar. İzleme/alarm/LB readiness tüketicileri `/readyz` kullanmalıdır; canlı izleme ayarları burada değiştirilmedi.

- Arıza: GET `/readyz` →503, Retry-After60, no-store, `{"status":"degraded","privacy":"unavailable"}`.
- Configured prefs açıldı:200, `{"status":"ready","privacy":"available"}`.
- Prefs path hiç yapılandırılmadı (mevcut opsiyonel mod):200, `{"status":"ready","privacy":"not_configured"}`; prefs olmadan export çalışabilir, preferences503 döner.
- Erase kasıtlı kapalıysa ve prefs sağlıklı/opsiyonelse rota hâlâ404'tür. Configured prefs arızası olduğunda flag kapalı olsa bile unavailable503 gösterilir; hata ile ürün flag'ı karıştırılmaz.
- Readiness startup prefs açılış sonucudur; sürekli disk yazma/okuma veya model/GPU/SMTP sağlık probu değildir. Onarım/config değişikliğinden sonra handler/restart gerekir. Retry-After iyileşme garantisi değil istemci yeniden-deneme tavsiyesidir.

## Önceki korumalar ve sınırlar

v4 R3 düzeltmeleri korunur: Linux fd-relative openat/nofollow+Fstat, FIFO reddi, temp cleanup/fsync, commit-belirsizliği hata türü; doğrulama→idempotent prefs cleanup→kalıcı ledger işleminden sonra kod tüketimi; retention1..3650/varsayılan90; HMAC; normalize alıcı eşleşmesi; Close challenge temizliği.

Anonimleştirmede aktif lisanslar mevcut revokeLocked ile iptal, rezervasyonlar release, lisans PII scrub ve LicenseOwners bağı kaldırılır. Orders/önceki Settlements/Refunds korunur. Otomatik iade yok (`refundTriggered:false`). Kullanılmamış süre iadesi ve mali saklama süreleri **HUKUKİ/TİCARİ İNCELEME** gerektirir. 1..3650 operasyonel sınırdır, hukuki saklama kararı değildir. `US_PRIVACY_HMAC_KEY` üretim değeri üretilmedi/okunmadı/yazdırılmadı; HOST VPS env0600 kararı korunur.

Windows/non-Linux kalıcı prefs hâlâ **fail-closed / uygulanmamış**; çapraz build runtime işlevsellik iddiası değildir. Yapılandırılmış prefs açılamazsa artık404 yerine503+readiness degraded görülür. İki sidecar/ledger dosyası tek atomik transaction değildir; ledger kayıt hatasında önceden temizlenmiş prefs idempotent retry ile tekrar temizlenebilir. Paylaşılan mailde unrelated adresler korunur, ortak PII payload temizlenir ve tüm kayıt gönderime kapatılır. Daha ayrıntılı v4 tasarımı tarihsel `../uyum-v4/README.md` içindedir; bu README arıza/Origin sözleşmesi bakımından onun yerini alır.

## .11 teslim kontrolünün anlamı

Gerçek teslim dizini: `/workspace/us-panel/dist/2026.11.06.11/`. İçinde EXE/Linux binary/TGZ/rapor/SHA manifestleri vardır; **Go kaynak ağacı yoktur**. TGZ listesi yalnız dokuz deploy dosyası içerir. Bu dizine literal `git apply --check` kaynak dosyaları yokluğu nedeniyle tam geçemez; bunu PASS diye sunmuyoruz.

Dist `TREE_MANIFEST.sha256`, kaynak205 dosyasını tanımlar. `/workspace/us-panel-dev/pinned/v11` **205/205** bu manifestle birebir doğrulandı; değişmekte olan `/workspace/us-panel/work2/kiyas/term` yalnız165/205 eşleştiğinden sabit .11 kanıtı olarak kullanılmadı. Dolayısıyla anlamlı .11 kontrolü manifestle doğrulanmış pinned/v11 üzerinde salt-okunur --check ve ayrı kopyasında birikimli apply/vet/testtir. Dist, pinned/v10 ve pinned/v11 değiştirilmedi.

## Doğrulama sonuçları

| Kontrol | Nihai sonuç | Kanıt |
|---|---|---|
| `go vet ./...` | PASS / exit0 | `vet.log`, `vet.exit` |
| `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet ./...` + `go build ./...` | PASS / exit0 | `windows-vet.log`, `windows-build.log`, `windows.exit` |
| Temiz uygulanmış v10 kopyasında `go test -race -count=1 -json` (licensesrv, internal/privacy, cmd/us-license-hook ayrı süreçlerde) | **173 PASS / 0 FAIL / 1 SKIP**; üç exit0 | `race-licensesrv.jsonl`, `race-privacy.jsonl`, `race-hook.jsonl` ve `.exit` dosyaları |
| 18 patch 01→04, `patch -p1 --batch --fuzz=0`; her klasör sonrası üç paket `go test -run '^$'` | PASS, dört aşama derlenir | `cumulative-apply.log`, `stages.json` |
| `r11_base` salt-okunur `git apply --check` | **18/18 PASS** | `r11_base-check.log`, `apply-checks.json` |
| Literal binary dist dizininde `git apply --check` | **15/18 check PASS, 3 kaynak dosyası yok: NOT APPLICABLE**; dist kaynak ağacı değildir | `dist11_literal-check.log` |
| Dist manifestiyle doğrulanmış pinned/v11 kaynak ağacında --check | **18/18 PASS**, kaynak205/205manifest eşleşir | `dist11_manifest_source-check.log`, `v11-manifest.json` |
| Aynı sabit .11'in ayrı kopyasında birikimli check/apply + `go vet ./...` | PASS / exit0 | `v11-cumulative-apply.log`, `v11-cumulative.exit` |
| Aynı .11 kopyasında `go test -race -count=1 -json -run 'TestV41\|TestReviewInvalidOriginPort' ./internal/privacy ./cmd/us-license-hook` | **53 PASS / 0 FAIL**, exit0 | `v11-probes.jsonl`, `v11-probes.exit` |
| Değişen Go dosyalarında gofmt | PASS / boş çıktı | `gofmt.log` |
| pinned/v10 değişmezliği | **181/181 aynı** | `pinned-before.json`, `pinned-after.json` |
| pinned/v11 son manifest / dist artefakt SHA256SUMS kontrolü | **205/205 aynı / SHA PASS** | `summary.json`, `dist-sha256-check.log` |
| Çalışma ve temiz apply kaynakları | Teslim18dosya birebir aynı | `summary.json` |

Race sayıları test/alt test action toplamlarıdır; tek SKIP mevcut `TestStoreWriterProcess` subprocess helper'ıdır. İlk sıralı race koşusu assertion FAIL olmadan **SIGTERM/exit143** ile kesildi; nedeni kanıtlanmadı (`race-interrupted.*`). Eksik koşu başarı sayılmadı; ardından temiz apply ağacında üç paketin tamamı ayrı süreçlerde yeniden çalıştırılıp exit0 verdi. İlk dar denemede startup config snapshot davranışı nedeniyle eski localhost testi handler'ı tekrar oluşturmadan env değiştiriyordu; fixture yeni sözleşmeye göre güncellendi, nihai tam paket koşuları geçti (`initial.log` tarihsel kayıttır).

Ortam: GOPROXY/GOSUMDB off, gerçek model/network test bayrakları0, US_GPU_WAKE_DISABLE=1. Gerçek e-posta, model veya GPU çağrısı yok. Bağımsız eski R4 port probu .11 öneri kopyasına aynen kopyalandı, teslim setine dış probe eklenmedi; yeni V4.1 testleri teslim setindedir.

**Açık kalan:** US'in02/04 entegrasyonu, PANEL DEV v4.1 bağımsız onayı ve nihai API/GLOBAL birleşik kabul. Native Windows kalıcı prefs uygulanmamış sınırı sürer. SRE canlı ayarı, native Windows runtime, gerçek SMTP/model/GPU, full API/GLOBAL entegrasyonu ve §2'nin36 senaryosu bu işte çalıştırılmış sayılmaz. PANEL DEV'in v4 Linux PASS'i v4.1'in bağımsız onayı değildir.
