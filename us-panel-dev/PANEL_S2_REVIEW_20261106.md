# PANEL §2 — US yamalarının bağımsız incelemesi (06.11.2026, TSİ)

İnceleyen: PANEL DEV (yalnız okuma + /tmp kopyalarında test). Kaynak ağaçlarına yazılmadı. Ham çıktılar: `/workspace/us-panel-dev/panel-evidence/review/`.
Probe testleri yalnız geçici kopyada (`/tmp/pr/a/b/probe/zz_review_probe_test.go`, kopyası `panel-evidence/review/zz_review_probe_test.go.txt`) çalıştırıldı. Çıktıları `20_probes.txt` dosyasında.

## Özet kararlar

| Madde | Test/derleme | Güvenlik iddiası | Karar | Entegrasyon önerisi |
|---|---|---|---|---|
| 1 Güvenli çalıştırma | PASS | Kısmen yanlış: 2 YÜKSEK kaçış kanıtlandı | **FAIL** (güvenlik) | **Düzeltmeyle kabul**: baseline'dan daha iyi, ama "sandbox/env sızmaz" iddiası düzeltilmeden EXE notlarına yazılmamalı |
| 2 Dosya araçları | PASS | Hard link + TOCTOU açığı (ORTA) | **PASS (bulgularla)** | **Düzeltmeyle kabul** |
| 3 Onarım döngüsü (mock) | PASS (gerçek model NOT TESTED, GUI'ye bağlı değil) | Onay ve bütçe uygulanıyor | **PASS** | **Kabul** (GUI bağlantısı ayrı iş) |
| 4 Görev durumu | PASS | Durum dosyası kurcalanırsa onaysız yazım (ORTA) | **PASS (bulgularla)** | **Düzeltmeyle kabul** |
| Windows Job Object canlı | — | — | **NOT TESTED** | — |
| EXE .8 (yalnız okuma) | sürüm/sha doğru | — | **PASS (statik)**, çalıştırma NOT TESTED | — |

## 1. Yama uygulama ve ağaç eşitliği
- Kök: `/workspace/us-panel/work2/kiyas/term` **zaten .8 birleşimini içeriyor**, temiz taban olamaz. Taban olarak `/workspace/us-panel-dev/baseline_tree` kullanıldı. Yedek `backups/term_pre_round8_20261106_0846.tgz` ile farkı yalnız Round 7 accounts dosyaları: gateway.go, licensesrv/accounts{,_test}.go, licensesrv/model_proxy.go (`00_backup_vs_baseline.txt`).
- item1→item2→item3→item4 sırayla `patch -p1 --dry-run` ve ardından gerçek uygulama: 4/4 temiz (fuzz/offset yok) (`01_patch_apply.txt`).
- Birleşik yama, `-p3` ile temiz uygulanıyor. `-p4` başarısız; bu yamanın yolları `work2/kiyas/term/` ile başlıyor.
- Sıralı (4 madde) ağaç ile birleşik ağaç **birebir aynı** (`diff -r` boş, `02_sequential_vs_combined.txt`). Birleşik yama = 4 maddenin toplamı.
- Birleşik ağaç = `/workspace/us-panel-dev/work2/kiyas/term` **birebir aynı** (`03_…` boş).
- Ana kaynak ile birleşik ağacın farkı (`04_…`):
  - 21 yeni veya değişen §2 dosyası. Bunlar ana kaynakta birleşik ağaçla aynı.
  - Round 7 dosyaları: gateway.go, licensesrv/accounts{,_test}.go, model_proxy.go.
  - release.go (`2026.11.06.8`) ve US_DERLE.bat (`…_8.exe`).

## 2. Testler (GOPROXY=off GOSUMDB=off US_REAL_MODEL_TEST=0 US_REAL_MODEL_E2E=0 US_DOOR_PROBE=0 US_REA_NETWORK_TEST=0)
Her ara adımda `gofmt -l`, `go vet ./...`, `GOOS=windows go vet ./...`, `GOOS=windows go build ./...` ve `go test -race -count=1 -json ./...` çalıştırıldı (`10_*_static.txt`, `11_*_race.json`).

| Ağaç | vet | win vet | win build | race (alt testler dahil) pass/fail/skip | üst düzey |
|---|---|---|---|---|---|
| baseline | 0 | 0 | 0 | 338/0/6 | 301/0/6 |
| +item1 | 0 | 0 | 0 | 347/0/6 | 310/0/6 |
| +item2 | 0 | 0 | 0 | 351/0/6 | 314/0/6 |
| +item3 | 0 | 0 | 0 | 355/0/6 | 318/0/6 |
| +item4 = birleşik | 0 | 0 | 0 | **358/0/7** | 321/0/7 |

- Not: İlk koşuda `licensesrv.TestParsePlansBothExportSchemas` her ağaçta FAIL verdi. Test, `../../../../reports/US_PLANS_JSON_{TEST,CANLI}.json` dosyalarını göreli yolla okuyor ve /tmp düzeninde bu yol yoktu. Bu bir ortam sorunu, yamalarla ilgisi yok. İki JSON `/tmp/pr/reports`'a kopyalandı ve licensesrv yeniden koşturuldu (`12_*_licensesrv_rerun.json`). Hepsi PASS. Tablo bu düzeltilmiş sayılardır (`13_test_counts_summary.txt`).
- **US iddiası 361/0/7:** Birleşik yamalı ağaçta 358/0/7 çıktı. Fark 3 test: ana kaynaktaki Round 7 `licensesrv/accounts_test.go` dosyasında 12 Test/t.Run var, baseline'da 9. US'in `reports/main_merged_checks.txt` sayımı ana kaynak üzerinde yapılmış. Bu açıklamayla iddia **tutarlı**. Ana kaynak ağacında aynı koşu tekrar edilmedi, çünkü o ağaca yazmamak için testi orada koşturmadım.
- gofmt -l: tüm ağaçlarda aynı 30 eski dosya. Yeni §2 dosyalarının hiçbiri listede yok. change_history.go baseline'da da listedeydi.
- Yeni 20 test `-race -count=3`: **60 pass / 0 fail / 3 skip**. Skip'ler, yalnız alt süreç olarak çalışan `TestTaskStateCrashChild` yardımcısı. Testler kararlı (`14_new_tests_count3.json`).
- Skip listesi (birleşik):
  - TestPHPLicenseMatchesThePanel: "php yok"
  - TestLiveDoorAnswers: opt-in canlı kapı (US_DOOR_PROBE)
  - TestOfficialREAInspectsOwnedFile: "opt-in external npm REA integration"
  - TestRealModelPanelMultiFileRepair: "opt-in real model test"
  - TestRealModelPanelBuildRentalProject: "opt-in: US_REAL_MODEL_E2E=1…"
  - TestTaskStateCrashChild: yalnız çökme alt süreci (yeni)
  - licensesrv.TestStoreWriterProcess: alt süreç yardımcısı

## 3. Korunan dosyalar (`05_protected.txt`)
- §2 yamaları korunan dosyaların **hiçbirine dokunmuyor**: US_DERLE.bat, cmd/us-license-hook/**, licensesrv/**, worker_client.go, account_client.go, accounts_mode_round7_test.go, hook_release_test.go, gateway.go, gateway_credit_test.go, embedded_key.go, embedded_license.go, release.go, commercial_policy.go.
- Ana kaynakta baseline'a göre değişenler (US'in kendi Round 7/8 işi):
  - US_DERLE.bat
  - release.go
  - gateway.go
  - licensesrv/accounts.go
  - licensesrv/accounts_test.go
  - licensesrv/model_proxy.go
- go.mod: yalnız `golang.org/x/sys v0.30.0` satırından `// indirect` kaldırılmış.

## 4. Güvenlik bulguları (önem sırasıyla)

### YÜKSEK-1 — Çalıştır ve Testler yolunda env sızıntısı: proje içi `.venv/bin/python` tüm panel ortamıyla çalıştırılıyor (KANITLANDI)
- `run.go:424-426` → `findPython` → `projectPython` (`run.go:578-596`) → `pythonRuns` (`run.go:492-497`): `exec.CommandContext(ctx, path, "--version")`. **cmd.Env ayarlanmadığı için** ebeveyn ortamı devralınıyor. Rlimit ve süreç grubu da yok.
- Bu yolu kullananlar: Çalıştır (`run.go:285`, detectLaunch), Testler (`project_test_runner.go:43`) ve Python sözdizimi denetimi (`model.go:1371-1378`; burada asıl komut da Env'siz çalışıyor).
- Kanıt: probe P4 ve P9. Proje kökündeki `.venv/bin/python` shell betiği `env` dökümü yaptı. `REVIEW_FAKE_SECRET` değeri dökümde göründü (`20_probes.txt`: "secret in child env=true", Tests yolundan sonra sayaç 2).
- Sonuç: `TestRunPathDoesNotLeakSecretsToProgram` yalnız son newSafeCmd çağrısını kapsıyor. "Run yolu sır sızdırmaz" iddiası **eksik**.
- Ayrıca bu sanal ortam yolunda argv[0] proje kontrolündeki bir dosya. Bkz. YÜKSEK-3.
- Düzeltme: `pythonRuns` ve `model.go:1378` için `cmd.Env = safeExecEnv()`, süreç grubu ve limit kullanılmalı. Tercihen proje içi yorumlayıcı da newSafeCmd üzerinden çalıştırılmalı.

### YÜKSEK-2 — Terminal allowlist'i `-toolexec` ile kabuk komutu çalıştırmaya izin veriyor (KANITLANDI)
- `terminal.go:37-75`: Engellenenler yalnız şunlar: `;|&`$<>`, mutlak yol, `..`, `:` ve tam eşleşen `-c/-e/-C/--eval…` argümanları. `-toolexec=sh -c '…'` tek argüman olduğu için `-c` kontrolüne takılmıyor. Değerde `/` olmadığı için yol kontrolü de devreye girmiyor.
- Kanıt: probe P2. `["go","build","-toolexec=sh -c 'touch PWNED_SHELL'"]` **kabul edildi**, newSafeCmd ile çalıştırıldı ve `PWNED_SHELL` dosyası oluştu (state=failed, yan etki gerçekleşti).
- Ayrıca kabul edilenler: `["go","test","-exec","bash","./..."]` ve `["go","run","-exec=sh","main.go"]`.
- Düzeltme: go alt komutları için bayrak allowlist'i kullanılmalı (`-run`, `-v`, `-count`, `-race`, `-o` …). `-exec`, `-toolexec`, `-ldflags`, `-gcflags`, `-modfile`, `-overlay` reddedilmeli. Argümanlarda boşluk da reddedilebilir.

### YÜKSEK-3 — İkili allowlist yalnız taban adına bakıyor (KANITLANDI)
- `safe_exec.go:54-58`: `filepath.Base(bin)`. Herhangi bir mutlak yoldaki `go`, `python` vb. adlı dosya kabul ediliyor.
- Kanıt: probe P1. `/tmp/…/go` adlı shell betiği newSafeCmd ile çalıştı (stdout "PWNED_NOT_GO", state=success).
- Bugünkü çağıranlar `resolveBin` veya `projectPython` sonucunu veriyor. Sanal ortam yolu (proje içi `python`) bu açığı fiilen kullanılabilir hale getiriyor. Docker veya `package.json` kaynaklı argv yolları NOT TESTED.
- Ayrıca `npm` allowlist'te ve `run.go:312-317` `npm start` script'ini çalıştırıyor. Bu, `package.json` içindeki kabuk komutunun çalışması demek. Kod okumasına dayalı.
- Düzeltme: newSafeCmd içinde `exec.LookPath` ile çözülmüş ve beklenen dizinlerdeki ikili zorunlu tutulmalı. Proje içi yorumlayıcı açık bir kullanıcı onayına bağlanmalı.

### ORTA-1 — Zaman aşımı veya iptalden sonra `setsid` yapan torun süreç yetim kalıyor (KANITLANDI)
- `process_tree_unix.go:17`: yalnız `kill(-pgid)`.
- Kanıt: probe P5, timeout ve cancel için. Ana süreç ve aynı pgid'deki `sleep` öldü, `start_new_session=True` ile başlatılan `sleep` **hayatta kaldı** (`setsidSleep_alive=true`). Probe sonunda elle öldürüldü.
- Normal torun süreç (aynı grup) için US'in iddiası doğru.
- Düzeltme (Linux): cgroup v2 (`cgroup.kill`) veya PR_SET_CHILD_SUBREAPER ile alt ağaç toplanmalı.

### ORTA-2 — Linux'ta süreç sayısı limiti yok, rlimit başlatmadan sonra uygulanıyor
- `safe_limits_linux.go:38-41`: RLIMIT_NPROC bilerek konmamış. Fork bombası paneli ve makineyi etkileyebilir. Kod okumasına dayalı; fork bombası denenmedi.
- `safe_exec.go:143-155`: `prlimit` işlem `Start()` sonrası çağrılıyor. İlk anda fork edilen torunlar limitsiz kalabilir. Kod okumasına dayalı; P6'da yarış görülmedi, alt süreç t0'da `(600,600)` okudu.
- **Olumlu:** rlimit yalnız alt sürece uygulanıyor, panel etkilenmiyor. P6: panelin RLIMIT_CPU değeri önce de sonra da sınırsız.
- Windows Job Object da süreç başladıktan sonra atanıyor (CREATE_SUSPENDED yok). Kod okumasına dayalı, NOT TESTED.

### ORTA-3 — Testler yolu ve diğer exec yolları güvenli çalıştırıcıyı kullanmıyor
- `project_test_runner.go:53-59`: env filtreli, ama rlimit ve allowlist yok.
- Ayrıca `model.go:1060,1225,1378`, `rea.go:66`, `studio.go:373`, `pack.go:51` (`reaEnv` yalnız 3 anahtarı siliyor, `rea.go:192-201`) ve `proposal_syntax.go:49`.
- Prompt'taki "kaynak sınırları" (T06) yalnız Terminal, Çalıştır ve onarım döngüsünde karşılanıyor. Kod okumasına dayalı.

### ORTA-4 — Dosya araçları: hard link ile kök dışı okuma ve symlink TOCTOU
- Kanıt: probe P7. Kök içindeki hard link, kök dışındaki dosyayı okuttu (`content="OUTSIDE-SECRET"`).
- `file_tools.go:94` `os.OpenFile`, `O_NOFOLLOW` kullanmıyor. resolve (`:45-87`) ile açma arasında bileşen symlink'e çevrilebilir (TOCTOU). Kod okumasına dayalı, yarış denenmedi.
- Yazma `rename` ile yapıldığı için hard link üzerinden dışarı yazma **yok**: rename dizin girdisini değiştiriyor.
- Düzeltme: okumada `Nlink>1` reddi, `O_NOFOLLOW`, Linux'ta `openat2(RESOLVE_BENEATH)`.

### ORTA-5 — Görev durumu kurcalanırsa onaysız düzeltme uygulanıyor (KANITLANDI)
- `task_state.go:143-162`: Devam sırasında `PendingEdits` yeniden onay sorulmadan yazılıyor. Durum dosyası proje içinde (`.us-tasks/agent/`), yani proje kodu (ör. `go test`) tarafından yazılabilir. Bütünlük kontrolü veya imza yok.
- Kanıt: probe P8. Elle hazırlanan `phase=applying` kaydıyla `Approve` hiç çağrılmadı (`approveCalled=false`) ama `main.go` dosyasına enjekte içerik yazıldı (`injectedOnDisk=true`).
- Yollar kök içinde kalıyor, ft.resolve bunu sağlıyor.
- Ayrıca kilit yok: aynı kimlikle eşzamanlı iki koşuda son yazan kazanır. Kod okumasına dayalı.
- Düzeltme: durum panel veri dizininde tutulmalı veya HMAC ile korunmalı. Devamda bekleyen düzeltmeler yeniden onaya sunulmalı. Dosya kilidi eklenmeli.
- **Olumlu:** id ile yol kaçışı yok. `^[A-Za-z0-9_-]{1,64}$`; `../x` ve `..` reddedildi (P8b).

### DÜŞÜK-1 — Windows yolları (kod okumasına dayalı)
- `file_tools.go:45-62`: `a.txt:stream` (ADS) ve sürücüye göreli `C:evil.txt` resolve'dan geçiyor (P7b, Linux'ta).
- Windows'ta `C:evil.txt` için `filepath.Join(root, …)` kök altında kalır. ADS ise aynı dosyanın gizli akışına yazar.
- Büyük/küçük harf: `insideRoot` `filepath.Rel` kullanıyor; Windows'ta karşılaştırma harf duyarlı olabilir. Bu bir kaçış değil, yanlış ret riski. NOT TESTED.
- Öneri: `:` içeren bileşenler reddedilmeli.

### DÜŞÜK-2 — Diğer
- Env allowlist (`safe_exec.go:25`): Terminal yolunda görülen anahtarlar PATH, HOME, LANG, LC_ALL, GOPROXY (P3). Gizli değer sızmadı (`NOLEAK`).
  - HOME, GOCACHE ve GOPATH go için gerekli. Bunlar sır değil, ama HOME alt süreç tarafından okunabilir; sandbox olmadığı için ~/.ssh vb. erişilebilir. Kod okumasına dayalı.
  - GOFLAGS ve GOPROXY panel ortamından geçiyor. Panel ortamında `GOFLAGS=-toolexec=…` varsa her go çağrısına uygulanır, ama bu ortamı panel sahibi kontrol ediyor. DÜŞÜK.
- Çalıştır zaman aşımı 4 saat (`run.go:115`).
- Onarım döngüsü: `MaxIterations<=0` olunca bütçenin tamamı varsayılana dönüyor (`repair_loop.go:113-115`). MaxDuration her devamda yeniden başlıyor.
- 'Program başlatılamadı' log yolu doğru: `gui.go:817-821`, `run.go:57,82`.
  - Test: `TestRunLogNotWrittenWhenProgramDidNotStart` (GUI üzerinden, PATH boş) ve `TestSafeExecStartFailureIsReported`. İkisi de PASS.
  - GUI testi "program bulunamadı" dalını vuruyor. `sc.Start()` hata dalı yalnız birim testle kapsanıyor.

### Onarım döngüsü (Madde 3), olumlu kontroller (kod okuma + US testleri)
- Model yolları `ft.resolve` ile kök içinde tutuluyor (`repair_loop.go:186-192`). Toplam boyut bütçesi `:193`.
- Onay: `Approve==nil` ise ret (`:197`). Bypass yalnız ORTA-5 senaryosunda var.
- Bütçe: iterasyon, model çağrısı ve süre kontrolleri `:139-165`. `TestRepairLoopStopsAtBudget` PASS.
- Not: Onaylanan düzeltme Go kodu olduğu için `go test` ile çalıştırılıyor. Onay bu yüzden kritik; UI'de diff gösterimi şart (GUI bağlantısı yok).
- `.git/hooks` gibi nokta dizinlerine yazım engellenmiyor. Yalnız `.us-*` özel dizinleri korunuyor (`change_history.go:36`). DÜŞÜK.

## 5. §2 uygunluğu (gap raporu ID'leri)

| ID | Önce | Sonra | Not |
|---|---|---|---|
| T01 Yetkili komut | KISMİ | **KISMİ** | Allowlist ortak, ama YÜKSEK-2 ve YÜKSEK-3 açık |
| T02 cwd koruma | VAR | VAR (güçlendi) | lockCwd symlink ve `..` reddi |
| T03 stdout/stderr | VAR | VAR | Çalıştır yolunda da ayrı |
| T04 İptal | VAR | **KISMİ** | setsid torunu yetim (ORTA-1); Windows NOT TESTED |
| T05 Zaman aşımı | VAR | KISMİ | Aynı nedenle |
| T06 Kaynak sınırı | KISMİ | **KISMİ** | CPU/DATA/FSIZE var; NPROC yok, Testler yolu limitsiz |
| T07 Hata kodu | VAR | VAR | |
| T08 Alan izolasyonu | KISMİ | **KISMİ** | Dosya araçlarında iki alan testi var, ama OS düzeyinde izolasyon yok; hard link |
| U06 Çalıştır env | HATALI | **KISMİ** | Son komut filtreli; sanal ortam yoklaması sızdırıyor (YÜKSEK-1) |
| Bulgu 4 ("Program çalıştırıldı") | HATALI | **KAPANDI** | |
| A06 / M13 Güvenli dosya değişikliği | VAR/KISMİ | KISMİ+ | fileTools yalnız repair_loop ve task_state'te kullanılıyor; editör, öneri ve desk yolları hâlâ eski kodda (appendFile atomik oldu) |
| A08 Onarma | KISMİ | **KISMİ** | Döngü var, ama GUI'ye bağlı değil, gerçek model NOT TESTED |
| A12 Derleme | KISMİ | KISMİ+ | Yalnız Go (`repair_loop.go:80-85`); Python/Node yok |
| A13 Hata yorumlama | KISMİ | KISMİ | Ham çıktı modele gidiyor |
| A14 Düzeltme döngüsü | KISMİ | **VAR (mock, yalnız Go, UI yok)** | |
| A15 / A16 Durum kaydı / devam | VAR / KISMİ | **A16 KISMİ+** | Onarım döngüsü için devam var; genel ajan için yok; ORTA-5 |
| A09, L05-L08, U04, U11 | — | Değişmedi | Kapsam dışı |

## 6. EXE .8 (yalnız okuma, çalıştırılmadı)
- `/workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_8.exe`, 14 759 936 bayt, 08:49 TSİ.
- sha256 `e7e54a42…b26b`, ROUND8 raporuyla **eşleşiyor**.
- `release.go:11` `2026.11.06.8`, `US_DERLE.bat:23` `…_8.exe`.
- İkilide "2026.11.06.8", "Program başlatılamadı", "kaynak sınırı uygulanamadı", "geçersiz görev kimliği" ve `AssignProcessToJobObject` dizgeleri var. §2 kodu derlemeye girmiş görünüyor.
- Windows'ta çalıştırma: **NOT TESTED**.

## 7. Kanıt komutları
```
cp -a /workspace/us-panel-dev/baseline_tree base; patch -p1 --dry-run < s2_itemN_*.patch; patch -p3 < s2_all_items_combined.patch
diff -r s1 comb; diff -r -x kiyas -x '*.exe' comb /workspace/us-panel-dev/work2/kiyas/term
/tmp/panel-review/run_checks.sh <ağaç> <etiket>   # gofmt, vet, win vet/build, go test -race -count=1 -json
go test -race -count=3 -json -run '^(…20 yeni test…)$' .
go test -race -count=1 -run 'TestProbe' -v .   # /tmp/pr/a/b/probe
sha256sum dist/US_Ticari_Aday_2026_11_06_8.exe; grep -a -c "<dizge>" dist/…_8.exe
```
