# FINDING-01 — Landlock / restricted-token Low IL azaltım adayı

Tarih: 2026-11-08 01:29 TSİ. Taban: .12. Çalışma kopyası: `/workspace/us-panel-dev/f01/term`.
**Durum: MITIGATION adayı; KABUL TAMAM DEĞİL. FINDING-01 AÇIK.** Canlı deploy, servis restartı veya gerçek GPU/VPS/Shopify/TCMB çağrısı yapılmadı. SMTP kapalı tutuldu. `/workspace/secrets/*` okunmadı/yazılmadı.

## Teslimatlar
- Yama: `/workspace/us-panel-dev/patches/f01/f01_landlock_lowil_on_12.diff`
  - SHA256: `a5f70b5f772f114f781d780435f9d631c21ab7b0d51d4a796cef9f316255c2bf`
- Windows native test paketi: `/workspace/us-panel-dev/f01/wintest_f01.zip`
  - SHA256: `deafd5e36501b0c75b1aae01b463087cb6ef23c232e44c74385310ed8d6895d2`
- Sürüm notu: `/workspace/us-panel-dev/patches/f01/SURUM_NOTU_F01_TR.md`
- Kanıtlar: `/workspace/us-panel-dev/f01/logs`
- Değişen dosyalar: `safe_exec.go`, `terminal.go`, `run.go`, `project_test_runner.go`, `pack.go`, `rea_safe_test.go`, `sandbox_env.go`, `sandbox_linux.go`, `sandbox_other.go`, `sandbox_windows.go`, `sandbox_linux_test.go`, `sandbox_windows_test.go`.
- `model_proxy.go`, `accounts.go`, gateway rezervasyon/idempotency, sembol bağlamı ve r13 Windows file-ops dosyaları değiştirilmedi; diğer .13 yamalarıyla birleştirilmedi. Ana work2 ağacına değişiklik uygulanmadı.

## Kök neden ve test-önce kanıtı
Aynı OS kullanıcısına ait çocuk süreçte cwd/argv/env/rlimit kontrolleri dosya erişimini sınırlandırmıyordu. İzinli Python proje kodu mutlak yolla kardeş workspace'i okuyabiliyordu.

| Kanıt | Önce | Sonra |
|---|---|---|
| QA audit12 `TestQAAuditExecSiblingIsolation` | `synthetic_sibling_read=true state=success` / FAIL | `synthetic_sibling_read=false state=failed` / PASS |
| Yeni `TestF01LandlockSiblingHomeAndWrite` | `OUTSIDE_READ_ALLOWED` / FAIL | kardeş ve sentetik HOME okuması EACCES; dışarı yazma EACCES; içeride yazma OK |
| Go araç zinciri | çalışıyordu | workspace içinde gerçek `go test ./...` PASS |
| PyInstaller doğrudan başlatma | ayrı repro `F01_BYPASS_READ` / FAIL | ortak güvenli başlatıcıya taşındı / PASS |
| Destek yokken kullanıcı uyarısı | Terminal/Çalıştır genel hata ile nedeni gizliyordu / FAIL | Terminal, Çalıştır ve proje testleri UYARI + Landlock nedenini gösterir / PASS |
| Re-exec başlangıç başarısızlığı | mevcut `TestSafeExecStartFailureIsReported` yeni helper ile `failed` verdi / FAIL | CLOEXEC status-pipe handshake; gerçek exec başarılı olmadan Start başarılı sayılmaz; test PASS |

Önce/sonra dosyaları: `qa_before_raw.log`, `qa_after_raw.log`, `qa_final_raw.log`, `f01_before.log`, `pack_before.log`, `warning_before.log`, `f01_final.log`.
QA'nın `run.py --tree ... --case file_isolation` koşucusu önce ve ilk fix sonrasında kullanıldı. Son kaynak aynı QA test overlay'iyle ayrıca doğrulandı (`qa-final-overlay.json`).

### MUTATION
Yalnız test overlay'inde helper'ın `applyLandlock(...)` çağrısı kaldırıldı; üründe kapatma bayrağı bırakılmadı.
`go test -overlay /workspace/us-panel-dev/f01/mutation/overlay.json . -run '^TestF01LandlockSiblingHomeAndWrite$' -v -count=1`
Sonuç: exit 1; `OUTSIDE_READ_ALLOWED`, beklenen FAIL. Üretim kaynak aynı testte PASS.
Dosyalar: `mutation/sandbox_linux.go`, `mutation/overlay.json`, `logs/mutation.log`.

