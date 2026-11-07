# Round 6 — yerel yük/güvenlik, proxy E2E ve Windows smoke

06.11.2026 07:20 TSİ (Europe/Istanbul)

## §8 — PASS (yalnız loopback fake SSE)
Gerçek ModelProxy/CreditHandler HTTP handler'ları httptest TCP sunucusunda çalıştı; sahte upstream ilk SSE chunk'ı flush edip gecikmeyle sonraki chunk'ı gönderdi. Test harness yerel JSON ledger kullandı. Bu standalone cmd/systemd veya gerçek GPU benchmark'ı değildir. Raw: LOAD_TEST_20261106.txt; her request'in index/HTTP/ms/SSE_DONE satırı ve aşağıdaki metrikler bulunur. Ana senaryo SSE tamamlanmasına kadar tüm request'leri ölçer; reject süreleri dahil. Gate kuyruğu deterministik doldurmak içindir, üretim SLA çıkarılmaz.

| Senaryo | Durum | Sonuç | p50 / p95 ms |
|---|---|---|---|
| 50 ayrı lisans, aktif1+queue32 | PASS | HTTP200=33; temiz HTTP429=17; unexpected=0; başarılı lisans başına1 debit, hold0 |104.970 /334.002|
| Tek lisans1 krediye paralel50 reserve | PASS | HTTP200=1; beklenen402=49; iki kez settle aynı sonucu korur; bakiye0/reserved0; double-spend0 |5.084 /6.136|
| Expired/revoked/hourly3 istek | PASS | 401=2,200=1; 1.1sn SSE için ceil=2sn,3600→3598; hold0 |0.240 /1105.780|

Reserve süreleri HTTP header'a kadardır. Üçlü auth/hourly p95, kasıtlı1.1sn sahte gecikmedir. Yetkisiz/lisansı iptal edilmiş istek upstream'e alınmaz. Kapsamlı penetration, gerçek TTFT/kapasite/uzun soak NOT TESTED.

## §9 — PASS kod/skip; BLOCKED canlı yürütme
TestRealModelPanelBuildRentalProject artık yalnız US_REAL_MODEL_E2E=1 ile açılır. US_PROXY_URL güvenli HTTPS (fixture için loopback HTTP) köküdür; bütün model çağrıları /v1/chat/completions üzerinden lisansla gider. US_TEST_LICENSE bir dosya yoludur, inline key değildir; en fazla4096byte UTF-8/BOM toleranslı tek satır okunur. Eski US_REAL_PROJECT_TEST tek başına bu testi açmaz. Aşama timeout50dk; gerçek panel queue/review/apply ve bağımsız proje HTTP acceptance akışı korunur. Env unset olduğunda temiz SKIP v3 raw'da görüldü. Gerçek çağrı bu tur yapılmadı.

Yetkili sonraki test için tam Bash komutu (bu tur çalıştırılmadı; model test kredisi tüketir):
```bash
cd /workspace/us-panel/work2/kiyas/term
GOPROXY=off GOSUMDB=off US_REAL_MODEL_E2E=1 \
US_PROXY_URL=https://api-test.kanus.org \
US_TEST_LICENSE=/absolute/private/test-license.txt \
US_REAL_MODEL_NAME=deepseek-v4.1-flash \
go test . -run '^TestRealModelPanelBuildRentalProject$' -count=1 -v -timeout=12h
```
Modelün ürettiği kod yürütülür; izole test ortamı ve ayrı operatör onayı gerekir. Tunnel/izinler tamamlanmadan çalıştırmayın. Private dosyanın içeriğini env'e, CLI'ye veya rapora kopyalamayın. Aşama başarısızlığı FAIL'dir; mock kabulü gerçek model kabulü sayılmaz.

## §10 — PASS kaynak/Linux CLI/cross-build; BLOCKED native Windows
Erken dispatch --version / --selftest / --check-license <file> / --check-server için newView, trial, GUI veya model başlatmaz. Version/selftest gerçek Linux main üzerinden offline çalıştırıldı; log SMOKE_RUN_LOG_20261106.txt. HTTP flag'leri yerel fixture testlerinde GET/status ve health, no-redirect, key-mask,401/403 ve dosya zorunluluğu ile doğrulandı. selftest embedded gui/config ve çözümlenmiş US_FIYAT placeholder kontrolüdür, canlı sunucu testi değildir.

