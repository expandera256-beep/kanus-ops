# .13 Windows handle-tabanlı dosya katmanı — yerel test paketi

Bu paket **sürüm EXE'si değildir**. Linux'ta Windows için çapraz derlenmiştir;
yerel Windows çalıştırması henüz yapılmadı (NOT TESTED).

## Adımlar
1. Paketi Windows 10/11 PC'de kullanıcı hesabının yazabildiği **yerel NTFS**
   bir klasöre açın (ör. Masaüstü\r13-wintest). Ağ paylaşımı veya FAT/exFAT kullanmayın.
2. Normal (yönetici olmayan) Komut İstemi açın. Bu klasöre `cd /d` ile geçin.
3. İsterseniz `certutil -hashfile r13_windows_amd64.test.exe SHA256` ile
   SHA256SUMS.txt içindeki özeti karşılaştırın. 32-bit Windows için 386 dosyası vardır.
4. `run.cmd` çalıştırın. Go kurulumu, internet, GPU, canlı servis veya lisans gerekmez.
   Testler yalnız Windows'un geçici klasöründe kendi ürettikleri dosyaları kullanır.
5. `windows-results.log` dosyasını KAN.US/PANEL DEV'e iletin. Betiğin exit kodu 0 olmalı.

## PASS nasıl görünür?
- Çıktı sonunda Go `PASS`, ardından betik `PASS - Windows dosya katmani...` yazmalı.
- `FAIL` veya `panic` olmamalı. Her test üç defa çalışır.
- `TestR13WindowsRootSwapStress/write` ve `/delete` satırlarının her tekrarında
  `swaps` ve `successful_operations` sıfırdan büyük olmalı.
- Yalnız `TestR13WindowsShortAlias`, diskte 8.3 ad üretimi kapalıysa SKIP olabilir.
  Bu durumda gerçek 8.3 alias deneyi NOT TESTED olarak raporlanır; tilde/ADS/device
  politika testleri yine çalışır. 8.3'ü açmak için yönetici ayarı değiştirmeyin.
- Parent-swap yardımcı testi, Windows açık temp dosyalı dizini taşımayı reddederse
  güvenli hata dönüşünü kabul eder; dış dizindeki sentinel kesinlikle aynı kalmalıdır.

## Kapsam / izinler
Okuma, yeni dosya, üzerine atomik yazma/rename, silme, hard link reddi, junction
reddi, kök junction-swap (deterministik + yarış), ADS, 8.3, ayrılmış aygıt adları,
masaüstü okuma ve geçmişten geri alma kontrol edilir.
Junction `cmd /d /c mklink /J` ile kurulur; yerel NTFS'te yönetici veya Developer
Mode gerektirmez. Symlink kullanılmaz. Junction kurulamıyorsa test FAIL olur;
kurum ilkelerini değiştirmeyin, hata kaydını iletin.

## Önemli teknik ayrıntı
Rename, hedef ebeveyn HANDLE'ını `RootDirectory` olarak kullanan
`NtSetInformationFile(FileRenameInformation)` ile yapılır. Win32
`SetFileInformationByHandle(FileRenameInfo)` bazı Windows sürümlerinde göreli adı
cwd'ye göre dönüştürebildiğinden kullanılmaz; mutlak yol fallback'i yoktur.
Silme `SetFileInformationByHandle(FileDispositionInfo)` üzerindendir.
Yerel Windows testleri geçmeden .13 güvenlik kabulü veya dağıtım kararı verilmemeli.
