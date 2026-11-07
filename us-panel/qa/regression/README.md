# R11 QA regresyonu — yalnız api-test

Şimdi güvenli plan/doğrulama (DNS/ağ/sır okuma/alt süreç yok):

```bash
python3 /workspace/us-panel/qa/regression/run.py --dry-run
```

**Aşağıdaki komut henüz ÇALIŞTIRILMADI.** HOST .11 hook'u yeni IPye taşıyıp gerçek hashli attestation sağladıktan ve QA mutasyon/lockout izni verildikten sonra tek komut:

```bash
python3 /workspace/us-panel/qa/regression/run.py --execute \
  --attestation /workspace/us-panel/reports/qa-20261108/r11-host-attestation.json \
  --fixtures /workspace/us-panel/qa/regression/fixtures.example.json \
  --allow-qa-mutations --account-limit \
  --limiter-evidence /workspace/us-panel/reports/qa-20261108/r11-limiter-ip-evidence.json
```

Attestation şablonu: `host-attestation.example.json` (bilerek geçersiz örnek; migration_complete=false/hash placeholder). Fixture şablonu sadece dosya yolları içerir. İki mevcut owned lisans taşınmış olmalı; yeni `r11-fresh-license.txt` ve revoked `r11-revoked-license.txt` koordinatörden gelmeli. Sadece açık izinli QA secret dosya adları okunur; webhook/session sırları asla okunmaz. Önceki fresh order404 artık owned fixture Bdir, **fresh fixture değildir**.

DNSnin tüm A/AAAA sonuçları yalnız82.29.184.225 olmalı. Her bağlantı gerçek peer IPyi kontrol eder; HTTPS SNI/sertifika api-test için doğrulanır. Doğrudan public HTTPS yoludur; proxyye/başka adrese otomatik fallback yoktur. Böyle bir yol erişilemiyorsa BLOCKED; HOST-onaylı erişilebilir test runner gerekir. Aktif ağ kontrolleri bu tur yapılmadı.

Varsayılan max2; sadece fresh ilk-claimde2 paralel, diğer çağrılar sıralı.75 istek tavanı,120s timeout,5xx/network exception/redirectte dur; retry/escalation yok. Reserve yalnız commit=false release ile kullanılır; bir kesinti olursa pending_holds sonuçta kaynak IDsiyle listelenir; otomatik tekrar/commit yok, HOST TTL/cleanup kontrol etmeli. İlk-claim kalıcı tek-koşum markerı oluşturur; otomatik silinmez veya tekrar edilmez. Register/reset/SMTP/admin/restart/webhook secret/model/chat/GPU işlemi yok.

Account-limit fazı en sonda; QA1i geçici kilitler.10 adet **geçerli uzunlukta yanlış parola**401,11.429+Retry-After beklenir. E-mail case/boşluk varyantları aynı normalized hesaba gitmelidir.13s minimum ara nginx etkisini azaltır; aynı kaynak IP Go IP bucketını hâlâ doldurabilir. Erken429 → NOT TESTED/confounded, yük artırılmaz.900s pencere aşılırsa NOT TESTED. Önceki tüm QA girişleriyle çakışmayan sakin pencere gerekir.

**Client XFF/X-Real-IP spoof etmek kaynak IP değiştirmez.** Varying-IP PASS için HOST aynı koşumun `X-QA-Run-ID` ve `X-QA-Step` başlıklarını access-logla ilişkilendirip aşağıdaki JSONu koşum bitmeden belirtilen dosyaya koymalı. Sunucu zaten böyle loglamıyorsa ek config bu workerın işi değildir; HOST kendi onaylı gözlem yöntemiyle kanıt sağlamalı. Dosya yoksa doğru11.429 paterni ayrı PASS olabilir ama varying-IP kabulü NOT TESTED kalır. Tek başına11.429 per-account kanıtı değildir.

```json
{
  "run_id": "RUNNERIN_SONUC_JSONUNDAKI_ID",
  "server_ip": "82.29.184.225",
  "attempts": [
    {"attempt":1,"step":"account_limit_1","http":401,"remote_addr":"GERCEK_KAYNAK_IP_1"},
    {"attempt":11,"step":"account_limit_11","http":429,"remote_addr":"GERCEK_KAYNAK_IP_11"}
  ]
}
```

Bu kısaltılmış schema örneğidir: gerçek dosya sıralı11 kayıt ve11 farklı gerçek remote_addr içermeli. Header değil HOST kanıtı gerekir. Koşum `events` alanında step/run korelasyonu sağlar; token/parola/key/body kaydetmez. Sonuç `reports/qa-20261108/r11-regression-results.json`; FAIL exit1, BLOCKED exit2, NOT TESTEDler JSONda açıkça kalır. Dry-run sıfır ağ garantisini guard testiyle doğruladık; aktif runner henüz doğrulanmadı.

Detaylı kabul planı: `reports/QA_PLAN_R11_20261108.md`. GPU/gerçek model ve native Windows .11 testleri ayrı sonraki kabul fazlarıdır.
