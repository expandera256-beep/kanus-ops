# Arşiv güvenliği ve kapsam

- Kaynaklar yalnız okunarak yeni bir klasöre kopyalandı. Orijinaller değiştirilmedi.
- Gitleaks v8.30.1 (varsayılan kurallar, satır içi izin açıklamaları dikkate alınmadan), zorunlu desenler için ripgrep ve ek kimlik bilgisi atama taraması kullanıldı.
- İlk Gitleaks geçişi 884 bulgu verdi (883 genel anahtar kuralı, 1 ödeme anahtarı kuralı). Bunlar doğrulanmış aktif sır sayısı değildir; test verileri ve yanlış pozitifler de içerebilir.
- Ad/dosya türü filtresiyle 69 dosya; içerik taramalarının birleşimiyle 7852 dosya tamamen dışarıda bırakıldı. Şüpheli içeriğin gerçek sır mı örnek mi olduğunu tahmin etmek yerine güvenli tarafta kalındı. Kaynak kod, test ve bazı raporlar bu nedenle eksik olabilir.
- Özel anahtarlar, ortam dosyaları, kimlik bilgisi depoları, çerezler ve canlı sırlar yüklenmez. Hassas dosyalar için değer veya içerik özeti yayımlanmaz. Ayrıntılı dosya yolu denetim kayıtları yalnız kutuda `/workspace/kanus-archive-audit/` altında tutulur.
- 103 ikili/büyük arşiv çıktısı, toplam 1,121,075,808 bayt, depoya alınmadı. İzlenebilir SHA256 listesi `EXCLUDED_ARTIFACTS.md` içindedir.
- İç içe Git geçmişleri, bağımlılık/derleme önbellekleri, sembolik bağlar ve ilgisiz çalışma alanları dahil edilmedi. Bağımlılık dizinleri tümüyle atlandığı için artifact sayıları bunların dosyalarını saymaz.
- Kök çalışma alanından yalnız ilgili `PANEL_S2_GAP_20261106.md` eklendi; Hanedan/Turkdini, tarayıcı profilleri, oturum günlükleri, araç çıktıları ve ilgisiz scratch içerikleri eklenmedi.
- Kutuya rsync kurulumu paket kaynağı hatasıyla başarısız olduğundan filtreli kopyalama Python standart kütüphanesiyle yapıldı.
- Bu arşiv derlenebilir tam yedek veya canlı yapılandırma yedeği değildir. Değişen kaynak klasörlerden atomik olmayan bir anlık kopyadır.
- İlk gönderimden önce temizlenmiş dosya ağacı ve Git index'i yeniden taranır; temiz sonuç olmadan gönderim yapılmaz. Tarama sonucu gelecekte eklenen dosyalar için garanti oluşturmaz.
