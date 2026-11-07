# §2 saldırı regresyon planı (06.11.2026, TSİ)

Kaynak: `PANEL_S2_REVIEW_20261106.md` bulguları. Bu plan kaynak ağaçlarına yazmaz.
- Otomasyon (Linux): `panel-evidence/review/regression/run_s2_regression.sh` ve PoC dosyaları `panel-evidence/review/regression/poc/`.
- .8 sonucu: `panel-evidence/review/regression/results/v8_20261106/table.md`.

**Karar kuralı:** Her senaryoda **PASS = güvenli davranış** (saldırı başarısız), **FAIL = açık hâlâ var**, **BLOCKED = PoC çalıştırılamadı** (derleme/API uyumu, eksik araç).

## 0. Koşucu nasıl çalışır ve .9 geldiğinde ne yapılır
```bash
# .8 (koşuldu) / .9 (ağaç geldiğinde):
/workspace/us-panel-dev/panel-evidence/review/regression/run_s2_regression.sh /workspace/us-panel-dev/work2/kiyas/term
/workspace/us-panel-dev/panel-evidence/review/regression/run_s2_regression.sh <.9 term ağacı> /workspace/us-panel-dev/panel-evidence/review/regression/results/v9_<tarih>
```
1. Verilen ağacın sha256 manifestini alır, ağacı `/tmp/s2reg.XXXX/term`'e kopyalar ve yalnız kopyaya `poc/*_test.go` ekler.
2. `go vet` ile derler. Derlenmeyen PoC dosyasını çıkarır ve o dosyanın senaryolarını BLOCKED yazar. Helpers derlenmezse çıkış kodu 2 olur.
3. `go test -race -count=1 -v -run '^TestS2Reg_'` koşar ve `S2REG|ID|SONUÇ|ayrıntı` satırlarından `table.md` üretir.
4. Koşu sonrası manifesti karşılaştırır: kaynak "DEĞİŞMEDİ" olmalı. Kopyayı siler, yetim `sleep 47/48` süreçlerini öldürür.
5. Çıkış kodu: 0 = hepsi PASS, 1 = FAIL veya BLOCKED var, 2 = kurulum hatası.

**.9 uyum noktaları:**
- `poc/s2reg_http_test.go` yalnız panelin dışa açık HTTP girişini kullanır: `POST /api/act` (op: `run`, `project-tests`, `terminal-start{text,n}`, `terminal-stop`) ve `GET /api/state` (`terminal.state/stdout/stderr`, `notice`). JSON alan adları değişmedikçe uyum gerekmez.
- `poc/s2reg_helpers_test.go` → `s2regServer` sunucuyu `newView`, `session.root/.browse` ve `guiMux` ile kurar. .9'da bu üçü değişirse **yalnız bu fonksiyon** güncellenir.
- `poc/s2reg_internal_test.go` iç API kullanır: `newSafeCmd`, `newFileTools/.Read/.resolve`, `taskState/pendingEdit/saveTaskState/runPersistentRepair/repairLoop/hashContent`. Değişirse koşucu dosyayı çıkarır ve H3-SAFECMD, M2–M4 ve L1-* BLOCKED olur. Ya PoC yeni adlara uyarlanır ya da senaryo HTTP girişi üzerinden yeniden yazılır.
- Senaryo ID listesi koşucudaki `ORDER` / `FILE_OF` tablosundadır. Yeni senaryo eklenirse oraya da yazılmalı.

## 1. Senaryolar (36 otomatik Linux + 9 Windows manuel = 45)

Sütunlar: ID | ilgili bulgu | ön koşul | adımlar | beklenen güvenli sonuç | PASS / FAIL ölçütü | otomasyon. Go testlerinin hepsi koşucu üzerinden çalışır.

