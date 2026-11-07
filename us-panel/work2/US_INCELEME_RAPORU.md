# US paneli inceleme ve onarım raporu

Durum: Bu çalışma bir ara sürümdür. Aşağıdaki eksikler nedeniyle panelin tüm istenen özellikleri tamamlanmış veya üretim için hazır olduğu iddia edilmez.

## Proje ve kapsam

- Kaynak ağacı: `kiyas/term`, giriş noktası `main.go`. Go arka ucu, `gui.html` gömülü HTML/JavaScript arayüzü ve yerel HTTP servisi kullanır. `kanus` altında ayrı deneysel Python ticari servis kodu vardır.
- Model istemcisi `model.go`, otomatik tünel `tunnel.go`; dosya işlemleri `desk.go`, değişiklik önerileri `proposal.go`, geçmiş `change_history.go`; test çalıştırıcısı `project_test_runner.go`; uzun görev kontrol noktası `long_checkpoint.go` içindedir.
- İncelenen kaynak ağacında 116 dosya vardı; dosya sayısı bu rapordan sonra yeni testlerle değişebilir. Bu sayım tüm bağımlılıkları veya kullanıcının bilgisayarındaki proje verilerini kapsamaz.

## Tespitler ve yapılan değişiklikler

1. `desk.go:workspaceNames` görünümü 80 öğeyle sınırlandırıyordu. Yeni `workspacePage` ve `gui.go:/api/files` ile dosya ekranına sayfalama eklendi. Gizli ve bağımlılık klasörleri ile sembolik bağlantılar atlanır; listeleme hatası başarı sayılmaz. Modelin kullandığı eski indeks ve içerik okuma sınırları tamamen kaldırılmadı.
2. `desk.go:localTextWithLimit` ikili ve geçersiz UTF-8 dosyalarını metin gibi işleyebiliyordu. Artık reddediliyor; okuma miktarı 8 MiB ile sınırlandırılıyor, açılan dosyanın türü ve kimliği kontrol ediliyor. Bölüm bölüm tam okuma henüz yok.
3. `model.go:askModel` bazı normal görevlerde okunmamış mevcut dosyaları değiştirebiliyordu. Mevcut hedefler için tam okuma zorunlu hale getirildi. `applyReplyFilesCheckedVersioned` bitiş işaretinden sonra açıklama metnini dosyaya eklemiyor, geçersiz yolu engellenen işlem olarak raporluyor ve yazılan içeriğin hash'ini yeniden okuyarak doğruluyor.
4. `proposal.go` daha önce dosya gövdesine modelin `OKU` komutunu ekleyebiliyordu. Önceki oturumda bulunan düzeltme, yeni ve diskten yüklenmiş önerilerde bu durumu reddediyor. Kullanıcının eski açık panelinde bekleyen öneri ayrıca gözden geçirilmelidir.
5. `project_test_runner.go` Python seçimini yalnızca PATH'deki `python3`/`python` üzerinden yapıyordu. Mevcut proje ortamı ve Python launcher algılama işlevi kullanıldı; sonuçta süre ve hata durumu gösteriliyor. Testler proje kodu çalıştırabilir ve kullanıcı tarafından başlatılır.
6. `gui.go:guiMux` yerel sunucuda yabancı Host başlıklarını kabul ediyordu. Loopback ve localhost dışındaki Host başlıkları reddedilir. Bu önlem tek başına kapsamlı CSRF veya güvenlik denetimi değildir.

## Gerçek testler

- Linux sunucusunda `go test ./...`: başarılı (`ok kiyas`, 5.768 saniye).
- Linux sunucusunda sayfalama, okunmamış hedef, öneri çakışması ve Host denetimi testleri `go test -race -run ... ./...` ile başarılı (`ok kiyas`, 1.092 saniye). Bu seçim tüm eşzamanlı senaryoları kapsamaz.
- Yerel ortamda `python3 -m unittest discover -s kanus -p 'test_*.py'`: 16 test başarılı, çıkış kodu 0.
- Yerel ortamda `node --check /tmp/us-gui-check.js`: HTML içindeki JavaScript için sözdizimi kontrolü başarılı, çıkış kodu 0.
- 155 dosyalı sayfalama, eksik okunan hedefi koruma, ikili dosyayı reddetme, geçersiz Host reddi, başarısız Python testi raporlama, kayıtlı bozuk öneriyi reddetme, dosya değişikliği çakışması ve geri alma, kontrollü sahte model/HTTP yanıtlarıyla otomatik testler içinde sınandı.
- Gerçek kullanıcı bilgisayarında yeni EXE ile uçtan uca görev, Windows GUI ve gerçek model testleri bu sürüm için yapılmadı. Eşzamanlı dosya işlemlerinin tüm yarış koşulları, bağlantı kopması, kaynak sınırları ve komut zaman aşımı tüm istenen kombinasyonlarla doğrulanmadı.

## Kalan eksikler

- İstenen yapılandırılmış model işlemleri şeması ve birden çok dosya için tümüyle atomik işlem/geri alma uygulanmadı. Ayrı dosya yazımları arasında kısmi başarı olabilir; yarış koşullarına karşı sistem çağrısı düzeyinde tam koruma ve bağımsız güvenlik denetimi gerekli.
- Dosya okuma aralıkları, parça bazlı bütünlük ve büyük dosyayı eksiksiz modele aktaracak uçtan uca mekanizma eksik. Sayfalama kullanıcı arayüzündedir; modelin eski indeksinde ve önizlemede sınırlamalar sürer.
- Tüm görev durumları, benzersiz ve kalıcı task_id, idempotency, görev sırası ve güvenli alt süreç iptali tamamlanmadı. Mevcut uzun görev kontrol noktası genel görev yöneticisinin yerini tutmaz.
- Model akışının devam parçalarında doğrulama başarısızlığı gerçek denemede gözlenmiştir; uzun proje kurma uçtan uca başarıyla tamamlanmadı. SSH tüneline bağımlılık ve tüm arayüz sekmelerinin gerçek kullanımla doğrulaması sürüyor.
- Test çalıştırıcısı genel bir sandbox değildir; otomatik Node test seçimi yapılmaz ve süreç ağa/dosya sistemine erişebilir. Zaman aşımı alt süreç ağacını her platformda tamamen sonlandırmayı garanti etmez.
- Yerel HTTP için kimlik doğrulama, CSRF/oturum sınırları, Windows yarış durumları, gizli veri işleme ve kapsamlı güvenlik denetimi tamamlanmadı.

## Kanus.org hazırlığı

Kaynakta ticari entegrasyon taslakları bulunmakla birlikte Shopify webhook'u, lisans e-postası, saatlik paket ve doğru istem başına kota üretime alınmış veya uçtan uca doğrulanmış değildir. VPS/GPU yaşam döngüsü, PostgreSQL/Redis kuyruğu, çok kullanıcılı kota ve maliyet denetimi, 10 eşzamanlı istem ve 10 dakikalık hedef için doğrulanmış üretim altyapısı kurulmadı. Canlıya dağıtım veya ücretli kaynak oluşturma yapılmadı.
