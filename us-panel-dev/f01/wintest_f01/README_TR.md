# FINDING-01 Windows Low IL native kabul paketi

Bu paket Linux'ta çapraz derlendi; native Windows sonuçları NOT TESTED.
KAN.US, Demir'in PC'sinde çalıştırıp `sonuc_amd64.log` / `sonuc_386.log` dosyasını geri iletmeli.
Yönetici olarak çalıştırmayın. Normal kullanıcı hesabı, NTFS ve Windows 10/11 önerilir.
Go, node ve python yoksa her biri için `TOOL_SKIP ... araç PATH içinde yok` açıkça yazılır;
araç var fakat Low IL altında başlayamazsa test FAIL olur. Paket mevcut proje dosyalarına dokunmaz.

## Çalıştırma
1. ZIP'i yeni ve yalnız bu teste ayrılmış bir klasöre açın.
2. PowerShell'de `Get-FileHash .\f01_amd64.test.exe -Algorithm SHA256` ve 386 için aynı komutla SHA256SUMS'u karşılaştırın.
3. CMD'de `run.cmd` çalıştırın (otomatik amd64/x86 seçimi). İkinci doğrulama: `run.cmd 386`.
4. Sonuç kodu sıfır ve `PASS` beklenir. Loglar paketin açıldığı klasöre yazılır.
5. `TOOL_LOW_IL_OK go/node/python` satırlarını ve varsa TOOL_SKIP satırlarını rapora ekleyin.

## Tam olarak ne test edilir?
- Üretimle aynı `safeCmd.Start` + `prepareChildSandbox` kullanılır; mevcut Job Object uygulanır.
- Çocuk tokenının integrity SID'si S-1-16-4096 (Low), admin SID'si varsa deny-only ve
  SeChangeNotifyPrivilege dışındaki ayrıcalıkların silinmiş olduğu doğrulanır.
- Workspace içindeki yeni dosyaya VE önceden var olan dosyaya yazma çalışır.
- Aynı geçici kökün kardeşi olan, açıkça Medium IL etiketli klasöre yazma ERROR_ACCESS_DENIED verir.
- Dışarıdaki sentetik dosyayı OKUYABİLDİĞİ ayrıca doğrulanır: OUTSIDE_READ_STILL_ALLOWED.
- Low IL yardımcı çocuğu go version / node --version / python --version çalıştırır;
  alt araçlar bu Low IL tokenını ve Job Object'i kalıtır. Ağ veya proje test/build işlemi yapılmaz.

## Workspace Low IL etiketi nasıl uygulanır?
Üretim kodu `labelLowWorkspace` önce mevcut reparse point'leri reddeder.
`windows.SecurityDescriptorFromString("S:(ML;OICI;NW;;;LW)")` ile mandatory-label SACL oluşturur;
`sd.SACL()` sonucunu `windows.SetNamedSecurityInfo(root, SE_FILE_OBJECT,
LABEL_SECURITY_INFORMATION, nil, nil, nil, sacl)` çağrısına verir.
OI (object inherit) ve CI (container inherit) alt dosya/klasörlere Low etiketi taşır.
NW no-write-up demektir; DACL/owner değiştirilmez. Başarısızsa komut başlamaz ve UYARI görünür.
Bu etiket kalıcı dosya sistemi metaverisidir. Native test yalnız `testing.T.TempDir()` altını
etiketler ve test bitince kendi geçici klasörünü temizler; gerçek projeleri etiketlemez.
Gerçek panel bir projeyi çalıştırırken o projenin Low etiketini değiştirir; deployment değildir.

## Sınırlar
Low IL yalnız daha yüksek integrity'ye yazmayı engeller. Diğer Low IL konumlarına yazmayı
veya ağ erişimini engelleyen bir workspace sınırı değildir. Windows'ta okuma izolasyonu yok.
AppContainer ya da proje başına ayrı OS kullanıcısı olmadan FINDING-01 kapanmaz.
İş nesnesine atama mevcut tasarımda Start sonrası yapılır; bu paket onu değiştirmez.
Reparse-point ön kontrolü ve SetNamedSecurityInfo yol tabanlıdır; eşzamanlı root-swap için
handle-temelli etiketleme ayrıca incelenmelidir. r13 dosya araçları yamasından bağımsızdır.

## Sürüm notu
FINDING-01 için bu sürüm yalnız risk azaltımı sağlar; bulgu kapanmamıştır. Linux'ta Landlock, proje kodunun kardeş çalışma alanlarını okumasını sınırlar. Windows'ta okuma izolasyonu yok: proje kodu çalışma alanı dışındaki dosyaları hâlâ OKUYABİLİR. Restricted token ve Low integrity level (Low IL), bu okuma riskini ortadan kaldırmaz. KAN.US üzerinden aktarılan karara göre Demir, FINDING-01 kapsamındaki bu Windows riskini 2026-11-08 tarihinde geçici olarak kabul etmiştir (seçenek e). AppContainer ve proje başına ayrı işletim sistemi kullanıcısı seçenekleri daha sonra yeniden değerlendirilecektir.