### YÜKSEK
| ID | Bulgu | Ön koşul | Adımlar | Beklenen güvenli sonuç | PASS / FAIL | Otomasyon |
|---|---|---|---|---|---|---|
| H1-RUN | YÜKSEK-1 env sızıntısı | Panel ortamında `S2REG_SECRET`; projede `main.py` ve env dökümü yapan `.venv/bin/python` betiği | `POST /api/act {"op":"run"}` | Proje yorumlayıcısı ya çalıştırılmaz ya da yalnız `safeExecEnv` ile çalışır | PASS: dökümde sır yok. FAIL: sır görüldü | Go testi (HTTP) |
| H1-TEST | YÜKSEK-1 | Aynısı + `test_a.py` | `{"op":"project-tests"}` | Aynı | Aynı | Go testi (HTTP) |
| H2-TOOLEXEC | YÜKSEK-2 kabuk kaçışı | Go projesi | Terminal: `["go","build","-toolexec=sh -c 'touch PWN_TOOLEXEC'"]` | Komut reddedilir | PASS: `PWN_TOOLEXEC` yok. FAIL: dosya oluştu | Go testi (HTTP) |
| H3-RUN | YÜKSEK-3 taban adı allowlist'i | `.venv/bin/python` adlı keyfi betik | `{"op":"run"}` | Proje içi ikili onaysız çalışmaz | PASS: betik çalışmadı. FAIL: `.exec` işareti oluştu | Go testi (HTTP) |
| H3-SAFECMD | YÜKSEK-3 | Kök dışında `go` adlı betik | `newSafeCmd(..., ["/tmp/x/go"])` → Run | Reddedilir (çözülmüş ve güvenilir yol şartı) | PASS: hata ya da çıktı yok. FAIL: "NOT_REAL_GO" çıktısı | Go testi (iç API) |

### ORTA
| ID | Bulgu | Ön koşul | Adımlar | Beklenen | PASS / FAIL | Otomasyon |
|---|---|---|---|---|---|---|
| M1-SETSID-TIMEOUT | ORTA-1 setsid torun | python3 | Terminal `["python3","tree.py"]` n=2 sn. Betik aynı grupta `sleep 47` ve `start_new_session` ile `sleep 48` başlatır | Zaman aşımında tüm ağaç ölür | PASS: iki torun da ölü. FAIL: biri canlı | Go testi (HTTP) |
| M1-SETSID-CANCEL | ORTA-1 | Aynı | n=60, pid'ler yazılınca `terminal-stop` | İptalde tüm ağaç ölür | Aynı | Go testi (HTTP) |
| M2-TASKSTATE | ORTA-5 durum kurcalama | Go projesi | `.us-tasks/agent/t1.json` içine `phase=applying` + `PendingEdits` yazılır; `runPersistentRepair` (Approve=false) | Devamda bekleyen düzeltme yeniden onaya sunulur ya da imza kontrolü yüzünden reddedilir | PASS: `main.go` değişmedi. FAIL: enjekte içerik yazıldı | Go testi (iç API) |
| M3-HARDLINK | ORTA-4 hard link | Kök ve dış dosya aynı dosya sisteminde | Kök içinde dış dosyaya hard link → `fileTools.Read` | `Nlink>1` reddedilir | PASS: hata. FAIL: "OUTSIDE-SECRET" okundu | Go testi (iç API) |
| M4-SYMLINK-RACE | ORTA-4 TOCTOU | — | Bir goroutine `d`'yi 3 sn boyunca gerçek dizin ile kök dışı symlink arasında değiştirir; ana döngü `Read("d/f.txt")` | `O_NOFOLLOW` / `openat2(RESOLVE_BENEATH)` ile sızıntı 0 | PASS: kök dışı okuma 0. FAIL: ≥1 | Go testi (iç API). 0 sonucu yarışın kaybedildiğini de gösterebilir |
| L2-TESTS-RLIMIT | ORTA-2/3 Testler yolunda limit yok | python3 | `{"op":"project-tests"}`; unittest `getrlimit(CPU,DATA,NPROC)` değerini dosyaya yazar | CPU ve DATA sınırlı | PASS: ikisi de ≠ -1. FAIL: -1 | Go testi (HTTP) |
| (ORTA-2 NPROC) | Fork bombası | — | **Otomatikleştirilmedi.** Fork bombası yıkıcı. Yerine `L2` NPROC değeri ve kod incelemesi | cgroup `pids.max` ya da NPROC | Kod okuması | Elle inceleme |

### Kod okumasına dayalı notlar (Windows yolları), Linux'ta `fileTools.resolve` ile politika denetimi
Güvenli sonuç: aşağıdaki yolların hepsi reddedilir. Linux'ta bunlar sıradan dosya adları olduğundan **FAIL = politika henüz yok** demektir. Gerçek Windows etkisi W-6/W-7'de elle sınanır.

