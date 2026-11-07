# US Panel — Hesap (accounts) modu kullanıcı rehberi

> **TASLAK, yayımlanmadı** — 6 Kasım 2026 (Europe/Istanbul, UTC+3)  
> Bu belge yalnızca yerel box dosyasıdır. Canlı sitede, mağazada veya e-postada yayınlanmamıştır.  
> Ton: marka kararı A (sade, sakin, güven veren; menşe / “yerli ve milli” sloganı kullanılmaz).  
> Olgular: Round7 raporları, accounts deploy/gap notları ve `work2/kiyas/term` kaynak kodu. Kaynakta olmayan akış uydurulmamıştır.

---

## Bu rehber ne içindir?

US Panel’in **hesap modu** (accounts) ile çalışan Windows sürümünü anlatır. Panel bir Windows EXE’dir; model ve kredi işlemleri VPS üzerindeki Go **us-license-hook** geçidinden geçer.

| Ortam | Geçit adresi (kaynak / rapor) |
|--------|-------------------------------|
| Test | `https://api-test.kanus.org` |
| Üretim (hedef) | `https://api.kanus.org` |

Hangi EXE’nin hangi adrese gömülü olduğu derleme ayarına bağlıdır. Adresleri elle Cloud Run veya eski adreslerle değiştirmeyin.

Destek: `support@kanus.org`. Telefon ve destek saatleri henüz sabitlenmedi (karar bekliyor).

---

## 1. Hesap açma (kayıt)

1. Paneli açın. **Modeller → Model bağlantısı** bölümünde hesap formunu bulun.
2. **E-posta** ve **Parola** girin.
   - Sunucu parolayı UTF-8 bayt olarak denetler: **en az 12, en fazla 72 bayt**.
   - Türkçe karakterler birden fazla bayt tutabilir; kısa görünen parola sunucuda reddedilebilir.
3. Varsa mağazadan gelen **lisans / kredi paketi anahtarı**nı (`us1.…` biçiminde) ilgili alana yazın. Kayıtta anahtar **isteğe bağlıdır**; yoksa sonra da tanımlayabilirsiniz.
4. **Lisansla hesap aç** düğmesine basın.
5. Başarıda panel **“Hesap açıldı. Giriş yapın.”** gösterir. Kayıt **otomatik giriş yapmaz**.

Aynı e-posta ile ikinci kayıt sunucuda **409** ile reddedilir. E-posta doğrulama kodu sunucu tarafında outbox’a yazılabilir; panelde doğrulanmış bir “doğrulama ekranı” **yoktur**. Doğrulama yapılmadan da giriş mümkün olacak şekilde tasarlanmıştır (posta sahipliği kanıtı sayılmaz).

---

## 2. Giriş

1. Aynı e-posta ve parola ile **Hesaba giriş** düğmesine basın.
2. Başarıda **“Hesap girişi yapıldı.”** görünür; oturum paneli belleğinde tutulur.
3. Giriş yavaş olabilir (sunucu parola doğrulaması; test ortamında onlarca saniye görülebilir). Bir süre bekleyin; peş peşe çok deneme hız sınırına takılabilir.

Yanlış parola veya tanımsız hesap için sunucu genel bir **401** döner; ayrıntılı neden panoda gösterilmez. EXE bunu genelde **“hesap işlemi reddedildi”** olarak yansıtır.

---

## 3. Lisans / kredi tanımlama (redeem)

Mağaza (Shopify) siparişi sonrası size verilen anahtar `us1.…` biçimindedir. Bu anahtar panoda “yeni ödeme” yapmaz; mevcut lisansı hesabınıza **bağlar**.

1. Önce doğru geçitte hesabınıza giriş yapın.
2. Anahtarı **Lisans veya kredi paketi anahtarı** alanına yapıştırın.
3. **Kredi yükle** düğmesine basın.
4. Başarıda bakiye satırı güncellenir (aşağıdaki biçim).

Kurallar (kaynak davranışı):

- Aynı anahtarı **aynı hesaba** tekrar yüklemek idempotenttir; ikinci kez kredi üretilmez.
- Anahtar **başka bir hesaba** bağlıysa işlem reddedilir (**409**). Krediyi kopyalayıp taşımaz.
- Süresi bitmiş, iptal/iade ile iptal edilmiş veya üzerinde açık rezervasyon bulunan anahtar bağlanamaz.
- Anahtarı / parolayı destek mesajına yazmayın.

---

## 4. Bakiyeyi görüntüleme

**Bakiyeyi göster** düğmesi sunucudan güncel bakiyeyi ister. Başarı metni kabaca şöyle görünür:

> Kalan model isteği: *N* · Saatlik bakiye: *S* saniye

