Yerini R13 aldı
DURUM (2026-11-08 01:10 TSİ): BEKLEMEDE — .12 PANEL DEV FAIL (kredi tükenmesi/revoke/iptal sonrası GPU wake; model_proxy.go:127–135, 535–550; rapor /workspace/us-panel-dev/PANEL_S2_REVIEW_R12_20261108.md). Deploy yok. Bir sonraki sürüm için PANEL DEV PASS + Demir onayı gerekli.

# .12 deploy planı — api-test (us-license-hook-test, prod VPS 82.29.184.225)
Durum: BEKLEMEDE: PANEL DEV PASS ve Demir'in .12 için yeni onayı gerekli. Restart ayarı eklenmeyecek. Sunucuda hiçbir adım uygulanmadı.

## Paket
- /workspace/us-panel/dist/2026.11.06.12/us-license-hook_hostinger.tgz sha256 b5489468efddfde838f940e64b40c2e512788d64acd880a860ea85a1822f539d (SHA256SUMS ile eşleşti)
- Yeni binary us-license-hook (= us-license-hook_linux) sha256 b2d7ab28f65dc98b6b75373f867575b73c030547b7745e8cf71a527dd1ec2c72, 11411690 B
- Mevcut: /opt/us-license-hook-test/us-license-hook.real sha256 7d72f449…56f0 (r10_prev ile aynı)