| ID | Yol | PASS / FAIL |
|---|---|---|
| L1-ADS | `a.txt:gizli` | PASS: resolve hata verdi |
| L1-ADS-DATA | `a.txt::$DATA` | Aynı |
| L1-DRIVE-REL | `C:evil.txt` | Aynı |
| L1-RESERVED-CON / -NUL / -AUX | `CON`, `sub/nul.txt`, `AUX.log` | Aynı |
| L1-LONGPATH | `\\?\C:\Windows\win.ini` | Aynı |
| L1-UNC | `\\server\share\x.txt` | Aynı |
| L1-DEVICE | `\\.\PhysicalDrive0` | Aynı |

### US'in yeni dar bayrakları (.9'da gelmesi beklenen politika)
Senaryoların hepsi Terminal üzerinden çalışır: `terminal-start` + Go testi (HTTP). Güvenli sonuç: komut **başlamadan reddedilir**.
- Yan etkili senaryolarda ölçüt, işaret dosyasının veya çıktının oluşup oluşmaması.
- Politika senaryolarında: PASS = başlamadı, FAIL = kabul edildi.

| ID | argv / ön koşul | PASS / FAIL ölçütü |
|---|---|---|
| N-EXEC | `["go","test","-exec=sh -c 'touch PWN_EXEC'","."]` | PASS: `PWN_EXEC` yok |
| N-TOOLEXEC-SPLIT | `["go","vet","-toolexec","sh -c 'touch PWN_SPLIT'","."]` | PASS: `PWN_SPLIT` yok |
| N-OVERLAY | `["go","run","-overlay=ov.json","."]`; ov.json `main.go` dosyasını kök dışı `evil.go` ile değiştirir | PASS: stdout'ta OUTSIDE_CODE_RAN yok |
| N-LDFLAGS | `["go","build","-ldflags=-linkmode=external -extld=sh"]` | PASS: reddedildi |
| N-GCFLAGS | `["go","build","-gcflags=all=-trimpath=x","-asmflags=-trimpath=x"]` | PASS: reddedildi (dar bayrak listesi dışında) |
| N-MODFILE | `["go","build","-modfile=other.mod"]` | PASS: reddedildi |
| N-GOFLAGS-ENV | Panel ortamında `GOFLAGS=-toolexec=<kök dışı betik>`, ardından `["go","build","."]` | PASS: betik işaret dosyası yok (GOFLAGS düşürülmüş) |
| N-CC-ENV | Panel ortamında `CC=<işaretleyen sarmalayıcı>`, `CGO_ENABLED=1`, cgo projesi; gcc gerekir, yoksa BLOCKED | PASS: işaret yok |
| N-PY-C | `["python3","-c","print(1)"]` | PASS: reddedildi |
| N-PY-C-AFTER | `["python3","a.py","-c","x"]` | Aynı |
| N-PY-M-OTHER | `["python3","-m","http.server"]` | Aynı |
| N-NODE-E | `["node","-e","1"]` | Aynı |
| N-NODE-EVAL | `["node","--eval=1"]` | Aynı |
| N-NODE-P | `["node","--test","-p","1"]` | Aynı |
| N-NODE-IMPORT | `["node","--test","--import=./a.js"]` (`=` biçimi) | Aynı |
| N-NPM | `["npm","run","x"]` (script `touch PWN_NPM`) | Aynı |

## 2. .8 ağacında gerçek koşu (06.11.2026 09:14–09:18 TSİ)
- Ağaç: `/workspace/us-panel-dev/work2/kiyas/term`.
  - §2 dosyaları ana kaynaktaki .8 ile birebir aynı (inceleme §1).
  - Bu ağaçtaki `release.go` hâlâ `2026.11.06.7`, çünkü sürüm artışı yalnız ana kaynakta yapıldı.
- Kaynak ağacı koşu sonrası: **DEĞİŞMEDİ** (sha256 manifest).
- **Toplam: 7 PASS / 29 FAIL / 0 BLOCKED.** Tablo: `results/v8_20261106/table.md`, ham log: `raw_test.log`.

**FAIL veren senaryolar (açık hâlâ var):**
- YÜKSEK: H1-RUN, H1-TEST, H2-TOOLEXEC, H3-RUN, H3-SAFECMD. Beklendiği gibi hepsi FAIL.
- ORTA: M1-SETSID-TIMEOUT, M1-SETSID-CANCEL, M2-TASKSTATE, M3-HARDLINK.
- **M4-SYMLINK-RACE: 2281 başarılı okumanın 639'u kök dışından geldi.** İncelemede bu bulgu yalnız kod okumasına dayanıyordu; artık kanıtlandı.
- L2-TESTS-RLIMIT: CPU ve DATA sınırsız (-1), NPROC 63991.
- L1-*: 9 yolun 9'u da reddedilmiyor.
- N: N-EXEC, N-TOOLEXEC-SPLIT, N-OVERLAY (kök dışı kod çalıştı), N-LDFLAGS, N-GCFLAGS, N-MODFILE, N-GOFLAGS-ENV.
- **Yeni açıklar:** N-NODE-P ve N-NODE-IMPORT. `--import=` biçimi ve `-p` engellenmiyor.

