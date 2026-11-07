# US Panel — EXE hata ve uyarı metni önerileri (accounts)

> **TASLAK, yayımlanmadı** — 6 Kasım 2026 (Europe/Istanbul, UTC+3)  
> Yalnızca yerel box dosyası. EXE veya kaynak kod **değiştirilmedi**; bu tablo öneri listesidir.  
> Ton: marka kararı A — kısa, sakin, ne olduğu + ne yapılabileceği. Abartı / aciliyet yok.  
> Satır numaraları: `/workspace/us-panel/work2/kiyas/term/` (Round7 dönemi kaynak).

---

## Nasıl okunur?

| Sütun | Anlam |
|--------|--------|
| Mevcut metin | Kullanıcıya gidebilecek mevcut Türkçe (veya SSE ASCII) ifade |
| Kaynak | `dosya:satır` |
| Önerilen | Daha kısa ve net Türkçe taslak (henüz uygulanmadı) |

Öneri metinleri ürün onayı olmadan EXE’ye alınmamalıdır.

---

## A — Hesap istemcisi (`account_client.go`)

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| hesap işlemi geçersiz | `account_client.go:49` | Desteklenmeyen hesap işlemi. |
| hesap bilgisi çok uzun | `account_client.go:52` | E-posta, parola veya lisans alanı çok uzun. Kısaltıp yeniden deneyin. |
| hesap sunucusuna ulaşılamadı | `account_client.go:57`, `:101`, `:142` | Hesap sunucusuna bağlanılamadı. Ağı kontrol edip biraz sonra yeniden deneyin. |
| hesap işlemi reddedildi | `account_client.go:62` | Hesap işlemi kabul edilmedi. Bilgileri kontrol edin veya biraz bekleyip yeniden deneyin. |
| geçersiz hesap yanıtı | `account_client.go:71` | Sunucu yanıtı beklenen biçimde değil. Daha sonra yeniden deneyin. |
| hesap bilgisi geçersiz | `account_client.go:89`, `:131` | Oturum veya lisans bilgisi geçersiz. Yeniden giriş yapın. |
| hesap bağlantısı kurulamadı | `account_client.go:94`, `:135` | Hesap isteği hazırlanamadı. Uygulamayı yeniden açıp deneyin. |
| kredi işlemi reddedildi | `account_client.go:106` | Kredi / lisans işlemi kabul edilmedi. Giriş ve anahtarı kontrol edin. |
| geçersiz bakiye yanıtı | `account_client.go:113` | Bakiye yanıtı okunamadı. Daha sonra yeniden deneyin. |
| istek geçmişi alınamadı | `account_client.go:147` | Çağrı geçmişi alınamadı. Giriş durumunu kontrol edin. |
| geçersiz istek geçmişi | `account_client.go:153` | Çağrı geçmişi beklenen biçimde değil. |
| geçersiz istek kaydı | `account_client.go:157` | Çağrı kaydı geçersiz. |
| geçersiz istek kimliği | `account_client.go:161` | Çağrı kimliği geçersiz. |
| geçersiz istek durumu | `account_client.go:165` | Çağrı durumu tanınmıyor. |

---

