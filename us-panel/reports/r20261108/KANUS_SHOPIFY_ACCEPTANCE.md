# Shopify otomasyon kabulü — §3

Denetim kesiti: 08.11.2026 00:26 TSİ (Europe/Istanbul). Canlı okuma ve izole yerel test; canlı değişiklik, VM start/stop, müşteri mesajı, ödeme/iade veya deploy yapılmadı. Kaynak sahipleri geliştirmeye devam ediyor; sonuçlar yalnız kayıtlı snapshot/hash ve gözlem anına aittir. 6 Kasım 13:38 TSİ geri yüklemesi nedeniyle mtime sürüm kanıtı değildir. Kanıt: `/workspace/us-panel/reports/r20261108/evidence/versions.json`, `/workspace/us-panel/reports/r20261108/evidence/work2-v11-manifest.json`.

Etiketler: VERIFIED=bu dar kapsamda doğrudan doğrulandı; IMPLEMENTED=uygulama/yerel kanıt var, canlı kabul yok; PARTIAL=kapsam eksik; BROKEN=kanıtlı başarısızlık; MISSING=belirtilen kapsamda yok; BLOCKED=erişim/önkoşul engeli; NOT TESTED=koşulmadı. PASS ticari kabulün tümüne taşınmaz.

## Başlangıç durumu
DNS artık **82.29.184.225**, TTL300 (00:21 TSİ); önceki “eski IP” bilgisi güncel değil. DNS değişimi, store aktarımının / kesinti süresinin / GPU tünelinin kabulü değildir. curl remote_ip=198.18.0.1 yerel proxy ara adresidir, origin IP diye kullanılamaz. Kanıt: /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json; /workspace/us-panel/reports/r20261108/evidence/hook-health.meta.

## Değiştirilen dosyalar
Ürün/altyapı dosyası değiştirilmedi. Bu rapor: `/workspace/us-panel/reports/r20261108/KANUS_SHOPIFY_ACCEPTANCE.md`. Ham kanıtlar: `/workspace/us-panel/reports/r20261108/evidence`; yerel test kopyaları: `/workspace/us-panel/reports/r20261108/test-snapshots`.

## Uygulanan çözüm
Sayfa vaatleri ile yerel/tarihsel backend kanıtları ayrıldı; yalnız health200 otomasyon kanıtı sayılmadı. /workspace/us-panel/reports/r20261108/evidence/hook-health.meta

## Gerçek testler
Shopify get-shop-info/search_products/get-product salt okuma; 7 açıklama kaydedildi. Public ürün uçlarının 6/7’si HTTP429, ürün JSON429; tekrar yük artırılmadı. Alternatif doğrudan yetkili MCP okuması başarılı. Kanıt: evidence/product-*.meta; evidence/shopify-*.txt.

## PASS/FAIL/BLOCKED sonuçları
| Kontrol | Durum | Kanıt / sınır |
|---|---|---|
| Mağaza + 7 aktif ürün | PASS | /workspace/us-panel/reports/r20261108/evidence/shopify-products.txt |
| HMAC / idempotency / atomic ledger / reversal / expiry | PASS (yerel/tarihsel sandbox) | /workspace/us-panel/reports/r20261108/evidence/work2-v11-test.log; /workspace/us-panel/reports/E2E_HOOK_KANUS_TEST_20261106.md |
| Yeni IP test order→kullanılabilir hesap/lisans→mail→model | BLOCKED | /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json; /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md §9 |
| Mail kuyruğu/retry/mutabakat/fraud | PARTIAL | /workspace/us-panel/reports/QA_TEST_RAPORU_20261106.md; /workspace/us-panel/reports/SMTP_HOSTINGER_20261106.md |
| Canlı webhook abonelik sahipliği/target ayarı | NOT TESTED | Shopify MCP keşfinde webhook okuma aracı yok; /workspace/us-panel/reports/SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md |


## Kalan eksikler
Yeni hostta tam test siparişi/hesap/SMTP/DeepSeek/ledger zinciri yok; metinleri otomasyon tamamlandı diye değiştirmek erken. Tarihsel sandbox refund/paid kanıtları yeni hostun E2E kabulü değildir. Kanıt: /workspace/us-panel/reports/E2E_HOOK_KANUS_TEST_20261106.md; /workspace/us-panel/reports/r20261108/evidence/dns-api-test.json

## Gereken kullanıcı onayları
Mevcut doğrudan onaylar: HOST 82.29.184.225 taşıması; KAN v4+discard kodu ve redeploy; PERF taşıma sonrası yalnız sentetik hesapla ≤20 eşzamanlı, ≤10 dk; Windows yalnız Demir bilgisayarı. Bunlar bu denetimde kullanılmadı. Yeni ücretli kaynak/VM operasyonu, kapsam dışı prod deploy, 25/50 gerçek yük, gerçek ödeme/iade, canlı fiyat, müşteri e-postası, veri silme ve erişim genişletme ayrıca açık onay ister. Kanıt: üst konuşma kullanıcı yanıtları; prompt §14 (`/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt`).