**PASS veren senaryolar:** N-CC-ENV, N-PY-C, N-PY-C-AFTER, N-PY-M-OTHER, N-NODE-E, N-NODE-EVAL, N-NPM.

## 3. Windows Job Object test planı (KAN.US, Demir'in Windows PC'sinde, PowerShell ile elle)

> Bu makinede çalıştırılmadı. Her şey `$env:TEMP\us-s2-wintest` altında yapılır ve sonunda silinir. Yıkıcı komut yok: yalnız bu klasör ve bu testte başlatılan süreçler kapatılır.
> Panel arayüzü adımları **elle** yapılır. Kullanılan adlar `gui.html`'deki gerçek öğelerdir:
> - Sol menü: **Dosya**, **Terminal**, **Önizleme**, **Servis**.
> - Terminal ekranı: "Komut" kutusu (`#terminalcommand`), "Bu proje kodunu çalıştırmayı onaylıyorum." kutucuğu, **Komutu çalıştır** ve **İptal et** düğmeleri, çıktı alanı.
> - Dosya ekranı: "C:\ veya D:\" yer tutuculu konum kutusu, **Aç** ve **Klasör seç** düğmeleri.
> - Önizleme ekranı, "Proje testleri": **Testleri çalıştır** ve **Testi iptal et**.
> - Sohbet alanı: **Çalıştır** ve **Durdur**.
>
> Arayüzdeki Terminal zaman aşımı sabit **120 sn**'dir (arayüz `n` göndermez).
> Ön koşul: Windows'ta Python 3 kurulu olmalı (`py` ya da `python` PATH'te).

### W-0 Ön koşullar ve hazırlık
```powershell
$Exe  = "C:\US\US_Ticari_Aday_2026_11_06_9.exe"      # .9 EXE yolu (gerçek yolu yazın)
$Sha  = "<US'in .9 raporundaki sha256>"
(Get-FileHash -Algorithm SHA256 $Exe).Hash.ToLower() -eq $Sha.ToLower()
$T = Join-Path $env:TEMP "us-s2-wintest"; New-Item -ItemType Directory -Force $T | Out-Null; Set-Location $T
py --version; if (-not $?) { python --version }
```
Beklenen sonuç: `True` ve Python sürümü. `False` çıkarsa **dur**; EXE doğrulanmadı.

Test betikleri:
```powershell
@'
import subprocess, time, os
p1 = subprocess.Popen(["ping","-t","127.0.0.1"], stdout=subprocess.DEVNULL)
p2 = subprocess.Popen(["cmd","/c","start","/b","ping","-t","127.0.0.2"], stdout=subprocess.DEVNULL)
p3 = subprocess.Popen(["powershell","-NoProfile","-Command","Start-Process ping -ArgumentList '-t','127.0.0.3' -WindowStyle Hidden"])
try:
    p4 = subprocess.Popen(["ping","-t","127.0.0.4"], creationflags=0x01000000, stdout=subprocess.DEVNULL)  # CREATE_BREAKAWAY_FROM_JOB
    br = "breakaway_started pid=%d" % p4.pid
except OSError as e:
    br = "breakaway_refused %s" % e
open("pids.txt","w").write("main=%d p1=%d p2=%d %s\n" % (os.getpid(), p1.pid, p2.pid, br))
print(open("pids.txt").read(), flush=True)
time.sleep(600)
'@ | Set-Content -Encoding UTF8 tree.py

@'
import time
chunks=[]
for i in range(1,33):
    chunks.append(bytearray(256*1024*1024)); print("ayrildi_MiB", i*256, flush=True); time.sleep(0.2)
print("BELLEK_SINIRI_YOK", flush=True)
'@ | Set-Content -Encoding UTF8 mem.py

@'
import os
print("SECRET_TEST" in os.environ and "SIZINTI_VAR" or "SIZINTI_YOK")
print(sorted(os.environ.keys()))
'@ | Set-Content -Encoding UTF8 env.py
```

