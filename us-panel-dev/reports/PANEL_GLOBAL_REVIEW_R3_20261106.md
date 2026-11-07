# PANEL GLOBAL R3 incelemesi

Tarih: 08.11.2026 Europe/Istanbul. Kaynaklar/yamalar değiştirilmedi; yalnız kalıcı scratch kopyalarında çalışıldı. Mesaj, deploy, EXE çalıştırma yok.

## Sonuç
**FAIL (LOW kaynak yaşam döngüsü bulgusu açık); Linux güvenlik/işlev düzeltmeleri PASS.** Nihai .11→UYUM→API→GLOBAL yığını **BLOCKED**: sabit .11 teslimi yok. Windows ve macOS gerçek çalışma **NOT TESTED**.

| Yama | Sonuç |
|---|---|
| panel_core | PASS |
| 02_gui_strings | PASS |
| 03_pref | FAIL — R3-01 LOW |
| 04_errors | PASS |
| 05_email_loader | PASS |

## Açık bulgu
**R3-01 LOW — `i18n/pref.go:79-102`, `i18n/pref_unix.go:17-19`, `i18n/pref_windows.go:64-73`: kök handle'ları hiçbir zaman kapatılmıyor.** Kök önbelleğinde sınır/eviction ve PrefStore.Close/finalizer yok. Farklı Root değerleri biriktikçe descriptor sayısı artıyor; Store terk edilse de os.File değil ham fd tutulduğu için GC kapatmıyor. Bağımsız test: 64 kökte fd **7→71**, GC sonrası da 71; `fd_growth.log` gerçek FAIL. Sabit tek köklü GUI'de etkisi süreç başına bir handle'dır; uzaktan tetiklenebilen saldırı iddiası yok. Süreç ömürlü pinleme tasarımı korunarak açık yaşam-döngüsü ve dinamik kök için sınırlama gerekir.

## R2 kapanış / kod incelemesi
- R2-01 symlink HOME/config: **PASS**. İlk çözüm kilit altında tek kez; fd sabit, alt bileşenler `openat(O_NOFOLLOW|O_DIRECTORY)`, dosya O_NONBLOCK+fstat regular+4096 sınırı. `..`, mutlak/volume/ADS adları reddediliyor.
- FIFO/symlink/özel dosya okuma ve alt symlink yazma reddi PASS. Rastgele O_EXCL temp, fsync, aynı dirfd üzerinde renameat, hata halinde unlinkat; ara fd ve temp fd kapanıyor. Dir fsync hatası yutuluyor (dayanıklılık sınırı; yeni saldırı iddiası değil).
- Kök açılamazsa dil bellek map'inde, POST 503; mevcut oturum GET belleği okuyabilir. GUI POST önce, yalnız 2xx sonra locale yükler; 500/503/ağ hatasında eski seçici+yerelleştirilmiş hata. `gui.lang.save_failed` tr/en mevcut. Hızlı eşzamanlı kullanıcı değişimi için gerçek tarayıcı E2E NOT TESTED.
- Windows statik inceleme: NtCreateFile RootDirectory, FILE_OPEN_REPARSE_POINT, açılmış handle üzerinden reparse/tür/boyut doğrulaması; OBJECT_ATTRIBUTES.Length ve x/sys UNICODE_STRING kullanımı uygun. FileRenameInformation alan hizalaması ptr doğal hizalamasını kullanıyor (amd64 FileName offset 20, 386 offset 12); değişken UTF16 buffer uzunluğu null hariç. Alt/old/temp handle kapanışları var, kök kapanışı R3-01. Rename hedef reparse semantiği/ACL/junction **gerçek Windows'ta NOT TESTED**; cross-build çalışma kanıtı değildir.

## Gerçek önce/sonra kanıtı
Aynı davranış testleri yalnız eski Path / yeni Root+Name kurucusuna uyarlanmıştır (üretim kodu değiştirilmedi):
- Symlink kök: eski FAIL / yeni PASS.
- Kök symlink yeniden bağlama: eski ilk kayıtta FAIL / yeni PASS, yeni hedefe yazı yok.
- UI 500/503/network: eski 3 FAIL / yeni 3 PASS; 200 her ikisinde PASS.
- Kanıt: `/workspace/us-panel-dev/scratch/review-20261108/A/old_root.log`, `tree_root.log`, `old_ui.log`, `tree_ui.log`; probe kaynakları `probes/` altında.

## Testler
- Manifest **181/181 PASS**, `../manifest_verify.json`.
- 01→04 git apply --check/apply **4/4 PASS**; 05 Python kökünde PASS. 03/04/05 tam SHA beklenen değerlerle eş; `hashes.json`.
- Go build/vet, GOOS=windows vet/build **PASS**; Windows i18n test binary derleme PASS (çalıştırılmadı).
- Tam race ilk koşu **414 PASS / 1 FAIL / 7 SKIP**: yalnız fixture yolu. Test `licensesrv/credits_test.go:17` dört üst dizin reports bekliyor; fixture'lar önce kopya kökteydi. Kaynak test değiştirilmeden scratch/reports'a da kopyalandı.
- Fixture düzeltilmiş `go test -race -count=1 -p=1 -json ./...`: **415 PASS / 0 FAIL / 7 SKIP**, exit 0. DATA RACE raporu yok. Bağımsız beklenen-FAIL fd probe bu tam suite'e dahil değildir.
- §2 regresyon **36/36 PASS**, kaynak kopya manifesti değişmedi. Orijinal runner yalnız scratch'e kopyalanarak /tmp yolu kalıcı scratch'e yönlendirildi, temizlik kaldırıldı; PoC/assertion değişmedi.
- Tüm ağaç gofmt -l **30 taban dosya** listeliyor; değişen GLOBAL Go dosyalarında çıktı **boş/PASS**; kaynaklara format yazılmadı.
- node check_ui **PASS**, gui_lang_ui_test **4/4 PASS**; Python test_i18n_mail **10/10 PASS**.
- Test env: GOPROXY/GOSUMDB off; dört gerçek model/network bayrağı 0; geçici scratch task anahtarı; kişisel config anahtarı kullanılmadı.

Kalıcı kanıt: `/workspace/us-panel-dev/scratch/review-20261108/A/`. Önceki bot raporları yeni test sonucu yerine kullanılmadı. İşletim kararı kaynağı KAN.US aktarımıdır, Demir'in doğrudan onayı diye sunulmadı.

Ek: GLOBAL_R3_REVIEW_GUIDE.md okundu; satır 52 Close/eviction eksikliğini zaten bilinen sınır olarak bildiriyor. R3-01 keşif iddiası değil bağımsız yeniden üretimdir; rehberin test iddiaları bu rapordaki kendi koşularımızla doğrulandı.