## Canlı sayfa tam ifadeleri ve URL
- https://kanus.org/us : “US için Shopify üzerinden lisans ve hesap erişimi hazırlanıyor.”
- Aynı URL: “Otomatik hesap açılışı ve lisans teslimi henüz etkin değil. Satın alma öncesinde destek ekibiyle iletişime geçin.”
- https://shop.kanus.org : “İhtiyacınıza uygun tek seferlik paketi seçip siparişinizi oluşturun. Satın alma otomatik yenilenmez. API anahtarınız, siparişinizden sonra en geç 10 dakika içinde e-posta ile gönderilir.”
- Aynı URL: “Anahtarlar sipariş sonrası manuel olarak iletilir.”
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/site-us-automation-quotes.txt` (canlı yüklenen JS /assets/index-uscommerce.js?v=20261104-usticari); `/workspace/us-panel/reports/r20261108/evidence/shop.text`. Browser render bu denetimde yapılmadı; SPA bundle metni gözlendi.

Aktif ürün açıklamalarında “manuel” ifadesi bulunmadı; aylık/saatlik ürünler ≤10dk e-posta vaadini koruyor. Manuel metin mağaza ana sayfasında. Backendin yerel hesap/lisans/ledger yetenekleri metinden ileri; fakat gerçek canlı otomasyon zinciri kanıtlanmadığından metin “yanlış” veya otomasyon “tamam” sayılamaz. Kanıt: aşağıdaki tam açıklamalar ve QA raporu.

## Ürün URL listesi ve tam açıklamalar
- https://shop.kanus.org/products/kanus-ai-1-aylik-erisim-200-istem
- https://shop.kanus.org/products/kanus-ai-1-saatlik-kullanim
- https://shop.kanus.org/products/kanus-ai-5-model-i̇stegi
- https://shop.kanus.org/products/kanus-ai-10-model-i̇stegi
- https://shop.kanus.org/products/kanus-ai-25-model-i̇stegi
- https://shop.kanus.org/products/kanus-ai-50-model-i̇stegi
- https://shop.kanus.org/products/kanus-ai-100-model-i̇stegi

### Kanus AI – 1 Aylık Erişim / 200 İstem
Tam açıklama (MCP get-product):
> kanus.org üzerinden erişilebilen yapay zekâ paneli için bir aylık API erişim paketi.
> 
> 
> 
> 
> Bir ay boyunca toplam 200 istem hakkı.
> 
> 
> Tek seferlik satın alma; otomatik yenileme yoktur.
> 
> 
> API anahtarınız, siparişinizden sonra en geç 10 dakika içinde e-posta ile gönderilir.
> 
> 
> Dijital erişim ürünüdür; fiziksel gönderim yapılmaz.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-monthly.txt`.

### Kanus AI – 1 Saatlik Kullanım
Tam açıklama (MCP get-product):
> kanus.org üzerinden erişilebilen yapay zekâ paneli için bir saatlik kullanım paketi.
> 
> 
> 
> Bir saat işlem veya açık panel kullanımı.
> 
> 
> Tek seferlik satın alma; otomatik yenileme yoktur.
> 
> 
> API anahtarınız, siparişinizden sonra en geç 10 dakika içinde e-posta ile gönderilir.
> 
> 
> Dijital erişim ürünüdür; fiziksel gönderim yapılmaz.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-hourly.txt`.

### Kanus AI — 5 Model İsteği
Tam açıklama (MCP get-product):
> 5 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 90 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.
> 
> İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-credit5.txt`.

### Kanus AI — 10 Model İsteği
Tam açıklama (MCP get-product):
> 10 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 90 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.
> 
> İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-credit10.txt`.

### Kanus AI — 25 Model İsteği
Tam açıklama (MCP get-product):
> 25 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 180 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.
> 
> İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-credit25.txt`.

### Kanus AI — 50 Model İsteği
Tam açıklama (MCP get-product):
> 50 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 180 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.
> 
> İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-credit50.txt`.

### Kanus AI — 100 Model İsteği
Tam açıklama (MCP get-product):
> 100 model API isteği için ön ödemeli kredi. Kod, satın alma tarihinden itibaren 365 gün geçerlidir. Tek bir panel görevi birden fazla model isteği kullanabilir.
> 
> İstek başına en fazla 32 KiB metin girdisi ve 2.048 çıktı tokenı. Fiyata KDV dahildir.
Kanıt: `/workspace/us-panel/reports/r20261108/evidence/shopify-credit100.txt`.

## Kanıt bütünlüğü
Yerel gözlem/test kayıtlarının SHA256 manifesti: `/workspace/us-panel/reports/r20261108/EVIDENCE_SHA256.json`. Değişebilen ekip raporlarının denetim anı kopyaları: `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots`. Manifest ham kanıtları ve ekip raporu kopyalarını kapsar; bu raporların içerik hashlerini kapsamaz.
