# kanus.org fiyatlandırma güncelleme planı — SALT OKUMA

8 Kasım 2026, 01:08–01:12 TSİ. Hiçbir canlı dosya, Shopify ürünü, fiyat, DNS veya hesap ayarı değiştirilmedi. Bu dosya yalnız inceleme ve değişiklik taslağıdır.

## 1. Hosting ve kaynak
- Hostinger `hosting_websites_list(domain=kanus.org)` bu turda okundu: website_type=other, Horizons UUID yok, enabled=true; hesap `u346856246`, order_id `1010284397`.
- Canlı kök: `/home/u346856246/domains/kanus.org/public_html`.
- Statik React SPA (Vite biçimli derlenmiş bundle) + PHP api/. Shopify teması değildir. DNS IP adresleri tek başına kaynak yolu değildir.
- Bu tur dosya listesi: kökte index.html, deneme.html, us.html, fx.json, assets/, api/, downloads/, storage/. /tr, /en, /de, /ru dizinleri yok; dil `localStorage["kanus.lang"]`, varsayılan EN.
- Gerçek JSX/TSX SPA kaynak deposu bulunamadı. `/workspace/us-panel` önceki 6 Kasım envanterinde de yoktu. Mevcut GLOBAL çalışma alanı `/workspace/us-panel-dev/global-work/site/` içinde canlı bundle kopyaları ve paket üreticileri var. Bunlar orijinal kaynak repo değildir.
- Üretici giriş: `/workspace/us-panel-dev/global-work/site/us/live_app.js`; i18n üretici `build_us.py`, kur katmanı `build_us_fx.py`. Sonraki `build_us_fx2.py` ve release_us_fx2 mevcut ancak bu tur canlı hash fx2 değil fx ile eşleşti. Yeni çalışma öncesi GLOBAL ile sürüm çakışması kontrol edilmeli.
- Canlıyla birebir yerel paket: `/workspace/us-panel-dev/global-work/site/release_us_fx/public_html/assets/index-uscommerce-i18nfx-5a375f29.js`.
- Bu tur kanıt kopyaları: `reports/r20261108/pricing-evidence/{index.html,index-uscommerce-i18nfx-5a375f29.js,deneme.html}`.
- Canlı index SHA256: `4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8`.
- Canlı JS SHA256: `5a375f292604694e6f905a9654353700edf9dc4bb595c5df250354048991c5d0`.

## 2. Görünen metinler ve kapsam
Önemli: Fiyatlar sayısal fiyat bileşeninde değil, ana sayfa `usCard` paragrafında gömülü! Önceki GLOBAL raporundaki “sitede fiyat öğesi yok” ifadesi fiyat metni yok anlamına gelmez. Önceki canlı render `fx_tr_home.png` görseli de 200 USD / 500 istem ve 10 USD / 1 saat yazısını doğruluyor.

Bu tur canlı HTML/JS ve statik HTML GET ile okundu. Yeni tarayıcı render matrisi çalıştırılmadı: bu yürütücüde yetkili tarayıcı alt-ajan aracı yok. Önceki aynı hash'li render kanıtı `GLOBAL_LIVE_I18N_TEST_20261108.md` ve `/workspace/us-panel-dev/global-work/tur_live_20261108/` altında mevcut. Bu nedenle aşağıdaki rota eşleştirmesi canlı bundle incelemesidir, tüm dinamik kullanıcı içeriklerinin yeni render sertifikası değildir.

