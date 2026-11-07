# Windows doğrulama seçenekleri — 8 Kasım 2026, TSİ
Harici servise kod/binary yüklenmedi; ücretli kaynak açılmadı. Microsoft ISO için yalnız HEAD (gövde indirilmedi).

| Yol | Kanıt / hazırlanmış çıktı | Maliyet / kaynak | Onay ve durum |
|---|---|---|---|
| GitHub Actions windows-latest | `../ci/windows-i18n.yml`; manual workflow_dispatch; yalnız `./i18n/` race + dört TestWinPref. GCC/mingw runtime doğrulaması. Salt okuma GitHub araması `user:demirtumen08-design is:private`: iki görünür özel repo `project`, `project-2` (kodları incelenmedi). | Resmi fiyat sayfasında Windows standard $0.010/dk; tipik 3–10 dk = $0.03–$0.10, timeout 15 dk. Dahil kota/bütçe bilinmiyor, sıfır ücret garantisi yok. | Repo seçimi + özel repoya kod yükleme + olası Actions harcaması Demir onayı ister. Workflow yerelde HAZIR; çalışmadı. En az veri için yalnız i18n/ + minimal go.mod (x/sys v0.30.0) ayrı modül düşünülebilir, proje sırları eklenmemeli. |
| Yerel QEMU/KVM + Microsoft eval | `/dev/kvm` var fakat O_RDWR -> EACCES; `qemu-system-x86_64` PATH'te yok. Disk 119 GiB boş; toplam RAM 15 GiB, available ölçümünde 3.3 GiB. | Resmi ISO HEAD **8,225,329,152 bayt = 7.66 GiB (8.23 GB)**. Windows 11 için en az 4 GB RAM / 64 GB disk; pratik 8 GiB RAM, 2–4 vCPU, 64–80 GiB dinamik disk, ISO+kurulum+Go ile 35–55 GiB gerçek kullanım tahmini. KVM ile indirme+kurulum+test ~1–2 saat; TCG yazılım emülasyonu birkaç saat/4–12 saat tahmini, ölçülmedi. 90 günlük eval, lisans şartları geçerli. | Şu an uygulanabilir değil: KVM izni yok, RAM müsait değil, QEMU/UEFI/swtpm yok. Hypervisor izni, RAM boşaltma, eval indirme+kurulum ve gerekli Microsoft hesap/lisans adımı için onay gerekir. İmaj indirilmedi. |
| Kayıtlı kişisel bilgisayar | ListMachines: bağlı `DESKTOP-48D4KOL`, id 5f142279-a00b-4f29-8cd2-1f225104511a. Etiket Windows benzeri, OS bu tur sorgulanmadı. `../ci/i18n_windows_amd64.test.exe`, `run_windows_tests.ps1`, `SHA256SUMS.txt` hazır. | Mevcut lisanslı Windows amd64 ise ek servis maliyeti yok. Binary ~9.16 MB, test birkaç saniye–dakika tahmini. Binary non-race; Go+MinGW ile source mode -race mümkün. | Kullanıcı bilgisayarına dosya kopyalama ve komut çalıştırma yerel araç onayı gerektirir. Talimat gereği bu makinede hiçbir komut çalıştırılmadı. ÖNERİLEN ilk yol. |
| Başka mevcut lisanslı Windows VM/Hyper-V | Aynı binary+PS1 veya source/race betiği; yeni kod yükleme gerekmez (dosya kullanıcının seçtiği yerel makineye taşınır). | Var olan altyapı varsa ek ücret yok; yeni ücretli VM önermiyoruz. | Erişim/transfer ve yerel komut onayı gerekir; mevcut VM varlığı UNKNOWN. Wine, WSL veya Linux cross-compile gerçek Windows NTFS/junction kabulü yerine geçmez. |

## Tercih / komutlar (şimdi ÇALIŞTIRILMADI)
1. Kendi Windows amd64 makinesinde binary + PS1 kullan (Go/GCC gerekmez). Hash kontrolü yap, sonra:
   `powershell -NoProfile -File .\run_windows_tests.ps1`
   Gerekirse kurum politikası gereği imza/onay gerekir; ExecutionPolicy kalıcı değiştirilmez.
2. Kaynakla: `powershell -NoProfile -File .\run_windows_tests.ps1 -ModuleDir C:\onayli-klasor\term` (Go + modern MinGW-w64 GCC gerekir).
3. Dört Windows testinden herhangi biri symlink yetkisi nedeniyle skip ise sonuç PARTIAL/exit 2; PASS sayılmaz. Workflow da bu nedenle hata verir. Developer Mode/yetki değişikliği ayrı onay ister.
4. Daha sonra (Demir kod yüklemeyi onaylarsa) özel repo Actions ile tekrarlanabilir race koşusu.

## Hazırlanan dosyaların sınırı
Workflow ve PowerShell betiği statik olarak hazırlandı; kutuda pwsh/actionlint yok, Windows'ta koşulmadı. YAML/PowerShell runtime doğrulaması NOT TESTED. Cross-compile edilmiş exe gerçek Windows'ta çalıştırılmadı; race binary değildir. Kaynak kod yamaları değişmedi.

## Resmi kaynaklar ve boyut kanıtı
- https://www.microsoft.com/en-us/evalcenter/download-windows-11-enterprise — 90 günlük eval, kurulum/aktivasyon/Microsoft hesabı şartları.
- Public link (HEAD): https://go.microsoft.com/fwlink/?LinkId=2382600&clcid=0x409&culture=en-us&country=us
  `windows_iso_head.txt`: Content-Length 8225329152; final URL Microsoft software-static.download.prss.microsoft.com/...en-us.iso.
- https://go.dev/doc/articles/race_detector — Windows race için cgo+C compiler, mingw-w64 runtime >=8; libsynchronization.a kontrolü.
- https://docs.github.com/en/billing/reference/actions-runner-pricing — Windows standard $0.010/dk; hesap dahil kullanımını doğrulamadık.
- https://github.com/actions/setup-go/blob/v6/README.md — checkout/setup-go v6, Go version selection.
