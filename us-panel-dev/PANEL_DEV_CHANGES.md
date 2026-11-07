# PANEL DEV — §2 değişiklikleri (Round 8), 06.11.2026 TSİ

Çalışma kopyası: `/workspace/us-panel-dev/work2/kiyas/term/` · Baseline: `reports/baseline_s2.txt` (vet 0, race 0, windows build 0; gofmt -l 30 eski dosya listeledi — önceden var).
Kontrol çıktıları: `reports/item{1..4}_checks.txt`, birleşik ana kaynak: `reports/main_merged_checks.txt`.
licensesrv, gateway/account protokolleri, gömülü anahtarlar: **dokunulmadı**. Canlı/deploy/GPU: **yok**.
Not: x/sys doğrudan kullanıldığı için go tool `go.mod`'da `golang.org/x/sys v0.30.0` satırındaki `// indirect` yorumunu kaldırdı (sürüm/bağımlılık değişmedi).
Not: `change_history.go` baseline'da da gofmt dışıydı; yalnız `writeModelFile` gövdesi değişti, dosya bütünüyle yeniden biçimlendirilmedi.

## Madde 1 — Ortak güvenli çalıştırma — `patches/s2_item1_safe_exec.patch`
Dosyalar: safe_exec.go, safe_limits_{linux,windows,other}.go, terminal.go, run.go, project_test_runner.go, gui.go, main.go (+testler)
| Kontrol | Sonuç |
|---|---|
| Komut allowlist (ikili + Terminal alt komut politikası) | PASS (TestSafeExecRejectsNonAllowlistedBinary) |
| cwd kilidi: `..` ve symlink kaçışı | PASS (TestSafeExecCwdLockBlocksDotDotAndSymlink) |
| Env allowlist + gizli değer sızıntı testi (Terminal ve Run yolu) | PASS (TestSafeExecEnvDoesNotLeakSecrets, TestRunPathDoesNotLeakSecretsToProgram) |
| Ayrı stdout/stderr + 16 KiB çıktı sınırı + çıkış kodu 7 | PASS (TestSafeExecSeparateStreamsAndCap) |
| Zaman aşımı ve iptalde süreç grubu (torun süreç) ölümü — Linux | PASS (TestSafeExecTimeoutKillsProcessGroup) |
| Linux rlimit (CPU 600s, DATA 4GiB, FSIZE 512MiB) | PASS (TestSafeExecAppliesLinuxRlimits) |
| Windows Job Object (bellek/süreç sayısı/CPU, KILL_ON_JOB_CLOSE) — derleme | PASS (GOOS=windows vet+build) |
| Windows Job Object — canlı davranış | NOT TESTED (Linux box) |
| Başlatma başarısızsa 'Program çalıştırıldı' yazılmaz | PASS (TestRunLogNotWrittenWhenProgramDidNotStart, TestSafeExecStartFailureIsReported) |
| Linux RLIMIT_NPROC | Bilerek uygulanmadı: UID bazlı sayaç thread'li sunucuları bozdu (TestPlanLeavesTheServerRunning). Çok kiracılı izolasyon için cgroup/container gerekir — NOT TESTED |
| gofmt / vet / race / windows build | PASS / PASS / PASS / PASS |

## Madde 2 — Dosya araçları — `patches/s2_item2_file_tools.patch`
Dosyalar: file_tools.go, change_history.go (writeModelFile → atomik), desk.go (appendFile atomik + yedekli)
| Kontrol | Sonuç |
|---|---|
| Kök dışı `..`, mutlak, `\..`, symlink dizin/dosya, symlink kök reddi (okuma/yazma/silme) | PASS (TestFileToolsRefuseEscapes) |
| Atomik yazım (temp+fsync+rename+dir fsync), çökme simülasyonunda eski içerik korunur, temp kalmaz | PASS (TestFileToolsAtomicWriteAndLimits) |
| Okuma/yazma boyut sınırı | PASS |
| İki çalışma alanı ayrımı | PASS (TestFileToolsWorkspacesAreSeparate) |
| appendFile atomik + yedek + symlink reddi | PASS (TestAppendFileIsAtomicAndBackedUp) |
| gofmt / vet / race / windows build | PASS / PASS / PASS / PASS |

## Madde 3 — Bütçeli build→hata→düzeltme→retest — `patches/s2_item3_repair_loop.patch`
Dosyalar: repair_loop.go, repair_loop_test.go (httptest OpenAI-uyumlu mock; gerçek model yok)
| Kontrol | Sonuç |
|---|---|
| Uçtan uca: bozuk Go projesi (undefined: c) → mock model düzeltir → go build/vet/test PASS + bağımsız go test | PASS (TestRepairLoopFixesBrokenGoProjectE2E) |
| İterasyon/model çağrısı bütçesi | PASS (TestRepairLoopStopsAtBudget) |
| Kök dışı yol önerisi, onaysız düzeltme, bozuk JSON → dosya değişmez | PASS (TestRepairLoopRefusesEscapeUnapprovedAndBadJSON) |
| İptal | PASS (TestRepairLoopCancellation) |
| GUI'ye bağlanma, gerçek DeepSeek ile döngü | NOT TESTED (bu turda kapsam dışı) |
| gofmt / vet / race / windows build | PASS / PASS / PASS / PASS |