| URL | Plan/fiyat/kredi görünümü |
|---|---|
| https://kanus.org/ | TR/EN/DE/RU: `usCard`, aylık 200 USD / 500 istem, saatlik 10 USD / 1 saat. Tam metinler aşağıda. |
| https://kanus.org/applications | US bağlantısı; fiyat göstermez. `appsIntro` sözlükte var ama bileşende kullanılmıyor. |
| https://kanus.org/us | Fiyat göstermez. Lisans ve hesap erişimi hazırlanıyor, otomatik hesap/lisans teslimi etkin değil uyarısı; 4 dilde `trialLink`: 20 istem/prompts/Anfragen/запросов. `usIntro` sözlükte var ama kullanılmıyor. |
| https://kanus.org/payments | SPA yönlendirmesi /us. Eski Q5/Q8 ve 1 USD=10 kredi sözlükleri bundle'da var ama bu rota bunları render etmiyor. Bunları mağaza fiyatı sanıp değiştirmeyin. |
| https://kanus.org/deneme.html | Yalnız Türkçe: deneme 20 istem, aylık abonelik 200 USD / 500 istem, kullanım 10 USD / 1 saat işlem veya açık panel. |
| https://kanus.org/us.html | Meta refresh /us; kendi fiyatı yok. |
| https://kanus.org/news, /gallery, /forum, /contact | Statik bileşenlerde fiyat/plan yok. Haber/galeri/forum sunucu kaynaklı içerikleri bu tur tek tek denetlenmedi. |
| https://kanus.org/panel, /bookings, /snapshots | SPA yönlendirmesi sırasıyla /, /, /gallery. |

## 3. Mağaza hedefleriyle fark (TRY, KDV dahil)
Hedefler delegasyondaki canlı Shopify doğrulamasından ve PLANS_LIVE_DIFF_20261106.md / LIVE_STORE_AUDIT_20261106.md referanslarından alınmıştır; bu tur Shopify fiyatı değiştirilmedi.

| Paket | kanus.org mevcut | Hedef |
|---|---|---|
| 1 kredi | Yok | 149 TL; **DRAFT/satış dışı, satın alma seçeneği eklenmemeli** |
| 5 kredi | Yok | 695 TL |
| 10 kredi | Yok | 1.290 TL |
| 25 kredi | Yok | 2.990 TL |
| 50 kredi | Yok | 5.490 TL |
| 100 kredi | Yok | 9.900 TL |
| 1 Aylık Erişim / 200 İstem | 200 USD, 500 istem, abonelik ifadesi | 18.900 TL, 200 istem; tek seferlik, yenileme yok |
| 1 Saatlik Kullanım | 10 USD, 1 saat | 1.990 TL, 1 saat |
| Bulut denemesi | 20 istem | Mağaza ürünü değil; fiyat uydurulmadı, değişiklik önerilmiyor |

USD fiyatlar satış bedeli olarak kaldırılmalı; FX bileşeni yalnız açıklayıcı kur sunmalı, yeni bir sabit USD satış fiyatı icat edilmemeli.

## 4. Kesin dize değişiklikleri — öneri, uygulanmadı
Ana sayfa için zorunlu 4 değişiklik `usCard`; kullanılmayan `appsIntro` ve `usIntro` için 8 değişiklik eski iddiaları ileride yeniden görünür kılmamak içindir. `usIntro` yeni metni otomatik e-posta vaadi içermez; B-B onaylı teslimat metinlerinin canlı webhook ve e-posta doğrulaması koşulu aynen korunur. /us içindeki mevcut uyarı bu fiyat çalışmasında değiştirilmez.

Dosya yolları:
- Doğrudan güncel bundle kopyası: `/workspace/us-panel-dev/global-work/site/release_us_fx/public_html/assets/index-uscommerce-i18nfx-5a375f29.js`.
- Üretim zincirinde kalıcı giriş karşılığı: `/workspace/us-panel-dev/global-work/site/us/live_app.js` (aynı sözlük dizeleri). Gelecek geliştirici ikisini bağımsız yamamak yerine üretim zincirini tek kaynaktan güncellemeli ve hash'li yeni paket üretmeli; canlıdaki hash'li asset üzerine yazmamalı.

