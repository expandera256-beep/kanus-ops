# .13 adayı — Windows handle-tabanlı dosya işlemleri

Tarih: 2026-11-08 01:09 (Europe/Istanbul, UTC+3)
Durum: **Kod ve çapraz derleme DONE; yerel Windows güvenlik kabulü NOT TESTED.**

## Taban ve kapsam
- Çalışma kopyası: `/workspace/us-panel-dev/r13/term`.
- `.12/TREE_MANIFEST.sha256` ile başlangıçta **209/209** eşleşme doğrulandı;
  yama üretilirken asıl work2 Go kaynakları yine **209/209** eşleşti.
- 5 kaynak değişti, 3 test dosyası eklendi. `go.mod/go.sum` değişmedi;
  mevcut `golang.org/x/sys v0.30.0/windows` kullanıldı.
- `post12/term`, licensesrv protokolleri, gömülü anahtarlar ve dist dosyaları değiştirilmedi.
  Deploy, canlı GPU/VPS/Shopify işlemi yapılmadı; sürüm paketi üretilmedi.
  Derlemeler yalnız doğrulama amaçlı scratch çıktıları ve test EXE'leridir.

## Kök neden
`.12` Windows dosya katmanı, yol bileşenlerini denetledikten sonra aynı mutlak
adı yeniden açıyordu. Denetim ile kullanım arasında workspace veya ebeveyn dizin
junction ile değiştirilirse sonraki open/rename/delete dış dizine yönlenebilirdi.
Unix araçları zaten kalıcı root fd kullanıyordu; eski desk/history yardımcılarında
ise yol bazlı ek okuma/yazma işlemleri kalmıştı.

## Tasarım ve dosya:satır
- `file_open_windows.go:22`: kök bir kez `CreateFile` ile açılır;
  BACKUP_SEMANTICS + OPEN_REPARSE_POINT, reparse/dizin doğrulaması yapılır.
  HANDLE araç ömrü boyunca saklanır; `:43` her işlem için private duplicate üretir.
- `file_open_windows.go:81`: her bileşen, `NtCreateFile` ve
  `OBJECT_ATTRIBUTES.RootDirectory` ile önceki HANDLE'a göre açılır.
  `FILE_OPEN_REPARSE_POINT`, dizin/normal-dosya türü ve `:62` handle metadata
  kontrolü ile symlink, junction ve mount/reparse bileşenleri reddedilir.
  Dosyalarda `NumberOfLinks != 1` reddedilir (okuma/yazma/silme).
- `file_open_windows.go:114`: açılmış nesnenin normalize adı HANDLE üzerinden
  sorgulanır; dosya adı farklıysa DOS alias reddedilir. Dönen yol IO için kullanılmaz.
- `file_open_windows.go:227`: temp dosya ebeveyn HANDLE'a göre FILE_CREATE ile
  açılır; yazılır, flush edilir, aynı ebeveyn altında atomik rename yapılır.
  Hata halinde temp HANDLE üzerinden disposition ile temizlenir.
- `file_open_windows.go:201`: **tasarımda bilinçli API farkı:** istenen
  `SetFileInformationByHandle(FileRenameInfo)` yerine mevcut x/sys içindeki
  `NtSetInformationFile(FileRenameInformation)` kullanıldı. RootDirectory hedef
  ebeveyn HANDLE'ıdır, ad tek göreli bileşendir. Win32 wrapper'ın bazı sürümlerde
  göreli adı cwd'ye göre mutlaklaştırabildiğine dair uyumluluk bulgusu vardır;
  güvenlik sınırını korumak için path tabanlı fallback YOKTUR.
- `file_open_windows.go:287`: Delete, ebeveyne göre açılan tek-link normal dosya
  HANDLE'ında `SetFileInformationByHandle(FileDispositionInfo)` ile yapılır.
- `file_tools.go:61`: resolve yalnız sözcüksel denetimdir; kök yeniden
  Lstat/EvalSymlinks ile çözülmez. Özel uygulama klasörleri korunur.
  `:194` ADS, tilde/8.3 biçimleri, ayrılmış aygıtlar (üstsimge rakamlar dahil),
  nokta/boşlukla biten adlar reddedilir. Baş/son boşluk sessizce kırpılmaz.
  `Close` ile araç kök HANDLE/fd ömrü açıkça sonlandırılabilir.
- `desk.go:263`: içerik okumadan ve eski yol denetimlerinden önce kök pinlenir;
  gerçek okuma `openBeneathAt` kullanır. Metadata denetimi erişim yetkisi değildir.
- `change_history.go:71,121`: snapshot, geçmiş kaydı yazma/okuma, dizin listeleme,
  geri yükleme ve kayıt silme aynı işlem başında pinlenen kökten yürür.
  `file_tools.go:163` eski writeModelFile/atomicWriteFile yolu da önce pinler.
- `file_open_unix.go:116,179,277`: güvenli göreli dizin listeleme ve geçmiş
  işlemlerine destek eklendi; yeni history dizini 0700, kaydı 0600 kalır.
  Yeni nested dosya snapshot'ı için ENOENT güvenli biçimde ayrıştırılır.

## Testler ve kanıt
Sayaçlar aksi söylenmedikçe Go test/subtest terminal olaylarıdır, benzersiz fonksiyon sayısı değildir.
- Linux `go vet ./...`: PASS (`r13/linux-vet-final.log`).
- Linux masaüstü + hook `go build`: PASS (yalnız scratch doğrulama çıktıları).
- Windows amd64 + 386 `go vet ./...`: PASS.
- Windows amd64 masaüstü `go build`: PASS.
- Windows amd64 + 386 `go test -c`: PASS; **çalıştırılmış Windows testi değildir**.
- Son ağaç `go test -json ./...`: **485 PASS, 0 FAIL, 7 SKIP**, 3 paket PASS;
  üst düzey testler: 409 PASS, 7 SKIP (`linux-test-complete.json`).