Ortak kontrol fonksiyonları: ping ağacını listeleme ve Job üyeliği.
```powershell
function Show-Pings { Get-CimInstance Win32_Process -Filter "Name='ping.exe'" | Select-Object ProcessId,ParentProcessId,CommandLine | Format-Table -AutoSize }
Add-Type -Namespace W -Name J -MemberDefinition @'
[DllImport("kernel32.dll", SetLastError=true)] public static extern bool IsProcessInJob(IntPtr p, IntPtr j, out bool r);
'@
function Test-InJob([int]$procId) { $p = Get-Process -Id $procId -ErrorAction SilentlyContinue; if (-not $p) { return "yok" }; $r=$false; [void][W.J]::IsProcessInJob($p.Handle,[IntPtr]::Zero,[ref]$r); $r }
Show-Pings   # başlangıç: liste boş olmalı (ya da önceden var olanları not edin)
```

### W-1 Paneli env sırrıyla başlatma (env sızıntısı ön koşulu)
```powershell
$env:SECRET_TEST = "s2-win-fake-secret"
Start-Process -FilePath $Exe -WorkingDirectory $T
```
Elle: Panel açılınca **Dosya** ekranında konum kutusuna `$T`'nin tam yolunu yazın (`echo $T` ile görün), **Aç**'a basın, sonra **Klasör seç** ile bu klasörü proje yapın.
Beklenen sonuç: Dosyalar başlığı altındaki kök satırında test klasörü görünür; tree.py, mem.py ve env.py listelenir.

### W-2 Env sızıntısı (Terminal)
Elle: **Terminal** ekranında "Komut" kutusuna `["python","env.py"]` yazın (Python `py` ise `["py","env.py"]`). Onay kutucuğunu işaretleyip **Komutu çalıştır**'a basın.
Beklenen sonuç: Çıktıda `SIZINTI_YOK`. Anahtar listesinde yalnız allowlist anahtarları var (PATH, SYSTEMROOT, TEMP, TMP, USERPROFILE, APPDATA, LOCALAPPDATA, PATHEXT, COMSPEC, WINDIR, PYTHONIOENCODING…). `SIZINTI_VAR` görülürse **FAIL**.

### W-3 Job üyeliği
Elle: Terminal'de `["python","tree.py"]` → **Komutu çalıştır**. Çıktı alanında `main=… p1=… p2=…` satırı görünene kadar bekleyin. Sonra PowerShell'de:
```powershell
Get-Content .\pids.txt
$ids = (Get-Content .\pids.txt) -replace '[^0-9 ]',' ' -split '\s+' | Where-Object { $_ } | ForEach-Object {[int]$_}
$ids | ForEach-Object { "{0} inJob={1}" -f $_, (Test-InJob $_) }
Show-Pings
```
Beklenen sonuç:
- main, p1 ve p2 için `inJob=True`.
- `breakaway_refused` yazar (Job, BREAKAWAY_OK vermediği için). `breakaway_started` yazarsa **FAIL**.
- 127.0.0.3 ping'i (Start-Process ile kopan torun) da Job içinde olmalı. Pid'ini `Show-Pings` tablosundaki CommandLine'dan bulup `Test-InJob <pid>` ile doğrulayın.

### W-4 İptalde tüm ağacın ölmesi
Elle: W-3 çalışırken Terminal'de **İptal et**'e basın. Çıktı durumu "cancelled" olana kadar bekleyin.
```powershell
Start-Sleep 3; Show-Pings
Get-CimInstance Win32_Process -Filter "Name='python.exe' OR Name='py.exe'" | Where-Object CommandLine -match 'tree.py' | Select ProcessId,ParentProcessId
```
Beklenen sonuç: `127.0.0.1`–`127.0.0.4` ping'lerinin **hiçbiri** listelenmez; tree.py süreci yok. Herhangi biri kalırsa **FAIL**; o pid'i not edin.

### W-5 Zaman aşımında tüm ağacın ölmesi
Elle: Terminal'de tekrar `["python","tree.py"]` → **Komutu çalıştır**. Hiçbir şeye dokunmadan **125 sn** bekleyin.
```powershell
Start-Sleep 130; Show-Pings
```
Beklenen sonuç: Durum "timeout" olur; ping listesi boştur. Kalan ping varsa **FAIL**.