### EN · `usCard` · https://kanus.org/
Eski:
```text
Commercial access. Write to support@kanus.org. 200 USD for 500 prompts, or 10 USD for 1 hour.
```
Yeni:
```text
Commercial access. Write to support@kanus.org. 1 month / 200 prompts: 18,900 TRY. 1 hour: 1,990 TRY. VAT included. One-time purchases; no automatic renewal.
```

### TR · `usCard` · https://kanus.org/
Eski:
```text
Ticari erişim. support@kanus.org adresine yazın. 200 USD ile 500 istem, ya da 10 USD ile 1 saat.
```
Yeni:
```text
Ticari erişim. support@kanus.org adresine yazın. 1 Aylık Erişim / 200 İstem: 18.900 TL. 1 Saatlik Kullanım: 1.990 TL. KDV dahil. Tek seferlik satın alım; otomatik yenileme yok.
```

### DE · `usCard` · https://kanus.org/
Eski:
```text
Kommerzieller Zugang. Schreiben Sie an support@kanus.org. 200 USD für 500 Prompts oder 10 USD für 1 Stunde.
```
Yeni:
```text
Kommerzieller Zugang. Schreiben Sie an support@kanus.org. 1 Monat / 200 Prompts: 18.900 TRY. 1 Stunde: 1.990 TRY. Inklusive Mehrwertsteuer. Einmaliger Kauf; keine automatische Verlängerung.
```

### RU · `usCard` · https://kanus.org/
Eski:
```text
Коммерческий доступ. Напишите на support@kanus.org. 200 USD за 500 запросов или 10 USD за 1 час.
```
Yeni:
```text
Коммерческий доступ. Напишите на support@kanus.org. 1 месяц / 200 запросов: 18 900 TRY. 1 час: 1 990 TRY. НДС включён. Разовая покупка; без автоматического продления.
```