- İlk tam `go test -race -json ./...`: **484 PASS, 0 FAIL, 7 SKIP**;
  son eklenen desk/history parent-swap testini de içeren 4 .13 testi ayrıca
  `-race -count=3` ile **12 PASS, 0 FAIL** (`r13-extra-race.json`).
- Son donmuş ağacın tam race tekrar sonucu aşağıdaki son doğrulama bölümündedir.
- `TestPanelRootSwapWriteDelete -count=3`: **3/3**; write 3/3, delete 3/3
  (9 PASS olayı, `root-swap-3.json`).
- Yeni Linux testleri: path-taking writer parent-swap, desk/history parent-swap,
  geçmiş izinleri/nested yeni dosya, taşınabilir alias/device politikası.
- S2 final regresyon: **36 PASS, 0 FAIL, 0 BLOCKED**;
  kaynak öncesi/sonrası manifest aynı (`r13/s2-final.log`, `r13/s2-final/`).
- Python lifecycle, ayrı kaynak kopyasında: **24/24 PASS**
  (`r13/python-lifecycle-final.log`). İlk kopyalama eksik `storage.py` nedeniyle
  import hatası verdi; Python kaynak bağımlılıkları tamamlanınca tekrar PASS oldu.
- Yama `.12` temiz kopyasına uygulandı; adayla `diff -qr` birebir eşleşti
  (`r13/patch-apply.log`).
- `file_tools_windows_test.go`: 9 Windows üst düzey testi + 1 ortak politika
  testi, native koşucuda üçer tekrar; deterministik ve yarışmalı junction-swap,
  parent-swap, lifecycle, hard link, ADS/device/8.3 ve desk/history kapsanır.

## Teslim
- Yama: `/workspace/us-panel-dev/patches/r13/windows_handle_fileops.diff`
  SHA256: `480722b79ff15d10baf9c4b319b51eb3daf8badd3fea35c04520add048961e8a`
- Windows test klasörü: `/workspace/us-panel-dev/r13/wintest/`
  İki test EXE'si, `run.cmd`, `README_TR.md`, `SHA256SUMS.txt` içerir.
- Taşınabilir paket: `/workspace/us-panel-dev/r13/wintest_bundle.zip`
  SHA256: `635c40e5545e7dcf41b1abb53fc6f4bdcda1b9469258d5385e5e98c96ff2c77f`
- Değişen kaynak özetleri: `r13/CANDIDATE_FILES.sha256`.

## DONE / NOT DONE / NOT TESTED
**DONE:** handle-tabanlı Windows uygulaması, Unix/Windows eski yardımcı yolların
pinlenmesi, otomatik Linux testleri, Windows çapraz vet/build/test derlemesi,
yama ve yerel kullanıcı test paketi.

**NOT TESTED:** gerçek Windows üzerinde API davranışı, junction stres testi,
NTFS hard link/8.3/ACL uyumluluğu ve run.cmd'nin çalıştırılması.
Yerel NTFS'te normal kullanıcıyla junction testleri Developer Mode/yönetici istemez;
kurum politikası engellerse testi SKIP saymak yerine hata kaydı alınmalı.
8.3 üretimi kapalı diskte gerçek alias testi SKIP olabilir ve ayrıca belirtilmelidir.
Windows'a özgü kabul ancak bu paket PC'de çalıştırılıp PANEL DEV inceleyince verilebilir.

**NOT DONE (bu kapsam dışında):** dağıtım/ana ağaca birleştirme/sürüm EXE'si,
aynı UID altında çalışan keyfi proje koduna karşı OS kullanıcı izolasyonu.
Pinleme anındaki seçilen workspace güvenilen başlangıç noktasıdır; bu çalışma bir
kullanıcı/izin sınırı veya aynı kullanıcının keyfi süreçlerine karşı izolasyon iddiası değildir.
Yeni güvenli katman desteklenmeyen dosya sistemi/API hatalarında fail-closed davranır;
FAT, ağ paylaşımı, farklı Windows sürümleri ayrıca kabul testi ister.

## Teknik referanslar
- https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information
- https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info
- Win32 wrapper uyumluluk gözlemi: https://learn.microsoft.com/ja-jp/answers/questions/6015074/setfileinformationbyhandle-filerenameinfo-rootdire

## Son doğrulama

Donmuş son ağacın `go test -race -json ./...` tekrarı: **485 PASS, 0 FAIL, 7 SKIP**, 3 paket PASS (`r13/linux-race-final.json`).

Sürüm/artifact hazırlığı istenmediğinden `release.go` değiştirilmedi; test koşucularında görünen buildVersion `.12` tabanını belirtir, kaynak yamaları .13 adayıdır.

## Native Windows sonucu (2026-11-08, DESKTOP-48D4KOL, KAN.US koşumu)
- amd64: run.cmd, admin yok: 27/27 PASS, 0 FAIL, 0 SKIP, ERRORLEVEL 0. UnsafeNames 13/13 her turda.
- 386: 27/27 PASS, ERRORLEVEL 0.
- RootSwapStress: swap ve ops sayıları her turda >0. ShortAlias gerçek 8.3 kontrolüyle koştu (SKIP değil).
- ParentSwapAtomicHelper: Windows "Erişim engellendi" döndürdü; bu güvenli hata (fail-closed) olarak kabul edildi.
- Loglar: /workspace/us-panel-dev/r13/windows-results-DESKTOP-48D4KOL-20261108/
- Durum: native Windows kabulü TESTED / PASS (önceki "NOT TESTED" ibaresi bu bölümle güncellendi).
