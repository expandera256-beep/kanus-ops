# Latest EXE — 2026.11.06.7 accounts

06.11.2026 08:21 TSİ (Europe/Istanbul)
- PASS — Yerel Go304+37/race/vet, fake SSE account+ledger E2E; Linux headless main version/selftest; WindowsPE amd64 GUI cross-build. SHA256 3dc03a29a0ecb7b3ba55fe1a92e277d0ec05974617966e72c66b0834bf718cc9
- BLOCKED — Native Windows burada yok. Güncel tam PowerShell, maskeli account/credits çıktıları ve exit kodları WINDOWS_SMOKE_COMMANDS_20261106_TR.md; önce HOST yeni account hook route/env kurulumu gerekir. Eski .6go-hook ekranı değil, .7accounts ekranı beklenir.

| Durum | .7 native kabul adımı | Beklenen |
|---|---|---|
| BLOCKED | --version/--selftest/--check-server/--check-license/--check-account | GUI açılmadan log-first plain-text+exit; parola/key/session log'da yok; hash eşleşir |
| BLOCKED | Yeni hesap/register→login→redeem→balance→requests/workers | Aynı us1 tekrar eklenince artış yok; farklıhesap409; saniye/prompt ayrı |
| BLOCKED | Test modele sadece Go session proxy | GPU ısınıyor;45dk server/50dk client; gerçek SSEterminalden önce done yok; sub1ve hourlyceilsn; fail/cancelrelease |
| BLOCKED | Profil/altyapı | api-testGoHTTPS; eskiCloudRunenvyok, PCSSH/GCPyok; providerunknown durum gerçekGPUready sayılmaz |
| BLOCKED | Reset/sessionexpired ve source refund/cancel | Eski token tekrarlogin ister; source revoke accountbakiyesini kaldırır; ham us1 reset ile iptalolmaz |
| NOT TESTED | Gerçek model proje kalitesi/paid Bogus Gateway/SMTP | Ayrı Demir onayı + HOST/tunnel + Windows; SMTP0kalır |

## Önceki .6/.5 checklist (arşiv)

# Latest EXE — 2026.11.06.6

PASS — Cross-build/PE ve yerel CLI testleri; BLOCKED — Native Windows. SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327

Tam PowerShell komutları ve beklenen exit kodları ROUND6_RAPOR_TR.md / WINDOWS_SMOKE_COMMANDS_20261106_TR.md. İlk iki flag offline; GET health/status model başlatmaz. Key private file, log US_SMOKE_LOG; GUI subsystem için Start-Process -Wait kullanın.

## Önceki .5 checklist (arşiv)

# EXE smoke checklist — US_Ticari_Aday_2026_11_06_5.exe

Güncelleme: 06.11.2026 06:21 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Artifact Windows amd64 GUI PE; buildVersion .5 ve varsayılan api-test URL doğrulandı. Çapraz derleme Windows çalışması değildir.

```text
1ef0f70429953a79de7ea95f21b028f9fb0ad22412e9b399d792bc388423f863  /workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_5.exe
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/dist/us-license-hook_linux
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/deploy/us-license-hook/us-license-hook
0e599d2877087eb6efaf97f34fa711ad3263a18c272897d89cc52af88a6fa39a  /workspace/us-panel/deploy/us-license-hook_hostinger.tgz
```

| Durum | Native Windows kabul adımı | Beklenen gerçek sonuç |
|---|---|---|
| BLOCKED | Temiz Windows kullanıcısı, ilk açılış ve ikinci instance | Doğru .5 sürümü; eski .4 paneli yeni sürüm gibi açılmamalı; loopback host/origin koruması. |
| BLOCKED | Dosyalar/Projeler/AI/Editör/Günlük/Ayarlar/Terminal | Türkçe bölümler açılır; gerçek workspace; JSON/YAML/SQL/TS metin roundtrip, dış path/symlink ve stale edit reddi. |
| BLOCKED | Go lisansı ve bakiye | Python account formu/SSH butonu yok; key sadece Go proxy'ye; status gerçek sunucu zamanı/birim/bakiye. Bilinmeyen/iptal/expired lisans modelden önce reddedilir. |
| BLOCKED | US_GATEWAY_URL=yerel SSH ve dev flag ile ticari deneme | Ticari EXE yine Go API'ye gider; doğrudan model/SSH başlamaz. |
| BLOCKED | Model soğuk başlangıç / en az 45 dk | GPU ısınıyor/kuyruk, heartbeat; hazır/başarılı denmez. İptal ve timeout hold bırakmalı; gerçek token akışı sonrası sonuç. |
| BLOCKED | sub ve use modeli | sub başarı 1 kredi; hourly süre ceil saniye. Hata/kesik stream/cancel release, bilinçsiz retry yok. |
| BLOCKED | Terminal Python/Go/Node | Açık onay; cwd/stdout/stderr/exit code, başarısız komutun success olmaması; cancel/timeout alt süreçleri bitirir. Windows Job/CPU/RAM sandbox var diye sunulmaz. |
| BLOCKED | Proje agent/onay/test/onarım/resume | Model çok dosya üretir; apply onayı, transaction rollback ve checkpoint restart; gerçek test çıktısı/başarısızlığı görünür. |
| BLOCKED | Windows LockFileEx | Aynı ledger'a ikinci writer fail-closed; yeniden başlatma/atomik kalıcılık doğrulanır. |
| NOT TESTED | Lisanslı gerçek Bogus Gateway E2E | Önce Demir sipariş/ödeme onayı; kanus-test, gerçek kart yok; E2E_HAZIRLIK_TR.md. |
| NOT TESTED | Güncelleme/rollback/AV/code signing | Gerçek Windows paket dağıtımı, imza ve geri dönüş henüz kabul edilmedi. |

- BLOCKED — Tüm native Windows adımları bu Linux box'ta çalıştırılamadı. Bu checklist'in hazırlanması adımların geçtiği anlamına gelmez.

- BLOCKED — Windows yeni status özeti: GET /v1/credits/status kind/remaining/expiry TSİ gösterimi; önce yeni Go hook route'u yetkili deploy edilmiş olmalı. Native test bu box'ta yapılmadı.
