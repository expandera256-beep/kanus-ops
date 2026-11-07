# İşlem güvenliği ve entegrasyon durumu

Bu ara sürüm, talep edilen bütün işlerin tamamlandığı veya üretime hazır olduğu anlamına gelmez.

- İncelenen dosya önerilerinde bütün hedefler önceden doğrulanır ve yedeklenir. Yazma veya doğrulama hatasında önceki yazımlar geri alınmaya çalışılır; dış değişiklik tespit edilirse dosya üzerine yazılmaz ve geçmiş kaydı saklanır.
- Bu yaklaşım çok dosyalı atomik işlem değildir: dosyalar arasında ara durum görülebilir, uygulama çökmesi durumunda otomatik kurtarma garantisi yoktur ve diğer işlemlerin yarışlarına karşı tam koruma sağlamaz. Sadece onaylı öneri yolu kapsanmıştır.
- Okunamayan veya bozulmuş uzun görev kontrol noktasının üzerine yeni kayıt yazılmaz. Uzun bağlam hâlâ kayıplı özetleme kullanır; eksiksiz uçtan uca işleme tamamlanmamıştır.
- Görev durum makinesi, kalıcı benzersiz görev kimlikleri, idempotency ve tam iptal/kurtarma uygulanmamıştır.
- Shopify bağlantısı üzerinden yürütülen sorguda bu uygulamanın görebildiği mağaza kapsamlı webhook aboneliği bulunmadı. Bu sorgu başka uygulamaların veya uygulama yapılandırmasındaki abonelikleri kanıtlamaz. Gerçek webhook teslimi, lisans e-postası, masaüstü hesap girişi ve kota geçidinden model isteği canlı olarak doğrulanmadı.
- Go testleri Linux üzerinde: `go test ./...` başarılı, 5.859 saniye. Python ticari testleri: `python3 -m unittest discover -s kanus -p 'test_*.py'`, 16 başarılı. Bunlar canlı Shopify ya da gerçek Windows paneli testleri değildir.
- Daha önceki kapsam ve riskler için US_INCELEME_RAPORU.md dosyasına bakın.