## B — Panel bildirimleri (`gui.go`)

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| Hesap açıldı. Giriş yapın. | `gui.go:548` | Hesabınız oluşturuldu. Şimdi giriş yapın. *(mevcut metin zaten net; isteğe bağlı yumuşatma)* |
| Hesap girişi yapıldı. | `gui.go:553` | Giriş başarılı. |
| Önce bu geçitte hesabınıza giriş yapın. | `gui.go:563`, `:588`, `:609` | Bu sunucu için önce hesabınıza giriş yapın. |
| Bağlantı değişti; hesap işlemi uygulanmadı. | `gui.go:540` | Bağlantı adresi değişti; işlem iptal edildi. Yeniden deneyin. |
| Hesap değişti; bakiye gösterilmedi. | `gui.go:576` | Oturum veya bağlantı değişti; bakiye yenilenmedi. |
| Hesap değişti; istek geçmişi gösterilmedi. | `gui.go:597` | Oturum veya bağlantı değişti; geçmiş yenilenmedi. |
| Hesap değişti; sunucu durumu gösterilmedi. | `gui.go:618` | Oturum veya bağlantı değişti; durum yenilenmedi. |
| Kalan model isteği: N · Saatlik bakiye: S saniye | `gui.go:583` | Kalan istek: N · Saatlik süre: S sn *(birimleri ayrı tutmaya devam edin)* |
| Son model çağrılarının kredi durumu güncellendi. | `gui.go:605` | Son çağrıların kredi durumu güncellendi. |
| Sunucu durumu alındı; bu işlem model isteği başlatmaz. | `gui.go:622` | Sunucu durumu alındı. Bu işlem modeli açmaz. |
| Model çağrısı sürüyor. | `gui.go:526`, `:557` | Model çağrısı sürüyor. Bitmesini bekleyin. |
| Bağlantı adresi güncellendi. | `gui.go:523` | Bağlantı adresi kaydedildi. |
| Lisans değişti; önceki bakiye gösterilmedi. | `gui.go:318` | Lisans değişti; önceki bakiye gizlendi. |
| Go sunucusu kredi durumu alındı; model çağrılmadı. | `gui.go:327` | Sunucu bakiyesi alındı. Model çağrılmadı. |
| Görev yeniden sıraya alındı; model çağrısı yeni kredi kullanabilir. | `gui.go:373` | Görev yeniden sıraya alındı. Yeni çağrı kredi kullanabilir. |
| Python hesap geçidi bu Go lisans sürümünde kapalıdır. | `gui.go:273` | Bu sürümde eski hesap geçidi kapalıdır. Accounts sürümünü kullanın. |

---

## C — Form ve arayüz (`gui.html`)

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| E-posta / Parola / Lisans veya kredi paketi anahtarı | `gui.html:200` | Alan etiketleri yeterli; isteğe bağlı ipucu: “Parola en az 12 karakter (UTF-8).” |
| Hesaba giriş | `gui.html:200` | Giriş yap |
| Lisansla hesap aç | `gui.html:200` | Hesap oluştur |
| Kredi yükle | `gui.html:200` | Lisansı hesaba bağla *(ödeme çağrışımını azaltır)* |
| Bakiyeyi göster | `gui.html:200` | Bakiyeyi yenile |
| Model çağrısı ve kredi geçmişi | `gui.html:201` | Çağrı ve kredi geçmişi |
| Hesap geçmişi henüz alınmadı. | `gui.html:201` | Geçmiş henüz yüklenmedi. |
| Hesap bağlantısı bekleniyor. | `gui.html:101`, `:203` | Hesap bağlantısı bekleniyor. |
| Sunucu bakiyesi henüz doğrulanmadı. | `gui.html:199` | Sunucu bakiyesi henüz alınmadı. |
| Hesap girişi ve durum sorguları model sunucusunu açmaz. Model isteği kredi rezervasyonundan sonra işlenir. | `gui.html:205` | Giriş ve durum sorgusu modeli açmaz. Model isteği krediyi ayırdıktan sonra işlenir. |
| 1 kredi kullanıldı / Kredi iade edildi / İşleniyor (rezervasyon) | `gui.html:421` | 1 kredi kullanıldı / Çağrı kredisi serbest bırakıldı / İşleniyor (ayrılmış) *(banka iadesi sanılmasın)* |
| Hesap işlemi yapılamadı. | `gui.html:785` | Hesap işlemi tamamlanamadı. |
| İşlem tamamlandı. | `gui.html:785` | İşlem tamamlandı. |

---

## D — Sunucu durumu istemcisi (`worker_client.go`)

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| Sunucu durumu için hesap girişi gerekli. | `worker_client.go:27` | Sunucu durumu için önce giriş yapın. |
| Sunucu durum isteği hazırlanamadı. | `worker_client.go:31` | Durum isteği hazırlanamadı. |
| Sunucu durumu alınamadı. | `worker_client.go:38` | Sunucu durumu alınamadı. Biraz sonra yeniden deneyin. |
| Sunucu yönetimi doğrulanamadı. | `worker_client.go:43` | Sunucu durumu doğrulanamadı. |
| Geçersiz sunucu durumu. | `worker_client.go:49`, `:53` | Sunucu durumu beklenen biçimde değil. |

