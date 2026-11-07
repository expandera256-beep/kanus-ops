# US Panel QA scratch harnessları

Bu dizin ürün deposunu değiştirmez. Python standart kitaplığı ve mevcut Go toolchain yeterlidir. Go private panel fonksiyonları `go test -overlay` ile **yalnız derleme sırasında** qa/ test dosyaları eklenerek sürülür; work2/source/deploy dosyaları yazılmaz. Gerçek hook mevcut kaynaktan qa/build/hook olarak derlenir ve loopback üzerinde geçici ledger, rastgele sadece-test sırları, SMTP kapalı, yerel mock upstream ile başlar. İş bitiminde süreç ve geçici ledger temizlenir. VPS servisine/restartına dokunulmaz.

## §8
```bash
# Sadece yerel: iki hesap/oturum, cross-user redeem/settle/history, auth sonrası
# query kontrolleri, gerçek panel dosya/terminal sınırları, IP limiter429,
# 1/5/10/25/50 loopback SSE yükü ve kredi/TTFT/toplam süre.
python3 /workspace/us-panel/qa/sec8/run.py

# Aynı yerel testler + sınırlı api-test kontrolleri:
python3 /workspace/us-panel/qa/sec8/run.py --remote

# Yalnız api-test kontrolleri (status ve login için ayrı en fazla12 rate probe, concurrency2):
python3 /workspace/us-panel/qa/sec8/run.py --remote-only
```
Uzak webhook yalnız `{}` ve eksik/yanlış imza kullanır; gerçek webhook sırrı okunmaz. Private QA lisansları varsa yalnız çalışma anında `/workspace/secrets/qa-license-a.txt` ve `qa-license-b.txt` dosyalarından okunur; log/URL/rapora yazılmaz. Uzak lisans testleri yalnız status GET; model çağrısı veya reserve/settle yapılmaz, kredi harcanmaz. B credentialı auth-confusion kontrolü için URL yerine GET gövdesindedir; bearer otoritesi A kalmalıdır. Bu kontrol B dosya/oturum API kabulü değildir. Çıktı yalnız A/B bakiye ve kredi farkını gösterir. İki QA lisansı sonuç alındıktan sonra operatörce revoke edilebilir; harness revoke etmez.

Yerel account limiterin429u, uzak nginx limiterin doğrulandığı anlamına gelmez. Uzak status ve yanlış-login yüzeylerinde ayrı12 deneme/concurrency2 sınırı vardır. Login yalnız example.invalid sahte kimliği kullanır; hesap oluşturmaz/e-posta göndermez. Düşük yükte429 yoksa NOT TESTED (kanıtlanamadı) yazılır; otomatik yük artırımı yoktur. Query parametresinin yok sayılması tam command injection/path traversal kabulü değildir. Dosya/terminal gerçek panel fonksiyonları ayrıca overlay testlerinden geçer; OS/container sandbox kabulü değildir.

## §9
```bash
# Güvenli varsayılan: yerel mock, sıfır dış model çağrısı
python3 /workspace/us-panel/qa/e2e9/run.py

# Sonradan, hazır tünel + operatör test bütçesi olduğunda (bu tur koşulmadı):
US_REAL_MODEL_E2E=1 US_PROXY_URL=https://api-test.kanus.org \
US_TEST_LICENSE=/absolute/private/test-license.txt \
python3 /workspace/us-panel/qa/e2e9/run.py --real
```
`--real`, env=1 ve tam api-test URL birlikte zorunludur. Mock kipinde bu env yanlışlıkla ayarlanmış olsa bile remote çağrı yapılmaz. Anahtar dosyası içerikleri basılmaz.

