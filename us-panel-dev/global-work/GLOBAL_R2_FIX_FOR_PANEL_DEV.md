# GLOBAL → PANEL DEV: R2 düzeltmesi (06.11.2026 ~10:30 TSİ)

Yanıt: PANEL_GLOBAL_REVIEW_R2_20261106.md (R2-01, R2-02). Taban: pinned/v10 (yazılmadı). Çalışma: global-work/v10dev. Yayın/deploy/mesaj yok.

## Yamalar
| Yama | Durum | SHA-256 |
|---|---|---|
| global_i18n_panel_core.patch | değişmedi | e4f10a339c78818ceddbc49e2f687b4ee80afa86a7fce9d407c22bd8536892c7 |
| global_i18n_02_gui_strings.patch | değişmedi | 314fab48b009f06e221ce140758724105b9f5b972c69d0c15c7ad2c9db8d3669 |
| global_i18n_03_pref.patch | **YENİ** | 425a76afd9e07b9bdc711b5496a1539046d975e63b54380d6e4f503f2c2c40a0 |
| global_i18n_04_errors.patch | **yeniden üretildi (yalnız bağlam)** | 1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991 |
| global_i18n_05_email_loader.patch | değişmedi | 7c917564df055f151f8bd0a0b334a09e248db1481723123196fbfb0377093f3d |

04 neden değişti: eski 04'ün gui_lang.go hunk'ı bağlam olarak `return filepath.Join(filepath.Dir(desk), "lang.json")` satırını kullanıyordu; yeni 03'te langPref `Root`+`Name` biçimine geçtiği için o satır yok. `git apply --check` FAIL → 3-way ile yeniden üretildi. +/− satırları eski 04 ile birebir aynı (doğrulandı); yalnız index/hunk başlığı/bağlam farklı. Ağaç sonucu, 04'ü eski yığının üstüne uygulayıp R2 farkını eklemekle byte-eşit (git diff boş). GUI geri alma değişikliği tamamen 03'te (gui.html change handler + #langstatus, yeni anahtar `gui.lang.save_failed` 03'ün tr/en.json hunk'ında); 02 değişmedi.
Eski 03/04 arşivi: global-work/patch_archive_r4/.

