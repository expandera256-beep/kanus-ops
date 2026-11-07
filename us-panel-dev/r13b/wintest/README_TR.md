# r13b — Windows dosya katmanı yerel test paketi (v2)

Bu paket **sürüm EXE'si değildir**. Linux'ta çapraz derlenmiş Go test ikilileridir.
İnternet, Go kurulumu, GPU, lisans veya canlı servis gerekmez. Testler yalnız
Windows geçici klasöründe kendi oluşturdukları dosyalara dokunur.

## Adımlar
1. ZIP'i **yerel NTFS** bir klasöre açın (ör. Masaüstü\r13b). Ağ paylaşımı, FAT/exFAT olmaz.
2. **Normal (yönetici olmayan)** Komut İstemi açın, `cd /d` ile klasöre geçin.
3. `run.cmd` çalıştırın. Betik amd64 (64-bit Windows'ta) ve 386 ikilisini sırayla, her testi 3 tur koşar.
4. Ortaya çıkan `r13b-results.log` dosyasını ve ekranda yazan `LOG_SHA256` satırını PANEL DEV'e iletin.

Yönetici olarak çalıştırmak **gerekmez**. Yönetici gerektiren vakalar normal koşuda
`SKIP-ADMIN:` satırıyla atlanır (aşağıya bakın); bunlar PASS sayılmaz.

## Log metadata (PANEL DEV R13 iletim zinciri bulgusu)
`r13b-results.log` başında: `HOSTNAME`, `OS_VERSION`, `ADMIN=0/1`, bütünlük seviyesi
(`Mandatory Label`), her ikilinin, yamanın ve BUILDINFO'nun `certutil` SHA-256'sı,
beklenen `SHA256SUMS.txt` ve `BUILDINFO.txt` (Go sürümü, `go version -m` çıktısı).
Her ikili ilk testte (`TestR13bAAMetadata`) kendi SHA-256'sını, gömülü yama
SHA-256'sını, hostname, Go toolchain, GOOS/GOARCH, elevated bayrağı, kullanıcı SID ve
OS build'ini `R13B_META` satırlarıyla yazar. Sonda `AMD64_EXITCODE`, `I386_EXITCODE`,
`FINAL_EXITCODE` yazılır.

## PASS ölçütleri
- `FINAL_EXITCODE=0`, `AMD64_EXITCODE=0` (64-bit PC'de) ve `I386_EXITCODE=0`.
- Hiçbir yerde `--- FAIL` veya `panic` yok.
- `R13B_META binary_sha256` değeri `SHA256SUMS.txt` içindeki ilgili ikiliyle,
  `R13B_META patch_sha256` değeri `windows_handle_fileops_v2.diff` satırıyla aynı.
- `TestR13WindowsRootSwapStress/*` her turda `swaps>0` ve `successful_operations>0`.
- `TestR13WindowsParentSwapAtomicHelper` her turda `OUTCOME=SWAPPED` ya da
  `OUTCOME=REFUSED errno=5|32` yazar. REFUSED, kontrol rename'i başarılı olduğu için
  geçerlidir ama **saldırı penceresinde swap testi değildir**; gerçek swap kanıtı
  `TestR13bWindowsParentSwapPinnedRoot` (BARRIER satırı zorunlu).
- `TestR13bWindowsDeskAncestorSwap`: `OUTCOME=SWAPPED` veya `OUTCOME=REFUSED` (kontrollü).
- `BARRIER:` ile başlayan satır `swap performed` dışında bir hata içeriyorsa test FAIL'dir.
- Beklenen SKIP'ler (yönetici olmayan koşu): `SKIP-ADMIN` VolumeMountPoint,
  ShortNameWithoutTilde; `SKIP-ACCOUNT` DACLOtherAccount; `SKIP-ENV` CaseSensitiveDir;
  Developer Mode kapalıysa `SKIP-PRIVILEGE` SymlinkRejected; 8.3 kapalı diskte
  `TestR13WindowsShortAlias` SKIP. Başka SKIP beklenmiyor.

## Kapsam
Kontrol (no-op hook) yazma + undo; pinli kök altında üst dizin swap'ı (bariyerli);
yazma penceresinde üst dizin swap'ı (sayısal errno); desk save/append/okuma ve
history/undo için alt dizin, oturum kökü ve kökün atası swap'ı; desk hata temizliği;
dizin ve dosya symlink; birim köküne junction (IO_REPARSE_TAG_MOUNT_POINT);
var olan dizinin ve pinli kökün sonradan junction'a çevrilmesi (retag);
açıldıktan sonra ve yazma penceresinde hard link; tilde'siz 8.3 (yönetici);
NUL/UNC/sürücü/aygıt/ADS varyantları; DACL deny (NTSTATUS 0xC0000022);
salt okunur silme (errno 5) / değiştirme; temp hata yolu; 500 döngüde handle sızıntısı.
Önceki r13 testleri (UnsafeNames, HardLink, JunctionRejected, RootSwap*, ShortAlias,
DeskAndHistory, RelativeLifecycle) aynen koşar.

## Notlar
Junction `mklink /J`, retag `FSCTL_SET_REPARSE_POINT` ve deny ACE `icacls` ile yalnız
test klasöründe kurulur; yönetici gerektirmez. Kurum ilkesi bunları engellerse test
FAIL olur: ilkeyi değiştirmeyin, logu iletin. Yerel testler geçmeden .13 güvenlik
kabulü veya dağıtım kararı verilmez.