## Linux uygulaması
- Bu makine: kernel `6.12.113+`, **Landlock ABI 6** (gerçek syscall sorgusu).
- Yalnız re-exec çocukta `runtime.LockOSThread`, `PR_SET_NO_NEW_PRIVS`, Landlock, sonra `syscall.Exec`. Ana panel kısıtlanmaz; ana süreçte dış sentetik dosyaya erişimin sürdüğü test edilir.
- Mevcut `golang.org/x/sys v0.30.0` sabitleri/struct'ları ve raw Landlock syscall kullanıldı; bağımlılık eklenmedi.
- Workspace'e dosya hakları; `/usr`, `/bin`, `/lib`, `/lib64` ve derleyicinin GOROOT'una yalnız okuma/exec. PATH'teki seçilen araç için yalnız kesin ikili dosyası açılır, **PATH üst dizini açılmaz**. HOME içindeki araç dosyasının yanındaki sentetik dosya okunamaz testi var.
- `/etc` bütünü açılmaz: yalnız ld.so.cache/ld.so.conf, nsswitch.conf, hosts, resolv.conf, localtime, ssl/certs, ssl/openssl.cnf. OpenSSL dosyası Node'un başlatılması için dar biçimde eklendi.
- `/dev/urandom` ve `/dev/random` okuma; `/dev/null` okuma/yazma istisnası. `/proc`, genel `/tmp`, gerçek `$HOME`, `~/.config`, `~/.ssh` ve kardeş workspace köklerine izin yok.
- HOME/temp/Go ve npm cache'i workspace altındaki `.us-exec` dizinine yönlendirilir. Paylaşılan GOMODCACHE kullanılmaz; özel araç kurulumlarının ek bağımlılıkları otomatik olarak genişletilmez, gerekirse komut hata verir.
- **Sessiz fallback yok.** ENOSYS/EOPNOTSUPP/ABI<1 uyarı+blok; ABI1/2 de TRUNCATE koruması olmadığından açıklamalı blok. Asgari etkinleştirme ABI3. Hata loga ve panel çıktısına taşınır.
- CLOEXEC pipe, Landlock ve gerçek exec sonucunu ebeveyne bildirir; başlangıç hatası `start_failed` kalır. Rlimit/iptal mekanizması korunur.

## Windows uygulaması ve native kabul
- CreateRestrictedToken: DISABLE_MAX_PRIVILEGE; Administrators SID deny-only. TokenIntegrityLevel: `S-1-16-4096` (Low).
- Mevcut Job Object korunur. Token, Start'ın CreateProcessAsUser çağrısından sonra kapatılır.
- Workspace etiketi: `S:(ML;OICI;NW;;;LW)` mandatory-label SACL; `SetNamedSecurityInfo(root, SE_FILE_OBJECT, LABEL_SECURITY_INFORMATION, nil,nil,nil,sacl)`. OI/CI alt dosya ve klasörlere kalıtım verir. Owner/DACL değiştirilmez. Etiket kalıcı metaveridir; gerçek panel komutu çalıştırırken proje etiketini değiştirir. Mevcut reparse point'ler önce reddedilir; hata halinde komut başlamaz.
- Native testler: Low token SID / admin deny-only / privilege drop; yeni ve mevcut workspace dosyasına yazma; açıkça Medium IL kardeşe yazmada ERROR_ACCESS_DENIED; dış sentetik dosyayı okuyabildiğinin açık kanıtı. go/node/python `version` çocukları Low IL ve Job'u kalıtır; yoksa açık TOOL_SKIP, varsa başlayamazsa FAIL.
- Windows amd64 ve 386 test binary derlemesi ve iki mimaride `go vet ./...`: **PASS**.
- **Native Windows koşusu NOT TESTED**: KAN.US `run.cmd` ve `run.cmd 386` çalıştırmalı; logları geri iletmeli. ZIP'te run.cmd, README_TR.md, SHA256SUMS ve iki mimarinin native test EXE'si bulunur.
- **Windows'ta okuma izolasyonu yok.** Low IL, yalnız daha yüksek integrity'ye yazmayı kısıtlar; başka Low IL konumlarına yazmayı veya ağı engelleyen workspace sınırı değildir.
- Reparse kontrolü + SetNamedSecurityInfo yol tabanlıdır; eşzamanlı root-swap için handle tabanlı etiketleme ayrı inceleme ister. r13 file-tools root-swap yaması bu yeni etiketleme yolunu otomatik olarak düzeltmez.