---

## E — Geçit / kredi (`gateway.go`)

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| Ticari sürüm için uzak HTTPS hesap geçidi gerekir. | `gateway.go:38` | Ticari sürüm için HTTPS hesap sunucusu gerekir. |
| Geçerli bir kapı adresi girin (yalnızca sunucu ve isteğe bağlı port). | `gateway.go:42` | Geçerli bir sunucu adresi girin (isteğe bağlı port). |
| Model adresi geçersiz. | `gateway.go:49` | Model adresi geçersiz. |
| Uzak model için HTTPS gerekir. HTTP yalnızca bu bilgisayardaki bir SSH tüneli için kullanılabilir. | `gateway.go:59` | Uzak model için HTTPS gerekir. |
| kredi servisine ulaşılamadı | `gateway.go:115` | Kredi servisine ulaşılamadı. |
| kredi servisi HTTP %d | `gateway.go:119` | Kredi servisi yanıt kodu: %d. |
| kredi servisi yanıtı geçersiz | `gateway.go:123` | Kredi servisi yanıtı okunamadı. |
| Önce Go lisans anahtarını kaydedin. | `gateway.go:169` | Önce lisans anahtarını kaydedin. |
| Kredi durumu doğrulanamadı. | `gateway.go:184` | Kredi durumu alınamadı. |
| Kredi durumu HTTP %d; bakiye doğrulanmadı. | `gateway.go:188` | Kredi durumu alınamadı (kod %d). |
| Kredi durum yanıtı geçersiz. | `gateway.go:205` | Kredi durumu yanıtı okunamadı. |

---

## F — “GPU ısınıyor” / soğuk açılış

| Mevcut metin | Kaynak | Önerilen kısa Türkçe |
|--------------|--------|----------------------|
| GPU ısınıyor / sunucu kuyruğu · hazır durumu henüz doğrulanmadı | `life.go:30` | Model hazırlanıyor / kuyrukta. Hazır olduğu henüz doğrulanmadı. |
| GPU ısınıyor / sunucu kuyruğu · bekleniyor | `life.go:126` | Model hazırlanıyor / kuyrukta. Lütfen bekleyin. |
| Model yanıtı akıyor | `life.go:128`, `model.go:2000` | Model yanıtı geliyor. |
| GPU isiniyor; model hazirligi dogrulanmadi | `licensesrv/model_proxy.go:185` *(SSE, ASCII)* | Panel tarafında Türkçe gösterim: “Model hazırlanıyor; hazırlık henüz doğrulanmadı.” |
| GPU isiniyor / kuyruk bekleniyor | `licensesrv/model_proxy.go:199` | Model hazırlanıyor / kuyruk bekleniyor. |

**Kullanıcı rehberi notu (metin değil, bağlam):** soğuk açılış ölçümlerinde yaklaşık **4–28 dakika** görülebilir; bu süre garantisi değildir.

---

## G — Bilinçli olarak dokunulmayan / eksik UI

| Konu | Durum |
|------|--------|
| Parola sıfırlama ekranı | Sunucu API var; EXE formu **yok** → metin önerisi için önce UI kararı |
| E-posta doğrulama ekranı | Aynı |
| HTTP 401/409/429’u aynen gösteren metin | EXE genel cümleye çeviriyor; kod bazlı ayrıştırma **yok** → istenirse sonraki tur |

---

## Açık sorular

1. “Kredi yükle” düğme adı “Lisansı hesaba bağla” olarak değiştirilsin mi? (ödeme algısı)
2. SSE ASCII `isiniyor` satırları Türkçe Unicode’a çekilsin mi, yoksa yalnız panel `modelStage` mi gösterilsin?
3. Hata metinlerinde HTTP kodu kullanıcıya gösterilsin mi? (destek için faydalı, son kullanıcı için gürültülü olabilir)
4. Parola uzunluğu ipucu formda sabit metin olarak konulsun mu?
5. Bu önerilerin hangileri bir sonraki EXE derlemesine alınacak? (bu dosya tek başına kod değişikliği değildir)