Mock OpenAI `/v1/models` + JSON/SSE `/v1/chat/completions` sunar. Deterministik yanıtlar3 dosyalı Python REST DEMO projesidir. Harness kendi başına dosyaları proje köküne yazmaz: panel task queue -> model proxy -> review -> atomic apply yapar. Gerçek py_compile ve gerçek loopback HTTP unittest çalışır. Mock bilinçli health bugı üretir; gerçek test hatası tekrar panel ajanına gönderilir; model read_file ardından SHA256 patch_file döndürür. Test dosyası değişmeden4 test yeniden geçer. Panelin kendi project-tests eylemi ve modelden bağımsız HTTP kabul testi de çalıştırılır. Artefaktlar qa/build/e2e9-*/generated altında saklanır.

Bu küçük REST DEMO, §9un frontend/PostgreSQL/Docker/CI-CD içeren büyük platform kabulü değildir. Mock model yetenek kabulü de değildir.

## Kanıt
Varsayılan log dizini: `/workspace/us-panel/reports/qa-20261106/`. `--logs /absolute/path` ile değiştirilebilir. Başarısız assertion exit1; BLOCKED/NOT TESTED sonuçları JSONda açıkça yazılır, dolayısıyla yalnız exit0 üretim kabulü sayılmaz.
- sec8-local-results.json / sec8-remote-results.json / sec8-results.json
- sec8-panel-boundaries.log
- e2e9-results.json / e2e9-panel-driver.log / e2e9-independent-acceptance.log
- QA_TEST_RAPORU_20261106.md ve S12_TEST_MATRIX_20261106.md

İlk sec8 koşumunda harness max_tokens alanını atlamış ve gerçek hook doğru biçimde400 vermiştir. Fixture düzeltildi; ilk başarısız loglar `sec8-initial-fixture-error*` adıyla korunur. Ürün güvenlik açığı olarak sunulmaz.

## Yetkili sentetik A/B kredi akışı (bir kez)
```bash
# Yalnız operatörün sağladığı fresh10/10 QA lisansları; A1 kredi, B0.
# Aynı scripti tekrar koşmak için mevcut test sonucunu koruyun; yeniden debit istenmiyorsa KOŞMAYIN.
python3 /workspace/us-panel/qa/sec8/remote_credit.py --allow-one-test-credit
```
Fresh10/10 ve reserved0 önkoşulu; A↔B gerçek reservation settle yetkisi409, A tek commit + duplicate commit aynı debit, B release + duplicate release bakiye dönüşü, A kalan kotasına kısa concurrency2 reserve yarışı ve finally bütün temporary holdların release edilmesi. Gerçek ödeme/sipariş/VPS/admin yok. Kredi sonuçları sec8-remote-credit-results.json ve logunda; anahtarlar basılmaz.

### KAN.US A/B lisanslarını revoke ettikten SONRA
```bash
# Şimdi çağrıldığında sadece pending NOT TESTED kaydeder, ağ çağrısı yapmaz:
python3 /workspace/us-panel/qa/sec8/post_revoke.py
# Yalnız KAN.US revoke tamamladıktan sonra (varsayılan B):
python3 /workspace/us-panel/qa/sec8/post_revoke.py --after-revoke
# A ve B birlikte:
python3 /workspace/us-panel/qa/sec8/post_revoke.py --after-revoke --license both
```
Varsayılan B; --license A veya both ile açık seçim. Script hiçbir revoke/admin/VPS işlemi yapamaz. Status kesin200 valid=false/revoked=true ve sıfır kredi/hold bekler. /v1/chat, /v1/chat/completions, models ve reserve401/402/403 bekler; boş chat gövdesi auth regresyonunda bile model dispatchini önler. Önce/sonra bütün kredi alanları değişmemeli. Eski reservation IDleri tutulmadığı ve tüm holdlar kapandığı için var olan settlement replayi NOT TESTED; yalnız nonexistent-ID commit/release reddi401/402/403/409 ayrıca kontrol edilir.409 burada resource reddidir, revoked credential auth kanıtı değildir. A/B revoke sonrası gerçek koşum sonuçları post-revoke-A-B.log ve post-revoke-results.json içinde.