## Test sonucu ve açık doğrulama kapısı
| Kontrol | Sonuç |
|---|---|
| Son kaynak hedefli F01 + başlangıç regresyonları | 9 PASS / 0 FAIL (`f01_final.log`; F01 testleri 6 adet) |
| Son kaynak QA kardeş okuma repro | PASS (`qa_final_raw.log`) |
| Mutasyon | beklenen FAIL / exit 1 |
| Son kaynak Linux vet | PASS (`final_linux_vet.log`) |
| Son kaynak Windows test derlemesi amd64 + 386 | PASS (ikisi de exit 0) |
| Son kaynak Windows vet ./... amd64 + 386 | PASS (ikisi de exit 0) |
| gofmt | değişen dosyalarda temiz |
| Tam `go test ./...` | kabul tamam değil: önce 486 PASS / 7 SKIP / 1 FAIL (re-exec başlangıç sınıflaması); bu hata hedefli testte düzeltildi, son tam tekrar SIGTERM nedeniyle tamamlanamadı |
| Tam `go test -race ./...` / check.sh | SIGTERM exit143 ile kesildi; PASS iddiası yok |
| S2 eski fixture | 35 PASS / 0 FAIL / 1 BLOCKED: limit testinin çıktı dosyası workspace dışındaydı |
| S2 uyumlu fixture (handshake/son uyarı düzeltmelerinden önce) | **36 PASS / 0 FAIL / 0 BLOCKED**, `logs/s2_final/table.md` |
| Son kaynak S2 tekrarı | SIGTERM exit143 nedeniyle tamamlanamadı (`logs/s2_accepted_runner.log`) |

S2 fixture kopyasında yalnız L2 limit çıktısı proje içine taşındı; ölçülen CPU/DATA/NPROC kontrolü değiştirilmedi. Global `pkill sleep` temizliği kopya koşucudan kaldırıldı; başka ajan süreçlerine dokunulmadı. Asıl S2 planı/fixture'ları değiştirilmedi. Uyarlama: `/workspace/us-panel-dev/f01/s2_f01`.
`check.sh` kopyası `/workspace/us-panel-dev/f01/check_f01.sh` yalnız ortak `/tmp/testv.log` ve `/tmp/chk.exe` yollarını bu göreve özel yollara çevirir; test kapsamı değiştirilmez.

### Kesinti / yeniden kabul
Uzun koşular hem background hem uzun foreground Shell çağrısında tekrar tekrar **SIGTERM / exit143** ile sonlandı. Kesintiyi gönderen süreç belirlenemedi; ürün hatası veya platform nedeni olduğu iddia edilmiyor. Proje killProcessTree SIGKILL kullanır; bu tek başına dış kaynağı kesin kanıtlamaz. Sinyalden kaçınan nohup/disown/trap/alternatif yürütücü kullanılmadı.
Benzer Grok Bot yürütme kesintisi araştırıldı; bulunan kayıtlar çoğunlukla macOS local-exec içindi ve bu Linux box için kök nedeni kanıtlamadı: https://forum.cursor.com/t/grok-bot-local-execution-background-shell-exits-but-agent-chat-often-not-woken-silent-miss/176625 . Foreground bekleme önerisi zaten denendi; kesinti sürdü. Tekrarlar durduruldu; tamamlanmış kısa testler yeniden koşturulmadı.
Devam için `/workspace/us-panel-dev/f01/resume_acceptance.sh` yalnız sıralı, çevrimdışı testleri içerir; parent kesinti kaynağı netleştikten sonra çalıştırabilir. Tüm kabul testleri ve native Windows sonucu gelmeden bu aday için PASS/merge/deploy önerilmiyor.

## Açık işler ve sınıflama
1. Kesinti kaynağı netleştikten sonra son kaynakta full go test, -race ve S2 36/36 kabulünü tamamla; PANEL DEV incelemesi.
2. KAN.US'ın Demir'in PC'sinde Low IL native testleri (mevcut dosya yazımı ve go/node/python dahil).
3. FINDING-01 kapanmaz. Landlock etkin/testli Linux için teknik öneri MEDIUM; Windows için koşullu HIGH ve geçici risk kabulü. QA nihai sınıflama/kayıt sahibidir.
4. Landlock ağ veya tam container izolasyonu değildir; önceden açık descriptor/önceden workspace içine alınmış dosya alias'ları ve aynı-user saldırgan süreçlerin yarışları için tam mimari sınır garantisi verilmez. Windows Low IL read sınırı değildir.
5. AppContainer / proje başına ayrı OS kullanıcısı ileride yeniden değerlendirilecek. Mimari karar için KAN.US/Demir hattı.

## Sürüm notu (aynı metin)
FINDING-01 için bu sürüm yalnız risk azaltımı sağlar; bulgu kapanmamıştır. Linux'ta Landlock, proje kodunun kardeş çalışma alanlarını okumasını sınırlar. Windows'ta okuma izolasyonu yok: proje kodu çalışma alanı dışındaki dosyaları hâlâ OKUYABİLİR. Restricted token ve Low integrity level (Low IL), bu okuma riskini ortadan kaldırmaz. KAN.US üzerinden aktarılan karara göre Demir, FINDING-01 kapsamındaki bu Windows riskini 2026-11-08 tarihinde geçici olarak kabul etmiştir (seçenek e). AppContainer ve proje başına ayrı işletim sistemi kullanıcısı seçenekleri daha sonra yeniden değerlendirilecektir.
