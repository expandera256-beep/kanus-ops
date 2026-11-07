# PANEL DEV — r13b ve f01 native Windows log incelemesi (2026-11-08 01:50 TSİ)
Kaynak: KAN.US'ın ilettiği loglar (DESKTOP-48D4KOL, demir, admin değil, Win10 19045.6466). PANEL DEV hiçbir binary'yi çalıştırmadı; yalnız log ve sha doğrulaması yaptı.

## r13b native: PASS
- Box üzerinde sha256 doğrulandı: wintest_bundle_v2.zip 4a0a4432…c547, windows_handle_fileops_v2.diff 662fc40a…089c, amd64.test.exe b0816c2f…57da, 386.test.exe df796fa5…e9a9. Bu değerler logdaki certutil ve R13B_META satırlarıyla eşleşiyor.
- amd64 ve 386: her biri PASS=81 FAIL=0 SKIP=15, panic=0 (RUN_METADATA.txt:8).
- "BARRIER: swap performed" logda 18 kez geçiyor; sınama gerçekten koştu. Ancestor rename: OUTCOME=REFUSED errno=5 (:386). Kontrol vakası TestR13bWindowsParentHelperControl PASS. r13'teki geçersiz kanıt sorunu kapandı.
- NOT TESTED (SKIP): symlink (SeCreateSymbolicLinkPrivilege gerekiyor), volume mount point, tilde'siz 8.3 adı, dizin bazında büyük/küçük harf duyarlılığı, ikinci hesapla DACL. Hepsi admin, Developer Mode ya da ek bileşen gerektiriyor.
- BLOCKED: r13b'nin Linux tarafı (kod incelemesi, regresyon, race) ve kapsam dışı dosyalar için .13/.14 görüşü. İnceleme aracı güvenlik filtresine takıldı. US'in bildirdiği 491/0/7 bağımsız doğrulanmadı.

## f01 native: FAIL (amd64 PASS, 386 FAIL)
- Zip sha deafd5e3…95d2; box'ta ve PC'deki Get-FileHash ile eşleşiyor.
- amd64: TestF01WindowsLowIL PASS. Sonuçlar: LOW_IL_OK, ADMIN_DENY_ONLY_OK, PRIVILEGES_DROPPED_OK, INSIDE_WRITE_OK, OUTSIDE_WRITE_DENIED.
- 386: sandbox_windows_test.go:145 hata veriyor: state=start_failed err="kaynak sınırı uygulanamadı: Parametre hatalı" (ERROR_INVALID_PARAMETER).
- **Muhtemel kök neden (kodda doğrulandı, Windows 386'da ayrıca test edilmedi):**
  - safe_exec.go:47 `MemoryBytes: 4 << 30` tanımlıyor; safe_limits_windows.go:27 bunu `info.ProcessMemoryLimit = uintptr(l.MemoryBytes)` olarak atıyor.
  - 386'da uintptr 32 bit olduğu için 4 GiB 0'a kesiliyor. JOB_OBJECT_LIMIT_PROCESS_MEMORY bayrağı açık ve sınır 0 olunca SetInformationJobObject (:37) reddediyor.
  - Fail-closed olduğu için güvenlik açığı değil. Ama 386'da safe exec ile hiçbir program başlatılamaz. **MEDIUM (kullanılabilirlik), yalnız 386.** Aynı kod r13b/term/safe_exec.go:274'te ve .12 tabanında da var.
  - Öneri: GOARCH=386 için bellek sınırını taşmayacak bir değere kırp (ör. min(MemoryBytes, ^uintptr(0)) ya da 386'da 2–3 GiB). Ayrıca 32 bit taşmasını yakalayan bir test ekle. Ya da 386 EXE yayımlanmayacaksa bunu açıkça belgeleyip 386'yı derlemeden çıkar.
- Not: OUTSIDE_READ_STILL_ALLOWED. Low IL kök dışındaki okumayı engellemiyor. Bu, FINDING-01'deki kısmi önlemin beklenen sınırı ve Windows riski Demir'ce kabul edildi.
- LOW (kanıt hijyeni): run.cmd'deki `echo exit_code=%RC%>>file` satırı "0>>"/"1>>" ifadesini handle yönlendirmesi olarak yorumluyor. Bu yüzden sonuc_*.log'da exit_code boş; RC değeri ERRORLEVEL'den alındı (RUN_METADATA.txt). Düzeltme: `>>file echo exit_code=%RC%`.