İsteğe bağlı aynı kontrol, Çalıştır yolu için: sohbet alanında **Çalıştır**, ardından **Durdur**. Not: Çalıştır yolu tree.py'yi değil, projenin algılanan giriş dosyasını çalıştırır. Klasörde tree.py'yi `main.py` olarak kopyalayın: `Copy-Item tree.py main.py`.

### W-6 Bellek sınırı
Elle: Terminal'de `["python","mem.py"]` → **Komutu çalıştır**.
Beklenen sonuç: Çıktıdaki son `ayrildi_MiB` değeri yaklaşık **4096'nın altında** kalır; süreç MemoryError ile ya da sıfır dışı bir çıkış koduyla biter. `BELLEK_SINIRI_YOK` görülürse **FAIL**. Bu arada Görev Yöneticisi'nde python.exe yaklaşık 4 GB'ı aşmamalı.

### W-7 ADS, ayrılmış adlar ve `\\?\` öneki (Terminal yol politikası)
Elle: Her birini Terminal'de deneyin. Beklenen sonuç: hepsi **başlamadan reddedilir** ("Yalnız proje içi göreli yollar…" ya da "Komut izinli değil").
```
["python","env.py:gizli"]
["python","CON.py"]
["python","sub\\nul.py"]
["python","AUX.py"]
["python","\\\\?\\C:\\Windows\\win.ini"]
["python","\\\\localhost\\c$\\Windows\\win.ini"]
```
Ardından PowerShell'de ADS ve ayrılmış ad yan etkilerini kontrol edin:
```powershell
Get-Item .\env.py -Stream * | Select Stream,Length     # yalnız ':$DATA' olmalı
Get-ChildItem -Force | Select Name                       # 'CON', 'NUL', 'AUX' adlı garip girdi olmamalı
```
Beklenen sonuç: Komutların hiçbiri çalışmadı; ek akış ya da ayrılmış adlı dosya yok. `CON.py`, `AUX.py` ve `sub\nul.py` aslında olmayan dosyalar. Komut "başladı" deyip Python "can't open file" hatası verirse bu **politika FAIL**'dir (reddedilmesi gerekirdi), ama zararsızdır; sonucu not edin.

### W-8 Testler yolu (limit ve sızıntı)
```powershell
@'
import unittest, os
class T(unittest.TestCase):
    def test_env(self):
        open("test_env_out.txt","w").write("SIZINTI_VAR" if "SECRET_TEST" in os.environ else "SIZINTI_YOK")
'@ | Set-Content -Encoding UTF8 test_env.py
```
Elle: **Önizleme** ekranında "Proje testleri" altındaki **Testleri çalıştır**'a basın, sonra `Get-Content .\test_env_out.txt` çalıştırın.
Beklenen sonuç: `SIZINTI_YOK`. Testler yolunun Job ya da bellek sınırı olmadığı Linux'ta kanıtlandı (L2); Windows'ta bilgi amaçlıdır.

### W-9 Temizlik
Elle: panelde çalışan bir işlem varsa Terminal'de **İptal et**, sohbet alanında **Durdur**. Ardından paneli kapatın.
```powershell
# Yalnız bu testte başlatılan ve 127.0.0.1–4 hedefli ping'ler:
Get-CimInstance Win32_Process -Filter "Name='ping.exe'" | Where-Object CommandLine -match '127\.0\.0\.[1-4]' | ForEach-Object { Stop-Process -Id $_.ProcessId -ErrorAction SilentlyContinue }
Remove-Item Env:\SECRET_TEST -ErrorAction SilentlyContinue
Set-Location $env:TEMP
Remove-Item -Recurse -Force (Join-Path $env:TEMP "us-s2-wintest")
Test-Path (Join-Path $env:TEMP "us-s2-wintest")   # Beklenen: False
```
**W-9'da ping kapatmak zorunda kalındıysa, W-4/W-5 FAIL olarak raporlanmalıdır.**

### Windows sonuç kaydı (KAN.US doldurur)
| ID | Sonuç (PASS/FAIL/NOT TESTED) | Not (pid, çıktı) |
|---|---|---|
| W-0 sha256 | | |
| W-2 env | | |
| W-3 Job üyeliği / breakaway | | |
| W-4 iptal ağacı | | |
| W-5 zaman aşımı ağacı | | |
| W-6 bellek | | |
| W-7 ADS / ayrılmış ad / `\\?\` / UNC | | |
| W-8 Testler env | | |
| W-9 temizlik | | |