## R2-01 tasarımı
- `PrefStore{Root func() string; Name string}`. Kök (desk.json dizini, US_DESK'e uyar) her farklı değer için **bir kez** `MkdirAll + filepath.EvalSymlinks` ile çözülür ve dizin fd'si/handle'ı olarak sabitlenir (önbellek; istek başına EvalSymlinks yok). Kök ve ÜSTÜ symlink olabilir (/home→/var/home, macOS /var→/private/var).
- Kökün ALTINDAKİ her bileşen (Name'in ara dizinleri ve dosya) fd'ye göreli `openat(O_NOFOLLOW|O_DIRECTORY)`; dosya `O_NOFOLLOW|O_NONBLOCK` + fstat regular + ≤4096 B; yazma `O_EXCL|O_NOFOLLOW` temp + fsync + `renameat` (aynı dirfd) + dir fsync. G-01/G-02 korumaları aynen.
- Kök açılamazsa çökme yok: tercih bellekte tutulur, Set `ErrNotPersisted`, `/api/lang` POST **503** "language preference could not be saved (kept for this session only)".
- GUI: önce POST; yalnız 2xx'te `loadLocale(lang)`. 2xx değil/ağ hatası → seçici önceki dile döner, `#langstatus` (role=alert) önceki dilde `gui.lang.save_failed` gösterir (TR "Dil tercihi kaydedilemedi; önceki dil korunuyor." / EN "Could not save the language preference; keeping the previous language."). Koşulsuz `loadLocale(...)` catch/finally zinciri kaldırıldı.
- Test semantiği değişikliği (açıkça): TestPrefSetRejectsSymlinkParent artık güvenilir kök + `Name: "parent/lang.json"` ile "kökün altındaki symlink" olarak sınanıyor (eskiden dosyanın kendi dizini kök sayılmıyordu).

## R2-02 (Windows) tasarımı
- `NtCreateFile` + `OBJECT_ATTRIBUTES.RootDirectory` (handle'a göreli açılış) + `FILE_OPEN_REPARSE_POINT`: her ara dizin ve dosya, sabitlenmiş kök handle'ından göreli açılır; açılan **handle** `GetFileInformationByHandle` ile doğrulanır (REPARSE_POINT yok, beklenen dizin/dosya türü, boyut ≤4096).
- Yazma handle üzerinden: temp `NtCreateFile(FILE_CREATE)` dizin handle'ına göreli, WriteFile + FlushFileBuffers, `NtSetInformationFile(FileRenameInformation, RootDirectory=dirHandle, ReplaceIfExists)`; hata halinde `SetFileInformationByHandle(FileDispositionInfo)` ile temp silinir. Mevcut hedef önce reparse-nofollow açılıp handle'dan regular doğrulanır. Yeni bağımlılık yok (x/sys/windows v0.30.0 modülde).
- Kalan risk: hedef-varlık kontrolü ile rename arasında hedef adın junction/symlink ile değiştirilmesi — NTFS rename hedef reparse noktasını takip etmeyip onu değiştirir (hedef dosyaya yazılmaz); yine de **gerçek Windows'ta NOT TESTED**.

## Önce FAIL → sonra PASS
| Test | Eski 03 (171bf6fc…) | Yeni 03 |
|---|---|---|
| Symlink'li kök kaydı 200 + kalıcı (i18n, eski API eşdeğeri) | FAIL "unsafe preference path" | PASS (TestPrefSymlinkedRootPersists, TestPrefHandlerSymlinkedRootAndUnavailable 200) |
| Kök yeniden bağlama → yazım taşınmaz (fd sabitleme) | FAIL (ilk kayıt bile reddedildi) | PASS (TestPrefRootRebindPinned: b/lang.json yok, a/lang.json güncellendi) |
| UI 500 → geri al + mesaj | FAIL (value=en, /api/i18n?lang=en yüklendi, mesaj yok) | PASS |
| UI 503 / ağ hatası → geri al | FAIL / FAIL | PASS / PASS |
| UI 200 → en | PASS | PASS |
| Symlink kökün altında (ara dizin + dosya) | — (yeni) | PASS (TestPrefSymlinkBelowSymlinkedRootRejected) |
| Kök açılamaz → bellek + 503 | — (yeni) | PASS (TestPrefUnavailableRootMemoryOnly, handler 503) |
| Önceki 8 güvenlik testi (G-01…G-04) | — | 8/8 PASS |
Loglar: r5_before_i18n.log, r5_before_ui.log, r5_after_ui.log.

## Komutlar (pinned/v10 + 01→04, fixture'lı /tmp/fx3)
| Komut | Sonuç |
|---|---|
| git apply --check && apply 01→04 | PASS 4/4; ağaç v10dev ile birebir |
| go build ./... ; go vet ./... | PASS ; PASS |
| GOOS=windows go vet ./... ; GOOS=windows go build ./... | PASS ; PASS |
| GOOS=windows go test -c ./i18n (pref_windows_test.go dahil, 4 test) | PASS (derlendi, 9,1 MB exe) |
| Windows testlerinin koşulması | **NOT TESTED** — wine yok; apt kurulumu BLOCKED (deb.debian.org 500 / imzasız depo). Gerçek Windows NOT TESTED. |
| gofmt -l (yama dosyaları) | PASS (boş) |
| go test -race -count=1 -p=1 -json ./... | PASS 415 pass / 0 fail / 7 skip, exit 0, DATA RACE 0 (önce 410; +5 yeni Go testi) |
| node check_ui.mjs ; node gui_lang_ui_test.mjs | PASS ; PASS 4/4 |
| Python 05 kısa kontrol test_i18n_mail (US_GO_LOCALES=yeni locales) | PASS 10/10 (05 değişmedi) |

## Yığın uyumu (.11 → UYUM → API → GLOBAL), /tmp kopyalarında
| Taban | UYUM | API | GLOBAL 01→04 |
|---|---|---|---|
| /workspace/us-panel/work2/kiyas/term (buildVersion 2026.11.06.11, anlık kopya) | uyum-v3 01→04 patch --fuzz=0 PASS | api-r11/0001 PASS | check+apply 4/4 PASS; go build + go vet (. ./i18n) PASS |
| aynı | uyum-v2 PASS | api-r11 PASS | 4/4 PASS |
| /workspace/us-panel-dev/r11_base (buildVersion .10 — adı r11 ama sürüm .10) | v3 / v2 PASS | PASS | 4/4 PASS |
Not: .11 için sabit (pinned) manifest yok; canlı US ağacının o anki kopyası kullanıldı. Yığında tam test suite koşulmadı → **birleşik suite NOT TESTED**; yalnız uygulanabilirlik + build/vet.

## Kalan riskler
- Gerçek Windows (NtCreateFile/rename semantiği, ACL, junction) NOT TESTED; macOS NOT TESTED.
- Kök, süreç ömrü boyunca ilk çözülen dizine sabit: kullanıcı config'i çalışırken taşırsa yeniden başlatma gerekir (bilinçli).
- Kök çözümü yalnız MkdirAll+EvalSymlinks ile ilk kullanımda; kökün kendisini ilk kullanım öncesi yönlendirebilen yerel saldırgan zaten kullanıcının config'ine yazabilir (tehdit modeli dışı).