## Açık konular (onaydan önce)
1. İzin: /etc/us-model/{wake.token,wake.pin} uslic:uslic 600. Test servisi uslic-test ile çalışıyor → okuyamaz, wake fail-closed olur.
   Seçenek A (önerilen): /etc/us-model-test/ altına kopya, uslic-test:uslic-test 600; env yolları buna göre. (Token'ın ikinci kopyası → Demir onayı.)
   Seçenek B: dosyalara uslic-test için ACL read (prod hook dosyasını paylaşır).
2. GPU maliyeti: wake açık hedefleniyor (US_GPU_WAKE_DISABLE=0), ancak .11 için verilmiş onay .12'yi kapsamaz. .12 kredi kapısı PANEL DEV tarafından PASS edilmeden ve Demir yeni onay vermeden deploy yok. Smoke sırasında chat/GPU çağrısı yok.
3. pin_ok=yes log satırı yalnız gerçek wake/status çağrısında yazılır; GPU uyandırmadan log'da görülemez. Bunun yerine readiness: dosya izinleri, pin'in :8443 sertifikasıyla eşleşmesi (salt TLS el sıkışması, /a100 çağrısı yok), servis log'unda config hatası olmaması.
4. Restart ayarı eklenmeyecek; bu deploy kapsamına alınmayacak.

## Adımlar
0. Ön kontrol: tgz sha, binary sha, nginx -t, servis active, healthz 200.
1. Yedek (yazma dondurma ile tutarlı): systemctl stop us-license-hook-test → T0;
   cp -a licenses.json /opt/us-license-hook-test/licenses.json.bak-pre-r12-<ts>; box kopyası hook-move/backup/; env yedeği /etc/us-license-hook-test.env.bak-pre-r12-<ts>; binary yedeği us-license-hook.real.r10-7d72f449.
2. Binary: install -m 755 -o uslic-test -g uslic-test yeni binary → us-license-hook.real.new; sha doğrula; mv -f .new → us-license-hook.real (wrapper değişmez).
3. ENV (/etc/us-license-hook-test.env'e ekle, mevcut değerler ezilmez):
   US_GPU_WAKE_DISABLE=0 (Demir: wake açık kalsın)
   US_GPU_WAKE_URL=https://34.41.186.170:8443/a100/start
   US_GPU_STATUS_URL=https://34.41.186.170:8443/a100/status
   US_GPU_WAKE_PIN_FILE=/etc/us-model-test/wake.pin   (madde 1 A)
   US_GPU_WAKE_TOKEN_FILE=/etc/us-model-test/wake.token
   US_GPU_WAKE_USE_SYSTEM_CA=0
   US_GPU_WAKE_HTTP_ALLOW=
   US_GPU_WARMUP_TIMEOUT=15m
   US_BCRYPT_CONCURRENCY=2  US_BCRYPT_WAIT_SECONDS=10
   US_LOGIN_ACCOUNT_LIMIT=10  US_LOGIN_ACCOUNT_WINDOW_SECONDS=900
   US_STRICT_DONE=1
   SMTP ayarları ve allowlist aynı kalır.
4. systemctl start us-license-hook-test; journalctl kontrolü (başlangıç hatası yok).
5. Smoke: healthz 200 (local + public TLS); imzasız/yanlış HMAC 401; plan dışı imzalı webhook 200; qa-license-a/b /v1/credits/status 200; sahte lisans 401; /account/login geçersiz 401; store sayaçları değişmedi; /proc/<pid>/exe sha = b2d7ab28…; wake readiness (madde 3); chat çağrısı YAPILMAZ.
6. Raporla: kesinti, sha'lar, smoke, yedek yolları.

## Geri dönüş (7d72f449)
systemctl stop us-license-hook-test; cp -a us-license-hook.real.r10-7d72f449 us-license-hook.real; env'i .bak-pre-r12 ile geri koy; store'u yalnız bozulma varsa .bak-pre-r12'den geri koy (aradaki yeni kayıtlar kaybolur → önce diff); start; healthz + smoke.


## R12 paket doğrulaması — 08.11.2026 01:03 TSİ
- Dist: `/workspace/us-panel/dist/2026.11.06.12/`.
- SHA256SUMS dosyası SHA256: `38e96f7b3f147d06c28487d9ff6c9ba0859d73486285b9ca503bc591cf24b981`; bildirilen prefix/suffix eşleşti.
- `sha256sum -c SHA256SUMS`: 5/5 OK (exe, linux, tgz, TREE_MANIFEST, teslim raporu).
- SPKI kaldırılması **teslim notlarından doğrulandı**: `dist/2026.11.06.12/TESLIM_RAPORU_TR.md:14`, “PEM tek desteklenen mod”, `gpu_pin.go:137 pin_spki_unsupported`; `TestGPUPinSPKIAndPEMAndRotation` PEM kabul/3 SPKI red olarak bildirilmiş. Kaynak kodu bu tur ayrıca denetlenmedi; bağımsız PANEL DEV PASS kapısı korunur.

## Pin biçimi kapısı ve geçiş (yalnız plan, UYGULANMADI)
- Prod `/etc/us-model/wake.pin`: `head -c 27` çıktısı program içinde PEM başlığıyla karşılaştırıldı, çıktı gösterilmedi; **hash-like, 53 byte**. Bu biçim ölçümü SPKI içeriğinin geçerliliğini tek başına kanıtlamaz.
- `/etc/us-model-test/wake.pin` yok; `/etc` altında başka wake.pin kopyası bulunmadı. İçerik/token/PIN hiçbir rapora konulmaz.
- Eski “ACL veya prod pini doğrudan kopyala” seçenekleri R12 için geçerli değil; PEM test kopyası gerekir. Prod pin/token dosyaları olduğu gibi kalır.
- PASS + Demir'in .12 ve test wake kopyası için yeni onayından sonra: varsa test pin/token/env dosyalarını root-only `/root/uslic-test-pre-r12-<ts>/` altına `cp -a` ile yedekle; yokluk bilgisini kaydet.
- Geçerli pin SPKI ise yalnız TLS el sıkışmasıyla `openssl s_client -connect 34.41.186.170:8443 -servername 34.41.186.170 -showcerts </dev/null` üzerinden leaf PEM'i root-only geçici dosyaya al; stdout/stderr chat/loga dökülmez. /a100/start veya /status isteği yapılmaz.
- `openssl x509 -pubkey -noout` → `openssl pkey -pubin -outform DER` → SHA256 ile SPKI özetini üret. Mevcut pinin desteklenen hex/base64/sha256-prefiks kodlamasını yalnız bellekte normalize edip aynı 32 byte özetle sabit-zamanlı karşılaştır. Yalnız MATCH/FAIL yaz; pin/sertifika/özet içeriğini yazdırma. Normalize edilemiyorsa veya eşleşmiyorsa DUR; yeni sertifikaya kör güven yok, prod dosya değişmez.
- Eşleşme sonrası PEM leaf geçerliliği ve IP SAN uyumunu kontrol et; test için `/etc/us-model-test/wake.pin` dosyasını owner/group `uslic-test:uslic-test`, mode `600` oluştur. Token da onaylı test kopyası olarak 600/aynı owner; prod dosyalarına ACL/owner değişikliği yok. Yeni kullanıcı oluşturulmaz.
- Test servisinin env'i bu test PEM yolunu kullanır. Okunabilirlik ve .12 PEM yükleme kontrolü GPU çağrısı yapmadan doğrulanır. Sistemde PEM alınmış olması SPKI eşleşmesi yerine geçmez.

## Login 429 Retry-After — öneri diff, UYGULANMADI
PERF kaynağı `KANUS_PERF_REPORT.md:98–104`: 23 login429'da Retry-After görünmedi; limiter kaynağı kesinleşmedi. Güncel config salt okunur alındı: site HTTPS server -> location / -> test-locations snippet; login burst2, zone uslic_test_account, rate5r/m. Global limit_req_status 429 zaten var; global dosya değişmeyecek.

Yalnız api-test sitesi ve onun özel snippet'i değişir. Önce `/etc/nginx/sites-available/api-test.kanus.org.bak-pre-r12-<ts>` ve `/etc/nginx/snippets/us-license-hook-test-locations.conf.bak-pre-r12-<ts>` yedekleri alınır. Başka site veya ortak rate-limit conf dosyasına dokunulmaz.

```diff
--- /etc/nginx/snippets/us-license-hook-test-locations.conf
+++ /etc/nginx/snippets/us-license-hook-test-locations.conf (öneri)
@@ location = /account/login {
     limit_req zone=uslic_test_account burst=2 nodelay;
+    limit_req_status 429;
+    proxy_intercept_errors on;
+    error_page 429 = @apitest_login_429;
     client_max_body_size 16k;
--- /etc/nginx/sites-available/api-test.kanus.org
+++ /etc/nginx/sites-available/api-test.kanus.org (öneri)
@@ HTTPS server, location / bloğundan sonra, server içinde
+    location @apitest_login_429 {
+        internal;
+        default_type application/json;
+        add_header Retry-After "900" always;
+        add_header Cache-Control "no-store" always;
+        return 429 '{"error":"rate_limited"}';
+    }
```
900 saniye konservatif sabit geri çekilme önerisidir (account pencere hedefi 900 sn); limiterin ölçülmüş kalan süresi değildir. IP 5r/m için gereğinden uzun olabilir. `proxy_intercept_errors on` hem nginx hem upstream login429'u bu yanıtla normalleştirir; mevcut upstream 429 body/Retry-After varsa değişeceği için uygulama öncesi PANEL DEV/PERF incelemesi gerekir. Başka route'un 429'u değişmez. Mevcut güvenlik header'larının nginx add_header kalıtımı kontrol edilip aynı etkili başlıklar bu named location'da korunur.

Uygulama sırası (yalnız yeni onaydan sonra): yedek -> diff -> `nginx -t`; FAIL ise yedekleri geri koy, reload yok. PASS ise yalnız nginx config reload (VM restart yok); reload ayrıca onay kapsamına alınır. Login limiter için kontrollü sentetik doğrulama ayrı test iznine tabidir; 429 + Retry-After + güvenlik header'ları ve diğer route davranışı doğrulanır. Bu tur nginx -t/reload veya HTTP probu yapılmadı.

## R12 tam rollback
1. Test servisini durdur; binary'yi `us-license-hook.real.r10-7d72f449` yedeğinden ve env'i `.bak-pre-r12-<ts>` dosyasından geri koy; binary SHA256 eski `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0` ile karşılaştır.
2. Test pin/token kopyalarını `/root/uslic-test-pre-r12-<ts>/` yedeğinden owner/mode ile geri koy. Önceden yoksa yalnız bu deploy'da oluşturulan test kopyasını kaldırarak yokluk durumunu geri getir; prod `/etc/us-model/wake.pin` hiç değişmediği için restore edilmez.
3. Api-test site ve özel snippet'i `.bak-pre-r12-<ts>` yedeklerinden geri koy; `nginx -t` PASS olmadan reload yok. Başka site değişmez.
4. Test servisini başlat, healthz ve GPU'suz smoke; nginx başlıkları ve eski binary hash doğrulaması. Store restore yalnız bozulma varsa ayrıca onay/diff ile; yeni kaydı kaybettirecek kör restore yok. Restart= ayarı eklenmez; prod hook :8787 ve VM yaşam döngüsüne dokunulmaz.
