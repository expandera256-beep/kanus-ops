Yerini R12 planı aldı
# .11 deploy planı — api-test (us-license-hook-test, prod VPS 82.29.184.225)
Durum: BEKLEMEDE — .11 PANEL DEV FAIL (2026-11-08 00:46, GPU wake kredi kontrolü). Kurulmayacak. .12 PASS alırsa aynı plan; .12 için Demir onayı yeniden alınacak. Restart ayarı eklenmeyecek.

## Paket
- /workspace/us-panel/dist/2026.11.06.11/us-license-hook_hostinger.tgz sha256 9db1386d…87f9 (SHA256SUMS ile eşleşti)
- Yeni binary us-license-hook (= us-license-hook_linux) sha256 12772d120e42a414b712cc22f6442e5e52404cbdb25725f36e1fae49e2662b47, 11392797 B
- Mevcut: /opt/us-license-hook-test/us-license-hook.real sha256 7d72f449…56f0 (r10_prev ile aynı)

## Açık konular (onaydan önce)
1. İzin: /etc/us-model/{wake.token,wake.pin} uslic:uslic 600. Test servisi uslic-test ile çalışıyor → okuyamaz, wake fail-closed olur.
   Seçenek A (önerilen): /etc/us-model-test/ altına kopya, uslic-test:uslic-test 600; env yolları buna göre. (Token'ın ikinci kopyası → Demir onayı.)
   Seçenek B: dosyalara uslic-test için ACL read (prod hook dosyasını paylaşır).
2. GPU maliyeti: .11'de wake varsayılan AÇIK. Model hazır değilken yetkili her /v1/chat isteği :8443/a100/start'a POST atar. Deploy sonrası QA/PERF chat yükü GPU'yu uyandırır. Canlı wake testine kadar US_GPU_WAKE_DISABLE=1 ile deploy edip KAN'ın testinde 0'a çekmek önerilir.
3. pin_ok=yes log satırı yalnız gerçek wake/status çağrısında yazılır; GPU uyandırmadan log'da görülemez. Bunun yerine readiness: dosya izinleri, pin'in :8443 sertifikasıyla eşleşmesi (salt TLS el sıkışması, /a100 çağrısı yok), servis log'unda config hatası olmaması.
4. Restart=on-failure/RestartSec=5 değişikliği ayrıca Demir onayı bekliyor (bu deploy'la birleştirilebilir).

## Adımlar
0. Ön kontrol: tgz sha, binary sha, nginx -t, servis active, healthz 200.
1. Yedek (yazma dondurma ile tutarlı): systemctl stop us-license-hook-test → T0;
   cp -a licenses.json /opt/us-license-hook-test/licenses.json.bak-pre-r11-<ts>; box kopyası hook-move/backup/; env yedeği /etc/us-license-hook-test.env.bak-pre-r11-<ts>; binary yedeği us-license-hook.real.r10-7d72f449.
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
5. Smoke: healthz 200 (local + public TLS); imzasız/yanlış HMAC 401; plan dışı imzalı webhook 200; qa-license-a/b /v1/credits/status 200; sahte lisans 401; /account/login geçersiz 401; store sayaçları değişmedi; /proc/<pid>/exe sha = 12772d12…; wake readiness (madde 3); chat çağrısı YAPILMAZ.
6. Raporla: kesinti, sha'lar, smoke, yedek yolları.

## Geri dönüş (7d72f449)
systemctl stop us-license-hook-test; cp -a us-license-hook.real.r10-7d72f449 us-license-hook.real; env'i .bak-pre-r11 ile geri koy; store'u yalnız bozulma varsa .bak-pre-r11'den geri koy (aradaki yeni kayıtlar kaybolur → önce diff); start; healthz + smoke.