- **Model isteği** ve **saatlik saniye** ayrı birimlerdir; bunları toplamayın.
- Kullanım sırasında sunucu kredi ayırır (reserve) ve bitişte kesinleştirir veya serbest bırakır (settle / release).
- Saatlik paketlerde süre **saniye** ile ölçülür. Başarılı çağrıda upstream başladıktan sonraki süre yukarı yuvarlanır; **soğuk açılış beklemesi bu süreye dahil**, proxy kabul kuyruğunda beklenen süre hariç tutulur (kaynak: model ölçüm kodu).
- **Model çağrısı ve kredi geçmişi** son kayıtları gösterir: `reserved` (ayrılmış), `completed` (kesinleşmiş), `refunded` (serbest bırakılmış çağrı). Buradaki “refunded” banka/kart iadesi demek değildir.

**Sunucu durumunu yenile** yalnızca durum bilgisi ister; modeli açmaz ve “hazır” garantisi vermez.

---

## 5. “GPU ısınıyor” ve soğuk açılış

İlk model isteğinde panel, hazırlık doğrulanana kadar benzer bir durum gösterebilir:

- `GPU ısınıyor / sunucu kuyruğu · hazır durumu henüz doğrulanmadı`
- `GPU ısınıyor / sunucu kuyruğu · bekleniyor`

Sunucu tarafı SSE yorumlarında ASCII biçiminde `GPU isiniyor; …` satırları da üretebilir. Bunlar bilgisayarınızın ısındığını veya modelin kesin hazır olduğunu kanıtlamaz.

**Beklenen bekleme (ölçülmüş örnekler, garanti değil):**

| Durum | Yaklaşık süre |
|--------|----------------|
| Korunmuş SSD ile yeniden açılış | ~**4** dakika |
| Tam soğuk açılış (ağırlık indirme + model yükleme) | ~**26–28** dakika |

Bu yüzden soğuk açılışın **yaklaşık 4–28 dakika** sürebileceğini bekleyin. Spot kesinti veya uzun boşta kalma sonrası yine soğuk açılış gerekebilir. Aynı görevi peş peşe yeniden göndermeyin; her çağrı yeni kredi kullanımı oluşturabilir.

---

## 6. Sık sorular

**Kayıt oldum ama içeride değilim.**  
Beklenen davranış. Ayrı olarak **Hesaba giriş** kullanın.

**“Kredi yükle” kartımdan para çeker mi?**  
Hayır. Bu düğme mevcut `us1.…` anahtarını hesaba bağlama isteğidir; ödeme/checkout değildir. Ödeme Shopify siparişinde yapılır.

**Parolamı unuttum.**  
Sunucuda parola sıfırlama uçları vardır; **panelde kullanıcı arayüzü yoktur**. Destek ile iletişime geçin; destek sizden parola, oturum veya lisans anahtarı istememelidir. Elle çözüm / ürün kararı bekleyen alan: aşağıdaki açık sorular.

**Anahtarı yanlış hesaba yükledim.**  
Başka hesaba bağlı anahtar ikinci hesaba alınamaz. Destek sürecinde transfer aracı kaynakta yok; karar bekliyor. Yeniden satın almadan önce destek yanıtını bekleyin.

**Bakiye neden beklediğimden düşük?**  
Aktif rezervasyonlar kullanılabilir bakiyeden düşülür. İptal/iade webhook’ları kaynağı iptal eder; hesap bakiyesi bağlı lisanslardan türetilir. Kredi ve saniye ayrıdır.

**E-posta doğrulaması zorunlu mu?**  
Hayır (mevcut tasarım). Doğrulama isteği/onay uçları vardır; panel GUI’si yoktur; SMTP ile gerçek teslimat her ortamda doğrulanmış sayılmaz.

**EXE üretime hazır mı?**  
Bu taslak onay vermez. Native Windows kabulü ve canlı model yolu ayrı doğrulanır.

---

## Açık sorular

1. Panelde parola sıfırlama / e-posta doğrulama ekranı eklenecek mi? (API var, GUI yok.)
2. Yanlış hesaba redeem sonrası destek transfer prosedürü nedir? (Kaynakta otomatik transfer yok.)
3. Üretim geçidi `api.kanus.org` kesim tarihi ve kullanıcıya gösterilecek adres metni.
4. Soğuk açılış için kullanıcıya verilecek resmi süre aralığı / SLA ifadesi (ölçüm 4–28 dk; garanti değil).
5. Destek saatleri ve telefon.
6. İndirme kanalı, imza ve güncel EXE adı/SHA (sürüm numarası raporlarda .7 / .8 olarak geçiyor).
7. SMTP allowlist dışında müşteriye sıfırlama kodunun gerçekten ulaşıp ulaşmadığı (canlı posta kabulü NOT TESTED).