Windows GUI subsystem yüzünden konsola güvenilmez: AttachConsole(parent) denenir, geçerli redirect korunur. Her smoke çağrısı ayrıca US_SMOKE_LOG'a (yoksa TEMP\US_smoke.log) plain text append eder. Writer log-first; console bulunmasa bile log yazımı testli. Log açılamazsa exit5, GUI açılmaz. Native Windows console API/süreç davranışı bu box'ta BLOCKED; Windows PE cross-build bunu kabul etmiş sayılmaz.

Demir PC'sinde EXE C:\US altında ve private lisans dosyası hazır olduktan sonra:
```powershell
$exe = 'C:\US\US_Ticari_Aday_2026_11_06_6.exe'
$license = 'C:\US\test-license.txt' # UTF-8 single line; key stays in this private file
$env:US_LICENSE_URL = 'https://api-test.kanus.org'
function Invoke-USSmoke {
    param([string[]]$ModeArgs)
    $env:US_SMOKE_LOG = Join-Path $env:TEMP ('US-smoke-' + [guid]::NewGuid().ToString('N') + '.log')
    $p = Start-Process -FilePath $exe -ArgumentList $ModeArgs -Wait -PassThru
    Get-Content -LiteralPath $env:US_SMOKE_LOG -Encoding UTF8
    "exit=$($p.ExitCode) log=$env:US_SMOKE_LOG"
}
Get-FileHash -LiteralPath $exe -Algorithm SHA256
Invoke-USSmoke -ModeArgs @('--version')
Invoke-USSmoke -ModeArgs @('--selftest')
Invoke-USSmoke -ModeArgs @('--check-server')
Invoke-USSmoke -ModeArgs @('--check-license', ('"{0}"' -f $license))
```
NOT TESTED — Yukarıdaki Windows/live-read komutları bu yürütücü tarafından çalıştırılmadı. Beklenen formatlar:
- --version: PASS version=2026.11.06.6; exit0.
- --selftest: PASS selftest assets=configured pricing=resolved version=2026.11.06.6; exit0.
- --check-server: health body ok ise PASS healthz HTTP=200; exit0. Model/GPU başlatmaz.
- --check-license: key=us1.*** remaining=<N> kind=<sub/use> expiry=<yyyy-mm-dd hh:mm:ss> TSİ valid=true; PASS license; exit0. Değerler gerçek sunucudan gelir. Key loglanmaz.
Exit2: arg/file/config/selftest; exit3: network/HTTP/redirect/invalid response; exit4:401/403 veya known expired/revoked valid=false; exit5: log açılamadı. Start-Process -Wait kullanılmalı, doğrudan GUI çağrısının konsol/Powershell beklemesine güvenilmemeli.

## Artifact / hook
PASS — dist/US_Ticari_Aday_2026_11_06_6.exe; SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327. PE amd64 GUI, .6 string/default api-test URL doğrulandı.
PASS — Hook production source/binary değişmedi (yalnız yeni licensesrv test dosyası var); yeniden build/package/deploy yapılmadı. Korunan Linux hook SHA7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf; tgz SHA0e599d2877087eb6efaf97f34fa711ad3263a18c272897d89cc52af88a6fa39a.

## Test/engel
PASS — Go289 üst düzey+37 alt,6 skip satırı,0 fail; tam race ./... exit0, vet exit0. Raw go_test/go_race/go_vet_topyekun_20261106_v3.txt; TEST_EXIT_STATUS_20261106_v3.json. Yeni kaynak listesi ROUND6_CHANGED_FILES_20261106.txt.
NOT TESTED — Atlananlar: PHP CLI yok; LiveDoor/OfficialREA opt-in kapalı; gerçek repair model ve yeni rental E2E opt-in kapalı. StoreWriterProcess yardımcı doğrudan skip, parent ayrı süreç kilidini gerçekten çalıştırır. Önceki Python109 sonucu değişmedi; bu tur yeniden çalıştırılmadı.
BLOCKED — Native Windows; VPS→127.0.0.1:18000 upstream tunnel ve aktif Go/GCP start/15dk idle-stop/watchdog bağlantısı. Önceki paket README'sindeki1923 örneği yeni18000 tunnel planının yapıldığını göstermez; runtime US_MODEL_UPSTREAM yetkili HOST tarafından uygun18000 endpoint'ine ayarlanıp kabul edilmeli. Bu tur hiçbir VPS/Shopify/GCP/SMTP/SSH çağrısı veya dış network write yapılmadı.
NOT TESTED — Round5 hook7061… deployed ve US_MODEL_NAME=deepseek-v4.1-flash bilgisi Demir/parent tarafından aktarıldı; bu yürütücü remote doğrulamadı. SMTP kapalı kalır. Önceki OS sandbox/tam hesap/crash-result replay açıkları kapanmış sayılmaz.