## Madde 4 — Kalıcı atomik JSON görev durumu — `patches/s2_item4_task_state.patch`
Dosyalar: task_state.go (`.us-tasks/agent/<id>.json`, sürüm 1), repair_loop.go (kanca noktaları)
| Kontrol | Sonuç |
|---|---|
| Gerçek alt süreç çökmesi (os.Exit) dosya uygulamanın ortasında → devam: düzeltme çift uygulanmaz, model tekrar çağrılmaz, sonuç fixed | PASS (TestTaskStateResumeAfterCrashMidApply) |
| Tamamlanan görev yeniden çalıştırılmaz | PASS |
| Model isteği sırasında çökme → otomatik ücretli tekrar YOK; yalnız açık onayla tekrar | PASS (TestTaskStateUncertainModelCallNotReplayed) |
| Bozuk kayıt korunur (.corrupt-*), kimlikle yol kaçışı reddi | PASS (TestTaskStateCorruptFileIsPreserved) |
| gofmt / vet / race / windows build | PASS / PASS / PASS / PASS |

## Birleştirme
Ana kaynak `/workspace/us-panel/work2/kiyas/term`: yalnız bu turda değişen 21 dosya kopyalandı; bu dosyalar ana kaynakta baseline ile birebirdi. Round 7 accounts farkları (gateway.go, licensesrv/accounts*.go, model_proxy.go) korunarak dokunulmadı. Yedek: `/workspace/us-panel/backups/term_pre_round8_*.tgz`.
Birleşik: vet PASS, race PASS (361 PASS / 0 FAIL / 7 SKIP, alt testler dahil), windows build PASS. release.go → 2026.11.06.8, US_DERLE.bat çıktı adı .8.

## Round 9 — login/register zaman aşımı — `patches/r9_login_timeout.patch`
| Kontrol | Sonuç |
|---|---|
| login/register 90 sn, diğer çağrılar 20 sn | PASS |
| Zaman aşımı veya ağ hatasında tam 1 tekrar | PASS |
| 401/429/500/503'te tekrar yok | PASS |
| --check-account ek süre sınırı yok, yeni bütçeye uyuyor | PASS (statik + test) |
| vet/winvet/race (365/0/7)/windows build | PASS |
| .9 Windows çalıştırma | NOT TESTED |

## Round 9b — PANEL DEV inceleme düzeltmeleri — `patches/s2_review_fixes.patch` (nihai .9 içinde)
Kırmızı→yeşil kanıtı: `reports/review_red.txt`. Ayrıntı: `/workspace/us-panel/reports/ROUND9_RAPOR_TR.md`.
| Bulgu | Durum | Test |
|---|---|---|
| YÜKSEK-1 venv env sızıntısı | FIXED, PASS | TestReviewH1VenvProbeDoesNotLeak |
| YÜKSEK-2 Terminal exec bayrakları + GOFLAGS/CC | FIXED, PASS | TestReviewH2TerminalRejectsExecFlags |
| YÜKSEK-3 taban adı / tam yol + realpath | FIXED, PASS | TestReviewH3AbsoluteFakeBinaryRejected |
| ORTA-1 setsid torunu | FIXED (kısmi; yetim çift-fork NOT FIXED) | TestReviewM1SetsidGrandchildKilled |
| ORTA-2 NPROC / başlatma sonrası rlimit | NOT FIXED (belgelendi) | — |
| ORTA-3 Testler yolu limitleri + exec göçü | FIXED (gh/PyInstaller rlimit NOT FIXED) | TestReviewTestsPathHasLimits |
| ORTA-4 hard link | FIXED Unix; Windows NOT FIXED | TestReviewM4HardlinkReadRejected, TestReviewEditorRejectsHardlink |
| ORTA-5 görev durumu kurcalama | FIXED; eşzamanlılık kilidi NOT FIXED | TestReviewM5TamperedTaskStateNeedsApproval |
| DÜŞÜK-1 Windows yol biçimleri | FIXED, PASS | TestReviewWindowsPathFormsRejected |
| Windows Job Object canlı | NOT TESTED | — |
Tam takım race: 374/0/7, win vet/build PASS. .9 sha256 6e1f6fad7d906cbf39e4dd2aee4969bcee4b57a2b00b59235d86faffe96e85a9

## Round 9c — ek bulgular
- file_open_unix.go / file_open_windows.go: openBeneath (fd tabanlı, fstat dev/ino/nlink yerine tip+nlink fd üzerinde). file_tools.go, desk.go bunu kullanır.
- terminal.go: nodeArgsAllowed.
- Testler: TestReviewSymlinkSwapRaceNoEscape (linux), TestReviewNodeArgAllowlist. Kırmızı kanıt: reports/review_red_round2.txt.
- Regresyon 36/36 PASS. .9 sha 816696e5…7fbf.