### EN · `appsIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe is the Windows application. Monthly subscription is 200 USD for 500 prompts. Usage billing is 10 USD for 1 hour of processing or an open panel.
```
Yeni:
```text
US.exe is the Windows application. 1 month / 200 prompts: 18,900 TRY. 1 hour of processing or an open panel: 1,990 TRY. VAT included. One-time purchases; no automatic renewal.
```

### TR · `appsIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe Windows uygulamasıdır. Aylık abonelik 200 USD, 500 istem. Kullanım 10 USD, 1 saat işlem veya açık panel.
```
Yeni:
```text
US.exe Windows uygulamasıdır. 1 Aylık Erişim / 200 İstem: 18.900 TL. 1 saat işlem veya açık panel: 1.990 TL. KDV dahil. Tek seferlik satın alım; otomatik yenileme yok.
```

### DE · `appsIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe ist die Windows-Anwendung. Monatsabo 200 USD für 500 Prompts. Nutzung 10 USD für 1 Stunde Verarbeitung oder ein offenes Panel.
```
Yeni:
```text
US.exe ist die Windows-Anwendung. 1 Monat / 200 Prompts: 18.900 TRY. 1 Stunde Verarbeitung oder ein offenes Panel: 1.990 TRY. Inklusive Mehrwertsteuer. Einmaliger Kauf; keine automatische Verlängerung.
```

### RU · `appsIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe — приложение для Windows. Подписка 200 USD в месяц на 500 запросов. Оплата использования 10 USD за 1 час обработки или открытую панель.
```
Yeni:
```text
US.exe — приложение для Windows. 1 месяц / 200 запросов: 18 900 TRY. 1 час обработки или открытую панель: 1 990 TRY. НДС включён. Разовая покупка; без автоматического продления.
```

### EN · `usIntro` · kullanılmayan sözlük girdisi
Eski:
```text
Download US.exe. Buy the license at shop.kanus.org. A paid order emails the license key. Monthly subscription is 200 USD for 500 prompts. Usage is 10 USD for 1 hour of processing or an open panel.
```
Yeni:
```text
Download US.exe. Buy the license at shop.kanus.org. 1 month / 200 prompts: 18,900 TRY. 1 hour of processing or an open panel: 1,990 TRY. VAT included. One-time purchases; no automatic renewal.
```

### TR · `usIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe indirilir. Lisans shop.kanus.org adresinden alınır. Ödenmiş sipariş anahtarı e-posta ile gönderir. Aylık abonelik 200 USD, 500 istem. Kullanım 10 USD, 1 saat işlem veya açık panel.
```
Yeni:
```text
US.exe indirilir. Lisans shop.kanus.org adresinden alınır. 1 Aylık Erişim / 200 İstem: 18.900 TL. 1 saat işlem veya açık panel: 1.990 TL. KDV dahil. Tek seferlik satın alım; otomatik yenileme yok.
```

### DE · `usIntro` · kullanılmayan sözlük girdisi
Eski:
```text
US.exe herunterladen. Die Lizenz gibt es auf shop.kanus.org. Eine bezahlte Bestellung schickt den Schlüssel per E-Mail. Monatsabo 200 USD für 500 Prompts. Nutzung 10 USD für 1 Stunde Verarbeitung oder ein offenes Panel.
```
Yeni:
```text
US.exe herunterladen. Die Lizenz gibt es auf shop.kanus.org. 1 Monat / 200 Prompts: 18.900 TRY. 1 Stunde Verarbeitung oder ein offenes Panel: 1.990 TRY. Inklusive Mehrwertsteuer. Einmaliger Kauf; keine automatische Verlängerung.
```

### RU · `usIntro` · kullanılmayan sözlük girdisi
Eski:
```text
Скачайте US.exe. Лицензия покупается на shop.kanus.org. Оплаченный заказ присылает ключ по почте. Подписка 200 USD в месяц на 500 запросов. Оплата использования 10 USD за 1 час обработки или открытую панель.
```
Yeni:
```text
Скачайте US.exe. Лицензия покупается на shop.kanus.org. 1 месяц / 200 запросов: 18 900 TRY. 1 час обработки или открытую панель: 1 990 TRY. НДС включён. Разовая покупка; без автоматического продления.
```

### Türkçe statik deneme sayfası
Canlı dosya `/home/u346856246/domains/kanus.org/public_html/deneme.html`.
Yerel orijinal kaynak bulunamadı; salt okunur kopya `/workspace/us-panel/reports/r20261108/pricing-evidence/deneme.html`.
Eski dize:
```text
Aylık abonelik 200 USD, 500 istem. Kullanım 10 USD, 1 saat işlem veya açık panel.
```
Yeni dize:
```text
1 Aylık Erişim / 200 İstem: 18.900 TL. 1 Saatlik Kullanım: 1.990 TL; 1 saat işlem veya açık panel. KDV dahil. Tek seferlik satın alım; otomatik yenileme yok.
```
Deneme sınırını (20) veya işlevsel kalan istem sayacını değiştirmeyin.

## 5. Yayın yöntemi ve geri dönüş
Doğrulanmış son yöntem: GLOBAL_SITE_UPLOAD_20261108.md ve GLOBAL_FX_UPLOAD_20261108.md: Hostinger connector `hosting_files_generate-upload-url` + TUS, hesap u346856246, hedef public_html, override=true. Git auto-deploy/Node build değil; Horizons değil. Bu tur upload URL üretilmedi ve hiçbir yükleme yapılmadı.

Yayın sırası (uygulama yetkisi verilmedi):
1. GLOBAL ile fx2/cron paralel çalışmasını kontrol et; index ve asset hash'lerini yeniden oku. Fark varsa dur ve güncel canlı üstüne yeniden hazırla.
2. Güncel index.html ve deneme.html için zaman damgalı yerel yedek al; bundle yerinde kalmalı. Yeni fiyat paketinin kendi rollback/index.html'i **bu turdaki fx sürümü** olmalı.
3. Yeni hash'li bundle'ı public_html/assets/ altına yükle. Yeni index.html script src'sini yeni bundle'a bağla, index'i en son yükle. deneme.html güncellemesi ayrı yükleme ve ayrı yedek gerektirir.
4. TR/EN/DE/RU ana sayfa ve deneme.html render et; 200 USD/500 istem/10 USD ve abonelik iddiaları kalmadığını, KDV/tek seferlik ifadesini doğrula. /us teslimat uyarısı koşulsuz değişmemeli. Shopify'a dokunma.
5. Geri dönüş: yeni paketin rollback/index.html'ini public_html/index.html üzerine yükle; deneme.html eski kopyasını geri yükle. Eski asset silinmediği sürece index geri dönüşü yeterli. api/, storage/, downloads/, fx.json, .htaccess ve cron bu işe dahil değil.

Mevcut geçmiş yedekler:
- `/workspace/us-panel-dev/global-work/site/live_backup_20261108/index.html.live_root` (i18n öncesi).
- `.../index.html.release_us` (fx öncesi).
- `.../release_us_fx/rollback/index.html` de fx öncesine döndürür: **yeni fiyat işi için doğru bir-adım rollback değildir**.
- Bu tur güncel fx index salt okunur kopyası: `/workspace/us-panel/reports/r20261108/pricing-evidence/index.html` (hash yukarıda); deneme kopyası aynı dizinde.

## 6. Açık sorular / tamamlanmamış doğrulama
1. Demir yalnız mevcut ana sayfa fiyat metnini düzeltmeyi istedi. Mağazada olup sitede olmayan 5/10/25/50/100 kredi paketleri ayrıca sitede listelensin mi? Yeni kart ve içerik eklemek kapsam genişlemesidir; otomatik eklenmedi.
2. 1 kredi ürün DRAFT olduğundan 149 TL satın alınabilir paket gibi gösterilmemeli. Bilgi amaçlı satış dışı satır isteniyor mu? Varsayılan: ekleme.
3. 20 istem bulut denemesi mağaza SKU'su değil. Ücretsiz deneme politikası devam ediyor mu? Şimdilik olduğu gibi koru; fiyat atama.
4. Orijinal SPA kaynak deposu veya deneme.html kaynak dosyası GLOBAL'de nerede? Şu an yalnız derlenmiş asset tabanlı yeniden paketleme yolu doğrulandı.
5. Bu tur yeni tam tarayıcı render yapılmadı; önceki aynı hash'li 4 dil / 4 sayfa raporu ve canlı metin/bileşen eşleştirmesi kullanıldı. /applications, /deneme.html ve dinamik forum/galeri/haber içerikleri için yeni render kontrolü yayın kapısıdır.

Sonuç: Ana sayfada zorunlu 4 görünür dize + deneme.html'de 1 görünür dize; ayrıca 8 atıl sözlük dizesi temizliği önerildi. Hiçbir canlı değişiklik yapılmadı.

## 7. UYGULANDI — 01:1x TSİ (yalnız TR)
Onay: Demir, SHOP sohbetinde 01:10 TSİ: "Türkçe sayfada TL fiyatları koy, diğer dillerde USD fiyatları şimdilik değiştirme".
- Yeni asset: public_html/assets/index-uscommerce-i18nfx-trtl-ee14acd6.js (sha ee14acd6…). Yalnız TR usCard, TR appsIntro, TR usIntro değişti. EN/DE/RU aynen (USD).
- index.html → yeni asset'e bağlı (sha 8fd6127d…). deneme.html TR fiyat satırı güncellendi (sha a82cb9ed…).
- Yükleme öncesi canlı hash'ler değişmemişti (4ab2f920… / b15b6cd4…). Canlı GET ile hash eşleşmesi doğrulandı; node --check OK. Tarayıcı render yapılamadı (playwright yok).
- Geri dönüş: pricing-release-tr/rollback/index.html → public_html/index.html ve rollback/deneme.html → public_html/deneme.html (TUS, override=true). Eski asset 5a375f29 yerinde.
